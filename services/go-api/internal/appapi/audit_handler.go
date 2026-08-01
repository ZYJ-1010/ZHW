package appapi

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/audit"
	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/users"
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
	items, err := s.audit.BehaviorLogsStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取行为日志失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) adminBehaviorEvents(w http.ResponseWriter, r *http.Request) {
	items, err := s.audit.QueryBehaviorLogsStrict(audit.BehaviorQuery{
		UserID:    parseInt64Query(r, "userId"),
		EventType: strings.TrimSpace(r.URL.Query().Get("eventType")),
		EventCode: strings.TrimSpace(r.URL.Query().Get("eventCode")),
	})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "查询行为日志失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) adminDashboard(w http.ResponseWriter, r *http.Request) {
	dashboard, err := s.audit.DashboardStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取首页行为统计失败，请稍后重试")
		return
	}
	usersList, err := s.auth.AdminUsers(users.Filter{})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "获取首页用户统计失败")
		return
	}
	gamesList, err := s.games.ListStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取首页组局统计失败，请稍后重试")
		return
	}
	rooms, err := s.im.AdminRoomsStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取聊天室统计失败，请稍后重试")
		return
	}
	registeredUserCount := len(usersList)
	verifiedUserCount := 0
	userStatuses := map[string]int{}
	for _, user := range usersList {
		userStatuses[user.RealnameStatus]++
		if user.RealnameStatus == "verified" {
			verifiedUserCount++
		}
	}
	activeGameCount := 0
	gameStatuses := map[string]int{
		games.StatusDraft:          0,
		games.StatusPendingAudit:   0,
		games.StatusRecruiting:     0,
		games.StatusFull:           0,
		games.StatusInProgress:     0,
		games.StatusPendingConfirm: 0,
		games.StatusPendingReview:  0,
		games.StatusCompleted:      0,
		games.StatusRejected:       0,
		games.StatusCancelled:      0,
		games.StatusDisputed:       0,
		games.StatusSettling:       0,
		games.StatusClosed:         0,
	}
	gameTypes := map[string]int{}
	for _, game := range gamesList {
		gameStatuses[game.Status]++
		primaryCategory := strings.TrimSpace(game.PrimaryCategory)
		if primaryCategory == "" {
			primaryCategory = "unclassified"
		}
		gameTypes[primaryCategory]++
		if game.Status == "recruiting" || game.Status == "in_progress" {
			activeGameCount++
		}
	}
	httpx.OK(w, map[string]interface{}{
		"funnel": dashboard.Funnel, "retention": dashboard.Retention, "updatedAt": dashboard.UpdatedAt,
		"summary": map[string]interface{}{
			"registeredUserCount": registeredUserCount, "verifiedUserCount": verifiedUserCount,
			"activeGameCount": activeGameCount, "imRoomCount": len(rooms),
		},
		"distributions": map[string]interface{}{
			"userStatuses": userStatuses, "gameStatuses": gameStatuses, "gameTypes": gameTypes,
		},
	})
}

func (s *Server) adminFunnel(w http.ResponseWriter, r *http.Request) {
	snapshot, err := s.audit.FunnelStrict(stringsFromQuery(r, "eventCodes"))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取漏斗数据失败，请稍后重试")
		return
	}
	httpx.OK(w, snapshot)
}

func (s *Server) adminRetention(w http.ResponseWriter, r *http.Request) {
	snapshot, err := s.audit.RetentionStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取留存数据失败，请稍后重试")
		return
	}
	httpx.OK(w, snapshot)
}

func (s *Server) createBehaviorEvent(w http.ResponseWriter, r *http.Request) {
	var req behaviorEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	eventType := strings.TrimSpace(req.EventType)
	if eventType == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "事件类型不能为空")
		return
	}
	eventCode := strings.TrimSpace(firstNonEmptyText(req.EventCode, eventType))
	if eventCode == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "事件编码不能为空")
		return
	}
	userID, _ := s.currentUserID(r)
	log, err := s.audit.RecordBehaviorStrict(audit.BehaviorRequest{
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
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "保存行为事件失败，请稍后重试")
		return
	}
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
	r.Header.Set("X-Admin-ID", strconv.FormatInt(adminID, 10))
	logs, err := s.audit.OperationLogsStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取操作日志失败，请稍后重试")
		return
	}
	items := filterOperationLogs(logs, r, adminID, hasFull)
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
	if _, err := s.audit.RecordBehaviorStrict(audit.BehaviorRequest{
		UserID:       userID,
		EventType:    eventType,
		EventCode:    eventType,
		TargetType:   targetType,
		BusinessType: targetType,
		TargetID:     targetID,
		BusinessID:   targetID,
		Source:       "server",
		Extra:        extra,
	}); err != nil {
		log.Printf("behavior audit persistence degraded event_type=%q user_id=%d err=%v", eventType, userID, err)
	}
}

func (s *Server) recordOperation(r *http.Request, action string, targetType string, targetID string, detail map[string]interface{}) {
	if s.audit == nil {
		return
	}
	if _, err := s.audit.RecordOperationStrict(audit.OperationRequest{
		AdminUserID: parseInt64Header(r, "X-Admin-ID"),
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		RequestID:   r.Header.Get("X-Request-ID"),
		IP:          clientIP(r),
		Detail:      detail,
	}); err != nil {
		log.Printf("admin operation audit persistence degraded action=%q target_type=%q target_id=%q err=%v", action, targetType, targetID, err)
	}
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
	r.Header.Set("X-Admin-ID", strconv.FormatInt(adminID, 10))
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
