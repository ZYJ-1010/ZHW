package connections

import (
	"context"
	"errors"
	"testing"
)

func TestRepositoryBackedConnectionsAndFollowLog(t *testing.T) {
	service := NewServiceWithRepository(newFakeConnectionRepository())
	service.UpsertPair(1, 2, "co_game", "game", 10, 2)
	service.UpsertPair(1, 2, "co_game", "game", 10, 3)

	userConnections := service.My(1)
	if len(userConnections) != 1 || userConnections[0].StrengthScore != 5 {
		t.Fatalf("expected merged repository connection, got %+v", userConnections)
	}
	if len(service.My(2)) != 1 || len(service.All()) != 2 {
		t.Fatalf("expected bidirectional repository connections, all=%+v", service.All())
	}

	if _, err := service.AddFollowLog(3, userConnections[0].ID, FollowRequest{FollowType: "call", Content: "call"}, false); err != ErrConnectionForbidden {
		t.Fatalf("expected ErrConnectionForbidden, got %v", err)
	}
	log, err := service.AddFollowLog(3, userConnections[0].ID, FollowRequest{FollowType: "call", Content: "call"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if log.ID == 0 || service.My(1)[0].StrengthScore != 6 {
		t.Fatalf("expected follow log and strength increment, log=%+v connections=%+v", log, service.My(1))
	}
}

type fakeConnectionRepository struct {
	nextConnectionID int64
	nextFollowID     int64
	items            map[int64]Connection
	follows          map[int64][]FollowLog
	pairErr          error
	followAtomicErr  error
	readErr          error
}

func newFakeConnectionRepository() *fakeConnectionRepository {
	return &fakeConnectionRepository{
		nextConnectionID: 1,
		nextFollowID:     1,
		items:            make(map[int64]Connection),
		follows:          make(map[int64][]FollowLog),
	}
}

func (r *fakeConnectionRepository) UpsertConnection(ctx context.Context, connection Connection, strengthDelta int) (Connection, error) {
	for id, item := range r.items {
		if item.UserID == connection.UserID &&
			item.ConnectedUserID == connection.ConnectedUserID &&
			item.RelationType == connection.RelationType &&
			item.SourceType == connection.SourceType &&
			item.SourceID == connection.SourceID {
			item.StrengthScore += strengthDelta
			item.UpdatedAt = connection.UpdatedAt
			r.items[id] = item
			return item, nil
		}
	}
	connection.ID = r.nextConnectionID
	r.nextConnectionID++
	r.items[connection.ID] = connection
	return connection, nil
}

func (r *fakeConnectionRepository) UpsertConnectionPair(ctx context.Context, first Connection, second Connection, strengthDelta int) (Connection, Connection, error) {
	if r.pairErr != nil {
		return Connection{}, Connection{}, r.pairErr
	}
	firstSaved, err := r.UpsertConnection(ctx, first, strengthDelta)
	if err != nil {
		return Connection{}, Connection{}, err
	}
	secondSaved, err := r.UpsertConnection(ctx, second, strengthDelta)
	return firstSaved, secondSaved, err
}

func (r *fakeConnectionRepository) ListConnections(ctx context.Context, userID int64) ([]Connection, error) {
	if r.readErr != nil {
		return nil, r.readErr
	}
	result := make([]Connection, 0)
	for _, item := range r.items {
		if item.UserID == userID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *fakeConnectionRepository) ListAllConnections(ctx context.Context) ([]Connection, error) {
	if r.readErr != nil {
		return nil, r.readErr
	}
	result := make([]Connection, 0, len(r.items))
	for _, item := range r.items {
		result = append(result, item)
	}
	return result, nil
}

func (r *fakeConnectionRepository) FindConnection(ctx context.Context, connectionID int64) (Connection, bool, error) {
	item, ok := r.items[connectionID]
	return item, ok, nil
}

func (r *fakeConnectionRepository) AddFollowLog(ctx context.Context, log FollowLog) (FollowLog, error) {
	log.ID = r.nextFollowID
	r.nextFollowID++
	r.follows[log.ConnectionID] = append(r.follows[log.ConnectionID], log)
	return log, nil
}

func (r *fakeConnectionRepository) IncrementStrength(ctx context.Context, connectionID int64, delta int) (Connection, error) {
	item := r.items[connectionID]
	item.StrengthScore += delta
	r.items[connectionID] = item
	return item, nil
}

func (r *fakeConnectionRepository) AddFollowLogAndIncrement(ctx context.Context, log FollowLog, strengthDelta int) (FollowLog, Connection, error) {
	if r.followAtomicErr != nil {
		return FollowLog{}, Connection{}, r.followAtomicErr
	}
	saved, err := r.AddFollowLog(ctx, log)
	if err != nil {
		return FollowLog{}, Connection{}, err
	}
	connection, err := r.IncrementStrength(ctx, log.ConnectionID, strengthDelta)
	return saved, connection, err
}

func TestStrictPairAndFollowDoNotReportSuccessOnRepositoryFailure(t *testing.T) {
	repo := newFakeConnectionRepository()
	service := NewServiceWithRepository(repo)
	repo.pairErr = errors.New("pair transaction failed")
	if err := service.UpsertPairStrict(1, 2, "co_game", "game", 10, 2); err == nil {
		t.Fatal("expected pair transaction failure")
	}
	if len(repo.items) != 0 {
		t.Fatalf("failed pair must not leave one-sided relation: %+v", repo.items)
	}

	repo.pairErr = nil
	if err := service.UpsertPairStrict(1, 2, "co_game", "game", 10, 2); err != nil {
		t.Fatalf("expected pair transaction success: %v", err)
	}
	connection := service.My(1)[0]
	repo.followAtomicErr = errors.New("follow transaction failed")
	if _, err := service.AddFollowLog(1, connection.ID, FollowRequest{FollowType: "call", Content: "测试跟进"}, false); err == nil {
		t.Fatal("expected follow transaction failure")
	}
	if len(repo.follows[connection.ID]) != 0 || service.My(1)[0].StrengthScore != 2 {
		t.Fatalf("failed follow must not change log or strength, follows=%+v connection=%+v", repo.follows, service.My(1)[0])
	}
}

func TestStrictListsDoNotUseStaleConnectionCache(t *testing.T) {
	repo := newFakeConnectionRepository()
	service := NewServiceWithRepository(repo)
	service.connections[99] = Connection{ID: 99, UserID: 1, ConnectedUserID: 2}
	repo.readErr = errors.New("database unavailable")

	if _, err := service.MyStrict(1); !errors.Is(err, repo.readErr) {
		t.Fatalf("expected user connection read error, got %v", err)
	}
	if _, err := service.AllStrict(); !errors.Is(err, repo.readErr) {
		t.Fatalf("expected all connection read error, got %v", err)
	}
}
