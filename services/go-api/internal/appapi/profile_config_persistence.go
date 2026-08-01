package appapi

import (
	"net/http"

	"zhw-mini/services/go-api/internal/common/httpx"
)

func (s *Server) saveProfileConfig(w http.ResponseWriter, userID int64, key string, payload map[string]interface{}) (map[string]interface{}, bool) {
	saved, err := s.profiles.SaveSystemManagementConfigStrict(userID, key, payload)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "保存失败，请稍后重试")
		return nil, false
	}
	return saved, true
}

func (s *Server) loadProfileConfig(w http.ResponseWriter, userID int64, key string, fallback map[string]interface{}) (map[string]interface{}, bool) {
	value, err := s.profiles.SystemManagementConfigStrict(userID, key, fallback)
	if err != nil {
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "数据读取失败，请稍后重试")
		return nil, false
	}
	return value, true
}
