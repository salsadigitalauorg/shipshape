package lagoon_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/lagoon"

	"github.com/hasura/go-graphql-client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestInitClient(t *testing.T) {
	assert := assert.New(t)

	assert.Nil(lagoon.Client)
	lagoon.InitClient()
	assert.NotNil(lagoon.Client)
}

func TestMustHaveEnvVars(t *testing.T) {
	assert := assert.New(t)

	origOutput := logrus.StandardLogger().Out
	defer func() {
		logrus.StandardLogger().ExitFunc = nil
		logrus.SetOutput(origOutput)
	}()

	var buf bytes.Buffer
	logrus.SetOutput(&buf)

	var fatal bool
	logrus.StandardLogger().ExitFunc = func(retCode int) {
		fatal = (retCode == 1)
	}

	t.Run("varsSet", func(t *testing.T) {
		os.Setenv("LAGOON_PROJECT", "foo")
		os.Setenv("LAGOON_ENVIRONMENT", "bar")
		defer func() {
			os.Unsetenv("LAGOON_PROJECT")
			os.Unsetenv("LAGOON_ENVIRONMENT")
		}()

		lagoon.MustHaveEnvVars()
		assert.False(fatal)
	})

	t.Run("varsUnset", func(t *testing.T) {
		lagoon.MustHaveEnvVars()
		assert.True(fatal)
		assert.Contains(buf.String(), "project & environment name required; please ensure both LAGOON_PROJECT & LAGOON_ENVIRONMENT are set")
	})
}

func TestGetEnvironmentIdFromEnvVars(t *testing.T) {
	assert := assert.New(t)

	svr := internal.MockLagoonServer()
	lagoon.Client = graphql.NewClient(svr.URL, http.DefaultClient)
	origOutput := logrus.StandardLogger().Out
	os.Setenv("LAGOON_PROJECT", "foo")
	os.Setenv("LAGOON_ENVIRONMENT", "bar")
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	defer func() {
		svr.Close()
		internal.MockLagoonReset()
		lagoon.Client = nil
		os.Unsetenv("LAGOON_PROJECT")
		os.Unsetenv("LAGOON_ENVIRONMENT")
		logrus.SetOutput(origOutput)
	}()

	_, err := lagoon.GetEnvironmentIdFromEnvVars()
	assert.NoError(err)
	assert.Equal(1, internal.MockLagoonNumCalls)
	assert.Equal("{\"query\":\"query ($ns:String!){"+
		"environmentByKubernetesNamespaceName(kubernetesNamespaceName: $ns)"+
		"{id}}\",\"variables\":{\"ns\":\"foo-bar\"}}\n", internal.MockLagoonRequestBodies[0])
}

func TestDeleteProblems(t *testing.T) {
	assert := assert.New(t)

	svr := internal.MockLagoonServer()
	lagoon.Client = graphql.NewClient(svr.URL, http.DefaultClient)
	origOutput := logrus.StandardLogger().Out
	os.Setenv("LAGOON_PROJECT", "foo")
	os.Setenv("LAGOON_ENVIRONMENT", "bar")
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	defer func() {
		svr.Close()
		internal.MockLagoonReset()
		lagoon.Client = nil
		os.Unsetenv("LAGOON_PROJECT")
		os.Unsetenv("LAGOON_ENVIRONMENT")
		logrus.SetOutput(origOutput)
	}()

	err := lagoon.DeleteProblems()
	assert.NoError(err)
	assert.Equal(2, internal.MockLagoonNumCalls)
	assert.Equal("{\"query\":\"query ($ns:String!){"+
		"environmentByKubernetesNamespaceName(kubernetesNamespaceName: $ns)"+
		"{id}}\",\"variables\":{\"ns\":\"foo-bar\"}}\n", internal.MockLagoonRequestBodies[0])
	assert.Equal("{\"query\":\"mutation ($envId:Int!$service:String!$sourceName:String!)"+
		"{deleteProblemsFromSource(input: {environment: $envId, source: "+
		"$sourceName, service:$service})}\",\"variables\":{\"envId\":50,\"service\":\"\",\"sourceName\":"+
		"\"Shipshape\"}}\n", internal.MockLagoonRequestBodies[1])
}

func Test_GetBearerTokenFromDisk(t *testing.T) {
	type args struct {
		tokenLocation string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Successfully loads token from disk",
			args: args{
				tokenLocation: "./testdata/insightsbearertoken",
			},
			want:    "bearertokentext",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lagoon.GetBearerTokenFromDisk(tt.args.tokenLocation)
			if !tt.wantErr(t, err, fmt.Sprintf("getBearerTokenFromDisk(%v)", tt.args.tokenLocation)) {
				return
			}
			assert.Equalf(t, tt.want, got, "getBearerTokenFromDisk(%v)", tt.args.tokenLocation)
		})
	}
}

func Test_ProblemsToInsightsRemote(t *testing.T) {
	type args struct {
		problems    []lagoon.Problem
		bearerToken string
	}
	tests := []struct {
		name          string
		args          args
		wantErr       assert.ErrorAssertionFunc
		testBodyEqual bool
	}{
		{
			name: "Successful post",
			args: args{
				problems: []lagoon.Problem{
					{
						EnvironmentId: 1,
						Identifier:    "problem1",
						Severity:      "HIGH",
						Service:       "serviceName",
					},
				},
				bearerToken: "bearertoken",
			},
			wantErr:       assert.NoError,
			testBodyEqual: true,
		},
	}
	for _, tt := range tests {

		mockServerData := internal.MockInsightsRemoteTestState{}

		serv := internal.MockRemoteInsightsServer(&mockServerData)

		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, lagoon.ProblemsToInsightsRemote(tt.args.problems, serv.URL, tt.args.bearerToken), fmt.Sprintf("problemsToInsightsRemote(%v, %v)", tt.args.problems, tt.args.bearerToken))
			if tt.testBodyEqual == true { //let's check the data we sent through seems correct
				bodyString := mockServerData.LastCallBody
				var problems []lagoon.Problem
				err := json.Unmarshal([]byte(bodyString), &problems)
				assert.NoError(t, err)
				assert.Equalf(t, tt.args.problems, problems, "Unmarshalled Body not Equal")
			}
		})
	}
}
