package files

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrCOSSignerNotConfigured = errors.New("cos signer not configured")

type COSPostSigner struct {
	SecretID  string
	SecretKey string
	TTL       time.Duration
	Now       func() time.Time
}

func NewCOSPostSigner(secretID string, secretKey string) *COSPostSigner {
	return &COSPostSigner{
		SecretID:  strings.TrimSpace(secretID),
		SecretKey: strings.TrimSpace(secretKey),
		TTL:       15 * time.Minute,
	}
}

func (s *COSPostSigner) SignForm(storageKey string, maxSize int64) (map[string]string, error) {
	if strings.TrimSpace(s.SecretID) == "" || strings.TrimSpace(s.SecretKey) == "" {
		return nil, ErrCOSSignerNotConfigured
	}
	storageKey = strings.TrimLeft(strings.TrimSpace(storageKey), "/")
	if storageKey == "" {
		return nil, ErrInvalidFile
	}
	if maxSize <= 0 {
		maxSize = 20 * mib
	}
	now := time.Now()
	if s.Now != nil {
		now = s.Now()
	}
	ttl := s.TTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	expiresAt := now.Add(ttl)
	keyTime := fmt.Sprintf("%d;%d", now.Unix(), expiresAt.Unix())
	policy := map[string]interface{}{
		"expiration": expiresAt.UTC().Format(time.RFC3339),
		"conditions": []interface{}{
			map[string]string{"q-sign-algorithm": "sha1"},
			map[string]string{"q-ak": s.SecretID},
			map[string]string{"q-sign-time": keyTime},
			map[string]string{"q-key-time": keyTime},
			[]interface{}{"eq", "$key", storageKey},
			[]interface{}{"content-length-range", 1, maxSize},
		},
	}
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return nil, err
	}
	encodedPolicy := base64.StdEncoding.EncodeToString(policyJSON)
	policyHash := sha1.Sum(policyJSON)
	stringToSign := hex.EncodeToString(policyHash[:])
	signKey := hmacSHA1Hex([]byte(s.SecretKey), keyTime)
	signature := hmacSHA1Hex([]byte(signKey), stringToSign)

	return map[string]string{
		"key":                   storageKey,
		"policy":                encodedPolicy,
		"q-sign-algorithm":      "sha1",
		"q-ak":                  s.SecretID,
		"q-key-time":            keyTime,
		"q-sign-time":           keyTime,
		"q-signature":           signature,
		"success_action_status": "200",
	}, nil
}

func hmacSHA1Hex(key []byte, value string) string {
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}
