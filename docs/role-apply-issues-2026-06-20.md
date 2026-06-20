# 角色申请问题记录 - 2026-06-20

本文记录“首页-其他”角色申请流程当前发现的待确认问题。此文档只做问题记录，不代表最终产品方案。

## 1. 申请入口 / 申请动作重复

- 模块：`首页 / 角色申请`
- 涉及页面：
  - `pages/home-other/index`
  - `pages/home/player/index`
  - `pages/role/apply/index`
  - `pages/role/status/index`
- 涉及接口：
  - `POST /api/app/role-applications`
  - `GET /api/app/role-applications/my`
  - `GET /api/app/roles/my`
- 当前情况：
  - `pages/home-other/index` 新页面中有申请领路人浏览页和申请领路人内页提交动作。
  - 项目原有 `pages/role/apply/index` 也承载角色申请入口。
  - 当前临时实现为：`pages/home-other/index` 提交 `roleType: guide` 后 toast 提示并回玩家首页，`pages/home/player/index` 根据 pending 状态展示审核中。
- 待确认问题：
  - 玩家从首页点击申请角色时，正式入口到底进入 `pages/home-other/index` 里的申请页，还是进入原有 `pages/role/apply/index`？
  - 是否需要同时保留两个申请入口？
  - 若用户重复点击提交申请，前端和后端分别如何处理？
  - 行家申请和领路人申请是否走同一个入口，只靠 `roleType` 区分？
- 来源：用户 2026-06-20 反馈“这里有两个申请这是有问题的”。
- 状态：待业主确认。

## 2. 行家计划书 / 领路计划书填写规则不明确

- 模块：`首页 / 角色申请`
- 涉及页面 / 文档：
  - `pages/home-other/index`
  - `pages/home/index`
  - `pages/role/status/index`
  - `docs/api-additions.md`
- 涉及接口：
  - `POST /api/app/role-applications`
  - `GET /api/app/role-applications/my`
  - 如后续需要动态配置申请表单，还需确认是否新增配置接口。
- 当前情况：
  - `pages/home-other/index` 的领路人内页目前参考申请行家内页样式制作。
  - 当前领路人内页包含：领路计划书、资质证明、3 组业务定价 / 成本。
  - `pages/home/index` 中已有行家计划书相关展示。
  - `pages/role/status/index` 中已有领路计划书审核状态和驳回建议文案。
- 待确认问题：
  - 行家计划书具体需要填写哪些字段？
  - 领路计划书具体需要填写哪些字段？
  - 行家计划书和领路计划书是否共用同一套表单结构？
  - 领路计划书是否也需要 3 组业务定价 / 服务成本？
  - 资质证明要上传哪些材料，是否行家和领路人一致？
  - 提交接口字段如何组织，是否需要后端返回表单配置和校验规则？
- 来源：用户 2026-06-20 反馈“领路计划书也是不清楚怎么填和行家的计划书，需要业主确认”。
- 状态：待业主确认。
