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

// statusTransitionRepository persists the game row and its lifecycle log in
// one transaction. A transition log must never become visible unless the
// corresponding game status update commits successfully.
type statusTransitionRepository interface {
	UpdateGameWithStatusLog(ctx context.Context, game Game, log StatusLog) (Game, StatusLog, error)
}

// cancelGameTransitionRepository extends a regular lifecycle transition with
// the membership changes required by service cancellation. Implementations
// must commit the active-member cancellation, game row and status log as one
// transaction.
type cancelGameTransitionRepository interface {
	CancelGameWithStatusLog(ctx context.Context, game Game, reason string, log StatusLog) (Game, StatusLog, error)
}

func validStatusTransition(from, to string) bool {
	if from == to {
		return true
	}
	return gameStatusTransitions[from][to]
}

// StatusText is the single source of truth for the user-facing game lifecycle
// label. API consumers must use this instead of maintaining their own status
// translation tables, so the hall, game detail and "我的局" stay consistent.
func StatusText(status string) string {
	switch status {
	case StatusDraft:
		return "草稿"
	case StatusPendingAudit:
		return "待后台审核"
	case StatusRecruiting:
		return "招募中"
	case StatusFull:
		return "已满员"
	case StatusInProgress:
		return "进行中"
	case StatusPendingConfirm:
		return "待成员确认"
	case StatusPendingReview:
		return "待评价"
	case StatusCompleted:
		return "已完成"
	case StatusRejected:
		return "审核未通过"
	case StatusCancelled, "cancelled":
		return "已取消"
	case StatusDisputed:
		return "争议中"
	case StatusSettling:
		return "结算中"
	case StatusClosed:
		return "已关闭"
	default:
		return "状态处理中"
	}
}

// ApplicationStatusText keeps the player-side join flow separate from the
// game lifecycle. A player can see exactly whether a request is waiting,
// approved, rejected or withdrawn without inferring it from the game status.
func ApplicationStatusText(status string) string {
	switch status {
	case "pending":
		return "申请中"
	case "approved":
		return "已入局"
	case "rejected":
		return "未通过"
	case "cancelled":
		return "已取消报名"
	default:
		return "报名状态处理中"
	}
}
