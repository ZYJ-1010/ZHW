package lbs

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

func (r *SQLRepository) SaveLocation(ctx context.Context, location Location) (Location, error) {
	return scanLocation(r.db.QueryRowContext(ctx, `
insert into user_location_records (
  user_id, longitude, latitude, accuracy_meter, accuracy_warning,
  city_code, city_name, address, source, created_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
returning user_id, longitude, latitude, accuracy_meter, accuracy_warning, city_code, city_name, address, source, created_at
`, location.UserID, location.Longitude, location.Latitude, location.AccuracyMeter, location.AccuracyWarning, nullString(location.CityCode), nullString(location.CityName), nullString(location.Address), location.Source, location.UpdatedAt))
}

func (r *SQLRepository) CurrentLocation(ctx context.Context, userID int64) (Location, bool, error) {
	location, err := scanLocation(r.db.QueryRowContext(ctx, locationSelect()+` where user_id = $1 order by created_at desc, id desc limit 1`, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return Location{}, false, nil
	}
	if err != nil {
		return Location{}, false, err
	}
	return location, true, nil
}

func (r *SQLRepository) RecentLocations(ctx context.Context, userID int64, limit int) ([]Location, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.db.QueryContext(ctx, locationSelect()+` where user_id = $1 order by created_at desc, id desc limit $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Location, 0)
	for rows.Next() {
		item, err := scanLocation(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) RecentLocationsBySource(ctx context.Context, userID int64, source string, limit int) ([]Location, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.db.QueryContext(ctx, locationSelect()+` where user_id = $1 and source = $2 order by created_at desc, id desc limit $3`, userID, source, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Location, 0)
	for rows.Next() {
		item, err := scanLocation(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) LatestLocations(ctx context.Context, limit int) ([]Location, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `
select user_id, longitude, latitude, accuracy_meter, accuracy_warning, city_code, city_name, address, source, created_at
from (
  select distinct on (user_id) user_id, longitude, latitude, accuracy_meter, accuracy_warning, city_code, city_name, address, source, created_at, id
  from user_location_records
  order by user_id, created_at desc, id desc
) latest
order by created_at desc, id desc
limit $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Location, 0)
	for rows.Next() {
		item, err := scanLocation(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func locationSelect() string {
	return `select user_id, longitude, latitude, accuracy_meter, accuracy_warning, city_code, city_name, address, source, created_at from user_location_records`
}

func scanLocation(row interface {
	Scan(dest ...any) error
}) (Location, error) {
	var item Location
	var cityCode sql.NullString
	var cityName sql.NullString
	var address sql.NullString
	if err := row.Scan(&item.UserID, &item.Longitude, &item.Latitude, &item.AccuracyMeter, &item.AccuracyWarning, &cityCode, &cityName, &address, &item.Source, &item.UpdatedAt); err != nil {
		return Location{}, err
	}
	item.CityCode = cityCode.String
	item.CityName = cityName.String
	item.Address = address.String
	return item, nil
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}
