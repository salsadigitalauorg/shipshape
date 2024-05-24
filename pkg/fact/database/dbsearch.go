package database

import (
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/go-sql-driver/mysql"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
)

// DbSearch searches the provided text from all tables of a database.
type DbSearch struct {
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
	Tables  map[string][]string `yaml:"tables"`
	Search  string              `yaml:"search"`
	IdField string              `yaml:"id-field"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --plugin=DbSearch --package=database

func init() {
	fact.Registry["database.db-search"] = func(n string) fact.Facter { return &DbSearch{Name: n} }
}

func (p *DbSearch) PluginName() string {
	return "database.db-search"
}

func (p *DbSearch) SupportedConnections() (fact.SupportLevel, []string) {
	return fact.SupportRequired, []string{"mysql"}
}

func (p *DbSearch) SupportedInputs() (fact.SupportLevel, []string) {
	return fact.SupportNone, []string{}
}

func (p *DbSearch) Collect() {
	if p.Format == "" {
		p.Format = fact.FormatMapNestedString
	} else if p.Format != fact.FormatMapNestedString {
		p.errors = append(p.errors, fmt.Errorf("format '%s' not supported", p.Format))
		return
	}

	if p.IdField == "" {
		p.errors = append(p.errors, fmt.Errorf("id-field is required"))
		return
	}

	conn := p.connection.(*connection.Mysql)
	log.WithField("mysqlConn", conn).Debug("collecting data")

	if len(p.Tables) == 0 {
		if err := p.fetchTablesColumns(*conn); err != nil {
			log.WithError(err).Error("failed to fetch tables and columns")
			return
		}
	}
	log.WithField("tables", fmt.Sprintf("%+v", p.Tables)).Trace("tables")

	// Execute the connection to get the db instance.
	if _, err := conn.Run(); err != nil {
		p.errors = append(p.errors, err)
		return
	}

	unknownColMsg := fmt.Sprintf("Unknown column '%s' in 'field list'", p.IdField)

	occurrences := map[string]map[string]string{}
	for table, cols := range p.Tables {
		for _, col := range cols {
			log.WithFields(log.Fields{
				"table":  table,
				"column": col,
			}).Trace("searching")
			ids := []string{}
			if err := conn.Db.Select(goqu.C(p.IdField)).Distinct().From(table).Where(
				goqu.C(col).Like(p.Search)).ScanVals(&ids); err != nil {
				log.WithField("err", fmt.Sprintf("%#v", err)).Trace("failed to search")
				if mErr, ok := err.(*mysql.MySQLError); !ok || mErr.Message != unknownColMsg {
					p.errors = append(p.errors, err)
					continue
				}
			}
			if len(ids) == 0 {
				continue
			}
			if _, ok := occurrences[table]; !ok {
				occurrences[table] = map[string]string{}
			}
			occurrences[table][col] = strings.Join(ids, ",")
		}
	}

	log.WithField("occurrences", fmt.Sprintf("%+v", occurrences)).Trace("search done")
	p.data = occurrences
}

// fetchTablesColumns fetches list of tables and columns from the information_schema db.
func (p *DbSearch) fetchTablesColumns(conn connection.Mysql) error {
	origDb := conn.Database
	conn.Database = "information_schema"

	// Build the query to fetch the list of tables.
	type tableCols struct {
		Table string `db:"table_name"`
		Col   string `db:"column_name"`
	}

	// Execute the connection to get the db instance.
	if _, err := conn.Run(); err != nil {
		p.errors = append(p.errors, err)
		return err
	}

	var tablesCols []tableCols
	if err := conn.Db.Select(&tableCols{}).From("columns").Where(goqu.And(
		goqu.C("table_schema").Eq(origDb),
		goqu.C("data_type").In([]string{"char", "varchar", "longtext", "longblob"}),
	)).ScanStructs(&tablesCols); err != nil {
		p.errors = append(p.errors, err)
		return err
	}

	// Build the tables map.
	p.Tables = map[string][]string{}
	for _, tc := range tablesCols {
		if _, ok := p.Tables[tc.Table]; !ok {
			p.Tables[tc.Table] = []string{}
		}
		p.Tables[tc.Table] = append(p.Tables[tc.Table], tc.Col)
	}
	return nil
}
