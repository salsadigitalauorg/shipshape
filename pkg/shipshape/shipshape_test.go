package shipshape_test

import (
	"os"
	"salsadigitalauorg/shipshape/pkg/shipshape"
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	invalidData := `
checks:
  drupal-db-config: foo
`
	cfg := shipshape.Config{}
	err := shipshape.ParseConfig([]byte(invalidData), "", &cfg)
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

	if cfg.DrupalRoot != "web" {
		t.Errorf("drupal root should be 'web', got %s", cfg.DrupalRoot)
	}

	if len(cfg.Checks[shipshape.DrupalDBConfig]) == 0 {
		t.Fatalf("DbConfig checks count should be 1, got %d", len(cfg.Checks[shipshape.DrupalDBConfig]))
	}

	if len(cfg.Checks[shipshape.DrupalFileConfig]) == 0 {
		t.Fatalf("FileConfig checks count should be 1, got %d", len(cfg.Checks[shipshape.DrupalFileConfig]))
	}

	ddc, ok := cfg.Checks[shipshape.DrupalDBConfig][0].(*shipshape.DrupalDBConfigCheck)
	if !ok || ddc.ConfigName != "core.extension" {
		t.Fatalf("DbConfig check 1's config name should be core.extension, got %s", ddc.ConfigName)
	}

	dfc, ok := cfg.Checks[shipshape.DrupalFileConfig][0].(*shipshape.DrupalFileConfigCheck)
	if !ok || dfc.ConfigName != "core.extension" {
		t.Fatalf("FileConfig check 1's config name should be core.extension, got %s", dfc.ConfigName)
	}

}

func TestRunChecks(t *testing.T) {
	cfg := shipshape.Config{
		DrupalRoot: "",
		Checks: map[shipshape.CheckType][]shipshape.Check{
			shipshape.DrupalDBConfig: {
				&shipshape.DrupalDBConfigCheck{
					DrupalConfigBase: shipshape.DrupalConfigBase{},
					Drush:            shipshape.Drush{},
				},
			},
			shipshape.DrupalFileConfig: {
				&shipshape.DrupalFileConfigCheck{
					DrupalConfigBase: shipshape.DrupalConfigBase{},
					ConfigPath:       "",
				},
			},
		},
	}
	cfg.RunChecks()
}
