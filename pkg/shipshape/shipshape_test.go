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
`
	c, err = shipshape.ParseConfig([]byte(data))
	if err != nil {
		t.Errorf("Failed to read check file config: %+v", err)
	}

	if c.DrupalRoot != "web" {
		t.Errorf("drupal root should be 'web', got %s", c.DrupalRoot)
	}

	if len(c.Checks[shipshape.DrupalDBConfig]) == 0 {
		t.Fatalf("DbConfig checks count should be 1, got %d", len(c.Checks[shipshape.DrupalDBConfig]))
	}

	if len(c.Checks[shipshape.DrupalFileConfig]) == 0 {
		t.Fatalf("FileConfig checks count should be 1, got %d", len(c.Checks[shipshape.DrupalFileConfig]))
	}

	ddc, ok := c.Checks[shipshape.DrupalDBConfig][0].(*shipshape.DrupalDBConfigCheck)
	if !ok || ddc.ConfigName != "core.extension" {
		t.Fatalf("DbConfig check 1's config name should be core.extension, got %s", ddc.ConfigName)
	}

	dfc, ok := c.Checks[shipshape.DrupalFileConfig][0].(*shipshape.DrupalFileConfigCheck)
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
