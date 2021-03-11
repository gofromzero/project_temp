package svc

import (
	"github.com/tal-tech/go-zero/zrpc"
	"project_temp/service/order/cmd/api/internal/config"
	"project_temp/service/user/cmd/rpc/userclient"
)

type ServiceContext struct {
	Config  config.Config
	UserRpc userclient.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		UserRpc: userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}
}
