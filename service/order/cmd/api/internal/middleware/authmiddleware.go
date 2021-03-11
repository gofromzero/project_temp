package middleware

import (
	"github.com/tal-tech/go-zero/core/logx"
	"net/http"
	"project_temp/service/user/cmd/rpc/userclient"
)

type AuthMiddleware struct {
	userRpc userclient.User
}

func NewAuthMiddleware(user userclient.User) *AuthMiddleware {
	return &AuthMiddleware{
		userRpc: user,
	}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logx.Info("AuthCheck")

		next(w, r)
	}
}
