const state = {
  token: localStorage.getItem("zhw_admin_token") || "",
  admin: null,
  permissions: [],
  view: "dashboard",
  games: [],
  gameApplications: [],
  dashboard: null,
  imRooms: [],
  users: [],
  inviteCodes: [],
  inviteRelations: [],
  revenueTemplates: [],
  revenueRecords: [],
  revenueSettlements: [],
  redemptionItems: [],
  redemptionOrders: [],
  pointsLogs: [],
  memberTeams: [],
  memberReports: [],
  connections: [],
  expertProfile: null,
  guideProfile: null,
  userGrowthTrace: null,
  gameReviewTrace: null,
  reports: [],
  feedbackRecords: [],
  exportTemplates: [],
  exportTasks: [],
  funnel: null,
  retention: null,
  behaviorEvents: [],
  behaviorLogs: [],
  aiSnapshot: null,
  wechatTasks: [],
  wechatTemplates: [],
  deliveryDocuments: [],
  testCases: [],
  testRuns: [],
  operationLogs: [],
  sensitiveWords: [],
  riskLogs: [],
  readiness: null,
  permissionTree: null,
  guideQualificationRules: [],
  adminUsers: [],
  adminRoles: [],
  adminPermissionCatalog: [],
  aiExportConfig: null,
  identities: [],
  roleApplications: [],
  gameApplicationConfig: null,
  gameAuditConfig: null,
  conditionRuleConfig: null,
  roleBenefitConfig: null,
  creditDeductionRules: [],
};

const views = {
  dashboard: { title: "一期后台工作台", crumb: "总览" },
  users: { title: "用户管理", crumb: "用户 / 实名 / 身份" },
  invites: { title: "邀请管理", crumb: "邀请码 / 关系 / 入口" },
  games: { title: "组局管理", crumb: "组局 / 开局 / 审核" },
  audits: { title: "审核中心", crumb: "实名 / 角色" },
  revenue: { title: "分润结算", crumb: "分润 / 规则 / 结算" },
  redemption: { title: "积分兑换", crumb: "积分 / 兑换 / 订单" },
  members: { title: "会员团队", crumb: "会员 / 团队 / 报表" },
  profiles: { title: "画像关系", crumb: "用户画像 / 人脉 / 资源" },
  growth: { title: "评价成长", crumb: "评价 / 信用 / 足迹" },
  reports: { title: "举报申诉", crumb: "投诉 / 证据 / 处理" },
  exports: { title: "导出中心", crumb: "报表 / 任务 / 文件" },
  analytics: { title: "数据分析", crumb: "行为 / 留存 / AI 数据" },
  delivery: { title: "通知交付", crumb: "通知 / 文档 / 测试" },
  admins: { title: "管理员权限", crumb: "账号 / 角色 / 权限" },
  system: { title: "系统配置", crumb: "配置 / 内容安全 / 权限" },
  im: { title: "IM 证据", crumb: "房间 / 文件 / 争议" },
  logs: { title: "操作日志", crumb: "权限 / 审计" },
};

const $ = (selector, root = document) => root.querySelector(selector);
const $$ = (selector, root = document) => Array.from(root.querySelectorAll(selector));

document.addEventListener("DOMContentLoaded", () => {
  bindShell();
  if (state.token) {
    boot();
  }
});

function bindShell() {
  $("#login-form").addEventListener("submit", login);
  $("#logout-button").addEventListener("click", logout);
  $("#refresh-button").addEventListener("click", () => loadView(state.view));
  $("#main-nav").addEventListener("click", (event) => {
    const button = event.target.closest("button[data-view]");
    if (!button) return;
    state.view = button.dataset.view;
    setActiveNav();
    loadView(state.view);
  });
}

async function boot() {
  try {
    const data = await apiGet("/api/admin/auth/permissions");
    state.admin = data.adminUser;
    state.permissions = data.permissions || [];
    state.permissionTree = data;
    $("#login-screen").classList.add("hidden");
    $("#app-shell").classList.remove("hidden");
    $("#admin-role").textContent = `${state.admin?.username || "admin"} / ${(state.admin?.roles || []).join(",") || "role"}`;
    setAPIStatus(true);
    applyNavPermissions();
    ensureAllowedView();
    setActiveNav();
    await loadView(state.view);
  } catch (error) {
    setAPIStatus(false);
    showLogin();
    showLoginError(error.message);
  }
}

async function login(event) {
  event.preventDefault();
  showLoginError("");
  const username = $("#login-username").value.trim();
  const password = $("#login-password").value;
  try {
    const data = await apiPost("/api/admin/auth/login", { username, password }, false);
    state.token = data.token;
    localStorage.setItem("zhw_admin_token", state.token);
    await boot();
  } catch (error) {
    showLoginError(error.message);
  }
}

function logout() {
  state.token = "";
  state.admin = null;
  state.permissions = [];
  localStorage.removeItem("zhw_admin_token");
  showLogin();
}

function showLogin() {
  $("#login-screen").classList.remove("hidden");
  $("#app-shell").classList.add("hidden");
}

function showLoginError(message) {
  $("#login-error").textContent = message || "";
}

async function loadView(view) {
  if (!isViewAllowed(view)) {
    state.view = firstAllowedView();
    view = state.view;
    setActiveNav();
  }
  const meta = views[view] || views.dashboard;
  $("#page-title").textContent = meta.title;
  $("#breadcrumb").textContent = meta.crumb;
  clearToast();

  const root = $("#view-root");
  root.innerHTML = template(`${view}-template`);
  if (view === "dashboard") await renderDashboard();
  if (view === "users") await renderUsers();
  if (view === "invites") await renderInvites();
  if (view === "games") await renderGames();
  if (view === "audits") await renderAudits();
  if (view === "revenue") await renderRevenue();
  if (view === "redemption") await renderRedemption();
  if (view === "members") await renderMembers();
  if (view === "profiles") await renderProfiles();
  if (view === "growth") await renderGrowth();
  if (view === "reports") await renderReports();
  if (view === "exports") await renderExports();
  if (view === "analytics") await renderAnalytics();
  if (view === "delivery") await renderDelivery();
  if (view === "admins") await renderAdmins();
  if (view === "system") await renderSystem();
  if (view === "im") await renderIM();
  if (view === "logs") await renderLogs();
}

async function renderDashboard() {
  const [dashboard, games, imRooms] = await Promise.all([
    apiGet("/api/admin/dashboard"),
    apiGet("/api/admin/games"),
    can("im:room:read") ? apiGet("/api/admin/im/rooms") : Promise.resolve({ items: [] }),
  ]);
  state.dashboard = dashboard;
  state.games = games.items || [];
  state.imRooms = imRooms.items || [];

  setField("behaviorCount", countArray(dashboard.behaviorEvents || dashboard.events || dashboard.behaviorLogs));
  setField("gameCount", state.games.length);
  setField("pendingGameCount", state.games.filter((item) => item.status === "pending_audit").length);
  setField("imRoomCount", state.imRooms.length);
  renderGameTypeBars(state.games);
}

async function renderUsers() {
  $("#user-filter-form").addEventListener("submit", (event) => {
    event.preventDefault();
    loadUsers(new FormData(event.currentTarget));
  });
  $("#users-table").addEventListener("click", onUsersTableClick);
  await loadUsers();
}

async function loadUsers(formData) {
  const data = await apiGet(`/api/admin/users${querySuffix(formData)}`);
  state.users = data.items || [];
  $("#users-table").innerHTML = state.users.map(userRow).join("") || emptyRow(7, "暂无用户");
}

async function onUsersTableClick(event) {
  const button = event.target.closest("button[data-action='user-detail']");
  if (!button) return;
  try {
    await showUserDetail(button.dataset.id);
  } catch (error) {
    toast(error.message, true);
  }
}

async function showUserDetail(id) {
  const data = await apiGet(`/api/admin/users/${id}`);
  const user = data.user || {};
  const identity = data.identity || {};
  const growth = data.growth || user.growth || {};
  const income = data.incomeSummary || user.incomeSummary || {};
  const points = user.pointsSummary || {};
  const membership = user.membership || {};
  const [favoriteData, connectionData] = await Promise.all([
    can("user:read") ? apiGet(`/api/admin/users/${id}/favorites`) : Promise.resolve({ items: data.favorites || [] }),
    can("connection:read") ? apiGet(`/api/admin/users/${id}/connections`) : Promise.resolve({ items: data.connections || [] }),
  ]);
  const favorites = favoriteData.items || data.favorites || [];
  const connections = connectionData.items || data.connections || [];
  const panel = $("#user-detail-panel");
  panel.classList.remove("hidden");
  panel.innerHTML = `
    <div class="panel-head">
      <div>
        <h2>用户详情 #${escapeHTML(user.id)}</h2>
        <p>${escapeHTML(user.nickname || user.openId || "-")}</p>
      </div>
      <span class="${badgeClass(user.realnameStatus)}">${statusLabel(user.realnameStatus)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("openId", user.openId || "-")}
      ${detailCell("status", user.status || "-")}
      ${detailCell("realnameStatus", user.realnameStatus || "-")}
      ${detailCell("identity.status", identity.status || "-")}
      ${detailCell("inviteCode", user.inviteCode || "-")}
      ${detailCell("roles", roleSnapshot(data.roles))}
      ${detailCell("growth.level", growth.level || 0)}
      ${detailCell("growth.creditScore", growth.creditScore || 0)}
      ${detailCell("points.available", points.availablePoints || growth.points || 0)}
      ${detailCell("membership.status", membership.status || "-")}
      ${detailCell("favorites", favorites.length)}
      ${detailCell("connections", connections.length)}
      ${detailCell("income.totalCent", income.totalCent || 0)}
      ${detailCell("income.pendingCent", income.pendingCent || 0)}
      ${detailCell("income.settledCent", income.settledCent || 0)}
    </div>
    <div class="split profile-form-gap">
      ${gameOpsTable("收藏明细", ["userId", "gameId", "game.title", "createdAt"], favorites.map(userFavoriteRow).join("") || emptyRow(4, "暂无收藏"))}
      ${gameOpsTable("人脉明细", ["ID", "userId", "connectedUserId", "relationType", "source", "strengthScore", "updatedAt"], connections.map(connectionRow).join("") || emptyRow(7, "暂无人脉"))}
    </div>
  `;
  panel.scrollIntoView({ behavior: "smooth", block: "start" });
}

async function renderInvites() {
  ensureInviteDetailPanel();
  $("#invite-create-form").addEventListener("submit", createInviteCode);
  $("#invite-filter-form").addEventListener("submit", (event) => {
    event.preventDefault();
    loadInviteCodes(new FormData(event.currentTarget));
  });
  $("#relation-filter-form").addEventListener("submit", (event) => {
    event.preventDefault();
    loadInviteRelations(new FormData(event.currentTarget));
  });
  $("#invite-codes-table").addEventListener("click", onInviteTableClick);
  await Promise.all([loadInviteCodes(), loadInviteRelations()]);
}

async function createInviteCode(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const data = Object.fromEntries(new FormData(form).entries());
  const payload = {
    code: data.code.trim(),
    ownerUserId: numberOrZero(data.ownerUserId),
    entryType: data.entryType,
    maxUses: Number(data.maxUses || 1),
    batchCount: Number(data.batchCount || 1),
  };
  if (payload.batchCount > 1 && payload.code) {
    toast("批量生成时请留空 code，由系统自动生成", true);
    return;
  }
  try {
    const result = await apiPost("/api/admin/invite-codes", payload);
    toast(payload.batchCount > 1 ? `已批量生成 ${result.total || payload.batchCount} 个邀请码` : "邀请码已创建");
    const gameType = form.gameType.value;
    form.reset();
    form.gameType.value = gameType;
    form.entryType.value = "poster";
    form.maxUses.value = "1";
    form.batchCount.value = "1";
    await loadInviteCodes();
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadInviteCodes(formData) {
  const data = await apiGet(`/api/admin/invite-codes${querySuffix(formData)}`);
  state.inviteCodes = data.items || [];
  $("#invite-codes-table").innerHTML = state.inviteCodes.map(inviteCodeRow).join("") || emptyRow(8, "暂无邀请码");
}

async function loadInviteRelations(formData) {
  const data = await apiGet(`/api/admin/invite-relations${querySuffix(formData)}`);
  state.inviteRelations = data.items || [];
  $("#invite-relations-table").innerHTML = state.inviteRelations.map(inviteRelationRow).join("") || emptyRow(5, "暂无邀请关系");
}

async function onInviteTableClick(event) {
  const button = event.target.closest("button[data-action]");
  if (!button) return;
  try {
    if (button.dataset.action === "invite-detail") {
      await showInviteCodeDetail(button.dataset.code);
    }
    if (button.dataset.action === "invite-disable") {
      await apiPost(`/api/admin/invite-codes/${encodeURIComponent(button.dataset.code)}/disable`, {});
      toast(`邀请码 ${button.dataset.code} 已禁用`);
      await loadInviteCodes(new FormData($("#invite-filter-form")));
    }
  } catch (error) {
    toast(error.message, true);
  }
}

function ensureInviteDetailPanel() {
  if ($("#invite-detail-panel")) return;
  $("#view-root").insertAdjacentHTML("beforeend", `<section id="invite-detail-panel" class="panel hidden"></section>`);
}

async function showInviteCodeDetail(code) {
  if (!code) return;
  const data = await apiGet(`/api/admin/invite-codes/${encodeURIComponent(code)}`);
  const invite = data.inviteCode || {};
  const owner = data.ownerUser || {};
  const bound = data.boundUser || {};
  const relations = data.relations || [];
  const panel = $("#invite-detail-panel");
  panel.classList.remove("hidden");
  panel.innerHTML = `
    <div class="panel-head">
      <div>
        <h2>邀请码详情</h2>
        <p>字段对齐 invite_codes / invite_relations，用于核查三种邀请入口绑定链路。</p>
      </div>
      <div class="row-actions">
        <span class="${badgeClass(invite.status)}">${statusLabel(invite.status)}</span>
        ${invite.status === "active" && can("invite_code:manage") ? `<button class="ghost" data-action="detail-invite-disable" data-code="${escapeHTML(invite.code)}" type="button">禁用</button>` : ""}
      </div>
    </div>
    <div class="detail-grid">
      ${detailCell("id", invite.id)}
      ${detailCell("code", invite.code)}
      ${detailCell("entryType", inviteEntryLabel(invite.entryType))}
      ${detailCell("ownerUserId", invite.ownerUserId || "-")}
      ${detailCell("owner.nickname", owner.nickname || "-")}
      ${detailCell("usedCount/maxUses", `${invite.usedCount || 0}/${invite.maxUses || 0}`)}
      ${detailCell("boundWechatUserId", invite.boundWechatUserId || "-")}
      ${detailCell("bound.nickname", bound.nickname || invite.boundWechatNickname || "-")}
      ${detailCell("relations", relations.length)}
      ${detailCell("bindSources", compactList(relations.map((item) => item.bindSource)))}
    </div>
    <div class="sub-panel">
      <div class="panel-head">
        <div>
          <h2>入口校验</h2>
          <p>一期小程序只允许通过小程序卡片、二维码、链接进入；唯一邀请码绑定微信后进入登录页。</p>
        </div>
      </div>
      <div class="detail-grid profile-detail-grid">
        ${detailCell("allowedEntry", inviteEntryLabel(invite.entryType))}
        ${detailCell("uniqueBinding", Number(invite.maxUses || 0) === 1 ? "true" : "false")}
        ${detailCell("boundWechat", invite.boundWechatUserId ? "true" : "false")}
        ${detailCell("authPageMode", invite.boundWechatUserId ? "login" : "register")}
      </div>
    </div>
    <div class="sub-panel">
      <div class="panel-head">
        <div>
          <h2>绑定关系明细</h2>
          <p>用于核查 inviteCodeId、inviterUserId、inviteeUserId、bindSource 是否与小程序登录链路一致。</p>
        </div>
      </div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>inviteCodeId</th>
              <th>inviterUserId</th>
              <th>inviteeUserId</th>
              <th>bindSource</th>
              <th>入口</th>
            </tr>
          </thead>
          <tbody>${relations.map((item) => inviteRelationDetailRow(item, invite.entryType)).join("") || emptyRow(5, "暂无绑定关系")}</tbody>
        </table>
      </div>
    </div>
  `;
  const disableButton = panel.querySelector("button[data-action='detail-invite-disable']");
  if (disableButton) {
    disableButton.addEventListener("click", async () => {
      await apiPost(`/api/admin/invite-codes/${encodeURIComponent(disableButton.dataset.code)}/disable`, {});
      toast(`邀请码 ${disableButton.dataset.code} 已禁用`);
      await Promise.all([showInviteCodeDetail(disableButton.dataset.code), loadInviteCodes(new FormData($("#invite-filter-form")))]);
    });
  }
  panel.scrollIntoView({ behavior: "smooth", block: "start" });
}

async function renderGames() {
  $("#game-filter-form").insertAdjacentHTML("beforeend", `<button id="game-batch-audit-button" class="ghost" type="button">批量通过待审核</button>`);
  $("#game-batch-audit-button").addEventListener("click", batchAuditGames);
  $("#game-create-form").addEventListener("submit", createAdminGame);
  $("#game-map-search-button")?.addEventListener("click", searchGameMapPlaces);
  $("#game-map-results")?.addEventListener("click", selectGameMapPlace);
  $("#game-filter-form").addEventListener("submit", (event) => {
    event.preventDefault();
    loadGames(new FormData(event.currentTarget));
  });
  $("#games-table").addEventListener("click", onGameTableClick);
  ensureGameApplicationsPanel();
  await loadGameTypeOptionsForGames();
  await loadGames();
}

async function searchGameMapPlaces() {
  const keyword = $("#game-map-keyword")?.value.trim() || "";
  const form = $("#game-create-form");
  const city = form?.cityName?.value.trim() || "";
  const resultBox = $("#game-map-results");

  if (!keyword) {
    toast("请输入地点关键词", true);
    return;
  }

  resultBox.innerHTML = `<p class="muted">搜索中...</p>`;
  try {
    const data = await apiGet(`/api/admin/map/search${querySuffixFromObject({ keyword, city, page: 1, pageSize: 8 })}`);
    const places = Array.isArray(data.items) ? data.items : [];
    resultBox.innerHTML = places.length
      ? places.map(gameMapPlaceButton).join("")
      : `<p class="muted">未找到地点</p>`;
  } catch (error) {
    resultBox.innerHTML = "";
    toast(error.message, true);
  }
}

function gameMapPlaceButton(place, index) {
  const payload = encodeURIComponent(JSON.stringify(place || {}));
  const title = place.title || place.name || `地点 ${index + 1}`;
  const address = place.address || [place.city, place.district].filter(Boolean).join(" ") || "-";

  return `
    <button class="map-result-button" type="button" data-place="${payload}">
      <strong>${escapeHTML(title)}</strong>
      <span>${escapeHTML(address)}</span>
    </button>
  `;
}

function selectGameMapPlace(event) {
  const button = event.target.closest("button[data-place]");
  if (!button) return;

  let place = {};
  try {
    place = JSON.parse(decodeURIComponent(button.dataset.place || "{}"));
  } catch (error) {
    toast("地点数据解析失败", true);
    return;
  }

  const form = $("#game-create-form");
  if (!form) return;
  form.cityCode.value = place.cityCode || form.cityCode.value;
  form.cityName.value = place.city || place.cityName || form.cityName.value;
  form.address.value = place.address || place.title || form.address.value;
  form.longitude.value = Number.isFinite(Number(place.longitude)) ? place.longitude : form.longitude.value;
  form.latitude.value = Number.isFinite(Number(place.latitude)) ? place.latitude : form.latitude.value;
  toast("地点已填入开局表单");
}

async function loadGameTypeOptionsForGames() {
  if (!can("system_config:read")) return;
  try {
    const data = await apiGet("/api/admin/games/category-config");
    const config = data.config || data;
    const options = normalizeGameTypeOptions(config.typeFilters);
    if (!options.length) return;
    state.gameTypeOptions = options;
    renderGameTypeSelect("#game-create-type-select", options, false);
    renderGameTypeSelect("#game-filter-type-select", options, true);
  } catch (error) {
    state.gameTypeOptions = state.gameTypeOptions || [];
  }
}

function normalizeGameTypeOptions(items = []) {
  return (Array.isArray(items) ? items : [])
    .filter((item) => item && item.visible !== false && item.key && item.key !== "all")
    .sort((a, b) => Number(a.order || 0) - Number(b.order || 0))
    .map((item) => ({
      key: String(item.key),
      name: String(item.name || item.key),
      selectable: item.selectable !== false,
    }));
}

function renderGameTypeSelect(selector, options, includeAll) {
  const select = $(selector);
  if (!select) return;
  const current = select.value;
  const rows = [];
  if (includeAll) rows.push(`<option value="">全部局型</option>`);
  rows.push(...options.map((item) => `<option value="${escapeHTML(item.key)}" ${item.selectable ? "" : "disabled"}>${escapeHTML(item.name)} ${escapeHTML(item.key)}</option>`));
  select.innerHTML = rows.join("");
  if ([...select.options].some((option) => option.value === current)) {
    select.value = current;
  }
}

async function batchAuditGames() {
  const ids = state.games.filter((item) => item.status === "pending_audit").map((item) => item.id);
  if (!ids.length) {
    toast("当前列表没有待审核组局", true);
    return;
  }
  try {
    const result = await apiPost("/api/admin/games/batch-audit", { gameIds: ids, approve: true, remark: "batch audit from admin web" });
    toast(`批量通过 ${result.success || 0} 个组局，失败 ${result.failed || 0} 个`);
    await loadGames(new FormData($("#game-filter-form")));
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadGames(formData) {
  const data = await apiGet(`/api/admin/games${querySuffix(formData)}`);
  state.games = data.items || [];
  $("#games-table").innerHTML = state.games.map(gameRow).join("") || emptyRow(8, "暂无组局");
  await loadGameApplications();
}

function ensureGameApplicationsPanel() {
  if ($("#game-applications-panel")) return;
  $("#game-detail-panel").insertAdjacentHTML("beforebegin", `
    <section id="game-applications-panel" class="panel">
      <div class="panel-head">
        <div>
          <h2>入局申请</h2>
          <p>字段对齐 game_applications：gameId、userId、status、reason、fileIds。</p>
        </div>
        <form id="game-application-filter-form" class="filters">
          <select name="status">
            <option value="">全部状态</option>
            <option value="pending">pending</option>
            <option value="approved">approved</option>
            <option value="rejected">rejected</option>
            <option value="cancelled">cancelled</option>
          </select>
          <input name="gameId" inputmode="numeric" placeholder="局 ID" />
          <input name="userId" inputmode="numeric" placeholder="用户 ID" />
          <button class="ghost" type="submit">筛选</button>
        </form>
      </div>
      <div class="table-wrap">
        ${gameApplicationTable("game-applications-table")}
      </div>
    </section>
  `);
  $("#game-application-filter-form").addEventListener("submit", (event) => {
    event.preventDefault();
    loadGameApplications(new FormData(event.currentTarget));
  });
  $("#game-applications-table").addEventListener("click", async (event) => {
    const button = event.target.closest("button[data-action='download-application-file']");
    if (!button) return;
    try {
      await downloadAdminFile(button.dataset.id);
    } catch (error) {
      toast(error.message, true);
    }
  });
}

async function loadGameApplications(formData) {
  const data = can("game:read")
    ? await apiGet(`/api/admin/game-applications${querySuffix(formData)}`)
    : { items: [] };
  state.gameApplications = data.items || [];
  renderGameApplications();
}

function renderGameApplications() {
  const table = $("#game-applications-table");
  if (!table) return;
  table.innerHTML = state.gameApplications.map(gameApplicationRow).join("") || emptyRow(7, "暂无入局申请");
}

function gameApplicationTable(bodyID) {
  return `
    <table>
      <thead>
        <tr>
          <th>ID</th>
          <th>局 ID</th>
          <th>用户 ID</th>
          <th>状态</th>
          <th>申请说明</th>
          <th>材料</th>
          <th>时间</th>
        </tr>
      </thead>
      <tbody id="${escapeHTML(bodyID)}"></tbody>
    </table>
  `;
}

function gameApplicationRow(item) {
  const fileIDs = Array.isArray(item.fileIds) ? item.fileIds : [];
  const fileButtons = fileIDs.length
    ? fileIDs.map((id) => `<button class="ghost" data-action="download-application-file" data-id="${escapeHTML(id)}" type="button">材料 #${escapeHTML(id)}</button>`).join(" ")
    : "-";
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td>${escapeHTML(item.gameId)}</td>
      <td>${escapeHTML(item.userId)}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${escapeHTML(item.reason || "-")}</td>
      <td>${fileButtons}</td>
      <td>${formatDate(item.createdAt)}</td>
    </tr>
  `;
}

async function createAdminGame(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const data = Object.fromEntries(new FormData(form).entries());
  const payload = {
    title: data.title.trim(),
    creatorUserId: Number(data.creatorUserId),
    gameType: data.gameType,
    cityCode: data.cityCode.trim(),
    cityName: data.cityName.trim(),
    address: data.address.trim(),
    minPlayers: Number(data.minPlayers),
    maxPlayers: Number(data.maxPlayers),
  };
  const longitude = Number(data.longitude || 0);
  const latitude = Number(data.latitude || 0);
  if (Number.isFinite(longitude) && Number.isFinite(latitude)) {
    payload.longitude = longitude;
    payload.latitude = latitude;
  }
  const mainGuideUserId = Number(data.mainGuideUserId || 0);
  if (mainGuideUserId > 0 && mainGuideUserId === payload.creatorUserId) {
    toast("mainGuideUserId cannot equal creatorUserId", true);
    return;
  }
  if (payload.minPlayers < 5 || payload.maxPlayers > 8 || payload.minPlayers > payload.maxPlayers) {
    toast("每局人数必须为 5-8 人，满 5 人后可手动开始，最多 8 人", true);
    return;
  }
  if (mainGuideUserId > 0) payload.mainGuideUserId = mainGuideUserId;
  try {
    const game = await apiPost("/api/admin/games", payload);
    toast(`已创建 ${gameTypeLabel(game.gameType)}，ID ${game.id}`);
    form.reset();
    form.minPlayers.value = "5";
    form.maxPlayers.value = "8";
    await loadGames();
  } catch (error) {
    toast(error.message, true);
  }
}

async function onGameTableClick(event) {
  const button = event.target.closest("button[data-action]");
  if (!button) return;
  const id = button.dataset.id;
  try {
    if (button.dataset.action === "audit") {
      await apiPost(`/api/admin/games/${id}/audit`, { approve: true, remark: "后台审核通过" });
      toast(`局 ${id} 已审核通过`);
      await loadGames(new FormData($("#game-filter-form")));
    }
    if (button.dataset.action === "detail") {
      await showGameDetail(id);
    }
  } catch (error) {
    toast(error.message, true);
  }
}

async function showGameDetail(id) {
  if (!state.gameApplications.length) {
    await loadGameApplications();
  }
  const data = await apiGet(`/api/admin/games/${id}`);
  const panel = $("#game-detail-panel");
  panel.classList.remove("hidden");
  panel.innerHTML = `
    <div class="panel-head">
      <div>
        <h2>局链路 #${escapeHTML(data.game.id)}</h2>
        <p>${escapeHTML(data.game.title || "")}</p>
      </div>
      <span class="${badgeClass(data.game.status)}">${statusLabel(data.game.status)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("gameType", gameTypeLabel(data.game.gameType))}
      ${detailCell("gameSource", data.game.gameSource)}
      ${detailCell("creatorUserId", data.game.creatorUserId)}
      ${detailCell("mainGuideUserId", data.game.mainGuideUserId || "-")}
      ${detailCell("memberIds", (data.memberIds || []).join(", ") || "-")}
      ${detailCell("milestones", (data.milestones || []).length)}
      ${detailCell("checkins", (data.checkins || []).length)}
      ${detailCell("retrospectives", (data.retrospectives || []).length)}
      ${detailCell("continueDrafts", (data.continueDrafts || []).length)}
    </div>
    <div class="table-wrap">
      <h3>本局入局申请</h3>
      ${gameApplicationTable("game-detail-applications-table")}
    </div>
    ${gameOpsBlock(data.game.id, data)}
  `;
  renderGameDetailApplications(data.game.id);
  bindGameOpsPanel(panel, data.game.id);
  panel.scrollIntoView({ behavior: "smooth", block: "start" });
}

function renderGameDetailApplications(gameID) {
  const table = $("#game-detail-applications-table");
  if (!table) return;
  const items = state.gameApplications.filter((item) => String(item.gameId) === String(gameID));
  table.innerHTML = items.map(gameApplicationRow).join("") || emptyRow(7, "暂无入局申请");
}

function bindGameOpsPanel(panel, gameID) {
  const milestoneForm = panel.querySelector("[data-role='game-milestone-form']");
  if (milestoneForm) {
    milestoneForm.addEventListener("submit", async (event) => {
      event.preventDefault();
      if (!can("game:progress:manage")) {
        toast("当前角色没有进度管理权限", true);
        return;
      }
      const formData = new FormData(event.currentTarget);
      await apiPost(`/api/admin/games/${gameID}/milestones`, {
        title: String(formData.get("title") || "").trim(),
        status: String(formData.get("status") || "pending").trim(),
      });
      toast("里程碑已创建");
      await showGameDetail(gameID);
    });
  }
  panel.addEventListener("click", async (event) => {
    const fileButton = event.target.closest("button[data-action='download-application-file']");
    if (fileButton) {
      await downloadAdminFile(fileButton.dataset.id);
      return;
    }
    const button = event.target.closest("button[data-action='checkin-invalid']");
    if (!button) return;
    if (!can("game:progress:manage")) {
      toast("当前角色没有进度管理权限", true);
      return;
    }
    await apiPost(`/api/admin/game-checkins/${button.dataset.id}/mark-invalid`, { reason: "admin marked invalid" });
    toast("打卡已标记异常");
    await showGameDetail(gameID);
  });
}

async function renderAudits() {
  ensureAuditDetailPanel();
  const tasks = [];
  if (can("identity:read")) {
    tasks.push(loadIdentities());
  } else {
    renderNoAccess("#identity-list", "缺少 identity:read");
  }
  if (can("role:view")) {
    tasks.push(loadRoles());
  } else {
    renderNoAccess("#role-list", "缺少 role:view");
  }
  await Promise.all(tasks);
  $("#view-root").addEventListener("click", async (event) => {
    const button = event.target.closest("button[data-action]");
    if (!button) return;
    try {
      if (["reload-identities", "identity-detail"].includes(button.dataset.action) && !can("identity:read")) {
        toast("缺少 identity:read", true);
        return;
      }
      if (["reload-roles", "role-detail"].includes(button.dataset.action) && !can("role:view")) {
        toast("缺少 role:view", true);
        return;
      }
      if (["role-approve", "role-reject"].includes(button.dataset.action) && !can("role:update")) {
        toast("缺少 role:update", true);
        return;
      }
      if (button.dataset.action === "reload-identities") await loadIdentities();
      if (button.dataset.action === "reload-roles") await loadRoles();
      if (button.dataset.action === "identity-detail") await showIdentityDetail(Number(button.dataset.userId));
      if (button.dataset.action === "role-detail") showRoleApplicationDetail(Number(button.dataset.id));
      if (button.dataset.action === "role-approve") {
        await reviewRoleApplication(button.dataset.id, true, "后台审核通过");
        toast("角色申请已通过");
        await loadRoles();
      }
      if (button.dataset.action === "role-reject") {
        await reviewRoleApplication(button.dataset.id, false, "后台驳回：材料不完整");
        toast("角色申请已驳回");
        await loadRoles();
      }
    } catch (error) {
      toast(error.message, true);
    }
  });
}

async function loadIdentities() {
  if (!can("identity:read")) {
    renderNoAccess("#identity-list", "缺少 identity:read");
    return;
  }
  const data = await apiGet("/api/admin/identity-verifications");
  const list = $("#identity-list");
  const items = data.items || [];
  state.identities = items;
  list.innerHTML = items.map((item) => stackItem({
    title: `用户 ${item.userId}`,
    badge: item.status,
    meta: [`phone=${item.phone || "-"}`, `realname=${item.realname || "-"}`, `updated=${formatTime(item.updatedAt || item.createdAt)}`],
    action: `<button class="ghost" data-action="identity-detail" data-user-id="${item.userId}" type="button">详情</button>`,
  })).join("") || emptyBlock("暂无实名记录");
}

async function loadRoles() {
  if (!can("role:view")) {
    renderNoAccess("#role-list", "缺少 role:view");
    return;
  }
  const data = await apiGet("/api/admin/audits/role-applications");
  const list = $("#role-list");
  const items = data.items || [];
  state.roleApplications = items;
  list.innerHTML = items.map((item) => stackItem({
    title: `${roleLabel(item.roleCode)}申请 #${item.id}`,
    badge: item.status,
    meta: [`userId=${item.userId}`, `reason=${item.reason || "-"}`, `created=${formatTime(item.createdAt)}`],
    action: roleApplicationActions(item),
  })).join("") || emptyBlock("暂无角色申请");
}

function ensureAuditDetailPanel() {
  if ($("#audit-detail-panel")) return;
  $("#view-root").insertAdjacentHTML("beforeend", `<section id="audit-detail-panel" class="panel hidden"></section>`);
}

function roleApplicationActions(item) {
  const actions = [`<button class="ghost" data-action="role-detail" data-id="${item.id}" type="button">详情</button>`];
  if (item.status === "pending" && can("role:update")) {
    actions.push(`<button class="ghost" data-action="role-approve" data-id="${item.id}" type="button">通过</button>`);
    actions.push(`<button class="ghost" data-action="role-reject" data-id="${item.id}" type="button">驳回</button>`);
  }
  return actions.join("");
}

async function reviewRoleApplication(applicationID, approve, remark) {
  return apiPost(`/api/admin/audits/role-applications/${applicationID}/review`, { approve, remark });
}

async function showIdentityDetail(userID) {
  if (!userID) return;
  const item = await apiGet(`/api/admin/identity-verifications/${userID}`);
  const panel = $("#audit-detail-panel");
  panel.classList.remove("hidden");
  panel.innerHTML = `
    <div class="panel-head">
      <div>
        <h2>实名认证详情</h2>
        <p>字段对齐 identity.Record，用于后台核查实名、短信与人脸核验状态。</p>
      </div>
      <span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("userId", item.userId)}
      ${detailCell("phone", item.phone || "-")}
      ${detailCell("realname", item.realname || "-")}
      ${detailCell("status", statusLabel(item.status))}
      ${detailCell("faceIdRequestId", item.faceIdRequestId || "-")}
      ${detailCell("createdAt", formatTime(item.createdAt))}
      ${detailCell("updatedAt", formatTime(item.updatedAt))}
    </div>
  `;
  panel.scrollIntoView({ behavior: "smooth", block: "start" });
}

function showRoleApplicationDetail(applicationID) {
  const item = state.roleApplications.find((app) => Number(app.id) === Number(applicationID));
  if (!item) return;
  const panel = $("#audit-detail-panel");
  panel.classList.remove("hidden");
  panel.innerHTML = `
    <div class="panel-head">
      <div>
        <h2>角色申请详情</h2>
        <p>字段对齐 role_applications，后台审核结果会同步写入用户角色链路。</p>
      </div>
      <span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("id", item.id)}
      ${detailCell("userId", item.userId)}
      ${detailCell("roleCode", item.roleCode)}
      ${detailCell("status", statusLabel(item.status))}
      ${detailCell("reason", item.reason || "-")}
      ${detailCell("abilityDescription", item.abilityDescription || "-")}
      ${detailCell("proofFileIds", compactList(item.proofFileIds || []))}
      ${detailCell("reviewAdminId", item.reviewAdminId || "-")}
      ${detailCell("reviewRemark", item.reviewRemark || item.rejectReason || "-")}
      ${detailCell("createdAt", formatTime(item.createdAt))}
      ${detailCell("updatedAt", formatTime(item.updatedAt))}
    </div>
  `;
  panel.scrollIntoView({ behavior: "smooth", block: "start" });
}

async function renderRevenue() {
  ensureRevenueDetailPanel();
  $("#revenue-template-form").addEventListener("submit", createRevenueTemplate);
  $("#revenue-rule-form").addEventListener("submit", upsertRevenueRule);
  $("#revenue-calc-form").addEventListener("click", onRevenueCalcClick);
  $("#revenue-records-table").addEventListener("click", onRevenueRecordClick);
  await loadGameTypeOptionsForRevenue();
  const tasks = [];
  if (can("revenue:template:view")) {
    tasks.push(loadRevenueTemplates(), loadRevenueRules());
  } else {
    renderNoAccess("#revenue-template-list", "缺少 revenue:template:view");
    renderNoAccess("#revenue-rule-list", "缺少 revenue:template:view");
  }
  if (can("revenue:record:view")) {
    tasks.push(loadRevenueRecords());
  } else {
    $("#revenue-records-table").innerHTML = emptyRow(8, "无权限查看分润记录");
  }
  if (can("settlement:offline:create")) {
    tasks.push(loadRevenueSettlements());
  } else {
    $("#revenue-settlements-table").innerHTML = emptyRow(6, "无权限查看线下结算记录");
  }
  await Promise.all(tasks);
}

async function loadGameTypeOptionsForRevenue() {
  if (state.gameTypeOptions && state.gameTypeOptions.length) {
    renderGameTypeSelect("#revenue-template-type-select", state.gameTypeOptions, false);
    return;
  }
  await loadGameTypeOptionsForGames();
  if (state.gameTypeOptions && state.gameTypeOptions.length) {
    renderGameTypeSelect("#revenue-template-type-select", state.gameTypeOptions, false);
  }
}

async function createRevenueTemplate(event) {
  event.preventDefault();
  if (!can("revenue:template:update")) {
    toast("缺少 revenue:template:update", true);
    return;
  }
  const data = Object.fromEntries(new FormData(event.currentTarget).entries());
  const payload = {
    name: data.name.trim(),
    gameType: data.gameType,
    platformBps: Number(data.platformBps),
    creatorBps: Number(data.creatorBps),
    memberBps: Number(data.memberBps),
  };
  try {
    await apiPost("/api/admin/revenue/templates", payload);
    toast("分润模板已保存");
    await loadRevenueTemplates();
  } catch (error) {
    toast(error.message, true);
  }
}

async function upsertRevenueRule(event) {
  event.preventDefault();
  if (!can("revenue:template:update")) {
    toast("缺少 revenue:template:update", true);
    return;
  }
  const data = Object.fromEntries(new FormData(event.currentTarget).entries());
  const payload = {
    templateId: Number(data.templateId),
    ruleCode: data.ruleCode.trim(),
    ruleValue: data.ruleValue.trim(),
  };
  try {
    await apiPost("/api/admin/revenue/rules", payload);
    toast("分润规则已保存");
    await loadRevenueRules(payload.templateId);
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadRevenueTemplates() {
  if (!can("revenue:template:view")) {
    renderNoAccess("#revenue-template-list", "缺少 revenue:template:view");
    return;
  }
  const data = await apiGet("/api/admin/revenue/templates");
  state.revenueTemplates = data.items || [];
  $("#revenue-template-list").innerHTML = state.revenueTemplates.map((item) => stackItem({
    title: `${item.name} #${item.id}`,
    badge: item.status,
    meta: [`gameType=${item.gameType}`, `platformBps=${item.platformBps}`, `creatorBps=${item.creatorBps}`, `memberBps=${item.memberBps}`],
    action: `<button class="ghost" data-action="load-revenue-rules" data-template-id="${item.id}" type="button">查看规则</button>`,
  })).join("") || emptyBlock("暂无分润模板");
  $("#revenue-template-list").addEventListener("click", async (event) => {
    const button = event.target.closest("button[data-action='load-revenue-rules']");
    if (!button) return;
    $("#revenue-rule-form").templateId.value = button.dataset.templateId;
    await loadRevenueRules(Number(button.dataset.templateId));
  });
  if (state.revenueTemplates[0] && !$("#revenue-rule-form").templateId.value) {
    $("#revenue-rule-form").templateId.value = state.revenueTemplates[0].id;
    $("#revenue-calc-form").templateId.value = state.revenueTemplates[0].id;
  }
}

async function loadRevenueRules(templateId = numberOrZero($("#revenue-rule-form")?.templateId?.value)) {
  if (!can("revenue:template:view")) {
    renderNoAccess("#revenue-rule-list", "缺少 revenue:template:view");
    return;
  }
  const data = await apiGet(`/api/admin/revenue/rules${templateId ? `?templateId=${templateId}` : ""}`);
  $("#revenue-rule-list").innerHTML = (data.items || []).map((item) => stackItem({
    title: `${item.ruleCode} #${item.id}`,
    badge: "active",
    meta: [`templateId=${item.templateId}`, `ruleValue=${item.ruleValue}`, `created=${formatTime(item.createdAt)}`],
  })).join("") || emptyBlock("暂无分润规则");
}

async function onRevenueCalcClick(event) {
  const button = event.target.closest("button[data-action]");
  if (!button) return;
  const payload = revenueCalcPayload();
  try {
    if (button.dataset.action === "revenue-preview") {
      if (!can("revenue:simulate")) {
        toast("缺少 revenue:simulate", true);
        return;
      }
      renderRevenuePreview(await apiPost("/api/admin/revenue/preview", payload));
    }
    if (button.dataset.action === "revenue-generate") {
      if (!can("revenue:generate")) {
        toast("缺少 revenue:generate", true);
        return;
      }
      await apiPost("/api/admin/revenue/records/generate", payload);
      toast("分润记录已生成");
      await loadRevenueRecords();
    }
  } catch (error) {
    toast(error.message, true);
  }
}

function revenueCalcPayload() {
  const data = Object.fromEntries(new FormData($("#revenue-calc-form")).entries());
  return {
    gameId: Number(data.gameId),
    amountCent: Number(data.amountCent),
    templateId: Number(data.templateId),
    creatorUserId: numberOrZero(data.creatorUserId),
    memberIds: parseIDList(data.memberIds),
    expertUserId: numberOrZero(data.expertUserId),
    guideUserId: numberOrZero(data.guideUserId),
  };
}

function renderRevenuePreview(preview) {
  const panel = $("#revenue-preview-panel");
  panel.classList.remove("hidden");
  panel.innerHTML = `
    <div class="panel-head">
      <h2>试算结果</h2>
      <span class="${badgeClass(preview.canGenerateRecord ? "active" : "pending")}">${preview.canGenerateRecord ? "可生成" : "不可生成"}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("gameId", preview.gameId)}
      ${detailCell("templateId", preview.templateId)}
      ${detailCell("amountCent", preview.amountCent)}
      ${detailCell("blockReasons", (preview.blockReasons || []).join(", ") || "-")}
    </div>
    <div class="table-wrap nested-table">
      <table><thead><tr><th>role</th><th>userId</th><th>amountCent</th></tr></thead><tbody>
        ${(preview.items || []).map((item) => `<tr><td>${escapeHTML(item.role)}</td><td>${escapeHTML(item.userId || "-")}</td><td>${escapeHTML(item.amountCent)}</td></tr>`).join("")}
      </tbody></table>
    </div>
  `;
}

async function loadRevenueRecords() {
  if (!can("revenue:record:view")) {
    $("#revenue-records-table").innerHTML = emptyRow(8, "无权限查看分润记录");
    return;
  }
  const data = await apiGet("/api/admin/revenue/records");
  state.revenueRecords = data.items || [];
  $("#revenue-records-table").innerHTML = state.revenueRecords.map(revenueRecordRow).join("") || emptyRow(8, "暂无分润记录");
}

function ensureRevenueDetailPanel() {
  if ($("#revenue-record-detail-panel")) return;
  $("#view-root").insertAdjacentHTML("beforeend", `<section id="revenue-record-detail-panel" class="panel hidden"></section>`);
}

function showRevenueRecordDetail(recordID) {
  const item = state.revenueRecords.find((record) => Number(record.id) === Number(recordID));
  if (!item) return;
  const settlements = state.revenueSettlements.filter((settlement) => Number(settlement.recordId) === Number(recordID));
  const panel = $("#revenue-record-detail-panel");
  panel.classList.remove("hidden");
  panel.innerHTML = `
    <div class="panel-head">
      <div>
        <h2>分润记录详情</h2>
        <p>字段对齐 revenue_records / revenue_record_items / settlement_records。</p>
      </div>
      <span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("id", item.id)}
      ${detailCell("recordNo", item.recordNo)}
      ${detailCell("gameId", item.gameId)}
      ${detailCell("templateId", item.templateId)}
      ${detailCell("amountCent", item.amountCent)}
      ${detailCell("frozenReason", item.frozenReason || "-")}
      ${detailCell("settledAt", item.settledAt || "-")}
      ${detailCell("createdAt", formatTime(item.createdAt))}
      ${detailCell("items", (item.items || []).length)}
      ${detailCell("settlements", settlements.length)}
    </div>
    <div class="table-wrap nested-table">
      <table><thead><tr><th>role</th><th>userId</th><th>amountCent</th></tr></thead><tbody>
        ${(item.items || []).map((child) => `<tr><td>${escapeHTML(child.role)}</td><td>${escapeHTML(child.userId || "-")}</td><td>${escapeHTML(child.amountCent)}</td></tr>`).join("") || emptyRow(3, "暂无分润明细")}
      </tbody></table>
    </div>
  `;
  panel.scrollIntoView({ behavior: "smooth", block: "start" });
}

async function loadRevenueSettlements() {
  if (!can("settlement:offline:create")) {
    $("#revenue-settlements-table").innerHTML = emptyRow(6, "无权限查看线下结算记录");
    return;
  }
  const data = await apiGet("/api/admin/revenue/settlements");
  state.revenueSettlements = data.items || [];
  $("#revenue-settlements-table").innerHTML = state.revenueSettlements.map((item) => `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td>${escapeHTML(item.recordId)}</td>
      <td>${escapeHTML(item.method)}</td>
      <td>${escapeHTML(item.proofNo || "-")}</td>
      <td>${escapeHTML(item.amountCent)}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `).join("") || emptyRow(6, "暂无结算记录");
}

async function onRevenueRecordClick(event) {
  const button = event.target.closest("button[data-action]");
  if (!button) return;
  try {
    if (button.dataset.action === "revenue-detail") {
      showRevenueRecordDetail(button.dataset.id);
      return;
    }
    if (button.dataset.action === "revenue-freeze") {
      if (!can("revenue:freeze")) {
        toast("缺少 revenue:freeze", true);
        return;
      }
      await apiPost(`/api/admin/revenue/records/${button.dataset.id}/freeze`, { reason: "admin_freeze" });
      toast("分润记录已冻结");
    }
    if (button.dataset.action === "revenue-settle") {
      if (!can("settlement:offline:create")) {
        toast("缺少 settlement:offline:create", true);
        return;
      }
      await apiPost(`/api/admin/revenue/records/${button.dataset.id}/settle`, { method: "offline", proofNo: `OFF-${button.dataset.id}` });
      toast("线下结算已登记");
    }
    await Promise.all([
      can("revenue:record:view") ? loadRevenueRecords() : Promise.resolve(),
      can("settlement:offline:create") ? loadRevenueSettlements() : Promise.resolve(),
    ]);
  } catch (error) {
    toast(error.message, true);
  }
}

async function renderRedemption() {
  $("#redemption-item-form").addEventListener("submit", createRedemptionItem);
  $("#redemption-items-table").addEventListener("click", onRedemptionItemClick);
  $("#redemption-orders-table").addEventListener("click", onRedemptionOrderClick);
  $("#redemption-orders-refresh").addEventListener("click", loadRedemptionOrders);
  $("#points-logs-refresh").addEventListener("click", loadPointsLogs);
  const tasks = [];
  if (can("redemption:manage")) {
    tasks.push(loadRedemptionItems(), loadRedemptionOrders());
  } else {
    $("#redemption-items-table").innerHTML = emptyRow(7, "无权限管理积分兑换商品");
    $("#redemption-orders-table").innerHTML = emptyRow(8, "无权限管理积分兑换订单");
  }
  if (can("points:read")) {
    tasks.push(loadPointsLogs());
  } else {
    $("#points-logs-table").innerHTML = emptyRow(7, "无权限查看积分流水");
  }
  await Promise.all(tasks);
}

async function createRedemptionItem(event) {
  event.preventDefault();
  if (!can("redemption:manage")) {
    toast("缺少 redemption:manage", true);
    return;
  }
  const form = event.currentTarget;
  const data = Object.fromEntries(new FormData(form).entries());
  const payload = {
    name: data.name.trim(),
    description: data.description.trim(),
    pointsCost: Number(data.pointsCost),
    stock: Number(data.stock),
  };
  try {
    await apiPost("/api/admin/redemption/items", payload);
    toast("兑换商品已新增");
    form.reset();
    form.pointsCost.value = "10";
    form.stock.value = "1";
    await loadRedemptionItems();
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadRedemptionItems() {
  if (!can("redemption:manage")) {
    $("#redemption-items-table").innerHTML = emptyRow(7, "无权限管理积分兑换商品");
    return;
  }
  const data = await apiGet("/api/admin/redemption/items");
  state.redemptionItems = data.items || [];
  $("#redemption-items-table").innerHTML = state.redemptionItems.map(redemptionItemRow).join("") || emptyRow(7, "暂无兑换商品");
}

async function loadRedemptionOrders() {
  if (!can("redemption:manage")) {
    $("#redemption-orders-table").innerHTML = emptyRow(8, "无权限管理积分兑换订单");
    return;
  }
  const data = await apiGet("/api/admin/redemption/orders");
  state.redemptionOrders = data.items || [];
  $("#redemption-orders-table").innerHTML = state.redemptionOrders.map(redemptionOrderRow).join("") || emptyRow(8, "暂无兑换订单");
}

async function loadPointsLogs() {
  if (!can("points:read")) {
    $("#points-logs-table").innerHTML = emptyRow(7, "无权限查看积分流水");
    return;
  }
  const data = await apiGet("/api/admin/points/logs");
  state.pointsLogs = data.items || [];
  $("#points-logs-table").innerHTML = state.pointsLogs.map(pointsLogRow).join("") || emptyRow(7, "暂无积分流水");
}

async function onRedemptionItemClick(event) {
  const button = event.target.closest("button[data-action]");
  if (!button) return;
  if (!can("redemption:manage")) {
    toast("缺少 redemption:manage", true);
    return;
  }
  const row = button.closest("tr");
  const payload = {
    name: row.querySelector("[data-field='name']").value.trim(),
    description: row.querySelector("[data-field='description']").value.trim(),
    pointsCost: Number(row.querySelector("[data-field='pointsCost']").value),
    stock: Number(row.querySelector("[data-field='stock']").value),
    status: button.dataset.status || row.querySelector("[data-field='status']").value,
  };
  try {
    await apiPut(`/api/admin/redemption/items/${button.dataset.id}`, payload);
    toast("兑换商品已更新");
    await loadRedemptionItems();
  } catch (error) {
    toast(error.message, true);
  }
}

async function onRedemptionOrderClick(event) {
  const button = event.target.closest("button[data-action]");
  if (!button) return;
  if (!can("redemption:manage")) {
    toast("缺少 redemption:manage", true);
    return;
  }
  const approve = button.dataset.action === "redemption-order-approve";
  const payload = button.dataset.status
    ? { status: button.dataset.status, reason: button.dataset.reason || "admin review" }
    : { approve, reason: approve ? "admin approve" : "admin reject" };
  try {
    await apiPost(`/api/admin/redemption/orders/${button.dataset.id}/review`, payload);
    toast("兑换订单已处理");
    await Promise.all([loadRedemptionOrders(), loadPointsLogs()]);
  } catch (error) {
    toast(error.message, true);
  }
}

async function renderMembers() {
  $("#member-teams-refresh").addEventListener("click", () => {
    if (can("team:read")) {
      loadMemberTeams();
      return;
    }
    toast("当前角色没有团队查看权限", true);
  });
  $("#member-reports-refresh").addEventListener("click", () => {
    if (can("member_report:read")) {
      loadMemberReports();
      return;
    }
    toast("当前角色没有会员报表权限", true);
  });
  $("#member-teams-table").addEventListener("click", onMemberTeamClick);

  const tasks = [];
  if (can("team:read")) {
    tasks.push(loadMemberTeams());
  } else {
    $("#member-teams-table").innerHTML = emptyRow(6, "当前角色没有团队查看权限");
  }
  if (can("member_report:read")) {
    tasks.push(loadMemberReports());
  } else {
    $("#member-reports-table").innerHTML = emptyRow(7, "当前角色没有会员报表权限");
  }
  await Promise.all(tasks);
}

async function loadMemberTeams() {
  const data = await apiGet("/api/admin/teams");
  state.memberTeams = data.items || [];
  $("#member-teams-table").innerHTML = state.memberTeams.map(memberTeamRow).join("") || emptyRow(6, "暂无会员团队");
}

async function loadMemberReports() {
  const data = await apiGet("/api/admin/member-reports");
  state.memberReports = data.items || [];
  $("#member-reports-table").innerHTML = state.memberReports.map(memberReportRow).join("") || emptyRow(7, "暂无会员报表");
}

async function onMemberTeamClick(event) {
  const button = event.target.closest("button[data-action='member-team-detail']");
  if (!button) return;
  try {
    await showMemberTeamDetail(button.dataset.id);
  } catch (error) {
    toast(error.message, true);
  }
}

async function showMemberTeamDetail(teamID) {
  const data = await apiGet(`/api/admin/teams/${teamID}`);
  const team = data.team || {};
  const members = data.members || [];
  const revenue = data.revenueSummary || {};
  const panel = $("#member-team-detail-panel");
  panel.classList.remove("hidden");
  panel.innerHTML = `
    <div class="panel-head">
      <div>
        <h2>团队详情 #${escapeHTML(team.id || teamID)}</h2>
        <p>成员、层级和收益汇总来自团队后端服务。</p>
      </div>
      <span class="${badgeClass(team.status)}">${statusLabel(team.status)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("leaderUserId", team.leaderUserId || "-")}
      ${detailCell("name", team.name || "-")}
      ${detailCell("status", team.status || "-")}
      ${detailCell("memberCount", revenue.memberCount || members.length)}
      ${detailCell("totalCent", revenue.totalCent || 0)}
      ${detailCell("pendingCent", revenue.pendingCent || 0)}
      ${detailCell("settledCent", revenue.settledCent || 0)}
      ${detailCell("createdAt", formatTime(team.createdAt))}
    </div>
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>userId</th>
            <th>relationLevel</th>
            <th>source</th>
            <th>status</th>
            <th>joinedAt</th>
          </tr>
        </thead>
        <tbody>
          ${members.map(memberTeamMemberRow).join("") || emptyRow(5, "暂无团队成员")}
        </tbody>
      </table>
    </div>
  `;
  panel.scrollIntoView({ behavior: "smooth", block: "start" });
}

async function renderProfiles() {
  $("#connections-refresh").addEventListener("click", () => {
    if (can("connection:read")) {
      loadConnections();
      return;
    }
    toast("当前角色没有人脉关系查看权限", true);
  });
  $("#expert-profile-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!can("profile:read")) {
      toast("当前角色没有画像查看权限", true);
      return;
    }
    await loadExpertProfile(new FormData(event.currentTarget).get("userId"));
  });
  $("#guide-profile-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!can("profile:read")) {
      toast("当前角色没有画像查看权限", true);
      return;
    }
    await loadGuideProfile(new FormData(event.currentTarget).get("userId"));
  });
  $("#expert-profile-panel").innerHTML = emptyBlock("输入行家 userId 后查询技能画像");
  $("#guide-profile-panel").innerHTML = emptyBlock("输入领路人 userId 后查询资源画像");
  if (can("connection:read")) {
    await loadConnections();
  } else {
    $("#connections-table").innerHTML = emptyRow(7, "当前角色没有人脉关系查看权限");
  }
}

async function loadConnections() {
  const data = await apiGet("/api/admin/connections");
  state.connections = data.items || [];
  $("#connections-table").innerHTML = state.connections.map(connectionRow).join("") || emptyRow(7, "暂无人脉关系");
}

async function loadExpertProfile(userID) {
  const id = String(userID || "").trim();
  if (!id) return;
  const data = await apiGet(`/api/admin/experts/${encodeURIComponent(id)}/skills`);
  state.expertProfile = data;
  $("#expert-profile-panel").innerHTML = profileDetailBlock("行家技能", data, [
    ["userId", data.userId || id],
    ["skillTree", compactList(data.skillTree)],
    ["serviceTags", compactList(data.serviceTags)],
    ["caseFileIds", compactList(data.caseFileIds)],
    ["completeness", `${data.completeness || 0}%`],
    ["updatedAt", formatTime(data.updatedAt)],
  ]);
}

async function loadGuideProfile(userID) {
  const id = String(userID || "").trim();
  if (!id) return;
  const data = await apiGet(`/api/admin/guides/${encodeURIComponent(id)}/resources`);
  state.guideProfile = data;
  $("#guide-profile-panel").innerHTML = profileDetailBlock("领路人资源", data, [
    ["userId", data.userId || id],
    ["resourceTags", compactList(data.resourceTags)],
    ["industryTags", compactList(data.industryTags)],
    ["cityCodes", compactList(data.cityCodes)],
    ["connectionScale", data.connectionScale || "-"],
    ["completeness", `${data.completeness || 0}%`],
    ["updatedAt", formatTime(data.updatedAt)],
  ]);
}

async function renderGrowth() {
  $("#user-growth-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!can("user:view")) {
      toast("当前角色没有用户成长查看权限", true);
      return;
    }
    await loadUserGrowthTrace(new FormData(event.currentTarget).get("userId"));
  });
  $("#game-review-trace-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!can("game:view")) {
      toast("当前角色没有局评价查看权限", true);
      return;
    }
    await loadGameReviewTrace(new FormData(event.currentTarget).get("gameId"));
  });
  $("#user-growth-panel").innerHTML = emptyBlock("输入 userId 后查询成长、评价和信用轨迹");
  $("#game-review-trace-panel").innerHTML = emptyBlock("输入 gameId 后查询该局评价和信用证据");
}

async function loadUserGrowthTrace(userID) {
  const id = String(userID || "").trim();
  if (!id) return;
  const data = await apiGet(`/api/admin/users/${encodeURIComponent(id)}/growth`);
  state.userGrowthTrace = data;
  $("#user-growth-panel").innerHTML = growthTraceBlock("用户成长", data);
}

async function loadGameReviewTrace(gameID) {
  const id = String(gameID || "").trim();
  if (!id) return;
  const data = await apiGet(`/api/admin/games/${encodeURIComponent(id)}/review-trace`);
  state.gameReviewTrace = data;
  $("#game-review-trace-panel").innerHTML = growthTraceBlock("局评价", data);
}

async function renderReports() {
  $("#reports-table").closest(".panel").querySelector(".panel-head").insertAdjacentHTML("beforeend", `<button id="reports-batch-handle-button" class="ghost" type="button">批量处理</button>`);
  $("#reports-batch-handle-button").addEventListener("click", batchHandleReports);
  $("#reports-table").addEventListener("click", onReportsTableClick);
  await loadReports();
  if (can("feedback:view")) await loadFeedbackRecords();
}

async function batchHandleReports() {
  const ids = state.reports
    .filter((item) => item.status === "pending" || item.status === "assigned")
    .map((item) => item.id);
  if (!ids.length) {
    toast("当前列表没有待处理举报", true);
    return;
  }
  try {
    const result = await apiPost("/api/admin/reports/batch-handle", { reportIds: ids, action: "handle", adminId: adminID() || 1, result: "举报属实，已完成平台处理", outcome: "confirmed", rewardPoints: 20, creditDeduct: 10 });
    toast(`批量处理 ${result.success || 0} 条举报，失败 ${result.failed || 0} 条`);
    await loadReports();
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadReports() {
  const data = await apiGet("/api/admin/reports");
  state.reports = data.items || [];
  $("#reports-table").innerHTML = state.reports.map(reportRow).join("") || emptyRow(8, "暂无举报申诉");
}

async function loadFeedbackRecords() {
  const data = await apiGet("/api/admin/feedback-records");
  state.feedbackRecords = data.items || [];
  const panel = $("#report-detail-panel");
  panel.classList.remove("hidden");
  panel.innerHTML = `
    <div class="panel-head">
      <div>
        <h2>用户反馈</h2>
        <p>来自小程序个人中心的反馈记录，可在此回复并同步到用户端</p>
      </div>
      <span class="badge">${escapeHTML(data.total || 0)} 条</span>
    </div>
    <table>
      <thead>
        <tr>
          <th>ID</th>
          <th>用户</th>
          <th>类型</th>
          <th>状态</th>
          <th>内容</th>
          <th>时间</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>${state.feedbackRecords.map(feedbackRow).join("") || emptyRow(7, "暂无用户反馈")}</tbody>
    </table>
  `;
  panel.onclick = onFeedbackTableClick;
}

async function onFeedbackTableClick(event) {
  const button = event.target.closest("button[data-action]");
  if (!button || button.dataset.action !== "feedback-reply") return;
  const content = window.prompt("回复内容", "客服已收到反馈，正在跟进处理。");
  if (!content || !content.trim()) return;
  try {
    await apiPost(`/api/admin/feedback-records/${encodeURIComponent(button.dataset.id)}/reply`, {
      userId: Number(button.dataset.userId),
      content: content.trim(),
      status: "processing",
    });
    toast("反馈已回复并同步到用户端");
    await loadFeedbackRecords();
  } catch (error) {
    toast(error.message, true);
  }
}

async function onReportsTableClick(event) {
  const button = event.target.closest("button[data-action]");
  if (!button) return;
  try {
    if (button.dataset.action === "report-detail") await showReportDetail(button.dataset.id);
    if (button.dataset.action === "report-assign") {
      await apiPost(`/api/admin/reports/${button.dataset.id}/assign`, { adminId: adminID(), handlerAdminId: adminID() || 1 });
      toast("举报已分配处理人");
      await loadReports();
    }
    if (button.dataset.action === "report-handle") {
      await apiPost(`/api/admin/reports/${button.dataset.id}/handle`, { adminId: adminID(), result: "后台已处理" });
      toast("举报已处理");
      await loadReports();
    }
    if (button.dataset.action === "credit-appeal-approve") {
      await apiPost(`/api/admin/reports/${button.dataset.id}/handle`, { adminId: adminID(), result: "信用申诉通过，已补回对应信用分", outcome: "appeal_approved" });
      toast("信用申诉已通过");
      await loadReports();
    }
    if (button.dataset.action === "credit-appeal-reject") {
      await apiPost(`/api/admin/reports/${button.dataset.id}/handle`, { adminId: adminID(), result: "信用申诉驳回，维持原处理结果", outcome: "appeal_rejected" });
      toast("信用申诉已驳回");
      await loadReports();
    }
    if (button.dataset.action === "report-close") {
      await apiPost(`/api/admin/reports/${button.dataset.id}/close`, { adminId: adminID(), result: "后台关闭" });
      toast("举报已关闭");
      await loadReports();
    }
  } catch (error) {
    toast(error.message, true);
  }
}

async function showReportDetail(id) {
  const data = await apiGet(`/api/admin/reports/${id}`);
  const report = data.report || {};
  const evidence = data.evidence || {};
  const fileID = evidence.ids?.fileId || 0;
  const file = evidence.file || {};
  const game = evidence.game || {};
  const review = evidence.review || {};
  const chatMessages = evidence.chatMessages || [];
  const fileAction = fileID > 0
    ? `<button class="ghost" data-action="admin-file-download" data-file-id="${escapeHTML(fileID)}" type="button">下载附件</button>`
    : "";
  const panel = $("#report-detail-panel");
  panel.classList.remove("hidden");
  panel.innerHTML = `
    <div class="panel-head">
      <div>
        <h2>举报详情 #${escapeHTML(report.id)}</h2>
        <p>${escapeHTML(report.content || "")}</p>
      </div>
      <span class="${badgeClass(report.status)}">${statusLabel(report.status)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("gameId", report.gameId)}
      ${detailCell("reporterUserId", report.reporterUserId)}
      ${detailCell("targetUserId", report.targetUserId || "-")}
      ${detailCell("reportType", report.reportType)}
      ${detailCell("gameTitle", game.title || "-")}
      ${detailCell("reviewId", evidence.ids?.reviewId || "-")}
      ${detailCell("reviewScore", review.score || "-")}
      ${detailCell("revenueFrozen", report.revenueFrozen)}
      ${detailCell("chatMessages", chatMessages.length)}
      ${detailCell("fileId", evidence.ids?.fileId || "-")}
      ${detailCell("file.bizType", file.bizType || "-")}
      ${detailCell("file.name", file.fileName || "-")}
      ${detailCell("revenueRecordId", evidence.ids?.revenueRecordId || "-")}
    </div>
    ${review.id ? `
      <div class="detail-block">
        <h3>评价证据</h3>
        <p>评分：${escapeHTML(review.score || "-")} / 评价人：${escapeHTML(review.reviewerUserId || "-")} / 被评价人：${escapeHTML(review.targetUserId || "-")}</p>
        <p>${escapeHTML(review.content || "无评价文字")}</p>
      </div>
    ` : ""}
    ${chatMessages.length ? `
      <div class="detail-block">
        <h3>聊天证据</h3>
        ${chatMessages.slice(0, 5).map((message) => `<p>#${escapeHTML(message.id)} ${escapeHTML(message.messageType || "text")}：${escapeHTML(message.content || "")}</p>`).join("")}
      </div>
    ` : ""}
    ${fileAction ? `<div class="actions">${fileAction}</div>` : ""}
  `;
  panel.querySelector("button[data-action='admin-file-download']")?.addEventListener("click", (event) => {
    downloadAdminFile(event.currentTarget.dataset.fileId);
  });
  panel.scrollIntoView({ behavior: "smooth", block: "start" });
}

async function renderExports() {
  ensureExportDetailPanel();
  ensureExportFilters();
  $("#export-create-form").addEventListener("submit", createExportTask);
  $("#export-filter-form").addEventListener("submit", (event) => {
    event.preventDefault();
    loadExportTasks(new FormData(event.currentTarget));
  });
  $("#export-runner-button").addEventListener("click", runExportTasks);
  $("#export-tasks-table").addEventListener("click", onExportTaskClick);
  await Promise.all([loadExportTemplates(), loadExportTasks()]);
}

async function loadExportTemplates() {
  const data = await apiGet("/api/admin/reports/export-templates");
  state.exportTemplates = data.items || [];
  $("#export-template-list").innerHTML = state.exportTemplates.map((item) => stackItem({
    title: `${item.name} / ${item.code}`,
    badge: item.enabled ? "active" : "disabled",
    meta: [`exportType=${item.exportType}`, `fileName=${item.fileName}`, `columns=${(item.columns || []).join(",")}`],
    action: `<button class="ghost" data-action="select-export-template" data-code="${escapeHTML(item.code)}" type="button">选择</button>`,
  })).join("") || emptyBlock("暂无导出模板");
  $("#export-template-list").addEventListener("click", (event) => {
    const button = event.target.closest("button[data-action='select-export-template']");
    if (!button) return;
    $("#export-create-form").templateCode.value = button.dataset.code;
  });
}

async function createExportTask(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const data = Object.fromEntries(new FormData(form).entries());
  try {
    const filters = data.filters.trim() ? JSON.parse(data.filters) : {};
    await apiPost("/api/admin/reports/export", { templateCode: data.templateCode.trim(), filters });
    toast("导出任务已创建");
    await loadExportTasks();
  } catch (error) {
    toast(error.message, true);
  }
}

async function runExportTasks() {
  try {
    await apiPost("/api/internal/reports/export-runner", {});
    toast("导出 runner 已执行");
    await loadExportTasks();
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadExportTasks(formData) {
  const data = await apiGet(`/api/admin/export-tasks${querySuffix(formData)}`);
  state.exportTasks = data.items || [];
  $("#export-tasks-table").innerHTML = state.exportTasks.map(exportTaskRow).join("") || emptyRow(8, "暂无导出任务");
}

function ensureExportFilters() {
  if ($("#export-filter-form")) return;
  $("#export-tasks-table").closest(".panel").querySelector(".panel-head").insertAdjacentHTML("beforeend", `
    <form id="export-filter-form" class="filters">
      <input name="templateCode" placeholder="templateCode" />
      <input name="exportType" placeholder="exportType" />
      <select name="status">
        <option value="">全部状态</option>
        <option value="pending">pending</option>
        <option value="running">running</option>
        <option value="done">done</option>
        <option value="failed">failed</option>
      </select>
      <button class="ghost" type="submit">筛选</button>
    </form>
  `);
}

function ensureExportDetailPanel() {
  if ($("#export-task-detail-panel")) return;
  $("#view-root").insertAdjacentHTML("beforeend", `<section id="export-task-detail-panel" class="panel hidden"></section>`);
}

async function showExportTaskDetail(taskID) {
  const item = state.exportTasks.find((task) => Number(task.id) === Number(taskID));
  if (!item) return;
  let downloadURL = "";
  if (item.status === "done" && item.fileId) {
    try {
      const result = await apiGet(`/api/admin/export-tasks/${item.id}/download-url`);
      downloadURL = typeof result === "string" ? result : JSON.stringify(result);
    } catch (error) {
      downloadURL = error.message;
    }
  }
  const panel = $("#export-task-detail-panel");
  panel.classList.remove("hidden");
  panel.innerHTML = `
    <div class="panel-head">
      <div>
        <h2>导出任务详情</h2>
        <p>字段对齐 export_tasks，用于核查固定格式报表生成与下载状态。</p>
      </div>
      <span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("id", item.id)}
      ${detailCell("taskNo", item.taskNo)}
      ${detailCell("templateCode", item.templateCode)}
      ${detailCell("exportType", item.exportType)}
      ${detailCell("fileId", item.fileId || "-")}
      ${detailCell("createdBy", item.createdBy || "-")}
      ${detailCell("createdAt", formatTime(item.createdAt))}
      ${detailCell("finishedAt", item.finishedAt || "-")}
      ${detailCell("failReason", item.failReason || "-")}
      ${detailCell("filters", compactJSON(item.filters || {}))}
      ${detailCell("downloadUrl", downloadURL || "-")}
    </div>
  `;
  panel.scrollIntoView({ behavior: "smooth", block: "start" });
}

async function onExportTaskClick(event) {
  const button = event.target.closest("button[data-action]");
  if (!button) return;
  try {
    if (button.dataset.action === "export-detail") {
      await showExportTaskDetail(button.dataset.id);
    }
    if (button.dataset.action === "export-download") {
      const url = await apiGet(`/api/admin/export-tasks/${button.dataset.id}/download-url`);
      toast(`下载地址：${typeof url === "string" ? url : JSON.stringify(url)}`);
    }
  } catch (error) {
    toast(error.message, true);
  }
}

async function renderAnalytics() {
  $("#funnel-filter-form").addEventListener("submit", (event) => {
    event.preventDefault();
    loadFunnel(new FormData(event.currentTarget));
  });
  $("#retention-refresh").addEventListener("click", loadRetention);
  $("#behavior-filter-form").addEventListener("submit", (event) => {
    event.preventDefault();
    loadBehaviorEvents(new FormData(event.currentTarget));
  });
  $("#behavior-logs-refresh").addEventListener("click", loadBehaviorLogs);
  $("#ai-snapshot-refresh").addEventListener("click", loadAIDataSnapshot);
  $("#ai-fixture-button").addEventListener("click", seedAIDataFixture);
  $("#ai-im-export-button").addEventListener("click", exportAIIMData);

  const tasks = [];
  if (can("analytics:funnel:view")) {
    tasks.push(loadFunnel());
  } else {
    $("#analytics-funnel-table").innerHTML = emptyRow(4, "当前角色没有漏斗分析权限");
  }
  if (can("analytics:retention:view")) {
    tasks.push(loadRetention());
  } else {
    $("#analytics-retention-table").innerHTML = emptyRow(5, "当前角色没有留存分析权限");
  }
  if (can("data:behavior:read")) {
    tasks.push(loadBehaviorEvents());
  } else {
    $("#behavior-events-table").innerHTML = emptyRow(7, "当前角色没有行为事件权限");
  }
  if (can("analytics:timeline:view")) {
    tasks.push(loadBehaviorLogs());
  } else {
    $("#behavior-log-list").innerHTML = emptyBlock("当前角色没有行为时间线权限");
  }
  if (can("ai:data:read")) {
    tasks.push(loadAIDataSnapshot());
  } else {
    $("#ai-snapshot-panel").innerHTML = detailCell("permission", "当前角色没有 AI 数据权限");
    $("#ai-readiness-list").innerHTML = emptyBlock("当前角色没有 AI 数据权限");
  }
  if (!can("ai:data:export")) {
    $("#ai-im-export-panel").innerHTML = emptyBlock("当前角色没有 AI IM 导出权限");
  }
  await Promise.all(tasks);
}

async function loadFunnel(formData) {
  const data = await apiGet(`/api/admin/analytics/funnel${querySuffix(formData)}`);
  state.funnel = data;
  $("#analytics-funnel-table").innerHTML = (data.steps || []).map(funnelRow).join("") || emptyRow(4, "暂无漏斗数据");
}

async function loadRetention() {
  const data = await apiGet("/api/admin/analytics/retention");
  state.retention = data;
  $("#analytics-retention-table").innerHTML = (data.buckets || []).map(retentionRow).join("") || emptyRow(5, "暂无留存数据");
}

async function loadBehaviorEvents(formData) {
  const data = await apiGet(`/api/admin/behavior/events${querySuffix(formData)}`);
  state.behaviorEvents = data.items || [];
  $("#behavior-events-table").innerHTML = state.behaviorEvents.map(behaviorEventRow).join("") || emptyRow(7, "暂无行为事件");
}

async function loadBehaviorLogs() {
  const data = await apiGet("/api/admin/behavior-logs");
  state.behaviorLogs = data.items || [];
  $("#behavior-log-list").innerHTML = state.behaviorLogs.slice(0, 20).map(behaviorLogItem).join("") || emptyBlock("暂无行为时间线");
}

async function loadAIDataSnapshot() {
  const data = await apiGet("/api/admin/ai-data/snapshot");
  state.aiSnapshot = data;
  setField("aiUserCount", data.userCount || 0);
  setField("aiBehaviorCount", data.behaviorLogCount || 0);
  setField("aiReviewCount", data.reviewCount || 0);
  setField("aiAcceptanceReady", data.acceptanceReady ? "是" : "否");
  renderAIDataSnapshot(data);
}

async function seedAIDataFixture() {
  if (!can("ai:data:seed")) {
    toast("当前角色没有 AI 验收数据生成权限", true);
    return;
  }
  try {
    const result = await apiPost("/api/admin/ai-data/acceptance-fixture", {});
    toast(`验收数据已生成：users=${result.usersVerified || 0} games=${result.gamesCreated || 0}`);
    await loadAIDataSnapshot();
    if (can("data:behavior:read")) await loadBehaviorEvents();
  } catch (error) {
    toast(error.message, true);
  }
}

async function exportAIIMData() {
  if (!can("ai:data:export")) {
    toast("当前角色没有 AI IM 导出权限", true);
    return;
  }
  try {
    const result = await apiPost("/api/admin/ai-data/im-export", {});
    const items = result.items || [];
    renderAIIMExport(items);
    toast(`AI IM 数据已导出：${items.length} 条`);
  } catch (error) {
    $("#ai-im-export-panel").innerHTML = emptyBlock(`导出失败：${error.message}`);
    toast(error.message, true);
  }
}

function renderAIIMExport(items) {
  const sample = items.slice(0, 8);
  $("#ai-im-export-panel").innerHTML = [
    stackItem({
      title: `IM export items=${items.length}`,
      badge: items.length ? "ready" : "empty",
      meta: ["仅展示前 8 条样例，完整数据以接口返回为准"],
    }),
    ...sample.map((item) => stackItem({
      title: `message ${item.messageId || item.id || "-"}`,
      badge: item.messageType || item.type || "message",
      meta: [
        `room=${item.roomId || "-"}`,
        `game=${item.gameId || "-"}`,
        `sender=${item.senderUserId || item.senderId || "-"}`,
        `createdAt=${item.createdAt || "-"}`,
        `content=${String(item.content || "").slice(0, 80) || "-"}`,
      ],
    })),
  ].join("");
}

function renderAIDataSnapshot(data) {
  const checks = data.acceptanceChecks || {};
  $("#ai-snapshot-panel").innerHTML = `
    ${detailCell("dataReady", data.dataReady ? "true" : "false")}
    ${detailCell("acceptanceReady", data.acceptanceReady ? "true" : "false")}
    ${detailCell("gameCount", data.gameCount || 0)}
    ${detailCell("favoriteCount", data.favoriteCount || 0)}
    ${detailCell("connectionCount", data.connectionCount || 0)}
    ${detailCell("footprintCount", data.footprintCount || 0)}
    ${detailCell("imMessageCount", data.imMessageCount || 0)}
    ${detailCell("imExportEnabled", data.imExportEnabled ? "true" : "false")}
    ${detailCell("checks.users", checkText(checks.users))}
    ${detailCell("checks.games", checkText(checks.games))}
    ${detailCell("checks.behaviorLogs", checkText(checks.behaviorLogs))}
    ${detailCell("checks.favorites", checkText(checks.favorites))}
    ${detailCell("checks.reviews", checkText(checks.reviews))}
  `;
  const sections = data.dataReadinessSections || {};
  $("#ai-readiness-list").innerHTML = Object.entries(sections).map(([key, item]) => stackItem({
    title: key,
    badge: item.ready ? "ready" : "blocked",
    meta: [`count=${item.count || 0}`, `ready=${item.ready ? "true" : "false"}`],
  })).join("") || emptyBlock("暂无数据准备分区");
}

async function renderDelivery() {
  $("#wechat-tasks-refresh").addEventListener("click", loadWechatTasks);
  $("#wechat-templates-refresh").addEventListener("click", loadWechatTemplates);
  $("#wechat-send-pending").addEventListener("click", sendPendingWechatTasks);
  $("#wechat-tasks-table").addEventListener("click", onWechatTaskClick);
  $("#delivery-document-form").addEventListener("submit", createDeliveryDocument);
  $("#test-case-form").addEventListener("submit", createTestCase);
  $("#test-run-form").addEventListener("submit", createTestRun);
  $("#test-runs-table").addEventListener("click", onTestRunTableClick);

  const tasks = [];
  if (can("notification:wechat:view")) {
    tasks.push(loadWechatTasks(), loadWechatTemplates());
  } else {
    $("#wechat-tasks-table").innerHTML = emptyRow(8, "当前角色没有微信订阅消息权限");
    $("#wechat-template-list").innerHTML = emptyBlock("当前角色没有微信订阅消息权限");
  }
  if (can("delivery:manage")) {
    tasks.push(loadDeliveryDocuments());
  } else {
    $("#delivery-documents-table").innerHTML = emptyRow(6, "当前角色没有交付文档权限");
  }
  if (can("testcase:read")) {
    tasks.push(loadTestCases(), loadTestRuns());
  } else {
    $("#test-cases-table").innerHTML = emptyRow(5, "当前角色没有测试用例权限");
    $("#test-runs-table").innerHTML = emptyRow(6, "当前角色没有测试运行权限");
  }
  await Promise.all(tasks);
}

async function loadWechatTasks() {
  if (!can("notification:wechat:view")) {
    $("#wechat-tasks-table").innerHTML = emptyRow(8, "当前角色没有微信订阅消息权限");
    return;
  }
  const data = await apiGet("/api/admin/notifications/wechat-tasks");
  state.wechatTasks = data.items || [];
  $("#wechat-tasks-table").innerHTML = state.wechatTasks.map(wechatTaskRow).join("") || emptyRow(8, "暂无微信订阅任务");
}

async function loadWechatTemplates() {
  if (!can("notification:wechat:view")) {
    $("#wechat-template-list").innerHTML = emptyBlock("当前角色没有微信订阅消息权限");
    return;
  }
  const data = await apiGet("/api/admin/notifications/wechat-templates");
  state.wechatTemplates = data.items || [];
  $("#wechat-template-list").innerHTML = state.wechatTemplates.map((item) => stackItem({
    title: `${item.title || item.scene}`,
    badge: item.status,
    meta: [`scene=${item.scene}`, `templateId=${item.templateId}`],
  })).join("") || emptyBlock("暂无订阅模板");
}

async function sendPendingWechatTasks() {
  if (!can("notification:wechat:view")) {
    toast("当前角色没有微信订阅消息权限", true);
    return;
  }
  try {
    const result = await apiPost("/api/internal/notifications/wechat-tasks/send-pending", { limit: 20 });
    toast(`批量发送完成：sent=${result.sent || 0} failed=${result.failed || 0} skipped=${result.skipped || 0}`);
    await loadWechatTasks();
  } catch (error) {
    toast(error.message, true);
  }
}

async function onWechatTaskClick(event) {
  const button = event.target.closest("button[data-action]");
  if (!button) return;
  if (!can("notification:wechat:view")) {
    toast("当前角色没有微信订阅消息权限", true);
    return;
  }
  try {
    if (button.dataset.action === "wechat-send") {
      await apiPost(`/api/internal/notifications/wechat-tasks/${button.dataset.id}/send`, {});
      toast("订阅消息任务已发送");
    }
    if (button.dataset.action === "wechat-mark-sent") {
      await apiPost(`/api/internal/notifications/wechat-tasks/${button.dataset.id}/mark-sent`, { resultCode: "admin", resultMessage: "admin marked sent" });
      toast("订阅消息任务已标记发送");
    }
    await loadWechatTasks();
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadDeliveryDocuments() {
  if (!can("delivery:manage")) {
    $("#delivery-documents-table").innerHTML = emptyRow(6, "当前角色没有交付文档权限");
    return;
  }
  const data = await apiGet("/api/admin/delivery-documents");
  state.deliveryDocuments = data.items || [];
  $("#delivery-documents-table").innerHTML = state.deliveryDocuments.map(deliveryDocumentRow).join("") || emptyRow(6, "暂无交付文档");
}

async function createDeliveryDocument(event) {
  event.preventDefault();
  if (!can("delivery:manage")) {
    toast("当前角色没有交付文档权限", true);
    return;
  }
  const form = event.currentTarget;
  const data = Object.fromEntries(new FormData(form).entries());
  try {
    await apiPost("/api/admin/delivery-documents", {
      docType: data.docType,
      title: data.title.trim(),
      status: data.status,
      reason: data.reason.trim(),
    });
    toast("交付文档记录已新增");
    form.reset();
    await loadDeliveryDocuments();
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadTestCases() {
  if (!can("testcase:read")) {
    $("#test-cases-table").innerHTML = emptyRow(5, "当前角色没有测试用例权限");
    return;
  }
  const data = await apiGet("/api/admin/test-cases");
  state.testCases = data.items || [];
  $("#test-cases-table").innerHTML = state.testCases.map(testCaseRow).join("") || emptyRow(5, "暂无测试用例");
}

async function createTestCase(event) {
  event.preventDefault();
  if (!can("testcase:manage")) {
    toast("当前角色没有测试用例写入权限", true);
    return;
  }
  const form = event.currentTarget;
  const data = Object.fromEntries(new FormData(form).entries());
  try {
    await apiPost("/api/admin/test-cases", {
      module: data.module.trim(),
      caseName: data.caseName.trim(),
      priority: data.priority,
      expectedResult: data.expectedResult.trim(),
    });
    toast("测试用例已新增");
    form.reset();
    await loadTestCases();
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadTestRuns() {
  if (!can("testcase:read")) {
    $("#test-runs-table").innerHTML = emptyRow(6, "当前角色没有测试运行权限");
    return;
  }
  const data = await apiGet("/api/admin/test-runs");
  state.testRuns = data.items || [];
  $("#test-runs-table").innerHTML = state.testRuns.map(testRunRow).join("") || emptyRow(6, "暂无测试运行");
}

async function createTestRun(event) {
  event.preventDefault();
  if (!can("testcase:manage")) {
    toast("当前角色没有测试运行写入权限", true);
    return;
  }
  const form = event.currentTarget;
  const data = Object.fromEntries(new FormData(form).entries());
  try {
    await apiPost("/api/admin/test-runs", {
      caseId: Number(data.caseId),
      result: data.result,
      actualResult: data.actualResult.trim(),
      requestId: data.requestId.trim(),
      evidenceFileId: numberOrZero(data.evidenceFileId),
    });
    toast("测试运行已记录");
    form.reset();
    await loadTestRuns();
  } catch (error) {
    toast(error.message, true);
  }
}

async function onTestRunTableClick(event) {
  const button = event.target.closest("button[data-action='test-run-file-download']");
  if (!button) return;
  if (!can("testcase:read")) {
    toast("当前角色没有测试运行权限", true);
    return;
  }
  try {
    await downloadAdminFile(button.dataset.fileId);
  } catch (error) {
    toast(error.message, true);
  }
}

async function renderAdmins() {
  $("#admin-create-form").addEventListener("submit", createAdminUser);
  $("#admin-users-table").addEventListener("click", onAdminUserTableClick);
  await Promise.all([loadAdminAccounts(), loadAdminRoles(), loadAdminPermissionCatalog()]);
}

async function createAdminUser(event) {
  event.preventDefault();
  if (!can("admin_user:create")) {
    toast("当前角色没有创建后台账号权限", true);
    return;
  }
  const data = Object.fromEntries(new FormData(event.currentTarget).entries());
  try {
    await apiPost("/api/admin/admin-users", {
      username: data.username.trim(),
      password: data.password,
      roles: data.roles.split(",").map((item) => item.trim()).filter(Boolean),
      status: data.status,
    });
    toast("后台账号已创建");
    event.currentTarget.reset();
    event.currentTarget.status.value = "active";
    await loadAdminAccounts();
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadAdminAccounts() {
  const data = await apiGet("/api/admin/admin-users");
  state.adminUsers = data.items || [];
  $("#admin-users-table").innerHTML = state.adminUsers.map(adminUserRow).join("") || emptyRow(6, "暂无后台账号");
}

async function onAdminUserTableClick(event) {
  const button = event.target.closest("button[data-action]");
  if (!button) return;
  const id = button.dataset.id;
  try {
    if (button.dataset.action === "admin-user-disable") {
      await apiPut(`/api/admin/admin-users/${id}`, { roles: rolesFromRow(button.closest("tr")), status: "disabled" });
      toast("后台账号已禁用");
    }
    if (button.dataset.action === "admin-user-enable") {
      await apiPut(`/api/admin/admin-users/${id}`, { roles: rolesFromRow(button.closest("tr")), status: "active" });
      toast("后台账号已启用");
    }
    if (button.dataset.action === "admin-user-update") {
      const row = button.closest("tr");
      const roles = row.querySelector("input[data-field='roles']").value;
      const status = row.querySelector("select[data-field='status']").value;
      await apiPut(`/api/admin/admin-users/${id}`, {
        roles: roles.split(",").map((item) => item.trim()).filter(Boolean),
        status,
      });
      toast("后台账号已更新");
    }
    await loadAdminAccounts();
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadAdminRoles() {
  const data = await apiGet("/api/admin/admin-roles");
  state.adminRoles = data.items || [];
  $("#admin-role-list").innerHTML = state.adminRoles.map((item) => stackItem({
    title: `${item.name || item.code} / ${item.code}`,
    badge: item.adminCount > 0 ? "active" : "pending",
    meta: [
      `adminCount=${item.adminCount || 0}`,
      `permissionCount=${item.permissionCount || 0}`,
      `description=${item.description || "-"}`,
      `permissions=${compactList(item.permissions || [], 12)}`,
    ],
  })).join("") || emptyBlock("暂无后台角色");
}

async function loadAdminPermissionCatalog() {
  const data = await apiGet("/api/admin/admin-permissions/catalog");
  state.adminPermissionCatalog = data.items || [];
  $("#admin-permissions-table").innerHTML = state.adminPermissionCatalog.map(permissionCatalogRow).join("") || emptyRow(3, "暂无权限目录");
}

async function renderSystem() {
  ensureGuideRulePanel();
  $("#system-readiness-refresh").addEventListener("click", loadSystemReadiness);
  $("#game-category-refresh").addEventListener("click", loadGameCategoryConfig);
  $("#game-category-form").addEventListener("submit", saveGameCategoryConfig);
  $("#home-display-refresh").addEventListener("click", loadHomeDisplayConfig);
  $("#home-display-form").addEventListener("submit", saveHomeDisplayConfig);
  $("#game-application-refresh").addEventListener("click", loadGameApplicationConfig);
  $("#game-application-form").addEventListener("submit", saveGameApplicationConfig);
  $("#game-audit-refresh").addEventListener("click", loadGameAuditConfig);
  $("#game-audit-form").addEventListener("submit", saveGameAuditConfig);
  $("#condition-rule-refresh").addEventListener("click", loadConditionRuleConfig);
  $("#condition-rule-form").addEventListener("submit", saveConditionRuleConfig);
  $("#role-benefit-refresh").addEventListener("click", loadRoleBenefitConfig);
  $("#role-benefit-form").addEventListener("submit", saveRoleBenefitConfig);
  $("#review-complete-refresh").addEventListener("click", loadReviewCompleteConfig);
  $("#review-complete-form").addEventListener("submit", saveReviewCompleteConfig);
  $("#credit-rule-refresh").addEventListener("click", loadCreditDeductionRules);
  $("#credit-rule-form").addEventListener("submit", saveCreditDeductionRules);
  $("#profit-config-refresh").addEventListener("click", loadProfitTemplateConfig);
  $("#profit-config-form").addEventListener("submit", saveProfitTemplateConfig);
  $("#risk-logs-refresh").addEventListener("click", loadRiskLogs);
  $("#guide-rule-form").addEventListener("submit", updateGuideRule);
  $("#guide-qualification-form").addEventListener("submit", updateGuideQualification);
  $("#sensitive-word-form").addEventListener("submit", createSensitiveWord);
  $("#sensitive-import-form").addEventListener("submit", importSensitiveWords);
  $("#sensitive-words-table").addEventListener("click", onSensitiveWordClick);
  $("#view-root").addEventListener("click", onSystemActionClick);
  const tasks = [];
  if (can("system_config:read")) {
    tasks.push(loadSystemReadiness());
    tasks.push(loadGameCategoryConfig());
    tasks.push(loadHomeDisplayConfig());
    tasks.push(loadGameApplicationConfig());
    tasks.push(loadGameAuditConfig());
    tasks.push(loadConditionRuleConfig());
    tasks.push(loadRoleBenefitConfig());
    tasks.push(loadReviewCompleteConfig());
    tasks.push(loadCreditDeductionRules());
    tasks.push(loadProfitTemplateConfig());
  } else {
    renderNoAccess("#system-readiness-summary", "缺少 system_config:read");
    $("#system-readiness-table").innerHTML = emptyRow(5, "无权限查看生产就绪检查");
    renderNoAccess("#game-category-config-panel", "缺少 system_config:read");
    renderNoAccess("#home-display-config-panel", "缺少 system_config:read");
    renderNoAccess("#game-application-config-panel", "缺少 system_config:read");
    renderNoAccess("#game-audit-config-panel", "缺少 system_config:read");
    renderNoAccess("#condition-rule-config-panel", "缺少 system_config:read");
    renderNoAccess("#role-benefit-config-panel", "缺少 system_config:read");
    renderNoAccess("#review-complete-config-panel", "缺少 system_config:read");
    renderNoAccess("#credit-rule-config-panel", "缺少 system_config:read");
  }
  if (can("content:sensitive_word:view")) {
    tasks.push(loadSensitiveWords());
  } else {
    $("#sensitive-words-table").innerHTML = emptyRow(7, "无权限查看敏感词库");
  }
  if (can("content:risk_log:view")) {
    tasks.push(loadRiskLogs());
  } else {
    renderNoAccess("#risk-log-list", "缺少 content:risk_log:view");
  }
  tasks.push(loadPermissionTree());
  if (can("role:view")) {
    tasks.push(loadGuideQualificationRules());
  } else {
    renderNoAccess("#guide-rule-list", "缺少 role:view");
  }
  if (can("ai:data:read")) {
    tasks.push(loadAIExportConfig());
  } else {
    renderNoAccess("#ai-export-config-panel", "缺少 ai:data:read");
  }
  await Promise.all(tasks);
}

function ensureGuideRulePanel() {
  if ($("#guide-rule-panel")) return;
  const anchor = $("#permission-tree-panel")?.closest(".split");
  if (!anchor) return;
  anchor.insertAdjacentHTML("afterend", `
    <section id="guide-rule-panel" class="panel">
      <div class="panel-head">
        <div>
          <h2>领路人资格规则</h2>
          <p>后台控制领路人/行家测试门槛，字段对应 guide_qualification_rules。</p>
        </div>
        <button id="guide-rule-refresh" class="ghost" type="button">刷新</button>
      </div>
      <form id="guide-rule-form" class="form-grid compact-grid">
        <label>ruleId<input name="ruleId" inputmode="numeric" required placeholder="规则 ID" /></label>
        <label>minInviteCount<input name="minInviteCount" inputmode="numeric" placeholder="不填则不变" /></label>
        <label>minCreditScore<input name="minCreditScore" inputmode="numeric" placeholder="0-100，不填则不变" /></label>
        <label>minCompletedGames<input name="minCompletedGames" inputmode="numeric" placeholder="不填则不变" /></label>
        <label>paymentRequired
          <select name="paymentRequired">
            <option value="">不变</option>
            <option value="true">true</option>
            <option value="false">false</option>
          </select>
        </label>
        <label>status
          <select name="status">
            <option value="">不变</option>
            <option value="active">active</option>
            <option value="disabled">disabled</option>
          </select>
        </label>
        <button class="primary" type="submit">保存规则</button>
      </form>
      <div id="guide-rule-list" class="stack-list"></div>
      <form id="guide-qualification-form" class="form-grid compact-grid profile-form-gap">
        <label>userId<input name="userId" inputmode="numeric" required placeholder="领路人/行家 userId" /></label>
        <label>conditionMet
          <select name="conditionMet">
            <option value="">不变</option>
            <option value="true">true</option>
            <option value="false">false</option>
          </select>
        </label>
        <label>paymentMet
          <select name="paymentMet">
            <option value="">不变</option>
            <option value="true">true</option>
            <option value="false">false</option>
          </select>
        </label>
        <button class="ghost" data-action="guide-qualification-load" type="button">查询资格</button>
        <button class="primary" type="submit">保存用户资格</button>
      </form>
      <div id="guide-qualification-panel" class="detail-grid"></div>
    </section>
  `);
  $("#guide-rule-refresh").addEventListener("click", loadGuideQualificationRules);
}

async function loadSystemReadiness() {
  if (!can("system_config:read")) {
    renderNoAccess("#system-readiness-summary", "缺少 system_config:read");
    $("#system-readiness-table").innerHTML = emptyRow(5, "无权限查看生产就绪检查");
    return;
  }
  const data = await apiGet("/api/admin/system/readiness");
  state.readiness = data;
  $("#system-readiness-summary").innerHTML = [
    detailCell("ready", data.ready),
    detailCell("production", data.production),
    detailCell("environment", data.environment || data.appEnv || "-"),
    detailCell("items", (data.items || []).length),
  ].join("");
  $("#system-readiness-table").innerHTML = (data.items || []).map((item) => `
    <tr>
      <td>${escapeHTML(item.key || item.code || "-")}</td>
      <td>${escapeHTML(item.name || "-")}</td>
      <td><span class="${badgeClass(readinessStatus(item) === "ok" ? "active" : "failed")}">${escapeHTML(readinessStatus(item))}</span></td>
      <td>${escapeHTML(String(Boolean(item.critical || item.required)))}</td>
      <td>${escapeHTML(readinessMessage(item))}</td>
    </tr>
  `).join("") || emptyRow(5, "暂无就绪检查项");
}

function readinessStatus(item) {
  if (item.status) return item.status;
  return item.ready ? "ok" : "missing";
}

function readinessMessage(item) {
  const detailText = item.details ? Object.entries(item.details).map(([key, value]) => `${key}: ${value}`).join("；") : "";
  return [item.message || item.reason || "-", detailText].filter(Boolean).join("；");
}

async function loadGameCategoryConfig() {
  if (!can("system_config:read")) {
    renderNoAccess("#game-category-config-panel", "缺少 system_config:read");
    return;
  }
  const data = await apiGet("/api/admin/games/category-config");
  state.gameCategoryConfig = data.config || data;
  const textarea = $("#game-category-config-json");
  if (textarea) textarea.value = JSON.stringify(state.gameCategoryConfig, null, 2);
  renderGameCategoryConfig();
}

function renderGameCategoryConfig() {
  const config = state.gameCategoryConfig || {};
  $("#game-category-config-panel").innerHTML = [
    detailCell("primaryCategories", (config.primaryCategories || []).length),
    detailCell("typeFilters", (config.typeFilters || []).length),
    detailCell("locationFilters", (config.locationFilters || []).length),
    detailCell("defaultPrimaryCategory", config.defaultPrimaryCategory || "-"),
    detailCell("defaultSecondaryCategory", config.defaultSecondaryCategory || "-"),
    detailCell("defaultType", config.defaultType || "-"),
    detailCell("version", config.version || "-"),
  ].join("");
}

async function saveGameCategoryConfig(event) {
  event.preventDefault();
  if (!can("system_config:update")) {
    toast("缺少 system_config:update", true);
    return;
  }
  const raw = String(new FormData(event.currentTarget).get("configJson") || "").trim();
  if (!raw) {
    toast("配置 JSON 不能为空", true);
    return;
  }
  let payload;
  try {
    payload = JSON.parse(raw);
  } catch (error) {
    toast("配置 JSON 格式不正确", true);
    return;
  }
  try {
    const data = await apiPut("/api/admin/games/category-config", payload);
    state.gameCategoryConfig = data.config || data;
    const textarea = $("#game-category-config-json");
    if (textarea) textarea.value = JSON.stringify(state.gameCategoryConfig, null, 2);
    renderGameCategoryConfig();
    toast("组局分类配置已保存");
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadHomeDisplayConfig() {
  if (!can("system_config:read")) {
    renderNoAccess("#home-display-config-panel", "缺少 system_config:read");
    return;
  }
  const data = await apiGet("/api/admin/home/display-config");
  state.homeDisplayConfig = data.config || data;
  const textarea = $("#home-display-config-json");
  if (textarea) textarea.value = JSON.stringify(state.homeDisplayConfig, null, 2);
  renderHomeDisplayConfig();
}

function renderHomeDisplayConfig() {
  const config = state.homeDisplayConfig || {};
  $("#home-display-config-panel").innerHTML = [
    detailCell("onlineBaseCount", config.onlineBaseCount ?? 0),
    detailCell("onlineSuffix", config.onlineSuffix || "-"),
  ].join("");
}

async function saveHomeDisplayConfig(event) {
  event.preventDefault();
  if (!can("system_config:update")) {
    toast("缺少 system_config:update", true);
    return;
  }
  const raw = String(new FormData(event.currentTarget).get("configJson") || "").trim();
  if (!raw) {
    toast("配置 JSON 不能为空", true);
    return;
  }
  let payload;
  try {
    payload = JSON.parse(raw);
  } catch (error) {
    toast("配置 JSON 格式不正确", true);
    return;
  }
  try {
    const data = await apiPut("/api/admin/home/display-config", payload);
    state.homeDisplayConfig = data.config || data;
    const textarea = $("#home-display-config-json");
    if (textarea) textarea.value = JSON.stringify(state.homeDisplayConfig, null, 2);
    renderHomeDisplayConfig();
    toast("首页展示配置已保存");
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadGameApplicationConfig() {
  if (!can("system_config:read")) {
    renderNoAccess("#game-application-config-panel", "缺少 system_config:read");
    return;
  }
  const data = await apiGet("/api/admin/games/application-config");
  state.gameApplicationConfig = data.config || data;
  const textarea = $("#game-application-config-json");
  if (textarea) textarea.value = JSON.stringify(state.gameApplicationConfig, null, 2);
  renderGameApplicationConfig();
}

function renderGameApplicationConfig() {
  const config = state.gameApplicationConfig || {};
  $("#game-application-config-panel").innerHTML = [
    detailCell("agreementTitle", config.agreementTitle || "-"),
    detailCell("requireRealname", Boolean(config.requireRealname)),
    detailCell("requireAgreement", Boolean(config.requireAgreement)),
    detailCell("allowDuplicateApply", Boolean(config.allowDuplicateApply)),
    detailCell("maxUploadCount", config.maxUploadCount ?? 0),
    detailCell("searchEnabled", Boolean(config.searchEnabled)),
    detailCell("version", config.version || "-"),
  ].join("");
}

async function saveGameApplicationConfig(event) {
  event.preventDefault();
  await saveJSONSystemConfig({
    form: event.currentTarget,
    endpoint: "/api/admin/games/application-config",
    stateKey: "gameApplicationConfig",
    textareaSelector: "#game-application-config-json",
    render: renderGameApplicationConfig,
    successMessage: "入局申请配置已保存",
  });
}

async function loadGameAuditConfig() {
  if (!can("system_config:read")) {
    renderNoAccess("#game-audit-config-panel", "缺少 system_config:read");
    return;
  }
  const data = await apiGet("/api/admin/games/audit-config");
  state.gameAuditConfig = data.config || data;
  const textarea = $("#game-audit-config-json");
  if (textarea) textarea.value = JSON.stringify(state.gameAuditConfig, null, 2);
  renderGameAuditConfig();
}

function renderGameAuditConfig() {
  const config = state.gameAuditConfig || {};
  $("#game-audit-config-panel").innerHTML = [
    detailCell("autoApproveFreeGames", Boolean(config.autoApproveFreeGames)),
    detailCell("requireManualAuditTypes", compactList(config.requireManualAuditTypes || [])),
    detailCell("requiredRejectReason", Boolean(config.requiredRejectReason)),
    detailCell("applicationAuditMode", config.applicationAuditMode || "-"),
    detailCell("batchAuditMaxCount", config.batchAuditMaxCount ?? 0),
    detailCell("version", config.version || "-"),
  ].join("");
}

async function saveGameAuditConfig(event) {
  event.preventDefault();
  await saveJSONSystemConfig({
    form: event.currentTarget,
    endpoint: "/api/admin/games/audit-config",
    stateKey: "gameAuditConfig",
    textareaSelector: "#game-audit-config-json",
    render: renderGameAuditConfig,
    successMessage: "组局审核配置已保存",
  });
}

async function loadConditionRuleConfig() {
  if (!can("system_config:read")) {
    renderNoAccess("#condition-rule-config-panel", "缺少 system_config:read");
    return;
  }
  const data = await apiGet("/api/admin/games/condition-rule-config");
  state.conditionRuleConfig = data.config || data;
  const textarea = $("#condition-rule-config-json");
  if (textarea) textarea.value = JSON.stringify(state.conditionRuleConfig, null, 2);
  renderConditionRuleConfig();
}

function renderConditionRuleConfig() {
  const config = state.conditionRuleConfig || {};
  $("#condition-rule-config-panel").innerHTML = [
    detailCell("enabled", Boolean(config.enabled)),
    detailCell("visibleInMiniProgram", Boolean(config.visibleInMiniProgram)),
    detailCell("adminOnlyCreate", Boolean(config.adminOnlyCreate)),
    detailCell("ruleItems", (config.ruleItems || []).length),
    detailCell("defaultVisibility", config.defaultVisibility || "-"),
    detailCell("paymentRequired", Boolean(config.paymentRequired)),
    detailCell("version", config.version || "-"),
  ].join("");
}

async function saveConditionRuleConfig(event) {
  event.preventDefault();
  await saveJSONSystemConfig({
    form: event.currentTarget,
    endpoint: "/api/admin/games/condition-rule-config",
    stateKey: "conditionRuleConfig",
    textareaSelector: "#condition-rule-config-json",
    render: renderConditionRuleConfig,
    successMessage: "条件局规则配置已保存",
  });
}

async function loadRoleBenefitConfig() {
  if (!can("system_config:read")) {
    renderNoAccess("#role-benefit-config-panel", "缺少 system_config:read");
    return;
  }
  const data = await apiGet("/api/admin/roles/benefit-config");
  state.roleBenefitConfig = data.config || data;
  const textarea = $("#role-benefit-config-json");
  if (textarea) textarea.value = JSON.stringify(state.roleBenefitConfig, null, 2);
  renderRoleBenefitConfig();
}

function renderRoleBenefitConfig() {
  const config = state.roleBenefitConfig || {};
  const comparison = config.roleComparison || {};
  const prompts = config.permissionPrompts || {};
  $("#role-benefit-config-panel").innerHTML = [
    detailCell("permissionPrompts", Object.keys(prompts).length),
    detailCell("roles", (comparison.roles || []).length),
    detailCell("benefits", (comparison.benefits || []).length),
    detailCell("title", comparison.title || "-"),
    detailCell("version", config.version || "-"),
  ].join("");
}

async function saveRoleBenefitConfig(event) {
  event.preventDefault();
  await saveJSONSystemConfig({
    form: event.currentTarget,
    endpoint: "/api/admin/roles/benefit-config",
    stateKey: "roleBenefitConfig",
    textareaSelector: "#role-benefit-config-json",
    render: renderRoleBenefitConfig,
    successMessage: "角色权益配置已保存",
  });
}

async function loadReviewCompleteConfig() {
  if (!can("system_config:read")) {
    renderNoAccess("#review-complete-config-panel", "缺少 system_config:read");
    return;
  }
  const data = await apiGet("/api/admin/reviews/complete-config");
  state.reviewCompleteConfig = data.config || data;
  const textarea = $("#review-complete-config-json");
  if (textarea) textarea.value = JSON.stringify(state.reviewCompleteConfig, null, 2);
  renderReviewCompleteConfig();
}

function renderReviewCompleteConfig() {
  const config = state.reviewCompleteConfig || {};
  $("#review-complete-config-panel").innerHTML = [
    detailCell("benefits", (config.benefits || []).length),
    detailCell("playOptions", (config.playOptions || []).length),
    detailCell("defaultIntent", (config.playOptions || [])[0]?.intent || "-"),
    detailCell("version", config.version || "-"),
  ].join("");
}

async function saveReviewCompleteConfig(event) {
  event.preventDefault();
  await saveJSONSystemConfig({
    form: event.currentTarget,
    endpoint: "/api/admin/reviews/complete-config",
    stateKey: "reviewCompleteConfig",
    textareaSelector: "#review-complete-config-json",
    render: renderReviewCompleteConfig,
    successMessage: "评价完成配置已保存",
  });
}

async function loadCreditDeductionRules() {
  if (!can("system_config:read")) {
    renderNoAccess("#credit-rule-config-panel", "缺少 system_config:read");
    return;
  }
  const data = await apiGet("/api/admin/credit-deduction-rules");
  state.creditDeductionRules = data.items || [];
  const textarea = $("#credit-rule-config-json");
  if (textarea) textarea.value = JSON.stringify({ items: state.creditDeductionRules }, null, 2);
  renderCreditDeductionRules();
}

function renderCreditDeductionRules() {
  const items = state.creditDeductionRules || [];
  const enabledCount = items.filter((item) => item.enabled).length;
  $("#credit-rule-config-panel").innerHTML = [
    detailCell("rules", items.length),
    detailCell("enabled", enabledCount),
    detailCell("disabled", items.length - enabledCount),
    detailCell("low_review", items.find((item) => item.ruleCode === "low_review")?.changeValue ?? "-"),
  ].join("");
}

async function saveCreditDeductionRules(event) {
  event.preventDefault();
  await saveJSONSystemConfig({
    form: event.currentTarget,
    endpoint: "/api/admin/credit-deduction-rules",
    stateKey: "creditDeductionRules",
    textareaSelector: "#credit-rule-config-json",
    render: renderCreditDeductionRules,
    successMessage: "信用扣分规则已保存",
    pickState: (data) => data.items || [],
    normalizePayload: (payload) => Array.isArray(payload) ? { items: payload } : payload,
  });
}

async function saveJSONSystemConfig({ form, endpoint, stateKey, textareaSelector, render, successMessage, pickState, normalizePayload }) {
  if (!can("system_config:update")) {
    toast("缺少 system_config:update", true);
    return;
  }
  const raw = String(new FormData(form).get("configJson") || "").trim();
  if (!raw) {
    toast("配置 JSON 不能为空", true);
    return;
  }
  let payload;
  try {
    payload = JSON.parse(raw);
  } catch (error) {
    toast("配置 JSON 格式不正确", true);
    return;
  }
  if (typeof normalizePayload === "function") {
    payload = normalizePayload(payload);
  }
  try {
    const data = await apiPut(endpoint, payload);
    state[stateKey] = typeof pickState === "function" ? pickState(data) : (data.config || data);
    const textarea = $(textareaSelector);
    if (textarea) textarea.value = JSON.stringify(typeof pickState === "function" ? { items: state[stateKey] } : state[stateKey], null, 2);
    render();
    toast(successMessage);
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadProfitTemplateConfig() {
  if (!can("system_config:read")) {
    renderNoAccess("#profit-config-panel", "缺少 system_config:read");
    return;
  }
  const data = await apiGet("/api/admin/revenue/profit-template-config");
  state.profitTemplateConfig = data.config || data;
  const form = $("#profit-config-form");
  if (form) {
    form.depositRuleText.value = state.profitTemplateConfig.depositRuleText || "";
    form.depositNoticeText.value = state.profitTemplateConfig.depositNoticeText || "";
    form.version.value = state.profitTemplateConfig.version || "";
  }
  renderProfitTemplateConfig();
}

function renderProfitTemplateConfig() {
  const config = state.profitTemplateConfig || {};
  $("#profit-config-panel").innerHTML = [
    detailCell("depositRuleText", config.depositRuleText || "-"),
    detailCell("depositNoticeText", config.depositNoticeText || "-"),
    detailCell("version", config.version || "-"),
  ].join("");
}

async function saveProfitTemplateConfig(event) {
  event.preventDefault();
  if (!can("system_config:update")) {
    toast("缺少 system_config:update", true);
    return;
  }
  const data = Object.fromEntries(new FormData(event.currentTarget).entries());
  const payload = {
    depositRuleText: String(data.depositRuleText || "").trim(),
    depositNoticeText: String(data.depositNoticeText || "").trim(),
    version: String(data.version || "").trim(),
  };
  try {
    const result = await apiPut("/api/admin/revenue/profit-template-config", payload);
    state.profitTemplateConfig = result.config || result;
    renderProfitTemplateConfig();
    toast("押金局规则配置已保存");
  } catch (error) {
    toast(error.message, true);
  }
}

async function createSensitiveWord(event) {
  event.preventDefault();
  if (!can("content:sensitive_word:create")) {
    toast("缺少 content:sensitive_word:create", true);
    return;
  }
  const data = Object.fromEntries(new FormData(event.currentTarget).entries());
  try {
    await apiPost("/api/admin/sensitive-words", {
      word: data.word.trim(),
      level: data.level,
      action: data.action,
      status: "active",
    });
    toast("敏感词已新增");
    event.currentTarget.reset();
    await loadSensitiveWords();
  } catch (error) {
    toast(error.message, true);
  }
}

async function importSensitiveWords(event) {
  event.preventDefault();
  if (!can("content:sensitive_word:import")) {
    toast("缺少 content:sensitive_word:import", true);
    return;
  }
  const words = String(new FormData(event.currentTarget).get("words") || "")
    .split(/\r?\n/)
    .map((item) => item.trim())
    .filter(Boolean);
  try {
    await apiPost("/api/admin/sensitive-words/import", { words });
    toast(`已导入 ${words.length} 个敏感词`);
    event.currentTarget.reset();
    await loadSensitiveWords();
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadSensitiveWords() {
  if (!can("content:sensitive_word:view")) {
    $("#sensitive-words-table").innerHTML = emptyRow(7, "无权限查看敏感词库");
    return;
  }
  const data = await apiGet("/api/admin/sensitive-words");
  state.sensitiveWords = data.items || [];
  $("#sensitive-words-table").innerHTML = state.sensitiveWords.map(sensitiveWordRow).join("") || emptyRow(7, "暂无敏感词");
}

async function onSensitiveWordClick(event) {
  const button = event.target.closest("button[data-action='sensitive-status']");
  if (!button) return;
  if (!can("content:sensitive_word:update")) {
    toast("缺少 content:sensitive_word:update", true);
    return;
  }
  try {
    await apiPut(`/api/admin/sensitive-words/${button.dataset.id}`, { status: button.dataset.status });
    toast("敏感词状态已更新");
    await loadSensitiveWords();
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadRiskLogs() {
  if (!can("content:risk_log:view")) {
    renderNoAccess("#risk-log-list", "缺少 content:risk_log:view");
    return;
  }
  const data = await apiGet("/api/admin/content-risk/logs");
  state.riskLogs = data.items || [];
  $("#risk-log-list").innerHTML = state.riskLogs.map((item) => stackItem({
    title: `${item.action} / ${item.word || "content"} #${item.id}`,
    badge: item.status,
    meta: [`gameId=${item.gameId || "-"}`, `roomId=${item.roomId || "-"}`, `messageId=${item.messageId || "-"}`, `source=${item.source || "-"}`, `created=${formatTime(item.createdAt)}`],
  })).join("") || emptyBlock("暂无内容风险日志");
}

async function loadPermissionTree() {
  const data = await apiGet("/api/admin/permissions/tree");
  state.permissionTree = data;
  $("#permission-tree-panel").innerHTML = [
    stackItem({
      title: `${data.adminUser?.username || "admin"} / ${(data.roles || []).join(",") || "-"}`,
      badge: data.adminUser?.status || "active",
      meta: [`permissions=${(data.permissions || []).length}`, `menus=${(data.menus || []).length}`, `apis=${(data.apis || []).length}`],
    }),
    stackItem({
      title: "权限 codes",
      badge: "active",
      meta: [(data.permissions || []).join(", ") || "-"],
    }),
  ].join("");
}

async function loadGuideQualificationRules() {
  if (!can("role:view")) {
    renderNoAccess("#guide-rule-list", "缺少 role:view");
    return;
  }
  const data = await apiGet("/api/admin/guides/qualification-rules");
  state.guideQualificationRules = data.items || [];
  $("#guide-rule-list").innerHTML = state.guideQualificationRules.map((item) => stackItem({
    title: `${item.ruleCode || "default"} #${item.id}`,
    badge: item.status,
    meta: [
      `minInviteCount=${item.minInviteCount || 0}`,
      `minCreditScore=${item.minCreditScore || 0}`,
      `minCompletedGames=${item.minCompletedGames || 0}`,
      `paymentRequired=${Boolean(item.paymentRequired)}`,
      `updated=${formatTime(item.updatedAt)}`,
    ],
    action: `<button class="ghost" data-action="guide-rule-load" data-id="${item.id}" type="button">载入</button>`,
  })).join("") || emptyBlock("暂无领路人资格规则");
}

async function loadAIExportConfig() {
  if (!can("ai:data:read")) {
    renderNoAccess("#ai-export-config-panel", "缺少 ai:data:read");
    return;
  }
  const data = await apiGet("/api/admin/ai-data/im-export-config");
  state.aiExportConfig = data.config || data;
  renderAIExportConfig();
}

async function onSystemActionClick(event) {
  const button = event.target.closest("button[data-action]");
  if (!button) return;
  if (button.dataset.action === "guide-rule-load") {
    fillGuideRuleForm(button.dataset.id);
    return;
  }
  if (button.dataset.action === "guide-qualification-load") {
    await loadGuideQualification();
    return;
  }
  if (!["ai-export-enable", "ai-export-disable"].includes(button.dataset.action)) return;
  if (!can("system_config:update")) {
    toast("缺少 system_config:update", true);
    return;
  }
  try {
    await apiPut("/api/admin/ai-data/im-export-config", { enabled: button.dataset.action === "ai-export-enable" });
    toast("AI IM 导出开关已更新");
    await loadAIExportConfig();
  } catch (error) {
    toast(error.message, true);
  }
}

function fillGuideRuleForm(ruleID) {
  const rule = state.guideQualificationRules.find((item) => String(item.id) === String(ruleID));
  const form = $("#guide-rule-form");
  if (!rule || !form) return;
  form.ruleId.value = rule.id || "";
  form.minInviteCount.value = rule.minInviteCount ?? "";
  form.minCreditScore.value = rule.minCreditScore ?? "";
  form.minCompletedGames.value = rule.minCompletedGames ?? "";
  form.paymentRequired.value = String(Boolean(rule.paymentRequired));
  form.status.value = rule.status || "";
}

async function updateGuideRule(event) {
  event.preventDefault();
  if (!can("role:update")) {
    toast("缺少 role:update", true);
    return;
  }
  const form = event.currentTarget;
  const data = Object.fromEntries(new FormData(form).entries());
  const ruleID = Number(data.ruleId);
  const payload = {};
  addOptionalNumber(payload, "minInviteCount", data.minInviteCount);
  addOptionalNumber(payload, "minCreditScore", data.minCreditScore);
  addOptionalNumber(payload, "minCompletedGames", data.minCompletedGames);
  if (data.paymentRequired) payload.paymentRequired = data.paymentRequired === "true";
  if (data.status) payload.status = data.status;
  if (!Number.isFinite(ruleID) || ruleID <= 0) {
    toast("请输入有效的规则 ID", true);
    return;
  }
  try {
    await apiPut(`/api/admin/guides/qualification-rules/${ruleID}`, payload);
    toast("领路人资格规则已保存");
    await loadGuideQualificationRules();
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadGuideQualification() {
  if (!can("role:view")) {
    toast("缺少 role:view", true);
    return;
  }
  const form = $("#guide-qualification-form");
  const userID = Number(new FormData(form).get("userId"));
  if (!Number.isFinite(userID) || userID <= 0) {
    toast("请输入有效的用户 ID", true);
    return;
  }
  try {
    const qualification = await apiGet(`/api/admin/guide-qualification-rules?userId=${encodeURIComponent(userID)}`);
    renderGuideQualification(qualification);
  } catch (error) {
    toast(error.message, true);
  }
}

async function updateGuideQualification(event) {
  event.preventDefault();
  if (!can("role:update")) {
    toast("缺少 role:update", true);
    return;
  }
  const form = event.currentTarget;
  const data = Object.fromEntries(new FormData(form).entries());
  const userID = Number(data.userId);
  const payload = { userId: userID };
  if (data.conditionMet) payload.conditionMet = data.conditionMet === "true";
  if (data.paymentMet) payload.paymentMet = data.paymentMet === "true";
  if (!Number.isFinite(userID) || userID <= 0) {
    toast("请输入有效的用户 ID", true);
    return;
  }
  if (!("conditionMet" in payload) && !("paymentMet" in payload)) {
    toast("请选择要更新的资格字段", true);
    return;
  }
  try {
    const qualification = await apiPost("/api/admin/guide-qualification-rules", payload);
    renderGuideQualification(qualification);
    toast("用户资格已保存");
  } catch (error) {
    toast(error.message, true);
  }
}

function renderGuideQualification(item) {
  $("#guide-qualification-panel").innerHTML = [
    detailCell("userId", item.userId || "-"),
    detailCell("conditionMet", Boolean(item.conditionMet)),
    detailCell("paymentMet", Boolean(item.paymentMet)),
    detailCell("guideOpenStatus", item.guideOpenStatus || "-"),
    detailCell("updatedAt", item.updatedAt || "-"),
  ].join("");
}

function addOptionalNumber(payload, key, value) {
  const text = String(value || "").trim();
  if (!text) return;
  payload[key] = Number(text);
}

function renderAIExportConfig() {
  const config = state.aiExportConfig || {};
  $("#ai-export-config-panel").innerHTML = [
    detailCell("enabled", Boolean(config.enabled)),
    detailCell("updatedAt", config.updatedAt || "-"),
    detailCell("updatedBy", config.updatedBy || "-"),
    detailCell("risk", config.enabled ? "IM export allowed" : "IM export blocked"),
  ].join("");
}

async function renderIM() {
  $("#im-filter-form").addEventListener("submit", (event) => {
    event.preventDefault();
    loadIMRooms(new FormData(event.currentTarget));
  });
  $("#im-room-list").addEventListener("click", onIMRoomListClick);
  ensureIMDetailPanel();
  await loadIMRooms();
}

async function loadIMRooms(formData) {
  const data = await apiGet(`/api/admin/im/rooms${querySuffix(formData)}`);
  const list = $("#im-room-list");
  state.imRooms = data.items || [];
  list.innerHTML = (data.items || []).map((item) => stackItem({
    title: `房间 ${item.id || item.roomId} / Game ${item.gameId}`,
    badge: item.status,
    meta: [`engine=${item.engine || "-"}`, `messages=${item.messageCount || 0}`, `files=${item.fileMessageCount || 0}`, `openIM=${item.openIMGroupId || "-"}`],
    action: imRoomActions(item),
  })).join("") || emptyBlock("暂无 IM 房间");
}

function ensureIMDetailPanel() {
  if ($("#im-room-detail-panel")) return;
  $("#view-root").insertAdjacentHTML("beforeend", `<section id="im-room-detail-panel" class="panel hidden"></section>`);
}

function imRoomActions(item) {
  const roomID = item.id || item.roomId;
  const actions = [`<button class="ghost" data-action="im-room-detail" data-id="${escapeHTML(roomID)}" type="button">详情</button>`];
  if (can("im:message:view_dispute")) actions.push(`<button class="ghost" data-action="im-dispute-messages" data-id="${escapeHTML(roomID)}" type="button">争议消息</button>`);
  if (can("im:room:retry_create")) actions.push(`<button class="ghost" data-action="im-room-retry" data-id="${escapeHTML(roomID)}" type="button">重试同步</button>`);
  if (item.status !== "archived" && can("im:room:archive")) actions.push(`<button class="ghost" data-action="im-room-archive" data-id="${escapeHTML(roomID)}" type="button">归档</button>`);
  return `<div class="row-actions">${actions.join("")}</div>`;
}

async function onIMRoomListClick(event) {
  const button = event.target.closest("button[data-action]");
  if (!button) return;
  const roomID = button.dataset.id;
  try {
    if (button.dataset.action === "im-room-detail") {
      await showIMRoomDetail(roomID);
    }
    if (button.dataset.action === "im-dispute-messages") {
      await showIMDisputeMessages(roomID);
    }
    if (button.dataset.action === "im-room-retry") {
      await apiPost(`/api/admin/im/rooms/${roomID}/retry-create`, {});
      toast("IM 房间已重试同步");
      await loadIMRooms(new FormData($("#im-filter-form")));
    }
    if (button.dataset.action === "im-room-archive") {
      await apiPost(`/api/admin/im/rooms/${roomID}/archive`, { reason: "admin archived from web" });
      toast("IM 房间已归档");
      await loadIMRooms(new FormData($("#im-filter-form")));
    }
  } catch (error) {
    toast(error.message, true);
  }
}

async function showIMRoomDetail(roomID) {
  const data = await apiGet(`/api/admin/im/rooms/${roomID}`);
  renderIMRoomDetail(data, false);
}

async function showIMDisputeMessages(roomID) {
  const data = await apiGet(`/api/admin/im/rooms/${roomID}/dispute-messages`);
  renderIMRoomDetail({ room: data.room, messages: data.items || [], fileMessages: [] }, true);
}

function renderIMRoomDetail(data, disputeOnly) {
  const room = data.room || {};
  const messages = data.messages || [];
  const fileMessages = data.fileMessages || [];
  const panel = $("#im-room-detail-panel");
  panel.classList.remove("hidden");
  panel.innerHTML = `
    <div class="panel-head">
      <div>
        <h2>IM 房间详情 #${escapeHTML(room.id || "-")}</h2>
        <p>${disputeOnly ? "争议证据消息" : "房间消息、文件消息和 OpenIM 同步状态"}</p>
      </div>
      <span class="${badgeClass(room.status)}">${statusLabel(room.status)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("gameId", room.gameId || "-")}
      ${detailCell("engine", room.engine || "-")}
      ${detailCell("openIMGroupId", room.openIMGroupId || "-")}
      ${detailCell("memberIds", compactList(room.memberIds || []))}
      ${detailCell("messages", messages.length)}
      ${detailCell("fileMessages", fileMessages.length)}
      ${detailCell("archiveReason", room.archiveReason || "-")}
    </div>
    ${gameOpsTable(disputeOnly ? "争议消息" : "全部消息", ["ID", "sender", "type", "status", "fileId", "acked", "read", "createdAt", "操作"], messages.map(imMessageRow).join("") || emptyRow(9, "暂无消息"))}
    ${disputeOnly ? "" : gameOpsTable("文件消息", ["ID", "sender", "type", "status", "fileId", "acked", "read", "createdAt", "操作"], fileMessages.map(imMessageRow).join("") || emptyRow(9, "暂无文件消息"))}
  `;
  panel.addEventListener("click", onIMDetailClick);
  panel.scrollIntoView({ behavior: "smooth", block: "start" });
}

async function onIMDetailClick(event) {
  const fileButton = event.target.closest("button[data-action='admin-file-download']");
  if (fileButton) {
    try {
      await downloadAdminFile(fileButton.dataset.fileId);
    } catch (error) {
      toast(error.message, true);
    }
    return;
  }
  const button = event.target.closest("button[data-action='im-message-hide']");
  if (!button) return;
  try {
    await apiPost(`/api/admin/im/messages/${button.dataset.id}/hide`, { reason: "admin hidden from web" });
    toast("IM 消息已隐藏");
    await showIMRoomDetail(button.dataset.roomId);
  } catch (error) {
    toast(error.message, true);
  }
}

async function renderLogs() {
  ensureLogTools();
  $("#logs-table").addEventListener("click", onLogsTableClick);
  $("#log-filter-form").addEventListener("submit", (event) => {
    event.preventDefault();
    loadOperationLogs(new FormData(event.currentTarget));
  });
  if (!can("operation_log:view_full")) {
    const input = $("#log-filter-form")?.querySelector("input[name='adminUserId']");
    if (input) {
      input.disabled = true;
      input.placeholder = "仅可查看本人";
      input.value = adminID() || "";
    }
  }
  await loadOperationLogs();
}

function ensureLogTools() {
  if (!$("#log-filter-form")) {
    $("#logs-table").closest(".panel").querySelector(".panel-head").insertAdjacentHTML("afterend", `
      <form id="log-filter-form" class="filters">
        <input name="action" placeholder="action" />
        <input name="targetType" placeholder="targetType" />
        <input name="adminUserId" inputmode="numeric" placeholder="adminUserId" />
        <button class="ghost" type="submit">筛选</button>
      </form>
    `);
  }
  if (!$("#log-detail-panel")) {
    $("#view-root").insertAdjacentHTML("beforeend", `<section id="log-detail-panel" class="panel hidden"></section>`);
  }
}

async function loadOperationLogs(formData) {
  const data = await apiGet(`/api/admin/operation-logs${querySuffix(formData)}`);
  state.operationLogs = data.items || [];
  $("#logs-table").innerHTML = state.operationLogs.map(operationLogRow).join("") || emptyRow(5, "暂无操作日志");
}

function onLogsTableClick(event) {
  const button = event.target.closest("button[data-action='log-detail']");
  if (!button) return;
  showOperationLogDetail(button.dataset.index);
}

function showOperationLogDetail(index) {
  const item = state.operationLogs[Number(index)];
  if (!item) return;
  const panel = $("#log-detail-panel");
  panel.classList.remove("hidden");
  panel.innerHTML = `
    <div class="panel-head">
      <div>
        <h2>操作日志详情</h2>
        <p>字段对齐 operation_logs，用于追踪后台账号、动作对象和请求来源。</p>
      </div>
      <span class="badge neutral">${escapeHTML(item.action || "-")}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("adminUserId", item.adminUserId || "-")}
      ${detailCell("action", item.action || "-")}
      ${detailCell("targetType", item.targetType || "-")}
      ${detailCell("targetId", item.targetId || "-")}
      ${detailCell("requestId", item.requestId || "-")}
      ${detailCell("ip", item.ip || "-")}
      ${detailCell("createdAt", formatTime(item.createdAt))}
      ${detailCell("detail", compactJSON(item.detail || {}))}
    </div>
  `;
  panel.scrollIntoView({ behavior: "smooth", block: "start" });
}

async function apiGet(path) {
  return api(path, { method: "GET" });
}

async function apiPost(path, body, requireAuth = true) {
  return api(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  }, requireAuth);
}

async function apiPut(path, body, requireAuth = true) {
  return api(path, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  }, requireAuth);
}

async function downloadAdminFile(fileID) {
  if (!fileID || Number(fileID) <= 0) {
    toast("无有效文件 ID", true);
    return;
  }
  const result = await apiGet(`/api/admin/files/${encodeURIComponent(fileID)}/download-url`);
  const url = typeof result === "string" ? result : result.downloadUrl;
  if (!url) {
    toast("文件下载地址为空", true);
    return;
  }
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(url).catch(() => {});
  }
  window.open(url, "_blank", "noopener");
  toast("下载地址已生成并复制");
}

async function api(path, options = {}, requireAuth = true) {
  const headers = { ...(options.headers || {}) };
  if (requireAuth && state.token) {
    headers.Authorization = `Bearer ${state.token}`;
    const id = adminID();
    if (id > 0) headers["X-Admin-ID"] = String(id);
  }
  const response = await fetch(path, { ...options, headers });
  let payload = {};
  try {
    payload = await response.json();
  } catch (error) {
    payload = { message: response.statusText };
  }
  if (!response.ok || payload.code !== 0) {
    if (response.status === 401) logout();
    throw new Error(payload.message || `HTTP ${response.status}`);
  }
  return payload.data;
}

function userRow(user) {
  return `
    <tr>
      <td>${escapeHTML(user.id)}</td>
      <td><code>${escapeHTML(user.openId || "-")}</code></td>
      <td>${escapeHTML(user.nickname || "-")}</td>
      <td><span class="${badgeClass(user.realnameStatus)}">${statusLabel(user.realnameStatus)}</span></td>
      <td><span class="${badgeClass(user.status)}">${statusLabel(user.status)}</span></td>
      <td>${formatTime(user.createdAt)}</td>
      <td><button class="ghost" data-action="user-detail" data-id="${user.id}" type="button">详情</button></td>
    </tr>
  `;
}

function gameOpsBlock(gameID, data) {
  const milestones = data.milestones || [];
  const checkins = data.checkins || [];
  const retrospectives = data.retrospectives || [];
  const continueDrafts = data.continueDrafts || [];
  return `
    <div class="sub-panel">
      <div class="panel-head">
        <div>
          <h2>运营明细 #${escapeHTML(gameID)}</h2>
          <p>字段对齐 game_milestones、game_checkins、game_retrospectives、game_continue_drafts。</p>
        </div>
      </div>
      <form data-role="game-milestone-form" class="form-grid compact-grid">
        <label>里程碑 title<input name="title" required maxlength="120" placeholder="确认场地 / 完成服务" /></label>
        <label>status
          <select name="status">
            <option value="pending">pending</option>
            <option value="done">done</option>
          </select>
        </label>
        <button class="ghost" type="submit">新增里程碑</button>
      </form>
      <div class="split profile-form-gap">
        ${gameOpsTable("里程碑", ["ID", "title", "status", "createdAt"], milestones.map(gameMilestoneRow).join("") || emptyRow(4, "暂无里程碑"))}
        ${gameOpsTable("打卡", ["ID", "userId", "milestoneId", "type", "status", "createdAt", "操作"], checkins.map(gameCheckinRow).join("") || emptyRow(7, "暂无打卡"))}
      </div>
      <div class="split profile-form-gap">
        ${gameOpsTable("复盘", ["ID", "userId", "againIntent", "content", "createdAt"], retrospectives.map(gameRetrospectiveRow).join("") || emptyRow(5, "暂无复盘"))}
        ${gameOpsTable("续局草稿", ["originalGameId", "draftGameId", "creatorUserId", "title", "status", "createdAt"], continueDrafts.map(gameContinueDraftRow).join("") || emptyRow(6, "暂无续局草稿"))}
      </div>
    </div>
  `;
}

function gameOpsTable(title, headers, rows) {
  return `
    <section>
      <h3>${escapeHTML(title)}</h3>
      <div class="table-wrap">
        <table>
          <thead><tr>${headers.map((item) => `<th>${escapeHTML(item)}</th>`).join("")}</tr></thead>
          <tbody>${rows}</tbody>
        </table>
      </div>
    </section>
  `;
}

function userFavoriteRow(item) {
  const game = item.game || {};
  return `
    <tr>
      <td>${escapeHTML(item.userId)}</td>
      <td>${escapeHTML(item.gameId)}</td>
      <td>${escapeHTML(game.title || "-")}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function inviteCodeRow(item) {
  const canDisable = item.status === "active" && can("invite_code:manage");
  const actions = [`<button class="ghost" data-action="invite-detail" data-code="${escapeHTML(item.code)}" type="button">详情</button>`];
  if (canDisable) actions.push(`<button class="ghost" data-action="invite-disable" data-code="${escapeHTML(item.code)}" type="button">禁用</button>`);
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td><code>${escapeHTML(item.code)}</code></td>
      <td>${escapeHTML(item.ownerUserId || "-")}</td>
      <td>${escapeHTML(item.entryType || "-")}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${escapeHTML(item.usedCount || 0)} / ${escapeHTML(item.maxUses || 0)}</td>
      <td>${escapeHTML(item.boundWechatUserId || item.boundWechatNickname || "-")}</td>
      <td><div class="row-actions">${actions.join("")}</div></td>
    </tr>
  `;
}

function inviteRelationRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.inviteCodeId)}</td>
      <td>${escapeHTML(item.inviterUserId || "-")}</td>
      <td>${escapeHTML(item.inviteeUserId)}</td>
      <td>${escapeHTML(item.bindSource || "-")}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function inviteRelationDetailRow(item, entryType) {
  return `
    <tr>
      <td>${escapeHTML(item.inviteCodeId)}</td>
      <td>${escapeHTML(item.inviterUserId || "-")}</td>
      <td>${escapeHTML(item.inviteeUserId)}</td>
      <td><code>${escapeHTML(item.bindSource || "-")}</code></td>
      <td>${inviteEntryLabel(entryType)}</td>
    </tr>
  `;
}

function inviteEntryLabel(entryType) {
  const labels = {
    poster: "poster 小程序卡片",
    qrcode: "qrcode 二维码",
    link: "link 链接",
  };
  return labels[entryType] || entryType || "-";
}

function revenueRecordRow(item) {
  const canFreeze = item.status !== "frozen" && item.status !== "settled";
  const canSettle = item.status === "pending_settlement";
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td><code>${escapeHTML(item.recordNo)}</code></td>
      <td>${escapeHTML(item.gameId)}</td>
      <td>${escapeHTML(item.templateId)}</td>
      <td>${escapeHTML(item.amountCent)}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${escapeHTML((item.items || []).map((child) => `${child.role}:${child.userId || "-"}:${child.amountCent}`).join(" / "))}</td>
      <td>
        <div class="row-actions">
          <button class="ghost" data-action="revenue-detail" data-id="${item.id}" type="button">详情</button>
          ${canFreeze ? `<button class="ghost" data-action="revenue-freeze" data-id="${item.id}" type="button">冻结</button>` : ""}
          ${canSettle ? `<button class="ghost" data-action="revenue-settle" data-id="${item.id}" type="button">结算</button>` : ""}
        </div>
      </td>
    </tr>
  `;
}

function redemptionItemRow(item) {
  const nextStatus = item.status === "active" ? "inactive" : "active";
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td><input data-field="name" value="${escapeHTML(item.name || "")}" /></td>
      <td><input data-field="pointsCost" inputmode="numeric" value="${escapeHTML(item.pointsCost || 0)}" /></td>
      <td><input data-field="stock" inputmode="numeric" value="${escapeHTML(item.stock || 0)}" /></td>
      <td>
        <select data-field="status">
          <option value="active" ${item.status === "active" ? "selected" : ""}>active</option>
          <option value="inactive" ${item.status === "inactive" ? "selected" : ""}>inactive</option>
        </select>
        <input data-field="description" type="hidden" value="${escapeHTML(item.description || "")}" />
      </td>
      <td>${formatTime(item.updatedAt || item.createdAt)}</td>
      <td>
        <div class="row-actions">
          <button class="ghost" data-action="redemption-item-update" data-id="${item.id}" type="button">保存</button>
          <button class="ghost" data-action="redemption-item-status" data-id="${item.id}" data-status="${nextStatus}" type="button">${nextStatus}</button>
        </div>
      </td>
    </tr>
  `;
}

function redemptionOrderRow(item) {
  const actions = [];
  if (item.status === "pending") {
    actions.push(`<button class="ghost" data-action="redemption-order-approve" data-id="${item.id}" type="button">通过</button>`);
    actions.push(`<button class="ghost" data-action="redemption-order-reject" data-id="${item.id}" type="button">驳回</button>`);
  }
  if (item.status === "approved") {
    actions.push(`<button class="ghost" data-action="redemption-order-fulfill" data-id="${item.id}" data-status="fulfilled" data-reason="admin fulfilled" type="button">履约</button>`);
  }
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td><code>${escapeHTML(item.orderNo || "-")}</code></td>
      <td>${escapeHTML(item.userId)}</td>
      <td>${escapeHTML(item.itemName || item.itemId || "-")}</td>
      <td>${escapeHTML(item.pointsCost)}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${formatTime(item.createdAt)}</td>
      <td><div class="row-actions">${actions.join("") || "-"}</div></td>
    </tr>
  `;
}

function pointsLogRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td>${escapeHTML(item.userId)}</td>
      <td>${escapeHTML(item.changeValue)}</td>
      <td>${escapeHTML(item.beforePoints)} -> ${escapeHTML(item.afterPoints)}</td>
      <td>${escapeHTML(item.bizType || "-")} #${escapeHTML(item.bizId || "-")}</td>
      <td>${escapeHTML(item.reason || "-")}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function memberTeamRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td>${escapeHTML(item.leaderUserId || "-")}</td>
      <td>${escapeHTML(item.name || "-")}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${formatTime(item.createdAt)}</td>
      <td><button class="ghost" data-action="member-team-detail" data-id="${item.id}" type="button">详情</button></td>
    </tr>
  `;
}

function memberReportRow(item) {
  const income = item.incomeSummary || {};
  return `
    <tr>
      <td>${escapeHTML(item.userId)}</td>
      <td>${escapeHTML(item.period || item.reportType || "-")}</td>
      <td>${escapeHTML(item.membershipPlan || "-")}</td>
      <td>${escapeHTML(item.invitedCount || 0)}</td>
      <td>${escapeHTML(item.participatedGames || 0)} / ${escapeHTML(item.completedGames || 0)}</td>
      <td>${escapeHTML(income.totalCent || 0)} / ${escapeHTML(income.pendingCent || 0)} / ${escapeHTML(income.settledCent || 0)}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function memberTeamMemberRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.userId)}</td>
      <td>${escapeHTML(item.relationLevel || 0)}</td>
      <td>${escapeHTML(item.source || "-")}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${formatTime(item.joinedAt)}</td>
    </tr>
  `;
}

function connectionRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td>${escapeHTML(item.userId)}</td>
      <td>${escapeHTML(item.connectedUserId)}</td>
      <td><code>${escapeHTML(item.relationType || "-")}</code></td>
      <td>${escapeHTML(item.sourceType || "-")} #${escapeHTML(item.sourceId || "-")}</td>
      <td>${escapeHTML(item.strengthScore || 0)}</td>
      <td>${formatTime(item.updatedAt)}</td>
    </tr>
  `;
}

function reportRow(item) {
  const canAssign = item.status === "pending";
  const canHandle = item.status === "pending" || item.status === "assigned";
  const canHandleCreditAppeal = item.reportType === "credit_appeal" && item.status === "appealed";
  const canClose = item.status !== "closed";
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td>${escapeHTML(item.gameId)}</td>
      <td>${escapeHTML(item.reporterUserId)}</td>
      <td>${escapeHTML(item.targetUserId || "-")}</td>
      <td>${escapeHTML(item.reportType)}</td>
      <td>${escapeHTML(item.reviewId || "-")}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${escapeHTML(item.revenueFrozen ? "true" : "false")}</td>
      <td>
        <div class="row-actions">
          <button class="ghost" data-action="report-detail" data-id="${item.id}" type="button">详情</button>
          ${canAssign ? `<button class="ghost" data-action="report-assign" data-id="${item.id}" type="button">分配</button>` : ""}
          ${canHandle ? `<button class="ghost" data-action="report-handle" data-id="${item.id}" type="button">处理</button>` : ""}
          ${canHandleCreditAppeal ? `<button class="ghost" data-action="credit-appeal-approve" data-id="${item.id}" type="button">通过</button>` : ""}
          ${canHandleCreditAppeal ? `<button class="ghost" data-action="credit-appeal-reject" data-id="${item.id}" type="button">驳回</button>` : ""}
          ${canClose ? `<button class="ghost" data-action="report-close" data-id="${item.id}" type="button">关闭</button>` : ""}
        </div>
      </td>
    </tr>
  `;
}

function feedbackRow(item) {
  const canReply = can("feedback:reply") && item.statusClass !== "resolved";
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td>${escapeHTML(item.userId)}</td>
      <td>${escapeHTML(item.type || item.typeKey || "-")}</td>
      <td><span class="${badgeClass(item.statusClass || "pending")}">${escapeHTML(item.status || statusLabel(item.statusClass || "pending"))}</span></td>
      <td>${escapeHTML(item.content || "-")}</td>
      <td>${escapeHTML(item.time || formatTime(item.createdAt))}</td>
      <td>
        <div class="row-actions">
          ${canReply ? `<button class="ghost" data-action="feedback-reply" data-id="${escapeHTML(item.id)}" data-user-id="${escapeHTML(item.userId)}" type="button">回复</button>` : ""}
        </div>
      </td>
    </tr>
  `;
}

function exportTaskRow(item) {
  const canDownload = item.status === "done" && item.fileId;
  const actions = [`<button class="ghost" data-action="export-detail" data-id="${item.id}" type="button">详情</button>`];
  if (canDownload) actions.push(`<button class="ghost" data-action="export-download" data-id="${item.id}" type="button">下载地址</button>`);
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td><code>${escapeHTML(item.taskNo)}</code></td>
      <td>${escapeHTML(item.templateCode)}</td>
      <td>${escapeHTML(item.exportType)}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${escapeHTML(item.fileId || "-")}</td>
      <td>${formatTime(item.createdAt)}</td>
      <td><div class="row-actions">${actions.join("")}</div></td>
    </tr>
  `;
}

function funnelRow(item) {
  return `
    <tr>
      <td><code>${escapeHTML(item.eventCode || "-")}</code></td>
      <td>${escapeHTML(item.userCount || 0)}</td>
      <td>${percentText(item.conversionRate)}</td>
      <td>${percentText(item.dropOffRate)}</td>
    </tr>
  `;
}

function retentionRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.cohortDate || "-")}</td>
      <td>${escapeHTML(item.newUsers || 0)}</td>
      <td>${escapeHTML(item.day1Retained || 0)} / ${percentText(item.day1Rate)}</td>
      <td>${escapeHTML(item.day7Retained || 0)} / ${percentText(item.day7Rate)}</td>
      <td>${escapeHTML(item.day30Retained || 0)} / ${percentText(item.day30Rate)}</td>
    </tr>
  `;
}

function behaviorEventRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td>${escapeHTML(item.userId || "-")}</td>
      <td>${escapeHTML(item.eventType || "-")}</td>
      <td><code>${escapeHTML(item.eventCode || "-")}</code></td>
      <td>${escapeHTML(item.targetType || item.businessType || "-")} #${escapeHTML(item.targetId || item.businessId || "-")}</td>
      <td>${escapeHTML(item.source || "-")}</td>
      <td>${formatTime(item.createdAt || item.occurredAt)}</td>
    </tr>
  `;
}

function behaviorLogItem(item) {
  return stackItem({
    title: `${item.eventCode || item.eventType || "event"} #${item.id || "-"}`,
    badge: item.eventType || "behavior",
    meta: [
      `userId=${item.userId || "-"}`,
      `target=${item.targetType || item.businessType || "-"} #${item.targetId || item.businessId || "-"}`,
      `page=${item.pagePath || "-"}`,
      `source=${item.source || "-"}`,
      `device=${item.device || "-"}`,
      `keyword=${item.keyword || "-"}`,
      `occurredAt=${formatTime(item.occurredAt || item.createdAt)}`,
    ],
  });
}

function wechatTaskRow(item) {
  const actions = [];
  if (item.status === "pending") {
    actions.push(`<button class="ghost" data-action="wechat-send" data-id="${item.id}" type="button">发送</button>`);
    actions.push(`<button class="ghost" data-action="wechat-mark-sent" data-id="${item.id}" type="button">标记已发</button>`);
  }
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td>${escapeHTML(item.notificationId || "-")}</td>
      <td>${escapeHTML(item.userId || "-")}</td>
      <td>${escapeHTML(item.scene || "-")}</td>
      <td><code>${escapeHTML(item.templateId || "-")}</code></td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${escapeHTML(item.resultCode || "-")} ${escapeHTML(item.resultMessage || "")}</td>
      <td><div class="row-actions">${actions.join("") || "-"}</div></td>
    </tr>
  `;
}

function deliveryDocumentRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td>${escapeHTML(item.docType)}</td>
      <td>${escapeHTML(item.title)}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${escapeHTML(item.reason || "-")}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function testCaseRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td>${escapeHTML(item.module)}</td>
      <td>${escapeHTML(item.caseName)}</td>
      <td><span class="${badgeClass(item.priority)}">${escapeHTML(item.priority || "-")}</span></td>
      <td>${escapeHTML(item.expectedResult || "-")}</td>
    </tr>
  `;
}

function testRunRow(item) {
  const evidenceAction = item.evidenceFileId
    ? `<button class="ghost" data-action="test-run-file-download" data-file-id="${escapeHTML(item.evidenceFileId)}" type="button">下载</button>`
    : "-";
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td>${escapeHTML(item.caseId)}</td>
      <td><span class="${badgeClass(item.result)}">${statusLabel(item.result)}</span></td>
      <td><code>${escapeHTML(item.requestId || "-")}</code></td>
      <td>${escapeHTML(item.evidenceFileId || "-")} ${evidenceAction}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function sensitiveWordRow(item) {
  const nextStatus = item.status === "active" ? "disabled" : "active";
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td><code>${escapeHTML(item.word)}</code></td>
      <td>${escapeHTML(item.level)}</td>
      <td>${escapeHTML(item.action)}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${formatTime(item.createdAt)}</td>
      <td><button class="ghost" data-action="sensitive-status" data-id="${item.id}" data-status="${nextStatus}" type="button">${nextStatus}</button></td>
    </tr>
  `;
}

function imMessageRow(item) {
  const canHide = item.status !== "hidden" && can("im:message:hide");
  const fileAction = item.fileId ? `<button class="ghost" data-action="admin-file-download" data-file-id="${escapeHTML(item.fileId)}" type="button">文件</button>` : "";
  const hideAction = canHide ? `<button class="ghost" data-action="im-message-hide" data-id="${escapeHTML(item.id)}" data-room-id="${escapeHTML(item.roomId)}" type="button">隐藏</button>` : "";
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td>${escapeHTML(item.senderUserId || "-")}</td>
      <td><code>${escapeHTML(item.messageType || "-")}</code></td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${escapeHTML(item.fileId || "-")}</td>
      <td>${escapeHTML(compactList(item.ackedBy || [], 4))}</td>
      <td>${escapeHTML(compactList(item.readBy || [], 4))}</td>
      <td>${formatTime(item.createdAt)}</td>
      <td><div class="row-actions">${fileAction}${hideAction}</div></td>
    </tr>
  `;
}

function operationLogRow(item, index) {
  return `
    <tr>
      <td>${escapeHTML(item.action)}</td>
      <td>${escapeHTML(item.targetType)} #${escapeHTML(item.targetId)}</td>
      <td>${escapeHTML(item.adminUserId || "-")}</td>
      <td>${formatTime(item.createdAt)}</td>
      <td>
        <div class="row-actions">
          <code>${escapeHTML(compactJSON(item.detail || {}))}</code>
          <button class="ghost" data-action="log-detail" data-index="${index}" type="button">详情</button>
        </div>
      </td>
    </tr>
  `;
}

function adminUserRow(item) {
  const actions = [];
  if (can("admin_user:update")) {
    actions.push(`<button class="ghost" data-action="admin-user-update" data-id="${item.id}" type="button">保存</button>`);
    actions.push(item.status === "active"
      ? `<button class="ghost" data-action="admin-user-disable" data-id="${item.id}" type="button">禁用</button>`
      : `<button class="ghost" data-action="admin-user-enable" data-id="${item.id}" type="button">启用</button>`);
  }
  return `
    <tr>
      <td>${escapeHTML(item.id)}</td>
      <td><code>${escapeHTML(item.username)}</code></td>
      <td><input data-field="roles" value="${escapeHTML((item.roles || []).join(", "))}" /></td>
      <td>
        <select data-field="status">
          <option value="active" ${item.status === "active" ? "selected" : ""}>active</option>
          <option value="disabled" ${item.status === "disabled" ? "selected" : ""}>disabled</option>
        </select>
      </td>
      <td>${escapeHTML(item.permissionCount || 0)}</td>
      <td>
        <div class="row-actions">
          ${actions.join("") || "-"}
        </div>
      </td>
    </tr>
  `;
}

function permissionCatalogRow(item) {
  return `
    <tr>
      <td><code>${escapeHTML(item.code)}</code></td>
      <td>${escapeHTML(item.module || "-")}</td>
      <td>${escapeHTML(item.action || "-")}</td>
    </tr>
  `;
}

function gameRow(game) {
  const canAudit = game.status === "pending_audit" && can("game:update_status");
  return `
    <tr>
      <td>${game.id}</td>
      <td>${escapeHTML(game.title)}</td>
      <td>${gameTypeLabel(game.gameType)}</td>
      <td>${escapeHTML(game.gameSource)}</td>
      <td><span class="${badgeClass(game.status)}">${statusLabel(game.status)}</span></td>
      <td>${game.currentPlayers}/${game.minPlayers}-${game.maxPlayers}</td>
      <td>${escapeHTML(game.address || game.cityName || game.cityCode || "-")}</td>
      <td>
        <div class="row-actions">
          <button class="ghost" data-action="detail" data-id="${game.id}" type="button">链路</button>
          ${canAudit ? `<button class="ghost" data-action="audit" data-id="${game.id}" type="button">通过</button>` : ""}
        </div>
      </td>
    </tr>
  `;
}

function gameMilestoneRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.id || "-")}</td>
      <td>${escapeHTML(item.title || "-")}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function gameCheckinRow(item) {
  const canMarkInvalid = item.status !== "invalid" && can("game:progress:manage");
  return `
    <tr>
      <td>${escapeHTML(item.id || "-")}</td>
      <td>${escapeHTML(item.userId || "-")}</td>
      <td>${escapeHTML(item.milestoneId || "-")}</td>
      <td>${escapeHTML(item.checkinType || "-")}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${formatTime(item.createdAt)}</td>
      <td>${canMarkInvalid ? `<button class="ghost" data-action="checkin-invalid" data-id="${item.id}" type="button">标记异常</button>` : "-"}</td>
    </tr>
  `;
}

function gameRetrospectiveRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.id || "-")}</td>
      <td>${escapeHTML(item.userId || "-")}</td>
      <td>${escapeHTML(item.againIntent || "-")}</td>
      <td>${escapeHTML(item.content || "-")}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function gameContinueDraftRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.originalGameId || "-")}</td>
      <td>${escapeHTML(item.draftGameId || "-")}</td>
      <td>${escapeHTML(item.creatorUserId || "-")}</td>
      <td>${escapeHTML(item.title || "-")}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function renderGameTypeBars(games) {
  const counts = games.reduce((acc, item) => {
    acc[item.gameType] = (acc[item.gameType] || 0) + 1;
    return acc;
  }, {});
  const max = Math.max(1, ...Object.values(counts));
  $("#game-type-bars").innerHTML = ["free", "standard", "public_welfare", "aa", "crowdfund", "deposit", "condition"].map((type) => {
    const count = counts[type] || 0;
    return `
      <div class="bar-row">
        <span>${gameTypeLabel(type)}</span>
        <div class="bar-track"><div class="bar-fill" style="width:${Math.round((count / max) * 100)}%"></div></div>
        <strong>${count}</strong>
      </div>
    `;
  }).join("");
}

function stackItem({ title, badge, meta, action = "" }) {
  return `
    <article class="stack-item">
      <div class="stack-item-head">
        <strong>${escapeHTML(title)}</strong>
        <span class="${badgeClass(badge)}">${escapeHTML(statusLabel(badge))}</span>
      </div>
      <div class="kv">${meta.map((item) => `<span>${escapeHTML(item)}</span>`).join("")}</div>
      ${action ? `<div class="row-actions">${action}</div>` : ""}
    </article>
  `;
}

function profileDetailBlock(title, data, rows) {
  const completeness = Number(data?.completeness || 0);
  const badge = completeness >= 100 ? "success" : completeness > 0 ? "warning" : "neutral";
  return `
    <article class="stack-item">
      <div class="stack-item-head">
        <strong>${escapeHTML(title)}</strong>
        <span class="badge ${badge}">${escapeHTML(completeness)}%</span>
      </div>
      <div class="detail-grid profile-detail-grid">
        ${rows.map(([label, value]) => detailCell(label, value)).join("")}
      </div>
    </article>
  `;
}

function growthTraceBlock(title, trace) {
  const profile = trace?.profile || {};
  const reviews = trace?.reviews || [];
  const creditLogs = trace?.creditLogs || [];
  const footprints = trace?.footprints || [];
  const achievements = trace?.achievements || [];
  return `
    <article class="stack-item">
      <div class="stack-item-head">
        <strong>${escapeHTML(title)} #${escapeHTML(trace?.userId || trace?.gameId || "-")}</strong>
        <span class="${badgeClass(reviews.length ? "active" : "pending")}">${escapeHTML(reviews.length)} 条评价</span>
      </div>
      <div class="detail-grid profile-detail-grid">
        ${detailCell("level", profile.level || 0)}
        ${detailCell("experience", profile.experience || 0)}
        ${detailCell("creditScore", profile.creditScore || profile.todayCreditScore || 100)}
        ${detailCell("availablePoints", profile.availablePoints || 0)}
        ${detailCell("reviewCount", profile.reviewCount || reviews.length)}
        ${detailCell("achievements", compactList(profile.achievements || achievements.map((item) => item.code)))}
      </div>
      <div class="sub-panel">
        <h3>评价记录</h3>
        <div class="table-wrap">
          <table>
            <thead><tr><th>ID</th><th>gameId</th><th>reviewer</th><th>target</th><th>score</th><th>againIntent</th><th>createdAt</th></tr></thead>
            <tbody>${reviews.map(reviewTraceRow).join("") || emptyRow(7, "暂无评价记录")}</tbody>
          </table>
        </div>
      </div>
      <div class="sub-panel">
        <h3>信用流水</h3>
        <div class="table-wrap">
          <table>
            <thead><tr><th>ID</th><th>userId</th><th>gameId</th><th>change</th><th>before/after</th><th>reason</th><th>createdAt</th></tr></thead>
            <tbody>${creditLogs.map(creditLogRow).join("") || emptyRow(7, "暂无信用流水")}</tbody>
          </table>
        </div>
      </div>
      <div class="sub-panel">
        <h3>足迹证据</h3>
        <div class="table-wrap">
          <table>
            <thead><tr><th>userId</th><th>gameId</th><th>action</th><th>createdAt</th></tr></thead>
            <tbody>${footprints.map(footprintRow).join("") || emptyRow(4, "暂无足迹")}</tbody>
          </table>
        </div>
      </div>
      <div class="sub-panel">
        <h3>成就</h3>
        <div class="kv">${achievements.map((item) => `<span>${escapeHTML(item.code || "-")} / ${escapeHTML(item.title || "-")}</span>`).join("") || "<span>暂无成就</span>"}</div>
      </div>
    </article>
  `;
}

function reviewTraceRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.id || "-")}</td>
      <td>${escapeHTML(item.gameId || "-")}</td>
      <td>${escapeHTML(item.reviewerUserId || "-")}</td>
      <td>${escapeHTML(item.targetUserId || "-")} / ${escapeHTML(item.targetRole || "-")}</td>
      <td>${escapeHTML(item.score || 0)}</td>
      <td>${escapeHTML(item.againIntent || "-")}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function creditLogRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.id || "-")}</td>
      <td>${escapeHTML(item.userId || "-")}</td>
      <td>${escapeHTML(item.gameId || "-")}</td>
      <td>${escapeHTML(item.changeValue || 0)}</td>
      <td>${escapeHTML(item.beforeScore || 0)} -> ${escapeHTML(item.afterScore || 0)}</td>
      <td>${escapeHTML(item.reason || "-")}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function footprintRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.userId || "-")}</td>
      <td>${escapeHTML(item.gameId || "-")}</td>
      <td><code>${escapeHTML(item.action || "-")}</code></td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function detailCell(label, value) {
  return `<div class="detail-cell"><span>${escapeHTML(label)}</span><strong>${escapeHTML(value)}</strong></div>`;
}

function template(id) {
  const node = document.getElementById(id);
  return node ? node.innerHTML : "";
}

function querySuffix(formData) {
  if (!formData) return "";
  const query = new URLSearchParams();
  for (const [key, value] of formData.entries()) {
    const text = String(value).trim();
    if (text) query.set(key, text);
  }
  const qs = query.toString();
  return qs ? `?${qs}` : "";
}

function querySuffixFromObject(data) {
  const query = new URLSearchParams();
  Object.entries(data || {}).forEach(([key, value]) => {
    const text = String(value ?? "").trim();
    if (text) query.set(key, text);
  });
  const qs = query.toString();
  return qs ? `?${qs}` : "";
}

function numberOrZero(value) {
  const number = Number(value || 0);
  return Number.isFinite(number) ? number : 0;
}

function parseIDList(value) {
  return String(value || "")
    .split(",")
    .map((item) => Number(item.trim()))
    .filter((item) => Number.isFinite(item) && item > 0);
}

function adminID() {
  return numberOrZero(state.admin?.id || state.admin?.adminUserId || state.admin?.userId);
}

function roleSnapshot(value) {
  if (!value || typeof value !== "object") return "-";
  return Object.entries(value)
    .filter(([, enabled]) => Boolean(enabled))
    .map(([role]) => role)
    .join(", ") || "-";
}

function compactJSON(value) {
  const text = JSON.stringify(value || {});
  return text.length > 72 ? `${text.slice(0, 69)}...` : text;
}

function compactList(values, limit = 8) {
  const list = Array.isArray(values) ? values : [];
  if (list.length <= limit) return list.join(", ") || "-";
  return `${list.slice(0, limit).join(", ")} ... +${list.length - limit}`;
}

function percentText(value) {
  const number = Number(value || 0);
  if (!Number.isFinite(number)) return "0%";
  return `${Math.round(number * 10000) / 100}%`;
}

function checkText(item) {
  if (!item) return "-";
  return `${item.current || 0}/${item.required || 0} ${item.ready ? "ready" : "pending"}`;
}

function rolesFromRow(row) {
  if (!row) return [];
  const raw = row.querySelector("input[data-field='roles']")?.value || "";
  return raw.split(",").map((item) => item.trim()).filter(Boolean);
}

function setActiveNav() {
  $$("nav button").forEach((button) => button.classList.toggle("active", button.dataset.view === state.view));
}

function applyNavPermissions() {
  $$("nav button[data-view]").forEach((button) => {
    button.hidden = !isViewAllowed(button.dataset.view);
  });
}

function ensureAllowedView() {
  if (!isViewAllowed(state.view)) {
    state.view = firstAllowedView();
  }
}

function firstAllowedView() {
  const button = $$("nav button[data-view]").find((item) => isViewAllowed(item.dataset.view));
  return button?.dataset.view || "dashboard";
}

function isViewAllowed(view) {
  const menus = state.permissionTree?.menus || [];
  if (!menus.length) return view === "dashboard";
  return menus.some((item) => item.code === view);
}

function setAPIStatus(ok) {
  const el = $("#api-status");
  el.textContent = ok ? "接口正常" : "接口异常";
  el.className = `status-dot ${ok ? "ok" : "pending"}`;
}

function setField(name, value) {
  const el = document.querySelector(`[data-field="${name}"]`);
  if (el) el.textContent = value;
}

function can(permission) {
  return state.permissions.includes(permission);
}

function countArray(value) {
  return Array.isArray(value) ? value.length : Number(value || 0);
}

function emptyRow(colspan, text) {
  return `<tr><td colspan="${colspan}"><div class="empty">${text}</div></td></tr>`;
}

function emptyBlock(text) {
  return `<div class="empty">${text}</div>`;
}

function renderNoAccess(selector, message) {
  const element = $(selector);
  if (element) {
    element.innerHTML = emptyBlock(message || "无权限访问");
  }
}

function toast(message, isError = false) {
  const el = $("#toast");
  el.textContent = message;
  el.className = `toast ${isError ? "error" : ""}`;
  window.clearTimeout(toast.timer);
  toast.timer = window.setTimeout(clearToast, 3600);
}

function clearToast() {
  const el = $("#toast");
  el.textContent = "";
  el.className = "toast hidden";
}

function gameTypeLabel(value) {
  const option = (state.gameTypeOptions || []).find((item) => item.key === value);
  if (option) return option.name;
  return {
    free: "普通局",
    standard: "标准局",
    public_welfare: "公益局",
    aa: "AA 局",
    crowdfund: "众筹局",
    deposit: "押金局",
    condition: "条件局",
  }[value] || value || "-";
}

function roleLabel(value) {
  return { player: "玩家", expert: "行家", guide: "领路人" }[value] || value || "角色";
}

function statusLabel(value) {
  return {
    pending: "待处理",
    phone_bound: "已绑手机",
    pending_audit: "待审核",
    verified: "已实名",
    rejected: "已驳回",
    recruiting: "招募中",
    in_progress: "进行中",
    pending_review: "待评价",
    completed: "已完成",
    active: "正常",
    disabled: "已禁用",
    inactive: "已下架",
    archived: "已归档",
    approved: "已通过",
    fulfilled: "已履约",
    exhausted: "已用尽",
    pending_settlement: "待结算",
    frozen: "已冻结",
    settled: "已结算",
    assigned: "已分配",
    handled: "已处理",
    closed: "已关闭",
    done: "已完成",
    failed: "失败",
  }[value] || value || "-";
}

function badgeClass(value) {
  if (["verified", "recruiting", "completed", "approved", "fulfilled", "active", "settled", "handled", "closed", "done"].includes(value)) return "badge success";
  if (["pending", "pending_audit", "pending_review", "phone_bound", "pending_settlement", "assigned"].includes(value)) return "badge warning";
  if (["rejected", "failed", "disabled", "inactive", "frozen"].includes(value)) return "badge danger";
  return "badge neutral";
}

function formatTime(value) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return String(value);
  return date.toLocaleString("zh-CN", { hour12: false });
}

function escapeHTML(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}
