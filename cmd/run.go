package cmd

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/lagoon"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run checks",
	Long:  `Run checks against the specified directory`,
	Args:  cobra.ExactArgs(1),

	PreRun: func(cmd *cobra.Command, args []string) {
		// Parse env vars, overriding flags.
		errorCodeOnFailureEnv := os.Getenv("SHIPSHAPE_ERROR_ON_FAILURE")
		if errorCodeOnFailureEnv != "" {
			if errorCodeOnFailureEnvBool, err := strconv.ParseBool(errorCodeOnFailureEnv); err == nil {
				shipshape.ErrorCodeOnFailure = errorCodeOnFailureEnvBool
			}
		}

		outputFormatEnv := os.Getenv("SHIPSHAPE_OUTPUT_FORMAT")
		if outputFormatEnv != "" {
			shipshape.OutputFormat = outputFormatEnv
		}

		lagoonApiBaseUrlEnv := os.Getenv("LAGOON_API_BASE_URL")
		if outputFormatEnv != "" {
			lagoon.ApiBaseUrl = lagoonApiBaseUrlEnv
		}

		lagoonApiTokenEnv := os.Getenv("LAGOON_API_TOKEN")
		if outputFormatEnv != "" {
			lagoon.ApiToken = lagoonApiTokenEnv
		}

		if !shipshape.ValidateOutputFormat() {
			log.Fatalf("Invalid output format; needs to be one of: %s.",
				strings.Join(shipshape.OutputFormats, "|"))
		}

		// simple check to ensure we have everything we need to write to the API if required.
		if lagoon.PushProblemsToInsightRemote {
			if lagoon.ApiBaseUrl == "" {
				log.Fatal("lagoon api base url not provided")
			}
			if lagoon.ApiToken == "" {
				log.Fatal("lagoon api token not provided")
			}
		}
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
		} else {
			shipshape.Run()
		}

		shipshape.Output()

		if lagoon.PushProblemsToInsightRemote {
			w := bufio.NewWriter(os.Stdout)
			err := lagoon.ProcessResultList(w, shipshape.RunResultList)
			if err != nil {
				log.Fatal(err)
			}
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

	// Output.
	runCmd.Flags().StringVarP(&shipshape.OutputFormat, "output",
		"o", "simple", `Output format [json|junit|simple|table]
(env: SHIPSHAPE_OUTPUT_FORMAT)`)

	// Lagoon.
	runCmd.Flags().StringVar(&lagoon.ApiBaseUrl, "lagoon-api-base-url",
		"", `Base url for the Lagoon API when pushing
problems to API (env: LAGOON_API_BASE_URL)`)
	runCmd.Flags().StringVar(&lagoon.ApiToken, "lagoon-api-token", "",
		`Lagoon API token when pushing problems
to API (env: LAGOON_API_TOKEN)`)
	runCmd.Flags().BoolVar(&lagoon.PushProblemsToInsightRemote,
		"lagoon-push-problems-to-insights", false,
		"Push audit facts to Lagoon via Insights Remote")
	runCmd.Flags().StringVar(&lagoon.LagoonInsightsRemoteEndpoint,
		"lagoon-insights-remote-endpoint",
		"http://lagoon-remote-insights-remote.lagoon.svc/problems",
		"Insights Remote Problems endpoint\n")

	rootCmd.AddCommand(runCmd)
}
