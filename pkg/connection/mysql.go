package connection

import (
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/go-sql-driver/mysql"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

type Mysql struct {
	BaseConnection `yaml:",inline"`
	Host           string `yaml:"host"`
	Port           string `yaml:"port"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	Database       string `yaml:"database"`
	Db             *goqu.Database
}

func init() {
	Manager().RegisterFactory("mysql", func(n string) Connectioner {
		return NewMysql(n)
	})
}

func NewMysql(id string) *Mysql {
	return &Mysql{
		BaseConnection: BaseConnection{
			BasePlugin: plugin.BasePlugin{
				Id: id,
			},
		},
	}
}

func (p *Mysql) GetName() string {
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
