package lagoon

import (
	"bufio"
	"bytes"
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

	PushProblemsToInsightsRemote bool   `yaml:"push-problems-to-insights"`
	InsightsRemoteEndpoint       string `yaml:"insights-remote-endpoint"`

	// InsightsRemoteBaseUrl is the base insights-remote URL used for
	// deletion requests, e.g. "http://lagoon-remote-insights-remote.lagoon.svc".
	// Distinct from InsightsRemoteEndpoint, which is the "/problems"-suffixed
	// POST endpoint.
	InsightsRemoteBaseUrl string `yaml:"insights-remote-base-url"`

	// Source can be specified when pushing Problems to Lagoon.
	// Default is "Shipshape".
	Source string `yaml:"source"`

	Project     string `yaml:"project"`
	Environment string `yaml:"environment"`
}

var l = &Lagoon{Source: "Shipshape"}

func init() {
	output.Outputters["lagoon"] = l
}

func (p *Lagoon) Output(rl *result.ResultList) ([]byte, error) {
	if !p.PushProblemsToInsightsRemote {
		log.Debug("skipping pushing problems to Lagoon")
		return nil, nil
	}

	log.WithFields(log.Fields{
		"insights-remote-endpoint": p.InsightsRemoteEndpoint,
		"project":                  p.Project,
		"environment":              p.Environment,
		"source":                   p.Source,
	}).Debug("pushing problems to Lagoon")
	buf := bytes.Buffer{}
	bufW := bufio.NewWriter(&buf)
	problems := []Problem{}

	if rl.TotalBreaches == 0 {
		log.WithFields(log.Fields{
			"project":     p.Project,
			"environment": p.Environment,
			"source":      p.Source,
		}).Debug("cleaning up existing problems")

		bearerToken, err := GetBearerTokenFromDisk(DefaultLagoonInsightsTokenLocation)
		if err != nil {
			return nil, err
		}
		if err := DeleteProblems([]string{p.Source}, p.InsightsRemoteBaseUrl, bearerToken); err != nil {
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

	bearerToken, err := GetBearerTokenFromDisk(DefaultLagoonInsightsTokenLocation)
	if err != nil {
		return nil, err
	}
	if err := ProblemsToInsightsRemote(problems, p.InsightsRemoteEndpoint, bearerToken); err != nil {
		return nil, err
	}
	fmt.Fprintln(bufW, "successfully pushed problems to Lagoon Remote")
	bufW.Flush()
	return buf.Bytes(), nil
}
