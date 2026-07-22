package lbs

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func newMapClientForTest(t *testing.T, body string) *TencentMapClient {
	t.Helper()
	client, err := NewTencentMapClient(TencentMapConfig{
		KeyServer: "test-key",
		APIBase:   "https://map.test",
		Client: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
				Request:    req,
			}, nil
		})},
	})
	if err != nil {
		t.Fatalf("new map client failed: %v", err)
	}
	return client
}

func TestTencentMapClientLocateIP(t *testing.T) {
	client := newMapClientForTest(t, `{"status":0,"message":"ok","result":{"location":{"lat":30.2741,"lng":120.1551},"ad_info":{"province":"浙江省","city":"杭州市","adcode":"330100","district":"西湖区"}}}`)
	place, err := client.LocateIP(context.Background(), "127.0.0.1")
	if err != nil {
		t.Fatalf("locate ip: %v", err)
	}
	if place.City != "杭州市" || place.CityCode != "330100" || place.Longitude != 120.1551 {
		t.Fatalf("unexpected IP location: %+v", place)
	}
}

func TestTencentMapSearchMapsAdCodeToCityCode(t *testing.T) {
	client := newMapClientForTest(t, `{
		"status":0,
		"message":"ok",
		"count":1,
		"data":[{
			"id":"poi-1",
			"title":"西湖",
			"address":"杭州市西湖区",
			"category":"景点",
			"location":{"lat":30.2741,"lng":120.1551},
			"ad_info":{"province":"浙江省","city":"杭州市","district":"西湖区","adcode":"330100"},
			"_distance":1200
		}]
	}`)

	result, err := client.Search(context.Background(), MapSearchRequest{Keyword: "西湖", City: "杭州"})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].CityCode != "330100" {
		t.Fatalf("expected cityCode from adcode, got %+v", result.Items)
	}
}

func TestTencentMapSearchAcceptsNumericAdCode(t *testing.T) {
	client := newMapClientForTest(t, `{
		"status":0,
		"message":"ok",
		"count":1,
		"data":[{
			"id":"poi-1",
			"title":"江油站",
			"address":"四川省绵阳市江油市建设北路",
			"category":"交通设施",
			"location":{"lat":31.7787,"lng":104.7459},
			"ad_info":{"province":"四川省","city":"绵阳市","district":"江油市","adcode":510781},
			"_distance":1200
		}]
	}`)

	result, err := client.Search(context.Background(), MapSearchRequest{Keyword: "江油站", City: "江油"})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].CityCode != "510781" {
		t.Fatalf("expected numeric adcode converted to string cityCode, got %+v", result.Items)
	}
}

func TestTencentMapSearchAcceptsMixedPrimitiveTypes(t *testing.T) {
	client := newMapClientForTest(t, `{
		"status":"0",
		"message":"ok",
		"count":"1",
		"data":[{
			"id":12909335601871599327,
			"title":"江油站",
			"address":"四川省绵阳市江油市建设北路",
			"category":"交通设施",
			"location":{"lat":"31.7787","lng":"104.7459"},
			"ad_info":{"province":"四川省","city":"绵阳市","district":"江油市","adcode":510781},
			"_distance":"1200"
		}]
	}`)

	result, err := client.Search(context.Background(), MapSearchRequest{Keyword: "江油站", City: "江油"})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if result.Total != 1 || len(result.Items) != 1 {
		t.Fatalf("expected one result, got %+v", result)
	}
	place := result.Items[0]
	if place.ID == "" || place.CityCode != "510781" || place.DistanceMeter != 1200 || place.Latitude != 31.7787 || place.Longitude != 104.7459 {
		t.Fatalf("expected mixed primitive types normalized, got %+v", place)
	}
}

func TestTencentMapGeocodeMapsAdCodeToCityCode(t *testing.T) {
	client := newMapClientForTest(t, `{
		"status":0,
		"message":"ok",
		"result":{
			"title":"杭州市西湖区",
			"location":{"lat":30.2741,"lng":120.1551},
			"ad_info":{"province":"浙江省","city":"杭州市","district":"西湖区","adcode":"330100"}
		}
	}`)

	place, err := client.Geocode(context.Background(), MapGeocodeRequest{Address: "西湖", City: "杭州"})
	if err != nil {
		t.Fatalf("geocode failed: %v", err)
	}
	if place.CityCode != "330100" {
		t.Fatalf("expected cityCode from adcode, got %+v", place)
	}
}
