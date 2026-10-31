package fact_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

func TestBaseFactCategoryAndDescription(t *testing.T) {
	t.Run("parse from yaml", func(t *testing.T) {
		var f fact.BaseFact
		err := yaml.Unmarshal([]byte("category: db-permissions\ndescription: Disallowed permissions on active site\n"), &f)
		assert.NoError(t, err)
		assert.Equal(t, "db-permissions", f.GetCategory())
		assert.Equal(t, "Disallowed permissions on active site", f.GetDescription())
	})

	t.Run("category defaults to fact id when absent", func(t *testing.T) {
		f := fact.BaseFact{BasePlugin: plugin.BasePlugin{Id: "db-permissions"}}
		assert.Equal(t, "db-permissions", f.GetCategory())
	})

	t.Run("description defaults to empty when absent", func(t *testing.T) {
		f := fact.BaseFact{BasePlugin: plugin.BasePlugin{Id: "db-permissions"}}
		assert.Equal(t, "", f.GetDescription())
	})
}
