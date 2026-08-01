package appapi

import (
	"net/http"
	"strconv"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/notifications"
	"zhw-mini/services/go-api/internal/users"
)

func (s *Server) runReviewRemindJob(w http.ResponseWriter, r *http.Request) {
	created := 0
	gameItems, err := s.games.ListStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取组局列表失败，未执行评价提醒")
		return
	}
	for _, game := range gameItems {
		if game.Status != "pending_review" && game.Status != "completed" {
			continue
		}
		for _, userID := range s.games.Members(game.ID) {
			exists, err := s.hasUnreadJobNotificationStrict(userID, "review_remind", game.ID)
			if err != nil {
				httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取提醒状态失败，未执行评价提醒")
				return
			}
			if exists {
				continue
			}
			todos, err := s.reviews.Todos(userID)
			if err != nil {
				httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取评价待办失败，未执行评价提醒")
				return
			}
			for _, todo := range todos {
				if todo.GameID != game.ID {
					continue
				}
				if _, err := s.createCriticalNotification(nil, "review_remind_job", notifications.CreateRequest{
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
				}); err != nil {
					httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "保存评价提醒失败，请稍后重试")
					return
				}
				created++
				break
			}
		}
	}
	httpx.OK(w, map[string]interface{}{"created": created})
}

func (s *Server) runProgressFeedbackRemindJob(w http.ResponseWriter, r *http.Request) {
	created := 0
	gameItems, err := s.games.ListStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取组局列表失败，未执行进度提醒")
		return
	}
	for _, game := range gameItems {
		if game.Status != "in_progress" && game.Status != "pending_confirm" {
			continue
		}
		exists, err := s.hasUnreadJobNotificationStrict(game.CreatorUserID, "progress_feedback_remind", game.ID)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取提醒状态失败，未执行进度提醒")
			return
		}
		if exists {
			continue
		}
		if _, err := s.createCriticalNotification(nil, "progress_feedback_remind_job", notifications.CreateRequest{
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
		}); err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "保存进度提醒失败，请稍后重试")
			return
		}
		created++
	}
	httpx.OK(w, map[string]interface{}{"created": created})
}

func (s *Server) hasUnreadJobNotificationStrict(userID int64, notifyType string, bizID int64) (bool, error) {
	items, err := s.notices.ListStrict(userID)
	if err != nil {
		return false, err
	}
	for _, item := range items {
		if item.NotifyType == notifyType && item.BizID == bizID && item.Status == "unread" {
			return true, nil
		}
	}
	return false, nil
}

func (s *Server) runPointsExpireJob(w http.ResponseWriter, r *http.Request) {
	rules, err := s.currentOperationRulesStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取积分规则失败，未执行积分到期处理")
		return
	}
	if !rules.Points.ExpireEnabled || rules.Points.ExpireDays <= 0 {
		httpx.OK(w, map[string]interface{}{"enabled": false, "expiredCount": 0, "items": []interface{}{}})
		return
	}
	cutoff := time.Now().Add(-time.Duration(rules.Points.ExpireDays) * 24 * time.Hour)
	items := make([]map[string]interface{}, 0)
	usersList, err := s.auth.AdminUsers(users.Filter{})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "获取用户列表失败")
		return
	}
	for _, user := range usersList {
		account, log, expireErr := s.points.Expire(user.ID, cutoff)
		if expireErr != nil || log.ID == 0 {
			continue
		}
		items = append(items, map[string]interface{}{"userId": user.ID, "expiredPoints": -log.ChangeValue, "account": account, "log": log})
		if _, err := s.createCriticalNotification(nil, "points_expire_job", notifications.CreateRequest{
			UserID: user.ID, NotifyType: "points_expired", Title: "积分到期提醒",
			Content: "本次有 " + strconv.Itoa(-log.ChangeValue) + " 积分到期扣减，剩余积分 " + strconv.Itoa(account.AvailablePoints) + "。",
			BizType: "points", BizID: log.ID,
		}); err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "保存积分到期提醒失败，请稍后重试")
			return
		}
	}
	httpx.OK(w, map[string]interface{}{"enabled": true, "cutoff": cutoff, "expiredCount": len(items), "items": items})
}

func (s *Server) archiveExpiredIMRooms(w http.ResponseWriter, r *http.Request) {
	gameIDs := make([]int64, 0)
	skipped := make([]map[string]interface{}, 0)
	reportItems, err := s.reports.ListStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取举报状态失败，未执行聊天室归档")
		return
	}
	openReportGames := make(map[int64]bool)
	for _, report := range reportItems {
		if report.Status != "closed" {
			openReportGames[report.GameID] = true
		}
	}
	gameItems, err := s.games.ListStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取组局列表失败，未执行聊天室归档")
		return
	}
	for _, game := range gameItems {
		if !imArchiveCandidateStatus(game.Status) {
			continue
		}
		if openReportGames[game.ID] {
			skipped = append(skipped, map[string]interface{}{
				"gameId": game.ID,
				"reason": "open_report",
			})
			continue
		}
		gameIDs = append(gameIDs, game.ID)
	}
	archived, err := s.im.ArchiveRoomsByGameIDsStrict(gameIDs, "archive_job")
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "归档聊天室失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{
		"archivedCount": len(archived),
		"archivedRooms": archived,
		"skippedCount":  len(skipped),
		"skippedRooms":  skipped,
	})
}

// retryFailedIMRooms gives create_failed rooms a bounded, repeatable recovery
// path for the internal scheduler. 管理员仍可通过后台单房间重试。
func (s *Server) retryFailedIMRooms(w http.ResponseWriter, r *http.Request) {
	items := make([]map[string]interface{}, 0)
	for _, room := range s.im.AdminRooms() {
		if room.Status != "create_failed" || room.ID <= 0 {
			continue
		}
		updated, err := s.im.AdminRetryCreateRoom(room.ID)
		item := map[string]interface{}{"roomId": room.ID, "gameId": room.GameID}
		if err != nil {
			item["success"] = false
			item["error"] = err.Error()
		} else {
			item["success"] = true
			item["room"] = updated
		}
		items = append(items, item)
	}
	httpx.OK(w, map[string]interface{}{"retriedCount": len(items), "items": items})
}

func imArchiveCandidateStatus(status string) bool {
	return status == "pending_review" || status == "completed"
}
