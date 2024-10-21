package flagsprovider

import "github.com/spf13/cobra"

type FlagsProvider interface {
	AddFlags(*cobra.Command)
	EnvironmentOverrides()
}

var Registry = map[string]func() FlagsProvider{}
var FlagProviders = map[string]FlagsProvider{}

func AddFlagsAll(c *cobra.Command) {
	for _, p := range FlagProviders {
		p.AddFlags(c)
	}
}

func ApplyEnvironmentOverridesAll() {
	for _, p := range FlagProviders {
		p.EnvironmentOverrides()
	}
}
