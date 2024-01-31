package drupal_test

import (
	"os/exec"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/checks/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/stretchr/testify/assert"
)

func TestForbiddenUserCheck_Init(t *testing.T) {
	c := drupal.ForbiddenUserCheck{}
	c.Init(drupal.ForbiddenUser)
	assert.True(t, c.RequiresDb)
	assert.Equal(t, "1", c.UserId)
}

func TestForbiddenUserCheck_Merge(t *testing.T) {
	c := drupal.ForbiddenUserCheck{}
	c.Init(drupal.ForbiddenUser)
	assert.Nil(t, c.Merge(&c))
}

func TestForbiddenUserCheck_HasData(t *testing.T) {
	c := drupal.ForbiddenUserCheck{}
	c.Init(drupal.ForbiddenUser)
	assert.True(t, c.HasData(true))
}

func TestForbiddenUserCheck_Init2(t *testing.T) {
	c := drupal.ForbiddenUserCheck{UserId: "2"}
	c.Init(drupal.ForbiddenUser)
	assert.True(t, c.RequiresDb)
	assert.Equal(t, "2", c.UserId)
}

func TestForbiddenUserCheck_RunCheck(t *testing.T) {
	assertions := assert.New(t)
	curShellCommander := command.ShellCommander
	defer func() { command.ShellCommander = curShellCommander }()

	t.Run("failOnDrushNotFound", func(t *testing.T) {
		c := drupal.ForbiddenUserCheck{}
		c.RunCheck()
		assertions.Equal(result.Fail, c.Result.Status)
		assertions.EqualValues(
			[]result.Breach{&result.ValueBreach{
				BreachType: "value",
				Value:      "vendor/drush/drush/drush: no such file or directory",
			}},
			c.Result.Breaches,
		)
	})

	t.Run("failOnDrushError", func(t *testing.T) {
		c := drupal.ForbiddenUserCheck{}
		c.Init(drupal.ForbiddenUser)
		assertions.True(c.RequiresDb)

		command.ShellCommander = internal.ShellCommanderMaker(
			nil,
			&exec.ExitError{Stderr: []byte("Unable to find a matching user")},
			nil,
		)
		c.RunCheck()
		assertions.Empty(c.Result.Passes)
		assertions.ElementsMatch(
			[]result.Breach{&result.ValueBreach{
				BreachType: "value",
				CheckType:  "drupal-user-forbidden",
				Severity:   "normal",
				Value:      "Unable to find a matching user",
			}},
			c.Result.Breaches,
		)
	})

	t.Run("failOnDrushInvalidResponse", func(t *testing.T) {
		c := drupal.ForbiddenUserCheck{}
		c.Init(drupal.ForbiddenUser)
		assertions.True(c.RequiresDb)

		stdout := "Unable to find a matching user"
		command.ShellCommander = internal.ShellCommanderMaker(
			&stdout,
			nil,
			nil,
		)
		c.RunCheck()
		assertions.Equal(result.Fail, c.Result.Status)
		assertions.Empty(c.Result.Passes)
		assertions.ElementsMatch(
			[]result.Breach{&result.ValueBreach{
				BreachType: "value",
				CheckType:  "drupal-user-forbidden",
				Severity:   "normal",
				Value:      "invalid character 'U' looking for beginning of value",
			}},
			c.Result.Breaches,
		)
	})

	t.Run("failOnActiveUser", func(t *testing.T) {
		c := drupal.ForbiddenUserCheck{}
		c.Init(drupal.ForbiddenUser)

		stdout := `
{
    "1": {
        "user_status": "1"
    }
}
`
		command.ShellCommander = internal.ShellCommanderMaker(
			&stdout,
			nil,
			nil,
		)
		c.RunCheck()
		assertions.Equal(result.Fail, c.Result.Status)
		assertions.Empty(c.Result.Passes)
		assertions.ElementsMatch(
			[]result.Breach{&result.KeyValueBreach{
				BreachType: "key-value",
				CheckType:  "drupal-user-forbidden",
				Severity:   "normal",
				Key:        "forbidden user is active",
				Value:      "1",
			}},
			c.Result.Breaches,
		)
	})

	t.Run("passOnInactiveUser", func(t *testing.T) {
		c := drupal.ForbiddenUserCheck{}
		c.Init(drupal.ForbiddenUser)

		stdout := `
{
    "1": {
        "user_status": "0"
    }
}
`
		command.ShellCommander = internal.ShellCommanderMaker(
			&stdout,
			nil,
			nil,
		)
		c.RunCheck()
		assertions.Equal(result.Pass, c.Result.Status)
		assertions.Empty(c.Result.Breaches)
		assertions.ElementsMatch(
			[]string{"No forbidden user is active."},
			c.Result.Passes,
		)
	})
}

func TestForbiddenUserCheck_Remediate(t *testing.T) {
	assertions := assert.New(t)
	curShellCommander := command.ShellCommander
	defer func() { command.ShellCommander = curShellCommander }()

	t.Run("failOnDrushError", func(t *testing.T) {
		c := drupal.ForbiddenUserCheck{}
		c.Init(drupal.ForbiddenUser)
		assertions.True(c.RequiresDb)

		command.ShellCommander = internal.ShellCommanderMaker(
			nil,
			&exec.ExitError{Stderr: []byte("Unable to find a matching user")},
			nil,
		)
		c.Remediate()
		assertions.EqualValues([]result.Breach{&result.ValueBreach{
			BreachType: "value",
			Value:      "Unable to find a matching user",
		}}, c.Result.Breaches)
	})

	t.Run("passOnBlockingInactiveUser", func(t *testing.T) {
		c := drupal.ForbiddenUserCheck{CheckBase: config.CheckBase{PerformRemediation: true}}
		c.Init(drupal.ForbiddenUser)

		stdout := `
{
    "1": {
        "user_status": "1"
    }
}
`
		command.ShellCommander = internal.ShellCommanderMaker(
			&stdout,
			nil,
			nil,
		)
		c.RunCheck()
		assertions.Equal(result.Pass, c.Result.Status)
		assertions.Empty(c.Result.Breaches)
		assertions.ElementsMatch(
			[]string{"Blocked the forbidden user [1]"},
			c.Result.Remediations,
		)
		assertions.ElementsMatch(
			[]string{"No forbidden user is active."},
			c.Result.Passes,
		)
	})
}
