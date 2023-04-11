package lagoon_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hasura/go-graphql-client"
	"github.com/salsadigitalauorg/shipshape/pkg/lagoon"
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

	var svr *httptest.Server
	var reqBody []byte

	svr = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ = io.ReadAll(r.Body)
		fmt.Fprintf(w, "{}")
	}))

	httpClient := http.DefaultClient
	lagoon.Client = graphql.NewClient(svr.URL, httpClient)
	origOutput := logrus.StandardLogger().Out
	defer func() {
		svr.Close()
		lagoon.Client = nil
		os.Unsetenv("LAGOON_PROJECT")
		os.Unsetenv("LAGOON_ENVIRONMENT")
		logrus.SetOutput(origOutput)
	}()

	os.Setenv("LAGOON_PROJECT", "foo")
	os.Setenv("LAGOON_ENVIRONMENT", "bar")

	var buf bytes.Buffer
	logrus.SetOutput(&buf)

	_, err := lagoon.GetEnvironmentIdFromEnvVars()
	assert.NoError(err)
	assert.Equal("{\"query\":\"query ($ns:String!){"+
		"environmentByKubernetesNamespaceName(kubernetesNamespaceName: $ns)"+
		"{id}}\",\"variables\":{\"ns\":\"foo-bar\"}}\n", string(reqBody))
}

func TestAddFacts(t *testing.T) {
	assert := assert.New(t)

	var svr *httptest.Server
	var reqBody []byte

	svr = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ = io.ReadAll(r.Body)
		fmt.Fprintf(w, "{}")
	}))

	httpClient := http.DefaultClient
	lagoon.Client = graphql.NewClient(svr.URL, httpClient)
	origOutput := logrus.StandardLogger().Out
	defer func() {
		svr.Close()
		lagoon.Client = nil
		os.Unsetenv("LAGOON_PROJECT")
		os.Unsetenv("LAGOON_ENVIRONMENT")
		logrus.SetOutput(origOutput)
	}()

	os.Setenv("LAGOON_PROJECT", "foo")
	os.Setenv("LAGOON_ENVIRONMENT", "bar")

	var buf bytes.Buffer
	logrus.SetOutput(&buf)

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
	assert.Equal("{\"query\":\"mutation ($input:AddFactsByNameInput!){"+
		"addFactsByName(input: $input){id}}\",\"variables\":{\"input\":{"+
		"\"environment\":\"bar\",\"facts\":[{\"name\":\"fact1\",\"value\":"+
		"\"value1\",\"source\":\"source1\",\"description\":\"\",\"category\":"+
		"\"cat1\"},{\"name\":\"fact2\",\"value\":\"value2\",\"source\":"+
		"\"source1\",\"description\":\"\",\"category\":\"cat2\"}],\"project\""+
		":\"foo\"}}}\n", string(reqBody))
}

func TestReplaceFacts(t *testing.T) {
	assert := assert.New(t)

	var svr *httptest.Server
	var reqBodies [][]byte

	svr = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := io.ReadAll(r.Body)
		reqBodies = append(reqBodies, reqBody)
		// Response for the first request, environment id.
		if len(reqBodies) == 1 {
			fmt.Fprintf(w, "{\"id\": 50}")
		} else if len(reqBodies) == 2 { // Response for the deletion.
			fmt.Fprintf(w, "{\"data\":{\"deleteFactsFromSource\":\"success\"}}")
		} else if len(reqBodies) == 3 { // Response for the add.
			fmt.Fprintf(w, "{}")
		}
	}))

	httpClient := http.DefaultClient
	lagoon.Client = graphql.NewClient(svr.URL, httpClient)
	origOutput := logrus.StandardLogger().Out
	defer func() {
		svr.Close()
		lagoon.Client = nil
		os.Unsetenv("LAGOON_PROJECT")
		os.Unsetenv("LAGOON_ENVIRONMENT")
		logrus.SetOutput(origOutput)
	}()

	os.Setenv("LAGOON_PROJECT", "foo")
	os.Setenv("LAGOON_ENVIRONMENT", "bar")

	var buf bytes.Buffer
	logrus.SetOutput(&buf)

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
	assert.Len(reqBodies, 3)
	assert.Equal("{\"query\":\"query ($ns:String!){"+
		"environmentByKubernetesNamespaceName(kubernetesNamespaceName: $ns)"+
		"{id}}\",\"variables\":{\"ns\":\"foo-bar\"}}\n", string(reqBodies[0]))
	assert.Equal("{\"query\":\"mutation ($envId:Int!$sourceName:String!){"+
		"deleteFactsFromSource(input: {environment: $envId, source: "+
		"$sourceName})}\",\"variables\":{\"envId\":0,\"sourceName\":\""+
		"Shipshape\"}}\n", string(reqBodies[1]))
	assert.Equal("{\"query\":\"mutation ($input:AddFactsByNameInput!){"+
		"addFactsByName(input: $input){id}}\",\"variables\":{\"input\":{"+
		"\"environment\":\"bar\",\"facts\":[{\"name\":\"fact1\",\"value\":"+
		"\"value1\",\"source\":\"source1\",\"description\":\"\",\"category\":"+
		"\"cat1\"},{\"name\":\"fact2\",\"value\":\"value2\",\"source\":"+
		"\"source1\",\"description\":\"\",\"category\":\"cat2\"}],\"project\""+
		":\"foo\"}}}\n", string(reqBodies[2]))
}
