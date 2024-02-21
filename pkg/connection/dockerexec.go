package connection

import "github.com/salsadigitalauorg/shipshape/pkg/command"

type DockerExec struct {
	// Common fields.
	Name   string `yaml:"name"`
	errors []error
	data   []byte

	// Plugin fields.
	Container string   `yaml:"container"`
	Command   []string `yaml:"command"`
}

//go:generate go run ../../cmd/gen.go connection-plugin --plugin=DockerExec

func (p *DockerExec) PluginName() string {
	return "docker.exec"
}

func (p *DockerExec) Run() ([]byte, error) {
	cmdArgs := []string{"exec", p.Container}
	cmdArgs = append(cmdArgs, p.Command...)
	return command.ShellCommander("docker", cmdArgs...).Output()
}
