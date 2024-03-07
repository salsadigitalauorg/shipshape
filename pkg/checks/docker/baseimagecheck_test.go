package docker_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/checks/docker"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
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
	assert.Equal(result.Pass, c.Result.Status)
	assert.EqualValues(
		[]string{"service1 is using valid base images"},
		c.Result.Passes,
	)
}

func TestInvalidDockerfileCheck(t *testing.T) {
	assert := assert.New(t)
	c := docker.BaseImageCheck{
		Allowed: []string{"bitnami/redis@latest"},
		Paths:   []string{"./fixtures/compose-dockerfile"},
	}
	c.RunCheck()
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues(
		[]breach.Breach{&breach.KeyValueBreach{
			BreachType: breach.BreachTypeKeyValue,
			KeyLabel:   "service",
			Key:        "service1",
			ValueLabel: "invalid base image",
			Value:      "bitnami/kubectl"},
		},
		c.Result.Breaches,
	)
}

func TestValidDockerfileImageVersion(t *testing.T) {
	assert := assert.New(t)
	c := docker.BaseImageCheck{
		Allowed: []string{"bitnami/kubectl@1.24"},
		Paths:   []string{"./fixtures/compose-dockerfile"},
	}
	c.RunCheck()
	assert.Equal(result.Pass, c.Result.Status)
	assert.EqualValues(
		[]string{"service1 is using valid base images"},
		c.Result.Passes,
	)
}

func TestInvalidDockerfileImageVersion(t *testing.T) {
	assert := assert.New(t)
	c := docker.BaseImageCheck{
		Allowed: []string{"bitnami/kubectl:1.26"},
		Paths:   []string{"./fixtures/compose-dockerfile"},
	}
	c.RunCheck()
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues(
		[]breach.Breach{&breach.KeyValueBreach{
			BreachType: breach.BreachTypeKeyValue,
			KeyLabel:   "service",
			Key:        "service1",
			ValueLabel: "invalid base image",
			Value:      "bitnami/kubectl"},
		},
		c.Result.Breaches,
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
	assert.Equal(result.Pass, c.Result.Status)
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

func TestValidImageVersions(t *testing.T) {
	assert := assert.New(t)
	c := docker.BaseImageCheck{
		Allowed: []string{
			"bitnami/kubectl@1.25.12-debian-11-r6",
			"bitnami/postgresql:15",
			"bitnami/redis",
			"bitnami/mongodb@latest",
		},
		Paths: []string{"./fixtures/compose-image"},
	}
	c.RunCheck()
	assert.Equal(result.Pass, c.Result.Status)
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
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues(
		[]breach.Breach{&breach.KeyValueBreach{
			BreachType: breach.BreachTypeKeyValue,
			KeyLabel:   "service",
			Key:        "service4",
			ValueLabel: "invalid base image",
			Value:      "bitnami/mongodb:5.0.19-debian-11-r11"},
		},
		c.Result.Breaches,
	)
}

func TestInvalidImageVersions(t *testing.T) {
	assert := assert.New(t)
	c := docker.BaseImageCheck{
		Allowed: []string{
			"bitnami/kubectl@latest",
			"bitnami/postgresql:17",
			"bitnami/redis",
		},
		Paths: []string{"./fixtures/compose-image"},
	}
	c.RunCheck()
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Fail, c.Result.Status)
	assert.ElementsMatch(
		[]breach.Breach{
			&breach.KeyValueBreach{
				BreachType: breach.BreachTypeKeyValue,
				KeyLabel:   "service",
				Key:        "service2",
				ValueLabel: "invalid base image",
				Value:      "bitnami/postgresql@16",
			},
			&breach.KeyValueBreach{
				BreachType: breach.BreachTypeKeyValue,
				KeyLabel:   "service",
				Key:        "service4",
				ValueLabel: "invalid base image",
				Value:      "bitnami/mongodb:5.0.19-debian-11-r11",
			},
		},
		c.Result.Breaches,
	)
}

func TestDockerfileWarning(t *testing.T) {
	assert := assert.New(t)
	c := docker.BaseImageCheck{
		Allowed: []string{"bitnami/kubectl"},
		Paths:   []string{"./fixtures/compose-dockerfile-missing"},
	}
	c.RunCheck()
	assert.Equal(result.Pass, c.Result.Status)
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
	assert.Equal(result.Pass, c.Result.Status)
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
	assert.Equal(result.Pass, c.Result.Status)
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
	assert.Equal(result.Pass, c.Result.Status)
	assert.EqualValues(
		[]string{"service1 is using deprecated image bitnami/kubectl"},
		c.Result.Warnings,
	)
}
