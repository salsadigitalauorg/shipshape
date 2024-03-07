package sca_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	. "github.com/salsadigitalauorg/shipshape/pkg/checks/sca"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

func TestIsDrupalCheck(t *testing.T) {
	assert := assert.New(t)
	c := AppTypeCheck{
		Disallowed:   []string{"drupal"},
		Threshold:    1,
		Markers:      make(map[string][]string),
		Dirs:         make(map[string][]string),
		Dependencies: make(map[string][]string),
		Paths:        []string{"./testdata/drupal"},
	}

	c.Markers["drupal"] = []string{"use Drupal\\Core\\DrupalKernel;"}
	c.Dirs["drupal"] = []string{"web", "docroot"}
	c.Dependencies["drupal"] = []string{"drupal/core-recommended"}

	c.RunCheck()
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues(
		[]breach.Breach{&breach.KeyValueBreach{
			BreachType: "key-value",
			Key:        "[./testdata/drupal] contains disallowed frameworks",
			Value:      "[drupal]",
		}},
		c.Result.Breaches,
	)
}

func TestThreshold(t *testing.T) {
	assert := assert.New(t)
	c := AppTypeCheck{
		Disallowed:   []string{"drupal"},
		Threshold:    100,
		Markers:      make(map[string][]string),
		Dirs:         make(map[string][]string),
		Dependencies: make(map[string][]string),
		Paths:        []string{"./testdata/drupal"},
	}
	c.RunCheck()
	assert.Equal(result.Pass, c.Result.Status)
}

func TestIsWordpressCheck(t *testing.T) {
	assert := assert.New(t)
	c := AppTypeCheck{
		Disallowed:   []string{"wordpress"},
		Threshold:    1,
		Markers:      make(map[string][]string),
		Dirs:         make(map[string][]string),
		Dependencies: make(map[string][]string),
		Paths:        []string{"./testdata/wordpress"},
	}

	c.Markers["wordpress"] = []string{" * @package WordPress"}
	c.Dirs["wordpress"] = []string{"wp-admin", "wp-content", "wp-includes"}

	c.RunCheck()
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues(
		[]breach.Breach{&breach.KeyValueBreach{
			BreachType: "key-value",
			Key:        "[./testdata/wordpress] contains disallowed frameworks",
			Value:      "[wordpress]",
		}},
		c.Result.Breaches,
	)
}

func TestIsSymfonyCheck(t *testing.T) {
	assert := assert.New(t)
	c := AppTypeCheck{
		Disallowed:   []string{"symfony"},
		Threshold:    1,
		Markers:      make(map[string][]string),
		Dirs:         make(map[string][]string),
		Dependencies: make(map[string][]string),
		Paths:        []string{"./testdata/symfony"},
	}
	c.Markers["symfony"] = []string{"    return new Kernel($context['APP_ENV'], (bool) $context['APP_DEBUG']);"}
	c.Dirs["symfony"] = []string{"bin", "config", "public"}
	c.Dependencies["symfony"] = []string{"symfony/runtime", "symfony/symfony", "symfony/framework"}

	c.RunCheck()
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues(
		[]breach.Breach{&breach.KeyValueBreach{
			BreachType: "key-value",
			Key:        "[./testdata/symfony] contains disallowed frameworks",
			Value:      "[symfony]",
		}},
		c.Result.Breaches,
	)
}

func TestIsLaravelCheck(t *testing.T) {
	assert := assert.New(t)
	c := AppTypeCheck{
		Disallowed:   []string{"laravel"},
		Threshold:    1,
		Markers:      make(map[string][]string),
		Dirs:         make(map[string][]string),
		Dependencies: make(map[string][]string),
		Paths:        []string{"./testdata/laravel"},
	}
	c.Markers["laravel"] = []string{"use Illuminate\\Contracts\\Http\\Kernel;"}
	c.Dirs["laravel"] = []string{"config", "public"}
	c.Dependencies["laravel"] = []string{"laravel/framework", "laravel/tinker"}

	c.RunCheck()
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues(
		[]breach.Breach{&breach.KeyValueBreach{
			BreachType: "key-value",
			Key:        "[./testdata/laravel] contains disallowed frameworks",
			Value:      "[laravel]",
		}},
		c.Result.Breaches,
	)
}

func TestMultipleTypes(t *testing.T) {
	assert := assert.New(t)
	c := AppTypeCheck{
		Disallowed:   []string{"laravel", "symfony", "wordpress"},
		Threshold:    1,
		Markers:      make(map[string][]string),
		Dirs:         make(map[string][]string),
		Dependencies: make(map[string][]string),
		Paths:        []string{"./testdata/drupal"},
	}
	c.Dirs["laravel"] = []string{"config", "public"}
	c.Dirs["symfony"] = []string{"config", "public"}
	c.Dirs["wordpress"] = []string{"wp-admin"}

	c.RunCheck()
	assert.Equal(result.Pass, c.Result.Status)
	assert.EqualValues(
		[]string{"No invalid application types detected"},
		c.Result.Passes,
	)
}

func TestOK(t *testing.T) {
	assert := assert.New(t)
	c := AppTypeCheck{
		Disallowed:   []string{"laravel", "wordpress", "symfony"},
		Threshold:    1,
		Markers:      make(map[string][]string),
		Dirs:         make(map[string][]string),
		Dependencies: make(map[string][]string),
		Paths:        []string{"./testdata/drupal"},
	}
	c.RunCheck()
	assert.Equal(result.Pass, c.Result.Status)
	assert.EqualValues(
		[]string{"No invalid application types detected"},
		c.Result.Passes,
	)
}
