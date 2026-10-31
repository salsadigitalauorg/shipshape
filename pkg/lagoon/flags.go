package lagoon

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/salsadigitalauorg/shipshape/pkg/flagsprovider"
)

func init() {
	flagsprovider.Registry["lagoon"] = func() flagsprovider.FlagsProvider { return l }
}

func (p *Lagoon) AddFlags(c *cobra.Command) {
	c.Flags().BoolVar(&p.PushProblemsToInsightsRemote,
		"lagoon-push-problems-to-insights", false,
		"Push audit facts to Lagoon via Insights Remote")

	c.Flags().StringVar(&p.InsightsRemoteEndpoint,
		"lagoon-insights-remote-endpoint",
		"http://lagoon-remote-insights-remote.lagoon.svc/problems",
		"Insights Remote Problems endpoint\n")

	c.Flags().StringVar(&p.InsightsRemoteBaseUrl,
		"lagoon-insights-remote-base-url",
		"http://lagoon-remote-insights-remote.lagoon.svc",
		"Insights Remote base url, used for deleting problems\n")

	c.Flags().StringVar(&p.Source, "lagoon-source", "Shipshape",
		"Source to use for Problems pushed to Lagoon")

	c.Flags().StringVar(&p.Project, "lagoon-project", "",
		"The Lagoon project name (env: LAGOON_PROJECT)")

	c.Flags().StringVar(&p.Environment, "lagoon-environment", "",
		"The Lagoon environment name (env: LAGOON_ENVIRONMENT)")
}

func (p *Lagoon) EnvironmentOverrides() {
	project := os.Getenv("LAGOON_PROJECT")
	if project != "" {
		p.Project = project
	}

	environment := os.Getenv("LAGOON_ENVIRONMENT")
	if environment != "" {
		p.Environment = environment
	}
}

func (p *Lagoon) MustHaveEnvVars() {
	if p.Project == "" || p.Environment == "" {
		log.Fatal("project & environment name required; please ensure both " +
			"LAGOON_PROJECT & LAGOON_ENVIRONMENT are set")
	}
}
