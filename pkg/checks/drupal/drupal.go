package drupal

import (
	"fmt"

	"github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
)

//go:generate go run ../../../cmd/gen.go registry --checkpackage=drupal

func RegisterChecks() {
	config.ChecksRegistry[DrushYaml] = func() config.Check { return &DrushYamlCheck{} }
	config.ChecksRegistry[FileModule] = func() config.Check { return &FileModuleCheck{} }
	config.ChecksRegistry[DbModule] = func() config.Check { return &DbModuleCheck{} }
	config.ChecksRegistry[DbPermissions] = func() config.Check { return &DbPermissionsCheck{} }
	config.ChecksRegistry[TrackingCode] = func() config.Check { return &TrackingCodeCheck{} }
	config.ChecksRegistry[UserRole] = func() config.Check { return &UserRoleCheck{} }
	config.ChecksRegistry[AdminUser] = func() config.Check { return &AdminUserCheck{} }
}

func init() {
	RegisterChecks()
}

// CheckModulesInYaml applies the Check logic for Drupal Modules in yaml content.
// It uses YamlBase to verify that the list of provided Required or
// Disallowed modules are installed or not.
var CheckModulesInYaml = func(c *yaml.YamlBase, ct config.CheckType, configName string, required []string, disallowed []string) {
	moduleKey := func(m string) yaml.KeyValue {
		if ct == FileModule {
			return yaml.KeyValue{
				Key:   "module." + m,
				Value: "0",
			}
		}
		return yaml.KeyValue{
			Key:   m + ".status",
			Value: "Enabled",
		}
	}

	for _, m := range required {
		kvr, _, err := c.CheckKeyValue(moduleKey(m), configName)
		// It could be a value different from 0, which still means it's enabled.
		if kvr == yaml.KeyValueEqual || kvr == yaml.KeyValueNotEqual {
			c.AddPass(fmt.Sprintf("'%s' is enabled", m))
		} else if kvr == yaml.KeyValueError {
			c.AddFail(err.Error())
		} else {
			c.AddFail(fmt.Sprintf("'%s' is not enabled", m))
		}
	}
	for _, m := range disallowed {
		kvr, _, err := c.CheckKeyValue(moduleKey(m), configName)
		// It could be a value different from 0, which still means it's enabled.
		if kvr == yaml.KeyValueEqual || kvr == yaml.KeyValueNotEqual {
			c.AddFail(fmt.Sprintf("'%s' is enabled", m))
		} else if kvr == yaml.KeyValueError {
			c.AddFail(err.Error())
		} else {
			c.AddPass(fmt.Sprintf("'%s' is not enabled", m))
		}
	}

	if len(c.Result.Failures) == 0 {
		c.Result.Status = config.Pass
	}
}
