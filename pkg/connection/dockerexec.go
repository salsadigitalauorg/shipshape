package connection

import (
	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

type DockerExec struct {
	BaseConnection `yaml:",inline"`
	Container      string   `yaml:"container"`
	Command        []string `yaml:"command"`
}

func init() {
	Manager().RegisterFactory("docker:exec", func(n string) Connectioner {
		return NewDockerExec(n)
	})
}

func NewDockerExec(id string) *DockerExec {
	return &DockerExec{
		BaseConnection: BaseConnection{
			BasePlugin: plugin.BasePlugin{
				Id: id,
			},
		},
	}
}

func (p *DockerExec) GetName() string {
	return "docker:exec"
}

func (p *DockerExec) Run() ([]byte, error) {
	cmdArgs := []string{"exec", p.Container}
	cmdArgs = append(cmdArgs, p.Command...)
	return command.ShellCommander("docker", cmdArgs...).Output()
}
