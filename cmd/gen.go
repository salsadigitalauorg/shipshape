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
	plugins     []string
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
			log.Fatal("breach-type missing flags; type is required")
		}
		gen.BreachType(breachTypes)
		break
	case "fact-plugin":
		if len(plugins) == 0 {
			log.Fatal("fact-plugin missing flags; plugin is required")
		}
		if pkg == "" {
			log.Fatal("fact-plugin missing flags; package is required")
		}
		gen.FactPlugin(plugins, pkg)
		break
	case "connection-plugin":
		if len(plugins) == 0 {
			log.Fatal("connection-plugin missing flags; plugin is required")
		}
		gen.ConnectionPlugin(plugins)
		break
	}
}

func parseFlags() {
	pflag.StringVar(&pkg, "checkpackage", "", "The package to which the check belongs")
	pflag.StringVar(&pkg, "package", "", "The package to which the plugin belongs")
	pflag.StringSliceVar(&breachTypes, "type", []string{}, "The breach type")
	pflag.StringSliceVar(&plugins, "plugin", []string{}, "The plugin struct")
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
