package appapi

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
)

func TestIdempotencyMiddlewareCoalescesConcurrentRequests(t *testing.T) {
	server := &Server{}
	release := make(chan struct{})
	entered := make(chan struct{}, 2)
	var calls atomic.Int32
	handler := server.IdempotencyMiddleware(func(w http.ResponseWriter, r *http.Request) {
		call := calls.Add(1)
		entered <- struct{}{}
		<-release
		httpx.OK(w, map[string]interface{}{"call": call})
	})

	results := make([]*httptest.ResponseRecorder, 2)
	var group sync.WaitGroup
	start := func(index int) {
		defer group.Done()
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/app/games", nil)
		request.Header.Set("Idempotency-Key", "same-create-request")
		handler(recorder, request)
		results[index] = recorder
	}

	group.Add(1)
	go start(0)
	<-entered
	group.Add(1)
	go start(1)
	select {
	case <-entered:
		close(release)
		group.Wait()
		t.Fatal("同一幂等请求仍被并发执行")
	case <-time.After(30 * time.Millisecond):
	}
	close(release)
	group.Wait()

	if got := calls.Load(); got != 1 {
		t.Fatalf("handler calls = %d, want 1", got)
	}
	if results[0].Code != http.StatusOK || results[1].Code != http.StatusOK || results[0].Body.String() != results[1].Body.String() {
		t.Fatalf("concurrent responses differ: first=%d %s second=%d %s", results[0].Code, results[0].Body.String(), results[1].Code, results[1].Body.String())
	}
}

func TestIdempotencyMiddlewareRetriesAfterFailedRequest(t *testing.T) {
	server := &Server{}
	var calls atomic.Int32
	handler := server.IdempotencyMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "服务暂不可用")
			return
		}
		httpx.OK(w, map[string]interface{}{"saved": true})
	})

	request := func() *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/app/games", nil)
		req.Header.Set("Idempotency-Key", "retry-after-failure")
		handler(recorder, req)
		return recorder
	}
	if first := request(); first.Code != http.StatusServiceUnavailable {
		t.Fatalf("first status = %d, want %d", first.Code, http.StatusServiceUnavailable)
	}
	if second := request(); second.Code != http.StatusOK {
		t.Fatalf("second status = %d, want %d", second.Code, http.StatusOK)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("handler calls = %d, want 2", got)
	}
}
