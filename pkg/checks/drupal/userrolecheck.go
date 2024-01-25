package drupal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os/exec"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

const UserRole config.CheckType = "drupal-user-role"

// UserRoleCheck fetches all users from the database and verifies them against
// the list of disallowed roles and allowed users.
type UserRoleCheck struct {
	config.CheckBase `yaml:",inline"`
	DrushCommand     `yaml:",inline"`
	// List of role machine names that users should not have.
	Roles []string `yaml:"roles"`
	// List of user ID's allowed to have the above roles.
	AllowedUsers []int `yaml:"allowed-users"`
	userRoles    map[int][]string
}

type userInfo struct {
	Roles []string
}

// Init implementation for the drush-based user role check.
func (c *UserRoleCheck) Init(ct config.CheckType) {
	c.CheckBase.Init(ct)
	c.RequiresDb = true
}

// Merge implementation for DbModuleCheck check.
func (c *UserRoleCheck) Merge(mergeCheck config.Check) error {
	userRoleMergeCheck := mergeCheck.(*UserRoleCheck)
	if err := c.CheckBase.Merge(&userRoleMergeCheck.CheckBase); err != nil {
		return err
	}

	c.DrushCommand.Merge(userRoleMergeCheck.DrushCommand)
	utils.MergeStringSlice(&c.Roles, userRoleMergeCheck.Roles)
	utils.MergeIntSlice(&c.AllowedUsers, userRoleMergeCheck.AllowedUsers)
	return nil
}

func (c *UserRoleCheck) getUserIds() string {
	userIds, err := Drush(c.DrushPath, c.Alias, c.Args).Query("SELECT GROUP_CONCAT(uid) FROM users")

	var pathErr *fs.PathError
	if err != nil && errors.As(err, &pathErr) {
		c.AddBreach(result.ValueBreach{
			Value: pathErr.Path + ": " + pathErr.Err.Error()})
	} else if err != nil {
		msg := string(err.(*exec.ExitError).Stderr)
		c.AddBreach(result.ValueBreach{
			Value: strings.ReplaceAll(strings.TrimSpace(msg), "  \n  ", "")})
	}
	return string(userIds)
}

// FetchData runs the drush command to populate data for the user role check.
func (c *UserRoleCheck) FetchData() {
	var err error

	userIds := c.getUserIds()
	if c.Result.Status == result.Fail {
		return
	}

	c.DataMap = map[string][]byte{}
	cmd := []string{"user:information", "--uid=" + userIds, "--fields=roles", "--format=json"}
	c.DataMap["user-info"], err = Drush(c.DrushPath, c.Alias, cmd).Exec()
	if err != nil {
		msg := string(err.(*exec.ExitError).Stderr)
		c.AddBreach(result.ValueBreach{
			Value: strings.ReplaceAll(strings.TrimSpace(msg), "  \n  ", "")})
	}
}

// UnmarshalDataMap parses the drush user info json
// into the userRoles for further processing.
func (c *UserRoleCheck) UnmarshalDataMap() {
	if len(c.DataMap["user-info"]) == 0 {
		c.AddBreach(result.ValueBreach{Value: "no data provided"})
		return
	}

	userInfoMap := map[int]userInfo{}
	err := json.Unmarshal(c.DataMap["user-info"], &userInfoMap)
	var synErr *json.SyntaxError
	if err != nil && errors.As(err, &synErr) {
		c.AddBreach(result.ValueBreach{Value: err.Error()})
		return
	}

	c.userRoles = map[int][]string{}
	for uid, uinf := range userInfoMap {
		c.userRoles[uid] = uinf.Roles
	}
}

// RunCheck implements the Check logic for disallowed user roles.
func (c *UserRoleCheck) RunCheck() {
	if len(c.Roles) == 0 {
		c.AddBreach(result.ValueBreach{Value: "no disallowed role provided"})
		return
	}

	for uid, roles := range c.userRoles {
		allowedUser := utils.IntSliceContains(c.AllowedUsers, uid)
		if allowedUser {
			continue
		}

		disallowed := utils.StringSlicesIntersect(roles, c.Roles)
		if len(disallowed) > 0 {
			c.AddBreach(result.KeyValuesBreach{
				KeyLabel:   "user",
				Key:        fmt.Sprintf("%d", uid),
				ValueLabel: "disallowed roles",
				Values:     disallowed,
			})
		}
	}

	if len(c.Result.Breaches) == 0 {
		c.Result.Status = result.Pass
	}
}
