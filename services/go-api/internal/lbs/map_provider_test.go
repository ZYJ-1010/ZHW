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
