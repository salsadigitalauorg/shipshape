package lagoon

import (
	"context"
	"errors"
	"os"

	"github.com/hasura/go-graphql-client"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Fact struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Source      string `json:"source"`
	Environment int    `json:"environment"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

var LagoonApiBaseUrl string
var LagoonApiToken string
var Client *graphql.Client

func InitClient() {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: LagoonApiToken},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	Client = graphql.NewClient(LagoonApiBaseUrl+"/graphql", httpClient)
}

// GetEnvironmentIdFromEnvVars derives the environment id from shell variables
// LAGOON_PROJECT & LAGOON_ENVIRONMENT.
func GetEnvironmentIdFromEnvVars() (int, error) {
	project := os.Getenv("LAGOON_PROJECT")
	if project == "" {
		return 0, errors.New("project name required")
	}

	environment := os.Getenv("LAGOON_ENVIRONMENT")
	if environment == "" {
		return 0, errors.New("environment name required")
	}

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
