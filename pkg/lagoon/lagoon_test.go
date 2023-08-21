package lagoon_test

import (
	"bytes"
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

func TestAddFacts(t *testing.T) {
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

	facts := []lagoon.Fact{
		{
			Name:     "fact1",
			Value:    "value1",
			Source:   "source1",
			Category: "cat1",
		},
		{
			Name:     "fact2",
			Value:    "value2",
			Source:   "source1",
			Category: "cat2",
		},
	}

	err := lagoon.AddFacts(facts)
	assert.NoError(err)
	assert.Equal(1, internal.MockLagoonNumCalls)
	assert.Equal("{\"query\":\"mutation ($input:AddFactsByNameInput!){"+
		"addFactsByName(input: $input){id}}\",\"variables\":{\"input\":{"+
		"\"environment\":\"bar\",\"facts\":[{\"name\":\"fact1\",\"value\":"+
		"\"value1\",\"source\":\"source1\",\"description\":\"\",\"category\":"+
		"\"cat1\"},{\"name\":\"fact2\",\"value\":\"value2\",\"source\":"+
		"\"source1\",\"description\":\"\",\"category\":\"cat2\"}],\"project\""+
		":\"foo\"}}}\n", internal.MockLagoonRequestBodies[0])
}

func TestDeleteFacts(t *testing.T) {
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

	err := lagoon.DeleteFacts()
	assert.NoError(err)
	assert.Equal(2, internal.MockLagoonNumCalls)
	assert.Equal("{\"query\":\"query ($ns:String!){"+
		"environmentByKubernetesNamespaceName(kubernetesNamespaceName: $ns)"+
		"{id}}\",\"variables\":{\"ns\":\"foo-bar\"}}\n", internal.MockLagoonRequestBodies[0])
	assert.Equal("{\"query\":\"mutation ($envId:Int!$sourceName:String!){"+
		"deleteFactsFromSource(input: {environment: $envId, source: "+
		"$sourceName})}\",\"variables\":{\"envId\":50,\"sourceName\":"+
		"\"Shipshape\"}}\n", internal.MockLagoonRequestBodies[1])
}

func TestReplaceFacts(t *testing.T) {
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

	facts := []lagoon.Fact{
		{
			Name:     "fact1",
			Value:    "value1",
			Source:   "source1",
			Category: "cat1",
		},
		{
			Name:     "fact2",
			Value:    "value2",
			Source:   "source1",
			Category: "cat2",
		},
	}

	err := lagoon.ReplaceFacts(facts)
	assert.NoError(err)
	assert.Equal(3, internal.MockLagoonNumCalls)
	assert.Equal("{\"query\":\"query ($ns:String!){"+
		"environmentByKubernetesNamespaceName(kubernetesNamespaceName: $ns)"+
		"{id}}\",\"variables\":{\"ns\":\"foo-bar\"}}\n", internal.MockLagoonRequestBodies[0])
	assert.Equal("{\"query\":\"mutation ($envId:Int!$sourceName:String!){"+
		"deleteFactsFromSource(input: {environment: $envId, source: "+
		"$sourceName})}\",\"variables\":{\"envId\":50,\"sourceName\":\""+
		"Shipshape\"}}\n", internal.MockLagoonRequestBodies[1])
	assert.Equal("{\"query\":\"mutation ($input:AddFactsByNameInput!){"+
		"addFactsByName(input: $input){id}}\",\"variables\":{\"input\":{"+
		"\"environment\":\"bar\",\"facts\":[{\"name\":\"fact1\",\"value\":"+
		"\"value1\",\"source\":\"source1\",\"description\":\"\",\"category\":"+
		"\"cat1\"},{\"name\":\"fact2\",\"value\":\"value2\",\"source\":"+
		"\"source1\",\"description\":\"\",\"category\":\"cat2\"}],\"project\""+
		":\"foo\"}}}\n", internal.MockLagoonRequestBodies[2])
}
