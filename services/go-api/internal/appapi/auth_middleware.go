package appapi

import (
	"context"
	"net/http"

	"zhw-mini/services/go-api/internal/common/httpx"
)

type appAuthContextKey struct{}

func (s *Server) AppAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get("Authorization"))
		if token == "" {
			httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "未登录")
			return
		}
		user, ok := s.auth.CurrentUser(token)
		if !ok {
			httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "登录已失效")
			return
		}
		ctx := context.WithValue(r.Context(), appAuthContextKey{}, user.ID)
		next(w, r.WithContext(ctx))
	}
}

func appUserIDFromRequest(r *http.Request) (int64, bool) {
	value := r.Context().Value(appAuthContextKey{})
	userID, ok := value.(int64)
	if !ok || userID <= 0 {
		return 0, false
	}
	return userID, true
}
