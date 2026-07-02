package appapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"zhw-mini/services/go-api/internal/aidata"
	"zhw-mini/services/go-api/internal/common/httpx"
)

func (s *Server) adminAIDataSnapshot(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.aiDataSnapshot())
}

func (s *Server) adminAIDataIMExport(w http.ResponseWriter, r *http.Request) {
	items, err := s.aidata.ExportIMMessages(s.im.AllMessages())
	if err != nil {
		if errors.Is(err, aidata.ErrIMExportDisabled) {
			httpx.Error(w, http.StatusForbidden, 40361, "ai im export disabled")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "ai data export failed")
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
		httpx.Error(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "invalid json body")
		return
	}
	httpx.OK(w, map[string]interface{}{"config": s.aidata.SetIMExportEnabled(req.Enabled)})
}
