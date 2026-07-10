package appapi

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"zhw-mini/services/go-api/internal/common/httpx"
)

type idempotencyStore struct {
	mu      sync.RWMutex
	records map[string]idempotencyRecord
}

type idempotencyRecord struct {
	status int
	header http.Header
	body   []byte
}

type idempotencyResponseWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func newIdempotencyStore() *idempotencyStore {
	return &idempotencyStore{records: make(map[string]idempotencyRecord)}
}

func (s *Server) IdempotencyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	if s.idempotency == nil {
		s.idempotency = newIdempotencyStore()
	}
	return func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if key == "" || !isWriteMethod(r.Method) {
			next(w, r)
			return
		}
		if len(key) > 128 {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "idempotency key too long")
			return
		}
		userID, _ := appUserIDFromRequest(r)
		cacheKey := idempotencyCacheKey(userID, r.Method, r.URL.EscapedPath(), key)
		if record, ok := s.idempotency.get(cacheKey); ok {
			copyHeader(w.Header(), record.header)
			w.WriteHeader(record.status)
			_, _ = w.Write(record.body)
			return
		}
		recorder := &idempotencyResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next(recorder, r)
		if recorder.status >= 200 && recorder.status < 300 {
			s.idempotency.set(cacheKey, idempotencyRecord{
				status: recorder.status,
				header: cloneHeader(w.Header()),
				body:   append([]byte(nil), recorder.body.Bytes()...),
			})
		}
	}
}

func (w *idempotencyResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *idempotencyResponseWriter) Write(body []byte) (int, error) {
	w.body.Write(body)
	return w.ResponseWriter.Write(body)
}

func (s *idempotencyStore) get(key string) (idempotencyRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.records[key]
	return record, ok
}

func (s *idempotencyStore) set(key string, record idempotencyRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[key] = record
}

func isWriteMethod(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete
}

func idempotencyCacheKey(userID int64, method string, path string, key string) string {
	return strings.Join([]string{method, path, key, strconv.FormatInt(userID, 10)}, "|")
}

func cloneHeader(header http.Header) http.Header {
	cloned := make(http.Header, len(header))
	for key, values := range header {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}

func copyHeader(dst http.Header, src http.Header) {
	for key := range dst {
		dst.Del(key)
	}
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}
