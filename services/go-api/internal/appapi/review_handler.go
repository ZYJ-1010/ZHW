package appapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/notifications"
	"zhw-mini/services/go-api/internal/points"
	"zhw-mini/services/go-api/internal/reviews"
)

type reviewService interface {
	MarkGameReviewable(gameID int64)
	AwardCompletedGame(gameID int64) []reviews.GrowthProfile
	AwardTaskReward(userID int64, taskCode string, points int, experience int) reviews.GrowthProfile
	Todos(userID int64) ([]reviews.Todo, error)
	Submit(userID int64, req reviews.SubmitRequest) (reviews.Review, reviews.GrowthProfile, error)
	SubmitWithPoints(userID int64, req reviews.SubmitRequest, rewardPoints int) (reviews.Review, reviews.GrowthProfile, error)
	GrowthRules() reviews.GrowthRules
	MyIntents(userID int64) []reviews.Review
	AllReviews() []reviews.Review
	Profile(userID int64) reviews.GrowthProfile
	Footprints(userID int64) []reviews.Footprint
	AllFootprints() []reviews.Footprint
	DeductCredit(userID int64, gameID int64, reason string) reviews.CreditLog
	CreditDeductionValue(reason string) int
	RestoreCredit(userID int64, gameID int64, reason string, amount int) reviews.CreditLog
	TraceByUser(userID int64) reviews.Trace
	TraceByGame(gameID int64) reviews.Trace
	GameReviewComplete(gameID int64) bool
}

type creditRuleRepository interface {
	ListCreditDeductionRules(ctx context.Context) ([]reviews.CreditDeductionRule, error)
	UpsertCreditDeductionRule(ctx context.Context, rule reviews.CreditDeductionRule) (reviews.CreditDeductionRule, error)
}

const (
	reviewCompleteConfigKey      = "review.complete_config"
	reviewPageConfigKey          = "review.page_config"
	growthAchievementConfigKey   = "growth.achievement_config"
	defaultAchievementIcon       = "/pages/profile/footprint/achievements/assets/icon-star.png"
	defaultAchievementOnlineText = "人成长中"
)

type reviewPageSatisfactionOptionDTO struct {
	ID    string `json:"id"`
	Emoji string `json:"emoji"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

type reviewPageRoleConfigDTO struct {
	ID          string   `json:"id"`
	AvatarText  string   `json:"avatarText"`
	AvatarTheme string   `json:"avatarTheme"`
	Title       string   `json:"title"`
	Desc        string   `json:"desc"`
	RatingTitle string   `json:"ratingTitle"`
	TagTitle    string   `json:"tagTitle"`
	Tags        []string `json:"tags"`
	Placeholder string   `json:"placeholder"`
}

type reviewPageConfigDTO struct {
	NavTitle                  string                             `json:"navTitle"`
	SkipText                  string                             `json:"skipText"`
	StatusTitle               string                             `json:"statusTitle"`
	StatusDesc                string                             `json:"statusDesc"`
	SatisfactionQuestion      string                             `json:"satisfactionQuestion"`
	SatisfactionOptions       []reviewPageSatisfactionOptionDTO  `json:"satisfactionOptions"`
	StoryTitle                string                             `json:"storyTitle"`
	AITip                     string                             `json:"aiTip"`
	StoryPlaceholder          string                             `json:"storyPlaceholder"`
	StoryMaxLength            int                                `json:"storyMaxLength"`
	AISummaryText             string                             `json:"aiSummaryText"`
	RatingHint                string                             `json:"ratingHint"`
	NPSHeadTitle              string                             `json:"npsHeadTitle"`
	NPSQuestion               string                             `json:"npsQuestion"`
	NPSLowLabel               string                             `json:"npsLowLabel"`
	NPSHighLabel              string                             `json:"npsHighLabel"`
	SubmitText                string                             `json:"submitText"`
	SubmitNote                string                             `json:"submitNote"`
	SkipToast                 string                             `json:"skipToast"`
	SubmitSuccessText         string                             `json:"submitSuccessText"`
	NoReviewTargetText        string                             `json:"noReviewTargetText"`
	MissingTargetText         string                             `json:"missingTargetText"`
	MissingScoreText          string                             `json:"missingScoreText"`
	SubmitFailedText          string                             `json:"submitFailedText"`
	DefaultSummary            string                             `json:"defaultSummary"`
	AgainIntentBySatisfaction map[string]string                  `json:"againIntentBySatisfaction"`
	RoleConfigs               map[string]reviewPageRoleConfigDTO `json:"roleConfigs"`
	Version                   string                             `json:"version"`
}

type reviewCompleteBenefitDTO struct {
	IconText string `json:"iconText"`
	Theme    string `json:"theme"`
	Title    string `json:"title"`
	Desc     string `json:"desc"`
	Order    int    `json:"order"`
	Visible  bool   `json:"visible"`
}

type reviewCompletePlayOptionDTO struct {
	ID      string `json:"id"`
	Theme   string `json:"theme"`
	Title   string `json:"title"`
	Desc    string `json:"desc"`
	Intent  string `json:"intent"`
	Route   string `json:"route"`
	Order   int    `json:"order"`
	Visible bool   `json:"visible"`
}

type reviewCompleteConfigDTO struct {
	Reward      map[string]interface{}        `json:"reward"`
	Benefits    []reviewCompleteBenefitDTO    `json:"benefits"`
	PlayOptions []reviewCompletePlayOptionDTO `json:"playOptions"`
	Version     string                        `json:"version"`
}

type growthAchievementFilterDTO struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Order int    `json:"order"`
}

type growthAchievementItemDTO struct {
	ID              string   `json:"id"`
	Code            string   `json:"code"`
	Title           string   `json:"title"`
	Desc            string   `json:"desc,omitempty"`
	Icon            string   `json:"icon,omitempty"`
	Tone            string   `json:"tone,omitempty"`
	Category        string   `json:"category,omitempty"`
	Roles           []string `json:"roles,omitempty"`
	StatusText      string   `json:"statusText,omitempty"`
	ProgressPercent int      `json:"progressPercent,omitempty"`
	Unlocked        bool     `json:"unlocked"`
	Order           int      `json:"order,omitempty"`
	Visible         bool     `json:"visible"`
	AchievedAt      string   `json:"achievedAt,omitempty"`
}

type growthAchievementSeasonDTO struct {
	Title     string `json:"title"`
	Status    string `json:"status"`
	RemainTpl string `json:"remainTpl"`
}

type growthAchievementConfigDTO struct {
	Filters          []growthAchievementFilterDTO `json:"filters"`
	Catalog          []growthAchievementItemDTO   `json:"catalog"`
	Locked           []growthAchievementItemDTO   `json:"locked"`
	Season           growthAchievementSeasonDTO   `json:"season"`
	OnlineSuffix     string                       `json:"onlineSuffix"`
	LevelTitlePrefix string                       `json:"levelTitlePrefix"`
	Version          string                       `json:"version"`
}

func (s *Server) reviewCompleteConfig(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.currentReviewCompleteConfig())
}

func (s *Server) adminGrowthAchievementConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentGrowthAchievementConfig()})
	case http.MethodPut:
		var req growthAchievementConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "成就配置格式错误")
			return
		}
		config := normalizeGrowthAchievementConfig(req)
		if len(config.Catalog) > 200 || len(config.Locked) > 200 {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "成就配置数量超限")
			return
		}
		if s.systemConfig == nil || s.systemConfig.Set(growthAchievementConfigKey, config) != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存成就配置失败")
			return
		}
		s.recordOperation(r, "achievement_config:update", "system_config", growthAchievementConfigKey, map[string]interface{}{"catalogCount": len(config.Catalog), "lockedCount": len(config.Locked)})
		httpx.OK(w, map[string]interface{}{"config": config})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "method not allowed")
	}
}

func (s *Server) adminReviewCompleteConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentReviewCompleteConfig()})
	case http.MethodPut:
		var req reviewCompleteConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid review complete config")
			return
		}
		config, err := normalizeReviewCompleteConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if err := s.systemConfig.Set(reviewCompleteConfigKey, config); err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "save review complete config failed")
			return
		}
		s.recordOperation(r, "review_complete_config:update", "system_config", "review_complete_config", map[string]interface{}{
			"benefitCount":    len(config.Benefits),
			"playOptionCount": len(config.PlayOptions),
			"version":         config.Version,
		})
		httpx.OK(w, map[string]interface{}{"config": s.currentReviewCompleteConfig()})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "method not allowed")
	}
}

func (s *Server) adminCreditDeductionRules(w http.ResponseWriter, r *http.Request) {
	repo, _ := s.reviewRepository().(creditRuleRepository)
	switch r.Method {
	case http.MethodGet:
		items := defaultCreditDeductionRules()
		if repo != nil {
			if stored, err := repo.ListCreditDeductionRules(r.Context()); err == nil && len(stored) > 0 {
				items = stored
			}
		}
		httpx.OK(w, map[string]interface{}{"items": items})
	case http.MethodPut:
		var req struct {
			Items []reviews.CreditDeductionRule `json:"items"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid credit deduction rules")
			return
		}
		items, err := normalizeCreditDeductionRules(req.Items)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if repo == nil {
			httpx.OK(w, map[string]interface{}{"items": items})
			return
		}
		saved := make([]reviews.CreditDeductionRule, 0, len(items))
		for _, item := range items {
			rule, err := repo.UpsertCreditDeductionRule(r.Context(), item)
			if err != nil {
				httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "save credit deduction rule failed")
				return
			}
			saved = append(saved, rule)
		}
		s.recordOperation(r, "credit_deduction_rule:update", "credit_deduction_rule", "batch", map[string]interface{}{"count": len(saved)})
		httpx.OK(w, map[string]interface{}{"items": saved})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "method not allowed")
	}
}

func (s *Server) reviewRepository() reviews.Repository {
	if service, ok := s.reviews.(*reviews.Service); ok {
		return service.Repository()
	}
	return nil
}

func (s *Server) currentReviewCompleteConfig() reviewCompleteConfigDTO {
	var stored reviewCompleteConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(reviewCompleteConfigKey, &stored) {
		config, err := normalizeReviewCompleteConfig(stored)
		if err == nil {
			return config
		}
	}
	return defaultReviewCompleteConfig()
}

func (s *Server) currentReviewPageConfig() reviewPageConfigDTO {
	var stored reviewPageConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(reviewPageConfigKey, &stored) {
		return normalizeReviewPageConfig(stored)
	}
	return defaultReviewPageConfig()
}

func defaultReviewCompleteConfig() reviewCompleteConfigDTO {
	return reviewCompleteConfigDTO{
		Reward: map[string]interface{}{
			"show":  false,
			"title": "评价完成",
			"desc":  "评价会计入信用与成长记录",
		},
		Benefits: []reviewCompleteBenefitDTO{
			{IconText: "✓", Theme: "green", Title: "信用成长", Desc: "完成评价后沉淀履约记录", Order: 10, Visible: true},
			{IconText: "↻", Theme: "blue", Title: "继续组局", Desc: "可继续邀请合作成员再开一局", Order: 20, Visible: true},
		},
		PlayOptions: []reviewCompletePlayOptionDTO{
			{ID: "play_again", Theme: "green", Title: "再玩一局", Desc: "继续发起或加入相似组局", Intent: "yes", Route: "play_again", Order: 10, Visible: true},
			{ID: "game_hall", Theme: "blue", Title: "返回大厅", Desc: "看看附近还有哪些局", Intent: "maybe", Route: "game_hall", Order: 20, Visible: true},
		},
		Version: "2026-07-01",
	}
}

func normalizeReviewCompleteConfig(config reviewCompleteConfigDTO) (reviewCompleteConfigDTO, error) {
	fallback := defaultReviewCompleteConfig()
	if config.Reward == nil {
		config.Reward = fallback.Reward
	}
	config.Benefits = normalizeReviewCompleteBenefits(config.Benefits)
	if len(config.Benefits) == 0 {
		config.Benefits = fallback.Benefits
	}
	config.PlayOptions = normalizeReviewCompletePlayOptions(config.PlayOptions)
	if len(config.PlayOptions) == 0 {
		config.PlayOptions = fallback.PlayOptions
	}
	if strings.TrimSpace(config.Version) == "" {
		config.Version = fallback.Version
	}
	return config, nil
}

func normalizeReviewPageConfig(config reviewPageConfigDTO) reviewPageConfigDTO {
	defaults := defaultReviewPageConfig()
	if strings.TrimSpace(config.NavTitle) == "" {
		config.NavTitle = defaults.NavTitle
	}
	if strings.TrimSpace(config.SkipText) == "" {
		config.SkipText = defaults.SkipText
	}
	if strings.TrimSpace(config.StatusTitle) == "" {
		config.StatusTitle = defaults.StatusTitle
	}
	if strings.TrimSpace(config.StatusDesc) == "" {
		config.StatusDesc = defaults.StatusDesc
	}
	if strings.TrimSpace(config.SatisfactionQuestion) == "" {
		config.SatisfactionQuestion = defaults.SatisfactionQuestion
	}
	if len(config.SatisfactionOptions) == 0 {
		config.SatisfactionOptions = defaults.SatisfactionOptions
	}
	if strings.TrimSpace(config.StoryTitle) == "" {
		config.StoryTitle = defaults.StoryTitle
	}
	if strings.TrimSpace(config.AITip) == "" {
		config.AITip = defaults.AITip
	}
	if strings.TrimSpace(config.StoryPlaceholder) == "" {
		config.StoryPlaceholder = defaults.StoryPlaceholder
	}
	if config.StoryMaxLength <= 0 {
		config.StoryMaxLength = defaults.StoryMaxLength
	}
	if strings.TrimSpace(config.AISummaryText) == "" {
		config.AISummaryText = defaults.AISummaryText
	}
	if strings.TrimSpace(config.RatingHint) == "" {
		config.RatingHint = defaults.RatingHint
	}
	if strings.TrimSpace(config.NPSHeadTitle) == "" {
		config.NPSHeadTitle = defaults.NPSHeadTitle
	}
	if strings.TrimSpace(config.NPSQuestion) == "" {
		config.NPSQuestion = defaults.NPSQuestion
	}
	if strings.TrimSpace(config.NPSLowLabel) == "" {
		config.NPSLowLabel = defaults.NPSLowLabel
	}
	if strings.TrimSpace(config.NPSHighLabel) == "" {
		config.NPSHighLabel = defaults.NPSHighLabel
	}
	if strings.TrimSpace(config.SubmitText) == "" {
		config.SubmitText = defaults.SubmitText
	}
	if strings.TrimSpace(config.SubmitNote) == "" {
		config.SubmitNote = defaults.SubmitNote
	}
	if strings.TrimSpace(config.SkipToast) == "" {
		config.SkipToast = defaults.SkipToast
	}
	if strings.TrimSpace(config.SubmitSuccessText) == "" {
		config.SubmitSuccessText = defaults.SubmitSuccessText
	}
	if strings.TrimSpace(config.NoReviewTargetText) == "" {
		config.NoReviewTargetText = defaults.NoReviewTargetText
	}
	if strings.TrimSpace(config.MissingTargetText) == "" {
		config.MissingTargetText = defaults.MissingTargetText
	}
	if strings.TrimSpace(config.MissingScoreText) == "" {
		config.MissingScoreText = defaults.MissingScoreText
	}
	if strings.TrimSpace(config.SubmitFailedText) == "" {
		config.SubmitFailedText = defaults.SubmitFailedText
	}
	if strings.TrimSpace(config.DefaultSummary) == "" {
		config.DefaultSummary = defaults.DefaultSummary
	}
	if len(config.AgainIntentBySatisfaction) == 0 {
		config.AgainIntentBySatisfaction = defaults.AgainIntentBySatisfaction
	}
	if len(config.RoleConfigs) == 0 {
		config.RoleConfigs = defaults.RoleConfigs
	}
	if strings.TrimSpace(config.Version) == "" {
		config.Version = defaults.Version
	}
	return config
}

func defaultReviewPageConfig() reviewPageConfigDTO {
	return reviewPageConfigDTO{
		NavTitle:             "服务评价",
		SkipText:             "跳过",
		StatusTitle:          "服务已完成！",
		StatusDesc:           "请对本次服务进行评价",
		SatisfactionQuestion: "这一局好玩吗？",
		SatisfactionOptions: []reviewPageSatisfactionOptionDTO{
			{ID: "great", Emoji: "😀", Title: "很好玩", Desc: "五星体验"},
			{ID: "ok", Emoji: "🙂", Title: "还行", Desc: "基本合格"},
			{ID: "bad", Emoji: "😕", Title: "不好玩", Desc: "有待改进"},
		},
		StoryTitle:         "发生了什么有趣的事？",
		AITip:              "AI小助手提示：可以从收获、惊喜、合作感受等方面描述哦",
		StoryPlaceholder:   "我们碰撞出了新的思路，对方的经验帮了大忙！",
		StoryMaxLength:     100,
		AISummaryText:      "AI帮我总结",
		RatingHint:         "点击星星评分",
		NPSHeadTitle:       "发起人专属",
		NPSQuestion:        "你愿意和这些小伙伴再玩一局吗？",
		NPSLowLabel:        "不可能",
		NPSHighLabel:       "极有可能",
		SubmitText:         "提交评价",
		SubmitNote:         "评价内容仅双方可见，请客观公正",
		SkipToast:          "已跳过评价",
		SubmitSuccessText:  "评价已提交",
		NoReviewTargetText: "暂无可评价对象",
		MissingTargetText:  "缺少评价对象，无法提交",
		MissingScoreText:   "请先为每个评价对象打分",
		SubmitFailedText:   "提交评价失败",
		DefaultSummary:     "本次合作沟通顺畅，交付清晰，整体体验不错。",
		AgainIntentBySatisfaction: map[string]string{
			"great": "yes",
			"ok":    "maybe",
			"bad":   "no",
		},
		RoleConfigs: map[string]reviewPageRoleConfigDTO{
			"expert": {
				ID: "expert", AvatarText: "ZH", AvatarTheme: "blue", Title: "评价行家", Desc: "本次服务已完成",
				RatingTitle: "服务质量", TagTitle: "行家标签（多选）",
				Tags:        []string{"专业能力强", "交付及时", "沟通顺畅", "超出预期", "性价比高", "推荐再合作"},
				Placeholder: "分享你对本次服务的评价...",
			},
			"player": {
				ID: "player", AvatarText: "WA", AvatarTheme: "pink", Title: "评价玩家", Desc: "需求已确认，开始反馈",
				RatingTitle: "合作满意度", TagTitle: "玩家标签（多选）",
				Tags:        []string{"需求明确", "配合度高", "付款及时", "沟通友好", "长期合作潜力"},
				Placeholder: "写下你对需求方的评价...",
			},
			"guide": {
				ID: "guide", AvatarText: "WA", AvatarTheme: "orange", Title: "评价领路人", Desc: "撮合已完成，协助交付",
				RatingTitle: "引荐满意度", TagTitle: "邀约标签（多选）",
				Tags:        []string{"匹配精准", "响应及时", "协助积极", "沟通高效", "值得信赖"},
				Placeholder: "写下你对引荐人的服务评价...",
			},
		},
		Version: "2026-07-01",
	}
}

func normalizeReviewCompleteBenefits(items []reviewCompleteBenefitDTO) []reviewCompleteBenefitDTO {
	result := make([]reviewCompleteBenefitDTO, 0, len(items))
	for _, item := range items {
		item.Title = strings.TrimSpace(item.Title)
		item.Desc = strings.TrimSpace(item.Desc)
		item.Theme = strings.TrimSpace(item.Theme)
		if item.Title == "" || (!item.Visible && item.Order != 0) {
			continue
		}
		if item.Theme == "" {
			item.Theme = "blue"
		}
		if item.Order == 0 {
			item.Visible = true
		}
		result = append(result, item)
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Order < result[j].Order })
	return result
}

func normalizeReviewCompletePlayOptions(items []reviewCompletePlayOptionDTO) []reviewCompletePlayOptionDTO {
	allowedIntents := map[string]bool{"yes": true, "maybe": true, "no": true}
	allowedRoutes := map[string]bool{"play_again": true, "game_hall": true}
	result := make([]reviewCompletePlayOptionDTO, 0, len(items))
	for _, item := range items {
		item.ID = strings.TrimSpace(item.ID)
		item.Title = strings.TrimSpace(item.Title)
		item.Desc = strings.TrimSpace(item.Desc)
		item.Theme = strings.TrimSpace(item.Theme)
		item.Intent = strings.TrimSpace(item.Intent)
		item.Route = strings.TrimSpace(item.Route)
		if item.ID == "" || item.Title == "" || !allowedIntents[item.Intent] || !allowedRoutes[item.Route] || (!item.Visible && item.Order != 0) {
			continue
		}
		if item.Theme == "" {
			item.Theme = "green"
		}
		if item.Order == 0 {
			item.Visible = true
		}
		result = append(result, item)
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Order < result[j].Order })
	return result
}

func (s *Server) currentGrowthAchievementConfig() growthAchievementConfigDTO {
	var stored growthAchievementConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(growthAchievementConfigKey, &stored) {
		return normalizeGrowthAchievementConfig(stored)
	}
	return defaultGrowthAchievementConfig()
}

func defaultGrowthAchievementConfig() growthAchievementConfigDTO {
	return growthAchievementConfigDTO{
		Filters: []growthAchievementFilterDTO{
			{Key: "all", Label: "全部", Order: 10},
		},
		Catalog:          []growthAchievementItemDTO{},
		Locked:           []growthAchievementItemDTO{},
		Season:           growthAchievementSeasonDTO{Title: "赛季", Status: "进行中", RemainTpl: "{footprintCount} 条记录"},
		OnlineSuffix:     defaultAchievementOnlineText,
		LevelTitlePrefix: "Lv.",
		Version:          "2026-07-01",
	}
}

func normalizeGrowthAchievementConfig(config growthAchievementConfigDTO) growthAchievementConfigDTO {
	fallback := defaultGrowthAchievementConfig()
	config.Filters = normalizeGrowthAchievementFilters(config.Filters)
	if len(config.Filters) == 0 {
		config.Filters = fallback.Filters
	}
	config.Catalog = normalizeGrowthAchievementItems(config.Catalog, true)
	if len(config.Catalog) == 0 {
		config.Catalog = fallback.Catalog
	}
	config.Locked = normalizeGrowthAchievementItems(config.Locked, false)
	if config.Season.Title == "" {
		config.Season.Title = fallback.Season.Title
	}
	if config.Season.Status == "" {
		config.Season.Status = fallback.Season.Status
	}
	if config.Season.RemainTpl == "" {
		config.Season.RemainTpl = fallback.Season.RemainTpl
	}
	if strings.TrimSpace(config.OnlineSuffix) == "" {
		config.OnlineSuffix = fallback.OnlineSuffix
	}
	if strings.TrimSpace(config.LevelTitlePrefix) == "" {
		config.LevelTitlePrefix = fallback.LevelTitlePrefix
	}
	if strings.TrimSpace(config.Version) == "" {
		config.Version = fallback.Version
	}
	return config
}

func normalizeGrowthAchievementFilters(items []growthAchievementFilterDTO) []growthAchievementFilterDTO {
	result := make([]growthAchievementFilterDTO, 0, len(items))
	seen := make(map[string]bool)
	for _, item := range items {
		item.Key = strings.TrimSpace(item.Key)
		item.Label = strings.TrimSpace(item.Label)
		if item.Key == "" || item.Label == "" || seen[item.Key] {
			continue
		}
		seen[item.Key] = true
		result = append(result, item)
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Order < result[j].Order })
	return result
}

func normalizeGrowthAchievementItems(items []growthAchievementItemDTO, unlocked bool) []growthAchievementItemDTO {
	result := make([]growthAchievementItemDTO, 0, len(items))
	for _, item := range items {
		item.ID = strings.TrimSpace(firstNonEmptyText(item.ID, item.Code))
		item.Code = strings.TrimSpace(firstNonEmptyText(item.Code, item.ID))
		item.Title = strings.TrimSpace(item.Title)
		if item.ID == "" || item.Code == "" || item.Title == "" || (!item.Visible && item.Order != 0) {
			continue
		}
		if item.Icon == "" {
			item.Icon = defaultAchievementIcon
		}
		if item.Tone == "" {
			item.Tone = "gold"
		}
		if item.Category == "" {
			item.Category = "city"
		}
		item.Roles = normalizeAchievementRoles(item.Roles)
		if item.StatusText == "" {
			if unlocked {
				item.StatusText = "已解锁"
			} else {
				item.StatusText = "进行中"
			}
		}
		item.Unlocked = unlocked
		if item.Order == 0 {
			item.Visible = true
		}
		result = append(result, item)
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Order < result[j].Order })
	return result
}

func normalizeAchievementRoles(items []string) []string {
	result := make([]string, 0, len(items))
	seen := make(map[string]bool, len(items))
	for _, value := range items {
		role := strings.ToLower(strings.TrimSpace(value))
		if role != "player" && role != "expert" && role != "guide" || seen[role] {
			continue
		}
		seen[role] = true
		result = append(result, role)
	}
	return result
}

func achievementDTOs(trace reviews.Trace, config growthAchievementConfigDTO) []growthAchievementItemDTO {
	return achievementDTOsForRole(trace, config, "")
}

func achievementDTOsForRole(trace reviews.Trace, config growthAchievementConfigDTO, role string) []growthAchievementItemDTO {
	catalog := make(map[string]growthAchievementItemDTO)
	for _, item := range config.Catalog {
		if !achievementVisibleForRole(item, role) {
			continue
		}
		catalog[item.Code] = item
	}
	result := make([]growthAchievementItemDTO, 0, len(trace.Achievements))
	for index, achievement := range trace.Achievements {
		item, ok := catalog[achievement.Code]
		if !ok {
			item = growthAchievementItemDTO{
				ID:         achievement.Code,
				Code:       achievement.Code,
				Title:      firstNonEmptyText(achievement.Title, achievement.Code),
				Desc:       "来自后端成长记录",
				Icon:       defaultAchievementIcon,
				Tone:       "gold",
				Category:   "city",
				StatusText: "已解锁",
				Order:      index + 1000,
				Visible:    true,
			}
		}
		item.ID = firstNonEmptyText(item.ID, achievement.Code)
		item.Code = achievement.Code
		item.Title = firstNonEmptyText(item.Title, achievement.Title, achievement.Code)
		item.Unlocked = true
		item.Visible = true
		item.AchievedAt = achievement.AchievedAt.Format(time.RFC3339)
		result = append(result, item)
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Order < result[j].Order })
	return result
}

func achievementVisibleForRole(item growthAchievementItemDTO, role string) bool {
	if len(item.Roles) == 0 || strings.TrimSpace(role) == "" {
		return true
	}
	role = strings.ToLower(strings.TrimSpace(role))
	for _, allowed := range item.Roles {
		if strings.ToLower(strings.TrimSpace(allowed)) == role {
			return true
		}
	}
	return false
}

func defaultCreditDeductionRules() []reviews.CreditDeductionRule {
	return []reviews.CreditDeductionRule{
		{RuleCode: "quit_after_confirm", ChangeValue: -10, Enabled: true, Description: "Deduct credit when a member exits after service confirmation starts."},
		{RuleCode: "quit_after_started", ChangeValue: -10, Enabled: true, Description: "Deduct credit when a member exits after the game has started."},
		{RuleCode: "player_cancel_service", ChangeValue: -3, Enabled: true, Description: "Deduct credit when a player cancels an active service."},
		{RuleCode: "expert_cancel_service", ChangeValue: -5, Enabled: true, Description: "Deduct credit when an expert cancels an active service."},
		{RuleCode: "low_review", ChangeValue: -5, Enabled: true, Description: "Deduct credit when a participant receives a low score review."},
		{RuleCode: "report_confirmed", ChangeValue: -10, Enabled: true, Description: "Deduct credit when a report is confirmed."},
		{RuleCode: "malicious_report", ChangeValue: -10, Enabled: true, Description: "Deduct credit when a report is judged malicious."},
	}
}

func normalizeCreditDeductionRules(items []reviews.CreditDeductionRule) ([]reviews.CreditDeductionRule, error) {
	if len(items) == 0 {
		return nil, errors.New("rules required")
	}
	result := make([]reviews.CreditDeductionRule, 0, len(items))
	seen := make(map[string]bool)
	for _, item := range items {
		item.RuleCode = strings.TrimSpace(item.RuleCode)
		item.Description = strings.TrimSpace(item.Description)
		if item.RuleCode == "" {
			return nil, errors.New("ruleCode required")
		}
		if seen[item.RuleCode] {
			return nil, errors.New("duplicate ruleCode")
		}
		if item.ChangeValue >= 0 || item.ChangeValue < -100 {
			return nil, errors.New("changeValue must be between -100 and -1")
		}
		seen[item.RuleCode] = true
		result = append(result, item)
	}
	return result, nil
}

func (s *Server) serviceConfirm(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/service-confirm")
	if !ok {
		return
	}
	var req struct {
		Note            string   `json:"note"`
		FileIDs         []int64  `json:"fileIds"`
		ConfirmItemKeys []string `json:"confirmItemKeys"`
		ConfirmSource   string   `json:"confirmSource"`
		PageGameID      int64    `json:"pageGameId"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	gameBefore, err := s.games.Get(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	submittedKeys := append([]string(nil), req.ConfirmItemKeys...)
	req.ConfirmItemKeys = s.normalizeDeliveryConfirmItems(gameBefore, req.ConfirmItemKeys)
	attemptExtra := map[string]interface{}{
		"gameStatusBefore": gameBefore.Status,
		"role":             s.userRoleForGame(gameBefore, userID),
		"submittedKeys":    submittedKeys,
		"normalizedKeys":   req.ConfirmItemKeys,
		"confirmSource":    strings.TrimSpace(req.ConfirmSource),
		"pageGameId":       req.PageGameID,
		"clientType":       strings.TrimSpace(r.Header.Get("X-Client-Type")),
		"appVersion":       strings.TrimSpace(r.Header.Get("X-App-Version")),
		"userAgent":        strings.TrimSpace(r.UserAgent()),
	}
	if len(req.ConfirmItemKeys) == 0 {
		attemptExtra["result"] = "rejected_incomplete_items"
		s.recordBehavior(userID, "service_confirm_attempt", "game", gameID, attemptExtra)
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "服务确认项不完整")
		return
	}
	encodedItems, _ := json.Marshal(req.ConfirmItemKeys)
	req.Note = strings.TrimSpace(req.Note + "\n确认项:" + string(encodedItems))
	wasReviewable := currentGameReviewable(s.games, gameID)
	confirm, items, game, err := s.games.ConfirmService(userID, gameID, req.Note, req.FileIDs...)
	if err != nil {
		attemptExtra["result"] = "rejected"
		attemptExtra["error"] = err.Error()
		s.recordBehavior(userID, "service_confirm_attempt", "game", gameID, attemptExtra)
		writeGameError(w, err)
		return
	}
	attemptExtra["result"] = "confirmed"
	attemptExtra["gameStatusAfter"] = game.Status
	s.recordBehavior(userID, "service_confirm_attempt", "game", gameID, attemptExtra)
	if game.Status == "pending_review" && !wasReviewable {
		s.reviews.MarkGameReviewable(gameID)
		s.reviews.AwardCompletedGame(gameID)
		s.createCoGameConnections(gameID)
		s.createReviewRemindNotifications(gameID, userID)
	} else if game.Status == "pending_confirm" {
		s.notifyPlayersAfterExpertConfirmation(game, userID)
	}
	s.recordBehavior(userID, "service_confirm", "game", gameID, map[string]interface{}{
		"gameStatus":       game.Status,
		"gameStatusBefore": gameBefore.Status,
		"fileIds":          req.FileIDs,
		"confirmItemKeys":  req.ConfirmItemKeys,
		"confirmSource":    strings.TrimSpace(req.ConfirmSource),
		"pageGameId":       req.PageGameID,
		"role":             s.userRoleForGame(gameBefore, userID),
	})
	httpx.OK(w, map[string]interface{}{"confirm": confirm, "items": items, "game": game})
}

func (s *Server) normalizeDeliveryConfirmItems(game games.Game, keys []string) []string {
	config := s.currentGameDeliveryPageConfig().Paid
	if successFundAmount(game) <= 0 {
		config = s.currentGameDeliveryPageConfig().Free
	}
	selected := make(map[string]bool, len(keys))
	for _, key := range keys {
		selected[strings.TrimSpace(key)] = true
	}
	if len(config.ConfirmItems) == 0 {
		return nil
	}
	if len(keys) == 0 {
		return nil
	}
	for _, item := range config.ConfirmItems {
		if !selected[item.ID] {
			return nil
		}
	}
	result := make([]string, 0, len(config.ConfirmItems))
	for _, item := range config.ConfirmItems {
		result = append(result, item.ID)
	}
	return result
}

func (s *Server) notifyPlayersAfterExpertConfirmation(game games.Game, confirmerUserID int64) {
	if s.userRoleForGame(game, confirmerUserID) != "expert" {
		return
	}
	confirmed := make(map[int64]bool)
	if _, items, ok := s.games.ServiceConfirmForGame(game.ID); ok {
		for _, item := range items {
			confirmed[item.UserID] = true
		}
	}
	for _, memberID := range s.boundPlayersForExpert(game.ID, confirmerUserID) {
		if confirmed[memberID] {
			continue
		}
		s.notices.Create(notifications.CreateRequest{
			UserID:     memberID,
			NotifyType: "expert_completion_confirmed",
			Title:      "行家已确认服务完成",
			Content:    "《" + game.Title + "》的行家已确认完成，请你进行最终确认。",
			BizType:    "game",
			BizID:      game.ID,
		})
	}
}

func (s *Server) boundPlayersForExpert(gameID int64, expertUserID int64) []int64 {
	result := make([]int64, 0, 1)
	seen := make(map[int64]bool)
	for _, invitation := range s.games.InvitationsForUser(expertUserID) {
		playerID := invitation.PlayerUserID
		if invitation.GameID != gameID || invitation.TargetUserID != expertUserID || invitation.Status != "accepted" || guideProgressInvitationRole(invitation.Role) != "expert" || playerID <= 0 || seen[playerID] || !s.games.IsMember(gameID, playerID) {
			continue
		}
		seen[playerID] = true
		result = append(result, playerID)
	}
	return result
}

func (s *Server) serviceConfirmItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/service-confirm-items")
	if !ok {
		return
	}
	var req struct {
		Note            string   `json:"note"`
		FileIDs         []int64  `json:"fileIds"`
		ConfirmItemKeys []string `json:"confirmItemKeys"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	gameBefore, err := s.games.Get(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	req.ConfirmItemKeys = s.normalizeDeliveryConfirmItems(gameBefore, req.ConfirmItemKeys)
	if len(req.ConfirmItemKeys) == 0 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "服务确认项不完整")
		return
	}
	encodedItems, _ := json.Marshal(req.ConfirmItemKeys)
	req.Note = strings.TrimSpace(req.Note + "\n确认项:" + string(encodedItems))
	wasReviewable := currentGameReviewable(s.games, gameID)
	confirm, items, game, err := s.games.ConfirmService(userID, gameID, req.Note, req.FileIDs...)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if game.Status == "pending_review" && !wasReviewable {
		s.reviews.MarkGameReviewable(gameID)
		s.reviews.AwardCompletedGame(gameID)
		s.createCoGameConnections(gameID)
		s.createReviewRemindNotifications(gameID, userID)
	}
	s.recordBehavior(userID, "service_confirm", "game", gameID, map[string]interface{}{"gameStatus": game.Status, "fileIds": req.FileIDs})
	httpx.OK(w, map[string]interface{}{"confirm": confirm, "items": items, "game": game})
}

func (s *Server) createCoGameConnections(gameID int64) {
	members := s.games.Members(gameID)
	for _, userID := range members {
		for _, connectedUserID := range members {
			if userID == connectedUserID {
				continue
			}
			s.connections.UpsertPair(userID, connectedUserID, "co_game", "co_game", gameID, 2)
		}
	}
}

func (s *Server) createReviewRemindNotifications(gameID int64, completionConfirmerIDs ...int64) {
	game, err := s.games.Get(gameID)
	if err != nil {
		return
	}
	excluded := make(map[int64]bool, len(completionConfirmerIDs))
	for _, userID := range completionConfirmerIDs {
		excluded[userID] = true
	}
	completionTriggered := len(completionConfirmerIDs) > 0
	hasReviewRoleRecipient := false
	if completionTriggered {
		for _, memberID := range s.games.Members(gameID) {
			role := s.userRoleForGame(game, memberID)
			if !excluded[memberID] && (role == "expert" || role == "guide" || role == "main_guide") {
				hasReviewRoleRecipient = true
				break
			}
		}
	}
	for _, userID := range s.games.Members(gameID) {
		if excluded[userID] {
			continue
		}
		if completionTriggered && hasReviewRoleRecipient {
			role := s.userRoleForGame(game, userID)
			if role != "expert" && role != "guide" && role != "main_guide" {
				continue
			}
		}
		todos, err := s.reviews.Todos(userID)
		if err != nil {
			continue
		}
		for _, todo := range todos {
			if todo.GameID != gameID {
				continue
			}
			s.notices.Create(notifications.CreateRequest{
				UserID:      userID,
				NotifyType:  "review_remind",
				Title:       "服务已确认完成，请评价",
				Content:     "《" + game.Title + "》的玩家已确认服务完成，请进入评价。",
				BizType:     "game",
				BizID:       gameID,
				NeedWechat:  true,
				WechatState: "pending",
				WechatData: map[string]string{
					"thing1": game.Title,
					"time2":  todo.DeadlineAt,
				},
			})
			break
		}
	}
}

func currentGameReviewable(games gameService, gameID int64) bool {
	game, err := games.Get(gameID)
	return err == nil && (game.Status == "pending_review" || game.Status == "completed")
}

func (s *Server) reviewTodos(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	items, err := s.reviews.Todos(userID)
	if err != nil {
		writeReviewError(w, err)
		return
	}
	requestedGameID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("gameId")), 10, 64)
	viewerIsOrganizer := false
	if requestedGameID > 0 {
		if game, err := s.games.Get(requestedGameID); err == nil {
			viewerIsOrganizer = game.CreatorUserID == userID
		}
	}
	todoItems := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		profile := s.inGameIdentity(item.TargetUserID, "待评价成员")
		dto := map[string]interface{}{
			"gameId": item.GameID, "targetUserId": item.TargetUserID, "targetRole": item.TargetRole,
			"deadlineAt": item.DeadlineAt, "targetName": profile.DisplayName,
			"realName": profile.RealName, "displayName": profile.DisplayName, "avatarText": profile.AvatarText,
		}
		if avatarType := realnameAvatarType(strings.TrimSpace(profile.RealName) != ""); avatarType != "" {
			dto["avatarType"] = avatarType
		}
		todoItems = append(todoItems, dto)
	}
	httpx.OK(w, map[string]interface{}{
		"items":      todoItems,
		"reward":     s.currentReviewCompleteConfig().Reward,
		"reviewPage": s.currentReviewPageConfig(),
		"viewer": map[string]interface{}{
			"userId":           userID,
			"gameId":           requestedGameID,
			"isOrganizer":      viewerIsOrganizer,
			"showOrganizerNps": viewerIsOrganizer,
		},
	})
}
func (s *Server) submitReview(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req reviews.SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	// The central points service below is the single reward writer. Passing zero
	// here prevents the review growth service from crediting the same reward twice.
	review, profile, err := s.reviews.SubmitWithPoints(userID, req, 0)
	if err != nil {
		writeReviewError(w, err)
		return
	}
	rewardPoints := s.reviews.GrowthRules().SubmittedReviewPoints
	pointsAccount := s.points.Summary(userID)
	var pointsLog points.Log
	if rewardPoints > 0 {
		var err error
		pointsAccount, pointsLog, err = s.points.Grant(userID, rewardPoints, "review_reward", review.ID, "提交评价奖励")
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "grant review points failed")
			return
		}
	}
	profile.AvailablePoints = pointsAccount.AvailablePoints
	s.connections.UpsertPair(userID, review.TargetUserID, "co_game", "review", review.GameID, 1)
	s.recordBehavior(userID, "submit_review", "game", req.GameID, map[string]interface{}{"reviewId": review.ID, "targetUserId": review.TargetUserID})
	var credit interface{}
	if review.Score <= 2 {
		log := s.reviews.DeductCredit(review.TargetUserID, review.GameID, "low_review")
		credit = log
		s.notices.Create(notifications.CreateRequest{
			UserID:     review.TargetUserID,
			NotifyType: "low_review_credit_deducted",
			Title:      "收到低分评价",
			Content:    "你收到一条低分评价，系统已记录信用变化，可在信用中心查看并申诉。",
			BizType:    "review",
			BizID:      review.ID,
		})
	}
	httpx.OK(w, map[string]interface{}{
		"review":        review,
		"profile":       profile,
		"credit":        credit,
		"pointsSummary": pointsAccount,
		"pointsLog":     pointsLog,
		"reward": map[string]interface{}{
			"points":       pointsLog.ChangeValue,
			"beforePoints": pointsLog.BeforePoints,
			"afterPoints":  pointsLog.AfterPoints,
			"logId":        pointsLog.ID,
		},
	})
}

func (s *Server) myReviewIntents(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	httpx.OK(w, map[string]interface{}{"items": s.reviews.MyIntents(userID)})
}

func (s *Server) reviewProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	trace := s.reviews.TraceByUser(userID)
	achievementConfig := s.currentGrowthAchievementConfig()
	role := "player"
	roles := s.profiles.RoleSnapshot(userID).Roles
	for _, candidate := range []string{"expert", "guide"} {
		for _, current := range roles {
			if current == candidate {
				role = candidate
				break
			}
		}
		if role == candidate {
			break
		}
	}
	achievements := achievementDTOsForRole(trace, achievementConfig, role)
	httpx.OK(w, map[string]interface{}{
		"profile":           trace.Profile,
		"footprints":        trace.Footprints,
		"achievements":      achievements,
		"achievementConfig": achievementConfig,
	})
}

func (s *Server) profileCreditCenter(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}

	trace := s.reviews.TraceByUser(userID)
	profile := trace.Profile
	creditRules := s.currentOperationRules().Credit
	levelText := creditLevelText(profile.CreditScore, creditRules.ExcellentThreshold, creditRules.RestrictedThreshold, creditRules.SuspendedThreshold)
	operationRules := s.currentOperationRules()
	creditNote := operationRules.Messages["creditRestricted"]
	if creditNote == "" {
		creditNote = "信用分低于{score}分将限制部分功能"
	}
	creditNote = strings.ReplaceAll(creditNote, "{score}", strconv.Itoa(operationRules.Credit.RestrictedThreshold))
	suspendedNote := operationRules.Messages["creditSuspended"]
	if suspendedNote == "" {
		suspendedNote = "信用分低于{score}分将暂停服务资格"
	}
	suspendedNote = strings.ReplaceAll(suspendedNote, "{score}", strconv.Itoa(operationRules.Credit.SuspendedThreshold))
	monthlyDelta := 0
	positiveCount := 0
	negativeCount := 0
	records := make([]map[string]interface{}, 0, len(trace.CreditLogs))
	for _, log := range trace.CreditLogs {
		monthlyDelta += log.ChangeValue
		if log.ChangeValue >= 0 {
			positiveCount++
		} else {
			negativeCount++
		}
		records = append(records, map[string]interface{}{
			"id":          log.ID,
			"creditLogId": log.ID,
			"title":       creditReasonTitle(log.Reason),
			"desc":        creditReasonDesc(log),
			"score":       signedIntText(log.ChangeValue),
			"tone":        creditTone(log.ChangeValue),
			"gameId":      log.GameID,
			"appealRoute": "/pages/profile/system-management/credit-appeal/index?creditLogId=" + strconv.FormatInt(log.ID, 10),
			"reason":      log.Reason,
			"beforeScore": log.BeforeScore,
			"afterScore":  log.AfterScore,
			"createdAt":   log.CreatedAt,
		})
	}

	httpx.OK(w, map[string]interface{}{
		"score":        profile.CreditScore,
		"scoreLabel":   "信用分",
		"todayScore":   profile.TodayCreditScore,
		"level":        levelText,
		"monthlyDelta": signedIntText(monthlyDelta),
		"summary": []map[string]string{
			{"label": "信用等级", "value": levelText},
			{"label": "奖励中心", "value": strconv.Itoa(positiveCount) + "条待查看"},
			{"label": "惩罚中心", "value": strconv.Itoa(negativeCount) + "条记录"},
		},
		"records": records,
		"appealEntry": map[string]interface{}{
			"enabled": len(records) > 0,
			"route":   "/pages/profile/system-management/credit-appeal/index",
			"text":    "信用申诉",
		},
		"bottomNote": creditNote + "，" + suspendedNote + "。",
	})
}

func (s *Server) myFootprints(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	httpx.OK(w, map[string]interface{}{"items": s.reviews.Footprints(userID)})
}

func creditLevelText(score int, excellent, restricted, suspended int) string {
	switch {
	case score >= excellent:
		return "优秀"
	case score >= restricted:
		return "良好"
	case score >= suspended:
		return "受限"
	default:
		return "暂停服务"
	}
}

func creditReasonTitle(reason string) string {
	switch reason {
	case "quit_after_started":
		return "开局后退出"
	case "quit_after_confirm":
		return "确认后退出"
	case "report_confirmed":
		return "举报核实"
	case "appeal_passed":
		return "申诉通过"
	case "low_review":
		return "低分评价"
	default:
		return "信用变更"
	}
}

func creditReasonDesc(log reviews.CreditLog) string {
	if log.GameID > 0 {
		return log.Reason + " · 局ID " + strconv.FormatInt(log.GameID, 10)
	}
	return log.Reason
}

func signedIntText(value int) string {
	if value > 0 {
		return "+" + strconv.Itoa(value)
	}
	return strconv.Itoa(value)
}

func creditTone(value int) string {
	if value >= 0 {
		return "plus"
	}
	return "minus"
}

func (s *Server) adminUserGrowth(w http.ResponseWriter, r *http.Request) {
	userID, ok := idFromAdminPath(w, r.URL.Path, "/api/admin/users/", "/growth")
	if !ok {
		return
	}
	httpx.OK(w, s.reviews.TraceByUser(userID))
}

func (s *Server) adminGameReviewTrace(w http.ResponseWriter, r *http.Request) {
	gameID, ok := idFromAdminPath(w, r.URL.Path, "/api/admin/games/", "/review-trace")
	if !ok {
		return
	}
	httpx.OK(w, s.reviews.TraceByGame(gameID))
}

func idFromAdminPath(w http.ResponseWriter, path string, prefix string, suffix string) (int64, bool) {
	idText := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	id, err := strconv.ParseInt(strings.Trim(idText, "/"), 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid id")
		return 0, false
	}
	return id, true
}

func writeReviewError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, games.ErrGameNotFound):
		httpx.Error(w, http.StatusNotFound, 40421, "game not found")
	case errors.Is(err, reviews.ErrForbidden):
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "review forbidden")
	case errors.Is(err, reviews.ErrGameNotReviewable):
		httpx.Error(w, http.StatusConflict, 40931, "game not reviewable")
	case errors.Is(err, reviews.ErrDuplicateReview):
		httpx.Error(w, http.StatusConflict, 40932, "duplicate review")
	case errors.Is(err, reviews.ErrInvalidReview):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid review")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "review operation failed")
	}
}
