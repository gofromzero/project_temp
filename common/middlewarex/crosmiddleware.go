package middlewarex

import (
	"github.com/tal-tech/go-zero/rest/httpx"
	"net/http"
)

type CrosMiddleware struct {
}

func NewCrosMiddleware() *CrosMiddleware {
	return &CrosMiddleware{}
}

func (m *CrosMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
		if r.Method == "OPTIONS" {
			httpx.Error(w, http.ErrBodyNotAllowed)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}
