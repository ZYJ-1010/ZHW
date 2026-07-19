package appapi

import (
	"hash/fnv"

	"zhw-mini/services/go-api/internal/reviews"
)

// awardCompletedGameRewards keeps experience and redeemable points in their
// respective ledgers. The points side is idempotent per user and game.
func (s *Server) awardCompletedGameRewards(gameID int64) {
	s.reviews.AwardCompletedGame(gameID)
	reward := s.reviews.GrowthRules().CompletedGamePoints
	if reward <= 0 {
		return
	}
	for _, userID := range s.games.Members(gameID) {
		_, _, _, _ = s.points.GrantOnce(userID, reward, "completed_game_reward", gameID, "完成组局奖励")
	}
}

func (s *Server) awardNewbieTaskReward(userID int64, rule taskRuleDTO) reviews.GrowthProfile {
	profile := s.reviews.AwardTaskReward(userID, rule.Code, 0, rule.RewardExperience)
	if rule.RewardPoints <= 0 {
		return profile
	}
	account, _, _, err := s.points.GrantOnce(userID, rule.RewardPoints, "newbie_task_reward", taskRewardBizID(rule.Code), "新手任务奖励："+rule.Title)
	if err == nil {
		profile.AvailablePoints = account.AvailablePoints
	}
	return profile
}

func taskRewardBizID(code string) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(code))
	value := int64(hasher.Sum64() & 0x7fffffffffffffff)
	if value == 0 {
		return 1
	}
	return value
}
