package appapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/files"
)

func (s *Server) createUploadToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	var req files.UploadTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	if req.BizType == "" {
		req.BizType = "chat_file"
	}
	if (req.BizType == "chat_file" || req.BizType == "delivery_proof") && !s.gamesMember(req.ObjectID, userID) {
		httpx.Error(w, http.StatusForbidden, 40331, "file upload requires game membership")
		return
	}
	token, file, err := s.files.CreateUploadToken(userID, req)
	if err != nil {
		log.Printf("file upload token failed user_id=%d biz_type=%q object_id=%d file_name=%q mime_type=%q size=%d err=%v", userID, req.BizType, req.ObjectID, req.FileName, req.MimeType, req.Size, err)
		if errors.Is(err, files.ErrStorageNotConfigured) {
			httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "storage base url not configured")
			return
		}
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid file request")
		return
	}
	httpx.OK(w, map[string]interface{}{"upload": token, "file": file})
}

func (s *Server) adminCreateUploadToken(w http.ResponseWriter, r *http.Request) {
	adminID, ok := s.admins.HasPermission(s.adminToken(r), "game:create_admin")
	if !ok {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "缺少后台接口权限")
		return
	}
	var req files.UploadTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	if req.BizType == "" {
		req.BizType = "game_cover"
	}
	if req.BizType != "game_cover" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid file request")
		return
	}
	token, file, err := s.files.CreateUploadToken(adminID, req)
	if err != nil {
		log.Printf("admin file upload token failed admin_id=%d biz_type=%q file_name=%q mime_type=%q size=%d err=%v", adminID, req.BizType, req.FileName, req.MimeType, req.Size, err)
		if errors.Is(err, files.ErrStorageNotConfigured) {
			httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "storage base url not configured")
			return
		}
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid file request")
		return
	}
	download, err := s.files.DownloadURLForFile(file)
	if err != nil {
		if errors.Is(err, files.ErrStorageNotConfigured) {
			httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "storage base url not configured")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "failed to generate download url")
		return
	}
	s.recordOperation(r, "file:upload_token", "file", strconv.FormatInt(file.ID, 10), map[string]interface{}{
		"bizType":  req.BizType,
		"fileName": file.FileName,
		"size":     file.Size,
	})
	httpx.OK(w, map[string]interface{}{"upload": token, "file": file, "download": download})
}

func (s *Server) downloadFileURL(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	fileID, ok := fileIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}
	file, err := s.files.Get(fileID)
	if err != nil {
		if errors.Is(err, files.ErrFileNotFound) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "file not found")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "failed to get file")
		return
	}
	if !s.canDownloadAppFile(file, userID) {
		httpx.Error(w, http.StatusForbidden, 40331, "file access denied")
		return
	}
	url, err := s.files.DownloadURL(fileID)
	if err != nil {
		if errors.Is(err, files.ErrFileExpired) {
			httpx.Error(w, http.StatusGone, httpx.CodeConflict, "file expired")
			return
		}
		if errors.Is(err, files.ErrStorageNotConfigured) {
			httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "storage base url not configured")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "failed to generate download url")
		return
	}
	httpx.OK(w, url)
}

func (s *Server) adminDownloadFileURL(w http.ResponseWriter, r *http.Request) {
	fileID, ok := adminFileIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}
	file, err := s.files.Get(fileID)
	if err != nil {
		if errors.Is(err, files.ErrFileNotFound) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "file not found")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "failed to get file")
		return
	}
	permission, allowed := adminFilePermission(file)
	if !allowed {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "file access denied")
		return
	}
	if _, ok := s.admins.HasPermission(s.adminToken(r), permission); !ok {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "file access denied")
		return
	}
	url, err := s.files.DownloadURL(fileID)
	if err != nil {
		if errors.Is(err, files.ErrFileExpired) {
			httpx.Error(w, http.StatusGone, httpx.CodeConflict, "file expired")
			return
		}
		if errors.Is(err, files.ErrStorageNotConfigured) {
			httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "storage base url not configured")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "failed to generate download url")
		return
	}
	s.recordOperation(r, "file:download_url", "file", strconv.FormatInt(fileID, 10), map[string]interface{}{
		"bizType":    file.BizType,
		"objectId":   file.ObjectID,
		"permission": permission,
	})
	httpx.OK(w, url)
}

func (s *Server) canDownloadAppFile(file files.File, userID int64) bool {
	switch file.BizType {
	case "avatar":
		return file.UploaderID == userID
	case "game_application":
		return file.UploaderID == userID || s.gameCreator(file.ObjectID, userID)
	case "chat_file", "delivery_proof":
		return s.gamesMember(file.ObjectID, userID)
	default:
		return false
	}
}

func adminFilePermission(file files.File) (string, bool) {
	switch file.BizType {
	case "realname_material":
		return "identity:read", true
	case "enterprise_material":
		return "identity:read", true
	case "report_attachment":
		return "report:view", true
	case "game_application":
		return "game:view", true
	case "chat_file":
		return "im:message:view_dispute", true
	case "delivery_proof":
		return "delivery:manage", true
	case "export_file":
		return "report_export:create", true
	case "game_cover":
		return "game:view", true
	default:
		return "", false
	}
}

func (s *Server) gameCreator(gameID int64, userID int64) bool {
	game, err := s.games.Get(gameID)
	return err == nil && game.CreatorUserID == userID
}

func (s *Server) gamesMember(gameID int64, userID int64) bool {
	type memberChecker interface {
		IsMember(gameID int64, userID int64) bool
	}
	if checker, ok := s.games.(memberChecker); ok {
		return checker.IsMember(gameID, userID)
	}
	return false
}

func adminFileIDFromPath(w http.ResponseWriter, path string) (int64, bool) {
	idText := strings.TrimPrefix(path, "/api/admin/files/")
	idText = strings.TrimSuffix(idText, "/download-url")
	id, err := strconv.ParseInt(strings.Trim(idText, "/"), 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid file id")
		return 0, false
	}
	return id, true
}

func fileIDFromPath(w http.ResponseWriter, path string) (int64, bool) {
	idText := strings.TrimPrefix(path, "/api/app/files/")
	idText = strings.TrimSuffix(idText, "/download-url")
	id, err := strconv.ParseInt(strings.Trim(idText, "/"), 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid file id")
		return 0, false
	}
	return id, true
}
