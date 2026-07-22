package appapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/games"
)

// rejectSensitiveGameContent 统一拦截组局可编辑文本。词库复用 IM 的后台词库，
// 避免昵称、聊天和局内容出现三套不一致的敏感词规则。
func (s *Server) rejectSensitiveGameContent(w http.ResponseWriter, values ...string) bool {
	if s.im == nil {
		return false
	}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if _, ok := s.im.CheckSensitiveWords(value); ok {
			httpx.Error(w, http.StatusUnavailableForLegalReasons, 45103, "局内容涉及敏感词，不能保存或发布")
			return true
		}
	}
	return false
}

func (s *Server) rejectSensitiveGameCreateRequest(w http.ResponseWriter, req games.CreateRequest) bool {
	values := []string{
		req.Title,
		req.Description,
		req.Highlights,
		req.Notice,
		req.Audience,
		req.Participation,
		req.PrimaryCategoryText,
		req.SecondaryCategoryText,
		req.CityName,
		req.Address,
	}
	values = append(values, req.Tags...)
	values = append(values, req.CompletionRules...)
	return s.rejectSensitiveGameContent(w, values...)
}

// rejectSensitiveDraftPayload 检查草稿 JSON 中所有字符串值。草稿不会因为还未发布
// 而绕过词库，避免含敏感内容先落库、再通过发布入口漏检。
func (s *Server) rejectSensitiveDraftPayload(w http.ResponseWriter, payload json.RawMessage) bool {
	if len(payload) == 0 || !json.Valid(payload) {
		return false
	}
	var value interface{}
	if err := json.Unmarshal(payload, &value); err != nil {
		return false
	}
	return s.rejectSensitiveGameContent(w, stringValuesInJSON(value)...)
}

func stringValuesInJSON(value interface{}) []string {
	values := make([]string, 0)
	var walk func(interface{})
	walk = func(item interface{}) {
		switch typed := item.(type) {
		case string:
			values = append(values, typed)
		case []interface{}:
			for _, child := range typed {
				walk(child)
			}
		case map[string]interface{}:
			for _, child := range typed {
				walk(child)
			}
		}
	}
	walk(value)
	return values
}
