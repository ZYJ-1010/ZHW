package appapi

import (
	"log"
	"net/http"
)

const imPersistenceHeader = "X-IM-Persistence"

func markIMPersistenceDegraded(w http.ResponseWriter, operation string, gameID int64, err error) {
	if w != nil {
		w.Header().Set(imPersistenceHeader, "degraded")
	}
	log.Printf("im persistence degraded operation=%q game_id=%d err=%v", operation, gameID, err)
}
