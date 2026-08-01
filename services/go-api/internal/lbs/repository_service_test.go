package lbs

import (
	"context"
	"errors"
	"testing"
)

func TestRepositoryBackedLocationSaveCurrentAndRecent(t *testing.T) {
	service := NewServiceWithRepository(newFakeLocationRepository())

	first, err := service.SaveCurrent(1, SaveRequest{Longitude: 116.397128, Latitude: 39.916527, AccuracyMeter: 80, CityCode: "110100", CityName: "Beijing"})
	if err != nil {
		t.Fatal(err)
	}
	if first.AccuracyWarning {
		t.Fatalf("expected accurate gps location, got %+v", first)
	}
	second, err := service.SaveManual(1, SaveRequest{Longitude: 116.4, Latitude: 39.9, AccuracyMeter: 500, CityCode: "110100", CityName: "Beijing", Address: "manual"})
	if err != nil {
		t.Fatal(err)
	}
	if !second.AccuracyWarning || second.Source != "manual" {
		t.Fatalf("expected manual warning location, got %+v", second)
	}

	current, ok := service.Current(1)
	if !ok || current.Address != "manual" {
		t.Fatalf("expected latest current location, ok=%v current=%+v", ok, current)
	}
	recent := service.Recent(1, 1)
	if len(recent) != 1 || recent[0].Source != "manual" {
		t.Fatalf("expected latest one recent location, got %+v", recent)
	}
}

func TestLocationValidationRejectsInvalidInput(t *testing.T) {
	service := NewService()

	cases := []SaveRequest{
		{Longitude: 181, Latitude: 39.9, AccuracyMeter: 80},
		{Longitude: 116.4, Latitude: -91, AccuracyMeter: 80},
		{Longitude: 116.4, Latitude: 39.9, AccuracyMeter: -1},
		{Longitude: 116.4, Latitude: 39.9, AccuracyMeter: 100001},
		{Longitude: 116.4, Latitude: 39.9, AccuracyMeter: 80, CityCode: "123456789012345678901234567890123"},
	}
	for _, req := range cases {
		if _, err := service.SaveCurrent(1, req); err == nil {
			t.Fatalf("expected invalid location error for %+v", req)
		}
	}
	if _, ok := service.Current(1); ok {
		t.Fatal("invalid location should not be saved")
	}
}

func TestNearbyUsersReturnsOtherLatestLocationsWithinRadius(t *testing.T) {
	service := NewServiceWithRepository(newFakeLocationRepository())
	center, err := service.SaveCurrent(1, SaveRequest{Longitude: 116.397128, Latitude: 39.916527, AccuracyMeter: 80})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SaveCurrent(2, SaveRequest{Longitude: 116.3972, Latitude: 39.9166, AccuracyMeter: 80}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SaveCurrent(3, SaveRequest{Longitude: 117.3972, Latitude: 40.9166, AccuracyMeter: 80}); err != nil {
		t.Fatal(err)
	}

	items := service.NearbyUsers(1, center, 1000, 10)
	if len(items) != 1 || items[0].UserID != 2 || items[0].AccuracyMeter <= 0 {
		t.Fatalf("expected only nearby other user with distance, got %+v", items)
	}
}

func TestRepositoryFailureDoesNotUseStaleLocationCache(t *testing.T) {
	repo := newFakeLocationRepository()
	service := NewServiceWithRepository(repo)
	service.locations[1] = Location{UserID: 1, CityName: "旧城市", Source: "manual"}
	service.history[1] = []Location{{UserID: 1, CityName: "旧城市", Source: "manual"}}
	repo.err = errors.New("database unavailable")

	if _, err := service.SaveManual(1, SaveRequest{Longitude: 120, Latitude: 30}); !errors.Is(err, repo.err) {
		t.Fatalf("expected repository save error, got %v", err)
	}
	if _, ok, err := service.CurrentStrict(1); ok || !errors.Is(err, repo.err) {
		t.Fatalf("expected strict current error without stale cache, ok=%v err=%v", ok, err)
	}
	if _, err := service.RecentStrict(1, 10); !errors.Is(err, repo.err) {
		t.Fatalf("expected strict recent error, got %v", err)
	}
	if _, err := service.RecentBySourceStrict(1, "manual", 10); !errors.Is(err, repo.err) {
		t.Fatalf("expected strict recent-by-source error, got %v", err)
	}
	if _, err := service.NearbyUsersStrict(1, Location{Longitude: 120, Latitude: 30}, 1000, 10); !errors.Is(err, repo.err) {
		t.Fatalf("expected strict nearby error, got %v", err)
	}
}

func TestMemoryCurrentUsesLatestManualSelection(t *testing.T) {
	service := NewService()
	if _, err := service.SaveCurrent(1, SaveRequest{Longitude: 116, Latitude: 39, CityName: "北京"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SaveManual(1, SaveRequest{Longitude: 120, Latitude: 30, CityName: "杭州"}); err != nil {
		t.Fatal(err)
	}
	current, ok, err := service.CurrentStrict(1)
	if err != nil || !ok || current.Source != "manual" || current.CityName != "杭州" {
		t.Fatalf("expected latest manual location, ok=%v err=%v current=%+v", ok, err, current)
	}
}

type fakeLocationRepository struct {
	items []Location
	err   error
}

func newFakeLocationRepository() *fakeLocationRepository {
	return &fakeLocationRepository{}
}

func (r *fakeLocationRepository) SaveLocation(ctx context.Context, location Location) (Location, error) {
	if r.err != nil {
		return Location{}, r.err
	}
	r.items = append([]Location{location}, r.items...)
	return location, nil
}

func (r *fakeLocationRepository) CurrentLocation(ctx context.Context, userID int64) (Location, bool, error) {
	if r.err != nil {
		return Location{}, false, r.err
	}
	for _, item := range r.items {
		if item.UserID == userID {
			return item, true, nil
		}
	}
	return Location{}, false, nil
}

func (r *fakeLocationRepository) RecentLocations(ctx context.Context, userID int64, limit int) ([]Location, error) {
	if r.err != nil {
		return nil, r.err
	}
	result := make([]Location, 0)
	for _, item := range r.items {
		if item.UserID == userID {
			result = append(result, item)
			if len(result) == limit {
				break
			}
		}
	}
	return result, nil
}

func (r *fakeLocationRepository) RecentLocationsBySource(ctx context.Context, userID int64, source string, limit int) ([]Location, error) {
	if r.err != nil {
		return nil, r.err
	}
	result := make([]Location, 0)
	for _, item := range r.items {
		if item.UserID == userID && item.Source == source {
			result = append(result, item)
			if len(result) == limit {
				break
			}
		}
	}
	return result, nil
}

func (r *fakeLocationRepository) LatestLocations(ctx context.Context, limit int) ([]Location, error) {
	if r.err != nil {
		return nil, r.err
	}
	result := make([]Location, 0)
	seen := make(map[int64]bool)
	for _, item := range r.items {
		if seen[item.UserID] {
			continue
		}
		seen[item.UserID] = true
		result = append(result, item)
		if len(result) == limit {
			break
		}
	}
	return result, nil
}
