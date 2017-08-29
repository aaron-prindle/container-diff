package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"

	"github.com/GoogleCloudPlatform/container-diff/utils"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

var analyzeFlagMap = map[string]*bool{
	"apt":     &apt,
	"node":    &node,
	"file":    &file,
	"history": &history,
	"pip":     &pip,
}

var json bool
var eng bool
var save bool

var apt bool
var node bool
var file bool
var history bool
var pip bool

var RootCmd = &cobra.Command{
	Use:   "container-diff",
	Short: "container-diff is a tool for comparing and analyzing container images",
	Long:  `container-diff is a tool for comparing and analyzing container images.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}

}

func init() {
	//TODO(aaron-prindle) see if this is still needed
	// pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
}

func outputResults(resultMap map[string]utils.Result) {
	// Outputs diff/analysis results in alphabetical order by analyzer name
	sortedTypes := []string{}
	for analyzerType := range resultMap {
		sortedTypes = append(sortedTypes, analyzerType)
	}
	sort.Strings(sortedTypes)

	results := make([]interface{}, len(resultMap))
	for i, analyzerType := range sortedTypes {
		result := resultMap[analyzerType]
		if json {
			results[i] = result.OutputStruct()
		} else {
			err := result.OutputText(analyzerType)
			if err != nil {
				glog.Error(err)
			}
		}
	}
	if json {
		err := utils.JSONify(results)
		if err != nil {
			glog.Error(err)
		}
	}
}

func cleanupImage(image utils.Image) {
	if !reflect.DeepEqual(image, (utils.Image{})) {
		glog.Infof("Removing image filesystem directory %s from system", image.FSPath)
		errMsg := remove(image.FSPath, true)
		if errMsg != "" {
			glog.Error(errMsg)
		}
	}
}

func getAllAnalyzers() []string {
	allAnalyzers := []string{}
	for name := range analyzeFlagMap {
		allAnalyzers = append(allAnalyzers, name)
	}
	return allAnalyzers
}

func validateArgs(args []string) (bool, error) {
	validArgNum, err := checkArgNum(args)
	if err != nil {
		return false, err
	} else if !validArgNum {
		return false, nil
	}
	validArgType, err := checkArgType(args)
	if err != nil {
		return false, err
	} else if !validArgType {
		return false, nil
	}
	return true, nil
}

func checkArgNum(args []string) (bool, error) {
	var errMessage string
	if len(args) < 1 {
		errMessage = "Too few arguments. Should have one or two images as arguments."
		return false, errors.New(errMessage)
	} else if len(args) > 2 {
		errMessage = "Too many arguments. Should have at most two images as arguments."
		return false, errors.New(errMessage)
	} else {
		return true, nil
	}
}

func checkImage(arg string) bool {
	if !utils.CheckImageID(arg) && !utils.CheckImageURL(arg) && !utils.CheckTar(arg) {
		return false
	}
	return true
}

func checkArgType(args []string) (bool, error) {
	var buffer bytes.Buffer
	valid := true
	for _, arg := range args {
		if !checkImage(arg) {
			valid = false
			errMessage := fmt.Sprintf("Argument %s is not an image ID, URL, or tar\n", args[0])
			buffer.WriteString(errMessage)
		}
	}
	if !valid {
		return false, errors.New(buffer.String())
	}
	return true, nil
}

func remove(path string, dir bool) string {
	var errStr string
	if path == "" {
		return ""
	}

	var err error
	if dir {
		err = os.RemoveAll(path)
	} else {
		err = os.Remove(path)
	}
	if err != nil {
		errStr = "\nUnable to remove " + path
	}
	return errStr
}
