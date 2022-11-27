package docker

import "github.com/salsadigitalauorg/shipshape/pkg/shipshape"

//go:generate go run ../../cmd/gen.go registry --checkpackage=docker

func RegisterChecks() {
	shipshape.ChecksRegistry[BaseImage] = func() shipshape.Check { return &BaseImageCheck{} }
}

func init() {
	RegisterChecks()
}
