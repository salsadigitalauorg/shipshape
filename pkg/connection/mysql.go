package connection

import (
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/go-sql-driver/mysql"
)

type Mysql struct {
	// Common fields.
	Name string `yaml:"name"`

	// Plugin fields.
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Db       *goqu.Database
}

//go:generate go run ../../cmd/gen.go connection-plugin --plugin=Mysql

func init() {
	Registry["mysql"] = func(n string) Connectioner { return &Mysql{Name: n} }
}

func (p *Mysql) PluginName() string {
	return "mysql"
}

func (p *Mysql) Run() ([]byte, error) {
	if p.Port == "" {
		p.Port = "3306"
	}

	cfg := mysql.NewConfig()
	cfg.User = p.User
	cfg.Passwd = p.Password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%s:%s", p.Host, p.Port)
	cfg.DBName = p.Database

	mysqlDb, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	dialect := goqu.Dialect("mysql")
	p.Db = dialect.DB(mysqlDb)
	return nil, nil
}
