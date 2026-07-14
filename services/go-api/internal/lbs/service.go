package lbs

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

var ErrInvalidLocation = errors.New("invalid location")

type Location struct {
	UserID          int64     `json:"userId"`
	Longitude       float64   `json:"longitude"`
	Latitude        float64   `json:"latitude"`
	AccuracyMeter   float64   `json:"accuracyMeter"`
	AccuracyWarning bool      `json:"accuracyWarning"`
	CityCode        string    `json:"cityCode"`
	CityName        string    `json:"cityName"`
	Address         string    `json:"address,omitempty"`
	Source          string    `json:"source"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type SaveRequest struct {
	Longitude     float64 `json:"longitude"`
	Latitude      float64 `json:"latitude"`
	AccuracyMeter float64 `json:"accuracyMeter"`
	CityCode      string  `json:"cityCode"`
	CityName      string  `json:"cityName"`
	Address       string  `json:"address"`
}

type Repository interface {
	SaveLocation(ctx context.Context, location Location) (Location, error)
	CurrentLocation(ctx context.Context, userID int64) (Location, bool, error)
	RecentLocations(ctx context.Context, userID int64, limit int) ([]Location, error)
	RecentLocationsBySource(ctx context.Context, userID int64, source string, limit int) ([]Location, error)
	LatestLocations(ctx context.Context, limit int) ([]Location, error)
}

type Service struct {
	mu        sync.RWMutex
	locations map[int64]Location
	history   map[int64][]Location
	repo      Repository
}

func NewService() *Service {
	return NewServiceWithRepository(nil)
}

func NewServiceWithRepository(repo Repository) *Service {
	return &Service{locations: make(map[int64]Location), history: make(map[int64][]Location), repo: repo}
}

func (s *Service) SaveCurrent(userID int64, req SaveRequest) (Location, error) {
	return s.save(userID, req, "gps")
}

func (s *Service) SaveManual(userID int64, req SaveRequest) (Location, error) {
	return s.save(userID, req, "manual")
}

func (s *Service) Current(userID int64) (Location, bool) {
	if s.repo != nil {
		location, ok, err := s.repo.CurrentLocation(context.Background(), userID)
		if err == nil {
			return location, ok
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.history[userID] {
		if item.Source == "gps" {
			return item, true
		}
	}
	location, ok := s.locations[userID]
	return location, ok
}

func (s *Service) Recent(userID int64, limit int) []Location {
	if limit <= 0 {
		limit = 10
	}
	if s.repo != nil {
		if items, err := s.repo.RecentLocations(context.Background(), userID, limit); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := append([]Location(nil), s.history[userID]...)
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}

func (s *Service) RecentBySource(userID int64, source string, limit int) []Location {
	source = strings.TrimSpace(source)
	if limit <= 0 {
		limit = 10
	}
	if source == "" {
		return s.Recent(userID, limit)
	}
	if s.repo != nil {
		if items, err := s.repo.RecentLocationsBySource(context.Background(), userID, source, limit); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Location, 0, limit)
	for _, item := range s.history[userID] {
		if item.Source != source {
			continue
		}
		items = append(items, item)
		if len(items) == limit {
			break
		}
	}
	return items
}

func (s *Service) NearbyUsers(userID int64, center Location, radiusMeter float64, limit int) []Location {
	if radiusMeter <= 0 || limit <= 0 {
		return nil
	}

	locations := s.latestLocations(limit * 4)
	items := make([]Location, 0, limit)
	for _, item := range locations {
		if item.UserID == userID {
			continue
		}
		distance := DistanceMeter(center.Longitude, center.Latitude, item.Longitude, item.Latitude)
		if distance > radiusMeter {
			continue
		}
		item.AccuracyMeter = math.Round(distance)
		items = append(items, item)
		if len(items) == limit {
			break
		}
	}
	return items
}

func (s *Service) latestLocations(limit int) []Location {
	if limit <= 0 {
		limit = 50
	}
	if s.repo != nil {
		if items, err := s.repo.LatestLocations(context.Background(), limit); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Location, 0, len(s.locations))
	for _, item := range s.locations {
		items = append(items, item)
	}
	return items
}

func (s *Service) save(userID int64, req SaveRequest, source string) (Location, error) {
	req, err := validateSaveRequest(req)
	if err != nil {
		return Location{}, err
	}
	location := Location{
		UserID:          userID,
		Longitude:       req.Longitude,
		Latitude:        req.Latitude,
		AccuracyMeter:   req.AccuracyMeter,
		AccuracyWarning: req.AccuracyMeter > 300,
		CityCode:        req.CityCode,
		CityName:        req.CityName,
		Address:         req.Address,
		Source:          source,
		UpdatedAt:       time.Now(),
	}
	if s.repo != nil {
		if saved, err := s.repo.SaveLocation(context.Background(), location); err == nil {
			return saved, nil
		}
	}
	s.mu.Lock()
	s.locations[userID] = location
	s.history[userID] = append([]Location{location}, s.history[userID]...)
	s.mu.Unlock()
	return location, nil
}

func validateSaveRequest(req SaveRequest) (SaveRequest, error) {
	req.CityCode = strings.TrimSpace(req.CityCode)
	req.CityName = strings.TrimSpace(req.CityName)
	req.Address = strings.TrimSpace(req.Address)
	if math.IsNaN(req.Longitude) || math.IsInf(req.Longitude, 0) || req.Longitude < -180 || req.Longitude > 180 {
		return req, ErrInvalidLocation
	}
	if math.IsNaN(req.Latitude) || math.IsInf(req.Latitude, 0) || req.Latitude < -90 || req.Latitude > 90 {
		return req, ErrInvalidLocation
	}
	if math.IsNaN(req.AccuracyMeter) || math.IsInf(req.AccuracyMeter, 0) || req.AccuracyMeter < 0 || req.AccuracyMeter > 100000 {
		return req, ErrInvalidLocation
	}
	if len(req.CityCode) > 32 || len(req.CityName) > 64 || len(req.Address) > 256 {
		return req, ErrInvalidLocation
	}
	return req, nil
}

func DistanceMeter(lon1 float64, lat1 float64, lon2 float64, lat2 float64) float64 {
	const earthRadius = 6371000
	rad := math.Pi / 180
	dLat := (lat2 - lat1) * rad
	dLon := (lon2 - lon1) * rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

func FormatDistanceLabel(distanceMeter float64) string {
	if math.IsNaN(distanceMeter) || math.IsInf(distanceMeter, 0) || distanceMeter < 0 {
		return ""
	}
	if distanceMeter < 1000 {
		return fmt.Sprintf("%.0fm", math.Round(distanceMeter))
	}
	return fmt.Sprintf("%.1fkm", distanceMeter/1000)
}
