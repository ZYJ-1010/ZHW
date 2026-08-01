package im

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrRoomNotFound     = errors.New("room not found")
	ErrForbidden        = errors.New("forbidden")
	ErrSensitive        = errors.New("sensitive word hit")
	ErrExternalIM       = errors.New("openim operation failed")
	ErrMessageNotFound  = errors.New("message not found")
	ErrInvalidMessage   = errors.New("invalid message")
	ErrInvalidImageFile = errors.New("image message requires image file")
	ErrInvalidVoiceFile = errors.New("voice message requires audio file")
)

type GameMemberChecker interface {
	Members(gameID int64) []int64
	IsMember(gameID int64, userID int64) bool
}

type GameReadOnlyChecker interface {
	IsIMReadOnly(gameID int64) bool
}

// GameRoomReadyChecker lets the IM layer avoid creating rooms for recruiting
// games that have not reached a valid start state yet.
type GameRoomReadyChecker interface {
	IsIMRoomReady(gameID int64) bool
}

type Repository interface {
	EnsureRoom(ctx context.Context, gameID int64, memberIDs []int64, engine string, openIMGroupID string) (Room, error)
	RoomByGame(ctx context.Context, gameID int64) (Room, bool, error)
	RoomByID(ctx context.Context, roomID int64) (Room, bool, error)
	ListRooms(ctx context.Context) ([]Room, error)
	SaveRoom(ctx context.Context, room Room) (Room, error)
	SaveMessage(ctx context.Context, message Message) (Message, error)
	MessageByID(ctx context.Context, roomID int64, messageID int64) (Message, bool, error)
	ListMessagesByRoom(ctx context.Context, roomID int64) ([]Message, error)
	ListMessages(ctx context.Context) ([]Message, error)
	UpdateMessage(ctx context.Context, message Message) (Message, error)
	EnsurePrivateConversation(ctx context.Context, userID int64, targetUserID int64, sourceGameID int64) (PrivateConversation, error)
	SavePrivateMessage(ctx context.Context, message PrivateMessage) (PrivateMessage, error)
	ListPrivateMessages(ctx context.Context, conversationID int64) ([]PrivateMessage, error)
}

type Room struct {
	ID            int64     `json:"id"`
	GameID        int64     `json:"gameId"`
	Status        string    `json:"status"`
	MemberIDs     []int64   `json:"memberIds"`
	Engine        string    `json:"engine"`
	OpenIMGroupID string    `json:"openIMGroupId,omitempty"`
	ArchivedAt    string    `json:"archivedAt,omitempty"`
	ArchiveReason string    `json:"archiveReason,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

type Message struct {
	ID        int64     `json:"id"`
	RoomID    int64     `json:"roomId"`
	GameID    int64     `json:"gameId"`
	SenderID  int64     `json:"senderUserId"`
	Type      string    `json:"messageType"`
	Content   string    `json:"content"`
	FileID    int64     `json:"fileId,omitempty"`
	Status    string    `json:"status"`
	AckedBy   []int64   `json:"ackedBy,omitempty"`
	ReadBy    []int64   `json:"readBy,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type PrivateConversation struct {
	ID           int64     `json:"id"`
	UserAID      int64     `json:"userAId"`
	UserBID      int64     `json:"userBId"`
	SourceGameID int64     `json:"sourceGameId,omitempty"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type PrivateMessage struct {
	ID             int64     `json:"id"`
	ConversationID int64     `json:"conversationId"`
	SenderID       int64     `json:"senderUserId"`
	TargetID       int64     `json:"targetUserId"`
	SourceGameID   int64     `json:"sourceGameId,omitempty"`
	Type           string    `json:"messageType"`
	Content        string    `json:"content"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}

type SendRequest struct {
	MessageType string             `json:"messageType"`
	Content     string             `json:"content"`
	FileID      int64              `json:"fileId"`
	Width       int32              `json:"width,omitempty"`
	Height      int32              `json:"height,omitempty"`
	DurationMS  int64              `json:"durationMs,omitempty"`
	Attachment  *MessageAttachment `json:"-"`
}

type MessageAttachment struct {
	URL      string
	FileName string
	MimeType string
	Size     int64
}

type Session struct {
	Engine        string `json:"engine"`
	IMUserID      string `json:"imUserId"`
	OpenIMGroupID string `json:"openIMGroupId,omitempty"`
	OpenIMToken   string `json:"openIMToken,omitempty"`
}

type WebhookEvent struct {
	ID        int64           `json:"id"`
	Command   string          `json:"command"`
	GameID    int64           `json:"gameId,omitempty"`
	RoomID    int64           `json:"roomId,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
}

type SensitiveWord struct {
	ID        int64     `json:"id"`
	Word      string    `json:"word"`
	Level     string    `json:"level"`
	Action    string    `json:"action"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type SensitiveWordRequest struct {
	Word   string `json:"word"`
	Level  string `json:"level"`
	Action string `json:"action"`
	Status string `json:"status"`
}

// SensitiveWordStore persists the moderation dictionary independently from
// the IM room/message repository. This keeps word changes effective after a
// process restart and allows the app layer to reuse the existing system
// configuration repository.
type SensitiveWordStore interface {
	LoadSensitiveWords(ctx context.Context) ([]SensitiveWord, error)
	SaveSensitiveWords(ctx context.Context, words []SensitiveWord) error
}

type UpdateSensitiveWordRequest struct {
	Status string `json:"status"`
}

type ContentRiskLog struct {
	ID        int64     `json:"id"`
	GameID    int64     `json:"gameId,omitempty"`
	RoomID    int64     `json:"roomId,omitempty"`
	MessageID int64     `json:"messageId,omitempty"`
	SenderID  int64     `json:"senderUserId,omitempty"`
	Content   string    `json:"content,omitempty"`
	Word      string    `json:"word,omitempty"`
	Action    string    `json:"action"`
	Status    string    `json:"status"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"createdAt"`
}

type ContentRiskCheckRequest struct {
	GameID  int64  `json:"gameId"`
	RoomID  int64  `json:"roomId"`
	UserID  int64  `json:"userId"`
	Content string `json:"content"`
}

type Service struct {
	mu                            sync.RWMutex
	nextRoomID                    int64
	nextMessageID                 int64
	nextPrivateConversationID     int64
	nextPrivateMessageID          int64
	nextWebhookID                 int64
	nextSensitiveID               int64
	nextRiskID                    int64
	roomsByGame                   map[int64]Room
	messagesByRoom                map[int64][]Message
	privateConversations          map[string]PrivateConversation
	privateMessagesByConversation map[int64][]PrivateMessage
	webhookEvents                 []WebhookEvent
	sensitiveWords                []SensitiveWord
	sensitiveWordStore            SensitiveWordStore
	riskLogs                      []ContentRiskLog
	members                       GameMemberChecker
	openim                        *OpenIMClient
	repo                          Repository
}

func NewService(members GameMemberChecker) *Service {
	return NewServiceWithOpenIM(members, OpenIMConfig{})
}

func NewServiceWithOpenIM(members GameMemberChecker, cfg OpenIMConfig) *Service {
	return &Service{
		nextRoomID:                    1,
		nextMessageID:                 1,
		nextPrivateConversationID:     1,
		nextPrivateMessageID:          1,
		nextWebhookID:                 1,
		nextSensitiveID:               2,
		nextRiskID:                    1,
		roomsByGame:                   make(map[int64]Room),
		messagesByRoom:                make(map[int64][]Message),
		privateConversations:          make(map[string]PrivateConversation),
		privateMessagesByConversation: make(map[int64][]PrivateMessage),
		webhookEvents:                 make([]WebhookEvent, 0),
		sensitiveWords: []SensitiveWord{{
			ID:        1,
			Word:      "敏感词",
			Level:     "high",
			Action:    "block",
			Status:    "active",
			CreatedAt: time.Now(),
		}},
		riskLogs: make([]ContentRiskLog, 0),
		members:  members,
		openim:   NewOpenIMClient(cfg),
	}
}

func (s *Service) UseRepository(repo Repository) {
	s.repo = repo
}

func (s *Service) UseSensitiveWordStore(store SensitiveWordStore) error {
	if store == nil {
		return nil
	}
	words, err := store.LoadSensitiveWords(context.Background())
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sensitiveWordStore = store
	if len(words) == 0 {
		return nil
	}
	s.sensitiveWords = normalizeSensitiveWords(words)
	maxID := int64(0)
	for _, word := range s.sensitiveWords {
		if word.ID > maxID {
			maxID = word.ID
		}
	}
	if maxID >= s.nextSensitiveID {
		s.nextSensitiveID = maxID + 1
	}
	return nil
}

func (s *Service) EnsureRoom(gameID int64) Room {
	room, _ := s.EnsureRoomStrict(gameID)
	return room
}

func (s *Service) EnsureRoomStrict(gameID int64) (Room, error) {
	if s.repo != nil {
		// Room reads are frequent (room, session and history load in parallel).
		// Reuse an existing room and only refresh its member snapshot; creating
		// the same OpenIM group again can turn a healthy room into create_failed.
		s.mu.Lock()
		memberIDs := s.members.Members(gameID)
		existing, found, err := s.repo.RoomByGame(context.Background(), gameID)
		if err != nil {
			s.mu.Unlock()
			return Room{}, err
		}
		if found {
			room, syncErr := s.repo.EnsureRoom(
				context.Background(),
				gameID,
				memberIDs,
				existing.Engine,
				existing.OpenIMGroupID,
			)
			if syncErr == nil {
				s.mu.Unlock()
				return room, nil
			}
			s.mu.Unlock()
			return Room{}, syncErr
		}
		engine := "local"
		openIMGroupID := ""
		createFailed := false
		if s.openim != nil {
			if groupID, err := s.openim.SyncGameRoom(context.Background(), gameID, memberIDs); err == nil {
				engine = "openim"
				openIMGroupID = groupID
			} else {
				engine = "openim"
				createFailed = true
			}
		}
		room, err := s.repo.EnsureRoom(context.Background(), gameID, memberIDs, engine, openIMGroupID)
		if err == nil {
			if createFailed {
				room.Status = "create_failed"
				room.Engine = engine
				room.OpenIMGroupID = ""
				saved, saveErr := s.repo.SaveRoom(context.Background(), room)
				if saveErr != nil {
					s.mu.Unlock()
					return Room{}, saveErr
				}
				room = saved
			}
			s.mu.Unlock()
			return room, nil
		}
		s.mu.Unlock()
		return Room{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if room, ok := s.roomsByGame[gameID]; ok {
		room.MemberIDs = append([]int64(nil), s.members.Members(gameID)...)
		s.roomsByGame[gameID] = room
		return room, nil
	}
	memberIDs := s.members.Members(gameID)
	engine := "local"
	openIMGroupID := ""
	status := "active"
	if s.openim != nil {
		if groupID, err := s.openim.SyncGameRoom(context.Background(), gameID, memberIDs); err == nil {
			engine = "openim"
			openIMGroupID = groupID
		} else {
			engine = "openim"
			status = "create_failed"
		}
	}
	room := Room{
		ID:            s.nextRoomID,
		GameID:        gameID,
		Status:        status,
		MemberIDs:     memberIDs,
		Engine:        engine,
		OpenIMGroupID: openIMGroupID,
		CreatedAt:     time.Now(),
	}
	s.nextRoomID++
	s.roomsByGame[gameID] = room
	return room, nil
}

func (s *Service) RoomForGame(userID int64, gameID int64) (Room, error) {
	if err := s.AuthorizeGameAccess(userID, gameID); err != nil {
		return Room{}, err
	}
	if checker, ok := s.members.(GameRoomReadyChecker); ok && !checker.IsIMRoomReady(gameID) {
		return Room{}, ErrRoomNotFound
	}
	if s.repo != nil {
		// EnsureRoom is idempotent and also synchronizes chat_room_members with
		// the current game members. Returning a previously-created room directly
		// left rooms created early with only their original member.
		room, err := s.EnsureRoomStrict(gameID)
		if err != nil {
			return Room{}, err
		}
		if room.Status == "create_failed" {
			return room, ErrExternalIM
		}
		if s.gameReadOnly(gameID) && !roomReadOnly(room) {
			updatedRooms, err := s.ReadOnlyRoomsByGameIDsStrict([]int64{gameID}, "game_ended")
			if err != nil {
				return Room{}, err
			}
			for _, updated := range updatedRooms {
				if updated.GameID == gameID {
					room = updated
					break
				}
			}
		}
		return room, nil
	}
	s.mu.RLock()
	room, ok := s.roomsByGame[gameID]
	s.mu.RUnlock()
	if !ok {
		room = s.EnsureRoom(gameID)
	}
	if room.Status == "create_failed" {
		return room, ErrExternalIM
	}
	if s.gameReadOnly(gameID) && !roomReadOnly(room) {
		updatedRooms, err := s.ReadOnlyRoomsByGameIDsStrict([]int64{gameID}, "game_ended")
		if err != nil {
			return Room{}, err
		}
		for _, updated := range updatedRooms {
			if updated.GameID == gameID {
				room = updated
				break
			}
		}
	}
	return room, nil
}

func (s *Service) RoomForID(userID int64, roomID int64) (Room, error) {
	if s.repo != nil {
		room, ok, err := s.repo.RoomByID(context.Background(), roomID)
		if err != nil {
			return Room{}, err
		}
		if !ok {
			return Room{}, ErrRoomNotFound
		}
		if err := s.AuthorizeGameAccess(userID, room.GameID); err != nil {
			return Room{}, err
		}
		return room, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, room := range s.roomsByGame {
		if room.ID == roomID {
			return s.authorizeRoomAccessLocked(userID, room)
		}
	}
	return Room{}, ErrRoomNotFound
}

func (s *Service) AuthorizeGameAccess(userID int64, gameID int64) error {
	if !s.members.IsMember(gameID, userID) {
		return ErrForbidden
	}
	return nil
}

func (s *Service) AuthorizeRoomAccess(userID int64, roomID int64) (Room, error) {
	return s.RoomForID(userID, roomID)
}

func (s *Service) authorizeRoomAccessLocked(userID int64, room Room) (Room, error) {
	if err := s.AuthorizeGameAccess(userID, room.GameID); err != nil {
		return Room{}, err
	}
	return room, nil
}

func (s *Service) Session(userID int64, gameID int64) (Session, error) {
	if err := s.AuthorizeGameAccess(userID, gameID); err != nil {
		return Session{}, err
	}
	room, err := s.RoomForGame(userID, gameID)
	if err != nil {
		return Session{}, err
	}
	session := Session{
		Engine:        room.Engine,
		IMUserID:      openIMUserID(userID),
		OpenIMGroupID: room.OpenIMGroupID,
	}
	if s.openim != nil && room.Engine == "openim" {
		token, err := s.openim.UserToken(context.Background(), userID)
		if err != nil {
			return Session{}, err
		}
		session.OpenIMToken = token
	}
	return session, nil
}

func (s *Service) Send(userID int64, gameID int64, req SendRequest) (Message, error) {
	return s.send(userID, gameID, req, false)
}

// SendSystem writes a server-generated lifecycle card even after the game room
// has become read-only. It deliberately accepts only the completion reminder
// message type; regular member messages must continue to use Send.
func (s *Service) SendSystem(userID int64, gameID int64, req SendRequest) (Message, error) {
	req = normalizeSendRequest(req)
	if req.MessageType != "service_confirm_remind" {
		return Message{}, ErrInvalidMessage
	}
	return s.send(userID, gameID, req, true)
}

func (s *Service) send(userID int64, gameID int64, req SendRequest, allowReadOnly bool) (Message, error) {
	if err := s.AuthorizeGameAccess(userID, gameID); err != nil {
		return Message{}, err
	}
	if !allowReadOnly && s.gameReadOnly(gameID) {
		return Message{}, ErrInvalidMessage
	}
	room, err := s.RoomForGame(userID, gameID)
	if err != nil {
		return Message{}, err
	}
	if !allowReadOnly && roomReadOnly(room) {
		return Message{}, ErrInvalidMessage
	}
	req = normalizeSendRequest(req)
	if !validMessageRequest(req) {
		return Message{}, ErrInvalidMessage
	}
	var flaggedWord SensitiveWord
	flagged := false
	if req.MessageType == "text" {
		if word, ok := s.CheckSensitiveWords(req.Content); ok {
			if word.Action == "flag" {
				flaggedWord = word
				flagged = true
			} else {
				s.mu.Lock()
				s.createRiskLogLocked(ContentRiskLog{
					GameID:   gameID,
					RoomID:   room.ID,
					SenderID: userID,
					Content:  req.Content,
					Word:     word.Word,
					Action:   word.Action,
					Status:   "blocked",
					Source:   "sensitive_word",
				})
				s.mu.Unlock()
				return Message{}, ErrSensitive
			}
		}
	}
	if s.openim != nil && room.Engine == "openim" && openIMSyncableMessage(req.MessageType) {
		if err := s.openim.SendGroupMessage(context.Background(), gameID, userID, req); err != nil {
			return Message{}, err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	status := "sent"
	if flagged {
		status = "risk_flagged"
	}
	message := Message{
		ID:        s.nextMessageID,
		RoomID:    room.ID,
		GameID:    gameID,
		SenderID:  userID,
		Type:      req.MessageType,
		Content:   req.Content,
		FileID:    req.FileID,
		Status:    status,
		CreatedAt: time.Now(),
	}
	if s.repo != nil {
		saved, err := s.repo.SaveMessage(context.Background(), message)
		if err != nil {
			return Message{}, err
		}
		message = saved
	}
	s.nextMessageID++
	s.messagesByRoom[room.ID] = append(s.messagesByRoom[room.ID], message)
	if flagged {
		s.createRiskLogLocked(ContentRiskLog{
			GameID:    gameID,
			RoomID:    room.ID,
			MessageID: message.ID,
			SenderID:  userID,
			Content:   req.Content,
			Word:      flaggedWord.Word,
			Action:    flaggedWord.Action,
			Status:    "risk_flagged",
			Source:    "sensitive_word",
		})
	}
	return message, nil
}

func (s *Service) gameReadOnly(gameID int64) bool {
	checker, ok := s.members.(GameReadOnlyChecker)
	return ok && checker.IsIMReadOnly(gameID)
}

func (s *Service) SendToRoom(userID int64, roomID int64, req SendRequest) (Message, error) {
	room, err := s.RoomForID(userID, roomID)
	if err != nil {
		return Message{}, err
	}
	return s.Send(userID, room.GameID, req)
}

func (s *Service) Messages(userID int64, gameID int64) ([]Message, error) {
	room, err := s.RoomForGame(userID, gameID)
	if err != nil {
		return nil, err
	}
	if s.repo != nil {
		messages, err := s.repo.ListMessagesByRoom(context.Background(), room.ID)
		if err != nil {
			return nil, err
		}
		return visibleMessages(messages), nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return visibleMessages(s.messagesByRoom[room.ID]), nil
}

func (s *Service) MessagesByRoom(userID int64, roomID int64) ([]Message, error) {
	room, err := s.RoomForID(userID, roomID)
	if err != nil {
		return nil, err
	}
	if s.repo != nil {
		messages, err := s.repo.ListMessagesByRoom(context.Background(), room.ID)
		if err != nil {
			return nil, err
		}
		return visibleMessages(messages), nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return visibleMessages(s.messagesByRoom[room.ID]), nil
}

func (s *Service) PrivateConversation(userID int64, targetUserID int64, sourceGameID int64) (PrivateConversation, error) {
	if userID <= 0 || targetUserID <= 0 || userID == targetUserID {
		return PrivateConversation{}, ErrInvalidMessage
	}
	if s.repo != nil {
		return s.repo.EnsurePrivateConversation(context.Background(), userID, targetUserID, sourceGameID)
	}
	key := privateConversationKey(userID, targetUserID)
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if conversation, ok := s.privateConversations[key]; ok {
		if sourceGameID > 0 && conversation.SourceGameID == 0 {
			conversation.SourceGameID = sourceGameID
			conversation.UpdatedAt = now
			s.privateConversations[key] = conversation
		}
		return conversation, nil
	}
	userAID, userBID := orderedPair(userID, targetUserID)
	conversation := PrivateConversation{
		ID:           s.nextPrivateConversationID,
		UserAID:      userAID,
		UserBID:      userBID,
		SourceGameID: sourceGameID,
		Status:       "active",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.nextPrivateConversationID++
	s.privateConversations[key] = conversation
	return conversation, nil
}

func (s *Service) PrivateMessages(userID int64, targetUserID int64, sourceGameID int64) ([]PrivateMessage, PrivateConversation, error) {
	conversation, err := s.PrivateConversation(userID, targetUserID, sourceGameID)
	if err != nil {
		return nil, PrivateConversation{}, err
	}
	if s.repo != nil {
		messages, err := s.repo.ListPrivateMessages(context.Background(), conversation.ID)
		if err != nil {
			return nil, PrivateConversation{}, err
		}
		return visiblePrivateMessages(messages), conversation, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return visiblePrivateMessages(s.privateMessagesByConversation[conversation.ID]), conversation, nil
}

func (s *Service) SendPrivate(userID int64, targetUserID int64, sourceGameID int64, req SendRequest) (PrivateMessage, PrivateConversation, error) {
	conversation, err := s.PrivateConversation(userID, targetUserID, sourceGameID)
	if err != nil {
		return PrivateMessage{}, PrivateConversation{}, err
	}
	req = normalizeSendRequest(req)
	if req.MessageType != "text" || !validMessageRequest(req) {
		return PrivateMessage{}, PrivateConversation{}, ErrInvalidMessage
	}
	status := "sent"
	var flaggedWord SensitiveWord
	flagged := false
	if word, ok := s.CheckSensitiveWords(req.Content); ok {
		if word.Action == "flag" {
			flaggedWord = word
			flagged = true
			status = "risk_flagged"
		} else {
			s.mu.Lock()
			s.createRiskLogLocked(ContentRiskLog{
				GameID:   sourceGameID,
				SenderID: userID,
				Content:  req.Content,
				Word:     word.Word,
				Action:   word.Action,
				Status:   "blocked",
				Source:   "sensitive_word",
			})
			s.mu.Unlock()
			return PrivateMessage{}, PrivateConversation{}, ErrSensitive
		}
	}
	message := PrivateMessage{
		ConversationID: conversation.ID,
		SenderID:       userID,
		TargetID:       targetUserID,
		SourceGameID:   sourceGameID,
		Type:           req.MessageType,
		Content:        req.Content,
		Status:         status,
		CreatedAt:      time.Now(),
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	message.ID = s.nextPrivateMessageID
	if s.repo != nil {
		saved, err := s.repo.SavePrivateMessage(context.Background(), message)
		if err != nil {
			return PrivateMessage{}, PrivateConversation{}, err
		}
		message = saved
	}
	s.nextPrivateMessageID++
	s.privateMessagesByConversation[conversation.ID] = append(s.privateMessagesByConversation[conversation.ID], message)
	if flagged {
		s.createRiskLogLocked(ContentRiskLog{
			GameID:    sourceGameID,
			MessageID: message.ID,
			SenderID:  userID,
			Content:   req.Content,
			Word:      flaggedWord.Word,
			Action:    flaggedWord.Action,
			Status:    "risk_flagged",
			Source:    "sensitive_word",
		})
	}
	return message, conversation, nil
}

func (s *Service) AdminMessagesByRoom(roomID int64) ([]Message, Room, error) {
	if s.repo != nil {
		room, ok, err := s.repo.RoomByID(context.Background(), roomID)
		if err != nil {
			return nil, Room{}, err
		}
		if !ok {
			return nil, Room{}, ErrRoomNotFound
		}
		messages, err := s.repo.ListMessagesByRoom(context.Background(), roomID)
		if err != nil {
			return nil, Room{}, err
		}
		return messages, room, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, room := range s.roomsByGame {
		if room.ID == roomID {
			return append([]Message(nil), s.messagesByRoom[room.ID]...), room, nil
		}
	}
	return nil, Room{}, ErrRoomNotFound
}

func (s *Service) AdminRooms() []Room {
	rooms, _ := s.AdminRoomsStrict()
	return rooms
}

func (s *Service) AdminRoomsStrict() ([]Room, error) {
	if s.repo != nil {
		rooms, err := s.repo.ListRooms(context.Background())
		if err != nil {
			return nil, err
		}
		return rooms, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Room, 0, len(s.roomsByGame))
	for _, room := range s.roomsByGame {
		result = append(result, room)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result, nil
}

func (s *Service) AdminArchiveRoom(roomID int64, reason string) (Room, error) {
	reason, err := normalizeArchiveReason(reason)
	if err != nil {
		return Room{}, ErrInvalidMessage
	}
	if s.repo != nil {
		room, ok, err := s.repo.RoomByID(context.Background(), roomID)
		if err != nil {
			return Room{}, err
		}
		if !ok {
			return Room{}, ErrRoomNotFound
		}
		room.Status = "archived"
		room.ArchivedAt = time.Now().Format(time.RFC3339)
		room.ArchiveReason = reason
		return s.repo.SaveRoom(context.Background(), room)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, room := range s.roomsByGame {
		if room.ID == roomID {
			room.Status = "archived"
			room.ArchivedAt = time.Now().Format(time.RFC3339)
			room.ArchiveReason = reason
			s.roomsByGame[room.GameID] = room
			return room, nil
		}
	}
	return Room{}, ErrRoomNotFound
}

func (s *Service) AdminRetryCreateRoom(roomID int64) (Room, error) {
	if s.repo != nil {
		target, ok, err := s.repo.RoomByID(context.Background(), roomID)
		if err != nil {
			return Room{}, err
		}
		if !ok {
			return Room{}, ErrRoomNotFound
		}
		if s.openim == nil {
			return target, nil
		}
		groupID, err := s.openim.SyncGameRoom(context.Background(), target.GameID, target.MemberIDs)
		if err != nil {
			target.Status = "create_failed"
			target.Engine = "openim"
			target.OpenIMGroupID = ""
			if _, saveErr := s.repo.SaveRoom(context.Background(), target); saveErr != nil {
				return Room{}, saveErr
			}
			return Room{}, err
		}
		target.Status = "active"
		target.Engine = "openim"
		target.OpenIMGroupID = groupID
		return s.repo.SaveRoom(context.Background(), target)
	}
	s.mu.RLock()
	var target Room
	found := false
	for _, room := range s.roomsByGame {
		if room.ID == roomID {
			target = room
			found = true
			break
		}
	}
	s.mu.RUnlock()
	if !found {
		return Room{}, ErrRoomNotFound
	}
	if s.openim == nil {
		return target, nil
	}
	groupID, err := s.openim.SyncGameRoom(context.Background(), target.GameID, target.MemberIDs)
	if err != nil {
		s.mu.Lock()
		target.Status = "create_failed"
		target.Engine = "openim"
		target.OpenIMGroupID = ""
		s.roomsByGame[target.GameID] = target
		s.mu.Unlock()
		return Room{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	target.Status = "active"
	target.Engine = "openim"
	target.OpenIMGroupID = groupID
	s.roomsByGame[target.GameID] = target
	return target, nil
}

func (s *Service) AdminHideMessage(messageID int64, reason string) (Message, error) {
	reason = strings.TrimSpace(reason)
	if len(reason) > 300 {
		return Message{}, ErrInvalidMessage
	}
	if s.repo != nil {
		messages, err := s.repo.ListMessages(context.Background())
		if err != nil {
			return Message{}, err
		}
		for _, message := range messages {
			if message.ID == messageID {
				message.Status = "hidden"
				return s.repo.UpdateMessage(context.Background(), message)
			}
		}
		return Message{}, ErrMessageNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for roomID, messages := range s.messagesByRoom {
		for index, message := range messages {
			if message.ID != messageID {
				continue
			}
			message.Status = "hidden"
			messages[index] = message
			s.messagesByRoom[roomID] = messages
			s.createRiskLogLocked(ContentRiskLog{
				GameID:    message.GameID,
				RoomID:    message.RoomID,
				MessageID: message.ID,
				SenderID:  message.SenderID,
				Content:   message.Content,
				Action:    "hide",
				Status:    "hidden",
				Source:    "admin",
			})
			return message, nil
		}
	}
	return Message{}, ErrMessageNotFound
}

func (s *Service) AllMessages() []Message {
	messages, _ := s.AllMessagesStrict()
	return messages
}

func (s *Service) AllMessagesStrict() ([]Message, error) {
	if s.repo != nil {
		messages, err := s.repo.ListMessages(context.Background())
		if err != nil {
			return nil, err
		}
		return messages, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Message, 0)
	for _, messages := range s.messagesByRoom {
		result = append(result, messages...)
	}
	return result, nil
}

func (s *Service) Ack(userID int64, roomID int64, messageID int64) (Message, error) {
	return s.markMessage(userID, roomID, messageID, "ack")
}

func (s *Service) MarkRead(userID int64, roomID int64, messageID int64) (Message, error) {
	return s.markMessage(userID, roomID, messageID, "read")
}

func (s *Service) ArchiveRoom(userID int64, roomID int64, reason string) (Room, error) {
	room, err := s.RoomForID(userID, roomID)
	if err != nil {
		return Room{}, err
	}
	reason, err = normalizeArchiveReason(reason)
	if err != nil {
		return Room{}, ErrInvalidMessage
	}
	if s.repo != nil {
		room.Status = "archived"
		room.ArchivedAt = time.Now().Format(time.RFC3339)
		room.ArchiveReason = reason
		return s.repo.SaveRoom(context.Background(), room)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	room.Status = "archived"
	room.ArchivedAt = time.Now().Format(time.RFC3339)
	room.ArchiveReason = reason
	s.roomsByGame[room.GameID] = room
	return room, nil
}

func (s *Service) ArchiveRoomsByGameIDs(gameIDs []int64, reason string) []Room {
	rooms, _ := s.ArchiveRoomsByGameIDsStrict(gameIDs, reason)
	return rooms
}

func (s *Service) ArchiveRoomsByGameIDsStrict(gameIDs []int64, reason string) ([]Room, error) {
	gameIDSet := make(map[int64]struct{}, len(gameIDs))
	for _, gameID := range gameIDs {
		if gameID > 0 {
			gameIDSet[gameID] = struct{}{}
		}
	}
	if len(gameIDSet) == 0 {
		return nil, nil
	}
	normalizedReason, err := normalizeArchiveReason(reason)
	if err != nil {
		normalizedReason = "archive_job"
	}
	if s.repo != nil {
		rooms, err := s.repo.ListRooms(context.Background())
		if err != nil {
			return nil, err
		}
		archived := make([]Room, 0)
		now := time.Now().Format(time.RFC3339)
		for _, room := range rooms {
			if _, ok := gameIDSet[room.GameID]; !ok || room.Status == "archived" {
				continue
			}
			room.Status = "archived"
			room.ArchivedAt = now
			room.ArchiveReason = normalizedReason
			updated, err := s.repo.SaveRoom(context.Background(), room)
			if err != nil {
				return nil, err
			}
			archived = append(archived, updated)
		}
		return archived, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	archived := make([]Room, 0)
	now := time.Now().Format(time.RFC3339)
	for gameID, room := range s.roomsByGame {
		if _, ok := gameIDSet[gameID]; !ok || room.Status == "archived" {
			continue
		}
		room.Status = "archived"
		room.ArchivedAt = now
		room.ArchiveReason = normalizedReason
		s.roomsByGame[gameID] = room
		archived = append(archived, room)
	}
	return archived, nil
}

func (s *Service) ReadOnlyRoomsByGameIDs(gameIDs []int64, reason string) []Room {
	rooms, _ := s.ReadOnlyRoomsByGameIDsStrict(gameIDs, reason)
	return rooms
}

func (s *Service) ReadOnlyRoomsByGameIDsStrict(gameIDs []int64, reason string) ([]Room, error) {
	gameIDSet := make(map[int64]struct{}, len(gameIDs))
	for _, gameID := range gameIDs {
		if gameID > 0 {
			gameIDSet[gameID] = struct{}{}
		}
	}
	if len(gameIDSet) == 0 {
		return nil, nil
	}
	normalizedReason, err := normalizeArchiveReason(reason)
	if err != nil {
		normalizedReason = "game_ended"
	}
	if s.repo != nil {
		rooms, err := s.repo.ListRooms(context.Background())
		if err != nil {
			return nil, err
		}
		updatedRooms := make([]Room, 0)
		for _, room := range rooms {
			if _, ok := gameIDSet[room.GameID]; !ok || room.Status == "readonly" || room.Status == "archived" {
				continue
			}
			room.Status = "readonly"
			room.ArchiveReason = normalizedReason
			updated, err := s.repo.SaveRoom(context.Background(), room)
			if err != nil {
				return nil, err
			}
			updatedRooms = append(updatedRooms, updated)
		}
		return updatedRooms, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	updatedRooms := make([]Room, 0)
	for gameID, room := range s.roomsByGame {
		if _, ok := gameIDSet[gameID]; !ok || room.Status == "readonly" || room.Status == "archived" {
			continue
		}
		room.Status = "readonly"
		room.ArchiveReason = normalizedReason
		s.roomsByGame[gameID] = room
		updatedRooms = append(updatedRooms, room)
	}
	return updatedRooms, nil
}

func (s *Service) RecordWebhook(command string, gameID int64, roomID int64, payload []byte) WebhookEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	event := WebhookEvent{
		ID:        s.nextWebhookID,
		Command:   command,
		GameID:    gameID,
		RoomID:    roomID,
		Payload:   append([]byte(nil), payload...),
		CreatedAt: time.Now(),
	}
	s.nextWebhookID++
	s.webhookEvents = append(s.webhookEvents, event)
	return event
}

func (s *Service) WebhookEvents() []WebhookEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]WebhookEvent(nil), s.webhookEvents...)
}

func (s *Service) SensitiveWords() []SensitiveWord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]SensitiveWord(nil), s.sensitiveWords...)
}

func (s *Service) CreateSensitiveWord(req SensitiveWordRequest) (SensitiveWord, error) {
	req.Word = strings.TrimSpace(req.Word)
	req.Level = strings.TrimSpace(req.Level)
	req.Action = strings.TrimSpace(req.Action)
	req.Status = strings.TrimSpace(req.Status)
	if req.Level == "" {
		req.Level = "medium"
	}
	if req.Action == "" {
		req.Action = "block"
	}
	if req.Status == "" {
		req.Status = "active"
	}
	if req.Word == "" || len(req.Word) > 80 || !validSensitiveWordAction(req.Action) || !validSensitiveWordStatus(req.Status) {
		return SensitiveWord{}, ErrInvalidMessage
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range s.sensitiveWords {
		if item.Word == req.Word {
			return item, nil
		}
	}
	word := SensitiveWord{
		ID:        s.nextSensitiveID,
		Word:      req.Word,
		Level:     req.Level,
		Action:    req.Action,
		Status:    req.Status,
		CreatedAt: time.Now(),
	}
	s.nextSensitiveID++
	s.sensitiveWords = append(s.sensitiveWords, word)
	if err := s.persistSensitiveWordsLocked(); err != nil {
		s.nextSensitiveID--
		s.sensitiveWords = s.sensitiveWords[:len(s.sensitiveWords)-1]
		return SensitiveWord{}, err
	}
	return word, nil
}

func (s *Service) ImportSensitiveWords(words []string) []SensitiveWord {
	result := make([]SensitiveWord, 0)
	for _, word := range words {
		item, err := s.CreateSensitiveWord(SensitiveWordRequest{Word: word, Level: "medium", Action: "block", Status: "active"})
		if err == nil {
			result = append(result, item)
		}
	}
	return result
}

func (s *Service) UpdateSensitiveWord(wordID int64, req UpdateSensitiveWordRequest) (SensitiveWord, error) {
	req.Status = strings.TrimSpace(req.Status)
	if wordID <= 0 || !validSensitiveWordStatus(req.Status) {
		return SensitiveWord{}, ErrInvalidMessage
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for index, item := range s.sensitiveWords {
		if item.ID == wordID {
			previous := item
			item.Status = req.Status
			s.sensitiveWords[index] = item
			if err := s.persistSensitiveWordsLocked(); err != nil {
				s.sensitiveWords[index] = previous
				return SensitiveWord{}, err
			}
			return item, nil
		}
	}
	return SensitiveWord{}, ErrMessageNotFound
}

func normalizeSensitiveWords(words []SensitiveWord) []SensitiveWord {
	result := make([]SensitiveWord, 0, len(words))
	seen := make(map[string]bool, len(words))
	for _, word := range words {
		word.Word = strings.TrimSpace(word.Word)
		if word.Word == "" || seen[word.Word] || !validSensitiveWordAction(word.Action) || !validSensitiveWordStatus(word.Status) {
			continue
		}
		if word.Level == "" {
			word.Level = "medium"
		}
		seen[word.Word] = true
		result = append(result, word)
	}
	return result
}

// persistSensitiveWordsLocked must be called while s.mu is held. The
// dictionary is small and the store write is deliberately synchronous so an
// admin success response always means the new rule is durable.
func (s *Service) persistSensitiveWordsLocked() error {
	if s.sensitiveWordStore == nil {
		return nil
	}
	return s.sensitiveWordStore.SaveSensitiveWords(context.Background(), append([]SensitiveWord(nil), s.sensitiveWords...))
}

func (s *Service) ContentRiskLogs() []ContentRiskLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]ContentRiskLog(nil), s.riskLogs...)
}

func (s *Service) CheckSensitiveWords(content string) (SensitiveWord, bool) {
	content = strings.TrimSpace(content)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, word := range s.sensitiveWords {
		if word.Status == "active" && strings.Contains(content, word.Word) {
			return word, true
		}
	}
	return SensitiveWord{}, false
}

func (s *Service) AiContentRiskPlaceholder(req ContentRiskCheckRequest) ContentRiskLog {
	content := strings.TrimSpace(req.Content)
	word, ok := s.matchAnySensitiveWord(content)
	status := "clean"
	action := "pass"
	matched := ""
	if ok {
		status = "risk_flagged"
		action = word.Action
		matched = word.Word
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.createRiskLogLocked(ContentRiskLog{
		GameID:   req.GameID,
		RoomID:   req.RoomID,
		SenderID: req.UserID,
		Content:  content,
		Word:     matched,
		Action:   action,
		Status:   status,
		Source:   "ai_placeholder",
	})
}

func (s *Service) matchAnySensitiveWord(content string) (SensitiveWord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, word := range s.sensitiveWords {
		if strings.Contains(content, word.Word) {
			return word, true
		}
	}
	return SensitiveWord{}, false
}

func (s *Service) markMessage(userID int64, roomID int64, messageID int64, action string) (Message, error) {
	room, err := s.RoomForID(userID, roomID)
	if err != nil {
		return Message{}, err
	}
	if s.repo != nil {
		message, ok, err := s.repo.MessageByID(context.Background(), room.ID, messageID)
		if err != nil {
			return Message{}, err
		}
		if !ok {
			return Message{}, ErrMessageNotFound
		}
		if action == "ack" {
			message.AckedBy = appendUniqueInt64(message.AckedBy, userID)
		}
		if action == "read" {
			message.AckedBy = appendUniqueInt64(message.AckedBy, userID)
			message.ReadBy = appendUniqueInt64(message.ReadBy, userID)
		}
		return s.repo.UpdateMessage(context.Background(), message)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	messages := s.messagesByRoom[room.ID]
	for index, message := range messages {
		if message.ID == messageID {
			if action == "ack" {
				message.AckedBy = appendUniqueInt64(message.AckedBy, userID)
			}
			if action == "read" {
				message.AckedBy = appendUniqueInt64(message.AckedBy, userID)
				message.ReadBy = appendUniqueInt64(message.ReadBy, userID)
			}
			messages[index] = message
			s.messagesByRoom[room.ID] = messages
			return message, nil
		}
	}
	return Message{}, ErrMessageNotFound
}

func appendUniqueInt64(values []int64, value int64) []int64 {
	for _, item := range values {
		if item == value {
			return values
		}
	}
	return append(values, value)
}

func visibleMessages(messages []Message) []Message {
	result := make([]Message, 0, len(messages))
	for _, message := range messages {
		if message.Status == "hidden" {
			continue
		}
		result = append(result, message)
	}
	return result
}

func visiblePrivateMessages(messages []PrivateMessage) []PrivateMessage {
	result := make([]PrivateMessage, 0, len(messages))
	for _, message := range messages {
		if message.Status == "hidden" {
			continue
		}
		result = append(result, message)
	}
	return result
}

func privateConversationKey(userID int64, targetUserID int64) string {
	userAID, userBID := orderedPair(userID, targetUserID)
	return strconvInt64(userAID) + ":" + strconvInt64(userBID)
}

func orderedPair(userID int64, targetUserID int64) (int64, int64) {
	if userID < targetUserID {
		return userID, targetUserID
	}
	return targetUserID, userID
}

func strconvInt64(value int64) string {
	return strconv.FormatInt(value, 10)
}

func normalizeSendRequest(req SendRequest) SendRequest {
	req.MessageType = strings.TrimSpace(req.MessageType)
	if req.MessageType == "" {
		req.MessageType = "text"
	}
	req.Content = strings.TrimSpace(req.Content)
	return req
}

func normalizeArchiveReason(reason string) (string, error) {
	reason = strings.TrimSpace(reason)
	if len(reason) > 300 {
		return "", ErrInvalidMessage
	}
	if reason == "" {
		return "archive_job", nil
	}
	return reason, nil
}

func roomReadOnly(room Room) bool {
	return room.Status == "archived" || room.Status == "readonly"
}

func validMessageRequest(req SendRequest) bool {
	switch req.MessageType {
	case "text":
		return req.Content != "" && len(req.Content) <= 1000 && req.FileID == 0
	case "image", "file", "voice":
		return req.FileID > 0 && len(req.Content) <= 300
	case "service_confirm_remind":
		return req.Content != "" && len(req.Content) <= 2000 && req.FileID == 0
	default:
		return false
	}
}

func openIMSyncableMessage(messageType string) bool {
	return messageType == "text" || messageType == "image" || messageType == "voice" || messageType == "file"
}

func validSensitiveWordAction(value string) bool {
	return value == "block" || value == "flag"
}

func validSensitiveWordStatus(value string) bool {
	return value == "active" || value == "disabled"
}

func (s *Service) createRiskLogLocked(log ContentRiskLog) ContentRiskLog {
	log.ID = s.nextRiskID
	log.CreatedAt = time.Now()
	if log.Action == "" {
		log.Action = "block"
	}
	if log.Status == "" {
		log.Status = "blocked"
	}
	if log.Source == "" {
		log.Source = "sensitive_word"
	}
	s.nextRiskID++
	s.riskLogs = append(s.riskLogs, log)
	return log
}
