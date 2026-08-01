package appapi

import (
	"log"
	"net/http"
)

const reviewPersistenceHeader = "X-Review-Persistence"

func markReviewPersistenceDegraded(w http.ResponseWriter, operation string, gameID int64, err error) {
	if w != nil {
		w.Header().Set(reviewPersistenceHeader, "degraded")
	}
	log.Printf("review persistence degraded operation=%q game_id=%d err=%v", operation, gameID, err)
}
