package sca

import (
	"os"
	"path/filepath"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

const AppType shipshape.CheckType = "sca:application_type"

/*
Example:
	threshold: 20
	disallowed:
		- symfyony
	markers:
		symfony:
			- "    return new Kernel($context['APP_ENV'], (bool) $context['APP_DEBUG']);"
	dirs:
		symfony:
			- bin
			- config
			- public
	dependencies:
		symfony:
			- symfony/runtime
			- symfony/symfony
			- symfony/framework

@TODO: currently only supports composer for dependency lookups.
*/
type AppTypeCheck struct {
	shipshape.CheckBase `yaml:",inline"`
	Threshold           int                 `yaml:"threshold"`
	Disallowed          []string            `yaml:"disallowed"`
	Entrypoint          string              `yaml:"entrypoint"`
	Paths               []string            `yaml:"paths"`
	Markers             map[string][]string `yaml:"markers"`
	Dirs                map[string][]string `yaml:"dirs"`
	Dependencies        map[string][]string `yaml:"dependencies"`
}

func (c *AppTypeCheck) Merge(mergeCheck shipshape.Check) error {
	appTypeCheck := mergeCheck.(*AppTypeCheck)
	if err := c.CheckBase.Merge(&appTypeCheck.CheckBase); err != nil {
		return err
	}

	utils.MergeStringSlice(&c.Disallowed, appTypeCheck.Disallowed)
	utils.MergeStringSlice(&c.Paths, appTypeCheck.Paths)
	c.Threshold = appTypeCheck.Threshold

	return nil
}

func (c *AppTypeCheck) RunCheck() {

	if c.Threshold == 0 {
		c.Threshold = 30
	}

	if c.Entrypoint == "" {
		c.Entrypoint = "index.php"
	}

	for _, path := range c.Paths {
		entrypoints, _ := utils.Glob(path, c.Entrypoint)

		for _, framework := range c.Disallowed {
			likelihood := 0
			if markers, ok := c.Markers[framework]; ok {
				for _, marker := range markers {
					for _, e := range entrypoints {
						if f, _ := utils.FileContains(e, marker); f {
							likelihood += 5
						}
					}
				}
			}

			if dirs, ok := c.Dirs[framework]; ok {
				filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
					if f.IsDir() && utils.StringSliceContains(dirs, f.Name()) {
						likelihood += 5
					}
					return nil
				})
			}

			if deps, ok := c.Dependencies[framework]; ok {
				if hasDep, _ := utils.HasComposerDependency(path, deps); hasDep {
					likelihood += 10
				}
			}

			if likelihood > c.Threshold {
				c.AddFail(framework + " detected at " + path)
			}
		}
	}
	if len(c.Result.Failures) == 0 {
		c.AddPass("No invalid application types detected")
		c.Result.Status = shipshape.Pass
	}
}
