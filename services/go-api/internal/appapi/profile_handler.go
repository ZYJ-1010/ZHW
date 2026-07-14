package appapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/profiles"
	"zhw-mini/services/go-api/internal/users"
)

const expertApplyConfigKey = "role.expert_apply_config"
const guideApplyConfigKey = "role.guide_apply_config"
const roleStatusPageConfigKey = "role.status_page_config"
const roleApplicationPageConfigKey = "role.application_page_config"
const roleBenefitConfigKey = "role.benefit_config"

func (s *Server) expertSkill(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	profile, err := s.profiles.ExpertSkill(userID)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	httpx.OK(w, profile)
}

func (s *Server) updateExpertSkill(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req profiles.ExpertSkillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	profile, err := s.profiles.UpdateExpertSkill(userID, req)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	s.recordBehavior(userID, "update_expert_skill_profile", "expert_skill_profile", userID, map[string]interface{}{"completeness": profile.Completeness})
	httpx.OK(w, profile)
}

func (s *Server) adminExpertSkill(w http.ResponseWriter, r *http.Request) {
	userID, ok := profileUserIDFromPath(w, r.URL.Path, "/api/admin/experts/", "/skills")
	if !ok {
		return
	}
	httpx.OK(w, s.profiles.AdminExpertSkill(userID))
}

func (s *Server) guideResource(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	profile, err := s.profiles.GuideResource(userID)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	httpx.OK(w, profile)
}

func (s *Server) adminGuideResource(w http.ResponseWriter, r *http.Request) {
	userID, ok := profileUserIDFromPath(w, r.URL.Path, "/api/admin/guides/", "/resources")
	if !ok {
		return
	}
	httpx.OK(w, s.profiles.AdminGuideResource(userID))
}

func (s *Server) adminProfileUsers(w http.ResponseWriter, r *http.Request) {
	keyword := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("keyword")))
	if keyword == "" {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "keyword required")
		return
	}
	items, err := s.auth.AdminUsers(users.Filter{})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "list users failed")
		return
	}
	identityByUser := make(map[int64]adminProfileUserIdentity)
	for _, record := range s.identity.AllRecords() {
		item := adminProfileUserIdentity{IDCardMasked: record.IDCardMasked}
		if plain, err := s.identity.RevealRecord(record); err == nil {
			item.PhoneFull = strings.TrimSpace(plain.Phone)
			item.IDCardFull = strings.TrimSpace(plain.IDCard)
		}
		identityByUser[record.UserID] = item
	}
	matches := make([]map[string]interface{}, 0)
	for _, user := range items {
		identityInfo := identityByUser[user.ID]
		matchFields := profileUserMatchFieldsWithIdentity(user, identityInfo, keyword)
		if len(matchFields) == 0 {
			continue
		}
		matches = append(matches, map[string]interface{}{
			"id":             user.ID,
			"nickname":       user.Nickname,
			"phoneMasked":    user.PhoneMasked,
			"phoneFull":      identityInfo.PhoneFull,
			"idCardMasked":   identityInfo.IDCardMasked,
			"idCardFull":     identityInfo.IDCardFull,
			"realnameStatus": user.RealnameStatus,
			"status":         user.Status,
			"matchFields":    matchFields,
		})
		if len(matches) >= 20 {
			break
		}
	}
	httpx.OK(w, map[string]interface{}{
		"items": matches,
		"total": len(matches),
	})
}

type adminProfileUserIdentity struct {
	PhoneFull    string
	IDCardMasked string
	IDCardFull   string
}

func profileUserMatchFields(user users.User, idCardMasked string, keyword string) []string {
	return profileUserMatchFieldsWithIdentity(user, adminProfileUserIdentity{IDCardMasked: idCardMasked}, keyword)
}

func profileUserMatchFieldsWithIdentity(user users.User, identityInfo adminProfileUserIdentity, keyword string) []string {
	fields := make([]string, 0, 4)
	if strings.Contains(strconv.FormatInt(user.ID, 10), keyword) {
		fields = append(fields, "用户编号")
	}
	if strings.Contains(strings.ToLower(user.Nickname), keyword) {
		fields = append(fields, "昵称")
	}
	phoneMasked := strings.ToLower(user.PhoneMasked)
	phoneFull := strings.ToLower(strings.TrimSpace(identityInfo.PhoneFull))
	if phoneMasked != "" && (strings.Contains(phoneMasked, keyword) || maskProfilePhoneKeyword(keyword) == phoneMasked) || phoneFull != "" && strings.Contains(phoneFull, keyword) {
		fields = append(fields, "手机号")
	}
	idMasked := strings.ToLower(identityInfo.IDCardMasked)
	idFull := strings.ToLower(strings.TrimSpace(identityInfo.IDCardFull))
	if idMasked != "" && (strings.Contains(idMasked, keyword) || maskProfileIDCardKeyword(keyword) == idMasked) || idFull != "" && strings.Contains(idFull, keyword) {
		fields = append(fields, "身份证号")
	}
	return fields
}

func maskProfilePhoneKeyword(keyword string) string {
	digits := digitsOnly(keyword)
	if len(digits) != 11 {
		return ""
	}
	return strings.ToLower(digits[:3] + "****" + digits[7:])
}

func maskProfileIDCardKeyword(keyword string) string {
	text := strings.ToUpper(strings.TrimSpace(keyword))
	if len(text) != 18 {
		return ""
	}
	return strings.ToLower(text[:3] + strings.Repeat("*", len(text)-7) + text[len(text)-4:])
}

func digitsOnly(value string) string {
	var builder strings.Builder
	for _, char := range value {
		if char >= '0' && char <= '9' {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func (s *Server) updateGuideResource(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req profiles.GuideResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	profile, err := s.profiles.UpdateGuideResource(userID, req)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	s.recordBehavior(userID, "update_guide_resource_profile", "guide_resource_profile", userID, map[string]interface{}{"completeness": profile.Completeness})
	httpx.OK(w, profile)
}

func (s *Server) expertApplyConfig(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireIdentityUser(w, r); !ok {
		return
	}
	httpx.OK(w, s.currentExpertApplyConfig())
}

func (s *Server) guideApplyConfig(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireIdentityUser(w, r); !ok {
		return
	}
	httpx.OK(w, s.currentGuideApplyConfig())
}

func (s *Server) roleStatusPageConfig(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireIdentityUser(w, r); !ok {
		return
	}
	httpx.OK(w, s.currentRoleStatusPageConfig())
}

func (s *Server) roleBenefitConfig(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireIdentityUser(w, r); !ok {
		return
	}
	httpx.OK(w, s.currentRoleBenefitConfig())
}

func (s *Server) adminRoleBenefitConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentRoleBenefitConfig()})
	case http.MethodPut:
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid role benefit config")
			return
		}
		config, err := normalizeRoleBenefitConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if err := s.systemConfig.Set(roleBenefitConfigKey, config); err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "save role benefit config failed")
			return
		}
		s.recordOperation(r, "role_benefit_config:update", "system_config", "role_benefit_config", map[string]interface{}{
			"version": config["version"],
		})
		httpx.OK(w, map[string]interface{}{"config": s.currentRoleBenefitConfig()})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "method not allowed")
	}
}

func (s *Server) currentExpertApplyConfig() map[string]interface{} {
	var config map[string]interface{}
	if s.systemConfig != nil && s.systemConfig.Get(expertApplyConfigKey, &config) && len(config) > 0 {
		return config
	}
	return defaultExpertApplyConfig()
}

func (s *Server) currentGuideApplyConfig() map[string]interface{} {
	var config map[string]interface{}
	if s.systemConfig != nil && s.systemConfig.Get(guideApplyConfigKey, &config) && len(config) > 0 {
		return config
	}
	return defaultGuideApplyConfig()
}

func (s *Server) currentRoleStatusPageConfig() map[string]interface{} {
	var config map[string]interface{}
	if s.systemConfig != nil && s.systemConfig.Get(roleStatusPageConfigKey, &config) && len(config) > 0 {
		return config
	}
	return defaultRoleStatusPageConfig()
}

func (s *Server) currentRoleApplicationPageConfig() map[string]interface{} {
	var config map[string]interface{}
	if s.systemConfig != nil && s.systemConfig.Get(roleApplicationPageConfigKey, &config) && len(config) > 0 {
		return config
	}
	return defaultRoleApplicationPageConfig()
}

func (s *Server) currentRoleBenefitConfig() map[string]interface{} {
	var config map[string]interface{}
	if s.systemConfig != nil && s.systemConfig.Get(roleBenefitConfigKey, &config) && len(config) > 0 {
		return config
	}
	return defaultRoleBenefitConfig()
}

func normalizeRoleBenefitConfig(req map[string]interface{}) (map[string]interface{}, error) {
	if len(req) == 0 {
		return nil, errors.New("role benefit config required")
	}
	if _, ok := req["roleComparison"].(map[string]interface{}); !ok {
		return nil, errors.New("roleComparison required")
	}
	return req, nil
}

func defaultRoleBenefitConfig() map[string]interface{} {
	return map[string]interface{}{
		"version": "2026-07-01",
		"permissionPrompts": map[string]interface{}{
			"expert": map[string]interface{}{
				"roleType":  "expert",
				"title":     "我懂玩家需要什么！我申请成为行家",
				"primary":   "申请成为行家",
				"secondary": "查看权益对比",
			},
			"guide": map[string]interface{}{
				"roleType":  "guide",
				"title":     "我愿意带领更多人一起玩！我申请成为领路人",
				"primary":   "申请成为领路人",
				"secondary": "查看权益对比",
			},
		},
		"roleComparison": map[string]interface{}{
			"name":      "权益对比页",
			"mode":      "roleComparison",
			"roleBadge": "权益",
			"title":     "角色权益对比",
			"subtitle":  "选择适合你的角色，开启不同玩法",
			"roles": []map[string]interface{}{
				{"key": "player", "name": "玩家", "level": "Lv.1+", "active": true},
				{"key": "guide", "name": "领路人", "level": "Lv.5+", "active": false},
				{"key": "expert", "name": "行家", "level": "Lv.20+", "active": false},
			},
			"benefits": []map[string]interface{}{
				{"name": "发起组局", "player": "✓", "leader": "—", "expert": "✓"},
				{"name": "加入组局", "player": "✓", "leader": "✓", "expert": "✓"},
				{"name": "创建路线", "player": "✓", "leader": "—", "expert": "✓"},
				{"name": "分润收益", "player": "—", "leader": "基础会员40%", "expert": "高级会员40%"},
				{"name": "服务交易", "player": "—", "leader": "—", "expert": "✓"},
				{"name": "数据看板", "player": "—", "leader": "✓", "expert": "✓"},
				{"name": "信用背书", "player": "—", "leader": "✓", "expert": "✓"},
			},
			"primary": "立即申请角色",
		},
	}
}

func defaultRoleApplicationPageConfig() map[string]interface{} {
	return map[string]interface{}{
		"pageTitle": "选择你的身份",
		"pageDesc":  "玩家为默认身份。行家和领路人需提交申请，审核通过后开放对应能力。",
		"texts": map[string]string{
			"loadingText":        "加载中...",
			"loadFailedText":     "角色申请加载失败",
			"conditionLabel":     "条件达成",
			"paymentLabel":       "付费状态",
			"pendingButtonText":  "已进入审核",
			"submittingText":     "提交中...",
			"submitButtonText":   "提交申请",
			"submitFailedText":   "提交失败",
			"requirementPrefix":  "• ",
			"defaultSuccessText": "申请已提交",
		},
	}
}

func defaultExpertApplyConfig() map[string]interface{} {
	yearOptions := make([]string, 0, 31)
	for i := 1; i <= 30; i++ {
		yearOptions = append(yearOptions, strconv.Itoa(i)+"年")
	}
	yearOptions = append(yearOptions, "30年以上")

	return map[string]interface{}{
		"skillOptions": []map[string]interface{}{
			{"name": "摄影", "active": true},
			{"name": "户外", "active": false},
			{"name": "美食", "active": false},
			{"name": "文化", "active": false},
			{"name": "手工", "active": false},
			{"name": "运动", "active": false},
			{"name": "音乐", "active": false},
			{"name": "+自定义", "custom": true},
		},
		"fields": []map[string]interface{}{
			{"type": "chips", "key": "skillDomain", "label": "选择技能领域", "required": true},
			{"type": "input", "key": "skillTags", "label": "技能标签", "required": true, "placeholder": "如：人像摄影、风光摄影、夜景拍摄", "helper": "添加具体标签，让用户更容易找到你", "maxlength": 30},
			{"type": "select", "key": "experienceYears", "label": "从业年限", "required": true, "placeholder": "请选择从业年限"},
			{"type": "textarea", "key": "intro", "label": "个人简介", "required": true, "placeholder": "介绍你的专业背景、服务风格、擅长领域...", "helper": "不少于 50 字，突出你的专业优势", "maxlength": 300},
		},
		"uploadField": map[string]interface{}{
			"label":       "资质证明",
			"required":    true,
			"icon":        "📎",
			"title":       "点击上传作品集及凭证",
			"acceptTypes": []string{"JPG", "PNG", "PDF"},
			"maxCount":    5,
		},
		"validationRules": map[string]interface{}{
			"skillTags":   map[string]interface{}{"minLength": 2, "maxLength": 30},
			"intro":       map[string]interface{}{"minLength": 50, "maxLength": 300},
			"serviceName": map[string]interface{}{"minLength": 2, "maxLength": 20},
			"customSkill": map[string]interface{}{"minLength": 2, "maxLength": 8},
			"money":       map[string]interface{}{"integerMaxLength": 8, "decimalMaxLength": 2},
		},
		"yearOptions":  yearOptions,
		"serviceCount": 3,
		"priceHint":    "平台将收取 10% 服务费",
	}
}

func defaultGuideApplyConfig() map[string]interface{} {
	return map[string]interface{}{
		"applyRoleType": "guide",
		"applyRoleName": "\u9886\u8def\u4eba",
		"requirements": []map[string]interface{}{
			{"title": "\u73a9\u5bb6\u7b49\u7ea7\u8fbe\u5230 Lv.5", "text": "\u4ee5\u540e\u53f0\u8d44\u683c\u89c4\u5219\u4e3a\u51c6", "done": false},
			{"title": "\u5b8c\u6210\u5b9e\u540d\u8ba4\u8bc1", "text": "\u9886\u8def\u4eba\u5fc5\u987b\u5b9e\u540d", "done": false},
			{"title": "\u4fe1\u7528\u5206 \u2265 80 \u5206", "text": "\u4ee5\u4fe1\u7528\u8bb0\u5f55\u4e3a\u51c6", "done": false},
		},
		"planTask": map[string]interface{}{
			"title":  "\u63d0\u4ea4\u9886\u8def\u8ba1\u5212\u4e66",
			"text":   "\u63cf\u8ff0\u4f60\u7684\u5e26\u961f\u98ce\u683c\u3001\u6218\u7ee9\u3001\u8d44\u6e90\u548c\u89c4\u5212",
			"done":   false,
			"action": "\u53bb\u586b\u5199 \u203a",
		},
		"perks": []map[string]interface{}{
			{"icon": "Y", "text": "\u6709\u6743\u76ca\u7684\u9886\u8def\u4eba\u5f15\u8350\u73a9\u5bb6\u7ec4\u5c40\u53ef\u83b7\u5f97\u76f8\u5e94\u6536\u5165"},
			{"icon": "*", "text": "\u4e13\u5c5e\u9886\u8def\u4eba\u6807\u8bc6\u4e0e\u4f18\u5148\u63a8\u8350\u4f4d"},
			{"icon": "D", "text": "\u6570\u636e\u770b\u677f\uff1a\u67e5\u770b\u9080\u7ea6\u6570\u636e\u4e0e\u5173\u7cfb\u7f51\u7edc"},
		},
		"fields": []map[string]interface{}{
			{"key": "city", "label": "\u6240\u5728\u57ce\u5e02", "type": "input", "required": true, "placeholder": "\u8bf7\u8f93\u5165\u5e38\u9a7b\u57ce\u5e02", "maxlength": 20, "helper": "\u7528\u4e8e\u5339\u914d\u540c\u57ce\u73a9\u5bb6\u4e0e\u7ec4\u5c40\u63a8\u8350"},
			{"key": "audience", "label": "\u53ef\u63a8\u8350\u4eba\u7fa4", "type": "chips", "required": true, "options": []map[string]interface{}{{"name": "\u670b\u53cb", "active": true}, {"name": "\u540c\u4e8b", "active": true}, {"name": "\u540c\u57ce\u73a9\u5bb6", "active": true}, {"name": "\u793e\u7fa4\u6210\u5458", "active": false}}, "helper": "\u53ef\u591a\u9009\uff0c\u540e\u7eed\u5c06\u7528\u4e8e\u5173\u7cfb\u7f51\u63a8\u8350"},
			{"key": "contact", "label": "\u5e38\u7528\u8054\u7cfb\u65b9\u5f0f", "type": "input", "required": true, "placeholder": "\u8bf7\u8f93\u5165\u5fae\u4fe1\u53f7\u6216\u624b\u673a\u53f7", "maxlength": 30},
			{"key": "guidePlan", "label": "\u9886\u8def\u8ba1\u5212\u4e66", "type": "textarea", "required": true, "placeholder": "\u8bf7\u63cf\u8ff0\u4f60\u7684\u5e26\u961f\u98ce\u683c\u3001\u6218\u7ee9\u3001\u8d44\u6e90\u548c\u89c4\u5212", "maxlength": 300, "helper": "\u4e0d\u5c11\u4e8e 50 \u5b57\uff0c\u8bf4\u660e\u4f60\u80fd\u5e2e\u52a9\u73a9\u5bb6\u5b8c\u6210\u7ec4\u5c40\u7684\u65b9\u5f0f"},
		},
		"uploadField": map[string]interface{}{
			"label":       "\u8d44\u8d28\u8bc1\u660e",
			"required":    true,
			"icon":        "+",
			"title":       "\u70b9\u51fb\u4e0a\u4f20\u4f5c\u54c1\u96c6\u53ca\u51ed\u8bc1",
			"helper":      "\u652f\u6301 JPG\u3001PNG\u3001PDF\uff0c\u6700\u591a 5 \u5f20",
			"acceptTypes": []string{"JPG", "PNG", "PDF"},
			"maxCount":    5,
		},
		"serviceCount": 3,
		"serviceBlocks": []map[string]interface{}{
			{"id": "guide-service-1", "title": "\u4e1a\u52a1"},
			{"id": "guide-service-2", "title": "\u4e1a\u52a1"},
			{"id": "guide-service-3", "title": "\u4e1a\u52a1"},
		},
		"validationRules": map[string]interface{}{
			"guidePlan":   map[string]interface{}{"minLength": 50, "maxLength": 300},
			"serviceName": map[string]interface{}{"minLength": 2, "maxLength": 20},
			"money":       map[string]interface{}{"integerMaxLength": 8, "decimalMaxLength": 2},
		},
		"priceHint":   "\u5e73\u53f0\u5c06\u6536\u53d6 10% \u670d\u52a1\u8d39",
		"primaryText": "\u63d0\u4ea4\u9886\u8def\u4eba\u7533\u8bf7",
		"helperText":  "\u5ba1\u6838\u9884\u8ba1 1-3 \u4e2a\u5de5\u4f5c\u65e5",
	}
}

func defaultRoleStatusPageConfig() map[string]interface{} {
	return map[string]interface{}{
		"roleAliases": map[string]string{
			"player": "player", "expert": "expert", "master": "expert", "guide": "guide", "leader": "guide",
			"\u73a9\u5bb6": "player", "\u884c\u5bb6": "expert", "\u9886\u8def\u4eba": "guide",
		},
		"statusMap": map[string]string{
			"active": "approved", "enabled": "approved", "passed": "approved", "success": "approved",
			"waiting": "pending", "reviewing": "pending", "auditing": "pending", "pending_audit": "pending",
			"rejected_audit": "rejected", "reject": "rejected", "disabled": "disabled",
			"available": "none", "locked": "none", "unavailable": "none",
		},
		"roleMeta": map[string]interface{}{
			"expert": map[string]interface{}{"roleName": "\u884c\u5bb6", "applyTitle": "\u884c\u5bb6\u7533\u8bf7", "successAccent": "cyan", "approvedCopy": "\u4f60\u5df2\u83b7\u5f97\u884c\u5bb6\u8eab\u4efd\uff0c\u53ef\u5728\u5e73\u53f0\u5185\u4f7f\u7528\u5bf9\u5e94\u80fd\u529b", "primaryText": "\u5f00\u542f\u884c\u5bb6\u4e4b\u65c5"},
			"guide":  map[string]interface{}{"roleName": "\u9886\u8def\u4eba", "applyTitle": "\u9886\u8def\u4eba\u7533\u8bf7", "successAccent": "orange", "approvedCopy": "\u4f60\u5df2\u83b7\u5f97\u9886\u8def\u4eba\u8eab\u4efd\uff0c\u53ef\u5728\u5e73\u53f0\u5185\u4f7f\u7528\u5bf9\u5e94\u80fd\u529b", "primaryText": "\u5f00\u542f\u9886\u8def\u4eba\u4e4b\u65c5"},
		},
		"pendingTimeline": []map[string]interface{}{
			{"title": "\u63d0\u4ea4\u7533\u8bf7", "descTemplate": "\u5df2\u6210\u529f\u63d0\u4ea4{roleName}\u7533\u8bf7\u8d44\u6599", "timeField": "submittedAt", "fallbackTime": "\u5df2\u63d0\u4ea4", "state": "done"},
			{"title": "\u8d44\u6599\u521d\u5ba1", "desc": "\u5e73\u53f0\u5ba1\u6838\u56e2\u961f\u5df2\u63a5\u6536\u5e76\u5f00\u59cb\u521d\u5ba1", "timeWhenSubmitted": "\u5df2\u63a5\u6536", "fallbackTime": "\u5f85\u7cfb\u7edf\u540c\u6b65", "state": "done"},
			{"title": "\u6df1\u5ea6\u5ba1\u6838", "descByRole": map[string]string{"guide": "\u6b63\u5728\u8bc4\u4f30\u4f60\u7684\u7ec4\u5c40\u8bb0\u5f55\u3001\u4fe1\u7528\u5206\u53ca\u9886\u8def\u8ba1\u5212\u4e66", "expert": "\u6b63\u5728\u8bc4\u4f30\u4f60\u7684\u4e13\u4e1a\u80fd\u529b\u3001\u8d44\u8d28\u6750\u6599\u53ca\u670d\u52a1\u8bf4\u660e"}, "time": "\u8fdb\u884c\u4e2d...", "state": "active"},
			{"title": "\u7ed3\u679c\u901a\u77e5", "desc": "\u5ba1\u6838\u7ed3\u679c\u5c06\u901a\u8fc7\u6d88\u606f\u63a8\u9001\u901a\u77e5\u4f60", "time": "\u5f85\u5b8c\u6210", "state": "pending"},
		},
		"approvedActions": []map[string]string{
			{"iconKey": "network", "text": "\u5173\u7cfb\u7f51\u5f00\u542f", "routeKey": "relationNetwork"},
			{"iconKey": "invite", "text": "\u9080\u8bf7\u73a9\u5bb6", "routeKey": "gameInvite"},
			{"iconKey": "profile", "text": "\u5b8c\u5584\u8d44\u6599", "routeKey": "profileSystemProfileInfo"},
		},
		"texts": map[string]interface{}{
			"loadingText": "\u52a0\u8f7d\u4e2d...", "errorTitle": "\u5ba1\u6838\u72b6\u6001\u52a0\u8f7d\u5931\u8d25", "backHomeText": "\u8fd4\u56de\u9996\u9875", "retryText": "\u91cd\u8bd5",
			"pendingPageTitle": "\u5ba1\u6838\u8fdb\u5ea6", "resultPageTitle": "\u5ba1\u6838\u7ed3\u679c", "pendingTitle": "\u5ba1\u6838\u4e2d", "approvedTitle": "\u606d\u559c\u5ba1\u6838\u901a\u8fc7\uff01", "rejectedTitle": "\u5ba1\u6838\u672a\u901a\u8fc7",
			"pendingSubtitleTemplate": "{roleName}\u7533\u8bf7\u6b63\u5728\u5ba1\u6838", "approvedSubtitleTemplate": "\u4f60\u5df2\u6210\u4e3a\u300c{roleName}\u300d", "rejectedSubtitle": "\u67e5\u770b\u539f\u56e0\u5e76\u5b8c\u5584\u540e\u53ef\u518d\u6b21\u7533\u8bf7",
			"pendingDesc": "\u5e73\u53f0\u6b63\u5728\u8bc4\u4f30\u4f60\u7684\u7533\u8bf7\u8d44\u6599\uff0c\u8bf7\u8010\u5fc3\u7b49\u5f85", "rejectedDesc": "\u611f\u8c22\u4f60\u7684\u7533\u8bf7\uff0c\u4f46\u672c\u6b21\u5ba1\u6838\u672a\u901a\u8fc7", "approvedAuditDesc": "\u4f60\u7684\u7533\u8bf7\u5df2\u901a\u8fc7\u5e73\u53f0\u5ba1\u6838",
			"expectedLabel": "\u9884\u8ba1\u5b8c\u6210\u65f6\u95f4", "expectedTemplate": "\u9884\u8ba1 {expectedReviewAt} \u524d\u5b8c\u6210\u5ba1\u6838\uff0c\u5c4a\u65f6\u5c06\u901a\u8fc7\u7ad9\u5185\u6d88\u606f\u901a\u77e5\u4f60\u5ba1\u6838\u7ed3\u679c\u3002", "expectedFallback": "\u5ba1\u6838\u9884\u8ba1 1-3 \u4e2a\u5de5\u4f5c\u65e5\uff0c\u7ed3\u679c\u5c06\u901a\u8fc7\u7ad9\u5185\u6d88\u606f\u901a\u77e5\u4f60\u3002",
			"detailTitle": "\u7533\u8bf7\u8be6\u60c5", "pendingHelper": "\u5ba1\u6838\u671f\u95f4\u4f60\u53ef\u4ee5\u7ee7\u7eed\u4f7f\u7528\u73a9\u5bb6\u8eab\u4efd", "certNoLabel": "\u8ba4\u8bc1\u7f16\u53f7", "certTimePrefix": "\u8ba4\u8bc1\u65f6\u95f4: ", "giftTitle": "\u65b0\u624b\u793c\u5305",
			"reasonTitle": "\u9a73\u56de\u539f\u56e0", "suggestionTitle": "\u6539\u8fdb\u5efa\u8bae", "reapplyTitle": "\u91cd\u65b0\u7533\u8bf7", "reapplyDesc": "\u5b8c\u5584\u8d44\u6599\u540e\u53ef\u518d\u6b21\u63d0\u4ea4\u7533\u8bf7\u3002\u5efa\u8bae\u6839\u636e\u9a73\u56de\u539f\u56e0\u9010\u9879\u6539\u8fdb\uff0c\u63d0\u9ad8\u901a\u8fc7\u7387\u3002", "recordTitle": "\u7533\u8bf7\u8bb0\u5f55",
			"pendingFooterHomeText": "\u8fd4\u56de\u73a9\u5bb6\u9996\u9875", "pendingFooterBenefitsText": "\u67e5\u770b\u6743\u76ca\u5bf9\u6bd4", "rejectedHelpText": "\u67e5\u770b\u5e2e\u52a9", "rejectedImproveText": "\u5b8c\u5584\u8d44\u6599", "routeMissingText": "\u8bf7\u9009\u62e9\u53ef\u7528\u5165\u53e3",
			"fieldRoleLabel": "\u7533\u8bf7\u89d2\u8272", "fieldApplyTimeLabel": "\u7533\u8bf7\u65f6\u95f4", "fieldApplicationNoLabel": "\u7533\u8bf7\u7f16\u53f7", "fieldCurrentStatusLabel": "\u5f53\u524d\u72b6\u6001", "fieldExpectedLabel": "\u9884\u8ba1\u5b8c\u6210", "fieldRejectTimeLabel": "\u9a73\u56de\u65f6\u95f4", "fieldReapplyLabel": "\u53ef\u91cd\u65b0\u7533\u8bf7",
			"submittedFallback": "\u5df2\u63d0\u4ea4", "backendRecordFallback": "\u4ee5\u540e\u53f0\u8bb0\u5f55\u4e3a\u51c6", "applicationNoFallback": "\u5ba1\u6838\u4e2d\u751f\u6210", "pendingStatusText": "\u6df1\u5ea6\u5ba1\u6838\u4e2d", "statusFallback": "\u5f85\u786e\u8ba4", "expectedDoneFallback": "\u9884\u8ba1 1-3 \u4e2a\u5de5\u4f5c\u65e5", "reapplySuffix": " \u540e", "reapplyNotifyFallback": "\u8bf7\u5173\u6ce8\u540e\u53f0\u901a\u77e5",
			"loadFailedText": "\u5ba1\u6838\u72b6\u6001\u52a0\u8f7d\u5931\u8d25",
		},
		"defaultRejectReasons":  []string{"\u7533\u8bf7\u8d44\u6599\u6682\u672a\u8fbe\u5230\u5f53\u524d\u89d2\u8272\u5ba1\u6838\u8981\u6c42", "\u90e8\u5206\u8bc1\u660e\u6750\u6599\u6216\u8ba1\u5212\u8bf4\u660e\u4ecd\u9700\u8865\u5145\u5b8c\u5584"},
		"suggestionTemplates":   []string{"\u591a\u53c2\u4e0e\u5e73\u53f0\u7ec4\u5c40\u6d3b\u52a8\uff0c\u79ef\u7d2f\u5e26\u961f\u7ecf\u9a8c", "\u5b8c\u5584\u4e2a\u4eba\u8d44\u6599\uff0c\u63d0\u5347\u4fe1\u7528\u8bc4\u5206", "{improvePlanText}\uff0c\u8be6\u7ec6\u63cf\u8ff0\u4f60\u7684\u670d\u52a1\u4f18\u52bf", "\u83b7\u5f97\u540c\u4f34\u63a8\u8350\u80cc\u4e66\u53ef\u63d0\u5347\u5ba1\u6838\u901a\u8fc7\u7387"},
		"improvePlanTextByRole": map[string]string{"guide": "\u91cd\u65b0\u64b0\u5199\u9886\u8def\u8ba1\u5212\u4e66", "expert": "\u8865\u5145\u670d\u52a1\u8bf4\u660e"},
		"reapplyDays":           7,
	}
}

func (s *Server) submitRoleApplication(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	var req profiles.SubmitRoleApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	app, err := s.profiles.SubmitRoleApplication(userID, req)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	s.recordBehavior(userID, "submit_role_application", "role_application", app.ID, map[string]interface{}{"roleCode": app.RoleCode})
	httpx.OK(w, app)
}

func (s *Server) myRoles(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	qualification, err := s.profiles.GuideQualification(userID)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	snapshot := s.profiles.RoleSnapshot(userID)
	httpx.OK(w, map[string]interface{}{
		"roles":              snapshot.Roles,
		"roleStatusMap":      snapshot.RoleStatusMap,
		"applications":       s.profiles.RoleApplicationsByUser(userID),
		"guideQualification": qualification,
	})
}

func (s *Server) guideQualificationMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	qualification, err := s.profiles.GuideQualification(userID)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	httpx.OK(w, map[string]interface{}{"qualification": qualification, "rules": s.profiles.GuideQualificationRules()})
}

func (s *Server) applyGuide(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	var req struct {
		Reason             string  `json:"reason"`
		AbilityDescription string  `json:"abilityDescription"`
		ProofFileIDs       []int64 `json:"proofFileIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	app, err := s.profiles.SubmitRoleApplication(userID, profiles.SubmitRoleApplicationRequest{RoleCode: "guide", Reason: req.Reason, AbilityDescription: req.AbilityDescription, ProofFileIDs: req.ProofFileIDs})
	if err != nil {
		writeProfileError(w, err)
		return
	}
	s.recordBehavior(userID, "submit_guide_application", "role_application", app.ID, map[string]interface{}{"roleCode": app.RoleCode})
	httpx.OK(w, app)
}

func (s *Server) myRoleApplications(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	httpx.OK(w, map[string]interface{}{
		"items":      s.profiles.RoleApplicationsByUser(userID),
		"pageConfig": s.currentRoleApplicationPageConfig(),
	})
}

func (s *Server) adminRoleApplications(w http.ResponseWriter, r *http.Request) {
	applications := s.profiles.AllRoleApplications()
	codeByID := map[int64]string{}
	if codes, err := s.auth.AdminInviteCodes(invites.CodeFilter{}); err == nil {
		for _, code := range codes {
			codeByID[code.ID] = code.Code
		}
	}
	payload := make([]map[string]interface{}, 0, len(applications))
	for _, application := range applications {
		payload = append(payload, s.adminRoleApplicationPayload(application, codeByID))
	}
	httpx.OK(w, map[string]interface{}{"items": payload})
}

func (s *Server) adminRoleApplicationPayload(application profiles.RoleApplication, codeByID map[int64]string) map[string]interface{} {
	payload := map[string]interface{}{
		"id":                 application.ID,
		"userId":             application.UserID,
		"roleCode":           application.RoleCode,
		"status":             application.Status,
		"reason":             application.Reason,
		"abilityDescription": application.AbilityDescription,
		"proofFileIds":       application.ProofFileIDs,
		"rejectReason":       application.RejectReason,
		"reviewAdminId":      application.ReviewAdminID,
		"reviewRemark":       application.ReviewRemark,
		"certNo":             application.CertificateNo,
		"certifiedAt":        application.CertifiedAt,
		"createdAt":          application.CreatedAt,
		"updatedAt":          application.UpdatedAt,
	}
	if relation, ok, err := s.auth.InviteRelationForUser(application.UserID); err == nil && ok {
		payload["inviteRelation"] = relation
		payload["inviteCodeId"] = relation.InviteCodeID
		if code := strings.TrimSpace(codeByID[relation.InviteCodeID]); code != "" {
			payload["inviteCode"] = code
		}
		if relation.InviterUserID > 0 {
			payload["inviter"] = s.adminUserSummary(relation.InviterUserID)
		}
	}
	return payload
}

func (s *Server) adminGuideQualificationRules(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"items": s.profiles.GuideQualificationRules()})
}

func (s *Server) routeAdminGuideQualificationRulePut(w http.ResponseWriter, r *http.Request) {
	ruleID, ok := profileUserIDFromPath(w, r.URL.Path, "/api/admin/guides/qualification-rules/", "")
	if !ok {
		return
	}
	var req profiles.UpdateGuideQualificationRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	rule, err := s.profiles.UpdateGuideQualificationRule(ruleID, req)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	s.recordOperation(r, "guide_qualification_rule:update", "guide_qualification_rule", strconv.FormatInt(rule.ID, 10), map[string]interface{}{"status": rule.Status})
	httpx.OK(w, rule)
}

func (s *Server) routeAdminRoleApplicationPost(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/review") {
		s.requireAdminPermission("role:update", s.reviewRoleApplication)(w, r)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) reviewRoleApplication(w http.ResponseWriter, r *http.Request) {
	applicationID, ok := profileUserIDFromPath(w, r.URL.Path, "/api/admin/audits/role-applications/", "/review")
	if !ok {
		return
	}
	var req profiles.ReviewRoleApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	app, err := s.profiles.ReviewRoleApplication(parseInt64Header(r, "X-Admin-ID"), applicationID, req)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	s.recordOperation(r, "role_application:review", "role_application", strconv.FormatInt(app.ID, 10), map[string]interface{}{"roleCode": app.RoleCode, "status": app.Status, "remark": req.Remark})
	httpx.OK(w, app)
}

func (s *Server) adminGuideQualification(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("userId")), 10, 64)
	if err != nil || userID <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid userId")
		return
	}
	qualification, err := s.profiles.GuideQualification(userID)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	httpx.OK(w, qualification)
}

func (s *Server) updateGuideQualification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID       int64 `json:"userId"`
		ConditionMet *bool `json:"conditionMet"`
		PaymentMet   *bool `json:"paymentMet"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	qualification, err := s.profiles.UpdateGuideQualification(req.UserID, profiles.UpdateGuideQualificationRequest{ConditionMet: req.ConditionMet, PaymentMet: req.PaymentMet})
	if err != nil {
		writeProfileError(w, err)
		return
	}
	s.recordOperation(r, "guide_qualification:update", "guide_qualification", strconv.FormatInt(req.UserID, 10), map[string]interface{}{"conditionMet": qualification.ConditionMet, "paymentMet": qualification.PaymentMet, "guideOpenStatus": qualification.GuideOpenStatus})
	httpx.OK(w, qualification)
}

func profileUserIDFromPath(w http.ResponseWriter, path string, prefix string, suffix string) (int64, bool) {
	idText := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	id, err := strconv.ParseInt(strings.Trim(idText, "/"), 10, 64)
	if err != nil || id <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "user id invalid")
		return 0, false
	}
	return id, true
}

func writeProfileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, profiles.ErrExpertForbidden):
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "not expert")
	case errors.Is(err, profiles.ErrGuideForbidden):
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "not guide")
	case errors.Is(err, profiles.ErrInvalidProfile):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid profile")
	case errors.Is(err, profiles.ErrInvalidRoleApplication):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid role application")
	case errors.Is(err, profiles.ErrDuplicateRoleApplication):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "duplicate role application")
	case errors.Is(err, profiles.ErrRoleAlreadyActive):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "role already active")
	case errors.Is(err, profiles.ErrRoleApplicationCooldown):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "role application reapply cooldown")
	case errors.Is(err, profiles.ErrRoleApplicationNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "role application not found")
	case errors.Is(err, profiles.ErrRoleApplicationReviewed):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "role application already reviewed")
	case errors.Is(err, profiles.ErrWaitingGuideCondition):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "waiting_condition")
	case errors.Is(err, profiles.ErrWaitingGuidePayment):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "waiting_payment")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "profile operation failed")
	}
}
