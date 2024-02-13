package drupal

import (
	"fmt"
	"strings"

	yamlv3 "gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

//go:generate go run ../../../cmd/gen.go registry --checkpackage=drupal

func RegisterChecks() {
	config.ChecksRegistry[DrushYaml] = func() config.Check { return &DrushYamlCheck{} }
	config.ChecksRegistry[FileModule] = func() config.Check { return &FileModuleCheck{} }
	config.ChecksRegistry[DbModule] = func() config.Check { return &DbModuleCheck{} }
	config.ChecksRegistry[DbPermissions] = func() config.Check { return &DbPermissionsCheck{} }
	config.ChecksRegistry[RolePermissions] = func() config.Check { return &RolePermissionsCheck{} }
	config.ChecksRegistry[TrackingCode] = func() config.Check { return &TrackingCodeCheck{} }
	config.ChecksRegistry[UserRole] = func() config.Check { return &UserRoleCheck{} }
	config.ChecksRegistry[AdminUser] = func() config.Check { return &AdminUserCheck{} }
	config.ChecksRegistry[DbUserTfa] = func() config.Check { return &DbUserTfaCheck{} }
	config.ChecksRegistry[ForbiddenUser] = func() config.Check { return &ForbiddenUserCheck{} }
}

func init() {
	RegisterChecks()
}

func ModuleYamlKey(ct config.CheckType, m string) yaml.KeyValue {
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

func DetermineModuleStatus(node yamlv3.Node, ct config.CheckType, modules []string) ([]string, []string, []string) {
	enabled := []string{}
	errored := []string{}
	disabled := []string{}

	for _, m := range modules {
		kvr, _, err := yaml.CheckKeyValue(node, ModuleYamlKey(ct, m))
		// It could be a value different from 0, which still means it's enabled.
		if ct == FileModule && (kvr == yaml.KeyValueEqual || kvr == yaml.KeyValueNotEqual) {
			enabled = append(enabled, m)
		} else if kvr == yaml.KeyValueEqual {
			enabled = append(enabled, m)
		} else if kvr == yaml.KeyValueError {
			errored = append(errored, err.Error())
		} else {
			disabled = append(disabled, m)
		}
	}

	return enabled, errored, disabled
}

// CheckModulesInYaml applies the Check logic for Drupal Modules in yaml content.
// It uses YamlBase to verify that the list of provided Required or
// Disallowed modules are installed or not.
func CheckModulesInYaml(c *yaml.YamlBase, ct config.CheckType, configName string, required []string, disallowed []string) {
	required_enabled,
		required_errored,
		required_disabled := DetermineModuleStatus(c.NodeMap[configName], ct, required)
	if len(required_errored) > 0 {
		c.AddBreach(&result.KeyValuesBreach{
			Key:    "error verifying status for required modules",
			Values: required_errored})
	}
	if len(required_disabled) > 0 {
		c.AddBreach(&result.KeyValuesBreach{
			Key:    "required modules are not enabled",
			Values: required_disabled})
	}
	if len(required_enabled) == len(required) {
		c.AddPass("all required modules are enabled")
	} else if len(required_enabled) > 0 {
		c.AddPass(fmt.Sprint(
			"some required modules are enabled: ",
			strings.Join(required_enabled, ",")))
	}

	disallowed_enabled,
		disallowed_errored,
		disallowed_disabled := DetermineModuleStatus(c.NodeMap[configName], ct, disallowed)
	if len(disallowed_errored) > 0 {
		c.AddBreach(&result.KeyValuesBreach{
			Key:    "error verifying status for disallowed modules",
			Values: disallowed_errored})
	}
	if len(disallowed_enabled) > 0 {
		c.AddBreach(&result.KeyValuesBreach{
			Key:    "disallowed modules are enabled",
			Values: disallowed_enabled})
	}
	if len(disallowed_disabled) == len(required) {
		c.AddPass("all disallowed modules are disabled")
	} else if len(disallowed_disabled) > 0 {
		c.AddPass(fmt.Sprint(
			"some disallowed modules are disabled: ",
			strings.Join(disallowed_disabled, ",")))
	}

	if len(c.Result.Breaches) == 0 {
		c.Result.Status = result.Pass
	}
}
