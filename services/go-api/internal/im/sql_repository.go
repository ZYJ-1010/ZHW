package im

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) EnsureRoom(ctx context.Context, gameID int64, memberIDs []int64, engine string, openIMGroupID string) (Room, error) {
	if engine == "" {
		engine = "local"
	}
	room, err := scanRoom(r.db.QueryRowContext(ctx, `
insert into chat_rooms (game_id, status, engine, openim_group_id, created_at)
values ($1,'active',$2,$3,now())
on conflict (game_id) do update set
  engine = excluded.engine,
  openim_group_id = excluded.openim_group_id
returning id, game_id, status, engine, openim_group_id, archived_at, archive_reason, created_at
`, gameID, engine, nullString(openIMGroupID)))
	if err != nil {
		return Room{}, err
	}
	for _, memberID := range memberIDs {
		if memberID <= 0 {
			continue
		}
		if _, err := r.db.ExecContext(ctx, `
insert into chat_room_members (room_id, user_id, status, created_at)
values ($1,$2,'active',now())
on conflict (room_id, user_id) do update set status = 'active'
`, room.ID, memberID); err != nil {
			return Room{}, err
		}
	}
	room.MemberIDs, err = r.roomMembers(ctx, room.ID)
	return room, err
}

func (r *SQLRepository) RoomByGame(ctx context.Context, gameID int64) (Room, bool, error) {
	room, err := scanRoom(r.db.QueryRowContext(ctx, `
select id, game_id, status, engine, openim_group_id, archived_at, archive_reason, created_at
from chat_rooms
where game_id = $1
`, gameID))
	return r.roomResult(ctx, room, err)
}

func (r *SQLRepository) RoomByID(ctx context.Context, roomID int64) (Room, bool, error) {
	room, err := scanRoom(r.db.QueryRowContext(ctx, `
select id, game_id, status, engine, openim_group_id, archived_at, archive_reason, created_at
from chat_rooms
where id = $1
`, roomID))
	return r.roomResult(ctx, room, err)
}

func (r *SQLRepository) ListRooms(ctx context.Context) ([]Room, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, game_id, status, engine, openim_group_id, archived_at, archive_reason, created_at
from chat_rooms
order by id asc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rooms := make([]Room, 0)
	for rows.Next() {
		room, err := scanRoom(rows)
		if err != nil {
			return nil, err
		}
		room.MemberIDs, err = r.roomMembers(ctx, room.ID)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, rows.Err()
}

func (r *SQLRepository) SaveRoom(ctx context.Context, room Room) (Room, error) {
	saved, err := scanRoom(r.db.QueryRowContext(ctx, `
update chat_rooms set
  status = $2,
  engine = $3,
  openim_group_id = $4,
  archived_at = $5,
  archive_reason = $6
where id = $1
returning id, game_id, status, engine, openim_group_id, archived_at, archive_reason, created_at
`, room.ID, room.Status, nullString(room.Engine), nullString(room.OpenIMGroupID), nullTime(room.ArchivedAt), nullString(room.ArchiveReason)))
	if err != nil {
		return Room{}, err
	}
	saved.MemberIDs, err = r.roomMembers(ctx, saved.ID)
	return saved, err
}

func (r *SQLRepository) SaveMessage(ctx context.Context, message Message) (Message, error) {
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}
	return scanMessage(r.db.QueryRowContext(ctx, `
with inserted as (
  insert into chat_messages (room_id, sender_user_id, message_type, content, file_id, status, created_at)
  values ($1,$2,$3,$4,$5,$6,$7)
  returning id, room_id, sender_user_id, message_type, content, file_id, status, created_at
)
select m.id, m.room_id, r.game_id, m.sender_user_id, m.message_type, m.content, m.file_id, m.status, m.created_at
from inserted m
join chat_rooms r on r.id = m.room_id
`, message.RoomID, message.SenderID, message.Type, nullString(message.Content), nullInt64(message.FileID), message.Status, message.CreatedAt))
}

func (r *SQLRepository) MessageByID(ctx context.Context, roomID int64, messageID int64) (Message, bool, error) {
	message, err := scanMessage(r.db.QueryRowContext(ctx, `
select m.id, m.room_id, r.game_id, m.sender_user_id, m.message_type, m.content, m.file_id, m.status, m.created_at
from chat_messages m
join chat_rooms r on r.id = m.room_id
where m.room_id = $1 and m.id = $2
`, roomID, messageID))
	if errors.Is(err, sql.ErrNoRows) {
		return Message{}, false, nil
	}
	return message, err == nil, err
}

func (r *SQLRepository) ListMessagesByRoom(ctx context.Context, roomID int64) ([]Message, error) {
	rows, err := r.db.QueryContext(ctx, `
select m.id, m.room_id, r.game_id, m.sender_user_id, m.message_type, m.content, m.file_id, m.status, m.created_at
from chat_messages m
join chat_rooms r on r.id = m.room_id
where m.room_id = $1
order by m.created_at asc, m.id asc
`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessages(rows)
}

func (r *SQLRepository) ListMessages(ctx context.Context) ([]Message, error) {
	rows, err := r.db.QueryContext(ctx, `
select m.id, m.room_id, r.game_id, m.sender_user_id, m.message_type, m.content, m.file_id, m.status, m.created_at
from chat_messages m
join chat_rooms r on r.id = m.room_id
order by m.created_at asc, m.id asc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessages(rows)
}

func (r *SQLRepository) UpdateMessage(ctx context.Context, message Message) (Message, error) {
	return scanMessage(r.db.QueryRowContext(ctx, `
with updated as (
  update chat_messages
  set status = $2
  where id = $1
  returning id, room_id, sender_user_id, message_type, content, file_id, status, created_at
)
select m.id, m.room_id, r.game_id, m.sender_user_id, m.message_type, m.content, m.file_id, m.status, m.created_at
from updated m
join chat_rooms r on r.id = m.room_id
`, message.ID, message.Status))
}

func (r *SQLRepository) EnsurePrivateConversation(ctx context.Context, userID int64, targetUserID int64, sourceGameID int64) (PrivateConversation, error) {
	userAID, userBID := orderedPair(userID, targetUserID)
	return scanPrivateConversation(r.db.QueryRowContext(ctx, `
insert into private_chat_conversations (user_a_id, user_b_id, source_game_id, status, created_at, updated_at)
values ($1,$2,$3,'active',now(),now())
on conflict (user_a_id, user_b_id) do update set
  source_game_id = case
    when excluded.source_game_id is not null then excluded.source_game_id
    else private_chat_conversations.source_game_id
  end,
  updated_at = now()
returning id, user_a_id, user_b_id, coalesce(source_game_id, 0), status, created_at, updated_at
`, userAID, userBID, nullInt64(sourceGameID)))
}

func (r *SQLRepository) SavePrivateMessage(ctx context.Context, message PrivateMessage) (PrivateMessage, error) {
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}
	return scanPrivateMessage(r.db.QueryRowContext(ctx, `
insert into private_chat_messages (conversation_id, sender_user_id, target_user_id, source_game_id, message_type, content, status, created_at)
values ($1,$2,$3,$4,$5,$6,$7,$8)
returning id, conversation_id, sender_user_id, target_user_id, coalesce(source_game_id, 0), message_type, content, status, created_at
`, message.ConversationID, message.SenderID, message.TargetID, nullInt64(message.SourceGameID), message.Type, message.Content, message.Status, message.CreatedAt))
}

func (r *SQLRepository) ListPrivateMessages(ctx context.Context, conversationID int64) ([]PrivateMessage, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, conversation_id, sender_user_id, target_user_id, coalesce(source_game_id, 0), message_type, content, status, created_at
from private_chat_messages
where conversation_id = $1
order by created_at asc, id asc
`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPrivateMessages(rows)
}

func (r *SQLRepository) roomResult(ctx context.Context, room Room, err error) (Room, bool, error) {
	if errors.Is(err, sql.ErrNoRows) {
		return Room{}, false, nil
	}
	if err != nil {
		return Room{}, false, err
	}
	room.MemberIDs, err = r.roomMembers(ctx, room.ID)
	return room, err == nil, err
}

func (r *SQLRepository) roomMembers(ctx context.Context, roomID int64) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, `
select user_id
from chat_room_members
where room_id = $1 and status = 'active'
order by created_at asc, user_id asc
`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]int64, 0)
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		items = append(items, userID)
	}
	return items, rows.Err()
}

func scanRoom(row interface {
	Scan(dest ...any) error
}) (Room, error) {
	var room Room
	var engine sql.NullString
	var openIMGroupID sql.NullString
	var archivedAt sql.NullTime
	var archiveReason sql.NullString
	if err := row.Scan(&room.ID, &room.GameID, &room.Status, &engine, &openIMGroupID, &archivedAt, &archiveReason, &room.CreatedAt); err != nil {
		return Room{}, err
	}
	room.Engine = engine.String
	if room.Engine == "" {
		room.Engine = "local"
	}
	room.OpenIMGroupID = openIMGroupID.String
	if archivedAt.Valid {
		room.ArchivedAt = archivedAt.Time.Format(time.RFC3339)
	}
	room.ArchiveReason = archiveReason.String
	return room, nil
}

func scanMessage(row interface {
	Scan(dest ...any) error
}) (Message, error) {
	var message Message
	var content sql.NullString
	var fileID sql.NullInt64
	if err := row.Scan(&message.ID, &message.RoomID, &message.GameID, &message.SenderID, &message.Type, &content, &fileID, &message.Status, &message.CreatedAt); err != nil {
		return Message{}, err
	}
	message.Content = content.String
	message.FileID = fileID.Int64
	return message, nil
}

func scanPrivateConversation(row interface {
	Scan(dest ...any) error
}) (PrivateConversation, error) {
	var conversation PrivateConversation
	if err := row.Scan(&conversation.ID, &conversation.UserAID, &conversation.UserBID, &conversation.SourceGameID, &conversation.Status, &conversation.CreatedAt, &conversation.UpdatedAt); err != nil {
		return PrivateConversation{}, err
	}
	return conversation, nil
}

func scanPrivateMessage(row interface {
	Scan(dest ...any) error
}) (PrivateMessage, error) {
	var message PrivateMessage
	if err := row.Scan(&message.ID, &message.ConversationID, &message.SenderID, &message.TargetID, &message.SourceGameID, &message.Type, &message.Content, &message.Status, &message.CreatedAt); err != nil {
		return PrivateMessage{}, err
	}
	return message, nil
}

func scanPrivateMessages(rows *sql.Rows) ([]PrivateMessage, error) {
	items := make([]PrivateMessage, 0)
	for rows.Next() {
		item, err := scanPrivateMessage(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanMessages(rows *sql.Rows) ([]Message, error) {
	items := make([]Message, 0)
	for rows.Next() {
		item, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func nullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value > 0}
}

func nullTime(value string) sql.NullTime {
	if value == "" {
		return sql.NullTime{}
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: parsed, Valid: true}
}
