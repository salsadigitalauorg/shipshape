package shipshape

import (
	"salsadigitalauorg/shipshape/pkg/core"
	"salsadigitalauorg/shipshape/pkg/drupal"
)

var AllChecks = []core.CheckType{
	core.File,
	drupal.DrupalDBConfig,
	drupal.DrupalFileConfig,
	drupal.DrupalModules,
	drupal.DrupalActiveModules,
}

type CheckMap map[core.CheckType][]core.Check

type Config struct {
	ProjectDir string   `yaml:"project-dir"`
	DrupalRoot string   `yaml:"drupal-root"`
	Checks     CheckMap `yaml:"checks"`
}
