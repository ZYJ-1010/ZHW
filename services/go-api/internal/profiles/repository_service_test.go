package profiles

import (
	"context"
	"errors"
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

func TestGrantRoleWhitelistReconcilesGuideState(t *testing.T) {
	repo := newFakeProfileRepository()
	service := NewServiceWithRepository(repo)
	if err := service.GrantRoleWhitelist(20, "guide", 7, "一期白名单开通"); err != nil {
		t.Fatal(err)
	}
	if !service.IsGuide(20) {
		t.Fatal("guide role should be active")
	}
	qualification, err := service.GuideQualification(20)
	if err != nil || !qualification.ConditionMet || !qualification.PaymentMet || qualification.GuideOpenStatus != "opened" {
		t.Fatalf("guide qualification was not reconciled: %+v, err=%v", qualification, err)
	}
}

func TestSubmitRoleApplicationStoresEligibilitySnapshot(t *testing.T) {
	service := NewService()
	app, err := service.SubmitRoleApplication(30, SubmitRoleApplicationRequest{
		RoleCode: "expert", Reason: "申请", EligibilitySnapshot: map[string]interface{}{"credit_score": 100},
	})
	if err != nil {
		t.Fatal(err)
	}
	if app.EligibilitySnapshot["credit_score"] != 100.0 && app.EligibilitySnapshot["credit_score"] != 100 {
		t.Fatalf("expected eligibility snapshot, got %+v", app.EligibilitySnapshot)
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
	if qualification.GuideOpenStatus != "ready_for_review" {
		t.Fatalf("expected ready_for_review, got %+v", qualification)
	}
	app, err := service.SubmitRoleApplication(3, SubmitRoleApplicationRequest{RoleCode: "guide", Reason: "want to guide"})
	if err != nil || app.RoleCode != "guide" {
		t.Fatalf("expected guide application without payment requirement, got app=%+v err=%v", app, err)
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

func TestStrictSystemManagementConfigSaveDoesNotFallbackToMemory(t *testing.T) {
	repo := newFakeProfileRepository()
	repo.saveSystemErr = errors.New("profile config persistence failed")
	service := NewServiceWithRepository(repo)

	if _, err := service.SaveSystemManagementConfigStrict(8, "profile-settings", map[string]interface{}{"enabled": true}); err == nil {
		t.Fatal("expected strict profile config persistence failure")
	}
	if _, ok := repo.system[8]; ok {
		t.Fatalf("failed repository save must not create process-local success state: %+v", repo.system[8])
	}
}

func TestStrictSystemManagementConfigReadDoesNotReturnFallbackOnRepositoryFailure(t *testing.T) {
	repo := newFakeProfileRepository()
	repo.getSystemErr = errors.New("profile config read failed")
	service := NewServiceWithRepository(repo)

	value, err := service.SystemManagementConfigStrict(8, "profile-settings", map[string]interface{}{"enabled": true})
	if err == nil || value != nil {
		t.Fatalf("expected strict read failure without fallback value, value=%+v err=%v", value, err)
	}
}

func TestStrictSystemManagementConfigListDoesNotReturnMemoryOnRepositoryFailure(t *testing.T) {
	repo := newFakeProfileRepository()
	repo.listSystemErr = errors.New("profile config list failed")
	service := NewServiceWithRepository(repo)
	service.system[8] = map[string]interface{}{"profile-info": map[string]interface{}{"personalInfo": map[string]interface{}{"avatarAuditStatus": "pending"}}}

	items, err := service.SystemManagementConfigsStrict("profile-info")
	if err == nil || items != nil {
		t.Fatalf("expected strict list failure without process-memory fallback, items=%+v err=%v", items, err)
	}
}

func TestStrictRoleApplicationListsDoNotReturnCacheOnRepositoryFailure(t *testing.T) {
	repo := newFakeProfileRepository()
	repo.listRoleErr = errors.New("role application database unavailable")
	service := NewServiceWithRepository(repo)
	service.apps[1] = RoleApplication{ID: 1, UserID: 8, RoleCode: "expert", Status: "pending"}

	if items, err := service.RoleApplicationsByUserStrict(8); err == nil || items != nil {
		t.Fatalf("expected strict user role application read failure, items=%+v err=%v", items, err)
	}
	if items, err := service.AllRoleApplicationsStrict(); err == nil || items != nil {
		t.Fatalf("expected strict admin role application read failure, items=%+v err=%v", items, err)
	}
}

func TestStrictRoleProfileReadsDoNotReturnStaleCache(t *testing.T) {
	repo := newFakeProfileRepository()
	service := NewServiceWithRepository(repo)
	service.rules[1] = GuideQualificationRule{ID: 1, RuleCode: "旧规则"}
	service.skills[8] = ExpertSkillProfile{UserID: 8, SkillTree: []string{"旧技能"}}
	service.resources[9] = GuideResourceProfile{UserID: 9, ResourceTags: []string{"旧资源"}}
	repo.profileReadErr = errors.New("profile database unavailable")

	if _, err := service.GuideQualificationRulesStrict(); !errors.Is(err, repo.profileReadErr) {
		t.Fatalf("expected qualification rule read error, got %v", err)
	}
	if _, err := service.AdminExpertSkillStrict(8); !errors.Is(err, repo.profileReadErr) {
		t.Fatalf("expected expert skill read error, got %v", err)
	}
	if _, err := service.AllExpertSkillsStrict(); !errors.Is(err, repo.profileReadErr) {
		t.Fatalf("expected expert skill list error, got %v", err)
	}
	if _, err := service.AdminGuideResourceStrict(9); !errors.Is(err, repo.profileReadErr) {
		t.Fatalf("expected guide resource read error, got %v", err)
	}
	if _, err := service.AllGuideResourcesStrict(); !errors.Is(err, repo.profileReadErr) {
		t.Fatalf("expected guide resource list error, got %v", err)
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
	saveSystemErr  error
	getSystemErr   error
	listSystemErr  error
	listRoleErr    error
	profileReadErr error
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
	if r.listRoleErr != nil {
		return nil, r.listRoleErr
	}
	result := make([]RoleApplication, 0, len(r.apps))
	for _, item := range r.apps {
		result = append(result, item)
	}
	return result, nil
}

func (r *fakeProfileRepository) ListRoleApplicationsByUser(ctx context.Context, userID int64) ([]RoleApplication, error) {
	if r.listRoleErr != nil {
		return nil, r.listRoleErr
	}
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
	if r.profileReadErr != nil {
		return GuideQualification{}, false, r.profileReadErr
	}
	q, ok := r.qualifications[userID]
	return q, ok, nil
}

func (r *fakeProfileRepository) SaveGuideQualification(ctx context.Context, q GuideQualification) (GuideQualification, error) {
	r.qualifications[q.UserID] = q
	return q, nil
}

func (r *fakeProfileRepository) ListGuideQualificationRules(ctx context.Context) ([]GuideQualificationRule, error) {
	if r.profileReadErr != nil {
		return nil, r.profileReadErr
	}
	if len(r.rules) == 0 {
		_, _ = r.SaveGuideQualificationRule(ctx, GuideQualificationRule{RuleCode: "default", PaymentRequired: false, Status: "active"})
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
	if r.profileReadErr != nil {
		return ExpertSkillProfile{}, false, r.profileReadErr
	}
	profile, ok := r.experts[userID]
	return profile, ok, nil
}

func (r *fakeProfileRepository) SaveExpertSkill(ctx context.Context, profile ExpertSkillProfile) (ExpertSkillProfile, error) {
	r.experts[profile.UserID] = profile
	return profile, nil
}

func (r *fakeProfileRepository) ListExpertSkills(ctx context.Context) ([]ExpertSkillProfile, error) {
	if r.profileReadErr != nil {
		return nil, r.profileReadErr
	}
	result := make([]ExpertSkillProfile, 0, len(r.experts))
	for _, item := range r.experts {
		result = append(result, item)
	}
	return result, nil
}

func (r *fakeProfileRepository) GetGuideResource(ctx context.Context, userID int64) (GuideResourceProfile, bool, error) {
	if r.profileReadErr != nil {
		return GuideResourceProfile{}, false, r.profileReadErr
	}
	profile, ok := r.resources[userID]
	return profile, ok, nil
}

func (r *fakeProfileRepository) SaveGuideResource(ctx context.Context, profile GuideResourceProfile) (GuideResourceProfile, error) {
	r.resources[profile.UserID] = profile
	return profile, nil
}

func (r *fakeProfileRepository) ListGuideResources(ctx context.Context) ([]GuideResourceProfile, error) {
	if r.profileReadErr != nil {
		return nil, r.profileReadErr
	}
	result := make([]GuideResourceProfile, 0, len(r.resources))
	for _, item := range r.resources {
		result = append(result, item)
	}
	return result, nil
}

func (r *fakeProfileRepository) GetSystemManagementConfig(ctx context.Context, userID int64, key string) (map[string]interface{}, bool, error) {
	if r.getSystemErr != nil {
		return nil, false, r.getSystemErr
	}
	if r.system[userID] == nil {
		return nil, false, nil
	}
	value, ok := r.system[userID][key]
	return cloneObjectMap(value), ok, nil
}

func (r *fakeProfileRepository) SaveSystemManagementConfig(ctx context.Context, userID int64, key string, payload map[string]interface{}) (map[string]interface{}, error) {
	if r.saveSystemErr != nil {
		return nil, r.saveSystemErr
	}
	if r.system[userID] == nil {
		r.system[userID] = make(map[string]map[string]interface{})
	}
	r.system[userID][key] = cloneObjectMap(payload)
	return cloneObjectMap(payload), nil
}

func (r *fakeProfileRepository) ListSystemManagementConfigs(ctx context.Context, key string) ([]SystemManagementConfigItem, error) {
	if r.listSystemErr != nil {
		return nil, r.listSystemErr
	}
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
