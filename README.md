# 真好玩一期小程序后端与后台工程

本仓库承载真好玩一期的小程序后端、PC 管理后台、资金服务占位、数据库迁移、部署配置和验收记录。

## 责任范围

- 小程序后端接口：登录、邀请、实名、角色、组局、入局、LBS、IM、评价、成长、分润记录、举报申诉、个人中心、文件上传、通知。
- PC 管理后台：后台前端页面、后台 API、RBAC 权限、审核、组局管理、IM 管理、分润收益、举报申诉、数据看板、导出、操作日志。
- 基础设施：PostgreSQL、Redis、对象存储、Nginx、Docker 部署、日志、监控、备份、回滚。
- 小程序前端联调：在不破坏前端既有原型样式的前提下，负责补齐接口调用、路由对接、字段绑定、测试数据下沉和验收用例。

## 当前骨架

```text
services/go-api        小程序后端与后台 API 主服务
services/funds-service 资金、分润、订单占位服务
admin-web             PC 管理后台前端
db/migrations         数据库迁移脚本
db/seeds              初始化数据
deploy                部署配置
docs/openapi          接口契约
docs/test-cases       测试用例
docs/progress         每小时项目记录
```

## 本地健康检查

Go API：

```powershell
.\scripts\restart-local-backend.ps1
```

默认使用 `E:\zhw-local-runtime` 里的本地 Go、模块缓存、构建缓存、临时目录和日志目录，避免本地重启时访问公网下载 Go 依赖。

访问：

```bash
curl http://127.0.0.1:8080/health
```

Java 资金服务：

```bash
cd services/funds-service
javac -encoding UTF-8 -d target/classes src/main/java/com/zhw/funds/common/*.java src/main/java/com/zhw/funds/placeholder/*.java src/main/java/com/zhw/funds/revenue/*.java src/main/java/com/zhw/funds/settlement/*.java src/main/java/com/zhw/funds/FundsApplication.java
java -cp target/classes com.zhw.funds.FundsApplication
```

访问：

```bash
curl http://127.0.0.1:8081/health
```

后台静态壳：

```bash
cd admin-web
npm run dev
```
