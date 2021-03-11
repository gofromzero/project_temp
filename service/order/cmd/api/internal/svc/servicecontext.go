package svc

import (
	"github.com/tal-tech/go-zero/rest"
	"github.com/tal-tech/go-zero/zrpc"
	"project_temp/common/middlewarex"
	"project_temp/service/order/cmd/api/internal/config"
	"project_temp/service/user/cmd/rpc/userclient"
)

type ServiceContext struct {
	Config config.Config

	AuthCheck rest.Middleware
	Cros      rest.Middleware

	UserRpc userclient.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	userRpc := userclient.NewUser(zrpc.MustNewClient(c.UserRpc))
	return &ServiceContext{
		Config:    c,
		AuthCheck: middlewarex.NewAuthMiddleware(userRpc).Handle,
		Cros:      middlewarex.NewCrosMiddleware().Handle,
		UserRpc:   userRpc,
	}
}
