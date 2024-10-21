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
	c.Flags().StringVar(&p.ApiBaseUrl, "lagoon-api-base-url",
		"", `Base url for the Lagoon API when pushing
problems to API (env: LAGOON_API_BASE_URL)`)

	c.Flags().StringVar(&p.ApiToken, "lagoon-api-token", "",
		`Lagoon API token when pushing problems
to API (env: LAGOON_API_TOKEN)`)

	c.Flags().BoolVar(&p.PushProblemsToInsightsRemote,
		"lagoon-push-problems-to-insights", false,
		"Push audit facts to Lagoon via Insights Remote")

	c.Flags().StringVar(&p.InsightsRemoteEndpoint,
		"lagoon-insights-remote-endpoint",
		"http://lagoon-remote-insights-remote.lagoon.svc/problems",
		"Insights Remote Problems endpoint\n")

	c.Flags().StringVar(&p.Source, "lagoon-source", "Shipshape",
		"Source to use for Problems pushed to Lagoon")

	c.Flags().StringVar(&p.Project, "lagoon-project", "",
		"The Lagoon project name (env: LAGOON_PROJECT)")

	c.Flags().StringVar(&p.Environment, "lagoon-environment", "",
		"The Lagoon environment name (env: LAGOON_ENVIRONMENT)")
}

func (p *Lagoon) EnvironmentOverrides() {
	apiBaseUrlEnv := os.Getenv("LAGOON_API_BASE_URL")
	if apiBaseUrlEnv != "" {
		p.ApiBaseUrl = apiBaseUrlEnv
	}

	apiTokenEnv := os.Getenv("LAGOON_API_TOKEN")
	if apiTokenEnv != "" {
		p.ApiToken = apiTokenEnv
	}

	// simple check to ensure we have everything we need to write to the API if required.
	if p.PushProblemsToInsightsRemote {
		if p.ApiBaseUrl == "" {
			log.Fatal("lagoon api base url not provided")
		}
		if p.ApiToken == "" {
			log.Fatal("lagoon api token not provided")
		}
	}

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
