package docker

import "github.com/salsadigitalauorg/shipshape/pkg/shipshape"

func RegisterChecks() {
	shipshape.ChecksRegistry[BaseImage] = func() shipshape.Check { return &BaseImageCheck{} }
}

func init() {
	RegisterChecks()
}
