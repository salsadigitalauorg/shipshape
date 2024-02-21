package docker

import (
	"errors"

	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

type DockerCommand struct {
	// Common fields.
	Name       string          `yaml:"name"`
	Format     fact.FactFormat `yaml:"format"`
	Connection string          `yaml:"connection"`
	errors     []error
	data       interface{}

	// Plugin fields.
	Command []string `yaml:"command"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --plugin=DockerCommand --package=docker

func (p *DockerCommand) PluginName() string {
	return "docker.command"
}

func (p *DockerCommand) Gather() {
	if p.Connection == "" {
		p.errors = append(p.errors, errors.New("connection is required"))
		return
	}

	cn := connection.GetConnection(p.Connection)
	if cn == nil {
		p.errors = append(p.errors, errors.New("connection not found"))
		return
	}

	if cn.PluginName() != "docker.exec" {
		p.errors = append(p.errors, errors.New("unsupported connection"))
		return
	}

	dockerConn := cn.(*connection.DockerExec)
	dockerConn.Command = p.Command
	data, err := dockerConn.Run()
	if err != nil {
		p.errors = append(p.errors, err)
		return
	}

	switch p.Format {
	case fact.FormatRaw:
		p.data = data
	case fact.FormatList:
		p.data = utils.MultilineOutputToSlice(data)
	default:
		p.errors = append(p.errors, errors.New("unsupported format"))
	}
}
