package appapi

import (
	"log"
	"net/http"
)

const connectionPersistenceHeader = "X-Connection-Persistence"

func markConnectionPersistenceDegraded(w http.ResponseWriter, operation string, userID, connectedUserID int64, err error) {
	if w != nil {
		w.Header().Set(connectionPersistenceHeader, "degraded")
	}
	log.Printf("connection persistence degraded operation=%q user_id=%d connected_user_id=%d err=%v", operation, userID, connectedUserID, err)
}
