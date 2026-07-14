# 2026-06-29 消息中心动作链路对接进度

## 本轮完成

- `GET /api/app/notifications` 的通知卡片动作已从“仅邀请可处理”扩展为多类型动作：
  - `review_remind`：返回立即评价与查看组局动作。
  - `progress_feedback_remind` / `trade_warning`：返回立即处理与联系发起人动作。
  - `game_member_quit` / `game_quit_result`：返回查看组局与联系发起人动作。
  - `report_created` / `report_handled` / `report_assigned` / `report_closed`：返回举报/申诉详情动作。
- `POST /api/app/notifications/{id}/actions` 已支持非邀请通知动作：
  - `detail / view / open`
  - `process / handle / deliver`
  - `route / navigation`
  - `contact / chat`
  - `delay`
- 后端动作会返回统一 `target`：
  - `target.route`
  - `target.routeKey`
  - `target.bizType`
  - `target.bizId`
- 小程序 `pages/message/index` 已根据后端动作返回的 `target.route` 直接跳转。
- 小程序消息中心的 `linkText` 按钮已修正为 `data-action="process"`，不再把消息 ID 当作动作名。
- 新增 HTTP 测试 `TestNotificationActionsReturnConcreteTargetsHTTP`，覆盖：
  - 评价提醒跳转评价页。
  - 路线动作跳转地图页。
  - 联系动作跳转 IM 房间。
  - 举报处理通知跳转举报记录详情页。
  - 不支持的 action 返回 422。

## 验证结果

- `go test -count=1 ./...` 通过。
- `Get-ChildItem -Recurse -LiteralPath miniprogram-client -Filter *.js | ForEach-Object { node --check $_.FullName }` 通过。
- `rg -n "resolveNotificationAction|notificationProcessTarget|notificationRouteTarget|notificationContactTarget|TestNotificationActionsReturnConcreteTargetsHTTP|navigateByActionResult" services/go-api/internal/appapi miniprogram-client/pages/message` 已确认代码在位。
- `rg -n "ponytail:" services miniprogram-client docs -g "!*node_modules*" -g "!*dist*" -g "!*.map"` 未发现本轮新增债务标记。

## 当前进度估算

- 小程序后端业务功能：约 `99.9%`。本地业务链路基本完整，剩余重点仍是线上真实第三方环境联调。
- 小程序前后端全链路对接：约 `92%`。本轮从约 `90%` 推进到约 `92%`，主要提升来自消息中心动作闭环。
- 后台后端业务功能：约 `99.2%`。后续重点仍是后台 Web 页面字段核对和 E2E。

## 下一步建议

1. 继续补系统管理剩余细分页接口：屏蔽设置、反馈、举报、协议签署。
2. 核对后台 Web 页面与后端接口字段，特别是审核、导出、审计记录、邀请码批量生成。
3. 对小程序消息详情页做一次真实参数联调，确认 `notificationId / gameId / reportId` 在页面内都能正确读取。
