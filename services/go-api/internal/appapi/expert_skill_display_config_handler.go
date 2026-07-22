package appapi

import (
	"encoding/json"
	"net/http"

	"zhw-mini/services/go-api/internal/common/httpx"
)

const expertSkillDisplayConfigKey = "expert.skill_display_config"

// expertSkillDisplayConfigDTO only controls explicit skills configured by an
// expert. It intentionally does not apply to system-evaluated hidden skills.
type expertSkillDisplayConfigDTO struct {
	VisibleSkillLimit int `json:"visibleSkillLimit"`
	HomePreviewCount  int `json:"homePreviewCount"`
}

func defaultExpertSkillDisplayConfig() expertSkillDisplayConfigDTO {
	return expertSkillDisplayConfigDTO{
		VisibleSkillLimit: 3,
		HomePreviewCount:  3,
	}
}

func normalizeExpertSkillDisplayConfig(config expertSkillDisplayConfigDTO) expertSkillDisplayConfigDTO {
	defaults := defaultExpertSkillDisplayConfig()
	if config.VisibleSkillLimit <= 0 {
		config.VisibleSkillLimit = defaults.VisibleSkillLimit
	}
	if config.VisibleSkillLimit > 20 {
		config.VisibleSkillLimit = 20
	}
	if config.HomePreviewCount <= 0 {
		config.HomePreviewCount = defaults.HomePreviewCount
	}
	if config.HomePreviewCount > config.VisibleSkillLimit {
		config.HomePreviewCount = config.VisibleSkillLimit
	}
	return config
}

func (s *Server) currentExpertSkillDisplayConfig() expertSkillDisplayConfigDTO {
	config := defaultExpertSkillDisplayConfig()
	if s.systemConfig != nil {
		var stored expertSkillDisplayConfigDTO
		if s.systemConfig.Get(expertSkillDisplayConfigKey, &stored) {
			config = stored
		}
	}
	return normalizeExpertSkillDisplayConfig(config)
}

func (s *Server) adminExpertSkillDisplayConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentExpertSkillDisplayConfig()})
	case http.MethodPut:
		var req expertSkillDisplayConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "行家技能配置格式错误")
			return
		}
		if req.VisibleSkillLimit < 1 || req.VisibleSkillLimit > 20 {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "显性技能上限须为 1-20")
			return
		}
		if req.HomePreviewCount < 1 || req.HomePreviewCount > req.VisibleSkillLimit {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "首页默认展示数量须为 1 至显性技能上限")
			return
		}
		if s.systemConfig == nil || s.systemConfig.Set(expertSkillDisplayConfigKey, req) != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存行家技能配置失败")
			return
		}
		s.recordOperation(r, "expert_skill_display_config:update", "system_config", expertSkillDisplayConfigKey, map[string]interface{}{
			"visibleSkillLimit": req.VisibleSkillLimit,
			"homePreviewCount":  req.HomePreviewCount,
		})
		httpx.OK(w, map[string]interface{}{"config": req})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "请求方式不支持")
	}
}
