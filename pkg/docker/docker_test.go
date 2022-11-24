package docker_test

import (
	"reflect"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/docker"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestRegisterChecks(t *testing.T) {
	checksMap := map[shipshape.CheckType]string{
		docker.BaseImage: "*docker.BaseImageCheck",
	}
	for ct, ts := range checksMap {
		c := shipshape.ChecksRegistry[ct]()
		ctype := reflect.TypeOf(c).String()
		assert.Equal(t, ts, ctype)
	}
}
