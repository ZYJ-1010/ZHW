package appapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/reviews"
	"zhw-mini/services/go-api/internal/users"
)

func TestGameDetailHTTPReturnsBackendDrivenPrimaryAction(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	creatorToken := loginForTestWithCode(t, mux, "detail-action-creator")

	createdBody := postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"详情动作测试","gameType":"free","minPlayers":5,"maxPlayers":8}`, http.StatusOK)
	var created struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createdBody, &created); err != nil {
		t.Fatal(err)
	}
	gamePath := "/api/app/games/" + strconv.FormatInt(created.Data.ID, 10)

	pendingBody := getJSON(t, mux, gamePath, creatorToken, http.StatusOK)
	var pending struct {
		Data struct {
			DetailDisplay GameDetailDisplayDTO `json:"detailDisplay"`
		} `json:"data"`
	}
	if err := json.Unmarshal(pendingBody, &pending); err != nil {
		t.Fatal(err)
	}
	if pending.Data.DetailDisplay.StatusText != "待后台审核" || pending.Data.DetailDisplay.PrimaryAction.Text != "后台审核中" || !pending.Data.DetailDisplay.PrimaryAction.Disabled {
		t.Fatalf("pending detail must be backend-review-only: %s", pendingBody)
	}
	if pending.Data.DetailDisplay.Organizer.Rating != "" || pending.Data.DetailDisplay.Organizer.RatingCount != 0 {
		t.Fatalf("new organizer must not expose a fake rating: %s", pendingBody)
	}

	homeBody := getJSON(t, mux, "/api/app/home", creatorToken, http.StatusOK)
	var home struct {
		Data struct {
			RecommendedGames []struct {
				ID         int64  `json:"id"`
				Scope      string `json:"scope"`
				StatusText string `json:"statusText"`
				Route      string `json:"route"`
			} `json:"recommendedGames"`
		} `json:"data"`
	}
	if err := json.Unmarshal(homeBody, &home); err != nil {
		t.Fatal(err)
	}
	if len(home.Data.RecommendedGames) == 0 || home.Data.RecommendedGames[0].ID != created.Data.ID || home.Data.RecommendedGames[0].Scope != "mine" || home.Data.RecommendedGames[0].StatusText != "待后台审核" || !strings.Contains(home.Data.RecommendedGames[0].Route, "id=") {
		t.Fatalf("creator home must include its pending game without publishing it to others: %s", homeBody)
	}

	postJSON(t, mux, gamePath+"/approve-local", creatorToken, `{}`, http.StatusOK)
	playerToken := loginForTestWithCode(t, mux, "detail-action-player")
	completeIdentityForTest(t, mux, playerToken)

	guestBody := getJSON(t, mux, gamePath, playerToken, http.StatusOK)
	var guest struct {
		Data struct {
			DetailDisplay GameDetailDisplayDTO `json:"detailDisplay"`
		} `json:"data"`
	}
	if err := json.Unmarshal(guestBody, &guest); err != nil {
		t.Fatal(err)
	}
	if guest.Data.DetailDisplay.PrimaryAction.Text != "立即报名" || guest.Data.DetailDisplay.PrimaryAction.Action != "apply" || guest.Data.DetailDisplay.PrimaryAction.Disabled {
		t.Fatalf("recruiting guest must receive apply action: %s", guestBody)
	}

	postJSON(t, mux, gamePath+"/applications", playerToken, `{"reason":"参加"}`, http.StatusOK)
	creatorBody := getJSON(t, mux, gamePath, creatorToken, http.StatusOK)
	var creator struct {
		Data struct {
			DetailDisplay GameDetailDisplayDTO `json:"detailDisplay"`
		} `json:"data"`
	}
	if err := json.Unmarshal(creatorBody, &creator); err != nil {
		t.Fatal(err)
	}
	if creator.Data.DetailDisplay.PendingApplicationCount != 1 || creator.Data.DetailDisplay.PrimaryAction.Action != "audit" || creator.Data.DetailDisplay.PrimaryAction.Text != "审核报名（1）" {
		t.Fatalf("creator must receive scoped audit action: %s", creatorBody)
	}
}

func TestGameDetailPrimaryActionByStatusAndRelation(t *testing.T) {
	tests := []struct {
		name         string
		status       string
		relation     GameMyRelationDTO
		pendingCount int
		reviewed     bool
		text         string
		action       string
		disabled     bool
		routePart    string
	}{
		{name: "pending audit", status: "pending_audit", text: "后台审核中", action: "none", disabled: true},
		{name: "creator can start", status: "recruiting", relation: GameMyRelationDTO{IsCreator: true, IsMember: true, CanStart: true}, text: "开始组局", action: "start"},
		{name: "creator audits applications", status: "recruiting", relation: GameMyRelationDTO{IsCreator: true, IsMember: true, CanAudit: true}, pendingCount: 2, text: "审核报名（2）", action: "audit", routePart: "gameId=42"},
		{name: "guest applies", status: "recruiting", relation: GameMyRelationDTO{Role: "guest", CanApply: true}, text: "立即报名", action: "apply", routePart: "gameId=42"},
		{name: "guest blocked by signup window", status: "recruiting", relation: GameMyRelationDTO{Role: "guest", ApplyDisabledReason: "signup closed"}, text: "signup closed", action: "none", disabled: true},
		{name: "application pending", status: "recruiting", relation: GameMyRelationDTO{Role: "guest", ApplicationStatus: "pending"}, text: "报名审核中", action: "none", disabled: true},
		{name: "member waits", status: "recruiting", relation: GameMyRelationDTO{IsMember: true}, text: "等待开局", action: "none", disabled: true},
		{name: "full creator starts", status: "full", relation: GameMyRelationDTO{IsCreator: true, IsMember: true, CanStart: true}, text: "开始组局", action: "start"},
		{name: "full member waits", status: "full", relation: GameMyRelationDTO{IsMember: true}, text: "等待开局", action: "none", disabled: true},
		{name: "full guest", status: "full", text: "已满员", action: "none", disabled: true},
		{name: "member enters collaboration", status: "in_progress", relation: GameMyRelationDTO{IsMember: true}, text: "进入组局", action: "collaboration", routePart: "collaboration"},
		{name: "guest sees in progress", status: "in_progress", text: "进行中", action: "none", disabled: true},
		{name: "ended while confirmation pending", status: "pending_confirm", relation: GameMyRelationDTO{IsMember: true, CanConfirm: true}, text: "已结束", action: "none", disabled: true},
		{name: "member reviews", status: "pending_review", relation: GameMyRelationDTO{IsMember: true, CanReview: true}, text: "去评价", action: "review", routePart: "review"},
		{name: "member reviewed", status: "pending_review", relation: GameMyRelationDTO{IsMember: true}, reviewed: true, text: "已评价", action: "none", disabled: true},
		{name: "completed", status: "completed", text: "已完成", action: "none", disabled: true},
		{name: "draft", status: "draft", text: "草稿", action: "none", disabled: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := gameDetailPrimaryAction(games.Game{ID: 42, Status: test.status, GameType: "free"}, test.relation, test.pendingCount, test.reviewed)
			if result.Text != test.text || result.Action != test.action || result.Disabled != test.disabled {
				t.Fatalf("unexpected action: %+v", result)
			}
			if test.routePart != "" && !strings.Contains(result.Route, test.routePart) {
				t.Fatalf("route %q must contain %q", result.Route, test.routePart)
			}
		})
	}
}

func TestOrganizerRatingUsesOnlyRealReceivedReviews(t *testing.T) {
	if rating, count := organizerRating(nil, 9); rating != "" || count != 0 {
		t.Fatalf("empty reviews must stay hidden, rating=%q count=%d", rating, count)
	}

	rating, count := organizerRating([]reviews.Review{
		{TargetUserID: 9, Score: 5},
		{TargetUserID: 9, Score: 4},
		{TargetUserID: 7, Score: 1},
	}, 9)
	if rating != "4.5" || count != 2 {
		t.Fatalf("expected 4.5 from two organizer reviews, rating=%q count=%d", rating, count)
	}

	payload, err := json.Marshal(GameDetailOrganizerDTO{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "rating\"") {
		t.Fatalf("empty organizer rating must be omitted: %s", payload)
	}
}

func TestFilterReceivedApplicationsByGameAndStatus(t *testing.T) {
	items := []games.Application{
		{ID: 1, GameID: 10, Status: "pending"},
		{ID: 2, GameID: 10, Status: "approved"},
		{ID: 3, GameID: 11, Status: "pending"},
	}
	filtered := filterReceivedApplications(items, "pending", 10)
	if len(filtered) != 1 || filtered[0].ID != 1 {
		t.Fatalf("expected only pending application for game 10: %+v", filtered)
	}
}
