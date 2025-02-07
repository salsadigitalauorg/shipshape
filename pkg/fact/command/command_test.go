package command_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	. "github.com/salsadigitalauorg/shipshape/pkg/fact/command"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
)

func TestCommandInit(t *testing.T) {
	assert := assert.New(t)

	// Test that the command plugin is registered.
	factPlugin := fact.Registry["command"]("TestCommand")
	assert.NotNil(factPlugin)
	keyFacter, ok := factPlugin.(*Command)
	assert.True(ok)
	assert.Equal("TestCommand", keyFacter.Name)
}

func TestCommandPluginName(t *testing.T) {
	commandF := Command{Name: "TestCommand"}
	assert.Equal(t, "command", commandF.PluginName())
}

func TestCommandSupportedConnections(t *testing.T) {
	commandF := Command{Name: "TestCommand"}
	supportLevel, connections := commandF.SupportedConnections()
	assert.Equal(t, fact.SupportNone, supportLevel)
	assert.Empty(t, connections)
}

func TestCommandSupportedInputs(t *testing.T) {
	commandF := Command{Name: "TestCommand"}
	supportLevel, inputs := commandF.SupportedInputs()
	assert.Equal(t, fact.SupportNone, supportLevel)
	assert.ElementsMatch(t, []string{}, inputs)
}

func TestCommandCollect(t *testing.T) {
	tests := []internal.FactCollectTest{
		{
			Name:   "emptyCommand",
			Facter: &Command{Name: "TestCommand"},
			ExpectedData: map[string]string{
				"code": "1", "stderr": "exec: no command", "stdout": "",
			},
		},
		{
			Name:   "echo",
			Facter: &Command{Name: "TestCommand", Cmd: "echo", Args: []string{"hello"}},
			ExpectedData: map[string]string{
				"code": "0", "stderr": "", "stdout": "hello",
			},
		},
		{
			Name:   "multiline",
			Facter: &Command{Name: "TestCommand", Cmd: "ls", Args: []string{"-A1"}},
			ExpectedData: map[string]string{
				"code": "0", "stderr": "", "stdout": "command.go\ncommand_gen.go\ncommand_test.go",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			internal.TestFactCollect(t, tt)
		})
	}
}
