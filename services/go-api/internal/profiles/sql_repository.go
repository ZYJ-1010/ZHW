package profiles

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) GrantRole(ctx context.Context, userID int64, roleCode string) error {
	_, err := r.db.ExecContext(ctx, `
insert into user_roles (user_id, role_code, status, created_at)
values ($1,$2,'active',now())
on conflict (user_id, role_code) do update set status = 'active'
`, userID, roleCode)
	return err
}

func (r *SQLRepository) SaveEnterpriseCertification(ctx context.Context, item EnterpriseCertification) (EnterpriseCertification, error) {
	return scanEnterpriseCertification(r.db.QueryRowContext(ctx, `
insert into enterprise_certifications (user_id, company_name, unified_social_credit_code, legal_person, business_license_file_id, public_account_file_id, status, reject_reason, review_admin_id, review_remark, created_at, updated_at)
values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
returning id, user_id, company_name, unified_social_credit_code, legal_person, business_license_file_id, public_account_file_id, status, reject_reason, review_admin_id, review_remark, created_at, updated_at
`, item.UserID, item.CompanyName, item.UnifiedSocialCreditCode, item.LegalPerson, item.BusinessLicenseFileID, item.PublicAccountFileID, item.Status, nullString(item.RejectReason), nullInt64(item.ReviewAdminID), nullString(item.ReviewRemark), item.CreatedAt, item.UpdatedAt))
}

func (r *SQLRepository) FindEnterpriseCertification(ctx context.Context, userID int64) (EnterpriseCertification, bool, error) {
	item, err := scanEnterpriseCertification(r.db.QueryRowContext(ctx, enterpriseCertificationSelect()+` where user_id = $1 order by id desc limit 1`, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return EnterpriseCertification{}, false, nil
	}
	if err != nil {
		return EnterpriseCertification{}, false, err
	}
	return item, true, nil
}

func (r *SQLRepository) ListEnterpriseCertifications(ctx context.Context, status string) ([]EnterpriseCertification, error) {
	query := enterpriseCertificationSelect()
	args := []any{}
	if status != "" {
		query += " where status = $1"
		args = append(args, status)
	}
	query += " order by updated_at desc, id desc"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]EnterpriseCertification, 0)
	for rows.Next() {
		item, err := scanEnterpriseCertification(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) ReviewEnterpriseCertification(ctx context.Context, item EnterpriseCertification) (EnterpriseCertification, error) {
	return scanEnterpriseCertification(r.db.QueryRowContext(ctx, `
update enterprise_certifications
set status = $2, reject_reason = $3, review_admin_id = $4, review_remark = $5, updated_at = $6
where user_id = $1 and status = 'pending'
returning id, user_id, company_name, unified_social_credit_code, legal_person, business_license_file_id, public_account_file_id, status, reject_reason, review_admin_id, review_remark, created_at, updated_at
`, item.UserID, item.Status, nullString(item.RejectReason), nullInt64(item.ReviewAdminID), nullString(item.ReviewRemark), item.UpdatedAt))
}

func enterpriseCertificationSelect() string {
	return `select id, user_id, company_name, unified_social_credit_code, legal_person, business_license_file_id, public_account_file_id, status, reject_reason, review_admin_id, review_remark, created_at, updated_at from enterprise_certifications`
}

func scanEnterpriseCertification(row interface{ Scan(dest ...any) error }) (EnterpriseCertification, error) {
	var item EnterpriseCertification
	var rejectReason, reviewRemark sql.NullString
	var reviewAdminID sql.NullInt64
	if err := row.Scan(&item.ID, &item.UserID, &item.CompanyName, &item.UnifiedSocialCreditCode, &item.LegalPerson, &item.BusinessLicenseFileID, &item.PublicAccountFileID, &item.Status, &rejectReason, &reviewAdminID, &reviewRemark, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return EnterpriseCertification{}, err
	}
	item.RejectReason = rejectReason.String
	item.ReviewRemark = reviewRemark.String
	item.ReviewAdminID = reviewAdminID.Int64
	return item, nil
}

func (r *SQLRepository) HasRole(ctx context.Context, userID int64, roleCode string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
select exists(select 1 from user_roles where user_id = $1 and role_code = $2 and status = 'active')
`, userID, roleCode).Scan(&exists)
	return exists, err
}

func (r *SQLRepository) SaveRoleApplication(ctx context.Context, app RoleApplication) (RoleApplication, error) {
	proofFileIDs, _ := json.Marshal(app.ProofFileIDs)
	eligibilitySnapshot, _ := json.Marshal(app.EligibilitySnapshot)
	return scanRoleApplication(r.db.QueryRowContext(ctx, `
insert into role_applications (user_id, role_code, status, reason, ability_description, proof_file_ids, eligibility_snapshot, reject_reason, review_admin_id, review_remark, certificate_no, certified_at, created_at, updated_at)
values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
returning `+roleApplicationSelect()+`
`, app.UserID, app.RoleCode, app.Status, app.Reason, nullString(app.AbilityDescription), string(proofFileIDs), string(eligibilitySnapshot), nullString(app.RejectReason), nullInt64(app.ReviewAdminID), nullString(app.ReviewRemark), nullString(app.CertificateNo), nullTimeString(app.CertifiedAt), app.CreatedAt, app.UpdatedAt))
}

func roleApplicationSelect() string {
	return `id, user_id, role_code, status, reason, ability_description, proof_file_ids, eligibility_snapshot, reject_reason, review_admin_id, review_remark, certificate_no, certified_at, created_at, updated_at`
}

func (r *SQLRepository) ListRoleApplications(ctx context.Context) ([]RoleApplication, error) {
	rows, err := r.db.QueryContext(ctx, `
	select `+roleApplicationSelect()+`
from role_applications
order by id desc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRoleApplications(rows)
}

func (r *SQLRepository) ListRoleApplicationsByUser(ctx context.Context, userID int64) ([]RoleApplication, error) {
	rows, err := r.db.QueryContext(ctx, `
	select `+roleApplicationSelect()+`
from role_applications
where user_id = $1
order by id desc
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRoleApplications(rows)
}

func (r *SQLRepository) FindRoleApplication(ctx context.Context, applicationID int64) (RoleApplication, bool, error) {
	app, err := scanRoleApplication(r.db.QueryRowContext(ctx, `
	select `+roleApplicationSelect()+`
from role_applications
where id = $1
`, applicationID))
	if errors.Is(err, sql.ErrNoRows) {
		return RoleApplication{}, false, nil
	}
	if err != nil {
		return RoleApplication{}, false, err
	}
	return app, true, nil
}

func (r *SQLRepository) UpdateRoleApplication(ctx context.Context, app RoleApplication) (RoleApplication, error) {
	return scanRoleApplication(r.db.QueryRowContext(ctx, `
update role_applications
set status = $2, reject_reason = $3, review_admin_id = $4, review_remark = $5, certificate_no = $6, certified_at = $7, updated_at = $8
where id = $1
returning `+roleApplicationSelect()+`
`, app.ID, app.Status, nullString(app.RejectReason), nullInt64(app.ReviewAdminID), nullString(app.ReviewRemark), nullString(app.CertificateNo), nullTimeString(app.CertifiedAt), app.UpdatedAt))
}

func (r *SQLRepository) SaveReviewedRoleApplication(ctx context.Context, app RoleApplication) (RoleApplication, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return RoleApplication{}, err
	}
	defer tx.Rollback()

	saved, err := scanRoleApplication(tx.QueryRowContext(ctx, `
update role_applications
set status = $2, reject_reason = $3, review_admin_id = $4, review_remark = $5, certificate_no = $6, certified_at = $7, updated_at = $8
where id = $1
returning `+roleApplicationSelect()+`
`, app.ID, app.Status, nullString(app.RejectReason), nullInt64(app.ReviewAdminID), nullString(app.ReviewRemark), nullString(app.CertificateNo), nullTimeString(app.CertifiedAt), app.UpdatedAt))
	if err != nil {
		return RoleApplication{}, err
	}

	if saved.Status == "approved" {
		if _, err := tx.ExecContext(ctx, `
insert into user_roles (user_id, role_code, status, created_at)
values ($1,$2,'active',now())
on conflict (user_id, role_code) do update set status = 'active'
`, saved.UserID, saved.RoleCode); err != nil {
			return RoleApplication{}, err
		}
		if saved.RoleCode == "guide" {
			if _, err := tx.ExecContext(ctx, `
insert into guide_qualification_records (user_id, condition_met, payment_met, guide_open_status, updated_at)
values ($1,true,true,'opened',now())
on conflict (user_id) do update set
  condition_met = true,
  payment_met = true,
  guide_open_status = 'opened',
  updated_at = now()
`, saved.UserID); err != nil {
				return RoleApplication{}, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return RoleApplication{}, err
	}
	return saved, nil
}

func (r *SQLRepository) GetGuideQualification(ctx context.Context, userID int64) (GuideQualification, bool, error) {
	q, err := scanGuideQualification(r.db.QueryRowContext(ctx, `
select user_id, condition_met, payment_met, guide_open_status, updated_at
from guide_qualification_records
where user_id = $1
`, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return GuideQualification{}, false, nil
	}
	if err != nil {
		return GuideQualification{}, false, err
	}
	return q, true, nil
}

func (r *SQLRepository) SaveGuideQualification(ctx context.Context, q GuideQualification) (GuideQualification, error) {
	return scanGuideQualification(r.db.QueryRowContext(ctx, `
insert into guide_qualification_records (user_id, condition_met, payment_met, guide_open_status, updated_at)
values ($1,$2,$3,$4,$5)
on conflict (user_id) do update set
  condition_met = excluded.condition_met,
  payment_met = excluded.payment_met,
  guide_open_status = excluded.guide_open_status,
  updated_at = excluded.updated_at
returning user_id, condition_met, payment_met, guide_open_status, updated_at
`, q.UserID, q.ConditionMet, q.PaymentMet, q.GuideOpenStatus, q.UpdatedAt))
}

func (r *SQLRepository) ListGuideQualificationRules(ctx context.Context) ([]GuideQualificationRule, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, rule_code, min_invite_count, min_credit_score, min_completed_games, payment_required, status, updated_at
from guide_qualification_rules
order by id asc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]GuideQualificationRule, 0)
	for rows.Next() {
		item, err := scanGuideQualificationRule(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) > 0 {
		return items, nil
	}
	defaultRule := GuideQualificationRule{
		RuleCode:          "default",
		MinInviteCount:    0,
		MinCreditScore:    0,
		MinCompletedGames: 0,
		PaymentRequired:   false,
		Status:            "active",
		UpdatedAt:         time.Now(),
	}
	rule, err := r.SaveGuideQualificationRule(ctx, defaultRule)
	if err != nil {
		return nil, err
	}
	return []GuideQualificationRule{rule}, nil
}

func (r *SQLRepository) SaveGuideQualificationRule(ctx context.Context, rule GuideQualificationRule) (GuideQualificationRule, error) {
	if rule.ID > 0 {
		return scanGuideQualificationRule(r.db.QueryRowContext(ctx, `
update guide_qualification_rules
set min_invite_count = $2,
  min_credit_score = $3,
  min_completed_games = $4,
  payment_required = $5,
  status = $6,
  updated_at = $7
where id = $1
returning id, rule_code, min_invite_count, min_credit_score, min_completed_games, payment_required, status, updated_at
`, rule.ID, rule.MinInviteCount, rule.MinCreditScore, rule.MinCompletedGames, rule.PaymentRequired, rule.Status, rule.UpdatedAt))
	}
	return scanGuideQualificationRule(r.db.QueryRowContext(ctx, `
insert into guide_qualification_rules (rule_code, min_invite_count, min_credit_score, min_completed_games, payment_required, status, updated_at)
values ($1,$2,$3,$4,$5,$6,$7)
on conflict (rule_code) do update set
  min_invite_count = excluded.min_invite_count,
  min_credit_score = excluded.min_credit_score,
  min_completed_games = excluded.min_completed_games,
  payment_required = excluded.payment_required,
  status = excluded.status,
  updated_at = excluded.updated_at
returning id, rule_code, min_invite_count, min_credit_score, min_completed_games, payment_required, status, updated_at
`, rule.RuleCode, rule.MinInviteCount, rule.MinCreditScore, rule.MinCompletedGames, rule.PaymentRequired, rule.Status, rule.UpdatedAt))
}

func (r *SQLRepository) GetExpertSkill(ctx context.Context, userID int64) (ExpertSkillProfile, bool, error) {
	profile, err := scanExpertSkill(r.db.QueryRowContext(ctx, `
select user_id, skill_tree, service_tags, case_file_ids, updated_at
from expert_skill_profiles
where user_id = $1
`, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return ExpertSkillProfile{}, false, nil
	}
	if err != nil {
		return ExpertSkillProfile{}, false, err
	}
	return profile, true, nil
}

func (r *SQLRepository) SaveExpertSkill(ctx context.Context, profile ExpertSkillProfile) (ExpertSkillProfile, error) {
	skillTree, _ := json.Marshal(profile.SkillTree)
	serviceTags, _ := json.Marshal(profile.ServiceTags)
	caseFileIDs, _ := json.Marshal(profile.CaseFileIDs)
	return scanExpertSkill(r.db.QueryRowContext(ctx, `
insert into expert_skill_profiles (user_id, skill_tree, service_tags, case_file_ids, updated_at)
values ($1,$2,$3,$4,now())
on conflict (user_id) do update set
  skill_tree = excluded.skill_tree,
  service_tags = excluded.service_tags,
  case_file_ids = excluded.case_file_ids,
  updated_at = now()
returning user_id, skill_tree, service_tags, case_file_ids, updated_at
`, profile.UserID, string(skillTree), string(serviceTags), string(caseFileIDs)))
}

func (r *SQLRepository) ListExpertSkills(ctx context.Context) ([]ExpertSkillProfile, error) {
	rows, err := r.db.QueryContext(ctx, `
select user_id, skill_tree, service_tags, case_file_ids, updated_at
from expert_skill_profiles
order by updated_at desc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ExpertSkillProfile, 0)
	for rows.Next() {
		item, err := scanExpertSkill(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) GetGuideResource(ctx context.Context, userID int64) (GuideResourceProfile, bool, error) {
	profile, err := scanGuideResource(r.db.QueryRowContext(ctx, `
select user_id, resource_tags, industry_tags, city_codes, connection_scale, updated_at
from guide_resource_profiles
where user_id = $1
`, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return GuideResourceProfile{}, false, nil
	}
	if err != nil {
		return GuideResourceProfile{}, false, err
	}
	return profile, true, nil
}

func (r *SQLRepository) SaveGuideResource(ctx context.Context, profile GuideResourceProfile) (GuideResourceProfile, error) {
	resourceTags, _ := json.Marshal(profile.ResourceTags)
	industryTags, _ := json.Marshal(profile.IndustryTags)
	cityCodes, _ := json.Marshal(profile.CityCodes)
	return scanGuideResource(r.db.QueryRowContext(ctx, `
insert into guide_resource_profiles (user_id, resource_tags, industry_tags, city_codes, connection_scale, updated_at)
values ($1,$2,$3,$4,$5,now())
on conflict (user_id) do update set
  resource_tags = excluded.resource_tags,
  industry_tags = excluded.industry_tags,
  city_codes = excluded.city_codes,
  connection_scale = excluded.connection_scale,
  updated_at = now()
returning user_id, resource_tags, industry_tags, city_codes, connection_scale, updated_at
`, profile.UserID, string(resourceTags), string(industryTags), string(cityCodes), nullString(profile.ConnectionScale)))
}

func (r *SQLRepository) ListGuideResources(ctx context.Context) ([]GuideResourceProfile, error) {
	rows, err := r.db.QueryContext(ctx, `
select user_id, resource_tags, industry_tags, city_codes, connection_scale, updated_at
from guide_resource_profiles
order by updated_at desc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]GuideResourceProfile, 0)
	for rows.Next() {
		item, err := scanGuideResource(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) GetSystemManagementConfig(ctx context.Context, userID int64, key string) (map[string]interface{}, bool, error) {
	var raw []byte
	err := r.db.QueryRowContext(ctx, `
select config_value
from user_system_management_configs
where user_id = $1 and config_key = $2
`, userID, key).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var value map[string]interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, false, err
	}
	return value, true, nil
}

func (r *SQLRepository) SaveSystemManagementConfig(ctx context.Context, userID int64, key string, payload map[string]interface{}) (map[string]interface{}, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var saved []byte
	err = r.db.QueryRowContext(ctx, `
insert into user_system_management_configs (user_id, config_key, config_value, updated_at)
values ($1,$2,$3,now())
on conflict (user_id, config_key) do update set
  config_value = excluded.config_value,
  updated_at = now()
returning config_value
`, userID, key, string(raw)).Scan(&saved)
	if err != nil {
		return nil, err
	}
	var value map[string]interface{}
	if err := json.Unmarshal(saved, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func (r *SQLRepository) ListSystemManagementConfigs(ctx context.Context, key string) ([]SystemManagementConfigItem, error) {
	rows, err := r.db.QueryContext(ctx, `
select user_id, config_key, config_value
from user_system_management_configs
where config_key = $1
order by updated_at desc
`, key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]SystemManagementConfigItem, 0)
	for rows.Next() {
		var item SystemManagementConfigItem
		var raw []byte
		if err := rows.Scan(&item.UserID, &item.Key, &raw); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &item.Value); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanExpertSkill(row interface {
	Scan(dest ...any) error
}) (ExpertSkillProfile, error) {
	var profile ExpertSkillProfile
	var skillTree []byte
	var serviceTags []byte
	var caseFileIDs []byte
	if err := row.Scan(&profile.UserID, &skillTree, &serviceTags, &caseFileIDs, &profile.UpdatedAt); err != nil {
		return ExpertSkillProfile{}, err
	}
	_ = json.Unmarshal(skillTree, &profile.SkillTree)
	_ = json.Unmarshal(serviceTags, &profile.ServiceTags)
	_ = json.Unmarshal(caseFileIDs, &profile.CaseFileIDs)
	profile.Completeness = completeness(len(profile.SkillTree) > 0, len(profile.ServiceTags) > 0, len(profile.CaseFileIDs) > 0)
	return profile, nil
}

func scanGuideResource(row interface {
	Scan(dest ...any) error
}) (GuideResourceProfile, error) {
	var profile GuideResourceProfile
	var resourceTags []byte
	var industryTags []byte
	var cityCodes []byte
	var connectionScale sql.NullString
	var updatedAt time.Time
	if err := row.Scan(&profile.UserID, &resourceTags, &industryTags, &cityCodes, &connectionScale, &updatedAt); err != nil {
		return GuideResourceProfile{}, err
	}
	_ = json.Unmarshal(resourceTags, &profile.ResourceTags)
	_ = json.Unmarshal(industryTags, &profile.IndustryTags)
	_ = json.Unmarshal(cityCodes, &profile.CityCodes)
	profile.ConnectionScale = connectionScale.String
	profile.Completeness = completeness(len(profile.ResourceTags) > 0, len(profile.IndustryTags) > 0, len(profile.CityCodes) > 0, profile.ConnectionScale != "")
	profile.UpdatedAt = updatedAt
	return profile, nil
}

func scanRoleApplication(row interface {
	Scan(dest ...any) error
}) (RoleApplication, error) {
	var app RoleApplication
	var reason sql.NullString
	var abilityDescription sql.NullString
	var proofFileIDs []byte
	var eligibilitySnapshot []byte
	var rejectReason sql.NullString
	var reviewAdminID sql.NullInt64
	var reviewRemark sql.NullString
	var certificateNo sql.NullString
	var certifiedAt sql.NullTime
	if err := row.Scan(&app.ID, &app.UserID, &app.RoleCode, &app.Status, &reason, &abilityDescription, &proofFileIDs, &eligibilitySnapshot, &rejectReason, &reviewAdminID, &reviewRemark, &certificateNo, &certifiedAt, &app.CreatedAt, &app.UpdatedAt); err != nil {
		return RoleApplication{}, err
	}
	app.Reason = reason.String
	app.AbilityDescription = abilityDescription.String
	_ = json.Unmarshal(proofFileIDs, &app.ProofFileIDs)
	_ = json.Unmarshal(eligibilitySnapshot, &app.EligibilitySnapshot)
	app.RejectReason = rejectReason.String
	if reviewAdminID.Valid {
		app.ReviewAdminID = reviewAdminID.Int64
	}
	app.ReviewRemark = reviewRemark.String
	app.CertificateNo = certificateNo.String
	if certifiedAt.Valid {
		app.CertifiedAt = certifiedAt.Time.Format(time.RFC3339)
	}
	return app, nil
}

func scanRoleApplications(rows *sql.Rows) ([]RoleApplication, error) {
	items := make([]RoleApplication, 0)
	for rows.Next() {
		item, err := scanRoleApplication(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanGuideQualification(row interface {
	Scan(dest ...any) error
}) (GuideQualification, error) {
	var q GuideQualification
	if err := row.Scan(&q.UserID, &q.ConditionMet, &q.PaymentMet, &q.GuideOpenStatus, &q.UpdatedAt); err != nil {
		return GuideQualification{}, err
	}
	return q, nil
}

func scanGuideQualificationRule(row interface {
	Scan(dest ...any) error
}) (GuideQualificationRule, error) {
	var rule GuideQualificationRule
	if err := row.Scan(&rule.ID, &rule.RuleCode, &rule.MinInviteCount, &rule.MinCreditScore, &rule.MinCompletedGames, &rule.PaymentRequired, &rule.Status, &rule.UpdatedAt); err != nil {
		return GuideQualificationRule{}, err
	}
	return rule, nil
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func nullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value > 0}
}

func nullTimeString(value string) sql.NullTime {
	parsed, err := time.Parse(time.RFC3339, value)
	return sql.NullTime{Time: parsed, Valid: value != "" && err == nil}
}
