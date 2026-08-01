package appapi

import (
	"encoding/json"
	"fmt"
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
	config, err := s.inviteCodeConfigStrict()
	if err != nil {
		return defaultInviteCodeConfig()
	}
	return config
}

func (s *Server) inviteCodeConfigStrict() (inviteCodeConfigDTO, error) {
	config := defaultInviteCodeConfig()
	if s.systemConfig == nil {
		return config, nil
	}
	found, err := s.systemConfig.GetStrict(inviteCodeConfigKey, &config)
	if err != nil {
		return inviteCodeConfigDTO{}, err
	}
	if found {
		if config.MaxBatchCount < 1 || config.MaxBatchCount > 1000 {
			return inviteCodeConfigDTO{}, fmt.Errorf("邀请码批量数量配置无效")
		}
		if config.MaxRequestCount < 1 || config.MaxRequestCount > 1000 {
			return inviteCodeConfigDTO{}, fmt.Errorf("邀请码申请数量配置无效")
		}
		if config.DefaultValidDays < 0 || config.DefaultValidDays > 3650 {
			return inviteCodeConfigDTO{}, fmt.Errorf("邀请码有效期配置无效")
		}
		return config, nil
	}
	if err := s.systemConfig.Set(inviteCodeConfigKey, config); err != nil {
		return inviteCodeConfigDTO{}, err
	}
	return config, nil
}

func (s *Server) adminInviteCodeConfig(w http.ResponseWriter, r *http.Request) {
	config, err := s.inviteCodeConfigStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取邀请码配置失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{"config": config})
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
