# P0 行家申请配置接口补齐 - 2026-06-30

## 范围口径

- 元宇宙功能不纳入当前交付和验收范围。
- AI 二期能力尽量推进，但不抢一期 P0 主链路。
- 本轮优先处理审查账本指出的真实断链接口：`GET /api/app/role-applications/expert/config`。

## 本轮处理

- `services/go-api/internal/appapi/profile_handler.go`
  - 新增 `expertApplyConfig`。
  - 配置优先读取 `system_configs.role.expert_apply_config`。
  - 未配置时保留后端默认配置，避免本地测试环境断链。

- `services/go-api/internal/appapi/server.go`
  - 注册 `GET /api/app/role-applications/expert/config`。

- `db/seeds/system_configs.sql`
  - 新增 `role.expert_apply_config` 测试配置数据。
  - 包含技能选项、字段配置、上传字段、校验规则、从业年限、服务数量和价格提示。

- `docs/openapi/app.openapi.yaml`
  - 补充 `GET /api/app/role-applications/expert/config`。

- `services/go-api/internal/appapi/server_test.go`
  - 新增 `TestExpertApplyConfigReadsSystemConfigHTTP`，验证接口读取系统配置。

## 验证

- `go test -count=1 ./internal/appapi -run "TestExpertApplyConfigReadsSystemConfigHTTP"`
- `go test -count=1 ./internal/appapi -run "TestRoleApplication|TestExpertApplyConfig"`
- `node --check miniprogram-client/pages/home/index.js`
- `node --check miniprogram-client/services/role.js`
