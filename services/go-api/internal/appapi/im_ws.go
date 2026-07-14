package appapi

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/files"
	"zhw-mini/services/go-api/internal/im"
)

const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

type imSocketHub struct {
	server *Server
	mu     sync.RWMutex
	rooms  map[int64]map[*imSocketClient]struct{}
}

type imSocketClient struct {
	hub     *imSocketHub
	conn    net.Conn
	reader  *bufio.Reader
	writeMu sync.Mutex
	userID  int64
	gameID  int64
	roomID  int64
}

type imSocketIncoming struct {
	Type        string         `json:"type"`
	RequestID   string         `json:"requestId"`
	MessageType string         `json:"messageType"`
	Content     string         `json:"content"`
	FileID      int64          `json:"fileId"`
	Payload     im.SendRequest `json:"payload"`
}

type imSocketOutgoing struct {
	Type      string      `json:"type"`
	RequestID string      `json:"requestId,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Code      int         `json:"code,omitempty"`
	Message   string      `json:"message,omitempty"`
}

func newIMSocketHub(server *Server) *imSocketHub {
	return &imSocketHub{
		server: server,
		rooms:  make(map[int64]map[*imSocketClient]struct{}),
	}
}

func (s *Server) imSocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "method not allowed")
		return
	}
	userID, ok := s.socketUserID(r)
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "未登录")
		return
	}
	gameID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("gameId")), 10, 64)
	if err != nil || gameID <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "gameId invalid")
		return
	}
	room, err := s.im.RoomForGame(userID, gameID)
	if err != nil {
		writeIMError(w, err)
		return
	}
	conn, reader, err := upgradeWebSocket(w, r)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "websocket upgrade failed")
		return
	}
	client := &imSocketClient{
		hub:    s.imSocketHub,
		conn:   conn,
		reader: reader,
		userID: userID,
		gameID: gameID,
		roomID: room.ID,
	}
	s.imSocketHub.add(client)
	s.recordBehavior(userID, "enter_im_ws", "game", gameID, map[string]interface{}{"roomId": room.ID})
	game, _ := s.games.Get(gameID)
	_ = client.writeJSON(imSocketOutgoing{Type: "connected", Data: s.chatRoomPayload(room, userID, game)})
	client.readLoop()
}

func (s *Server) socketUserID(r *http.Request) (int64, bool) {
	token := bearerToken(r.Header.Get("Authorization"))
	if token == "" {
		token = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	if token == "" {
		return 0, false
	}
	user, ok := s.auth.CurrentUser(token)
	if !ok {
		return 0, false
	}
	return user.ID, true
}

func (h *imSocketHub) add(client *imSocketClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[client.roomID] == nil {
		h.rooms[client.roomID] = make(map[*imSocketClient]struct{})
	}
	h.rooms[client.roomID][client] = struct{}{}
}

func (h *imSocketHub) remove(client *imSocketClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	clients := h.rooms[client.roomID]
	if clients == nil {
		return
	}
	delete(clients, client)
	if len(clients) == 0 {
		delete(h.rooms, client.roomID)
	}
}

func (h *imSocketHub) broadcast(roomID int64, payload imSocketOutgoing) {
	h.mu.RLock()
	clients := make([]*imSocketClient, 0, len(h.rooms[roomID]))
	for client := range h.rooms[roomID] {
		clients = append(clients, client)
	}
	h.mu.RUnlock()
	for _, client := range clients {
		_ = client.writeJSON(payload)
	}
}

func (c *imSocketClient) readLoop() {
	defer func() {
		c.hub.remove(c)
		_ = c.conn.Close()
	}()
	for {
		opcode, payload, err := readWSFrame(c.reader)
		if err != nil {
			return
		}
		switch opcode {
		case 0x1:
			c.handleText(payload)
		case 0x8:
			_ = c.writeFrame(0x8, nil)
			return
		case 0x9:
			_ = c.writeFrame(0xA, payload)
		}
	}
}

func (c *imSocketClient) handleText(payload []byte) {
	var req imSocketIncoming
	if err := json.Unmarshal(payload, &req); err != nil {
		_ = c.writeJSON(imSocketOutgoing{Type: "error", Code: httpx.CodeValidationError, Message: "消息格式错误"})
		return
	}
	if req.Type != "" && req.Type != "send_message" && req.Type != "message" {
		_ = c.writeJSON(imSocketOutgoing{Type: "error", RequestID: req.RequestID, Code: httpx.CodeValidationError, Message: "消息类型不支持"})
		return
	}
	sendReq := req.Payload
	if sendReq.MessageType == "" {
		sendReq.MessageType = req.MessageType
	}
	if sendReq.Content == "" {
		sendReq.Content = req.Content
	}
	if sendReq.FileID == 0 {
		sendReq.FileID = req.FileID
	}
	if err := c.hub.server.validateChatMessageFileForSocket(sendReq, c.gameID); err != nil {
		_ = c.writeJSON(socketError(req.RequestID, err))
		return
	}
	message, err := c.hub.server.im.Send(c.userID, c.gameID, sendReq)
	if err != nil {
		_ = c.writeJSON(socketError(req.RequestID, err))
		return
	}
	c.hub.server.recordBehavior(c.userID, "send_message_ws", "game", c.gameID, map[string]interface{}{"messageId": message.ID, "messageType": message.Type})
	c.hub.server.createIMMessageNotifications(c.userID, message)
	c.hub.broadcast(message.RoomID, imSocketOutgoing{Type: "message", RequestID: req.RequestID, Data: c.hub.server.inGameMessageDTO(message)})
}

func (s *Server) validateChatMessageFileForSocket(req im.SendRequest, gameID int64) error {
	if req.MessageType != "image" && req.MessageType != "file" {
		return nil
	}
	file, err := s.files.Get(req.FileID)
	if err != nil {
		if errors.Is(err, files.ErrFileNotFound) {
			return files.ErrFileNotFound
		}
		return err
	}
	if file.BizType != "chat_file" || file.ObjectID != gameID {
		return im.ErrForbidden
	}
	return nil
}

func socketError(requestID string, err error) imSocketOutgoing {
	switch {
	case errors.Is(err, im.ErrForbidden):
		return imSocketOutgoing{Type: "error", RequestID: requestID, Code: 40331, Message: "非局成员不能访问 IM"}
	case errors.Is(err, files.ErrFileNotFound):
		return imSocketOutgoing{Type: "error", RequestID: requestID, Code: httpx.CodeNotFound, Message: "file not found"}
	case errors.Is(err, im.ErrRoomNotFound):
		return imSocketOutgoing{Type: "error", RequestID: requestID, Code: 40431, Message: "IM 房间不存在"}
	case errors.Is(err, im.ErrSensitive):
		return imSocketOutgoing{Type: "error", RequestID: requestID, Code: 45101, Message: "消息命中敏感词"}
	case errors.Is(err, im.ErrInvalidMessage):
		return imSocketOutgoing{Type: "error", RequestID: requestID, Code: httpx.CodeValidationError, Message: "IM 消息参数错误"}
	default:
		return imSocketOutgoing{Type: "error", RequestID: requestID, Code: httpx.CodeSystemError, Message: "IM 操作失败"}
	}
}

func (c *imSocketClient) writeJSON(payload imSocketOutgoing) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.writeFrame(0x1, data)
}

func (c *imSocketClient) writeFrame(opcode byte, payload []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return writeWSFrame(c.conn, opcode, payload)
}

func upgradeWebSocket(w http.ResponseWriter, r *http.Request) (net.Conn, *bufio.Reader, error) {
	if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") || !strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade") {
		return nil, nil, errors.New("not websocket")
	}
	key := strings.TrimSpace(r.Header.Get("Sec-WebSocket-Key"))
	if key == "" {
		return nil, nil, errors.New("missing websocket key")
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack unsupported")
	}
	conn, rw, err := hijacker.Hijack()
	if err != nil {
		return nil, nil, err
	}
	accept := websocketAcceptKey(key)
	_, err = fmt.Fprintf(rw, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: %s\r\n\r\n", accept)
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}
	if err := rw.Flush(); err != nil {
		_ = conn.Close()
		return nil, nil, err
	}
	return conn, rw.Reader, nil
}

func websocketAcceptKey(key string) string {
	sum := sha1.Sum([]byte(key + websocketGUID))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func readWSFrame(reader *bufio.Reader) (byte, []byte, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(reader, header); err != nil {
		return 0, nil, err
	}
	opcode := header[0] & 0x0F
	masked := header[1]&0x80 != 0
	length := uint64(header[1] & 0x7F)
	switch length {
	case 126:
		extended := make([]byte, 2)
		if _, err := io.ReadFull(reader, extended); err != nil {
			return 0, nil, err
		}
		length = uint64(binary.BigEndian.Uint16(extended))
	case 127:
		extended := make([]byte, 8)
		if _, err := io.ReadFull(reader, extended); err != nil {
			return 0, nil, err
		}
		length = binary.BigEndian.Uint64(extended)
	}
	if length > 1<<20 {
		return 0, nil, errors.New("websocket payload too large")
	}
	var maskKey [4]byte
	if masked {
		if _, err := io.ReadFull(reader, maskKey[:]); err != nil {
			return 0, nil, err
		}
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return 0, nil, err
	}
	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}
	return opcode, payload, nil
}

func writeWSFrame(writer io.Writer, opcode byte, payload []byte) error {
	header := []byte{0x80 | opcode}
	length := len(payload)
	switch {
	case length < 126:
		header = append(header, byte(length))
	case length <= 65535:
		header = append(header, 126, byte(length>>8), byte(length))
	default:
		header = append(header, 127)
		var extended [8]byte
		binary.BigEndian.PutUint64(extended[:], uint64(length))
		header = append(header, extended[:]...)
	}
	if _, err := writer.Write(header); err != nil {
		return err
	}
	if length == 0 {
		return nil
	}
	_, err := writer.Write(payload)
	return err
}
