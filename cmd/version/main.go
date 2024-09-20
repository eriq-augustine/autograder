package main

import (
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	Out string `help:"Writes the output to the given file in JSON format."`
}

func main() {
	kong.Parse(&args,
		kong.Description("Get the autograder's version."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	if args.Out == "" {
		fmt.Printf("Short Version: %s\n", util.GetAutograderVersion())
		fmt.Printf("Full  Version: %s\n", util.Version.FullVersion(util.GetAutograderFullVersion()))
		fmt.Printf("API   Version: %d\n", util.MustGetAPIVersion())
	} else {
		version := util.GetAutograderFullVersion()

		err = util.ToJSONFileIndent(&version, args.Out)
		if err != nil {
			log.Error("Failed to write to the JSON file", err, log.NewAttr("path", args.Out))
		}
	}
}
