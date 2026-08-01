package appapi

import (
	"log"
	"net/http"
)

const creditPersistenceHeader = "X-Credit-Persistence"

// markCreditPersistenceDegraded keeps an already committed business action
// observable without returning a misleading retryable 5xx response.
func markCreditPersistenceDegraded(w http.ResponseWriter, operation string, userID, gameID int64, err error) {
	if w != nil {
		w.Header().Set(creditPersistenceHeader, "degraded")
	}
	log.Printf("credit persistence degraded operation=%q user_id=%d game_id=%d err=%v", operation, userID, gameID, err)
}
