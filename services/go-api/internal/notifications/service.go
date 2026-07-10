package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrForbidden            = errors.New("forbidden")
	ErrWechatSubscribeSend  = errors.New("wechat subscribe send failed")
)

type Notification struct {
	ID               int64     `json:"id"`
	UserID           int64     `json:"userId"`
	NotifyType       string    `json:"notifyType"`
	Title            string    `json:"title"`
	Content          string    `json:"content"`
	BizType          string    `json:"bizType,omitempty"`
	BizID            int64     `json:"bizId,omitempty"`
	Status           string    `json:"status"`
	NeedWechat       bool      `json:"needWechat"`
	WechatState      string    `json:"wechatState,omitempty"`
	WechatTemplateID string    `json:"wechatTemplateId,omitempty"`
	WechatTaskID     int64     `json:"wechatTaskId,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	ReadAt           string    `json:"readAt,omitempty"`
}

type CreateRequest struct {
	UserID           int64
	NotifyType       string
	Title            string
	Content          string
	BizType          string
	BizID            int64
	NeedWechat       bool
	WechatState      string
	WechatTemplateID string
	WechatData       map[string]string
}

type WechatTemplate struct {
	Scene      string `json:"scene"`
	TemplateID string `json:"templateId"`
	Title      string `json:"title"`
	Status     string `json:"status"`
}

type WechatTask struct {
	ID             int64             `json:"id"`
	NotificationID int64             `json:"notificationId"`
	UserID         int64             `json:"userId"`
	Scene          string            `json:"scene"`
	TemplateID     string            `json:"templateId"`
	Status         string            `json:"status"`
	RequestPayload map[string]string `json:"requestPayload,omitempty"`
	ResultCode     string            `json:"resultCode,omitempty"`
	ResultMessage  string            `json:"resultMessage,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
	SentAt         string            `json:"sentAt,omitempty"`
}

type WechatSubscribeSendRequest struct {
	UserID     int64
	OpenID     string
	TemplateID string
	Page       string
	Data       map[string]string
}

type WechatSubscribeSendResult struct {
	Code    string
	Message string
}

type WechatTaskSendBatchResult struct {
	Sent      int          `json:"sent"`
	Failed    int          `json:"failed"`
	Skipped   int          `json:"skipped"`
	FailedIDs []int64      `json:"failedIds,omitempty"`
	Items     []WechatTask `json:"items"`
}

type WechatSubscribeSender interface {
	Send(ctx context.Context, req WechatSubscribeSendRequest) (WechatSubscribeSendResult, error)
}

type OpenIDResolver func(userID int64) (string, bool)

type Repository interface {
	SaveNotification(ctx context.Context, notification Notification, task *WechatTask) (Notification, error)
	ListNotifications(ctx context.Context, userID int64) ([]Notification, error)
	FindNotification(ctx context.Context, notificationID int64) (Notification, bool, error)
	UpdateNotification(ctx context.Context, notification Notification) (Notification, error)
	ListWechatTasks(ctx context.Context) ([]WechatTask, error)
	ListWechatTemplates(ctx context.Context) ([]WechatTemplate, error)
	MarkWechatTaskSent(ctx context.Context, taskID int64, resultCode string, resultMessage string, sentAt string) (WechatTask, error)
}

type Service struct {
	mu              sync.RWMutex
	nextID          int64
	nextWechatID    int64
	notifications   map[int64]Notification
	wechatTasks     map[int64]WechatTask
	wechatTemplates map[string]WechatTemplate
	repo            Repository
	wechatSender    WechatSubscribeSender
	openIDResolver  OpenIDResolver
}

func NewService() *Service {
	return NewServiceWithRepository(nil)
}

func NewServiceWithRepository(repo Repository) *Service {
	return &Service{
		nextID:          1,
		nextWechatID:    1,
		notifications:   make(map[int64]Notification),
		wechatTasks:     make(map[int64]WechatTask),
		wechatTemplates: defaultWechatTemplates(),
		repo:            repo,
		wechatSender:    LocalWechatSubscribeSender{},
	}
}

func (s *Service) UseWechatSubscribeSender(sender WechatSubscribeSender) {
	if sender == nil {
		return
	}
	s.wechatSender = sender
}

func (s *Service) UseOpenIDResolver(resolver OpenIDResolver) {
	if resolver == nil {
		return
	}
	s.openIDResolver = resolver
}

func (s *Service) Create(req CreateRequest) Notification {
	if req.NotifyType == "" {
		req.NotifyType = "system"
	}
	if req.Title == "" {
		req.Title = "通知"
	}
	s.mu.Lock()
	notification := Notification{
		ID:               s.nextID,
		UserID:           req.UserID,
		NotifyType:       req.NotifyType,
		Title:            req.Title,
		Content:          req.Content,
		BizType:          req.BizType,
		BizID:            req.BizID,
		Status:           "unread",
		NeedWechat:       req.NeedWechat,
		WechatState:      defaultWechatState(req.NeedWechat, req.WechatState),
		WechatTemplateID: templateIDFor(s.wechatTemplates, req.NotifyType, req.WechatTemplateID),
		CreatedAt:        time.Now(),
	}
	s.nextID++
	s.notifications[notification.ID] = notification
	if notification.NeedWechat {
		notification.WechatTemplateID = s.templateIDForScene(req.NotifyType, req.WechatTemplateID)
		task := &WechatTask{
			ID:             s.nextWechatID,
			NotificationID: notification.ID,
			UserID:         notification.UserID,
			Scene:          notification.NotifyType,
			TemplateID:     notification.WechatTemplateID,
			Status:         "pending",
			RequestPayload: copyStringMap(req.WechatData),
			CreatedAt:      notification.CreatedAt,
		}
		s.nextWechatID++
		s.wechatTasks[task.ID] = *task
		notification.WechatTaskID = task.ID
		notification.WechatState = "pending"
		s.notifications[notification.ID] = notification
		if s.repo != nil {
			if saved, err := s.repo.SaveNotification(context.Background(), notification, task); err == nil {
				if saved.WechatTaskID > 0 {
					task.ID = saved.WechatTaskID
					task.NotificationID = saved.ID
					s.wechatTasks[task.ID] = *task
				}
				s.notifications[saved.ID] = saved
				s.mu.Unlock()
				return saved
			}
		}
		s.mu.Unlock()
		return notification
	}
	if s.repo != nil {
		if saved, err := s.repo.SaveNotification(context.Background(), notification, nil); err == nil {
			s.notifications[saved.ID] = saved
			s.mu.Unlock()
			return saved
		}
	}
	s.mu.Unlock()
	return notification
}

func (s *Service) List(userID int64) []Notification {
	if s.repo != nil {
		if items, err := s.repo.ListNotifications(context.Background(), userID); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Notification, 0)
	for _, notification := range s.notifications {
		if notification.UserID == userID {
			result = append(result, notification)
		}
	}
	return result
}

func (s *Service) MarkRead(userID int64, notificationID int64) (Notification, error) {
	if s.repo != nil {
		notification, ok, err := s.repo.FindNotification(context.Background(), notificationID)
		if err != nil {
			return Notification{}, err
		}
		if !ok {
			return Notification{}, ErrNotificationNotFound
		}
		if notification.UserID != userID {
			return Notification{}, ErrForbidden
		}
		if notification.Status != "read" {
			notification.Status = "read"
			notification.ReadAt = time.Now().Format(time.RFC3339)
			notification, err = s.repo.UpdateNotification(context.Background(), notification)
			if err != nil {
				return Notification{}, err
			}
		}
		s.mu.Lock()
		s.notifications[notification.ID] = notification
		s.mu.Unlock()
		return notification, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	notification, ok := s.notifications[notificationID]
	if !ok {
		return Notification{}, ErrNotificationNotFound
	}
	if notification.UserID != userID {
		return Notification{}, ErrForbidden
	}
	if notification.Status != "read" {
		notification.Status = "read"
		notification.ReadAt = time.Now().Format(time.RFC3339)
		s.notifications[notificationID] = notification
	}
	return notification, nil
}

func (s *Service) WechatTasks() []WechatTask {
	if s.repo != nil {
		if items, err := s.repo.ListWechatTasks(context.Background()); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]WechatTask, 0, len(s.wechatTasks))
	for _, task := range s.wechatTasks {
		result = append(result, task)
	}
	return result
}

func (s *Service) WechatTemplates() []WechatTemplate {
	if s.repo != nil {
		if items, err := s.repo.ListWechatTemplates(context.Background()); err == nil && len(items) > 0 {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]WechatTemplate, 0, len(s.wechatTemplates))
	for _, item := range s.wechatTemplates {
		result = append(result, item)
	}
	return result
}

func (s *Service) MarkWechatTaskSent(taskID int64, resultCode string, resultMessage string) (WechatTask, error) {
	if s.repo != nil {
		sentAt := time.Now().Format(time.RFC3339)
		task, err := s.repo.MarkWechatTaskSent(context.Background(), taskID, resultCode, resultMessage, sentAt)
		if err != nil {
			return WechatTask{}, err
		}
		s.mu.Lock()
		s.wechatTasks[task.ID] = task
		if notification, ok := s.notifications[task.NotificationID]; ok {
			notification.WechatState = "sent"
			s.notifications[notification.ID] = notification
		}
		s.mu.Unlock()
		return task, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.wechatTasks[taskID]
	if !ok {
		return WechatTask{}, ErrNotificationNotFound
	}
	task.Status = "sent"
	task.ResultCode = resultCode
	task.ResultMessage = resultMessage
	task.SentAt = time.Now().Format(time.RFC3339)
	s.wechatTasks[taskID] = task
	notification := s.notifications[task.NotificationID]
	notification.WechatState = "sent"
	s.notifications[notification.ID] = notification
	return task, nil
}

func (s *Service) SendWechatTask(taskID int64) (WechatTask, error) {
	task, err := s.findWechatTask(taskID)
	if err != nil {
		return WechatTask{}, err
	}
	if task.Status == "sent" {
		return task, nil
	}
	openID := ""
	if s.openIDResolver != nil {
		if resolved, ok := s.openIDResolver(task.UserID); ok {
			openID = strings.TrimSpace(resolved)
		}
	}
	if openID == "" || strings.TrimSpace(task.TemplateID) == "" {
		return WechatTask{}, ErrWechatSubscribeSend
	}
	data := copyStringMap(task.RequestPayload)
	page := strings.TrimSpace(data["page"])
	delete(data, "page")
	result, err := s.wechatSender.Send(context.Background(), WechatSubscribeSendRequest{
		UserID:     task.UserID,
		OpenID:     openID,
		TemplateID: task.TemplateID,
		Page:       page,
		Data:       data,
	})
	if err != nil {
		return WechatTask{}, err
	}
	return s.MarkWechatTaskSent(taskID, result.Code, result.Message)
}

func (s *Service) SendPendingWechatTasks(limit int) WechatTaskSendBatchResult {
	if limit <= 0 {
		limit = 20
	}
	result := WechatTaskSendBatchResult{Items: make([]WechatTask, 0)}
	attempted := 0
	for _, task := range s.WechatTasks() {
		if task.Status != "pending" {
			result.Skipped++
			continue
		}
		if attempted >= limit {
			break
		}
		attempted++
		sent, err := s.SendWechatTask(task.ID)
		if err != nil {
			result.Failed++
			result.FailedIDs = append(result.FailedIDs, task.ID)
			continue
		}
		result.Sent++
		result.Items = append(result.Items, sent)
	}
	return result
}

func (s *Service) findWechatTask(taskID int64) (WechatTask, error) {
	if s.repo != nil {
		tasks, err := s.repo.ListWechatTasks(context.Background())
		if err != nil {
			return WechatTask{}, err
		}
		for _, task := range tasks {
			if task.ID == taskID {
				return task, nil
			}
		}
		return WechatTask{}, ErrNotificationNotFound
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.wechatTasks[taskID]
	if !ok {
		return WechatTask{}, ErrNotificationNotFound
	}
	return task, nil
}

type LocalWechatSubscribeSender struct{}

func (LocalWechatSubscribeSender) Send(_ context.Context, req WechatSubscribeSendRequest) (WechatSubscribeSendResult, error) {
	if strings.TrimSpace(req.OpenID) == "" || strings.TrimSpace(req.TemplateID) == "" {
		return WechatSubscribeSendResult{}, ErrWechatSubscribeSend
	}
	return WechatSubscribeSendResult{Code: "local", Message: "local subscribe message sent"}, nil
}

type WechatSubscribeSenderHTTP struct {
	AppID         string
	AppSecret     string
	TokenEndpoint string
	SendEndpoint  string
	Client        *http.Client
}

func NewWechatSubscribeSender(appID string, appSecret string) *WechatSubscribeSenderHTTP {
	return &WechatSubscribeSenderHTTP{
		AppID:         strings.TrimSpace(appID),
		AppSecret:     strings.TrimSpace(appSecret),
		TokenEndpoint: "https://api.weixin.qq.com/cgi-bin/token",
		SendEndpoint:  "https://api.weixin.qq.com/cgi-bin/message/subscribe/send",
		Client:        &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *WechatSubscribeSenderHTTP) Send(ctx context.Context, req WechatSubscribeSendRequest) (WechatSubscribeSendResult, error) {
	accessToken, err := s.fetchAccessToken(ctx)
	if err != nil {
		return WechatSubscribeSendResult{}, err
	}
	endpoint, err := withAccessToken(s.SendEndpoint, accessToken)
	if err != nil {
		return WechatSubscribeSendResult{}, err
	}
	payload := map[string]interface{}{
		"touser":      strings.TrimSpace(req.OpenID),
		"template_id": strings.TrimSpace(req.TemplateID),
		"data":        subscribeData(req.Data),
	}
	if strings.TrimSpace(req.Page) != "" {
		payload["page"] = strings.TrimSpace(req.Page)
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return WechatSubscribeSendResult{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return WechatSubscribeSendResult{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient().Do(httpReq)
	if err != nil {
		return WechatSubscribeSendResult{}, err
	}
	defer resp.Body.Close()
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return WechatSubscribeSendResult{}, ErrWechatSubscribeSend
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || result.ErrCode != 0 {
		return WechatSubscribeSendResult{}, ErrWechatSubscribeSend
	}
	return WechatSubscribeSendResult{Code: fmt.Sprintf("%d", result.ErrCode), Message: result.ErrMsg}, nil
}

func (s *WechatSubscribeSenderHTTP) fetchAccessToken(ctx context.Context) (string, error) {
	if strings.TrimSpace(s.AppID) == "" || strings.TrimSpace(s.AppSecret) == "" {
		return "", ErrWechatSubscribeSend
	}
	endpoint := strings.TrimSpace(s.TokenEndpoint)
	if endpoint == "" {
		endpoint = "https://api.weixin.qq.com/cgi-bin/token"
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	q := parsed.Query()
	q.Set("grant_type", "client_credential")
	q.Set("appid", s.AppID)
	q.Set("secret", s.AppSecret)
	parsed.RawQuery = q.Encode()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return "", err
	}
	resp, err := s.httpClient().Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		AccessToken string `json:"access_token"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", ErrWechatSubscribeSend
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || result.ErrCode != 0 || strings.TrimSpace(result.AccessToken) == "" {
		return "", ErrWechatSubscribeSend
	}
	return strings.TrimSpace(result.AccessToken), nil
}

func (s *WechatSubscribeSenderHTTP) httpClient() *http.Client {
	if s.Client != nil {
		return s.Client
	}
	return &http.Client{Timeout: 5 * time.Second}
}

func withAccessToken(endpoint string, accessToken string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil {
		return "", err
	}
	q := parsed.Query()
	q.Set("access_token", accessToken)
	parsed.RawQuery = q.Encode()
	return parsed.String(), nil
}

func subscribeData(values map[string]string) map[string]map[string]string {
	result := make(map[string]map[string]string, len(values))
	for key, value := range values {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		result[key] = map[string]string{"value": value}
	}
	return result
}

func defaultWechatState(needWechat bool, value string) string {
	if value != "" {
		return value
	}
	if needWechat {
		return "pending"
	}
	return "not_required"
}

func defaultWechatTemplates() map[string]WechatTemplate {
	return map[string]WechatTemplate{
		"report_created": {
			Scene:      "report_created",
			TemplateID: "mock_report_created_tpl",
			Title:      "举报申诉提交提醒",
			Status:     "active",
		},
		"report_handled": {
			Scene:      "report_handled",
			TemplateID: "mock_report_handled_tpl",
			Title:      "举报申诉处理结果",
			Status:     "active",
		},
		"report_closed": {
			Scene:      "report_closed",
			TemplateID: "mock_report_closed_tpl",
			Title:      "举报申诉关闭提醒",
			Status:     "active",
		},
		"review_remind": {
			Scene:      "review_remind",
			TemplateID: "mock_review_remind_tpl",
			Title:      "评价提醒",
			Status:     "active",
		},
		"progress_feedback_remind": {
			Scene:      "progress_feedback_remind",
			TemplateID: "mock_progress_feedback_tpl",
			Title:      "行家进度提醒",
			Status:     "active",
		},
	}
}

func templateIDFor(templates map[string]WechatTemplate, scene string, fallback string) string {
	if fallback != "" {
		return fallback
	}
	template, ok := templates[scene]
	if ok && template.Status == "active" {
		return template.TemplateID
	}
	return ""
}

func (s *Service) templateIDForScene(scene string, fallback string) string {
	if fallback != "" {
		return fallback
	}
	if s.repo != nil {
		if templates, err := s.repo.ListWechatTemplates(context.Background()); err == nil {
			for _, template := range templates {
				if template.Scene == scene && template.Status == "active" {
					return template.TemplateID
				}
			}
		}
	}
	return templateIDFor(s.wechatTemplates, scene, fallback)
}

func copyStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}
