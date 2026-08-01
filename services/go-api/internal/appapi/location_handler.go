package appapi

import (
	"encoding/json"
	"errors"
	"math"
	"net"
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
	CurrentStrict(userID int64) (lbs.Location, bool, error)
	Recent(userID int64, limit int) []lbs.Location
	RecentStrict(userID int64, limit int) ([]lbs.Location, error)
	RecentBySource(userID int64, source string, limit int) []lbs.Location
	RecentBySourceStrict(userID int64, source string, limit int) ([]lbs.Location, error)
	NearbyUsers(userID int64, center lbs.Location, radiusMeter float64, limit int) []lbs.Location
	NearbyUsersStrict(userID int64, center lbs.Location, radiusMeter float64, limit int) ([]lbs.Location, error)
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
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "定位来源参数错误")
		return
	}
	items, err := s.lbs.RecentBySourceStrict(userID, source, limit)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取历史定位失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

// locationFallback supplies a same-city recommendation context without asking
// for device coordinates. The request IP is only used as a future provider
// input; until an IP-city provider is configured, the configured map city is
// returned explicitly as a default rather than being presented as GPS data.
func (s *Server) locationFallback(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	location, found, err := s.lbs.CurrentStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取定位信息失败，请稍后重试")
		return
	}
	if found && strings.TrimSpace(location.CityCode) != "" {
		httpx.OK(w, map[string]interface{}{"source": "saved", "cityCode": location.CityCode, "cityName": location.CityName, "latitude": location.Latitude, "longitude": location.Longitude, "sortMode": "city_then_time"})
		return
	}
	config := s.currentMapIndexConfig()
	cityName := "默认城市"
	cityCode := ""
	source := "default_city"
	if provider, ok := s.mapProvider.(lbs.IPGeoProvider); ok {
		ip := requestRemoteIP(r)
		place, err := provider.LocateIP(r.Context(), ip)
		if err == nil && strings.TrimSpace(place.City) != "" {
			cityName, cityCode, source = place.City, place.CityCode, "ip_city"
			if place.Longitude != 0 && place.Latitude != 0 {
				config.DefaultLocation.Longitude, config.DefaultLocation.Latitude = place.Longitude, place.Latitude
			}
		}
	} else if s.mapProvider != nil {
		place, err := s.mapProvider.ReverseGeocode(r.Context(), lbs.MapReverseGeocodeRequest{Longitude: config.DefaultLocation.Longitude, Latitude: config.DefaultLocation.Latitude})
		if err == nil && strings.TrimSpace(place.City) != "" {
			cityName, cityCode = place.City, place.CityCode
		}
	}
	httpx.OK(w, map[string]interface{}{
		"source": source, "cityCode": cityCode, "cityName": cityName,
		"latitude": config.DefaultLocation.Latitude, "longitude": config.DefaultLocation.Longitude,
		"sortMode": "city_then_time", "message": "未获取定位，已按默认城市坐标推荐",
	})
}

func requestRemoteIP(r *http.Request) string {
	// In production the API is normally behind a reverse proxy.  Prefer the
	// first forwarded address so IP-city fallback uses the visitor's address
	// rather than the proxy/container address.  This value only affects a
	// recommendation fallback and is never used for authentication or access
	// control.
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		for _, value := range strings.Split(r.Header.Get(header), ",") {
			value = strings.TrimSpace(value)
			if net.ParseIP(value) != nil {
				return value
			}
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
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
	var location lbs.Location
	hasQueryLocation := false
	if rawLon := r.URL.Query().Get("longitude"); rawLon != "" {
		rawLat := r.URL.Query().Get("latitude")
		lon, lonErr := strconv.ParseFloat(rawLon, 64)
		lat, latErr := strconv.ParseFloat(rawLat, 64)
		if lonErr != nil || latErr != nil || lon < -180 || lon > 180 || lat < -90 || lat > 90 || math.IsNaN(lon) || math.IsNaN(lat) || math.IsInf(lon, 0) || math.IsInf(lat, 0) {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "定位坐标参数错误")
			return
		}
		location.UserID = userID
		location.Longitude = lon
		location.Latitude = lat
		location.Source = "query"
		hasQueryLocation = true
	}
	if !hasQueryLocation {
		var ok bool
		var err error
		location, ok, err = s.lbs.CurrentStrict(userID)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取定位信息失败，请稍后重试")
			return
		}
		if !ok {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请先保存定位")
			return
		}
	}
	radius := float64(s.currentOperationRules().Map.DefaultRadiusMeters)
	if radius <= 0 {
		radius = s.nearbyDefaultRadiusMeter
	}
	if radius <= 0 {
		radius = 5000
	}
	maxRadius := s.currentOperationRules().Map.MaxRadiusMeters
	if maxRadius <= 0 {
		maxRadius = 50000
	}
	if raw := nearbyRadiusQuery(r); raw != "" {
		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil || parsed <= 0 || parsed > float64(maxRadius) || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "附近半径参数错误")
			return
		}
		radius = parsed
	}
	allGames, err := s.games.ListStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取附近组局失败，请稍后重试")
		return
	}
	items := publicGames(allGames)
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
	onlinePlayers, offlinePlayers, err := s.nearbyPlayerDTOs(userID, location, radius, 30)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取附近玩家失败，请稍后重试")
		return
	}
	s.recordBehavior(userID, "browse_games", "game", 0, map[string]interface{}{"scope": "nearby", "radiusMeter": radius, "count": len(result)})
	httpx.OK(w, map[string]interface{}{
		"items":              result,
		"total":              len(result),
		"radiusMeters":       radius,
		"center":             map[string]float64{"longitude": location.Longitude, "latitude": location.Latitude},
		"onlinePlayers":      onlinePlayers,
		"offlinePlayers":     offlinePlayers,
		"onlinePlayerCount":  len(onlinePlayers),
		"offlinePlayerCount": len(offlinePlayers),
	})
}

func nearbyRadiusQuery(r *http.Request) string {
	for _, key := range []string{"radiusMeter", "radiusMeters", "radius"} {
		if raw := r.URL.Query().Get(key); raw != "" {
			return raw
		}
	}
	return ""
}

func (s *Server) nearbyPlayerDTOs(userID int64, center lbs.Location, radius float64, limit int) ([]map[string]interface{}, []map[string]interface{}, error) {
	onlinePlayers := make([]map[string]interface{}, 0)
	offlinePlayers := make([]map[string]interface{}, 0)
	nearbyUsers, err := s.lbs.NearbyUsersStrict(userID, center, radius, limit)
	if err != nil {
		return nil, nil, err
	}
	for _, item := range nearbyUsers {
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
	return onlinePlayers, offlinePlayers, nil
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
