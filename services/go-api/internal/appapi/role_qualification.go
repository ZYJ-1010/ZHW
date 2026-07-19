package appapi

import (
	"fmt"
	"strings"

	"zhw-mini/services/go-api/internal/invites"
)

type roleApplyRequirement struct {
	Key      string `json:"key"`
	Title    string `json:"title"`
	Text     string `json:"text"`
	Met      bool   `json:"met"`
	Checked  bool   `json:"checked"`
	Required int    `json:"required,omitempty"`
	Current  int    `json:"current,omitempty"`
}

type roleApplyEligibility struct {
	RoleCode     string                 `json:"roleCode"`
	Eligible     bool                   `json:"eligible"`
	BaseEligible bool                   `json:"baseEligible"`
	Requirements []roleApplyRequirement `json:"requirements"`
}

type roleApplicationRequirementsError struct {
	RoleCode string
	Missing  []string
}

func roleApplicationRequirementsErrorIs(err error) bool {
	_, ok := err.(roleApplicationRequirementsError)
	return ok
}

func (e roleApplicationRequirementsError) Error() string {
	if len(e.Missing) == 0 {
		return "未满足角色申请条件"
	}
	return fmt.Sprintf("未满足申请条件：%s", strings.Join(e.Missing, "、"))
}

func (s *Server) roleApplyEligibility(userID int64, roleCode string) (roleApplyEligibility, error) {
	roleCode = strings.TrimSpace(roleCode)
	if userID <= 0 || (roleCode != "expert" && roleCode != "guide") {
		return roleApplyEligibility{}, fmt.Errorf("invalid role code")
	}

	growth := s.reviews.Profile(userID)
	createdGames, participatedGames, completedUsers := s.roleApplyGameStats()
	enterpriseMet := s.enterpriseCertificationMet(userID)
	rules := s.currentOperationRules().Roles
	requirements := make([]roleApplyRequirement, 0, 7)
	if roleCode == "expert" {
		requirements = append(requirements,
			roleApplyBoolRequirement("realname", "完成实名认证", s.identity.IsVerified(userID), "已完成", "未完成"),
			roleApplyBoolRequirement("enterprise", "完成企业认证", enterpriseMet, "已认证", "未认证"),
			roleApplyCountRequirement("created_games", fmt.Sprintf("发起过 %d 次以上组局", rules.ExpertCreatedGames), createdGames[userID], rules.ExpertCreatedGames),
			roleApplyCountRequirement("credit_score", fmt.Sprintf("信用分 ≥ %d 分", rules.ExpertCreditScore), growth.CreditScore, rules.ExpertCreditScore),
			roleApplyRequirement{Key: "plan", Title: "提交行家计划书", Text: "提交申请时填写计划书", Met: false, Checked: false},
		)
	} else {
		invitedCompleted := s.roleApplyInvitedCompletedCount(userID, completedUsers)
		requirements = append(requirements,
			roleApplyBoolRequirement("realname", "完成实名认证", s.identity.IsVerified(userID), "已完成", "未完成"),
			roleApplyBoolRequirement("enterprise", "完成企业认证", enterpriseMet, "已认证", "未认证"),
			roleApplyCountRequirement("participated_games", fmt.Sprintf("参与过 %d 次以上组局", rules.GuideParticipatedGames), participatedGames[userID], rules.GuideParticipatedGames),
			roleApplyCountRequirement("invited_completed_game", fmt.Sprintf("已成功邀请 ≥ %d 人完成组局", rules.GuideInvitedCompleted), invitedCompleted, rules.GuideInvitedCompleted),
			roleApplyCountRequirement("credit_score", fmt.Sprintf("信用分 ≥ %d 分", rules.GuideCreditScore), growth.CreditScore, rules.GuideCreditScore),
			roleApplyRequirement{Key: "plan", Title: "提交领路计划书", Text: "提交申请时填写计划书", Met: false, Checked: false},
		)
		// 后台明确标记的领路人资格是运营白名单，不参与普通条件计算。
		// 这是单独的人工开通路径，不再作为条件配置来源。
		if qualification, err := s.profiles.GuideQualification(userID); err == nil && qualification.ConditionMet {
			for index, item := range requirements {
				if item.Key == "plan" {
					continue
				}
				item.Met = true
				item.Checked = true
				item.Text = "后台已确认"
				requirements[index] = item
			}
		}
	}

	baseEligible := true
	eligible := true
	for _, item := range requirements {
		if item.Key != "plan" && !item.Met {
			baseEligible = false
		}
		if !item.Met {
			eligible = false
		}
	}
	return roleApplyEligibility{RoleCode: roleCode, Eligible: eligible, BaseEligible: baseEligible, Requirements: requirements}, nil
}

func roleApplyCountRequirement(key, title string, current, required int) roleApplyRequirement {
	return roleApplyRequirement{
		Key: key, Title: title, Text: fmt.Sprintf("当前 %d / 需要 %d", current, required),
		Met: current >= required, Checked: current >= required, Current: current, Required: required,
	}
}

func roleApplyBoolRequirement(key, title string, met bool, yes, no string) roleApplyRequirement {
	text := no
	if met {
		text = yes
	}
	return roleApplyRequirement{Key: key, Title: title, Text: text, Met: met, Checked: met}
}

func (s *Server) roleApplyGameStats() (map[int64]int, map[int64]int, map[int64]bool) {
	created := make(map[int64]int)
	participated := make(map[int64]int)
	completedUsers := make(map[int64]bool)
	for _, game := range s.games.List() {
		if game.CreatorUserID > 0 {
			created[game.CreatorUserID]++
		}
		members := s.games.Members(game.ID)
		for _, userID := range members {
			if userID <= 0 {
				continue
			}
			participated[userID]++
			if game.Status == "pending_review" || game.Status == "completed" {
				completedUsers[userID] = true
			}
		}
	}
	return created, participated, completedUsers
}

func (s *Server) roleApplyInvitedCompletedCount(userID int64, completedUsers map[int64]bool) int {
	relations, err := s.auth.AdminInviteRelations(invites.RelationFilter{InviterUserID: userID})
	if err != nil {
		return 0
	}
	seen := make(map[int64]bool)
	for _, relation := range relations {
		if relation.InviteeUserID > 0 && completedUsers[relation.InviteeUserID] {
			seen[relation.InviteeUserID] = true
		}
	}
	return len(seen)
}

func (s *Server) enterpriseCertificationMet(userID int64) bool {
	item, ok := s.profiles.EnterpriseCertification(userID)
	return ok && item.Status == "approved"
}

func (e roleApplyEligibility) missingBaseRequirements() []string {
	missing := make([]string, 0)
	for _, item := range e.Requirements {
		if item.Key != "plan" && !item.Met {
			missing = append(missing, item.Title)
		}
	}
	return missing
}

func (s *Server) decorateRoleApplyConfig(config map[string]interface{}, eligibility roleApplyEligibility) map[string]interface{} {
	result := cloneObjectMap(config)
	result["eligibility"] = eligibility
	result["canSubmit"] = eligibility.BaseEligible
	byKey := make(map[string]roleApplyRequirement, len(eligibility.Requirements))
	for _, item := range eligibility.Requirements {
		byKey[item.Key] = item
	}
	if raw, ok := result["requirements"]; ok {
		result["requirements"] = decorateRoleApplyRequirements(raw, byKey)
	}
	return result
}

func decorateRoleApplyRequirements(raw interface{}, byKey map[string]roleApplyRequirement) interface{} {
	decorate := func(item map[string]interface{}) map[string]interface{} {
		key := roleApplyRequirementKey(fmt.Sprint(item["title"]))
		condition, ok := byKey[key]
		if !ok {
			return item
		}
		item["key"] = condition.Key
		item["text"] = condition.Text
		item["status"] = condition.Text
		item["met"] = condition.Met
		item["checked"] = condition.Checked
		item["done"] = condition.Checked
		return item
	}
	switch values := raw.(type) {
	case []map[string]interface{}:
		result := make([]map[string]interface{}, 0, len(values))
		for _, item := range values {
			result = append(result, decorate(cloneObjectMap(item)))
		}
		return result
	case []interface{}:
		result := make([]interface{}, 0, len(values))
		for _, value := range values {
			if item, ok := value.(map[string]interface{}); ok {
				result = append(result, decorate(cloneObjectMap(item)))
			} else {
				result = append(result, value)
			}
		}
		return result
	default:
		return raw
	}
}

func roleApplyRequirementKey(title string) string {
	title = strings.TrimSpace(title)
	switch {
	case strings.Contains(title, "等级"):
		return "level"
	case strings.Contains(title, "实名认证"):
		return "realname"
	case strings.Contains(title, "企业认证"):
		return "enterprise"
	case strings.Contains(title, "发起") && strings.Contains(title, "组局"):
		return "created_games"
	case strings.Contains(title, "参与") && strings.Contains(title, "组局"):
		return "participated_games"
	case strings.Contains(title, "邀请") && strings.Contains(title, "完成组局"):
		return "invited_completed_game"
	case strings.Contains(title, "信用分"):
		return "credit_score"
	case strings.Contains(title, "计划书"):
		return "plan"
	default:
		return ""
	}
}
