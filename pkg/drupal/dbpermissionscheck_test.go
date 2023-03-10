package drupal_test

import (
	"os/exec"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	. "github.com/salsadigitalauorg/shipshape/pkg/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/yaml"

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
		assert.Equal(config.Fail, c.Result.Status)
		assert.Empty(c.Result.Passes)
		assert.ElementsMatch([]string{"no data provided"}, c.Result.Failures)
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
		assert.NotEqual(config.Fail, c.Result.Status)
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
			ExpectStatus: config.Fail,
			ExpectNoPass: true,
			ExpectFails:  []string{"list of disallowed perms not provided"},
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
			ExpectStatus: config.Pass,
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
			ExpectStatus: config.Fail,
			ExpectPasses: []string{
				"[anonymous] no disallowed permissions",
				"[authenticated] no disallowed permissions",
			},
			ExpectFails: []string{
				"[site_administrator] disallowed permissions: [administer modules, administer permissions]",
				"[site_editor] disallowed permissions: [administer modules]",
			},
		},
		{
			Name: "breachRemediation",
			Check: &DbPermissionsCheck{
				DrushYamlCheck: DrushYamlCheck{
					YamlBase: yaml.YamlBase{
						CheckBase: config.CheckBase{
							PerformRemediation: true,
						},
					},
				},
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
			Init: true,
			PreRun: func(t *testing.T) {
				command.ShellCommander = internal.ShellCommanderMaker(nil, nil, nil)
			},
			Sort:         true,
			ExpectStatus: config.Pass,
			ExpectPasses: []string{
				"[anonymous] no disallowed permissions",
				"[authenticated] no disallowed permissions",
			},
			ExpectNoFail: true,
			ExpectRemediations: []string{
				"[site_administrator] fixed disallowed permissions: [administer modules, administer permissions]",
				"[site_editor] fixed disallowed permissions: [administer modules]",
			},
		},
		{
			Name: "breachRemediationExitError",
			Check: &DbPermissionsCheck{
				DrushYamlCheck: DrushYamlCheck{
					YamlBase: yaml.YamlBase{
						CheckBase: config.CheckBase{
							PerformRemediation: true,
						},
					},
				},
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
			Init: true,
			PreRun: func(t *testing.T) {
				command.ShellCommander = internal.ShellCommanderMaker(
					nil,
					&exec.ExitError{Stderr: []byte("unable to run drush command")},
					nil,
				)
			},
			Sort:         true,
			ExpectStatus: config.Fail,
			ExpectPasses: []string{
				"[anonymous] no disallowed permissions",
				"[authenticated] no disallowed permissions",
			},
			ExpectFails: []string{
				"[site_administrator] failed to fix disallowed permissions [administer modules, administer permissions] due to error: unable to run drush command",
				"[site_editor] failed to fix disallowed permissions [administer modules] due to error: unable to run drush command",
			},
			ExpectNoRemediations: true,
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
		err := c.Remediate(DbPermissionsBreach{Role: "foo", Perms: "bar,baz"})
		assert.Error(err, "unable to run drush command")
	})

	t.Run("drushCommandIsCorrect", func(t *testing.T) {
		var generatedCommand string
		command.ShellCommander = internal.ShellCommanderMaker(nil, nil, &generatedCommand)

		c := DbPermissionsCheck{}
		c.Remediate(DbPermissionsBreach{Role: "foo", Perms: "bar,baz"})
		assert.Equal("vendor/drush/drush/drush role:perm:remove foo bar,baz", generatedCommand)
	})
}
