package drupal_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/checks/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

func TestRolePermissionsCheck_Init(t *testing.T) {
	c := drupal.RolePermissionsCheck{
		RoleId:                "authenticated",
		RequiredPermissions:   []string{"setup own tfa"},
		DisallowedPermissions: []string{},
	}
	c.Init(drupal.RolePermissions)
	assert.True(t, c.RequiresDb)
	assert.Equal(t, "authenticated", c.RoleId)
}

func TestRolePermissionsCheck_Merge(t *testing.T) {
	c := drupal.RolePermissionsCheck{}
	c.Init(drupal.RolePermissions)
	assert.Nil(t, c.Merge(&c))
}

func TestRolePermissionsCheck_HasData(t *testing.T) {
	c := drupal.RolePermissionsCheck{}
	c.Init(drupal.RolePermissions)
	assert.True(t, c.HasData(true))
}

func TestRolePermissionsCheck_RunCheck(t *testing.T) {
	assertions := assert.New(t)
	curShellCommander := command.ShellCommander
	defer func() { command.ShellCommander = curShellCommander }()

	t.Run("failOnNoRoleProvided", func(t *testing.T) {
		c := drupal.RolePermissionsCheck{}
		c.RunCheck()
		c.Result.DetermineResultStatus(false)
		assertions.Equal(result.Fail, c.Result.Status)
		assertions.ElementsMatch(
			[]breach.Breach{&breach.ValueBreach{
				BreachType: breach.BreachTypeValue,
				Value:      "no role ID provided"}},
			c.Result.Breaches,
		)
	})

	t.Run("failOnDrushNotFound", func(t *testing.T) {
		c := drupal.RolePermissionsCheck{
			RoleId: "authenticated",
		}
		c.RunCheck()
		c.Result.DetermineResultStatus(false)
		assertions.Equal(result.Fail, c.Result.Status)
		assertions.ElementsMatch(
			[]breach.Breach{&breach.ValueBreach{
				BreachType: breach.BreachTypeValue,
				Value:      "vendor/drush/drush/drush: no such file or directory"}},
			c.Result.Breaches)
	})

	t.Run("failOnDrushError", func(t *testing.T) {
		c := drupal.RolePermissionsCheck{
			RoleId: "authenticated",
		}
		c.Init(drupal.RolePermissions)
		assertions.True(c.RequiresDb)

		command.ShellCommander = internal.ShellCommanderMaker(
			nil,
			&exec.ExitError{Stderr: []byte("Unexpected error")},
			nil,
		)
		c.RunCheck()
		c.Result.DetermineResultStatus(false)
		assertions.Empty(c.Result.Passes)
		assertions.ElementsMatch(
			[]breach.Breach{&breach.ValueBreach{
				BreachType: breach.BreachTypeValue,
				CheckType:  "drupal-role-permissions",
				Severity:   "normal",
				Value:      "Unexpected error"}},
			c.Result.Breaches)
	})

	t.Run("failOnDrushInvalidResponse", func(t *testing.T) {
		c := drupal.RolePermissionsCheck{
			RoleId: "authenticated",
		}
		c.Init(drupal.RolePermissions)
		assertions.True(c.RequiresDb)

		stdout := "Unexpected error"
		command.ShellCommander = internal.ShellCommanderMaker(
			&stdout,
			nil,
			nil,
		)
		c.RunCheck()
		c.Result.DetermineResultStatus(false)
		assertions.Equal(result.Fail, c.Result.Status)
		assertions.Empty(c.Result.Passes)
		assertions.ElementsMatch(
			[]breach.Breach{&breach.ValueBreach{
				BreachType: breach.BreachTypeValue,
				CheckType:  "drupal-role-permissions",
				Severity:   "normal",
				Value:      "invalid character 'U' looking for beginning of value"}},
			c.Result.Breaches)
	})

	t.Run("failOnPermissions", func(t *testing.T) {
		c := drupal.RolePermissionsCheck{
			RoleId:                "authenticated",
			RequiredPermissions:   []string{"setup own tfa"},
			DisallowedPermissions: []string{"administer users"},
		}
		c.Init(drupal.RolePermissions)
		assertions.True(c.RequiresDb)

		stdout := `
{
    "authenticated": {
        "perms": [
            "access content",
			"administer users",
            "opt-in or out of google analytics tracking",
            "search content",
            "use text format webform_default",
            "view media",
            "view securitytxt"
        ]
    }
}
`
		command.ShellCommander = internal.ShellCommanderMaker(
			&stdout,
			nil,
			nil,
		)
		c.RunCheck()
		c.Result.DetermineResultStatus(false)
		assertions.Equal(result.Fail, c.Result.Status)
		assertions.Empty(c.Result.Passes)
		assertions.ElementsMatch(
			[]breach.Breach{
				&breach.KeyValueBreach{
					BreachType: breach.BreachTypeKeyValue,
					CheckType:  "drupal-role-permissions",
					Severity:   "normal",
					KeyLabel:   "role",
					Key:        "authenticated",
					ValueLabel: "missing permissions",
					Value:      "[setup own tfa]",
				},
				&breach.KeyValueBreach{
					BreachType: breach.BreachTypeKeyValue,
					CheckType:  "drupal-role-permissions",
					Severity:   "normal",
					KeyLabel:   "role",
					Key:        "authenticated",
					ValueLabel: "disallowed permissions",
					Value:      "[administer users]",
				},
			},
			c.Result.Breaches)
	})

	t.Run("passOnPermissions", func(t *testing.T) {
		c := drupal.RolePermissionsCheck{
			RoleId:              "authenticated",
			RequiredPermissions: []string{"setup own tfa"},
		}
		c.Init(drupal.RolePermissions)
		assertions.True(c.RequiresDb)

		stdout := `
{
    "authenticated": {
        "perms": [
            "access content",
            "opt-in or out of google analytics tracking",
            "search content",
			"setup own tfa",
            "use text format webform_default",
            "view media",
            "view securitytxt"
        ]
    }
}
`
		command.ShellCommander = internal.ShellCommanderMaker(
			&stdout,
			nil,
			nil,
		)
		c.RunCheck()
		c.Result.DetermineResultStatus(false)
		assertions.Equal(result.Pass, c.Result.Status)
		assertions.Empty(c.Result.Breaches)
	})
}
