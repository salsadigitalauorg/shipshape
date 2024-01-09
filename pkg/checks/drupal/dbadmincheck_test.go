package drupal_test

import (
	"fmt"
	"os/exec"
	"reflect"
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/checks/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/result"

	"github.com/stretchr/testify/assert"
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
		assert.Equal(result.Fail, c.Result.Status)
		assert.EqualValues([]string{"vendor/drush/drush/drush: no such file or directory"}, c.Result.Failures)

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
		assert.Equal(result.Fail, c.Result.Status)
		assert.EqualValues([]string{"unable to run drush command"}, c.Result.Failures)
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
		assert.Equal(result.Fail, c.Result.Status)
		assert.EqualValues([]string{"no data provided"}, c.Result.Failures)
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
		assert.Equal(result.Fail, c.Result.Status)
		assert.EqualValues([]string{"invalid character ']' after object key:value pair"}, c.Result.Failures)
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
			ExpectFails:  []string{"Role [anonymous] has `is_admin: true`"},
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

		// Role has is_admin:true, with remediation.
		{
			Name: "roleAdminWithRemediation",
			Check: &AdminUserCheck{
				CheckBase: config.CheckBase{
					DataMap: map[string][]byte{
						"anonymous": []byte(`{"is_admin":true, "id": "anonymous"}`)},
					PerformRemediation: true,
				},
				AllowedRoles: []string{"content-admin"},
			},
			PreRun: func(t *testing.T) {
				command.ShellCommander = internal.ShellCommanderMaker(nil, nil, nil)
			},
			ExpectStatus:       result.Pass,
			ExpectRemediations: []string{"Fixed disallowed admin setting for role [anonymous]"},
		},

		// Role has is_admin:true, with remediation error.
		{
			Name: "roleAdminWithRemediationError",
			Check: &AdminUserCheck{
				CheckBase: config.CheckBase{
					DataMap: map[string][]byte{
						"anonymous": []byte(`{"is_admin":true, "id": "anonymous"}`)},
					PerformRemediation: true,
				},
				AllowedRoles: []string{"content-admin"},
			},
			PreRun: func(t *testing.T) {
				command.ShellCommander = internal.ShellCommanderMaker(
					nil,
					&exec.ExitError{Stderr: []byte("unable to run drush command")},
					nil,
				)
			},
			ExpectStatus: result.Fail,
			ExpectFails:  []string{"Failed to fix disallowed admin setting for role [anonymous] due to error: unable to run drush command"},
		},
	}

	curShellCommander := command.ShellCommander
	defer func() { command.ShellCommander = curShellCommander }()

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			test.Check.UnmarshalDataMap()
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
		err := c.Remediate("foo")
		assert.Error(err, "unable to run drush command")
	})

	t.Run("drushCommandIsCorrect", func(t *testing.T) {
		var generatedCommand string
		command.ShellCommander = internal.ShellCommanderMaker(nil, nil, &generatedCommand)

		c := AdminUserCheck{}
		c.Remediate("foo")
		assert.Equal("vendor/drush/drush/drush config:set user.role.foo is_admin 0", generatedCommand)
	})
}
