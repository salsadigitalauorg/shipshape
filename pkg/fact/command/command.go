package command

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
)

// Command is a representation of a shell command.
type Command struct {
	// Common fields.
	Name                 string          `yaml:"name"`
	Format               data.DataFormat `yaml:"format"`
	ConnectionName       string          `yaml:"connection"`
	InputName            string          `yaml:"input"`
	AdditionalInputNames []string        `yaml:"additional-inputs"`

	connection       connection.Connectioner
	input            fact.Facter
	additionalInputs []fact.Facter
	errors           []error
	data             interface{}

	// Plugin fields.
	Cmd         string   `yaml:"cmd"`
	Args        []string `yaml:"args"`
	IgnoreError bool     `yaml:"ignore-error"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --plugin=Command --package=command

func init() {
	fact.Registry["command"] = func(n string) fact.Facter {
		return &Command{Name: n, Format: data.FormatMapString}
	}
}

func (p *Command) PluginName() string {
	return "command"
}

func (p *Command) SupportedConnections() (fact.SupportLevel, []string) {
	return fact.SupportNone, []string{}
}

func (p *Command) SupportedInputs() (fact.SupportLevel, []string) {
	return fact.SupportNone, []string{}
}

func (p *Command) Collect() {
	contextLogger := log.WithFields(log.Fields{
		"fact-plugin": p.PluginName(),
		"fact":        p.Name,
	})

	contextLogger.WithFields(log.Fields{
		"cmd":  p.Cmd,
		"args": p.Args,
	}).Debug("collecting data")

	res := map[string]string{
		"code":   "0",
		"stdout": "",
		"stderr": "",
	}

	data, err := command.ShellCommander(p.Cmd, p.Args...).Output()
	contextLogger.WithFields(log.Fields{
		"stdout": string(data),
		"stderr": fmt.Sprintf("%#v", err),
	}).Debug("command output")

	res["stdout"] = strings.Trim(string(data), " \n")
	if err != nil {
		res["code"] = strconv.Itoa(command.GetExitCode(err))
		res["stderr"] = command.GetMsgFromCommandError(err)

		if !p.IgnoreError {
			contextLogger.
				WithField("stdout", res["stdout"]).
				WithField("stderr", res["stderr"]).
				WithError(err).Error("command failed")
			p.errors = append(p.errors, err)
		}
	}

	p.data = res
}
