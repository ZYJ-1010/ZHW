# 开发记录与约定

## 文档边界

`00.文档` 是正式需求来源，保持只读参考。

工程侧新增记录放在 `docs/`：

```text
docs/architecture.md
docs/page-map.md
docs/api-additions.md
docs/login-test.md
docs/development-notes.md
```

## 当前进度

1. 已将登录模块拆为页面层、业务层、接口层、配置层。
2. 已加入 mock / prod 环境切换。
3. 已根据墨刀已读页面树建立小程序页面骨架。
4. 已为用户、角色、组局、定位、IM、个人中心、评价、收益建立服务和接口模块占位。

## 下一步建议

1. 先完成第 1 批入口闭环：登录、首页、个人中心基础。
2. 再接第 2 批角色与组局基础。
3. 每做一个业务模块，同步补 mock 数据和接口说明。
4. 正式接口缺口继续记到 `docs/api-additions.md`。
