package drupal_test

import (
	"github.com/salsadigitalauorg/shipshape/pkg/checks/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/stretchr/testify/assert"
	"os/exec"
	"testing"
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
		assertions.Equal(result.Fail, c.Result.Status)
		assertions.EqualValues([]string{"no role ID provided"}, c.Result.Failures)
	})

	t.Run("failOnDrushNotFound", func(t *testing.T) {
		c := drupal.RolePermissionsCheck{
			RoleId: "authenticated",
		}
		c.RunCheck()
		assertions.Equal(result.Fail, c.Result.Status)
		assertions.EqualValues([]string{"vendor/drush/drush/drush: no such file or directory"}, c.Result.Failures)
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
		assertions.Empty(c.Result.Passes)
		assertions.ElementsMatch(
			[]string{"Unexpected error"},
			c.Result.Failures,
		)
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
		assertions.Equal(result.Fail, c.Result.Status)
		assertions.Empty(c.Result.Passes)
		assertions.ElementsMatch(
			[]string{"invalid character 'U' looking for beginning of value"},
			c.Result.Failures,
		)
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
		assertions.Equal(result.Fail, c.Result.Status)
		assertions.Empty(c.Result.Passes)
		assertions.ElementsMatch(
			[]string{
				"The role [authenticated] does not have all required permissions.",
				"The role [authenticated] has disallowed permissions.",
			},
			c.Result.Failures,
		)
		assertions.Equal("Missing permissions", (c.Result.Breaches[0]).(result.ValueBreach).ValueLabel)
		assertions.Equal("setup own tfa", (c.Result.Breaches[0]).(result.ValueBreach).Value)
		assertions.Equal("Disallowed permissions", (c.Result.Breaches[1]).(result.ValueBreach).ValueLabel)
		assertions.Equal("administer users", (c.Result.Breaches[1]).(result.ValueBreach).Value)
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
		assertions.Equal(result.Pass, c.Result.Status)
		assertions.Empty(c.Result.Failures)
	})
}
