package membership

import (
	"context"
	"database/sql"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) ListPlans(ctx context.Context) ([]Plan, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, plan_code, plan_name, status, created_at
from membership_plans
order by id asc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Plan, 0)
	for rows.Next() {
		var item Plan
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Status, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) UpsertPlan(ctx context.Context, plan Plan) (Plan, error) {
	return scanPlan(r.db.QueryRowContext(ctx, `
insert into membership_plans (plan_code, plan_name, status, created_at)
values ($1,$2,$3,now())
on conflict (plan_code) do update set
  plan_name = excluded.plan_name,
  status = excluded.status
returning id, plan_code, plan_name, status, created_at
`, plan.Code, plan.Name, plan.Status))
}

func (r *SQLRepository) GetMembership(ctx context.Context, userID int64) (Membership, bool, error) {
	item, err := scanMembership(r.db.QueryRowContext(ctx, membershipSelect()+` where m.user_id = $1 and m.status = 'active' order by m.started_at desc limit 1`, userID))
	if err == sql.ErrNoRows {
		return Membership{UserID: userID, PlanCode: "none", PlanName: "none", Status: "none"}, false, nil
	}
	if err != nil {
		return Membership{}, false, err
	}
	return item, true, nil
}

func (r *SQLRepository) SaveMembership(ctx context.Context, membership Membership) (Membership, error) {
	return scanMembership(r.db.QueryRowContext(ctx, `
insert into user_memberships (user_id, plan_code, status, started_at, expires_at)
values ($1,$2,$3,$4,$5)
returning id, user_id, plan_code,
  coalesce((select plan_name from membership_plans where plan_code = user_memberships.plan_code), plan_code) as plan_name,
  status, started_at, expires_at
`, membership.UserID, membership.PlanCode, membership.Status, membership.StartedAt, membership.ExpiresAt))
}

func membershipSelect() string {
	return `
select m.id, m.user_id, m.plan_code, coalesce(p.plan_name, m.plan_code) as plan_name,
  m.status, m.started_at, m.expires_at
from user_memberships m
left join membership_plans p on p.plan_code = m.plan_code
`
}

func scanPlan(row interface {
	Scan(dest ...any) error
}) (Plan, error) {
	var item Plan
	err := row.Scan(&item.ID, &item.Code, &item.Name, &item.Status, &item.CreatedAt)
	return item, err
}

func scanMembership(row interface {
	Scan(dest ...any) error
}) (Membership, error) {
	var item Membership
	var expiresAt sql.NullTime
	err := row.Scan(&item.ID, &item.UserID, &item.PlanCode, &item.PlanName, &item.Status, &item.StartedAt, &expiresAt)
	if err == nil && expiresAt.Valid {
		value := expiresAt.Time
		item.ExpiresAt = &value
	}
	return item, err
}
