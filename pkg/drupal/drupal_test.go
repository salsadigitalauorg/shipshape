package drupal_test

import (
	"reflect"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestRegisterChecks(t *testing.T) {
	checksMap := map[shipshape.CheckType]string{
		drupal.DrushYaml:     "*drupal.DrushYamlCheck",
		drupal.FileModule:    "*drupal.FileModuleCheck",
		drupal.DbModule:      "*drupal.DbModuleCheck",
		drupal.DbPermissions: "*drupal.DbPermissionsCheck",
		drupal.TrackingCode:  "*drupal.TrackingCodeCheck",
		drupal.UserRole:      "*drupal.UserRoleCheck",
	}
	for ct, ts := range checksMap {
		c := shipshape.ChecksRegistry[ct]()
		ctype := reflect.TypeOf(c).String()
		assert.Equal(t, ts, ctype)
	}
}

func mockCheck(configName string) shipshape.YamlBase {
	return shipshape.YamlBase{
		CheckBase: shipshape.CheckBase{
			DataMap: map[string][]byte{
				configName: []byte(`
module:
  block: 0
  node: 0

`),
			},
		},
	}
}

func TestCheckKeyValueExists(t *testing.T) {
	c := mockCheck("shipshape.extension.yml")
	c.UnmarshalDataMap()

	m := shipshape.KeyValue{
		Key:   "module.block",
		Value: "0",
	}
	kvr, _, _ := c.CheckKeyValue(m, "shipshape.extension.yml")
	assert.Equal(t, kvr, shipshape.KeyValueEqual)
}
