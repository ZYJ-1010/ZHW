package lbs

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	ErrMapProviderUnavailable = errors.New("map provider unavailable")
	ErrMapRequestInvalid      = errors.New("map request invalid")
	ErrMapProviderFailed      = errors.New("map provider failed")
)

type MapProvider interface {
	Search(ctx context.Context, req MapSearchRequest) (MapSearchResult, error)
	Geocode(ctx context.Context, req MapGeocodeRequest) (MapPlace, error)
	ReverseGeocode(ctx context.Context, req MapReverseGeocodeRequest) (MapPlace, error)
	Route(ctx context.Context, req MapRouteRequest) (MapRoute, error)
}

type MapSearchRequest struct {
	Keyword     string
	City        string
	Longitude   float64
	Latitude    float64
	RadiusMeter float64
	Page        int
	PageSize    int
}

type MapSearchResult struct {
	Provider string     `json:"provider"`
	Items    []MapPlace `json:"items"`
	Total    int        `json:"total"`
}

type MapGeocodeRequest struct {
	Address string
	City    string
}

type MapReverseGeocodeRequest struct {
	Longitude float64
	Latitude  float64
}

type MapRouteRequest struct {
	FromLongitude float64
	FromLatitude  float64
	ToLongitude   float64
	ToLatitude    float64
	Mode          string
}

type MapPlace struct {
	ID            string  `json:"id,omitempty"`
	Title         string  `json:"title"`
	Address       string  `json:"address,omitempty"`
	Category      string  `json:"category,omitempty"`
	Province      string  `json:"province,omitempty"`
	City          string  `json:"city,omitempty"`
	CityCode      string  `json:"cityCode,omitempty"`
	District      string  `json:"district,omitempty"`
	Longitude     float64 `json:"longitude"`
	Latitude      float64 `json:"latitude"`
	DistanceMeter float64 `json:"distanceMeter,omitempty"`
}

type MapRoute struct {
	Provider       string `json:"provider"`
	Mode           string `json:"mode"`
	DistanceMeter  int    `json:"distanceMeter"`
	DurationSecond int    `json:"durationSecond"`
}

type TencentMapConfig struct {
	KeyServer string
	SK        string
	APIBase   string
	Timeout   time.Duration
	Client    *http.Client
}

type TencentMapClient struct {
	keyServer string
	sk        string
	apiBase   string
	client    *http.Client
}

func NewTencentMapClient(cfg TencentMapConfig) (*TencentMapClient, error) {
	key := strings.TrimSpace(cfg.KeyServer)
	if key == "" {
		return nil, ErrMapProviderUnavailable
	}
	apiBase := strings.TrimRight(strings.TrimSpace(cfg.APIBase), "/")
	if apiBase == "" {
		apiBase = "https://apis.map.qq.com"
	}
	parsed, err := url.Parse(apiBase)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return nil, ErrMapRequestInvalid
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	client := cfg.Client
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}
	return &TencentMapClient{
		keyServer: key,
		sk:        strings.TrimSpace(cfg.SK),
		apiBase:   apiBase,
		client:    client,
	}, nil
}

func (c *TencentMapClient) Search(ctx context.Context, req MapSearchRequest) (MapSearchResult, error) {
	req.Keyword = strings.TrimSpace(req.Keyword)
	req.City = strings.TrimSpace(req.City)
	if req.Keyword == "" || len([]rune(req.Keyword)) > 80 || len([]rune(req.City)) > 64 {
		return MapSearchResult{}, ErrMapRequestInvalid
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.Page > 50 || req.PageSize > 20 {
		return MapSearchResult{}, ErrMapRequestInvalid
	}
	values := url.Values{}
	values.Set("keyword", req.Keyword)
	values.Set("page_index", strconv.Itoa(req.Page))
	values.Set("page_size", strconv.Itoa(req.PageSize))
	if req.Longitude != 0 || req.Latitude != 0 {
		if !validCoordinate(req.Longitude, req.Latitude) || req.RadiusMeter <= 0 || req.RadiusMeter > 50000 {
			return MapSearchResult{}, ErrMapRequestInvalid
		}
		values.Set("boundary", fmt.Sprintf("nearby(%f,%f,%.0f)", req.Latitude, req.Longitude, req.RadiusMeter))
	} else {
		if req.City == "" {
			req.City = "全国"
		}
		values.Set("boundary", "region("+req.City+",0)")
	}
	var resp tencentSearchResponse
	if err := c.getJSON(ctx, "/ws/place/v1/search", values, &resp); err != nil {
		return MapSearchResult{}, err
	}
	if resp.Status != 0 {
		return MapSearchResult{}, fmt.Errorf("%w: %s", ErrMapProviderFailed, resp.Message)
	}
	items := make([]MapPlace, 0, len(resp.Data))
	for _, item := range resp.Data {
		items = append(items, MapPlace{
			ID:            item.ID,
			Title:         item.Title,
			Address:       item.Address,
			Category:      item.Category,
			Province:      item.AdInfo.Province,
			City:          item.AdInfo.City,
			CityCode:      item.AdInfo.AdCode,
			District:      item.AdInfo.District,
			Longitude:     item.Location.Lng,
			Latitude:      item.Location.Lat,
			DistanceMeter: item.Distance,
		})
	}
	return MapSearchResult{Provider: "tencent", Items: items, Total: resp.Count}, nil
}

func (c *TencentMapClient) Geocode(ctx context.Context, req MapGeocodeRequest) (MapPlace, error) {
	req.Address = strings.TrimSpace(req.Address)
	req.City = strings.TrimSpace(req.City)
	if req.Address == "" || len([]rune(req.Address)) > 160 || len([]rune(req.City)) > 64 {
		return MapPlace{}, ErrMapRequestInvalid
	}
	values := url.Values{}
	values.Set("address", req.Address)
	if req.City != "" {
		values.Set("region", req.City)
	}
	var resp tencentGeocodeResponse
	if err := c.getJSON(ctx, "/ws/geocoder/v1", values, &resp); err != nil {
		return MapPlace{}, err
	}
	if resp.Status != 0 {
		return MapPlace{}, fmt.Errorf("%w: %s", ErrMapProviderFailed, resp.Message)
	}
	return MapPlace{
		Title:     req.Address,
		Address:   resp.Result.Title,
		Province:  resp.Result.AdInfo.Province,
		City:      resp.Result.AdInfo.City,
		CityCode:  resp.Result.AdInfo.AdCode,
		District:  resp.Result.AdInfo.District,
		Longitude: resp.Result.Location.Lng,
		Latitude:  resp.Result.Location.Lat,
	}, nil
}

func (c *TencentMapClient) ReverseGeocode(ctx context.Context, req MapReverseGeocodeRequest) (MapPlace, error) {
	if !validCoordinate(req.Longitude, req.Latitude) {
		return MapPlace{}, ErrMapRequestInvalid
	}
	values := url.Values{}
	values.Set("location", fmt.Sprintf("%f,%f", req.Latitude, req.Longitude))
	values.Set("get_poi", "1")
	var resp tencentReverseGeocodeResponse
	if err := c.getJSON(ctx, "/ws/geocoder/v1", values, &resp); err != nil {
		return MapPlace{}, err
	}
	if resp.Status != 0 {
		return MapPlace{}, fmt.Errorf("%w: %s", ErrMapProviderFailed, resp.Message)
	}
	return MapPlace{
		Title:     resp.Result.Address,
		Address:   resp.Result.Address,
		Province:  resp.Result.AdInfo.Province,
		City:      resp.Result.AdInfo.City,
		CityCode:  resp.Result.AdInfo.AdCode,
		District:  resp.Result.AdInfo.District,
		Longitude: req.Longitude,
		Latitude:  req.Latitude,
	}, nil
}

func (c *TencentMapClient) Route(ctx context.Context, req MapRouteRequest) (MapRoute, error) {
	if !validCoordinate(req.FromLongitude, req.FromLatitude) || !validCoordinate(req.ToLongitude, req.ToLatitude) {
		return MapRoute{}, ErrMapRequestInvalid
	}
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "driving"
	}
	if mode != "driving" && mode != "walking" && mode != "bicycling" {
		return MapRoute{}, ErrMapRequestInvalid
	}
	values := url.Values{}
	values.Set("from", fmt.Sprintf("%f,%f", req.FromLatitude, req.FromLongitude))
	values.Set("to", fmt.Sprintf("%f,%f", req.ToLatitude, req.ToLongitude))
	var resp tencentRouteResponse
	if err := c.getJSON(ctx, "/ws/direction/v1/"+mode, values, &resp); err != nil {
		return MapRoute{}, err
	}
	if resp.Status != 0 {
		return MapRoute{}, fmt.Errorf("%w: %s", ErrMapProviderFailed, resp.Message)
	}
	if len(resp.Result.Routes) == 0 {
		return MapRoute{}, ErrMapProviderFailed
	}
	return MapRoute{
		Provider:       "tencent",
		Mode:           mode,
		DistanceMeter:  resp.Result.Routes[0].Distance,
		DurationSecond: resp.Result.Routes[0].Duration,
	}, nil
}

func (c *TencentMapClient) getJSON(ctx context.Context, path string, values url.Values, target interface{}) error {
	values.Set("key", c.keyServer)
	if c.sk != "" {
		values.Set("sig", tencentSignature(path, values, c.sk))
	}
	endpoint := c.apiBase + path + "?" + values.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: http %d", ErrMapProviderFailed, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func tencentSignature(path string, values url.Values, sk string) string {
	clone := url.Values{}
	for key, vals := range values {
		if key == "sig" {
			continue
		}
		for _, value := range vals {
			clone.Add(key, value)
		}
	}
	sum := md5.Sum([]byte(path + "?" + clone.Encode() + sk))
	return hex.EncodeToString(sum[:])
}

func validCoordinate(longitude float64, latitude float64) bool {
	return longitude >= -180 && longitude <= 180 && latitude >= -90 && latitude <= 90
}

type tencentLocation struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type tencentAdInfo struct {
	Province string `json:"province"`
	City     string `json:"city"`
	AdCode   string `json:"adcode"`
	District string `json:"district"`
}

type tencentSearchResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Count   int    `json:"count"`
	Data    []struct {
		ID       string          `json:"id"`
		Title    string          `json:"title"`
		Address  string          `json:"address"`
		Category string          `json:"category"`
		Location tencentLocation `json:"location"`
		AdInfo   tencentAdInfo   `json:"ad_info"`
		Distance float64         `json:"_distance"`
	} `json:"data"`
}

type tencentGeocodeResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Title    string          `json:"title"`
		Location tencentLocation `json:"location"`
		AdInfo   tencentAdInfo   `json:"ad_info"`
	} `json:"result"`
}

type tencentReverseGeocodeResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Address string        `json:"address"`
		AdInfo  tencentAdInfo `json:"ad_info"`
	} `json:"result"`
}

type tencentRouteResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Routes []struct {
			Distance int `json:"distance"`
			Duration int `json:"duration"`
		} `json:"routes"`
	} `json:"result"`
}
