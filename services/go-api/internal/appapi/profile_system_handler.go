package appapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/files"
	"zhw-mini/services/go-api/internal/notifications"
)

func (s *Server) getSystemProfileInfo(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	payload := s.profiles.SystemManagementConfig(userID, "profile-info", s.defaultSystemProfileInfo(userID))
	httpx.OK(w, s.withCurrentProfileAvatar(userID, payload))
}

func (s *Server) saveSystemProfileInfo(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	if personal, ok := objectField(payload, "personalInfo"); ok {
		name := strings.TrimSpace(stringField(personal, "name"))
		avatarFileID := parseFlexibleInt64(personal["avatarFileId"])
		if name != "" && s.rejectSensitiveNickname(w, name) {
			return
		}
		if name != "" || avatarFileID > 0 {
			user, _ := s.auth.UserByID(userID)
			nextName := user.Nickname
			if name != "" {
				nextName = name
			}
			if avatarFileID > 0 {
				clearAvatarFields(personal)
				clearPayloadAvatarFields(payload)
				if avatarFileID != user.AvatarFileID {
					avatarURL, ok := s.avatarURLForOwnedFile(w, userID, avatarFileID)
					if !ok {
						return
					}
					personal["pendingAvatarFileId"] = avatarFileID
					personal["pendingAvatarUrl"] = avatarURL
					personal["avatarAuditStatus"] = "pending"
					personal["avatarAuditText"] = "待审核"
					delete(personal, "avatarAuditReason")
					delete(personal, "rejectedAvatarFileId")
					delete(personal, "rejectedAvatarUrl")
				}
			}
			if name != "" {
				if _, err := s.auth.UpdateProfile(userID, nextName, user.AvatarURL, user.AvatarFileID); err != nil {
					httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid profile")
					return
				}
			}
		}
	}
	payload = mergeObjectMap(s.defaultSystemProfileInfo(userID), payload)
	payload = normalizeSystemProfilePayload(payload)
	// certifications 的状态只允许由后端认证记录生成，不能接受客户端伪造。
	payload["certifications"] = s.profileCertificationSummary(userID)
	payload = s.withCurrentProfileAvatar(userID, payload)
	saved := s.profiles.SaveSystemManagementConfig(userID, "profile-info", payload)
	s.recordBehavior(userID, "update_system_profile_info", "profile", userID, map[string]interface{}{"sections": len(saved)})
	httpx.OK(w, s.withCurrentProfileAvatar(userID, saved))
}

// normalizeSystemProfilePayload keeps the legacy profile mirror and the newer
// personalInfo/enterpriseInfo sections consistent for all clients.
func normalizeSystemProfilePayload(payload map[string]interface{}) map[string]interface{} {
	result := cloneObjectMap(payload)
	profile, _ := objectField(result, "profile")
	personal, _ := objectField(result, "personalInfo")
	enterprise, _ := objectField(result, "enterpriseInfo")
	if profile == nil {
		profile = map[string]interface{}{}
	}
	if personal == nil {
		personal = map[string]interface{}{}
	}
	if enterprise == nil {
		enterprise = map[string]interface{}{}
	}
	for _, key := range []string{"name", "phoneMasked", "contactVisibility", "hobby", "avatarText", "avatarFileId", "avatarUrl", "pendingAvatarFileId", "pendingAvatarUrl", "avatarAuditStatus", "avatarAuditText", "avatarAuditReason"} {
		if value, ok := personal[key]; ok {
			profile[key] = value
		} else if value, ok := profile[key]; ok {
			personal[key] = value
		}
	}
	for _, key := range []string{"company", "jobTitle", "businessCountText", "resources", "publicBusinessInfo"} {
		if value, ok := enterprise[key]; ok {
			profile[key] = value
		} else if value, ok := profile[key]; ok {
			enterprise[key] = value
		}
	}
	result["profile"] = profile
	result["personalInfo"] = personal
	result["enterpriseInfo"] = enterprise
	return result
}

func (s *Server) avatarURLForOwnedFile(w http.ResponseWriter, userID int64, avatarFileID int64) (string, bool) {
	file, err := s.files.Get(avatarFileID)
	if err != nil || file.UploaderID != userID || file.BizType != "avatar" {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "invalid avatar file")
		return "", false
	}
	download, err := s.files.DownloadURLForFile(file)
	if err != nil {
		if errors.Is(err, files.ErrStorageNotConfigured) {
			httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "storage base url not configured")
			return "", false
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "failed to generate avatar url")
		return "", false
	}
	return download.DownloadURL, true
}

func (s *Server) currentUserAvatar(userID int64) (int64, string) {
	user, ok := s.auth.UserByID(userID)
	if !ok {
		return 0, ""
	}
	avatarURL := strings.TrimSpace(user.AvatarURL)
	if user.AvatarFileID <= 0 {
		return 0, avatarURL
	}
	file, err := s.files.Get(user.AvatarFileID)
	if err != nil || file.UploaderID != userID || file.BizType != "avatar" {
		return user.AvatarFileID, avatarURL
	}
	if download, err := s.files.DownloadURLForFile(file); err == nil {
		avatarURL = download.DownloadURL
	}
	return user.AvatarFileID, avatarURL
}

func (s *Server) withCurrentProfileAvatar(userID int64, payload map[string]interface{}) map[string]interface{} {
	result := cloneObjectMap(payload)
	avatarFileID, avatarURL := s.currentUserAvatar(userID)
	if avatarFileID <= 0 && avatarURL == "" {
		return result
	}
	for _, sectionKey := range []string{"profile", "personalInfo"} {
		section, _ := objectField(result, sectionKey)
		if section == nil {
			section = map[string]interface{}{}
		} else {
			section = cloneObjectMap(section)
		}
		if avatarFileID > 0 {
			section["avatarFileId"] = avatarFileID
		}
		if avatarURL != "" {
			section["avatarUrl"] = avatarURL
		}
		result[sectionKey] = section
	}
	return result
}

func clearPayloadAvatarFields(payload map[string]interface{}) {
	for _, sectionKey := range []string{"profile", "personalInfo"} {
		section, ok := objectField(payload, sectionKey)
		if !ok {
			continue
		}
		clearAvatarFields(section)
		payload[sectionKey] = section
	}
}

func clearAvatarFields(section map[string]interface{}) {
	delete(section, "avatarFileId")
	delete(section, "avatarUrl")
}

func (s *Server) rejectSensitiveNickname(w http.ResponseWriter, nickname string) bool {
	if strings.TrimSpace(nickname) == "" || s.im == nil {
		return false
	}
	if _, ok := s.im.CheckSensitiveWords(nickname); ok {
		httpx.Error(w, http.StatusUnavailableForLegalReasons, 45102, "昵称涉及敏感词，不能保存")
		return true
	}
	return false
}

func (s *Server) adminAvatarAudits(w http.ResponseWriter, r *http.Request) {
	statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))
	items := make([]map[string]interface{}, 0)
	for _, config := range s.profiles.SystemManagementConfigs("profile-info") {
		personal, ok := objectField(config.Value, "personalInfo")
		if !ok {
			continue
		}
		status := strings.TrimSpace(stringField(personal, "avatarAuditStatus"))
		pendingFileID := parseFlexibleInt64(personal["pendingAvatarFileId"])
		rejectedFileID := parseFlexibleInt64(personal["rejectedAvatarFileId"])
		if status == "" && pendingFileID > 0 {
			status = "pending"
		}
		if status == "" {
			continue
		}
		if statusFilter != "" && status != statusFilter {
			continue
		}
		user, _ := s.auth.UserByID(config.UserID)
		items = append(items, map[string]interface{}{
			"userId":               config.UserID,
			"userName":             s.displayName(config.UserID, "用户 "),
			"nickname":             strings.TrimSpace(user.Nickname),
			"status":               status,
			"statusText":           stringField(personal, "avatarAuditText"),
			"currentAvatarFileId":  user.AvatarFileID,
			"currentAvatarUrl":     strings.TrimSpace(user.AvatarURL),
			"pendingAvatarFileId":  pendingFileID,
			"pendingAvatarUrl":     stringField(personal, "pendingAvatarUrl"),
			"rejectedAvatarFileId": rejectedFileID,
			"rejectedAvatarUrl":    stringField(personal, "rejectedAvatarUrl"),
		})
	}
	httpx.OK(w, map[string]interface{}{"items": items, "total": len(items)})
}

func (s *Server) routeAdminAvatarAuditPost(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/review") {
		s.requireAdminPermission("identity:update", s.reviewAvatarAudit)(w, r)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) reviewAvatarAudit(w http.ResponseWriter, r *http.Request) {
	userID, ok := idFromAdminPath(w, r.URL.Path, "/api/admin/avatar-audits/", "/review")
	if !ok {
		return
	}
	var req struct {
		Approve bool   `json:"approve"`
		Reason  string `json:"reason"`
		Remark  string `json:"remark"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	reason := strings.TrimSpace(firstNonEmpty(req.Reason, req.Remark))
	if !req.Approve && reason == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "驳回审核必须填写原因")
		return
	}
	payload := s.profiles.SystemManagementConfig(userID, "profile-info", s.defaultSystemProfileInfo(userID))
	personal, ok := objectField(payload, "personalInfo")
	if !ok {
		personal = map[string]interface{}{}
	}
	pendingFileID := parseFlexibleInt64(personal["pendingAvatarFileId"])
	if pendingFileID <= 0 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "no pending avatar")
		return
	}
	avatarURL, ok := s.avatarURLForOwnedFile(w, userID, pendingFileID)
	if !ok {
		return
	}
	if req.Approve {
		user, _ := s.auth.UserByID(userID)
		if _, err := s.auth.UpdateProfile(userID, user.Nickname, avatarURL, pendingFileID); err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid profile")
			return
		}
		personal["avatarFileId"] = pendingFileID
		personal["avatarUrl"] = avatarURL
		personal["avatarAuditStatus"] = "approved"
		personal["avatarAuditText"] = "已通过"
		delete(personal, "avatarAuditReason")
		delete(personal, "rejectedAvatarFileId")
		delete(personal, "rejectedAvatarUrl")
	} else {
		personal["avatarAuditStatus"] = "rejected"
		personal["avatarAuditText"] = "已驳回"
		personal["avatarAuditReason"] = reason
		personal["rejectedAvatarFileId"] = pendingFileID
		personal["rejectedAvatarUrl"] = avatarURL
	}
	delete(personal, "pendingAvatarFileId")
	delete(personal, "pendingAvatarUrl")
	payload["personalInfo"] = personal
	saved := s.profiles.SaveSystemManagementConfig(userID, "profile-info", payload)
	notifyTitle := "头像审核已通过"
	notifyContent := "你的头像已通过审核，已更新为正式头像。"
	notifyType := "avatar_review_approved"
	if !req.Approve {
		notifyTitle = "头像审核未通过"
		notifyContent = "你的头像未通过审核。原因：" + reason
		notifyType = "avatar_review_rejected"
	}
	s.notices.Create(notifications.CreateRequest{UserID: userID, NotifyType: notifyType, Title: notifyTitle, Content: notifyContent, BizType: "avatar_audit", BizID: userID})
	s.recordOperation(r, "avatar:review", "user", strconv.FormatInt(userID, 10), map[string]interface{}{
		"approve": req.Approve,
		"fileId":  pendingFileID,
		"reason":  reason,
	})
	httpx.OK(w, s.withCurrentProfileAvatar(userID, saved))
}

func (s *Server) getSystemSkillConfig(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	httpx.OK(w, s.systemSkillConfig(userID))
}

func (s *Server) getSystemServiceCaseDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	caseID := serviceCaseIDFromPath(r.URL.Path)
	if caseID == "" {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "case id required")
		return
	}
	httpx.OK(w, s.systemServiceCaseDetail(userID, caseID))
}

func (s *Server) saveSystemSkillConfig(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	existing := s.systemSkillConfig(userID)
	payload = mergeObjectMap(existing, payload)
	limit := s.currentExpertSkillDisplayConfig().VisibleSkillLimit
	if err := validateSystemSkillConfig(payload, limit, systemSkillVisibleCount(existing)); err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
		return
	}
	payload = normalizeSystemSkillConfigForDisplay(payload, limit)
	saved := s.profiles.SaveSystemManagementConfig(userID, "skill-config", payload)
	s.recordBehavior(userID, "update_system_skill_config", "profile", userID, map[string]interface{}{"activeTab": saved["activeTab"]})
	httpx.OK(w, saved)
}

func validateSystemSkillConfig(payload map[string]interface{}, visibleLimit int, existingVisibleCount int) error {
	allowedCount := visibleLimit
	if existingVisibleCount > allowedCount {
		// 下调上限不会删除存量显性技能；只禁止继续增加，确保存量行家资料可继续编辑。
		allowedCount = existingVisibleCount
	}
	if len(systemSkillItemList(payload["skillSlots"])) > allowedCount {
		return errors.New("显性技能数量超过后台配置上限")
	}
	if systemSkillVisibleCount(payload) > allowedCount {
		return errors.New("显性技能数量超过后台配置上限")
	}
	return nil
}

func (s *Server) getSystemFeedbackHome(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	httpx.OK(w, s.systemFeedbackHomeConfig(userID))
}

func (s *Server) submitSystemFeedback(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	var req struct {
		TypeKey    string  `json:"typeKey"`
		SessionKey string  `json:"sessionKey"`
		Content    string  `json:"content"`
		Contact    string  `json:"contact"`
		FileIDs    []int64 `json:"fileIds"`
		Quick      bool    `json:"quick"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	req.TypeKey = strings.TrimSpace(req.TypeKey)
	req.SessionKey = strings.TrimSpace(req.SessionKey)
	req.Content = strings.TrimSpace(req.Content)
	req.Contact = strings.TrimSpace(req.Contact)
	if req.TypeKey == "" || (req.Content == "" && len(req.FileIDs) == 0) || len(req.Content) > 500 || len(req.Contact) > 80 || len(req.FileIDs) > 9 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid feedback")
		return
	}
	if req.Content == "" {
		req.Content = "\u9644\u4ef6\u53cd\u9988"
	}
	records := s.systemFeedbackRecords(userID)
	id := "FB" + time.Now().Format("20060102150405") + strconv.FormatInt(int64(len(records)+1), 10)
	record := map[string]interface{}{
		"id":          id,
		"typeKey":     req.TypeKey,
		"type":        systemFeedbackTypeLabel(req.TypeKey),
		"typeTone":    systemFeedbackTypeTone(req.TypeKey),
		"sessionKey":  req.SessionKey,
		"status":      "\u5f85\u5904\u7406",
		"statusClass": "pending",
		"content":     req.Content,
		"contact":     req.Contact,
		"fileIds":     req.FileIDs,
		"quick":       req.Quick,
		"time":        time.Now().Format("2006-01-02 15:04"),
		"createdAt":   time.Now().Format(time.RFC3339),
		"images":      []interface{}{},
	}
	records = append([]map[string]interface{}{record}, records...)
	s.profiles.SaveSystemManagementConfig(userID, "feedback-records", map[string]interface{}{"items": records})
	s.recordBehavior(userID, "submit_system_feedback", "feedback", 0, map[string]interface{}{"typeKey": req.TypeKey, "recordId": id})
	successPage := s.systemFeedbackSuccessPageConfig(userID)
	httpx.OK(w, map[string]interface{}{
		"record":       record,
		"successTitle": successPage["successTitle"],
		"successDesc":  successPage["successDesc"],
		"successPage":  successPage,
	})
}

func (s *Server) systemFeedbackSuccessPageConfig(userID int64) map[string]interface{} {
	home := s.systemFeedbackHomeConfig(userID)
	if config, ok := objectField(home, "successPage"); ok {
		return config
	}
	return defaultSystemFeedbackSuccessPageConfig()
}

func (s *Server) systemFeedbackHomeConfig(userID int64) map[string]interface{} {
	defaults := s.defaultSystemFeedbackHome(userID)
	config := s.profiles.SystemManagementConfig(userID, "feedback-home", defaults)
	config = mergeObjectMap(defaults, config)

	if defaultLimits, ok := objectField(defaults, "limits"); ok {
		if limits, ok := objectField(config, "limits"); ok {
			config["limits"] = mergeObjectMap(defaultLimits, limits)
		} else {
			config["limits"] = defaultLimits
		}
	}
	if defaultSuccessPage, ok := objectField(defaults, "successPage"); ok {
		if successPage, ok := objectField(config, "successPage"); ok {
			config["successPage"] = mergeObjectMap(defaultSuccessPage, successPage)
		} else {
			config["successPage"] = defaultSuccessPage
		}
	}

	return config
}

func defaultSystemFeedbackSuccessPageConfig() map[string]interface{} {
	return map[string]interface{}{
		"successTitle":      "\u63d0\u4ea4\u6210\u529f",
		"successDesc":       "\u611f\u8c22\u60a8\u7684\u53cd\u9988\uff0c\u6211\u4eec\u4f1a\u8ba4\u771f\u9605\u8bfb\u6bcf\u4e00\u6761\u5efa\u8bae",
		"successDescSecond": "\u5904\u7406\u8fdb\u5ea6\u5c06\u901a\u8fc7\u6d88\u606f\u901a\u77e5\u60a8",
		"backHomeText":      "\u8fd4\u56de\u9996\u9875",
		"viewRecordsText":   "\u67e5\u770b\u53cd\u9988\u8bb0\u5f55",
		"reward": map[string]interface{}{
			"show":  true,
			"title": "\u83b7\u5f97\u7ecf\u9a8c\u503c\u5956\u52b1",
			"desc":  "\u4f18\u8d28\u53cd\u9988\u88ab\u91c7\u7eb3\u540e\u53ef\u83b7\u5f97\u66f4\u591a\u7ecf\u9a8c\u503c",
			"value": "+20",
		},
		"rating": map[string]interface{}{
			"title":     "\u60a8\u5bf9\u6211\u4eec\u7684\u53cd\u9988\u4f53\u9a8c\u6ee1\u610f\u5417\uff1f",
			"desc":      "0 = \u975e\u5e38\u4e0d\u6ee1\u610f\uff0c10 = \u975e\u5e38\u6ee1\u610f",
			"minLabel":  "\u975e\u5e38\u4e0d\u6ee1\u610f",
			"maxLabel":  "\u975e\u5e38\u6ee1\u610f",
			"default":   7,
			"scoreFrom": 0,
			"scoreTo":   10,
		},
		"reasonsTitle": "\u54ea\u4e9b\u65b9\u9762\u8ba9\u60a8\u6ee1\u610f\uff1f\uff08\u53ef\u591a\u9009\uff09",
		"reasons": []map[string]interface{}{
			{"value": "\u53cd\u9988\u6d41\u7a0b\u7b80\u5355"},
			{"value": "\u54cd\u5e94\u901f\u5ea6\u5feb"},
			{"value": "\u5ba2\u670d\u6001\u5ea6\u597d"},
			{"value": "\u95ee\u9898\u89e3\u51b3\u5f7b\u5e95"},
			{"value": "\u754c\u9762\u6e05\u6670\u6613\u7528"},
			{"value": "\u6709\u79ef\u5206\u6fc0\u52b1"},
		},
		"submitRatingText": "\u63d0\u4ea4\u8bc4\u4ef7",
		"ratingSavedText":  "\u8bc4\u4ef7\u5df2\u8bb0\u5f55",
	}
}

func (s *Server) getSystemFeedbackRecords(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	records := s.systemFeedbackRecords(userID)
	activeTab := strings.TrimSpace(r.URL.Query().Get("tab"))
	if activeTab == "" {
		activeTab = "all"
	}
	filtered := make([]map[string]interface{}, 0, len(records))
	for _, item := range records {
		statusClass, _ := item["statusClass"].(string)
		if activeTab == "resolved" && statusClass != "resolved" {
			continue
		}
		if activeTab == "processing" && statusClass == "resolved" {
			continue
		}
		filtered = append(filtered, item)
	}
	httpx.OK(w, map[string]interface{}{
		"activeTab": activeTab,
		"tabs": []map[string]interface{}{
			{"key": "all", "label": "\u5168\u90e8", "count": len(records)},
			{"key": "processing", "label": "\u5904\u7406\u4e2d", "count": countFeedbackByStatus(records, false)},
			{"key": "resolved", "label": "\u5df2\u89e3\u51b3", "count": countFeedbackByStatus(records, true)},
		},
		"records":    filtered,
		"allRecords": records,
	})
}

func (s *Server) getSystemFeedbackDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	recordID, parseOK := feedbackRecordIDFromPath(w, r.URL.Path, "")
	if !parseOK {
		return
	}
	record, found := s.findSystemFeedbackRecord(userID, recordID)
	if !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "feedback not found")
		return
	}
	httpx.OK(w, map[string]interface{}{
		"record":   record,
		"messages": systemFeedbackMessages(record),
	})
}

func (s *Server) appendSystemFeedbackMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	recordID, parseOK := feedbackRecordIDFromPath(w, r.URL.Path, "/messages")
	if !parseOK {
		return
	}
	var req struct {
		Content string  `json:"content"`
		FileIDs []int64 `json:"fileIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if (req.Content == "" && len(req.FileIDs) == 0) || len(req.Content) > 500 || len(req.FileIDs) > 9 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid feedback message")
		return
	}
	if req.Content == "" {
		req.Content = "\u9644\u4ef6\u8865\u5145"
	}
	records := s.systemFeedbackRecords(userID)
	for index, record := range records {
		if stringField(record, "id") != recordID {
			continue
		}
		message := map[string]interface{}{
			"id":        "MSG" + time.Now().Format("20060102150405"),
			"role":      "me",
			"avatar":    "我",
			"time":      time.Now().Format("15:04"),
			"content":   req.Content,
			"fileIds":   req.FileIDs,
			"createdAt": time.Now().Format(time.RFC3339),
		}
		messages := systemFeedbackMessages(record)
		messages = append(messages, message)
		record["messages"] = messages
		record["replyText"] = "已补充说明"
		record["replyClass"] = "processing"
		record["status"] = "处理中"
		record["statusClass"] = "processing"
		records[index] = record
		s.profiles.SaveSystemManagementConfig(userID, "feedback-records", map[string]interface{}{"items": records})
		s.recordBehavior(userID, "append_system_feedback_message", "feedback", 0, map[string]interface{}{"recordId": recordID})
		httpx.OK(w, map[string]interface{}{"record": record, "message": message, "messages": messages})
		return
	}
	httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "feedback not found")
}

func (s *Server) adminSystemFeedbackRecords(w http.ResponseWriter, r *http.Request) {
	statusClass := strings.TrimSpace(r.URL.Query().Get("statusClass"))
	items := make([]map[string]interface{}, 0)
	for _, config := range s.profiles.SystemManagementConfigs("feedback-records") {
		for _, record := range systemFeedbackRecordsFromPayload(config.Value) {
			if statusClass != "" && stringField(record, "statusClass") != statusClass {
				continue
			}
			item := cloneObjectMap(record)
			item["userId"] = config.UserID
			items = append(items, item)
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		return systemFeedbackSortTime(items[i]).After(systemFeedbackSortTime(items[j]))
	})
	s.recordOperation(r, "feedback:view", "feedback", "feedback-records", map[string]interface{}{"total": len(items), "statusClass": statusClass})
	httpx.OK(w, map[string]interface{}{"items": items, "total": len(items)})
}

func (s *Server) adminSystemFeedbackReply(w http.ResponseWriter, r *http.Request) {
	recordID, ok := adminFeedbackRecordIDFromPath(w, r.URL.Path, "/reply")
	if !ok {
		return
	}
	var req struct {
		UserID  int64  `json:"userId"`
		Content string `json:"content"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	req.Status = strings.TrimSpace(req.Status)
	if req.UserID <= 0 || req.Content == "" || len(req.Content) > 500 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid feedback reply")
		return
	}
	status, statusClass, statusOK := normalizeFeedbackReplyStatus(req.Status)
	if !statusOK {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid feedback status")
		return
	}
	records := s.systemFeedbackRecords(req.UserID)
	for index, record := range records {
		if stringField(record, "id") != recordID {
			continue
		}
		message := map[string]interface{}{
			"id":        "SVC" + time.Now().Format("20060102150405"),
			"role":      "service",
			"avatar":    "客",
			"time":      time.Now().Format("15:04"),
			"content":   req.Content,
			"createdAt": time.Now().Format(time.RFC3339),
		}
		messages := append(systemFeedbackMessages(record), message)
		record["messages"] = messages
		record["replyText"] = req.Content
		record["replyClass"] = statusClass
		record["status"] = status
		record["statusClass"] = statusClass
		record["handledAt"] = time.Now().Format(time.RFC3339)
		records[index] = record
		s.profiles.SaveSystemManagementConfig(req.UserID, "feedback-records", map[string]interface{}{"items": records})
		s.notices.Create(notifications.CreateRequest{
			UserID:     req.UserID,
			NotifyType: "feedback_replied",
			Title:      "客服已回复你的反馈",
			Content:    req.Content,
			BizType:    "feedback",
		})
		s.recordOperation(r, "feedback:reply", "feedback", recordID, map[string]interface{}{"userId": req.UserID, "statusClass": statusClass})
		httpx.OK(w, map[string]interface{}{"record": record, "message": message, "messages": messages})
		return
	}
	httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "feedback not found")
}

func (s *Server) getSystemBlockSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	httpx.OK(w, s.profiles.SystemManagementConfig(userID, "block-settings", s.defaultSystemBlockSettings(userID)))
}

func (s *Server) saveSystemBlockSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	base := mergeObjectMap(s.defaultSystemBlockSettings(userID), s.profiles.SystemManagementConfig(userID, "block-settings", s.defaultSystemBlockSettings(userID)))
	payload = mergeObjectMap(base, payload)
	saved := s.profiles.SaveSystemManagementConfig(userID, "block-settings", payload)
	s.recordBehavior(userID, "update_system_block_settings", "profile", userID, map[string]interface{}{"enabled": saved["enabled"]})
	httpx.OK(w, saved)
}

func (s *Server) getProfileSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	settings := s.profiles.SystemManagementConfig(userID, "profile-settings", s.defaultProfileSettings(userID))
	httpx.OK(w, s.normalizeProfileAccountSecurityRows(userID, settings))
}

func (s *Server) normalizeProfileAccountSecurityRows(userID int64, settings map[string]interface{}) map[string]interface{} {
	result := cloneObjectMap(settings)
	sections, _ := result["sections"].([]interface{})
	if len(sections) == 0 {
		return s.defaultProfileSettings(userID)
	}
	user, _ := s.auth.UserByID(userID)
	securityRows := map[string]map[string]interface{}{
		"wechatBind":    {"id": "wechatBind", "label": "绑定微信", "iconKey": "wechatBind", "value": map[bool]string{true: "已绑定", false: "未绑定"}[strings.TrimSpace(user.OpenID) != ""], "arrow": true, "action": "bind_wechat"},
		"loginPassword": {"id": "loginPassword", "label": "登录密码", "iconKey": "loginPassword", "value": map[bool]string{true: "已设置", false: "未设置"}[s.auth.HasPassword(userID)], "arrow": true, "action": "set_password"},
	}
	for index, rawSection := range sections {
		section, ok := rawSection.(map[string]interface{})
		if !ok || section["title"] != "账号安全" {
			continue
		}
		rows, _ := section["rows"].([]interface{})
		found := map[string]bool{}
		for rowIndex, rawRow := range rows {
			row, ok := rawRow.(map[string]interface{})
			if !ok {
				continue
			}
			id, _ := row["id"].(string)
			if dynamic, exists := securityRows[id]; exists {
				rows[rowIndex] = dynamic
				found[id] = true
			}
		}
		for id, dynamic := range securityRows {
			if !found[id] {
				rows = append(rows, dynamic)
			}
		}
		section["rows"] = rows
		sections[index] = section
		result["sections"] = sections
		return result
	}
	return s.defaultProfileSettings(userID)
}

func (s *Server) saveProfileSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	payload = mergeObjectMap(s.defaultProfileSettings(userID), payload)
	saved := s.profiles.SaveSystemManagementConfig(userID, "profile-settings", payload)
	s.recordBehavior(userID, "update_profile_settings", "profile", userID, map[string]interface{}{"sections": len(saved)})
	httpx.OK(w, saved)
}

func (s *Server) listProfileAgreements(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	httpx.OK(w, map[string]interface{}{"items": s.profileAgreementItems(userID)})
}

func (s *Server) getProfileAgreementDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	key := agreementKeyFromPath(r.URL.Path, "")
	agreement, found := s.profileAgreementByKey(userID, key)
	if !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "agreement not found")
		return
	}
	httpx.OK(w, agreement)
}

func (s *Server) signProfileAgreement(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	key := agreementKeyFromPath(r.URL.Path, "/sign")
	agreement, found := s.profileAgreementByKey(userID, key)
	if !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "agreement not found")
		return
	}
	now := time.Now().Format(time.RFC3339)
	version := firstNonEmptyString(agreement, "version")
	if version == "" {
		version = "2026-07-01"
	}
	agreement["signed"] = true
	agreement["signedAt"] = now
	agreement["signedVersion"] = version
	agreement["requiresResign"] = false
	agreements := s.profileAgreementItems(userID)
	for index, item := range agreements {
		if stringField(item, "key") == key {
			item["signed"] = true
			item["signedAt"] = now
			item["signedVersion"] = version
			item["requiresResign"] = false
			agreements[index] = item
			break
		}
	}
	s.profiles.SaveSystemManagementConfig(userID, "agreements", map[string]interface{}{"items": agreements})
	s.recordBehavior(userID, "sign_agreement", "agreement", 0, map[string]interface{}{"agreementKey": key})
	httpx.OK(w, agreement)
}

func (s *Server) defaultSystemProfileInfo(userID int64) map[string]interface{} {
	user, _ := s.auth.UserByID(userID)
	record := s.identity.Status(userID)
	name := strings.TrimSpace(user.Nickname)
	if name == "" {
		name = "\u7528\u6237" + strconv.FormatInt(userID, 10)
	}
	phoneMasked := strings.TrimSpace(record.PhoneMasked)
	if phoneMasked == "" {
		phoneMasked = "\u672a\u7ed1\u5b9a"
	}
	profile := map[string]interface{}{
		"avatarText":         avatarTextForName(name, userID),
		"name":               name,
		"phone":              phoneMasked,
		"contactVisibility":  "all",
		"hobby":              "\u672a\u586b\u5199",
		"company":            "\u672a\u586b\u5199",
		"jobTitle":           "\u672a\u586b\u5199",
		"businessCountText":  "\u672a\u8bbe\u7f6e",
		"resources":          "\u672a\u586b\u5199",
		"publicBusinessInfo": false,
	}
	return map[string]interface{}{
		"profile": profile,
		"personalInfo": map[string]interface{}{
			"avatarText":        profile["avatarText"],
			"name":              name,
			"phoneMasked":       phoneMasked,
			"contactVisibility": "all",
			"hobby":             "\u672a\u586b\u5199",
		},
		"enterpriseInfo": map[string]interface{}{
			"company":            "\u672a\u586b\u5199",
			"jobTitle":           "\u672a\u586b\u5199",
			"businessCountText":  "\u672a\u8bbe\u7f6e",
			"resources":          "\u672a\u586b\u5199",
			"publicBusinessInfo": false,
		},
		"certifications": s.profileCertificationSummary(userID),
		"visibilityOptions": []map[string]interface{}{
			{"key": "all", "label": "\u5168\u90e8\u5c55\u793a"},
			{"key": "member", "label": "\u4ec5\u4f1a\u5458\u53ef\u89c1"},
			{"key": "hidden", "label": "\u5b8c\u5168\u9690\u85cf"},
		},
		"identity": record,
	}
}

func (s *Server) profileCertificationSummary(userID int64) []map[string]interface{} {
	personalStatus := "\u672a\u8ba4\u8bc1"
	personalClass := ""
	if s.identity.IsRealnameVerified(userID) {
		personalStatus = "\u5df2\u8ba4\u8bc1"
		personalClass = "verified"
	}
	enterpriseStatus := "\u672a\u8ba4\u8bc1"
	enterpriseClass := ""
	enterpriseDesc := "\u8425\u4e1a\u6267\u7167+\u5bf9\u516c\u8d26\u6237"
	if item, ok := s.profiles.EnterpriseCertification(userID); ok {
		switch item.Status {
		case "pending":
			enterpriseStatus = "\u5ba1\u6838\u4e2d"
			enterpriseClass = "pending"
		case "approved":
			enterpriseStatus = "\u5df2\u8ba4\u8bc1"
			enterpriseClass = "verified"
		case "rejected":
			enterpriseStatus = "\u5df2\u9a73\u56de"
			enterpriseClass = "rejected"
			if strings.TrimSpace(item.RejectReason) != "" {
				enterpriseDesc = item.RejectReason
			}
		}
	}
	return []map[string]interface{}{
		{"key": "personal", "status": personalStatus, "statusClass": personalClass, "tone": "green"},
		{"key": "enterprise", "status": enterpriseStatus, "statusClass": enterpriseClass, "tone": "blue", "desc": enterpriseDesc},
	}
}

func (s *Server) defaultSystemSkillConfig(userID int64) map[string]interface{} {
	displayConfig := s.currentExpertSkillDisplayConfig()
	visibleLimit := displayConfig.VisibleSkillLimit
	skillProfile := s.profiles.AdminExpertSkill(userID)
	tags := uniqueSystemSkillTags(skillProfile.SkillTree, skillProfile.ServiceTags)
	visible := make([]map[string]interface{}, 0, len(tags))
	slots := make([]map[string]interface{}, 0, visibleLimit)
	for index, tag := range tags {
		if index >= visibleLimit {
			break
		}
		id := "skill-" + strconv.Itoa(index+1)
		item := map[string]interface{}{
			"id":             id,
			"title":          tag,
			"iconText":       "\u2605",
			"tone":           "blue",
			"badge":          "\u540e\u53f0\u540c\u6b65",
			"badgeTone":      "info",
			"visibilityText": "\u663e\u6027\u6280\u80fd \u00b7 \u73a9\u5bb6\u53ef\u89c1",
			"sourceText":     "\u884c\u5bb6\u6280\u80fd\u6863\u6848",
			"lockedAt":       skillProfile.UpdatedAt.Format("2006-01-02"),
			"caseTitle":      "\u670d\u52a1\u6848\u4f8b",
			"caseBadge":      "\u5f85\u7ed1\u5b9a",
			"caseDesc":       "\u7531\u540e\u53f0\u6280\u80fd\u6863\u6848\u751f\u6210\uff0c\u540e\u7eed\u53ef\u7ed1\u5b9a\u771f\u5b9e\u670d\u52a1\u6848\u4f8b\u3002",
			"caseDate":       skillProfile.UpdatedAt.Format("2006-01-02"),
			"casePlayers":    strconv.Itoa(s.games.StatsForUser(userID).Participated) + "\u6b21\u53c2\u4e0e",
			"ratingText":     strconv.Itoa(s.reviews.Profile(userID).CreditScore) + "\u4fe1\u7528\u5206",
		}
		visible = append(visible, item)
		slots = append(slots, map[string]interface{}{"id": id, "title": tag, "iconText": "\u2605", "tone": "blue", "active": index == 0, "empty": false})
	}
	for len(slots) < visibleLimit {
		index := len(slots) + 1
		slots = append(slots, map[string]interface{}{
			"id":       "empty-" + strconv.Itoa(index),
			"title":    "\u6dfb\u52a0\u6280\u80fd",
			"iconKey":  "plus",
			"iconText": "+",
			"tone":     "gray",
			"empty":    true,
		})
	}
	return map[string]interface{}{
		"activeTab": "visible",
		"roleSummary": map[string]interface{}{
			"roleName":        "\u884c\u5bb6",
			"maxSkillCount":   visibleLimit,
			"monthlyLimit":    3,
			"usedCount":       0,
			"remainingCount":  3,
			"configuredCount": configuredSkillCount(slots),
			"resetDate":       nextMonthDate(),
		},
		"skillSlots": slots,
		"tabs": []map[string]string{
			{"key": "visible", "label": "\u663e\u6027\u6280\u80fd"},
			{"key": "hidden", "label": "\u9690\u5f62\u6280\u80fd"},
			{"key": "cases", "label": "\u670d\u52a1\u6848\u4f8b"},
		},
		"sectionMap": map[string]interface{}{
			"visible": map[string]string{"title": "\u5df2\u914d\u7f6e\u663e\u6027\u6280\u80fd", "desc": "\u73a9\u5bb6\u53ef\u89c1\uff0c\u7528\u4e8e\u5efa\u7acb\u4fe1\u4efb"},
			"hidden":  map[string]string{"title": "\u7cfb\u7edf\u8bc6\u522b\u9690\u5f62\u6280\u80fd", "desc": "\u7531\u5c65\u7ea6\u3001\u8bc4\u4ef7\u4e0e\u590d\u76d8\u5185\u5bb9\u6c89\u6dc0"},
			"cases":   map[string]string{"title": "\u670d\u52a1\u6848\u4f8b\u6c89\u6dc0", "desc": "\u7528\u4e8e\u652f\u6491\u6280\u80fd\u6807\u7b7e\u548c\u540e\u7eed\u667a\u80fd\u63a8\u8350"},
		},
		"skillGroups": map[string]interface{}{
			"visible": visible,
			"hidden":  []map[string]interface{}{},
			"cases":   []map[string]interface{}{},
		},
		"addableSkills": defaultAddableSkills(),
		"unlockSuggestion": map[string]string{
			"title":      "\u914d\u7f6e\u663e\u6027\u6280\u80fd",
			"desc":       "\u663e\u6027\u6280\u80fd\u6570\u91cf\u7531\u5e73\u53f0\u8fd0\u8425\u89c4\u5219\u7edf\u4e00\u63a7\u5236",
			"actionText": "\u53bb\u6dfb\u52a0",
		},
	}
}

func (s *Server) systemSkillConfig(userID int64) map[string]interface{} {
	config := s.profiles.SystemManagementConfig(userID, "skill-config", s.defaultSystemSkillConfig(userID))
	return normalizeSystemSkillConfigForDisplay(config, s.currentExpertSkillDisplayConfig().VisibleSkillLimit)
}

func normalizeSystemSkillConfigForDisplay(config map[string]interface{}, visibleLimit int) map[string]interface{} {
	result := cloneObjectMap(config)
	slots := systemSkillItemList(result["skillSlots"])
	for len(slots) < visibleLimit {
		index := len(slots) + 1
		slots = append(slots, map[string]interface{}{
			"id":       "empty-" + strconv.Itoa(index),
			"title":    "\u6dfb\u52a0\u6280\u80fd",
			"iconKey":  "plus",
			"iconText": "+",
			"tone":     "gray",
			"empty":    true,
		})
	}
	result["skillSlots"] = slots
	roleSummary, _ := result["roleSummary"].(map[string]interface{})
	if roleSummary == nil {
		roleSummary = map[string]interface{}{}
	}
	roleSummary["roleName"] = "\u884c\u5bb6"
	roleSummary["maxSkillCount"] = visibleLimit
	roleSummary["configuredCount"] = configuredSkillCount(slots)
	result["roleSummary"] = roleSummary
	return result
}

func systemSkillItemList(value interface{}) []map[string]interface{} {
	items := make([]map[string]interface{}, 0)
	switch list := value.(type) {
	case []map[string]interface{}:
		for _, item := range list {
			items = append(items, cloneObjectMap(item))
		}
	case []interface{}:
		for _, raw := range list {
			if item, ok := raw.(map[string]interface{}); ok {
				items = append(items, cloneObjectMap(item))
			}
		}
	}
	return items
}

func systemSkillVisibleCount(config map[string]interface{}) int {
	groups, _ := config["skillGroups"].(map[string]interface{})
	if groups != nil {
		if visible := systemSkillItemList(groups["visible"]); len(visible) > 0 {
			return len(visible)
		}
	}
	return configuredSkillCount(systemSkillItemList(config["skillSlots"]))
}

func uniqueSystemSkillTags(groups ...[]string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0)
	for _, group := range groups {
		for _, raw := range group {
			value := strings.TrimSpace(raw)
			if value == "" {
				continue
			}
			if _, exists := seen[value]; exists {
				continue
			}
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}

func defaultAddableSkills() []map[string]interface{} {
	return []map[string]interface{}{
		{"id": "review", "title": "\u590d\u76d8\u80fd\u529b\u8005", "desc": "\u5584\u4e8e\u603b\u7ed3\u7ec4\u5c40\u8fc7\u7a0b", "iconText": "\U0001F9E0", "tone": "orange"},
		{"id": "mood", "title": "\u6c14\u6c1b\u8c03\u89e3\u5458", "desc": "\u64c5\u957f\u5316\u89e3\u5c34\u5c2c\u548c\u51b7\u573a", "iconText": "\U0001F3AA", "tone": "pink"},
		{"id": "rules", "title": "\u89c4\u5219\u8bb2\u89e3\u5458", "desc": "\u64c5\u957f\u62c6\u89e3\u6e38\u620f\u89c4\u5219", "iconText": "\U0001F4CB", "tone": "green"},
		{"id": "photo", "title": "\u6444\u5f71\u8bb0\u5f55\u8005", "desc": "\u64c5\u957f\u8bb0\u5f55\u7ec4\u5c40\u9ad8\u5149", "iconText": "\U0001F4F8", "tone": "cyan"},
	}
}

func (s *Server) defaultProfileSettings(userID int64) map[string]interface{} {
	user, _ := s.auth.UserByID(userID)
	record := s.identity.Status(userID)
	phoneMasked := strings.TrimSpace(record.PhoneMasked)
	if phoneMasked == "" {
		phoneMasked = "\u672a\u7ed1\u5b9a"
	}
	version := "v1.0.0"
	if strings.TrimSpace(user.Nickname) == "" {
		version = "v1.0.0"
	}
	return map[string]interface{}{
		"sections": []map[string]interface{}{
			{
				"title": "\u8d26\u53f7\u5b89\u5168",
				"rows": []map[string]interface{}{
					{"id": "wechatBind", "label": "\u7ed1\u5b9a\u5fae\u4fe1", "iconKey": "wechatBind", "value": map[bool]string{true: "\u5df2\u7ed1\u5b9a", false: "\u672a\u7ed1\u5b9a"}[strings.TrimSpace(user.OpenID) != ""], "arrow": true, "action": "bind_wechat"},
					{"id": "payPassword", "label": "\u652f\u4ed8\u5bc6\u7801", "iconKey": "payPassword", "value": "\u4e00\u671f\u672a\u5f00\u653e", "arrow": true, "disabledReason": "\u4e00\u671f\u672a\u63a5\u771f\u5b9e\u652f\u4ed8\uff0c\u652f\u4ed8\u5bc6\u7801\u6682\u672a\u5f00\u653e"},
					{"id": "loginPassword", "label": "\u767b\u5f55\u5bc6\u7801", "iconKey": "loginPassword", "value": map[bool]string{true: "\u5df2\u8bbe\u7f6e", false: "\u672a\u8bbe\u7f6e"}[s.auth.HasPassword(userID)], "arrow": true, "action": "set_password"},
					{"id": "phone", "label": "\u66f4\u6362\u624b\u673a\u53f7", "iconKey": "phone", "value": phoneMasked, "arrow": true, "disabledReason": "\u8bf7\u5728\u6211\u7684\u8d44\u6599\u4e2d\u66f4\u65b0\u8054\u7cfb\u65b9\u5f0f"},
					{"id": "facePay", "label": "\u6307\u7eb9/\u9762\u5bb9\u652f\u4ed8", "iconKey": "facePay", "switch": true, "enabled": true},
				},
			},
			{
				"title": "\u901a\u77e5\u8bbe\u7f6e",
				"rows": []map[string]interface{}{
					{"id": "gamePush", "label": "\u5c40\u6d88\u606f\u63a8\u9001", "iconKey": "gamePush", "switch": true, "enabled": true},
					{"id": "systemNotice", "label": "\u7cfb\u7edf\u901a\u77e5", "iconKey": "systemNotice", "switch": true, "enabled": true},
					{"id": "subscribeNotice", "label": "\u8ba2\u9605\u6d88\u606f", "iconKey": "subscribeNotice", "switch": true, "enabled": true},
					{"id": "emailNotice", "label": "\u90ae\u4ef6\u901a\u77e5", "iconKey": "emailNotice", "switch": true, "enabled": false},
					{"id": "quietHours", "label": "\u6d88\u606f\u514d\u6253\u6270", "iconKey": "quietHours", "value": "22:00 - 08:00", "arrow": true, "disabledReason": "\u6d88\u606f\u514d\u6253\u6270\u8be6\u7ec6\u914d\u7f6e\u6682\u672a\u5f00\u653e"},
				},
			},
			{
				"title": "\u9690\u79c1\u8bbe\u7f6e",
				"rows": []map[string]interface{}{
					{"id": "showGames", "label": "\u5141\u8bb8\u4ed6\u4eba\u67e5\u770b\u6211\u7684\u5c40", "iconKey": "showGames", "switch": true, "enabled": true},
					{"id": "showReviews", "label": "\u5141\u8bb8\u4ed6\u4eba\u67e5\u770b\u6211\u7684\u8bc4\u4ef7", "iconKey": "showReviews", "switch": true, "enabled": true},
					{"id": "findByPhone", "label": "\u5141\u8bb8\u901a\u8fc7\u624b\u673a\u53f7\u627e\u5230\u6211", "iconKey": "findByPhone", "switch": true, "enabled": false},
					{"id": "personalized", "label": "\u4e2a\u6027\u5316\u63a8\u9001", "iconKey": "personalized", "switch": true, "enabled": true},
					{"id": "privacySummary", "label": "\u9690\u79c1\u653f\u7b56\u6458\u8981", "iconKey": "privacySummary", "arrow": true, "agreementKey": "privacy"},
					{"id": "thirdPartyList", "label": "\u7b2c\u4e09\u65b9\u5171\u4eab\u6e05\u5355", "iconKey": "thirdPartyList", "arrow": true, "agreementKey": "third_party_sharing"},
					{"id": "collectionList", "label": "\u4fe1\u606f\u6536\u96c6\u6e05\u5355", "iconKey": "collectionList", "arrow": true, "agreementKey": "personal_info_collection"},
				},
			},
			{
				"title": "\u901a\u7528\u8bbe\u7f6e",
				"rows": []map[string]interface{}{
					{"id": "clearCache", "label": "\u6e05\u9664\u7f13\u5b58", "iconKey": "clearCache", "value": "\u70b9\u51fb\u6e05\u7406", "arrow": true, "action": "clear_cache"},
					{"id": "about", "label": "\u5173\u4e8e\u6211\u4eec", "iconKey": "about", "value": version, "arrow": true, "route": "/pages/profile/system-management/agreement-detail/index?agreement=about&title=%E5%85%B3%E4%BA%8E%E6%88%91%E4%BB%AC"},
				},
			},
		},
		"logoutText": "\u9000\u51fa\u767b\u5f55",
	}
}

func (s *Server) profileAgreementItems(userID int64) []map[string]interface{} {
	payload := s.profiles.SystemManagementConfig(userID, "agreements", map[string]interface{}{"items": defaultProfileAgreements()})
	if typed, ok := payload["items"].([]map[string]interface{}); ok && len(typed) > 0 {
		return normalizeProfileAgreements(typed)
	}
	items, _ := payload["items"].([]interface{})
	agreements := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		if agreement, ok := item.(map[string]interface{}); ok {
			agreements = append(agreements, agreement)
		}
	}
	if len(agreements) > 0 {
		return normalizeProfileAgreements(agreements)
	}
	return normalizeProfileAgreements(defaultProfileAgreements())
}

func (s *Server) profileAgreementByKey(userID int64, key string) (map[string]interface{}, bool) {
	for _, item := range s.profileAgreementItems(userID) {
		if stringField(item, "key") == key {
			return item, true
		}
	}
	return nil, false
}

func normalizeProfileAgreements(items []map[string]interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		version := firstNonEmptyString(item, "version")
		if version == "" {
			version = "2026-07-01"
			item["version"] = version
		}
		signedVersion := firstNonEmptyString(item, "signedVersion")
		signed, _ := item["signed"].(bool)
		if signedVersion == "" && signed {
			signedVersion = version
			item["signedVersion"] = signedVersion
		}
		requiresResign := signed && signedVersion != "" && signedVersion != version
		if requiresResign {
			item["signed"] = false
			item["requiresResign"] = true
		} else {
			item["requiresResign"] = false
		}
		ensureAgreementDetailTexts(item)
		result = append(result, item)
	}
	return result
}

func ensureAgreementDetailTexts(item map[string]interface{}) {
	if strings.TrimSpace(stringField(item, "signActionText")) == "" {
		item["signActionText"] = "\u540c\u610f\u5e76\u7b7e\u7f72"
	}
	if strings.TrimSpace(stringField(item, "signSuccessText")) == "" {
		item["signSuccessText"] = "\u7b7e\u7f72\u6210\u529f"
	}
	if _, ok := item["signConfirm"].(map[string]interface{}); !ok {
		item["signConfirm"] = map[string]interface{}{
			"title":       "\u786e\u8ba4\u7b7e\u7f72\u534f\u8bae\uff1f",
			"desc":        "\u7b7e\u7f72\u540e\u5c06\u89c6\u4e3a\u540c\u610f\u534f\u8bae\u5168\u90e8\u6761\u6b3e\uff0c\u534f\u8bae\u7acb\u5373\u751f\u6548",
			"cancelText":  "\u53d6\u6d88",
			"confirmText": "\u786e\u8ba4\u7b7e\u7f72",
		}
	}
}

func defaultProfileAgreements() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"key":      "user-service",
			"title":    "\u7528\u6237\u670d\u52a1\u534f\u8bae",
			"desc":     "\u5e73\u53f0\u670d\u52a1\u6761\u6b3e\u4e0e\u89c4\u5219",
			"signed":   true,
			"version":  "2026-07-01",
			"tone":     "green",
			"last":     false,
			"iconKey":  "doc",
			"sections": defaultAgreementSections(),
		},
		{
			"key":      "privacy",
			"title":    "\u9690\u79c1\u653f\u7b56",
			"desc":     "\u4e2a\u4eba\u4fe1\u606f\u4fdd\u62a4\u8bf4\u660e",
			"signed":   true,
			"version":  "2026-07-01",
			"tone":     "deep-green",
			"last":     false,
			"iconKey":  "lock",
			"sections": defaultAgreementSections(),
		},
		{
			"key":      "settlement",
			"title":    "\u5165\u9a7b\u534f\u8bae",
			"desc":     "\u670d\u52a1\u4e0e\u5206\u6da6\u534f\u8bae",
			"signed":   false,
			"version":  "2026-07-01",
			"tone":     "orange",
			"last":     true,
			"iconKey":  "box",
			"sections": defaultAgreementSections(),
		},
		{
			"key":      "third_party_sharing",
			"title":    "\u7b2c\u4e09\u65b9\u5171\u4eab\u6e05\u5355",
			"desc":     "\u7b2c\u4e09\u65b9\u670d\u52a1\u4e0e\u5171\u4eab\u573a\u666f",
			"signed":   true,
			"version":  "2026-07-01",
			"tone":     "deep-green",
			"last":     false,
			"iconKey":  "doc",
			"sections": defaultAgreementSections(),
		},
		{
			"key":      "personal_info_collection",
			"title":    "\u4fe1\u606f\u6536\u96c6\u6e05\u5355",
			"desc":     "\u5e73\u53f0\u6536\u96c6\u548c\u4f7f\u7528\u4fe1\u606f\u7684\u8bf4\u660e",
			"signed":   true,
			"version":  "2026-07-01",
			"tone":     "green",
			"last":     false,
			"iconKey":  "lock",
			"sections": defaultAgreementSections(),
		},
		{
			"key":      "about",
			"title":    "\u5173\u4e8e\u6211\u4eec",
			"desc":     "\u5e73\u53f0\u4ecb\u7ecd\u4e0e\u670d\u52a1\u8bf4\u660e",
			"signed":   true,
			"version":  "2026-07-01",
			"tone":     "green",
			"last":     true,
			"iconKey":  "doc",
			"sections": defaultAgreementSections(),
		},
	}
}

func defaultAgreementSections() []map[string]string {
	return []map[string]string{
		{"title": "\u4e00\u3001\u534f\u8bae\u8303\u56f4", "content": "\u672c\u534f\u8bae\u662f\u60a8\u4e0e\u672c\u5e73\u53f0\u4e4b\u95f4\u5173\u4e8e\u4f7f\u7528\u5e73\u53f0\u670d\u52a1\u6240\u8ba2\u7acb\u7684\u534f\u8bae\u3002\u8bf7\u60a8\u4ed4\u7ec6\u9605\u8bfb\u672c\u534f\u8bae\uff0c\u5982\u60a8\u4e0d\u540c\u610f\u672c\u534f\u8bae\u7684\u4efb\u4f55\u5185\u5bb9\uff0c\u8bf7\u505c\u6b62\u4f7f\u7528\u5e73\u53f0\u670d\u52a1\u3002"},
		{"title": "\u4e8c\u3001\u8d26\u53f7\u6ce8\u518c", "content": "\u60a8\u627f\u8bfa\u4ee5\u771f\u5b9e\u8eab\u4efd\u6ce8\u518c\u8d26\u53f7\uff0c\u5e76\u4fdd\u8bc1\u6240\u63d0\u4f9b\u7684\u4e2a\u4eba\u8d44\u6599\u771f\u5b9e\u3001\u51c6\u786e\u3001\u5b8c\u6574\u3001\u5408\u6cd5\u6709\u6548\u3002\u5982\u6709\u53d8\u52a8\uff0c\u5e94\u53ca\u65f6\u66f4\u65b0\u3002"},
		{"title": "\u4e09\u3001\u670d\u52a1\u5185\u5bb9", "content": "\u5e73\u53f0\u5411\u60a8\u63d0\u4f9b\u7ec4\u5c40\u7ba1\u7406\u3001\u6280\u80fd\u5c55\u793a\u3001\u793e\u4ea4\u4e92\u52a8\u7b49\u670d\u52a1\u3002\u60a8\u6709\u6743\u6309\u7167\u5e73\u53f0\u89c4\u5219\u4f7f\u7528\u5404\u9879\u670d\u52a1\u3002"},
		{"title": "\u56db\u3001\u7528\u6237\u884c\u4e3a\u89c4\u8303", "content": "\u60a8\u5728\u4f7f\u7528\u5e73\u53f0\u670d\u52a1\u65f6\uff0c\u5e94\u9075\u5b88\u6cd5\u5f8b\u6cd5\u89c4\uff0c\u4e0d\u5f97\u53d1\u5e03\u8fdd\u6cd5\u8fdd\u89c4\u4fe1\u606f\uff0c\u4e0d\u5f97\u4fb5\u72af\u4ed6\u4eba\u5408\u6cd5\u6743\u76ca\u3002"},
		{"title": "\u4e94\u3001\u77e5\u8bc6\u4ea7\u6743", "content": "\u5e73\u53f0\u6240\u6709\u5185\u5bb9\uff0c\u5305\u62ec\u4f46\u4e0d\u9650\u4e8e\u6587\u5b57\u3001\u56fe\u7247\u3001\u97f3\u9891\u3001\u89c6\u9891\u3001\u8f6f\u4ef6\u7b49\uff0c\u5747\u53d7\u77e5\u8bc6\u4ea7\u6743\u6cd5\u5f8b\u4fdd\u62a4\u3002"},
		{"title": "\u516d\u3001\u514d\u8d23\u58f0\u660e", "content": "\u5e73\u53f0\u4e0d\u5bf9\u56e0\u4e0d\u53ef\u6297\u529b\u6216\u7b2c\u4e09\u65b9\u539f\u56e0\u5bfc\u81f4\u7684\u670d\u52a1\u4e2d\u65ad\u627f\u62c5\u8d23\u4efb\u3002"},
		{"title": "\u4e03\u3001\u534f\u8bae\u53d8\u66f4", "content": "\u5e73\u53f0\u6709\u6743\u6839\u636e\u9700\u8981\u4fee\u6539\u672c\u534f\u8bae\uff0c\u4fee\u6539\u540e\u7684\u534f\u8bae\u5c06\u5728\u5e73\u53f0\u516c\u793a\uff0c\u516c\u793a\u671f\u6ee1\u5373\u751f\u6548\u3002"},
	}
}

func agreementKeyFromPath(path string, suffix string) string {
	text := strings.TrimSuffix(strings.TrimPrefix(path, "/api/app/profile/agreements/"), suffix)
	return strings.Trim(text, "/")
}

func objectField(payload map[string]interface{}, key string) (map[string]interface{}, bool) {
	if payload == nil {
		return nil, false
	}
	value, ok := payload[key].(map[string]interface{})
	return value, ok
}

func stringField(payload map[string]interface{}, key string) string {
	value, _ := payload[key].(string)
	return value
}

func mergeObjectMap(base map[string]interface{}, patch map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(base)+len(patch))
	for key, value := range base {
		result[key] = value
	}
	for key, value := range patch {
		result[key] = value
	}
	return result
}

func cloneObjectMap(value map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(value))
	for key, item := range value {
		result[key] = item
	}
	return result
}

func serviceCaseIDFromPath(path string) string {
	text := strings.TrimPrefix(path, "/api/app/profile/system-management/service-cases/")
	return strings.Trim(text, "/")
}

func (s *Server) systemServiceCaseDetail(userID int64, caseID string) map[string]interface{} {
	config := s.profiles.SystemManagementConfig(userID, "skill-config", s.defaultSystemSkillConfig(userID))
	caseItem := findSystemServiceCase(config, caseID)
	if caseItem == nil {
		caseItem = map[string]interface{}{
			"id":          caseID,
			"title":       "\u670d\u52a1\u6848\u4f8b",
			"caseTitle":   "\u670d\u52a1\u6848\u4f8b",
			"caseDate":    time.Now().Format("2006-01-02"),
			"casePlayers": "\u5f85\u7ed1\u5b9a",
			"caseDesc":    "\u5b8c\u6210\u7ec4\u5c40\u548c\u8bc4\u4ef7\u540e\uff0c\u8fd9\u91cc\u4f1a\u81ea\u52a8\u6c89\u6dc0\u670d\u52a1\u6848\u4f8b\u3002",
			"iconText":    "\u2605",
			"tone":        "blue",
		}
	}
	score := serviceCaseScore(caseItem, s.reviews.Profile(userID).CreditScore)
	title := firstNonEmptyString(caseItem, "caseTitle", "title")
	if title == "" {
		title = "\u670d\u52a1\u6848\u4f8b"
	}
	date := firstNonEmptyString(caseItem, "caseDate", "lockedAt")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	playersText := firstNonEmptyString(caseItem, "casePlayers")
	if playersText == "" {
		participated := s.games.StatsForUser(userID).Participated
		if participated > 0 {
			playersText = strconv.Itoa(participated) + "\u6b21\u53c2\u4e0e"
		} else {
			playersText = "\u5f85\u7ed1\u5b9a"
		}
	}
	return map[string]interface{}{
		"caseInfo": map[string]interface{}{
			"id":           caseID,
			"title":        title,
			"date":         date,
			"playersText":  playersText,
			"totalPlayers": serviceCasePlayers(playersText),
			"ratingText":   serviceCaseRatingText(caseItem, score),
			"iconText":     firstNonEmptyString(caseItem, "iconText"),
			"tone":         firstNonEmptyString(caseItem, "tone"),
		},
		"rating": map[string]interface{}{
			"score": serviceCaseScoreText(score),
			"tags":  serviceCaseTags(caseItem),
		},
		"detailSections": serviceCaseDetailSections(caseItem),
		"players":        serviceCasePlayersList(caseItem),
	}
}

func findSystemServiceCase(config map[string]interface{}, caseID string) map[string]interface{} {
	groups, ok := objectField(config, "skillGroups")
	if !ok {
		return nil
	}
	for _, groupKey := range []string{"cases", "visible"} {
		if item := findCaseInValue(groups[groupKey], caseID); item != nil {
			return item
		}
	}
	return nil
}

func findCaseInValue(value interface{}, caseID string) map[string]interface{} {
	items, ok := value.([]interface{})
	if ok {
		for _, raw := range items {
			if item, ok := raw.(map[string]interface{}); ok && stringField(item, "id") == caseID {
				return item
			}
		}
	}
	typed, ok := value.([]map[string]interface{})
	if ok {
		for _, item := range typed {
			if stringField(item, "id") == caseID {
				return item
			}
		}
	}
	return nil
}

func firstNonEmptyString(payload map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(stringField(payload, key)); value != "" {
			return value
		}
	}
	return ""
}

func serviceCaseScore(payload map[string]interface{}, creditScore int) float64 {
	switch value := payload["rating"].(type) {
	case float64:
		if value > 0 {
			return value
		}
	case int:
		if value > 0 {
			return float64(value)
		}
	case string:
		if parsed, err := strconv.ParseFloat(strings.TrimSuffix(value, "\u5206"), 64); err == nil && parsed > 0 {
			return parsed
		}
	}
	if creditScore <= 0 {
		return 5
	}
	score := float64(creditScore) / 20
	if score < 1 {
		return 1
	}
	if score > 5 {
		return 5
	}
	return score
}

func serviceCaseScoreText(score float64) string {
	return strconv.FormatFloat(score, 'f', 1, 64)
}

func serviceCaseRatingText(payload map[string]interface{}, score float64) string {
	if text := firstNonEmptyString(payload, "ratingText"); text != "" {
		return text
	}
	return serviceCaseScoreText(score) + "\u5206"
}

func serviceCasePlayers(playersText string) int {
	for _, field := range strings.FieldsFunc(playersText, func(r rune) bool { return r < '0' || r > '9' }) {
		if value, err := strconv.Atoi(field); err == nil && value > 0 {
			return value
		}
	}
	return 0
}

func serviceCaseTags(payload map[string]interface{}) []string {
	tags := stringListField(payload, "tags")
	for _, key := range []string{"linkedSkillTitle", "badge", "sourceText"} {
		if value := firstNonEmptyString(payload, key); value != "" {
			tags = append(tags, value)
		}
	}
	return tags
}

func serviceCaseDetailSections(payload map[string]interface{}) []map[string]string {
	if sections := serviceCaseConfiguredSections(payload["detailSections"]); len(sections) > 0 {
		return sections
	}
	sections := make([]map[string]string, 0, 3)
	if text := firstNonEmptyString(payload, "caseDesc", "desc"); text != "" {
		sections = append(sections, map[string]string{"label": "\u670d\u52a1\u4eae\u70b9\uff1a", "text": text})
	}
	if text := firstNonEmptyString(payload, "sourceText", "visibilityText"); text != "" {
		sections = append(sections, map[string]string{"label": "\u80fd\u529b\u6765\u6e90\uff1a", "text": text})
	}
	if text := firstNonEmptyString(payload, "feedbackText"); text != "" {
		sections = append(sections, map[string]string{"label": "\u73a9\u5bb6\u53cd\u9988\uff1a", "text": text})
	}
	return sections
}

func serviceCaseConfiguredSections(value interface{}) []map[string]string {
	items, ok := value.([]interface{})
	if !ok {
		return nil
	}
	sections := make([]map[string]string, 0, len(items))
	for _, raw := range items {
		item, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		label := firstNonEmptyString(item, "label")
		text := firstNonEmptyString(item, "text")
		if label != "" && text != "" {
			sections = append(sections, map[string]string{"label": label, "text": text})
		}
	}
	return sections
}

func serviceCasePlayersList(payload map[string]interface{}) []map[string]interface{} {
	items, ok := payload["players"].([]interface{})
	if !ok {
		return []map[string]interface{}{}
	}
	players := make([]map[string]interface{}, 0, len(items))
	for index, raw := range items {
		item, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		name := firstNonEmptyString(item, "name")
		desc := firstNonEmptyString(item, "desc")
		if name == "" {
			continue
		}
		players = append(players, map[string]interface{}{
			"name": name,
			"desc": desc,
			"last": index == len(items)-1,
		})
	}
	return players
}

func stringListField(payload map[string]interface{}, key string) []string {
	raw, ok := payload[key].([]interface{})
	if !ok {
		if typed, ok := payload[key].([]string); ok {
			return append([]string{}, typed...)
		}
		return []string{}
	}
	values := make([]string, 0, len(raw))
	for _, item := range raw {
		if value := strings.TrimSpace(stringField(map[string]interface{}{"value": item}, "value")); value != "" {
			values = append(values, value)
		}
	}
	return values
}

func configuredSkillCount(slots []map[string]interface{}) int {
	count := 0
	for _, slot := range slots {
		if empty, _ := slot["empty"].(bool); !empty {
			count++
		}
	}
	return count
}

func nextMonthDate() string {
	now := time.Now()
	return time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
}

func (s *Server) defaultSystemFeedbackHome(userID int64) map[string]interface{} {
	return map[string]interface{}{
		"activeType":    "feature",
		"activeSession": "general",
		"feedbackTypes": []map[string]string{
			{"key": "feature", "label": "\u529f\u80fd\u5efa\u8bae", "iconKey": "pencil"},
			{"key": "problem", "label": "\u95ee\u9898\u53cd\u9988", "iconKey": "alert"},
			{"key": "experience", "label": "\u4f53\u9a8c\u4f18\u5316", "iconKey": "star"},
			{"key": "game", "label": "\u7ec4\u5c40\u76f8\u5173", "iconKey": "problem"},
			{"key": "expert", "label": "\u884c\u5bb6\u76f8\u5173", "iconKey": "problem"},
			{"key": "points", "label": "\u79ef\u5206/\u63d0\u73b0", "iconKey": "notify"},
			{"key": "other", "label": "\u5176\u4ed6", "iconKey": "alert"},
		},
		"sessions": systemFeedbackSessions(s.games.ApplicationsForUser(userID), s.games.InvitationsForUser(userID)),
		"quickTypes": []map[string]interface{}{
			{"key": "problem", "label": "\u9047\u5230\u95ee\u9898", "iconKey": "problem", "tone": "red"},
			{"key": "feature", "label": "\u529f\u80fd\u5efa\u8bae", "iconKey": "pencil", "tone": "cyan"},
			{"key": "experience", "label": "\u4f53\u9a8c\u4f18\u5316", "iconKey": "star", "tone": "gold"},
			{"key": "other", "label": "\u5176\u4ed6", "iconKey": "alert", "tone": "gray"},
		},
		"quickActions": []map[string]string{
			{"key": "feature", "label": "\u529f\u80fd\u5efa\u8bae", "iconKey": "pencil"},
			{"key": "problem", "label": "\u95ee\u9898\u53cd\u9988", "iconKey": "alert"},
			{"key": "screenshot", "label": "\u622a\u56fe\u53cd\u9988", "iconKey": "image"},
		},
		"limits": map[string]interface{}{
			"contentMaxLength":       500,
			"fileMaxCount":           9,
			"uploadNote":             "\u652f\u6301 JPG/PNG \u56fe\u7247\u548c\u8bed\u97f3\u6587\u4ef6\uff0c\u5355\u4e2a\u9644\u4ef6\u4e0d\u8d85\u8fc7 5MB\uff0c\u6700\u591a 9 \u4e2a\u9644\u4ef6",
			"uploadFullText":         "\u6700\u591a\u4e0a\u4f20 9 \u4e2a\u9644\u4ef6",
			"uploadSelectedTemplate": "\u5df2\u9009\u62e9 {selected}/{max} \u4e2a\u9644\u4ef6",
		},
		"successPage": defaultSystemFeedbackSuccessPageConfig(),
	}
}

func systemFeedbackSessions(apps interface{}, invitations interface{}) []map[string]string {
	return []map[string]string{
		{"key": "general", "title": "\u901a\u7528\u53cd\u9988", "meta": "\u4e0d\u5173\u8054\u5177\u4f53\u7ec4\u5c40"},
		{"key": "latest_game", "title": "\u6700\u8fd1\u7ec4\u5c40", "meta": "\u53ef\u5728\u63d0\u4ea4\u65f6\u8865\u5145\u8bf4\u660e"},
	}
}

func (s *Server) systemFeedbackRecords(userID int64) []map[string]interface{} {
	payload := s.profiles.SystemManagementConfig(userID, "feedback-records", map[string]interface{}{"items": []map[string]interface{}{}})
	return systemFeedbackRecordsFromPayload(payload)
}

func systemFeedbackRecordsFromPayload(payload map[string]interface{}) []map[string]interface{} {
	items, _ := payload["items"].([]interface{})
	records := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		if record, ok := item.(map[string]interface{}); ok {
			records = append(records, record)
		}
	}
	if len(records) > 0 {
		return records
	}
	if typed, ok := payload["items"].([]map[string]interface{}); ok {
		return typed
	}
	return []map[string]interface{}{}
}

func (s *Server) findSystemFeedbackRecord(userID int64, recordID string) (map[string]interface{}, bool) {
	for _, record := range s.systemFeedbackRecords(userID) {
		if stringField(record, "id") == recordID {
			return record, true
		}
	}
	return nil, false
}

func systemFeedbackMessages(record map[string]interface{}) []map[string]interface{} {
	items, _ := record["messages"].([]interface{})
	messages := make([]map[string]interface{}, 0, len(items)+2)
	for _, item := range items {
		if message, ok := item.(map[string]interface{}); ok {
			messages = append(messages, message)
		}
	}
	if typed, ok := record["messages"].([]map[string]interface{}); ok && len(typed) > 0 {
		return typed
	}
	if len(messages) > 0 {
		return messages
	}
	createdAt := stringField(record, "time")
	content := stringField(record, "content")
	if content != "" {
		messages = append(messages, map[string]interface{}{
			"id":      "initial",
			"role":    "me",
			"avatar":  "我",
			"time":    createdAt,
			"content": content,
		})
	}
	if reply := stringField(record, "replyText"); reply != "" {
		messages = append(messages, map[string]interface{}{
			"id":      "service-reply",
			"role":    "service",
			"avatar":  "客",
			"time":    createdAt,
			"content": reply,
		})
	}
	return messages
}

func feedbackRecordIDFromPath(w http.ResponseWriter, path string, suffix string) (string, bool) {
	text := strings.TrimSuffix(strings.TrimPrefix(path, "/api/app/profile/system-management/feedback-records/"), suffix)
	id := strings.Trim(text, "/")
	if id == "" {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "feedback id required")
		return "", false
	}
	return id, true
}

func adminFeedbackRecordIDFromPath(w http.ResponseWriter, path string, suffix string) (string, bool) {
	text := strings.TrimSuffix(strings.TrimPrefix(path, "/api/admin/feedback-records/"), suffix)
	id := strings.Trim(text, "/")
	if id == "" {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "feedback id required")
		return "", false
	}
	return id, true
}

func normalizeFeedbackReplyStatus(value string) (string, string, bool) {
	switch value {
	case "", "processing":
		return "处理中", "processing", true
	case "pending":
		return "待处理", "pending", true
	case "resolved":
		return "已解决", "resolved", true
	default:
		return "", "", false
	}
}

func systemFeedbackSortTime(record map[string]interface{}) time.Time {
	for _, key := range []string{"handledAt", "createdAt"} {
		if value := stringField(record, key); value != "" {
			if parsed, err := time.Parse(time.RFC3339, value); err == nil {
				return parsed
			}
		}
	}
	if value := stringField(record, "time"); value != "" {
		if parsed, err := time.Parse("2006-01-02 15:04", value); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func systemFeedbackTypeLabel(value string) string {
	switch value {
	case "problem":
		return "\u95ee\u9898\u53cd\u9988"
	case "experience":
		return "\u4f53\u9a8c\u4f18\u5316"
	case "game":
		return "\u7ec4\u5c40\u76f8\u5173"
	case "expert":
		return "\u884c\u5bb6\u76f8\u5173"
	case "points":
		return "\u79ef\u5206/\u63d0\u73b0"
	case "other":
		return "\u5176\u4ed6"
	default:
		return "\u529f\u80fd\u5efa\u8bae"
	}
}

func systemFeedbackTypeTone(value string) string {
	switch value {
	case "problem":
		return "red"
	case "experience":
		return "gold"
	case "other":
		return "gray"
	default:
		return "blue"
	}
}

func countFeedbackByStatus(items []map[string]interface{}, resolved bool) int {
	count := 0
	for _, item := range items {
		statusClass, _ := item["statusClass"].(string)
		if resolved && statusClass == "resolved" {
			count++
		}
		if !resolved && statusClass != "resolved" {
			count++
		}
	}
	return count
}

func (s *Server) defaultSystemBlockSettings(userID int64) map[string]interface{} {
	keywords := []string{"\u57f9\u8bad", "\u8bfe\u7a0b", "\u6536\u8d39\u6559\u5b66", "\u52a0\u5fae\u4fe1", "\u79c1\u4e0b"}
	whitelist := []map[string]interface{}{
		{"id": "guide-" + strconv.FormatInt(userID, 10), "name": s.displayName(userID, "\u7528\u6237"), "role": "\u6211", "reason": "\u9ed8\u8ba4\u4e0d\u53d7\u5c4f\u853d"},
	}
	renewalDays := 30
	return map[string]interface{}{
		"enabled":        true,
		"protectionMode": "hard",
		"renewalDays":    renewalDays,
		"summary": map[string]interface{}{
			"protectedUserText": strconv.Itoa(len(s.connections.My(userID))) + "\u4f4d\u7528\u6237",
			"blockedExpertText": strconv.Itoa(len(keywords)) + "\u4e2a\u5173\u952e\u8bcd",
			"renewalDaysText":   strconv.Itoa(renewalDays) + "\u5929",
		},
		"stats": []map[string]interface{}{
			{"value": len(s.connections.My(userID)), "label": "\u4fdd\u62a4\u7528\u6237"},
			{"value": len(keywords), "label": "\u5173\u952e\u8bcd"},
			{"value": "100%", "label": "\u751f\u6548\u7387"},
		},
		"configRows": []map[string]interface{}{
			{"key": "protection", "title": "\u4fdd\u62a4\u6a21\u5f0f", "desc": "\u5f53\u524d\uff1a\u786c\u4fdd\u62a4", "badge": "\u5df2\u5f00\u542f"},
			{"key": "scene", "title": "\u5206\u573a\u666f\u914d\u7f6e", "desc": "\u63a8\u8350/\u9644\u8fd1/\u5217\u8868", "badge": "4\u5f00"},
			{"key": "whitelist", "title": "\u767d\u540d\u5355", "desc": strconv.Itoa(len(whitelist)) + "\u4f4d\u4e0d\u53d7\u4fdd\u62a4", "badge": strconv.Itoa(len(whitelist)) + "/20"},
		},
		"manageRows": []map[string]interface{}{
			{"key": "users", "title": "\u7528\u6237\u5c4f\u853d", "desc": "\u53cc\u65b9\u4e92\u4e0d\u53ef\u89c1", "badge": "0\u4eba"},
			{"key": "keywords", "title": "\u5173\u952e\u8bcd\u5c4f\u853d", "desc": "\u81ea\u52a8\u8fc7\u6ee4\u5185\u5bb9", "badge": strconv.Itoa(len(keywords)) + "/20"},
		},
		"rules": []map[string]string{
			{"prefix": "\u4fdd\u62a4\u671f\u9ed8\u8ba4", "strong": "30\u5929", "suffix": ""},
			{"prefix": "\u914d\u7f6e\u53d8\u66f4", "strong": "5\u79d2\u5185", "suffix": "\u5bf9\u65b0\u8bf7\u6c42\u751f\u6548"},
		},
		"protectionModes": []map[string]interface{}{
			{
				"key": "hard", "title": "\u786c\u4fdd\u62a4", "desc": "\u5b8c\u5168\u8fc7\u6ee4\uff0c\u7528\u6237\u65e0\u611f\u77e5",
				"rules": []map[string]string{
					{"prefix": "", "strong": "\u5b8c\u5168\u8fc7\u6ee4", "suffix": "\uff0c\u540c\u7c7b\u884c\u5bb6\u5185\u5bb9\u4e0d\u5c55\u793a"},
					{"prefix": "\u88ab\u4fdd\u62a4\u7528\u6237 ", "strong": "\u96f6\u66dd\u5149", "suffix": ""},
					{"prefix": "\u9002\u5408\u7ade\u4e89\u6fc0\u70c8\u7684\u540c\u57ce\u5e02", "strong": "", "suffix": ""},
				},
			},
			{
				"key": "soft", "title": "\u8f6f\u4fdd\u62a4", "desc": "\u6392\u5e8f\u964d\u6743\u00d70.1\uff0c\u63a8\u81f3\u7b2c5\u9875\u540e",
				"rules": []map[string]string{
					{"prefix": "\u6392\u5e8f\u6743\u91cd ", "strong": "\u00d70.1", "suffix": "\uff0c\u5927\u5e45\u964d\u6743"},
					{"prefix": "\u7ffb\u9875\u81f3\u5c40\u5217\u8868\u7b2c5\u9875\u540e", "strong": "\u4e0d\u53d7\u4fdd\u62a4", "suffix": ""},
					{"prefix": "\u9002\u5408\u5185\u5bb9\u4e0d\u8db3\u6216\u8fc7\u6e21\u671f", "strong": "", "suffix": ""},
				},
			},
		},
		"renewalOptions": []map[string]interface{}{
			{"days": 30, "label": "\u6807\u51c6\u5468\u671f"},
			{"days": 60, "label": "\u53cc\u500d\u4fdd\u62a4"},
			{"days": 90, "label": "\u5b63\u5ea6\u4fdd\u62a4"},
		},
		"renewalRules": []map[string]string{
			{"prefix": "\u4fdd\u62a4\u671f\u6700\u957f ", "strong": "90\u5929", "suffix": "\uff0c\u5230\u671f\u9700\u91cd\u65b0\u7eed\u671f"},
			{"prefix": "\u7eed\u671f\u540e\u7acb\u5373\u751f\u6548\uff0c", "strong": "5\u79d2\u5185", "suffix": " \u8986\u76d6\u6240\u6709\u65b0\u8bf7\u6c42"},
			{"prefix": "\u5230\u671f\u524d ", "strong": "3\u5929", "suffix": " \u5c06\u53d1\u9001\u63d0\u9192\u901a\u77e5"},
		},
		"keywords":        keywords,
		"suggestions":     []string{"\u5fae\u5546", "\u76f4\u9500", "\u5237\u5355", "\u8d37\u6b3e", "\u517c\u804c", "\u4ee3\u7406", "\u62c9\u7fa4", "\u63a8\u5e7f"},
		"keywordMaxCount": 20,
		"keywordStats": []map[string]interface{}{
			{"value": len(keywords), "label": "\u5df2\u8bbe\u7f6e"},
			{"value": 20, "label": "\u4e0a\u9650", "color": "gold"},
			{"value": "\u6a21\u7cca", "label": "\u5339\u914d\u6a21\u5f0f", "color": "green"},
		},
		"keywordRules": []map[string]string{
			{"prefix": "\u652f\u6301", "strong": "\u6a21\u7cca\u5339\u914d", "suffix": "\uff0c\u5982\u201c\u57f9\u8bad\u201d\u5339\u914d\u201c\u57f9\u8bad\u673a\u6784\u201d\u201c\u57f9\u8bad\u8bfe\u7a0b\u201d"},
			{"prefix": "\u6700\u591a\u53ef\u8bbe", "strong": "20\u4e2a", "suffix": "\u5173\u952e\u8bcd"},
			{"prefix": "\u5173\u952e\u8bcd\u5c4f\u853d\u4ec5\u5f71\u54cd\u5185\u5bb9\u5c55\u793a\uff0c", "strong": "\u4e0d\u5f71\u54cd\u7528\u6237\u95f4\u4ea4\u4e92", "suffix": ""},
			{"prefix": "\u751f\u6548\u8303\u56f4\uff1a\u5c40\u6807\u9898\u3001\u63cf\u8ff0\u3001\u8bc4\u8bba\u3001\u79c1\u4fe1\u5185\u5bb9", "strong": "", "suffix": ""},
		},
		"whitelist":           whitelist,
		"whitelistCandidates": []map[string]interface{}{},
		"whitelistRules": []map[string]string{
			{"prefix": "\u767d\u540d\u5355\u884c\u5bb6", "strong": "\u4e0d\u53d7\u4fdd\u62a4", "suffix": "\u5f71\u54cd"},
			{"prefix": "\u6700\u591a\u6dfb\u52a0", "strong": "20\u4f4d", "suffix": ""},
		},
		"blockedUsers":    []map[string]interface{}{},
		"blockCandidates": []map[string]interface{}{},
		"blockRules": []map[string]string{
			{"prefix": "\u53ef\u901a\u8fc7\u7528\u6237\u4e3b\u9875\u53f3\u4e0a\u89d2\u83dc\u5355\u5feb\u901f\u5c4f\u853d", "strong": "", "suffix": ""},
			{"prefix": "\u5c4f\u853d\u540e\u53cc\u65b9", "strong": "\u4e92\u4e0d\u53ef\u89c1", "suffix": "\uff0c\u5386\u53f2\u4e92\u52a8\u8bb0\u5f55\u4fdd\u7559"},
			{"prefix": "\u89e3\u9664\u5c4f\u853d\u540e", "strong": "24\u5c0f\u65f6\u51b7\u5374\u671f", "suffix": "\u624d\u80fd\u518d\u6b21\u5c4f\u853d"},
			{"prefix": "\u5c4f\u853d\u4eba\u6570\u4e0a\u9650", "strong": "100\u4eba", "suffix": ""},
		},
		"scenes": []map[string]interface{}{
			{"key": "recommend", "title": "\u63a8\u8350\u573a\u666f", "desc": "\u9996\u9875/\u53d1\u73b0\u9875\u63a8\u8350", "mode": "hard", "enabled": true},
			{"key": "nearby", "title": "\u9644\u8fd1\u573a\u666f", "desc": "LBS\u5730\u7406\u4f4d\u7f6e\u63a8\u8350", "mode": "soft", "enabled": true},
			{"key": "message", "title": "\u79c1\u4fe1\u573a\u666f", "desc": "\u5c40\u5185\u4e0e\u79c1\u4fe1\u5185\u5bb9", "mode": "hard", "enabled": true},
			{"key": "list", "title": "\u5217\u8868\u6d4f\u89c8", "desc": "\u5c40\u5217\u8868/\u884c\u5bb6\u5217\u8868", "mode": "none", "enabled": true},
		},
		"sceneRules": []map[string]string{
			{"strong": "\u786c\u4fdd\u62a4", "suffix": " = \u5b8c\u5168\u8fc7\u6ee4"},
			{"strong": "\u8f6f\u4fdd\u62a4", "suffix": " = \u964d\u6743\u81f3\u7b2c5\u9875\u540e"},
			{"strong": "\u4e0d\u8fc7\u6ee4", "suffix": " = \u6b63\u5e38\u5c55\u793a"},
		},
	}
}
