package shipshape_test

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
)

func TestReadAndParseConfig(t *testing.T) {
	_, err := shipshape.ReadAndParseConfig("", "testdata/nonexistent.yml")
	if err == nil || err.Error() != "open testdata/nonexistent.yml: no such file or directory" {
		t.Errorf("file read should fail, got %#v", err.Error())
	}

	_, err = shipshape.ReadAndParseConfig("", "testdata/shipshape.yml")
	if err != nil {
		t.Errorf("file read should pass, got %#v", err.Error())
	}
}

func TestParseConfig(t *testing.T) {
	invalidData := `
checks:
  yaml: foo
`
	cfg := shipshape.Config{}
	err := shipshape.ParseConfig([]byte(invalidData), "", &cfg)
	if err == nil || !strings.Contains(err.Error(), "yaml: unmarshal errors") {
		t.Error("file parsing should fail")
	}

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
	if err != nil {
		t.Errorf("Failed to read check file config: %+v", err)
	}
	cfg.Init()

	currDir, err := os.Getwd()
	if err != nil {
		t.Fatal("Unable to get current dir.")
	}
	if cfg.ProjectDir != currDir {
		t.Errorf("Project dir should be '%s', got '%s'", currDir, cfg.ProjectDir)
	}

	if len(cfg.Checks[drupal.DrushYaml]) == 0 {
		t.Fatalf("DrushYaml checks count should be 1, got %d", len(cfg.Checks[drupal.DrushYaml]))
	}

	if len(cfg.Checks[shipshape.Yaml]) == 0 {
		t.Fatalf("YamlCheck checks count should be 1, got %d", len(cfg.Checks[shipshape.Yaml]))
	}

	dyc, ok := cfg.Checks[drupal.DrushYaml][0].(*drupal.DrushYamlCheck)
	if !ok || dyc.ConfigName != "shipshape.extension" {
		t.Fatalf("DrushYamlCheck check 1's config name should be shipshape.extension, got %s", dyc.ConfigName)
	}

	yc, ok := cfg.Checks[shipshape.Yaml][0].(*shipshape.YamlCheck)
	if !ok || yc.File != "shipshape.extension.yml" {
		t.Fatalf("YamlCheck check 1's config name should be shipshape.extension.yml, got %s", yc.File)
	}

	yc2, ok := cfg.Checks[shipshape.Yaml][1].(*shipshape.YamlCheck)
	if !ok || yc2.File != "shipshape.extension.yml" {
		t.Fatalf("YamlCheck check 2's config name should be shipshape.extension.yml, got %s", yc.File)
	}
	if yc2.IgnoreMissing != true {
		t.Fatalf("IgnoreMissing should be true, got %#v", yc2.IgnoreMissing)
	}

	rl := cfg.RunChecks([]string(nil))
	expectedRl := shipshape.ResultList{Results: []shipshape.Result{
		{
			Name:      "File check - Ignore missing",
			CheckType: "yaml",
			Status:    "Pass",
			Passes:    []string{fmt.Sprintf("File %s/config/sync/shipshape.extension.yml does not exist", shipshape.ProjectDir)},
			Failures:  []string(nil),
		},
		{
			Name:      "My db check",
			CheckType: "drush-yaml",
			Status:    "Fail",
			Passes:    []string(nil),
			Failures:  []string{fmt.Sprintf("%s/vendor/drush/drush/drush: no such file or directory", shipshape.ProjectDir)},
		},
		{
			Name:      "My file check",
			CheckType: "yaml",
			Status:    "Fail",
			Passes:    []string(nil),
			Failures:  []string{fmt.Sprintf("open %s/config/sync/shipshape.extension.yml: no such file or directory", shipshape.ProjectDir)},
		},
	}}
	if !reflect.DeepEqual(rl.Results, expectedRl.Results) {
		t.Errorf("Results are not as expected, got: %#v", rl)
	}
}

func TestRunChecks(t *testing.T) {
	cfg := shipshape.Config{
		Checks: map[shipshape.CheckType][]shipshape.Check{
			shipshape.File: {&shipshape.FileCheck{}},
			shipshape.Yaml: {&shipshape.YamlCheck{}},
		},
	}
	cfg.RunChecks([]string(nil))
}
