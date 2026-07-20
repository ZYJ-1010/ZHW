package appapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/gamedrafts"
)

type gameDraftSaveRequest struct {
	Title   string          `json:"title"`
	Payload json.RawMessage `json:"payload"`
}

func gameDraftIDFromPath(path string) int64 {
	value := strings.Trim(strings.TrimPrefix(path, "/api/app/game-drafts/"), "/")
	if value == "" || strings.Contains(value, "/") {
		return 0
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0
	}
	return id
}

func (s *Server) listGameDrafts(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	items, err := s.gameDrafts.List(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "获取草稿箱失败")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) createGameDraft(w http.ResponseWriter, r *http.Request) {
	s.saveGameDraft(w, r, 0)
}

func (s *Server) updateGameDraft(w http.ResponseWriter, r *http.Request) {
	draftID := gameDraftIDFromPath(r.URL.Path)
	if draftID <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "草稿不存在")
		return
	}
	s.saveGameDraft(w, r, draftID)
}

func (s *Server) saveGameDraft(w http.ResponseWriter, r *http.Request, draftID int64) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req gameDraftSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "草稿内容格式错误")
		return
	}
	draft, err := s.gameDrafts.Save(userID, draftID, req.Title, req.Payload)
	if err != nil {
		writeGameDraftError(w, err)
		return
	}
	s.recordBehavior(userID, "game_create_draft_save", "game_create_draft", draft.ID, map[string]interface{}{"title": draft.Title})
	httpx.OK(w, draft)
}

func (s *Server) getGameDraft(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	draft, err := s.gameDrafts.Get(userID, gameDraftIDFromPath(r.URL.Path))
	if err != nil {
		writeGameDraftError(w, err)
		return
	}
	httpx.OK(w, draft)
}

func (s *Server) deleteGameDraft(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	draftID := gameDraftIDFromPath(r.URL.Path)
	if err := s.gameDrafts.Delete(userID, draftID); err != nil {
		writeGameDraftError(w, err)
		return
	}
	s.recordBehavior(userID, "game_create_draft_delete", "game_create_draft", draftID, nil)
	httpx.OK(w, map[string]bool{"deleted": true})
}

func writeGameDraftError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, gamedrafts.ErrNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "草稿不存在或已删除")
	case errors.Is(err, gamedrafts.ErrForbidden):
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "无权操作该草稿")
	case errors.Is(err, gamedrafts.ErrInvalid):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "草稿内容不合法")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "草稿保存失败")
	}
}
