package docker

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

type DockerCommand struct {
	fact.BaseFact `yaml:",inline"`

	// Plugin fields.
	Command []string `yaml:"command"`
	AsList  bool     `yaml:"as-list"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --package=docker

func init() {
	fact.Manager().RegisterFactory("docker:command", func(n string) fact.Facter {
		return NewDockerCommand(n)
	})
}

func NewDockerCommand(id string) *DockerCommand {
	return &DockerCommand{
		BaseFact: fact.BaseFact{
			BasePlugin: plugin.BasePlugin{
				Id: id,
			},
		},
	}
}

func (p *DockerCommand) GetName() string {
	return "docker:command"
}

func (p *DockerCommand) SupportedConnections() (plugin.SupportLevel, []string) {
	return plugin.SupportRequired, []string{"docker:exec"}
}

func (p *DockerCommand) Collect() {
	log.WithFields(log.Fields{
		"fact-plugin":       p.GetName(),
		"fact":              p.GetId(),
		"connection":        p.GetConnectionName(),
		"connection-plugin": p.GetConnection().GetName(),
	}).Debug("collecting data")

	dockerConn := p.GetConnection().(*connection.DockerExec)
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
		p.AddErrors(err)
		return
	}

	if !p.AsList {
		p.Format = data.FormatRaw
		p.SetData(rawData)
		return
	}

	p.Format = data.FormatListString
	p.SetData(utils.MultilineOutputToSlice(rawData))
}
