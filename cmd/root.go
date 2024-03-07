package cmd

import (
	"fmt"
	"os"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// Version information.
var (
	version string
	commit  string
)

// Logging flags.
var (
	logLevel string
	verbose  bool
	debug    bool
)

var rootCmd = &cobra.Command{
	Use:   "shipshape",
	Short: "shipshape is a tool for identifying breaches from data",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Determine log level.
		if verbose {
			logLevel = "info"
		}
		if debug {
			logLevel = "debug"
		}
		initLogger(logLevel)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays the shipshape version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Initial config.
	rootCmd.PersistentFlags().StringSliceVarP(&config.Files, "file", "f",
		[]string{"shipshape.yml"}, `Path to the file containing the checks.
Can be specified as comma-separated single argument or
using --file (-f) multiple times`)

	// Filter checks.
	rootCmd.PersistentFlags().StringSliceVarP(&config.CheckTypesToRun, "types",
		"t", []string(nil), `List of checks to run; default is empty, which will
run all checks. Can be specified as comma-separated
single argument or using --types (-t) multiple times`)
	rootCmd.PersistentFlags().BoolVarP(&config.ExcludeDb, "exclude-db", "x",
		false, `Exclude checks requiring a database; overrides
any db checks specified by '--types'`)

	// Logging flags.
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "warn",
		"Level of logs to display")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"Display verbose output - equivalent to --log-level info")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false,
		"Display debug information - equivalent to --log-level debug")

	rootCmd.AddCommand(versionCmd)
}

func initLogger(logLevel string) {
	if logLevel == "" {
		logLevel = "warn"
	}
	if logrusLevel, err := log.ParseLevel(logLevel); err != nil {
		panic(err)
	} else {
		log.SetLevel(logrusLevel)
	}
}
