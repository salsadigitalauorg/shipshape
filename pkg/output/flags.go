package output

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salsadigitalauorg/shipshape/pkg/flagsprovider"
)

func init() {
	flagsprovider.Registry["stdout"] = func() flagsprovider.FlagsProvider {
		return s
	}
}

func (f *Stdout) ValidateOutputFormat() bool {
	valid := false
	for _, fm := range OutputFormats {
		if f.Format == fm {
			valid = true
			break
		}
	}
	return valid
}

func (f *Stdout) AddFlags(c *cobra.Command) {
	c.Flags().StringVarP(&f.Format, "output-format",
		"o", "pretty", `Output format [pretty|table|json|junit]
(env: SHIPSHAPE_OUTPUT_FORMAT)`)
}

func (f *Stdout) EnvironmentOverrides() {
	if outputFormatEnv := os.Getenv("SHIPSHAPE_OUTPUT_FORMAT"); outputFormatEnv != "" {
		f.Format = outputFormatEnv
	}

	if !f.ValidateOutputFormat() {
		log.Fatalf("Invalid output format; needs to be one of: %s.",
			strings.Join(OutputFormats, "|"))
	}
}
