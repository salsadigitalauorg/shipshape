package drupal_test

import (
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/checks/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/stretchr/testify/assert"
)

func TestFileModuleMerge(t *testing.T) {
	assert := assert.New(t)

	c := FileModuleCheck{
		YamlCheck: yaml.YamlCheck{
			Path:           "path1",
			File:           "file1.yml",
			Pattern:        "pattern1",
			ExcludePattern: "excludePattern1",
		},
		Required:   []string{"req1"},
		Disallowed: []string{"disallowed1"},
	}
	c.Merge(&FileModuleCheck{
		YamlCheck: yaml.YamlCheck{
			Path:  "path2",
			Files: []string{"slcFile1.yml", "slcFile2.yml"},
		},
		Required:   []string{"req2"},
		Disallowed: []string{"disallowed2"},
	})
	assert.EqualValues(FileModuleCheck{
		YamlCheck: yaml.YamlCheck{
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

func TestFileModuleCheck(t *testing.T) {
	assert := assert.New(t)

	c := FileModuleCheck{
		YamlCheck: yaml.YamlCheck{
			YamlBase: mockCheck("core.extension.yml"),
		},
		Required:   []string{"node", "block"},
		Disallowed: []string{"views_ui", "field_ui"},
	}
	c.Init(FileModule)
	assert.Equal("core.extension.yml", c.File)
	assert.True(*c.IgnoreMissing)

	c.UnmarshalDataMap()
	c.RunCheck()
	assert.Equal(result.Pass, c.Result.Status)
	assert.Empty(c.Result.Failures)
	assert.ElementsMatch(c.Result.Passes, []string{
		"all required modules are enabled",
		"all disallowed modules are disabled",
	})
}
