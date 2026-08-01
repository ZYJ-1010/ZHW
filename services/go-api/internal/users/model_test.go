package users

import "testing"

func TestBindPhoneAuthRemovesPreviousPhoneLookup(t *testing.T) {
	store := NewStore()
	user, err := store.Create("phone-change-user")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.BindPhoneAuth(user.ID, "old-phone-hash", "138****0001"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.BindPhoneAuth(user.ID, "new-phone-hash", "139****0002"); err != nil {
		t.Fatal(err)
	}
	if old, found, err := store.FindByPhoneHash("old-phone-hash"); err != nil || found {
		t.Fatalf("old phone lookup must be removed, user=%+v found=%v err=%v", old, found, err)
	}
	if current, found, err := store.FindByPhoneHash("new-phone-hash"); err != nil || !found || current.ID != user.ID {
		t.Fatalf("new phone lookup must resolve account, user=%+v found=%v err=%v", current, found, err)
	}
}
