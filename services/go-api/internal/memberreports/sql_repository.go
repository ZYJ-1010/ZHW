package memberreports

import (
	"context"
	"database/sql"
	"encoding/json"

	"zhw-mini/services/go-api/internal/revenue"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) SaveMembership(ctx context.Context, membership Membership) (Membership, error) {
	return scanMembership(r.db.QueryRowContext(ctx, `
insert into member_report_memberships (user_id, plan_name, invited_count, created_at, updated_at)
values ($1,$2,$3,now(),now())
on conflict (user_id) do update set
  plan_name = excluded.plan_name,
  invited_count = excluded.invited_count,
  updated_at = now()
returning user_id, plan_name, invited_count, updated_at
`, membership.UserID, membership.PlanName, membership.InvitedCount))
}

func (r *SQLRepository) GetMembership(ctx context.Context, userID int64) (Membership, bool, error) {
	membership, err := scanMembership(r.db.QueryRowContext(ctx, membershipSelect()+` where user_id = $1`, userID))
	if err == sql.ErrNoRows {
		return Membership{}, false, nil
	}
	if err != nil {
		return Membership{}, false, err
	}
	return membership, true, nil
}

func (r *SQLRepository) ListMemberships(ctx context.Context) ([]Membership, error) {
	rows, err := r.db.QueryContext(ctx, membershipSelect()+` order by updated_at desc, user_id desc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Membership, 0)
	for rows.Next() {
		item, err := scanMembership(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) SaveSnapshot(ctx context.Context, snapshot Snapshot) (Snapshot, error) {
	metrics, err := json.Marshal(snapshotMetrics{
		Participated:   snapshot.Participated,
		Completed:      snapshot.Completed,
		AverageScore:   snapshot.AverageScore,
		IncomeSummary:  snapshot.IncomeSummary,
		InvitedCount:   snapshot.InvitedCount,
		MembershipPlan: snapshot.MembershipPlan,
	})
	if err != nil {
		return Snapshot{}, err
	}
	return scanSnapshot(r.db.QueryRowContext(ctx, `
insert into member_report_snapshots (user_id, report_type, period, metrics, created_at)
values ($1,$2,$3,$4,$5)
returning id, user_id, report_type, period, metrics, created_at
`, snapshot.UserID, snapshot.ReportType, snapshot.Period, string(metrics), snapshot.CreatedAt))
}

type snapshotMetrics struct {
	Participated   int                   `json:"participatedGames"`
	Completed      int                   `json:"completedGames"`
	AverageScore   float64               `json:"averageScore"`
	IncomeSummary  revenue.IncomeSummary `json:"incomeSummary"`
	InvitedCount   int                   `json:"invitedCount"`
	MembershipPlan string                `json:"membershipPlan"`
}

func membershipSelect() string {
	return `select user_id, plan_name, invited_count, updated_at from member_report_memberships`
}

func scanMembership(row interface {
	Scan(dest ...any) error
}) (Membership, error) {
	var item Membership
	err := row.Scan(&item.UserID, &item.PlanName, &item.InvitedCount, &item.UpdatedAt)
	return item, err
}

func scanSnapshot(row interface {
	Scan(dest ...any) error
}) (Snapshot, error) {
	var item Snapshot
	var rawMetrics []byte
	if err := row.Scan(&item.ID, &item.UserID, &item.ReportType, &item.Period, &rawMetrics, &item.CreatedAt); err != nil {
		return Snapshot{}, err
	}
	var metrics snapshotMetrics
	_ = json.Unmarshal(rawMetrics, &metrics)
	item.Participated = metrics.Participated
	item.Completed = metrics.Completed
	item.AverageScore = metrics.AverageScore
	item.IncomeSummary = metrics.IncomeSummary
	item.InvitedCount = metrics.InvitedCount
	item.MembershipPlan = metrics.MembershipPlan
	return item, nil
}
