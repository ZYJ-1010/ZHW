package im

import "testing"

func TestVoiceMessageUsesFileAttachment(t *testing.T) {
	state := &readonlyGameState{members: []int64{1, 2}}
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
