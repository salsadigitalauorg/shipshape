package drupal_test

import (
	"fmt"
	"os/exec"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	. "github.com/salsadigitalauorg/shipshape/pkg/checks/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

func TestAdminUserInit(t *testing.T) {
	c := AdminUserCheck{}
	c.Init(AdminUser)
	assert.True(t, c.RequiresDb)
}

func TestAdminUserMerge(t *testing.T) {
	assert := assert.New(t)

	c := AdminUserCheck{
		DrushCommand: DrushCommand{
			DrushPath: "/path/to/drush",
		},
		AllowedRoles: []string{"role1"},
	}
	c.Merge(&AdminUserCheck{
		DrushCommand: DrushCommand{
			DrushPath: "/new/path/to/drush",
		},
		AllowedRoles: []string{"role2", "role3"},
	})
	assert.EqualValues(AdminUserCheck{
		DrushCommand: DrushCommand{
			DrushPath: "/new/path/to/drush",
		},
		AllowedRoles: []string{"role2", "role3"},
	}, c)
}

func TestAdminUserFetchData(t *testing.T) {
	assert := assert.New(t)

	t.Run("drushNotFound", func(t *testing.T) {
		c := AdminUserCheck{}
		c.FetchData()
		assert.EqualValues(
			[]breach.Breach{&breach.ValueBreach{
				BreachType: "value",
				Value:      "vendor/drush/drush/drush: no such file or directory",
			}},
			c.Result.Breaches,
		)
	})

	curShellCommander := command.ShellCommander
	defer func() { command.ShellCommander = curShellCommander }()
	t.Run("drushError", func(t *testing.T) {
		command.ShellCommander = internal.ShellCommanderMaker(
			nil,
			&exec.ExitError{Stderr: []byte("unable to run drush command")},
			nil,
		)
		c := AdminUserCheck{}
		c.FetchData()
		assert.EqualValues(
			[]breach.Breach{&breach.ValueBreach{
				BreachType: "value",
				Value:      "unable to run drush command",
			}},
			c.Result.Breaches,
		)
	})

	// correct data.
	t.Run("correctData", func(t *testing.T) {
		expectedStdout := `{"anonymous":{"is_admin": false}}`

		curShellCommander := command.ShellCommander
		defer func() { command.ShellCommander = curShellCommander }()
		command.ShellCommander = internal.ShellCommanderMaker(&expectedStdout, nil, nil)

		c := AdminUserCheck{}
		c.FetchData()
		assert.NotEqual(result.Fail, c.Result.Status)
		assert.NotEqual(result.Pass, c.Result.Status)
		assert.Equal([]byte(`{"anonymous":{"is_admin": false}}`), c.DataMap["anonymous"])
	})
}

func TestAdminUserUnmarshalData(t *testing.T) {
	assert := assert.New(t)
	c := AdminUserCheck{}

	// Empty datamap.
	t.Run("emptyDataMap", func(t *testing.T) {
		c.UnmarshalDataMap()
		assert.EqualValues(
			[]breach.Breach{&breach.ValueBreach{
				BreachType: "value",
				Value:      "no data provided",
			}},
			c.Result.Breaches,
		)
	})

	// Incorrect json.
	t.Run("incorrectJSON", func(t *testing.T) {
		c = AdminUserCheck{
			CheckBase: config.CheckBase{
				DataMap: map[string][]byte{
					"anonymous": []byte(`{"is_admin":false, "id": "anonymous"]}`)},
			},
		}
		c.UnmarshalDataMap()
		assert.EqualValues(
			[]breach.Breach{&breach.ValueBreach{
				BreachType: "value",
				Value:      "invalid character ']' after object key:value pair",
			}},
			c.Result.Breaches,
		)
	})

	// Correct json.
	t.Run("correctJSON", func(t *testing.T) {
		c = AdminUserCheck{
			CheckBase: config.CheckBase{
				DataMap: map[string][]byte{
					"anonymous": []byte(`{"is_admin":false, "id": "anonymous"}`)},
			},
		}

		c.UnmarshalDataMap()
		assert.NotEqual(result.Fail, c.Result.Status)
		assert.NotEqual(result.Pass, c.Result.Status)
		roleConfigsVal := reflect.ValueOf(c).FieldByName("roleConfigs")
		assert.Equal("map[string]bool{\"anonymous\":false}", fmt.Sprintf("%#v", roleConfigsVal))
	})
}

func TestAdminUserRunCheck(t *testing.T) {
	tests := []internal.RunCheckTest{
		// Role does not have is_admin:true.
		{
			Name: "roleNotAdmin",
			Check: &AdminUserCheck{
				CheckBase: config.CheckBase{
					DataMap: map[string][]byte{
						"anonymous": []byte(`{"is_admin":false, "id": "anonymous"}`)},
				},
				AllowedRoles: []string{"authenticated", "content-admin"},
			},
			ExpectStatus: result.Pass,
		},

		// Role has is_admin:true.
		{
			Name: "roleAdmin",
			Check: &AdminUserCheck{
				CheckBase: config.CheckBase{
					DataMap: map[string][]byte{
						"anonymous": []byte(`{"is_admin":true, "id": "anonymous"}`)},
				},
				AllowedRoles: []string{"content-admin"},
			},
			ExpectStatus: result.Fail,
			ExpectFails: []breach.Breach{&breach.KeyValueBreach{
				BreachType: "key-value",
				Key:        "is_admin: true",
				ValueLabel: "role",
				Value:      "anonymous",
			}},
		},

		// Role has is_admin:true but is allowed.
		{
			Name: "roleAdminAllowed",
			Check: &AdminUserCheck{
				CheckBase: config.CheckBase{
					DataMap: map[string][]byte{
						"anonymous": []byte(`{"is_admin":true, "id": "anonymous"}`)},
				},
				AllowedRoles: []string{"anonymous", "content-admin"},
			},
			ExpectStatus: result.Pass,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			test.Check.UnmarshalDataMap()

			curShellCommander := command.ShellCommander
			defer func() { command.ShellCommander = curShellCommander }()

			internal.TestRunCheck(t, test)
		})
	}
}

func TestAdminUserRemediate(t *testing.T) {
	assert := assert.New(t)

	curShellCommander := command.ShellCommander
	defer func() { command.ShellCommander = curShellCommander }()

	t.Run("drushError", func(t *testing.T) {
		command.ShellCommander = internal.ShellCommanderMaker(
			nil,
			&exec.ExitError{Stderr: []byte("unable to run drush command")},
			nil)

		c := AdminUserCheck{}
		c.AddBreach(&breach.KeyValueBreach{
			Key:        "is_admin: true",
			ValueLabel: "role",
			Value:      "foo",
		})
		c.Remediate()
		assert.EqualValues([]breach.Breach{&breach.KeyValueBreach{
			BreachType: "key-value",
			Key:        "is_admin: true",
			ValueLabel: "role",
			Value:      "foo",
			Remediation: breach.Remediation{
				Status: "failed",
				Messages: []string{"failed to set is_admin to false for role 'foo' " +
					"due to error: unable to run drush command"},
			},
		}}, c.Result.Breaches)
	})

	t.Run("drushCommandIsCorrect", func(t *testing.T) {
		var generatedCommand string
		command.ShellCommander = internal.ShellCommanderMaker(nil, nil, &generatedCommand)

		c := AdminUserCheck{}
		c.AddBreach(&breach.KeyValueBreach{
			Key:        "is_admin: true",
			ValueLabel: "role",
			Value:      "foo",
		})
		c.Remediate()
		assert.Equal("vendor/drush/drush/drush config:set user.role.foo is_admin 0", generatedCommand)
	})
}
