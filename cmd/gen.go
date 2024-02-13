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
	breachTypes  []string
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
	case "breach-type":
		if len(breachTypes) == 0 {
			log.Fatal("missing flags; struct is required")
		}
		gen.BreachType(breachTypes)
		break
	}
}

func parseFlags() {
	pflag.StringVar(&checkpackage, "checkpackage", "", "The package to which the check belongs")
	pflag.StringSliceVar(&breachTypes, "type", []string{}, "The breach type")
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
