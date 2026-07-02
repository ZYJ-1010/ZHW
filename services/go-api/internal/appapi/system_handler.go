package appapi

import (
	"net/http"

	"zhw-mini/services/go-api/internal/common/httpx"
)

func (s *Server) adminSystemReadiness(w http.ResponseWriter, r *http.Request) {
	report := s.cfg.ReadinessReport()
	s.recordOperation(r, "system:readiness:view", "system", "readiness", map[string]interface{}{
		"ready":      report.Ready,
		"production": report.Production,
	})
	httpx.OK(w, report)
}
