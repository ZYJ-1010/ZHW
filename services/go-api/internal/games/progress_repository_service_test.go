package games

import (
	"context"
	"testing"
	"time"
)

func TestProgressRepositoryPersistsMilestonesCheckinsRetrospectivesAndContinueDrafts(t *testing.T) {
	service := newVerifiedGameService()
	progressRepo := newFakeProgressRepository()
	service.UseProgressRepository(progressRepo)
	game, members := mustCreateStartedGame(t, service)

	milestone, err := service.CreateMilestone(1, game.ID, MilestoneRequest{Title: "phase one"})
	if err != nil {
		t.Fatal(err)
	}
	if !progressRepo.createdMilestone || milestone.ID == 0 {
		t.Fatalf("expected repository milestone create, milestone=%+v repo=%+v", milestone, progressRepo)
	}
	if _, err := service.UpdateMilestone(1, game.ID, milestone.ID, MilestoneRequest{Status: "completed"}); err != nil {
		t.Fatal(err)
	}
	items, err := service.Milestones(2, game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !progressRepo.listedMilestones || len(items) != 1 || items[0].Status != "completed" {
		t.Fatalf("expected repository milestones, items=%+v repo=%+v", items, progressRepo)
	}

	checkin, err := service.CreateCheckin(2, game.ID, CheckinRequest{MilestoneID: milestone.ID, CheckinType: "proof", Content: "done", FileIDs: []int64{9}})
	if err != nil {
		t.Fatal(err)
	}
	if !progressRepo.createdCheckin || len(checkin.FileIDs) != 1 {
		t.Fatalf("expected repository checkin create, checkin=%+v repo=%+v", checkin, progressRepo)
	}
	if _, err := service.MarkCheckinInvalid(checkin.ID); err != nil {
		t.Fatal(err)
	}
	checkins, err := service.Checkins(2, game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(checkins) != 0 {
		t.Fatalf("expected invalid checkin hidden, got %+v", checkins)
	}
	adminCheckins, err := service.AdminCheckins(game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !progressRepo.listedCheckins || len(adminCheckins) != 1 || adminCheckins[0].Status != "invalid" {
		t.Fatalf("expected admin repository checkins, got %+v repo=%+v", adminCheckins, progressRepo)
	}

	mustMoveGameToPendingReview(t, service, game.ID, members)
	retro, err := service.CreateRetrospective(2, game.ID, RetrospectiveRequest{Content: "great", AgainIntent: "yes"})
	if err != nil {
		t.Fatal(err)
	}
	if !progressRepo.createdRetrospective || retro.ID == 0 {
		t.Fatalf("expected repository retrospective create, retro=%+v repo=%+v", retro, progressRepo)
	}
	if _, err := service.CreateRetrospective(2, game.ID, RetrospectiveRequest{Content: "duplicate"}); err != ErrDuplicateRetrospective {
		t.Fatalf("expected duplicate retrospective from repository, got %v", err)
	}

	continued, err := service.ContinueDraft(1, game.ID, ContinueDraftRequest{Title: "next draft"})
	if err != nil {
		t.Fatal(err)
	}
	records, err := service.AdminContinueDrafts(game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !progressRepo.createdContinueDraft || len(records) != 1 || records[0].DraftGameID != continued.Draft.ID {
		t.Fatalf("expected repository continue draft, continued=%+v records=%+v repo=%+v", continued, records, progressRepo)
	}
}

type fakeProgressRepository struct {
	nextMilestoneID      int64
	nextCheckinID        int64
	nextRetrospectiveID  int64
	milestones           map[int64]Milestone
	checkins             map[int64]Checkin
	retrospectives       map[int64]Retrospective
	continueDrafts       []ContinueDraftRecord
	createdMilestone     bool
	listedMilestones     bool
	createdCheckin       bool
	listedCheckins       bool
	createdRetrospective bool
	createdContinueDraft bool
}

func newFakeProgressRepository() *fakeProgressRepository {
	return &fakeProgressRepository{
		nextMilestoneID:     1,
		nextCheckinID:       1,
		nextRetrospectiveID: 1,
		milestones:          make(map[int64]Milestone),
		checkins:            make(map[int64]Checkin),
		retrospectives:      make(map[int64]Retrospective),
		continueDrafts:      make([]ContinueDraftRecord, 0),
	}
}

func (r *fakeProgressRepository) CreateMilestone(ctx context.Context, milestone Milestone) (Milestone, error) {
	r.createdMilestone = true
	milestone.ID = r.nextMilestoneID
	r.nextMilestoneID++
	r.milestones[milestone.ID] = milestone
	return milestone, nil
}

func (r *fakeProgressRepository) UpdateMilestone(ctx context.Context, milestone Milestone) (Milestone, error) {
	r.milestones[milestone.ID] = milestone
	return milestone, nil
}

func (r *fakeProgressRepository) ListMilestones(ctx context.Context, gameID int64) ([]Milestone, error) {
	r.listedMilestones = true
	result := make([]Milestone, 0)
	for _, item := range r.milestones {
		if item.GameID == gameID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *fakeProgressRepository) CreateCheckin(ctx context.Context, checkin Checkin) (Checkin, error) {
	r.createdCheckin = true
	checkin.ID = r.nextCheckinID
	r.nextCheckinID++
	r.checkins[checkin.ID] = checkin
	return checkin, nil
}

func (r *fakeProgressRepository) UpdateCheckinStatus(ctx context.Context, checkinID int64, status string) (Checkin, error) {
	item, ok := r.checkins[checkinID]
	if !ok {
		return Checkin{}, ErrCheckinNotFound
	}
	item.Status = status
	r.checkins[checkinID] = item
	return item, nil
}

func (r *fakeProgressRepository) ListCheckins(ctx context.Context, gameID int64, includeInvalid bool) ([]Checkin, error) {
	r.listedCheckins = true
	result := make([]Checkin, 0)
	for _, item := range r.checkins {
		if item.GameID != gameID {
			continue
		}
		if !includeInvalid && item.Status == "invalid" {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

func (r *fakeProgressRepository) CreateRetrospective(ctx context.Context, retrospective Retrospective) (Retrospective, error) {
	r.createdRetrospective = true
	retrospective.ID = r.nextRetrospectiveID
	r.nextRetrospectiveID++
	r.retrospectives[retrospective.ID] = retrospective
	return retrospective, nil
}

func (r *fakeProgressRepository) RetrospectiveExists(ctx context.Context, gameID int64, userID int64) (bool, error) {
	for _, item := range r.retrospectives {
		if item.GameID == gameID && item.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeProgressRepository) ListRetrospectives(ctx context.Context, gameID int64) ([]Retrospective, error) {
	result := make([]Retrospective, 0)
	for _, item := range r.retrospectives {
		if item.GameID == gameID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *fakeProgressRepository) CreateContinueDraft(ctx context.Context, record ContinueDraftRecord) (ContinueDraftRecord, error) {
	r.createdContinueDraft = true
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now()
	}
	r.continueDrafts = append(r.continueDrafts, record)
	return record, nil
}

func (r *fakeProgressRepository) ListContinueDrafts(ctx context.Context, gameID int64) ([]ContinueDraftRecord, error) {
	result := make([]ContinueDraftRecord, 0)
	for _, item := range r.continueDrafts {
		if item.OriginalGameID == gameID {
			result = append(result, item)
		}
	}
	return result, nil
}
