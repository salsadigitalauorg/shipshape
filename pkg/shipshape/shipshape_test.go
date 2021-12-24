package shipshape_test

import (
	"salsadigitalauorg/shipshape/pkg/shipshape"
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	invalidData := `
checks:
  db-config: foo
`
	c, err := shipshape.ParseConfig([]byte(invalidData))
	if err == nil || !strings.Contains(err.Error(), "yaml: unmarshal errors") {
		t.Error("file parsing should fail")
	}

	data := `
drupal-root: web
checks:
  db-config:
    - name: My db check
      config-name: core.extension
  file-config:
    - name: My file check
      config-name: core.extension
      config-path: config/sync
  drush-command:
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

	if len(c.Checks.DbConfig) == 0 {
		t.Fatalf("DbConfig checks count should be 1, got %d", len(c.Checks.DbConfig))
	}

	if len(c.Checks.FileConfig) == 0 {
		t.Fatalf("FileConfig checks count should be 1, got %d", len(c.Checks.FileConfig))
	}

	if len(c.Checks.DrushCommand) == 0 {
		t.Fatalf("DrushCommand checks count should be 1, got %d", len(c.Checks.DrushCommand))
	}

	if c.Checks.DbConfig[0].ConfigName != "core.extension" {
		t.Fatalf("ActiveConfig check 1's config name should be core.extension, got %s", c.Checks.DbConfig[0].ConfigName)
	}

	if c.Checks.FileConfig[0].ConfigName != "core.extension" {
		t.Fatalf("FileConfig check 1's config name should be core.extension, got %s", c.Checks.FileConfig[0].ConfigName)
	}

}

func TestRunChecks(t *testing.T) {
	cfg := shipshape.Config{
		Checks: shipshape.CheckList{
			DbConfig: []shipshape.DbConfigCheck{
				{ConfigName: "core.extension"},
			},
			FileConfig: []shipshape.FileConfigCheck{
				{ConfigName: "core.extensions"},
			},
			DrushCommand: []shipshape.DrushCommandCheck{
				{Command: "status"},
			},
		},
	}
	cfg.RunChecks()
}
