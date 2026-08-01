package appapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"zhw-mini/services/go-api/internal/aidata"
	"zhw-mini/services/go-api/internal/common/httpx"
)

func (s *Server) adminAIDataSnapshot(w http.ResponseWriter, r *http.Request) {
	snapshot, err := s.aiDataSnapshot()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取 AI 数据验收快照失败，请稍后重试")
		return
	}
	httpx.OK(w, snapshot)
}

func (s *Server) adminAIDataIMExport(w http.ResponseWriter, r *http.Request) {
	messages, err := s.im.AllMessagesStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取聊天导出数据失败，请稍后重试")
		return
	}
	items, err := s.aidata.ExportIMMessages(messages)
	if err != nil {
		if errors.Is(err, aidata.ErrIMExportDisabled) {
			httpx.Error(w, http.StatusForbidden, 40361, "AI 聊天数据导出未启用")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "AI 数据导出失败")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) adminAIDataIMExportConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		httpx.OK(w, map[string]interface{}{"config": s.aidata.IMExportConfig()})
		return
	}
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "请求参数错误")
		return
	}
	httpx.OK(w, map[string]interface{}{"config": s.aidata.SetIMExportEnabled(req.Enabled)})
}
