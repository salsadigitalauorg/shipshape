package drupal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os/exec"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
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
		c.AddBreach(result.ValueBreach{
			Value: pathError.Path + ": " + pathError.Err.Error()})
	} else if err != nil {
		msg := string(err.(*exec.ExitError).Stderr)
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
			c.AddBreach(result.ValueBreach{Value: err.Error()})
		}

		if userStatusMap[c.UserId]["user_status"] == "1" {
			return true
		}
	}

	return false
}

// Remediate attempts to block an active forbidden user.
func (c *ForbiddenUserCheck) Remediate() {
	for _, b := range c.Result.Breaches {
		if _, ok := b.(result.KeyValueBreach); !ok {
			continue
		}

		_, err := Drush(c.DrushPath, c.Alias, []string{"user:block", "--uid=" + c.UserId}).Exec()
		if err != nil {
			c.AddBreach(result.KeyValueBreach{
				KeyLabel:   "user",
				Key:        c.UserId,
				ValueLabel: "error blocking forbidden user",
				Value:      command.GetMsgFromCommandError(err),
			})
		} else {
			c.AddRemediation(fmt.Sprintf("Blocked the forbidden user [%s]", c.UserId))
		}
	}

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
		c.AddBreach(result.KeyValueBreach{
			Key:   "forbidden user is active",
			Value: c.UserId,
		})
	}

	if len(c.Result.Breaches) == 0 {
		c.Result.Status = result.Pass
		c.AddPass("No forbidden user is active.")
	}
}
