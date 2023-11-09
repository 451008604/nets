package sql

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/451008604/nets/config"
	"github.com/451008604/nets/logs"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io"
)

type Module struct {
	Ctx   context.Context
	Query *Query
}

var SQL = newSqlModel()

func newSqlModel() *Module {
	sqlConf := config.GetGlobalObject().Mysql
	if sqlConf.Address == "" {
		return nil
	}
	DB := &Module{Ctx: context.Background()}
	db, err := gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/house_new?charset=utf8mb4&parseTime=true&loc=Local", sqlConf.Username, sqlConf.Password, sqlConf.Address)), &gorm.Config{})
	logs.PrintLogPanic(err)
	// 开启调试模式
	DB.Query = Use(db)

	return DB
}

func GetSqlQuery() *Query {
	return SQL.Query
}

func passwordToMd5(password string) string {
	encryption := md5.New()
	_, _ = io.WriteString(encryption, password)
	return fmt.Sprintf("%x", encryption.Sum(nil))
}
