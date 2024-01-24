package sca

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

const AppType config.CheckType = "sca:application_type"

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
	config.CheckBase `yaml:",inline"`
	Threshold        int                 `yaml:"threshold"`
	Disallowed       []string            `yaml:"disallowed"`
	Entrypoint       string              `yaml:"entrypoint"`
	Paths            []string            `yaml:"paths"`
	Markers          map[string][]string `yaml:"markers"`
	Dirs             map[string][]string `yaml:"dirs"`
	Dependencies     map[string][]string `yaml:"dependencies"`
}

func (c *AppTypeCheck) Merge(mergeCheck config.Check) error {
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

		disallowedFound := []string{}
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
				disallowedFound = append(disallowedFound, framework)
			}
		}
		if len(disallowedFound) > 0 {
			c.AddBreach(result.KeyValueBreach{
				Key:   fmt.Sprintf("[%s] contains disallowed frameworks", path),
				Value: "[" + strings.Join(disallowedFound, ", ") + "]",
			})
		}
	}
	if len(c.Result.Breaches) == 0 {
		c.AddPass("No invalid application types detected")
		c.Result.Status = result.Pass
	}
}
