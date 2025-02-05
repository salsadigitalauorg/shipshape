package drupal_test

import (
	"os/exec"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	. "github.com/salsadigitalauorg/shipshape/pkg/checks/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/result"

	"github.com/stretchr/testify/assert"
)

func TestDbPermissionsInit(t *testing.T) {
	assert := assert.New(t)

	c := DbPermissionsCheck{}
	c.Init(DbPermissions)
	assert.Equal("role:list", c.Command)
	assert.Equal("permissions", c.ConfigName)
}

func TestDbPermissionsMerge(t *testing.T) {
	assert := assert.New(t)

	c := DbPermissionsCheck{
		DrushYamlCheck: DrushYamlCheck{
			YamlBase: yaml.YamlBase{
				Values: []yaml.KeyValue{
					{Key: "key1", Value: "val1", Optional: false},
				},
			},
		},
		Disallowed:   []string{"disallowed1"},
		ExcludeRoles: []string{"role1"},
	}
	c.Merge(&DbPermissionsCheck{
		DrushYamlCheck: DrushYamlCheck{
			YamlBase: yaml.YamlBase{
				Values: []yaml.KeyValue{
					{Key: "key1", Value: "val1", Optional: true},
				},
			},
		},
		Disallowed:   []string{"disallowed2"},
		ExcludeRoles: []string{"role2"},
	})
	assert.EqualValues(DbPermissionsCheck{
		DrushYamlCheck: DrushYamlCheck{
			YamlBase: yaml.YamlBase{
				Values: []yaml.KeyValue{
					{Key: "key1", Value: "val1", Optional: true},
				},
			},
		},
		Disallowed:   []string{"disallowed2"},
		ExcludeRoles: []string{"role2"},
	}, c)
}

func TestDbPermissionsUnmarshalDataMap(t *testing.T) {
	assert := assert.New(t)

	t.Run("noDataProvided", func(t *testing.T) {
		c := DbPermissionsCheck{}
		c.UnmarshalDataMap()
		c.Result.DetermineResultStatus(false)
		assert.Equal(result.Fail, c.Result.Status)
		assert.Empty(c.Result.Passes)
		assert.EqualValues(
			[]breach.Breach{&breach.ValueBreach{
				BreachType: "value",
				Value:      "no data provided",
			}},
			c.Result.Breaches,
		)
	})

	t.Run("validData", func(t *testing.T) {
		c := DbPermissionsCheck{}
		c.Init(DbPermissions)
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
		assert.NotEqual(result.Fail, c.Result.Status)
		assert.EqualValues(map[string]DrushRole{
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
	})
}

func TestDbPermissionsRunCheck(t *testing.T) {
	tests := []internal.RunCheckTest{
		{
			Name:         "disallowedNotProvided",
			Check:        &DbPermissionsCheck{},
			Init:         true,
			ExpectStatus: result.Fail,
			ExpectNoPass: true,
			ExpectFails: []breach.Breach{&breach.ValueBreach{
				BreachType: "value",
				Severity:   "normal",
				Value:      "list of disallowed perms not provided",
			}},
		},
		{
			Name: "noBreaches",
			Check: &DbPermissionsCheck{
				Permissions: map[string]DrushRole{
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
				},
				Disallowed: []string{"administer modules"},
			},
			Init:         true,
			Sort:         true,
			ExpectStatus: result.Pass,
			ExpectPasses: []string{
				"[anonymous] no disallowed permissions",
				"[authenticated] no disallowed permissions",
				"[site_administrator] no disallowed permissions",
				"[site_editor] no disallowed permissions",
			},
			ExpectNoFail: true,
		},
		{
			Name: "hasSomeBreaches",
			Check: &DbPermissionsCheck{
				Permissions: map[string]DrushRole{
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
				},
				Disallowed:   []string{"administer modules", "administer permissions"},
				ExcludeRoles: []string{"another_site_administrator"},
			},
			Init:         true,
			Sort:         true,
			ExpectStatus: result.Fail,
			ExpectPasses: []string{
				"[anonymous] no disallowed permissions",
				"[authenticated] no disallowed permissions",
			},
			ExpectFails: []breach.Breach{
				&breach.KeyValuesBreach{
					BreachType: "key-values",
					Severity:   "normal",
					KeyLabel:   "role",
					Key:        "site_administrator",
					ValueLabel: "permissions",
					Values:     []string{"administer modules", "administer permissions"},
				},
				&breach.KeyValuesBreach{
					BreachType: "key-values",
					Severity:   "normal",
					KeyLabel:   "role",
					Key:        "site_editor",
					ValueLabel: "permissions",
					Values:     []string{"administer modules"},
				},
			},
		},
	}

	curShellCommander := command.ShellCommander
	defer func() { command.ShellCommander = curShellCommander }()

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			internal.TestRunCheck(t, test)
		})
	}
}

func TestDbPermissionsRemediate(t *testing.T) {
	assert := assert.New(t)

	curShellCommander := command.ShellCommander
	defer func() { command.ShellCommander = curShellCommander }()

	t.Run("drushError", func(t *testing.T) {
		command.ShellCommander = internal.ShellCommanderMaker(
			nil,
			&exec.ExitError{Stderr: []byte("unable to run drush command")},
			nil)

		c := DbPermissionsCheck{}
		c.AddBreach(&breach.KeyValuesBreach{
			BreachType: "key-values",
			KeyLabel:   "role",
			Key:        "foo",
			ValueLabel: "permissions",
			Values:     []string{"bar", "baz"},
		})
		c.Remediate()
		assert.EqualValues([]breach.Breach{&breach.KeyValuesBreach{
			BreachType: "key-values",
			KeyLabel:   "role",
			Key:        "foo",
			ValueLabel: "permissions",
			Values:     []string{"bar", "baz"},
			Remediation: breach.Remediation{
				Status:   breach.RemediationStatusFailed,
				Messages: []string{"failed to fix disallowed permissions for role 'foo' due to error: <nil>: unable to run drush command"},
			},
		}}, c.Result.Breaches)
	})

	t.Run("drushCommandIsCorrect", func(t *testing.T) {
		var generatedCommand string
		command.ShellCommander = internal.ShellCommanderMaker(nil, nil, &generatedCommand)

		c := DbPermissionsCheck{}
		c.AddBreach(&breach.KeyValuesBreach{
			KeyLabel:   "role",
			Key:        "foo",
			ValueLabel: "permissions",
			Values:     []string{"bar", "baz"},
		})
		c.Remediate()
		assert.Equal("vendor/drush/drush/drush role:perm:remove foo bar,baz", generatedCommand)
	})
}
