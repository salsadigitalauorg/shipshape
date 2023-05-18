package drupal_test

import (
	"os/exec"
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/checks/drupal"
	"github.com/stretchr/testify/assert"
	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
)

func TestDbTfaUserCheck(t *testing.T) {
	assert := assert.New(t)
	curShellCommander := command.ShellCommander
	defer func() { command.ShellCommander = curShellCommander }()
	t.Run("drushError", func(t *testing.T) {
		c := DbUserTfaCheck{}
		c.Init(DbUserTfa)
		assert.True(c.RequiresDb)
		
		command.ShellCommander = internal.ShellCommanderMaker(
			nil,
			&exec.ExitError{Stderr: []byte("unable to run drush command")},
			nil,
		)
		c.FetchData()
		assert.Equal(config.Fail, c.Result.Status)
		assert.Empty(c.Result.Passes)
		assert.ElementsMatch(
			[]string{"Error calling drush ev."},
			c.Result.Failures,
		)
	})
	
	t.Run("failOnSingleUserWithoutTFA", func(t *testing.T) {
		c := DbUserTfaCheck{}
		c.Init(DbUserTfa)
		
		stdout := `
[
  {
    "uid": "1",
    "name": "shipshape-1"
  }
]
`
		command.ShellCommander = internal.ShellCommanderMaker(
			&stdout,
			nil,
			nil,
		)
		c.FetchData()
		c.RunCheck()
		assert.Equal(config.Fail, c.Result.Status)
		assert.Empty(c.Result.Passes)
		assert.ElementsMatch(
			[]string{"Two-factor authentication not enabled for active user shipshape-1, with UID 1."},
			c.Result.Failures,
		)
	})
	
	t.Run("failOnMultipleUserWithoutTFA", func(t *testing.T) {
		c := DbUserTfaCheck{}
		c.Init(DbUserTfa)

		stdout := `
[
  {
    "uid": "1",
    "name": "shipshape-1"
  },
  {
    "uid": "2",
    "name": "shipshape-2"
  }
]
`
		command.ShellCommander = internal.ShellCommanderMaker(
			&stdout,
			nil,
			nil,
		)
		c.FetchData()
		c.RunCheck()
		assert.Equal(config.Fail, c.Result.Status)
		assert.Empty(c.Result.Passes)
		assert.ElementsMatch(
			[]string{"Two-factor authentication not enabled for active user shipshape-1, with UID 1.", "Two-factor authentication not enabled for active user shipshape-2, with UID 2."},
			c.Result.Failures,
			)
	})
	t.Run("passOnEmptyQueryResult", func(t *testing.T) {
		c := DbUserTfaCheck{}
		c.Init(DbUserTfa)

		stdout := `
[]
`
		command.ShellCommander = internal.ShellCommanderMaker(
			&stdout,
			nil,
			nil,
			)
		c.FetchData()
		c.RunCheck()
		assert.Equal(config.Pass, c.Result.Status)
		assert.Empty(c.Result.Failures)
		assert.ElementsMatch(
			[]string{"All active users have two-factor authentication enabled."},
			c.Result.Passes,
			)
	})
}