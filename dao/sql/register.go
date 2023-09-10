package sql

import (
	"github.com/451008604/socketServerFrame/dao/sqlmodel"
)

// 插入用户数据
func (r *Module) insertUserData(q *Query, data *sqlmodel.HouseUser) (*sqlmodel.HouseUser, error) {
	return data, q.HouseUser.WithContext(r.Ctx).Create(data)
}

// 插入帐户数据
func (r *Module) insertAccountData(q *Query, data *sqlmodel.HouseAccount) (*sqlmodel.HouseAccount, error) {
	return data, q.HouseAccount.WithContext(r.Ctx).Create(data)
}

// 插入通用数据
func (r *Module) insertCommonData(q *Query, data *sqlmodel.HouseCommon) (*sqlmodel.HouseCommon, error) {
	return data, q.HouseCommon.WithContext(r.Ctx).Create(data)
}
