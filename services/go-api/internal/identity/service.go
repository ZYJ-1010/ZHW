package identity

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/mozillazg/go-pinyin"
)

var (
	ErrPhoneRequired     = errors.New("phone required")
	ErrCodeInvalid       = errors.New("sms code invalid")
	ErrPhoneNotVerified  = errors.New("phone not verified")
	ErrFaceIDNotStarted  = errors.New("faceid not started")
	ErrRealnameRequired  = errors.New("realname required")
	ErrPhoneInvalid      = errors.New("phone invalid")
	ErrIDCardInvalid     = errors.New("id card invalid")
	ErrSMSRateLimited    = errors.New("sms code send rate limited")
	ErrSMSDailyLimited   = errors.New("sms code send daily limited")
	ErrSMSSendFailed     = errors.New("sms code send failed")
	ErrFaceIDStartFailed = errors.New("faceid start failed")
	ErrRecordNotFound    = errors.New("identity record not found")
)

const (
	smsResendInterval = 60 * time.Second
	smsDailyLimit     = 5
	temporarySMSCode  = "000000"
)

type Status string

const (
	StatusWechatLoggedIn             Status = "wechat_logged_in"
	StatusPhoneBound                 Status = "phone_bound"
	StatusSMSVerified                Status = "sms_verified"
	StatusPhoneVerified              Status = "phone_verified"
	StatusFaceIDProcessing           Status = "faceid_processing"
	StatusFaceVerified               Status = "face_verified"
	StatusWechatRealnameNotSupported Status = "wechat_realname_not_supported"
	StatusPendingManualReview        Status = "pending"
	StatusRejected                   Status = "rejected"
	StatusVerified                   Status = "verified"
)

var idCard18Pattern = regexp.MustCompile(`^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$`)

type Record struct {
	UserID                    int64  `json:"userId"`
	Status                    Status `json:"status"`
	PhoneMasked               string `json:"phoneMasked,omitempty"`
	RealNameMasked            string `json:"realNameMasked,omitempty"`
	RealNameCiphertext        string `json:"-"`
	RealNameInitials          string `json:"-"`
	IDCardMasked              string `json:"idCardMasked,omitempty"`
	IDCardCiphertext          string `json:"-"`
	SMSVerified               bool   `json:"smsVerified"`
	PhoneVerified             bool   `json:"phoneVerified"`
	FaceVerified              bool   `json:"faceVerified"`
	WechatRealnameConsistency string `json:"wechatRealnameConsistency"`
	FailureReason             string `json:"failureReason,omitempty"`
	UpdatedAt                 string `json:"updatedAt"`
}

type PlainIdentity struct {
	RealName string
	IDCard   string
}

type InGameIdentity struct {
	RealName    string `json:"realName"`
	DisplayName string `json:"displayName"`
	AvatarText  string `json:"avatarText"`
}

type Service struct {
	mu            sync.RWMutex
	records       map[int64]Record
	smsCodes      map[int64]smsCodeState
	faceTokens    map[int64]string
	repo          Repository
	smsSender     SMSSender
	faceIDStarter FaceIDStarter
	dataKey       [32]byte
}

type smsCodeState struct {
	code     string
	sentAt   time.Time
	dayKey   string
	dayCount int
}

func NewService() *Service {
	return NewServiceWithRepository(nil)
}

func NewServiceWithRepository(repo Repository) *Service {
	service := &Service{
		records:       make(map[int64]Record),
		smsCodes:      make(map[int64]smsCodeState),
		faceTokens:    make(map[int64]string),
		repo:          repo,
		smsSender:     LocalSMSSender{},
		faceIDStarter: LocalFaceIDStarter{},
	}
	service.UseDataEncryptionKey("local-development-only")
	return service
}

func (s *Service) UseDataEncryptionKey(secret string) {
	s.dataKey = sha256.Sum256([]byte("zhw-in-game-identity-v1\x00" + strings.TrimSpace(secret)))
}

func (s *Service) UseSMSSender(sender SMSSender) {
	if sender == nil {
		return
	}
	s.smsSender = sender
}

func (s *Service) UseFaceIDStarter(starter FaceIDStarter) {
	if starter == nil {
		return
	}
	s.faceIDStarter = starter
}

func (s *Service) Ensure(userID int64) Record {
	s.mu.Lock()
	record := s.ensureLocked(userID)
	s.mu.Unlock()
	_ = s.persistRecord(record)
	return record
}

func (s *Service) BindPhone(userID int64, phone string) (Record, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return Record{}, ErrPhoneRequired
	}
	if !validMainlandPhone(phone) {
		return Record{}, ErrPhoneInvalid
	}
	s.mu.Lock()
	record := s.ensureLocked(userID)
	record.PhoneMasked = maskPhone(phone)
	record.Status = StatusPhoneBound
	record.UpdatedAt = now()
	s.records[userID] = record
	s.mu.Unlock()
	return record, s.persistRecord(record)
}

func (s *Service) SendSMSCode(userID int64) (SMSDispatchResult, error) {
	s.mu.Lock()
	record := s.ensureLocked(userID)
	nowTime := time.Now()
	dayKey := nowTime.Format("20060102")
	state := s.smsCodes[userID]
	if !state.sentAt.IsZero() && nowTime.Sub(state.sentAt) < smsResendInterval {
		s.mu.Unlock()
		return SMSDispatchResult{}, ErrSMSRateLimited
	}
	if state.dayKey != dayKey {
		state.dayKey = dayKey
		state.dayCount = 0
	}
	if state.dayCount >= smsDailyLimit {
		s.mu.Unlock()
		return SMSDispatchResult{}, ErrSMSDailyLimited
	}
	state.code = s.smsSender.GenerateCode()
	state.sentAt = nowTime
	state.dayCount++
	s.smsCodes[userID] = state
	s.mu.Unlock()
	result, err := s.smsSender.Send(context.Background(), SMSDispatchRequest{
		UserID:      userID,
		PhoneMasked: record.PhoneMasked,
		Scene:       "strong_identity",
		Code:        state.code,
		ExpiresAt:   state.sentAt.Add(5 * time.Minute),
	})
	if err != nil {
		return SMSDispatchResult{}, err
	}
	if s.repo != nil {
		err := s.repo.SaveSMSCode(context.Background(), SMSCodeRecord{
			UserID:      userID,
			Scene:       "strong_identity",
			PhoneMasked: record.PhoneMasked,
			CodeHash:    hashValue(state.code),
			SentAt:      state.sentAt,
			ExpiresAt:   state.sentAt.Add(5 * time.Minute),
		})
		if err != nil {
			return SMSDispatchResult{}, err
		}
	}
	return result, nil
}

func (s *Service) VerifySMSCode(userID int64, code string) (Record, error) {
	code = strings.TrimSpace(code)
	s.mu.Lock()
	if s.smsCodes[userID].code != code && code != temporarySMSCode {
		s.mu.Unlock()
		return Record{}, ErrCodeInvalid
	}
	record := s.ensureLocked(userID)
	record.SMSVerified = true
	record.Status = StatusSMSVerified
	record.UpdatedAt = now()
	s.records[userID] = record
	s.mu.Unlock()
	return record, s.persistRecord(record)
}

func (s *Service) VerifyPhone(userID int64, realName string, idCard string) (Record, error) {
	realName = strings.TrimSpace(realName)
	idCard = strings.ToUpper(strings.TrimSpace(idCard))
	s.mu.Lock()
	record := s.ensureLocked(userID)
	if !record.SMSVerified {
		s.mu.Unlock()
		return Record{}, ErrCodeInvalid
	}
	if realName == "" || idCard == "" {
		s.mu.Unlock()
		return Record{}, ErrRealnameRequired
	}
	if !validIDCard18(idCard) {
		s.mu.Unlock()
		return Record{}, ErrIDCardInvalid
	}
	ciphertext, err := s.encryptRealName(realName)
	if err != nil {
		s.mu.Unlock()
		return Record{}, err
	}
	idCardCiphertext, err := s.encryptIdentitySecret(idCard)
	if err != nil {
		s.mu.Unlock()
		return Record{}, err
	}
	record.RealNameMasked = maskRealName(realName)
	record.RealNameCiphertext = ciphertext
	record.RealNameInitials = RealNameInitials(realName)
	record.IDCardMasked = maskIDCard(idCard)
	record.IDCardCiphertext = idCardCiphertext
	record.PhoneVerified = true
	record.Status = StatusPhoneVerified
	record.UpdatedAt = now()
	s.records[userID] = record
	s.mu.Unlock()
	return record, s.persistRecord(record)
}

func (s *Service) SubmitManualRealname(userID int64, realName string, idCard string) (Record, error) {
	realName = strings.TrimSpace(realName)
	idCard = strings.ToUpper(strings.TrimSpace(idCard))
	if realName == "" || idCard == "" {
		return Record{}, ErrRealnameRequired
	}
	if !validIDCard18(idCard) {
		return Record{}, ErrIDCardInvalid
	}
	if len([]rune(realName)) < 2 || len([]rune(realName)) > 20 {
		return Record{}, ErrRealnameRequired
	}
	ciphertext, err := s.encryptRealName(realName)
	if err != nil {
		return Record{}, err
	}
	idCardCiphertext, err := s.encryptIdentitySecret(idCard)
	if err != nil {
		return Record{}, err
	}
	s.mu.Lock()
	record := s.ensureLocked(userID)
	record.RealNameMasked = maskRealName(realName)
	record.RealNameCiphertext = ciphertext
	record.RealNameInitials = RealNameInitials(realName)
	record.IDCardMasked = maskIDCard(idCard)
	record.IDCardCiphertext = idCardCiphertext
	record.PhoneVerified = false
	record.FaceVerified = false
	record.Status = StatusPendingManualReview
	record.FailureReason = ""
	record.UpdatedAt = now()
	s.records[userID] = record
	s.mu.Unlock()
	return record, s.persistRecord(record)
}

func (s *Service) ReviewManualRealname(userID int64, approve bool, reason string) (Record, error) {
	reason = strings.TrimSpace(reason)
	s.mu.Lock()
	record, ok := s.records[userID]
	if !ok && s.repo != nil {
		s.mu.Unlock()
		saved, found, err := s.repo.FindRecord(context.Background(), userID)
		if err != nil {
			return Record{}, err
		}
		if !found {
			return Record{}, ErrRecordNotFound
		}
		s.mu.Lock()
		record = saved
		ok = true
	}
	if !ok {
		s.mu.Unlock()
		return Record{}, ErrRecordNotFound
	}
	if strings.TrimSpace(record.RealNameMasked) == "" || strings.TrimSpace(record.IDCardMasked) == "" {
		s.mu.Unlock()
		return Record{}, ErrRealnameRequired
	}
	if approve {
		record.Status = StatusVerified
		record.PhoneVerified = true
		record.FailureReason = ""
	} else {
		record.Status = StatusRejected
		if reason == "" {
			reason = "后台审核未通过"
		}
		record.FailureReason = reason
	}
	record.UpdatedAt = now()
	s.records[userID] = record
	s.mu.Unlock()
	return record, s.persistRecord(record)
}

func (s *Service) InGameIdentity(userID int64) (InGameIdentity, bool) {
	record := s.Status(userID)
	if record.Status != StatusVerified && record.Status != StatusFaceVerified && !record.PhoneVerified {
		return InGameIdentity{}, false
	}
	if strings.TrimSpace(record.RealNameCiphertext) == "" {
		return InGameIdentity{}, false
	}
	realName, err := s.decryptRealName(record.RealNameCiphertext)
	if err != nil || strings.TrimSpace(realName) == "" {
		return InGameIdentity{}, false
	}
	initials := strings.ToUpper(strings.TrimSpace(record.RealNameInitials))
	if initials == "" {
		initials = RealNameInitials(realName)
	}
	return InGameIdentity{RealName: realName, DisplayName: realName, AvatarText: initials}, true
}

func (s *Service) RevealRecord(record Record) (PlainIdentity, error) {
	var result PlainIdentity
	if strings.TrimSpace(record.RealNameCiphertext) != "" {
		realName, err := s.decryptRealName(record.RealNameCiphertext)
		if err != nil {
			return PlainIdentity{}, err
		}
		result.RealName = realName
	}
	if strings.TrimSpace(record.IDCardCiphertext) != "" {
		idCard, err := s.decryptIdentitySecret(record.IDCardCiphertext)
		if err != nil {
			return PlainIdentity{}, err
		}
		result.IDCard = idCard
	}
	return result, nil
}

func (s *Service) encryptIdentitySecret(value string) (string, error) {
	return s.encryptRealName(value)
}

func (s *Service) decryptIdentitySecret(value string) (string, error) {
	return s.decryptRealName(value)
}

func (s *Service) encryptRealName(value string) (string, error) {
	block, err := aes.NewCipher(s.dataKey[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(strings.TrimSpace(value)), nil)
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

func (s *Service) decryptRealName(value string) (string, error) {
	data, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(s.dataKey[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("invalid real name ciphertext")
	}
	plain, err := gcm.Open(nil, data[:gcm.NonceSize()], data[gcm.NonceSize():], nil)
	return string(plain), err
}

func RealNameInitials(name string) string {
	runes := make([]rune, 0, 2)
	for _, r := range []rune(strings.TrimSpace(name)) {
		if unicode.Is(unicode.Han, r) || unicode.IsLetter(r) {
			runes = append(runes, r)
			if len(runes) == 2 {
				break
			}
		}
	}
	args := pinyin.NewArgs()
	args.Style = pinyin.FirstLetter
	initials := strings.Builder{}
	for _, r := range runes {
		values := pinyin.SinglePinyin(r, args)
		if len(values) > 0 && values[0] != "" {
			initials.WriteString(strings.ToUpper(values[0][:1]))
		} else if unicode.IsLetter(r) {
			initials.WriteString(strings.ToUpper(string(r)))
		}
	}
	return initials.String()
}

func (s *Service) StartFaceID(userID int64) (string, error) {
	s.mu.Lock()
	record := s.ensureLocked(userID)
	if !record.PhoneVerified {
		s.mu.Unlock()
		return "", ErrPhoneNotVerified
	}
	starter := s.faceIDStarter
	s.mu.Unlock()
	result, err := starter.Start(context.Background(), FaceIDStartRequest{
		UserID:         userID,
		PhoneMasked:    record.PhoneMasked,
		RealNameMasked: record.RealNameMasked,
		IDCardMasked:   record.IDCardMasked,
	})
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(result.FaceToken)
	if token == "" {
		return "", ErrFaceIDStartFailed
	}
	s.mu.Lock()
	record = s.ensureLocked(userID)
	s.faceTokens[userID] = token
	record.Status = StatusFaceIDProcessing
	record.UpdatedAt = now()
	s.records[userID] = record
	s.mu.Unlock()
	if err := s.persistRecord(record); err != nil {
		return "", err
	}
	if s.repo != nil {
		err := s.repo.SaveFaceIDSession(context.Background(), FaceIDSessionRecord{
			UserID:        userID,
			FaceTokenHash: hashValue(token),
			Status:        string(StatusFaceIDProcessing),
			CreatedAt:     time.Now(),
		})
		if err != nil {
			return "", err
		}
	}
	return token, nil
}

func (s *Service) CompleteFaceID(userID int64, token string) (Record, error) {
	s.mu.Lock()
	if s.faceTokens[userID] == "" || s.faceTokens[userID] != strings.TrimSpace(token) {
		s.mu.Unlock()
		return Record{}, ErrFaceIDNotStarted
	}
	record := s.ensureLocked(userID)
	record.FaceVerified = true
	record.WechatRealnameConsistency = "not_supported"
	record.Status = StatusVerified
	record.UpdatedAt = now()
	s.records[userID] = record
	s.mu.Unlock()
	if err := s.persistRecord(record); err != nil {
		return Record{}, err
	}
	if s.repo != nil {
		completedAt := time.Now()
		err := s.repo.SaveFaceIDSession(context.Background(), FaceIDSessionRecord{
			UserID:        userID,
			FaceTokenHash: hashValue(token),
			Status:        string(StatusVerified),
			CreatedAt:     completedAt,
			CompletedAt:   &completedAt,
		})
		if err != nil {
			return Record{}, err
		}
	}
	return record, nil
}

func (s *Service) Status(userID int64) Record {
	s.mu.RLock()
	record, ok := s.records[userID]
	s.mu.RUnlock()
	if ok {
		return record
	}
	if s.repo != nil {
		record, ok, err := s.repo.FindRecord(context.Background(), userID)
		if err == nil && ok {
			s.mu.Lock()
			s.records[userID] = record
			s.mu.Unlock()
			return record
		}
	}
	return s.Ensure(userID)
}

func (s *Service) AllRecords() []Record {
	if s.repo != nil {
		if items, err := s.repo.ListRecords(context.Background()); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Record, 0, len(s.records))
	for _, record := range s.records {
		items = append(items, record)
	}
	return items
}

func (s *Service) IsVerified(userID int64) bool {
	record := s.Status(userID)
	return record.PhoneVerified || record.FaceVerified || record.Status == StatusVerified
}

func (s *Service) ensureLocked(userID int64) Record {
	if record, ok := s.records[userID]; ok {
		return record
	}
	record := Record{
		UserID:                    userID,
		Status:                    StatusWechatLoggedIn,
		WechatRealnameConsistency: "pending",
		UpdatedAt:                 now(),
	}
	s.records[userID] = record
	return record
}

func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

func maskRealName(name string) string {
	runes := []rune(strings.TrimSpace(name))
	if len(runes) == 0 {
		return ""
	}
	if len(runes) == 1 {
		return "*"
	}
	return string(runes[0]) + strings.Repeat("*", len(runes)-1)
}

func maskIDCard(idCard string) string {
	idCard = strings.TrimSpace(idCard)
	if len(idCard) < 8 {
		return strings.Repeat("*", len(idCard))
	}
	return idCard[:3] + strings.Repeat("*", len(idCard)-7) + idCard[len(idCard)-4:]
}

func validMainlandPhone(phone string) bool {
	if len(phone) != 11 || phone[0] != '1' {
		return false
	}
	for _, char := range phone {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func validIDCard18(idCard string) bool {
	return idCard18Pattern.MatchString(strings.TrimSpace(idCard))
}

func now() string {
	return time.Now().Format("2006-01-02T15:04:05Z07:00")
}

func (s *Service) persistRecord(record Record) error {
	if s.repo == nil {
		return nil
	}
	return s.repo.SaveRecord(context.Background(), record)
}

func hashValue(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

type SMSDispatchRequest struct {
	UserID      int64
	PhoneMasked string
	Scene       string
	Code        string
	ExpiresAt   time.Time
}

type SMSDispatchResult struct {
	Provider  string `json:"provider"`
	MessageID string `json:"messageId,omitempty"`
	MockCode  string `json:"mockCode,omitempty"`
}

type SMSSender interface {
	GenerateCode() string
	Send(ctx context.Context, req SMSDispatchRequest) (SMSDispatchResult, error)
}

type LocalSMSSender struct{}

func (LocalSMSSender) GenerateCode() string {
	return temporarySMSCode
}

func (LocalSMSSender) Send(_ context.Context, req SMSDispatchRequest) (SMSDispatchResult, error) {
	return SMSDispatchResult{Provider: "local", MockCode: req.Code}, nil
}

type HTTPSMSSender struct {
	Endpoint string
	Secret   string
	Client   *http.Client
}

func NewHTTPSMSSender(endpoint string, secret string) *HTTPSMSSender {
	return &HTTPSMSSender{
		Endpoint: strings.TrimSpace(endpoint),
		Secret:   strings.TrimSpace(secret),
		Client:   &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *HTTPSMSSender) GenerateCode() string {
	return temporarySMSCode
}

func (s *HTTPSMSSender) Send(ctx context.Context, req SMSDispatchRequest) (SMSDispatchResult, error) {
	endpoint := strings.TrimSpace(s.Endpoint)
	if endpoint == "" {
		return SMSDispatchResult{}, ErrSMSSendFailed
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme != "https" && parsed.Scheme != "http" || parsed.Host == "" {
		return SMSDispatchResult{}, ErrSMSSendFailed
	}
	payload := map[string]interface{}{
		"userId":      req.UserID,
		"phoneMasked": req.PhoneMasked,
		"scene":       req.Scene,
		"code":        req.Code,
		"expiresAt":   req.ExpiresAt.Format(time.RFC3339),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return SMSDispatchResult{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return SMSDispatchResult{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if s.Secret != "" {
		httpReq.Header.Set("X-SMS-Secret", s.Secret)
	}
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return SMSDispatchResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return SMSDispatchResult{}, ErrSMSSendFailed
	}
	var result struct {
		MessageID string `json:"messageId"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return SMSDispatchResult{Provider: "http", MessageID: result.MessageID}, nil
}

type FaceIDStartRequest struct {
	UserID         int64
	PhoneMasked    string
	RealNameMasked string
	IDCardMasked   string
}

type FaceIDStartResult struct {
	Provider  string
	FaceToken string
	RequestID string
}

type FaceIDStarter interface {
	Start(ctx context.Context, req FaceIDStartRequest) (FaceIDStartResult, error)
}

type LocalFaceIDStarter struct{}

func (LocalFaceIDStarter) Start(context.Context, FaceIDStartRequest) (FaceIDStartResult, error) {
	return FaceIDStartResult{Provider: "local", FaceToken: "mock_face_token"}, nil
}

type HTTPFaceIDStarter struct {
	Endpoint string
	Secret   string
	Client   *http.Client
}

func NewHTTPFaceIDStarter(endpoint string, secret string) *HTTPFaceIDStarter {
	return &HTTPFaceIDStarter{
		Endpoint: strings.TrimSpace(endpoint),
		Secret:   strings.TrimSpace(secret),
		Client:   &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *HTTPFaceIDStarter) Start(ctx context.Context, req FaceIDStartRequest) (FaceIDStartResult, error) {
	endpoint := strings.TrimSpace(s.Endpoint)
	if endpoint == "" {
		return FaceIDStartResult{}, ErrFaceIDStartFailed
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme != "https" && parsed.Scheme != "http" || parsed.Host == "" {
		return FaceIDStartResult{}, ErrFaceIDStartFailed
	}
	payload := map[string]interface{}{
		"userId":         req.UserID,
		"phoneMasked":    req.PhoneMasked,
		"realNameMasked": req.RealNameMasked,
		"idCardMasked":   req.IDCardMasked,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return FaceIDStartResult{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return FaceIDStartResult{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if s.Secret != "" {
		httpReq.Header.Set("X-FaceID-Secret", s.Secret)
	}
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return FaceIDStartResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return FaceIDStartResult{}, ErrFaceIDStartFailed
	}
	var result struct {
		FaceToken string `json:"faceToken"`
		BizToken  string `json:"bizToken"`
		RequestID string `json:"requestId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return FaceIDStartResult{}, ErrFaceIDStartFailed
	}
	token := strings.TrimSpace(result.FaceToken)
	if token == "" {
		token = strings.TrimSpace(result.BizToken)
	}
	if token == "" {
		return FaceIDStartResult{}, ErrFaceIDStartFailed
	}
	return FaceIDStartResult{Provider: "http", FaceToken: token, RequestID: result.RequestID}, nil
}
