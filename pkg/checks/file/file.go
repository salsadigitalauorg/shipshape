package file

import "github.com/salsadigitalauorg/shipshape/pkg/config"

//go:generate go run ../../../cmd/gen.go registry --checkpackage=file

func RegisterChecks() {
	config.ChecksRegistry[File] = func() config.Check { return &FileCheck{} }
	config.ChecksRegistry[FileDiff] = func() config.Check { return &FileDiffCheck{} }
}

func init() {
	RegisterChecks()
}
