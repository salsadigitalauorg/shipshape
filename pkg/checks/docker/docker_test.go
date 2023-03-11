package docker_test

import (
	"reflect"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/checks/docker"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestRegisterChecks(t *testing.T) {
	checksMap := map[config.CheckType]string{
		docker.BaseImage: "*docker.BaseImageCheck",
	}
	for ct, ts := range checksMap {
		c := config.ChecksRegistry[ct]()
		ctype := reflect.TypeOf(c).String()
		assert.Equal(t, ts, ctype)
	}
}
