package im

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	openIMTextContentType = 101
	openIMGroupSession    = 3
	openIMWorkingGroup    = 2
	openIMMiniProgram     = 5
)

type OpenIMConfig struct {
	APIAddr     string
	Secret      string
	AdminUserID string
	Enabled     bool
}

type OpenIMClient struct {
	baseURL     string
	secret      string
	adminUserID string
	httpClient  *http.Client
}

func NewOpenIMClient(cfg OpenIMConfig) *OpenIMClient {
	if !cfg.Enabled {
		return nil
	}
	return &OpenIMClient{
		baseURL:     strings.TrimRight(cfg.APIAddr, "/"),
		secret:      cfg.Secret,
		adminUserID: cfg.AdminUserID,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *OpenIMClient) SyncGameRoom(ctx context.Context, gameID int64, memberIDs []int64) (string, error) {
	if len(memberIDs) == 0 {
		return "", ErrForbidden
	}
	token, err := c.adminToken(ctx)
	if err != nil {
		return "", err
	}
	users := make([]openIMUser, 0, len(memberIDs))
	userIDs := make([]string, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		userID := openIMUserID(memberID)
		userIDs = append(userIDs, userID)
		users = append(users, openIMUser{UserID: userID, Nickname: userID})
	}
	if err := c.post(ctx, token, "/user/user_register", map[string]any{"users": users}, nil); err != nil {
		return "", err
	}
	groupID := openIMGroupID(gameID)
	req := map[string]any{
		"ownerUserID":   userIDs[0],
		"memberUserIDs": userIDs,
		"groupInfo": map[string]any{
			"groupID":   groupID,
			"groupName": fmt.Sprintf("game-%d", gameID),
			"groupType": openIMWorkingGroup,
		},
	}
	if err := c.post(ctx, token, "/group/create_group", req, nil); err != nil {
		return "", err
	}
	return groupID, nil
}

func (c *OpenIMClient) SendGroupText(ctx context.Context, gameID int64, senderUserID int64, content string) error {
	token, err := c.adminToken(ctx)
	if err != nil {
		return err
	}
	req := map[string]any{
		"sendID":           openIMUserID(senderUserID),
		"senderNickname":   openIMUserID(senderUserID),
		"senderPlatformID": openIMMiniProgram,
		"groupID":          openIMGroupID(gameID),
		"contentType":      openIMTextContentType,
		"sessionType":      openIMGroupSession,
		"content": map[string]any{
			"content": content,
		},
	}
	return c.post(ctx, token, "/msg/send_msg", req, nil)
}

func (c *OpenIMClient) UserToken(ctx context.Context, userID int64) (string, error) {
	var resp struct {
		Token string `json:"token"`
	}
	req := map[string]any{
		"secret":     c.secret,
		"platformID": openIMMiniProgram,
		"userID":     openIMUserID(userID),
	}
	if err := c.post(ctx, "", "/auth/get_user_token", req, &resp); err != nil {
		return "", err
	}
	if resp.Token == "" {
		return "", ErrExternalIM
	}
	return resp.Token, nil
}

func (c *OpenIMClient) adminToken(ctx context.Context) (string, error) {
	var resp struct {
		Token string `json:"token"`
	}
	req := map[string]any{
		"secret": c.secret,
		"userID": c.adminUserID,
	}
	if err := c.post(ctx, "", "/auth/get_admin_token", req, &resp); err != nil {
		return "", err
	}
	if resp.Token == "" {
		return "", ErrExternalIM
	}
	return resp.Token, nil
}

func (c *OpenIMClient) post(ctx context.Context, token string, path string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("operationID", "zhw-"+strconv.FormatInt(time.Now().UnixNano(), 10))
	if token != "" {
		req.Header.Set("token", token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var envelope struct {
		ErrCode int             `json:"errCode"`
		ErrMsg  string          `json:"errMsg"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return err
	}
	if resp.StatusCode >= 400 || envelope.ErrCode != 0 {
		if envelope.ErrMsg == "" {
			envelope.ErrMsg = resp.Status
		}
		return fmt.Errorf("%w: %s", ErrExternalIM, envelope.ErrMsg)
	}
	if out != nil && len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return err
		}
	}
	return nil
}

func openIMUserID(userID int64) string {
	return "zhw_user_" + strconv.FormatInt(userID, 10)
}

func openIMGroupID(gameID int64) string {
	return "zhw_game_" + strconv.FormatInt(gameID, 10)
}

type openIMUser struct {
	UserID   string `json:"userID"`
	Nickname string `json:"nickname"`
	FaceURL  string `json:"faceURL,omitempty"`
}
