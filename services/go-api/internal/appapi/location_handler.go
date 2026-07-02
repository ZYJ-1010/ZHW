package appapi

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/lbs"
)

type lbsService interface {
	SaveCurrent(userID int64, req lbs.SaveRequest) (lbs.Location, error)
	SaveManual(userID int64, req lbs.SaveRequest) (lbs.Location, error)
	Current(userID int64) (lbs.Location, bool)
	Recent(userID int64, limit int) []lbs.Location
	RecentBySource(userID int64, source string, limit int) []lbs.Location
	NearbyUsers(userID int64, center lbs.Location, radiusMeter float64, limit int) []lbs.Location
}

func (s *Server) saveCurrentLocation(w http.ResponseWriter, r *http.Request) {
	s.saveLocation(w, r, true)
}

func (s *Server) saveManualLocation(w http.ResponseWriter, r *http.Request) {
	s.saveLocation(w, r, false)
}

func (s *Server) myRecentLocations(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	limit := 10
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}
	source := r.URL.Query().Get("source")
	if source != "" && source != "gps" && source != "manual" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid location source")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": s.lbs.RecentBySource(userID, source, limit)})
}

func (s *Server) saveLocation(w http.ResponseWriter, r *http.Request, current bool) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req lbs.SaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	var (
		location lbs.Location
		err      error
	)
	if current {
		location, err = s.lbs.SaveCurrent(userID, req)
	} else {
		location, err = s.lbs.SaveManual(userID, req)
	}
	if err != nil {
		if errors.Is(err, lbs.ErrInvalidLocation) {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "定位参数错误")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存定位失败")
		return
	}
	httpx.OK(w, location)
}

func (s *Server) sameCityGames(w http.ResponseWriter, r *http.Request) {
	cityCode := r.URL.Query().Get("cityCode")
	items := publicGames(s.games.SameCity(cityCode))
	if userID, ok := s.currentUserID(r); ok {
		s.recordBehavior(userID, "browse_games", "game", 0, map[string]interface{}{"scope": "same_city", "cityCode": cityCode, "count": len(items)})
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) nearbyGames(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	location, ok := s.lbs.Current(userID)
	if !ok {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请先保存定位")
		return
	}
	if rawLon := r.URL.Query().Get("longitude"); rawLon != "" {
		rawLat := r.URL.Query().Get("latitude")
		lon, lonErr := strconv.ParseFloat(rawLon, 64)
		lat, latErr := strconv.ParseFloat(rawLat, 64)
		if lonErr != nil || latErr != nil || lon < -180 || lon > 180 || lat < -90 || lat > 90 || math.IsNaN(lon) || math.IsNaN(lat) || math.IsInf(lon, 0) || math.IsInf(lat, 0) {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid coordinates")
			return
		}
		location.Longitude = lon
		location.Latitude = lat
	}
	radius := s.nearbyDefaultRadiusMeter
	if radius <= 0 {
		radius = 5000
	}
	if raw := r.URL.Query().Get("radiusMeter"); raw != "" {
		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil || parsed <= 0 || parsed > 50000 || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "附近半径参数错误")
			return
		}
		radius = parsed
	}
	items := publicGames(s.games.List())
	result := make([]games.Game, 0)
	for _, game := range items {
		if game.Longitude == 0 && game.Latitude == 0 {
			continue
		}
		distance := lbs.DistanceMeter(location.Longitude, location.Latitude, game.Longitude, game.Latitude)
		if distance <= radius {
			game.DistanceMeter = math.Round(distance)
			game.DistanceLabel = lbs.FormatDistanceLabel(game.DistanceMeter)
			result = append(result, game)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].DistanceMeter < result[j].DistanceMeter
	})
	onlinePlayers, offlinePlayers := s.nearbyPlayerDTOs(userID, location, radius, 30)
	s.recordBehavior(userID, "browse_games", "game", 0, map[string]interface{}{"scope": "nearby", "radiusMeter": radius, "count": len(result)})
	httpx.OK(w, map[string]interface{}{
		"items":              result,
		"onlinePlayers":      onlinePlayers,
		"offlinePlayers":     offlinePlayers,
		"onlinePlayerCount":  len(onlinePlayers),
		"offlinePlayerCount": len(offlinePlayers),
	})
}

func (s *Server) nearbyPlayerDTOs(userID int64, center lbs.Location, radius float64, limit int) ([]map[string]interface{}, []map[string]interface{}) {
	onlinePlayers := make([]map[string]interface{}, 0)
	offlinePlayers := make([]map[string]interface{}, 0)
	for _, item := range s.lbs.NearbyUsers(userID, center, radius, limit) {
		distance := lbs.DistanceMeter(center.Longitude, center.Latitude, item.Longitude, item.Latitude)
		status := "online"
		if time.Since(item.UpdatedAt) > 15*time.Minute {
			status = "offline"
		}
		dto := map[string]interface{}{
			"id":           strconv.FormatInt(item.UserID, 10),
			"userId":       item.UserID,
			"name":         s.safeNearbyUserName(item.UserID),
			"status":       status,
			"statusText":   mapNearbyPlayerStatusText(status),
			"activeText":   mapNearbyPlayerActiveText(item.UpdatedAt),
			"longitude":    item.Longitude,
			"latitude":     item.Latitude,
			"distanceText": lbs.FormatDistanceLabel(distance),
		}
		if status == "online" {
			onlinePlayers = append(onlinePlayers, dto)
		} else {
			offlinePlayers = append(offlinePlayers, dto)
		}
	}
	return onlinePlayers, offlinePlayers
}

func (s *Server) safeNearbyUserName(userID int64) string {
	if user, ok := s.auth.UserByID(userID); ok && strings.TrimSpace(user.Nickname) != "" {
		return strings.TrimSpace(user.Nickname)
	}
	return "玩家" + strconv.FormatInt(userID, 10)
}

func mapNearbyPlayerStatusText(status string) string {
	if status == "offline" {
		return "离线玩家"
	}
	return "在线玩家"
}

func mapNearbyPlayerActiveText(updatedAt time.Time) string {
	if updatedAt.IsZero() {
		return "刚刚活跃"
	}
	minutes := int(time.Since(updatedAt).Minutes())
	if minutes <= 1 {
		return "刚刚活跃"
	}
	if minutes < 60 {
		return strconv.Itoa(minutes) + "分钟前活跃"
	}
	return strconv.Itoa(minutes/60) + "小时前活跃"
}
