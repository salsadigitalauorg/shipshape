package drupal_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestDbPermissionsMerge(t *testing.T) {
	assert := assert.New(t)

	c := drupal.DbPermissionsCheck{
		DrushYamlCheck: drupal.DrushYamlCheck{
			YamlBase: shipshape.YamlBase{
				Values: []shipshape.KeyValue{
					{Key: "key1", Value: "val1", Optional: false},
				},
			},
		},
		Disallowed:   []string{"disallowed1"},
		ExcludeRoles: []string{"role1"},
	}
	c.Merge(&drupal.DbPermissionsCheck{
		DrushYamlCheck: drupal.DrushYamlCheck{
			YamlBase: shipshape.YamlBase{
				Values: []shipshape.KeyValue{
					{Key: "key1", Value: "val1", Optional: true},
				},
			},
		},
		Disallowed:   []string{"disallowed2"},
		ExcludeRoles: []string{"role2"},
	})
	assert.EqualValues(drupal.DbPermissionsCheck{
		DrushYamlCheck: drupal.DrushYamlCheck{
			YamlBase: shipshape.YamlBase{
				Values: []shipshape.KeyValue{
					{Key: "key1", Value: "val1", Optional: true},
				},
			},
		},
		Disallowed:   []string{"disallowed2"},
		ExcludeRoles: []string{"role2"},
	}, c)
}

func TestDbPermissionsCheck(t *testing.T) {
	assert := assert.New(t)

	// Test init.
	c := drupal.DbPermissionsCheck{}
	c.Init(drupal.DbPermissions)
	assert.Equal("role:list", c.Command)
	assert.Equal("permissions", c.ConfigName)

	c.UnmarshalDataMap()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.Empty(c.Result.Passes)
	assert.ElementsMatch([]string{"no data provided"}, c.Result.Failures)

	// Test UnmarshalDataMap.
	c = drupal.DbPermissionsCheck{}
	c.Init(drupal.DbPermissions)
	c.DataMap = map[string][]byte{
		"permissions": []byte(`
site_administrator:
  label: 'Site Administrator'
  perms: {  }
anonymous:
  label: 'Anonymous user'
  perms:
    - 'access content'
    - 'view media'
authenticated:
  label: 'Authenticated user'
  perms:
    - 'access content'
    - 'view media'
site_editor:
  label: 'Site Editor'
  perms: {  }
`),
	}
	c.UnmarshalDataMap()
	assert.NotEqual(shipshape.Fail, c.Result.Status)
	assert.EqualValues(map[string]drupal.DrushRole{
		"anonymous": {
			Label: "Anonymous user",
			Perms: []string{"access content", "view media"},
		},
		"authenticated": {
			Label: "Authenticated user",
			Perms: []string{"access content", "view media"},
		},
		"site_administrator": {
			Label: "Site Administrator",
			Perms: []string(nil),
		},
		"site_editor": {
			Label: "Site Editor",
			Perms: []string(nil),
		},
	}, c.Permissions)

	// Test RunCheck.
	c = drupal.DbPermissionsCheck{}
	c.Init(drupal.DbPermissions)
	c.RunCheck(false)
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.Empty(c.Result.Passes)
	assert.ElementsMatch(
		[]string{"list of disallowed perms not provided"},
		c.Result.Failures,
	)

	c = drupal.DbPermissionsCheck{}
	c.Init(drupal.DbPermissions)
	c.Permissions = map[string]drupal.DrushRole{
		"anonymous": {
			Label: "Anonymous user",
			Perms: []string{"access content", "view media"},
		},
		"authenticated": {
			Label: "Authenticated user",
			Perms: []string{"access content", "view media"},
		},
		"site_administrator": {
			Label: "Site Administrator",
			Perms: []string(nil),
		},
		"site_editor": {
			Label: "Site Editor",
			Perms: []string(nil),
		},
	}
	c.Disallowed = []string{"administer modules"}
	c.RunCheck(false)
	c.Result.Sort()
	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.Empty(c.Result.Failures)
	assert.ElementsMatch([]string{
		"[anonymous] no disallowed permissions",
		"[authenticated] no disallowed permissions",
		"[site_administrator] no disallowed permissions",
		"[site_editor] no disallowed permissions",
	}, c.Result.Passes)

	c = drupal.DbPermissionsCheck{}
	c.Init(drupal.DbPermissions)
	c.Permissions = map[string]drupal.DrushRole{
		"anonymous": {
			Label: "Anonymous user",
			Perms: []string{"access content", "view media"},
		},
		"authenticated": {
			Label: "Authenticated user",
			Perms: []string{"access content", "view media"},
		},
		"site_administrator": {
			Label: "Site Administrator",
			Perms: []string{"administer modules", "administer permissions"},
		},
		"another_site_administrator": {
			Label: "Site Administrator",
			Perms: []string{"administer modules", "administer permissions"},
		},
		"site_editor": {
			Label: "Site Editor",
			Perms: []string{"administer modules"},
		},
	}
	c.Disallowed = []string{"administer modules", "administer permissions"}
	c.ExcludeRoles = []string{"another_site_administrator"}
	c.RunCheck(false)
	c.Result.Sort()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.ElementsMatch([]string{
		"[anonymous] no disallowed permissions",
		"[authenticated] no disallowed permissions",
	}, c.Result.Passes)
	assert.ElementsMatch([]string{
		"[site_administrator] disallowed permissions: [administer modules, administer permissions]",
		"[site_editor] disallowed permissions: [administer modules]",
	}, c.Result.Failures)
}
