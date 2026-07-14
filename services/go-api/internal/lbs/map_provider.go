package lbs

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	if resp.Status.Int() != 0 {
		return MapSearchResult{}, fmt.Errorf("%w: %s", ErrMapProviderFailed, resp.Message.String())
	}
	items := make([]MapPlace, 0, len(resp.Data))
	for _, item := range resp.Data {
		items = append(items, MapPlace{
			ID:            item.ID.String(),
			Title:         item.Title.String(),
			Address:       item.Address.String(),
			Category:      item.Category.String(),
			Province:      item.AdInfo.Province.String(),
			City:          item.AdInfo.City.String(),
			CityCode:      item.AdInfo.AdCode.String(),
			District:      item.AdInfo.District.String(),
			Longitude:     item.Location.Lng.Float64(),
			Latitude:      item.Location.Lat.Float64(),
			DistanceMeter: item.Distance.Float64(),
		})
	}
	return MapSearchResult{Provider: "tencent", Items: items, Total: resp.Count.Int()}, nil
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
	if resp.Status.Int() != 0 {
		return MapPlace{}, fmt.Errorf("%w: %s", ErrMapProviderFailed, resp.Message.String())
	}
	return MapPlace{
		Title:     req.Address,
		Address:   resp.Result.Title.String(),
		Province:  resp.Result.AdInfo.Province.String(),
		City:      resp.Result.AdInfo.City.String(),
		CityCode:  resp.Result.AdInfo.AdCode.String(),
		District:  resp.Result.AdInfo.District.String(),
		Longitude: resp.Result.Location.Lng.Float64(),
		Latitude:  resp.Result.Location.Lat.Float64(),
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
	if resp.Status.Int() != 0 {
		return MapPlace{}, fmt.Errorf("%w: %s", ErrMapProviderFailed, resp.Message.String())
	}
	return MapPlace{
		Title:     resp.Result.Address.String(),
		Address:   resp.Result.Address.String(),
		Province:  resp.Result.AdInfo.Province.String(),
		City:      resp.Result.AdInfo.City.String(),
		CityCode:  resp.Result.AdInfo.AdCode.String(),
		District:  resp.Result.AdInfo.District.String(),
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
	if resp.Status.Int() != 0 {
		return MapRoute{}, fmt.Errorf("%w: %s", ErrMapProviderFailed, resp.Message.String())
	}
	if len(resp.Result.Routes) == 0 {
		return MapRoute{}, ErrMapProviderFailed
	}
	return MapRoute{
		Provider:       "tencent",
		Mode:           mode,
		DistanceMeter:  resp.Result.Routes[0].Distance.Int(),
		DurationSecond: resp.Result.Routes[0].Duration.Int(),
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
		return fmt.Errorf("%w: build request: %v", ErrMapProviderFailed, err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: request: %v", ErrMapProviderFailed, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("%w: read response: %v", ErrMapProviderFailed, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: http %d %s", ErrMapProviderFailed, resp.StatusCode, previewBody(body))
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("%w: decode response: %v body=%s", ErrMapProviderFailed, err, previewBody(body))
	}
	return nil
}

func previewBody(body []byte) string {
	text := strings.TrimSpace(string(body))
	if len(text) > 200 {
		return text[:200]
	}
	return text
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
	Lat tencentFloat `json:"lat"`
	Lng tencentFloat `json:"lng"`
}

type tencentAdInfo struct {
	Province tencentString     `json:"province"`
	City     tencentString     `json:"city"`
	AdCode   tencentStringCode `json:"adcode"`
	District tencentString     `json:"district"`
}

type tencentString string
type tencentStringCode string
type tencentFloat float64
type tencentInt int

func (s *tencentString) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*s = tencentString(strings.TrimSpace(text))
		return nil
	}
	var number json.Number
	if err := json.Unmarshal(data, &number); err == nil {
		*s = tencentString(number.String())
		return nil
	}
	if string(data) == "null" {
		*s = ""
		return nil
	}
	return fmt.Errorf("invalid tencent string: %s", previewBody(data))
}

func (s tencentString) String() string {
	return string(s)
}

func (c *tencentStringCode) UnmarshalJSON(data []byte) error {
	var value tencentString
	if err := value.UnmarshalJSON(data); err != nil {
		return err
	}
	*c = tencentStringCode(value.String())
	return nil
}

func (c tencentStringCode) String() string {
	return string(c)
}

func (f *tencentFloat) UnmarshalJSON(data []byte) error {
	var number float64
	if err := json.Unmarshal(data, &number); err == nil {
		*f = tencentFloat(number)
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		parsed, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
		if err != nil {
			return fmt.Errorf("invalid tencent float: %s", previewBody(data))
		}
		*f = tencentFloat(parsed)
		return nil
	}
	if string(data) == "null" {
		*f = 0
		return nil
	}
	return fmt.Errorf("invalid tencent float: %s", previewBody(data))
}

func (f tencentFloat) Float64() float64 {
	return float64(f)
}

func (i *tencentInt) UnmarshalJSON(data []byte) error {
	var number int
	if err := json.Unmarshal(data, &number); err == nil {
		*i = tencentInt(number)
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		parsed, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil {
			return fmt.Errorf("invalid tencent int: %s", previewBody(data))
		}
		*i = tencentInt(parsed)
		return nil
	}
	if string(data) == "null" {
		*i = 0
		return nil
	}
	return fmt.Errorf("invalid tencent int: %s", previewBody(data))
}

func (i tencentInt) Int() int {
	return int(i)
}

type tencentSearchResponse struct {
	Status  tencentInt    `json:"status"`
	Message tencentString `json:"message"`
	Count   tencentInt    `json:"count"`
	Data    []struct {
		ID       tencentString   `json:"id"`
		Title    tencentString   `json:"title"`
		Address  tencentString   `json:"address"`
		Category tencentString   `json:"category"`
		Location tencentLocation `json:"location"`
		AdInfo   tencentAdInfo   `json:"ad_info"`
		Distance tencentFloat    `json:"_distance"`
	} `json:"data"`
}

type tencentGeocodeResponse struct {
	Status  tencentInt    `json:"status"`
	Message tencentString `json:"message"`
	Result  struct {
		Title    tencentString   `json:"title"`
		Location tencentLocation `json:"location"`
		AdInfo   tencentAdInfo   `json:"ad_info"`
	} `json:"result"`
}

type tencentReverseGeocodeResponse struct {
	Status  tencentInt    `json:"status"`
	Message tencentString `json:"message"`
	Result  struct {
		Address tencentString `json:"address"`
		AdInfo  tencentAdInfo `json:"ad_info"`
	} `json:"result"`
}

type tencentRouteResponse struct {
	Status  tencentInt    `json:"status"`
	Message tencentString `json:"message"`
	Result  struct {
		Routes []struct {
			Distance tencentInt `json:"distance"`
			Duration tencentInt `json:"duration"`
		} `json:"routes"`
	} `json:"result"`
}
