package cmd

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/output"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
)

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
		for _, f := range fact.Manager().GetPlugins() {
			if shouldSkipFact(f) {
				continue
			}

			log.WithFields(log.Fields{
				"fact":   f.GetName(),
				"format": f.GetFormat(),
			}).Debug("printing collected fact")
			fmt.Printf("%s:", f.GetName())
			switch f.GetFormat() {

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

			default:
				log.WithField("data", fmt.Sprintf("%s", f.GetData())).Warn("collect not yet implemented for this format")
				fmt.Println("  collect not yet implemented for", f.GetFormat())
			}
			fmt.Println()
		}
	},
}

func shouldSkipFact(f fact.Facter) bool {
	if len(fact.OnlyFactNames) == 0 {
		return false
	}
	for _, n := range fact.OnlyFactNames {
		if f.GetName() == n {
			return false
		}
	}
	return true
}

func init() {
	collectCmd.Flags().StringSliceVarP(&fact.OnlyFactNames, "facts", "n",
		[]string{}, "Collect only these facts")
	rootCmd.AddCommand(collectCmd)
}
