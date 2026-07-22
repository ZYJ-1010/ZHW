package appapi

import (
	"encoding/json"
	"net/http"

	"zhw-mini/services/go-api/internal/common/httpx"
)

const inviteCodeConfigKey = "invite_code_config"

type inviteCodeConfigDTO struct {
	MaxBatchCount    int `json:"maxBatchCount"`
	MaxRequestCount  int `json:"maxRequestCount"`
	DefaultValidDays int `json:"defaultValidDays"`
}

func defaultInviteCodeConfig() inviteCodeConfigDTO {
	return inviteCodeConfigDTO{MaxBatchCount: 200, MaxRequestCount: 200, DefaultValidDays: 0}
}

func (s *Server) inviteCodeConfig() inviteCodeConfigDTO {
	config := defaultInviteCodeConfig()
	if s.systemConfig.Get(inviteCodeConfigKey, &config) {
		if config.MaxBatchCount < 1 || config.MaxBatchCount > 1000 {
			config.MaxBatchCount = 200
		}
		if config.MaxRequestCount < 1 || config.MaxRequestCount > 1000 {
			config.MaxRequestCount = 200
		}
		if config.DefaultValidDays < 0 || config.DefaultValidDays > 3650 {
			config.DefaultValidDays = 0
		}
		return config
	}
	_ = s.systemConfig.Set(inviteCodeConfigKey, config)
	return config
}

func (s *Server) adminInviteCodeConfig(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"config": s.inviteCodeConfig()})
}

func (s *Server) updateAdminInviteCodeConfig(w http.ResponseWriter, r *http.Request) {
	var config inviteCodeConfigDTO
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数不正确")
		return
	}
	if config.MaxBatchCount < 1 || config.MaxBatchCount > 1000 || config.MaxRequestCount < 1 || config.MaxRequestCount > 1000 || config.DefaultValidDays < 0 || config.DefaultValidDays > 3650 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "邀请码配置范围不正确")
		return
	}
	if err := s.systemConfig.Set(inviteCodeConfigKey, config); err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "保存邀请码配置失败")
		return
	}
	s.recordOperation(r, "invite_code:config_update", "system_config", inviteCodeConfigKey, map[string]interface{}{
		"maxBatchCount": config.MaxBatchCount, "maxRequestCount": config.MaxRequestCount, "defaultValidDays": config.DefaultValidDays,
	})
	httpx.OK(w, map[string]interface{}{"config": config})
}
