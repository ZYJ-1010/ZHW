package notifications

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCreateWechatNotificationCreatesTaskAndTemplate(t *testing.T) {
	service := NewService()

	notification := service.Create(CreateRequest{
		UserID:     2,
		NotifyType: "report_created",
		Title:      "举报申诉已提交",
		Content:    "已提交",
		BizType:    "report",
		BizID:      1,
		NeedWechat: true,
		WechatData: map[string]string{
			"thing1": "举报申诉已提交",
		},
	})

	if notification.Status != "unread" || notification.WechatState != "pending" || notification.WechatTemplateID == "" || notification.WechatTaskID == 0 {
		t.Fatalf("expected pending wechat notification, got %+v", notification)
	}
	tasks := service.WechatTasks()
	if len(tasks) != 1 || tasks[0].NotificationID != notification.ID || tasks[0].Status != "pending" || tasks[0].RequestPayload["thing1"] == "" {
		t.Fatalf("expected pending wechat task, got %+v", tasks)
	}
}

func TestMarkReadAndWechatTaskSent(t *testing.T) {
	service := NewService()
	notification := service.Create(CreateRequest{UserID: 2, NotifyType: "report_handled", NeedWechat: true})

	if _, err := service.MarkRead(3, notification.ID); err != ErrForbidden {
		t.Fatalf("expected ErrForbidden for other user, got %v", err)
	}
	read, err := service.MarkRead(2, notification.ID)
	if err != nil {
		t.Fatal(err)
	}
	if read.Status != "read" || read.ReadAt == "" {
		t.Fatalf("expected read notification, got %+v", read)
	}

	task, err := service.MarkWechatTaskSent(notification.WechatTaskID, "0", "ok")
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != "sent" || task.SentAt == "" {
		t.Fatalf("expected sent task, got %+v", task)
	}
	items := service.List(2)
	if len(items) != 1 || items[0].WechatState != "sent" {
		t.Fatalf("expected notification wechat state sent, got %+v", items)
	}
}

func TestSendWechatTaskUsesSubscribeSender(t *testing.T) {
	sender := &recordingWechatSender{}
	service := NewService()
	service.UseOpenIDResolver(func(userID int64) (string, bool) {
		if userID == 2 {
			return "openid-2", true
		}
		return "", false
	})
	service.UseWechatSubscribeSender(sender)

	notification := service.Create(CreateRequest{
		UserID:     2,
		NotifyType: "review_remind",
		NeedWechat: true,
		WechatData: map[string]string{
			"thing1": "请评价",
			"page":   "pages/review/index",
		},
	})
	if sender.called {
		t.Fatal("expected create to only enqueue task")
	}
	task, err := service.SendWechatTask(notification.WechatTaskID)
	if err != nil {
		t.Fatal(err)
	}

	if !sender.called || sender.req.OpenID != "openid-2" || sender.req.TemplateID == "" || sender.req.Page != "pages/review/index" || sender.req.Data["thing1"] != "请评价" {
		t.Fatalf("expected subscribe sender called, sender=%+v task=%+v", sender, task)
	}
	items := service.List(2)
	if len(items) != 1 || items[0].WechatState != "sent" {
		t.Fatalf("expected sent notification after successful sender, got %+v", items)
	}
}

func TestSendWechatTaskRequiresOpenID(t *testing.T) {
	service := NewService()
	notification := service.Create(CreateRequest{UserID: 2, NotifyType: "review_remind", NeedWechat: true})

	if _, err := service.SendWechatTask(notification.WechatTaskID); err != ErrWechatSubscribeSend {
		t.Fatalf("expected ErrWechatSubscribeSend, got %v", err)
	}
	tasks := service.WechatTasks()
	if len(tasks) != 1 || tasks[0].Status != "pending" {
		t.Fatalf("expected pending task when send fails, got %+v", tasks)
	}
}

func TestSendPendingWechatTasksHonorsLimit(t *testing.T) {
	sender := &recordingWechatSender{}
	service := NewService()
	service.UseOpenIDResolver(func(userID int64) (string, bool) {
		return "openid", true
	})
	service.UseWechatSubscribeSender(sender)
	first := service.Create(CreateRequest{UserID: 1, NotifyType: "review_remind", NeedWechat: true})
	second := service.Create(CreateRequest{UserID: 2, NotifyType: "review_remind", NeedWechat: true})

	result := service.SendPendingWechatTasks(1)
	if result.Sent != 1 || result.Failed != 0 || len(result.Items) != 1 {
		t.Fatalf("unexpected batch result: %+v first=%+v second=%+v", result, first, second)
	}
	tasks := service.WechatTasks()
	sentCount := 0
	pendingCount := 0
	for _, task := range tasks {
		if task.Status == "sent" {
			sentCount++
		}
		if task.Status == "pending" {
			pendingCount++
		}
	}
	if sentCount != 1 || pendingCount != 1 {
		t.Fatalf("expected one sent and one pending task, got %+v", tasks)
	}
}

func TestCreateOrUpdateRoomMessageAggregatesUnreadRoomNotification(t *testing.T) {
	service := NewService()
	first := service.CreateOrUpdateRoomMessage(CreateRequest{
		UserID: 2, NotifyType: "im_message", Title: "收到局内消息", Content: "甲：第一条", BizType: "game", BizID: 9,
	})
	second := service.CreateOrUpdateRoomMessage(CreateRequest{
		UserID: 2, NotifyType: "im_message", Title: "收到局内消息", Content: "乙：第二条", BizType: "game", BizID: 9,
	})
	if first.ID == 0 || second.ID != first.ID {
		t.Fatalf("expected same unread room notification, first=%+v second=%+v", first, second)
	}
	items := service.List(2)
	if len(items) != 1 || items[0].Content != "乙：第二条" {
		t.Fatalf("expected latest content in one notification, got %+v", items)
	}
}

func TestServiceUsesRepositoryWhenConfigured(t *testing.T) {
	repo := &fakeNotificationRepository{
		notifications: map[int64]Notification{
			8: {
				ID:         8,
				UserID:     2,
				NotifyType: "report_created",
				Title:      "report",
				Status:     "unread",
				CreatedAt:  time.Now(),
			},
		},
		tasks: map[int64]WechatTask{
			6: {
				ID:             6,
				NotificationID: 8,
				UserID:         2,
				Scene:          "report_created",
				TemplateID:     "tpl",
				Status:         "pending",
				CreatedAt:      time.Now(),
			},
		},
		templates: []WechatTemplate{{Scene: "report_created", TemplateID: "tpl", Title: "report", Status: "active"}},
	}
	service := NewServiceWithRepository(repo)

	notification := service.Create(CreateRequest{UserID: 2, NotifyType: "report_created", NeedWechat: true})
	if !repo.saved || notification.ID != 9 || notification.WechatTaskID != 7 {
		t.Fatalf("expected repository create, saved=%v notification=%+v", repo.saved, notification)
	}
	if len(service.notifications) != 1 || service.notifications[9].ID != 9 || len(service.wechatTasks) != 1 || service.wechatTasks[7].ID != 7 {
		t.Fatalf("expected repository IDs to replace temporary in-memory IDs, notifications=%+v tasks=%+v", service.notifications, service.wechatTasks)
	}
	if items := service.List(2); !repo.listed || len(items) != 1 || items[0].ID != 8 {
		t.Fatalf("expected repository list, listed=%v items=%+v", repo.listed, items)
	}
	read, err := service.MarkRead(2, 8)
	if err != nil {
		t.Fatal(err)
	}
	if !repo.updated || read.Status != "read" || read.ReadAt == "" {
		t.Fatalf("expected repository mark read, updated=%v notification=%+v", repo.updated, read)
	}
	if tasks := service.WechatTasks(); !repo.listedTasks || len(tasks) != 1 || tasks[0].ID != 6 {
		t.Fatalf("expected repository tasks, listed=%v tasks=%+v", repo.listedTasks, tasks)
	}
	if templates := service.WechatTemplates(); !repo.listedTemplates || len(templates) != 1 || templates[0].TemplateID != "tpl" {
		t.Fatalf("expected repository templates, listed=%v templates=%+v", repo.listedTemplates, templates)
	}
	task, err := service.MarkWechatTaskSent(6, "0", "ok")
	if err != nil {
		t.Fatal(err)
	}
	if !repo.markedSent || task.Status != "sent" || task.SentAt == "" {
		t.Fatalf("expected repository mark sent, marked=%v task=%+v", repo.markedSent, task)
	}
}

func TestCreatePersistedReportsRepositoryFailureAndKeepsFallback(t *testing.T) {
	persistErr := errors.New("database unavailable")
	repo := &fakeNotificationRepository{
		notifications: map[int64]Notification{},
		tasks:         map[int64]WechatTask{},
		saveErr:       persistErr,
		listErr:       persistErr,
	}
	service := NewServiceWithRepository(repo)

	notification, err := service.CreatePersisted(CreateRequest{
		UserID: 2, NotifyType: "application_approved", Title: "入局申请已通过", NeedWechat: true,
	})
	if !errors.Is(err, ErrNotificationPersist) {
		t.Fatalf("expected persistence error, got %v", err)
	}
	if notification.ID == 0 || notification.WechatTaskID == 0 {
		t.Fatalf("expected in-memory fallback notification and task, got %+v", notification)
	}
	items := service.List(2)
	if len(items) != 1 || items[0].ID != notification.ID {
		t.Fatalf("expected fallback notification to remain readable while repository is unavailable, got %+v", items)
	}
}

func TestStrictNotificationListDoesNotReturnProcessFallbackOnRepositoryFailure(t *testing.T) {
	persistErr := errors.New("database unavailable")
	repo := &fakeNotificationRepository{
		notifications: map[int64]Notification{},
		tasks:         map[int64]WechatTask{},
		saveErr:       persistErr,
		listErr:       persistErr,
	}
	service := NewServiceWithRepository(repo)
	_, _ = service.CreatePersisted(CreateRequest{UserID: 2, NotifyType: "report_created", Title: "举报已提交"})

	items, err := service.ListStrict(2)
	if !errors.Is(err, persistErr) || items != nil {
		t.Fatalf("expected strict repository error without local fallback, items=%+v err=%v", items, err)
	}
}

func TestStrictWechatBatchDoesNotSendProcessFallbackTasksOnRepositoryFailure(t *testing.T) {
	listErr := errors.New("task database unavailable")
	repo := &fakeNotificationRepository{
		notifications: map[int64]Notification{},
		tasks:         map[int64]WechatTask{},
		taskListErr:   listErr,
	}
	service := NewServiceWithRepository(repo)
	service.wechatTasks[1] = WechatTask{ID: 1, UserID: 2, Status: "pending", TemplateID: "local-template"}

	result, err := service.SendPendingWechatTasksStrict(20)
	if !errors.Is(err, listErr) || result.Sent != 0 || result.Failed != 0 {
		t.Fatalf("expected strict task read failure without sending local tasks, result=%+v err=%v", result, err)
	}
}

type recordingWechatSender struct {
	called bool
	req    WechatSubscribeSendRequest
}

func (s *recordingWechatSender) Send(ctx context.Context, req WechatSubscribeSendRequest) (WechatSubscribeSendResult, error) {
	s.called = true
	s.req = req
	return WechatSubscribeSendResult{Code: "0", Message: "ok"}, nil
}

type fakeNotificationRepository struct {
	notifications   map[int64]Notification
	tasks           map[int64]WechatTask
	templates       []WechatTemplate
	saved           bool
	listed          bool
	listedTasks     bool
	listedTemplates bool
	updated         bool
	markedSent      bool
	saveErr         error
	listErr         error
	taskListErr     error
}

func (r *fakeNotificationRepository) SaveNotification(ctx context.Context, notification Notification, task *WechatTask) (Notification, error) {
	r.saved = true
	if r.saveErr != nil {
		return Notification{}, r.saveErr
	}
	notification.ID = 9
	if task != nil {
		notification.WechatTaskID = 7
		task.ID = 7
		task.NotificationID = notification.ID
		r.tasks[task.ID] = *task
	}
	r.notifications[notification.ID] = notification
	return notification, nil
}

func (r *fakeNotificationRepository) ListNotifications(ctx context.Context, userID int64) ([]Notification, error) {
	r.listed = true
	if r.listErr != nil {
		return nil, r.listErr
	}
	return []Notification{r.notifications[8]}, nil
}

func (r *fakeNotificationRepository) FindNotification(ctx context.Context, notificationID int64) (Notification, bool, error) {
	notification, ok := r.notifications[notificationID]
	return notification, ok, nil
}

func (r *fakeNotificationRepository) UpdateNotification(ctx context.Context, notification Notification) (Notification, error) {
	r.updated = true
	r.notifications[notification.ID] = notification
	return notification, nil
}

func (r *fakeNotificationRepository) ListWechatTasks(ctx context.Context) ([]WechatTask, error) {
	r.listedTasks = true
	if r.taskListErr != nil {
		return nil, r.taskListErr
	}
	return []WechatTask{r.tasks[6]}, nil
}

func (r *fakeNotificationRepository) ListWechatTemplates(ctx context.Context) ([]WechatTemplate, error) {
	r.listedTemplates = true
	return r.templates, nil
}

func (r *fakeNotificationRepository) MarkWechatTaskSent(ctx context.Context, taskID int64, resultCode string, resultMessage string, sentAt string) (WechatTask, error) {
	r.markedSent = true
	task := r.tasks[taskID]
	task.Status = "sent"
	task.ResultCode = resultCode
	task.ResultMessage = resultMessage
	task.SentAt = sentAt
	r.tasks[task.ID] = task
	return task, nil
}
