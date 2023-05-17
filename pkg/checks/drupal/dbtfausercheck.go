package drupal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os/exec"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

const DbUserTfa config.CheckType = "drupal-tfa-user"

// DbUserTfaCheck fetches a list of users and checks that they
// have TFA configured.
type DbUserTfaCheck struct {
	config.CheckBase `yaml:",inline"`
	DrushCommand     `yaml:",inline"`
}

// Init implementation for the drush-based check.
func (c *AdminUserCheck) Init(ct config.CheckType) {
	c.CheckBase.Init(ct)
	c.RequiresDb = true
}

func (c *AdminUserCheck) FetchData() {

	var err error
	
	cmd := []string{"ev", "return \\Drupal::database()->query(\"SELECT users.uid FROM users WHERE users.uid != '0' AND NOT EXISTS (SELECT 1 FROM users_data WHERE users.uid = users_data.uid AND users_data.module = 'tfa');\")->fetchAll()", "--format=json"}
}