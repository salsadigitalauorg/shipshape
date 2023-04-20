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

func TestInit(t *testing.T) {
	c := UserRoleCheck{}
	c.Init(UserRole)
	assert.True(t, c.RequiresDb)
}

func TestUserRoleMerge(t *testing.T) {
	assert := assert.New(t)

	c := UserRoleCheck{
		DrushCommand: DrushCommand{
			DrushPath: "/path/to/drush",
		},
		Roles:        []string{"role1"},
		AllowedUsers: []int{1, 2},
	}
	c.Merge(&UserRoleCheck{
		DrushCommand: DrushCommand{
			DrushPath: "/new/path/to/drush",
		},
		Roles:        []string{"role2"},
		AllowedUsers: []int{2, 3},
	})
	assert.EqualValues(UserRoleCheck{
		DrushCommand: DrushCommand{
			DrushPath: "/new/path/to/drush",
		},
		Roles:        []string{"role2"},
		AllowedUsers: []int{2, 3},
	}, c)
}

func TestFetchData(t *testing.T) {
	assert := assert.New(t)

	t.Run("drushNotFound", func(t *testing.T) {
		c := UserRoleCheck{}
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
		c := UserRoleCheck{}
		c.FetchData()
		assert.Equal(result.Fail, c.Result.Status)
		assert.EqualValues([]string{"unable to run drush command"}, c.Result.Failures)
	})

	// correct data.
	t.Run("correctData", func(t *testing.T) {
		command.ShellCommander = internal.ShellCommanderMaker(
			&[]string{`{"1":{"roles":["authenticated"]}}`}[0],
			nil,
			nil,
		)
		c := UserRoleCheck{}
		c.FetchData()
		assert.NotEqual(result.Fail, c.Result.Status)
		assert.NotEqual(result.Pass, c.Result.Status)
		assert.Equal([]byte(`{"1":{"roles":["authenticated"]}}`), c.DataMap["user-info"])
	})
}

func TestUnmarshalData(t *testing.T) {
	assert := assert.New(t)

	// Empty datamap.
	c := UserRoleCheck{}
	c.UnmarshalDataMap()
	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues([]string{"no data provided"}, c.Result.Failures)

	// Incorrect json.
	c = UserRoleCheck{
		CheckBase: config.CheckBase{
			DataMap: map[string][]byte{
				"user-info": []byte(`{"1":{"roles":"authenticated"]}}`)},
		},
	}
	c.UnmarshalDataMap()
	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues([]string{"invalid character ']' after object key:value pair"}, c.Result.Failures)

	// Correct json.
	c = UserRoleCheck{
		CheckBase: config.CheckBase{
			DataMap: map[string][]byte{
				"user-info": []byte(`{"1":{"roles":["authenticated"]}}`)},
		},
	}
	c.UnmarshalDataMap()
	assert.NotEqual(result.Fail, c.Result.Status)
	assert.NotEqual(result.Pass, c.Result.Status)
	userRolesVal := reflect.ValueOf(c).FieldByName("userRoles")
	assert.Equal("map[int][]string{1:[]string{\"authenticated\"}}", fmt.Sprintf("%#v", userRolesVal))
}

func TestRunCheck(t *testing.T) {
	assert := assert.New(t)

	// No disallowed roles provided.
	c := UserRoleCheck{
		CheckBase: config.CheckBase{
			DataMap: map[string][]byte{
				"user-info": []byte(`{"1":{"roles":["authenticated"]}}`)},
		},
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues([]string{"no disallowed role provided"}, c.Result.Failures)

	// User has disallowed roles.
	c = UserRoleCheck{
		CheckBase: config.CheckBase{
			DataMap: map[string][]byte{
				"user-info": []byte(`{"1":{"roles":["authenticated","site-admin","content-admin"]}}`)},
		},
		Roles: []string{"site-admin", "content-admin"},
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues([]string{"User 1 has disallowed roles: [site-admin, content-admin]"}, c.Result.Failures)

	// User allowed to have disallowed roles.
	c = UserRoleCheck{
		CheckBase: config.CheckBase{
			DataMap: map[string][]byte{
				"user-info": []byte(`
				{
					"1":{"roles":["authenticated"]},
					"2":{"roles":["authenticated","site-admin","content-admin"]}
				}
				`)},
		},
		Roles:        []string{"site-admin", "content-admin"},
		AllowedUsers: []int{2},
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	assert.Equal(result.Pass, c.Result.Status)
}
