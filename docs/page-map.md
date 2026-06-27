# 墨刀页面树与小程序路由映射

本文是工程实现文档，只记录当前根据墨刀可见页面树整理出的开发映射。

## 第 1 批：入口闭环

| 业务域 | 墨刀页面 | 小程序路由 |
| --- | --- | --- |
| 邀请注册与登录 | 进入页面、登录前首页、用户注册登录、新用户邀请注册登录、微信授权登录（登录页内部流程）、找回密码 | `pages/entry/index`、`pages/home/guest/index`、`pages/login/index`、`pages/login/invite/index`、`pages/login/forgot/index` |
| 首页与角色入口 | 登录后玩家首页、行家首页、领路人首页、权限提示、权益对比 | `pages/home/index`、`pages/home/player/index`、`pages/home/expert/index`、`pages/home/guide/index` |
| 我的基础 | 我的 | `pages/profile/index` |

目标：用户能进入、识别身份、看到角色和下一步动作。

## 第 2 批：角色与组局基础

| 业务域 | 墨刀页面 | 小程序路由 |
| --- | --- | --- |
| 角色申请 | 申请行家操作页、申请领路人操作页 | `pages/role/apply/index` |
| 审核状态 | 审核进度、审核通过、审核驳回、申请状态提示 | `pages/role/status/index` |
| 局列表 | 局前大厅 | `pages/game/hall/index` |
| 发起组局 | 发起组局 | `pages/game/create/index` |
| 局详情 | 详情、申请入口、邀约入口 | `pages/game/detail/index` |

目标：能浏览局、创建局、提交审核、看到状态。

## 第 3 批：入局与交付闭环

| 业务域 | 墨刀页面 | 小程序路由 |
| --- | --- | --- |
| 入局申请 | 玩家自申请入局 | `pages/game/applications/index` |
| 审核申请 | 行家审核列表、行家审核详情 | `pages/game/audit/index`、`pages/game/applications/index` |
| 领路人邀约 | 领路人发起引荐页、玩家被邀约确认页、领路人接收页 | `pages/game/applications/index` |
| 交付确认 | 交付操作页、交付确认页 | `pages/game/delivery/index` |
| 评价 | 评价页面 | `pages/game/review/index` |

目标：完成申请 / 邀约 -> 审核 -> 成员 -> 交付 -> 评价。

## 第 4 批：地图、IM、成长扩展

| 业务域 | 墨刀页面 | 小程序路由 |
| --- | --- | --- |
| 地图与城市探索 | 地图首页、组局分布地球网、组局盲盒、城市图鉴、足迹热力图、实景打卡、好友点亮城市、我的城市故事 | `pages/map/index`、`pages/map/blind-route/index`、`pages/map/city-atlas/index`、`pages/map/real-checkin/index`、`pages/map/footprint-heatmap/index`、`pages/map/friend-city/index`、`pages/map/my-city/index` |
| 消息 | 消息 | `pages/message/index` |
| 关系 | 关系网首页 | `pages/relation/network/index` |
| 局内 IM | 局内文字、图片、文件消息 | `pages/im/room/index` |
| 会员与成长 | 会员中心、资产管理、积分中心、成就页、我的足迹-我的成就墙 | `pages/profile/member/index`、`pages/profile/asset-center/manage/index`、`pages/profile/asset-center/points/index`、`pages/profile/achievements/index`、`pages/profile/footprint/achievements/index` |
| 系统管理 | 我的资料、技能配置 | `pages/profile/system-management/profile-info/index`、`pages/profile/system-management/skill-config/index` |
| 预留 | 元宇宙、元宇宙管理中心、母版 | `pages/placeholder/metaverse/index` |

目标：补齐一期验收入口和二期扩展基础。

## 当前说明

墨刀完整 50 个画布尚未全部导出，当前映射基于已读取到的页面树。当前普通登录先按“第一个用户不需要邀请码”处理；默认启动链路为 `pages/entry/index` -> `pages/home/guest/index` -> `pages/login/index`。邀请码注册已拆到 `pages/login/invite/index`，作为分享/扫码进入后的单独链路。后续拿到完整截图或标注后，只在本文和 `config/page-map.js` 中补齐，不改 `00.文档`。
