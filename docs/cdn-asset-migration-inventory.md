# 小程序 CDN 资源迁移清单

> 盘点日期：2026-07-12  
> 有效项目根目录：`E:\项目\03周总\03.预览版修改\code\ZHW-full-project-20260712`  
> 有效小程序源码：`miniprogram-client`  
> CDN 待发布目录：`cdn-assets/miniprogram`  
> CDN 访问前缀：`https://static.haowan.net.cn/miniprogram/`

## 1. 使用说明

本清单用于后续分批迁移图标和图片资源。当前仅完成盘点和分类，没有修改前端、后端或资源文件。

执行迁移时必须遵守：

1. 没有进入本清单的资源不自动修改。
2. `KEEP_LOCAL` 资源不上传、不替换。
3. `CDN_STATIC` 资源必须先上传并验证成功，再修改引用。
4. `BACKEND_ASSET` 资源必须同时检查接口字段和小程序消费位置。
5. `DYNAMIC_UPLOAD` 资源继续走文件上传接口，不进入静态资源迁移。
6. `REVIEW` 资源在路径确认前不得自动处理。
7. 每批迁移完成后保留原文件，回归通过后再单独决定是否清理。

## 2. 状态定义

| 标记 | 含义 | 后续动作 |
| --- | --- | --- |
| `CDN_DONE` | 已经使用 CDN 地址 | 保持现状，只做可访问性检查 |
| `KEEP_LOCAL` | 基础、首屏或弱网关键资源 | 保留本地路径 |
| `CDN_STATIC` | 普通静态业务资源 | 加入待上传目录并替换为统一 CDN 地址 |
| `BACKEND_ASSET` | Go 接口返回的小程序资源路径 | 改为资源键或统一生成 CDN URL |
| `DYNAMIC_UPLOAD` | 用户产生的动态文件 | 继续使用 `/api/app/files/upload-token` |
| `REVIEW` | 路径异常或用途不明确 | 人工确认后再重新分类 |

## 3. 当前盘点基线

| 项目 | 数量 | 结果 |
| --- | ---: | --- |
| 前端 CDN 引用次数 | 109 | 分布在 37 个文件 |
| 前端唯一 CDN 资源 | 80 | 80 个均能在 `cdn-assets/miniprogram` 找到对应文件 |
| 前端本地引用次数 | 79 | 分布在 28 个文件 |
| 前端唯一可归一化本地资源 | 73 | 70 个文件存在，3 个相对路径待核查 |
| 后端本地资源引用次数 | 59 | 主要集中在 8 个 handler 文件 |
| 后端唯一小程序本地资源 | 52 | 12 个已有 CDN 待发布副本，40 个尚未进入待发布目录 |
| 同一归一化资源同时使用本地和 CDN | 0 | 当前混用发生在不同资源之间 |

说明：该统计只覆盖 `.js`、`.wxml`、`.json`、`.wxss`、`.go` 和 `.sql` 中可识别的图片路径；使用字符串拼接生成的路径需在迁移批次中再次扫描。

## 4. A 类：CDN_DONE

当前 80 个唯一 CDN 地址全部有对应的待发布源文件，暂不改动。主要目录包括：

- `assets/game/`
- `assets/game-card/`
- `assets/map/`
- `components/game-card/assets/`
- `components/home-shell/assets/`
- `components/participant-card/assets/`
- `pages/game/*/assets/`
- `pages/home/player/assets/`
- `pages/map/my-city/assets/`
- `pages/message/assets/`
- `pages/profile/member/assets/`
- `pages/relation/network/assets/`

后续检查项：

- [ ] 检查 80 个 CDN URL 均返回 HTTP 200。
- [ ] 检查响应 `Content-Type` 与文件类型匹配。
- [ ] 检查微信公众平台已配置 `static.haowan.net.cn` 合法域名。
- [ ] 检查 COS/CDN 缓存规则。

## 5. B 类：KEEP_LOCAL

以下资源优先保留本地，避免入口首屏、基础交互或弱网场景依赖 CDN：

| 资源或目录 | 原因 |
| --- | --- |
| `pages/entry/assets/entry-brand-wordmark.svg` | 入口首屏品牌资源 |
| `pages/entry/assets/entry-ticket-icon.svg` | 入口首屏关键资源 |
| `pages/entry/assets/entry-title-mark.svg` | 入口首屏关键资源 |
| `pages/login/assets/icon-check.svg` | 登录基础交互资源 |
| `components/chat-input-bar/assets/add.svg` | 聊天输入基础操作，已按参考工程替换为 SVG |
| `components/chat-input-bar/assets/emoji.svg` | 聊天输入基础操作，已按参考工程替换为 SVG |
| `components/chat-input-bar/assets/voice.svg` | 聊天输入基础操作，已按参考工程替换为 SVG |
| 返回、关闭、勾选、右箭头等纯基础控件图标 | 体积小、调用频繁、网络收益低 |
| 后续新增的断网和加载失败占位图 | 必须在 CDN 不可用时显示 |

复核项：

- [ ] 最终迁移前确认是否新增原生 `tabBar`；如有，其图标继续保留本地。
- [ ] 建立明确的本地兜底图片，不使用业务图片作为兜底。

## 6. C 类：CDN_STATIC

以下本地资源适合分批迁移。当前均未进入 `cdn-assets/miniprogram`，因此上传完成前不得替换引用。

注意：2026-07-12 已按参考工程完成一批 PNG/CSS 图标到本地 SVG 的替换，具体项目见第 13 节。第 13 节标记为 `SVG_LOCAL_DONE` 的资源不再按本节旧 PNG 路径迁移；后续是否上 CDN 需另行决定。

### C1. 游戏模块

- `components/game-card/assets/action-greet.png`
- `components/game-info-card/assets/chevron-right.png`
- `components/game-info-card/assets/header-info.png`
- `components/game-info-card/assets/mini-program.png`
- `components/game-info-card/assets/title-info.png`
- `pages/game/assets/icons/icon-location-pin.svg`
- `pages/game/detail/assets/i41@3x.png`
- `pages/game/detail/assets/i45@3x.png`
- `pages/game/detail/assets/i46@3x.png`
- `pages/game/detail/assets/i47@3x.png`
- `pages/game/detail/assets/i48@3x.png`
- `pages/game/detail/assets/icon-calendar.png`
- `pages/game/guide-chat/assets/icon-calendar.png`
- `pages/game/guide-chat/assets/icon-invite.png`
- `pages/game/guide-chat/assets/icon-location.png`
- `pages/game/guide-progress/assets/chevron-right.png`
- `pages/game/guide-progress/assets/notice-alert.png`
- `pages/game/guide-progress/assets/result-canceled.png`
- `pages/game/guide-progress/assets/result-success.png`
- `pages/game/guide-progress/assets/section-active.png`
- `pages/game/guide-progress/assets/status-confirmed.png`
- `pages/game/hall/assets/category-growth.png`
- `pages/game/hall/assets/category-more.png`
- `pages/game/manage/assets/icon-weixin-contact.png`
- `pages/game/referral-record/assets/chevron.png`
- `pages/game/share/assets/share-private-message.png`

### C2. 地图、消息和 IM

- `pages/im/room/assets/icon-handshake.svg`
- `pages/im/room/assets/icon-location.svg`
- `pages/im/room/assets/icon-time.svg`
- `pages/map/blind-route/assets/i50.png`
- `pages/map/blind-route/assets/i52.png`
- `pages/map/blind-route/assets/i54.png`
- `pages/message/assets/i57@3x.png`
- `pages/message/assets/i61@3x.png`
- `pages/message/assets/i62@3x.png`

### C3. 个人中心和资产模块

- `pages/profile/assets/i66@3x.png` 至 `pages/profile/assets/i85@3x.png`
- `pages/profile/assets/icon-review-like.svg`
- `pages/profile/assets/icon-review-reply.svg`
- `pages/profile/asset-center/manage/assets/fa/bag-shopping.svg`
- `pages/profile/asset-center/manage/assets/fa/check.svg`
- `pages/profile/asset-center/manage/assets/fa/circle-question.svg`
- `pages/profile/asset-center/manage/assets/fa/credit-card.svg`
- `pages/profile/asset-center/manage/assets/fa/download.svg`
- `pages/profile/asset-center/manage/assets/fa/hourglass-half.svg`
- `pages/profile/asset-center/manage/assets/fa/list-ul.svg`
- `pages/profile/asset-center/manage/assets/fa/plus.svg`
- `pages/profile/asset-center/manage/assets/fa/rotate-left.svg`
- `pages/profile/asset-center/manage/assets/fa/spinner.svg`
- `pages/profile/asset-center/manage/assets/fa/star.svg`
- `pages/profile/asset-center/points/assets/icon-chevron-left.svg`
- `pages/profile/footprint/achievements/assets/icon-crown.png`
- `pages/profile/footprint/achievements/assets/icon-star.png`
- `pages/profile/service-center/invite/records/assets/warning-red.svg`
- `pages/profile/system-management/report-center/assets/icon-check.svg`

### C4. 自定义首页壳层

以下资源先放在 `CDN_STATIC` 候选中，迁移前需检查它们是否属于首屏关键资源：

- `components/home-shell/assets/master-comment.svg`
- `components/home-shell/assets/master-search.svg`

## 7. D 类：BACKEND_ASSET

后端不应长期返回与小程序包目录耦合的 `/pages/...` 路径。主要涉及：

| 后端文件 | 当前唯一资源规模 | 后续动作 |
| --- | ---: | --- |
| `services/go-api/internal/appapi/profile_home_handler.go` | 约 20 个 | 个人中心入口图标改为资源键或 CDN URL |
| `services/go-api/internal/appapi/profile_assets_handler.go` | 约 10 个 | 资产模块图标统一处理 |
| `services/go-api/internal/appapi/game_handler.go` | 约 9 个 | 游戏详情和交付图标统一处理 |
| `services/go-api/internal/appapi/game_invite_handler.go` | 约 8 个 | 邀请和进度图标统一处理 |
| `services/go-api/internal/appapi/notification_handler.go` | 约 5 个 | 消息分类图标统一处理 |
| `services/go-api/internal/appapi/map_handler.go` | 约 3 个 | 地图盲盒资源统一处理 |
| `services/go-api/internal/appapi/home_payload.go` | 约 2 个 | 封面路径核查并统一处理 |
| `services/go-api/internal/appapi/review_handler.go` | 约 1 个 | 成就图标统一处理 |
| `services/go-api/internal/appapi/map_my_city_handler.go` | 约 1 个 | 城市故事封面统一处理 |

推荐接口契约：

```json
{
  "iconKey": "profile.asset.balance"
}
```

小程序使用统一资源表解析 `iconKey`。如果最终决定由后端返回完整地址，则必须通过 `STORAGE_DOWNLOAD_BASE_URL + storage_key` 生成，不允许 handler 内手写 CDN 域名。

需要特别核查：

- `components/game-card/assets/cover-sunset.png`：后端引用路径对应的前端文件不存在，实际资源可能位于 `assets/game-card/cover-sunset.png`。
- 数据库种子和 mock 数据必须与正式 handler 使用同一资源键。

## 8. E 类：DYNAMIC_UPLOAD

以下资源不参与静态 CDN 清单：

- 用户头像
- 聊天图片和附件
- 实名认证材料
- 举报、反馈等用户上传凭证
- 其他运行时产生的文件

保持现有流程：

```text
小程序申请 /api/app/files/upload-token
→ 后端生成 COS 表单签名
→ 小程序直传 COS
→ 后端保存 storage_key
→ 使用 STORAGE_DOWNLOAD_BASE_URL 生成下载地址
```

COS Secret ID 和 Secret Key 只能保存在服务端环境变量中。

## 9. F 类：REVIEW

以下 3 个 mock 相对路径按引用文件目录解析时找不到文件：

- `miniprogram-client/api/mock.js` 中的 `./assets/follow-deal.png`
- `miniprogram-client/api/mock.js` 中的 `./assets/follow-feedback.png`
- `miniprogram-client/api/mock.js` 中的 `./assets/follow-schedule.png`

预期资源可能位于 `pages/game/success-guide/assets/`。处理前必须核对实际消费页面，不能直接按字符串替换。

## 10. 建议迁移批次

| 批次 | 范围 | 风险 | 当前状态 |
| --- | --- | --- | --- |
| 0 | 建立资源配置、资源键和检查脚本 | 低 | `PENDING` |
| 1 | 游戏模块普通业务资源 | 中 | `PENDING` |
| 2 | 地图、消息和 IM 资源 | 中 | `PENDING` |
| 3 | 个人中心和资产模块资源 | 中 | `PENDING` |
| 4 | 后端 `BACKEND_ASSET` 解耦 | 高 | `PENDING` |
| 5 | mock、SQL 种子及异常路径收口 | 中 | `PENDING` |
| 6 | 全量回归和冗余资源评估 | 高 | `PENDING` |

每批执行顺序：

```text
确认清单
→ 复制到 cdn-assets/miniprogram
→ 上传 COS
→ 验证 CDN URL
→ 修改该批引用
→ 静态扫描
→ 微信开发者工具回归
→ 更新清单状态
```

## 11. 验收清单

- [ ] 所有 `CDN_DONE` 和新迁移 CDN URL 可访问。
- [ ] 所有 `KEEP_LOCAL` 文件仍在小程序包中。
- [ ] 前端代码不再散落新增的完整 CDN 域名。
- [ ] Go handler 不再新增 `/pages/...` 图片路径。
- [ ] 后端、mock 和 SQL 种子使用一致的资源键或地址策略。
- [ ] 首页、登录、游戏、地图、消息、IM、个人中心完成真机或开发者工具回归。
- [ ] 弱网或 CDN 图片失败时，关键页面仍可操作。
- [ ] 未经确认不删除任何原始资源。

## 12. 本次记录

- [x] 完成只读盘点。
- [x] 完成初步分类。
- [x] 确认现有 80 个 CDN 资源均有待发布源文件。
- [x] 初次盘点阶段只新增清单文档，未修改前端、后端或图片资源。
- [ ] 等待确认后开始迁移批次 0。

## 13. SVG 参考替换记录（2026-07-12）

参考工程：`E:\项目\03周总\02.前端后台联调\05周总02\ZHW-HZ-Z`

本批只读取参考工程，将已确认的 SVG 复制到当前唯一有效项目，并替换同语义旧图标；没有修改参考工程，也没有上传 CDN。

### 13.1 SVG_LOCAL_DONE

- [x] 聊天输入栏：`voice.png`、`emoji.png`、`add.png` → 对应 SVG。
- [x] 游戏详情操作栏：拒绝、确认 PNG → `icon-close.svg`、`icon-confirm-green.svg`。
- [x] 成局双方确认标记：CSS 自绘勾选 → `icon-confirmed.svg`。
- [x] 游戏交付保障图标：`icon-shield.png` → `shield-service.svg`。
- [x] 游戏交付联系玩家快捷图标：统一使用 `contact-person.svg`。
- [x] 带局信息：时间、地点 PNG → `icon-time.svg`、`icon-location.svg`。
- [x] 带局进度：等待状态 PNG → `participant-waiting.svg`。
- [x] 行家业务管理：完成、取消、联系、握手、评价图标 → 对应 SVG。
- [x] 玩家组局管理：完成、取消、联系、警告及状态图标 → 对应 SVG。
- [x] 转介绍记录：处理中、完成、取消、握手、提醒、聊天和警告图标 → 对应 SVG。
- [x] 补齐系统管理页面代码已引用但当前目录缺失的 SVG 文件。

### 13.2 本批素材结果

- 当前项目 SVG 文件总数由 91 增至 153。
- 从参考工程补入 62 个当前项目缺失的业务 SVG。
- 参考工程 `pages`、`components`、`assets`、`styles`、`utils` 范围内的业务 SVG 已无缺失同路径文件。
- 现有同路径 SVG 不做批量覆盖，避免覆盖当前项目已经调整过的版本。

### 13.3 验证状态

- [x] 修改的 JavaScript 文件通过 `node --check`。
- [x] 本批新增的绝对本地 SVG 引用均能找到对应文件。
- [x] 页面事件处理器检查：`missing_event_handlers=0`。
- [x] 前端 API 与后端路由检查：`missing_backend_routes=0`。
- [ ] 项目现有路由注册检查仍报告 `missing_registered_pages=55`；属于当前 `app.json` 注册基线问题，本批未改路由。
- [ ] 微信开发者工具编译和页面视觉回归。
