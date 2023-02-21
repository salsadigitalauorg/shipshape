package drupal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os/exec"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

const AdminUser shipshape.CheckType = "drupal-admin-user"

// AdminUserCheck fetches all role configurations from the database and verifies
// they do not have is_admin set to true.
type AdminUserCheck struct {
	shipshape.CheckBase `yaml:",inline"`
	DrushCommand        `yaml:",inline"`
	// List of role names allowed to have is_admin set to true.
	AllowedRoles []string `yaml:"allowed-roles"`
	userRoles    []string
	roleConfigs  map[string]bool
}

type roleConf struct {
	IsAdmin bool `json:"is_admin"`
	Name string `json:"id"`
}

// Init implementation for the drush-based user role config check.
func (c *AdminUserCheck) Init(ct shipshape.CheckType) {
	c.CheckBase.Init(ct)
	c.RequiresDb = true
}

// Merge implementation for AdminUserCheck check.
func (c *AdminUserCheck) Merge(mergeCheck shipshape.Check) error {
	adminUserMergeCheck := mergeCheck.(*AdminUserCheck)
	if err := c.CheckBase.Merge(&adminUserMergeCheck.CheckBase); err != nil {
		return err
	}

	c.DrushCommand.Merge(adminUserMergeCheck.DrushCommand)
	utils.MergeStringSlice(&c.AllowedRoles, adminUserMergeCheck.AllowedRoles)
	return nil
}

// getActiveRoles runs the drush command to populate data for the roles config check.
func (c *AdminUserCheck) getActiveRoles() map[string]string {
  var err error

	activeRoles := map[string][]byte{}
	rolesListMap := map[string]string{}

  cmd := []string{"role:list", "--fields=.", "--format=json"}
  activeRoles["user-roles"], err = Drush(c.DrushPath, c.Alias, cmd).Exec()
  var pathErr *fs.PathError
  if err != nil && errors.As(err, &pathErr) {
    c.AddFail(pathErr.Path + ": " + pathErr.Err.Error())
  } else if err != nil {
  		msg := string(err.(*exec.ExitError).Stderr)
  		c.AddFail(strings.ReplaceAll(strings.TrimSpace(msg), "  \n  ", ""))
  } else {
    // Unmarshal roles JSON.
    err = json.Unmarshal(activeRoles["user-roles"], &rolesListMap)
    var synErr *json.SyntaxError
    if err != nil && errors.As(err, &synErr) {
      c.AddFail(err.Error())
    }
  }

  return rolesListMap
}

// FetchData runs the drush command for each active role to extract its config.
func (c *AdminUserCheck) FetchData() {
	var err error

	activeRoles := c.getActiveRoles()
	if c.Result.Status == shipshape.Fail {
		return
	}

	// Loop through active roles and pull active config with drush.
	rolesMap := map[string][]byte{}
	for i := range activeRoles {
	  cmd := []string{"cget", "user.role." + i, "--format=json"}
  	rolesMap[i], err = Drush(c.DrushPath, c.Alias, cmd).Exec()
  	c.DataMap = rolesMap
  }

	if err != nil {
		msg := string(err.(*exec.ExitError).Stderr)
		c.AddFail(strings.ReplaceAll(strings.TrimSpace(msg), "  \n  ", ""))
	}
}

// UnmarshalDataMap parses the data map json entries
// into the roleConfigs for further processing.
func (c *AdminUserCheck) UnmarshalDataMap() {
  if len(c.DataMap) == 0 {
		c.AddFail("no data provided")
		return
	}

	c.roleConfigs = map[string]bool{}
  for _, element := range c.DataMap {
    var role roleConf
    err := json.Unmarshal([]byte(element), &role)
    var synErr *json.SyntaxError
    if err != nil && errors.As(err, &synErr) {
      c.AddFail(err.Error())
      return
    }
    // Collect role config.
    c.roleConfigs[role.Name] = role.IsAdmin
  }
}

// RunCheck implements the Check logic for all active roles.
func (c *AdminUserCheck) RunCheck() {
  for roleName, isAdmin := range c.roleConfigs {
    allowedRole := utils.StringSliceContains(c.AllowedRoles, roleName)
		if allowedRole {
			continue
		}

		if (isAdmin) {
			c.AddFail(fmt.Sprintf("Role [%s] has `is_admin: true`", roleName))
		}
	}

	if len(c.Result.Failures) == 0 {
		c.Result.Status = shipshape.Pass
	}
}
