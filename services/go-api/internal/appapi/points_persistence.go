package appapi

import (
	"log"
	"net/http"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/points"
)

const pointsPersistenceHeader = "X-Points-Persistence"

func markPointsPersistenceDegraded(w http.ResponseWriter, operation string, userID int64, err error) {
	if w != nil {
		w.Header().Set(pointsPersistenceHeader, "degraded")
	}
	log.Printf("points persistence degraded operation=%q user_id=%d err=%v", operation, userID, err)
}

func (s *Server) loadPointsData(w http.ResponseWriter, userID int64, includeLogs bool) (points.Account, []points.Log, bool) {
	account, err := s.points.SummaryStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "积分数据读取失败，请稍后重试")
		return points.Account{}, nil, false
	}
	if !includeLogs {
		return account, nil, true
	}
	logs, err := s.points.LogsStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "积分明细读取失败，请稍后重试")
		return points.Account{}, nil, false
	}
	return account, logs, true
}
