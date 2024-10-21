package lagoon

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/output"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

// Lagoon is the output plugin for pushing problems to Lagoon.
// TODO: Split Lagoon config into separate config and output plugins if
// more output plugins are added.
type Lagoon struct {
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
	output.Outputters["lagoon"] = l
}

func (p *Lagoon) Output(rl *result.ResultList) ([]byte, error) {
	if !p.PushProblemsToInsightsRemote {
		return nil, nil
	}

	buf := bytes.Buffer{}
	bufW := bufio.NewWriter(&buf)
	problems := []Problem{}

	if rl.TotalBreaches == 0 {
		InitClient(p.ApiBaseUrl, p.ApiToken)
		err := p.DeleteProblems()
		if err != nil {
			return nil, err
		}
		fmt.Fprint(bufW, "no breach to push to Lagoon; only deleted previous problems")
		bufW.Flush()
		return buf.Bytes(), nil
	}

	for _, r := range rl.Results {
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
			Source:            p.Source,
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
