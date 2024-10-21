package lagoon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/config"

	"github.com/hasura/go-graphql-client"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
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

var Client *graphql.Client

func InitClient(apiBaseUrl, apiToken string) {
	if Client != nil {
		return
	}
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: apiToken},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	Client = graphql.NewClient(apiToken+"/graphql", httpClient)
}

// GetEnvironmentId derives the environment id from a Lagoon project
// name & environment name.
func GetEnvironmentId(project string, environment string) (int, error) {
	ns := project + "-" + environment
	log.WithField("namespace", ns).Info("fetching environment id")
	var q struct {
		EnvironmentByKubernetesNamespaceName struct {
			Id int
		} `graphql:"environmentByKubernetesNamespaceName(kubernetesNamespaceName: $ns)"`
	}
	variables := map[string]interface{}{"ns": ns}
	err := Client.Query(context.Background(), &q, variables)
	if err != nil {
		return 0, err
	}
	return q.EnvironmentByKubernetesNamespaceName.Id, nil
}

const DefaultLagoonInsightsTokenLocation = "/var/run/secrets/lagoon/dynamic/insights-token/INSIGHTS_TOKEN"

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

	req, _ := http.NewRequest(http.MethodPost, serviceEndpoint, bytes.NewBuffer(bodyString))
	req.Header.Set("Authorization", bearerToken)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("there was an error sending the problems to '%s' : %s", serviceEndpoint, response.Body)
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
