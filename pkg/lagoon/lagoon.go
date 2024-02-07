package lagoon

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"net/http"
	"os"
	"strings"

	"github.com/hasura/go-graphql-client"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
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

const SourceName string = "Shipshape"
const FactMaxValueLength int = 300

var ApiBaseUrl string
var ApiToken string
var PushProblemsToInsightRemote bool
var LagoonInsightsRemoteEndpoint string

var project string
var environment string

var Client *graphql.Client

func InitClient() {
	if Client != nil {
		return
	}
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ApiToken},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	Client = graphql.NewClient(ApiBaseUrl+"/graphql", httpClient)
}

func MustHaveEnvVars() {
	project = os.Getenv("LAGOON_PROJECT")
	environment = os.Getenv("LAGOON_ENVIRONMENT")
	if project == "" || environment == "" {
		log.Fatal("project & environment name required; please ensure both " +
			"LAGOON_PROJECT & LAGOON_ENVIRONMENT are set")
	}
}

// GetEnvironmentIdFromEnvVars derives the environment id from shell variables
// LAGOON_PROJECT & LAGOON_ENVIRONMENT.
func GetEnvironmentIdFromEnvVars() (int, error) {
	MustHaveEnvVars()

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
		return "", fmt.Errorf("Unable to load insights token from disk")
	}

	b, err := os.ReadFile(tokenLocation)
	if err != nil {
		return "", err
	}
	return strings.Trim(string(b), "\n"), nil
}

func ProcessResultList(w *bufio.Writer, list result.ResultList) error {
	problems := []Problem{}

	if list.TotalBreaches == 0 {
		InitClient()
		err := DeleteProblems()
		if err != nil {
			return err
		}
		fmt.Fprint(w, "no breach to push to Lagoon; only deleted previous problems")
		w.Flush()
		return nil
	}

	for iR, r := range list.Results {
		// let's marshall the breaches, they can be attached to the problem in the data field
		_, err := json.Marshal(r.Breaches)
		if err != nil {
			log.WithError(err).Fatal("Unable to marshall breach information")
		}

		breachMap := map[string]string{}
		for iB, b := range r.Breaches {
			breachName := fmt.Sprintf("[%d] %s", iR+iB+1, BreachFactName(b))
			value := BreachFactValue(b)
			if len(value) > FactMaxValueLength {
				value = value[:FactMaxValueLength-12] + "...TRUNCATED"
			}
			breachMap[breachName] = value
		}

		breachMapJson, err := json.Marshal(breachMap)
		if err != nil {
			log.WithError(err).Fatal("Unable to write problems to Insights Remote")
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

	InitClient()
	// first, let's try doing this via in-cluster functionality
	bearerToken, err := GetBearerTokenFromDisk(DefaultLagoonInsightsTokenLocation)
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
		return fmt.Errorf("There was an error sending the problems to '%s' : %s\n", serviceEndpoint, response.Body)
	}
	return nil
}

func DeleteProblems() error {
	envId, err := GetEnvironmentIdFromEnvVars()
	if err != nil {
		return err
	}
	var m struct {
		DeleteFactsFromSource string `graphql:"deleteProblemsFromSource(input: {environment: $envId, source: $sourceName, service:$service})"`
	}
	variables := map[string]interface{}{
		"envId":      envId,
		"sourceName": SourceName,
		"service":    "",
	}
	return Client.Mutate(context.Background(), &m, variables)
}

func BreachFactName(b result.Breach) string {
	var name string
	if result.BreachGetKeyLabel(b) != "" {
		name = fmt.Sprintf("%s: %s", result.BreachGetKeyLabel(b),
			result.BreachGetKey(b))
	} else if result.BreachGetKey(b) != "" {
		name = result.BreachGetKey(b)
	} else if result.BreachGetValueLabel(b) != "" {
		name = result.BreachGetValueLabel(b)
	} else {
		name = result.BreachGetCheckName(b) + " - " +
			string(result.BreachGetCheckType(b))
	}
	return name
}

func BreachFactValue(b result.Breach) string {
	value := result.BreachGetValue(b)
	if value == "" {
		value = strings.Join(result.BreachGetValues(b), ", ")
	}

	label := result.BreachGetValueLabel(b)
	if label == "" || BreachFactName(b) == label {
		return value
	} else {
		value = fmt.Sprintf("%s: %s", label, value)
	}

	expected := result.BreachGetExpectedValue(b)
	if expected == "" {
		return value
	} else {
		value = fmt.Sprintf("expected: %s, %s", expected, value)
	}
	return value
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
