package games

import (
	"context"
	"time"
)

const (
	StatusDraft          = "draft"
	StatusPendingAudit   = "pending_audit"
	StatusRecruiting     = "recruiting"
	StatusFull           = "full"
	StatusInProgress     = "in_progress"
	StatusPendingConfirm = "pending_confirm"
	StatusPendingReview  = "pending_review"
	StatusCompleted      = "completed"
	StatusRejected       = "rejected"
	StatusCancelled      = "canceled"
	StatusDisputed       = "disputed"
	StatusClosed         = "closed"
	StatusSettling       = "settling"
)

var gameStatusTransitions = map[string]map[string]bool{
	StatusDraft:          {StatusPendingAudit: true, StatusRejected: true},
	StatusPendingAudit:   {StatusRecruiting: true, StatusRejected: true},
	StatusRecruiting:     {StatusFull: true, StatusInProgress: true, StatusCancelled: true},
	StatusFull:           {StatusInProgress: true, StatusCancelled: true},
	StatusInProgress:     {StatusPendingConfirm: true, StatusPendingReview: true, StatusCancelled: true},
	StatusPendingConfirm: {StatusPendingReview: true, StatusCompleted: true, StatusDisputed: true},
	StatusPendingReview:  {StatusCompleted: true, StatusDisputed: true, StatusClosed: true},
	StatusCompleted:      {StatusSettling: true, StatusClosed: true},
	StatusSettling:       {StatusCompleted: true, StatusClosed: true},
	StatusDisputed:       {StatusSettling: true, StatusCompleted: true, StatusClosed: true},
}

type StatusLog struct {
	ID         int64     `json:"id"`
	GameID     int64     `json:"gameId"`
	FromStatus string    `json:"fromStatus"`
	ToStatus   string    `json:"toStatus"`
	OperatorID int64     `json:"operatorUserId"`
	Reason     string    `json:"reason,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

type statusLogRepository interface {
	SaveStatusLog(ctx context.Context, log StatusLog) (StatusLog, error)
}

func validStatusTransition(from, to string) bool {
	if from == to {
		return true
	}
	return gameStatusTransitions[from][to]
}
