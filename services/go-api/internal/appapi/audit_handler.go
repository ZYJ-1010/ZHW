package appapi

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/audit"
	"zhw-mini/services/go-api/internal/common/httpx"
)

type behaviorEventRequest struct {
	EventType    string                 `json:"eventType"`
	EventCode    string                 `json:"eventCode"`
	TargetType   string                 `json:"targetType"`
	BusinessType string                 `json:"businessType"`
	TargetID     int64                  `json:"targetId"`
	BusinessID   int64                  `json:"businessId"`
	PagePath     string                 `json:"pagePath"`
	Keyword      string                 `json:"keyword"`
	Source       string                 `json:"source"`
	Device       string                 `json:"device"`
	Extra        map[string]interface{} `json:"extra"`
}

func (s *Server) adminBehaviorLogs(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"items": s.audit.BehaviorLogs()})
}

func (s *Server) adminBehaviorEvents(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"items": s.audit.QueryBehaviorLogs(audit.BehaviorQuery{
		UserID:    parseInt64Query(r, "userId"),
		EventType: strings.TrimSpace(r.URL.Query().Get("eventType")),
		EventCode: strings.TrimSpace(r.URL.Query().Get("eventCode")),
	})})
}

func (s *Server) adminDashboard(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.audit.Dashboard())
}

func (s *Server) adminFunnel(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.audit.Funnel(stringsFromQuery(r, "eventCodes")))
}

func (s *Server) adminRetention(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.audit.Retention())
}

func (s *Server) createBehaviorEvent(w http.ResponseWriter, r *http.Request) {
	var req behaviorEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	eventType := strings.TrimSpace(req.EventType)
	if eventType == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "eventType required")
		return
	}
	eventCode := strings.TrimSpace(firstNonEmptyText(req.EventCode, eventType))
	if eventCode == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "eventCode 不能为空")
		return
	}
	userID, _ := s.currentUserID(r)
	log := s.audit.RecordBehavior(audit.BehaviorRequest{
		UserID:       userID,
		EventType:    eventType,
		EventCode:    eventCode,
		TargetType:   strings.TrimSpace(req.TargetType),
		BusinessType: strings.TrimSpace(req.BusinessType),
		TargetID:     req.TargetID,
		BusinessID:   req.BusinessID,
		PagePath:     strings.TrimSpace(req.PagePath),
		Keyword:      strings.TrimSpace(req.Keyword),
		Source:       strings.TrimSpace(firstNonEmptyText(req.Source, "app")),
		Device:       strings.TrimSpace(firstNonEmptyText(req.Device, r.Header.Get("X-Device"))),
		IP:           clientIP(r),
		Extra:        req.Extra,
	})
	httpx.OK(w, log)
}

func (s *Server) adminOperationLogs(w http.ResponseWriter, r *http.Request) {
	adminID, hasFull := s.admins.HasPermission(s.adminToken(r), "operation_log:view_full")
	if adminID == 0 {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "缺少后台接口权限")
		return
	}
	if !hasFull {
		var hasSelf bool
		adminID, hasSelf = s.admins.HasPermission(s.adminToken(r), "operation_log:view_self")
		if !hasSelf {
			httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "缺少后台接口权限")
			return
		}
	}
	if r.Header.Get("X-Admin-ID") == "" {
		r.Header.Set("X-Admin-ID", strconv.FormatInt(adminID, 10))
	}
	items := filterOperationLogs(s.audit.OperationLogs(), r, adminID, hasFull)
	httpx.OK(w, map[string]interface{}{"items": items, "total": len(items)})
}

func filterOperationLogs(items []audit.OperationLog, r *http.Request, currentAdminID int64, hasFull bool) []audit.OperationLog {
	action := strings.TrimSpace(r.URL.Query().Get("action"))
	targetType := strings.TrimSpace(r.URL.Query().Get("targetType"))
	targetID := strings.TrimSpace(r.URL.Query().Get("targetId"))
	adminUserID := parseInt64Query(r, "adminUserId")
	if !hasFull {
		adminUserID = currentAdminID
	}
	filtered := make([]audit.OperationLog, 0, len(items))
	for _, item := range items {
		if action != "" && !strings.Contains(item.Action, action) {
			continue
		}
		if targetType != "" && item.TargetType != targetType {
			continue
		}
		if targetID != "" && item.TargetID != targetID {
			continue
		}
		if adminUserID > 0 && item.AdminUserID != adminUserID {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func stringsFromQuery(r *http.Request, key string) []string {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func (s *Server) recordBehavior(userID int64, eventType string, targetType string, targetID int64, extra map[string]interface{}) {
	if s.audit == nil {
		return
	}
	s.audit.RecordBehavior(audit.BehaviorRequest{
		UserID:       userID,
		EventType:    eventType,
		EventCode:    eventType,
		TargetType:   targetType,
		BusinessType: targetType,
		TargetID:     targetID,
		BusinessID:   targetID,
		Source:       "server",
		Extra:        extra,
	})
}

func (s *Server) recordOperation(r *http.Request, action string, targetType string, targetID string, detail map[string]interface{}) {
	if s.audit == nil {
		return
	}
	s.audit.RecordOperation(audit.OperationRequest{
		AdminUserID: parseInt64Header(r, "X-Admin-ID"),
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		RequestID:   r.Header.Get("X-Request-ID"),
		IP:          clientIP(r),
		Detail:      detail,
	})
}

func (s *Server) currentUserID(r *http.Request) (int64, bool) {
	token := bearerToken(r.Header.Get("Authorization"))
	if token == "" {
		return 0, false
	}
	user, ok := s.auth.CurrentUser(token)
	if !ok {
		return 0, false
	}
	return user.ID, true
}

func (s *Server) requireAdminPermission(permission string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := s.requireAdminPermissionID(w, r, permission); !ok {
			return
		}
		next(w, r)
	}
}

func (s *Server) requireAdminPermissionID(w http.ResponseWriter, r *http.Request, permission string) (int64, bool) {
	adminID, ok := s.admins.HasPermission(s.adminToken(r), permission)
	if !ok {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "缺少后台接口权限")
		return 0, false
	}
	if r.Header.Get("X-Admin-ID") == "" {
		r.Header.Set("X-Admin-ID", strconv.FormatInt(adminID, 10))
	}
	return adminID, true
}

func parseInt64Header(r *http.Request, key string) int64 {
	value, _ := strconv.ParseInt(strings.TrimSpace(r.Header.Get(key)), 10, 64)
	return value
}

func parseInt64Query(r *http.Request, key string) int64 {
	value, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get(key)), 10, 64)
	return value
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func firstNonEmptyText(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
