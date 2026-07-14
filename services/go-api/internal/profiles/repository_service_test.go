package profiles

import (
	"context"
	"testing"
	"time"
)

func TestSubmitRoleApplicationRejectsActiveRoleAndRecentReapply(t *testing.T) {
	repo := newFakeProfileRepository()
	service := NewServiceWithRepository(repo)

	if err := repo.GrantRole(context.Background(), 10, "expert"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SubmitRoleApplication(10, SubmitRoleApplicationRequest{RoleCode: "expert", Reason: "already active"}); err != ErrRoleAlreadyActive {
		t.Fatalf("expected ErrRoleAlreadyActive, got %v", err)
	}

	now := time.Now()
	_, _ = repo.SaveRoleApplication(context.Background(), RoleApplication{
		UserID: 11, RoleCode: "expert", Status: "rejected", Reason: "retry", CreatedAt: now.Add(-time.Hour), UpdatedAt: now.Add(-time.Hour),
	})
	if _, err := service.SubmitRoleApplication(11, SubmitRoleApplicationRequest{RoleCode: "expert", Reason: "too soon"}); err != ErrRoleApplicationCooldown {
		t.Fatalf("expected ErrRoleApplicationCooldown, got %v", err)
	}

	_, _ = repo.SaveRoleApplication(context.Background(), RoleApplication{
		UserID: 12, RoleCode: "expert", Status: "rejected", Reason: "old rejection", CreatedAt: now.Add(-8 * 24 * time.Hour), UpdatedAt: now.Add(-8 * 24 * time.Hour),
	})
	if _, err := service.SubmitRoleApplication(12, SubmitRoleApplicationRequest{RoleCode: "expert", Reason: "retry allowed"}); err != nil {
		t.Fatalf("expected retry after cooldown to succeed, got %v", err)
	}
}

func TestSubmitRoleApplicationKeepsExpertAndGuideStateSeparate(t *testing.T) {
	repo := newFakeProfileRepository()
	service := NewServiceWithRepository(repo)
	now := time.Now()
	_, _ = repo.SaveRoleApplication(context.Background(), RoleApplication{
		UserID: 20, RoleCode: "expert", Status: "rejected", Reason: "expert rejected", CreatedAt: now, UpdatedAt: now,
	})
	_, _ = repo.SaveGuideQualification(context.Background(), GuideQualification{
		UserID: 20, ConditionMet: true, PaymentMet: true, GuideOpenStatus: "opened", UpdatedAt: now,
	})

	app, err := service.SubmitRoleApplication(20, SubmitRoleApplicationRequest{RoleCode: "guide", Reason: "guide application"})
	if err != nil {
		t.Fatalf("expected guide application to ignore expert rejection, got %v", err)
	}
	if app.RoleCode != "guide" {
		t.Fatalf("expected guide role code, got %+v", app)
	}
}

func TestRepositoryBackedProfilesRequireRolesAndPersist(t *testing.T) {
	service := NewServiceWithRepository(newFakeProfileRepository())

	if _, err := service.UpdateExpertSkill(1, ExpertSkillRequest{SkillTree: []string{"boardgame"}}); err != ErrExpertForbidden {
		t.Fatalf("expected ErrExpertForbidden, got %v", err)
	}
	service.GrantRole(1, "expert")
	expert, err := service.UpdateExpertSkill(1, ExpertSkillRequest{
		SkillTree:   []string{"boardgame"},
		ServiceTags: []string{"host"},
		CaseFileIDs: []int64{11},
	})
	if err != nil {
		t.Fatal(err)
	}
	if expert.Completeness != 100 || len(service.AllExpertSkills()) != 1 {
		t.Fatalf("expected repository expert profile, got %+v", expert)
	}

	if _, err := service.UpdateGuideResource(2, GuideResourceRequest{ResourceTags: []string{"venue"}}); err != ErrGuideForbidden {
		t.Fatalf("expected ErrGuideForbidden, got %v", err)
	}
	service.GrantRole(2, "guide")
	guide, err := service.UpdateGuideResource(2, GuideResourceRequest{
		ResourceTags:    []string{"venue"},
		IndustryTags:    []string{"entertainment"},
		CityCodes:       []string{"110100"},
		ConnectionScale: "100-500",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !service.IsGuide(2) || guide.Completeness != 100 || len(service.AllGuideResources()) != 1 {
		t.Fatalf("expected repository guide profile, got guide=%+v isGuide=%v", guide, service.IsGuide(2))
	}

	qualification, err := service.UpdateGuideQualification(3, UpdateGuideQualificationRequest{ConditionMet: boolPtr(true), PaymentMet: boolPtr(false)})
	if err != nil {
		t.Fatal(err)
	}
	if qualification.GuideOpenStatus != "waiting_payment" {
		t.Fatalf("expected waiting_payment, got %+v", qualification)
	}
	app, err := service.SubmitRoleApplication(3, SubmitRoleApplicationRequest{RoleCode: "guide", Reason: "want to guide"})
	if err != ErrWaitingGuidePayment {
		t.Fatalf("expected waiting payment, got app=%+v err=%v", app, err)
	}

	rules := service.GuideQualificationRules()
	if len(rules) != 1 || rules[0].RuleCode != "default" {
		t.Fatalf("expected default guide rule, got %+v", rules)
	}
	minCredit := 80
	updatedRule, err := service.UpdateGuideQualificationRule(rules[0].ID, UpdateGuideQualificationRuleRequest{MinCreditScore: &minCredit})
	if err != nil {
		t.Fatal(err)
	}
	if updatedRule.MinCreditScore != 80 {
		t.Fatalf("expected updated rule, got %+v", updatedRule)
	}

	saved := service.SaveSystemManagementConfig(2, "feedback-records", map[string]interface{}{
		"items": []map[string]interface{}{{"id": "fb-1", "statusClass": "pending"}},
	})
	loaded := service.SystemManagementConfig(2, "feedback-records", nil)
	if len(saved) == 0 || len(loaded) == 0 {
		t.Fatalf("expected repository-backed system management config, saved=%+v loaded=%+v", saved, loaded)
	}
	listed := service.SystemManagementConfigs("feedback-records")
	if len(listed) != 1 || listed[0].UserID != 2 || listed[0].Key != "feedback-records" {
		t.Fatalf("expected repository-backed system management config list, got %+v", listed)
	}
}

type fakeProfileRepository struct {
	roles          map[int64]map[string]bool
	apps           map[int64]RoleApplication
	nextAppID      int64
	qualifications map[int64]GuideQualification
	rules          map[int64]GuideQualificationRule
	nextRuleID     int64
	experts        map[int64]ExpertSkillProfile
	resources      map[int64]GuideResourceProfile
	system         map[int64]map[string]map[string]interface{}
}

func newFakeProfileRepository() *fakeProfileRepository {
	return &fakeProfileRepository{
		roles:          make(map[int64]map[string]bool),
		apps:           make(map[int64]RoleApplication),
		nextAppID:      1,
		qualifications: make(map[int64]GuideQualification),
		rules:          make(map[int64]GuideQualificationRule),
		nextRuleID:     1,
		experts:        make(map[int64]ExpertSkillProfile),
		resources:      make(map[int64]GuideResourceProfile),
		system:         make(map[int64]map[string]map[string]interface{}),
	}
}

func (r *fakeProfileRepository) GrantRole(ctx context.Context, userID int64, roleCode string) error {
	if r.roles[userID] == nil {
		r.roles[userID] = make(map[string]bool)
	}
	r.roles[userID][roleCode] = true
	return nil
}

func (r *fakeProfileRepository) HasRole(ctx context.Context, userID int64, roleCode string) (bool, error) {
	return r.roles[userID][roleCode], nil
}

func (r *fakeProfileRepository) SaveRoleApplication(ctx context.Context, app RoleApplication) (RoleApplication, error) {
	if app.ID == 0 {
		app.ID = r.nextAppID
		r.nextAppID++
	}
	r.apps[app.ID] = app
	return app, nil
}

func (r *fakeProfileRepository) ListRoleApplications(ctx context.Context) ([]RoleApplication, error) {
	result := make([]RoleApplication, 0, len(r.apps))
	for _, item := range r.apps {
		result = append(result, item)
	}
	return result, nil
}

func (r *fakeProfileRepository) ListRoleApplicationsByUser(ctx context.Context, userID int64) ([]RoleApplication, error) {
	result := make([]RoleApplication, 0)
	for _, item := range r.apps {
		if item.UserID == userID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *fakeProfileRepository) FindRoleApplication(ctx context.Context, applicationID int64) (RoleApplication, bool, error) {
	app, ok := r.apps[applicationID]
	return app, ok, nil
}

func (r *fakeProfileRepository) UpdateRoleApplication(ctx context.Context, app RoleApplication) (RoleApplication, error) {
	r.apps[app.ID] = app
	return app, nil
}

func (r *fakeProfileRepository) GetGuideQualification(ctx context.Context, userID int64) (GuideQualification, bool, error) {
	q, ok := r.qualifications[userID]
	return q, ok, nil
}

func (r *fakeProfileRepository) SaveGuideQualification(ctx context.Context, q GuideQualification) (GuideQualification, error) {
	r.qualifications[q.UserID] = q
	return q, nil
}

func (r *fakeProfileRepository) ListGuideQualificationRules(ctx context.Context) ([]GuideQualificationRule, error) {
	if len(r.rules) == 0 {
		_, _ = r.SaveGuideQualificationRule(ctx, GuideQualificationRule{RuleCode: "default", PaymentRequired: true, Status: "active"})
	}
	result := make([]GuideQualificationRule, 0, len(r.rules))
	for _, item := range r.rules {
		result = append(result, item)
	}
	return result, nil
}

func (r *fakeProfileRepository) SaveGuideQualificationRule(ctx context.Context, rule GuideQualificationRule) (GuideQualificationRule, error) {
	if rule.ID == 0 {
		rule.ID = r.nextRuleID
		r.nextRuleID++
	}
	r.rules[rule.ID] = rule
	return rule, nil
}

func (r *fakeProfileRepository) GetExpertSkill(ctx context.Context, userID int64) (ExpertSkillProfile, bool, error) {
	profile, ok := r.experts[userID]
	return profile, ok, nil
}

func (r *fakeProfileRepository) SaveExpertSkill(ctx context.Context, profile ExpertSkillProfile) (ExpertSkillProfile, error) {
	r.experts[profile.UserID] = profile
	return profile, nil
}

func (r *fakeProfileRepository) ListExpertSkills(ctx context.Context) ([]ExpertSkillProfile, error) {
	result := make([]ExpertSkillProfile, 0, len(r.experts))
	for _, item := range r.experts {
		result = append(result, item)
	}
	return result, nil
}

func (r *fakeProfileRepository) GetGuideResource(ctx context.Context, userID int64) (GuideResourceProfile, bool, error) {
	profile, ok := r.resources[userID]
	return profile, ok, nil
}

func (r *fakeProfileRepository) SaveGuideResource(ctx context.Context, profile GuideResourceProfile) (GuideResourceProfile, error) {
	r.resources[profile.UserID] = profile
	return profile, nil
}

func (r *fakeProfileRepository) ListGuideResources(ctx context.Context) ([]GuideResourceProfile, error) {
	result := make([]GuideResourceProfile, 0, len(r.resources))
	for _, item := range r.resources {
		result = append(result, item)
	}
	return result, nil
}

func (r *fakeProfileRepository) GetSystemManagementConfig(ctx context.Context, userID int64, key string) (map[string]interface{}, bool, error) {
	if r.system[userID] == nil {
		return nil, false, nil
	}
	value, ok := r.system[userID][key]
	return cloneObjectMap(value), ok, nil
}

func (r *fakeProfileRepository) SaveSystemManagementConfig(ctx context.Context, userID int64, key string, payload map[string]interface{}) (map[string]interface{}, error) {
	if r.system[userID] == nil {
		r.system[userID] = make(map[string]map[string]interface{})
	}
	r.system[userID][key] = cloneObjectMap(payload)
	return cloneObjectMap(payload), nil
}

func (r *fakeProfileRepository) ListSystemManagementConfigs(ctx context.Context, key string) ([]SystemManagementConfigItem, error) {
	items := make([]SystemManagementConfigItem, 0)
	for userID, byKey := range r.system {
		if value, ok := byKey[key]; ok {
			items = append(items, SystemManagementConfigItem{UserID: userID, Key: key, Value: cloneObjectMap(value)})
		}
	}
	return items, nil
}

func boolPtr(value bool) *bool {
	return &value
}
