package aidata

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"zhw-mini/services/go-api/internal/audit"
	"zhw-mini/services/go-api/internal/connections"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/im"
	"zhw-mini/services/go-api/internal/profiles"
	"zhw-mini/services/go-api/internal/reviews"
)

var ErrIMExportDisabled = errors.New("ai im export disabled")

type Service struct {
	mu            sync.RWMutex
	allowIMExport bool
	repository    Repository
}

func NewService() *Service {
	return &Service{}
}

type Repository interface {
	IMExportConfig(ctx context.Context) (IMExportConfig, error)
	SetIMExportEnabled(ctx context.Context, enabled bool) (IMExportConfig, error)
}

func NewServiceWithRepository(repository Repository) *Service {
	return &Service{repository: repository}
}

type IMExportConfig struct {
	Enabled   bool   `json:"enabled"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	UpdatedBy string `json:"updatedBy,omitempty"`
}

type SnapshotInput struct {
	BehaviorLogs   []audit.BehaviorLog
	Games          []games.Game
	Favorites      []games.Favorite
	Reviews        []reviews.Review
	Footprints     []reviews.Footprint
	Connections    []connections.Connection
	ExpertSkills   []profiles.ExpertSkillProfile
	GuideResources []profiles.GuideResourceProfile
	IMMessages     []im.Message
}

type Snapshot struct {
	UserCount             int                    `json:"userCount"`
	GameCount             int                    `json:"gameCount"`
	BehaviorLogCount      int                    `json:"behaviorLogCount"`
	FavoriteCount         int                    `json:"favoriteCount"`
	ReviewCount           int                    `json:"reviewCount"`
	FootprintCount        int                    `json:"footprintCount"`
	ConnectionCount       int                    `json:"connectionCount"`
	ExpertProfileCount    int                    `json:"expertProfileCount"`
	GuideProfileCount     int                    `json:"guideProfileCount"`
	IMMessageCount        int                    `json:"imMessageCount"`
	BehaviorStats         BehaviorStats          `json:"behaviorStats"`
	FavoriteStats         FavoriteStats          `json:"favoriteStats"`
	ProfileStats          ProfileStats           `json:"profileStats"`
	ConnectionStats       ConnectionStats        `json:"connectionStats"`
	DataReadinessSections map[string]SectionInfo `json:"dataReadinessSections"`
	DataReady             bool                   `json:"dataReady"`
	AcceptanceReady       bool                   `json:"acceptanceReady"`
	AcceptanceChecks      Checks                 `json:"acceptanceChecks"`
	IMExportEnabled       bool                   `json:"imExportEnabled"`
	UpdatedAt             time.Time              `json:"updatedAt"`
}

type Checks struct {
	Users        Check `json:"users"`
	Games        Check `json:"games"`
	BehaviorLogs Check `json:"behaviorLogs"`
	Favorites    Check `json:"favorites"`
	Reviews      Check `json:"reviews"`
}

type Check struct {
	Current  int  `json:"current"`
	Required int  `json:"required"`
	Ready    bool `json:"ready"`
}

type BehaviorStats struct {
	ByDate      map[string]int `json:"byDate"`
	ByUser      map[string]int `json:"byUser"`
	ByEventType map[string]int `json:"byEventType"`
}

type FavoriteStats struct {
	ByUser     map[string]int `json:"byUser"`
	ByGameType map[string]int `json:"byGameType"`
}

type ProfileStats struct {
	ExpertCompletenessAverage float64 `json:"expertCompletenessAverage"`
	GuideCompletenessAverage  float64 `json:"guideCompletenessAverage"`
	ExpertCompleteCount       int     `json:"expertCompleteCount"`
	GuideCompleteCount        int     `json:"guideCompleteCount"`
}

type ConnectionStats struct {
	BySourceType   map[string]int `json:"bySourceType"`
	ByRelationType map[string]int `json:"byRelationType"`
}

type SectionInfo struct {
	Count int  `json:"count"`
	Ready bool `json:"ready"`
}

type IMExportItem struct {
	MessageID   int64     `json:"messageId"`
	RoomID      int64     `json:"roomId"`
	GameID      int64     `json:"gameId"`
	SenderID    int64     `json:"senderUserId"`
	MessageType string    `json:"messageType"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (s *Service) Snapshot(input SnapshotInput) Snapshot {
	config := s.IMExportConfig()
	snapshot := Snapshot{
		UserCount:          uniqueUserCount(input),
		GameCount:          len(input.Games),
		BehaviorLogCount:   len(input.BehaviorLogs),
		FavoriteCount:      len(input.Favorites),
		ReviewCount:        len(input.Reviews),
		FootprintCount:     len(input.Footprints),
		ConnectionCount:    len(input.Connections),
		ExpertProfileCount: len(input.ExpertSkills),
		GuideProfileCount:  len(input.GuideResources),
		IMMessageCount:     len(input.IMMessages),
		BehaviorStats:      behaviorStats(input.BehaviorLogs),
		FavoriteStats:      favoriteStats(input.Favorites, input.Games),
		ProfileStats:       profileStats(input.ExpertSkills, input.GuideResources),
		ConnectionStats:    connectionStats(input.Connections),
		IMExportEnabled:    config.Enabled,
		UpdatedAt:          time.Now(),
	}
	snapshot.DataReadinessSections = map[string]SectionInfo{
		"behaviorLogs":   {Count: snapshot.BehaviorLogCount, Ready: snapshot.BehaviorLogCount > 0},
		"favorites":      {Count: snapshot.FavoriteCount, Ready: snapshot.FavoriteCount > 0},
		"reviews":        {Count: snapshot.ReviewCount, Ready: snapshot.ReviewCount > 0},
		"connections":    {Count: snapshot.ConnectionCount, Ready: snapshot.ConnectionCount > 0},
		"expertProfiles": {Count: snapshot.ExpertProfileCount, Ready: snapshot.ExpertProfileCount > 0},
		"guideProfiles":  {Count: snapshot.GuideProfileCount, Ready: snapshot.GuideProfileCount > 0},
	}
	snapshot.DataReady = snapshot.BehaviorLogCount > 0 ||
		snapshot.FavoriteCount > 0 ||
		snapshot.ReviewCount > 0 ||
		snapshot.FootprintCount > 0 ||
		snapshot.ConnectionCount > 0 ||
		snapshot.ExpertProfileCount > 0 ||
		snapshot.GuideProfileCount > 0 ||
		snapshot.IMMessageCount > 0
	snapshot.AcceptanceChecks = Checks{
		Users:        readiness(snapshot.UserCount, 5),
		Games:        readiness(snapshot.GameCount, 3),
		BehaviorLogs: readiness(snapshot.BehaviorLogCount, 20),
		Favorites:    readiness(snapshot.FavoriteCount, 5),
		Reviews:      readiness(snapshot.ReviewCount, 5),
	}
	snapshot.AcceptanceReady = snapshot.AcceptanceChecks.Users.Ready &&
		snapshot.AcceptanceChecks.Games.Ready &&
		snapshot.AcceptanceChecks.BehaviorLogs.Ready &&
		snapshot.AcceptanceChecks.Favorites.Ready &&
		snapshot.AcceptanceChecks.Reviews.Ready
	return snapshot
}

func (s *Service) ExportIMMessages(messages []im.Message) ([]IMExportItem, error) {
	if !s.IMExportConfig().Enabled {
		return nil, ErrIMExportDisabled
	}
	items := make([]IMExportItem, 0, len(messages))
	for _, message := range messages {
		items = append(items, IMExportItem{
			MessageID:   message.ID,
			RoomID:      message.RoomID,
			GameID:      message.GameID,
			SenderID:    message.SenderID,
			MessageType: message.Type,
			Content:     message.Content,
			CreatedAt:   message.CreatedAt,
		})
	}
	return items, nil
}

func (s *Service) IMExportConfig() IMExportConfig {
	if s.repository != nil {
		config, err := s.repository.IMExportConfig(context.Background())
		if err == nil {
			return config
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return IMExportConfig{Enabled: s.allowIMExport}
}

func (s *Service) SetIMExportEnabled(enabled bool) IMExportConfig {
	if s.repository != nil {
		config, err := s.repository.SetIMExportEnabled(context.Background(), enabled)
		if err == nil {
			return config
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.allowIMExport = enabled
	return IMExportConfig{Enabled: s.allowIMExport}
}

func readiness(current int, required int) Check {
	return Check{Current: current, Required: required, Ready: current >= required}
}

func behaviorStats(items []audit.BehaviorLog) BehaviorStats {
	stats := BehaviorStats{
		ByDate:      make(map[string]int),
		ByUser:      make(map[string]int),
		ByEventType: make(map[string]int),
	}
	for _, item := range items {
		occurredAt := item.OccurredAt
		if occurredAt.IsZero() {
			occurredAt = item.CreatedAt
		}
		if !occurredAt.IsZero() {
			stats.ByDate[occurredAt.UTC().Format("2006-01-02")]++
		}
		if item.UserID > 0 {
			stats.ByUser[int64Key(item.UserID)]++
		}
		if item.EventType != "" {
			stats.ByEventType[item.EventType]++
		}
	}
	return stats
}

func favoriteStats(items []games.Favorite, gameItems []games.Game) FavoriteStats {
	gameTypes := make(map[int64]string, len(gameItems))
	for _, game := range gameItems {
		gameTypes[game.ID] = game.GameType
	}
	stats := FavoriteStats{
		ByUser:     make(map[string]int),
		ByGameType: make(map[string]int),
	}
	for _, item := range items {
		if item.UserID > 0 {
			stats.ByUser[int64Key(item.UserID)]++
		}
		gameType := item.Game.GameType
		if gameType == "" {
			gameType = gameTypes[item.GameID]
		}
		if gameType == "" {
			gameType = "unknown"
		}
		stats.ByGameType[gameType]++
	}
	return stats
}

func profileStats(experts []profiles.ExpertSkillProfile, guides []profiles.GuideResourceProfile) ProfileStats {
	stats := ProfileStats{}
	expertTotal := 0
	for _, item := range experts {
		expertTotal += item.Completeness
		if item.Completeness >= 100 {
			stats.ExpertCompleteCount++
		}
	}
	if len(experts) > 0 {
		stats.ExpertCompletenessAverage = float64(expertTotal) / float64(len(experts))
	}
	guideTotal := 0
	for _, item := range guides {
		guideTotal += item.Completeness
		if item.Completeness >= 100 {
			stats.GuideCompleteCount++
		}
	}
	if len(guides) > 0 {
		stats.GuideCompletenessAverage = float64(guideTotal) / float64(len(guides))
	}
	return stats
}

func connectionStats(items []connections.Connection) ConnectionStats {
	stats := ConnectionStats{
		BySourceType:   make(map[string]int),
		ByRelationType: make(map[string]int),
	}
	for _, item := range items {
		sourceType := item.SourceType
		if sourceType == "" {
			sourceType = "unknown"
		}
		relationType := item.RelationType
		if relationType == "" {
			relationType = "unknown"
		}
		stats.BySourceType[sourceType]++
		stats.ByRelationType[relationType]++
	}
	return stats
}

func int64Key(value int64) string {
	return fmt.Sprintf("%d", value)
}

func uniqueUserCount(input SnapshotInput) int {
	users := make(map[int64]bool)
	add := func(userID int64) {
		if userID > 0 {
			users[userID] = true
		}
	}
	for _, item := range input.BehaviorLogs {
		add(item.UserID)
	}
	for _, item := range input.Games {
		add(item.CreatorUserID)
	}
	for _, item := range input.Favorites {
		add(item.UserID)
	}
	for _, item := range input.Reviews {
		add(item.ReviewerUserID)
		add(item.TargetUserID)
	}
	for _, item := range input.Footprints {
		add(item.UserID)
	}
	for _, item := range input.Connections {
		add(item.UserID)
		add(item.ConnectedUserID)
	}
	for _, item := range input.ExpertSkills {
		add(item.UserID)
	}
	for _, item := range input.GuideResources {
		add(item.UserID)
	}
	for _, item := range input.IMMessages {
		add(item.SenderID)
	}
	return len(users)
}
