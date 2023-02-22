package drupal_test

import (
	"fmt"
	"os/exec"
	"reflect"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestAdminUserInit(t *testing.T) {
	c := drupal.AdminUserCheck{}
	c.Init(drupal.AdminUser)
	assert.True(t, c.RequiresDb)
}

func TestAdminUserMerge(t *testing.T) {
	assert := assert.New(t)

	c := drupal.AdminUserCheck{
		DrushCommand: drupal.DrushCommand{
			DrushPath: "/path/to/drush",
		},
		AllowedRoles: []string{"role1"},
	}
	c.Merge(&drupal.AdminUserCheck{
		DrushCommand: drupal.DrushCommand{
			DrushPath: "/new/path/to/drush",
		},
		AllowedRoles: []string{"role2", "role3"},
	})
	assert.EqualValues(drupal.AdminUserCheck{
		DrushCommand: drupal.DrushCommand{
			DrushPath: "/new/path/to/drush",
		},
		AllowedRoles: []string{"role2", "role3"},
	}, c)
}

func TestAdminUserFetchData(t *testing.T) {
	assert := assert.New(t)

	t.Run("drushNotFound", func(t *testing.T) {
		c := drupal.AdminUserCheck{}
		c.FetchData()
		assert.Equal(shipshape.Fail, c.Result.Status)
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
		c := drupal.AdminUserCheck{}
		c.FetchData()
		assert.Equal(shipshape.Fail, c.Result.Status)
		assert.EqualValues([]string{"unable to run drush command"}, c.Result.Failures)
	})

	// correct data.
	t.Run("correctData", func(t *testing.T) {
		expectedStdout := `{"anonymous":{"is_admin": false}}`

		curShellCommander := command.ShellCommander
		defer func() { command.ShellCommander = curShellCommander }()
		command.ShellCommander = internal.ShellCommanderMaker(&expectedStdout, nil, nil)

		c := drupal.AdminUserCheck{}
		c.FetchData()
		assert.NotEqual(shipshape.Fail, c.Result.Status)
		assert.NotEqual(shipshape.Pass, c.Result.Status)
		assert.Equal([]byte(`{"anonymous":{"is_admin": false}}`), c.DataMap["anonymous"])
	})
}

func TestAdminUserUnmarshalData(t *testing.T) {
	assert := assert.New(t)
	c := drupal.AdminUserCheck{}

	// Empty datamap.
	t.Run("emptyDataMap", func(t *testing.T) {
		c.UnmarshalDataMap()
		assert.Equal(shipshape.Fail, c.Result.Status)
		assert.EqualValues([]string{"no data provided"}, c.Result.Failures)
	})

	// Incorrect json.
	t.Run("incorrectJSON", func(t *testing.T) {
		c = drupal.AdminUserCheck{
			CheckBase: shipshape.CheckBase{
				DataMap: map[string][]byte{
					"anonymous": []byte(`{"is_admin":false, "id": "anonymous"]}`)},
			},
		}
		c.UnmarshalDataMap()
		assert.Equal(shipshape.Fail, c.Result.Status)
		assert.EqualValues([]string{"invalid character ']' after object key:value pair"}, c.Result.Failures)
	})

	// Correct json.
	t.Run("correctJSON", func(t *testing.T) {
		c = drupal.AdminUserCheck{
			CheckBase: shipshape.CheckBase{
				DataMap: map[string][]byte{
					"anonymous": []byte(`{"is_admin":false, "id": "anonymous"}`)},
			},
		}

		c.UnmarshalDataMap()
		assert.NotEqual(shipshape.Fail, c.Result.Status)
		assert.NotEqual(shipshape.Pass, c.Result.Status)
		roleConfigsVal := reflect.ValueOf(c).FieldByName("roleConfigs")
		assert.Equal("map[string]bool{\"anonymous\":false}", fmt.Sprintf("%#v", roleConfigsVal))
	})
}

func TestAdminUserRunCheck(t *testing.T) {
	assert := assert.New(t)
	c := drupal.AdminUserCheck{}

	// Role does not have is_admin:true.
	t.Run("roleNotAdmin", func(t *testing.T) {
		c = drupal.AdminUserCheck{
			CheckBase: shipshape.CheckBase{
				DataMap: map[string][]byte{
					"anonymous": []byte(`{"is_admin":false, "id": "anonymous"}`)},
			},
			AllowedRoles: []string{"authenticated", "content-admin"},
		}
		c.UnmarshalDataMap()
		c.RunCheck()
		assert.Equal(shipshape.Pass, c.Result.Status)
	})

	// Role has is_admin:true.
	t.Run("roleAdmin", func(t *testing.T) {
		c = drupal.AdminUserCheck{
			CheckBase: shipshape.CheckBase{
				DataMap: map[string][]byte{
					"anonymous": []byte(`{"is_admin":true, "id": "anonymous"}`)},
			},
			AllowedRoles: []string{"content-admin"},
		}
		c.UnmarshalDataMap()
		c.RunCheck()
		assert.Equal(shipshape.Fail, c.Result.Status)
		assert.EqualValues([]string{"Role [anonymous] has `is_admin: true`"}, c.Result.Failures)
	})

	// Role has is_admin:true but is allowed.
	t.Run("roleAdminAllowed", func(t *testing.T) {
		c = drupal.AdminUserCheck{
			CheckBase: shipshape.CheckBase{
				DataMap: map[string][]byte{
					"anonymous": []byte(`{"is_admin":true, "id": "anonymous"}`)},
			},
			AllowedRoles: []string{"anonymous", "content-admin"},
		}
		c.UnmarshalDataMap()
		c.RunCheck()
		assert.Equal(shipshape.Pass, c.Result.Status)
	})

	c.UnmarshalDataMap()
	c.RunCheck()
	assert.Equal(shipshape.Pass, c.Result.Status)
}
