package database

import (
	"errors"

	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

type Query struct {
	// Common fields.
	Name           string          `yaml:"name"`
	Format         fact.FactFormat `yaml:"format"`
	ConnectionName string          `yaml:"connection"`
	InputName      string          `yaml:"input"`
	connection     connection.Connectioner
	input          fact.Facter
	errors         []error
	data           interface{}

	// Plugin fields.
	Query string `yaml:"query"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --plugin=Query --package=database

func init() {
	fact.Registry["database.query"] = func(n string) fact.Facter { return &Query{Name: n} }
}

func (p *Query) PluginName() string {
	return "database.query"
}

func (p *Query) SupportedConnections() (fact.SupportLevel, []string) {
	return fact.SupportRequired, []string{"mysql"}
}

func (p *Query) SupportedInputs() (fact.SupportLevel, []string) {
	return fact.SupportNone, []string{}
}

func (p *Query) Gather() {
	mysqlConn := p.connection.(*connection.Mysql)
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
