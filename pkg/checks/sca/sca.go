package sca

import "github.com/salsadigitalauorg/shipshape/pkg/config"

//go:generate go run ../../../cmd/gen.go registry --checkpackage=sca

func RegisterChecks() {
	config.ChecksRegistry[AppType] = func() config.Check { return &AppTypeCheck{} }
}

func init() {
	RegisterChecks()
}
