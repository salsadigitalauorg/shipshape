package sca

import "github.com/salsadigitalauorg/shipshape/pkg/shipshape"

//go:generate go run ../../cmd/gen.go registry --checkpackage=sca

func RegisterChecks() {
	shipshape.ChecksRegistry[AppType] = func() shipshape.Check { return &AppTypeCheck{} }
}

func init() {
	RegisterChecks()
}
