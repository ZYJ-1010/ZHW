#!/usr/bin/env python3
"""Generate code-derived database, API, and UI inventories for delivery."""

from __future__ import annotations

import argparse
import csv
import html
import json
import re
from collections import Counter, defaultdict
from datetime import date
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def write_csv(path: Path, fieldnames: list[str], rows: list[dict[str, object]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8-sig", newline="") as handle:
        writer = csv.DictWriter(handle, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(rows)


def write_text(path: Path, content: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content.rstrip() + "\n", encoding="utf-8")


def clean_identifier(value: str) -> str:
    return value.strip().strip('"').split(".")[-1]


def parse_table_columns(block: str) -> list[str]:
    columns: list[str] = []
    for raw_line in block.splitlines():
        line = raw_line.strip().rstrip(",")
        if not line or line.startswith("--"):
            continue
        lowered = line.lower()
        if lowered.startswith(("constraint ", "primary key", "unique(", "unique (", "check (", "foreign key")):
            continue
        match = re.match(r'"?([a-zA-Z_][a-zA-Z0-9_]*)"?\s+', line)
        if match:
            columns.append(match.group(1))
    return columns


def database_inventory() -> list[dict[str, object]]:
    migrations_dir = ROOT / "db" / "migrations"
    tables: dict[str, dict[str, object]] = {}

    create_pattern = re.compile(
        r"create\s+table(?:\s+if\s+not\s+exists)?\s+([^\s(]+)\s*\((.*?)\);",
        re.IGNORECASE | re.DOTALL,
    )
    alter_pattern = re.compile(
        r"alter\s+table\s+([^\s;]+)\s+(.*?);",
        re.IGNORECASE | re.DOTALL,
    )
    add_column_pattern = re.compile(
        r"add\s+column(?:\s+if\s+not\s+exists)?\s+([a-zA-Z_][a-zA-Z0-9_]*)",
        re.IGNORECASE,
    )
    index_pattern = re.compile(
        r"create\s+(unique\s+)?index(?:\s+if\s+not\s+exists)?\s+[^\s]+\s+on\s+([^\s(]+)",
        re.IGNORECASE,
    )

    for migration in sorted(migrations_dir.glob("*.sql")):
        text = migration.read_text(encoding="utf-8")
        for match in create_pattern.finditer(text):
            table = clean_identifier(match.group(1))
            line = text.count("\n", 0, match.start()) + 1
            tables.setdefault(
                table,
                {
                    "table_name": table,
                    "created_in": migration.name,
                    "create_line": line,
                    "columns": [],
                    "alter_migrations": set(),
                    "index_count": 0,
                    "unique_index_count": 0,
                },
            )
            tables[table]["columns"] = parse_table_columns(match.group(2))

        for match in alter_pattern.finditer(text):
            table = clean_identifier(match.group(1))
            if table not in tables:
                continue
            added = add_column_pattern.findall(match.group(2))
            if not added:
                continue
            columns = tables[table]["columns"]
            assert isinstance(columns, list)
            for column in added:
                if column not in columns:
                    columns.append(column)
            alter_migrations = tables[table]["alter_migrations"]
            assert isinstance(alter_migrations, set)
            alter_migrations.add(migration.name)

        for match in index_pattern.finditer(text):
            table = clean_identifier(match.group(2))
            if table not in tables:
                continue
            tables[table]["index_count"] = int(tables[table]["index_count"]) + 1
            if match.group(1):
                tables[table]["unique_index_count"] = int(tables[table]["unique_index_count"]) + 1

    rows: list[dict[str, object]] = []
    for table in sorted(tables):
        item = tables[table]
        columns = item["columns"]
        alter_migrations = item["alter_migrations"]
        assert isinstance(columns, list)
        assert isinstance(alter_migrations, set)
        rows.append(
            {
                "table_name": table,
                "created_in": item["created_in"],
                "create_line": item["create_line"],
                "column_count": len(columns),
                "columns": ";".join(columns),
                "alter_migrations": ";".join(sorted(alter_migrations)),
                "index_count": item["index_count"],
                "unique_index_count": item["unique_index_count"],
            }
        )
    return rows


def classify_surface(path: str) -> str:
    for prefix, surface in (
        ("/api/app/", "小程序 API"),
        ("/api/admin/", "管理后台 API"),
        ("/api/internal/", "内部 API"),
        ("/api/funds/", "资金 API"),
    ):
        if path.startswith(prefix):
            return surface
    return "健康检查"


def auth_label(path: str, handler: str) -> str:
    permission = re.search(r'requireAdminPermission\("([^"]+)"', handler)
    if permission:
        return f"后台权限:{permission.group(1)}"
    if "AppAuthMiddleware" in handler:
        return "小程序登录态"
    if path.startswith("/api/internal/"):
        return "内部调用"
    if "admin" in handler.lower() and "Auth" not in handler:
        return "由处理器校验"
    return "公开或处理器内校验"


def api_inventory() -> list[dict[str, object]]:
    rows: list[dict[str, object]] = []
    server_path = ROOT / "services" / "go-api" / "internal" / "appapi" / "server.go"
    route_pattern = re.compile(r'^\s*handle\("(GET|POST|PUT|DELETE|PATCH)\s+([^"]+)",\s*(.+)\)\s*$')
    for line_no, line in enumerate(server_path.read_text(encoding="utf-8").splitlines(), 1):
        match = route_pattern.match(line)
        if not match:
            continue
        method, path, handler = match.groups()
        rows.append(
            {
                "service": "go-api",
                "surface": classify_surface(path),
                "method": method,
                "route_pattern": path,
                "dynamic_prefix": "是" if path.endswith("/") else "否",
                "auth_or_permission": auth_label(path, handler),
                "handler": handler,
                "source_file": str(server_path.relative_to(ROOT)),
                "source_line": line_no,
            }
        )

    main_path = ROOT / "services" / "go-api" / "cmd" / "server" / "main.go"
    health_pattern = re.compile(r'mux\.HandleFunc\("(GET|POST|PUT|DELETE|PATCH)\s+([^"]+)"')
    for line_no, line in enumerate(main_path.read_text(encoding="utf-8").splitlines(), 1):
        match = health_pattern.search(line)
        if not match:
            continue
        method, path = match.groups()
        rows.append(
            {
                "service": "go-api",
                "surface": classify_surface(path),
                "method": method,
                "route_pattern": path,
                "dynamic_prefix": "否",
                "auth_or_permission": "公开",
                "handler": "health handler",
                "source_file": str(main_path.relative_to(ROOT)),
                "source_line": line_no,
            }
        )

    funds_root = ROOT / "services" / "funds-service" / "src" / "main" / "java"
    context_pattern = re.compile(r'createContext\("([^"]+)"')
    for java_path in sorted(funds_root.rglob("*.java")):
        for line_no, line in enumerate(java_path.read_text(encoding="utf-8").splitlines(), 1):
            match = context_pattern.search(line)
            if not match:
                continue
            path = match.group(1)
            rows.append(
                {
                    "service": "funds-service-placeholder",
                    "surface": classify_surface(path),
                    "method": "ANY",
                    "route_pattern": path,
                    "dynamic_prefix": "否",
                    "auth_or_permission": "未实现统一鉴权",
                    "handler": java_path.stem,
                    "source_file": str(java_path.relative_to(ROOT)),
                    "source_line": line_no,
                }
            )

    return sorted(rows, key=lambda item: (str(item["service"]), str(item["route_pattern"]), str(item["method"])))


def strip_html(value: str) -> str:
    return re.sub(r"\s+", " ", re.sub(r"<[^>]+>", "", value)).strip()


def ui_inventory() -> list[dict[str, object]]:
    rows: list[dict[str, object]] = []
    app_json_path = ROOT / "miniprogram-client" / "app.json"
    app_config = json.loads(app_json_path.read_text(encoding="utf-8"))
    registered: set[str] = set()

    route_entries = [("main", route) for route in app_config.get("pages", [])]
    for package in app_config.get("subPackages", []):
        root = package["root"]
        route_entries.extend((root, f"{root}/{route}") for route in package.get("pages", []))

    for package, route in route_entries:
        registered.add(route)
        page_json = ROOT / "miniprogram-client" / f"{route}.json"
        title = ""
        if page_json.exists():
            try:
                title = json.loads(page_json.read_text(encoding="utf-8")).get("navigationBarTitleText", "")
            except json.JSONDecodeError:
                title = ""
        rows.append(
            {
                "surface": "微信小程序",
                "module": package,
                "identifier": route,
                "title": title,
                "status": "已注册",
                "source_file": str(page_json.relative_to(ROOT)) if page_json.exists() else str(app_json_path.relative_to(ROOT)),
                "notes": "app.json 页面路由",
            }
        )

    pages_root = ROOT / "miniprogram-client" / "pages"
    for wxml in sorted(pages_root.rglob("index.wxml")):
        route = str(wxml.relative_to(ROOT / "miniprogram-client")).removesuffix(".wxml")
        if route in registered or "/components/" in route:
            continue
        rows.append(
            {
                "surface": "微信小程序",
                "module": route.split("/")[1] if "/" in route else "pages",
                "identifier": route,
                "title": "",
                "status": "存在页面文件但未注册",
                "source_file": str(wxml.relative_to(ROOT)),
                "notes": "不能从 app.json 正常进入",
            }
        )

    admin_html = ROOT / "admin-web" / "src" / "index.html"
    html = admin_html.read_text(encoding="utf-8")
    rows.append(
        {
            "surface": "PC 管理后台",
            "module": "auth",
            "identifier": "login",
            "title": "后台登录",
            "status": "已实现",
            "source_file": str(admin_html.relative_to(ROOT)),
            "notes": "登录视图",
        }
    )
    view_pattern = re.compile(r'<button[^>]*data-view="([^"]+)"[^>]*>(.*?)</button>', re.DOTALL)
    for view, label in view_pattern.findall(html):
        rows.append(
            {
                "surface": "PC 管理后台",
                "module": "primary-view",
                "identifier": view,
                "title": strip_html(label),
                "status": "已实现",
                "source_file": str(admin_html.relative_to(ROOT)),
                "notes": "一级导航视图",
            }
        )
    tab_pattern = re.compile(
        r'<button[^>]*data-tab-group="([^"]+)"[^>]*data-tab-target="([^"]+)"[^>]*>(.*?)</button>',
        re.DOTALL,
    )
    for group, target, label in tab_pattern.findall(html):
        rows.append(
            {
                "surface": "PC 管理后台",
                "module": group,
                "identifier": target.lstrip("#"),
                "title": strip_html(label),
                "status": "已实现",
                "source_file": str(admin_html.relative_to(ROOT)),
                "notes": "页签/子视图",
            }
        )

    prototype_readme = ROOT / "docs" / "prototype-pt-20260719" / "README.md"
    prototype_pattern = re.compile(
        r"^\d+\.\s+(.+?)（`([^`]+)`；元素\s+(\d+)）$",
        re.MULTILINE,
    )
    for title, canvas_id, element_count in prototype_pattern.findall(prototype_readme.read_text(encoding="utf-8")):
        rows.append(
            {
                "surface": "墨刀原型",
                "module": title.split("/")[0].strip(),
                "identifier": canvas_id,
                "title": title.strip(),
                "status": "已归档 PT 数据",
                "source_file": str(prototype_readme.relative_to(ROOT)),
                "notes": f"元素数:{element_count}; 375x812pt",
            }
        )

    return rows


def markdown_cell(value: object) -> str:
    text = "" if value is None else str(value)
    return text.replace("|", "\\|").replace("\n", "<br>")


def markdown_code_path(value: object) -> str:
    text = html.escape(str(value or ""))
    for separator in ("/", "_", "-", "."):
        text = text.replace(separator, separator + "<wbr>")
    return f"<code>{text}</code>"


def compact_handler(value: object) -> str:
    text = str(value or "")
    matches = re.findall(r"\bs\.([A-Za-z0-9_]+)", text)
    return matches[-1] if matches else text


def render_database_markdown(rows: list[dict[str, object]], inventory_date: str) -> str:
    by_migration: dict[str, list[dict[str, object]]] = defaultdict(list)
    for row in rows:
        by_migration[str(row["created_in"])].append(row)

    lines = [
        "# 数据库设计清单（当前实现）",
        "",
        "## 1. 基础设计",
        "",
        "| 项目 | 内容 |",
        "| --- | --- |",
        f"| 数据库 | PostgreSQL |",
        f"| 迁移文件 | {len(list((ROOT / 'db' / 'migrations').glob('*.sql')))} |",
        f"| 建表对象 | {len(rows)} |",
        f"| 显式索引 | {sum(int(row['index_count']) for row in rows)} |",
        "| 主键 | `bigserial`；关联表使用联合主键 |",
        "| 时间 | `timestamptz` 为主；部分组局计划时间为 `varchar(32)` |",
        "| 金额 | 资金域使用整数分；`games.price` 使用 `decimal(10,2)` |",
        "| 扩展字段 | `jsonb` |",
        "| 敏感数据 | 密文、掩码、哈希字段 |",
        "| 删除 | 主要通过 `status`；未统一使用 `deleted_at` |",
        "| 外键 | 少量显式外键；多数关系由应用层维护 |",
        "| 迁移台账 | `schema_migrations` 按完整迁移文件名记录 |",
        "",
        "## 2. 核心关系",
        "",
        "| 主表 | 关联表 | 关联字段 |",
        "| --- | --- | --- |",
        "| `users` | `user_wechat_accounts` | `user_id` |",
        "| `users` | `invite_codes` | `owner_user_id` |",
        "| `invite_codes` | `invite_relations` | `invite_code_id` |",
        "| `users` | `user_roles` | `user_id` |",
        "| `users` | `role_applications` | `user_id` |",
        "| `users` | `games` | `creator_user_id` |",
        "| `games` | `game_locations` | `game_id` |",
        "| `games` | `game_members` | `game_id` |",
        "| `games` | `game_applications` | `game_id` |",
        "| `games` | `game_invitations` | `game_id` |",
        "| `games` | `chat_rooms` | `game_id` |",
        "| `chat_rooms` | `chat_room_members` | `room_id` |",
        "| `chat_rooms` | `chat_messages` | `room_id` |",
        "| `games` | `reviews` | `game_id` |",
        "| `users` | `user_growth_profiles` | `user_id` |",
        "| `users` | `credit_accounts` | `user_id` |",
        "| `users` | `points_accounts` | `user_id` |",
        "| `games` | `revenue_records` | `game_id` |",
        "| `revenue_records` | `revenue_record_items` | `revenue_record_id` |",
        "| `users` | `reports` | `reporter_user_id`、`target_user_id` |",
        "| `users` | `notifications` | `user_id` |",
        "",
        "## 3. 表结构清单",
    ]

    for migration in sorted(by_migration):
        lines.extend(["", f"### {migration}"])
        for row in sorted(by_migration[migration], key=lambda item: str(item["table_name"])):
            columns = "、".join(str(row["columns"]).split(";"))
            alters = "、".join(str(row["alter_migrations"]).split(";")) if row["alter_migrations"] else "-"
            lines.extend(
                [
                    "",
                    f"#### `{markdown_cell(row['table_name'])}`",
                    "",
                    f"- 字段：{markdown_cell(columns)}",
                    f"- 后续变更：{markdown_cell(alters)}",
                    f"- 索引：{markdown_cell(row['index_count'])}；唯一索引：{markdown_cell(row['unique_index_count'])}",
                ]
            )

    lines.extend(
        [
            "",
            "## 4. 数据约束清单",
            "",
            "| 对象 | 约束 |",
            "| --- | --- |",
            "| 微信账号 | `openid` 唯一 |",
            "| 邀请码 | `code` 唯一 |",
            "| 用户角色 | `user_id + role_code` 唯一 |",
            "| 待审核角色申请 | 同一用户同一角色仅允许一条 `pending` |",
            "| 局成员 | `game_id + user_id` 主键 |",
            "| 入局申请 | 同一用户同一局仅允许一条待审核申请 |",
            "| 局邀约 | 按局、目标用户和状态控制重复 |",
            "| 群聊客户端消息 | `room_id + sender_user_id + client_msg_id` 唯一 |",
            "| 评价 | `game_id + reviewer + target + target_role` 唯一 |",
            "| 信用流水 | 用户和幂等键唯一 |",
            "| 成长事件 | `idempotency_key` 唯一 |",
            "| 分润模板 | 三方基点合计 10000 |",
            "| 分润记录 | 每个局一条记录 |",
            "| 私聊会话 | 双方用户唯一且不能相同 |",
            "| 企业认证 | 同一用户仅允许一条待审核记录 |",
            "",
            "## 5. 当前问题清单",
            "",
            "| 优先级 | 问题 |",
            "| --- | --- |",
            "| P0 | 多数 `user_id`、`game_id`、`file_id` 没有外键 |",
            "| P0 | 组局计划时间部分使用字符串 |",
            "| P0 | 元和分两套金额单位并存 |",
            "| P0 | 实名、位置、信用存在新旧表并存 |",
            "| P1 | 大量状态字段为自由字符串 |",
            "| P1 | `JSONB` 配置缺少统一版本 Schema |",
            "| P1 | 消息、行为、通知、日志未设计分区和归档 |",
            "| P1 | 敏感数据需补密钥轮换、访问审计和保留期限 |",
            "",
            "## 6. 迁移文件",
            "",
            "| 用途 | 文件 |",
            "| --- | --- |",
            "| 新库初始化 | `deploy/postgres/init-prod.sh` |",
            "| 增量迁移 | `db/migrate-prod.sh` |",
            "| 迁移目录 | `db/migrations/*.sql` |",
            "| 种子目录 | `db/seeds/*.sql` |",
            "| 字段机器清单 | `docs/数据库表清单-{}.csv` |".format(inventory_date),
        ]
    )
    return "\n".join(lines)


def render_api_markdown(rows: list[dict[str, object]], inventory_date: str) -> str:
    service_counts = Counter(str(row["service"]) for row in rows)
    surface_counts = Counter(str(row["surface"]) for row in rows)
    lines = [
        "# 接口清单（当前实现）",
        "",
        "## 1. 接口统计",
        "",
        "| 分类 | 数量 |",
        "| --- | ---: |",
        f"| Go API | {service_counts.get('go-api', 0)} |",
        f"| Java 资金预留服务 | {service_counts.get('funds-service-placeholder', 0)} |",
        f"| 小程序 API | {surface_counts.get('小程序 API', 0)} |",
        f"| 管理后台 API | {surface_counts.get('管理后台 API', 0)} |",
        f"| 内部 API | {surface_counts.get('内部 API', 0)} |",
        f"| 资金 API | {surface_counts.get('资金 API', 0)} |",
        f"| 健康检查 | {surface_counts.get('健康检查', 0)} |",
        f"| 合计 | {len(rows)} |",
        "",
        "## 2. 接口约定",
        "",
        "| 项目 | 内容 |",
        "| --- | --- |",
        "| 小程序鉴权 | `Authorization: Bearer <token>`；`AppAuthMiddleware` |",
        "| 后台鉴权 | 管理员 Token；`requireAdminPermission` |",
        "| 幂等 | `IdempotencyMiddleware` |",
        "| 返回 | `code`、`message`、`data`、`requestId` |",
        "| 动态路由 | 路径以 `/` 结尾，由处理器解析 ID 和动作后缀 |",
        "| WebSocket | `GET /api/app/im/ws` |",
        "| App OpenAPI | `docs/openapi/app.openapi.yaml` |",
        "| Admin OpenAPI | `docs/openapi/admin.openapi.yaml` |",
        "",
        "## 3. 全量注册路由",
    ]

    ordered_groups = [
        ("go-api", "小程序 API"),
        ("go-api", "管理后台 API"),
        ("go-api", "内部 API"),
        ("go-api", "资金 API"),
        ("go-api", "健康检查"),
        ("funds-service-placeholder", "资金 API"),
        ("funds-service-placeholder", "内部 API"),
        ("funds-service-placeholder", "健康检查"),
    ]
    group_titles = {
        ("go-api", "小程序 API"): "Go API - 小程序",
        ("go-api", "管理后台 API"): "Go API - 管理后台",
        ("go-api", "内部 API"): "Go API - 内部任务",
        ("go-api", "资金 API"): "Go API - 资金兼容",
        ("go-api", "健康检查"): "Go API - 健康检查",
        ("funds-service-placeholder", "资金 API"): "Java 资金预留服务 - 资金",
        ("funds-service-placeholder", "内部 API"): "Java 资金预留服务 - 内部",
        ("funds-service-placeholder", "健康检查"): "Java 资金预留服务 - 健康检查",
    }
    for key in ordered_groups:
        group_rows = [row for row in rows if (row["service"], row["surface"]) == key]
        if not group_rows:
            continue
        lines.extend(
            [
                "",
                f"### {group_titles[key]}（{len(group_rows)}）",
                "",
                "| 方法 | 接口信息 | 鉴权/权限 |",
                "| --- | --- | --- |",
            ]
        )
        for row in sorted(group_rows, key=lambda item: (str(item["route_pattern"]), str(item["method"]))):
            source = f"{row['source_file']}:{row['source_line']}"
            details = [
                markdown_code_path(row["route_pattern"]),
                f"处理器：<code>{html.escape(compact_handler(row['handler']))}</code>",
                f"源码：{markdown_code_path(source)}",
            ]
            if row["dynamic_prefix"] == "是":
                details.insert(1, "类型：动态路由前缀")
            lines.append(
                "| {} | {} | {} |".format(
                    markdown_cell(row["method"]),
                    "<br>".join(details),
                    markdown_cell(row["auth_or_permission"]),
                )
            )

    lines.extend(
        [
            "",
            "## 4. OpenAPI 差异",
            "",
            "| 项目 | 数量 |",
            "| --- | ---: |",
            "| App OpenAPI 操作 | 201 |",
            "| Admin OpenAPI 操作 | 151 |",
            "| OpenAPI 合计 | 352 |",
            "| 代码静态路由未进入 OpenAPI | 30 |",
            "",
            "```text",
            "POST /api/app/auth/wechat-entry-precheck",
            "POST /api/app/auth/password-login",
            "POST /api/app/auth/password-reset",
            "POST /api/app/account/wechat-bind",
            "PUT  /api/app/account/login-password",
            "POST /api/app/account/delete",
            "POST /api/app/newbie-tasks/guide-profile-reminder",
            "GET  /api/app/game-drafts",
            "POST /api/app/game-drafts",
            "GET  /api/app/games/create-template-config",
            "GET  /api/app/profile/service-center/invite/codes",
            "POST /api/app/profile/service-center/invite/quota-requests",
            "GET  /api/app/locations/fallback",
            "GET  /api/admin/invite-code-config",
            "PUT  /api/admin/invite-code-config",
            "GET  /api/admin/invite-codes/export",
            "GET  /api/admin/invite-owners",
            "GET  /api/admin/invite-quota-requests",
            "GET  /api/admin/games/create-template-config",
            "PUT  /api/admin/games/create-template-config",
            "GET  /api/admin/experts/skill-display-config",
            "PUT  /api/admin/experts/skill-display-config",
            "GET  /api/admin/growth/role-level-config",
            "PUT  /api/admin/growth/role-level-config",
            "GET  /api/admin/growth/role-metric-config",
            "PUT  /api/admin/growth/role-metric-config",
            "GET  /api/admin/credit/restriction-config",
            "PUT  /api/admin/credit/restriction-config",
            "GET  /api/admin/revenue/points-reward-config",
            "PUT  /api/admin/revenue/points-reward-config",
            "```",
            "",
            "## 5. 配套文件",
            "",
            "| 文件 | 路径 |",
            "| --- | --- |",
            "| 注册路由 CSV | `docs/接口注册路由清单-{}.csv` |".format(inventory_date),
            "| App OpenAPI | `docs/openapi/app.openapi.yaml` |",
            "| Admin OpenAPI | `docs/openapi/admin.openapi.yaml` |",
            "| DTO 示例 | `docs/openapi/dto-samples.md` |",
            "| 错误码 | `docs/openapi/error-codes.md` |",
            "| WebSocket | `docs/openapi/ws-protocol.md` |",
        ]
    )
    return "\n".join(lines)


def render_ui_markdown(rows: list[dict[str, object]], inventory_date: str) -> str:
    surface_counts = Counter(str(row["surface"]) for row in rows)
    lines = [
        "# 全量 UI 页面清单（当前实现）",
        "",
        "## 1. 数量",
        "",
        "| UI 来源 | 数量 |",
        "| --- | ---: |",
    ]
    for surface, count in surface_counts.items():
        lines.append(f"| {surface} | {count} |")
    lines.extend(
        [
            f"| 合计 | {len(rows)} |",
            "",
            "## 2. 页面/视图清单",
        ]
    )
    surface_order = {"微信小程序": 0, "PC 管理后台": 1, "墨刀原型": 2}
    ordered_rows = sorted(rows, key=lambda item: (surface_order.get(str(item["surface"]), 9), str(item["module"]), str(item["identifier"])))
    row_number = 0
    for surface in sorted({str(row["surface"]) for row in ordered_rows}, key=lambda value: surface_order.get(value, 9)):
        lines.extend(["", f"### {surface}"])
        surface_rows = [row for row in ordered_rows if row["surface"] == surface]
        for module in sorted({str(row["module"]) for row in surface_rows}):
            lines.extend(
                [
                    "",
                    f"#### {markdown_cell(module)}",
                    "",
                    "| 编号 | 页面/视图信息 | 状态 |",
                    "| ---: | --- | --- |",
                ]
            )
            for row in [item for item in surface_rows if item["module"] == module]:
                row_number += 1
                title = markdown_cell(row["title"] or "未设置页面标题")
                details = [f"**{title}**", markdown_code_path(row["identifier"])]
                if row["source_file"]:
                    details.append(f"源文件：{markdown_code_path(row['source_file'])}")
                if row["notes"]:
                    details.append(markdown_cell(row["notes"]))
                lines.append(
                    "| {} | {} | {} |".format(
                        row_number,
                        "<br>".join(details),
                        markdown_cell(row["status"]),
                    )
                )
    lines.extend(
        [
            "",
            "## 3. 交付状态定义",
            "",
            "| 状态 | 含义 |",
            "| --- | --- |",
            "| 已注册 | 已写入小程序 `app.json` |",
            "| 已实现 | 已写入后台视图或登录视图 |",
            "| 存在页面文件但未注册 | 有 `index.wxml`，但不能从 `app.json` 正常进入 |",
            "| 已归档 PT 数据 | 墨刀原型画布，需与代码路由和截图另行映射 |",
        ]
    )
    return "\n".join(lines)


def parse_markdown_row(line: str) -> list[str]:
    return [cell.strip() for cell in line.strip().strip("|").split("|")]


def break_inline_code_paths(value: str) -> str:
    def replace(match: re.Match[str]) -> str:
        code = match.group(1)
        if any(separator in code for separator in ("/", "_", "-", ".")):
            return markdown_code_path(code)
        return match.group(0)

    return re.sub(r"`([^`]+)`", replace, value)


def reformat_feature_markdown(path: Path) -> None:
    source = path.read_text(encoding="utf-8").splitlines()
    output: list[str] = []
    index = 0
    while index < len(source):
        if not source[index].startswith("|") or index + 1 >= len(source) or not source[index + 1].startswith("| ---"):
            output.append(source[index])
            index += 1
            continue

        block: list[str] = []
        while index < len(source) and source[index].startswith("|"):
            block.append(source[index])
            index += 1
        header = parse_markdown_row(block[0])
        data_rows = [parse_markdown_row(line) for line in block[2:]]

        if len(header) == 6:
            output.extend(["| 功能 | 页面、后端及后台实现 | 状态 |", "| --- | --- | --- |"]) 
            for cells in data_rows:
                if len(cells) != 6:
                    continue
                feature = f"**{cells[0]}**<br>{cells[1]}"
                implementation = "<br>".join(
                    [
                        f"页面：{break_inline_code_paths(cells[2])}",
                        f"后端/数据：{break_inline_code_paths(cells[3])}",
                        f"后台：{break_inline_code_paths(cells[4])}",
                    ]
                )
                output.append(f"| {feature} | {implementation} | {cells[5]} |")
        elif len(header) == 4:
            output.extend(["| 功能 | 相关实现 | 状态 |", "| --- | --- | --- |"]) 
            for cells in data_rows:
                if len(cells) != 4:
                    continue
                feature = f"**{cells[0]}**<br>{cells[1]}"
                output.append(f"| {feature} | {break_inline_code_paths(cells[2])} | {cells[3]} |")
        else:
            output.append("| " + " | ".join(header) + " |")
            output.append(block[1])
            for cells in data_rows:
                output.append("| " + " | ".join(break_inline_code_paths(cell) for cell in cells) + " |")

    write_text(path, "\n".join(output))


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--date", default=date.today().strftime("%Y%m%d"))
    args = parser.parse_args()

    database_rows = database_inventory()
    api_rows = api_inventory()
    ui_rows = ui_inventory()

    docs_dir = ROOT / "docs"
    write_csv(
        docs_dir / f"数据库表清单-{args.date}.csv",
        [
            "table_name",
            "created_in",
            "create_line",
            "column_count",
            "columns",
            "alter_migrations",
            "index_count",
            "unique_index_count",
        ],
        database_rows,
    )
    write_csv(
        docs_dir / f"接口注册路由清单-{args.date}.csv",
        [
            "service",
            "surface",
            "method",
            "route_pattern",
            "dynamic_prefix",
            "auth_or_permission",
            "handler",
            "source_file",
            "source_line",
        ],
        api_rows,
    )
    write_csv(
        docs_dir / f"全量UI页面清单-{args.date}.csv",
        ["surface", "module", "identifier", "title", "status", "source_file", "notes"],
        ui_rows,
    )
    write_text(
        docs_dir / "数据库设计-当前实现.md",
        render_database_markdown(database_rows, args.date),
    )
    write_text(
        docs_dir / "接口清单-当前实现.md",
        render_api_markdown(api_rows, args.date),
    )
    write_text(
        docs_dir / f"全量UI页面清单-{args.date}.md",
        render_ui_markdown(ui_rows, args.date),
    )
    feature_path = docs_dir / "功能清单-当前实现.md"
    if feature_path.exists():
        reformat_feature_markdown(feature_path)

    print(f"database_tables={len(database_rows)}")
    print(f"api_route_registrations={len(api_rows)}")
    print(f"ui_inventory_items={len(ui_rows)}")


if __name__ == "__main__":
    main()
