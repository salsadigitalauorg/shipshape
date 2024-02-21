package database

import (
	"errors"

	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

type Query struct {
	// Common fields.
	Name       string          `yaml:"name"`
	Format     fact.FactFormat `yaml:"format"`
	Connection string          `yaml:"connection"`
	errors     []error
	data       interface{}

	// Plugin fields.
	Query string `yaml:"query"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --plugin=Query --package=database

func (p *Query) PluginName() string {
	return "database.query"
}

func (p *Query) Gather() {
	if p.Connection == "" {
		p.errors = append(p.errors, errors.New("connection is required"))
		return
	}

	cn := connection.GetConnection(p.Connection)
	if cn == nil {
		p.errors = append(p.errors, errors.New("connection not found"))
		return
	}

	if cn.PluginName() != "mysql" {
		p.errors = append(p.errors, errors.New("unsupported connection"))
		return
	}

	mysqlConn := cn.(*connection.Mysql)
	mysqlConn.Query = p.Query
	data, err := mysqlConn.Run()
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
