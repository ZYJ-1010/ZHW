package appapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/memberreports"
	"zhw-mini/services/go-api/internal/teams"
)

func (s *Server) memberReportMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	snapshot, err := s.memberReports.Me(userID)
	if err != nil {
		writeMemberTeamError(w, err)
		return
	}
	s.recordBehavior(userID, "view_member_report", "member_report", snapshot.ID, nil)
	httpx.OK(w, snapshot)
}

func (s *Server) adminMemberReports(w http.ResponseWriter, r *http.Request) {
	items, err := s.memberReports.AdminSnapshotsStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取会员报表失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) myTeam(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	team, err := s.teams.My(userID)
	if err != nil {
		writeMemberTeamError(w, err)
		return
	}
	s.recordBehavior(userID, "view_my_team", "team", team.ID, nil)
	httpx.OK(w, team)
}

func (s *Server) adminTeams(w http.ResponseWriter, r *http.Request) {
	items, err := s.teams.AllStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取团队列表失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) adminTeamDetail(w http.ResponseWriter, r *http.Request) {
	teamID, ok := teamIDFromAdminPath(w, r.URL.Path)
	if !ok {
		return
	}
	detail, err := s.teams.Detail(teamID)
	if err != nil {
		writeMemberTeamError(w, err)
		return
	}
	httpx.OK(w, detail)
}

func (s *Server) myTeamMembers(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	members, err := s.teams.Members(userID)
	if err != nil {
		writeMemberTeamError(w, err)
		return
	}
	s.recordBehavior(userID, "view_team_members", "team", userID, map[string]interface{}{"count": len(members)})
	httpx.OK(w, map[string]interface{}{"items": members})
}

func (s *Server) myTeamRevenueSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	summary, err := s.teams.RevenueSummary(userID)
	if err != nil {
		writeMemberTeamError(w, err)
		return
	}
	s.recordBehavior(userID, "view_team_revenue_summary", "team", summary.TeamID, nil)
	httpx.OK(w, summary)
}

func teamIDFromAdminPath(w http.ResponseWriter, path string) (int64, bool) {
	idText := strings.TrimPrefix(path, "/api/admin/teams/")
	id, err := strconv.ParseInt(strings.Trim(idText, "/"), 10, 64)
	if err != nil || id <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "团队 ID 错误")
		return 0, false
	}
	return id, true
}

func writeMemberTeamError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, memberreports.ErrMemberReportForbidden):
		httpx.Error(w, http.StatusForbidden, 40351, "无会员报表权限")
	case errors.Is(err, teams.ErrTeamForbidden):
		httpx.Error(w, http.StatusForbidden, 40352, "无团队管理权限")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "会员团队操作失败")
	}
}
