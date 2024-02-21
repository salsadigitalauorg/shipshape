//go:build ignore

package main

import (
	"log"

	"github.com/salsadigitalauorg/shipshape/cmd/gen"
	"github.com/spf13/pflag"
)

var (
	arg         string
	arg2        string
	pkg         string
	breachTypes []string
)

func main() {
	parseFlags()
	parseArgs()

	switch arg {
	case "registry":
		if arg2 == "" && pkg == "" {
			log.Fatal("check registry missing flags; checkpackage is required")
		} else if arg2 == "create" {
			gen.RegistryCreateFile()
			break
		}
		gen.Registry(pkg)
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
	pflag.StringVar(&pkg, "checkpackage", "", "The package to which the check belongs")
	pflag.StringSliceVar(&breachTypes, "type", []string{}, "The breach type")
	pflag.Parse()
}

func parseArgs() {
	args := pflag.Args()
	if len(args) == 0 {
		log.Fatalf("at least 1 argument expected, got '%+v'\n", args)
	} else {
		arg = args[0]
		if len(args) > 1 {
			arg2 = args[1]
		}
	}
}
