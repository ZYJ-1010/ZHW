package appapi

import (
	"net/http"

	"zhw-mini/services/go-api/internal/common/httpx"
)

const mapMyCityConfigKey = "map.my_city_config"

type mapMyCityStatDTO struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Tone  string `json:"tone"`
}

type mapMyCityLegendDTO struct {
	Label string `json:"label"`
	Tone  string `json:"tone"`
}

type mapMyCityMapStatDTO struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

type mapMyCityStoryDTO struct {
	ID            string   `json:"id"`
	Type          string   `json:"type,omitempty"`
	Tag           string   `json:"tag"`
	SubTag        string   `json:"subTag,omitempty"`
	TagTone       string   `json:"tagTone"`
	HighlightTag  string   `json:"highlightTag,omitempty"`
	Title         string   `json:"title"`
	Date          string   `json:"date"`
	SortDate      string   `json:"sortDate"`
	Location      string   `json:"location"`
	Cover         string   `json:"cover,omitempty"`
	CoverLocation string   `json:"coverLocation,omitempty"`
	Desc          string   `json:"desc"`
	Participants  []string `json:"participants"`
	BadgeTitle    string   `json:"badgeTitle"`
	BadgeDesc     string   `json:"badgeDesc"`
	ActionText    string   `json:"actionText"`
	ActionTone    string   `json:"actionTone"`
}

type mapMyCityStoryGroupDTO struct {
	Year    int                 `json:"year"`
	Stories []mapMyCityStoryDTO `json:"stories"`
}

type mapMyCityConfigDTO struct {
	OnlineText       string                   `json:"onlineText"`
	PageTitle        string                   `json:"pageTitle"`
	ProfileName      string                   `json:"profileName"`
	RouteTip         string                   `json:"routeTip"`
	SwitchMapText    string                   `json:"switchMapText"`
	JourneyTitle     string                   `json:"journeyTitle"`
	JourneyDesc      string                   `json:"journeyDesc"`
	ParticipantLabel string                   `json:"participantLabel"`
	EndingTitle      string                   `json:"endingTitle"`
	EndingDesc       string                   `json:"endingDesc"`
	SummaryStats     []mapMyCityStatDTO       `json:"summaryStats"`
	MapLegends       []mapMyCityLegendDTO     `json:"mapLegends"`
	MapStats         []mapMyCityMapStatDTO    `json:"mapStats"`
	TagEmojis        map[string]string        `json:"tagEmojis"`
	BadgeEmojis      map[string]string        `json:"badgeEmojis"`
	StoryGroups      []mapMyCityStoryGroupDTO `json:"storyGroups"`
	Version          string                   `json:"version"`
}

func (s *Server) mapMyCity(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.currentMapMyCityConfig())
}

func (s *Server) currentMapMyCityConfig() mapMyCityConfigDTO {
	var stored mapMyCityConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(mapMyCityConfigKey, &stored) && len(stored.StoryGroups) > 0 {
		return stored
	}
	return defaultMapMyCityConfig()
}

func defaultMapMyCityConfig() mapMyCityConfigDTO {
	return mapMyCityConfigDTO{
		OnlineText:       "在线",
		PageTitle:        "我的城市故事",
		ProfileName:      "我的信息",
		RouteTip:         "收集8条更早行程，航线图更完整",
		SwitchMapText:    "切换为火车",
		JourneyTitle:     "我的局迹",
		JourneyDesc:      "通过 12 个局，认识了 28 位朋友",
		ParticipantLabel: "参与者：",
		EndingTitle:      "我是有底线的",
		EndingDesc:       "继续探索，创造更多故事",
		SummaryStats: []mapMyCityStatDTO{
			{Value: "28", Label: "故事总数", Tone: "blue"},
			{Value: "6", Label: "覆盖城市", Tone: "violet"},
			{Value: "52", Label: "参与组局数", Tone: "pink"},
		},
		MapLegends: []mapMyCityLegendDTO{
			{Label: "第一次", Tone: "pink"},
			{Label: "夜游", Tone: "violet"},
			{Label: "社交局", Tone: "blue"},
			{Label: "最难忘", Tone: "gold"},
		},
		MapStats: []mapMyCityMapStatDTO{
			{Label: "里程", Value: "40244", Unit: "公里"},
			{Label: "次数", Value: "33", Unit: "次"},
			{Label: "国家/地区", Value: "1", Unit: "个"},
			{Label: "城市", Value: "10", Unit: "个"},
		},
		TagEmojis: map[string]string{
			"最难忘": "👑",
			"桌游局": "🎲",
			"微醺局": "🍷",
			"脑暴局": "💡",
			"篮球局": "🏀",
			"第一次": "🌱",
			"起点":  "🌱",
			"摄影局": "📷",
		},
		BadgeEmojis: map[string]string{
			"创业伙伴": "🤝",
			"深度密友": "💬",
			"合伙人":  "🤝",
			"室友":   "🏠",
			"固定局友": "📌",
		},
		StoryGroups: []mapMyCityStoryGroupDTO{
			{
				Year: 2024,
				Stories: []mapMyCityStoryDTO{
					{
						ID:            "countdown-night",
						Tag:           "最难忘",
						TagTone:       "gold",
						Title:         "跨年夜的倒计时",
						Date:          "12.31",
						SortDate:      "2024-12-31",
						Cover:         "/pages/map/my-city/assets/story-river-cover.png",
						CoverLocation: "上海 · 外滩",
						Desc:          "和刚认识的摄影局朋友们一起在外滩等待新年钟声。江风吹得发抖，但倒数的那一刻，所有的陌生人都变成了朋友。",
						Participants:  []string{"a", "b", "c"},
					},
					{
						ID:           "script-rain",
						Tag:          "桌游局",
						SubTag:       "新手场",
						TagTone:      "violet",
						Title:        "暴雨中的剧本杀",
						Date:         "2024.09.20",
						SortDate:     "2024-09-20",
						Location:     "杭州 · 西湖区 · 14:00-22:00",
						Desc:         "原定5人的局因为暴雨只来了3人，却因此有了最深入的交谈。认识了做AI的@阿杰，现在我们是创业合伙人。",
						Participants: []string{"a", "b", "c", "d"},
						BadgeTitle:   "创业伙伴",
						BadgeDesc:    "已共同发起 3 个项目",
						ActionText:   "查看项目",
						ActionTone:   "violet",
					},
					{
						ID:           "truth-night",
						Tag:          "微醺局",
						SubTag:       "深夜场",
						TagTone:      "orange",
						Title:        "周五晚上的坦白局",
						Date:         "2024.11.03",
						SortDate:     "2024-11-03",
						Location:     "北京 · 三里屯 · 21:00",
						Desc:         "\"你最后悔的事是什么？\"那个问题让陌生人变成了知己。和@Lucy约定每月一次深度对话。",
						Participants: []string{"a", "b"},
						BadgeTitle:   "深度密友",
						ActionText:   "约下次",
						ActionTone:   "orange",
					},
					{
						ID:           "business-canvas",
						Tag:          "脑暴局",
						SubTag:       "创始人专场",
						TagTone:      "gold",
						Title:        "凌晨的商业模式画布",
						Date:         "2024.12.15",
						SortDate:     "2024-12-15",
						Location:     "深圳 · 科技园 · 通宵",
						Desc:         "从晚上8点到早上6点，8个人在黑板上画满了想法。这个局让我找到了技术合伙人@老K。",
						Participants: []string{"a", "b", "c", "d"},
						BadgeTitle:   "合伙人",
						BadgeDesc:    "公司估值 500w",
						ActionText:   "查看公司",
						ActionTone:   "gold",
					},
					{
						ID:           "court-weekly",
						Tag:          "篮球局",
						SubTag:       "每周固定",
						TagTone:      "cyan",
						Title:        "东华球场的汗水",
						Date:         "每周六",
						SortDate:     "2024-06-15",
						Location:     "上海 · 东华大学 · 16:00",
						Desc:         "最纯粹的快乐。这里没有身份，只有队友。通过球局认识了现在的室友@阿强。",
						Participants: []string{"a", "b", "c"},
						BadgeTitle:   "室友",
						BadgeDesc:    "合租 6 个月",
						ActionText:   "加入球局",
						ActionTone:   "cyan",
					},
					{
						ID:           "first-use",
						Type:         "first",
						Tag:          "起点",
						TagTone:      "green",
						HighlightTag: "第一次",
						Title:        "第一次使用真好玩",
						Date:         "03.12",
						SortDate:     "2024-03-12",
						Location:     "广州 · 天河公园",
						Desc:         "抱着试试看的心态参加了第一次飞盘局，从此打开了城市探索的新世界。",
					},
				},
			},
			{
				Year: 2023,
				Stories: []mapMyCityStoryDTO{
					{
						ID:           "sunrise-photo",
						Tag:          "摄影局",
						SubTag:       "第1次参与",
						TagTone:      "pink",
						Title:        "外滩 sunrise 拍摄",
						Date:         "2023.03.12",
						SortDate:     "2023-03-12",
						Location:     "上海 · 外滩观景台 · 06:00",
						Desc:         "为了拍日出早上5点起床，认识了同样疯狂的@小林和@大为。后来我们组成了固定摄影小队，每周六早扫街。",
						Participants: []string{"a", "b", "c"},
						BadgeTitle:   "固定局友",
						BadgeDesc:    "已持续组队 8 个月",
						ActionText:   "再组一局",
						ActionTone:   "blue",
					},
				},
			},
		},
		Version: "2026-06-30",
	}
}
