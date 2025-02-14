package command

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

// Command is a representation of a shell command.
type Command struct {
	fact.BaseFact

	// Plugin-specific fields
	Cmd         string   `yaml:"cmd"`
	Args        []string `yaml:"args"`
	IgnoreError bool     `yaml:"ignore-error"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --package=command

func init() {
	fact.GetManager().Register("command", func(n string) fact.Facter {
		return New(n)
	})
}

func New(id string) *Command {
	return &Command{
		BaseFact: fact.BaseFact{
			BasePlugin: plugin.BasePlugin{
				Id: id,
			},
			Format: data.FormatMapString,
		},
	}
}

func (p *Command) GetName() string {
	return "command"
}

func (p *Command) Collect() {
	contextLogger := log.WithFields(log.Fields{
		"fact-plugin": p.GetName(),
		"fact":        p.GetId(),
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
			p.AddErrors(err)
		}
	}

	p.SetData(res)
}
