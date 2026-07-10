package invites

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) UpsertCode(ctx context.Context, invite InviteCode) (InviteCode, error) {
	return scanInviteCode(r.db.QueryRowContext(ctx, `
insert into invite_codes (code, owner_user_id, status, max_uses, used_count, entry_type, created_at, updated_at)
values ($1,$2,$3,$4,0,$5,now(),now())
on conflict (code) do update set
  owner_user_id = excluded.owner_user_id,
  status = excluded.status,
  max_uses = excluded.max_uses,
  entry_type = excluded.entry_type,
  updated_at = now()
returning id, code, owner_user_id, status, max_uses, used_count, entry_type
`, invite.Code, nullInt64(invite.OwnerID), invite.Status, nullInt(invite.MaxUses), NormalizeEntryType(invite.EntryType)))
}

func (r *SQLRepository) FindCode(ctx context.Context, code string) (InviteCode, bool, error) {
	invite, err := scanInviteCode(r.db.QueryRowContext(ctx, `
select id, code, owner_user_id, status, max_uses, used_count, entry_type
from invite_codes
where code = $1
`, code))
	if err == sql.ErrNoRows {
		return InviteCode{}, false, nil
	}
	if err != nil {
		return InviteCode{}, false, err
	}
	return invite, true, nil
}

func (r *SQLRepository) ListCodes(ctx context.Context, filter CodeFilter) ([]InviteCode, error) {
	where := []string{"1=1"}
	args := []any{}
	if strings.TrimSpace(filter.Status) != "" {
		args = append(args, strings.TrimSpace(filter.Status))
		where = append(where, "ic.status = $"+strconv.Itoa(len(args)))
	}
	if entryType, ok := ParseEntryType(filter.EntryType); ok && entryType != "" {
		args = append(args, entryType)
		where = append(where, "ic.entry_type = $"+strconv.Itoa(len(args)))
	}
	if filter.OwnerID > 0 {
		args = append(args, filter.OwnerID)
		where = append(where, "ic.owner_user_id = $"+strconv.Itoa(len(args)))
	}
	rows, err := r.db.QueryContext(ctx, `
select ic.id, ic.code, ic.owner_user_id, ic.status, ic.max_uses, ic.used_count, ic.entry_type,
       coalesce(bound.invitee_user_id, 0), coalesce(u.nickname, '')
from invite_codes ic
left join lateral (
  select ir.invitee_user_id
  from invite_relations ir
  where ir.invite_code_id = ic.id
  order by ir.id asc
  limit 1
) bound on true
left join users u on u.id = bound.invitee_user_id
where `+strings.Join(where, " and ")+`
order by ic.id desc
limit 500
`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]InviteCode, 0)
	for rows.Next() {
		var invite InviteCode
		var ownerID sql.NullInt64
		var maxUses sql.NullInt64
		var entryType sql.NullString
		var boundUserID sql.NullInt64
		var nickname sql.NullString
		if err := rows.Scan(&invite.ID, &invite.Code, &ownerID, &invite.Status, &maxUses, &invite.UsedCount, &entryType, &boundUserID, &nickname); err != nil {
			return nil, err
		}
		invite.OwnerID = ownerID.Int64
		invite.MaxUses = int(maxUses.Int64)
		invite.EntryType = NormalizeEntryType(entryType.String)
		invite.BoundWechatUserID = boundUserID.Int64
		invite.BoundWechatNickname = nickname.String
		result = append(result, invite)
	}
	return result, rows.Err()
}

func (r *SQLRepository) Bind(ctx context.Context, invite InviteCode, inviteeUserID int64, source string) (Relation, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Relation{}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `select id from invite_codes where id = $1 for update`, invite.ID); err != nil {
		return Relation{}, err
	}

	relation, ok, err := r.relationForUserTx(ctx, tx, inviteeUserID)
	if err != nil {
		return Relation{}, err
	}
	if ok {
		if err := tx.Commit(); err != nil {
			return Relation{}, err
		}
		return relation, nil
	}

	if invite.MaxUses == 1 {
		if bound, ok, err := r.findBoundCodeTx(ctx, tx, invite.ID); err != nil {
			return Relation{}, err
		} else if ok {
			if bound.BoundWechatUserID != inviteeUserID {
				return Relation{}, ErrInviteAlreadyBound
			}
			relation, err := scanRelation(tx.QueryRowContext(ctx, `
select invite_code_id, inviter_user_id, invitee_user_id, bind_source
from invite_relations
where invite_code_id = $1
limit 1
`, invite.ID))
			if err != nil {
				return Relation{}, err
			}
			if err := tx.Commit(); err != nil {
				return Relation{}, err
			}
			return relation, nil
		}
	}
	relation, err = scanRelation(tx.QueryRowContext(ctx, `
insert into invite_relations (invite_code_id, inviter_user_id, invitee_user_id, bind_source, created_at)
values ($1,$2,$3,$4,now())
returning invite_code_id, inviter_user_id, invitee_user_id, bind_source
`, invite.ID, nullInt64(invite.OwnerID), inviteeUserID, source))
	if err != nil {
		return Relation{}, err
	}

	_, err = tx.ExecContext(ctx, `
update invite_codes
set used_count = used_count + 1, updated_at = now()
where id = $1
`, invite.ID)
	if err != nil {
		return Relation{}, err
	}
	if err := tx.Commit(); err != nil {
		return Relation{}, err
	}
	return relation, nil
}

func (r *SQLRepository) RelationForUser(ctx context.Context, userID int64) (Relation, bool, error) {
	return r.relationForUserTx(ctx, r.db, userID)
}

func (r *SQLRepository) ListRelations(ctx context.Context, filter RelationFilter) ([]Relation, error) {
	where := []string{"1=1"}
	args := []any{}
	if filter.InviterUserID > 0 {
		args = append(args, filter.InviterUserID)
		where = append(where, "ir.inviter_user_id = $"+strconv.Itoa(len(args)))
	}
	if filter.InviteeUserID > 0 {
		args = append(args, filter.InviteeUserID)
		where = append(where, "ir.invitee_user_id = $"+strconv.Itoa(len(args)))
	}
	if filter.InviteCodeID > 0 {
		args = append(args, filter.InviteCodeID)
		where = append(where, "ir.invite_code_id = $"+strconv.Itoa(len(args)))
	}
	rows, err := r.db.QueryContext(ctx, `
select ir.invite_code_id, ir.inviter_user_id, ir.invitee_user_id, ir.bind_source
from invite_relations ir
where `+strings.Join(where, " and ")+`
order by ir.id desc
limit 500
`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Relation, 0)
	for rows.Next() {
		var relation Relation
		var inviterID sql.NullInt64
		if err := rows.Scan(&relation.InviteCodeID, &inviterID, &relation.InviteeUserID, &relation.BindSource); err != nil {
			return nil, err
		}
		relation.InviterUserID = inviterID.Int64
		result = append(result, relation)
	}
	return result, rows.Err()
}

func (r *SQLRepository) relationForUserTx(ctx context.Context, q queryRower, userID int64) (Relation, bool, error) {
	relation, err := scanRelation(q.QueryRowContext(ctx, `
select invite_code_id, inviter_user_id, invitee_user_id, bind_source
from invite_relations
where invitee_user_id = $1
order by id asc
limit 1
`, userID))
	if err == sql.ErrNoRows {
		return Relation{}, false, nil
	}
	if err != nil {
		return Relation{}, false, err
	}
	return relation, true, nil
}

func (r *SQLRepository) FindBoundCode(ctx context.Context, inviteCodeID int64) (InviteCode, bool, error) {
	return r.findBoundCodeTx(ctx, r.db, inviteCodeID)
}

func (r *SQLRepository) findBoundCodeTx(ctx context.Context, q queryRower, inviteCodeID int64) (InviteCode, bool, error) {
	invite, err := scanBoundInviteCode(q.QueryRowContext(ctx, `
select ic.id, ic.code, ic.owner_user_id, ic.status, ic.max_uses, ic.used_count,
       ic.entry_type, ir.invitee_user_id, coalesce(wa.openid, ''), coalesce(u.nickname, '')
from invite_codes ic
join invite_relations ir on ir.invite_code_id = ic.id
left join user_wechat_accounts wa on wa.user_id = ir.invitee_user_id
left join users u on u.id = ir.invitee_user_id
where ic.id = $1
order by ir.id asc, wa.id asc
limit 1
`, inviteCodeID))
	if err == sql.ErrNoRows {
		return InviteCode{}, false, nil
	}
	if err != nil {
		return InviteCode{}, false, err
	}
	return invite, true, nil
}

type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func scanInviteCode(row interface {
	Scan(dest ...any) error
}) (InviteCode, error) {
	var invite InviteCode
	var ownerID sql.NullInt64
	var maxUses sql.NullInt64
	var entryType sql.NullString
	err := row.Scan(&invite.ID, &invite.Code, &ownerID, &invite.Status, &maxUses, &invite.UsedCount, &entryType)
	invite.OwnerID = ownerID.Int64
	invite.MaxUses = int(maxUses.Int64)
	invite.EntryType = NormalizeEntryType(entryType.String)
	return invite, err
}

func scanBoundInviteCode(row interface {
	Scan(dest ...any) error
}) (InviteCode, error) {
	var invite InviteCode
	var ownerID sql.NullInt64
	var maxUses sql.NullInt64
	var entryType sql.NullString
	var boundUserID sql.NullInt64
	var openID sql.NullString
	var nickname sql.NullString
	err := row.Scan(&invite.ID, &invite.Code, &ownerID, &invite.Status, &maxUses, &invite.UsedCount, &entryType, &boundUserID, &openID, &nickname)
	if errors.Is(err, sql.ErrNoRows) {
		return InviteCode{}, err
	}
	invite.OwnerID = ownerID.Int64
	invite.MaxUses = int(maxUses.Int64)
	invite.EntryType = NormalizeEntryType(entryType.String)
	invite.BoundWechatUserID = boundUserID.Int64
	invite.BoundWechatOpenID = openID.String
	invite.BoundWechatNickname = nickname.String
	return invite, err
}

func scanRelation(row interface {
	Scan(dest ...any) error
}) (Relation, error) {
	var relation Relation
	var inviterID sql.NullInt64
	err := row.Scan(&relation.InviteCodeID, &inviterID, &relation.InviteeUserID, &relation.BindSource)
	relation.InviterUserID = inviterID.Int64
	return relation, err
}

func nullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value > 0}
}

func nullInt(value int) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(value), Valid: value > 0}
}
