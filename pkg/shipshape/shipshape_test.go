package shipshape_test

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	. "github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape/testdata/testchecks"
)

func TestInit(t *testing.T) {
	assert := assert.New(t)

	t.Run("defaultValues", func(t *testing.T) {
		config.Files = []string{"testdata/shipshape.yml"}
		err := Init()
		assert.NoError(err)
		assert.Equal(config.Config{
			Checks: config.CheckMap{},
		}, RunConfig)
	})
}

func TestRunChecks(t *testing.T) {
	currLogOut := logrus.StandardLogger().Out
	defer logrus.SetOutput(currLogOut)
	logrus.SetOutput(io.Discard)

	assert := assert.New(t)

	test1stCheck := &testchecks.TestCheck1Check{}
	test2ndCheck := &testchecks.TestCheck2Check{}
	yaml.Unmarshal([]byte("name: test1stcheck"), test1stCheck)
	test1stCheck.Init(testchecks.TestCheck1)
	yaml.Unmarshal([]byte("name: test2ndcheck"), test2ndCheck)
	test2ndCheck.Init(testchecks.TestCheck2)
	RunConfig = config.Config{
		Checks: config.CheckMap{
			testchecks.TestCheck1: {test1stCheck},
			testchecks.TestCheck2: {test2ndCheck},
		},
	}

	RunResultList = result.NewResultList(false)
	Run()
	assert.Equal(uint32(2), RunResultList.TotalChecks)
	assert.Equal(uint32(2), RunResultList.TotalBreaches)
	assert.EqualValues(map[string]int{
		string(testchecks.TestCheck1): 1,
		string(testchecks.TestCheck2): 1,
	}, RunResultList.BreachCountByType)
	assert.ElementsMatch([]result.Result{
		{
			Name:      "test1stcheck",
			Severity:  "normal",
			CheckType: "test-check-1",
			Status:    "Fail",
			Passes:    []string(nil),
			Breaches: []breach.Breach{&breach.ValueBreach{
				BreachType: "value",
				CheckType:  "test-check-1",
				CheckName:  "test1stcheck",
				Severity:   "normal",
				Value:      "no data available",
			}},
			Warnings: []string(nil),
		},
		{
			Name:      "test2ndcheck",
			Severity:  "normal",
			CheckType: "test-check-2",
			Status:    "Fail",
			Passes:    []string(nil),
			Breaches: []breach.Breach{&breach.ValueBreach{
				BreachType: "value",
				CheckType:  "test-check-2",
				CheckName:  "test2ndcheck",
				Severity:   "normal",
				Value:      "no data available",
			}},
			Warnings: []string(nil),
		}},
		RunResultList.Results)
}
