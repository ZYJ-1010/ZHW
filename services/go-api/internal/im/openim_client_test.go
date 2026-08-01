package im

import "testing"

func TestOpenIMMessageContentMapsSupportedMessages(t *testing.T) {
	tests := []struct {
		name        string
		request     SendRequest
		contentType int
		field       string
	}{
		{name: "text", request: SendRequest{MessageType: "text", Content: "你好"}, contentType: openIMTextContentType, field: "content"},
		{name: "image", request: SendRequest{MessageType: "image", FileID: 7, Width: 640, Height: 480, Attachment: &MessageAttachment{URL: "https://static.example/image.png", FileName: "image.png", MimeType: "image/png", Size: 128}}, contentType: openIMPictureContentType, field: "sourcePicture"},
		{name: "voice", request: SendRequest{MessageType: "voice", FileID: 8, DurationMS: 3200, Attachment: &MessageAttachment{URL: "https://static.example/voice.mp3", FileName: "voice.mp3", MimeType: "audio/mpeg", Size: 256}}, contentType: openIMVoiceContentType, field: "sourceUrl"},
		{name: "file", request: SendRequest{MessageType: "file", FileID: 9, Attachment: &MessageAttachment{URL: "https://static.example/document.pdf", FileName: "document.pdf", MimeType: "application/pdf", Size: 512}}, contentType: openIMFileContentType, field: "sourceUrl"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contentType, content, err := openIMMessageContent(test.request)
			if err != nil {
				t.Fatal(err)
			}
			if contentType != test.contentType {
				t.Fatalf("content type = %d, want %d", contentType, test.contentType)
			}
			if _, ok := content[test.field]; !ok {
				t.Fatalf("content field %q missing: %+v", test.field, content)
			}
			if test.name == "voice" && content["duration"] != int64(3200) {
				t.Fatalf("voice duration = %v, want 3200 milliseconds", content["duration"])
			}
		})
	}
}

func TestOpenIMMediaMessageRequiresAttachment(t *testing.T) {
	if _, _, err := openIMMessageContent(SendRequest{MessageType: "image", FileID: 7}); err != ErrInvalidMessage {
		t.Fatalf("missing media attachment error = %v, want ErrInvalidMessage", err)
	}
}

func TestOpenIMSyncableMessageExcludesLocalSystemCards(t *testing.T) {
	if !openIMSyncableMessage("text") || !openIMSyncableMessage("image") || !openIMSyncableMessage("voice") || !openIMSyncableMessage("file") {
		t.Fatal("supported member messages must be synchronized")
	}
	if openIMSyncableMessage("service_confirm_remind") {
		t.Fatal("local system card must not be sent with an unsupported OpenIM payload")
	}
}
