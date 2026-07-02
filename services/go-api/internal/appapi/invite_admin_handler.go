package appapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/invites"
)

func (s *Server) adminInviteCodes(w http.ResponseWriter, r *http.Request) {
	ownerID, _, ok := optionalInt64Query(w, r, "ownerUserId")
	if !ok {
		return
	}
	items, err := s.auth.AdminInviteCodes(invites.CodeFilter{
		Status:    strings.TrimSpace(r.URL.Query().Get("status")),
		EntryType: strings.TrimSpace(r.URL.Query().Get("entryType")),
		OwnerID:   ownerID,
	})
	if err != nil {
		if errors.Is(err, invites.ErrInvalidEntryType) {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid entry type")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "list invite codes failed")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items, "total": len(items)})
}

func (s *Server) createAdminInviteCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code        string `json:"code"`
		OwnerUserID int64  `json:"ownerUserId"`
		MaxUses     int    `json:"maxUses"`
		EntryType   string `json:"entryType"`
		BatchCount  int    `json:"batchCount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	if req.BatchCount < 0 || req.BatchCount > 200 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid batch count")
		return
	}
	if req.BatchCount > 1 {
		if strings.TrimSpace(req.Code) != "" {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "batch code must be auto generated")
			return
		}
		items := make([]invites.InviteCode, 0, req.BatchCount)
		for i := 0; i < req.BatchCount; i++ {
			invite, err := s.auth.AdminCreateInviteCode("", req.OwnerUserID, req.MaxUses, req.EntryType)
			if err != nil {
				if errors.Is(err, invites.ErrInvalidEntryType) {
					httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid entry type")
					return
				}
				httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "create invite code failed")
				return
			}
			items = append(items, invite)
		}
		s.recordOperation(r, "invite_code:batch_create", "invite_code", "", map[string]interface{}{
			"entryType":   req.EntryType,
			"ownerUserId": req.OwnerUserID,
			"maxUses":     req.MaxUses,
			"batchCount":  req.BatchCount,
		})
		httpx.OK(w, map[string]interface{}{"items": items, "total": len(items)})
		return
	}
	invite, err := s.auth.AdminCreateInviteCode(req.Code, req.OwnerUserID, req.MaxUses, req.EntryType)
	if err != nil {
		if errors.Is(err, invites.ErrInvalidEntryType) {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid entry type")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "create invite code failed")
		return
	}
	s.recordOperation(r, "invite_code:create", "invite_code", strconv.FormatInt(invite.ID, 10), map[string]interface{}{
		"code":        invite.Code,
		"entryType":   invite.EntryType,
		"ownerUserId": invite.OwnerID,
	})
	httpx.OK(w, invite)
}

func (s *Server) adminInviteCodeDetail(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/admin/invite-codes/"))
	code = strings.Trim(code, "/")
	if code == "" || strings.Contains(code, "/") {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invite code required")
		return
	}
	items, err := s.auth.AdminInviteCodes(invites.CodeFilter{})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "list invite codes failed")
		return
	}
	var invite invites.InviteCode
	found := false
	for _, item := range items {
		if item.Code == code {
			invite = item
			found = true
			break
		}
	}
	if !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "invite code not found")
		return
	}
	relations, err := s.auth.AdminInviteRelations(invites.RelationFilter{InviteCodeID: invite.ID})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "list invite relations failed")
		return
	}
	httpx.OK(w, map[string]interface{}{
		"inviteCode": invite,
		"relations":  relations,
		"ownerUser":  s.adminUserSummary(invite.OwnerID),
		"boundUser":  s.adminUserSummary(invite.BoundWechatUserID),
	})
}

func (s *Server) routeAdminInviteCodePost(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/disable") {
		s.requireAdminPermission("invite_code:manage", s.disableAdminInviteCode)(w, r)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) disableAdminInviteCode(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/admin/invite-codes/"), "/disable")
	code = strings.TrimSpace(strings.Trim(code, "/"))
	if code == "" {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invite code required")
		return
	}
	invite, err := s.auth.AdminDisableInviteCode(code)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "invite code not found")
		return
	}
	s.recordOperation(r, "invite_code:disable", "invite_code", strconv.FormatInt(invite.ID, 10), map[string]interface{}{"code": invite.Code})
	httpx.OK(w, invite)
}

func (s *Server) adminInviteRelations(w http.ResponseWriter, r *http.Request) {
	inviterID, _, ok := optionalInt64Query(w, r, "inviterUserId")
	if !ok {
		return
	}
	inviteeID, _, ok := optionalInt64Query(w, r, "inviteeUserId")
	if !ok {
		return
	}
	inviteCodeID, _, ok := optionalInt64Query(w, r, "inviteCodeId")
	if !ok {
		return
	}
	items, err := s.auth.AdminInviteRelations(invites.RelationFilter{
		InviterUserID: inviterID,
		InviteeUserID: inviteeID,
		InviteCodeID:  inviteCodeID,
	})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "list invite relations failed")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items, "total": len(items)})
}

func (s *Server) adminUserSummary(userID int64) interface{} {
	if userID <= 0 {
		return nil
	}
	user, ok := s.auth.UserByID(userID)
	if !ok {
		return map[string]interface{}{"id": userID, "missing": true}
	}
	return map[string]interface{}{
		"id":             user.ID,
		"openId":         user.OpenID,
		"nickname":       user.Nickname,
		"realnameStatus": user.RealnameStatus,
		"status":         user.Status,
	}
}
