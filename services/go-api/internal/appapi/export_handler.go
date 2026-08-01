package appapi

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/exports"
)

func (s *Server) adminExportTemplates(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"items": s.exports.Templates()})
}

func (s *Server) createExportTask(w http.ResponseWriter, r *http.Request) {
	var req exports.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	task, err := s.exports.Create(parseInt64Header(r, "X-Admin-ID"), req)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "导出模板不存在或已停用")
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
	tasks, err := s.exports.TasksStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取导出任务失败，请稍后重试")
		return
	}
	items := filterExportTasks(tasks, r)
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
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "导出任务编号错误")
		return
	}
	fileID, err := s.exports.ReadyFileID(taskID)
	if err != nil {
		if errors.Is(err, exports.ErrTaskNotFound) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "导出任务不存在")
			return
		}
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "导出文件尚未生成")
		return
	}
	file, err := s.files.Get(fileID)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "导出文件不存在")
		return
	}
	if expiresAt, ok := parseExportFileExpiry(file.ExpiresAt); ok && time.Now().After(expiresAt) {
		httpx.Error(w, http.StatusGone, httpx.CodeConflict, "导出文件已过期")
		return
	}
	if _, err := s.exports.FileContent(fileID); err != nil {
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "导出内容尚未生成")
		return
	}
	s.recordOperation(r, "export:download_url", "export_task", strconv.FormatInt(taskID, 10), map[string]interface{}{
		"fileId": fileID,
	})
	httpx.OK(w, map[string]interface{}{
		"downloadUrl": "/api/admin/export-tasks/" + strconv.FormatInt(taskID, 10) + "/download",
		"fileName":    file.FileName,
		"fileId":      fileID,
	})
}

func (s *Server) runExportTasks(w http.ResponseWriter, r *http.Request) {
	results, err := s.exports.RunPending(20, func(task exports.Task, template exports.Template) (int64, string, error) {
		content, err := s.exportCSV(task, template)
		if err != nil {
			return 0, "", err
		}
		file, err := s.files.CreateGeneratedFileWithTTL("export_file", task.ID, template.FileName, "text/csv", int64(len(content)), 24*time.Hour)
		if err != nil {
			return 0, "", err
		}
		if err := s.exports.SaveFileContent(file.ID, content); err != nil {
			return 0, "", err
		}
		return file.ID, file.StorageKey, nil
	})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "执行导出任务失败")
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

func (s *Server) routeAdminExportTaskGet(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/download-url") {
		s.adminExportTaskDownloadURL(w, r)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/download") {
		s.adminExportTaskDownload(w, r)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) adminExportTaskDownload(w http.ResponseWriter, r *http.Request) {
	taskID, err := pathID(r.URL.Path, "/api/admin/export-tasks/", "/download")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "导出任务编号错误")
		return
	}
	fileID, err := s.exports.ReadyFileID(taskID)
	if err != nil {
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "导出文件尚未生成")
		return
	}
	file, err := s.files.Get(fileID)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "导出文件不存在")
		return
	}
	if expiresAt, ok := parseExportFileExpiry(file.ExpiresAt); ok && time.Now().After(expiresAt) {
		httpx.Error(w, http.StatusGone, httpx.CodeConflict, "导出文件已过期")
		return
	}
	content, err := s.exports.FileContent(fileID)
	if err != nil {
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "导出内容尚未生成")
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+safeExportFileName(file.FileName)+`"`)
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	_, _ = w.Write(content)
	s.recordOperation(r, "export:download", "export_task", strconv.FormatInt(taskID, 10), map[string]interface{}{"fileId": fileID})
}

func (s *Server) exportCSV(task exports.Task, template exports.Template) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	// Excel on Windows recognises UTF-8 correctly with BOM.
	buffer.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(buffer)
	if err := writer.Write(template.Columns); err != nil {
		return nil, err
	}
	switch template.ExportType {
	case "reports":
		items, err := s.reports.ListStrict()
		if err != nil {
			return nil, fmt.Errorf("读取举报导出数据失败: %w", err)
		}
		for _, item := range items {
			if err := writer.Write([]string{strconv.FormatInt(item.ID, 10), strconv.FormatInt(item.GameID, 10), strconv.FormatInt(item.ReporterUserID, 10), item.ReportType, item.Status, item.CreatedAt.Format(time.RFC3339)}); err != nil {
				return nil, err
			}
		}
	case "operation_logs":
		items, err := s.audit.OperationLogsStrict()
		if err != nil {
			return nil, fmt.Errorf("读取操作日志导出数据失败: %w", err)
		}
		for _, item := range items {
			if err := writer.Write([]string{strconv.FormatInt(item.ID, 10), strconv.FormatInt(item.AdminUserID, 10), item.Action, item.TargetType, item.TargetID, item.CreatedAt.Format(time.RFC3339)}); err != nil {
				return nil, err
			}
		}
	case "reviews":
		items, err := s.reviews.AllReviewsStrict()
		if err != nil {
			return nil, fmt.Errorf("读取评价导出数据失败: %w", err)
		}
		for _, item := range items {
			if err := writer.Write([]string{strconv.FormatInt(item.ID, 10), strconv.FormatInt(item.GameID, 10), strconv.FormatInt(item.ReviewerUserID, 10), strconv.FormatInt(item.TargetUserID, 10), strconv.Itoa(item.Score), strings.Join(item.Tags, "|"), item.AgainIntent, item.CreatedAt.Format(time.RFC3339)}); err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("unsupported export type: %s", template.ExportType)
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func parseExportFileExpiry(value string) (time.Time, bool) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, value)
	return parsed, err == nil
}

func safeExportFileName(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), `"`, "")
	if value == "" {
		return "export.csv"
	}
	return value
}

func pathID(path string, prefix string, suffix string) (int64, error) {
	value := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	value = strings.Trim(value, "/")
	return strconv.ParseInt(value, 10, 64)
}
