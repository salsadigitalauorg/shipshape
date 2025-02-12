package cmd

import (
	"fmt"
	"maps"
	"slices"
	"sort"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/analyse"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/output"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage shipshape configuration",
}

var configListChecksCmd = &cobra.Command{
	Use:   "list-checks",
	Short: "List available checks",
	Long:  `List all available checks that can be run by shipshape`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Type of checks available:")
		checks := []string{}
		for c := range config.ChecksRegistry {
			checks = append(checks, string(c))
		}
		sort.Strings(checks)
		for _, c := range checks {
			fmt.Println("  - " + c)
		}
	},
}

var configListPluginsCmd = &cobra.Command{
	Use:   "list-plugins",
	Short: "List available plugins",
	Long:  `List all available plugins that can be used in shipshape`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Connection plugins:")
		plugins := slices.Collect(maps.Keys(connection.Registry))
		sort.Strings(plugins)
		for _, p := range plugins {
			fmt.Println("  - " + p)
		}

		fmt.Println("\nFact plugins:")
		plugins = slices.Collect(maps.Keys(fact.Registry))
		sort.Strings(plugins)
		for _, p := range plugins {
			fmt.Println("  - " + p)
		}

		fmt.Println("\nAnalyse plugins:")
		plugins = slices.Collect(maps.Keys(analyse.Registry))
		sort.Strings(plugins)
		for _, p := range plugins {
			fmt.Println("  - " + p)
		}

		fmt.Println("\nOutput plugins:")
		plugins = slices.Collect(maps.Keys(output.Outputters))
		sort.Strings(plugins)
		for _, p := range plugins {
			fmt.Println("  - " + p)
		}
	},
}

var configDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dumps the current configuration",
	Long: `Dumps the final merged configuration - useful to make sure
multiple config files are being merged as expected`,
	Run: func(cmd *cobra.Command, args []string) {
		if !config.ConfigFilesExist() {
			shipshape.Exit(1)
		}

		err := shipshape.Init()
		if err != nil {
			log.Fatal(err)
		}

		var out []byte
		if shipshape.IsV2 {
			out, err = yaml.Marshal(shipshape.RunConfigV2)
		} else {
			out, err = yaml.Marshal(shipshape.RunConfig)
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", string(out))
	},
}

func init() {
	configCmd.AddCommand(configListPluginsCmd)
	configCmd.AddCommand(configListChecksCmd)
	configCmd.AddCommand(configDumpCmd)
	rootCmd.AddCommand(configCmd)
}
