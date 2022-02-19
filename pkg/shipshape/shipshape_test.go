package shipshape_test

import (
	"os"
	"salsadigitalauorg/shipshape/pkg/core"
	"salsadigitalauorg/shipshape/pkg/drupal"
	"salsadigitalauorg/shipshape/pkg/shipshape"
	"strings"
	"testing"
)

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
      config-name: core.extension
  yaml:
    - name: My file check
      file: core.extension
      path: config/sync
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
		t.Fatalf("DbConfig checks count should be 1, got %d", len(cfg.Checks[drupal.DrushYaml]))
	}

	if len(cfg.Checks[core.Yaml]) == 0 {
		t.Fatalf("FileConfig checks count should be 1, got %d", len(cfg.Checks[core.Yaml]))
	}

	dyc, ok := cfg.Checks[drupal.DrushYaml][0].(*drupal.DrushYamlCheck)
	if !ok || dyc.ConfigName != "core.extension" {
		t.Fatalf("DbConfig check 1's config name should be core.extension, got %s", dyc.ConfigName)
	}

	yc, ok := cfg.Checks[core.Yaml][0].(*core.YamlCheck)
	if !ok || yc.File != "core.extension" {
		t.Fatalf("FileConfig check 1's config name should be core.extension, got %s", yc.File)
	}

}

func TestRunChecks(t *testing.T) {
	cfg := shipshape.Config{
		Checks: map[core.CheckType][]core.Check{},
	}
	cfg.RunChecks()
}
