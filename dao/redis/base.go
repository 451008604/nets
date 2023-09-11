package redis

import (
	"context"
	"github.com/451008604/socketServerFrame/config"
	"github.com/451008604/socketServerFrame/dao/sql"
	"github.com/451008604/socketServerFrame/logs"
	"github.com/redis/go-redis/v9"
)

type Module struct {
	*redis.Client
	Ctx    context.Context
	sql    *sql.Query
	player *PlayerInfo
}

var Redis = newRedisModel()

func newRedisModel() *Module {
	DB := &Module{}
	DB.Ctx = context.Background()
	DB.sql = sql.GetSqlQuery()
	DB.player = NewPlayerInfo()

	DB.Client = redis.NewClient(&redis.Options{
		Addr:     config.GetGlobalObject().RedisAddress,
		Password: config.GetGlobalObject().RedisPassWord,
	})
	if val, err := DB.Ping(context.Background()).Result(); val != "PONG" {
		logs.PrintLogPanic(err)
	}

	return DB
}