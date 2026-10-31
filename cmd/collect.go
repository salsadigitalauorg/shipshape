package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/lagoon"
	"github.com/salsadigitalauorg/shipshape/pkg/output"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
)

var collectFormat string

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collect facts only, no checks",
	Long: `Collect all facts or only the one specified and
output them in the format specified`,
	PreRun: func(cmd *cobra.Command, args []string) {
		runCmd.PreRun(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		shipshape.FactsOnly = true
		runCmd.Run(cmd, args)

		if collectFormat == "lagoon-facts" {
			printLagoonFacts()
			return
		}

		for _, f := range fact.Manager().GetPlugins() {
			if shouldSkipFact(f) {
				continue
			}

			log.WithFields(log.Fields{
				"fact":        f.GetId(),
				"fact-plugin": f.GetName(),
				"format":      f.GetFormat(),
			}).Debug("printing collected fact")
			fmt.Printf("%s:", f.GetId())
			switch f.GetFormat() {

			case data.FormatMapBytes:
				loadedData := data.AsMapBytes(f.GetData())
				fmt.Println()
				for k, v := range loadedData {
					fmt.Printf("  %s: %d bytes\n", k, len(v))
				}

			case data.FormatMapListString:
				loadedData := data.AsMapListString(f.GetData())
				for k, vList := range loadedData {
					fmt.Printf("\n  %s:\n", k)
					for _, v := range vList {
						fmt.Printf("    - %s\n", v)
					}
				}

			case data.FormatMapString:
				loadedData := data.AsMapString(f.GetData())
				fmt.Println()
				for k, v := range loadedData {
					if strings.Contains(v, "\n") {
						fmt.Printf("  %s:\n", k)
						fmt.Println(output.TabbedMultiline("    ", v))
					} else {
						fmt.Printf("  %s: %s\n", k, v)
					}
				}

			case data.FormatMapNestedString:
				loadedData := data.AsMapNestedString(f.GetData())
				fmt.Println()
				for k, vMap := range loadedData {
					fmt.Printf("  %s:\n", k)
					for k2, v := range vMap {
						fmt.Printf("    %s: %s\n", k2, v)
					}
				}

			case data.FormatString:
				fmt.Printf(" %s\n", f.GetData())

			case data.FormatRaw:
				fmt.Println("\n",
					output.TabbedMultiline("  ", fmt.Sprintf("%s", f.GetData())))

			case data.FormatNil:
				fmt.Println()

			default:
				log.WithField("data", fmt.Sprintf("%s", f.GetData())).Warn("collect not yet implemented for this format")
				fmt.Println("  collect not yet implemented for", f.GetFormat())
			}
			fmt.Println()
		}
	},
}

// lagoonFactsPayload mirrors the addFactsByName mutation's input shape.
type lagoonFactsPayload struct {
	Project     string        `json:"project"`
	Environment string        `json:"environment"`
	Facts       []lagoon.Fact `json:"facts"`
}

// printLagoonFacts emits collected facts as JSON matching the Lagoon
// addFactsByName mutation input. It only emits the payload - no API call is
// made, so no Lagoon API credentials are needed. project/environment are
// read from the Lagoon outputter's config, populated by
// LAGOON_PROJECT/LAGOON_ENVIRONMENT (collectCmd does not register the
// --lagoon-* flags; see MustHaveEnvVars for the fail-fast check below).
func printLagoonFacts() {
	l := lagoon.Instance()
	l.MustHaveEnvVars()

	facters := map[string]fact.Facter{}
	for id, f := range fact.Manager().GetPlugins() {
		if shouldSkipFact(f) {
			continue
		}
		facters[id] = f
	}

	source := l.Source
	if source == "" {
		source = lagoon.DefaultSource
	}

	payload := lagoonFactsPayload{
		Project:     l.Project,
		Environment: l.Environment,
		Facts:       lagoon.FactsFromFacters(facters, source),
	}

	b, err := json.Marshal(payload)
	if err != nil {
		log.WithError(err).Fatal("unable to marshal lagoon facts payload")
	}
	fmt.Println(string(b))
}

func shouldSkipFact(f fact.Facter) bool {
	if len(fact.OnlyFactNames) == 0 {
		return false
	}
	for _, n := range fact.OnlyFactNames {
		if f.GetId() == n {
			return false
		}
	}
	return true
}

func init() {
	collectCmd.Flags().StringSliceVarP(&fact.OnlyFactNames, "facts", "n",
		[]string{}, "Collect only these facts")
	collectCmd.Flags().StringVarP(&collectFormat, "output-format", "o", "pretty",
		"Output format [pretty|lagoon-facts]")
	rootCmd.AddCommand(collectCmd)
}
