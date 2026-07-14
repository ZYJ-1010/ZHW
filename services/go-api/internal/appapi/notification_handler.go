package appapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/notifications"
)

var defaultTradeWarningDetail = map[string]interface{}{
	"pageTitle":  "\u4ea4\u6613\u9884\u8b66",
	"onlineText": "\u5728\u7ebf",
	"warning": map[string]interface{}{
		"title":         "\u4ea4\u6613\u9884\u8b66",
		"prefixText":    "",
		"highlightText": "",
		"suffixText":    "",
	},
	"countdown": []map[string]interface{}{},
	"order": map[string]interface{}{
		"orderNo":            "",
		"statusText":         "",
		"customerAvatarText": "",
		"customerTitle":      "",
		"customerDesc":       "",
		"detailRows":         []map[string]interface{}{},
	},
	"deliveryMethods": []map[string]interface{}{
		{"id": "online", "title": "\u7ebf\u4e0a\u786e\u8ba4", "desc": "\u53cc\u65b9\u5728\u7ebf\u786e\u8ba4\u670d\u52a1\u5b8c\u6210", "active": true},
		{"id": "upload", "title": "\u4e0a\u4f20\u51ed\u8bc1", "desc": "\u4e0a\u4f20\u670d\u52a1\u5b8c\u6210\u622a\u56fe\u6216\u6587\u4ef6", "active": false},
	},
	"actions": map[string]interface{}{
		"delayText":   "\u7533\u8bf7\u5ef6\u671f",
		"deliverText": "\u7acb\u5373\u4ea4\u4ed8",
	},
	"texts": map[string]interface{}{
		"loadingText":        "\u52a0\u8f7d\u4e2d...",
		"loadFailedText":     "\u83b7\u53d6\u4ea4\u6613\u9884\u8b66\u5931\u8d25",
		"invalidActionText":  "\u4ea4\u6613\u9884\u8b66\u64cd\u4f5c\u65e0\u6548",
		"actionFailedText":   "\u4ea4\u6613\u9884\u8b66\u5904\u7406\u5931\u8d25",
		"delaySuccessText":   "\u5ef6\u671f\u7533\u8bf7\u5df2\u63d0\u4ea4",
		"delayStatusText":    "\u5df2\u7533\u8bf7\u5ef6\u671f",
		"deliverSuccessText": "\u5df2\u8fdb\u5165\u4ea4\u4ed8\u786e\u8ba4",
		"countdownTitle":     "\u5269\u4f59\u4ea4\u4ed8\u65f6\u95f4",
		"orderNoLabel":       "\u8ba2\u5355\u7f16\u53f7",
		"deliveryTitle":      "\u4ea4\u4ed8\u65b9\u5f0f",
	},
}

const tradeWarningConfigKey = "message.trade_warning_config"

var defaultSystemNotificationDetail = map[string]interface{}{
	"pageTitle":  "\u7cfb\u7edf\u901a\u77e5",
	"onlineText": "\u5728\u7ebf",
	"article": map[string]interface{}{
		"tagText":         "\u7cfb\u7edf\u901a\u77e5",
		"title":           "",
		"author":          "",
		"publishedAtText": "",
		"readText":        "",
		"blocks":          []map[string]interface{}{},
	},
	"feedback": map[string]interface{}{
		"question": "\u8fd9\u7bc7\u901a\u77e5\u5bf9\u4f60\u6709\u5e2e\u52a9\u5417\uff1f",
		"useful":   map[string]interface{}{"icon": "\u8d5e", "label": "\u6709\u7528", "countText": "0"},
		"useless":  map[string]interface{}{"icon": "\u8e29", "label": "\u6ca1\u7528", "countText": "0"},
	},
	"texts": map[string]interface{}{
		"loadFailedText":      "\u83b7\u53d6\u7cfb\u7edf\u901a\u77e5\u5931\u8d25",
		"feedbackFailedText":  "\u53cd\u9988\u63d0\u4ea4\u5931\u8d25",
		"feedbackSuccessText": "\u5df2\u8bb0\u5f55{label}\u53cd\u9988",
	},
}

const systemNotificationConfigKey = "message.system_notification_config"

var defaultMessageCenterConfig = map[string]interface{}{
	"pageTitle":  "\u6d88\u606f\u4e2d\u5fc3",
	"onlineText": "\u5728\u7ebf",
	"quickActions": []map[string]interface{}{
		{"key": "join", "label": "\u7ec4\u5c40\u52a0\u5165", "iconSrc": "/pages/message/assets/i53@3x.png", "tone": "blue", "bucket": "group"},
		{"key": "system", "label": "\u7cfb\u7edf\u901a\u77e5", "iconSrc": "/pages/message/assets/i54@3x.png", "tone": "green", "bucket": "system"},
		{"key": "achievement", "label": "\u6210\u5c31\u89e3\u9501", "iconSrc": "/pages/message/assets/i55@3x.png", "tone": "yellow", "bucket": "achievement"},
		{"key": "warning", "label": "\u9884\u8b66\u901a\u77e5", "iconSrc": "/pages/message/assets/i56@3x.png", "tone": "red", "bucket": "warning"},
		{"key": "friend", "label": "\u597d\u53cb", "iconSrc": "/pages/message/assets/i57@3x.png", "tone": "cyan", "bucket": "friend"},
	},
	"tabs": []map[string]interface{}{
		{"key": "all", "label": "\u5168\u90e8\u6d88\u606f", "countSource": "unread"},
		{"key": "unread", "label": "\u672a\u8bfb", "countSource": "unread"},
		{"key": "trade", "label": "\u4ea4\u6613\u901a\u77e5", "countSource": "trade"},
	},
	"sections": []map[string]interface{}{
		{"key": "group", "title": "\u7ec4\u5c40\u52a8\u6001", "bucket": "group", "moreText": "\u67e5\u770b\u5168\u90e8"},
		{"key": "system", "title": "\u7cfb\u7edf\u901a\u77e5", "bucket": "system", "moreText": "\u67e5\u770b\u5168\u90e8"},
		{"key": "achievement", "title": "\u6210\u5c31\u89e3\u9501", "bucket": "achievement", "moreText": "\u67e5\u770b\u5168\u90e8"},
		{"key": "warning", "title": "\u9884\u8b66\u63d0\u9192", "bucket": "warning", "moreText": "\u67e5\u770b\u5168\u90e8"},
		{"key": "friend", "title": "\u597d\u53cb\u6d88\u606f", "bucket": "friend", "moreText": "\u67e5\u770b\u5168\u90e8"},
	},
	"actionTexts": map[string]interface{}{
		"accept":  "\u786e\u8ba4\u53c2\u52a0",
		"reject":  "\u5a49\u62d2",
		"process": "\u7acb\u5373\u5904\u7406",
		"review":  "\u7acb\u5373\u8bc4\u4ef7",
		"detail":  "\u67e5\u770b\u8be6\u60c5",
		"game":    "\u67e5\u770b\u7ec4\u5c40",
		"contact": "\u8054\u7cfb\u53d1\u8d77\u4eba",
	},
	"texts": map[string]interface{}{
		"loadFailedText":     "\u6d88\u606f\u4e2d\u5fc3\u52a0\u8f7d\u5931\u8d25",
		"entryMissingText":   "\u6682\u65e0\u53ef\u6253\u5f00\u7684\u6d88\u606f\u5165\u53e3",
		"openFailedText":     "\u6d88\u606f\u6253\u5f00\u5931\u8d25",
		"actionMissingText":  "\u64cd\u4f5c\u4fe1\u606f\u4e0d\u5b8c\u6574",
		"actionSuccessText":  "\u64cd\u4f5c\u6210\u529f",
		"actionHandledText":  "\u5df2\u6807\u8bb0\u5904\u7406",
		"actionFailedText":   "\u6d88\u606f\u64cd\u4f5c\u5931\u8d25",
		"serviceMissingText": "\u7f3a\u5c11\u6d88\u606f\u64cd\u4f5c\u4fe1\u606f",
		"serviceFailedText":  "\u6d88\u606f\u64cd\u4f5c\u5931\u8d25",
	},
}

const messageCenterConfigKey = "message.center_config"

var defaultMessageMyConfig = map[string]interface{}{
	"pageTitle":  "\u597d\u53cb\u6d88\u606f",
	"onlineText": "\u5728\u7ebf",
	"friend": map[string]interface{}{
		"defaultInitials": "\u0049\u004d",
		"defaultName":     "\u5c40\u5185\u4f1a\u8bdd",
		"defaultStatus":   "\u5728\u7ebf",
		"nameTemplate":    "\u6210\u5458 {userId}",
	},
	"quickActions": []map[string]interface{}{
		{"key": "friend", "label": "\u52a0\u597d\u53cb"},
		{"key": "greet", "label": "\u6253\u62db\u547c", "messageText": "\u4f60\u597d\uff0c\u6211\u770b\u5230\u4f60\u7684\u6d88\u606f\u4e86\u3002"},
		{"key": "card", "label": "\u53d1\u540d\u7247", "messageText": "\u8fd9\u662f\u6211\u7684\u540d\u7247\uff0c\u540e\u7eed\u53ef\u4ee5\u5728\u5c40\u5185\u7ee7\u7eed\u6c9f\u901a\u3002"},
		{"key": "location", "label": "\u53d1\u5b9a\u4f4d"},
	},
	"texts": map[string]interface{}{
		"loadFailedText":       "\u52a0\u8f7d\u4f1a\u8bdd\u5931\u8d25",
		"actionMissingText":    "\u64cd\u4f5c\u4fe1\u606f\u4e0d\u5b8c\u6574",
		"sendFailedText":       "\u53d1\u9001\u5931\u8d25",
		"recordStartText":      "\u5f00\u59cb\u5f55\u97f3",
		"recordStopText":       "\u5f53\u524d\u652f\u6301\u6587\u5b57\u3001\u56fe\u7247\u548c\u6587\u4ef6\u6d88\u606f",
		"recordErrorText":      "\u5f55\u97f3\u5931\u8d25",
		"fileEntryMissingText": "\u8bf7\u4ece\u5c40\u5185\u6d88\u606f\u5165\u53e3\u53d1\u9001\u6587\u4ef6",
		"justNowText":          "\u521a\u521a",
	},
}

const messageMyConfigKey = "message.my_config"

func (s *Server) myNotifications(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	allItems := s.notices.List(userID)
	activeBucket := r.URL.Query().Get("bucket")
	showAll := r.URL.Query().Get("detail") == "1" || r.URL.Query().Get("showAll") == "1"
	items := filterNotifications(allItems, r.URL.Query().Get("tab"), r.URL.Query().Get("type"), activeBucket)
	httpx.OK(w, s.notificationCenterPayload(userID, items, allItems, r.URL.Query().Get("tab"), activeBucket, showAll))
}

func (s *Server) markNotificationRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	notificationID, ok := notificationIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}
	notification, err := s.notices.MarkRead(userID, notificationID)
	if err != nil {
		writeNotificationError(w, err)
		return
	}
	httpx.OK(w, notification)
}

func defaultMessageMyPageConfig() map[string]interface{} {
	return cloneMap(defaultMessageMyConfig)
}

func (s *Server) currentMessageMyConfig() map[string]interface{} {
	var config map[string]interface{}
	if s.systemConfig != nil && s.systemConfig.Get(messageMyConfigKey, &config) && len(config) > 0 {
		return config
	}
	return defaultMessageMyPageConfig()
}

func (s *Server) messageMyConfig(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireUser(w, r); !ok {
		return
	}
	httpx.OK(w, s.currentMessageMyConfig())
}

func defaultMessageCenterPageConfig() map[string]interface{} {
	return cloneMap(defaultMessageCenterConfig)
}

func (s *Server) currentMessageCenterConfig() map[string]interface{} {
	var config map[string]interface{}
	if s.systemConfig != nil && s.systemConfig.Get(messageCenterConfigKey, &config) && len(config) > 0 {
		return config
	}
	return defaultMessageCenterPageConfig()
}

func (s *Server) notificationCenterPayload(userID int64, items []notifications.Notification, allItems []notifications.Notification, activeTab string, activeBucket string, showAll bool) map[string]interface{} {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	config := s.currentMessageCenterConfig()
	unread := 0
	trade := 0
	for _, item := range items {
		if item.Status == "unread" {
			unread++
		}
		if item.NotifyType == "trade_warning" || item.NotifyType == "progress_feedback_remind" {
			trade++
		}
	}
	if strings.TrimSpace(activeTab) == "" {
		activeTab = "all"
	}
	activeBucket = strings.TrimSpace(activeBucket)
	imCardTotal := len(s.messageCenterIMCards(userID, true))
	return map[string]interface{}{
		"items":        items,
		"pageTitle":    stringFromConfig(config, "pageTitle"),
		"onlineText":   stringFromConfig(config, "onlineText"),
		"activeTab":    activeTab,
		"activeBucket": activeBucket,
		"showAll":      showAll,
		"quickActions": messageCenterQuickActions(config, allItems, imCardTotal),
		"tabs":         messageCenterTabs(config, unread, trade),
		"sections":     s.messageCenterSections(userID, config, items, allItems, activeBucket, showAll),
		"texts":        stringMapFromConfig(config, "texts"),
	}
}

func stringFromConfig(config map[string]interface{}, key string) string {
	value, _ := config[key].(string)
	return value
}

func stringValue(value interface{}) string {
	text, _ := value.(string)
	return text
}

func stringMapFromConfig(config map[string]interface{}, key string) map[string]string {
	result := map[string]string{}
	raw, _ := config[key].(map[string]interface{})
	for itemKey, itemValue := range raw {
		if text, ok := itemValue.(string); ok {
			result[itemKey] = text
		}
	}
	return result
}

func sliceMapFromConfig(config map[string]interface{}, key string) []map[string]interface{} {
	rawItems, _ := config[key].([]interface{})
	result := make([]map[string]interface{}, 0, len(rawItems))
	for _, rawItem := range rawItems {
		if item, ok := rawItem.(map[string]interface{}); ok {
			result = append(result, item)
		}
	}
	return result
}

func messageCenterQuickActions(config map[string]interface{}, items []notifications.Notification, imCardTotal int) []map[string]interface{} {
	actions := sliceMapFromConfig(config, "quickActions")
	result := make([]map[string]interface{}, 0, len(actions))
	for _, action := range actions {
		item := cloneMap(action)
		bucket, _ := item["bucket"].(string)
		unreadCount, totalCount := countNotificationsByBucket(items, bucket)
		if bucket == "friend" {
			totalCount += imCardTotal
		}
		item["unreadCount"] = unreadCount
		item["totalCount"] = totalCount
		item["countText"] = strconv.Itoa(unreadCount) + "/" + strconv.Itoa(totalCount)
		result = append(result, item)
	}
	return result
}

func messageCenterTabs(config map[string]interface{}, unread int, trade int) []map[string]interface{} {
	tabs := sliceMapFromConfig(config, "tabs")
	result := make([]map[string]interface{}, 0, len(tabs))
	for _, tab := range tabs {
		item := cloneMap(tab)
		switch item["countSource"] {
		case "trade":
			item["unreadCount"] = trade
		case "unread":
			item["unreadCount"] = unread
		}
		delete(item, "countSource")
		result = append(result, item)
	}
	return result
}

func (s *Server) messageCenterSections(userID int64, config map[string]interface{}, items []notifications.Notification, allItems []notifications.Notification, activeBucket string, showAll bool) []map[string]interface{} {
	sectionConfigs := messageCenterSectionConfigs(config)
	result := make([]map[string]interface{}, 0, len(sectionConfigs))
	imCards := s.messageCenterIMCards(userID, showAll)
	imCardTotal := len(s.messageCenterIMCards(userID, true))
	for _, section := range sectionConfigs {
		item := cloneMap(section)
		bucket, _ := item["bucket"].(string)
		if bucket == "" {
			continue
		}
		if activeBucket != "" && bucket != activeBucket {
			continue
		}
		item["moreText"] = "\u67e5\u770b\u5168\u90e8"
		item["moreRoute"] = "/pages/message/index?bucket=" + url.QueryEscape(bucket) + "&detail=1"
		unreadCount, totalCount := countNotificationsByBucket(allItems, bucket)
		cards := s.notificationCards(items, bucket, showAll)
		if bucket == "friend" && len(imCards) > 0 {
			totalCount += imCardTotal
			cards = append(imCards, cards...)
			if !showAll && len(cards) > 3 {
				cards = cards[:3]
			}
		}
		item["unreadCount"] = unreadCount
		item["totalCount"] = totalCount
		item["countText"] = strconv.Itoa(unreadCount) + "/" + strconv.Itoa(totalCount)
		item["items"] = cards
		delete(item, "bucket")
		result = append(result, item)
	}
	return result
}

func (s *Server) messageCenterIMCards(userID int64, showAll bool) []map[string]interface{} {
	cards := make([]map[string]interface{}, 0, 2)
	for _, room := range s.im.AdminRooms() {
		if !s.games.IsMember(room.GameID, userID) {
			continue
		}
		title := ""
		if game, err := s.games.Get(room.GameID); err == nil {
			title = strings.TrimSpace(game.Title)
		}
		if title == "" {
			title = "局内群聊"
		}
		memberCount := len(room.MemberIDs)
		lastText := "进入群聊查看消息"
		if messages, _, err := s.im.AdminMessagesByRoom(room.ID); err == nil && len(messages) > 0 {
			for i := len(messages) - 1; i >= 0; i-- {
				if messages[i].Status == "hidden" {
					continue
				}
				if messages[i].Type == "image" {
					lastText = "最新消息：[图片]"
				} else if messages[i].Type == "file" || messages[i].FileID > 0 {
					lastText = "最新消息：[文件]"
				} else if summary := readableIMMessageSummary(messages[i].Content); summary != "" {
					lastText = "最新消息：" + summary
				}
				break
			}
		}
		cards = append(cards, map[string]interface{}{
			"id":          "im-room-" + strconv.FormatInt(room.GameID, 10),
			"routeKey":    "im_room",
			"route":       "/pages/im/room/index?gameId=" + strconv.FormatInt(room.GameID, 10),
			"detailRoute": "/pages/im/room/index?gameId=" + strconv.FormatInt(room.GameID, 10),
			"tone":        "green",
			"icon":        "IM",
			"title":       title + "IM",
			"timeText":    room.CreatedAt.Format("01-02 15:04"),
			"desc":        lastText,
			"tagText":     "成员" + strconv.Itoa(memberCount) + "人",
			"tagTone":     "green",
			"metaText":    imRoomStatusText(room.Status),
		})
		if !showAll && len(cards) >= 1 {
			break
		}
	}
	return cards
}

func readableIMMessageSummary(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	var payload map[string]interface{}
	if json.Unmarshal([]byte(content), &payload) == nil {
		for _, key := range []string{"content", "message", "memberDesc", "description", "desc", "text", "subtitle", "title"} {
			if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
		return "系统消息"
	}
	return content
}

func imRoomStatusText(status string) string {
	switch strings.TrimSpace(status) {
	case "active":
		return "群聊中"
	case "readonly":
		return "只读"
	case "archived":
		return "已归档"
	default:
		return "局内群聊"
	}
}
func messageCenterSectionConfigs(config map[string]interface{}) []map[string]interface{} {
	sections := sliceMapFromConfig(config, "sections")
	seen := map[string]bool{}
	for _, section := range sections {
		if key := strings.TrimSpace(stringValue(section["key"])); key != "" {
			seen[key] = true
		}
	}
	for _, section := range sliceMapFromConfig(defaultMessageCenterPageConfig(), "sections") {
		key := strings.TrimSpace(stringValue(section["key"]))
		if key != "" && !seen[key] {
			sections = append(sections, section)
			seen[key] = true
		}
	}
	return sections
}

func filterNotifications(items []notifications.Notification, tab string, notifyType string, bucket string) []notifications.Notification {
	tab = strings.TrimSpace(tab)
	notifyType = strings.TrimSpace(notifyType)
	bucket = strings.TrimSpace(bucket)
	result := make([]notifications.Notification, 0, len(items))
	for _, item := range items {
		if notifyType != "" && item.NotifyType != notifyType {
			continue
		}
		if bucket != "" && notificationBucket(item) != bucket {
			continue
		}
		if tab == "unread" && item.Status != "unread" {
			continue
		}
		if tab == "trade" && item.NotifyType != "trade_warning" && item.NotifyType != "progress_feedback_remind" {
			continue
		}
		result = append(result, item)
	}
	return result
}

func (s *Server) notificationCards(items []notifications.Notification, bucket string, showAll bool) []map[string]interface{} {
	cards := make([]map[string]interface{}, 0)
	actionTexts := stringMapFromConfig(s.currentMessageCenterConfig(), "actionTexts")
	for _, item := range items {
		if notificationBucket(item) != bucket {
			continue
		}
		if !showAll && len(cards) >= 3 {
			break
		}
		card := map[string]interface{}{
			"id":       item.ID,
			"routeKey": notificationRouteKey(item),
			"tone":     notificationTone(item),
			"unread":   item.Status == "unread",
			"title":    item.Title,
			"timeText": item.CreatedAt.Format("01-02 15:04"),
			"desc":     item.Content,
			"bizType":  item.BizType,
			"bizId":    item.BizID,
		}
		if target := s.notificationDetailTarget(item); target != nil {
			if route, ok := target["route"].(string); ok && strings.TrimSpace(route) != "" {
				card["route"] = route
				card["detailRoute"] = route
			}
			if routeKey, ok := target["routeKey"].(string); ok && strings.TrimSpace(routeKey) != "" {
				card["routeKey"] = routeKey
			}
		}
		if actions := s.notificationActions(item, actionTexts); len(actions) > 0 {
			for _, action := range actions {
				actionKey, _ := action["key"].(string)
				target, err := s.resolveNotificationAction(item, actionKey)
				if err != nil || target == nil {
					continue
				}
				if route, ok := target["route"].(string); ok && strings.TrimSpace(route) != "" {
					action["route"] = route
					action["detailRoute"] = route
				}
				if routeKey, ok := target["routeKey"].(string); ok && strings.TrimSpace(routeKey) != "" {
					action["routeKey"] = routeKey
				}
			}
			card["actions"] = actions
		}
		if item.BizType == "game_invitation" {
			card = s.gameInvitationNotificationCard(item, card)
		}
		cards = append(cards, card)
	}
	return cards
}

func (s *Server) gameInvitationNotificationCard(item notifications.Notification, card map[string]interface{}) map[string]interface{} {
	invitation, ok := s.notificationInvitation(item)
	if !ok {
		return card
	}
	detail := s.invitationProgressItem(invitation, item.UserID)
	gameInfo, _ := detail["gameInfo"].(map[string]interface{})
	player, _ := detail["player"].(map[string]interface{})
	expert, _ := detail["expert"].(map[string]interface{})
	inviter, _ := detail["inviter"].(map[string]interface{})
	statusText := stringFromMap(detail, "statusText")
	viewerRole := invitationViewerRole(invitation, item.UserID)
	viewerRoleText := gameInvitationViewerRoleText(viewerRole)
	route, routeKey := s.gameInvitationConfirmRoute(invitation, item.UserID)
	card["cardType"] = "game_invitation"
	card["cardClass"] = "game-invitation"
	card["routeKey"] = routeKey
	card["route"] = route
	card["detailRoute"] = route
	if item.NotifyType == "game_invitation_success" {
		if successRoute, successRouteKey := gameInvitationSuccessRoute(invitation.GameID, viewerRole); successRoute != "" {
			card["routeKey"] = successRouteKey
			card["route"] = successRoute
			card["detailRoute"] = successRoute
		}
	}
	card["tone"] = "orange"
	card["icon"] = "\u5c40"
	card["title"] = gameInvitationNotificationTitle(invitation.Status, viewerRoleText)
	descMessage := invitation.Message
	if item.NotifyType == "game_invitation_success" {
		descMessage = item.Content
	}
	card["desc"] = gameInvitationNotificationDesc(descMessage, viewerRoleText)
	card["tagText"] = gameInvitationNotificationTag(invitation.Status, statusText, viewerRoleText)
	card["tagTone"] = gameInvitationNotificationTagTone(invitation.Status)
	if item.NotifyType == "game_invitation_success" {
		card["title"] = "\u7ec4\u5c40\u6210\u529f"
		card["tagText"] = "\u7ec4\u5c40\u6210\u529f"
		card["tagTone"] = "success"
	}
	card["summaryText"] = stringFromMap(gameInfo, "title")
	card["subText"] = stringFromMap(gameInfo, "locationText")
	card["metaText"] = gameInvitationNotificationMeta(inviter)
	card["infoRows"] = gameInvitationNotificationInfoRows(gameInfo, inviter)
	card["participants"] = gameInvitationNotificationParticipants(player, expert)
	if item.NotifyType == "game_invitation_success" || invitation.Status != "pending" {
		card["actions"] = []map[string]interface{}{
			{
				"key":     "detail",
				"text":    gameInvitationNotificationActionText(invitation.Status, item.NotifyType),
				"primary": true,
				"orange":  true,
			},
		}
	}
	return card
}

func gameInvitationViewerRoleText(role string) string {
	switch role {
	case "expert":
		return "\u884c\u5bb6"
	case "player":
		return "\u73a9\u5bb6"
	case "guide":
		return "\u9886\u8def\u4eba"
	default:
		return "\u7528\u6237"
	}
}

func gameInvitationNotificationTitle(status string, viewerRoleText string) string {
	switch status {
	case "accepted":
		return "\u7ec4\u5c40\u786e\u8ba4\u5df2\u5b8c\u6210"
	case "rejected":
		return "\u7ec4\u5c40\u786e\u8ba4\u5df2\u53d6\u6d88"
	default:
		return viewerRoleText + "\u7ec4\u5c40\u786e\u8ba4\u901a\u77e5"
	}
}

func gameInvitationNotificationDesc(message string, viewerRoleText string) string {
	message = strings.TrimSpace(message)
	if message != "" {
		return message
	}
	return "\u8bf7\u67e5\u770b\u7ec4\u5c40\u4fe1\u606f\u5e76\u5b8c\u6210" + viewerRoleText + "\u786e\u8ba4"
}

func gameInvitationNotificationTag(status string, statusText string, viewerRoleText string) string {
	statusText = strings.TrimSpace(statusText)
	switch status {
	case "accepted", "rejected":
		return statusText
	default:
		if viewerRoleText != "" && viewerRoleText != "\u7528\u6237" {
			return viewerRoleText + "\u5f85\u786e\u8ba4"
		}
		if statusText != "" {
			return statusText
		}
		return "\u5f85\u786e\u8ba4"
	}
}

func gameInvitationNotificationTagTone(status string) string {
	switch status {
	case "accepted":
		return "success"
	case "rejected":
		return "muted"
	default:
		return "warning"
	}
}

func gameInvitationNotificationActionText(status string, notifyType string) string {
	if notifyType == "game_invitation_success" {
		return "\u67e5\u770b\u6210\u529f\u9875"
	}
	switch status {
	case "accepted", "rejected":
		return "\u67e5\u770b\u8be6\u60c5"
	default:
		return "\u67e5\u770b\u5e76\u786e\u8ba4"
	}
}

func gameInvitationNotificationMeta(inviter map[string]interface{}) string {
	name := stringFromMap(inviter, "name")
	if name == "" {
		return ""
	}
	return "\u9886\u8def\u4eba\uff1a" + name
}

func gameInvitationNotificationInfoRows(gameInfo map[string]interface{}, inviter map[string]interface{}) []map[string]string {
	rows := []map[string]string{}
	add := func(label string, value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		rows = append(rows, map[string]string{"label": label, "value": value})
	}
	add("\u7ec4\u5c40\u4e3b\u9898", stringFromMap(gameInfo, "title"))
	add("\u65f6\u95f4", stringFromMap(gameInfo, "timeText"))
	add("\u5730\u70b9", stringFromMap(gameInfo, "locationText"))
	add("\u9886\u8def\u4eba", stringFromMap(inviter, "name"))
	return rows
}

func gameInvitationNotificationParticipants(player map[string]interface{}, expert map[string]interface{}) []map[string]interface{} {
	participants := make([]map[string]interface{}, 0, 2)
	add := func(member map[string]interface{}) {
		name := stringFromMap(member, "name")
		if name == "" {
			return
		}
		participants = append(participants, map[string]interface{}{
			"name":        name,
			"roleLabel":   stringFromMap(member, "roleLabel"),
			"state":       stringFromMap(member, "state"),
			"stateClass":  stringFromMap(member, "stateClass"),
			"avatarText":  stringFromMap(member, "avatarText"),
			"avatarSrc":   stringFromMap(member, "avatarSrc"),
			"avatarClass": stringFromMap(member, "avatarClass"),
		})
	}
	add(player)
	add(expert)
	return participants
}

func notificationBucket(item notifications.Notification) string {
	switch item.NotifyType {
	case "game_application", "game_apply", "game_approved", "application_approved", "application_rejected", "game_invitation", "game_invitation_progress", "game_invitation_success", "game_ready_to_start", "game_started", "game_member_quit", "game_quit_result", "service_confirm_remind", "review_remind", "game_ended", "game_completion_requested", "expert_completion_confirmed", "player_cancel_request", "expert_cancel_request", "guide_cancel_request":
		return "group"
	case "achievement", "achievement_unlocked", "growth_achievement":
		return "achievement"
	case "friend", "friend_message", "im_message", "private_message":
		return "friend"
	case "trade_warning", "progress_feedback_remind":
		return "warning"
	default:
		return "system"
	}
}

func notificationTone(item notifications.Notification) string {
	switch notificationBucket(item) {
	case "group":
		return "orange"
	case "warning":
		return "red"
	default:
		return "blue"
	}
}

func notificationRouteKey(item notifications.Notification) string {
	if notificationBucket(item) == "warning" {
		return "warning"
	}
	return "system"
}

func (s *Server) notificationActions(item notifications.Notification, texts map[string]string) []map[string]interface{} {
	text := func(key string) string {
		return texts[key]
	}
	if item.BizType == "game_invitation" && item.Status == "unread" {
		return []map[string]interface{}{
			{"key": "detail", "text": text("detail"), "primary": true},
		}
	}
	switch item.NotifyType {
	case "player_cancel_request", "expert_cancel_request", "guide_cancel_request":
		return []map[string]interface{}{
			{"key": "detail", "text": text("detail"), "primary": true},
		}
	case "game_ready_to_start":
		return []map[string]interface{}{
			{"key": "detail", "text": "去开局", "primary": true},
		}
	case "review_remind", "game_ended":
		return []map[string]interface{}{
			{"key": "process", "text": text("review"), "primary": true},
			{"key": "detail", "text": text("game")},
		}
	case "service_confirm_remind":
		return []map[string]interface{}{
			{"key": "process", "text": text("process"), "primary": true},
			{"key": "contact", "text": text("contact")},
		}
	case "game_completion_requested", "expert_completion_confirmed":
		return []map[string]interface{}{
			{"key": "process", "text": text("process"), "primary": true},
			{"key": "detail", "text": text("game")},
		}
	case "progress_feedback_remind", "trade_warning":
		return []map[string]interface{}{
			{"key": "process", "text": text("process"), "primary": true},
			{"key": "contact", "text": text("contact")},
		}
	case "game_member_quit", "game_quit_result":
		return []map[string]interface{}{
			{"key": "detail", "text": text("game"), "primary": true},
			{"key": "contact", "text": text("contact")},
		}
	case "report_created", "report_handled", "report_assigned", "report_closed":
		return []map[string]interface{}{
			{"key": "detail", "text": text("detail"), "primary": true},
		}
	}
	return nil
}

func countNotificationsByBucket(items []notifications.Notification, bucket string) (int, int) {
	unreadCount := 0
	totalCount := 0
	for _, item := range items {
		if notificationBucket(item) != bucket {
			continue
		}
		totalCount++
		if item.Status == "unread" {
			unreadCount++
		}
	}
	return unreadCount, totalCount
}

func cloneMap(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return map[string]interface{}{}
	}
	raw, _ := json.Marshal(src)
	var dst map[string]interface{}
	_ = json.Unmarshal(raw, &dst)
	if dst == nil {
		dst = map[string]interface{}{}
	}
	return dst
}

func defaultTradeWarningConfig() map[string]interface{} {
	return cloneMap(defaultTradeWarningDetail)
}

func (s *Server) currentTradeWarningConfig() map[string]interface{} {
	var config map[string]interface{}
	if s.systemConfig != nil && s.systemConfig.Get(tradeWarningConfigKey, &config) && len(config) > 0 {
		return config
	}
	return defaultTradeWarningConfig()
}

func defaultSystemNotificationConfig() map[string]interface{} {
	return cloneMap(defaultSystemNotificationDetail)
}

func (s *Server) currentSystemNotificationConfig() map[string]interface{} {
	var config map[string]interface{}
	if s.systemConfig != nil && s.systemConfig.Get(systemNotificationConfigKey, &config) && len(config) > 0 {
		return config
	}
	return defaultSystemNotificationConfig()
}

func (s *Server) tradeWarningDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	warningID := strings.TrimSpace(r.URL.Query().Get("warningId"))
	if warningID == "" {
		warningID = strings.TrimSpace(r.URL.Query().Get("id"))
	}
	orderID := strings.TrimSpace(r.URL.Query().Get("orderId"))
	gameID := strings.TrimSpace(r.URL.Query().Get("gameId"))
	notice, hasNotice := s.findNotificationForUser(userID, warningID)
	if gameID == "" {
		gameID = gameIDFromNotification(notice)
	}
	detail := cloneMap(s.currentTradeWarningConfig())
	detail["warningId"] = warningID
	detail["id"] = warningID
	detail["gameId"] = gameID
	if warning, ok := detail["warning"].(map[string]interface{}); ok {
		title, content := notificationTitleContent(notice, hasNotice)
		warning["title"] = defaultString(title, "\u4ea4\u6613\u9884\u8b66")
		warning["prefixText"] = ""
		warning["highlightText"] = content
		warning["suffixText"] = ""
	}
	if order, ok := detail["order"].(map[string]interface{}); ok {
		order["orderNo"] = orderID
		order["statusText"] = notificationStatusText(notice, hasNotice)
		if hasNotice {
			order["customerTitle"] = notice.Title
			order["customerDesc"] = notice.Content
			order["detailRows"] = []map[string]interface{}{
				{"label": "\u4e1a\u52a1\u7c7b\u578b", "value": notice.BizType},
				{"label": "\u4e1a\u52a1 ID", "value": strconv.FormatInt(notice.BizID, 10)},
			}
		}
	}
	httpx.OK(w, detail)
}

func (s *Server) handleTradeWarningAction(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req struct {
		Action         string `json:"action"`
		WarningID      string `json:"warningId"`
		OrderID        string `json:"orderId"`
		GameID         string `json:"gameId"`
		DeliveryMethod string `json:"deliveryMethod"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if action != "delay" && action != "deliver" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid trade warning action")
		return
	}
	warningID := strings.TrimSpace(req.WarningID)
	orderID := strings.TrimSpace(req.OrderID)
	gameID := strings.TrimSpace(req.GameID)
	if gameID == "" {
		gameID = s.tradeWarningGameID(userID, warningID)
	}
	s.recordBehavior(userID, "trade_warning_action", "trade_warning", 0, map[string]interface{}{
		"action":         action,
		"warningId":      warningID,
		"orderId":        orderID,
		"gameId":         gameID,
		"deliveryMethod": strings.TrimSpace(req.DeliveryMethod),
	})
	if action == "deliver" {
		httpx.OK(w, map[string]interface{}{
			"action":  action,
			"handled": true,
			"message": "\u5df2\u8fdb\u5165\u4ea4\u4ed8\u786e\u8ba4",
			"target": map[string]interface{}{
				"route":    "pages/game/delivery/index?gameId=" + url.QueryEscape(gameID) + "&warningId=" + url.QueryEscape(warningID) + "&orderId=" + url.QueryEscape(orderID),
				"routeKey": "gameDelivery",
			},
		})
		return
	}
	httpx.OK(w, map[string]interface{}{
		"action":     action,
		"handled":    true,
		"message":    "\u5ef6\u671f\u7533\u8bf7\u5df2\u63d0\u4ea4",
		"statusText": "\u5df2\u7533\u8bf7\u5ef6\u671f",
		"warningId":  warningID,
		"orderId":    orderID,
	})
}

func (s *Server) tradeWarningGameID(userID int64, warningID string) string {
	notice, ok := s.findNotificationForUser(userID, warningID)
	if !ok {
		return ""
	}
	return gameIDFromNotification(notice)
}

func (s *Server) findNotificationForUser(userID int64, notificationID string) (notifications.Notification, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(notificationID), 10, 64)
	if err != nil || id <= 0 {
		return notifications.Notification{}, false
	}
	for _, item := range s.notices.List(userID) {
		if item.ID == id {
			return item, true
		}
	}
	return notifications.Notification{}, false
}

func gameIDFromNotification(notification notifications.Notification) string {
	if notification.BizID <= 0 {
		return ""
	}
	return strconv.FormatInt(notification.BizID, 10)
}

func notificationTitleContent(notification notifications.Notification, ok bool) (string, string) {
	if !ok {
		return "", ""
	}
	return strings.TrimSpace(notification.Title), strings.TrimSpace(notification.Content)
}

func notificationStatusText(notification notifications.Notification, ok bool) string {
	if !ok {
		return ""
	}
	if notification.Status == "read" {
		return "\u5df2\u8bfb"
	}
	return "\u672a\u8bfb"
}

func (s *Server) systemNotificationDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	messageID := strings.TrimSpace(r.URL.Query().Get("messageId"))
	if messageID == "" {
		messageID = strings.TrimSpace(r.URL.Query().Get("notificationId"))
	}
	if messageID == "" {
		messageID = strings.TrimSpace(r.URL.Query().Get("id"))
	}
	detail := cloneMap(s.currentSystemNotificationConfig())
	detail["feedback"] = s.systemNotificationFeedback(userID)
	if messageID != "" {
		detail["messageId"] = messageID
		detail["notificationId"] = messageID
		detail["id"] = messageID
	}
	if notice, ok := s.findNotificationForUser(userID, messageID); ok {
		if article, ok := detail["article"].(map[string]interface{}); ok {
			article["tagText"] = notificationTypeLabel(notice.NotifyType)
			article["title"] = notice.Title
			article["author"] = "\u7cfb\u7edf\u901a\u77e5"
			article["publishedAtText"] = notice.CreatedAt.Format("2006-01-02")
			article["readText"] = ""
			article["blocks"] = []map[string]interface{}{
				{"id": "content", "type": "paragraph", "text": notice.Content, "lead": true},
			}
		}
		detail["bizType"] = notice.BizType
		detail["bizId"] = notice.BizID
	}
	httpx.OK(w, detail)
}

func notificationTypeLabel(notifyType string) string {
	switch notifyType {
	case "game_approved", "game_apply", "application_approved", "application_rejected":
		return "\u7ec4\u5c40\u52a8\u6001"
	case "trade_warning", "progress_feedback_remind":
		return "\u9884\u8b66\u63d0\u9192"
	case "review_remind":
		return "\u8bc4\u4ef7\u63d0\u9192"
	case "service_confirm_remind":
		return "\u4ea4\u4ed8\u786e\u8ba4"
	case "report_created", "report_handled", "report_assigned", "report_closed":
		return "\u7533\u8bc9\u901a\u77e5"
	default:
		return "\u7cfb\u7edf\u901a\u77e5"
	}
}
func (s *Server) submitSystemNotificationFeedback(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req struct {
		MessageID string `json:"messageId"`
		Value     string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	value := strings.ToLower(strings.TrimSpace(req.Value))
	if value != "useful" && value != "useless" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid feedback")
		return
	}

	config := s.profiles.SystemManagementConfig(userID, "system-notification-feedback", map[string]interface{}{"useful": 128, "useless": 10})
	config[value] = interfaceToInt(config[value]) + 1
	s.profiles.SaveSystemManagementConfig(userID, "system-notification-feedback", config)
	feedback := systemNotificationFeedbackPayload(config)
	s.recordBehavior(userID, "system_notification_feedback", "notification", 0, map[string]interface{}{"messageId": req.MessageID, "value": value})
	httpx.OK(w, map[string]interface{}{
		"value":    value,
		"feedback": feedback,
	})
}

func (s *Server) systemNotificationFeedback(userID int64) map[string]interface{} {
	config := s.profiles.SystemManagementConfig(userID, "system-notification-feedback", map[string]interface{}{"useful": 128, "useless": 10})
	return systemNotificationFeedbackPayload(config)
}

func systemNotificationFeedbackPayload(config map[string]interface{}) map[string]interface{} {
	useful := interfaceToInt(config["useful"])
	useless := interfaceToInt(config["useless"])
	if useful <= 0 {
		useful = 128
	}
	if useless <= 0 {
		useless = 10
	}
	return map[string]interface{}{
		"question": "\u8fd9\u7bc7\u6587\u7ae0\u5bf9\u4f60\u6709\u5e2e\u52a9\u5417\uff1f",
		"useful":   map[string]interface{}{"icon": "👍", "label": "\u6709\u7528", "count": useful, "countText": strconv.Itoa(useful)},
		"useless":  map[string]interface{}{"icon": "👎", "label": "\u6ca1\u7528", "count": useless, "countText": strconv.Itoa(useless)},
	}
}

func interfaceToInt(value interface{}) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		number, _ := typed.Int64()
		return int(number)
	case string:
		number, _ := strconv.Atoi(strings.TrimSpace(typed))
		return number
	default:
		return 0
	}
}

func (s *Server) handleNotificationAction(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	notificationID, ok := notificationIDFromActionPath(w, r.URL.Path)
	if !ok {
		return
	}
	var req struct {
		Action string `json:"action"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	notification, err := s.notices.MarkRead(userID, notificationID)
	if err != nil {
		writeNotificationError(w, err)
		return
	}
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if notification.BizType == "game_invitation" {
		s.handleGameInvitationNotificationAction(w, r, notification, action, req.Reason)
		return
	}
	result, err := s.resolveNotificationAction(notification, action)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
		return
	}
	s.recordBehavior(userID, "notification_action", "notification", notification.ID, map[string]interface{}{"action": action, "bizType": notification.BizType, "bizId": notification.BizID})
	httpx.OK(w, map[string]interface{}{
		"notification": notification,
		"action":       action,
		"handled":      true,
		"target":       result,
	})
}

func (s *Server) handleGameInvitationNotificationAction(w http.ResponseWriter, r *http.Request, notification notifications.Notification, action string, reason string) {
	userID, _ := appUserIDFromRequest(r)
	if action == "" || action == "detail" || action == "view" || action == "open" {
		result, err := s.resolveNotificationAction(notification, action)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		s.recordBehavior(userID, "notification_action", "notification", notification.ID, map[string]interface{}{"action": action, "bizType": notification.BizType, "bizId": notification.BizID})
		httpx.OK(w, map[string]interface{}{
			"notification": notification,
			"action":       action,
			"handled":      true,
			"target":       result,
		})
		return
	}
	httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "please open invitation detail to respond")
}

func (s *Server) adminWechatSubscribeTasks(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"items": s.notices.WechatTasks()})
}

func (s *Server) adminWechatSubscribeTemplates(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"items": s.notices.WechatTemplates()})
}

func (s *Server) markWechatSubscribeTaskSent(w http.ResponseWriter, r *http.Request) {
	taskID, ok := wechatTaskIDFromPath(w, r.URL.Path, "/mark-sent")
	if !ok {
		return
	}
	var req struct {
		ResultCode    string `json:"resultCode"`
		ResultMessage string `json:"resultMessage"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	task, err := s.notices.MarkWechatTaskSent(taskID, req.ResultCode, req.ResultMessage)
	if err != nil {
		writeNotificationError(w, err)
		return
	}
	httpx.OK(w, task)
}

func (s *Server) sendWechatSubscribeTask(w http.ResponseWriter, r *http.Request) {
	taskID, ok := wechatTaskIDFromPath(w, r.URL.Path, "/send")
	if !ok {
		return
	}
	task, err := s.notices.SendWechatTask(taskID)
	if err != nil {
		writeNotificationError(w, err)
		return
	}
	httpx.OK(w, task)
}

func (s *Server) sendPendingWechatSubscribeTasks(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Limit int `json:"limit"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	httpx.OK(w, s.notices.SendPendingWechatTasks(req.Limit))
}

func notificationIDFromPath(w http.ResponseWriter, path string) (int64, bool) {
	text := strings.TrimSuffix(strings.TrimPrefix(path, "/api/app/notifications/"), "/read")
	id, err := strconv.ParseInt(strings.Trim(text, "/"), 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "通知 ID 错误")
		return 0, false
	}
	return id, true
}

func notificationIDFromActionPath(w http.ResponseWriter, path string) (int64, bool) {
	text := strings.TrimSuffix(strings.TrimPrefix(path, "/api/app/notifications/"), "/actions")
	id, err := strconv.ParseInt(strings.Trim(text, "/"), 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "notification id invalid")
		return 0, false
	}
	return id, true
}

func notificationTarget(notification notifications.Notification) map[string]interface{} {
	return map[string]interface{}{
		"bizType": notification.BizType,
		"bizId":   notification.BizID,
	}
}

func (s *Server) notificationInvitation(notification notifications.Notification) (games.Invitation, bool) {
	if notification.BizType != "game_invitation" || notification.BizID <= 0 {
		return games.Invitation{}, false
	}
	for _, invitation := range s.games.InvitationsForUser(notification.UserID) {
		if invitation.ID == notification.BizID {
			return invitation, true
		}
	}
	return games.Invitation{}, false
}

func (s *Server) resolveNotificationAction(notification notifications.Notification, action string) (map[string]interface{}, error) {
	if action == "" {
		action = "detail"
	}
	switch action {
	case "detail", "view", "open":
		return s.notificationDetailTarget(notification), nil
	case "process", "handle", "deliver":
		return s.notificationProcessTarget(notification), nil
	case "route", "navigation":
		return s.notificationRouteTarget(notification), nil
	case "contact", "chat":
		return notificationContactTarget(notification), nil
	case "delay":
		target := s.notificationDetailTarget(notification)
		target["defer"] = true
		return target, nil
	default:
		return nil, errors.New("unsupported action")
	}
}

func (s *Server) notificationDetailTarget(notification notifications.Notification) map[string]interface{} {
	target := notificationTarget(notification)
	if isServiceCancelNotification(notification.NotifyType) && notification.BizType == "game" {
		target["route"] = "pages/game/guide-cancel/index?notificationId=" + strconv.FormatInt(notification.ID, 10) + "&gameId=" + strconv.FormatInt(notification.BizID, 10)
		target["routeKey"] = "gameCancelDetail"
		return target
	}
	if (notification.NotifyType == "game_started" || notification.NotifyType == "im_message") && notification.BizType == "game" && notification.BizID > 0 {
		target["route"] = "pages/im/room/index?gameId=" + strconv.FormatInt(notification.BizID, 10)
		target["routeKey"] = "imRoom"
		return target
	}
	if notification.NotifyType == "private_message" && notification.BizType == "private_chat" && notification.BizID > 0 {
		target["route"] = "pages/message/my/index?mode=private&targetUserId=" + strconv.FormatInt(notification.BizID, 10)
		target["routeKey"] = "messageMy"
		return target
	}
	if notificationBucket(notification) == "group" && notification.BizType == "game" {
		route, routeKey := s.notificationGameGroupRoute(notification)
		target["route"] = route
		target["routeKey"] = routeKey
		return target
	}
	switch notification.BizType {
	case "report":
		target["route"] = "pages/profile/system-management/report-record-detail/index?id=" + strconv.FormatInt(notification.BizID, 10)
		target["routeKey"] = "reportDetail"
	case "game":
		target["route"] = "pages/game/detail/index?id=" + strconv.FormatInt(notification.BizID, 10)
		target["routeKey"] = "gameDetail"
	case "game_application":
		target["route"] = s.gameApplicationAuditListRoute(notification)
		target["routeKey"] = "gameAudit"
	case "game_invitation":
		if notification.NotifyType == "game_invitation_success" {
			if invitation, ok := s.notificationInvitation(notification); ok {
				if route, routeKey := gameInvitationSuccessRoute(invitation.GameID, invitationViewerRole(invitation, notification.UserID)); route != "" {
					target["route"] = route
					target["routeKey"] = routeKey
					return target
				}
			}
		}
		if invitation, ok := s.notificationInvitation(notification); ok {
			route, routeKey := s.gameInvitationConfirmRoute(invitation, notification.UserID)
			target["route"] = route
			target["routeKey"] = routeKey
			return target
		}
		target["route"] = "pages/game/audit-detail/index?invitationId=" + strconv.FormatInt(notification.BizID, 10)
		target["routeKey"] = "gameAuditDetail"
	default:
		target["route"] = "pages/message/system-detail/index?notificationId=" + strconv.FormatInt(notification.ID, 10)
		target["routeKey"] = "systemDetail"
	}
	return target
}

func (s *Server) notificationGameGroupRoute(notification notifications.Notification) (string, string) {
	if notification.NotifyType == "game_approved" {
		return "pages/game/detail/index?id=" + strconv.FormatInt(notification.BizID, 10), "gameDetail"
	}
	if notification.NotifyType == "game_completion_requested" || notification.NotifyType == "expert_completion_confirmed" || notification.NotifyType == "review_remind" || notification.NotifyType == "game_ended" {
		return "pages/game/detail/index?id=" + strconv.FormatInt(notification.BizID, 10), "gameDetail"
	}
	if notification.NotifyType == "game_invitation_progress" {
		return "pages/game/detail/index?id=" + strconv.FormatInt(notification.BizID, 10), "gameDetail"
	}
	if notification.NotifyType == "game_ready_to_start" {
		if notification.BizID <= 0 {
			return "pages/game/detail/index", "gameDetail"
		}
		return "pages/game/detail/index?id=" + strconv.FormatInt(notification.BizID, 10), "gameDetail"
	}
	if notification.NotifyType == "service_confirm_remind" {
		if notification.BizID <= 0 {
			return "pages/game/delivery/index", "gameDelivery"
		}
		return s.deliveryNotificationRoute(notification.BizID), "gameDelivery"
	}

	const routeKey = "gameGuideProgressDetail"
	const routePath = "pages/game/guide-progress-detail/index"

	if notification.BizID <= 0 {
		return routePath, routeKey
	}
	gameIDText := strconv.FormatInt(notification.BizID, 10)
	if s.games != nil {
		for _, invitation := range s.games.InvitationsForUser(notification.UserID) {
			if invitation.GameID == notification.BizID && invitation.ID > 0 {
				return routePath + "?invitationId=" + strconv.FormatInt(invitation.ID, 10) + "&gameId=" + gameIDText, routeKey
			}
		}
	}
	return routePath + "?gameId=" + gameIDText, routeKey
}

func (s *Server) deliveryNotificationRoute(gameID int64) string {
	route := "pages/game/delivery/index?gameId=" + strconv.FormatInt(gameID, 10)
	if gameID <= 0 || s.games == nil {
		return route
	}
	game, err := s.games.Get(gameID)
	if err != nil {
		return route
	}
	mode := "paid"
	if successFundAmount(game) <= 0 {
		mode = "free"
	}
	return route + "&mode=" + mode
}

func (s *Server) notificationProcessTarget(notification notifications.Notification) map[string]interface{} {
	target := notificationTarget(notification)
	switch notification.NotifyType {
	case "game_application":
		target["route"] = s.gameApplicationAuditListRoute(notification)
		target["routeKey"] = "gameAudit"
	case "review_remind", "game_ended":
		target["route"] = "pages/game/review/index?gameId=" + strconv.FormatInt(notification.BizID, 10)
		target["routeKey"] = "gameReview"
	case "service_confirm_remind", "progress_feedback_remind", "trade_warning", "game_completion_requested", "expert_completion_confirmed":
		target["route"] = s.deliveryNotificationRoute(notification.BizID)
		target["routeKey"] = "gameDelivery"
	case "report_created", "report_handled", "report_assigned", "report_closed":
		target["route"] = "pages/profile/system-management/report-record-detail/index?id=" + strconv.FormatInt(notification.BizID, 10)
		target["routeKey"] = "reportDetail"
	default:
		return s.notificationDetailTarget(notification)
	}
	return target
}

func (s *Server) notificationRouteTarget(notification notifications.Notification) map[string]interface{} {
	if notification.BizType == "game_application" || notification.NotifyType == "game_application" {
		target := notificationTarget(notification)
		target["route"] = s.gameApplicationAuditListRoute(notification)
		target["routeKey"] = "gameAudit"
		return target
	}
	target := notificationTarget(notification)
	target["route"] = "pages/map/index?gameId=" + strconv.FormatInt(notification.BizID, 10) + "&mode=route"
	target["routeKey"] = "mapRoute"
	return target
}

func (s *Server) gameApplicationAuditListRoute(notification notifications.Notification) string {
	gameID := int64(0)
	for _, item := range s.games.ApplicationsForCreator(notification.UserID) {
		if item.ID == notification.BizID {
			gameID = item.GameID
			break
		}
	}
	if gameID > 0 {
		return "pages/game/audit/index?gameId=" + strconv.FormatInt(gameID, 10)
	}
	return "pages/game/audit/index"
}

func notificationContactTarget(notification notifications.Notification) map[string]interface{} {
	target := notificationTarget(notification)
	if notification.BizType == "game" && notification.BizID > 0 {
		target["route"] = "pages/im/room/index?gameId=" + strconv.FormatInt(notification.BizID, 10)
		target["routeKey"] = "imRoom"
		return target
	}
	target["route"] = "pages/message/my/index"
	target["routeKey"] = "messageMy"
	return target
}

func wechatTaskIDFromPath(w http.ResponseWriter, path string, suffix string) (int64, bool) {
	text := strings.TrimSuffix(strings.TrimPrefix(path, "/api/internal/notifications/wechat-tasks/"), suffix)
	id, err := strconv.ParseInt(strings.Trim(text, "/"), 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "微信订阅消息任务 ID 错误")
		return 0, false
	}
	return id, true
}

func writeNotificationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, notifications.ErrNotificationNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "通知不存在")
	case errors.Is(err, notifications.ErrForbidden):
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "无权读取该通知")
	case errors.Is(err, notifications.ErrWechatSubscribeSend):
		httpx.Error(w, http.StatusBadGateway, httpx.CodeSystemError, "wechat subscribe message send failed")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "通知操作失败")
	}
}
