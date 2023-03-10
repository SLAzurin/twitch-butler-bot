package apidb

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/slazurin/twitch-butler-bot/pkg/data"
)

var DB *sql.DB

var RedisDB *redis.Client

func ManualInit() {
	var err error
	var connStr = "postgresql://" + data.AppCfg.PostgresUser + ":" + data.AppCfg.PostgresPassword + "@" + data.AppCfg.PostgresHost + ":" + data.AppCfg.PostgresPort + "/" + data.AppCfg.PostgresDB + "?sslmode=disable"
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	RedisDB = redis.NewClient(&redis.Options{
		Addr:     data.AppCfg.RedisHost + ":" + data.AppCfg.RedisPort,
		Password: "", // always use no pw, this is a private redis.
		DB:       0,
	})
}
