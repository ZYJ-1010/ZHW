package appapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"zhw-mini/services/go-api/internal/aidata"
	"zhw-mini/services/go-api/internal/audit"
	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/im"
	"zhw-mini/services/go-api/internal/profiles"
	"zhw-mini/services/go-api/internal/reviews"
)

const aiAcceptanceSeedUserBase int64 = 910000

type aiDataSeedResult struct {
	UsersVerified      int             `json:"usersVerified"`
	GamesCreated       int             `json:"gamesCreated"`
	BehaviorLogsAdded  int             `json:"behaviorLogsAdded"`
	FavoritesAdded     int             `json:"favoritesAdded"`
	ReviewsAdded       int             `json:"reviewsAdded"`
	IMMessagesAdded    int             `json:"imMessagesAdded"`
	AcceptanceSnapshot aidata.Snapshot `json:"acceptanceSnapshot"`
}

func (s *Server) adminAIDataAcceptanceFixture(w http.ResponseWriter, r *http.Request) {
	result, err := s.seedAIDataAcceptanceFixture()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "生成 AI 数据验收样例失败")
		return
	}
	s.recordOperation(r, "ai_data:seed_acceptance_fixture", "ai_data", "acceptance", map[string]interface{}{
		"acceptanceReady": result.AcceptanceSnapshot.AcceptanceReady,
		"gamesCreated":    result.GamesCreated,
		"behaviorAdded":   result.BehaviorLogsAdded,
		"favoritesAdded":  result.FavoritesAdded,
		"reviewsAdded":    result.ReviewsAdded,
	})
	httpx.OK(w, result)
}

func (s *Server) seedAIDataAcceptanceFixture() (aiDataSeedResult, error) {
	result := aiDataSeedResult{}
	snapshot, err := s.aiDataSnapshot()
	if err != nil {
		return result, err
	}
	if snapshot.AcceptanceReady {
		result.AcceptanceSnapshot = snapshot
		return result, nil
	}

	seedUsers := []int64{
		aiAcceptanceSeedUserBase + 1,
		aiAcceptanceSeedUserBase + 2,
		aiAcceptanceSeedUserBase + 3,
		aiAcceptanceSeedUserBase + 4,
		aiAcceptanceSeedUserBase + 5,
	}
	for _, userID := range seedUsers {
		verified, err := s.ensureSeedIdentity(userID)
		if err != nil {
			return result, err
		}
		if verified {
			result.UsersVerified++
		}
	}

	seedGames, created, err := s.ensureAcceptanceGames(seedUsers[0])
	if err != nil {
		return result, err
	}
	result.GamesCreated = created
	for _, game := range seedGames {
		if err := s.ensureAcceptanceGameReady(game.ID, seedUsers); err != nil {
			return result, err
		}
	}

	result.BehaviorLogsAdded, err = s.ensureAcceptanceBehaviorLogs(seedUsers, seedGames)
	if err != nil {
		return result, err
	}
	favoritesAdded, err := s.ensureAcceptanceFavorites(seedUsers, seedGames)
	if err != nil {
		return result, err
	}
	result.FavoritesAdded = favoritesAdded
	reviewsAdded, err := s.ensureAcceptanceReviews(seedUsers, seedGames)
	if err != nil {
		return result, err
	}
	result.ReviewsAdded = reviewsAdded
	result.IMMessagesAdded, err = s.ensureAcceptanceIMMessages(seedUsers, seedGames[0].ID)
	if err != nil {
		return result, err
	}
	if err := s.ensureAcceptanceProfiles(seedUsers); err != nil {
		return result, err
	}

	result.AcceptanceSnapshot, err = s.aiDataSnapshot()
	if err != nil {
		return result, err
	}
	return result, nil
}

func (s *Server) aiDataSnapshot() (aidata.Snapshot, error) {
	behaviorLogs, err := s.audit.BehaviorLogsStrict()
	if err != nil {
		return aidata.Snapshot{}, fmt.Errorf("读取行为日志失败: %w", err)
	}
	gameItems, err := s.games.ListStrict()
	if err != nil {
		return aidata.Snapshot{}, fmt.Errorf("读取组局数据失败: %w", err)
	}
	favorites, err := s.games.AllFavoritesStrict()
	if err != nil {
		return aidata.Snapshot{}, fmt.Errorf("读取收藏数据失败: %w", err)
	}
	reviewItems, err := s.reviews.AllReviewsStrict()
	if err != nil {
		return aidata.Snapshot{}, fmt.Errorf("读取评价数据失败: %w", err)
	}
	footprints, err := s.reviews.AllFootprintsStrict()
	if err != nil {
		return aidata.Snapshot{}, fmt.Errorf("读取成长足迹失败: %w", err)
	}
	connectionItems, err := s.connections.AllStrict()
	if err != nil {
		return aidata.Snapshot{}, fmt.Errorf("读取关系数据失败: %w", err)
	}
	expertSkills, err := s.profiles.AllExpertSkillsStrict()
	if err != nil {
		return aidata.Snapshot{}, fmt.Errorf("读取行家技能失败: %w", err)
	}
	guideResources, err := s.profiles.AllGuideResourcesStrict()
	if err != nil {
		return aidata.Snapshot{}, fmt.Errorf("读取领路人资源失败: %w", err)
	}
	imMessages, err := s.im.AllMessagesStrict()
	if err != nil {
		return aidata.Snapshot{}, fmt.Errorf("读取聊天数据失败: %w", err)
	}
	return s.aidata.Snapshot(aidata.SnapshotInput{
		BehaviorLogs: behaviorLogs, Games: gameItems, Favorites: favorites, Reviews: reviewItems,
		Footprints: footprints, Connections: connectionItems, ExpertSkills: expertSkills,
		GuideResources: guideResources, IMMessages: imMessages,
	}), nil
}

func (s *Server) ensureSeedIdentity(userID int64) (bool, error) {
	if s.identity.IsRealnameVerified(userID) {
		return false, nil
	}
	_, err := s.identity.BindPhone(userID, fmt.Sprintf("139%08d", userID%100000000))
	if err != nil {
		return false, err
	}
	smsResult, err := s.identity.SendSMSCode(userID)
	if err != nil {
		return false, err
	}
	if smsResult.MockCode == "" {
		return false, errors.New("ai data fixture requires local sms sender")
	}
	if _, err = s.identity.VerifySMSCode(userID, smsResult.MockCode); err != nil {
		return false, err
	}
	if _, err = s.identity.VerifyPhone(userID, "E5 Fixture User", "110101199001011234"); err != nil {
		return false, err
	}
	token, err := s.identity.StartFaceID(userID)
	if err != nil {
		return false, err
	}
	if _, err = s.identity.CompleteFaceID(userID, token); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Server) ensureAcceptanceGames(creatorUserID int64) ([]games.Game, int, error) {
	seedGames, err := s.acceptanceGames()
	if err != nil {
		return nil, 0, err
	}
	created := 0
	for len(seedGames) < 3 {
		game, err := s.games.Create(creatorUserID, games.CreateRequest{
			Title:         fmt.Sprintf("E5 AI acceptance game %d", len(seedGames)+1),
			GameType:      "free",
			MinPlayers:    5,
			MaxPlayers:    8,
			SignupStartAt: "2026-01-01 00:00",
			SignupEndAt:   "2026-12-31 23:59",
			StartAt:       "2027-01-01 10:00",
			EndAt:         "2027-01-01 12:00",
			CityCode:      "110100",
			CityName:      "Beijing",
			Longitude:     116.397,
			Latitude:      39.908,
		})
		if err != nil {
			return seedGames, created, err
		}
		seedGames = append(seedGames, game)
		created++
	}
	return seedGames[:3], created, nil
}

func (s *Server) acceptanceGames() ([]games.Game, error) {
	gamesList, err := s.games.ListStrict()
	if err != nil {
		return nil, err
	}
	items := make([]games.Game, 0, 3)
	for _, item := range gamesList {
		if strings.HasPrefix(item.Title, "E5 AI acceptance game ") {
			items = append(items, item)
		}
	}
	return items, nil
}

func (s *Server) ensureAcceptanceGameReady(gameID int64, seedUsers []int64) error {
	game, err := s.games.Get(gameID)
	if err != nil {
		return err
	}
	if game.Status == "pending_audit" {
		game, err = s.games.ApproveGame(gameID)
		if err != nil {
			return err
		}
	}
	for _, userID := range seedUsers[1:] {
		if s.games.IsMember(gameID, userID) {
			continue
		}
		app, err := s.games.Apply(userID, gameID, games.ApplyRequest{Reason: "E5 acceptance fixture"})
		if err != nil {
			if errors.Is(err, games.ErrAlreadyMember) {
				continue
			}
			return err
		}
		if _, err = s.games.ReviewApplication(game.CreatorUserID, app.ID, true); err != nil {
			return err
		}
	}
	game, err = s.games.Get(gameID)
	if err != nil {
		return err
	}
	if game.Status == "recruiting" || game.Status == "full" {
		if _, err = s.games.ManualStart(game.CreatorUserID, gameID); err != nil {
			return err
		}
	}
	for _, userID := range seedUsers {
		_, _, game, err = s.games.ConfirmService(userID, gameID, "E5 acceptance fixture")
		if err != nil {
			return err
		}
	}
	if game.Status == "pending_review" {
		if err = s.reviews.MarkGameReviewableStrict(gameID); err != nil {
			return err
		}
		if err = s.createCoGameConnections(gameID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) ensureAcceptanceBehaviorLogs(seedUsers []int64, seedGames []games.Game) (int, error) {
	added := 0
	logs, err := s.audit.BehaviorLogsStrict()
	if err != nil {
		return 0, err
	}
	for len(logs)+added < 20 {
		index := len(logs) + added
		userID := seedUsers[index%len(seedUsers)]
		game := seedGames[index%len(seedGames)]
		s.audit.RecordBehavior(audit.BehaviorRequest{
			UserID:     userID,
			EventType:  "e5_acceptance_view",
			TargetType: "game",
			TargetID:   game.ID,
			PagePath:   "/pages/game/detail/index",
			Keyword:    "ai",
			Extra:      map[string]interface{}{"fixture": "e5_ai_acceptance"},
		})
		added++
	}
	return added, nil
}

func (s *Server) ensureAcceptanceFavorites(seedUsers []int64, seedGames []games.Game) (int, error) {
	added := 0
	for _, userID := range seedUsers {
		for _, game := range seedGames {
			favorites, err := s.games.AllFavoritesStrict()
			if err != nil {
				return added, err
			}
			if len(favorites) >= 5 {
				return added, nil
			}
			before := len(favorites)
			if _, err := s.games.FavoriteGame(userID, game.ID); err != nil {
				return added, err
			}
			favorites, err = s.games.AllFavoritesStrict()
			if err != nil {
				return added, err
			}
			if len(favorites) > before {
				added++
			}
		}
	}
	return added, nil
}

func (s *Server) ensureAcceptanceReviews(seedUsers []int64, seedGames []games.Game) (int, error) {
	added := 0
	pairs := [][2]int64{
		{seedUsers[0], seedUsers[1]},
		{seedUsers[1], seedUsers[0]},
		{seedUsers[2], seedUsers[0]},
		{seedUsers[3], seedUsers[0]},
		{seedUsers[4], seedUsers[0]},
	}
	for _, game := range seedGames {
		for _, pair := range pairs {
			reviewItems, err := s.reviews.AllReviewsStrict()
			if err != nil {
				return added, err
			}
			if len(reviewItems) >= 5 {
				return added, nil
			}
			_, _, err = s.reviews.Submit(pair[0], reviews.SubmitRequest{
				GameID:       game.ID,
				TargetUserID: pair[1],
				TargetRole:   "member",
				Score:        5,
				Content:      "E5 acceptance fixture",
				AgainIntent:  "yes",
			})
			if err == nil {
				added++
				continue
			}
			if errors.Is(err, reviews.ErrDuplicateReview) {
				continue
			}
			return added, err
		}
	}
	return added, nil
}

func (s *Server) ensureAcceptanceIMMessages(seedUsers []int64, gameID int64) (int, error) {
	messages, err := s.im.AllMessagesStrict()
	if err != nil {
		return 0, err
	}
	if len(messages) > 0 {
		return 0, nil
	}
	added := 0
	for _, userID := range seedUsers {
		if _, err := s.im.Send(userID, gameID, im.SendRequest{MessageType: "text", Content: "E5 acceptance fixture"}); err == nil {
			added++
		}
	}
	return added, nil
}

func (s *Server) ensureAcceptanceProfiles(seedUsers []int64) error {
	if err := s.profiles.GrantRoleStrict(seedUsers[0], "expert"); err != nil {
		return err
	}
	if _, err := s.profiles.UpdateExpertSkill(seedUsers[0], profiles.ExpertSkillRequest{
		SkillTree:   []string{"boardgame"},
		ServiceTags: []string{"host"},
	}); err != nil {
		return err
	}
	if err := s.profiles.GrantRoleStrict(seedUsers[1], "guide"); err != nil {
		return err
	}
	if _, err := s.profiles.UpdateGuideResource(seedUsers[1], profiles.GuideResourceRequest{
		ResourceTags:    []string{"venue"},
		IndustryTags:    []string{"entertainment"},
		CityCodes:       []string{"110100"},
		ConnectionScale: "100-500",
	}); err != nil {
		return err
	}
	return nil
}
