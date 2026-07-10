package httpx

import (
	"log"
	"net/http"
	"strconv"
	"time"
)

const RequestIDHeader = "X-Request-Id"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = strconv.FormatInt(time.Now().UnixNano(), 36)
		}
		r.Header.Set(RequestIDHeader, requestID)
		w.Header().Set(RequestIDHeader, requestID)
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(data)
}

func AccessLog(next http.Handler, logger *log.Logger) http.Handler {
	if logger == nil {
		logger = log.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		logger.Printf("request method=%s path=%s status=%d duration_ms=%d request_id=%s",
			r.Method,
			r.URL.EscapedPath(),
			status,
			time.Since(start).Milliseconds(),
			r.Header.Get(RequestIDHeader),
		)
	})
}
