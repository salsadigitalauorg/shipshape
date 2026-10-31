package lagoon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/config"

	log "github.com/sirupsen/logrus"
)

type Fact struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Source      string `json:"source"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type ProblemSeverityRating string

type Problem struct {
	EnvironmentId     int                   `json:"environment"`
	Identifier        string                `json:"identifier"`
	Version           string                `json:"version,omitempty"`
	FixedVersion      string                `json:"fixedVersion,omitempty"`
	Source            string                `json:"source,omitempty"`
	Service           string                `json:"service,omitempty"`
	Data              string                `json:"data"`
	Severity          ProblemSeverityRating `json:"severity,omitempty"`
	SeverityScore     float64               `json:"severityScore,omitempty"`
	AssociatedPackage string                `json:"associatedPackage,omitempty"`
	Description       string                `json:"description,omitempty"`
	Links             string                `json:"links,omitempty"`
}

const DefaultLagoonInsightsTokenLocation = "/var/run/secrets/lagoon/dynamic/insights-token/INSIGHTS_TOKEN"

// DefaultSource is the Lagoon Problem/Fact source used when neither a
// per-analyser nor a plugin-level source is configured.
const DefaultSource = "Shipshape"

func GetBearerTokenFromDisk(tokenLocation string) (string, error) {
	//first, we check that the token exists on disk
	_, err := os.Stat(tokenLocation)
	if err != nil {
		return "", fmt.Errorf("unable to load insights token from disk")
	}

	b, err := os.ReadFile(tokenLocation)
	if err != nil {
		return "", err
	}
	return strings.Trim(string(b), "\n"), nil
}

func ProblemsToInsightsRemote(problems []Problem, serviceEndpoint string, bearerToken string) error {
	bodyString, err := json.Marshal(problems)
	if err != nil {
		return err
	}

	log.WithField("body", string(bodyString)).
		Trace("sending problems to Lagoon Insights")

	req, _ := http.NewRequest(http.MethodPost, serviceEndpoint, bytes.NewBuffer(bodyString))
	req.Header.Set("Authorization", bearerToken)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("there was an error sending the problems to '%s' : %s", serviceEndpoint, response.Status)
	}
	return nil
}

// DeleteProblems deletes all problems for the given sources via
// insights-remote's HTTP DELETE endpoint. One request per distinct source;
// duplicate sources are only requested once. insightsRemoteBaseUrl is the
// base insights-remote URL (e.g.
// "http://lagoon-remote-insights-remote.lagoon.svc"), not the
// "/problems"-suffixed POST endpoint.
func DeleteProblems(sources []string, insightsRemoteBaseUrl string, bearerToken string) error {
	seen := map[string]bool{}
	client := &http.Client{}
	for _, source := range sources {
		if source == "" || seen[source] {
			continue
		}
		seen[source] = true

		endpoint := strings.TrimSuffix(insightsRemoteBaseUrl, "/") + "/problems/" + url.PathEscape(source)
		log.WithField("endpoint", endpoint).Trace("deleting problems from Lagoon Insights")

		req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", bearerToken)

		response, err := client.Do(req)
		if err != nil {
			return err
		}
		response.Body.Close()
		if response.StatusCode >= 300 {
			return fmt.Errorf("there was an error deleting problems at '%s': %s", endpoint, response.Status)
		}
	}
	return nil
}

// SeverityTranslation will convert a ShipShape severity rating to a Lagoon rating
func SeverityTranslation(ssSeverity config.Severity) ProblemSeverityRating {
	// Currently supported severity levels in Lagoon
	//NONE
	//UNKNOWN
	//NEGLIGIBLE
	//LOW
	//MEDIUM
	//HIGH
	//CRITICAL

	switch ssSeverity {
	case config.LowSeverity:
		return "LOW"
	case config.NormalSeverity:
		return "MEDIUM"
	case config.HighSeverity:
		return "HIGH"
	case config.CriticalSeverity:
		return "CRITICAL"
	}

	return "UNKNOWN"
}
