package svc

import (
	"github.com/tal-tech/go-zero/core/stores/sqlc"
	"github.com/tal-tech/go-zero/core/stores/sqlx"
	"project_temp/service/user/cmd/rpc/internal/config"
	"project_temp/service/user/model"
)

type ServiceContext struct {
	Config    config.Config
	BookModel model.BookModel
	Sqlc      sqlc.CachedConn
}

func NewServiceContext(c config.Config) *ServiceContext {
	//  defaultExpiry         = time.Hour * 24 * 7
	//	defaultNotFoundExpiry = time.Minute
	return &ServiceContext{
		Config:    c,
		Sqlc:      sqlc.NewConn(sqlx.NewMysql(c.Mysql.DataSource), c.Cache),
		BookModel: model.NewBookModel(sqlx.NewMysql(c.Mysql.DataSource), c.Cache), // 手动代码
	}
}
