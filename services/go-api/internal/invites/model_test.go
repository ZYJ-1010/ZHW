package invites

import (
	"context"
	"testing"
	"time"
)

func TestRepositoryNotFoundDoesNotUseStaleInviteCache(t *testing.T) {
	store := NewStoreWithRepository(authoritativeInviteRepository{})
	store.codes["STALE001"] = InviteCode{ID: 1, Code: "STALE001", Status: StatusActive}
	store.boundCode[1] = 88
	store.relations[88] = Relation{InviteCodeID: 1, InviterUserID: 9, InviteeUserID: 88}

	if _, found, err := store.FindCode("STALE001"); err != nil || found {
		t.Fatalf("repository not-found must not return stale invite, found=%v err=%v", found, err)
	}
	if _, found, err := store.FindBoundCode(1); err != nil || found {
		t.Fatalf("repository not-found must not return stale binding, found=%v err=%v", found, err)
	}
	if _, found, err := store.RelationForUser(88); err != nil || found {
		t.Fatalf("repository not-found must not return stale relation, found=%v err=%v", found, err)
	}
}

type authoritativeInviteRepository struct{}

func (authoritativeInviteRepository) UpsertCode(context.Context, InviteCode) (InviteCode, error) {
	return InviteCode{}, nil
}
func (authoritativeInviteRepository) CreateCodes(_ context.Context, items []InviteCode) ([]InviteCode, error) {
	return items, nil
}
func (authoritativeInviteRepository) FindCode(context.Context, string) (InviteCode, bool, error) {
	return InviteCode{}, false, nil
}
func (authoritativeInviteRepository) Bind(context.Context, InviteCode, int64, string) (Relation, error) {
	return Relation{}, nil
}
func (authoritativeInviteRepository) SetRelationInviter(context.Context, int64, int64, int64, string) (Relation, error) {
	return Relation{}, nil
}
func (authoritativeInviteRepository) RelationForUser(context.Context, int64) (Relation, bool, error) {
	return Relation{}, false, nil
}
func (authoritativeInviteRepository) FindBoundCode(context.Context, int64) (InviteCode, bool, error) {
	return InviteCode{}, false, nil
}
func (authoritativeInviteRepository) ListCodes(context.Context, CodeFilter) ([]InviteCode, error) {
	return nil, nil
}
func (authoritativeInviteRepository) ListRelations(context.Context, RelationFilter) ([]Relation, error) {
	return nil, nil
}
func (authoritativeInviteRepository) CreateQuotaRequest(context.Context, QuotaRequest) (QuotaRequest, error) {
	return QuotaRequest{}, nil
}
func (authoritativeInviteRepository) ListQuotaRequests(context.Context, int64, string) ([]QuotaRequest, error) {
	return nil, nil
}
func (authoritativeInviteRepository) ReviewQuotaRequest(context.Context, int64, string, string, int64) (QuotaRequest, error) {
	return QuotaRequest{}, nil
}
func (authoritativeInviteRepository) ClearInviteeBindings(context.Context, int64) error { return nil }

func TestUnusedInviteStatusTransitionsAndUsedLock(t *testing.T) {
	store := NewStore()
	invite := store.UpsertCodeWithEntryType("ADMIN001", 1001, 1, EntryTypeLink)

	updated, err := store.UpdateUnusedCode(invite.Code, 1002, EntryTypeQRCode, time.Now().Add(time.Hour))
	if err != nil || updated.OwnerID != 1002 || updated.EntryType != EntryTypeQRCode {
		t.Fatalf("expected unused invite to update, invite=%+v err=%v", updated, err)
	}
	if _, err := store.DisableCode(invite.Code); err != nil {
		t.Fatalf("expected unused invite to disable: %v", err)
	}
	if _, err := store.EnableCode(invite.Code); err != nil {
		t.Fatalf("expected disabled invite to enable: %v", err)
	}
	if _, err := store.Bind(updated, 2001, "invite_qrcode"); err != nil {
		t.Fatalf("expected invite bind: %v", err)
	}
	if _, err := store.UpdateUnusedCode(invite.Code, 1003, EntryTypeLink, time.Time{}); err == nil {
		t.Fatal("used invite must not be editable")
	}
	if _, err := store.DisableCode(invite.Code); err == nil {
		t.Fatal("used invite must not be disabled")
	}
	if _, err := store.VoidCode(invite.Code); err == nil {
		t.Fatal("used invite must not be voided")
	}
}

func TestInviteVoidAndQuotaReview(t *testing.T) {
	store := NewStore()
	invite := store.UpsertCodeWithEntryType("ADMIN002", 1001, 1, EntryTypeLink)
	voided, err := store.VoidCode(invite.Code)
	if err != nil || voided.Status != StatusVoided {
		t.Fatalf("expected unused invite to be voided, invite=%+v err=%v", voided, err)
	}
	if _, err := store.EnableCode(invite.Code); err == nil {
		t.Fatal("voided invite must not be re-enabled")
	}

	request, err := store.CreateQuotaRequest(1001, 3, "邀请活动")
	if err != nil || request.Status != "pending" {
		t.Fatalf("expected pending quota request: %+v err=%v", request, err)
	}
	reviewed, err := store.ReviewQuotaRequest(request.ID, "approved", "通过", 1)
	if err != nil || reviewed.Status != "approved" {
		t.Fatalf("expected approved quota request: %+v err=%v", reviewed, err)
	}
	if _, err := store.ReviewQuotaRequest(request.ID, "rejected", "重复审核", 1); err == nil {
		t.Fatal("reviewed quota request must not be reviewed twice")
	}
}

func TestQuotaApprovalCreatesAllCodesBeforeChangingStatus(t *testing.T) {
	store := NewStore()
	request, err := store.CreateQuotaRequest(1001, 3, "活动加量")
	if err != nil {
		t.Fatal(err)
	}
	items, err := store.PrepareEntryCodes(1001, request.Quantity, EntryTypeQRCode, time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	approved, created, err := store.ApproveQuotaRequestWithCodes(request.ID, "通过", 9, items)
	if err != nil {
		t.Fatal(err)
	}
	if approved.Status != "approved" || len(created) != request.Quantity {
		t.Fatalf("unexpected atomic quota approval: request=%+v items=%+v", approved, created)
	}
	for _, item := range created {
		if item.OwnerID != request.OwnerUserID || item.MaxUses != 1 || item.EntryType != EntryTypeQRCode || item.ID == 0 {
			t.Fatalf("unexpected generated invite: %+v", item)
		}
	}
	if _, _, err := store.ApproveQuotaRequestWithCodes(request.ID, "重复", 9, items); err == nil {
		t.Fatal("reviewed quota request must not create another batch")
	}
}

func TestQuotaApprovalRejectsQuantityMismatchWithoutChangingRequest(t *testing.T) {
	store := NewStore()
	request, err := store.CreateQuotaRequest(1001, 2, "活动加量")
	if err != nil {
		t.Fatal(err)
	}
	items, err := store.PrepareEntryCodes(1001, 1, EntryTypeLink, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.ApproveQuotaRequestWithCodes(request.ID, "通过", 9, items); err == nil {
		t.Fatal("quantity mismatch must fail")
	}
	requests, err := store.ListQuotaRequests(1001, "pending")
	if err != nil || len(requests) != 1 || requests[0].ID != request.ID {
		t.Fatalf("failed approval must leave request pending, items=%+v err=%v", requests, err)
	}
}

func TestInviteListDerivesUsageAndExpiryStatuses(t *testing.T) {
	store := NewStore()
	expired := store.UpsertCodeWithEntryType("EXPIRED001", 1001, 1, EntryTypeLink)
	expired.ExpiresAt = time.Now().Add(-time.Minute)
	store.codes[expired.Code] = expired
	used := store.UpsertCodeWithEntryType("USED001", 1001, 1, EntryTypeLink)
	if _, err := store.Bind(used, 2001, "invite_link"); err != nil {
		t.Fatalf("bind used invite: %v", err)
	}

	expiredItems, err := store.ListCodes(CodeFilter{Status: StatusExpired})
	if err != nil || len(expiredItems) != 1 || expiredItems[0].Code != expired.Code || expiredItems[0].DisplayStatus != StatusExpired || expiredItems[0].UseStatus != "unused" {
		t.Fatalf("expected expired filter and display status, items=%+v err=%v", expiredItems, err)
	}
	usedItems, err := store.ListCodes(CodeFilter{Status: "used"})
	if err != nil || len(usedItems) != 1 || usedItems[0].Code != used.Code || usedItems[0].UseStatus != "used" {
		t.Fatalf("expected used filter and usage status, items=%+v err=%v", usedItems, err)
	}
	if _, err := store.UpdateUnusedCode(expired.Code, 1001, EntryTypeLink, time.Now().Add(time.Hour)); err == nil {
		t.Fatal("expired invite must not be reactivated by editing")
	}
}

func TestCreateCodesIsAtomicAndDoesNotOverwriteExistingCodes(t *testing.T) {
	store := NewStore()
	existing := store.UpsertCodeWithEntryType("ABC000001", 1001, 1, EntryTypeLink)
	items, err := store.CreateCodes([]InviteCode{
		{Code: "ABC000002", OwnerID: 1001, MaxUses: 1, EntryType: EntryTypeQRCode},
		{Code: "ABC000003", OwnerID: 1001, MaxUses: 1, EntryType: EntryTypeQRCode},
	})
	if err != nil || len(items) != 2 {
		t.Fatalf("expected two new invite codes, items=%+v err=%v", items, err)
	}
	if _, err := store.CreateCodes([]InviteCode{
		{Code: "ABC000004", OwnerID: 1002, MaxUses: 1, EntryType: EntryTypeLink},
		{Code: existing.Code, OwnerID: 1002, MaxUses: 1, EntryType: EntryTypeLink},
	}); err != ErrInviteCodeExists {
		t.Fatalf("expected duplicate code conflict, got %v", err)
	}
	if _, found, err := store.FindCode("ABC000004"); err != nil || found {
		t.Fatalf("conflicting batch must not create a partial code, found=%v err=%v", found, err)
	}
	current, found, err := store.FindCode(existing.Code)
	if err != nil || !found || current.OwnerID != 1001 || current.EntryType != EntryTypeLink {
		t.Fatalf("existing invite must remain unchanged, item=%+v found=%v err=%v", current, found, err)
	}
}

func TestSetRelationInviterReplacesInMemoryRelationWithoutDuplicate(t *testing.T) {
	store := NewStore()
	original := store.UpsertCodeWithEntryType("OLD000001", 1001, 1, EntryTypeLink)
	if _, err := store.Bind(original, 2001, "invite_link"); err != nil {
		t.Fatalf("bind original invite: %v", err)
	}

	updated, err := store.SetRelationInviter(2001, 1002, "admin_reassign")
	if err != nil {
		t.Fatalf("reassign inviter: %v", err)
	}
	if updated.InviterUserID != 1002 || updated.InviteCodeID == original.ID {
		t.Fatalf("expected relation to move to the new inviter code, got %+v", updated)
	}
	relations, err := store.ListRelations(RelationFilter{InviteeUserID: 2001})
	if err != nil {
		t.Fatalf("list relations: %v", err)
	}
	if len(relations) != 1 || relations[0].InviterUserID != 1002 {
		t.Fatalf("expected exactly one current relation, got %+v", relations)
	}
	if _, found, err := store.FindBoundCode(original.ID); err != nil || found {
		t.Fatalf("old invite must no longer point at the reassigned user, found=%v err=%v", found, err)
	}
}
