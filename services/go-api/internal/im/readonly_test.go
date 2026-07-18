package im

import (
	"errors"
	"testing"
)

type readonlyGameState struct {
	members   []int64
	readonly  bool
	roomReady bool
}

func (s *readonlyGameState) Members(gameID int64) []int64 {
	return append([]int64(nil), s.members...)
}

func (s *readonlyGameState) IsMember(gameID int64, userID int64) bool {
	for _, memberID := range s.members {
		if memberID == userID {
			return true
		}
	}
	return false
}

func (s *readonlyGameState) IsIMReadOnly(gameID int64) bool {
	return s.readonly
}

func (s *readonlyGameState) IsIMRoomReady(gameID int64) bool {
	return s.roomReady
}

func TestEndedGameRoomIsReadableButRejectsNewMessages(t *testing.T) {
	state := &readonlyGameState{members: []int64{1, 2}, roomReady: true}
	service := NewService(state)
	room := service.EnsureRoom(99)
	if room.Status != "active" {
		t.Fatalf("initial room status = %q", room.Status)
	}

	state.readonly = true
	room, err := service.RoomForGame(1, 99)
	if err != nil {
		t.Fatal(err)
	}
	if room.Status != "readonly" {
		t.Fatalf("ended game room status = %q, want readonly", room.Status)
	}
	if _, err := service.Messages(1, 99); err != nil {
		t.Fatalf("history should remain readable: %v", err)
	}
	if _, err := service.Send(1, 99, SendRequest{MessageType: "text", Content: "after end"}); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("send after end error = %v, want ErrInvalidMessage", err)
	}
	if _, err := service.SendSystem(1, 99, SendRequest{MessageType: "service_confirm_remind", Content: `{"title":"待确认"}`}); err != nil {
		t.Fatalf("completion system reminder should be written after room becomes read-only: %v", err)
	}
	messages, err := service.Messages(1, 99)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 || messages[0].Type != "service_confirm_remind" {
		t.Fatalf("completion system reminder missing: %+v", messages)
	}
	if _, err := service.SendSystem(1, 99, SendRequest{MessageType: "text", Content: "system bypass"}); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("system bypass accepted a regular text message: %v", err)
	}
}

func TestRecruitingGameDoesNotCreateRoomOnAccess(t *testing.T) {
	state := &readonlyGameState{members: []int64{1, 2}, roomReady: false}
	service := NewService(state)
	if _, err := service.RoomForGame(1, 99); !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("recruiting game access error = %v, want ErrRoomNotFound", err)
	}
	if rooms := service.AdminRooms(); len(rooms) != 0 {
		t.Fatalf("recruiting game access created rooms: %+v", rooms)
	}
}
