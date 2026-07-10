package lbs

import (
	"context"
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

type fakeLocationRepository struct {
	items []Location
}

func newFakeLocationRepository() *fakeLocationRepository {
	return &fakeLocationRepository{}
}

func (r *fakeLocationRepository) SaveLocation(ctx context.Context, location Location) (Location, error) {
	r.items = append([]Location{location}, r.items...)
	return location, nil
}

func (r *fakeLocationRepository) CurrentLocation(ctx context.Context, userID int64) (Location, bool, error) {
	for _, item := range r.items {
		if item.UserID == userID {
			return item, true, nil
		}
	}
	return Location{}, false, nil
}

func (r *fakeLocationRepository) RecentLocations(ctx context.Context, userID int64, limit int) ([]Location, error) {
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
