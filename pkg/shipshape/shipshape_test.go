package shipshape_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestReadAndParseConfigFileExistence(t *testing.T) {
	assert := assert.New(t)

	_, err := shipshape.ReadAndParseConfig("", []string{"testdata/nonexistent.yml"})
	assert.Error(err)
	assert.Equal("open testdata/nonexistent.yml: no such file or directory", err.Error())

	_, err = shipshape.ReadAndParseConfig("", []string{"testdata/shipshape.yml"})
	assert.NoError(err)
}

func TestReadAndParseConfigFileMerge(t *testing.T) {
	assert := assert.New(t)

	_, err := shipshape.ReadAndParseConfig("", []string{
		"testdata/merge/config-a.yml",
		"testdata/merge/config-b.yml",
	})
	assert.NoError(err)
}

func TestParseConfig(t *testing.T) {
	assert := assert.New(t)

	invalidData := `
checks:
  yaml: foo
`
	cfg := shipshape.Config{}
	err := shipshape.ParseConfig([]byte(invalidData), "", &cfg)
	assert.Error(err)
	assert.Contains(err.Error(), "yaml: unmarshal errors")

	data := `
checks:
  drush-yaml:
    - name: My db check
      config-name: shipshape.extension
  yaml:
    - name: My file check
      file: shipshape.extension.yml
      path: config/sync
      values:
        - key: profile
          value: govcms
    - name: File check - Ignore missing
      file: shipshape.extension.yml
      path: config/sync
      ignore-missing: true
      values:
        - key: profile
          value: govcms
  foo:
    - name: bar
`
	cfg = shipshape.Config{}
	err = shipshape.ParseConfig([]byte(data), "", &cfg)
	assert.NoError(err)
	cfg.Init()

	currDir, err := os.Getwd()
	assert.NoError(err)
	assert.Equal(currDir, cfg.ProjectDir)
	assert.Len(cfg.Checks[drupal.DrushYaml], 1)
	assert.Len(cfg.Checks[shipshape.Yaml], 2)

	dyc, ok := cfg.Checks[drupal.DrushYaml][0].(*drupal.DrushYamlCheck)
	assert.True(ok)
	assert.Equal("shipshape.extension", dyc.ConfigName)

	yc, ok := cfg.Checks[shipshape.Yaml][0].(*shipshape.YamlCheck)
	assert.True(ok)
	assert.Equal("shipshape.extension.yml", yc.File)

	yc2, ok := cfg.Checks[shipshape.Yaml][1].(*shipshape.YamlCheck)
	assert.True(ok)
	assert.Equal("shipshape.extension.yml", yc2.File)
	assert.True(*yc2.IgnoreMissing)

	rl := cfg.RunChecks()
	expectedRl := shipshape.ResultList{Results: []shipshape.Result{
		{
			Name:      "File check - Ignore missing",
			Severity:  "normal",
			CheckType: "yaml",
			Status:    "Pass",
			Passes:    []string{fmt.Sprintf("File %s/config/sync/shipshape.extension.yml does not exist", shipshape.ProjectDir)},
			Failures:  []string(nil),
		},
		{
			Name:      "My db check",
			Severity:  "normal",
			CheckType: "drush-yaml",
			Status:    "Fail",
			Passes:    []string(nil),
			Failures:  []string{fmt.Sprintf("%s/vendor/drush/drush/drush: no such file or directory", shipshape.ProjectDir)},
		},
		{
			Name:      "My file check",
			Severity:  "normal",
			CheckType: "yaml",
			Status:    "Fail",
			Passes:    []string(nil),
			Failures:  []string{fmt.Sprintf("open %s/config/sync/shipshape.extension.yml: no such file or directory", shipshape.ProjectDir)},
		},
	}}
	assert.ElementsMatch(expectedRl.Results, rl.Results)
}
