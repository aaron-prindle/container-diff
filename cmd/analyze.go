package cmd

import (
	"errors"
	"os"

	"github.com/GoogleCloudPlatform/container-diff/differs"
	"github.com/GoogleCloudPlatform/container-diff/utils"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
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

		if err := analyzeImage(args[0], analyzeArgs); err != nil {
			glog.Error(err)
			os.Exit(1)
		}
	},
}

func analyzeImage(imageArg string, analyzerArgs []string) error {
	image, err := utils.ImagePrepper{imageArg}.GetImage()
	if err != nil {
		glog.Error(err.Error())
		cleanupImage(image)
		return errors.New("Could not perform image analysis")
	}
	analyzeTypes, err := differs.GetAnalyzers(analyzerArgs)
	if err != nil {
		glog.Error(err.Error())
		cleanupImage(image)
		return errors.New("Could not perform image analysis")
	}

	req := differs.SingleRequest{image, analyzeTypes}
	if analyses, err := req.GetAnalysis(); err == nil {
		glog.Info("Retrieving analyses")
		outputResults(analyses)
		if !save {
			cleanupImage(image)
		} else {
			dir, _ := os.Getwd()
			glog.Infof("Image was saved at %s as %s", dir, image.FSPath)
		}
	} else {
		glog.Error(err.Error())
		cleanupImage(image)
		return errors.New("Could not perform image analysis")
	}

	return nil
}

func init() {
	RootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().BoolVarP(&json, "json", "j", false, "JSON Output defines if the diff should be returned in a human readable format (false) or a JSON (true).")
	analyzeCmd.Flags().BoolVarP(&eng, "eng", "e", false, "By default the docker calls are shelled out locally, set this flag to use the Docker Engine Client (version compatibility required).")
	analyzeCmd.Flags().BoolVarP(&pip, "pip", "p", false, "Set this flag to use the pip differ.")
	analyzeCmd.Flags().BoolVarP(&node, "node", "n", false, "Set this flag to use the node differ.")
	analyzeCmd.Flags().BoolVarP(&apt, "apt", "a", false, "Set this flag to use the apt differ.")
	analyzeCmd.Flags().BoolVarP(&file, "file", "f", false, "Set this flag to use the file differ.")
	analyzeCmd.Flags().BoolVarP(&history, "history", "d", false, "Set this flag to use the dockerfile history differ.")
	analyzeCmd.Flags().BoolVarP(&save, "save", "s", false, "Set this flag to save rather than remove the final image filesystems on exit.")
}
