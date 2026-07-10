package revenueclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

type Response struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type TemplateRequest struct {
	Name        string `json:"name"`
	GameType    string `json:"gameType"`
	PlatformBps int    `json:"platformBps"`
	CreatorBps  int    `json:"creatorBps"`
	MemberBps   int    `json:"memberBps"`
}

type CalculateRequest struct {
	GameID     int64   `json:"gameId"`
	AmountCent int64   `json:"amountCent"`
	TemplateID int64   `json:"templateId"`
	MemberIDs  []int64 `json:"memberIds"`
	CreatorID  int64   `json:"creatorUserId"`
}

type SettlementRequest struct {
	RecordID int64  `json:"recordId"`
	Method   string `json:"method"`
	ProofNo  string `json:"proofNo"`
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) Health(ctx context.Context) (Response, error) {
	return c.get(ctx, "/health")
}

func (c *Client) PaymentPrecreatePlaceholder(ctx context.Context, payload map[string]interface{}) (Response, error) {
	return c.post(ctx, "/api/funds/payment-precreate-placeholder", payload)
}

func (c *Client) PaymentCallbackPlaceholder(ctx context.Context, payload map[string]interface{}) (Response, error) {
	return c.post(ctx, "/api/funds/payment-callback-placeholder", payload)
}

func (c *Client) ProfitSharingOrderPlaceholder(ctx context.Context, payload map[string]interface{}) (Response, error) {
	return c.post(ctx, "/api/funds/profit-sharing/orders", payload)
}

func (c *Client) ProfitSharingReturnPlaceholder(ctx context.Context, payload map[string]interface{}) (Response, error) {
	return c.post(ctx, "/api/funds/profit-sharing/return-orders", payload)
}

func (c *Client) CreateTemplate(ctx context.Context, req TemplateRequest) (Response, error) {
	return c.post(ctx, "/api/funds/revenue/templates", req)
}

func (c *Client) Simulate(ctx context.Context, req CalculateRequest) (Response, error) {
	return c.post(ctx, "/api/funds/revenue/simulate", req)
}

func (c *Client) Generate(ctx context.Context, req CalculateRequest) (Response, error) {
	return c.post(ctx, "/api/funds/revenue/generate", req)
}

func (c *Client) SettleOffline(ctx context.Context, req SettlementRequest) (Response, error) {
	return c.post(ctx, "/api/funds/settlements/offline", req)
}

func (c *Client) get(ctx context.Context, path string) (Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return Response{}, err
	}
	return c.do(request)
}

func (c *Client) post(ctx context.Context, path string, payload interface{}) (Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return Response{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	return c.do(request)
}

func (c *Client) do(request *http.Request) (Response, error) {
	resp, err := c.http.Do(request)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()
	var decoded Response
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return Response{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decoded, fmt.Errorf("funds service status %d: %s", resp.StatusCode, decoded.Message)
	}
	if decoded.Code != 0 {
		return decoded, fmt.Errorf("funds service code %d: %s", decoded.Code, decoded.Message)
	}
	return decoded, nil
}
