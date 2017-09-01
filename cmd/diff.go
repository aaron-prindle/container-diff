package cmd

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/GoogleCloudPlatform/container-diff/differs"
	"github.com/GoogleCloudPlatform/container-diff/utils"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare two images: [image1] [image2]",
	Long:  `Compares two images using the specifed analyzers as indicated via flags (see documentation for available ones).`,
	Run: func(cmd *cobra.Command, args []string) {
		if validArgs, err := validateArgs(args, checkDiffArgNum, checkArgType); !validArgs {
			glog.Error(err.Error())
			os.Exit(1)
		}
		analyzeArgs := []string{}
		allAnalyzers := getAllAnalyzers()
		for _, name := range allAnalyzers {
			if *analyzeFlagMap[name] == true {
				analyzeArgs = append(analyzeArgs, name)
			}
		}

		// If no analyzers are specified, perform them all as the default
		if len(analyzeArgs) == 0 {
			analyzeArgs = allAnalyzers
		}

		if err := diffImages(args[0], args[1], analyzeArgs); err != nil {
			glog.Error(err)
			os.Exit(1)
		}
	},
}

func checkDiffArgNum(args []string) (bool, error) {
	var errMessage string
	if len(args) != 2 {
		errMessage = "'diff' requires two images as arguments: container diff [image1] [image2]"
		return false, errors.New(errMessage)
	}
	return true, nil
}

func diffImages(image1Arg, image2Arg string, diffArgs []string) error {
	cli, err := NewClient()
	if err != nil {
		return fmt.Errorf("Error getting docker client for differ: %s", err)
	}
	defer cli.Close()
	var wg sync.WaitGroup
	wg.Add(2)

	glog.Infof("Starting diff on images %s and %s, using differs: %s", image1Arg, image2Arg, diffArgs)

	imageMap := map[string]*utils.Image{
		image1Arg: &utils.Image{},
		image2Arg: &utils.Image{},
	}
	for imageArg, _ := range imageMap {
		go func(imageMap map[string]*utils.Image) {
			defer wg.Done()
			ip := utils.ImagePrepper{
				Source: imageArg,
				Client: cli,
			}
			image, err := ip.GetImage()
			imageMap[imageArg] = &image
			if err != nil {
				glog.Error(err.Error())
			}
		}(imageMap)
	}
	wg.Wait()

	if !save {
		defer cleanupImage(*imageMap[image1Arg])
		defer cleanupImage(*imageMap[image2Arg])
	}

	diffTypes, err := differs.GetAnalyzers(diffArgs)
	if err != nil {
		glog.Error(err.Error())
		return errors.New("Could not perform image diff")
	}

	req := differs.DiffRequest{*imageMap[image1Arg], *imageMap[image2Arg], diffTypes}
	if diffs, err := req.GetDiff(); err != nil {
		glog.Error(err.Error())
		return errors.New("Could not perform image diff")
	}
	glog.Info("Retrieving diffs")
	outputResults(diffs)

	if save {
		dir, _ := os.Getwd()
		glog.Infof("Images were saved at %s as %s and %s", dir, imageMap[image1Arg].FSPath,
			imageMap[image2Arg].FSPath)

	}
	return nil
}

func init() {
	RootCmd.AddCommand(diffCmd)
	addSharedFlags(diffCmd)
}
