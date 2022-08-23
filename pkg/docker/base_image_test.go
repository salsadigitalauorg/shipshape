package docker_test

import (
	"sort"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/docker"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestDockerfileCheck(t *testing.T) {
	assert := assert.New(t)
	c := docker.BaseImageCheck{
		Allowed: []string{"bitnami/kubectl"},
		Paths:   []string{"./fixtures/compose-dockerfile"},
	}
	c.RunCheck()
	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.EqualValues(
		[]string{"service1 is using valid base images"},
		c.Result.Passes,
	)
}

func TestInvalidDockerfileCheck(t *testing.T) {
	assert := assert.New(t)
	c := docker.BaseImageCheck{
		Allowed: []string{"bitnami/redis"},
		Paths:   []string{"./fixtures/compose-dockerfile"},
	}
	c.RunCheck()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues(
		[]string{"service1 is using invalid base image bitnami/kubectl"},
		c.Result.Failures,
	)
}

func TestValidImage(t *testing.T) {
	assert := assert.New(t)
	c := docker.BaseImageCheck{
		Allowed: []string{
			"bitnami/kubectl",
			"bitnami/postgresql",
			"bitnami/redis",
			"bitnami/mongodb",
		},
		Paths: []string{"./fixtures/compose-image"},
	}
	c.RunCheck()
	assert.Equal(shipshape.Pass, c.Result.Status)
	sort.Strings(c.Result.Passes)
	assert.EqualValues(
		[]string{
			"service1 is using valid base images",
			"service2 is using valid base images",
			"service3 is using valid base images",
			"service4 is using valid base images",
		},
		c.Result.Passes,
	)
}

func TestInvalidImageCheck(t *testing.T) {
	assert := assert.New(t)
	c := docker.BaseImageCheck{
		Allowed: []string{
			"bitnami/kubectl",
			"bitnami/postgresql",
			"bitnami/redis",
		},
		Paths: []string{"./fixtures/compose-image"},
	}
	c.RunCheck()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues(
		[]string{"service4 is using invalid base image bitnami/mongodb"},
		c.Result.Failures,
	)
}

func TestDockerfileWarning(t *testing.T) {
	assert := assert.New(t)
	c := docker.BaseImageCheck{
		Allowed: []string{"bitnami/kubectl"},
		Paths:   []string{"./fixtures/compose-dockerfile-missing"},
	}
	c.RunCheck()
	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.EqualValues(
		[]string{"Unable to find Dockerfile"},
		c.Result.Warnings,
	)
}

func TestExcludeServiceSingle(t *testing.T) {
	assert := assert.New(t)
	c := docker.BaseImageCheck{
		Allowed: []string{"bitnami/redis"},
		Exclude: []string{"service1"},
		Paths:   []string{"./fixtures/compose-dockerfile"},
	}
	c.RunCheck()
	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.EqualValues(
		[]string(nil),
		c.Result.Passes,
	)
}

func TestExcludeServiceMany(t *testing.T) {
	assert := assert.New(t)
	c := docker.BaseImageCheck{
		Allowed: []string{
			"bitnami/kubectl",
			"bitnami/postgresql",
			"bitnami/mongodb",
		},
		Exclude: []string{"service3"},
		Paths:   []string{"./fixtures/compose-image"},
	}
	c.RunCheck()
	assert.Equal(shipshape.Pass, c.Result.Status)
	sort.Strings(c.Result.Passes)
	assert.EqualValues(
		[]string{
			"service1 is using valid base images",
			"service2 is using valid base images",
			"service4 is using valid base images",
		},
		c.Result.Passes,
	)
}

func TestDeprecatedImage(t *testing.T) {
	assert := assert.New(t)
	c := docker.BaseImageCheck{
		Deprecated: []string{"bitnami/kubectl"},
		Paths:      []string{"./fixtures/compose-dockerfile"},
	}
	c.RunCheck()
	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.EqualValues(
		[]string{"service1 is using deprecated image bitnami/kubectl"},
		c.Result.Warnings,
	)
}
