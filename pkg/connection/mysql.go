package connection

import (
	"errors"

	"github.com/salsadigitalauorg/shipshape/pkg/command"
)

type Mysql struct {
	// Common fields.
	Name string `yaml:"name"`

	// Plugin fields.
	Connection string `yaml:"connection"`
	DbHost     string `yaml:"db-host"`
	DbPort     string `yaml:"db-port"`
	DbUser     string `yaml:"db-user"`
	DbPass     string `yaml:"db-pass"`
	DbName     string `yaml:"db-name"`
	Query      string `yaml:"query"`
}

//go:generate go run ../../cmd/gen.go connection-plugin --plugin=Mysql

func (p *Mysql) PluginName() string {
	return "mysql"
}

func (p *Mysql) Run() ([]byte, error) {
	cmdArgs := []string{
		"-h", p.DbHost,
		"-P", p.DbPort,
		"-u", p.DbUser,
		"-p" + p.DbPass,
		p.DbName,
		p.Query,
	}

	cn := GetConnection(p.Connection)
	if cn != nil {
		// We currently only support docker connections.
		if cn.PluginName() != "docker.exec" {
			return nil, errors.New("unsupported connection")
		}

		dockerConn := cn.(*DockerExec)
		dockerConn.Command = append([]string{"mysql"}, cmdArgs...)
		return dockerConn.Run()
	}

	return command.ShellCommander("mysql", cmdArgs...).Output()
}
