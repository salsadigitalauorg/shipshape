package drupal

import (
	"fmt"
	"sort"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
	"gopkg.in/yaml.v3"
)

// Init implementation for the DB-based permissions check.
func (c *DbPermissionsCheck) Init(ct shipshape.CheckType) {
	c.CheckBase.Init(ct)
	c.RequiresDb = true
	c.Command = "role:list"
	c.ConfigName = "permissions"
}

// Merge implementation for DbPermissionsCheck check.
func (c *DbPermissionsCheck) Merge(mergeCheck shipshape.Check) error {
	dbPermissionsMergeCheck := mergeCheck.(*DbPermissionsCheck)
	if err := c.DrushYamlCheck.Merge(&dbPermissionsMergeCheck.DrushYamlCheck); err != nil {
		return err
	}

	utils.MergeStringSlice(&c.Disallowed, dbPermissionsMergeCheck.Disallowed)
	utils.MergeStringSlice(&c.ExcludeRoles, dbPermissionsMergeCheck.ExcludeRoles)
	return nil
}

// UnmarshalDataMap parses the drush permissions yaml into the DrushRoles
// type for further processing.
func (c *DbPermissionsCheck) UnmarshalDataMap() {
	if len(c.DataMap[c.ConfigName]) == 0 {
		c.AddFail("no data provided")
	}

	c.Permissions = map[string]DrushRole{}
	yaml.Unmarshal(c.DataMap[c.ConfigName], &c.Permissions)
}

// RunCheck implements the Check logic for Drupal Permissions in database config.
func (c *DbPermissionsCheck) RunCheck() {
	if len(c.Disallowed) == 0 {
		c.AddFail("list of disallowed perms not provided")
	}

	for r, perms := range c.Permissions {
		if utils.StringSliceContains(c.ExcludeRoles, r) {
			continue
		}

		fails := utils.StringSlicesIntersect(perms.Perms, c.Disallowed)
		if len(fails) == 0 {
			c.AddPass(fmt.Sprintf("[%s] no disallowed permissions", r))
			continue
		}

		// Sort fails.
		sort.Slice(fails, func(i int, j int) bool {
			return fails[i] < fails[j]
		})
		c.AddFail(fmt.Sprintf("[%s] disallowed permissions: [%s]", r, strings.Join(fails, ", ")))
	}

	if len(c.Result.Failures) == 0 {
		c.Result.Status = shipshape.Pass
	}
}
