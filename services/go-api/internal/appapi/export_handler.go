package appapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/exports"
	"zhw-mini/services/go-api/internal/files"
)

func (s *Server) adminExportTemplates(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"items": s.exports.Templates()})
}

func (s *Server) createExportTask(w http.ResponseWriter, r *http.Request) {
	var req exports.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	task, err := s.exports.Create(parseInt64Header(r, "X-Admin-ID"), req)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "export template not found or disabled")
		return
	}
	s.recordOperation(r, "export:create", "export_task", strconv.FormatInt(task.ID, 10), map[string]interface{}{
		"taskNo":       task.TaskNo,
		"templateCode": task.TemplateCode,
		"exportType":   task.ExportType,
	})
	httpx.OK(w, task)
}

func (s *Server) adminExportTasks(w http.ResponseWriter, r *http.Request) {
	items := filterExportTasks(s.exports.Tasks(), r)
	httpx.OK(w, map[string]interface{}{"items": items, "total": len(items)})
}

func filterExportTasks(items []exports.Task, r *http.Request) []exports.Task {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	templateCode := strings.TrimSpace(r.URL.Query().Get("templateCode"))
	exportType := strings.TrimSpace(r.URL.Query().Get("exportType"))
	createdBy := parseInt64Query(r, "createdBy")
	filtered := make([]exports.Task, 0, len(items))
	for _, item := range items {
		if status != "" && item.Status != status {
			continue
		}
		if templateCode != "" && item.TemplateCode != templateCode {
			continue
		}
		if exportType != "" && item.ExportType != exportType {
			continue
		}
		if createdBy > 0 && item.CreatedBy != createdBy {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func (s *Server) adminExportTaskDownloadURL(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "/download-url") {
		http.NotFound(w, r)
		return
	}
	taskID, err := pathID(r.URL.Path, "/api/admin/export-tasks/", "/download-url")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid task id")
		return
	}
	fileID, err := s.exports.ReadyFileID(taskID)
	if err != nil {
		if errors.Is(err, exports.ErrTaskNotFound) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "export task not found")
			return
		}
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "export file not ready")
		return
	}
	url, err := s.files.DownloadURL(fileID)
	if err != nil {
		if errors.Is(err, files.ErrFileExpired) {
			httpx.Error(w, http.StatusGone, httpx.CodeConflict, "export file expired")
			return
		}
		if errors.Is(err, files.ErrStorageNotConfigured) {
			httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "storage base url not configured")
			return
		}
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "export file not found")
		return
	}
	s.recordOperation(r, "export:download_url", "export_task", strconv.FormatInt(taskID, 10), map[string]interface{}{
		"fileId": fileID,
	})
	httpx.OK(w, url)
}

func (s *Server) runExportTasks(w http.ResponseWriter, r *http.Request) {
	results, err := s.exports.RunPending(20, func(task exports.Task, template exports.Template) (int64, string, error) {
		file, err := s.files.CreateGeneratedFileWithTTL("export_file", task.ID, template.FileName, "text/csv", int64(len(template.Columns)*16), 24*time.Hour)
		if err != nil {
			return 0, "", err
		}
		return file.ID, file.StorageKey, nil
	})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "export task run failed")
		return
	}
	for _, result := range results {
		s.recordOperation(r, "export:run", "export_task", strconv.FormatInt(result.Task.ID, 10), map[string]interface{}{
			"fileId":     result.Task.FileID,
			"storageKey": result.StorageKey,
		})
	}
	httpx.OK(w, map[string]interface{}{"items": results, "processed": len(results)})
}

func pathID(path string, prefix string, suffix string) (int64, error) {
	value := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	value = strings.Trim(value, "/")
	return strconv.ParseInt(value, 10, 64)
}
