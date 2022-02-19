package shipshape

import (
	"salsadigitalauorg/shipshape/pkg/core"
	"salsadigitalauorg/shipshape/pkg/drupal"
)

var AllChecks = []core.CheckType{
	core.File,
	drupal.DBConfig,
	drupal.FileConfig,
	drupal.Modules,
	drupal.ActiveModules,
}

type CheckMap map[core.CheckType][]core.Check

type Config struct {
	ProjectDir string   `yaml:"project-dir"`
	Checks     CheckMap `yaml:"checks"`
}
