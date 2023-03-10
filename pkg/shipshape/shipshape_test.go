package shipshape_test

import (
	"io"
	"os"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	. "github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape/testdata/testchecks"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestInit(t *testing.T) {
	assert := assert.New(t)

	t.Run("defaultValues", func(t *testing.T) {
		currDir, _ := os.Getwd()
		err := Init("", []string{}, []string{}, false, false, "")
		assert.NoError(err)
		assert.Equal(currDir, config.ProjectDir)
		assert.Equal(config.Config{
			ProjectDir:   currDir,
			Checks:       config.CheckMap{},
			FailSeverity: config.HighSeverity,
		}, RunConfig)
		assert.Equal(logrus.WarnLevel, logrus.GetLevel())
	})

	t.Run("projectDirIsSet", func(t *testing.T) {
		err := Init("foo", []string{}, []string{}, false, false, "warn")
		assert.NoError(err)
		assert.Equal("foo", config.ProjectDir)
	})
}

func TestReadAndParseConfig(t *testing.T) {
	assert := assert.New(t)

	currLogOut := logrus.StandardLogger().Out
	defer logrus.SetOutput(currLogOut)
	logrus.SetOutput(io.Discard)

	t.Run("nonExistentFile", func(t *testing.T) {
		err := ReadAndParseConfig("", []string{"testdata/nonexistent.yml"})
		assert.Error(err)
		assert.Equal("open testdata/nonexistent.yml: no such file or directory", err.Error())
	})

	t.Run("existingFile", func(t *testing.T) {
		err := ReadAndParseConfig("", []string{"testdata/shipshape.yml"})
		assert.NoError(err)
	})

	t.Run("configFileMerge", func(t *testing.T) {
		err := ReadAndParseConfig("", []string{
			"testdata/merge/config-a.yml",
			"testdata/merge/config-b.yml",
		})
		assert.NoError(err)
		mergedCfg := RunConfig

		RunConfig = config.Config{}
		err = ReadAndParseConfig("", []string{
			"testdata/merge/config-result.yml",
		})
		assert.NoError(err)
		resultingCfg := RunConfig

		assert.EqualValues(resultingCfg, mergedCfg)
	})
}

func TestParseConfigData(t *testing.T) {
	assert := assert.New(t)

	currLogOut := logrus.StandardLogger().Out
	defer logrus.SetOutput(currLogOut)

	t.Run("invalidData", func(t *testing.T) {
		testchecks.RegisterChecks()
		logrus.SetOutput(io.Discard)
		invalidData := `
checks:
  test-check-1: foo
`
		err := ParseConfigData([][]byte{[]byte(invalidData)})
		assert.Error(err)
		assert.Contains(err.Error(), "yaml: unmarshal errors")

	})

	t.Run("validData", func(t *testing.T) {
		testchecks.RegisterChecks()
		data := `
checks:
  test-check-1:
    - name: My test check 1
      foo: baz
  test-check-2:
    - name: My first test check 2
      bar: zoom
    - name: My second test check 2
      bar: zap
`
		err := ParseConfigData([][]byte{[]byte(data)})
		assert.NoError(err)

		if !assert.Len(RunConfig.Checks[testchecks.TestCheck1], 1) {
			t.FailNow()
		}
		if !assert.Len(RunConfig.Checks[testchecks.TestCheck2], 2) {
			t.FailNow()
		}

		tc1, ok := RunConfig.Checks[testchecks.TestCheck1][0].(*testchecks.TestCheck1Check)
		assert.True(ok)
		assert.Equal("My test check 1", tc1.Name)
		assert.Equal("baz", tc1.Foo)

		tc2, ok := RunConfig.Checks[testchecks.TestCheck2][0].(*testchecks.TestCheck2Check)
		assert.True(ok)
		assert.Equal("My first test check 2", tc2.Name)
		assert.Equal("zoom", tc2.Bar)

		tc22, ok := RunConfig.Checks[testchecks.TestCheck2][1].(*testchecks.TestCheck2Check)
		assert.True(ok)
		assert.Equal("My second test check 2", tc22.Name)
		assert.Equal("zap", tc22.Bar)
	})
}

func TestRunChecks(t *testing.T) {
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

	rl := RunChecks()
	assert.Equal(uint32(2), rl.TotalChecks)
	assert.Equal(uint32(2), rl.TotalBreaches)
	assert.EqualValues(map[config.CheckType]int{
		testchecks.TestCheck1: 1,
		testchecks.TestCheck2: 1,
	}, rl.BreachCountByType)
	assert.ElementsMatch([]config.Result{
		{Name: "test1stcheck", Severity: "normal", CheckType: "test-check-1", Status: "Fail", Passes: []string(nil), Failures: []string{"no data available"}, Warnings: []string(nil)},
		{Name: "test2ndcheck", Severity: "normal", CheckType: "test-check-2", Status: "Fail", Passes: []string(nil), Failures: []string{"no data available"}, Warnings: []string(nil)}},
		rl.Results)
}
