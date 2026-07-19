package appapi

import (
	"net/http"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/notifications"
	"zhw-mini/services/go-api/internal/users"
)

func (s *Server) runReviewRemindJob(w http.ResponseWriter, r *http.Request) {
	created := 0
	for _, game := range s.games.List() {
		if game.Status != "pending_review" && game.Status != "completed" {
			continue
		}
		for _, userID := range s.games.Members(game.ID) {
			todos, err := s.reviews.Todos(userID)
			if err != nil {
				continue
			}
			for _, todo := range todos {
				if todo.GameID != game.ID {
					continue
				}
				s.notices.Create(notifications.CreateRequest{
					UserID:     userID,
					NotifyType: "review_remind",
					Title:      "评价提醒",
					Content:    "你参与的局已结束，请及时完成评价。",
					BizType:    "game",
					BizID:      game.ID,
					NeedWechat: true,
					WechatData: map[string]string{
						"thing1": game.Title,
						"time2":  todo.DeadlineAt,
					},
				})
				created++
				break
			}
		}
	}
	httpx.OK(w, map[string]interface{}{"created": created})
}

func (s *Server) runProgressFeedbackRemindJob(w http.ResponseWriter, r *http.Request) {
	created := 0
	for _, game := range s.games.List() {
		if game.Status != "in_progress" && game.Status != "pending_confirm" {
			continue
		}
		s.notices.Create(notifications.CreateRequest{
			UserID:     game.CreatorUserID,
			NotifyType: "progress_feedback_remind",
			Title:      "进度反馈提醒",
			Content:    "你发起的局正在进行中，请及时反馈服务进度。",
			BizType:    "game",
			BizID:      game.ID,
			NeedWechat: true,
			WechatData: map[string]string{
				"thing1": game.Title,
				"thing2": game.Status,
			},
		})
		created++
	}
	httpx.OK(w, map[string]interface{}{"created": created})
}

func (s *Server) runPointsExpireJob(w http.ResponseWriter, r *http.Request) {
	rules := s.currentOperationRules()
	if !rules.Points.ExpireEnabled || rules.Points.ExpireDays <= 0 {
		httpx.OK(w, map[string]interface{}{"enabled": false, "expiredCount": 0, "items": []interface{}{}})
		return
	}
	cutoff := time.Now().Add(-time.Duration(rules.Points.ExpireDays) * 24 * time.Hour)
	items := make([]map[string]interface{}, 0)
	usersList, err := s.auth.AdminUsers(users.Filter{})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "list users failed")
		return
	}
	for _, user := range usersList {
		account, log, expireErr := s.points.Expire(user.ID, cutoff)
		if expireErr != nil || log.ID == 0 {
			continue
		}
		items = append(items, map[string]interface{}{"userId": user.ID, "expiredPoints": -log.ChangeValue, "account": account, "log": log})
	}
	httpx.OK(w, map[string]interface{}{"enabled": true, "cutoff": cutoff, "expiredCount": len(items), "items": items})
}

func (s *Server) archiveExpiredIMRooms(w http.ResponseWriter, r *http.Request) {
	gameIDs := make([]int64, 0)
	skipped := make([]map[string]interface{}, 0)
	for _, game := range s.games.List() {
		if !imArchiveCandidateStatus(game.Status) {
			continue
		}
		if s.hasOpenReport(game.ID) {
			skipped = append(skipped, map[string]interface{}{
				"gameId": game.ID,
				"reason": "open_report",
			})
			continue
		}
		gameIDs = append(gameIDs, game.ID)
	}
	archived := s.im.ArchiveRoomsByGameIDs(gameIDs, "archive_job")
	httpx.OK(w, map[string]interface{}{
		"archivedCount": len(archived),
		"archivedRooms": archived,
		"skippedCount":  len(skipped),
		"skippedRooms":  skipped,
	})
}

func imArchiveCandidateStatus(status string) bool {
	return status == "pending_review" || status == "completed"
}

func (s *Server) hasOpenReport(gameID int64) bool {
	for _, report := range s.reports.List() {
		if report.GameID == gameID && report.Status != "closed" {
			return true
		}
	}
	return false
}
