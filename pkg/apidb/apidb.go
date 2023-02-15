package apidb

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/slazurin/twitch-butler-bot/pkg/data"
)

var connStr = "postgresql://" + data.AppCfg.PostgresUser + ":" + data.AppCfg.PostgresPassword + "@" + data.AppCfg.PostgresHost + ":" + data.AppCfg.PostgresPort + "/" + data.AppCfg.PostgresDB

var DB *sql.DB

func init() {
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
}
