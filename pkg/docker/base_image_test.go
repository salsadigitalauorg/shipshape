package docker_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/docker"
)

func TestDockerfileCheck(t *testing.T) {
	c := docker.BaseImageCheck{
		Allowed: []string{"bitnami/kubectl"},
		Paths:   []string{"./fixtures/compose-dockerfile"},
	}
	c.RunCheck()
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"Dockerfiles adhere to the policy"}); !ok {
		t.Error(msg)
	}
}

func TestInvalidDockerfileCheck(t *testing.T) {
	c := docker.BaseImageCheck{
		Allowed: []string{"bitnami/redis"},
		Paths:   []string{"./fixtures/compose-dockerfile"},
	}
	c.RunCheck()
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"service1 is using invalid base image FROM bitnami/kubectl"}); !ok {
		t.Error(msg)
	}
}

func TestValidImage(t *testing.T) {
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
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"Dockerfiles adhere to the policy"}); !ok {
		t.Error(msg)
	}
}

func TestInvalidImageCheck(t *testing.T) {
	c := docker.BaseImageCheck{
		Allowed: []string{
			"bitnami/kubectl",
			"bitnami/postgresql",
			"bitnami/redis",
		},
		Paths: []string{"./fixtures/compose-image"},
	}
	c.RunCheck()

	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"service4 is using an invalid base image bitnami/mongodb"}); !ok {
		t.Error(msg)
	}
}

func TestDockerfileWarning(t *testing.T) {
	c := docker.BaseImageCheck{
		Allowed: []string{"bitnami/kubectl"},
		Paths:   []string{"./fixtures/compose-dockerfile-missing"},
	}
	c.RunCheck()
	if msg, ok := internal.EnsureWarnings(t, &c.CheckBase, []string{"Unable to find Dockerfile"}); !ok {
		t.Error(msg)
	}
}

func TestExcludeServiceSingle(t *testing.T) {
	c := docker.BaseImageCheck{
		Allowed: []string{"bitnami/redis"},
		Exclude: []string{"service1"},
		Paths:   []string{"./fixtures/compose-dockerfile"},
	}
	c.RunCheck()
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"Dockerfiles adhere to the policy"}); !ok {
		t.Error(msg)
	}
}

func TestExcludeServiceMany(t *testing.T) {
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
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"Dockerfiles adhere to the policy"}); !ok {
		t.Error(msg)
	}
}
