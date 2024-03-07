package drupal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os/exec"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

const AdminUser config.CheckType = "drupal-admin-user"

// AdminUserCheck fetches all role configurations from the database and verifies
// they do not have is_admin set to true.
type AdminUserCheck struct {
	config.CheckBase `yaml:",inline"`
	DrushCommand     `yaml:",inline"`
	// List of role names allowed to have is_admin set to true.
	AllowedRoles []string `yaml:"allowed-roles"`
	roleConfigs  map[string]bool
}

type roleConf struct {
	IsAdmin bool   `json:"is_admin"`
	Name    string `json:"id"`
}

// Init implementation for the drush-based user role config check.
func (c *AdminUserCheck) Init(ct config.CheckType) {
	c.CheckBase.Init(ct)
	c.RequiresDb = true
}

// Merge implementation for AdminUserCheck check.
func (c *AdminUserCheck) Merge(mergeCheck config.Check) error {
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
	rolesListMap := map[string]string{}

	cmd := []string{"role:list", "--fields=.", "--format=json"}

	activeRoles, err := Drush(c.DrushPath, c.Alias, cmd).Exec()
	var pathErr *fs.PathError
	if err != nil && errors.As(err, &pathErr) {
		c.AddBreach(&breach.ValueBreach{
			Value: pathErr.Path + ": " + pathErr.Err.Error()})
	} else if err != nil {
		msg := string(err.(*exec.ExitError).Stderr)
		c.AddBreach(&breach.ValueBreach{
			Value: strings.ReplaceAll(strings.TrimSpace(msg), "  \n  ", "")})
	} else {
		// Unmarshal roles JSON.
		err = json.Unmarshal(activeRoles, &rolesListMap)
		var synErr *json.SyntaxError
		if err != nil && errors.As(err, &synErr) {
			c.AddBreach(&breach.ValueBreach{Value: err.Error()})
		}
	}

	return rolesListMap
}

// FetchData runs the drush command for each active role to extract its config.
func (c *AdminUserCheck) FetchData() {
	var err error

	activeRoles := c.getActiveRoles()
	if c.Result.Status == result.Fail {
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
		c.AddBreach(&breach.ValueBreach{
			Value: strings.ReplaceAll(strings.TrimSpace(msg), "  \n  ", "")})
	}
}

// UnmarshalDataMap parses the data map json entries
// into the roleConfigs for further processing.
func (c *AdminUserCheck) UnmarshalDataMap() {
	if len(c.DataMap) == 0 {
		c.AddBreach(&breach.ValueBreach{Value: "no data provided"})
		return
	}

	c.roleConfigs = map[string]bool{}
	for _, element := range c.DataMap {
		var role roleConf
		err := json.Unmarshal([]byte(element), &role)
		var synErr *json.SyntaxError
		if err != nil && errors.As(err, &synErr) {
			c.AddBreach(&breach.ValueBreach{Value: err.Error()})
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

		if isAdmin {
			c.AddBreach(&breach.KeyValueBreach{
				Key:        "is_admin: true",
				ValueLabel: "role",
				Value:      roleName,
			})
		}
	}
}

// Remediate attempts to fix a breach.
func (c *AdminUserCheck) Remediate() {
	for _, b := range c.Result.Breaches {
		b, ok := b.(*breach.KeyValueBreach)
		if !ok {
			continue
		}

		_, err := Drush(c.DrushPath, c.Alias, []string{"config:set", "user.role." + b.Value, "is_admin", "0"}).Exec()
		if err != nil {
			b.SetRemediation(breach.RemediationStatusFailed, fmt.Sprintf(
				"failed to set is_admin to false for role '%s' due to error: %s",
				b.Value, command.GetMsgFromCommandError(err)))
		} else {
			b.SetRemediation(breach.RemediationStatusSuccess, fmt.Sprintf(
				"Fixed disallowed admin setting for role [%s]", b.Value))
		}
	}
}
