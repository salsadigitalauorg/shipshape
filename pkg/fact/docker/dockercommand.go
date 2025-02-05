package docker

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

type DockerCommand struct {
	// Common fields.
	Name                 string          `yaml:"name"`
	Format               data.DataFormat `yaml:"format"`
	ConnectionName       string          `yaml:"connection"`
	InputName            string          `yaml:"input"`
	AdditionalInputNames []string        `yaml:"additional-inputs"`
	connection           connection.Connectioner
	input                fact.Facter
	additionalInputs     []fact.Facter
	errors               []error
	data                 interface{}

	// Plugin fields.
	Command []string `yaml:"command"`
	AsList  bool     `yaml:"as-list"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --plugin=DockerCommand --package=docker

func init() {
	fact.Registry["docker:command"] = func(n string) fact.Facter { return &DockerCommand{Name: n} }
}

func (p *DockerCommand) PluginName() string {
	return "docker:command"
}

func (p *DockerCommand) SupportedConnections() (fact.SupportLevel, []string) {
	return fact.SupportRequired, []string{"docker:exec"}
}

func (p *DockerCommand) SupportedInputs() (fact.SupportLevel, []string) {
	return fact.SupportNone, []string{}
}

func (p *DockerCommand) Collect() {
	log.WithFields(log.Fields{
		"fact-plugin":       p.PluginName(),
		"fact":              p.Name,
		"connection":        p.GetConnectionName(),
		"connection-plugin": p.connection.PluginName(),
	}).Debug("collecting data")

	dockerConn := p.connection.(*connection.DockerExec)
	dockerConn.Command = p.Command
	rawData, err := dockerConn.Run()
	if err != nil {
		errMsg := command.GetMsgFromCommandError(err)
		if errMsg == "" {
			errMsg = string(rawData)
		} else {
			errMsg = errMsg + ": " + string(rawData)
		}
		err = errors.New(errMsg)
		log.WithError(err).Error("docker command failed")
		p.errors = append(p.errors, err)
		return
	}

	if !p.AsList {
		p.Format = data.FormatRaw
		p.data = rawData
		return
	}

	p.Format = data.FormatListString
	p.data = utils.MultilineOutputToSlice(rawData)
}
