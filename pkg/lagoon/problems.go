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

	// Source can be specified when pushing Problems to Lagoon. Falls back
	// to the per-Result source (see result.Result.Source) when set there;
	// default is "Shipshape".
	Source string `yaml:"source"`

	// ClearProblems controls whether Shipshape deletes existing Lagoon
	// problems for its owned sources before pushing new ones. Defaults to
	// true, and runs on every invocation, not only zero-breach runs.
	ClearProblems bool `yaml:"clear-problems"`

	Project     string `yaml:"project"`
	Environment string `yaml:"environment"`
}

var l = &Lagoon{Source: DefaultSource, ClearProblems: true}

func init() {
	output.Outputters["lagoon"] = l
}

// getBearerToken is overridable in tests so Output() can be exercised without
// the real in-cluster token file at DefaultLagoonInsightsTokenLocation.
var getBearerToken = func() (string, error) {
	return GetBearerTokenFromDisk(DefaultLagoonInsightsTokenLocation)
}

// resolveSource returns the effective Lagoon Problem source for a Result: the
// per-analyser override when set, falling back to the plugin-level Source,
// falling back to DefaultSource if both are empty (so a YAML `source:` with
// no value never produces an empty-string deletion target).
func (p *Lagoon) resolveSource(r result.Result) string {
	if r.Source != "" {
		return r.Source
	}
	if p.Source != "" {
		return p.Source
	}
	return DefaultSource
}

// ownedSources returns the deduplicated set of Lagoon Problem sources this
// run owns: the plugin-level default plus every per-Result override, derived
// from all Results (not only breached ones), so a fully-clean run still
// clears its own stale problems.
func (p *Lagoon) ownedSources(rl *result.ResultList) []string {
	seen := map[string]bool{}
	sources := []string{}
	add := func(s string) {
		if s == "" || seen[s] {
			return
		}
		seen[s] = true
		sources = append(sources, s)
	}

	add(p.resolveSource(result.Result{}))
	for _, r := range rl.Results {
		add(p.resolveSource(r))
	}
	return sources
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
		"clear-problems":           p.ClearProblems,
	}).Debug("pushing problems to Lagoon")
	buf := bytes.Buffer{}
	bufW := bufio.NewWriter(&buf)
	problems := []Problem{}

	bearerToken, tokenErr := getBearerToken()

	deleted := false
	if p.ClearProblems {
		if tokenErr != nil {
			return nil, tokenErr
		}
		sources := p.ownedSources(rl)
		log.WithField("sources", sources).Debug("clearing existing problems")
		if err := DeleteProblems(sources, p.InsightsRemoteBaseUrl, bearerToken); err != nil {
			return nil, err
		}
		deleted = true
	}

	if rl.TotalBreaches == 0 {
		switch {
		case deleted:
			fmt.Fprint(bufW, "no breach to push to Lagoon; only deleted previous problems")
		default:
			fmt.Fprint(bufW, "no breach to push to Lagoon")
		}
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
			Source:            p.resolveSource(r),
			Service:           "",
			Data:              string(breachMapJson),
			Severity:          SeverityTranslation(config.Severity(r.Severity)),
			SeverityScore:     0,
			AssociatedPackage: "",
			Description:       "",
			Links:             "",
		})
	}

	if tokenErr != nil {
		return nil, tokenErr
	}
	if err := ProblemsToInsightsRemote(problems, p.InsightsRemoteEndpoint, bearerToken); err != nil {
		return nil, err
	}
	fmt.Fprintln(bufW, "successfully pushed problems to Lagoon Remote")
	bufW.Flush()
	return buf.Bytes(), nil
}
