package redis

import (
	"context"
	"github.com/451008604/socketServerFrame/config"
	"github.com/451008604/socketServerFrame/logs"
	"github.com/redis/go-redis/v9"
)

type Module struct {
	*redis.Client
	Ctx     context.Context
	account *AccountInfo
	player  *PlayerInfo
}

var DB *Module

func NewRedisModel() *Module {
	DB = &Module{
		Ctx:     context.Background(),
		account: NewAccountInfo(),
	}

	DB.Client = redis.NewClient(&redis.Options{
		Addr:     config.GetGlobalObject().RedisAddress,
		Password: config.GetGlobalObject().RedisPassWord,
	})
	if val, err := DB.Ping(context.Background()).Result(); val != "PONG" {
		logs.PrintLogPanic(err)
	}

	return DB
}
