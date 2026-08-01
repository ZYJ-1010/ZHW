package appapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/connections"
)

func (s *Server) myConnections(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	items, err := s.connections.MyStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取关系网络失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{
		"items":      items,
		"onlineText": strconv.Itoa(maxInt(1, len(items)+1)) + "\u4eba\u5728\u7ebf",
		"header": map[string]interface{}{
			"title":      "\u6211\u7684\u5173\u7cfb\u7f51",
			"titleIcon":  "\u2605",
			"statusText": "\u52a8\u6001\u66f4\u65b0",
			"address":    "\u57fa\u4e8e\u9080\u8bf7\u3001\u7ec4\u5c40\u548c\u8bc4\u4ef7\u751f\u6210",
		},
		"tabs": []map[string]string{
			{"key": "network", "text": "\u4eba\u8109\u7f51\u7edc"},
			{"key": "nearby", "text": "\u9644\u8fd1\u73a9\u5bb6"},
		},
		"activeTab": "network",
		"network":   s.connectionNetworkPayload(userID, items),
	})
}

func (s *Server) adminConnections(w http.ResponseWriter, r *http.Request) {
	items, err := s.connections.AllStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取关系网络失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) adminUserConnections(w http.ResponseWriter, r *http.Request) {
	userID, ok := idFromAdminPath(w, r.URL.Path, "/api/admin/users/", "/connections")
	if !ok {
		return
	}
	items, err := s.connections.MyStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取用户关系网络失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) routeConnectionPost(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/follow-up") || strings.HasSuffix(r.URL.Path, "/follow-logs") {
		s.createConnectionFollowLog(w, r)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) createConnectionFollowLog(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	connectionID, ok := connectionIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}
	var req connections.FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	log, err := s.connections.AddFollowLog(userID, connectionID, req, s.profiles.IsGuide(userID))
	if err != nil {
		writeConnectionError(w, err)
		return
	}
	s.recordBehavior(userID, "create_connection_follow_up", "connection", connectionID, map[string]interface{}{"followLogId": log.ID})
	httpx.OK(w, log)
}

func connectionIDFromPath(w http.ResponseWriter, path string) (int64, bool) {
	text := strings.TrimPrefix(path, "/api/app/connections/")
	text = strings.TrimSuffix(strings.TrimSuffix(text, "/follow-up"), "/follow-logs")
	id, err := strconv.ParseInt(strings.Trim(text, "/"), 10, 64)
	if err != nil || id <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "人脉 ID 错误")
		return 0, false
	}
	return id, true
}

func writeConnectionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, connections.ErrConnectionNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "人脉不存在")
	case errors.Is(err, connections.ErrConnectionForbidden):
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "无权跟进该人脉")
	case errors.Is(err, connections.ErrInvalidFollowLog):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "跟进记录参数错误")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "人脉操作失败")
	}
}

func (s *Server) connectionNetworkPayload(userID int64, items []connections.Connection) map[string]interface{} {
	nodes := []map[string]interface{}{
		{
			"id":        "user-" + strconv.FormatInt(userID, 10),
			"userId":    userID,
			"type":      "self",
			"label":     "\u6211",
			"avatar":    avatarTextForName("\u6211", userID),
			"level":     0,
			"x":         50,
			"y":         50,
			"strength":  100,
			"highlight": true,
		},
	}
	for index, item := range items {
		name := s.displayName(item.ConnectedUserID, "\u7528\u6237")
		nodes = append(nodes, map[string]interface{}{
			"id":           "user-" + strconv.FormatInt(item.ConnectedUserID, 10),
			"userId":       item.ConnectedUserID,
			"type":         "relation",
			"label":        name,
			"avatar":       avatarTextForName(name, item.ConnectedUserID),
			"relationType": item.RelationType,
			"sourceType":   item.SourceType,
			"strength":     item.StrengthScore,
			"level":        1,
			"x":            20 + (index%4)*20,
			"y":            25 + (index/4)*18,
		})
	}
	return map[string]interface{}{
		"nodes": nodes,
		"edges": s.connectionEdges(userID, items),
		"stats": []map[string]interface{}{
			{"key": "relations", "label": "\u5173\u7cfb\u6570", "value": len(items)},
			{"key": "strongRelations", "label": "\u5f3a\u5173\u7cfb", "value": countStrongConnections(items)},
			{"key": "onlineNodes", "label": "\u52a8\u6001\u8282\u70b9", "value": len(nodes)},
		},
	}
}

func (s *Server) connectionEdges(userID int64, items []connections.Connection) []map[string]interface{} {
	edges := make([]map[string]interface{}, 0, len(items))
	sourceID := "user-" + strconv.FormatInt(userID, 10)
	for _, item := range items {
		edges = append(edges, map[string]interface{}{
			"id":           "edge-" + strconv.FormatInt(item.ID, 10),
			"source":       sourceID,
			"target":       "user-" + strconv.FormatInt(item.ConnectedUserID, 10),
			"relationType": item.RelationType,
			"sourceType":   item.SourceType,
			"strength":     item.StrengthScore,
		})
	}
	return edges
}

func countStrongConnections(items []connections.Connection) int {
	count := 0
	for _, item := range items {
		if item.StrengthScore >= 3 {
			count++
		}
	}
	return count
}
