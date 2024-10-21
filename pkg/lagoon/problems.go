package lagoon

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/flagsprovider"
	"github.com/salsadigitalauorg/shipshape/pkg/output"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

// Lagoon is the output plugin for pushing problems to Lagoon.
// TODO: Split Lagoon config into separate config and output plugins if
// more output plugins are added.
type Lagoon struct {
	// Common fields.
	ResultList *result.ResultList `yaml:"-"`

	// Plugin-specific fields.

	// ApiBaseUrl is the base URL for the Lagoon API.
	ApiBaseUrl string `yaml:"api-base-url"`

	// ApiToken is the token for the Lagoon API.
	ApiToken string `yaml:"api-token"`

	PushProblemsToInsightsRemote bool   `yaml:"push-problems-to-insights"`
	InsightsRemoteEndpoint       string `yaml:"insights-remote-endpoint"`

	// Source can be specified when pushing Problems to Lagoon.
	// Default is "shipshape".
	Source string `yaml:"source"`

	Project     string `yaml:"project"`
	Environment string `yaml:"environment"`
}

var l = &Lagoon{Source: "shipshape"}

func init() {
	output.Registry["lagoon"] = func(rl *result.ResultList) output.Outputter {
		l.ResultList = rl
		return l
	}

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

func (p *Lagoon) Output() ([]byte, error) {
	if !p.PushProblemsToInsightsRemote {
		return nil, nil
	}

	buf := bytes.Buffer{}
	bufW := bufio.NewWriter(&buf)
	problems := []Problem{}

	if p.ResultList.TotalBreaches == 0 {
		InitClient(p.ApiBaseUrl, p.ApiToken)
		err := p.DeleteProblems()
		if err != nil {
			return nil, err
		}
		fmt.Fprint(bufW, "no breach to push to Lagoon; only deleted previous problems")
		bufW.Flush()
		return buf.Bytes(), nil
	}

	for _, r := range p.ResultList.Results {
		if len(r.Breaches) == 0 {
			continue
		}

		// let's marshal the breaches, they can be attached to the problem in the data field
		breachMapJson, err := json.Marshal(r.Breaches)
		if err != nil {
			log.WithError(err).Fatal("Unable to marshal breach information")
		}

		problems = append(problems, Problem{
			Identifier:        r.Name,
			Version:           "1",
			FixedVersion:      "",
			Source:            "shipshape",
			Service:           "",
			Data:              string(breachMapJson),
			Severity:          SeverityTranslation(config.Severity(r.Severity)),
			SeverityScore:     0,
			AssociatedPackage: "",
			Description:       "",
			Links:             "",
		})
	}

	InitClient(p.ApiBaseUrl, p.ApiToken)
	// first, let's try doing this via in-cluster functionality
	bearerToken, err := GetBearerTokenFromDisk(DefaultLagoonInsightsTokenLocation)
	if err == nil { // we have a token, and so we can proceed via the internal service call
		err = ProblemsToInsightsRemote(problems, p.InsightsRemoteEndpoint, bearerToken)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	fmt.Fprintln(bufW, "successfully pushed problems to Lagoon Remote")
	bufW.Flush()
	return buf.Bytes(), nil
}

func (p *Lagoon) DeleteProblems() error {
	envId, err := GetEnvironmentId(p.Project, p.Environment)
	if err != nil {
		return err
	}
	var m struct {
		DeleteFactsFromSource string `graphql:"deleteProblemsFromSource(input: {environment: $envId, source: $sourceName, service:$service})"`
	}
	variables := map[string]interface{}{
		"envId":      envId,
		"sourceName": p.Source,
		"service":    "",
	}
	return Client.Mutate(context.Background(), &m, variables)
}
