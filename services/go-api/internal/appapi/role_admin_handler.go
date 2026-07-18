package appapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
)

// adminGrantRole supports the一期白名单/导入场景。调用方可以提交单个 userId
// 或 userIds 批量赋予已审核的行家、领路人身份。
func (s *Server) adminGrantRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   int64   `json:"userId"`
		UserIDs  []int64 `json:"userIds"`
		RoleCode string  `json:"roleCode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid role grant request")
		return
	}
	roleCode := strings.ToLower(strings.TrimSpace(req.RoleCode))
	if roleCode == "master" || roleCode == "行家" {
		roleCode = "expert"
	}
	if roleCode == "leader" || roleCode == "领路人" {
		roleCode = "guide"
	}
	if roleCode != "expert" && roleCode != "guide" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "roleCode must be expert or guide")
		return
	}
	ids := append([]int64{}, req.UserIDs...)
	if req.UserID > 0 {
		ids = append(ids, req.UserID)
	}
	ids = uniquePositiveIDs(ids)
	if len(ids) == 0 || len(ids) > 500 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "userIds required")
		return
	}
	items := make([]map[string]interface{}, 0, len(ids))
	for _, userID := range ids {
		if _, found := s.auth.UserByID(userID); !found {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "user not found: "+strconv.FormatInt(userID, 10))
			return
		}
		s.profiles.GrantRole(userID, roleCode)
		items = append(items, map[string]interface{}{"userId": userID, "roleCode": roleCode, "status": "approved"})
		s.recordOperation(r, "role:grant", "user", strconv.FormatInt(userID, 10), map[string]interface{}{"roleCode": roleCode, "source": "admin_whitelist"})
	}
	httpx.OK(w, map[string]interface{}{"items": items, "total": len(items)})
}
