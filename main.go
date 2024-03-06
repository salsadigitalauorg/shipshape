package main

import (
	"github.com/salsadigitalauorg/shipshape/cmd"
)

//go:generate go run cmd/gen.go registry create

func main() {
	cmd.Execute()
}
