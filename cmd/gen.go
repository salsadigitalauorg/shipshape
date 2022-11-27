//go:build ignore

package main

import (
	"log"

	"github.com/salsadigitalauorg/shipshape/cmd/gen"
	"github.com/spf13/pflag"
)

var (
	arg          string
	checkpackage string
	checktype    string
	checkstruct  string
)

func main() {
	parseFlags()
	parseArgs()

	switch arg {
	case "registry":
		if checkpackage == "" {
			log.Fatal("missing flags; checkpackage is required")
		}
		gen.Registry(checkpackage)
		break
	}
}

func parseFlags() {
	pflag.StringVar(&checkpackage, "checkpackage", "", "The package to which the check belongs")
	pflag.StringVar(&checktype, "checktype", "", "The checktype constant name")
	pflag.StringVar(&checkstruct, "checkstruct", "", "The struct defined by the check")
	pflag.Parse()
}

func parseArgs() {
	args := pflag.Args()
	if len(args) == 0 || len(args) > 1 {
		log.Fatalf("1 argument expected, got '%+v'\n", args)
	} else {
		arg = args[0]
	}
}
