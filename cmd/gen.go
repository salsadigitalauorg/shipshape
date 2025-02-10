//go:build ignore

package main

import (
	"log"

	"github.com/spf13/pflag"

	"github.com/salsadigitalauorg/shipshape/cmd/gen"
)

var (
	arg               string
	arg2              string
	pkg               string
	enableEnvResolver bool
	breachTypes       []string
	plugins           []string
	names             []string
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
		gen.CheckRegistry(pkg)
		break
	case "breach-type":
		if len(breachTypes) == 0 {
			log.Fatal("breach-type missing flags; type is required")
		}
		gen.BreachType(breachTypes)
		break
	case "connection-plugin":
		if len(plugins) == 0 {
			log.Fatal("connection-plugin missing flags; plugin is required")
		}
		gen.ConnectionPlugin(plugins)
		break
	case "fact-plugin":
		if len(plugins) == 0 {
			log.Fatal("fact-plugin missing flags; plugin is required")
		}
		if pkg == "" {
			log.Fatal("fact-plugin missing flags; package is required")
		}
		gen.FactPlugin(plugins, pkg, enableEnvResolver)
		gen.FactRegistry(pkg)
		break
	case "analyse-plugin":
		if len(plugins) == 0 {
			log.Fatal("analyse-plugin missing flags; plugin is required")
		}
		gen.AnalysePlugin(plugins)
		break
	case "remediator":
		if len(plugins) == 0 {
			log.Fatal("remediator missing flags; plugins is required")
		}
		if len(names) == 0 {
			log.Fatal("remediator missing flags; name is required")
		}
		gen.RemediatorPlugin(plugins, names)
		break
	default:
		log.Fatalf("unknown argument '%s'", arg)
	}
}

func parseFlags() {
	pflag.StringVar(&pkg, "checkpackage", "", "The package to which the check belongs")
	pflag.StringVar(&pkg, "package", "", "The package to which the plugin belongs")
	pflag.BoolVar(&enableEnvResolver, "envresolver", false, "Add envResolver methods to the plugin")
	pflag.StringSliceVar(&breachTypes, "type", []string{}, "The breach type")
	pflag.StringSliceVar(&plugins, "plugin", []string{}, "The plugin struct")
	pflag.StringSliceVar(&names, "name", []string{}, "The plugin name")
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
