package games

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type transitionOrderRepository struct {
	*fakeGameRepository
	updateErr  error
	statusErr  error
	operations []string
	statusLogs []StatusLog
}

type cancelTransitionRepository struct {
	*fakeGameRepository
	cancelErr  error
	cancelLogs []StatusLog
}

func (r *cancelTransitionRepository) CancelGameWithStatusLog(_ context.Context, game Game, _ string, statusLog StatusLog) (Game, StatusLog, error) {
	if r.cancelErr != nil {
		return Game{}, StatusLog{}, r.cancelErr
	}
	persisted, ok := r.games[game.ID]
	if !ok || persisted.Status != statusLog.FromStatus {
		return Game{}, StatusLog{}, errors.New("stale game status")
	}
	delete(r.members, game.ID)
	r.games[game.ID] = game
	statusLog.ID = int64(len(r.cancelLogs) + 1)
	r.cancelLogs = append(r.cancelLogs, statusLog)
	return game, statusLog, nil
}

func (r *transitionOrderRepository) UpdateGame(ctx context.Context, game Game) (Game, error) {
	r.operations = append(r.operations, "update_game")
	if r.updateErr != nil {
		return Game{}, r.updateErr
	}
	return r.fakeGameRepository.UpdateGame(ctx, game)
}

func (r *transitionOrderRepository) SaveStatusLog(_ context.Context, statusLog StatusLog) (StatusLog, error) {
	r.operations = append(r.operations, "save_status_log")
	if r.statusErr != nil {
		return StatusLog{}, r.statusErr
	}
	statusLog.ID = int64(len(r.statusLogs) + 1)
	r.statusLogs = append(r.statusLogs, statusLog)
	return statusLog, nil
}

func TestStatusTransitionDoesNotWriteLogBeforeFailedGameUpdate(t *testing.T) {
	repo := &transitionOrderRepository{fakeGameRepository: newFakeGameRepository()}
	service := NewServiceWithRepositories(fakeIdentity{verified: true}, repo, nil)
	game, err := service.Create(1, CreateRequest{
		Title: "状态日志顺序", GameType: "free", MinPlayers: 5, MaxPlayers: 8,
		StartAt: "2026-08-01 10:00", EndAt: "2026-08-01 12:00",
	})
	if err != nil {
		t.Fatal(err)
	}

	repo.updateErr = errors.New("database update failed")
	if _, err := service.ApproveGame(game.ID); err == nil {
		t.Fatal("expected game update failure")
	}
	if !reflect.DeepEqual(repo.operations, []string{"update_game"}) {
		t.Fatalf("status log must not be attempted before a successful update: %+v", repo.operations)
	}
	if len(repo.statusLogs) != 0 || len(service.StatusLogs(game.ID)) != 0 {
		t.Fatalf("failed update must leave no transition log: repo=%+v memory=%+v", repo.statusLogs, service.StatusLogs(game.ID))
	}
	if current := repo.games[game.ID]; current.Status != StatusPendingAudit {
		t.Fatalf("failed update must keep persisted status unchanged: %+v", current)
	}
}

func TestLegacyStatusTransitionWritesLogAfterSuccessfulGameUpdate(t *testing.T) {
	repo := &transitionOrderRepository{fakeGameRepository: newFakeGameRepository()}
	service := NewServiceWithRepositories(fakeIdentity{verified: true}, repo, nil)
	game, err := service.Create(1, CreateRequest{
		Title: "状态日志成功顺序", GameType: "free", MinPlayers: 5, MaxPlayers: 8,
		StartAt: "2026-08-01 10:00", EndAt: "2026-08-01 12:00",
	})
	if err != nil {
		t.Fatal(err)
	}

	approved, err := service.ApproveGame(game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if approved.Status != StatusRecruiting {
		t.Fatalf("expected recruiting status, got %+v", approved)
	}
	if !reflect.DeepEqual(repo.operations, []string{"update_game", "save_status_log"}) {
		t.Fatalf("expected update before status log: %+v", repo.operations)
	}
	logs := service.StatusLogs(game.ID)
	if len(repo.statusLogs) != 1 || len(logs) != 1 {
		t.Fatalf("expected one committed transition log: repo=%+v memory=%+v", repo.statusLogs, logs)
	}
	if logs[0].FromStatus != StatusPendingAudit || logs[0].ToStatus != StatusRecruiting {
		t.Fatalf("unexpected transition log: %+v", logs[0])
	}
}

func TestCancelTransitionFailureKeepsGameMembersLogsAndCacheUnchanged(t *testing.T) {
	repo := &cancelTransitionRepository{fakeGameRepository: newFakeGameRepository()}
	service := NewServiceWithRepositories(fakeIdentity{verified: true}, repo, nil)
	game, members := mustCreateStartedGame(t, service)
	logsBefore := len(service.StatusLogs(game.ID))
	repo.cancelErr = errors.New("cancel transaction failed")

	if _, err := service.CancelService(game.ID, "运营取消"); err == nil {
		t.Fatal("expected cancellation transaction failure")
	}
	if current := repo.games[game.ID]; current.Status != StatusInProgress || current.CurrentPlayers != len(members) {
		t.Fatalf("failed cancellation must keep persisted game unchanged: %+v", current)
	}
	if len(repo.members[game.ID]) != len(members) {
		t.Fatalf("failed cancellation must keep persisted active members: %+v", repo.members[game.ID])
	}
	if len(repo.cancelLogs) != 0 || len(service.StatusLogs(game.ID)) != logsBefore {
		t.Fatalf("failed cancellation must not append a status log: repo=%+v memory=%+v", repo.cancelLogs, service.StatusLogs(game.ID))
	}
	if current := service.games[game.ID]; current.Status != StatusInProgress || current.CurrentPlayers != len(members) {
		t.Fatalf("failed cancellation must keep cached game unchanged: %+v", current)
	}
	if len(service.members[game.ID]) != len(members) {
		t.Fatalf("failed cancellation must keep cached members: %+v", service.members[game.ID])
	}
}

func TestCancelTransitionCommitsGameMembersAndLogTogether(t *testing.T) {
	repo := &cancelTransitionRepository{fakeGameRepository: newFakeGameRepository()}
	service := NewServiceWithRepositories(fakeIdentity{verified: true}, repo, nil)
	game, _ := mustCreateStartedGame(t, service)
	logsBefore := len(service.StatusLogs(game.ID))
	service.progressFeedbacks[game.ID] = []ProgressFeedback{{ID: 1}}

	canceled, err := service.CancelService(game.ID, "运营取消")
	if err != nil {
		t.Fatal(err)
	}
	if canceled.Status != StatusCancelled || canceled.CurrentPlayers != 0 {
		t.Fatalf("unexpected canceled game: %+v", canceled)
	}
	if current := repo.games[game.ID]; current.Status != StatusCancelled || current.CurrentPlayers != 0 {
		t.Fatalf("persisted game was not canceled: %+v", current)
	}
	if len(repo.members[game.ID]) != 0 || len(service.members[game.ID]) != 0 {
		t.Fatalf("canceled game must have no active members: repo=%+v memory=%+v", repo.members[game.ID], service.members[game.ID])
	}
	if len(repo.cancelLogs) != 1 || len(service.StatusLogs(game.ID)) != logsBefore+1 {
		t.Fatalf("cancellation must append exactly one status log: repo=%+v memory=%+v", repo.cancelLogs, service.StatusLogs(game.ID))
	}
	log := repo.cancelLogs[0]
	if log.FromStatus != StatusInProgress || log.ToStatus != StatusCancelled || log.Reason != "运营取消" {
		t.Fatalf("unexpected cancellation status log: %+v", log)
	}
	if _, ok := service.progressFeedbacks[game.ID]; ok {
		t.Fatal("successful cancellation must clear cached collaboration progress")
	}
}
