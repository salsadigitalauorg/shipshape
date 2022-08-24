package drupal

import (
	"fmt"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
)

func RegisterChecks() {
	shipshape.ChecksRegistry[DrushYaml] = func() shipshape.Check { return &DrushYamlCheck{} }
	shipshape.ChecksRegistry[FileModule] = func() shipshape.Check { return &FileModuleCheck{} }
	shipshape.ChecksRegistry[DbModule] = func() shipshape.Check { return &DbModuleCheck{} }
	shipshape.ChecksRegistry[DbPermissions] = func() shipshape.Check { return &DbPermissionsCheck{} }
	shipshape.ChecksRegistry[TrackingCode] = func() shipshape.Check { return &TrackingCodeCheck{} }
	shipshape.ChecksRegistry[UserRole] = func() shipshape.Check { return &UserRoleCheck{} }
}

func init() {
	RegisterChecks()
}

// CheckModulesInYaml applies the Check logic for Drupal Modules in yaml content.
// It uses YamlBase to verify that the list of provided Required or
// Disallowed modules are installed or not.
func CheckModulesInYaml(c *shipshape.YamlBase, ct shipshape.CheckType, configName string, required []string, disallowed []string) {
	moduleKey := func(m string) shipshape.KeyValue {
		if ct == FileModule {
			return shipshape.KeyValue{
				Key:   "module." + m,
				Value: "0",
			}
		}
		return shipshape.KeyValue{
			Key:   m + ".status",
			Value: "Enabled",
		}
	}

	for _, m := range required {
		kvr, _, err := c.CheckKeyValue(moduleKey(m), configName)
		// It could be a value different from 0, which still means it's enabled.
		if kvr == shipshape.KeyValueEqual || kvr == shipshape.KeyValueNotEqual {
			c.AddPass(fmt.Sprintf("'%s' is enabled", m))
		} else if kvr == shipshape.KeyValueError {
			c.AddFail(err.Error())
		} else {
			c.AddFail(fmt.Sprintf("'%s' is not enabled", m))
		}
	}
	for _, m := range disallowed {
		kvr, _, err := c.CheckKeyValue(moduleKey(m), configName)
		// It could be a value different from 0, which still means it's enabled.
		if kvr == shipshape.KeyValueEqual || kvr == shipshape.KeyValueNotEqual {
			c.AddFail(fmt.Sprintf("'%s' is enabled", m))
		} else if kvr == shipshape.KeyValueError {
			c.AddFail(err.Error())
		} else {
			c.AddPass(fmt.Sprintf("'%s' is not enabled", m))
		}
	}

	if len(c.Result.Failures) == 0 {
		c.Result.Status = shipshape.Pass
	}
}

// RunCheck applies the Check logic for Drupal Modules in config files.
func (c *FileModuleCheck) RunCheck() {
	CheckModulesInYaml(&c.YamlBase, FileModule, c.File, c.Required, c.Disallowed)
}

// Init implementation for the File-based module check.
func (c *FileModuleCheck) Init(ct shipshape.CheckType) {
	c.CheckBase.Init(ct)
	c.File = "core.extension.yml"
	if c.IgnoreMissing == nil {
		cTrue := true
		c.IgnoreMissing = &cTrue
	}
}
