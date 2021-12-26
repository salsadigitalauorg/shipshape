package shipshape_test

import (
	"salsadigitalauorg/shipshape/pkg/shipshape"
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	invalidData := `
checks:
  drupal-db-config: foo
`
	c, err := shipshape.ParseConfig([]byte(invalidData))
	if err == nil || !strings.Contains(err.Error(), "yaml: unmarshal errors") {
		t.Error("file parsing should fail")
	}

	data := `
drupal-root: web
checks:
  drupal-db-config:
    - name: My db check
      config-name: core.extension
  drupal-file-config:
    - name: My file check
      config-name: core.extension
      config-path: config/sync
  drush:
    - name: My drush command check
      command: pm:list --status=enabled
`
	c, err = shipshape.ParseConfig([]byte(data))
	if err != nil {
		t.Error("Failed to read check file config")
	}

	if c.DrupalRoot != "web" {
		t.Errorf("drupal root should be 'web', got %s", c.DrupalRoot)
	}

	if len(c.Checks.DrupalDBConfig) == 0 {
		t.Fatalf("DbConfig checks count should be 1, got %d", len(c.Checks.DrupalDBConfig))
	}

	if len(c.Checks.DrupalFileConfig) == 0 {
		t.Fatalf("FileConfig checks count should be 1, got %d", len(c.Checks.DrupalFileConfig))
	}

	if len(c.Checks.Drush) == 0 {
		t.Fatalf("Drush checks count should be 1, got %d", len(c.Checks.Drush))
	}

	if c.Checks.DrupalDBConfig[0].ConfigName != "core.extension" {
		t.Fatalf("DbConfig check 1's config name should be core.extension, got %s", c.Checks.DrupalDBConfig[0].ConfigName)
	}

	if c.Checks.DrupalFileConfig[0].ConfigName != "core.extension" {
		t.Fatalf("FileConfig check 1's config name should be core.extension, got %s", c.Checks.DrupalFileConfig[0].ConfigName)
	}

}

func TestRunChecks(t *testing.T) {
	cfg := shipshape.Config{
		Checks: shipshape.CheckList{
			DrupalDBConfig:   []shipshape.DrupalDBConfigCheck{},
			DrupalFileConfig: []shipshape.DrupalFileConfigCheck{},
			Drush:            []shipshape.DrushCheck{},
		},
	}
	cfg.RunChecks()
}
