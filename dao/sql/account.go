package sql

import (
	"github.com/451008604/socketServerFrame/dao/sqlmodel"
	"github.com/google/uuid"
	"time"
)

func (r *Module) GetAccountInfo(account, password string) (register uint32, accountInfo *sqlmodel.HouseAccount, userInfo *sqlmodel.HouseUser, err error) {
	accountInfo, _ = r.Query.HouseAccount.WithContext(r.Ctx).Where(
		r.Query.HouseAccount.Account.Eq(account),
		r.Query.HouseAccount.Password.Eq(passwordToMd5(password)),
	).First()

	// 注册新账号
	if accountInfo == nil {
		accountInfo, userInfo, err = r.createNewAccount(account, password)
		return 1, accountInfo, userInfo, err
	}

	// 获取玩家数据
	userInfo, err = r.Query.HouseUser.WithContext(r.Ctx).Where(r.Query.HouseUser.ID.Eq(accountInfo.UserID)).First()
	if err != nil {
		return 0, nil, nil, err
	}

	return 0, accountInfo, userInfo, nil
}

func (r *Module) createNewAccount(account, password string) (accountData *sqlmodel.HouseAccount, userData *sqlmodel.HouseUser, err error) {
	rand, _ := uuid.NewRandom()

	// 创建事务
	session := r.Query.Begin()
	// ===========================================创建玩家关联表数据===========================================
	if userData, err = r.insertUserData(session.Query, &sqlmodel.HouseUser{
		UniID:        int64(rand.ID()) + 10000000000,
		Nickname:     account,
		HeadImage:    "1",
		RegisterTime: int32(time.Now().Unix()),
	}); session.Error != nil {
		_ = session.Rollback()
		return nil, nil, err
	}

	if accountData, err = r.insertAccountData(session.Query, &sqlmodel.HouseAccount{
		UserID:   userData.ID,
		Account:  account,
		Password: passwordToMd5(password),
	}); err != nil {
		_ = session.Rollback()
		return nil, nil, err
	}

	if _, err = r.insertCommonData(session.Query, &sqlmodel.HouseCommon{
		UserID: userData.ID,
	}); err != nil {
		_ = session.Rollback()
		return nil, nil, err
	}

	// =====================================================================================================
	// 提交事务
	if err = session.Commit(); err != nil {
		return nil, nil, err
	}

	return accountData, userData, nil
}
