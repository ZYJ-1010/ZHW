package profiles

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	ErrExpertForbidden                 = errors.New("expert role forbidden")
	ErrGuideForbidden                  = errors.New("guide role forbidden")
	ErrInvalidProfile                  = errors.New("invalid profile")
	ErrInvalidRoleApplication          = errors.New("invalid role application")
	ErrDuplicateRoleApplication        = errors.New("duplicate role application")
	ErrRoleAlreadyActive               = errors.New("role already active")
	ErrRoleApplicationCooldown         = errors.New("role application reapply cooldown")
	ErrRoleApplicationNotFound         = errors.New("role application not found")
	ErrRoleApplicationReviewed         = errors.New("role application already reviewed")
	ErrWaitingGuideCondition           = errors.New("waiting guide condition")
	ErrWaitingGuidePayment             = errors.New("waiting guide payment")
	ErrEnterpriseCertificationNotFound = errors.New("enterprise certification not found")
	ErrEnterpriseCertificationPending  = errors.New("enterprise certification already pending")
	ErrInvalidEnterpriseCertification  = errors.New("invalid enterprise certification")
	ErrInvalidSystemManagementConfig   = errors.New("invalid system management config")
)

const roleApplicationReapplyCooldown = 7 * 24 * time.Hour

var unifiedSocialCreditCodePattern = regexp.MustCompile(`^[0-9A-Z]{18}$`)

type ExpertSkillProfile struct {
	UserID       int64     `json:"userId"`
	SkillTree    []string  `json:"skillTree"`
	ServiceTags  []string  `json:"serviceTags"`
	CaseFileIDs  []int64   `json:"caseFileIds"`
	Completeness int       `json:"completeness"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type GuideResourceProfile struct {
	UserID          int64     `json:"userId"`
	ResourceTags    []string  `json:"resourceTags"`
	IndustryTags    []string  `json:"industryTags"`
	CityCodes       []string  `json:"cityCodes"`
	ConnectionScale string    `json:"connectionScale"`
	Completeness    int       `json:"completeness"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type RoleApplication struct {
	ID                  int64                  `json:"id"`
	UserID              int64                  `json:"userId"`
	RoleCode            string                 `json:"roleCode"`
	Status              string                 `json:"status"`
	Reason              string                 `json:"reason,omitempty"`
	AbilityDescription  string                 `json:"abilityDescription,omitempty"`
	EligibilitySnapshot map[string]interface{} `json:"eligibilitySnapshot,omitempty"`
	ProofFileIDs        []int64                `json:"proofFileIds,omitempty"`
	RejectReason        string                 `json:"rejectReason,omitempty"`
	ReviewAdminID       int64                  `json:"reviewAdminId,omitempty"`
	ReviewRemark        string                 `json:"reviewRemark,omitempty"`
	CertificateNo       string                 `json:"certNo,omitempty"`
	CertifiedAt         string                 `json:"certifiedAt,omitempty"`
	CreatedAt           time.Time              `json:"createdAt"`
	UpdatedAt           time.Time              `json:"updatedAt"`
}

type GuideQualification struct {
	UserID          int64     `json:"userId"`
	ConditionMet    bool      `json:"conditionMet"`
	PaymentMet      bool      `json:"paymentMet"`
	GuideOpenStatus string    `json:"guideOpenStatus"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type GuideQualificationRule struct {
	ID                int64     `json:"id"`
	RuleCode          string    `json:"ruleCode"`
	MinInviteCount    int       `json:"minInviteCount"`
	MinCreditScore    int       `json:"minCreditScore"`
	MinCompletedGames int       `json:"minCompletedGames"`
	PaymentRequired   bool      `json:"paymentRequired"`
	Status            string    `json:"status"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type RoleSnapshot struct {
	Roles         []string          `json:"roles"`
	RoleStatusMap map[string]string `json:"roleStatusMap"`
}

type EnterpriseCertification struct {
	ID                      int64     `json:"id"`
	UserID                  int64     `json:"userId"`
	CompanyName             string    `json:"companyName"`
	UnifiedSocialCreditCode string    `json:"unifiedSocialCreditCode"`
	LegalPerson             string    `json:"legalPerson"`
	BusinessLicenseFileID   int64     `json:"businessLicenseFileId"`
	PublicAccountFileID     int64     `json:"publicAccountFileId"`
	Status                  string    `json:"status"`
	RejectReason            string    `json:"rejectReason,omitempty"`
	ReviewAdminID           int64     `json:"reviewAdminId,omitempty"`
	ReviewRemark            string    `json:"reviewRemark,omitempty"`
	CreatedAt               time.Time `json:"createdAt"`
	UpdatedAt               time.Time `json:"updatedAt"`
}

type SubmitEnterpriseCertificationRequest struct {
	CompanyName             string `json:"companyName"`
	UnifiedSocialCreditCode string `json:"unifiedSocialCreditCode"`
	LegalPerson             string `json:"legalPerson"`
	BusinessLicenseFileID   int64  `json:"businessLicenseFileId"`
	PublicAccountFileID     int64  `json:"publicAccountFileId"`
}

type ReviewEnterpriseCertificationRequest struct {
	Approve bool   `json:"approve"`
	Remark  string `json:"remark"`
}

type ExpertSkillRequest struct {
	SkillTree   []string `json:"skillTree"`
	ServiceTags []string `json:"serviceTags"`
	CaseFileIDs []int64  `json:"caseFileIds"`
}

type GuideResourceRequest struct {
	ResourceTags    []string `json:"resourceTags"`
	IndustryTags    []string `json:"industryTags"`
	CityCodes       []string `json:"cityCodes"`
	ConnectionScale string   `json:"connectionScale"`
}

type SubmitRoleApplicationRequest struct {
	RoleCode            string                 `json:"roleCode"`
	Reason              string                 `json:"reason"`
	AbilityDescription  string                 `json:"abilityDescription"`
	ProofFileIDs        []int64                `json:"proofFileIds"`
	EligibilitySnapshot map[string]interface{} `json:"eligibilitySnapshot,omitempty"`
}

type ReviewRoleApplicationRequest struct {
	Approve bool   `json:"approve"`
	Remark  string `json:"remark"`
}

type UpdateGuideQualificationRequest struct {
	ConditionMet *bool `json:"conditionMet"`
	PaymentMet   *bool `json:"paymentMet"`
}

type UpdateGuideQualificationRuleRequest struct {
	MinInviteCount    *int    `json:"minInviteCount"`
	MinCreditScore    *int    `json:"minCreditScore"`
	MinCompletedGames *int    `json:"minCompletedGames"`
	PaymentRequired   *bool   `json:"paymentRequired"`
	Status            *string `json:"status"`
}

type SystemManagementConfigItem struct {
	UserID int64                  `json:"userId"`
	Key    string                 `json:"key"`
	Value  map[string]interface{} `json:"value"`
}

type Repository interface {
	GrantRole(ctx context.Context, userID int64, roleCode string) error
	HasRole(ctx context.Context, userID int64, roleCode string) (bool, error)
	SaveRoleApplication(ctx context.Context, app RoleApplication) (RoleApplication, error)
	ListRoleApplications(ctx context.Context) ([]RoleApplication, error)
	ListRoleApplicationsByUser(ctx context.Context, userID int64) ([]RoleApplication, error)
	FindRoleApplication(ctx context.Context, applicationID int64) (RoleApplication, bool, error)
	UpdateRoleApplication(ctx context.Context, app RoleApplication) (RoleApplication, error)
	GetGuideQualification(ctx context.Context, userID int64) (GuideQualification, bool, error)
	SaveGuideQualification(ctx context.Context, qualification GuideQualification) (GuideQualification, error)
	ListGuideQualificationRules(ctx context.Context) ([]GuideQualificationRule, error)
	SaveGuideQualificationRule(ctx context.Context, rule GuideQualificationRule) (GuideQualificationRule, error)
	GetExpertSkill(ctx context.Context, userID int64) (ExpertSkillProfile, bool, error)
	SaveExpertSkill(ctx context.Context, profile ExpertSkillProfile) (ExpertSkillProfile, error)
	ListExpertSkills(ctx context.Context) ([]ExpertSkillProfile, error)
	GetGuideResource(ctx context.Context, userID int64) (GuideResourceProfile, bool, error)
	SaveGuideResource(ctx context.Context, profile GuideResourceProfile) (GuideResourceProfile, error)
	ListGuideResources(ctx context.Context) ([]GuideResourceProfile, error)
	GetSystemManagementConfig(ctx context.Context, userID int64, key string) (map[string]interface{}, bool, error)
	SaveSystemManagementConfig(ctx context.Context, userID int64, key string, payload map[string]interface{}) (map[string]interface{}, error)
	ListSystemManagementConfigs(ctx context.Context, key string) ([]SystemManagementConfigItem, error)
}

type roleWhitelistRepository interface {
	GrantRoleWhitelist(ctx context.Context, userID int64, roleCode string, adminID int64, reason string) error
}

type enterpriseRepository interface {
	SaveEnterpriseCertification(ctx context.Context, item EnterpriseCertification) (EnterpriseCertification, error)
	FindEnterpriseCertification(ctx context.Context, userID int64) (EnterpriseCertification, bool, error)
	ListEnterpriseCertifications(ctx context.Context, status string) ([]EnterpriseCertification, error)
	ReviewEnterpriseCertification(ctx context.Context, item EnterpriseCertification) (EnterpriseCertification, error)
}

type reviewedRoleApplicationRepository interface {
	SaveReviewedRoleApplication(ctx context.Context, app RoleApplication) (RoleApplication, error)
}

type Service struct {
	mu         sync.RWMutex
	experts    map[int64]bool
	guides     map[int64]bool
	nextAppID  int64
	nextRuleID int64
	apps       map[int64]RoleApplication
	guideQual  map[int64]GuideQualification
	rules      map[int64]GuideQualificationRule
	skills     map[int64]ExpertSkillProfile
	resources  map[int64]GuideResourceProfile
	enterprise map[int64]EnterpriseCertification
	system     map[int64]map[string]interface{}
	repo       Repository
}

func cloneMap(input map[string]interface{}) map[string]interface{} {
	if len(input) == 0 {
		return nil
	}
	result := make(map[string]interface{}, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func NewService() *Service {
	return NewServiceWithRepository(nil)
}

func NewServiceWithRepository(repo Repository) *Service {
	return &Service{
		experts:    make(map[int64]bool),
		guides:     make(map[int64]bool),
		nextAppID:  1,
		nextRuleID: 1,
		apps:       make(map[int64]RoleApplication),
		guideQual:  make(map[int64]GuideQualification),
		rules:      make(map[int64]GuideQualificationRule),
		skills:     make(map[int64]ExpertSkillProfile),
		resources:  make(map[int64]GuideResourceProfile),
		enterprise: make(map[int64]EnterpriseCertification),
		system:     make(map[int64]map[string]interface{}),
		repo:       repo,
	}
}

func (s *Service) GrantRole(userID int64, roleCode string) {
	_ = s.GrantRoleStrict(userID, roleCode)
}

func (s *Service) GrantRoleStrict(userID int64, roleCode string) error {
	if s.repo != nil {
		return s.repo.GrantRole(context.Background(), userID, roleCode)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	switch roleCode {
	case "expert":
		s.experts[userID] = true
	case "guide":
		s.guides[userID] = true
	}
	return nil
}

// GrantRoleWhitelist is the一期运营开通路径. It still requires the caller
// to verify real-name status at the app layer, while atomically reconciling
// the role, guide qualification and any pending application in storage.
func (s *Service) GrantRoleWhitelist(userID int64, roleCode string, adminID int64, reason string) error {
	roleCode = strings.ToLower(strings.TrimSpace(roleCode))
	if userID <= 0 || (roleCode != "expert" && roleCode != "guide") {
		return ErrInvalidRoleApplication
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "后台白名单开通"
	}
	if s.repo != nil {
		if repository, ok := s.repo.(roleWhitelistRepository); ok {
			return repository.GrantRoleWhitelist(context.Background(), userID, roleCode, adminID, reason)
		}
		if err := s.repo.GrantRole(context.Background(), userID, roleCode); err != nil {
			return err
		}
		if roleCode == "guide" {
			qualification, _ := s.GuideQualification(userID)
			qualification.UserID = userID
			qualification.ConditionMet = true
			qualification.PaymentMet = true
			qualification.GuideOpenStatus = "opened"
			qualification.UpdatedAt = time.Now()
			_, err := s.repo.SaveGuideQualification(context.Background(), qualification)
			return err
		}
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if roleCode == "expert" {
		s.experts[userID] = true
	} else {
		s.guides[userID] = true
		qualification := s.guideQual[userID]
		qualification.UserID = userID
		qualification.ConditionMet = true
		qualification.PaymentMet = true
		qualification.GuideOpenStatus = "opened"
		qualification.UpdatedAt = time.Now()
		s.guideQual[userID] = qualification
	}
	for id, application := range s.apps {
		if application.UserID != userID || application.RoleCode != roleCode || application.Status != "pending" {
			continue
		}
		application.Status = "approved"
		application.ReviewAdminID = adminID
		application.ReviewRemark = reason
		application.UpdatedAt = time.Now()
		if application.CertificateNo == "" {
			application.CertificateNo = fmt.Sprintf("ZHW-WL-%05d-%d", application.ID, application.UpdatedAt.Year())
		}
		application.CertifiedAt = application.UpdatedAt.Format(time.RFC3339)
		s.apps[id] = application
	}
	return nil
}

func (s *Service) EnterpriseCertification(userID int64) (EnterpriseCertification, bool) {
	item, found, _ := s.EnterpriseCertificationStrict(userID)
	return item, found
}

func (s *Service) EnterpriseCertificationStrict(userID int64) (EnterpriseCertification, bool, error) {
	if userID <= 0 {
		return EnterpriseCertification{}, false, nil
	}
	if repository, ok := s.repo.(enterpriseRepository); ok {
		item, found, err := repository.FindEnterpriseCertification(context.Background(), userID)
		if err != nil {
			return EnterpriseCertification{}, false, err
		}
		return item, found, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.enterprise[userID]
	return item, ok, nil
}

func (s *Service) SubmitEnterpriseCertification(userID int64, req SubmitEnterpriseCertificationRequest) (EnterpriseCertification, error) {
	req.CompanyName = strings.TrimSpace(req.CompanyName)
	req.UnifiedSocialCreditCode = strings.ToUpper(strings.TrimSpace(req.UnifiedSocialCreditCode))
	req.LegalPerson = strings.TrimSpace(req.LegalPerson)
	if userID <= 0 || len([]rune(req.CompanyName)) < 2 || len([]rune(req.CompanyName)) > 100 ||
		!unifiedSocialCreditCodePattern.MatchString(req.UnifiedSocialCreditCode) || req.LegalPerson == "" || len([]rune(req.LegalPerson)) > 50 ||
		req.BusinessLicenseFileID <= 0 || req.PublicAccountFileID <= 0 {
		return EnterpriseCertification{}, ErrInvalidEnterpriseCertification
	}
	if item, ok := s.EnterpriseCertification(userID); ok && item.Status == "pending" {
		return EnterpriseCertification{}, ErrEnterpriseCertificationPending
	}
	now := time.Now()
	item := EnterpriseCertification{
		UserID: userID, CompanyName: req.CompanyName, UnifiedSocialCreditCode: req.UnifiedSocialCreditCode,
		LegalPerson: req.LegalPerson, BusinessLicenseFileID: req.BusinessLicenseFileID,
		PublicAccountFileID: req.PublicAccountFileID, Status: "pending", CreatedAt: now, UpdatedAt: now,
	}
	if repository, ok := s.repo.(enterpriseRepository); ok {
		saved, err := repository.SaveEnterpriseCertification(context.Background(), item)
		return saved, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item.ID = int64(len(s.enterprise) + 1)
	s.enterprise[userID] = item
	return item, nil
}

func (s *Service) ReviewEnterpriseCertification(adminID int64, userID int64, req ReviewEnterpriseCertificationRequest) (EnterpriseCertification, error) {
	req.Remark = strings.TrimSpace(req.Remark)
	if adminID <= 0 || userID <= 0 || (!req.Approve && req.Remark == "") || len(req.Remark) > 500 {
		return EnterpriseCertification{}, ErrInvalidEnterpriseCertification
	}
	item, ok := s.EnterpriseCertification(userID)
	if !ok {
		return EnterpriseCertification{}, ErrEnterpriseCertificationNotFound
	}
	if item.Status != "pending" {
		return EnterpriseCertification{}, ErrInvalidEnterpriseCertification
	}
	item.ReviewAdminID = adminID
	item.ReviewRemark = req.Remark
	item.RejectReason = ""
	item.Status = "approved"
	if !req.Approve {
		item.Status = "rejected"
		item.RejectReason = req.Remark
	}
	item.UpdatedAt = time.Now()
	if repository, ok := s.repo.(enterpriseRepository); ok {
		return repository.ReviewEnterpriseCertification(context.Background(), item)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enterprise[userID] = item
	return item, nil
}

func (s *Service) AllEnterpriseCertifications(status string) []EnterpriseCertification {
	items, _ := s.AllEnterpriseCertificationsStrict(status)
	return items
}

func (s *Service) AllEnterpriseCertificationsStrict(status string) ([]EnterpriseCertification, error) {
	status = strings.TrimSpace(status)
	if repository, ok := s.repo.(enterpriseRepository); ok {
		items, err := repository.ListEnterpriseCertifications(context.Background(), status)
		if err != nil {
			return nil, err
		}
		return items, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]EnterpriseCertification, 0, len(s.enterprise))
	for _, item := range s.enterprise {
		if status != "" && item.Status != status {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) IsGuide(userID int64) bool {
	if s.repo != nil {
		ok, err := s.repo.HasRole(context.Background(), userID, "guide")
		return err == nil && ok
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.guides[userID]
}

func (s *Service) RoleSnapshot(userID int64) RoleSnapshot {
	snapshot, _ := s.RoleSnapshotStrict(userID)
	return snapshot
}

func (s *Service) RoleSnapshotStrict(userID int64) (RoleSnapshot, error) {
	statusMap := map[string]string{
		"player": "approved",
		"expert": "none",
		"guide":  "none",
	}
	roles := []string{"player"}
	expert, err := s.hasRoleStrict(userID, "expert")
	if err != nil {
		return RoleSnapshot{}, err
	}
	if expert {
		statusMap["expert"] = "approved"
		roles = append(roles, "expert")
	}
	guide, err := s.hasRoleStrict(userID, "guide")
	if err != nil {
		return RoleSnapshot{}, err
	}
	if guide {
		statusMap["guide"] = "approved"
		roles = append(roles, "guide")
	}
	return RoleSnapshot{Roles: roles, RoleStatusMap: statusMap}, nil
}

func (s *Service) hasRole(userID int64, roleCode string) bool {
	ok, _ := s.hasRoleStrict(userID, roleCode)
	return ok
}

func (s *Service) hasRoleStrict(userID int64, roleCode string) (bool, error) {
	if s.repo != nil {
		return s.repo.HasRole(context.Background(), userID, roleCode)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	switch roleCode {
	case "expert":
		return s.experts[userID], nil
	case "guide":
		return s.guides[userID], nil
	default:
		return roleCode == "player", nil
	}
}

func (s *Service) SubmitRoleApplication(userID int64, req SubmitRoleApplicationRequest) (RoleApplication, error) {
	req.RoleCode = strings.TrimSpace(req.RoleCode)
	req.Reason = strings.TrimSpace(req.Reason)
	req.AbilityDescription = strings.TrimSpace(req.AbilityDescription)
	if userID <= 0 || !validRoleCode(req.RoleCode) || req.Reason == "" || len(req.Reason) > 500 || len(req.AbilityDescription) > 1000 || !validFileIDs(req.ProofFileIDs, 12) {
		return RoleApplication{}, ErrInvalidRoleApplication
	}
	if s.repo != nil {
		active, err := s.repo.HasRole(context.Background(), userID, req.RoleCode)
		if err != nil {
			return RoleApplication{}, err
		}
		if active {
			return RoleApplication{}, ErrRoleAlreadyActive
		}
	} else if s.hasRole(userID, req.RoleCode) {
		return RoleApplication{}, ErrRoleAlreadyActive
	}
	if s.repo != nil {
		items, err := s.repo.ListRoleApplicationsByUser(context.Background(), userID)
		if err != nil {
			return RoleApplication{}, err
		}
		if hasPendingRoleApplication(items, req.RoleCode) {
			return RoleApplication{}, ErrDuplicateRoleApplication
		}
		now := time.Now()
		if roleApplicationInCooldown(items, req.RoleCode, now) {
			return RoleApplication{}, ErrRoleApplicationCooldown
		}
		return s.repo.SaveRoleApplication(context.Background(), RoleApplication{
			UserID:              userID,
			RoleCode:            req.RoleCode,
			Status:              "pending",
			Reason:              req.Reason,
			AbilityDescription:  req.AbilityDescription,
			ProofFileIDs:        append([]int64(nil), req.ProofFileIDs...),
			EligibilitySnapshot: cloneMap(req.EligibilitySnapshot),
			CreatedAt:           now,
			UpdatedAt:           now,
		})
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	items := s.roleApplicationsByUserLocked(userID)
	if hasPendingRoleApplication(items, req.RoleCode) {
		return RoleApplication{}, ErrDuplicateRoleApplication
	}
	now := time.Now()
	if roleApplicationInCooldown(items, req.RoleCode, now) {
		return RoleApplication{}, ErrRoleApplicationCooldown
	}
	app := RoleApplication{
		ID:                  s.nextAppID,
		UserID:              userID,
		RoleCode:            req.RoleCode,
		Status:              "pending",
		Reason:              req.Reason,
		AbilityDescription:  req.AbilityDescription,
		ProofFileIDs:        append([]int64(nil), req.ProofFileIDs...),
		EligibilitySnapshot: cloneMap(req.EligibilitySnapshot),
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	s.nextAppID++
	s.apps[app.ID] = app
	return app, nil
}

func roleApplicationInCooldown(items []RoleApplication, roleCode string, now time.Time) bool {
	for _, item := range items {
		if item.RoleCode != roleCode || item.Status != "rejected" {
			continue
		}
		rejectedAt := item.UpdatedAt
		if rejectedAt.IsZero() {
			rejectedAt = item.CreatedAt
		}
		if !rejectedAt.IsZero() && now.Before(rejectedAt.Add(roleApplicationReapplyCooldown)) {
			return true
		}
	}
	return false
}

func (s *Service) RoleApplicationsByUser(userID int64) []RoleApplication {
	items, _ := s.RoleApplicationsByUserStrict(userID)
	return items
}

func (s *Service) RoleApplicationsByUserStrict(userID int64) ([]RoleApplication, error) {
	if s.repo != nil {
		return s.repo.ListRoleApplicationsByUser(context.Background(), userID)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.roleApplicationsByUserLocked(userID), nil
}

func (s *Service) AllRoleApplications() []RoleApplication {
	items, _ := s.AllRoleApplicationsStrict()
	return items
}

func (s *Service) AllRoleApplicationsStrict() ([]RoleApplication, error) {
	if s.repo != nil {
		return s.repo.ListRoleApplications(context.Background())
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]RoleApplication, 0, len(s.apps))
	for _, app := range s.apps {
		items = append(items, app)
	}
	return items, nil
}

func (s *Service) ReviewRoleApplication(adminID int64, applicationID int64, req ReviewRoleApplicationRequest) (RoleApplication, error) {
	req.Remark = strings.TrimSpace(req.Remark)
	if applicationID <= 0 || req.Remark == "" || len(req.Remark) > 500 {
		return RoleApplication{}, ErrInvalidRoleApplication
	}
	if s.repo != nil {
		app, ok, err := s.repo.FindRoleApplication(context.Background(), applicationID)
		if err != nil {
			return RoleApplication{}, err
		}
		if !ok {
			return RoleApplication{}, ErrRoleApplicationNotFound
		}
		updated, err := s.reviewRoleApplication(app, adminID, req)
		if err != nil {
			return RoleApplication{}, err
		}
		if repository, ok := s.repo.(reviewedRoleApplicationRepository); ok {
			return repository.SaveReviewedRoleApplication(context.Background(), updated)
		}
		saved, err := s.repo.UpdateRoleApplication(context.Background(), updated)
		if err != nil {
			return RoleApplication{}, err
		}
		if saved.Status == "approved" {
			if err := s.repo.GrantRole(context.Background(), saved.UserID, saved.RoleCode); err != nil {
				return RoleApplication{}, err
			}
		}
		return saved, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	app, ok := s.apps[applicationID]
	if !ok {
		return RoleApplication{}, ErrRoleApplicationNotFound
	}
	updated, err := s.reviewRoleApplication(app, adminID, req)
	if err != nil {
		return RoleApplication{}, err
	}
	s.apps[applicationID] = updated
	if updated.Status == "approved" {
		switch updated.RoleCode {
		case "expert":
			s.experts[updated.UserID] = true
		case "guide":
			s.guides[updated.UserID] = true
			q := s.guideQual[updated.UserID]
			q.UserID = updated.UserID
			q.ConditionMet = true
			q.PaymentMet = true
			q.GuideOpenStatus = "opened"
			q.UpdatedAt = time.Now()
			s.guideQual[updated.UserID] = q
		}
	}
	return updated, nil
}

func (s *Service) GuideQualification(userID int64) (GuideQualification, error) {
	if userID <= 0 {
		return GuideQualification{}, ErrInvalidRoleApplication
	}
	if s.repo != nil {
		if q, ok, err := s.repo.GetGuideQualification(context.Background(), userID); err != nil {
			return GuideQualification{}, err
		} else if ok {
			return q, nil
		}
		return GuideQualification{UserID: userID, GuideOpenStatus: "waiting_condition", UpdatedAt: time.Now()}, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if q, ok := s.guideQual[userID]; ok {
		return q, nil
	}
	return GuideQualification{UserID: userID, GuideOpenStatus: "waiting_condition", UpdatedAt: time.Now()}, nil
}

func (s *Service) UpdateGuideQualification(userID int64, req UpdateGuideQualificationRequest) (GuideQualification, error) {
	if userID <= 0 || (req.ConditionMet == nil && req.PaymentMet == nil) {
		return GuideQualification{}, ErrInvalidRoleApplication
	}
	current, err := s.GuideQualification(userID)
	if err != nil {
		return GuideQualification{}, err
	}
	if req.ConditionMet != nil {
		current.ConditionMet = *req.ConditionMet
	}
	if req.PaymentMet != nil {
		current.PaymentMet = *req.PaymentMet
	}
	current.GuideOpenStatus = guideOpenStatus(current.ConditionMet, current.PaymentMet, s.IsGuide(userID))
	current.UpdatedAt = time.Now()
	if s.repo != nil {
		return s.repo.SaveGuideQualification(context.Background(), current)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.guideQual[userID] = current
	return current, nil
}

func (s *Service) GuideQualificationRules() []GuideQualificationRule {
	items, _ := s.GuideQualificationRulesStrict()
	return items
}

func (s *Service) GuideQualificationRulesStrict() ([]GuideQualificationRule, error) {
	if s.repo != nil {
		return s.repo.ListGuideQualificationRules(context.Background())
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureDefaultGuideRuleLocked()
	items := make([]GuideQualificationRule, 0, len(s.rules))
	for _, rule := range s.rules {
		items = append(items, rule)
	}
	return items, nil
}

func (s *Service) UpdateGuideQualificationRule(ruleID int64, req UpdateGuideQualificationRuleRequest) (GuideQualificationRule, error) {
	if ruleID <= 0 || (req.MinInviteCount == nil && req.MinCreditScore == nil && req.MinCompletedGames == nil && req.PaymentRequired == nil && req.Status == nil) {
		return GuideQualificationRule{}, ErrInvalidRoleApplication
	}
	if req.MinInviteCount != nil && *req.MinInviteCount < 0 {
		return GuideQualificationRule{}, ErrInvalidRoleApplication
	}
	if req.MinCreditScore != nil && (*req.MinCreditScore < 0 || *req.MinCreditScore > 100) {
		return GuideQualificationRule{}, ErrInvalidRoleApplication
	}
	if req.MinCompletedGames != nil && *req.MinCompletedGames < 0 {
		return GuideQualificationRule{}, ErrInvalidRoleApplication
	}
	status := ""
	if req.Status != nil {
		status = strings.TrimSpace(*req.Status)
		if status != "active" && status != "disabled" {
			return GuideQualificationRule{}, ErrInvalidRoleApplication
		}
	}
	if s.repo != nil {
		items, err := s.repo.ListGuideQualificationRules(context.Background())
		if err != nil {
			return GuideQualificationRule{}, err
		}
		var current GuideQualificationRule
		found := false
		for _, item := range items {
			if item.ID == ruleID {
				current = item
				found = true
				break
			}
		}
		if !found {
			return GuideQualificationRule{}, ErrRoleApplicationNotFound
		}
		return s.repo.SaveGuideQualificationRule(context.Background(), applyGuideRuleUpdate(current, req, status))
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureDefaultGuideRuleLocked()
	current, ok := s.rules[ruleID]
	if !ok {
		return GuideQualificationRule{}, ErrRoleApplicationNotFound
	}
	updated := applyGuideRuleUpdate(current, req, status)
	s.rules[ruleID] = updated
	return updated, nil
}

func (s *Service) ExpertSkill(userID int64) (ExpertSkillProfile, error) {
	if s.repo != nil {
		ok, err := s.repo.HasRole(context.Background(), userID, "expert")
		if err != nil {
			return ExpertSkillProfile{}, err
		}
		if !ok {
			return ExpertSkillProfile{}, ErrExpertForbidden
		}
		if profile, found, err := s.repo.GetExpertSkill(context.Background(), userID); err == nil && found {
			return profile, nil
		} else if err != nil {
			return ExpertSkillProfile{}, err
		}
		return ExpertSkillProfile{UserID: userID, SkillTree: []string{}, ServiceTags: []string{}, CaseFileIDs: []int64{}, UpdatedAt: time.Now()}, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.experts[userID] {
		return ExpertSkillProfile{}, ErrExpertForbidden
	}
	if profile, ok := s.skills[userID]; ok {
		return profile, nil
	}
	return ExpertSkillProfile{UserID: userID, SkillTree: []string{}, ServiceTags: []string{}, CaseFileIDs: []int64{}, UpdatedAt: time.Now()}, nil
}

func (s *Service) AdminExpertSkill(userID int64) ExpertSkillProfile {
	profile, _ := s.AdminExpertSkillStrict(userID)
	return profile
}

func (s *Service) AdminExpertSkillStrict(userID int64) (ExpertSkillProfile, error) {
	if s.repo != nil {
		profile, ok, err := s.repo.GetExpertSkill(context.Background(), userID)
		if err != nil {
			return ExpertSkillProfile{}, err
		}
		if ok {
			return profile, nil
		}
		return ExpertSkillProfile{UserID: userID, SkillTree: []string{}, ServiceTags: []string{}, CaseFileIDs: []int64{}, UpdatedAt: time.Now()}, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if profile, ok := s.skills[userID]; ok {
		return profile, nil
	}
	return ExpertSkillProfile{UserID: userID, SkillTree: []string{}, ServiceTags: []string{}, CaseFileIDs: []int64{}, UpdatedAt: time.Now()}, nil
}

func (s *Service) AllExpertSkills() []ExpertSkillProfile {
	items, _ := s.AllExpertSkillsStrict()
	return items
}

func (s *Service) AllExpertSkillsStrict() ([]ExpertSkillProfile, error) {
	if s.repo != nil {
		return s.repo.ListExpertSkills(context.Background())
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]ExpertSkillProfile, 0, len(s.skills))
	for _, profile := range s.skills {
		items = append(items, profile)
	}
	return items, nil
}

func (s *Service) UpdateExpertSkill(userID int64, req ExpertSkillRequest) (ExpertSkillProfile, error) {
	req.SkillTree = normalizeTags(req.SkillTree, 20, 40)
	req.ServiceTags = normalizeTags(req.ServiceTags, 20, 40)
	if (len(req.SkillTree) == 0 && len(req.ServiceTags) == 0) || !validFileIDs(req.CaseFileIDs, 12) {
		return ExpertSkillProfile{}, ErrInvalidProfile
	}
	if s.repo != nil {
		ok, err := s.repo.HasRole(context.Background(), userID, "expert")
		if err != nil {
			return ExpertSkillProfile{}, err
		}
		if !ok {
			return ExpertSkillProfile{}, ErrExpertForbidden
		}
		return s.repo.SaveExpertSkill(context.Background(), expertProfileFromRequest(userID, req))
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.experts[userID] {
		return ExpertSkillProfile{}, ErrExpertForbidden
	}
	profile := expertProfileFromRequest(userID, req)
	s.skills[userID] = profile
	return profile, nil
}

func (s *Service) GuideResource(userID int64) (GuideResourceProfile, error) {
	if s.repo != nil {
		ok, err := s.repo.HasRole(context.Background(), userID, "guide")
		if err != nil {
			return GuideResourceProfile{}, err
		}
		if !ok {
			return GuideResourceProfile{}, ErrGuideForbidden
		}
		if profile, found, err := s.repo.GetGuideResource(context.Background(), userID); err == nil && found {
			return profile, nil
		} else if err != nil {
			return GuideResourceProfile{}, err
		}
		return GuideResourceProfile{UserID: userID, ResourceTags: []string{}, IndustryTags: []string{}, CityCodes: []string{}, UpdatedAt: time.Now()}, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.guides[userID] {
		return GuideResourceProfile{}, ErrGuideForbidden
	}
	if profile, ok := s.resources[userID]; ok {
		return profile, nil
	}
	return GuideResourceProfile{UserID: userID, ResourceTags: []string{}, IndustryTags: []string{}, CityCodes: []string{}, UpdatedAt: time.Now()}, nil
}

func (s *Service) AdminGuideResource(userID int64) GuideResourceProfile {
	profile, _ := s.AdminGuideResourceStrict(userID)
	return profile
}

func (s *Service) AdminGuideResourceStrict(userID int64) (GuideResourceProfile, error) {
	if s.repo != nil {
		profile, ok, err := s.repo.GetGuideResource(context.Background(), userID)
		if err != nil {
			return GuideResourceProfile{}, err
		}
		if ok {
			return profile, nil
		}
		return GuideResourceProfile{UserID: userID, ResourceTags: []string{}, IndustryTags: []string{}, CityCodes: []string{}, UpdatedAt: time.Now()}, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if profile, ok := s.resources[userID]; ok {
		return profile, nil
	}
	return GuideResourceProfile{UserID: userID, ResourceTags: []string{}, IndustryTags: []string{}, CityCodes: []string{}, UpdatedAt: time.Now()}, nil
}

func (s *Service) AllGuideResources() []GuideResourceProfile {
	items, _ := s.AllGuideResourcesStrict()
	return items
}

func (s *Service) AllGuideResourcesStrict() ([]GuideResourceProfile, error) {
	if s.repo != nil {
		return s.repo.ListGuideResources(context.Background())
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]GuideResourceProfile, 0, len(s.resources))
	for _, profile := range s.resources {
		items = append(items, profile)
	}
	return items, nil
}

func (s *Service) SystemManagementConfig(userID int64, key string, fallback map[string]interface{}) map[string]interface{} {
	value, err := s.SystemManagementConfigStrict(userID, key, fallback)
	if err == nil {
		return value
	}
	return cloneObjectMap(fallback)
}

func (s *Service) SystemManagementConfigStrict(userID int64, key string, fallback map[string]interface{}) (map[string]interface{}, error) {
	if userID <= 0 || strings.TrimSpace(key) == "" {
		return nil, ErrInvalidSystemManagementConfig
	}
	key = strings.TrimSpace(key)
	if s.repo != nil {
		value, ok, err := s.repo.GetSystemManagementConfig(context.Background(), userID, key)
		if err != nil {
			return nil, err
		}
		if ok {
			return cloneObjectMap(value), nil
		}
		return cloneObjectMap(fallback), nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if byKey, ok := s.system[userID]; ok {
		if value, ok := byKey[key].(map[string]interface{}); ok {
			return cloneObjectMap(value), nil
		}
	}
	return cloneObjectMap(fallback), nil
}

func (s *Service) SaveSystemManagementConfig(userID int64, key string, payload map[string]interface{}) map[string]interface{} {
	saved, err := s.SaveSystemManagementConfigStrict(userID, key, payload)
	if err == nil {
		return saved
	}
	return cloneObjectMap(payload)
}

func (s *Service) SaveSystemManagementConfigStrict(userID int64, key string, payload map[string]interface{}) (map[string]interface{}, error) {
	key = strings.TrimSpace(key)
	if userID <= 0 || key == "" {
		return nil, ErrInvalidSystemManagementConfig
	}
	if s.repo != nil {
		saved, err := s.repo.SaveSystemManagementConfig(context.Background(), userID, key, payload)
		if err != nil {
			return nil, err
		}
		return cloneObjectMap(saved), nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.system[userID] == nil {
		s.system[userID] = make(map[string]interface{})
	}
	s.system[userID][key] = cloneObjectMap(payload)
	return cloneObjectMap(s.system[userID][key].(map[string]interface{})), nil
}

func (s *Service) SystemManagementConfigs(key string) []SystemManagementConfigItem {
	items, _ := s.SystemManagementConfigsStrict(key)
	return items
}

func (s *Service) SystemManagementConfigsStrict(key string) ([]SystemManagementConfigItem, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrInvalidSystemManagementConfig
	}
	if s.repo != nil {
		items, err := s.repo.ListSystemManagementConfigs(context.Background(), key)
		if err != nil {
			return nil, err
		}
		return cloneSystemManagementConfigItems(items), nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]SystemManagementConfigItem, 0)
	for userID, byKey := range s.system {
		if value, ok := byKey[key].(map[string]interface{}); ok {
			items = append(items, SystemManagementConfigItem{UserID: userID, Key: key, Value: cloneObjectMap(value)})
		}
	}
	return items, nil
}

func (s *Service) UpdateGuideResource(userID int64, req GuideResourceRequest) (GuideResourceProfile, error) {
	req.ResourceTags = normalizeTags(req.ResourceTags, 20, 40)
	req.IndustryTags = normalizeTags(req.IndustryTags, 20, 40)
	req.CityCodes = normalizeTags(req.CityCodes, 20, 32)
	req.ConnectionScale = strings.TrimSpace(req.ConnectionScale)
	if (len(req.ResourceTags) == 0 && len(req.IndustryTags) == 0 && len(req.CityCodes) == 0) || len(req.ConnectionScale) > 32 {
		return GuideResourceProfile{}, ErrInvalidProfile
	}
	if s.repo != nil {
		ok, err := s.repo.HasRole(context.Background(), userID, "guide")
		if err != nil {
			return GuideResourceProfile{}, err
		}
		if !ok {
			return GuideResourceProfile{}, ErrGuideForbidden
		}
		return s.repo.SaveGuideResource(context.Background(), guideProfileFromRequest(userID, req))
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.guides[userID] {
		return GuideResourceProfile{}, ErrGuideForbidden
	}
	profile := guideProfileFromRequest(userID, req)
	s.resources[userID] = profile
	return profile, nil
}

func cloneObjectMap(value map[string]interface{}) map[string]interface{} {
	if value == nil {
		return map[string]interface{}{}
	}
	data, err := json.Marshal(value)
	if err != nil {
		return map[string]interface{}{}
	}
	var cloned map[string]interface{}
	if err := json.Unmarshal(data, &cloned); err != nil {
		return map[string]interface{}{}
	}
	return cloned
}

func cloneSystemManagementConfigItems(items []SystemManagementConfigItem) []SystemManagementConfigItem {
	cloned := make([]SystemManagementConfigItem, 0, len(items))
	for _, item := range items {
		cloned = append(cloned, SystemManagementConfigItem{
			UserID: item.UserID,
			Key:    item.Key,
			Value:  cloneObjectMap(item.Value),
		})
	}
	return cloned
}

func normalizeTags(values []string, maxCount int, maxLen int) []string {
	if len(values) > maxCount {
		return nil
	}
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{})
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || len(value) > maxLen {
			return nil
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func validFileIDs(fileIDs []int64, maxCount int) bool {
	if len(fileIDs) > maxCount {
		return false
	}
	for _, fileID := range fileIDs {
		if fileID <= 0 {
			return false
		}
	}
	return true
}

func completeness(parts ...bool) int {
	if len(parts) == 0 {
		return 0
	}
	done := 0
	for _, ok := range parts {
		if ok {
			done++
		}
	}
	return done * 100 / len(parts)
}

func expertProfileFromRequest(userID int64, req ExpertSkillRequest) ExpertSkillProfile {
	return ExpertSkillProfile{
		UserID:       userID,
		SkillTree:    append([]string(nil), req.SkillTree...),
		ServiceTags:  append([]string(nil), req.ServiceTags...),
		CaseFileIDs:  append([]int64(nil), req.CaseFileIDs...),
		Completeness: completeness(len(req.SkillTree) > 0, len(req.ServiceTags) > 0, len(req.CaseFileIDs) > 0),
		UpdatedAt:    time.Now(),
	}
}

func guideProfileFromRequest(userID int64, req GuideResourceRequest) GuideResourceProfile {
	return GuideResourceProfile{
		UserID:          userID,
		ResourceTags:    append([]string(nil), req.ResourceTags...),
		IndustryTags:    append([]string(nil), req.IndustryTags...),
		CityCodes:       append([]string(nil), req.CityCodes...),
		ConnectionScale: req.ConnectionScale,
		Completeness:    completeness(len(req.ResourceTags) > 0, len(req.IndustryTags) > 0, len(req.CityCodes) > 0, req.ConnectionScale != ""),
		UpdatedAt:       time.Now(),
	}
}

func validRoleCode(roleCode string) bool {
	return roleCode == "expert" || roleCode == "guide"
}

func hasPendingRoleApplication(items []RoleApplication, roleCode string) bool {
	for _, item := range items {
		if item.RoleCode == roleCode && item.Status == "pending" {
			return true
		}
	}
	return false
}

func (s *Service) roleApplicationsByUserLocked(userID int64) []RoleApplication {
	items := make([]RoleApplication, 0)
	for _, app := range s.apps {
		if app.UserID == userID {
			items = append(items, app)
		}
	}
	return items
}

func (s *Service) reviewRoleApplication(app RoleApplication, adminID int64, req ReviewRoleApplicationRequest) (RoleApplication, error) {
	if app.Status != "pending" {
		return RoleApplication{}, ErrRoleApplicationReviewed
	}
	app.ReviewAdminID = adminID
	app.ReviewRemark = req.Remark
	app.UpdatedAt = time.Now()
	if req.Approve {
		app.Status = "approved"
		if app.CertificateNo == "" {
			app.CertificateNo = fmt.Sprintf("ZHW-%05d-%d", app.ID, app.UpdatedAt.Year())
		}
		app.CertifiedAt = app.UpdatedAt.Format(time.RFC3339)
		return app, nil
	}
	app.Status = "rejected"
	app.RejectReason = req.Remark
	return app, nil
}

func guideOpenStatus(conditionMet bool, _ bool, isGuide bool) string {
	switch {
	case !conditionMet:
		return "waiting_condition"
	case isGuide:
		return "opened"
	default:
		return "ready_for_review"
	}
}

func (s *Service) ensureDefaultGuideRuleLocked() {
	if len(s.rules) > 0 {
		return
	}
	s.rules[1] = GuideQualificationRule{
		ID:                1,
		RuleCode:          "default",
		MinInviteCount:    0,
		MinCreditScore:    0,
		MinCompletedGames: 0,
		PaymentRequired:   false,
		Status:            "active",
		UpdatedAt:         time.Now(),
	}
	if s.nextRuleID <= 1 {
		s.nextRuleID = 2
	}
}

func applyGuideRuleUpdate(rule GuideQualificationRule, req UpdateGuideQualificationRuleRequest, status string) GuideQualificationRule {
	if req.MinInviteCount != nil {
		rule.MinInviteCount = *req.MinInviteCount
	}
	if req.MinCreditScore != nil {
		rule.MinCreditScore = *req.MinCreditScore
	}
	if req.MinCompletedGames != nil {
		rule.MinCompletedGames = *req.MinCompletedGames
	}
	if req.PaymentRequired != nil {
		// 一期关闭会员购买条件，保留字段仅用于历史数据兼容。
		rule.PaymentRequired = false
	}
	if status != "" {
		rule.Status = status
	}
	rule.UpdatedAt = time.Now()
	return rule
}
