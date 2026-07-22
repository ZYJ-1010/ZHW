package invites

import (
	"testing"
	"time"
)

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
