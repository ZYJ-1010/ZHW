package appapi

import (
	"log"
	"net/http"
)

const growthPersistenceHeader = "X-Growth-Persistence"

func markGrowthPersistenceDegraded(w http.ResponseWriter, operation string, userID, gameID int64, err error) {
	if w != nil {
		w.Header().Set(growthPersistenceHeader, "degraded")
	}
	log.Printf("growth persistence degraded operation=%q user_id=%d game_id=%d err=%v", operation, userID, gameID, err)
}
