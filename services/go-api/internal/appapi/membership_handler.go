package appapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
)

const membershipPageConfigKey = "membership.page_config"
const membershipRadarConfigKey = "membership.radar_config"

type membershipPageConfigDTO struct {
	Items []map[string]interface{} `json:"items"`
}

type membershipRadarConfigDTO struct {
	Pages          map[string]map[string]interface{} `json:"pages"`
	FormRows       []map[string]interface{}          `json:"formRows"`
	RadarNodes     []map[string]interface{}          `json:"radarNodes"`
	Profile        map[string]interface{}            `json:"profile"`
	Result         map[string]interface{}            `json:"result"`
	Texts          map[string]interface{}            `json:"texts"`
	ActionMessages map[string]interface{}            `json:"actionMessages"`
}

type membershipRadarActionRequest struct {
	Action        string                   `json:"action"`
	TargetID      string                   `json:"targetId"`
	TargetUserID  int64                    `json:"targetUserId"`
	MatchMode     string                   `json:"matchMode"`
	MatchCriteria map[string]interface{}   `json:"matchCriteria"`
	FormRows      []map[string]interface{} `json:"formRows"`
}

func (s *Server) membershipPlans(w http.ResponseWriter, r *http.Request) {
	var config membershipPageConfigDTO
	if s.systemConfig != nil {
		found, err := s.systemConfig.GetStrict(membershipPageConfigKey, &config)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取会员页面配置失败，请稍后重试")
			return
		}
		if found && len(config.Items) > 0 {
			httpx.OK(w, map[string]interface{}{"items": config.Items})
			return
		}
	}
	plans, err := s.membership.PlansStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取会员方案失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": plans})
}

func (s *Server) membershipMy(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	membershipInfo, err := s.membership.MyStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取会员信息失败，请稍后重试")
		return
	}
	httpx.OK(w, membershipInfo)
}

func (s *Server) membershipRadarConfig(w http.ResponseWriter, r *http.Request) {
	var config membershipRadarConfigDTO
	if s.systemConfig != nil {
		found, err := s.systemConfig.GetStrict(membershipRadarConfigKey, &config)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取会员雷达配置失败，请稍后重试")
			return
		}
		if found {
			httpx.OK(w, config)
			return
		}
	}

	httpx.OK(w, membershipRadarConfigDTO{
		Pages:      map[string]map[string]interface{}{},
		FormRows:   []map[string]interface{}{},
		RadarNodes: []map[string]interface{}{},
		Profile:    map[string]interface{}{},
		Result:     map[string]interface{}{"total": 0},
		Texts:      map[string]interface{}{},
	})
}

func (s *Server) membershipRadarAction(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	var req membershipRadarActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	req.Action = strings.TrimSpace(req.Action)
	req.TargetID = strings.TrimSpace(req.TargetID)
	req.MatchMode = strings.TrimSpace(req.MatchMode)
	if req.MatchMode == "" {
		req.MatchMode = "all"
	}
	if !validMembershipRadarAction(req.Action) {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "雷达操作无效")
		return
	}

	state, loaded := s.loadProfileConfig(w, userID, "membership-radar", map[string]interface{}{
		"savedProfile": map[string]interface{}{},
		"lastMatch":    map[string]interface{}{},
		"followed":     []interface{}{},
	})
	if !loaded {
		return
	}
	payload := map[string]interface{}{
		"action":        req.Action,
		"targetId":      req.TargetID,
		"targetUserId":  req.TargetUserID,
		"matchMode":     req.MatchMode,
		"matchCriteria": req.MatchCriteria,
	}
	route := ""
	switch req.Action {
	case "save":
		state["savedProfile"] = map[string]interface{}{
			"formRows":      req.FormRows,
			"matchCriteria": req.MatchCriteria,
		}
	case "start", "rematch", "next":
		state["lastMatch"] = map[string]interface{}{
			"matchMode":     req.MatchMode,
			"matchCriteria": req.MatchCriteria,
			"targetId":      req.TargetID,
			"targetUserId":  req.TargetUserID,
		}
	case "follow":
		if req.TargetUserID > 0 && req.TargetUserID != userID {
			if err := s.connections.UpsertPairStrict(userID, req.TargetUserID, "radar_follow", "membership_radar", 0, 1); err != nil {
				markConnectionPersistenceDegraded(w, "membership_radar_follow", userID, req.TargetUserID, err)
			}
		}
		state["followed"] = appendRadarFollow(state["followed"], req.TargetID, req.TargetUserID)
	case "profile":
		if req.TargetUserID > 0 && req.TargetUserID != userID {
			if err := s.connections.UpsertPairStrict(userID, req.TargetUserID, "radar_view", "membership_radar", 0, 0); err != nil {
				markConnectionPersistenceDegraded(w, "membership_radar_profile", userID, req.TargetUserID, err)
			}
		}
		route = radarProfileRoute(req)
	}
	if _, saved := s.saveProfileConfig(w, userID, "membership-radar", state); !saved {
		return
	}
	s.recordBehavior(userID, "membership_radar_"+req.Action, "membership_radar", req.TargetUserID, payload)
	httpx.OK(w, map[string]interface{}{
		"action":  req.Action,
		"status":  "ok",
		"route":   route,
		"profile": state["savedProfile"],
		"last":    state["lastMatch"],
	})
}

func validMembershipRadarAction(action string) bool {
	switch action {
	case "start", "save", "next", "follow", "profile", "rematch":
		return true
	default:
		return false
	}
}

func appendRadarFollow(value interface{}, targetID string, targetUserID int64) []interface{} {
	items, _ := value.([]interface{})
	key := targetID
	if key == "" && targetUserID > 0 {
		key = strconv.FormatInt(targetUserID, 10)
	}
	for _, item := range items {
		if existing, ok := item.(map[string]interface{}); ok {
			if existing["targetId"] == key {
				return items
			}
		}
	}
	return append(items, map[string]interface{}{
		"targetId":     key,
		"targetUserId": targetUserID,
	})
}

func radarProfileRoute(req membershipRadarActionRequest) string {
	if req.TargetUserID > 0 {
		return "/pages/profile/service-center/invite/member-detail/index?id=" + strconv.FormatInt(req.TargetUserID, 10)
	}
	if req.TargetID != "" {
		return "/pages/profile/service-center/invite/member-detail/index?id=" + req.TargetID
	}
	return ""
}
