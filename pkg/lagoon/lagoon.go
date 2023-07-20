package lagoon

import (
	"context"
	"fmt"
	"os"

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

const SourceName string = "Shipshape"

var ApiBaseUrl string
var ApiToken string
var PushFacts bool

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

// AddFacts pushes the given facts to the Lagoon API.
func AddFacts(facts []Fact) error {
	MustHaveEnvVars()

	type AddFactInput struct{ Fact }
	type AddFactsByNameInput map[string]interface{}

	factsInput := []AddFactInput{}
	for _, f := range facts {
		factsInput = append(factsInput, AddFactInput{f})
	}
	var m struct {
		AddFactsByName []struct{ Id int } `graphql:"addFactsByName(input: $input)"`
	}
	variables := map[string]interface{}{"input": AddFactsByNameInput{
		"project":     os.Getenv("LAGOON_PROJECT"),
		"environment": os.Getenv("LAGOON_ENVIRONMENT"),
		"facts":       factsInput,
	}}

	qryStr, _ := graphql.ConstructMutation(&m, variables)
	log.WithFields(log.Fields{
		"query":     qryStr,
		"variables": fmt.Sprintf("%+v", variables),
	}).Debug("executing API mutation")
	err := Client.Mutate(context.Background(), &m, variables)
	if err != nil {
		return err
	}
	return nil
}

func DeleteFacts() error {
	envId, err := GetEnvironmentIdFromEnvVars()
	if err != nil {
		return err
	}
	var m struct {
		DeleteFactsFromSource string `graphql:"deleteFactsFromSource(input: {environment: $envId, source: $sourceName})"`
	}
	variables := map[string]interface{}{
		"envId":      envId,
		"sourceName": SourceName,
	}
	return Client.Mutate(context.Background(), &m, variables)
}

// ReplaceFacts deletes all the Shipshape facts and then adds the new ones.
func ReplaceFacts(facts []Fact) error {
	log.Debug("deleting facts before adding new")
	err := DeleteFacts()
	if err != nil {
		return err
	}
	log.Debug("adding new facts")
	return AddFacts(facts)
}
