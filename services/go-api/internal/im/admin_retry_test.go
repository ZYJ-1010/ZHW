package im

import (
	"context"
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

type retryRoomRepository struct {
	room Room
}

func (r *retryRoomRepository) EnsureRoom(context.Context, int64, []int64, string, string) (Room, error) {
	return r.room, nil
}

func (r *retryRoomRepository) RoomByGame(_ context.Context, gameID int64) (Room, bool, error) {
	return r.room, r.room.GameID == gameID, nil
}

func (r *retryRoomRepository) RoomByID(_ context.Context, roomID int64) (Room, bool, error) {
	return r.room, r.room.ID == roomID, nil
}

func (r *retryRoomRepository) ListRooms(context.Context) ([]Room, error) {
	return []Room{r.room}, nil
}

func (r *retryRoomRepository) SaveRoom(_ context.Context, room Room) (Room, error) {
	r.room = room
	return room, nil
}

func (r *retryRoomRepository) SaveMessage(_ context.Context, message Message) (Message, error) {
	return message, nil
}

func (r *retryRoomRepository) MessageByID(context.Context, int64, int64) (Message, bool, error) {
	return Message{}, false, nil
}

func (r *retryRoomRepository) ListMessagesByRoom(context.Context, int64) ([]Message, error) {
	return nil, nil
}

func (r *retryRoomRepository) ListMessages(context.Context) ([]Message, error) {
	return nil, nil
}

func (r *retryRoomRepository) UpdateMessage(_ context.Context, message Message) (Message, error) {
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
