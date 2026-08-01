package im

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestAdminRetryCreateRoomUsesRepositoryRoom(t *testing.T) {
	repo := &retryRoomRepository{
		room: Room{
			ID:        12,
			GameID:    92012,
			MemberIDs: []int64{10001, 10002},
			Status:    "normal",
			Engine:    "local",
			CreatedAt: time.Now(),
		},
	}
	service := NewService(nil)
	service.UseRepository(repo)

	room, err := service.AdminRetryCreateRoom(12)
	if err != nil {
		t.Fatalf("AdminRetryCreateRoom returned error: %v", err)
	}
	if room.ID != 12 || room.GameID != 92012 {
		t.Fatalf("unexpected room: %+v", room)
	}
}

func TestEnsureRoomReusesPersistedOpenIMGroup(t *testing.T) {
	repo := &retryRoomRepository{
		room: Room{
			ID:            12,
			GameID:        92012,
			MemberIDs:     []int64{10001, 10002},
			Status:        "active",
			Engine:        "openim",
			OpenIMGroupID: "zhw_game_92012",
			CreatedAt:     time.Now(),
		},
	}
	service := NewServiceWithOpenIM(&readonlyGameState{members: []int64{10001, 10002}, roomReady: true}, OpenIMConfig{
		Enabled:     true,
		APIAddr:     "http://127.0.0.1:1",
		Secret:      "secret",
		AdminUserID: "admin",
	})
	service.UseRepository(repo)

	room := service.EnsureRoom(92012)
	if room.Status != "active" || room.Engine != "openim" || room.OpenIMGroupID != "zhw_game_92012" {
		t.Fatalf("persisted OpenIM room was recreated or degraded: %+v", room)
	}
}

func TestArchiveRoomsByGameIDsPersistsWithRepository(t *testing.T) {
	repo := &retryRoomRepository{room: Room{ID: 12, GameID: 92012, Status: "active", CreatedAt: time.Now()}}
	service := NewService(nil)
	service.UseRepository(repo)
	rooms := service.ArchiveRoomsByGameIDs([]int64{92012}, "game ended")
	if len(rooms) != 1 || rooms[0].Status != "archived" {
		t.Fatalf("expected archived room result, got %+v", rooms)
	}
	if repo.room.Status != "archived" || repo.room.ArchiveReason != "game ended" {
		t.Fatalf("repository room was not archived: %+v", repo.room)
	}
}

func TestStrictIMReadsDoNotUseProcessCacheOnRepositoryFailure(t *testing.T) {
	repo := &retryRoomRepository{roomErr: errors.New("room database unavailable"), listErr: errors.New("list database unavailable")}
	service := NewServiceWithOpenIM(&readonlyGameState{members: []int64{1}, roomReady: true}, OpenIMConfig{})
	service.UseRepository(repo)
	service.roomsByGame[99] = Room{ID: 99, GameID: 99, Status: "active"}
	service.messagesByRoom[99] = []Message{{ID: 9, RoomID: 99, GameID: 99, Content: "旧消息"}}

	if _, err := service.EnsureRoomStrict(99); !errors.Is(err, repo.roomErr) {
		t.Fatalf("expected room read error without local room fallback, got %v", err)
	}
	if _, err := service.AdminRoomsStrict(); !errors.Is(err, repo.listErr) {
		t.Fatalf("expected room list error without local room fallback, got %v", err)
	}
	if _, err := service.AllMessagesStrict(); !errors.Is(err, repo.listErr) {
		t.Fatalf("expected message list error without local message fallback, got %v", err)
	}
}

func TestRepositoryMarkReadPersistsMessageState(t *testing.T) {
	repo := &retryRoomRepository{
		room:     Room{ID: 12, GameID: 92012, Status: "active", MemberIDs: []int64{1}, CreatedAt: time.Now()},
		messages: []Message{{ID: 7, RoomID: 12, GameID: 92012, SenderID: 2, Status: "sent"}},
	}
	service := NewServiceWithOpenIM(&readonlyGameState{members: []int64{1}, roomReady: true}, OpenIMConfig{})
	service.UseRepository(repo)

	message, err := service.MarkRead(1, 12, 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(message.ReadBy) != 1 || message.ReadBy[0] != 1 || len(repo.updatedMessage.ReadBy) != 1 {
		t.Fatalf("expected read state persisted, message=%+v persisted=%+v", message, repo.updatedMessage)
	}
}

type retryRoomRepository struct {
	room           Room
	messages       []Message
	updatedMessage Message
	roomErr        error
	listErr        error
}

func (r *retryRoomRepository) EnsureRoom(context.Context, int64, []int64, string, string) (Room, error) {
	return r.room, nil
}

func (r *retryRoomRepository) RoomByGame(_ context.Context, gameID int64) (Room, bool, error) {
	if r.roomErr != nil {
		return Room{}, false, r.roomErr
	}
	return r.room, r.room.GameID == gameID, nil
}

func (r *retryRoomRepository) RoomByID(_ context.Context, roomID int64) (Room, bool, error) {
	if r.roomErr != nil {
		return Room{}, false, r.roomErr
	}
	return r.room, r.room.ID == roomID, nil
}

func (r *retryRoomRepository) ListRooms(context.Context) ([]Room, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return []Room{r.room}, nil
}

func (r *retryRoomRepository) SaveRoom(_ context.Context, room Room) (Room, error) {
	r.room = room
	return room, nil
}

func (r *retryRoomRepository) SaveMessage(_ context.Context, message Message) (Message, error) {
	return message, nil
}

func (r *retryRoomRepository) MessageByID(_ context.Context, roomID int64, messageID int64) (Message, bool, error) {
	for _, message := range r.messages {
		if message.RoomID == roomID && message.ID == messageID {
			return message, true, nil
		}
	}
	return Message{}, false, nil
}

func (r *retryRoomRepository) ListMessagesByRoom(context.Context, int64) ([]Message, error) {
	return nil, nil
}

func (r *retryRoomRepository) ListMessages(context.Context) ([]Message, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return append([]Message(nil), r.messages...), nil
}

func (r *retryRoomRepository) UpdateMessage(_ context.Context, message Message) (Message, error) {
	r.updatedMessage = message
	return message, nil
}

func (r *retryRoomRepository) EnsurePrivateConversation(context.Context, int64, int64, int64) (PrivateConversation, error) {
	return PrivateConversation{}, nil
}

func (r *retryRoomRepository) SavePrivateMessage(_ context.Context, message PrivateMessage) (PrivateMessage, error) {
	return message, nil
}

func (r *retryRoomRepository) ListPrivateMessages(context.Context, int64) ([]PrivateMessage, error) {
	return nil, nil
}
