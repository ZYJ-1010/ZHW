package identity

import (
	"context"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentsms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

// TencentSMSConfig contains only runtime configuration. Its values must come
// from the deployment environment and must never be persisted in application
// data, logs, responses or source control.
type TencentSMSConfig struct {
	SecretID   string
	SecretKey  string
	SDKAppID   string
	SignName   string
	TemplateID string
	Region     string
}

// TencentSMSSender sends mainland-China verification messages through the
// official Tencent Cloud SMS API. The caller supplies a decrypted phone number
// and this sender converts it to the E.164 format required by Tencent Cloud.
type TencentSMSSender struct {
	config TencentSMSConfig
}

func NewTencentSMSSender(config TencentSMSConfig) *TencentSMSSender {
	config.SecretID = strings.TrimSpace(config.SecretID)
	config.SecretKey = strings.TrimSpace(config.SecretKey)
	config.SDKAppID = strings.TrimSpace(config.SDKAppID)
	config.SignName = strings.TrimSpace(config.SignName)
	config.TemplateID = strings.TrimSpace(config.TemplateID)
	config.Region = strings.TrimSpace(config.Region)
	if config.Region == "" {
		config.Region = "ap-guangzhou"
	}
	return &TencentSMSSender{config: config}
}

func (s *TencentSMSSender) GenerateCode() string {
	return generateSMSCode()
}

func (s *TencentSMSSender) Send(ctx context.Context, req SMSDispatchRequest) (SMSDispatchResult, error) {
	if s == nil || !validMainlandPhone(req.Phone) || s.config.SecretID == "" || s.config.SecretKey == "" || s.config.SDKAppID == "" || s.config.SignName == "" || s.config.TemplateID == "" {
		return SMSDispatchResult{}, ErrSMSSendFailed
	}

	credential := common.NewCredential(s.config.SecretID, s.config.SecretKey)
	clientProfile := profile.NewClientProfile()
	client, err := tencentsms.NewClient(credential, s.config.Region, clientProfile)
	if err != nil {
		return SMSDispatchResult{}, ErrSMSSendFailed
	}
	request := tencentsms.NewSendSmsRequest()
	request.PhoneNumberSet = common.StringPtrs([]string{"+86" + req.Phone})
	request.SmsSdkAppId = common.StringPtr(s.config.SDKAppID)
	request.SignName = common.StringPtr(s.config.SignName)
	request.TemplateId = common.StringPtr(s.config.TemplateID)
	request.TemplateParamSet = common.StringPtrs([]string{req.Code})
	request.SessionContext = common.StringPtr("zhw-identity-" + strings.TrimSpace(req.Scene))

	response, err := client.SendSmsWithContext(ctx, request)
	if err != nil || response == nil || response.Response == nil || len(response.Response.SendStatusSet) != 1 {
		return SMSDispatchResult{}, ErrSMSSendFailed
	}
	status := response.Response.SendStatusSet[0]
	if status == nil || status.Code == nil || *status.Code != "Ok" {
		return SMSDispatchResult{}, ErrSMSSendFailed
	}
	messageID := ""
	if status.SerialNo != nil {
		messageID = strings.TrimSpace(*status.SerialNo)
	}
	return SMSDispatchResult{Provider: "tencent_sms", MessageID: messageID}, nil
}
