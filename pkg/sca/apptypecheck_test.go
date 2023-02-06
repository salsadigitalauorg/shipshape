package sca_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/sca"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestIsDrupalCheck(t *testing.T) {
	assert := assert.New(t)
	c := sca.AppTypeCheck{
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
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues(
		[]string{"drupal detected at ./testdata/drupal"},
		c.Result.Failures,
	)
}

func TestThreshold(t *testing.T) {
	assert := assert.New(t)
	c := sca.AppTypeCheck{
		Disallowed:   []string{"drupal"},
		Threshold:    100,
		Markers:      make(map[string][]string),
		Dirs:         make(map[string][]string),
		Dependencies: make(map[string][]string),
		Paths:        []string{"./testdata/drupal"},
	}
	c.RunCheck()
	assert.Equal(shipshape.Pass, c.Result.Status)
}

func TestIsWordpressCheck(t *testing.T) {
	assert := assert.New(t)
	c := sca.AppTypeCheck{
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
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues(
		[]string{"wordpress detected at ./testdata/wordpress"},
		c.Result.Failures,
	)
}

func TestIsSymfonyCheck(t *testing.T) {
	assert := assert.New(t)
	c := sca.AppTypeCheck{
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
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues(
		[]string{"symfony detected at ./testdata/symfony"},
		c.Result.Failures,
	)
}

func TestIsLaravelCheck(t *testing.T) {
	assert := assert.New(t)
	c := sca.AppTypeCheck{
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
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues(
		[]string{"laravel detected at ./testdata/laravel"},
		c.Result.Failures,
	)
}

func TestMultipleTypes(t *testing.T) {
	assert := assert.New(t)
	c := sca.AppTypeCheck{
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
	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.EqualValues(
		[]string{"No invalid application types detected"},
		c.Result.Passes,
	)
}

func TestOK(t *testing.T) {
	assert := assert.New(t)
	c := sca.AppTypeCheck{
		Disallowed:   []string{"laravel", "wordpress", "symfony"},
		Threshold:    1,
		Markers:      make(map[string][]string),
		Dirs:         make(map[string][]string),
		Dependencies: make(map[string][]string),
		Paths:        []string{"./testdata/drupal"},
	}
	c.RunCheck()
	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.EqualValues(
		[]string{"No invalid application types detected"},
		c.Result.Passes,
	)
}
