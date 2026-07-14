# TC-144 AI 数据准备开关关闭时禁止导出 IM 给 AI

## 用例目标

验证一期 AI 数据准备只允许后台查看数据沉淀统计，不允许在默认关闭状态下导出 IM 聊天内容给 AI。

## 前置条件

- 已完成至少 1 个测试局的创建、审核、入局、手动开始、IM 发送、服务确认和评价。
- 已写入至少 1 条行为日志、1 条收藏、1 条评价、1 条用户足迹、1 条人脉关系、1 条行家画像、1 条领路人画像和 1 条 IM 消息。
- 后台管理员具备 `ai:data:read` 权限。
- 后台管理员具备 `ai:data:seed` 权限。
- 后台管理员具备 `ai:data:export` 权限。
- AI IM 导出开关保持默认关闭。

## 测试步骤

1. 使用带 `ai:data:seed` 权限的后台管理员访问 `POST /api/admin/ai-data/acceptance-fixture`。
2. 检查接口返回：
   - `usersVerified = 5`
   - `gamesCreated = 3`
   - `behaviorLogsAdded = 20`
   - `favoritesAdded = 5`
   - `reviewsAdded = 5`
   - `acceptanceSnapshot.acceptanceReady = true`
3. 重复访问 `POST /api/admin/ai-data/acceptance-fixture`，检查接口仍返回 HTTP `200`，且不继续追加已达标数据。
4. 使用带 `ai:data:read` 权限的后台管理员访问 `GET /api/admin/ai-data/snapshot`。
5. 检查返回字段：
   - `behaviorLogCount > 0`
   - `favoriteCount > 0`
   - `reviewCount > 0`
   - `footprintCount > 0`
   - `connectionCount > 0`
   - `expertProfileCount > 0`
   - `guideProfileCount > 0`
   - `imMessageCount > 0`
   - `dataReady = true`
   - `userCount > 0`
   - `gameCount > 0`
   - `acceptanceChecks.users.required = 5`
   - `acceptanceChecks.games.required = 3`
   - `acceptanceChecks.behaviorLogs.required = 20`
   - `acceptanceChecks.favorites.required = 5`
   - `acceptanceChecks.reviews.required = 5`
   - `acceptanceReady` 按上述 5 项阈值自动计算
   - `imExportEnabled = false`
6. 使用带 `ai:data:export` 权限的后台管理员访问 `POST /api/admin/ai-data/im-export`。
7. 检查接口返回 HTTP `403`，业务错误码为 `40361`。

## 预期结果

- 后台可以看到 AI 数据准备快照。
- 后台可以直接看到 E5 验收阈值当前数量、要求数量和是否达标。
- 后台可用受控夹具一键补齐 5 用户、3 局、20 行为、5 收藏、5 评价的验收数据。
- 快照只返回 IM 消息数量，不返回聊天内容。
- 默认关闭开关时，后台即使具备 `ai:data:export` 权限，也不能导出 IM 聊天内容。
- 禁止导出时返回 `40361`，用于区分普通权限不足和业务开关关闭。

## 自动化覆盖

- `services/go-api/internal/appapi/server_test.go`
- `TestAIDataSnapshotAndIMExportGateHTTP`
- `TestAIDataAcceptanceFixtureHTTP`

## 当前状态

- 已通过 `go test -count=1 ./...`。
