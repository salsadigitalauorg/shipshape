package command_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	. "github.com/salsadigitalauorg/shipshape/pkg/fact/command"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

func TestCommandInit(t *testing.T) {
	assert := assert.New(t)

	// Test that the command plugin is registered.
	factPlugin := fact.Manager().GetFactories()["command"]("TestCommand")
	assert.NotNil(factPlugin)
	keyFacter, ok := factPlugin.(*Command)
	assert.True(ok)
	assert.Equal("TestCommand", keyFacter.Id)
}

func TestCommandPluginName(t *testing.T) {
	commandF := New("TestCommand")
	assert.Equal(t, "command", commandF.GetName())
}

func TestCommandSupportedConnections(t *testing.T) {
	commandF := New("TestCommand")
	supportLevel, connections := commandF.SupportedConnections()
	assert.Equal(t, plugin.SupportNone, supportLevel)
	assert.Empty(t, connections)
}

func TestCommandSupportedInputs(t *testing.T) {
	commandF := New("TestCommand")
	supportLevel, inputs := commandF.SupportedInputFormats()
	assert.Equal(t, plugin.SupportNone, supportLevel)
	assert.ElementsMatch(t, []string{}, inputs)
}

func TestCommandCollect(t *testing.T) {
	tests := []internal.FactCollectTest{
		{
			Name:           "emptyCommand",
			Facter:         New("TestCommand"),
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"code": "1", "stderr": "exec: no command", "stdout": "",
			},
			ExpectedErrors: []error{errors.New("exec: no command")},
		},
		{
			Name: "emptyCommand/ignoreError",
			FactFn: func() fact.Facter {
				f := New("TestCommand")
				f.IgnoreError = true
				return f
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"code": "1", "stderr": "exec: no command", "stdout": "",
			},
		},
		{
			Name: "echo",
			FactFn: func() fact.Facter {
				f := New("TestCommand")
				f.Cmd = "echo"
				f.Args = []string{"hello"}
				return f
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"code": "0", "stderr": "", "stdout": "hello",
			},
		},
		{
			Name: "multiline",
			FactFn: func() fact.Facter {
				f := New("TestCommand")
				f.Cmd = "ls"
				f.Args = []string{"-1"}
				return f
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"code": "0", "stderr": "", "stdout": "command.go\ncommand_test.go",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			internal.TestFactCollect(t, tt)
		})
	}
}
