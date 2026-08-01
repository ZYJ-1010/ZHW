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
	mu       sync.Mutex
	records  map[string]idempotencyRecord
	inflight map[string]chan struct{}
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
	return &idempotencyStore{
		records:  make(map[string]idempotencyRecord),
		inflight: make(map[string]chan struct{}),
	}
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
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "幂等请求标识过长")
			return
		}
		userID, _ := appUserIDFromRequest(r)
		cacheKey := idempotencyCacheKey(userID, r.Method, r.URL.EscapedPath(), key)
		for {
			if record, wait, execute := s.idempotency.begin(cacheKey); !execute {
				if wait == nil {
					replayIdempotencyRecord(w, record)
					return
				}
				select {
				case <-wait:
					continue
				case <-r.Context().Done():
					return
				}
			}
			break
		}
		recorder := &idempotencyResponseWriter{ResponseWriter: w, status: http.StatusOK}
		var record *idempotencyRecord
		defer func() { s.idempotency.finish(cacheKey, record) }()
		next(recorder, r)
		if recorder.status >= 200 && recorder.status < 300 {
			saved := idempotencyRecord{
				status: recorder.status,
				header: cloneHeader(w.Header()),
				body:   append([]byte(nil), recorder.body.Bytes()...),
			}
			record = &saved
		}
	}
}

func replayIdempotencyRecord(w http.ResponseWriter, record idempotencyRecord) {
	copyHeader(w.Header(), record.header)
	w.WriteHeader(record.status)
	_, _ = w.Write(record.body)
}

func (w *idempotencyResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *idempotencyResponseWriter) Write(body []byte) (int, error) {
	w.body.Write(body)
	return w.ResponseWriter.Write(body)
}

func (s *idempotencyStore) begin(key string) (idempotencyRecord, <-chan struct{}, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if record, ok := s.records[key]; ok {
		return record, nil, false
	}
	if wait, ok := s.inflight[key]; ok {
		return idempotencyRecord{}, wait, false
	}
	s.inflight[key] = make(chan struct{})
	return idempotencyRecord{}, nil, true
}

func (s *idempotencyStore) finish(key string, record *idempotencyRecord) {
	s.mu.Lock()
	if record != nil {
		s.records[key] = *record
	}
	wait := s.inflight[key]
	delete(s.inflight, key)
	if wait != nil {
		close(wait)
	}
	s.mu.Unlock()
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
