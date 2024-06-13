package docker

import (
	"errors"

	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

type DockerCommand struct {
	// Common fields.
	Name           string          `yaml:"name"`
	Format         data.DataFormat `yaml:"format"`
	ConnectionName string          `yaml:"connection"`
	InputName      string          `yaml:"input"`
	connection     connection.Connectioner
	input          fact.Facter
	errors         []error
	data           interface{}

	// Plugin fields.
	Command []string `yaml:"command"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --plugin=DockerCommand --package=docker

func init() {
	fact.Registry["docker.command"] = func(n string) fact.Facter { return &DockerCommand{Name: n} }
}

func (p *DockerCommand) PluginName() string {
	return "docker.command"
}

func (p *DockerCommand) SupportedConnections() (fact.SupportLevel, []string) {
	return fact.SupportRequired, []string{"docker.exec"}
}

func (p *DockerCommand) SupportedInputs() (fact.SupportLevel, []string) {
	return fact.SupportNone, []string{}
}

func (p *DockerCommand) Collect() {
	dockerConn := p.connection.(*connection.DockerExec)
	dockerConn.Command = p.Command
	rawData, err := dockerConn.Run()
	if err != nil {
		p.errors = append(p.errors, err)
		return
	}

	switch p.Format {
	case data.FormatRaw:
		p.data = rawData
	case data.FormatList:
		p.data = utils.MultilineOutputToSlice(rawData)
	default:
		p.errors = append(p.errors, errors.New("unsupported format"))
	}
}
