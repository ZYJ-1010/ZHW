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
	"zhw-mini/services/go-api/internal/reviews"
)

type reviewService interface {
	MarkGameReviewable(gameID int64)
	MarkGameReviewableStrict(gameID int64) error
	AwardCompletedGame(gameID int64) []reviews.GrowthProfile
	AwardCompletedGameStrict(gameID int64) ([]reviews.GrowthProfile, error)
	RecordGrowthEvent(event reviews.GrowthEvent)
	AwardTaskReward(userID int64, taskCode string, points int, experience int) reviews.GrowthProfile
	AwardTaskRewardStrict(userID int64, taskCode string, points int, experience int) (reviews.GrowthProfile, error)
	Todos(userID int64) ([]reviews.Todo, error)
	Submit(userID int64, req reviews.SubmitRequest) (reviews.Review, reviews.GrowthProfile, error)
	SubmitWithPoints(userID int64, req reviews.SubmitRequest, rewardPoints int) (reviews.Review, reviews.GrowthProfile, error)
	GrowthRules() reviews.GrowthRules
	MyIntents(userID int64) []reviews.Review
	MyIntentsStrict(userID int64) ([]reviews.Review, error)
	AllReviews() []reviews.Review
	AllReviewsStrict() ([]reviews.Review, error)
	Profile(userID int64) reviews.GrowthProfile
	ProfileStrict(userID int64) (reviews.GrowthProfile, error)
	Footprints(userID int64) []reviews.Footprint
	FootprintsStrict(userID int64) ([]reviews.Footprint, error)
	AllFootprints() []reviews.Footprint
	AllFootprintsStrict() ([]reviews.Footprint, error)
	DeductCredit(userID int64, gameID int64, reason string) reviews.CreditLog
	DeductCreditStrict(userID int64, gameID int64, reason string) (reviews.CreditLog, error)
	DeductCreditOnceStrict(userID int64, gameID int64, reason string, idempotencyKey string) (reviews.CreditLog, bool, error)
	CreditDeductionValue(reason string) int
	CreditDeductionValueStrict(reason string) (int, error)
	RestoreCredit(userID int64, gameID int64, reason string, amount int) reviews.CreditLog
	RestoreCreditStrict(userID int64, gameID int64, reason string, amount int) (reviews.CreditLog, error)
	CreditScoreStrict(userID int64) (int, error)
	RestoreCreditForAppeal(userID int64, gameID int64, sourceCreditLogID int64, appealID int64, amount int) (reviews.CreditLog, bool, error)
	TraceByUser(userID int64) reviews.Trace
	TraceByUserStrict(userID int64) (reviews.Trace, error)
	TraceByGame(gameID int64) reviews.Trace
	TraceByGameStrict(gameID int64) (reviews.Trace, error)
	GameReviewComplete(gameID int64) bool
}

type gameReviewCompletionFinalizer interface {
	CompleteAfterReviews(gameID int64) (games.Game, bool, error)
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
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "不支持当前请求方式")
	}
}

func (s *Server) adminReviewCompleteConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentReviewCompleteConfig()})
	case http.MethodPut:
		var req reviewCompleteConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "评价完成页配置格式错误")
			return
		}
		config, err := normalizeReviewCompleteConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if err := s.systemConfig.Set(reviewCompleteConfigKey, config); err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存评价完成页配置失败")
			return
		}
		s.recordOperation(r, "review_complete_config:update", "system_config", "review_complete_config", map[string]interface{}{
			"benefitCount":    len(config.Benefits),
			"playOptionCount": len(config.PlayOptions),
			"version":         config.Version,
		})
		httpx.OK(w, map[string]interface{}{"config": s.currentReviewCompleteConfig()})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "不支持当前请求方式")
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
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "信用扣分规则格式错误")
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
				httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存信用扣分规则失败")
				return
			}
			saved = append(saved, rule)
		}
		s.recordOperation(r, "credit_deduction_rule:update", "credit_deduction_rule", "batch", map[string]interface{}{"count": len(saved)})
		httpx.OK(w, map[string]interface{}{"items": saved})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "不支持当前请求方式")
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
		Catalog: []growthAchievementItemDTO{
			{ID: "first_game", Code: "first_game", Title: "首局达成", Desc: "参与第一局", Category: "city", StatusText: "进行中", Order: 10, Visible: true},
			{ID: "credit_keeper", Code: "credit_keeper", Title: "信用守护", Desc: "保持良好信用", Category: "city", StatusText: "进行中", Order: 20, Visible: true},
			{ID: "earth", Code: "earth", Title: "地球漫游者", Desc: "探索附近组局", Category: "city", StatusText: "进行中", Order: 30, Visible: true},
		},
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
	visibleCatalog := make(map[string]growthAchievementItemDTO)
	for _, item := range config.Catalog {
		if achievementVisibleForRole(item, role) {
			visibleCatalog[item.Code] = item
		}
	}
	result := make([]growthAchievementItemDTO, 0, len(trace.Achievements))
	for index, achievement := range trace.Achievements {
		item, ok := visibleCatalog[achievement.Code]
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

func growthAchievementConfigForRole(config growthAchievementConfigDTO, role string) growthAchievementConfigDTO {
	filtered := config
	filtered.Catalog = make([]growthAchievementItemDTO, 0, len(config.Catalog))
	for _, item := range config.Catalog {
		if achievementVisibleForRole(item, role) {
			filtered.Catalog = append(filtered.Catalog, item)
		}
	}
	filtered.Locked = make([]growthAchievementItemDTO, 0, len(config.Locked))
	for _, item := range config.Locked {
		if achievementVisibleForRole(item, role) {
			filtered.Locked = append(filtered.Locked, item)
		}
	}
	return filtered
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
		{RuleCode: "quit_after_admitted", ChangeValue: -10, Enabled: true, Description: "成员入局后、开局前主动退出时扣除信用分。"},
		{RuleCode: "quit_after_confirm", ChangeValue: -10, Enabled: true, Description: "成员确认服务后主动退出时扣除信用分。"},
		{RuleCode: "quit_after_started", ChangeValue: -10, Enabled: true, Description: "局已开局后主动退出时扣除信用分。"},
		{RuleCode: "player_cancel_service", ChangeValue: -3, Enabled: true, Description: "玩家取消已确认服务时扣除信用分。"},
		{RuleCode: "expert_cancel_service", ChangeValue: -5, Enabled: true, Description: "行家取消已确认服务时扣除信用分。"},
		{RuleCode: "low_review", ChangeValue: -5, Enabled: true, Description: "收到低分评价时扣除信用分。"},
		{RuleCode: "report_confirmed", ChangeValue: -10, Enabled: true, Description: "举报核实成立时扣除信用分。"},
		{RuleCode: "malicious_report", ChangeValue: -10, Enabled: true, Description: "恶意举报核实成立时扣除信用分。"},
	}
}

func normalizeCreditDeductionRules(items []reviews.CreditDeductionRule) ([]reviews.CreditDeductionRule, error) {
	if len(items) == 0 {
		return nil, errors.New("请至少保留一条信用扣分规则")
	}
	result := make([]reviews.CreditDeductionRule, 0, len(items))
	seen := make(map[string]bool)
	for _, item := range items {
		item.RuleCode = strings.TrimSpace(item.RuleCode)
		item.Description = strings.TrimSpace(item.Description)
		if item.RuleCode == "" {
			return nil, errors.New("信用扣分规则编码不能为空")
		}
		if seen[item.RuleCode] {
			return nil, errors.New("信用扣分规则编码不能重复")
		}
		if item.ChangeValue >= 0 || item.ChangeValue < -100 {
			return nil, errors.New("信用扣分值必须在 -100 至 -1 之间")
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
		if reviewErr := s.reviews.MarkGameReviewableStrict(gameID); reviewErr != nil {
			markReviewPersistenceDegraded(w, "service_confirm", gameID, reviewErr)
		}
		if growthErr := s.awardCompletedGameRewards(gameID); growthErr != nil {
			markGrowthPersistenceDegraded(w, "service_confirm", userID, gameID, growthErr)
		}
		if connectionErr := s.createCoGameConnections(gameID); connectionErr != nil {
			markConnectionPersistenceDegraded(w, "service_confirm", userID, 0, connectionErr)
		}
		s.createReviewRemindNotifications(w, gameID, userID)
	} else if game.Status == "pending_confirm" {
		s.notifyPlayersAfterExpertConfirmation(w, game, userID)
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

func (s *Server) notifyPlayersAfterExpertConfirmation(w http.ResponseWriter, game games.Game, confirmerUserID int64) {
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
		_, _ = s.createCriticalNotification(w, "expert_completion_confirmed", notifications.CreateRequest{
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
		if reviewErr := s.reviews.MarkGameReviewableStrict(gameID); reviewErr != nil {
			markReviewPersistenceDegraded(w, "service_confirm_detail", gameID, reviewErr)
		}
		if growthErr := s.awardCompletedGameRewards(gameID); growthErr != nil {
			markGrowthPersistenceDegraded(w, "service_confirm_detail", userID, gameID, growthErr)
		}
		if connectionErr := s.createCoGameConnections(gameID); connectionErr != nil {
			markConnectionPersistenceDegraded(w, "service_confirm_detail", userID, 0, connectionErr)
		}
		s.createReviewRemindNotifications(w, gameID, userID)
	}
	s.recordBehavior(userID, "service_confirm", "game", gameID, map[string]interface{}{"gameStatus": game.Status, "fileIds": req.FileIDs})
	httpx.OK(w, map[string]interface{}{"confirm": confirm, "items": items, "game": game})
}

func (s *Server) createCoGameConnections(gameID int64) error {
	members := s.games.Members(gameID)
	for _, userID := range members {
		for _, connectedUserID := range members {
			if userID == connectedUserID {
				continue
			}
			if err := s.connections.UpsertPairStrict(userID, connectedUserID, "co_game", "co_game", gameID, 2); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Server) createReviewRemindNotifications(w http.ResponseWriter, gameID int64, completionConfirmerIDs ...int64) {
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
			_, _ = s.createCriticalNotification(w, "review_remind", notifications.CreateRequest{
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
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	// A review produces growth experience only. Redeemable points are granted
	// exclusively after a paid-game revenue settlement.
	review, profile, err := s.reviews.SubmitWithPoints(userID, req, 0)
	if err != nil {
		writeReviewError(w, err)
		return
	}
	gameCompleted := false
	gameCompletionPending := false
	if s.reviews.GameReviewComplete(req.GameID) {
		finalizer, ok := s.games.(gameReviewCompletionFinalizer)
		if !ok {
			gameCompletionPending = true
		} else if _, completed, completeErr := finalizer.CompleteAfterReviews(req.GameID); completeErr != nil {
			// 评价已经原子落库，不能因后续状态写入失败向客户端伪报“评价失败”，
			// 否则用户重试会触发重复评价。响应明确告知状态仍待刷新即可。
			gameCompletionPending = true
		} else {
			gameCompleted = completed
		}
	}
	pointsAccount := s.points.Summary(userID)
	profile.AvailablePoints = pointsAccount.AvailablePoints
	if connectionErr := s.connections.UpsertPairStrict(userID, review.TargetUserID, "co_game", "review", review.GameID, 1); connectionErr != nil {
		markConnectionPersistenceDegraded(w, "submit_review", userID, review.TargetUserID, connectionErr)
	}
	s.recordBehavior(userID, "submit_review", "game", req.GameID, map[string]interface{}{"reviewId": review.ID, "targetUserId": review.TargetUserID})
	var credit interface{}
	if review.Score <= 2 {
		creditLog, creditErr := s.reviews.DeductCreditStrict(review.TargetUserID, review.GameID, "low_review")
		if creditErr != nil {
			markCreditPersistenceDegraded(w, "review_low_score", review.TargetUserID, review.GameID, creditErr)
		} else {
			credit = creditLog
		}
		_, _ = s.createCriticalNotification(w, "review_low_score", notifications.CreateRequest{
			UserID:     review.TargetUserID,
			NotifyType: "low_review_credit_deducted",
			Title:      "收到低分评价",
			Content:    "你收到一条低分评价，系统已记录信用变化，可在信用中心查看并申诉。",
			BizType:    "review",
			BizID:      review.ID,
		})
	}
	httpx.OK(w, map[string]interface{}{
		"review":                review,
		"profile":               profile,
		"credit":                credit,
		"gameCompleted":         gameCompleted,
		"gameCompletionPending": gameCompletionPending,
		"pointsSummary":         pointsAccount,
		"reward": map[string]interface{}{
			"experience": s.reviews.GrowthRules().SubmittedReviewExperience,
		},
	})
}

func (s *Server) myReviewIntents(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	items, err := s.reviews.MyIntentsStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取再来一局意向失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) reviewProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	trace, err := s.reviews.TraceByUserStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取成长与信用数据失败，请稍后重试")
		return
	}
	achievementConfig := s.currentGrowthAchievementConfig()
	role := normalizeHomeRoleType(r.URL.Query().Get("roleType"))
	if role == "" || (role != "player" && !s.userHasActiveRole(userID, role)) {
		role = s.homeRoleType(userID)
	}
	roleGrowth := s.roleGrowthProfilesForUser(userID, trace.Profile)
	activeRoleGrowth := homeRoleGrowthProfile(role, roleGrowth)
	achievements := achievementDTOsForRole(trace, achievementConfig, role)
	achievementConfig = growthAchievementConfigForRole(achievementConfig, role)
	httpx.OK(w, map[string]interface{}{
		"profile":           trace.Profile,
		"activeRoleCode":    role,
		"activeRoleName":    homeRoleName(role),
		"activeRoleGrowth":  activeRoleGrowth,
		"roleGrowth":        roleGrowth,
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

	trace, err := s.reviews.TraceByUserStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取成长与信用数据失败，请稍后重试")
		return
	}
	profile := trace.Profile
	creditConfig := s.currentCreditRestrictionConfig()
	creditState, creditStateText := creditRestrictionState(profile.CreditScore, creditConfig)
	creditNote := "信用为永久账户，不按天重置；低于" + strconv.Itoa(creditConfig.CreateRestrictedBelow) + "分不能发局，低于" + strconv.Itoa(creditConfig.JoinRestrictedBelow) + "分不能报名或接受邀请，低于" + strconv.Itoa(creditConfig.FrozenBelow) + "分仅可查看和申诉"
	monthlyDelta := 0
	positiveCount := 0
	negativeCount := 0
	appealableCount := 0
	firstAppealRoute := ""
	appealsByCreditLogID := make(map[int64]creditAppealSummary)
	appealsByID := make(map[int64]creditAppealSummary)
	appeals, err := s.reports.AppealsStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取信用申诉记录失败，请稍后重试")
		return
	}
	for _, appeal := range appeals {
		if appeal.ReportType != "credit_appeal" || appeal.CreditLogID <= 0 {
			continue
		}
		summary := creditAppealSummary{
			ID:         appeal.ID,
			Status:     appeal.Status,
			Outcome:    appeal.HandleOutcome,
			StatusText: creditAppealStatusText(appeal.Status, appeal.HandleOutcome),
		}
		appealsByCreditLogID[appeal.CreditLogID] = summary
		appealsByID[appeal.ID] = summary
	}
	now := time.Now()
	records := make([]map[string]interface{}, 0, len(trace.CreditLogs))
	for _, log := range trace.CreditLogs {
		if log.CreatedAt.Year() == now.Year() && log.CreatedAt.Month() == now.Month() {
			monthlyDelta += log.ChangeValue
		}
		if log.ChangeValue > 0 {
			positiveCount++
		} else if log.ChangeValue < 0 {
			negativeCount++
		}
		appeal, appealed := appealsByCreditLogID[log.ID]
		if !appealed && log.AppealID > 0 {
			appeal, appealed = appealsByID[log.AppealID]
			if !appealed {
				appeal = creditAppealSummary{ID: log.AppealID, StatusText: "已提交申诉"}
			}
			appealed = true
		}
		canAppeal := log.ChangeValue < 0 && !appealed
		appealRoute := ""
		if appealed && appeal.ID > 0 {
			appealRoute = "/pages/profile/system-management/appeal-detail/index?reportId=" + strconv.FormatInt(appeal.ID, 10)
		} else if canAppeal {
			appealRoute = "/pages/profile/system-management/credit-appeal/index?creditLogId=" + strconv.FormatInt(log.ID, 10)
			appealableCount++
			if firstAppealRoute == "" {
				firstAppealRoute = appealRoute
			}
		}
		records = append(records, map[string]interface{}{
			"id":               log.ID,
			"creditLogId":      log.ID,
			"title":            creditReasonTitle(log.Reason),
			"desc":             creditReasonDesc(log),
			"score":            signedIntText(log.ChangeValue),
			"tone":             creditTone(log.ChangeValue),
			"gameId":           log.GameID,
			"canAppeal":        canAppeal,
			"appealId":         appeal.ID,
			"appealStatus":     appeal.Status,
			"appealStatusText": appeal.StatusText,
			"appealRoute":      appealRoute,
			"reason":           log.Reason,
			"beforeScore":      log.BeforeScore,
			"afterScore":       log.AfterScore,
			"createdAt":        log.CreatedAt,
		})
	}

	httpx.OK(w, map[string]interface{}{
		"score":       profile.CreditScore,
		"scoreLabel":  "信用分",
		"todayScore":  profile.TodayCreditScore,
		"level":       creditStateText,
		"status":      creditState,
		"isPermanent": true,
		"restrictions": map[string]interface{}{
			"createRestrictedBelow": creditConfig.CreateRestrictedBelow,
			"joinRestrictedBelow":   creditConfig.JoinRestrictedBelow,
			"frozenBelow":           creditConfig.FrozenBelow,
		},
		"monthlyDelta": signedIntText(monthlyDelta),
		"summary": []map[string]string{
			{"label": "当前状态", "value": creditStateText},
			{"label": "正向记录", "value": strconv.Itoa(positiveCount) + "条"},
			{"label": "扣分记录", "value": strconv.Itoa(negativeCount) + "条"},
		},
		"records": records,
		"appealEntry": map[string]interface{}{
			"enabled": appealableCount > 0,
			"count":   appealableCount,
			"route":   singleCreditAppealRoute(appealableCount, firstAppealRoute),
			"text":    creditAppealEntryText(appealableCount),
		},
		"bottomNote": creditNote + "。",
	})
}

type creditAppealSummary struct {
	ID         int64
	Status     string
	Outcome    string
	StatusText string
}

func creditAppealStatusText(status string, outcome string) string {
	switch strings.ToLower(strings.TrimSpace(outcome)) {
	case "appeal_approved":
		return "申诉已通过"
	case "appeal_rejected":
		return "申诉未通过"
	case "processing":
		return "申诉处理中"
	}
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "appealed", "assigned", "pending":
		return "申诉处理中"
	case "appeal_withdrawn":
		return "申诉已撤回"
	case "handled", "closed":
		return "申诉已处理"
	default:
		return "已提交申诉"
	}
}

func creditAppealEntryText(count int) string {
	if count <= 0 {
		return "暂无可申诉记录"
	}
	return "信用申诉（" + strconv.Itoa(count) + "）"
}

func singleCreditAppealRoute(count int, route string) string {
	if count == 1 {
		return route
	}
	return ""
}

func (s *Server) myFootprints(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	items, err := s.reviews.FootprintsStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取成长足迹失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
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
	case "player_cancel_service":
		return "玩家取消已确认服务"
	case "expert_cancel_service":
		return "行家取消已确认服务"
	case "malicious_report":
		return "恶意举报"
	default:
		return "信用变更"
	}
}

func creditReasonDesc(log reviews.CreditLog) string {
	title := creditReasonTitle(log.Reason)
	if log.GameID > 0 {
		return title + " · 局ID " + strconv.FormatInt(log.GameID, 10)
	}
	return title
}

func signedIntText(value int) string {
	if value > 0 {
		return "+" + strconv.Itoa(value)
	}
	return strconv.Itoa(value)
}

func creditTone(value int) string {
	if value > 0 {
		return "plus"
	}
	if value < 0 {
		return "minus"
	}
	return "neutral"
}

func (s *Server) adminUserGrowth(w http.ResponseWriter, r *http.Request) {
	userID, ok := idFromAdminPath(w, r.URL.Path, "/api/admin/users/", "/growth")
	if !ok {
		return
	}
	trace, err := s.reviews.TraceByUserStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取成长与信用数据失败，请稍后重试")
		return
	}
	httpx.OK(w, struct {
		reviews.Trace
		RoleProfiles []roleGrowthProfileDTO `json:"roleProfiles"`
	}{
		Trace:        trace,
		RoleProfiles: s.roleGrowthProfilesForUser(userID, trace.Profile),
	})
}

func (s *Server) adminGameReviewTrace(w http.ResponseWriter, r *http.Request) {
	gameID, ok := idFromAdminPath(w, r.URL.Path, "/api/admin/games/", "/review-trace")
	if !ok {
		return
	}
	trace, err := s.reviews.TraceByGameStrict(gameID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取局内成长与信用轨迹失败，请稍后重试")
		return
	}
	httpx.OK(w, trace)
}

func idFromAdminPath(w http.ResponseWriter, path string, prefix string, suffix string) (int64, bool) {
	idText := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	id, err := strconv.ParseInt(strings.Trim(idText, "/"), 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "记录编号错误")
		return 0, false
	}
	return id, true
}

func writeReviewError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, games.ErrGameNotFound):
		httpx.Error(w, http.StatusNotFound, 40421, "组局不存在或已被删除")
	case errors.Is(err, reviews.ErrForbidden):
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "当前账号无评价权限")
	case errors.Is(err, reviews.ErrGameNotReviewable):
		httpx.Error(w, http.StatusConflict, 40931, "当前组局暂不满足评价条件")
	case errors.Is(err, reviews.ErrDuplicateReview):
		httpx.Error(w, http.StatusConflict, 40932, "请勿重复提交评价")
	case errors.Is(err, reviews.ErrInvalidReview):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "评价内容不符合要求")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "评价操作失败，请稍后重试")
	}
}
