package cmd

import (
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/flagsprovider"
	"github.com/salsadigitalauorg/shipshape/pkg/output"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
)

var runCmd = &cobra.Command{
	Use:   "run [dir|.]",
	Short: "Execute policies",
	Long:  `Execute policies against the specified directory`,
	Args:  cobra.ExactArgs(1),

	PreRun: func(cmd *cobra.Command, args []string) {
		config.ProjectDir = args[0]

		// Parse env vars, overriding flags.
		errorCodeOnFailureEnv := os.Getenv("SHIPSHAPE_ERROR_ON_FAILURE")
		if errorCodeOnFailureEnv != "" {
			if errorCodeOnFailureEnvBool, err := strconv.ParseBool(errorCodeOnFailureEnv); err == nil {
				shipshape.ErrorCodeOnFailure = errorCodeOnFailureEnvBool
			}
		}

		flagsprovider.ApplyEnvironmentOverridesAll()
	},

	Run: func(cmd *cobra.Command, args []string) {
		if !config.ConfigFilesExist() {
			shipshape.Exit(1)
		}

		err := shipshape.Init()
		if err != nil {
			log.Fatal(err)
		}

		if shipshape.IsV2 {
			shipshape.RunV2()

			// If we're only collecting facts, we don't need to output anything,
			// but not exit either, since it is then assumed this command was called
			// from collectCmd.
			if shipshape.FactsOnly {
				return
			}
		} else {
			shipshape.Run()
		}

		log.Print("outputting results")
		if err := output.OutputAll(os.Stdout); err != nil {
			log.Fatal(err)
		}

		if shipshape.RunResultList.Status() == result.Fail &&
			len(shipshape.RunResultList.GetBreachesBySeverity(shipshape.FailSeverity)) > 0 {

			shipshape.Exit(2)
		}
	},
}

func init() {
	// Remediation.
	runCmd.Flags().BoolVarP(&shipshape.Remediate, "remediate", "r",
		false, "Run remediation for supported checks")

	// Failure state.
	runCmd.Flags().StringVar(&shipshape.FailSeverity,
		"fail-severity", string(config.HighSeverity),
		`The severity level at which the program
should exit with an error`)
	runCmd.Flags().BoolVarP(&shipshape.ErrorCodeOnFailure,
		"error-code", "e", false, `Exit with error code if a failure is
detected (env: SHIPSHAPE_ERROR_ON_FAILURE)`)

	flagsprovider.AddFlagsAll(runCmd)

	rootCmd.AddCommand(runCmd)
}
