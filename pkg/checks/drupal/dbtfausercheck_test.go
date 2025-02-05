package drupal_test

import (
	"os/exec"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	. "github.com/salsadigitalauorg/shipshape/pkg/checks/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/stretchr/testify/assert"
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
		assert.Equal(result.Fail, c.Result.Status)
		assert.Empty(c.Result.Passes)
		assert.EqualValues(
			[]breach.Breach{&breach.ValueBreach{
				BreachType: "value",
				CheckType:  "drupal-db-user-tfa",
				Severity:   "normal",
				ValueLabel: "error fetching drush user info",
				Value:      "<nil>: unable to run drush command",
			}},
			c.Result.Breaches,
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
		assert.Equal(result.Fail, c.Result.Status)
		assert.Empty(c.Result.Passes)
		assert.EqualValues(
			[]breach.Breach{&breach.ValueBreach{
				BreachType: "value",
				CheckType:  "drupal-db-user-tfa",
				Severity:   "normal",
				ValueLabel: "users with TFA disabled",
				Value:      "shipshape-1:1",
			}},
			c.Result.Breaches,
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
		assert.Equal(result.Fail, c.Result.Status)
		assert.Empty(c.Result.Passes)
		assert.EqualValues(
			[]breach.Breach{&breach.ValueBreach{
				BreachType: "value",
				CheckType:  "drupal-db-user-tfa",
				Severity:   "normal",
				ValueLabel: "users with TFA disabled",
				Value:      "shipshape-1:1, shipshape-2:2",
			}},
			c.Result.Breaches,
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
		assert.Equal(result.Pass, c.Result.Status)
		assert.Empty(c.Result.Breaches)
		assert.EqualValues(
			[]string{"All active users have two-factor authentication enabled."},
			c.Result.Passes,
		)
	})
}
