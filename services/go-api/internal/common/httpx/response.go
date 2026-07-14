package httpx

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var fractionalRFC3339Pattern = regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})\.\d+(Z|[+-]\d{2}:\d{2})`)

type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"requestId,omitempty"`
}

func OK(w http.ResponseWriter, data interface{}) {
	Write(w, http.StatusOK, Response{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

func Error(w http.ResponseWriter, status int, code int, message string) {
	Write(w, status, Response{
		Code:    code,
		Message: message,
	})
}

func Write(w http.ResponseWriter, status int, payload Response) {
	if payload.RequestID == "" {
		payload.RequestID = w.Header().Get(RequestIDHeader)
	}
	if payload.RequestID == "" {
		payload.RequestID = strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set(RequestIDHeader, payload.RequestID)
	w.WriteHeader(status)
	encoded, err := json.Marshal(payload)
	if err != nil {
		return
	}
	encoded = fractionalRFC3339Pattern.ReplaceAll(encoded, []byte("$1$2"))
	_, _ = w.Write(append(encoded, '\n'))
}
