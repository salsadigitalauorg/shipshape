package drupal_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestFileModuleMerge(t *testing.T) {
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
		Required:   []string{"req2"},
		Disallowed: []string{"disallowed2"},
	}, c)
}

func TestFileModuleConfigName(t *testing.T) {
	assert := assert.New(t)

	configNameVal := ""
	origCheckModulesInYaml := drupal.CheckModulesInYaml
	mockCheckModulesInYaml := func(c *shipshape.YamlBase, ct shipshape.CheckType, configName string, required, disallowed []string) {
		configNameVal = configName
	}

	t.Run("noPath", func(t *testing.T) {
		c := drupal.FileModuleCheck{YamlCheck: shipshape.YamlCheck{File: "foo.bar"}}
		drupal.CheckModulesInYaml = mockCheckModulesInYaml
		defer func() {
			drupal.CheckModulesInYaml = origCheckModulesInYaml
		}()
		c.RunCheck(false)
		assert.Equal("foo.bar", configNameVal)
	})

	t.Run("pathWithoutSlash", func(t *testing.T) {
		c := drupal.FileModuleCheck{YamlCheck: shipshape.YamlCheck{
			File: "foo.bar",
			Path: "/some/path",
		}}
		drupal.CheckModulesInYaml = mockCheckModulesInYaml
		defer func() {
			drupal.CheckModulesInYaml = origCheckModulesInYaml
		}()
		c.RunCheck(false)
		assert.Equal("/some/path/foo.bar", configNameVal)
	})

	t.Run("pathWithSlash", func(t *testing.T) {
		c := drupal.FileModuleCheck{YamlCheck: shipshape.YamlCheck{
			File: "foo.bar",
			Path: "/some/path/",
		}}
		drupal.CheckModulesInYaml = mockCheckModulesInYaml
		defer func() {
			drupal.CheckModulesInYaml = origCheckModulesInYaml
		}()
		c.RunCheck(false)
		assert.Equal("/some/path/foo.bar", configNameVal)
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
	c.RunCheck(false)
	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.Empty(c.Result.Failures)
	assert.ElementsMatch(c.Result.Passes, []string{
		"'node' is enabled",
		"'block' is enabled",
		"'views_ui' is not enabled",
		"'field_ui' is not enabled",
	})
}
