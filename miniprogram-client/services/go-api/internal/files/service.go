package files

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	ErrFileNotFound         = errors.New("file not found")
	ErrInvalidFile          = errors.New("invalid file")
	ErrFileExpired          = errors.New("file expired")
	ErrStorageNotConfigured = errors.New("storage base url not configured")
)

const mib int64 = 1024 * 1024

type uploadPolicy struct {
	maxSize       int64
	mimeTypes     map[string]struct{}
	accessLevel   string
	requireObject bool
}

var uploadPolicies = map[string]uploadPolicy{
	"avatar": {
		maxSize:     5 * mib,
		accessLevel: "private",
		mimeTypes: allowMIMEs(
			"image/jpeg",
			"image/png",
			"image/webp",
		),
	},
	"chat_file": {
		maxSize:       20 * mib,
		accessLevel:   "game_member",
		requireObject: true,
		mimeTypes: allowMIMEs(
			"image/jpeg",
			"image/png",
			"image/webp",
			"application/pdf",
			"text/plain",
			"application/zip",
		),
	},
	"realname_material": {
		maxSize:     10 * mib,
		accessLevel: "private",
		mimeTypes: allowMIMEs(
			"image/jpeg",
			"image/png",
			"image/webp",
			"application/pdf",
			"audio/mpeg",
			"audio/mp4",
			"audio/aac",
			"audio/amr",
			"audio/wav",
		),
	},
	"report_attachment": {
		maxSize:     20 * mib,
		accessLevel: "private",
		mimeTypes: allowMIMEs(
			"image/jpeg",
			"image/png",
			"image/webp",
			"application/pdf",
		),
	},
	"game_application": {
		maxSize:       20 * mib,
		accessLevel:   "private",
		requireObject: true,
		mimeTypes: allowMIMEs(
			"image/jpeg",
			"image/png",
			"image/webp",
			"application/pdf",
		),
	},
	"delivery_proof": {
		maxSize:       20 * mib,
		accessLevel:   "game_member",
		requireObject: true,
		mimeTypes: allowMIMEs(
			"image/jpeg",
			"image/png",
			"image/webp",
			"application/pdf",
		),
	},
}

type UploadTokenRequest struct {
	BizType  string `json:"bizType"`
	ObjectID int64  `json:"objectId"`
	FileName string `json:"fileName"`
	MimeType string `json:"mimeType"`
	Size     int64  `json:"size"`
	SHA256   string `json:"sha256"`
}

type File struct {
	ID          int64     `json:"fileId"`
	UploaderID  int64     `json:"uploaderUserId"`
	BizType     string    `json:"bizType"`
	ObjectID    int64     `json:"objectId"`
	FileName    string    `json:"fileName"`
	MimeType    string    `json:"mimeType"`
	Size        int64     `json:"size"`
	SHA256      string    `json:"sha256,omitempty"`
	StorageKey  string    `json:"storageKey"`
	AccessLevel string    `json:"accessLevel"`
	ExpiresAt   string    `json:"expiresAt,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

type UploadToken struct {
	FileID     int64             `json:"fileId"`
	UploadURL  string            `json:"uploadUrl"`
	StorageKey string            `json:"storageKey"`
	Headers    map[string]string `json:"headers"`
	ExpiresAt  string            `json:"expiresAt"`
}

type DownloadURL struct {
	FileID      int64  `json:"fileId"`
	DownloadURL string `json:"downloadUrl"`
	ExpiresAt   string `json:"expiresAt"`
}

type Repository interface {
	SaveFile(ctx context.Context, file File) (File, error)
	FindFile(ctx context.Context, fileID int64) (File, bool, error)
}

type Service struct {
	mu               sync.RWMutex
	nextID           int64
	files            map[int64]File
	repo             Repository
	uploadBaseURL    string
	downloadBaseURL  string
	requirePublicURL bool
}

func NewService() *Service {
	return NewServiceWithRepository(nil)
}

func NewServiceWithRepository(repo Repository) *Service {
	return &Service{
		nextID: 1,
		files:  make(map[int64]File),
		repo:   repo,
	}
}

func (s *Service) UsePublicBaseURLs(uploadBaseURL string, downloadBaseURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.uploadBaseURL = strings.TrimRight(strings.TrimSpace(uploadBaseURL), "/")
	s.downloadBaseURL = strings.TrimRight(strings.TrimSpace(downloadBaseURL), "/")
}

func (s *Service) RequirePublicBaseURLs(require bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requirePublicURL = require
}

func (s *Service) CreateUploadToken(userID int64, req UploadTokenRequest) (UploadToken, File, error) {
	req.BizType = strings.TrimSpace(req.BizType)
	if req.BizType == "" {
		req.BizType = "chat_file"
	}
	req.FileName = safeFileName(req.FileName)
	req.MimeType = strings.ToLower(strings.TrimSpace(req.MimeType))
	policy, ok := uploadPolicies[req.BizType]
	if !ok || !validUploadRequest(req, policy) {
		return UploadToken{}, File{}, ErrInvalidFile
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	file := File{
		ID:          s.nextID,
		UploaderID:  userID,
		BizType:     req.BizType,
		ObjectID:    req.ObjectID,
		FileName:    req.FileName,
		MimeType:    req.MimeType,
		Size:        req.Size,
		SHA256:      req.SHA256,
		StorageKey:  fmt.Sprintf("%s/%d/%d-%s", req.BizType, req.ObjectID, s.nextID, req.FileName),
		AccessLevel: policy.accessLevel,
		CreatedAt:   time.Now(),
	}
	s.nextID++
	if s.repo != nil {
		saved, err := s.repo.SaveFile(context.Background(), file)
		if err != nil {
			return UploadToken{}, File{}, err
		}
		file = saved
	}
	s.files[file.ID] = file

	expiresAt := time.Now().Add(15 * time.Minute).Format(time.RFC3339)
	uploadURL, err := s.uploadURL(file.StorageKey)
	if err != nil {
		return UploadToken{}, File{}, err
	}
	return UploadToken{
		FileID:     file.ID,
		UploadURL:  uploadURL,
		StorageKey: file.StorageKey,
		Headers: map[string]string{
			"x-zhw-file-id": fmt.Sprintf("%d", file.ID),
		},
		ExpiresAt: expiresAt,
	}, file, nil
}

func allowMIMEs(values ...string) map[string]struct{} {
	allowed := make(map[string]struct{}, len(values))
	for _, value := range values {
		allowed[value] = struct{}{}
	}
	return allowed
}

func validUploadRequest(req UploadTokenRequest, policy uploadPolicy) bool {
	if req.FileName == "" || req.FileName == "." || req.FileName == ".." || req.Size <= 0 || req.Size > policy.maxSize {
		return false
	}
	if policy.requireObject && req.ObjectID <= 0 {
		return false
	}
	_, ok := policy.mimeTypes[req.MimeType]
	return ok
}

func safeFileName(fileName string) string {
	return filepath.Base(strings.TrimSpace(fileName))
}

func (s *Service) CreateGeneratedFile(bizType string, objectID int64, fileName string, mimeType string, size int64) (File, error) {
	return s.CreateGeneratedFileWithTTL(bizType, objectID, fileName, mimeType, size, 0)
}

func (s *Service) CreateGeneratedFileWithTTL(bizType string, objectID int64, fileName string, mimeType string, size int64, ttl time.Duration) (File, error) {
	if strings.TrimSpace(fileName) == "" {
		return File{}, ErrInvalidFile
	}
	if bizType == "" {
		bizType = "export_file"
	}
	if size <= 0 {
		size = 1
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	file := File{
		ID:          s.nextID,
		BizType:     bizType,
		ObjectID:    objectID,
		FileName:    filepath.Base(fileName),
		MimeType:    mimeType,
		Size:        size,
		StorageKey:  fmt.Sprintf("%s/%d/%d-%s", bizType, objectID, s.nextID, filepath.Base(fileName)),
		AccessLevel: "admin_permission",
		CreatedAt:   time.Now(),
	}
	if ttl != 0 {
		file.ExpiresAt = time.Now().Add(ttl).Format(time.RFC3339)
	}
	s.nextID++
	if s.repo != nil {
		saved, err := s.repo.SaveFile(context.Background(), file)
		if err != nil {
			return File{}, err
		}
		file = saved
	}
	s.files[file.ID] = file
	return file, nil
}

func (s *Service) Get(fileID int64) (File, error) {
	if s.repo != nil {
		file, ok, err := s.repo.FindFile(context.Background(), fileID)
		if err != nil {
			return File{}, err
		}
		if !ok {
			return File{}, ErrFileNotFound
		}
		return file, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	file, ok := s.files[fileID]
	if !ok {
		return File{}, ErrFileNotFound
	}
	return file, nil
}

func (s *Service) DownloadURL(fileID int64) (DownloadURL, error) {
	file, err := s.Get(fileID)
	if err != nil {
		return DownloadURL{}, err
	}
	expiresAt := time.Now().Add(10 * time.Minute)
	if fileExpiresAt, ok := parseFileExpiresAt(file.ExpiresAt); ok {
		if !time.Now().Before(fileExpiresAt) {
			return DownloadURL{}, ErrFileExpired
		}
		if fileExpiresAt.Before(expiresAt) {
			expiresAt = fileExpiresAt
		}
	}
	downloadURL, err := s.downloadURL(file.StorageKey)
	if err != nil {
		return DownloadURL{}, err
	}
	return DownloadURL{
		FileID:      file.ID,
		DownloadURL: downloadURL,
		ExpiresAt:   expiresAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) DownloadURLForFile(file File) (DownloadURL, error) {
	expiresAt := time.Now().Add(10 * time.Minute)
	if fileExpiresAt, ok := parseFileExpiresAt(file.ExpiresAt); ok && fileExpiresAt.Before(expiresAt) {
		expiresAt = fileExpiresAt
	}
	downloadURL, err := s.downloadURL(file.StorageKey)
	if err != nil {
		return DownloadURL{}, err
	}
	return DownloadURL{
		FileID:      file.ID,
		DownloadURL: downloadURL,
		ExpiresAt:   expiresAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) uploadURL(storageKey string) (string, error) {
	if s.uploadBaseURL == "" {
		if s.requirePublicURL {
			return "", ErrStorageNotConfigured
		}
		return "mock://upload/" + storageKey, nil
	}
	return s.uploadBaseURL + "/" + urlPathEscape(storageKey), nil
}

func (s *Service) downloadURL(storageKey string) (string, error) {
	if s.downloadBaseURL == "" {
		if s.requirePublicURL {
			return "", ErrStorageNotConfigured
		}
		return "mock://download/" + storageKey, nil
	}
	return s.downloadBaseURL + "/" + urlPathEscape(storageKey), nil
}

func urlPathEscape(storageKey string) string {
	parts := strings.Split(storageKey, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func parseFileExpiresAt(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	expiresAt, err := time.Parse(time.RFC3339, value)
	return expiresAt, err == nil
}
