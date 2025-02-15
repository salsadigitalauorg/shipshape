package database

import (
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

// Search searches the provided text from all tables of a database.
type Search struct {
	fact.BaseFact

	// Plugin fields.
	Tables  map[string][]string `yaml:"tables"`
	Search  string              `yaml:"search"`
	IdField string              `yaml:"id-field"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --package=database

func init() {
	fact.GetManager().Register("database:search", func(n string) fact.Facter {
		return New(n)
	})
}

func New(id string) *Search {
	return &Search{
		BaseFact: fact.BaseFact{
			BasePlugin: plugin.BasePlugin{
				Id: id,
			},
			Format: data.FormatMapNestedString,
		},
	}
}

func (p *Search) GetName() string {
	return "database:search"
}

func (p *Search) SupportedConnections() (plugin.SupportLevel, []string) {
	return plugin.SupportRequired, []string{"mysql"}
}

func (p *Search) Collect() {
	if p.IdField == "" {
		p.AddErrors(fmt.Errorf("id-field is required"))
		return
	}

	conn := p.GetConnection().(*connection.Mysql)
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
		p.AddErrors(err)
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
					p.AddErrors(err)
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
	p.SetData(occurrences)
}

// fetchTablesColumns fetches list of tables and columns from the information_schema db.
func (p *Search) fetchTablesColumns(conn connection.Mysql) error {
	origDb := conn.Database
	conn.Database = "information_schema"

	// Build the query to fetch the list of tables.
	type tableCols struct {
		Table string `db:"table_name"`
		Col   string `db:"column_name"`
	}

	// Execute the connection to get the db instance.
	if _, err := conn.Run(); err != nil {
		p.AddErrors(err)
		return err
	}

	var tablesCols []tableCols
	if err := conn.Db.Select(&tableCols{}).From("columns").Where(goqu.And(
		goqu.C("table_schema").Eq(origDb),
		goqu.C("data_type").In([]string{"char", "varchar", "longtext", "longblob"}),
	)).ScanStructs(&tablesCols); err != nil {
		p.AddErrors(err)
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
