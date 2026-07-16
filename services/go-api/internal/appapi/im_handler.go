package appapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/files"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/im"
	"zhw-mini/services/go-api/internal/notifications"
	"zhw-mini/services/go-api/internal/users"
)

type imService interface {
	EnsureRoom(gameID int64) im.Room
	RoomForGame(userID int64, gameID int64) (im.Room, error)
	Session(userID int64, gameID int64) (im.Session, error)
	Send(userID int64, gameID int64, req im.SendRequest) (im.Message, error)
	SendSystem(userID int64, gameID int64, req im.SendRequest) (im.Message, error)
	SendToRoom(userID int64, roomID int64, req im.SendRequest) (im.Message, error)
	SendPrivate(userID int64, targetUserID int64, sourceGameID int64, req im.SendRequest) (im.PrivateMessage, im.PrivateConversation, error)
	Messages(userID int64, gameID int64) ([]im.Message, error)
	MessagesByRoom(userID int64, roomID int64) ([]im.Message, error)
	PrivateMessages(userID int64, targetUserID int64, sourceGameID int64) ([]im.PrivateMessage, im.PrivateConversation, error)
	AuthorizeGameAccess(userID int64, gameID int64) error
	AuthorizeRoomAccess(userID int64, roomID int64) (im.Room, error)
	AdminMessagesByRoom(roomID int64) ([]im.Message, im.Room, error)
	AdminRooms() []im.Room
	AdminArchiveRoom(roomID int64, reason string) (im.Room, error)
	AdminRetryCreateRoom(roomID int64) (im.Room, error)
	AdminHideMessage(messageID int64, reason string) (im.Message, error)
	AllMessages() []im.Message
	Ack(userID int64, roomID int64, messageID int64) (im.Message, error)
	MarkRead(userID int64, roomID int64, messageID int64) (im.Message, error)
	ArchiveRoom(userID int64, roomID int64, reason string) (im.Room, error)
	ArchiveRoomsByGameIDs(gameIDs []int64, reason string) []im.Room
	ReadOnlyRoomsByGameIDs(gameIDs []int64, reason string) []im.Room
	RecordWebhook(command string, gameID int64, roomID int64, payload []byte) im.WebhookEvent
	CheckSensitiveWords(content string) (im.SensitiveWord, bool)
	SensitiveWords() []im.SensitiveWord
	CreateSensitiveWord(req im.SensitiveWordRequest) (im.SensitiveWord, error)
	UpdateSensitiveWord(wordID int64, req im.UpdateSensitiveWordRequest) (im.SensitiveWord, error)
	ImportSensitiveWords(words []string) []im.SensitiveWord
	ContentRiskLogs() []im.ContentRiskLog
	AiContentRiskPlaceholder(req im.ContentRiskCheckRequest) im.ContentRiskLog
}

func (s *Server) chatRoom(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/chat-room")
	if !ok {
		return
	}
	room, err := s.im.RoomForGame(userID, gameID)
	if err != nil {
		if errors.Is(err, im.ErrRoomNotFound) {
			s.im.EnsureRoom(gameID)
			room, err = s.im.RoomForGame(userID, gameID)
		}
	}
	if err != nil {
		writeIMError(w, err)
		return
	}
	s.recordBehavior(userID, "enter_im", "game", gameID, nil)
	game, _ := s.games.Get(gameID)
	httpx.OK(w, s.chatRoomPayload(room, userID, game))
}

func (s *Server) chatRoomPayload(room im.Room, currentUserID int64, game games.Game) map[string]interface{} {
	return map[string]interface{}{
		"id":                           room.ID,
		"gameId":                       room.GameID,
		"title":                        imRoomTitle(room.GameID, game.Title),
		"status":                       room.Status,
		"memberIds":                    room.MemberIDs,
		"memberCount":                  len(room.MemberIDs),
		"members":                      s.imRoomMembers(room, game, currentUserID),
		"engine":                       room.Engine,
		"openIMGroupId":                room.OpenIMGroupID,
		"archivedAt":                   room.ArchivedAt,
		"archiveReason":                room.ArchiveReason,
		"createdAt":                    appDisplayTime(room.CreatedAt),
		"createdAtText":                formatAppDisplayTime(room.CreatedAt, "2006-01-02 15:04"),
		"currentUserId":                currentUserID,
		"canAccessServiceConfirmation": s.canAccessServiceConfirmation(game, currentUserID),
		"collaborationEntry": map[string]interface{}{
			"text":  "协作",
			"route": "pages/game/collaboration/index?gameId=" + strconv.FormatInt(room.GameID, 10),
		},
	}
}

func (s *Server) chatSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/chat-session")
	if !ok {
		return
	}
	session, err := s.im.Session(userID, gameID)
	if err != nil {
		writeIMError(w, err)
		return
	}
	s.recordBehavior(userID, "enter_im", "game", gameID, nil)
	payload := map[string]interface{}{
		"engine":        session.Engine,
		"imUserId":      session.IMUserID,
		"openIMGroupId": session.OpenIMGroupID,
		"openIMToken":   session.OpenIMToken,
		"currentUserId": userID,
	}
	if game, err := s.games.Get(gameID); err == nil {
		payload["title"] = imRoomTitle(gameID, game.Title)
	}
	httpx.OK(w, payload)
}

func (s *Server) sendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/chat/messages")
	if !ok {
		return
	}
	var req im.SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	if !s.validateChatMessageFile(w, req, gameID) {
		return
	}
	message, err := s.im.Send(userID, gameID, req)
	if err != nil {
		writeIMError(w, err)
		return
	}
	s.recordBehavior(userID, "send_message", "game", gameID, map[string]interface{}{"messageId": message.ID, "messageType": message.Type})
	s.createIMMessageNotifications(userID, message)
	httpx.OK(w, s.inGameMessageDTO(message))
}

func (s *Server) sendMessageToRoom(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	roomID, ok := roomIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}
	var req im.SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	room, err := s.im.AuthorizeRoomAccess(userID, roomID)
	if err != nil {
		writeIMError(w, err)
		return
	}
	if !s.validateChatMessageFile(w, req, room.GameID) {
		return
	}
	message, err := s.im.SendToRoom(userID, roomID, req)
	if err != nil {
		writeIMError(w, err)
		return
	}
	s.recordBehavior(userID, "send_message", "room", roomID, map[string]interface{}{"messageId": message.ID, "messageType": message.Type})
	s.createIMMessageNotifications(userID, message)
	httpx.OK(w, s.inGameMessageDTO(message))
}

func (s *Server) historyMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/chat/messages")
	if !ok {
		return
	}
	messages, err := s.im.Messages(userID, gameID)
	if err != nil {
		writeIMError(w, err)
		return
	}
	httpx.OK(w, map[string]interface{}{"items": s.inGameMessageDTOs(messages)})
}

func (s *Server) historyMessagesByRoom(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	roomID, ok := roomIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}
	messages, err := s.im.MessagesByRoom(userID, roomID)
	if err != nil {
		writeIMError(w, err)
		return
	}
	httpx.OK(w, map[string]interface{}{"items": s.inGameMessageDTOs(messages)})
}

func (s *Server) privateChatMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	targetUserID, ok := requiredInt64Query(w, r, "targetUserId")
	if !ok {
		return
	}
	sourceGameID, _, ok := optionalInt64Query(w, r, "sourceGameId")
	if !ok {
		return
	}
	target, found := s.auth.UserByID(targetUserID)
	if !found || target.Status != "active" {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "用户不存在")
		return
	}
	messages, conversation, err := s.im.PrivateMessages(userID, targetUserID, sourceGameID)
	if err != nil {
		writeIMError(w, err)
		return
	}
	s.recordBehavior(userID, "enter_private_chat", "user", targetUserID, map[string]interface{}{"sourceGameId": sourceGameID})
	httpx.OK(w, map[string]interface{}{
		"conversation":  conversation,
		"currentUserId": userID,
		"targetUser":    s.privateChatUserDTO(target),
		"items":         s.privateChatMessageDTOs(messages),
	})
}

func (s *Server) sendPrivateChatMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req struct {
		TargetUserID int64  `json:"targetUserId"`
		SourceGameID int64  `json:"sourceGameId"`
		MessageType  string `json:"messageType"`
		Content      string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	if req.TargetUserID <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "targetUserId invalid")
		return
	}
	target, found := s.auth.UserByID(req.TargetUserID)
	if !found || target.Status != "active" {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "用户不存在")
		return
	}
	message, conversation, err := s.im.SendPrivate(userID, req.TargetUserID, req.SourceGameID, im.SendRequest{
		MessageType: req.MessageType,
		Content:     req.Content,
	})
	if err != nil {
		writeIMError(w, err)
		return
	}
	s.recordBehavior(userID, "send_private_message", "user", req.TargetUserID, map[string]interface{}{"conversationId": conversation.ID, "messageId": message.ID, "sourceGameId": req.SourceGameID})
	s.createPrivateMessageNotification(userID, req.TargetUserID, message)
	httpx.OK(w, map[string]interface{}{
		"conversation": conversation,
		"targetUser":   s.privateChatUserDTO(target),
		"message":      s.privateChatMessageDTO(message),
	})
}

func (s *Server) createIMMessageNotifications(senderID int64, message im.Message) {
	if s == nil || s.notices == nil || s.im == nil || message.Status != "sent" {
		return
	}
	_, room, err := s.im.AdminMessagesByRoom(message.RoomID)
	if err != nil {
		return
	}
	senderName := s.displayName(senderID, "用户 "+strconv.FormatInt(senderID, 10))
	summary := imNotificationSummary(message.Type, message.Content, message.FileID)
	content := senderName + "：" + summary
	for _, memberID := range room.MemberIDs {
		if memberID <= 0 || memberID == senderID {
			continue
		}
		s.notices.Create(notifications.CreateRequest{
			UserID:     memberID,
			NotifyType: "im_message",
			Title:      "收到局内消息",
			Content:    content,
			BizType:    "game",
			BizID:      message.GameID,
		})
	}
}

func (s *Server) createPrivateMessageNotification(senderID int64, targetUserID int64, message im.PrivateMessage) {
	if s == nil || s.notices == nil || targetUserID <= 0 || message.Status != "sent" {
		return
	}
	senderName := s.displayName(senderID, "用户 "+strconv.FormatInt(senderID, 10))
	s.notices.Create(notifications.CreateRequest{
		UserID:     targetUserID,
		NotifyType: "private_message",
		Title:      "收到私聊消息",
		Content:    senderName + "：" + imNotificationSummary(message.Type, message.Content, 0),
		BizType:    "private_chat",
		BizID:      senderID,
	})
}

func imNotificationSummary(messageType string, content string, fileID int64) string {
	switch messageType {
	case "image":
		return "[图片]"
	case "file":
		return "[文件]"
	}
	if fileID > 0 {
		return "[文件]"
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return "新消息"
	}
	if len([]rune(content)) > 48 {
		return string([]rune(content)[:48]) + "..."
	}
	return content
}

func (s *Server) inGameMessageDTOs(messages []im.Message) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(messages))
	for _, message := range messages {
		items = append(items, s.inGameMessageDTO(message))
	}
	return items
}

func (s *Server) privateChatMessageDTOs(messages []im.PrivateMessage) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(messages))
	for _, message := range messages {
		items = append(items, s.privateChatMessageDTO(message))
	}
	return items
}

func (s *Server) privateChatMessageDTO(message im.PrivateMessage) map[string]interface{} {
	return map[string]interface{}{
		"id":             message.ID,
		"messageId":      message.ID,
		"conversationId": message.ConversationID,
		"senderUserId":   message.SenderID,
		"targetUserId":   message.TargetID,
		"sourceGameId":   message.SourceGameID,
		"messageType":    message.Type,
		"type":           message.Type,
		"content":        message.Content,
		"status":         message.Status,
		"createdAt":      appDisplayTime(message.CreatedAt),
		"createdAtText":  formatAppDisplayTime(message.CreatedAt, "2006-01-02 15:04"),
	}
}

func (s *Server) privateChatUserDTO(user users.User) map[string]interface{} {
	nickname := strings.TrimSpace(user.Nickname)
	if nickname == "" {
		nickname = "用户 " + strconv.FormatInt(user.ID, 10)
	}
	return map[string]interface{}{
		"id":        user.ID,
		"userId":    user.ID,
		"nickname":  nickname,
		"name":      nickname,
		"avatarUrl": strings.TrimSpace(user.AvatarURL),
	}
}

func (s *Server) inGameMessageDTO(message im.Message) map[string]interface{} {
	profile := s.inGameIdentity(message.SenderID, "成员")
	roleText := "系统"
	if message.SenderID <= 0 {
		profile.DisplayName = "组局助手"
		profile.AvatarText = "助"
	} else {
		roleText = imRoleText(s.imSenderRole(message.GameID, message.SenderID))
	}
	return map[string]interface{}{
		"id": message.ID, "messageId": message.ID, "roomId": message.RoomID, "gameId": message.GameID,
		"senderUserId": message.SenderID, "senderId": message.SenderID, "senderName": profile.DisplayName,
		"senderRoleText": roleText, "roleText": roleText,
		"realName": profile.RealName, "displayName": profile.DisplayName, "avatarText": profile.AvatarText,
		// In-message avatars are real-name initials (for example 陈思 -> CS).
		// Profile image URLs remain available on room members/group previews only.
		"avatarSrc": "", "avatarUrl": "",
		"messageType": message.Type, "type": message.Type, "content": message.Content, "fileName": message.Content, "fileId": message.FileID,
		"status": message.Status, "ackedBy": message.AckedBy, "readBy": message.ReadBy,
		"createdAt": appDisplayTime(message.CreatedAt), "createdAtText": formatAppDisplayTime(message.CreatedAt, "2006-01-02 15:04"),
	}
}

func (s *Server) imRoomMembers(room im.Room, game games.Game, currentUserID int64) []map[string]interface{} {
	memberRoles := gameMemberRoleMap(s.games.MemberRoles(room.GameID))
	items := make([]map[string]interface{}, 0, len(room.MemberIDs))
	for _, memberID := range room.MemberIDs {
		profile := s.inGameIdentity(memberID, "成员")
		role := gameMemberRole(game, memberID, true, memberRoles)
		avatarURL := s.imUserAvatarURL(memberID)
		items = append(items, map[string]interface{}{
			"id": memberID, "userId": memberID,
			"name": profile.DisplayName, "nickname": profile.DisplayName,
			"realName": profile.RealName, "displayName": profile.DisplayName,
			"role": role, "roleText": imRoleText(role), "roleLabel": imRoleText(role),
			"avatarText": profile.AvatarText, "avatarSrc": avatarURL, "avatarUrl": avatarURL,
			"isCurrentUser": memberID == currentUserID,
		})
	}
	return items
}

func (s *Server) imSenderRole(gameID int64, userID int64) string {
	game, err := s.games.Get(gameID)
	if err != nil {
		return "member"
	}
	memberRoles := gameMemberRoleMap(s.games.MemberRoles(gameID))
	return gameMemberRole(game, userID, true, memberRoles)
}

func (s *Server) imUserAvatarURL(userID int64) string {
	if user, ok := s.auth.UserByID(userID); ok {
		return strings.TrimSpace(user.AvatarURL)
	}
	return ""
}

func imRoomTitle(gameID int64, title string) string {
	if title = strings.TrimSpace(title); title != "" {
		return title
	}
	return "局" + strconv.FormatInt(gameID, 10)
}

func imRoleText(role string) string {
	switch role {
	case "creator":
		return "发起人"
	case "main_guide", "guide":
		return "领路人"
	case "expert":
		return "行家"
	case "guest":
		return "局外用户"
	default:
		return "玩家"
	}
}

func (s *Server) ackMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	roomID, messageID, ok := roomMessageIDFromPath(w, r.URL.Path, "/ack")
	if !ok {
		return
	}
	message, err := s.im.Ack(userID, roomID, messageID)
	if err != nil {
		writeIMError(w, err)
		return
	}
	httpx.OK(w, message)
}

func (s *Server) readMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	roomID, messageID, ok := roomMessageIDFromPath(w, r.URL.Path, "/read")
	if !ok {
		return
	}
	message, err := s.im.MarkRead(userID, roomID, messageID)
	if err != nil {
		writeIMError(w, err)
		return
	}
	httpx.OK(w, message)
}

func (s *Server) archiveRoom(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	idText := strings.TrimPrefix(r.URL.Path, "/api/app/chat/rooms/")
	idText = strings.TrimSuffix(idText, "/archive")
	roomID, err := strconv.ParseInt(strings.Trim(idText, "/"), 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "房间 ID 错误")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	room, err := s.im.ArchiveRoom(userID, roomID, req.Reason)
	if err != nil {
		writeIMError(w, err)
		return
	}
	httpx.OK(w, room)
}

func (s *Server) openIMWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "回调参数错误")
		return
	}
	var req struct {
		CallbackCommand string `json:"callbackCommand"`
		GroupID         string `json:"groupID"`
		GroupIDAlt      string `json:"groupId"`
	}
	_ = json.Unmarshal(payload, &req)
	groupID := req.GroupID
	if groupID == "" {
		groupID = req.GroupIDAlt
	}
	event := s.im.RecordWebhook(req.CallbackCommand, gameIDFromOpenIMGroupID(groupID), 0, payload)
	httpx.OK(w, event)
}

func (s *Server) validateChatMessageFile(w http.ResponseWriter, req im.SendRequest, gameID int64) bool {
	if req.MessageType != "image" && req.MessageType != "file" {
		return true
	}
	file, err := s.files.Get(req.FileID)
	if err != nil {
		if errors.Is(err, files.ErrFileNotFound) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "file not found")
			return false
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "failed to get file")
		return false
	}
	if file.BizType != "chat_file" || file.ObjectID != gameID {
		httpx.Error(w, http.StatusForbidden, 40331, "file access denied")
		return false
	}
	return true
}

func (s *Server) routeAdminIMRoomGet(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/dispute-messages") {
		s.requireAdminPermission("im:message:view_dispute", s.adminDisputeMessages)(w, r)
		return
	}
	s.requireAdminPermission("im:room:read", s.adminIMRoomDetail)(w, r)
}

func (s *Server) routeAdminIMRoomPost(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasSuffix(r.URL.Path, "/retry-create"):
		s.requireAdminPermission("im:room:retry_create", s.adminRetryCreateIMRoom)(w, r)
	case strings.HasSuffix(r.URL.Path, "/archive"):
		s.requireAdminPermission("im:room:archive", s.adminArchiveIMRoom)(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) routeAdminIMMessagePost(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/hide") {
		s.requireAdminPermission("im:message:hide", s.adminHideIMMessage)(w, r)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) adminIMRooms(w http.ResponseWriter, r *http.Request) {
	gameFilter, hasGameFilter, ok := optionalInt64Query(w, r, "gameId")
	if !ok {
		return
	}
	statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))
	messages := s.im.AllMessages()
	items := make([]map[string]interface{}, 0)
	for _, room := range s.im.AdminRooms() {
		if hasGameFilter && room.GameID != gameFilter {
			continue
		}
		if statusFilter != "" && room.Status != statusFilter {
			continue
		}
		messageCount, fileMessageCount := countRoomMessages(messages, room.ID)
		items = append(items, adminIMRoomDTO(room, messageCount, fileMessageCount))
	}
	s.recordOperation(r, "im:room:read", "im_room", "list", map[string]interface{}{
		"count":  len(items),
		"gameId": gameFilter,
		"status": statusFilter,
	})
	httpx.OK(w, map[string]interface{}{"items": items, "total": len(items)})
}

func (s *Server) adminIMRoomDetail(w http.ResponseWriter, r *http.Request) {
	roomID, ok := adminRoomIDFromPath(w, r.URL.Path, "")
	if !ok {
		return
	}
	messages, room, err := s.im.AdminMessagesByRoom(roomID)
	if err != nil {
		writeIMError(w, err)
		return
	}
	fileMessages := filterFileMessages(messages)
	s.recordOperation(r, "im:room:read", "im_room", strconv.FormatInt(roomID, 10), map[string]interface{}{
		"gameId":           room.GameID,
		"messageCount":     len(messages),
		"fileMessageCount": len(fileMessages),
	})
	httpx.OK(w, map[string]interface{}{
		"room":             room,
		"messages":         messages,
		"fileMessages":     fileMessages,
		"messageCount":     len(messages),
		"fileMessageCount": len(fileMessages),
	})
}

func (s *Server) adminDisputeMessages(w http.ResponseWriter, r *http.Request) {
	roomID, ok := adminRoomIDFromPath(w, r.URL.Path, "/dispute-messages")
	if !ok {
		return
	}
	messages, room, err := s.im.AdminMessagesByRoom(roomID)
	if err != nil {
		writeIMError(w, err)
		return
	}
	s.recordOperation(r, "im:message:view_dispute", "im_room", strconv.FormatInt(roomID, 10), map[string]interface{}{
		"gameId":       room.GameID,
		"messageCount": len(messages),
	})
	httpx.OK(w, map[string]interface{}{"room": room, "items": messages})
}

func (s *Server) adminRetryCreateIMRoom(w http.ResponseWriter, r *http.Request) {
	roomID, ok := adminRoomIDFromPath(w, r.URL.Path, "/retry-create")
	if !ok {
		return
	}
	room, err := s.im.AdminRetryCreateRoom(roomID)
	if err != nil {
		writeIMError(w, err)
		return
	}
	s.recordOperation(r, "im:room:retry_create", "im_room", strconv.FormatInt(roomID, 10), map[string]interface{}{
		"gameId":        room.GameID,
		"engine":        room.Engine,
		"openIMGroupId": room.OpenIMGroupID,
	})
	httpx.OK(w, room)
}

func (s *Server) adminArchiveIMRoom(w http.ResponseWriter, r *http.Request) {
	roomID, ok := adminRoomIDFromPath(w, r.URL.Path, "/archive")
	if !ok {
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	room, err := s.im.AdminArchiveRoom(roomID, req.Reason)
	if err != nil {
		writeIMError(w, err)
		return
	}
	s.recordOperation(r, "im:room:archive", "im_room", strconv.FormatInt(roomID, 10), map[string]interface{}{
		"gameId": room.GameID,
		"reason": room.ArchiveReason,
	})
	httpx.OK(w, room)
}

func (s *Server) adminHideIMMessage(w http.ResponseWriter, r *http.Request) {
	messageID, ok := adminIMMessageIDFromPath(w, r.URL.Path, "/hide")
	if !ok {
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	message, err := s.im.AdminHideMessage(messageID, req.Reason)
	if err != nil {
		writeIMError(w, err)
		return
	}
	s.recordOperation(r, "im:message:hide", "im_message", strconv.FormatInt(messageID, 10), map[string]interface{}{
		"gameId": message.GameID,
		"roomId": message.RoomID,
		"reason": strings.TrimSpace(req.Reason),
	})
	httpx.OK(w, message)
}

func (s *Server) adminSensitiveWords(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"items": s.im.SensitiveWords()})
}

func (s *Server) createAdminSensitiveWord(w http.ResponseWriter, r *http.Request) {
	var req im.SensitiveWordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	word, err := s.im.CreateSensitiveWord(req)
	if err != nil {
		writeIMError(w, err)
		return
	}
	s.recordOperation(r, "content:sensitive_word:create", "sensitive_word", strconv.FormatInt(word.ID, 10), map[string]interface{}{"word": word.Word, "action": word.Action})
	httpx.OK(w, word)
}

func (s *Server) updateAdminSensitiveWord(w http.ResponseWriter, r *http.Request) {
	wordID, ok := sensitiveWordIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}
	var req im.UpdateSensitiveWordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	word, err := s.im.UpdateSensitiveWord(wordID, req)
	if err != nil {
		writeIMError(w, err)
		return
	}
	s.recordOperation(r, "content:sensitive_word:update", "sensitive_word", strconv.FormatInt(word.ID, 10), map[string]interface{}{"status": word.Status})
	httpx.OK(w, word)
}

func (s *Server) importAdminSensitiveWords(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Words []string `json:"words"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	items := s.im.ImportSensitiveWords(req.Words)
	s.recordOperation(r, "content:sensitive_word:import", "sensitive_word", "batch", map[string]interface{}{"count": len(items)})
	httpx.OK(w, map[string]interface{}{"items": items, "count": len(items)})
}

func (s *Server) adminContentRiskLogs(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"items": s.im.ContentRiskLogs()})
}

func (s *Server) aiContentRiskPlaceholder(w http.ResponseWriter, r *http.Request) {
	var req im.ContentRiskCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	log := s.im.AiContentRiskPlaceholder(req)
	httpx.OK(w, log)
}

func roomIDFromPath(w http.ResponseWriter, path string) (int64, bool) {
	idText := strings.TrimPrefix(path, "/api/app/chat/rooms/")
	idText = strings.TrimSuffix(idText, "/messages")
	id, err := strconv.ParseInt(strings.Trim(idText, "/"), 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "房间 ID 错误")
		return 0, false
	}
	return id, true
}

func sensitiveWordIDFromPath(w http.ResponseWriter, path string) (int64, bool) {
	text := strings.TrimPrefix(path, "/api/admin/sensitive-words/")
	id, err := strconv.ParseInt(strings.Trim(text, "/"), 10, 64)
	if err != nil || id <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "敏感词 ID 错误")
		return 0, false
	}
	return id, true
}

func adminRoomIDFromPath(w http.ResponseWriter, path string, suffix string) (int64, bool) {
	text := strings.TrimSuffix(strings.TrimPrefix(path, "/api/admin/im/rooms/"), suffix)
	id, err := strconv.ParseInt(strings.Trim(text, "/"), 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "房间 ID 错误")
		return 0, false
	}
	return id, true
}

func adminIMMessageIDFromPath(w http.ResponseWriter, path string, suffix string) (int64, bool) {
	text := strings.TrimSuffix(strings.TrimPrefix(path, "/api/admin/im/messages/"), suffix)
	id, err := strconv.ParseInt(strings.Trim(text, "/"), 10, 64)
	if err != nil || id <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "娑堟伅 ID 閿欒")
		return 0, false
	}
	return id, true
}

func optionalInt64Query(w http.ResponseWriter, r *http.Request, name string) (int64, bool, bool) {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if value == "" {
		return 0, false, true
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, name+" invalid")
		return 0, true, false
	}
	return id, true, true
}

func requiredInt64Query(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	id, err := strconv.ParseInt(value, 10, 64)
	if value == "" || err != nil || id <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, name+" invalid")
		return 0, false
	}
	return id, true
}

func adminIMRoomDTO(room im.Room, messageCount int, fileMessageCount int) map[string]interface{} {
	return map[string]interface{}{
		"id":               room.ID,
		"gameId":           room.GameID,
		"status":           room.Status,
		"memberIds":        room.MemberIDs,
		"engine":           room.Engine,
		"openIMGroupId":    room.OpenIMGroupID,
		"archivedAt":       room.ArchivedAt,
		"archiveReason":    room.ArchiveReason,
		"createdAt":        appDisplayTime(room.CreatedAt),
		"createdAtText":    formatAppDisplayTime(room.CreatedAt, "2006-01-02 15:04"),
		"messageCount":     messageCount,
		"fileMessageCount": fileMessageCount,
	}
}

func countRoomMessages(messages []im.Message, roomID int64) (int, int) {
	messageCount := 0
	fileMessageCount := 0
	for _, message := range messages {
		if message.RoomID != roomID {
			continue
		}
		messageCount++
		if isFileMessage(message) {
			fileMessageCount++
		}
	}
	return messageCount, fileMessageCount
}

func filterFileMessages(messages []im.Message) []im.Message {
	items := make([]im.Message, 0)
	for _, message := range messages {
		if isFileMessage(message) {
			items = append(items, message)
		}
	}
	return items
}

func isFileMessage(message im.Message) bool {
	return message.Type == "image" || message.Type == "file" || message.FileID > 0
}

func roomMessageIDFromPath(w http.ResponseWriter, path string, suffix string) (int64, int64, bool) {
	text := strings.TrimPrefix(path, "/api/app/chat/rooms/")
	text = strings.TrimSuffix(text, suffix)
	parts := strings.Split(strings.Trim(text, "/"), "/")
	if len(parts) != 3 || parts[1] != "messages" {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "消息路径错误")
		return 0, 0, false
	}
	roomID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "房间 ID 错误")
		return 0, 0, false
	}
	messageID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "消息 ID 错误")
		return 0, 0, false
	}
	return roomID, messageID, true
}

func gameIDFromOpenIMGroupID(groupID string) int64 {
	value := strings.TrimPrefix(groupID, "zhw_game_")
	id, _ := strconv.ParseInt(value, 10, 64)
	return id
}

func writeIMError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, im.ErrForbidden):
		httpx.Error(w, http.StatusForbidden, 40331, "非局成员不能访问 IM")
	case errors.Is(err, im.ErrRoomNotFound):
		httpx.Error(w, http.StatusNotFound, 40431, "IM 房间不存在")
	case errors.Is(err, im.ErrMessageNotFound):
		httpx.Error(w, http.StatusNotFound, 40432, "IM 消息不存在")
	case errors.Is(err, im.ErrInvalidMessage):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "IM 消息参数错误")
	case errors.Is(err, im.ErrSensitive):
		httpx.Error(w, http.StatusUnavailableForLegalReasons, 45101, "消息命中敏感词")
	case errors.Is(err, im.ErrExternalIM):
		httpx.Error(w, http.StatusBadGateway, 50231, "OpenIM 服务调用失败")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "IM 操作失败")
	}
}
