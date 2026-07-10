package files

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCreateUploadTokenValidatesPolicy(t *testing.T) {
	service := NewService()

	tests := []struct {
		name string
		req  UploadTokenRequest
	}{
		{
			name: "unknown biz type",
			req: UploadTokenRequest{
				BizType:  "export_file",
				FileName: "a.csv",
				MimeType: "text/csv",
				Size:     128,
			},
		},
		{
			name: "avatar too large",
			req: UploadTokenRequest{
				BizType:  "avatar",
				FileName: "avatar.png",
				MimeType: "image/png",
				Size:     5*mib + 1,
			},
		},
		{
			name: "avatar mime not allowed",
			req: UploadTokenRequest{
				BizType:  "avatar",
				FileName: "avatar.gif",
				MimeType: "image/gif",
				Size:     128,
			},
		},
		{
			name: "chat file requires object",
			req: UploadTokenRequest{
				BizType:  "chat_file",
				FileName: "a.png",
				MimeType: "image/png",
				Size:     128,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := service.CreateUploadToken(1, tt.req)
			if !errors.Is(err, ErrInvalidFile) {
				t.Fatalf("expected ErrInvalidFile, got %v", err)
			}
		})
	}
}

func TestCreateUploadTokenAppliesAccessPolicy(t *testing.T) {
	service := NewService()

	_, avatar, err := service.CreateUploadToken(1, UploadTokenRequest{
		BizType:  "avatar",
		FileName: "avatar.png",
		MimeType: "IMAGE/PNG",
		Size:     128,
	})
	if err != nil {
		t.Fatal(err)
	}
	if avatar.AccessLevel != "private" || avatar.MimeType != "image/png" {
		t.Fatalf("expected private normalized avatar, got %+v", avatar)
	}

	_, chatFile, err := service.CreateUploadToken(1, UploadTokenRequest{
		BizType:  "chat_file",
		ObjectID: 99,
		FileName: "a.pdf",
		MimeType: "application/pdf",
		Size:     256,
	})
	if err != nil {
		t.Fatal(err)
	}
	if chatFile.AccessLevel != "game_member" {
		t.Fatalf("expected game member chat file, got %+v", chatFile)
	}

	_, reportAudio, err := service.CreateUploadToken(1, UploadTokenRequest{
		BizType:  "report_attachment",
		FileName: "feedback.m4a",
		MimeType: "audio/mp4",
		Size:     256,
	})
	if err != nil {
		t.Fatal(err)
	}
	if reportAudio.AccessLevel != "private" || reportAudio.MimeType != "audio/mp4" {
		t.Fatalf("expected private report audio, got %+v", reportAudio)
	}
}

func TestServiceUsesConfiguredPublicBaseURLs(t *testing.T) {
	service := NewService()
	service.UsePublicBaseURLs("https://upload.example.com/private/", "https://download.example.com/private/")

	token, file, err := service.CreateUploadToken(1, UploadTokenRequest{
		BizType:  "avatar",
		FileName: "my avatar.png",
		MimeType: "image/png",
		Size:     128,
	})
	if err != nil {
		t.Fatal(err)
	}
	if token.UploadURL != "https://upload.example.com/private/avatar/0/1-my%20avatar.png" {
		t.Fatalf("unexpected upload url for %+v: %s", file, token.UploadURL)
	}
	download, err := service.DownloadURL(file.ID)
	if err != nil {
		t.Fatal(err)
	}
	if download.DownloadURL != "https://download.example.com/private/avatar/0/1-my%20avatar.png" {
		t.Fatalf("unexpected download url: %s", download.DownloadURL)
	}
}

func TestServiceUsesCOSPostFormWhenConfigured(t *testing.T) {
	service := NewService()
	service.UsePublicBaseURLs("https://bucket.cos.ap-nanjing.myqcloud.com", "https://static.example.com")
	service.UseCOSPostSigner(NewCOSPostSigner("AKID-test", "secret-test"))

	token, file, err := service.CreateUploadToken(1, UploadTokenRequest{
		BizType:  "avatar",
		FileName: "avatar.png",
		MimeType: "image/png",
		Size:     128,
	})
	if err != nil {
		t.Fatal(err)
	}
	if token.UploadURL != "https://bucket.cos.ap-nanjing.myqcloud.com" {
		t.Fatalf("expected COS form upload url, got %s", token.UploadURL)
	}
	if token.FormData["key"] != file.StorageKey || token.FormData["q-signature"] == "" || token.FormData["policy"] == "" {
		t.Fatalf("expected COS signed form data, token=%+v file=%+v", token, file)
	}
	if token.Headers["x-zhw-file-id"] != "" {
		t.Fatalf("COS upload should not require custom headers, got %+v", token.Headers)
	}
}

func TestServiceCanRequirePublicBaseURLs(t *testing.T) {
	service := NewService()
	service.RequirePublicBaseURLs(true)

	_, _, err := service.CreateUploadToken(1, UploadTokenRequest{
		BizType:  "avatar",
		FileName: "avatar.png",
		MimeType: "image/png",
		Size:     128,
	})
	if !errors.Is(err, ErrStorageNotConfigured) {
		t.Fatalf("expected ErrStorageNotConfigured for upload, got %v", err)
	}

	service.UsePublicBaseURLs("https://upload.example.com", "")
	token, file, err := service.CreateUploadToken(1, UploadTokenRequest{
		BizType:  "avatar",
		FileName: "avatar.png",
		MimeType: "image/png",
		Size:     128,
	})
	if err != nil {
		t.Fatal(err)
	}
	if token.UploadURL == "" {
		t.Fatalf("expected configured upload URL: %+v", token)
	}
	if _, err := service.DownloadURL(file.ID); !errors.Is(err, ErrStorageNotConfigured) {
		t.Fatalf("expected ErrStorageNotConfigured for download, got %v", err)
	}

	service.UsePublicBaseURLs("https://upload.example.com", "https://download.example.com")
	download, err := service.DownloadURL(file.ID)
	if err != nil {
		t.Fatal(err)
	}
	if download.DownloadURL == "" || strings.HasPrefix(download.DownloadURL, "mock://") {
		t.Fatalf("expected public download URL, got %+v", download)
	}
}

func TestServiceUsesRepositoryWhenConfigured(t *testing.T) {
	repo := &fakeFileRepository{
		files: map[int64]File{
			9: {
				ID:          9,
				UploaderID:  2,
				BizType:     "chat_file",
				ObjectID:    10,
				FileName:    "a.png",
				MimeType:    "image/png",
				Size:        128,
				StorageKey:  "chat_file/10/9-a.png",
				AccessLevel: "game_member",
				CreatedAt:   time.Now(),
			},
		},
	}
	service := NewServiceWithRepository(repo)

	token, file, err := service.CreateUploadToken(2, UploadTokenRequest{
		BizType:  "chat_file",
		ObjectID: 10,
		FileName: "a.png",
		MimeType: "image/png",
		Size:     128,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !repo.saved || file.ID != 10 || token.FileID != 10 || token.Headers["x-zhw-file-id"] != "10" {
		t.Fatalf("expected repository upload file, saved=%v file=%+v token=%+v", repo.saved, file, token)
	}

	generated, err := service.CreateGeneratedFile("export_file", 3, "report.csv", "text/csv", 64)
	if err != nil {
		t.Fatal(err)
	}
	if generated.ID != 11 || generated.AccessLevel != "admin_permission" {
		t.Fatalf("expected repository generated file, got %+v", generated)
	}

	got, err := service.Get(9)
	if err != nil {
		t.Fatal(err)
	}
	if !repo.found || got.ID != 9 {
		t.Fatalf("expected repository get, found=%v file=%+v", repo.found, got)
	}
	url, err := service.DownloadURL(9)
	if err != nil {
		t.Fatal(err)
	}
	if url.FileID != 9 || url.DownloadURL == "" {
		t.Fatalf("expected repository download url, got %+v", url)
	}
}

type fakeFileRepository struct {
	files map[int64]File
	saved bool
	found bool
	next  int64
}

func (r *fakeFileRepository) SaveFile(ctx context.Context, file File) (File, error) {
	r.saved = true
	if r.next == 0 {
		r.next = 10
	}
	file.ID = r.next
	r.next++
	if file.CreatedAt.IsZero() {
		file.CreatedAt = time.Now()
	}
	r.files[file.ID] = file
	return file, nil
}

func (r *fakeFileRepository) FindFile(ctx context.Context, fileID int64) (File, bool, error) {
	r.found = true
	file, ok := r.files[fileID]
	return file, ok, nil
}
