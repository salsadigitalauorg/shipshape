package docker

import (
	"github.com/salsadigitalauorg/shipshape/pkg/config"
)

//go:generate go run ../../cmd/gen.go registry --checkpackage=docker

func RegisterChecks() {
	config.ChecksRegistry[BaseImage] = func() config.Check { return &BaseImageCheck{} }
}

func init() {
	RegisterChecks()
}
