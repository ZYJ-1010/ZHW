package appapi

import (
	"hash/fnv"

	"zhw-mini/services/go-api/internal/reviews"
)

// awardCompletedGameRewards grants only growth experience. Redeemable points
// belong to the separately settled paid-game revenue flow.
func (s *Server) awardCompletedGameRewards(gameID int64) error {
	_, err := s.reviews.AwardCompletedGameStrict(gameID)
	return err
}

func (s *Server) awardNewbieTaskReward(userID int64, rule taskRuleDTO) (reviews.GrowthProfile, error) {
	return s.awardTaskReward(userID, rule, rule.Code)
}

func (s *Server) awardTaskReward(userID int64, rule taskRuleDTO, rewardKey string) (reviews.GrowthProfile, error) {
	profile, err := s.reviews.AwardTaskRewardStrict(userID, rewardKey, 0, rule.RewardExperience)
	if err != nil {
		return reviews.GrowthProfile{}, err
	}
	if rule.RewardPoints <= 0 {
		return profile, nil
	}
	account, _, _, err := s.points.GrantOnce(userID, rule.RewardPoints, "task_reward", taskRewardBizID(rewardKey), "任务奖励："+rule.Title)
	if err != nil {
		return reviews.GrowthProfile{}, err
	}
	profile.AvailablePoints = account.AvailablePoints
	return profile, nil
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
