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
		Disallowed: []string{"drupal"},
		Threshold:  1,
		Paths:      []string{"./fixtures/drupal"},
	}
	c.RunCheck()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues(
		[]string{"Drupal detected at ./fixtures/drupal"},
		c.Result.Failures,
	)
}

func TestThreshold(t *testing.T) {
	assert := assert.New(t)
	c := sca.AppTypeCheck{
		Disallowed: []string{"drupal"},
		Threshold:  100,
		Paths:      []string{"./fixtures/drupal"},
	}
	c.RunCheck()
	assert.Equal(shipshape.Pass, c.Result.Status)
}

func TestIsWordpressCheck(t *testing.T) {
	assert := assert.New(t)
	c := sca.AppTypeCheck{
		Disallowed: []string{"wordpress"},
		Threshold:  1,
		Paths:      []string{"./fixtures/wordpress"},
	}
	c.RunCheck()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues(
		[]string{"Wordpress detected at ./fixtures/wordpress"},
		c.Result.Failures,
	)
}

func TestIsSymfonyCheck(t *testing.T) {
	assert := assert.New(t)
	c := sca.AppTypeCheck{
		Disallowed: []string{"symfony"},
		Threshold:  1,
		Paths:      []string{"./fixtures/symfony"},
	}
	c.RunCheck()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues(
		[]string{"Symfony detected at ./fixtures/symfony"},
		c.Result.Failures,
	)
}

func TestIsLaravelCheck(t *testing.T) {
	assert := assert.New(t)
	c := sca.AppTypeCheck{
		Disallowed: []string{"laravel"},
		Threshold:  1,
		Paths:      []string{"./fixtures/laravel"},
	}
	c.RunCheck()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues(
		[]string{"Laravel detected at ./fixtures/laravel"},
		c.Result.Failures,
	)
}

func TestOK(t *testing.T) {
	assert := assert.New(t)
	c := sca.AppTypeCheck{
		Disallowed: []string{"laravel", "wordpress", "symfony"},
		Threshold:  1,
		Paths:      []string{"./fixtures/drupal"},
	}
	c.RunCheck()
	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.EqualValues(
		[]string{"No invalid application types detected"},
		c.Result.Passes,
	)
}
