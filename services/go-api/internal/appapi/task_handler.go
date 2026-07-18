package appapi

import (
	"net/http"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
)

func (s *Server) completeTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	code := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/app/newbie-tasks/"), "/")
	if code == "" || s.tasks == nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "任务编号无效")
		return
	}
	progress, err := s.tasks.MarkCompleted(userID, code)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "任务完成记录失败")
		return
	}
	httpx.OK(w, progress)
}
