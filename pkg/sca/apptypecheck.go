package sca

import (
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

const AppType shipshape.CheckType = "sca:application_type"

type AppTypeCheck struct {
	shipshape.CheckBase `yaml:",inline"`
	Disallowed          []string `yaml:"disallowed"`
	Paths               []string `yaml:"paths"`
	Threshold           int      `yaml:"threshold"`
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
	for _, path := range c.Paths {
		if utils.StringSliceContains(c.Disallowed, "wordpress") && isWordpress(path) > c.Threshold {
			c.AddFail("Wordpress detected at " + path)
		}
		if utils.StringSliceContains(c.Disallowed, "symfony") && isSymfony(path) > c.Threshold {
			c.AddFail("Symfony detected at " + path)
		}
		if utils.StringSliceContains(c.Disallowed, "laravel") && isLaravel(path) > c.Threshold {
			c.AddFail("Laravel detected at " + path)
		}
		if utils.StringSliceContains(c.Disallowed, "drupal") && isDrupal(path) > c.Threshold {
			c.AddFail("Drupal detected at " + path)
		}
	}
	if len(c.Result.Failures) == 0 {
		c.AddPass("No invalid application types detected")
		c.Result.Status = shipshape.Pass
	}
}

func isWordpress(path string) int {
	likelihood := 0

	marker := " * @package WordPress"
	paths, _ := utils.Glob(path, "index.php")
	for _, index := range paths {
		if f, _ := utils.FileContains(index, marker); f {
			likelihood += 10
		}
	}

	dirs := []string{"wp-admin", "wp-content", "wp-includes"}
	for _, dir := range dirs {
		if d, _ := utils.IsDirectory(dir); d {
			likelihood += 10
		}
	}

	// Wordpress is not detected.
	return likelihood
}

func isDrupal(path string) int {
	likelihood := 0

	paths, _ := utils.Glob(path, "index.php")
	marker := "use Drupal\\Core\\DrupalKernel;"
	for _, index := range paths {
		if f, _ := utils.FileContains(index, marker); f {
			likelihood += 10
		}
	}

	deps := []string{"drupal/core-recommended"}
	if hasDep, _ := utils.HasComposerDependency(path, deps); hasDep {
		likelihood += 10
	}

	dirs := []string{"web", "docroot"}
	for _, dir := range dirs {
		if d, _ := utils.IsDirectory(dir); d {
			likelihood += 5
		}
	}

	return likelihood
}

func isSymfony(path string) int {
	likelihood := 0
	marker := "    return new Kernel($context['APP_ENV'], (bool) $context['APP_DEBUG']);"
	paths, _ := utils.Glob(path, "index.php")
	for _, index := range paths {
		if f, _ := utils.FileContains(index, marker); f {
			likelihood += 10
		}
	}

	// Check composer.json
	deps := []string{"symfony/runtime", "symfony/symfony", "symfony/framework"}
	if hasDep, _ := utils.HasComposerDependency(path, deps); hasDep {
		likelihood += 10
	}

	dirs := []string{"bin", "config", "public"}
	for _, dir := range dirs {
		if d, _ := utils.IsDirectory(dir); d {
			likelihood += 5
		}
	}

	return likelihood
}

func isLaravel(path string) int {
	likelihood := 0
	marker := "use Illuminate\\Contracts\\Http\\Kernel;"
	paths, _ := utils.Glob(path, "index.php")
	for _, index := range paths {
		if f, _ := utils.FileContains(index, marker); f {
			likelihood += 10
		}
	}

	deps := []string{"laravel/framework", "laravel/tinker"}
	if hasDep, _ := utils.HasComposerDependency(path, deps); hasDep {
		likelihood += 10
	}

	dirs := []string{"bin", "config", "public"}
	for _, dir := range dirs {
		if d, _ := utils.IsDirectory(dir); d {
			likelihood += 5
		}
	}

	return likelihood
}
