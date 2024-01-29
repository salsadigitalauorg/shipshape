package shipshape_test

import (
	"bufio"
	"bytes"
	"net/http"
	"os"
	"testing"
	"text/tabwriter"

	"github.com/salsadigitalauorg/shipshape/pkg/checks/file"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/lagoon"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	. "github.com/salsadigitalauorg/shipshape/pkg/shipshape"

	"github.com/hasura/go-graphql-client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestTableDisplay(t *testing.T) {
	assert := assert.New(t)

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)
	RunResultList = result.ResultList{}
	TableDisplay(w)
	assert.Equal(
		"No result available; ensure your shipshape.yml is configured correctly.\n",
		buf.String())

	buf = bytes.Buffer{}
	RunResultList = result.ResultList{Results: []result.Result{{Name: "a", Status: result.Pass}}}
	TableDisplay(w)
	assert.Equal("NAME   STATUS   PASSES   FAILS\n"+
		"a      Pass              \n", buf.String())

	buf = bytes.Buffer{}
	RunResultList = result.ResultList{
		Results: []result.Result{
			{Name: "a", Status: result.Pass},
			{Name: "b", Status: result.Pass},
			{Name: "c", Status: result.Pass},
		},
	}
	TableDisplay(w)
	assert.Equal("NAME   STATUS   PASSES   FAILS\n"+
		"a      Pass              \n"+
		"b      Pass              \n"+
		"c      Pass              \n",
		buf.String())

	buf = bytes.Buffer{}
	RunResultList = result.ResultList{
		Results: []result.Result{
			{
				Name:   "a",
				Status: result.Pass,
				Passes: []string{"Pass a", "Pass ab"},
			},
			{
				Name:   "b",
				Status: result.Pass,
				Passes: []string{"Pass b", "Pass bb", "Pass bc"},
			},
			{
				Name:     "c",
				Status:   result.Fail,
				Failures: []string{"Fail c", "Fail cb"},
			},
			{
				Name:     "d",
				Status:   result.Fail,
				Passes:   []string{"Pass d", "Pass db"},
				Failures: []string{"Fail c", "Fail cb"},
			},
		},
	}
	TableDisplay(w)
	assert.Equal("NAME   STATUS   PASSES    FAILS\n"+
		"a      Pass     Pass a    \n"+
		"                Pass ab   \n"+
		"b      Pass     Pass b    \n"+
		"                Pass bb   \n"+
		"                Pass bc   \n"+
		"c      Fail               Fail c\n"+
		"                          Fail cb\n"+
		"d      Fail     Pass d    Fail c\n"+
		"                Pass db   Fail cb\n",
		buf.String())
}

func TestSimpleDisplay(t *testing.T) {
	assert := assert.New(t)

	t.Run("noResult", func(t *testing.T) {
		RunResultList = result.NewResultList(false)
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		SimpleDisplay(w)
		assert.Equal("No result available; ensure your shipshape.yml is configured correctly.\n", buf.String())
	})

	t.Run("topShape", func(t *testing.T) {
		RunResultList = result.NewResultList(false)
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		RunResultList.Results = append(RunResultList.Results, result.Result{
			Name: "a", Status: result.Pass})
		buf = bytes.Buffer{}
		SimpleDisplay(w)
		assert.Equal("Ship is in top shape; no breach detected!\n", buf.String())
	})

	t.Run("breachesDetected", func(t *testing.T) {
		RunResultList = result.NewResultList(false)
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		RunResultList.Results = append(RunResultList.Results, result.Result{
			Name:     "b",
			Status:   result.Fail,
			Failures: []string{"Fail b"}})
		buf = bytes.Buffer{}
		SimpleDisplay(w)
		assert.Equal("# Breaches were detected\n\n  ### b\n     -- Fail b\n\n", buf.String())
	})

	t.Run("topShapeRemediating", func(t *testing.T) {
		RunResultList = result.ResultList{RemediationPerformed: true}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		RunResultList.Results = append(RunResultList.Results, result.Result{
			Name: "a", Status: result.Pass})
		buf = bytes.Buffer{}
		SimpleDisplay(w)
		assert.Equal("Ship is in top shape; no breach detected!\n", buf.String())
	})

	t.Run("allBreachesRemediated", func(t *testing.T) {
		RunResultList = result.ResultList{RemediationPerformed: true}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		RunResultList.TotalRemediations = 1
		RunResultList.Results = append(RunResultList.Results, result.Result{
			Name: "a", Status: result.Pass, Remediations: []string{"fixed 1"}})
		buf = bytes.Buffer{}
		SimpleDisplay(w)
		assert.Equal("Breaches were detected but were all fixed successfully!\n\n"+
			"  ### a\n     -- fixed 1\n\n", buf.String())
	})

	t.Run("someBreachesRemediated", func(t *testing.T) {
		RunResultList = result.ResultList{RemediationPerformed: true}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		RunResultList.TotalRemediations = 1
		RunResultList.TotalBreaches = 1
		RunResultList.Results = append(RunResultList.Results, result.Result{
			Name: "a", Status: result.Fail, Remediations: []string{"fixed 1"}})
		buf = bytes.Buffer{}
		SimpleDisplay(w)
		assert.Equal("Breaches were detected but not all of them could be "+
			"fixed as they are either not supported yet or there were errors "+
			"when trying to remediate.\n\n"+
			"# Remediations\n\n  ### a\n     -- fixed 1\n\n"+
			"# Non-remediated breaches\n\n", buf.String())
	})

	t.Run("noBreachRemediated", func(t *testing.T) {
		RunResultList = result.ResultList{RemediationPerformed: true}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		RunResultList.TotalBreaches = 1
		RunResultList.TotalRemediations = 0
		RunResultList.Results = append(RunResultList.Results, result.Result{
			Name: "a", Status: result.Fail})
		buf = bytes.Buffer{}
		SimpleDisplay(w)
		assert.Equal("Breaches were detected but not all of them could be "+
			"fixed as they are either not supported yet or there were errors "+
			"when trying to remediate.\n\n"+
			"# Remediations\n\n"+
			"# Non-remediated breaches\n\n", buf.String())
	})
}
func TestJUnit(t *testing.T) {
	assert := assert.New(t)

	RunResultList = result.NewResultList(false)
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	JUnit(w)
	assert.Equal(`<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" errors="0"></testsuites>
`, buf.String())

	RunConfig.Checks = config.CheckMap{file.File: []config.Check{&file.FileCheck{
		CheckBase: config.CheckBase{Name: "a"},
	}}}
	RunResultList.Results = append(RunResultList.Results, result.Result{
		Name: "a", Status: result.Pass})
	buf = bytes.Buffer{}
	JUnit(w)
	assert.Equal(`<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" errors="0">
    <testsuite name="file" tests="0" errors="0">
        <testcase name="a" classname="a"></testcase>
    </testsuite>
</testsuites>
`, buf.String())

	RunConfig.Checks[file.File] = append(RunConfig.Checks[file.File], &file.FileCheck{
		CheckBase: config.CheckBase{Name: "b"},
	})
	RunResultList.Results = append(RunResultList.Results, result.Result{
		Name:     "b",
		Status:   result.Fail,
		Failures: []string{"Fail b"}})
	buf = bytes.Buffer{}
	JUnit(w)
	assert.Equal(`<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" errors="0">
    <testsuite name="file" tests="0" errors="0">
        <testcase name="a" classname="a"></testcase>
        <testcase name="b" classname="b">
            <error message="Fail b"></error>
        </testcase>
    </testsuite>
</testsuites>
`, buf.String())
}

func TestLagoonProblems(t *testing.T) {
	assert := assert.New(t)

	t.Run("noResult", func(t *testing.T) {
		RunResultList = result.NewResultList(false)

		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		LagoonProblems(w)
		assert.Equal("[]", buf.String())
	})

	t.Run("noResultPushProblems", func(t *testing.T) {
		RunResultList = result.NewResultList(false)

		svr := internal.MockLagoonServer()
		lagoon.Client = graphql.NewClient(svr.URL, http.DefaultClient)
		lagoon.PushProblems = true
		origOutput := logrus.StandardLogger().Out
		var logbuf bytes.Buffer
		logrus.SetOutput(&logbuf)
		os.Setenv("LAGOON_PROJECT", "foo")
		os.Setenv("LAGOON_ENVIRONMENT", "bar")
		defer func() {
			svr.Close()
			internal.MockLagoonReset()
			lagoon.Client = nil
			os.Unsetenv("LAGOON_PROJECT")
			os.Unsetenv("LAGOON_ENVIRONMENT")
			logrus.SetOutput(origOutput)
			lagoon.PushProblems = false
		}()

		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		LagoonProblems(w)
		assert.Equal(2, internal.MockLagoonNumCalls)
		assert.Equal("{\"query\":\"query ($ns:String!){"+
			"environmentByKubernetesNamespaceName(kubernetesNamespaceName: "+
			"$ns){id}}\",\"variables\":{\"ns\":\"foo-bar\"}}\n", internal.MockLagoonRequestBodies[0])
		assert.Equal("{\"query\":\"mutation ($envId:Int!$sourceName:String!)"+
			"{deleteFactsFromSource(input: {environment: $envId, source: "+
			"$sourceName})}\",\"variables\":{\"envId\":50,\"sourceName\":\""+
			"Shipshape\"}}\n", internal.MockLagoonRequestBodies[1])
		assert.Equal("no breach to push to Lagoon; only deleted previous facts", buf.String())
	})

	t.Run("breachesDetected", func(t *testing.T) {
		RunConfig.Checks = config.CheckMap{file.File: []config.Check{
			&file.FileCheck{CheckBase: config.CheckBase{Name: "a"}}}}
		RunResultList = result.NewResultList(false)
		RunResultList.Results = append(RunResultList.Results, result.Result{
			Name:   "a",
			Status: result.Fail,
			Breaches: []result.Breach{result.ValueBreach{
				CheckName: "a",
				Value:     "Fail a",
				CheckType: "file",
			}},
		})
		RunResultList.TotalBreaches = 1

		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		LagoonProblems(w)
		assert.Equal("[{\"name\":\"[1] a - file\",\"value\":\"Fail a\",\"source\":"+
			"\"Shipshape\",\"description\":\"a\",\"category\":\"file\"}]",
			buf.String())
	})

	t.Run("pushToLagoon", func(t *testing.T) {
		RunConfig.Checks = config.CheckMap{file.File: []config.Check{
			&file.FileCheck{CheckBase: config.CheckBase{Name: "a"}}}}
		RunResultList = result.NewResultList(false)
		RunResultList.Results = append(RunResultList.Results, result.Result{
			Name:   "a",
			Status: result.Fail,
			Breaches: []result.Breach{result.ValueBreach{
				CheckName: "a",
				Value:     "Fail a",
				CheckType: "file",
			}},
		})
		RunResultList.TotalBreaches = 1

		lagoon.PushProblems = true

		svr := internal.MockLagoonServer()
		lagoon.Client = graphql.NewClient(svr.URL, http.DefaultClient)
		origOutput := logrus.StandardLogger().Out
		defer func() {
			svr.Close()
			internal.MockLagoonReset()
			lagoon.Client = nil
			os.Unsetenv("LAGOON_PROJECT")
			os.Unsetenv("LAGOON_ENVIRONMENT")
			logrus.SetOutput(origOutput)
			lagoon.PushProblems = false
		}()

		var logbuf bytes.Buffer
		logrus.SetOutput(&logbuf)

		os.Setenv("LAGOON_PROJECT", "foo")
		os.Setenv("LAGOON_ENVIRONMENT", "bar")

		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		LagoonProblems(w)
		assert.Equal(3, internal.MockLagoonNumCalls)
		assert.Equal("{\"query\":\"query ($ns:String!){"+
			"environmentByKubernetesNamespaceName(kubernetesNamespaceName: "+
			"$ns){id}}\",\"variables\":{\"ns\":\"foo-bar\"}}\n", internal.MockLagoonRequestBodies[0])
		assert.Equal("{\"query\":\"mutation ($envId:Int!$sourceName:String!)"+
			"{deleteFactsFromSource(input: {environment: $envId, source: "+
			"$sourceName})}\",\"variables\":{\"envId\":50,\"sourceName\":\""+
			"Shipshape\"}}\n", internal.MockLagoonRequestBodies[1])
		assert.Equal("{\"query\":\"mutation ($input:AddFactsByNameInput!){"+
			"addFactsByName(input: $input){id}}\",\"variables\":{\"input\":{"+
			"\"environment\":\"bar\",\"facts\":[{\"name\":\"[1] a - file\",\"value\""+
			":\"Fail a\",\"source\":\"Shipshape\",\"description\":\"a\",\"category"+
			"\":\"file\"}],\"project\":\"foo\"}}}\n",
			internal.MockLagoonRequestBodies[2])
		assert.Contains(logbuf.String(),
			"level=info msg=\"fetching environment id\" namespace=foo-bar\n")
		assert.Equal("successfully pushed facts to the Lagoon api", buf.String())
	})

	t.Run("pushToLagoonOversizedText", func(t *testing.T) {
		RunConfig.Checks = config.CheckMap{file.File: []config.Check{
			&file.FileCheck{CheckBase: config.CheckBase{Name: "a"}}}}
		RunResultList = result.NewResultList(false)

		// Oversized text
		oversizedText := ""
		for i := 0; i <= lagoon.FactMaxValueLength; i++ {
			oversizedText += "a"
		}

		RunResultList.Results = append(RunResultList.Results, result.Result{
			Name:   "a",
			Status: result.Fail,
			Breaches: []result.Breach{result.ValueBreach{
				CheckName: "a",
				Value:     oversizedText,
				CheckType: "file",
			}},
		})
		RunResultList.TotalBreaches = 1

		lagoon.PushProblems = true

		svr := internal.MockLagoonServer()
		lagoon.Client = graphql.NewClient(svr.URL, http.DefaultClient)
		origOutput := logrus.StandardLogger().Out
		defer func() {
			svr.Close()
			internal.MockLagoonReset()
			lagoon.Client = nil
			os.Unsetenv("LAGOON_PROJECT")
			os.Unsetenv("LAGOON_ENVIRONMENT")
			logrus.SetOutput(origOutput)
			lagoon.PushProblems = false
		}()

		var logbuf bytes.Buffer
		logrus.SetOutput(&logbuf)

		os.Setenv("LAGOON_PROJECT", "foo")
		os.Setenv("LAGOON_ENVIRONMENT", "bar")

		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		LagoonProblems(w)
		assert.Equal(3, internal.MockLagoonNumCalls)
		assert.Equal("{\"query\":\"query ($ns:String!){"+
			"environmentByKubernetesNamespaceName(kubernetesNamespaceName: "+
			"$ns){id}}\",\"variables\":{\"ns\":\"foo-bar\"}}\n", internal.MockLagoonRequestBodies[0])
		assert.Equal("{\"query\":\"mutation ($envId:Int!$sourceName:String!)"+
			"{deleteFactsFromSource(input: {environment: $envId, source: "+
			"$sourceName})}\",\"variables\":{\"envId\":50,\"sourceName\":\""+
			"Shipshape\"}}\n", internal.MockLagoonRequestBodies[1])
		assert.Equal("{\"query\":\"mutation ($input:AddFactsByNameInput!){"+
			"addFactsByName(input: $input){id}}\",\"variables\":{\"input\":{"+
			"\"environment\":\"bar\",\"facts\":[{\"name\":\"[1] a - file\",\"value\""+
			":\"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"+
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"+
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"+
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"+
			"aaaaaaaaaaa...TRUNCATED\",\"source\":\"Shipshape\",\"description\":\"a"+
			"\",\"category\":\"file\"}],\"project\":\"foo\"}}}\n",
			internal.MockLagoonRequestBodies[2])
		assert.Contains(logbuf.String(),
			"level=info msg=\"fetching environment id\" namespace=foo-bar\n")
		assert.Equal("successfully pushed facts to the Lagoon api", buf.String())
	})
}
