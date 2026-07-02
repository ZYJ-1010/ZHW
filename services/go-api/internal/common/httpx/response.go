package httpx

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

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
	_ = json.NewEncoder(w).Encode(payload)
}
