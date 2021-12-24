package shipshape_test

import (
	"salsadigitalauorg/shipshape/pkg/shipshape"
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	invalidData := `
checks:
  active-config: foo
`
	c, err := shipshape.ParseConfig([]byte(invalidData))
	if err == nil || !strings.Contains(err.Error(), "yaml: unmarshal errors") {
		t.Error("file parsing should fail")
	}

	data := `
drupal-root: web
checks:
  active-config:
    - config-name: core.extension
  file-config:
    - config-name: core.extension
      config-path: config/sync
`
	c, err = shipshape.ParseConfig([]byte(data))
	if err != nil {
		t.Error("Failed to read check file config")
	}

	if c.DrupalRoot != "web" {
		t.Errorf("drupal root should be 'web', got %s", c.DrupalRoot)
	}

	if len(c.Checks.ActiveConfig) == 0 {
		t.Fatalf("ActiveConfig checks count should be 1, got %d", len(c.Checks.ActiveConfig))
	}

	if len(c.Checks.FileConfig) == 0 {
		t.Fatalf("FileConfig checks count should be 1, got %d", len(c.Checks.FileConfig))
	}

	if c.Checks.ActiveConfig[0].ConfigName != "core.extension" {
		t.Fatalf("ActiveConfig check 1's config name should be core.extension, got %s", c.Checks.ActiveConfig[0].ConfigName)
	}

	if c.Checks.FileConfig[0].ConfigName != "core.extension" {
		t.Fatalf("FileConfig check 1's config name should be core.extension, got %s", c.Checks.FileConfig[0].ConfigName)
	}

}
