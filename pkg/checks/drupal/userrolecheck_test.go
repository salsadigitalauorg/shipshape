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

func TestUserRoleCheckInit(t *testing.T) {
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

func TestUserRoleCheckFetchData(t *testing.T) {
	assert := assert.New(t)

	t.Run("drushNotFound", func(t *testing.T) {
		c := UserRoleCheck{}
		c.FetchData()
		assert.EqualValues(
			[]breach.Breach{&breach.ValueBreach{
				BreachType: "value",
				Value:      "vendor/drush/drush/drush: no such file or directory",
			}},
			c.Result.Breaches,
		)
	})

	t.Run("drushError", func(t *testing.T) {
		curShellCommander := command.ShellCommander
		defer func() { command.ShellCommander = curShellCommander }()
		sqlQueryFail := true
		command.ShellCommander = func(name string, arg ...string) command.IShellCommand {
			var stdout []byte
			return internal.TestShellCommand{
				OutputterFunc: func() ([]byte, error) {
					if arg[0] == "sql:query" {
						if sqlQueryFail {
							return nil, &exec.ExitError{Stderr: []byte("unable to run drush sql query")}
						}
						return []byte(`1,0`), nil
					}
					return stdout, &exec.ExitError{Stderr: []byte("unable to run drush command")}
				},
			}
		}
		c := UserRoleCheck{}
		c.FetchData()
		assert.EqualValues(
			[]breach.Breach{&breach.ValueBreach{
				BreachType: "value",
				Value:      "unable to run drush sql query",
			}},
			c.Result.Breaches,
		)

		sqlQueryFail = false
		c = UserRoleCheck{}
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

func TestUserRoleCheckUnmarshalData(t *testing.T) {
	assert := assert.New(t)

	// Empty datamap.
	c := UserRoleCheck{}
	c.UnmarshalDataMap()
	assert.EqualValues(
		[]breach.Breach{&breach.ValueBreach{
			BreachType: "value",
			Value:      "no data provided",
		}},
		c.Result.Breaches,
	)

	// Incorrect json.
	c = UserRoleCheck{
		CheckBase: config.CheckBase{
			DataMap: map[string][]byte{
				"user-info": []byte(`{"1":{"roles":"authenticated"]}}`)},
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

func TestUserRoleCheckRunCheck(t *testing.T) {
	tt := []internal.RunCheckTest{
		{
			Name: "noDisallowedRoleProvided",
			Check: &UserRoleCheck{
				CheckBase: config.CheckBase{
					DataMap: map[string][]byte{
						"user-info": []byte(`{"1":{"roles":["authenticated"]}}`)},
				},
			},
			ExpectStatus: result.Fail,
			ExpectFails: []breach.Breach{&breach.ValueBreach{
				BreachType: "value",
				Value:      "no disallowed role provided"}}},

		{
			Name: "userHasDisallowedRoles",
			Check: &UserRoleCheck{
				CheckBase: config.CheckBase{
					DataMap: map[string][]byte{
						"user-info": []byte(`{"1":{"roles":["authenticated","site-admin","content-admin"]}}`)},
				},
				Roles: []string{"site-admin", "content-admin"},
			},
			ExpectStatus: result.Fail,
			ExpectFails: []breach.Breach{&breach.KeyValuesBreach{
				BreachType: "key-values",
				KeyLabel:   "user",
				Key:        "1",
				ValueLabel: "disallowed roles",
				Values:     []string{"site-admin", "content-admin"}}}},

		{
			Name: "userAllowedToHaveDisallowedRoles",
			Check: &UserRoleCheck{
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
			},
			ExpectStatus: result.Pass},
	}

	for _, tc := range tt {
		tc.Check.UnmarshalDataMap()
		internal.TestRunCheck(t, tc)
	}
}
