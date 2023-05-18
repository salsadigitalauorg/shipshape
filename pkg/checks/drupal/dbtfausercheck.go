package drupal

import (
	"encoding/json"
	"fmt"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
//	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

// DbUserTfaCheck fetches a list of users and checks that they
// have TFA configured.
type DbUserTfaCheck struct {
	config.CheckBase `yaml:",inline"`
	DrushCommand     `yaml:",inline"`
}

type User struct {
	UID string `json:"uid"`
	Name string `json:"name"`
}

// Init implementation for the drush-based check.
func (c *DbUserTfaCheck) Init(ct config.CheckType) {
	c.CheckBase.Init(ct)
	c.RequiresDb = true
}

func (c *DbUserTfaCheck) FetchData() {
	cmd := []string{"ev", "return \\Drupal::database()->query(\"SELECT users.uid, users_field_data.name FROM users LEFT JOIN users_field_data ON users.uid = users_field_data.uid WHERE users.uid != '0' AND NOT EXISTS (SELECT 1 FROM users_data WHERE users.uid = users_data.uid AND users_data.module = 'tfa');\")->fetchAll()", "--format=json"}
	result, err := Drush(c.DrushPath, c.Alias, cmd).Exec()
	if err != nil {
		c.AddFail(err.Error())
	}
	c.DataMap = map[string][]byte{}
	c.DataMap["db-tfa-check"] = result
}

func (c *DbUserTfaCheck) RunCheck() {
	var users []User
	err := json.Unmarshal(c.DataMap["db-tfa-check"], &users)
	if err != nil {
		fmt.Println(err)
	}

	if len(users) == 0 {
		c.AddPass("All users have TFA enabled.")
		c.Result.Status = config.Pass
	} else {
		for _, user := range users {
			c.AddFail(fmt.Sprintf("Found UID %s with name %s without TFA.", user.UID, user.Name))
		}
		c.Result.Status = config.Fail
	}
}


func (c *DbUserTfaCheck) Merge(mergeCheck config.Check) error {
	// not implemented yet
	return nil
}