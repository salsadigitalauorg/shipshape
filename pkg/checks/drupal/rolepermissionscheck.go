package drupal

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os/exec"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

const RolePermissions config.CheckType = "drupal-role-permissions"

// RolePermissionsCheck checks the permissions of a role.
type RolePermissionsCheck struct {
	config.CheckBase `yaml:",inline"`
	DrushCommand     `yaml:",inline"`
	// The Role ID to check.
	RoleId string `yaml:"rid"`
	// List permissions the above role is required to have.
	RequiredPermissions []string `yaml:"required-permissions"`
	// List permissions the above role must not have.
	DisallowedPermissions []string `yaml:"disallowed-permissions"`
}

// Init implementation for the drush-based role permissions check.
func (c *RolePermissionsCheck) Init(ct config.CheckType) {
	c.CheckBase.Init(ct)
	c.RequiresDb = true
}

// GetRolePermissions get the permissions of the role.
func (c *RolePermissionsCheck) GetRolePermissions() []string {
	// Command: drush role:list --filter=id=anonymous --fields=perms --format=json
	cmd := []string{"role:list", "--filter=id=" + c.RoleId, "--fields=perms", "--format=json"}

	drushOutput, err := Drush(c.DrushPath, c.Alias, cmd).Exec()

	var pathError *fs.PathError
	if err != nil && errors.As(err, &pathError) {
		c.AddBreach(&result.ValueBreach{
			Value: pathError.Path + ": " + pathError.Err.Error()})
	} else if err != nil {
		msg := string(err.(*exec.ExitError).Stderr)
		c.AddBreach(&result.ValueBreach{
			Value: strings.ReplaceAll(strings.TrimSpace(msg), "  \n  ", "")})
	} else {
		// Unmarshal role:list JSON.
		// {
		//    "anonymous": {
		//        "perms": [
		//            "access content",
		//            "search content",
		//            "view media",
		//            "view securitytxt"
		//        ]
		//    }
		//}
		rolePermissionsMap := map[string]map[string][]string{}
		err = json.Unmarshal(drushOutput, &rolePermissionsMap)
		var syntaxError *json.SyntaxError
		if err != nil && errors.As(err, &syntaxError) {
			c.AddBreach(&result.ValueBreach{Value: err.Error()})
		}

		if len(rolePermissionsMap[c.RoleId]["perms"]) > 0 {
			return rolePermissionsMap[c.RoleId]["perms"]
		}
	}

	return nil
}

// HasData implementation for RolePermissionsCheck check.
func (c *RolePermissionsCheck) HasData(failCheck bool) bool {
	return true
}

// Merge implementation for RolePermissionsCheck check.
func (c *RolePermissionsCheck) Merge(mergeCheck config.Check) error {
	return nil
}

// RunCheck implements the Check logic for role permissions.
func (c *RolePermissionsCheck) RunCheck() {
	if c.RoleId == "" {
		c.AddBreach(&result.ValueBreach{Value: "no role ID provided"})
		return
	}

	rolePermissions := c.GetRolePermissions()
	// Check for required permissions.
	diff := utils.StringSlicesInterdiffUnique(rolePermissions, c.RequiredPermissions)
	if len(diff) > 0 {
		c.AddBreach(&result.KeyValueBreach{
			KeyLabel:   "role",
			Key:        c.RoleId,
			ValueLabel: "missing permissions",
			Value:      "[" + strings.Join(diff, ", ") + "]",
		})
	}

	// Check for disallowed permissions.
	diff = utils.StringSlicesIntersectUnique(rolePermissions, c.DisallowedPermissions)
	if len(diff) > 0 {
		c.AddBreach(&result.KeyValueBreach{
			KeyLabel:   "role",
			Key:        c.RoleId,
			ValueLabel: "disallowed permissions",
			Value:      "[" + strings.Join(diff, ", ") + "]",
		})
	}

	if len(c.Result.Breaches) == 0 {
		c.Result.Status = result.Pass
	}
}
