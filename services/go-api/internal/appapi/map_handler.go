package appapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/lbs"
)

const (
	mapIndexConfigKey      = "map.index_config"
	mapPlayPagesConfigKey  = "map.play_pages_config"
	mapProviderMinInterval = 2 * time.Second
)

type mapIndexLocationDTO struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type mapIndexConfigDTO struct {
	OnlineText          string              `json:"onlineText"`
	DefaultLocation     mapIndexLocationDTO `json:"defaultLocation"`
	DefaultRadiusMeters int                 `json:"defaultRadiusMeters"`
	RadiusOptions       []int               `json:"radiusOptions"`
	MapFilters          []string            `json:"mapFilters"`
	Texts               map[string]string   `json:"texts"`
	Version             string              `json:"version"`
}

type mapCheckinRequest struct {
	PointID     string  `json:"pointId"`
	ChallengeID string  `json:"challengeId"`
	GameID      string  `json:"gameId"`
	Story       string  `json:"story"`
	FileIDs     []int64 `json:"fileIds"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
}

type mapBlindRouteDTO struct {
	ID         int64  `json:"id"`
	CardID     string `json:"cardId"`
	Title      string `json:"title"`
	TimeText   string `json:"timeText"`
	Status     string `json:"status"`
	StatusText string `json:"statusText"`
	CreatedAt  string `json:"createdAt"`
}

type mapChallengeDTO struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	TimeLeft     string `json:"timeLeft"`
	StatusText   string `json:"statusText"`
	StatusTone   string `json:"statusTone"`
	SelfValue    int    `json:"selfValue"`
	RivalValue   int    `json:"rivalValue"`
	Total        int    `json:"total"`
	SelfPercent  int    `json:"selfPercent"`
	RivalPercent int    `json:"rivalPercent"`
	CreatedAt    string `json:"createdAt"`
}

func (s *Server) mapIndexConfig(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.currentMapIndexConfig())
}

func (s *Server) currentMapIndexConfig() mapIndexConfigDTO {
	var stored mapIndexConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(mapIndexConfigKey, &stored) && len(stored.MapFilters) > 0 {
		return normalizeMapIndexConfig(stored)
	}
	return defaultMapIndexConfig()
}

func normalizeMapIndexConfig(config mapIndexConfigDTO) mapIndexConfigDTO {
	defaults := defaultMapIndexConfig()
	if strings.TrimSpace(config.OnlineText) == "" {
		config.OnlineText = defaults.OnlineText
	}
	if config.DefaultLocation.Latitude == 0 || config.DefaultLocation.Longitude == 0 {
		config.DefaultLocation = defaults.DefaultLocation
	}
	if config.DefaultRadiusMeters <= 0 {
		config.DefaultRadiusMeters = defaults.DefaultRadiusMeters
	}
	if len(config.RadiusOptions) == 0 {
		config.RadiusOptions = defaults.RadiusOptions
	}
	config.MapFilters = defaults.MapFilters
	if config.Texts == nil {
		config.Texts = map[string]string{}
	}
	for key, value := range defaults.Texts {
		if strings.TrimSpace(config.Texts[key]) == "" {
			config.Texts[key] = value
		}
	}
	if strings.TrimSpace(config.Version) == "" {
		config.Version = defaults.Version
	}
	return config
}

func defaultMapIndexConfig() mapIndexConfigDTO {
	return mapIndexConfigDTO{
		OnlineText: "在线",
		DefaultLocation: mapIndexLocationDTO{
			Latitude:  31.2304,
			Longitude: 121.4737,
		},
		DefaultRadiusMeters: 3000,
		RadiusOptions:       []int{1000, 3000, 5000},
		MapFilters:          []string{"附近组局"},
		Texts: map[string]string{
			"searchPlaceholder":      "搜局、搜人、搜地块...",
			"loadingNearbyText":      "正在获取附近数据...",
			"locateToolText":         "定",
			"refreshToolText":        "刷",
			"dateRangeText":          "2025.09.23 - 2026.03.30",
			"detailActionText":       "查看详情",
			"joinActionText":         "去组队",
			"playerDetailActionText": "查看资料",
			"currentInteractText":    "当前可互动",
			"offlineInteractText":    "暂未在线，可查看轨迹",
			"nearbySectionTitle":     "附近玩法点",
			"nearbyEmptyText":        "当前位置附近暂无玩法点",
		},
		Version: "2026-07-01",
	}
}

func (s *Server) mapPlayPages(w http.ResponseWriter, r *http.Request) {
	config := s.currentMapPlayPagesConfig()
	pageKey := strings.TrimSpace(r.URL.Query().Get("pageKey"))
	if pageKey == "" {
		httpx.OK(w, config)
		return
	}
	pages, _ := config["pages"].(map[string]interface{})
	if page, ok := pages[pageKey]; ok {
		httpx.OK(w, page)
		return
	}
	httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "地图玩法配置不存在")
}

func (s *Server) submitMapCheckin(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req mapCheckinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	req.PointID = strings.TrimSpace(req.PointID)
	req.ChallengeID = strings.TrimSpace(req.ChallengeID)
	req.GameID = strings.TrimSpace(req.GameID)
	req.Story = strings.TrimSpace(req.Story)
	if req.PointID == "" && req.ChallengeID == "" && req.GameID == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "缺少地图打卡对象")
		return
	}
	if req.Story == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请填写打卡故事")
		return
	}
	if len(req.FileIDs) == 0 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请上传打卡照片")
		return
	}
	if req.GameID != "" {
		gameID, err := strconv.ParseInt(req.GameID, 10, 64)
		if err != nil || gameID <= 0 {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "局 ID 错误")
			return
		}
		game, err := s.games.Get(gameID)
		if err != nil {
			writeGameError(w, err)
			return
		}
		if !s.games.IsMember(gameID, userID) {
			httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "仅局内成员可提交组局打卡")
			return
		}
		if game.Status != "in_progress" && game.Status != "pending_confirm" {
			httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "当前局不可签到")
			return
		}
	}

	now := time.Now()
	id := now.UnixNano()
	s.recordBehavior(userID, "submit_map_checkin", "map_checkin", id, map[string]interface{}{
		"pointId":     req.PointID,
		"challengeId": req.ChallengeID,
		"gameId":      req.GameID,
		"fileIds":     req.FileIDs,
	})
	httpx.OK(w, map[string]interface{}{
		"id":          id,
		"pointId":     req.PointID,
		"challengeId": req.ChallengeID,
		"gameId":      req.GameID,
		"status":      "pending_audit",
		"statusText":  "待审核",
		"statusDesc":  "打卡材料已提交，等待后台审核",
		"createdAt":   now.Format(time.RFC3339),
	})
}

func (s *Server) createMapBlindRoute(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req struct {
		CardID string `json:"cardId"`
		Title  string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	req.CardID = strings.TrimSpace(req.CardID)
	req.Title = strings.TrimSpace(req.Title)
	if req.CardID == "" || req.Title == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请选择路线")
		return
	}
	now := time.Now()
	item := mapBlindRouteDTO{
		ID:         now.UnixNano(),
		CardID:     req.CardID,
		Title:      req.Title,
		TimeText:   "刚刚开启",
		Status:     "opened",
		StatusText: "进行中",
		CreatedAt:  now.Format(time.RFC3339),
	}
	s.mapPlayMu.Lock()
	if s.mapBlindRoutes == nil {
		s.mapBlindRoutes = make(map[int64]mapBlindRouteDTO)
	}
	s.mapBlindRoutes[item.ID] = item
	s.mapPlayMu.Unlock()
	s.recordBehavior(userID, "create_map_blind_route", "map_blind_route", item.ID, map[string]interface{}{"cardId": item.CardID})
	httpx.OK(w, item)
}

func (s *Server) routeMapBlindRoutePost(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/complete") {
		s.completeMapBlindRoute(w, r)
		return
	}
	httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "地图路线接口不存在")
}

func (s *Server) completeMapBlindRoute(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	idText := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/app/map/blind-routes/"), "/complete")
	id, err := strconv.ParseInt(strings.Trim(idText, "/"), 10, 64)
	if err != nil || id <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "路线 ID 错误")
		return
	}
	s.mapPlayMu.Lock()
	item, exists := s.mapBlindRoutes[id]
	if !exists {
		item = mapBlindRouteDTO{ID: id, Title: "盲盒路线"}
	}
	item.Status = "completed"
	item.StatusText = "已完成"
	item.TimeText = "刚刚完成"
	s.mapBlindRoutes[id] = item
	s.mapPlayMu.Unlock()
	s.recordBehavior(userID, "complete_map_blind_route", "map_blind_route", id, nil)
	httpx.OK(w, item)
}

func (s *Server) createMapChallenge(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req struct {
		Title string `json:"title"`
		Total int    `json:"total"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		req.Title = "城市挑战"
	}
	if req.Total <= 0 {
		req.Total = 3
	}
	now := time.Now()
	item := mapChallengeDTO{
		ID:           now.UnixNano(),
		Title:        req.Title,
		TimeLeft:     "24小时",
		StatusText:   "进行中",
		StatusTone:   "active",
		SelfValue:    0,
		RivalValue:   0,
		Total:        req.Total,
		SelfPercent:  0,
		RivalPercent: 0,
		CreatedAt:    now.Format(time.RFC3339),
	}
	s.mapPlayMu.Lock()
	if s.mapChallenges == nil {
		s.mapChallenges = make(map[int64]mapChallengeDTO)
	}
	s.mapChallenges[item.ID] = item
	s.mapPlayMu.Unlock()
	s.recordBehavior(userID, "create_map_challenge", "map_challenge", item.ID, nil)
	httpx.OK(w, item)
}

func (s *Server) currentMapPlayPagesConfig() map[string]interface{} {
	var config map[string]interface{}
	if s.systemConfig != nil && s.systemConfig.Get(mapPlayPagesConfigKey, &config) {
		return mergeMapPlayPagesConfig(defaultMapPlayPagesConfig(), config)
	}
	return defaultMapPlayPagesConfig()
}

func mergeMapPlayPagesConfig(defaults map[string]interface{}, config map[string]interface{}) map[string]interface{} {
	result := cloneObjectMap(defaults)
	for key, value := range config {
		if key != "pages" {
			result[key] = value
			continue
		}
		defaultPages, _ := result["pages"].(map[string]interface{})
		pages, _ := value.(map[string]interface{})
		if defaultPages == nil || pages == nil {
			result[key] = value
			continue
		}
		mergedPages := cloneObjectMap(defaultPages)
		for pageKey, pageValue := range pages {
			mergedPages[pageKey] = pageValue
		}
		result["pages"] = mergedPages
	}
	return result
}

func defaultMapPlayPagesConfig() map[string]interface{} {
	return map[string]interface{}{
		"version": "2026-07-01",
		"pages": map[string]interface{}{
			"blind-route": map[string]interface{}{
				"title":        "组局盲盒",
				"description":  "不知道去哪局？让命运决定你的下一次城市冒险",
				"sectionTitle": "最近开启",
				"selectToast":  "已选择 {title}",
				"cards": []map[string]interface{}{
					{"id": "walk", "title": "城市漫步盲盒", "desc": "30 分钟内出发，随机匹配附近轻量局", "tone": "blue", "icon": "/pages/map/blind-route/assets/i50.png", "tags": []string{"轻松", "附近", "低门槛"}},
					{"id": "food", "title": "深夜食堂盲盒", "desc": "匹配同城饭搭子和夜宵路线", "tone": "orange", "icon": "/pages/map/blind-route/assets/i52.png", "tags": []string{"饭局", "夜间", "社交"}},
					{"id": "photo", "title": "拍照路线盲盒", "desc": "用一个主题串起三处城市机位", "tone": "purple", "icon": "/pages/map/blind-route/assets/i54.png", "tags": []string{"拍照", "路线", "打卡"}},
				},
				"recentRoutes": []map[string]interface{}{
					{"title": "徐汇夜风路线", "timeText": "20 分钟前", "statusText": "已成局"},
					{"title": "周末咖啡搭子", "timeText": "1 小时前", "statusText": "招募中"},
				},
			},
			"city-atlas": map[string]interface{}{
				"title":             "上海探索图鉴",
				"progressLabel":     "探索进度",
				"lockedText":        "未解锁",
				"hiddenBadge":       "隐藏",
				"routeSectionTitle": "主题路线",
				"filters":           []string{"全部", "已解锁", "未解锁", "隐藏点"},
				"unlockInfo":        map[string]interface{}{"unlockedCount": 1, "totalCount": 36, "progressPercent": 33},
				"unlockToast":       "{name}已解锁",
				"lockedToast":       "{name}待解锁",
				"unlockPoints": []map[string]interface{}{
					{"id": "bund-night", "name": "外滩夜景", "statusType": "unlocked", "unlockText": "2024.01.15 解锁", "footprintValue": "+50 足迹值", "tone": "blue", "iconType": "building", "checked": true},
					{"id": "tianzifang", "name": "田子坊", "statusType": "unlocked", "unlockText": "2024.02.03 解锁", "footprintValue": "+30 足迹值", "tone": "green", "iconType": "lantern", "checked": true},
					{"id": "wukang-road", "name": "武康路街角", "statusType": "locked", "unlockText": "完成 2 次附近打卡后解锁", "footprintValue": "+40 足迹值", "tone": "locked", "iconType": "lock", "checked": false},
					{"id": "hidden-rooftop", "name": "城市天台", "statusType": "hidden", "unlockText": "隐藏点待发现", "footprintValue": "+80 足迹值", "tone": "purple", "iconType": "hidden", "checked": false},
				},
				"themeRoutes": []map[string]interface{}{
					{"id": "couple-walk", "name": "情侣漫步", "meta": "6个地点 · 预计3小", "progressText": "已解锁 2/6", "tone": "sunset", "iconText": "💕"},
					{"id": "coffee-shop", "name": "咖啡探店", "meta": "8个地点 · 预计4小", "progressText": "已解锁 0/8", "tone": "cyan", "iconText": "☕"},
				},
			},
			"footprint-heatmap": map[string]interface{}{
				"title":          "足迹热力图",
				"heatTitle":      "全国城市打卡热力",
				"legendLabel":    "城市打卡热度",
				"friendTitle":    "好友也在打卡",
				"friendMoreText": "查看全部",
				"hotTitle":       "城市热点排行",
				"rangeTabs":      []string{"今日", "本周", "本月", "全部"},
				"rangeStats": map[string]interface{}{
					"今日": []map[string]interface{}{{"value": "12", "label": "打卡城市"}, {"value": "1.8k", "label": "玩家足迹"}, {"value": "3", "label": "热门城市"}},
					"本周": []map[string]interface{}{{"value": "38", "label": "打卡城市"}, {"value": "8.5k", "label": "玩家足迹"}, {"value": "9", "label": "热门城市"}},
					"本月": []map[string]interface{}{{"value": "76", "label": "打卡城市"}, {"value": "26k", "label": "玩家足迹"}, {"value": "18", "label": "热门城市"}},
					"全部": []map[string]interface{}{{"value": "126", "label": "打卡城市"}, {"value": "92k", "label": "玩家足迹"}, {"value": "31", "label": "热门城市"}},
				},
				"cityHeatPoints": []map[string]interface{}{
					{"id": "beijing", "city": "北京", "level": "mid", "className": "footprint-city-point beijing level-mid"},
					{"id": "shanghai", "city": "上海", "level": "hot", "className": "footprint-city-point shanghai level-hot"},
					{"id": "chengdu", "city": "成都", "level": "hot", "className": "footprint-city-point chengdu level-hot"},
					{"id": "guangzhou", "city": "广州", "level": "mid", "className": "footprint-city-point guangzhou level-mid"},
					{"id": "shenzhen", "city": "深圳", "level": "high", "className": "footprint-city-point shenzhen level-high"},
					{"id": "xian", "city": "西安", "level": "low", "className": "footprint-city-point xian level-low"},
					{"id": "hangzhou", "city": "杭州", "level": "high", "className": "footprint-city-point hangzhou level-high"},
				},
				"friendUpdates": []map[string]interface{}{
					{"id": "alex", "avatarText": "AL", "name": "Alex", "desc": "刚刚在成都宽窄巷子打卡", "online": true},
					{"id": "sarah", "avatarText": "SA", "name": "Sarah", "desc": "25分钟前在西安城墙打卡", "online": false},
				},
				"hotCities": []map[string]interface{}{
					{"id": "shanghai", "rank": 1, "city": "上海市中心", "desc": "2456人在这里打卡", "progress": 86, "level": "hot"},
					{"id": "chengdu", "rank": 2, "city": "成都市", "desc": "1892人在这里打卡", "progress": 72, "level": "warm"},
					{"id": "shenzhen", "rank": 3, "city": "深圳湾", "desc": "1567人在这里打卡", "progress": 58, "level": "active"},
				},
			},
			"friend-city": map[string]interface{}{
				"title":              "好友",
				"challengeTitle":     "进行中的挑战",
				"rankingTitle":       "城　市　榜",
				"nationalRankText":   "查看全国榜",
				"challengeToast":     "{title}进行中",
				"emptyChallengeText": "暂无挑战详情",
				"rankingToast":       "第{rank}名城市榜",
				"emptyRankingText":   "暂无城市榜详情",
				"nationalToast":      "已展示当前城市榜",
				"duel": map[string]interface{}{
					"selfName": "我", "selfCount": "12区已点亮", "rivalName": "Sarah", "rivalCount": "10区已点亮", "vsText": "VS", "subtitle": "友谊赛", "startButtonText": "发起挑战", "recordButtonText": "查看记录",
				},
				"challenges": []map[string]interface{}{
					{"id": "jingan-first", "title": "率先点亮静安区", "timeLeft": "2天", "statusText": "进行中", "statusTone": "pending", "selfValue": 3, "rivalValue": 2, "total": 5, "selfPercent": 60, "rivalPercent": 40},
					{"id": "landmark-speed", "title": "10个地标速通", "timeLeft": "5天", "statusText": "领先中", "statusTone": "leading", "selfValue": 7, "rivalValue": 4, "total": 10, "selfPercent": 70, "rivalPercent": 40},
				},
				"rankings": []map[string]interface{}{
					{"rank": 1, "name": "Mike", "desc": "已点亮 28 区", "score": "2,450", "tone": "gold"},
					{"rank": 2, "name": "Sarah", "desc": "已点亮 24 区", "score": "2,180", "tone": "silver"},
					{"rank": 3, "name": "David", "desc": "已点亮 22 区", "score": "1,950", "tone": "bronze"},
					{"rank": 4, "name": "我", "desc": "已点亮 12 区", "score": "1,240", "tone": "normal"},
				},
			},
			"real-checkin": map[string]interface{}{
				"taskSectionTitle":   "选择打卡任务",
				"generateButtonText": "生成足迹碎片",
				"texts":              map[string]interface{}{"unsupportedCamera": "当前基础库不支持拍照", "photoSelected": "打卡照片已选择", "storySaved": "打卡文字已记录", "storyRequired": "请先填写打卡文字", "taskSelected": "{title}已选中", "taskRequired": "请选择打卡任务", "fragmentPending": "足迹碎片待生成"},
				"checkinDetail":      map[string]interface{}{"distanceText": "距离目标 15米", "spotName": "外滩观景台", "statusTitle": "地点已解锁", "statusDesc": "完成打卡任务获得足迹值", "storyTitle": "留下你的故事", "storyPlaceholder": "用20个字记录此刻的心情...", "storyMinLength": 20, "storyMaxLength": 120, "rewards": []string{"+20 足迹值", "+1 成就点"}},
				"checkinTasks": []map[string]interface{}{
					{"id": "photo", "title": "拍摄地标合影", "desc": "与标志性建筑合影", "scoreText": "+10分"},
					{"id": "angle", "title": "发现隐藏角度", "desc": "拍摄独特的视角", "scoreText": "+20分"},
				},
			},
		},
	}
}

func (s *Server) mapSearch(w http.ResponseWriter, r *http.Request) {
	if !s.ensureMapProvider(w) {
		return
	}
	if !s.ensureMapProviderRequestAllowed(w, r) {
		return
	}
	query := r.URL.Query()
	radius, err := optionalQueryFloat(query.Get("radiusMeter"))
	if err != nil {
		writeMapResponse(w, nil, lbs.ErrMapRequestInvalid)
		return
	}
	req := lbs.MapSearchRequest{
		Keyword:     query.Get("keyword"),
		City:        query.Get("city"),
		RadiusMeter: radius,
		Page:        queryInt(query.Get("page")),
		PageSize:    queryInt(query.Get("pageSize")),
	}
	if rawLon := query.Get("longitude"); rawLon != "" {
		longitude, err := requiredQueryFloat(rawLon)
		if err != nil {
			writeMapResponse(w, nil, lbs.ErrMapRequestInvalid)
			return
		}
		latitude, err := requiredQueryFloat(query.Get("latitude"))
		if err != nil {
			writeMapResponse(w, nil, lbs.ErrMapRequestInvalid)
			return
		}
		req.Longitude = longitude
		req.Latitude = latitude
	}
	result, err := s.mapProvider.Search(r.Context(), req)
	writeMapResponse(w, result, err)
}

func (s *Server) mapGeocode(w http.ResponseWriter, r *http.Request) {
	if !s.ensureMapProvider(w) {
		return
	}
	if !s.ensureMapProviderRequestAllowed(w, r) {
		return
	}
	place, err := s.mapProvider.Geocode(r.Context(), lbs.MapGeocodeRequest{
		Address: r.URL.Query().Get("address"),
		City:    r.URL.Query().Get("city"),
	})
	writeMapResponse(w, map[string]interface{}{"provider": "tencent", "place": place}, err)
}

func (s *Server) mapReverseGeocode(w http.ResponseWriter, r *http.Request) {
	if !s.ensureMapProvider(w) {
		return
	}
	if !s.ensureMapProviderRequestAllowed(w, r) {
		return
	}
	longitude, err := requiredQueryFloat(r.URL.Query().Get("longitude"))
	if err != nil {
		writeMapResponse(w, nil, lbs.ErrMapRequestInvalid)
		return
	}
	latitude, err := requiredQueryFloat(r.URL.Query().Get("latitude"))
	if err != nil {
		writeMapResponse(w, nil, lbs.ErrMapRequestInvalid)
		return
	}
	place, err := s.mapProvider.ReverseGeocode(r.Context(), lbs.MapReverseGeocodeRequest{
		Longitude: longitude,
		Latitude:  latitude,
	})
	writeMapResponse(w, map[string]interface{}{"provider": "tencent", "place": place}, err)
}

func (s *Server) mapRoute(w http.ResponseWriter, r *http.Request) {
	if !s.ensureMapProvider(w) {
		return
	}
	if !s.ensureMapProviderRequestAllowed(w, r) {
		return
	}
	query := r.URL.Query()
	fromLongitude, err := requiredQueryFloat(query.Get("fromLongitude"))
	if err != nil {
		writeMapResponse(w, nil, lbs.ErrMapRequestInvalid)
		return
	}
	fromLatitude, err := requiredQueryFloat(query.Get("fromLatitude"))
	if err != nil {
		writeMapResponse(w, nil, lbs.ErrMapRequestInvalid)
		return
	}
	toLongitude, err := requiredQueryFloat(query.Get("toLongitude"))
	if err != nil {
		writeMapResponse(w, nil, lbs.ErrMapRequestInvalid)
		return
	}
	toLatitude, err := requiredQueryFloat(query.Get("toLatitude"))
	if err != nil {
		writeMapResponse(w, nil, lbs.ErrMapRequestInvalid)
		return
	}
	route, err := s.mapProvider.Route(r.Context(), lbs.MapRouteRequest{
		FromLongitude: fromLongitude,
		FromLatitude:  fromLatitude,
		ToLongitude:   toLongitude,
		ToLatitude:    toLatitude,
		Mode:          query.Get("mode"),
	})
	writeMapResponse(w, route, err)
}

func (s *Server) ensureMapProvider(w http.ResponseWriter) bool {
	if s.mapProvider == nil {
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "腾讯地图服务端 Key 未配置")
		return false
	}
	return true
}

func (s *Server) ensureMapProviderRequestAllowed(w http.ResponseWriter, r *http.Request) bool {
	if _, ok := s.mapProvider.(*lbs.TencentMapClient); !ok {
		return true
	}

	key := s.mapProviderLimitKey(r)
	now := time.Now()

	s.mapProviderLimitMu.Lock()
	if s.mapProviderLastSeen == nil {
		s.mapProviderLastSeen = make(map[string]time.Time)
	}
	if last, ok := s.mapProviderLastSeen[key]; ok {
		elapsed := now.Sub(last)
		if elapsed < mapProviderMinInterval {
			remaining := mapProviderMinInterval - elapsed
			retryAfter := int((remaining + time.Second - time.Nanosecond) / time.Second)
			s.mapProviderLimitMu.Unlock()
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			httpx.Error(w, http.StatusTooManyRequests, httpx.CodeValidationError, "地图请求太频繁，请稍后再试")
			return false
		}
	}
	s.mapProviderLastSeen[key] = now
	if len(s.mapProviderLastSeen) > 512 {
		cutoff := now.Add(-time.Minute)
		for itemKey, last := range s.mapProviderLastSeen {
			if last.Before(cutoff) {
				delete(s.mapProviderLastSeen, itemKey)
			}
		}
	}
	s.mapProviderLimitMu.Unlock()
	return true
}

func (s *Server) mapProviderLimitKey(r *http.Request) string {
	if userID, ok := appUserIDFromRequest(r); ok {
		return "app:" + strconv.FormatInt(userID, 10) + ":" + r.URL.Path
	}
	if adminID := strings.TrimSpace(r.Header.Get("X-Admin-ID")); adminID != "" {
		return "admin:" + adminID + ":" + r.URL.Path
	}
	return "ip:" + clientIP(r) + ":" + r.URL.Path
}

func writeMapResponse(w http.ResponseWriter, payload interface{}, err error) {
	if err == nil {
		httpx.OK(w, payload)
		return
	}
	if errors.Is(err, lbs.ErrMapRequestInvalid) {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "地图请求参数错误")
		return
	}
	if errors.Is(err, lbs.ErrMapProviderUnavailable) {
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "地图服务未配置")
		return
	}
	if errors.Is(err, lbs.ErrMapProviderFailed) {
		log.Printf("tencent map provider failed: %v", err)
		httpx.Error(w, http.StatusBadGateway, httpx.CodeSystemError, "腾讯地图服务调用失败")
		return
	}
	log.Printf("map service unexpected error: %v", err)
	httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "地图服务异常")
}

func optionalQueryFloat(value string) (float64, error) {
	if value == "" {
		return 0, nil
	}
	return requiredQueryFloat(value)
}

func requiredQueryFloat(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

func queryInt(value string) int {
	parsed, _ := strconv.Atoi(value)
	return parsed
}
