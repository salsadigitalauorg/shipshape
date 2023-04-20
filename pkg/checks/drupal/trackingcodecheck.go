package drupal

import (
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
	"gopkg.in/yaml.v3"
)

func (c *TrackingCodeCheck) Init(ct config.CheckType) {
	c.CheckBase.Init(ct)
	c.RequiresDb = true
	c.Command = "status"
	c.ConfigName = "uri"
}

// Merge implementation for DbModuleCheck check.
func (c *TrackingCodeCheck) Merge(mergeCheck config.Check) error {
	trackingCodeMergeCheck := mergeCheck.(*TrackingCodeCheck)
	if err := c.DrushYamlCheck.Merge(&trackingCodeMergeCheck.DrushYamlCheck); err != nil {
		return err
	}

	utils.MergeString(&c.Code, trackingCodeMergeCheck.Code)
	utils.MergeString(&c.DrushStatus.Uri, trackingCodeMergeCheck.DrushStatus.Uri)
	return nil
}

// UnmarshalDataMap parses the drush status yaml into the DrushStatus
// type for further processing.
func (c *TrackingCodeCheck) UnmarshalDataMap() {
	if len(c.DataMap[c.ConfigName]) == 0 {
		c.AddFail("no data provided")
	}

	c.DrushStatus = DrushStatus{}
	err := yaml.Unmarshal(c.DataMap[c.ConfigName], &c.DrushStatus)
	if err != nil {
		if _, ok := err.(*yaml.TypeError); !ok {
			c.AddFail(err.Error())
			return
		}
	}
}

func (c *TrackingCodeCheck) RunCheck() {
	resp, err := http.Get(c.DrushStatus.Uri)

	if err != nil {
		c.AddFail("could not determine site uri")
		return
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	r, _ := regexp.Compile(c.Code)

	if r.Match(body) {
		c.AddPass(fmt.Sprintf("tracking code [%s] present", c.Code))
		c.Result.Status = result.Pass
	} else {
		c.AddFail(fmt.Sprintf("tracking code [%s] not present", c.Code))
	}

}
