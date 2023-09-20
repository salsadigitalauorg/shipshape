package drupal

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"io/fs"
	"os/exec"
	"strings"
)

const ForbiddenUser config.CheckType = "drupal-user-forbidden"

// ForbiddenUserCheck checks whether a forbidden user is active.
type ForbiddenUserCheck struct {
	config.CheckBase `yaml:",inline"`
	DrushCommand     `yaml:",inline"`
	UserId           string `yaml:"uid"`
}

// Init implementation for the drush-based user status check.
func (c *ForbiddenUserCheck) Init(ct config.CheckType) {
	c.CheckBase.Init(ct)
	c.RequiresDb = true
	// Default to User 1.
	if c.UserId == "" {
		c.UserId = "1"
	}
}

// CheckUserStatus check the status of a forbidden user.
func (c *ForbiddenUserCheck) CheckUserStatus() bool {
	// Command: drush user:info --uid=1 --fields=user_status --format=json
	cmd := []string{"user:info", "--uid=" + c.UserId, "--fields=user_status", "--format=json"}

	userStatus, err := Drush(c.DrushPath, c.Alias, cmd).Exec()
	var pathError *fs.PathError
	if err != nil && errors.As(err, &pathError) {
		c.AddFail(pathError.Path + ": " + pathError.Err.Error())
		c.AddBreach(result.ValueBreach{
			Value: pathError.Path + ": " + pathError.Err.Error()})
	} else if err != nil {
		msg := string(err.(*exec.ExitError).Stderr)
		c.AddFail(strings.ReplaceAll(strings.TrimSpace(msg), "  \n  ", ""))
		c.AddBreach(result.ValueBreach{
			Value: strings.ReplaceAll(strings.TrimSpace(msg), "  \n  ", "")})
	} else {
		// Unmarshal user:info JSON.
		// {
		//	 "1": {
		//		 "user_status": "1"
		//	 }
		// }
		userStatusMap := map[string]map[string]string{}
		err = json.Unmarshal(userStatus, &userStatusMap)
		var syntaxError *json.SyntaxError
		if err != nil && errors.As(err, &syntaxError) {
			c.AddFail(err.Error())
			c.AddBreach(result.ValueBreach{Value: err.Error()})
		}

		if userStatusMap[c.UserId]["user_status"] == "1" {
			return true
		}
	}

	return false
}

// Remediate attempts to block an active forbidden user.
func (c *ForbiddenUserCheck) Remediate(breachIfc interface{}) error {
	_, err := Drush(c.DrushPath, c.Alias, []string{"user:block", "--uid=" + c.UserId}).Exec()
	if err != nil {
		c.AddFail(fmt.Sprintf(
			"Failed to block the active forbidden user [%s] due to error: %s",
			c.UserId, command.GetMsgFromCommandError(err)))
		return err
	}
	return nil
}

// Merge implementation for ForbiddenUserCheck check.
func (c *ForbiddenUserCheck) Merge(mergeCheck config.Check) error {
	return nil
}

// HasData implementation for ForbiddenUserCheck check.
func (c *ForbiddenUserCheck) HasData(failCheck bool) bool {
	return true
}

// RunCheck applies the Check logic for active forbidden users.
func (c *ForbiddenUserCheck) RunCheck() {
	userActive := c.CheckUserStatus()
	if userActive {
		if c.PerformRemediation {
			if err := c.Remediate(nil); err == nil {
				c.AddRemediation(fmt.Sprintf("Blocked the forbidden user [%s]", c.UserId))
			}
		} else {
			c.AddFail(fmt.Sprintf("Forbidden user [%s] is active", c.UserId))
			c.AddBreach(result.KeyValueBreach{
				Key:   "uid",
				Value: c.UserId,
			})
		}
	}

	if len(c.Result.Failures) == 0 {
		c.Result.Status = result.Pass
		c.AddPass("No forbidden user is active.")
	}
}
