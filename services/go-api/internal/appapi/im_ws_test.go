package appapi

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

func TestIMWebSocketFlow(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	creatorToken := loginForTestWithCode(t, mux, "im-ws-creator")
	completeIdentityForTest(t, mux, creatorToken)
	playerToken := loginForTestWithCode(t, mux, "im-ws-player")
	completeIdentityForTest(t, mux, playerToken)

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"IM websocket game","gameType":"free","minPlayers":5,"maxPlayers":8}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/applications", playerToken, `{"reason":"join"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/applications/1/review", creatorToken, `{"approve":true}`, http.StatusOK)
	approveExtraMembersForHTTP(t, mux, creatorToken, 1, "im-ws", 3)
	postJSON(t, mux, "/api/app/games/1/manual-start", creatorToken, `{}`, http.StatusOK)

	conn, reader := openIMWebSocket(t, server, playerToken, 1, http.StatusSwitchingProtocols)
	defer conn.Close()
	connected := readWSJSON(t, reader)
	if connected.Type != "connected" {
		t.Fatalf("expected connected frame, got %+v", connected)
	}

	writeClientWSJSON(t, conn, map[string]interface{}{
		"type":    "send_message",
		"payload": map[string]interface{}{"messageType": "text", "content": "hello websocket"},
	})
	message := readWSJSON(t, reader)
	if message.Type != "message" {
		t.Fatalf("expected message frame, got %+v", message)
	}
	messageData := marshalMap(t, message.Data)
	if messageData["content"] != "hello websocket" || messageData["messageType"] != "text" {
		t.Fatalf("expected websocket text message, got %+v", messageData)
	}

	uploadBody := postJSON(t, mux, "/api/app/files/upload-token", playerToken, `{"bizType":"chat_file","objectId":1,"fileName":"a.png","mimeType":"image/png","size":128}`, http.StatusOK)
	var uploadResp struct {
		Data struct {
			Upload struct {
				FileID int64 `json:"fileId"`
			} `json:"upload"`
		} `json:"data"`
	}
	if err := json.Unmarshal(uploadBody, &uploadResp); err != nil {
		t.Fatal(err)
	}
	writeClientWSJSON(t, conn, map[string]interface{}{
		"type":    "send_message",
		"payload": map[string]interface{}{"messageType": "image", "fileId": uploadResp.Data.Upload.FileID, "content": "a.png"},
	})
	imageMessage := readWSJSON(t, reader)
	imageData := marshalMap(t, imageMessage.Data)
	if imageMessage.Type != "message" || int64FromJSONNumber(imageData["fileId"]) != uploadResp.Data.Upload.FileID {
		t.Fatalf("expected websocket image/file message, got %+v", imageMessage)
	}

	outsiderToken := loginForTestWithCode(t, mux, "im-ws-outsider")
	completeIdentityForTest(t, mux, outsiderToken)
	outsiderConn, _ := openIMWebSocket(t, server, outsiderToken, 1, http.StatusForbidden)
	if outsiderConn != nil {
		_ = outsiderConn.Close()
	}

	postJSON(t, mux, "/api/app/chat/rooms/1/archive", creatorToken, `{"reason":"completed_without_dispute"}`, http.StatusOK)
	writeClientWSJSON(t, conn, map[string]interface{}{
		"type":    "send_message",
		"payload": map[string]interface{}{"messageType": "text", "content": "after archive"},
	})
	readonlyError := readWSJSON(t, reader)
	if readonlyError.Type != "error" || readonlyError.Code == 0 {
		t.Fatalf("expected readonly error after archive, got %+v", readonlyError)
	}

	historyBody := getJSON(t, mux, "/api/app/games/1/chat/messages", playerToken, http.StatusOK)
	if !bytes.Contains(historyBody, []byte("hello websocket")) || !bytes.Contains(historyBody, []byte(`"fileId":`+strconv.FormatInt(uploadResp.Data.Upload.FileID, 10))) {
		t.Fatalf("expected websocket messages persisted in history: %s", string(historyBody))
	}
}

type wsTestFrame struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
	Code int         `json:"code"`
}

func openIMWebSocket(t *testing.T, server *httptest.Server, token string, gameID int64, expectedStatus int) (net.Conn, *bufio.Reader) {
	t.Helper()
	address := strings.TrimPrefix(server.URL, "http://")
	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	key := base64.StdEncoding.EncodeToString([]byte("zhw-mini-ws-test"))
	request := "GET /api/app/im/ws?gameId=" + strconv.FormatInt(gameID, 10) + " HTTP/1.1\r\n" +
		"Host: " + address + "\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Key: " + key + "\r\n" +
		"Sec-WebSocket-Version: 13\r\n" +
		"Authorization: Bearer " + token + "\r\n\r\n"
	if _, err := conn.Write([]byte(request)); err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewReader(conn)
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(statusLine, strconv.Itoa(expectedStatus)) {
		t.Fatalf("expected websocket status %d, got %s", expectedStatus, strings.TrimSpace(statusLine))
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if line == "\r\n" {
			break
		}
	}
	if expectedStatus != http.StatusSwitchingProtocols {
		return conn, reader
	}
	return conn, reader
}

func readWSJSON(t *testing.T, reader *bufio.Reader) wsTestFrame {
	t.Helper()
	opcode, payload, err := readWSFrame(reader)
	if err != nil {
		t.Fatal(err)
	}
	if opcode != 0x1 {
		t.Fatalf("expected text websocket frame, got opcode %d", opcode)
	}
	var frame wsTestFrame
	if err := json.Unmarshal(payload, &frame); err != nil {
		t.Fatalf("invalid websocket json %s: %v", string(payload), err)
	}
	return frame
}

func writeClientWSJSON(t *testing.T, conn net.Conn, payload interface{}) {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeClientWSFrame(conn, 0x1, data); err != nil {
		t.Fatal(err)
	}
}

func writeClientWSFrame(writer net.Conn, opcode byte, payload []byte) error {
	header := []byte{0x80 | opcode}
	length := len(payload)
	switch {
	case length < 126:
		header = append(header, 0x80|byte(length))
	case length <= 65535:
		header = append(header, 0x80|126, byte(length>>8), byte(length))
	default:
		header = append(header, 0x80|127)
		var extended [8]byte
		binary.BigEndian.PutUint64(extended[:], uint64(length))
		header = append(header, extended[:]...)
	}
	mask := [4]byte{1, 2, 3, 4}
	header = append(header, mask[:]...)
	masked := append([]byte(nil), payload...)
	for i := range masked {
		masked[i] ^= mask[i%4]
	}
	if _, err := writer.Write(header); err != nil {
		return err
	}
	_, err := writer.Write(masked)
	return err
}

func marshalMap(t *testing.T, value interface{}) map[string]interface{} {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func int64FromJSONNumber(value interface{}) int64 {
	number, _ := value.(float64)
	return int64(number)
}
