package lagoon

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"

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

const SourceName string = "Shipshape"
const FactMaxValueLength int = 300

var PushProblemsToInsightRemote bool
var LagoonInsightsRemoteEndpoint string

var project string
var environment string

func MustHaveEnvVars() {
	project = os.Getenv("LAGOON_PROJECT")
	environment = os.Getenv("LAGOON_ENVIRONMENT")
	if project == "" || environment == "" {
		log.Fatal("project & environment name required; please ensure both " +
			"LAGOON_PROJECT & LAGOON_ENVIRONMENT are set")
	}
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

func ProcessResultList(w *bufio.Writer, list result.ResultList) error {
	problems := []Problem{}

	// first, let's try doing this via in-cluster functionality
	bearerToken, err := GetBearerTokenFromDisk(DefaultLagoonInsightsTokenLocation)

	if list.TotalBreaches == 0 {
		err := DeleteProblems(LagoonInsightsRemoteEndpoint, bearerToken)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, "no breach to push to Lagoon; only deleted previous problems")
		w.Flush()
		return nil
	}

	for _, r := range list.Results {
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

	if err == nil { // we have a token, and so we can proceed via the internal service call
		err = ProblemsToInsightsRemote(problems, LagoonInsightsRemoteEndpoint, bearerToken)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	fmt.Fprintln(w, "successfully pushed problems to Lagoon Remote")
	w.Flush()
	return nil
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

func DeleteProblems(serviceEndpoint string, bearerToken string) error {

	deleteEndpoint := fmt.Sprintf("%v/%v", serviceEndpoint, SourceName)

	bodyString, err := json.Marshal("{}")
	if err != nil {
		return err
	}

	req, _ := http.NewRequest(http.MethodDelete, deleteEndpoint, bytes.NewBuffer(bodyString))
	req.Header.Set("Authorization", bearerToken)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("there was an error deleting the problems at '%s' : %s", deleteEndpoint, response.Body)
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
