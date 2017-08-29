package cmd

import (
	"errors"
	"os"
	"sort"
	"sync"

	"github.com/GoogleCloudPlatform/container-diff/differs"
	"github.com/GoogleCloudPlatform/container-diff/utils"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare two images.",
	Long:  `Compares two images using the specifed differ(s) as indicated via flags (see documentation for available ones).`,
	Run: func(cmd *cobra.Command, args []string) {
		if validArgs, err := validateArgs(args); !validArgs {
			glog.Error(err.Error())
			os.Exit(1)
		}

		utils.SetDockerEngine(eng)

		analyzeArgs := []string{}
		allAnalyzers := getAllAnalyzers()
		for _, name := range allAnalyzers {
			if *analyzeFlagMap[name] == true {
				analyzeArgs = append(analyzeArgs, name)
			}
		}

		// If no differs/analyzers are specified, perform them all as the default
		if len(analyzeArgs) == 0 {
			analyzeArgs = allAnalyzers
		}

		if err := diffImages(args[0], args[1], analyzeArgs); err != nil {
			glog.Error(err)
			os.Exit(1)
		}
	},
}

func diffImages(image1Arg, image2Arg string, diffArgs []string) error {
	var wg sync.WaitGroup
	wg.Add(2)

	glog.Infof("Starting diff on images %s and %s, using differs: %s", image1Arg, image2Arg, diffArgs)

	var image1, image2 utils.Image
	var err error
	go func() {
		defer wg.Done()
		image1, err = utils.ImagePrepper{image1Arg}.GetImage()
		if err != nil {
			glog.Error(err.Error())
		}
	}()

	go func() {
		defer wg.Done()
		image2, err = utils.ImagePrepper{image2Arg}.GetImage()
		if err != nil {
			glog.Error(err.Error())
		}
	}()
	wg.Wait()
	if err != nil {
		cleanupImage(image1)
		cleanupImage(image2)
		return errors.New("Could not perform image diff")
	}

	diffTypes, err := differs.GetAnalyzers(diffArgs)
	if err != nil {
		glog.Error(err.Error())
		cleanupImage(image1)
		cleanupImage(image2)
		return errors.New("Could not perform image diff")
	}

	req := differs.DiffRequest{image1, image2, diffTypes}
	if diffs, err := req.GetDiff(); err == nil {
		// Outputs diff results in alphabetical order by differ name
		sortedTypes := []string{}
		for name := range diffs {
			sortedTypes = append(sortedTypes, name)
		}
		sort.Strings(sortedTypes)
		glog.Info("Retrieving diffs")
		diffResults := []utils.DiffResult{}
		for _, diffType := range sortedTypes {
			diff := diffs[diffType]
			if json {
				diffResults = append(diffResults, diff.GetStruct())
			} else {
				err = diff.OutputText(diffType)
				if err != nil {
					glog.Error(err)
				}
			}
		}
		if json {
			err = utils.JSONify(diffResults)
			if err != nil {
				glog.Error(err)
			}
		}
		if !save {
			cleanupImage(image1)
			cleanupImage(image2)

		} else {
			dir, _ := os.Getwd()
			glog.Infof("Images were saved at %s as %s and %s", dir, image1.FSPath, image2.FSPath)
		}
	} else {
		glog.Error(err.Error())
		cleanupImage(image1)
		cleanupImage(image2)
		return errors.New("Could not perform image diff")
	}

	return nil
}

func init() {
	RootCmd.AddCommand(diffCmd)
	diffCmd.Flags().BoolVarP(&json, "json", "j", false, "JSON Output defines if the diff should be returned in a human readable format (false) or a JSON (true).")
	diffCmd.Flags().BoolVarP(&eng, "eng", "e", false, "By default the docker calls are shelled out locally, set this flag to use the Docker Engine Client (version compatibility required).")
	diffCmd.Flags().BoolVarP(&pip, "pip", "p", false, "Set this flag to use the pip differ.")
	diffCmd.Flags().BoolVarP(&node, "node", "n", false, "Set this flag to use the node differ.")
	diffCmd.Flags().BoolVarP(&apt, "apt", "a", false, "Set this flag to use the apt differ.")
	diffCmd.Flags().BoolVarP(&file, "file", "f", false, "Set this flag to use the file differ.")
	diffCmd.Flags().BoolVarP(&history, "history", "d", false, "Set this flag to use the dockerfile history differ.")
	diffCmd.Flags().BoolVarP(&save, "save", "s", false, "Set this flag to save rather than remove the final image filesystems on exit.")
}
