package drupal

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

// DbUserTfaCheck fetches a list of users and checks that they
// have TFA configured.
type DbUserTfaCheck struct {
	config.CheckBase `yaml:",inline"`
	DrushCommand     `yaml:",inline"`
}

type User struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

// Init implementation for the drush-based check.
func (c *DbUserTfaCheck) Init(ct config.CheckType) {
	c.CheckBase.Init(ct)
	c.RequiresDb = true
}

// FetchData runs the Drush command to extract user information from the Drupal database.
func (c *DbUserTfaCheck) FetchData() {
	cmd := []string{
		"ev",
		`return \\Drupal::database()->query(
			\"SELECT users.uid, users_field_data.name
				FROM users
				LEFT JOIN users_field_data
					ON users.uid = users_field_data.uid
				WHERE users.uid != '0'
					AND users_field_data.status = '1'
					AND NOT EXISTS (
						SELECT 1
						FROM users_data
						WHERE users.uid = users_data.uid
						 	AND users_data.module = 'tfa');\")->fetchAll()`,
		"--format=json"}
	res, err := Drush(c.DrushPath, c.Alias, cmd).Exec()
	if err != nil {
		c.Result.Status = result.Fail
		c.AddBreach(&breach.ValueBreach{
			ValueLabel: "error fetching drush user info",
			Value:      command.GetMsgFromCommandError(err),
		})
	}
	c.DataMap = map[string][]byte{}
	c.DataMap["db-tfa-check"] = res
}

// RunCheck checks to see if any results were returned from the Drupal database query.
func (c *DbUserTfaCheck) RunCheck() {
	var users []User
	err := json.Unmarshal(c.DataMap["db-tfa-check"], &users)
	if err != nil {
		fmt.Println(err)
	}

	if len(users) == 0 {
		c.AddPass("All active users have two-factor authentication enabled.")
		c.Result.Status = result.Pass
	} else {
		tfaDisabled := []string{}
		for _, user := range users {
			tfaDisabled = append(tfaDisabled, fmt.Sprintf("%s:%s", user.Name, user.UID))
		}
		c.AddBreach(&breach.ValueBreach{
			ValueLabel: "users with TFA disabled",
			Value:      strings.Join(tfaDisabled, ", "),
		})
		c.Result.Status = result.Fail
	}
}

// Merge implmentation for DbUserTfaCheck check.
func (c *DbUserTfaCheck) Merge(mergeCheck config.Check) error {
	return nil
}
