package im

import (
	"context"
	"testing"
)

type memorySensitiveWordStore struct {
	words []SensitiveWord
}

func (s *memorySensitiveWordStore) LoadSensitiveWords(context.Context) ([]SensitiveWord, error) {
	return append([]SensitiveWord(nil), s.words...), nil
}

func (s *memorySensitiveWordStore) SaveSensitiveWords(_ context.Context, words []SensitiveWord) error {
	s.words = append([]SensitiveWord(nil), words...)
	return nil
}

func TestSensitiveWordsPersistAcrossServiceRestart(t *testing.T) {
	store := &memorySensitiveWordStore{}
	first := NewService(&readonlyGameState{members: []int64{1}})
	if err := first.UseSensitiveWordStore(store); err != nil {
		t.Fatal(err)
	}
	created, err := first.CreateSensitiveWord(SensitiveWordRequest{Word: "持久化词"})
	if err != nil {
		t.Fatal(err)
	}
	if len(store.words) != 2 {
		t.Fatalf("expected default and created words to be stored, got %d", len(store.words))
	}
	if _, err := first.UpdateSensitiveWord(created.ID, UpdateSensitiveWordRequest{Status: "disabled"}); err != nil {
		t.Fatal(err)
	}

	second := NewService(&readonlyGameState{members: []int64{1}})
	if err := second.UseSensitiveWordStore(store); err != nil {
		t.Fatal(err)
	}
	if _, ok := second.CheckSensitiveWords("命中持久化词"); ok {
		t.Fatal("disabled persisted word must not block messages")
	}
	if _, ok := second.CheckSensitiveWords("敏感词"); !ok {
		t.Fatal("persisted dictionary should replace the in-memory default")
	}
}

func TestGroupIMSendUsesSensitiveWordDictionary(t *testing.T) {
	store := &memorySensitiveWordStore{}
	service := NewService(&readonlyGameState{members: []int64{1}, roomReady: true})
	if err := service.UseSensitiveWordStore(store); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateSensitiveWord(SensitiveWordRequest{Word: "群聊拦截词", Action: "block", Status: "active"}); err != nil {
		t.Fatal(err)
	}
	service.EnsureRoom(99)
	if _, err := service.Send(1, 99, SendRequest{MessageType: "text", Content: "这里有群聊拦截词"}); err != ErrSensitive {
		t.Fatalf("group IM should reject sensitive content, got %v", err)
	}
}

func TestVoiceMessageUsesFileAttachment(t *testing.T) {
	state := &readonlyGameState{members: []int64{1, 2}, roomReady: true}
	service := NewService(state)
	message, err := service.Send(1, 99, SendRequest{MessageType: "voice", Content: "voice.mp3", FileID: 7})
	if err != nil {
		t.Fatalf("voice message should be accepted: %v", err)
	}
	if message.Type != "voice" || message.FileID != 7 || message.Content != "voice.mp3" {
		t.Fatalf("unexpected voice message: %+v", message)
	}
}

func TestPrivateChatStillAcceptsTextOnly(t *testing.T) {
	state := &readonlyGameState{members: []int64{1, 2}}
	service := NewService(state)
	if _, _, err := service.SendPrivate(1, 2, 99, SendRequest{MessageType: "voice", Content: "voice.mp3", FileID: 7}); err != ErrInvalidMessage {
		t.Fatalf("private voice message should remain disabled, got %v", err)
	}
}
