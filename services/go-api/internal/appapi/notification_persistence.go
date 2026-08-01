package appapi

import (
	"log"
	"net/http"

	"zhw-mini/services/go-api/internal/notifications"
)

const notificationPersistenceHeader = "X-Notification-Persistence"

// createCriticalNotification preserves the already-committed business result
// while making notification storage degradation observable. Returning a 5xx
// here would invite a retry after the game/application state has changed.
func (s *Server) createCriticalNotification(w http.ResponseWriter, operation string, req notifications.CreateRequest) (notifications.Notification, error) {
	notification, err := s.notices.CreatePersisted(req)
	if err == nil {
		return notification, nil
	}

	if w != nil {
		w.Header().Set(notificationPersistenceHeader, "degraded")
	}
	log.Printf(
		"notification persistence degraded operation=%q user_id=%d notify_type=%q biz_type=%q biz_id=%d err=%v",
		operation,
		req.UserID,
		req.NotifyType,
		req.BizType,
		req.BizID,
		err,
	)
	return notification, err
}
