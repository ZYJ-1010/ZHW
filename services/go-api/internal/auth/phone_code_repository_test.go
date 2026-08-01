package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"sync"
	"testing"
	"time"

	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

type sharedPhoneCodeState struct {
	phoneMasked      string
	scene            string
	codeHash         string
	sentAt           time.Time
	expiresAt        time.Time
	failedAttempts   int
	consumedAt       time.Time
	claimToken       string
	dayKey           string
	dayCount         int
	pendingToken     string
	pendingScene     string
	pendingCodeHash  string
	pendingExpiresAt time.Time
	pendingStartedAt time.Time
}

type sharedPhoneCodeRepository struct {
	mu              sync.Mutex
	states          map[string]sharedPhoneCodeState
	sessions        map[string]Session
	failNextSession bool
}

func newSharedPhoneCodeRepository() *sharedPhoneCodeRepository {
	return &sharedPhoneCodeRepository{states: make(map[string]sharedPhoneCodeState), sessions: make(map[string]Session)}
}

func (r *sharedPhoneCodeRepository) SaveSession(_ context.Context, session Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.failNextSession {
		r.failNextSession = false
		return errors.New("session repository unavailable")
	}
	r.sessions[hashToken(session.Token)] = session
	return nil
}

func (r *sharedPhoneCodeRepository) FindSessionByTokenHash(_ context.Context, tokenHash string) (Session, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	session, ok := r.sessions[tokenHash]
	return session, ok, nil
}

func (r *sharedPhoneCodeRepository) RevokeUserSessions(_ context.Context, userID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for tokenHash, session := range r.sessions {
		if session.UserID == userID {
			delete(r.sessions, tokenHash)
		}
	}
	return nil
}

func (r *sharedPhoneCodeRepository) RevokeSessionByTokenHash(_ context.Context, tokenHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, tokenHash)
	return nil
}

func (r *sharedPhoneCodeRepository) ReservePhoneCodeSend(_ context.Context, reservation PhoneCodeSendReservation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	state := r.states[reservation.PhoneKey]
	if state.pendingToken != "" && reservation.StartedAt.Before(state.pendingStartedAt.Add(reservation.ResendInterval)) {
		return ErrPhoneCodeRateLimit
	}
	if !state.sentAt.IsZero() && reservation.StartedAt.Before(state.sentAt.Add(reservation.ResendInterval)) {
		return ErrPhoneCodeRateLimit
	}
	dayCount := state.dayCount
	if state.dayKey != reservation.DayKey {
		dayCount = 0
	}
	if dayCount >= reservation.DailyLimit {
		return ErrPhoneCodeDailyLimit
	}
	state.phoneMasked = reservation.PhoneMasked
	state.pendingToken = reservation.ReservationToken
	state.pendingScene = reservation.Scene
	state.pendingCodeHash = reservation.CodeHash
	state.pendingExpiresAt = reservation.ExpiresAt
	state.pendingStartedAt = reservation.StartedAt
	r.states[reservation.PhoneKey] = state
	return nil
}

func (r *sharedPhoneCodeRepository) ConfirmPhoneCodeSend(_ context.Context, phoneKey string, reservationToken string, sentAt time.Time, dayKey string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	state := r.states[phoneKey]
	if state.pendingToken != reservationToken {
		return ErrPhoneCodeSendFailed
	}
	state.scene = state.pendingScene
	state.codeHash = state.pendingCodeHash
	state.sentAt = sentAt
	state.expiresAt = state.pendingExpiresAt
	state.failedAttempts = 0
	state.consumedAt = time.Time{}
	state.claimToken = ""
	if state.dayKey == dayKey {
		state.dayCount++
	} else {
		state.dayKey = dayKey
		state.dayCount = 1
	}
	state.pendingToken = ""
	state.pendingScene = ""
	state.pendingCodeHash = ""
	state.pendingExpiresAt = time.Time{}
	state.pendingStartedAt = time.Time{}
	r.states[phoneKey] = state
	return nil
}

func (r *sharedPhoneCodeRepository) CancelPhoneCodeSend(_ context.Context, phoneKey string, reservationToken string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	state := r.states[phoneKey]
	if state.pendingToken == reservationToken {
		state.pendingToken = ""
		state.pendingScene = ""
		state.pendingCodeHash = ""
		state.pendingExpiresAt = time.Time{}
		state.pendingStartedAt = time.Time{}
		r.states[phoneKey] = state
	}
	return nil
}

func (r *sharedPhoneCodeRepository) ClaimPhoneCode(_ context.Context, phoneKey string, candidateHashes map[string]string, allowedScenes map[string]bool, claimToken string, nowTime time.Time, maxFailures int, pendingTimeout time.Duration) (PhoneCodeClaim, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	state, ok := r.states[phoneKey]
	if !ok || state.pendingToken != "" && nowTime.Before(state.pendingStartedAt.Add(pendingTimeout)) || state.scene == "" || state.codeHash == "" || !state.consumedAt.IsZero() || !nowTime.Before(state.expiresAt) || state.failedAttempts >= maxFailures || !allowedScenes[state.scene] {
		return PhoneCodeClaim{}, false, nil
	}
	candidateHash, ok := candidateHashes[state.scene]
	if !ok || subtle.ConstantTimeCompare([]byte(state.codeHash), []byte(candidateHash)) != 1 {
		state.failedAttempts++
		if state.failedAttempts > maxFailures {
			state.failedAttempts = maxFailures
		}
		r.states[phoneKey] = state
		return PhoneCodeClaim{}, false, nil
	}
	state.consumedAt = nowTime
	state.claimToken = claimToken
	r.states[phoneKey] = state
	return PhoneCodeClaim{PhoneKey: phoneKey, Scene: state.scene, ClaimToken: claimToken, SentAt: state.sentAt, ExpiresAt: state.expiresAt}, true, nil
}

func (r *sharedPhoneCodeRepository) RestorePhoneCodeClaim(_ context.Context, phoneKey string, claimToken string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	state := r.states[phoneKey]
	if state.claimToken == claimToken {
		state.consumedAt = time.Time{}
		state.claimToken = ""
		r.states[phoneKey] = state
	}
	return nil
}

func newRepositoryPhoneCodeServices(repo *sharedPhoneCodeRepository) (*Service, *Service, *users.Store) {
	userStore := users.NewStore()
	inviteStore := invites.NewStore()
	first := NewService(userStore, inviteStore, NewTokenStoreWithRepository(repo))
	second := NewService(userStore, inviteStore, NewTokenStoreWithRepository(repo))
	first.UsePhoneSMSSender(fixedPhoneSMSSender{})
	second.UsePhoneSMSSender(fixedPhoneSMSSender{})
	return first, second, userStore
}

func createRepositoryPhoneUser(t *testing.T, service *Service, store *users.Store, phone string) users.User {
	t.Helper()
	user, err := store.CreateWithPhone(service.phoneLookupHash(phone), maskPhone(phone))
	if err != nil {
		t.Fatal(err)
	}
	return user
}

func TestPhoneCodeRepositoryVerifiesAcrossInstancesOnce(t *testing.T) {
	repo := newSharedPhoneCodeRepository()
	first, second, store := newRepositoryPhoneCodeServices(repo)
	phone := "13800138101"
	createRepositoryPhoneUser(t, first, store, phone)
	if _, err := first.SendPhoneCode(context.Background(), phone, "login"); err != nil {
		t.Fatal(err)
	}
	if _, err := second.PhoneLogin(PhoneLoginRequest{Phone: phone, Code: "654321"}); err != nil {
		t.Fatalf("another instance must consume the shared code: %v", err)
	}
	if _, err := first.PhoneLogin(PhoneLoginRequest{Phone: phone, Code: "654321"}); !errors.Is(err, ErrPhoneCodeInvalid) {
		t.Fatalf("shared code must be one-time, got %v", err)
	}
}

func TestPhoneCodeRepositorySharesSceneFailuresAndRateLimit(t *testing.T) {
	repo := newSharedPhoneCodeRepository()
	first, second, store := newRepositoryPhoneCodeServices(repo)
	phone := "13800138102"
	user := createRepositoryPhoneUser(t, first, store, phone)
	if _, err := first.SetPassword(user.ID, "OldPass123"); err != nil {
		t.Fatal(err)
	}
	if _, err := first.SendPhoneCode(context.Background(), phone, "password_reset"); err != nil {
		t.Fatal(err)
	}
	if _, err := second.SendPhoneCode(context.Background(), phone, "login"); !errors.Is(err, ErrPhoneCodeRateLimit) {
		t.Fatalf("send throttling must be shared, got %v", err)
	}
	if _, err := second.PhoneLogin(PhoneLoginRequest{Phone: phone, Code: "654321"}); !errors.Is(err, ErrPhoneCodeInvalid) {
		t.Fatalf("password-reset code must not log in, got %v", err)
	}
	if _, err := second.ResetPasswordByPhone(phone, "654321", "NewPass456"); err != nil {
		t.Fatalf("scene rejection must not consume the shared code: %v", err)
	}
}

func TestPhoneCodeRepositorySharesFailureLimit(t *testing.T) {
	repo := newSharedPhoneCodeRepository()
	first, second, store := newRepositoryPhoneCodeServices(repo)
	phone := "13800138103"
	createRepositoryPhoneUser(t, first, store, phone)
	if _, err := first.SendPhoneCode(context.Background(), phone, "login"); err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < phoneCodeMaxFailures; attempt++ {
		service := first
		if attempt%2 == 1 {
			service = second
		}
		if _, err := service.PhoneLogin(PhoneLoginRequest{Phone: phone, Code: "111111"}); !errors.Is(err, ErrPhoneCodeInvalid) {
			t.Fatalf("wrong attempt %d: %v", attempt+1, err)
		}
	}
	if _, err := second.PhoneLogin(PhoneLoginRequest{Phone: phone, Code: "654321"}); !errors.Is(err, ErrPhoneCodeInvalid) {
		t.Fatalf("correct code must remain locked across instances, got %v", err)
	}
}

func TestPhoneCodeRepositoryRestoresClaimAfterDownstreamFailure(t *testing.T) {
	repo := newSharedPhoneCodeRepository()
	first, second, store := newRepositoryPhoneCodeServices(repo)
	phone := "13800138104"
	createRepositoryPhoneUser(t, first, store, phone)
	if _, err := first.SendPhoneCode(context.Background(), phone, "login"); err != nil {
		t.Fatal(err)
	}
	repo.mu.Lock()
	repo.failNextSession = true
	repo.mu.Unlock()
	if _, err := second.PhoneLogin(PhoneLoginRequest{Phone: phone, Code: "654321"}); err == nil {
		t.Fatal("expected downstream session failure")
	}
	if _, err := first.PhoneLogin(PhoneLoginRequest{Phone: phone, Code: "654321"}); err != nil {
		t.Fatalf("claim must be restored for a retry on another instance: %v", err)
	}
}

func TestRollbackPhoneLoginRevokesSessionAndRestoresSharedCode(t *testing.T) {
	repo := newSharedPhoneCodeRepository()
	first, second, store := newRepositoryPhoneCodeServices(repo)
	phone := "13800138106"
	createRepositoryPhoneUser(t, first, store, phone)
	if _, err := first.SendPhoneCode(context.Background(), phone, "login"); err != nil {
		t.Fatal(err)
	}
	login, err := second.PhoneLogin(PhoneLoginRequest{Phone: phone, Code: "654321"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := second.CurrentUser(login.Token); !ok {
		t.Fatal("issued session must be available before rollback")
	}
	if err := second.RollbackPhoneLogin(login); err != nil {
		t.Fatal(err)
	}
	if _, ok := first.CurrentUser(login.Token); ok {
		t.Fatal("rollback must revoke the token in the shared session repository")
	}
	if _, err := first.PhoneLogin(PhoneLoginRequest{Phone: phone, Code: "654321"}); err != nil {
		t.Fatalf("rollback must restore the code for a retry: %v", err)
	}
}

func TestPhoneCodeRepositoryCancelsFailedSendReservation(t *testing.T) {
	repo := newSharedPhoneCodeRepository()
	first, second, _ := newRepositoryPhoneCodeServices(repo)
	phone := "13800138105"
	first.UsePhoneSMSSender(failingPhoneSMSSender{})
	if _, err := first.SendPhoneCode(context.Background(), phone, "login"); !errors.Is(err, ErrPhoneCodeSendFailed) {
		t.Fatalf("provider failure = %v", err)
	}
	second.UsePhoneSMSSender(fixedPhoneSMSSender{})
	if _, err := second.SendPhoneCode(context.Background(), phone, "login"); err != nil {
		t.Fatalf("failed reservation must not throttle another instance: %v", err)
	}
}

var _ identity.SMSSender = fixedPhoneSMSSender{}
