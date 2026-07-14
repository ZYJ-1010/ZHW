package appapi

import (
	"testing"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

func TestHomeRoleNetworkAlwaysReturnsConfiguredItems(t *testing.T) {
	server := newTestAppServer(
		auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore()),
		identity.NewService(),
	)

	network := server.homeRoleNetwork(10003, nil, 0, 0)
	items, ok := network["items"].([]map[string]interface{})
	if !ok || len(items) < 2 {
		t.Fatalf("expected guide industry fallback and all-items entry, got %#v", network["items"])
	}
	for _, item := range items {
		if item["icon"] == "" || item["name"] == "" {
			t.Fatalf("expected non-empty guide network icon and name, got %#v", item)
		}
	}
}

func TestParseGuideApplicationIndustries(t *testing.T) {
	industries, audiences, city := parseGuideApplicationIndustries("所在城市：宁波\n可推荐人群：学生、宝妈\n业务说明：业务1：剧本杀组局，定价100；业务2：AI技术交流，定价200")
	if len(industries) != 2 || industries[0] != "剧本杀组局" || industries[1] != "AI技术交流" {
		t.Fatalf("unexpected industries: %#v", industries)
	}
	if len(audiences) != 2 || city != "宁波" {
		t.Fatalf("unexpected audience or city: audiences=%#v city=%q", audiences, city)
	}
}
