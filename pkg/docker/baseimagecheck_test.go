package docker_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/docker"
	"github.com/stretchr/testify/assert"
)

func TestBaseImageMerge(t *testing.T) {
	assert := assert.New(t)

	c := docker.BaseImageCheck{
		Allowed:    []string{"allowed1"},
		Exclude:    []string{"excluded1"},
		Deprecated: []string{"depr1"},
		Paths:      []string{"path1"},
	}
	c.Merge(&docker.BaseImageCheck{
		Allowed: []string{"allowed2"},
		Exclude: []string{"excluded2"},
		Pattern: []string{"patt2"},
		Paths:   []string{"path2"},
	})
	assert.EqualValues(docker.BaseImageCheck{
		Allowed:    []string{"allowed2"},
		Exclude:    []string{"excluded2"},
		Deprecated: []string{"depr1"},
		Pattern:    []string{"patt2"},
		Paths:      []string{"path2"},
	}, c)
}

func TestDockerfileCheck(t *testing.T) {
	assert := assert.New(t)
	c := docker.BaseImageCheck{
		Allowed: []string{"bitnami/kubectl"},
		Paths:   []string{"./fixtures/compose-dockerfile"},
	}
	c.RunCheck()
	assert.Equal(config.Pass, c.Result.Status)
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
	assert.Equal(config.Fail, c.Result.Status)
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
	assert.Equal(config.Pass, c.Result.Status)
	assert.ElementsMatch(
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
	assert.Equal(config.Fail, c.Result.Status)
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
	assert.Equal(config.Pass, c.Result.Status)
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
	assert.Equal(config.Pass, c.Result.Status)
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
	assert.Equal(config.Pass, c.Result.Status)
	assert.ElementsMatch(
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
	assert.Equal(config.Pass, c.Result.Status)
	assert.EqualValues(
		[]string{"service1 is using deprecated image bitnami/kubectl"},
		c.Result.Warnings,
	)
}
