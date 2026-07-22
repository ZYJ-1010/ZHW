package appapi

import (
	"encoding/json"
	"net/http"

	"zhw-mini/services/go-api/internal/common/httpx"
)

// deleteAccount requires an explicit confirmation to prevent an accidental
// request from irreversibly releasing the user's login bindings.
func (s *Server) deleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := appUserIDFromRequest(r)
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "未登录")
		return
	}
	var req struct {
		Confirm bool `json:"confirm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !req.Confirm {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请确认注销账号")
		return
	}
	if err := s.auth.DeleteAccount(userID); err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "账号注销失败，请稍后重试")
		return
	}
	s.recordBehavior(userID, "account_deleted", "user", userID, nil)
	httpx.OK(w, map[string]bool{"deleted": true})
}
