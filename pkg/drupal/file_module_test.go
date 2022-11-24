package drupal_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestMerge(t *testing.T) {
	assert := assert.New(t)

	c := drupal.FileModuleCheck{
		YamlCheck: shipshape.YamlCheck{
			Path:           "path1",
			File:           "file1.yml",
			Pattern:        "pattern1",
			ExcludePattern: "excludePattern1",
		},
		Required:   []string{"req1"},
		Disallowed: []string{"disallowed1"},
	}
	c.Merge(&drupal.FileModuleCheck{
		YamlCheck: shipshape.YamlCheck{
			Path:  "path2",
			Files: []string{"slcFile1.yml", "slcFile2.yml"},
		},
		Required:   []string{"req2"},
		Disallowed: []string{"disallowed2"},
	})
	assert.EqualValues(drupal.FileModuleCheck{
		YamlCheck: shipshape.YamlCheck{
			Path:           "path2",
			File:           "file1.yml",
			Files:          []string{"slcFile1.yml", "slcFile2.yml"},
			Pattern:        "pattern1",
			ExcludePattern: "excludePattern1",
		},
		Required:   []string{"req1", "req2"},
		Disallowed: []string{"disallowed1", "disallowed2"},
	}, c)
}

func TestDisallowedIsEnabled(t *testing.T) {
	assert := assert.New(t)

	c := mockCheck("shipshape.extension.yml")
	c.DataMap = map[string][]byte{
		"shipshape.extension.yml": []byte(`
module:
  clamav: 0
  tfa: 0
  dblog: 0

`),
	}
	c.UnmarshalDataMap()

	required := []string{
		"clamav",
		"tfa",
	}
	disallowed := []string{
		"dblog",
		"module_permissions_ui",
		"update",
	}

	drupal.CheckModulesInYaml(&c, drupal.FileModule, "shipshape.extension.yml", required, disallowed)
	assert.ElementsMatch(c.Result.Passes, []string{
		"'clamav' is enabled",
		"'tfa' is enabled",
		"'module_permissions_ui' is not enabled",
		"'update' is not enabled",
	})
	assert.ElementsMatch(c.Result.Failures, []string{"'dblog' is enabled"})
}

func TestCheckModulesInYaml(t *testing.T) {
	assert := assert.New(t)

	// Invalid yaml key.
	c := mockCheck("shipshape.extension.yml")
	required := []string{
		"node&foo",
		"block",
	}
	disallowed := []string{
		"views_ui",
		"field_ui&bar",
	}
	c.UnmarshalDataMap()
	drupal.CheckModulesInYaml(&c, drupal.FileModule, "shipshape.extension.yml", required, disallowed)
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.ElementsMatch(c.Result.Passes, []string{
		"'block' is enabled",
		"'views_ui' is not enabled",
	})
	assert.ElementsMatch(c.Result.Failures, []string{
		"invalid character '&' at position 11, following \".node\"",
		"invalid character '&' at position 15, following \".field_ui\"",
	})

	// Required is not enabled & disallowed is enabled.
	c = mockCheck("shipshape.extension.yml")
	c.DataMap = map[string][]byte{
		"shipshape.extension.yml": []byte(`
module:
  block: 0
  views_ui: 0

`),
	}
	required = []string{
		"node",
		"block",
	}
	disallowed = []string{
		"views_ui",
		"field_ui",
	}
	c.UnmarshalDataMap()
	drupal.CheckModulesInYaml(&c, drupal.FileModule, "shipshape.extension.yml", required, disallowed)
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.ElementsMatch(c.Result.Passes, []string{
		"'block' is enabled",
		"'field_ui' is not enabled",
	})
	assert.ElementsMatch(c.Result.Failures, []string{
		"'node' is not enabled",
		"'views_ui' is enabled",
	})

	c = mockCheck("shipshape.extension.yml")
	required = []string{
		"node",
		"block",
	}
	disallowed = []string{
		"views_ui",
		"field_ui",
	}
	c.UnmarshalDataMap()
	drupal.CheckModulesInYaml(&c, drupal.FileModule, "shipshape.extension.yml", required, disallowed)
	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.Empty(c.Result.Failures)
	assert.ElementsMatch(c.Result.Passes, []string{
		"'node' is enabled",
		"'block' is enabled",
		"'views_ui' is not enabled",
		"'field_ui' is not enabled",
	})
}

func TestFileModuleCheck(t *testing.T) {
	assert := assert.New(t)

	c := drupal.FileModuleCheck{
		YamlCheck: shipshape.YamlCheck{
			YamlBase: mockCheck("core.extension.yml"),
		},
		Required:   []string{"node", "block"},
		Disallowed: []string{"views_ui", "field_ui"},
	}
	c.Init(drupal.FileModule)
	assert.Equal("core.extension.yml", c.File)
	assert.True(*c.IgnoreMissing)

	c.UnmarshalDataMap()
	c.RunCheck()
	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.Empty(c.Result.Failures)
	assert.ElementsMatch(c.Result.Passes, []string{
		"'node' is enabled",
		"'block' is enabled",
		"'views_ui' is not enabled",
		"'field_ui' is not enabled",
	})
}
