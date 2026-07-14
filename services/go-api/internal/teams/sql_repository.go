package teams

import (
	"context"
	"database/sql"
	"errors"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) SaveTeam(ctx context.Context, team Team) (Team, error) {
	return scanTeam(r.db.QueryRowContext(ctx, `
insert into teams (leader_user_id, name, status, created_at)
values ($1,$2,$3,$4)
on conflict (leader_user_id) do update set
  name = excluded.name,
  status = excluded.status
returning id, leader_user_id, name, status, created_at
`, team.LeaderUserID, team.Name, team.Status, team.CreatedAt))
}

func (r *SQLRepository) FindTeamByLeader(ctx context.Context, leaderUserID int64) (Team, bool, error) {
	team, err := scanTeam(r.db.QueryRowContext(ctx, teamSelect()+` where leader_user_id = $1`, leaderUserID))
	if errors.Is(err, sql.ErrNoRows) {
		return Team{}, false, nil
	}
	if err != nil {
		return Team{}, false, err
	}
	return team, true, nil
}

func (r *SQLRepository) FindTeamByID(ctx context.Context, teamID int64) (Team, bool, error) {
	team, err := scanTeam(r.db.QueryRowContext(ctx, teamSelect()+` where id = $1`, teamID))
	if errors.Is(err, sql.ErrNoRows) {
		return Team{}, false, nil
	}
	if err != nil {
		return Team{}, false, err
	}
	return team, true, nil
}

func (r *SQLRepository) ListTeams(ctx context.Context) ([]Team, error) {
	rows, err := r.db.QueryContext(ctx, teamSelect()+` order by created_at desc, id desc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Team, 0)
	for rows.Next() {
		item, err := scanTeam(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) SaveMember(ctx context.Context, leaderUserID int64, member Member) (Member, error) {
	return scanMember(r.db.QueryRowContext(ctx, `
insert into team_relations (leader_user_id, member_user_id, relation_level, source, status, created_at)
values ($1,$2,$3,$4,$5,$6)
on conflict (leader_user_id, member_user_id, relation_level) do update set
  source = excluded.source,
  status = excluded.status
returning member_user_id, relation_level, source, status, created_at
`, leaderUserID, member.UserID, member.RelationLevel, member.Source, member.Status, member.JoinedAt))
}

func (r *SQLRepository) ListMembers(ctx context.Context, leaderUserID int64) ([]Member, error) {
	rows, err := r.db.QueryContext(ctx, `
select member_user_id, relation_level, source, status, created_at
from team_relations
where leader_user_id = $1
order by created_at asc, id asc
`, leaderUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Member, 0)
	for rows.Next() {
		item, err := scanMember(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func teamSelect() string {
	return `select id, leader_user_id, name, status, created_at from teams`
}

func scanTeam(row interface {
	Scan(dest ...any) error
}) (Team, error) {
	var item Team
	err := row.Scan(&item.ID, &item.LeaderUserID, &item.Name, &item.Status, &item.CreatedAt)
	return item, err
}

func scanMember(row interface {
	Scan(dest ...any) error
}) (Member, error) {
	var item Member
	err := row.Scan(&item.UserID, &item.RelationLevel, &item.Source, &item.Status, &item.JoinedAt)
	return item, err
}
