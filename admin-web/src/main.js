const state = {
  token: localStorage.getItem("zhw_admin_token") || "",
  admin: null,
  permissions: [],
  view: "dashboard",
  games: [],
  gameApplications: [],
  selectedGameAuditIds: new Set(),
  gameCreatorOptions: [],
  dashboard: null,
  imRooms: [],
  users: [],
  userDisplayMap: new Map(),
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
  adminApplications: [],
  aiExportConfig: null,
  identities: [],
  avatarAudits: [],
  roleApplications: [],
  gameApplicationConfig: null,
  gameAuditConfig: null,
  conditionRuleConfig: null,
  roleBenefitConfig: null,
  growthRewardRules: null,
  operationRules: null,
  creditDeductionRules: [],
  pagination: {},
  drawerRestore: null,
};

const DEFAULT_PAGE_SIZE = 10;
const GAME_MAP_SEARCH_COOLDOWN_MS = 3000;
const NAV_COLLAPSE_STORAGE_KEY = "zhw_admin_collapsed_nav_groups";

let gameMapSearchCooldownTimer = 0;

const views = {
  dashboard: { title: "运营工作台", crumb: "总览" },
  users: { title: "用户管理", crumb: "用户 / 实名 / 身份" },
  invites: { title: "邀请管理", crumb: "邀请码 / 关系 / 入口" },
  games: { title: "组局管理", crumb: "组局 / 开局 / 审核" },
  audits: { title: "角色审核", crumb: "认证记录 / 行家领路人申请" },
  revenue: { title: "分润结算", crumb: "分润 / 规则 / 结算" },
  redemption: { title: "积分兑换", crumb: "积分 / 兑换 / 订单" },
  members: { title: "会员团队", crumb: "会员 / 团队 / 报表" },
  profiles: { title: "画像关系", crumb: "用户画像 / 人脉 / 资源" },
  growth: { title: "评价成长", crumb: "评价 / 信用 / 足迹" },
  reports: { title: "举报申诉", crumb: "投诉 / 证据 / 处理" },
  exports: { title: "导出中心", crumb: "报表 / 任务 / 文件" },
  analytics: { title: "数据分析", crumb: "行为 / 漏斗 / 智能数据" },
  delivery: { title: "通知交付", crumb: "微信通知 / 上线材料 / 验收" },
  admins: { title: "后台账号", crumb: "账号 / 角色" },
  system: { title: "运营规则", crumb: "局规则 / 内容安全 / 权限" },
  im: { title: "局内消息证据", crumb: "房间 / 文件 / 争议" },
  logs: { title: "操作日志", crumb: "管理员操作 / 审计" },
};

const ADMIN_VISIBLE_VIEWS = new Set([
  "dashboard",
  "users",
  "invites",
  "games",
  "audits",
  "revenue",
  "redemption",
  "members",
  "profiles",
  "growth",
  "reports",
  "exports",
  "analytics",
  "delivery",
  "admins",
  "system",
  "im",
  "logs",
]);

const DEFAULT_GAME_TYPE_OPTIONS = [
  { key: "free", name: "普通局", selectable: true },
  { key: "standard", name: "标准局", selectable: true },
  { key: "public_welfare", name: "公益局", selectable: true },
  { key: "aa", name: "均摊局", selectable: true },
  { key: "crowdfund", name: "众筹局", selectable: true },
  { key: "deposit", name: "押金局", selectable: true },
  { key: "condition", name: "条件局", selectable: true },
];

const $ = (selector, root = document) => root.querySelector(selector);
const $$ = (selector, root = document) => Array.from(root.querySelectorAll(selector));

document.addEventListener("DOMContentLoaded", () => {
  installAdminCopyObserver();
  bindShell();
  if (state.token) {
    boot();
  }
});

function installAdminCopyObserver() {
  const root = $("#app-shell");
  if (!root || installAdminCopyObserver.started) return;
  installAdminCopyObserver.started = true;
  let timer = 0;
  const observer = new MutationObserver((mutations) => {
    if (!mutations.some((item) => item.addedNodes.length > 0)) return;
    window.clearTimeout(timer);
    timer = window.setTimeout(() => polishAdminFragment(root), 0);
  });
  observer.observe(root, { childList: true, subtree: true });
}

function bindShell() {
  $("#login-form").addEventListener("submit", login);
  $("#admin-apply-form").addEventListener("submit", submitAdminApplication);
  $$("[data-login-panel]").forEach((button) => {
    button.addEventListener("click", () => switchLoginPanel(button.dataset.loginPanel));
  });
  $("#logout-button").addEventListener("click", logout);
  $("#refresh-button").addEventListener("click", () => loadView(state.view));
  document.addEventListener("click", (event) => {
    if (event.target.closest("[data-action='drawer-close']") || event.target.classList.contains("admin-drawer-backdrop")) {
      closeAdminDrawer();
    }
  });
  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape") closeAdminDrawer();
  });
  $("#main-nav").addEventListener("click", (event) => {
    const groupButton = event.target.closest("button[data-nav-group]");
    if (groupButton) {
      toggleNavGroup(groupButton.dataset.navGroup);
      return;
    }
    const button = event.target.closest("button[data-view]");
    if (!button) return;
    state.view = button.dataset.view;
    setActiveNav();
    expandActiveNavGroup();
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
    $("#admin-role").textContent = `账号：${state.admin?.username || "-"} / ${roleSnapshot(state.admin?.roles || [])}`;
    setAPIStatus(true);
    applyNavPermissions();
    ensureAllowedView();
    setActiveNav();
    restoreNavGroups();
    await loadView(state.view);
  } catch (error) {
    setAPIStatus(false);
    showLogin();
    showLoginError(error.message);
  }
}

function switchLoginPanel(panel) {
  const target = panel === "apply" ? "apply" : "login";
  $$("[data-login-content]").forEach((item) => {
    item.classList.toggle("hidden", item.dataset.loginContent !== target);
  });
  showLoginError("");
  const result = $("#admin-apply-result");
  if (result) result.textContent = "";
}

async function login(event) {
  event.preventDefault();
  showLoginError("");
  const username = $("#login-username").value.trim();
  const password = $("#login-password").value;
  try {
    const data = await apiPost("/api/admin/auth/login", { username, password }, false);
    state.token = data.token;
    state.userDisplayMap.clear();
    localStorage.setItem("zhw_admin_token", state.token);
    await boot();
  } catch (error) {
    showLoginError(error.message);
  }
}

async function submitAdminApplication(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const resultBox = $("#admin-apply-result");
  const data = Object.fromEntries(new FormData(form).entries());
  try {
    const item = await apiPost("/api/admin/auth/applications", {
      name: String(data.name || "").trim(),
      contact: String(data.contact || "").trim(),
      desiredRole: String(data.desiredRole || "").trim(),
      reason: String(data.reason || "").trim(),
    }, false);
    resultBox.textContent = `申请已提交，编号 ${item.id}。请等待超级管理员在“后台账号 / 账号申请”里处理。`;
    form.reset();
  } catch (error) {
    resultBox.textContent = humanMessage(error.message);
  }
}

function logout() {
  state.token = "";
  state.admin = null;
  state.permissions = [];
  state.userDisplayMap.clear();
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
  await ensureUserDisplayCache();
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
  polishAdminFragment(root);
}

async function renderDashboard() {
  const canReadGames = can("game:read") || can("game:view");
  const [dashboard, games, imRooms] = await Promise.all([
    apiGet("/api/admin/dashboard"),
    canReadGames ? apiGet("/api/admin/games") : Promise.resolve({ items: [] }),
    can("im:room:read") ? apiGet("/api/admin/im/rooms") : Promise.resolve({ items: [] }),
  ]);
  state.dashboard = dashboard;
  state.games = games.items || [];
  state.imRooms = imRooms.items || [];

  setField("behaviorCount", countArray(dashboard.behaviorEvents || dashboard.events || dashboard.behaviorLogs));
  setField("gameCount", state.games.length);
  const pendingGames = state.games.filter((item) => item.status === "pending_audit").length;
  setField("pendingGameCount", pendingGames);
  const gamesNavButton = document.querySelector('#main-nav button[data-view="games"]');
  if (gamesNavButton) {
    gamesNavButton.classList.toggle("has-pending-dot", pendingGames > 0);
    gamesNavButton.setAttribute("data-pending-count", String(pendingGames));
  }
  setField("imRoomCount", state.imRooms.length);
  try {
    const pending = await apiGet("/api/admin/pending-counts");
    const counts = pending.counts || {};
    const navMap = {
      games: "games",
      reports: "reports",
      redemption: "redemption",
      audits: "audits",
    };
    Object.entries(navMap).forEach(([key, view]) => {
      const button = document.querySelector(`#main-nav button[data-view="${view}"]`);
      if (!button) return;
      const count = view === "audits"
        ? Number(counts.identity || 0) + Number(counts.enterprise || 0) + Number(counts.avatars || 0) + Number(counts.roles || 0)
        : Number(counts[key] || 0);
      button.classList.toggle("has-pending-dot", count > 0);
      button.setAttribute("data-pending-count", String(count));
    });
  } catch (error) {
    // Dashboard still renders when the optional aggregated counter permission
    // is unavailable; individual module pages remain the source of truth.
  }
  renderGameTypeBars(state.games);
  renderDashboardStatusBars(state.games);
  renderDashboardWorkQueue(state.games, state.imRooms);
  renderDashboardPermissionScope();
}

async function ensureUserDisplayCache() {
  if (!can("user:read")) return;
  try {
    const data = await apiGet("/api/admin/users");
    rememberUserDisplayItems(data.items || []);
  } catch (error) {
    // Nickname cache is best-effort; the page can still fall back to user IDs.
  }
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
  if (formData) resetPagination("users");
  const data = await apiGet(`/api/admin/users${querySuffix(formData)}`);
  state.users = data.items || [];
  rememberUserDisplayItems(state.users);
  renderPaginatedTable("#users-table", state.users, "users", userRow, 8, "暂无用户");
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
  rememberUserDisplayItems([user]);
  const identity = data.identity || {};
  const inviteRelation = data.inviteRelation || user.inviteRelation || {};
  const inviter = data.inviter || user.inviter || {};
  rememberUserDisplayItems([inviter]);
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
  openAdminDrawer({
    title: `${userText(user.id)}详情`,
    subtitle: user.nickname || wechatBindingText(user.openId),
    body: `
    <div class="row-actions">
      <span class="${badgeClass(user.realnameStatus)}">${statusLabel(user.realnameStatus)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("账号状态", statusLabel(user.status))}
      ${detailCell("实名状态", statusLabel(user.realnameStatus))}
      ${detailCell("认证状态", statusLabel(identity.status))}
      ${detailCell("绑定邀请码", user.inviteCode || "-")}
      ${detailCell("邀请人", inviterLabel(inviter, inviteRelation))}
      ${detailCell("用户身份", roleSnapshot(data.roles))}
      ${detailCell("成长等级", `Lv.${growth.level || 0}`)}
      ${detailCell("信用分", growth.creditScore || 0)}
      ${detailCell("可用积分", points.availablePoints || growth.points || 0)}
      ${detailCell("会员状态", statusLabel(membership.status))}
      ${detailCell("收藏数", `${favorites.length} 个`)}
      ${detailCell("人脉数", `${connections.length} 个`)}
      ${detailCell("累计收益", yuanText(income.totalCent))}
      ${detailCell("待结算收益", yuanText(income.pendingCent))}
      ${detailCell("已结算收益", yuanText(income.settledCent))}
    </div>
    ${technicalDetails("技术识别信息", [
      detailCell("微信标识", user.openId || "-"),
      detailCell("用户", user.id || "-"),
    ].join(""))}
    ${can("invite_code:manage") ? `
    <div class="sub-panel">
      <div class="panel-head">
        <div>
          <h2>邀请关系</h2>
        </div>
      </div>
      <form id="user-invite-relation-form" class="form-grid compact-grid invite-relation-form">
        <label class="user-picker-field">邀请人
          <div class="user-picker" data-user-picker>
            <input name="inviterUserText" type="text" autocomplete="off" required value="${escapeHTML(inviterPickerValue(inviter, inviteRelation))}" placeholder="输入用户编号或昵称搜索" data-user-picker-input />
            <input name="inviterUserId" type="hidden" value="${escapeHTML(inviterPickerID(inviter, inviteRelation))}" data-user-picker-value />
            <div class="user-picker-menu" data-user-picker-menu>
              ${userPickerOptions(inviter, inviteRelation).map(userPickerOptionButton).join("")}
              <div class="user-picker-empty" data-user-picker-empty hidden>没有匹配用户</div>
            </div>
          </div>
        </label>
        <button class="ghost" type="submit">保存邀请人</button>
      </form>
    </div>
    ` : ""}
    <div class="split profile-form-gap">
      ${gameOpsTable("收藏明细", ["用户", "组局", "局标题", "收藏时间"], favorites.map(userFavoriteRow).join("") || emptyRow(4, "暂无收藏"))}
      ${gameOpsTable("人脉明细", ["关系编号", "用户", "关联用户", "关系类型", "来源", "关系强度", "更新时间"], connections.map(connectionRow).join("") || emptyRow(7, "暂无人脉"))}
    </div>
  `,
  });
  bindUserInviteRelationForm(id);
}

async function renderInvites() {
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
  const entryType = ["poster", "qrcode", "link"].includes(data.entryType) ? data.entryType : "poster";
  const rawBatchCount = Number(data.batchCount || 1);
  const batchCount = Number.isFinite(rawBatchCount) ? Math.min(Math.max(rawBatchCount, 1), 200) : 1;
  const ownerUserId = Number(data.ownerUserId || 0);
  if (!Number.isInteger(ownerUserId) || ownerUserId <= 0) {
    toast("请输入有效的邀请人用户编号", true);
    return;
  }
  const payload = {
    code: String(data.code || "").trim(),
    ownerUserId,
    entryType,
    batchCount,
  };
  if (payload.code && !/^[A-Za-z0-9_-]{4,32}$/.test(payload.code)) {
    toast("邀请码只能使用 4-32 位字母、数字、下划线或短横线", true);
    return;
  }
  if (payload.batchCount > 1 && payload.code) {
    toast("批量生成时请留空邀请码，由系统自动生成", true);
    return;
  }
  try {
    const result = (await apiPost("/api/admin/invite-codes", payload)) || {};
    toast(payload.batchCount > 1 ? `已批量生成 ${result.total || payload.batchCount} 个邀请码` : "邀请码已创建");
    form.reset();
    const ownerInput = form.querySelector("[name='ownerUserId']");
    const entryTypeInput = form.querySelector("[name='entryType']");
    const batchCountInput = form.querySelector("[name='batchCount']");
    if (ownerInput) ownerInput.value = "";
    if (entryTypeInput) entryTypeInput.value = "poster";
    if (batchCountInput) batchCountInput.value = "1";
    const codeInput = form.querySelector("[name='code']");
    if (codeInput) codeInput.value = "";
    await loadInviteCodes();
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadInviteCodes(formData) {
  if (formData) resetPagination("inviteCodes");
  const data = await apiGet(`/api/admin/invite-codes${querySuffix(formData)}`);
  state.inviteCodes = data.items || [];
  renderPaginatedTable("#invite-codes-table", state.inviteCodes, "inviteCodes", inviteCodeRow, 7, "暂无邀请码");
}

async function loadInviteRelations(formData) {
  if (formData) resetPagination("inviteRelations");
  const data = await apiGet(`/api/admin/invite-relations${querySuffix(formData)}`);
  state.inviteRelations = data.items || [];
  renderPaginatedTable("#invite-relations-table", state.inviteRelations, "inviteRelations", inviteRelationRow, 5, "暂无邀请关系");
}

async function onInviteTableClick(event) {
  const button = event.target.closest("button[data-action]");
  if (!button) return;
  try {
    if (button.dataset.action === "invite-copy-usage") {
      copyInviteUsage(button.dataset.code);
    }
    if (button.dataset.action === "invite-materials") {
      await showInviteMaterials(button.dataset.code);
    }
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

async function showInviteCodeDetail(code) {
  if (!code) return;
  const data = await apiGet(`/api/admin/invite-codes/${encodeURIComponent(code)}`);
  const invite = data.inviteCode || {};
  const owner = data.ownerUser || {};
  const bound = data.boundUser || {};
  rememberUserDisplayItems([owner, bound]);
  const relations = data.relations || [];
  const usage = inviteUsageInfo(invite);
  const panel = openAdminDrawer({
    title: "邀请码详情",
    subtitle: invite.code || code,
    body: `
      <div class="row-actions">
        <span class="${badgeClass(invite.status)}">${statusLabel(invite.status)}</span>
        ${invite.status === "active" && can("invite_code:manage") ? `<button class="ghost" data-action="detail-invite-disable" data-code="${escapeHTML(invite.code)}" type="button">禁用</button>` : ""}
      </div>
    <div class="detail-grid">
      ${detailCell("编号", invite.id)}
      ${detailCell("邀请码", invite.code)}
      ${detailCell("入口类型", inviteEntryLabel(invite.entryType))}
      ${detailCell("创建来源", inviteOwnerText(invite, owner))}
      ${detailCell("使用状态", inviteUseText(invite))}
      ${detailCell("绑定用户", bound.nickname || invite.boundWechatNickname || boundUserText(invite))}
      ${detailCell("绑定关系数", relations.length)}
      ${detailCell("绑定来源", compactList(relations.map((item) => item.bindSource)))}
    </div>
    <div class="sub-panel">
      <div class="panel-head">
        <div>
          <h2>使用方式</h2>
          <p>把下面入口配置到对应的小程序卡片、二维码或海报中，用户进入后会自动带入邀请码。</p>
        </div>
        <div class="row-actions">
          <button class="ghost" data-action="detail-invite-materials" type="button">生成可发物料</button>
          <button class="ghost" data-action="detail-invite-copy-usage" type="button">复制发送文案</button>
        </div>
      </div>
      <div class="detail-grid profile-detail-grid">
        ${detailCell("小程序路径", usage.path)}
        ${detailCell("小程序码参数", usage.scene)}
        ${detailCell("链接参数", usage.query)}
        ${detailCell("使用说明", usage.tip)}
      </div>
      <div id="invite-materials-result"></div>
    </div>
    <div class="sub-panel">
      <div class="panel-head">
        <div>
          <h2>入口校验</h2>
          <p>一期小程序只允许通过小程序卡片、二维码、链接进入；唯一邀请码绑定微信后进入登录页。</p>
        </div>
      </div>
      <div class="detail-grid profile-detail-grid">
        ${detailCell("允许入口", inviteEntryLabel(invite.entryType))}
        ${detailCell("是否唯一绑定", "是")}
        ${detailCell("是否已绑定微信", invite.boundWechatUserId ? "是" : "否")}
        ${detailCell("进入页面", invite.boundWechatUserId ? "登录页" : "注册页")}
      </div>
    </div>
    <div class="sub-panel">
      <div class="panel-head">
        <div>
          <h2>绑定关系明细</h2>
        </div>
      </div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>邀请码</th>
              <th>被邀请用户</th>
              <th>绑定来源</th>
              <th>入口</th>
            </tr>
          </thead>
          <tbody>${relations.map((item) => inviteRelationDetailRow(item, invite.entryType)).join("") || emptyRow(4, "暂无绑定关系")}</tbody>
        </table>
      </div>
    </div>
  `,
  });
  const copyUsageButton = panel.querySelector("button[data-action='detail-invite-copy-usage']");
  if (copyUsageButton) {
    copyUsageButton.addEventListener("click", () => copyInviteUsage(invite.code));
  }
  const materialsButton = panel.querySelector("button[data-action='detail-invite-materials']");
  if (materialsButton) {
    materialsButton.addEventListener("click", () => showInviteMaterials(invite.code));
  }
  const disableButton = panel.querySelector("button[data-action='detail-invite-disable']");
  if (disableButton) {
    disableButton.addEventListener("click", async () => {
      await apiPost(`/api/admin/invite-codes/${encodeURIComponent(disableButton.dataset.code)}/disable`, {});
      toast(`邀请码 ${disableButton.dataset.code} 已禁用`);
      await Promise.all([showInviteCodeDetail(disableButton.dataset.code), loadInviteCodes(new FormData($("#invite-filter-form")))]);
    });
  }
}

async function renderGames() {
  $("#game-batch-audit-button").addEventListener("click", batchAuditGames);
  $("#game-batch-reject-button")?.addEventListener("click", batchRejectGames);
  $("#game-audit-select-all")?.addEventListener("change", onGameAuditSelectAllChange);
  $("#open-game-create-drawer")?.addEventListener("click", () => {
    openEmbeddedFormDrawer("#game-create-form", "后台开局", "填写封面、地点、报名时间和人数后创建并开放招募");
  });
  $("#game-create-form").addEventListener("submit", createAdminGame);
  $("#game-map-search-button")?.addEventListener("click", searchGameMapPlaces);
  $("#game-map-results")?.addEventListener("click", selectGameMapPlace);
  $("#game-cover-upload-button")?.addEventListener("click", uploadGameCoverImage);
  $("#game-create-form input[name='coverImage']")?.addEventListener("input", (event) => {
    renderGameCoverPreview(event.currentTarget.value);
  });
  $("#game-creator-select")?.addEventListener("change", syncGameCreatorHint);
  $("#game-filter-form").addEventListener("submit", (event) => {
    event.preventDefault();
    loadGames(new FormData(event.currentTarget));
  });
  $("#games-table").addEventListener("click", onGameTableClick);
  $("#games-table").addEventListener("change", onGameAuditCheckboxChange);
  ensureGameApplicationsPanel();
  await Promise.all([loadGameTypeOptionsForGames(), loadGameCreatorOptionsForGames()]);
  await loadGames();
}

async function uploadGameCoverImage() {
  const fileInput = $("#game-cover-file");
  const file = fileInput?.files?.[0];
  const status = $("#game-cover-upload-status");
  const uploadButton = $("#game-cover-upload-button");
  const form = $("#game-create-form");
  if (!file) {
    toast("请选择要上传的局封面图片", true);
    return;
  }
  const allowedTypes = new Set(["image/jpeg", "image/png", "image/webp"]);
  if (!allowedTypes.has(file.type)) {
    toast("局封面仅支持 JPG、PNG、WEBP", true);
    return;
  }
  if (file.size <= 0 || file.size > 10 * 1024 * 1024) {
    toast("局封面不能超过 10MB", true);
    return;
  }
  if (uploadButton) uploadButton.disabled = true;
  if (status) status.textContent = "正在获取上传凭证...";
  try {
    const result = await apiPost("/api/admin/files/upload-token", {
      bizType: "game_cover",
      objectId: 0,
      fileName: file.name || `game-cover-${Date.now()}.jpg`,
      mimeType: file.type,
      size: file.size,
    });
    const upload = result.upload || {};
    const download = result.download || {};
    if (status) status.textContent = "正在上传到 COS...";
    await uploadBrowserFile(file, upload);
    const coverURL = download.downloadUrl || download.downloadURL || result.downloadUrl || "";
    if (!coverURL) {
      throw new Error("上传成功但未返回封面访问地址");
    }
    setFormValue(form, "coverImage", coverURL);
    renderGameCoverPreview(coverURL);
    if (status) status.textContent = "封面已上传";
    toast("局封面已上传并填入");
  } catch (error) {
    if (status) status.textContent = humanMessage(error.message);
    toast(error.message, true);
  } finally {
    if (uploadButton) uploadButton.disabled = false;
  }
}

async function uploadBrowserFile(file, upload) {
  const uploadURL = upload?.uploadUrl || upload?.uploadURL || "";
  if (!uploadURL) {
    throw new Error("上传地址缺失");
  }
  if (/^(mock|local):\/\//i.test(uploadURL)) {
    return;
  }
  let response;
  try {
    if (upload.formData && Object.keys(upload.formData).length > 0) {
      const formData = new FormData();
      Object.entries(upload.formData).forEach(([key, value]) => {
        formData.append(key, value);
      });
      formData.append("file", file);
      response = await fetch(uploadURL, { method: "POST", body: formData });
    } else {
      response = await fetch(uploadURL, {
        method: "PUT",
        headers: upload.headers || {},
        body: file,
      });
    }
  } catch (error) {
    throw new Error("文件上传失败：浏览器无法连接 COS，请检查存储桶 CORS 是否允许后台域名");
  }
  if (!response.ok) {
    const detail = await response.text().catch(() => "");
    throw new Error(`文件上传失败：COS ${response.status}${detail ? ` ${detail.replace(/\s+/g, " ").slice(0, 160)}` : ""}`);
  }
}

function renderGameCoverPreview(url) {
  const preview = $("#game-cover-preview");
  if (!preview) return;
  const value = String(url || "").trim();
  preview.classList.toggle("hidden", !value);
  preview.innerHTML = value
    ? `<img src="${escapeHTML(value)}" alt="局封面预览" /><span>${escapeHTML(value)}</span>`
    : "";
}

function resetGameCoverUpload() {
  const fileInput = $("#game-cover-file");
  const status = $("#game-cover-upload-status");
  if (fileInput) fileInput.value = "";
  if (status) status.textContent = "支持 JPG/PNG/WEBP，不超过 10MB";
  renderGameCoverPreview("");
}

async function loadGameCreatorOptionsForGames() {
  const select = $("#game-creator-select");
  if (!select) return;
  if (!can("game:create_admin")) {
    select.innerHTML = `<option value="">缺少后台开局权限</option>`;
    select.disabled = true;
    syncGameCreatorHint();
    return;
  }
  select.disabled = true;
  select.innerHTML = `<option value="">正在加载用户...</option>`;
  try {
    const data = await apiGet("/api/admin/users/options");
    state.gameCreatorOptions = data.items || [];
    rememberUserDisplayItems(state.gameCreatorOptions);
    renderGameCreatorSelect();
    select.disabled = false;
  } catch (error) {
    select.innerHTML = `<option value="">用户列表加载失败</option>`;
    select.disabled = true;
    $("#game-creator-hint").textContent = humanMessage(error.message);
    toast(error.message, true);
  }
}

function renderGameCreatorSelect() {
  const select = $("#game-creator-select");
  if (!select) return;
  const current = select.value;
  const options = state.gameCreatorOptions || [];
  select.innerHTML = [
    `<option value="">请选择发起人</option>`,
    ...options.map((item) => `<option value="${Number(item.id) || ""}">${escapeHTML(gameCreatorOptionText(item))}</option>`),
  ].join("");
  if ([...select.options].some((option) => option.value === current)) {
    select.value = current;
  }
  syncGameCreatorHint();
}

function syncGameCreatorHint() {
  const hint = $("#game-creator-hint");
  const select = $("#game-creator-select");
  if (!hint || !select) return;
  const selected = state.gameCreatorOptions.find((item) => String(item.id) === String(select.value));
  hint.textContent = selected ? `已选择：${gameCreatorOptionText(selected)}` : "选择后将使用该用户编号创建局";
}

function gameCreatorOptionText(item) {
  const id = Number(item?.id) || 0;
  const nickname = String(item?.nickname || item?.nickName || item?.displayName || "").trim();
  if (id && nickname) return `用户 ${id} - ${nickname}`;
  if (id) return `用户 ${id}`;
  return nickname || "-";
}

async function searchGameMapPlaces() {
  const keyword = $("#game-map-keyword")?.value.trim() || "";
  const form = $("#game-create-form");
  const city = form?.cityName?.value.trim() || "";
  const resultBox = $("#game-map-results");
  const searchButton = $("#game-map-search-button");

  if (searchButton?.disabled) {
    toast("请稍后再搜索", true);
    return;
  }

  if (!keyword) {
    toast("请输入地点关键词", true);
    return;
  }

  startGameMapSearchCooldown(searchButton);
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

function startGameMapSearchCooldown(button) {
  if (!button) return;
  window.clearTimeout(gameMapSearchCooldownTimer);
  if (!button.dataset.defaultText) {
    button.dataset.defaultText = button.textContent || "搜索地点";
  }
  button.disabled = true;
  button.textContent = "稍后再搜";
  gameMapSearchCooldownTimer = window.setTimeout(() => {
    button.disabled = false;
    button.textContent = button.dataset.defaultText || "搜索地点";
    gameMapSearchCooldownTimer = 0;
  }, GAME_MAP_SEARCH_COOLDOWN_MS);
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
  setFormValue(form, "cityCode", place.cityCode || getFormValue(form, "cityCode"));
  setFormValue(form, "cityName", place.city || place.cityName || getFormValue(form, "cityName"));
  setFormValue(form, "address", place.address || place.title || getFormValue(form, "address"));
  setFormValue(form, "longitude", Number.isFinite(Number(place.longitude)) ? place.longitude : getFormValue(form, "longitude"));
  setFormValue(form, "latitude", Number.isFinite(Number(place.latitude)) ? place.latitude : getFormValue(form, "latitude"));
  toast("地点已填入开局表单");
}

async function loadGameTypeOptionsForGames() {
  if (!can("system_config:read")) return;
  try {
    const data = await apiGet("/api/admin/games/category-config");
    const config = data.config || data;
    const options = mergeGameTypeOptions(config.typeFilters);
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

function mergeGameTypeOptions(items = []) {
  const merged = new Map(DEFAULT_GAME_TYPE_OPTIONS.map((item) => [item.key, { ...item }]));
  normalizeGameTypeOptions(items).forEach((item) => {
    const preset = merged.get(item.key);
    merged.set(item.key, {
      ...item,
      name: preset?.name || item.name,
      selectable: item.selectable !== false,
    });
  });
  return [...merged.values()];
}

function renderGameTypeSelect(selector, options, includeAll) {
  const select = $(selector);
  if (!select) return;
  const current = select.value;
  const rows = [];
  if (includeAll) rows.push(`<option value="">全部局型</option>`);
  rows.push(...options.map((item) => `<option value="${escapeHTML(item.key)}" ${item.selectable ? "" : "disabled"}>${escapeHTML(item.name)}</option>`));
  select.innerHTML = rows.join("");
  if ([...select.options].some((option) => option.value === current)) {
    select.value = current;
  }
}

function categoryName(categories = [], key) {
  const list = Array.isArray(categories) ? categories : [];
  const found = list.find((item) => item && item.key === key);
  return found?.name || key || "-";
}

async function batchAuditGames() {
  const ids = [...state.selectedGameAuditIds].map((id) => Number(id)).filter(Boolean);
  if (!ids.length) {
    toast("请先勾选要批量通过的待审核组局", true);
    return;
  }
  try {
    const result = await apiPost("/api/admin/games/batch-audit", { gameIds: ids, approve: true, remark: "后台批量审核通过" });
    toast(`批量通过 ${result.success || 0} 个组局，失败 ${result.failed || 0} 个`);
    state.selectedGameAuditIds.clear();
    await loadGames(new FormData($("#game-filter-form")));
  } catch (error) {
    toast(error.message, true);
  }
}

async function batchRejectGames() {
  const ids = [...state.selectedGameAuditIds].map((id) => Number(id)).filter(Boolean)
  if (!ids.length) {
    toast("请先勾选要批量驳回的待审核组局", true)
    return
  }
  const reason = window.prompt("请输入批量驳回原因")
  if (reason == null || !String(reason).trim()) {
    toast("驳回必须填写原因", true)
    return
  }
  try {
    const result = await apiPost("/api/admin/games/batch-audit", { gameIds: ids, approve: false, remark: String(reason).trim() })
    toast(`批量驳回 ${result.success || 0} 个组局，失败 ${result.failed || 0} 个`)
    state.selectedGameAuditIds.clear()
    await loadGames(new FormData($("#game-filter-form")))
  } catch (error) {
    toast(error.message, true)
  }
}

function onGameAuditSelectAllChange(event) {
  const checked = Boolean(event.currentTarget.checked);
  document.querySelectorAll("#games-table .game-audit-checkbox").forEach((checkbox) => {
    checkbox.checked = checked;
    updateGameAuditSelection(checkbox.value, checked);
  });
  updateGameAuditSelectionUI();
}

function onGameAuditCheckboxChange(event) {
  const checkbox = event.target.closest(".game-audit-checkbox");
  if (!checkbox) return;
  updateGameAuditSelection(checkbox.value, checkbox.checked);
  updateGameAuditSelectionUI();
}

function updateGameAuditSelection(id, checked) {
  const value = String(id || "");
  if (!value) return;
  if (checked) {
    state.selectedGameAuditIds.add(value);
  } else {
    state.selectedGameAuditIds.delete(value);
  }
}

function pruneGameAuditSelection() {
  const pendingIDs = new Set(state.games.filter((item) => item.status === "pending_audit").map((item) => String(item.id)));
  state.selectedGameAuditIds = new Set([...state.selectedGameAuditIds].filter((id) => pendingIDs.has(String(id))));
}

function updateGameAuditSelectionUI() {
  const selectAll = $("#game-audit-select-all");
  if (!selectAll) return;
  const boxes = [...document.querySelectorAll("#games-table .game-audit-checkbox")];
  const checkedCount = boxes.filter((box) => box.checked).length;
  selectAll.checked = boxes.length > 0 && checkedCount === boxes.length;
  selectAll.indeterminate = checkedCount > 0 && checkedCount < boxes.length;
  selectAll.disabled = boxes.length === 0;
  const summary = $("#game-audit-summary");
  if (summary) {
    summary.textContent = boxes.length
      ? `本页待审核 ${boxes.length} 个，已勾选 ${state.selectedGameAuditIds.size} 个`
      : "当前页没有待审核局";
  }
  const button = $("#game-batch-audit-button");
  if (button) {
    button.textContent = state.selectedGameAuditIds.size ? `批量通过已选 ${state.selectedGameAuditIds.size} 个` : "批量通过已勾选";
    button.disabled = state.selectedGameAuditIds.size === 0;
  }
}

async function loadGames(formData) {
  if (formData) resetPagination("games");
  const data = await apiGet(`/api/admin/games${querySuffix(formData)}`);
  state.games = data.items || [];
  pruneGameAuditSelection();
  renderPaginatedTable("#games-table", state.games, "games", gameRow, 8, "暂无组局");
  updateGameAuditSelectionUI();
  await loadGameApplications();
}

function ensureGameApplicationsPanel() {
  if ($("#game-applications-panel")) return;
  $("#view-root").insertAdjacentHTML("beforeend", `
    <section id="game-applications-panel" class="panel">
      <div class="panel-head">
        <div>
          <h2>入局申请</h2>
        </div>
        <form id="game-application-filter-form" class="filters">
          <select name="status">
            <option value="">全部状态</option>
            <option value="pending">待处理</option>
            <option value="approved">已通过</option>
            <option value="rejected">已驳回</option>
            <option value="cancelled">已取消</option>
          </select>
          <input name="gameId" inputmode="numeric" placeholder="组局" />
          <input name="userId" inputmode="numeric" placeholder="用户" />
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
  if (formData) resetPagination("gameApplications");
  const data = can("game:read")
    ? await apiGet(`/api/admin/game-applications${querySuffix(formData)}`)
    : { items: [] };
  state.gameApplications = data.items || [];
  renderGameApplications();
}

function renderGameApplications() {
  const table = $("#game-applications-table");
  if (!table) return;
  renderPaginatedTable("#game-applications-table", state.gameApplications, "gameApplications", gameApplicationRow, 7, "暂无入局申请");
}

function gameApplicationTable(bodyID) {
  return `
    <table>
      <thead>
        <tr>
          <th>申请</th>
          <th>组局</th>
          <th>申请人</th>
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
    ? fileIDs.map((id, index) => `<button class="ghost" data-action="download-application-file" data-id="${escapeHTML(id)}" type="button">查看材料 ${index + 1}</button>`).join(" ")
    : "-";
  return `
    <tr>
      <td>${escapeHTML(item.id ? `申请 ${item.id}` : "-")}</td>
      <td>${escapeHTML(item.gameId ? `局 ${item.gameId}` : "-")}</td>
      <td>${escapeHTML(userText(item.userId))}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${escapeHTML(item.reason || "-")}</td>
      <td>${fileButtons}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

async function createAdminGame(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const data = Object.fromEntries(new FormData(form).entries());
  const creatorUserId = Number(data.creatorUserId);
  if (!Number.isInteger(creatorUserId) || creatorUserId <= 0) {
    toast("请选择发起人", true);
    return;
  }
  const payload = {
    title: data.title.trim(),
    creatorUserId,
    gameType: data.gameType,
    coverImage: String(data.coverImage || "").trim(),
    cityCode: data.cityCode.trim(),
    cityName: data.cityName.trim(),
    address: data.address.trim(),
    signupStartAt: String(data.signupStartAt || "").replace("T", " ").trim(),
    signupEndAt: String(data.signupEndAt || "").replace("T", " ").trim(),
    startAt: String(data.startAt || "").replace("T", " ").trim(),
    endAt: String(data.endAt || "").replace("T", " ").trim(),
    minPlayers: Number(data.minPlayers),
    maxPlayers: Number(data.maxPlayers),
  };
  const rawLongitude = String(data.longitude || "").trim();
  const rawLatitude = String(data.latitude || "").trim();
  if (rawLongitude || rawLatitude) {
    const longitude = Number(rawLongitude);
    const latitude = Number(rawLatitude);
    if (!Number.isFinite(longitude) || !Number.isFinite(latitude)) {
      toast("请选择有效的地点坐标", true);
      return;
    }
    payload.longitude = longitude;
    payload.latitude = latitude;
  }
  if (payload.minPlayers < 5 || payload.maxPlayers > 8 || payload.minPlayers > payload.maxPlayers) {
    toast("每局人数必须为 5-8 人，满 5 人后可手动开始，最多 8 人", true);
    return;
  }
  const signupStartTimestamp = Date.parse(String(data.signupStartAt || ""));
  const signupEndTimestamp = Date.parse(String(data.signupEndAt || ""));
  const startTimestamp = Date.parse(String(data.startAt || ""));
  const endTimestamp = Date.parse(String(data.endAt || ""));
  const timestamps = [signupStartTimestamp, signupEndTimestamp, startTimestamp, endTimestamp];
  if (
    !payload.signupStartAt ||
    !payload.signupEndAt ||
    !payload.startAt ||
    !payload.endAt ||
    timestamps.some((timestamp) => !Number.isFinite(timestamp)) ||
    signupEndTimestamp <= signupStartTimestamp ||
    signupEndTimestamp > startTimestamp ||
    endTimestamp <= startTimestamp
  ) {
    toast("请填写完整时间，且须满足：报名开始 < 报名截止 ≤ 组局开始 < 组局结束", true);
    return;
  }
  try {
    const game = await apiPost("/api/admin/games", payload);
    toast(`已创建 ${gameTypeLabel(game.gameType)}，编号 ${game.id}`);
    form.reset();
    setFormValue(form, "minPlayers", "5");
    setFormValue(form, "maxPlayers", "8");
    resetGameCoverUpload();
    syncGameCreatorHint();
    closeAdminDrawer();
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
    if (button.dataset.action === "reject-audit") {
      const reason = window.prompt("请输入驳回原因");
      if (reason == null || !String(reason).trim()) {
        toast("驳回必须填写原因", true);
        return;
      }
      await apiPost(`/api/admin/games/${id}/audit`, { approve: false, remark: String(reason).trim() });
      toast(`局 ${id} 已驳回`);
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
  const panel = openAdminDrawer({
    title: data.game.id ? `局 ${data.game.id}详情` : "组局详情",
    subtitle: data.game.title || "",
    body: `
    <div class="row-actions">
      <span class="${badgeClass(data.game.status)}">${statusLabel(data.game.status)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("局类型", gameTypeLabel(data.game.gameType))}
      ${detailCell("发起方式", sourceLabel(data.game.gameSource))}
      ${detailCell("发起人", userText(data.game.creatorUserId))}
      ${detailCell("封面", data.game.coverImage ? "已配置" : "未配置")}
      ${detailCell("主领路人", userText(data.game.mainGuideUserId))}
      ${detailCell("当前成员", memberListText(data.memberIds))}
      ${detailCell("进度节点", `${(data.milestones || []).length} 个`)}
      ${detailCell("打卡记录", `${(data.checkins || []).length} 条`)}
      ${detailCell("复盘记录", `${(data.retrospectives || []).length} 条`)}
      ${detailCell("再开一局草稿", `${(data.continueDrafts || []).length} 条`)}
    </div>
    <div class="table-wrap">
      <h3>本局入局申请</h3>
      ${gameApplicationTable("game-detail-applications-table")}
    </div>
    ${gameOpsBlock(data.game.id, data)}
  `,
  });
  renderGameDetailApplications(data.game.id);
  bindGameOpsPanel(panel, data.game.id);
}

function renderGameDetailApplications(gameID) {
  const table = $("#game-detail-applications-table");
  if (!table) return;
  const items = state.gameApplications.filter((item) => String(item.gameId) === String(gameID));
  renderPaginatedTable("#game-detail-applications-table", items, `gameDetailApplications-${gameID}`, gameApplicationRow, 7, "暂无入局申请", 6);
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
    const imRoomButton = event.target.closest("button[data-action='ensure-game-im-room']");
    if (imRoomButton) {
      if (!can("game:update_status")) {
        toast("当前角色没有创建 IM 房间权限", true);
        return;
      }
      imRoomButton.disabled = true;
      try {
        const result = await apiPost(`/api/admin/games/${gameID}/im-room`, {});
        const room = result.room || {};
        const actionText = result.created ? "已创建" : "已同步";
        toast(`IM 房间${actionText}：${room.id || "-"}`);
      } catch (error) {
        toast(error.message, true);
      } finally {
        imRoomButton.disabled = false;
      }
      return;
    }
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
        await apiPost(`/api/admin/game-checkins/${button.dataset.id}/mark-invalid`, { reason: "后台标记打卡异常" });
    toast("打卡已标记异常");
    await showGameDetail(gameID);
  });
}

async function renderAudits() {
  bindSectionTabs($("#view-root"));
  $("#role-grant-form")?.addEventListener("submit", grantRoleFromAdmin);
  const tasks = [];
  if (can("identity:read")) {
    tasks.push(loadIdentities());
    tasks.push(loadEnterpriseCertifications());
    tasks.push(loadAvatarAudits());
  } else {
    renderNoAccess("#identity-list", "缺少 identity:read");
    renderNoAccess("#enterprise-certification-list", "缺少 identity:read");
    renderNoAccess("#avatar-audit-list", "缺少 identity:read");
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
      if (["reload-identities", "identity-detail", "reload-enterprise-certifications", "enterprise-certification-detail", "reload-avatar-audits", "avatar-detail"].includes(button.dataset.action) && !can("identity:read")) {
        toast("缺少 identity:read", true);
        return;
      }
      if (["identity-approve", "identity-reject", "enterprise-certification-approve", "enterprise-certification-reject", "avatar-approve", "avatar-reject"].includes(button.dataset.action) && !can("identity:update")) {
        toast("缺少 identity:update", true);
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
      if (button.dataset.action === "reload-enterprise-certifications") await loadEnterpriseCertifications();
      if (button.dataset.action === "reload-avatar-audits") await loadAvatarAudits();
      if (button.dataset.action === "reload-roles") await loadRoles();
      if (button.dataset.action === "identity-detail") await showIdentityDetail(Number(button.dataset.userId));
      if (button.dataset.action === "enterprise-certification-detail") await showEnterpriseCertificationDetail(Number(button.dataset.userId));
      if (button.dataset.action === "identity-approve") {
        await reviewIdentityVerification(button.dataset.userId, true, "后台审核通过");
        toast("实名认证已通过");
        await loadIdentities();
      }
      if (button.dataset.action === "identity-reject") {
        const reason = askRejectReason("请输入实名认证驳回原因");
        if (!reason) return;
        await reviewIdentityVerification(button.dataset.userId, false, reason);
        toast("实名认证已驳回");
        await loadIdentities();
      }
      if (button.dataset.action === "enterprise-certification-approve") {
        await reviewEnterpriseCertification(button.dataset.userId, true, "后台审核通过");
        toast("企业认证已通过");
        await loadEnterpriseCertifications();
      }
      if (button.dataset.action === "enterprise-certification-reject") {
        const reason = askRejectReason("请输入企业认证驳回原因");
        if (!reason) return;
        await reviewEnterpriseCertification(button.dataset.userId, false, reason);
        toast("企业认证已驳回");
        await loadEnterpriseCertifications();
      }
      if (button.dataset.action === "avatar-detail") showAvatarAuditDetail(Number(button.dataset.userId));
      if (button.dataset.action === "avatar-approve") {
        await reviewAvatarAudit(button.dataset.userId, true, "后台审核通过");
        toast("头像已通过");
        await loadAvatarAudits();
      }
      if (button.dataset.action === "avatar-reject") {
        const reason = askRejectReason("请输入头像驳回原因");
        if (!reason) return;
        await reviewAvatarAudit(button.dataset.userId, false, reason);
        toast("头像已驳回");
        await loadAvatarAudits();
      }
      if (button.dataset.action === "role-detail") showRoleApplicationDetail(Number(button.dataset.id));
      if (button.dataset.action === "role-approve") {
        await reviewRoleApplication(button.dataset.id, true, "后台审核通过");
        toast("角色申请已通过");
        await loadRoles();
      }
      if (button.dataset.action === "role-reject") {
        const reason = askRejectReason("请输入角色申请驳回原因");
        if (!reason) return;
        await reviewRoleApplication(button.dataset.id, false, reason);
        toast("角色申请已驳回");
        await loadRoles();
      }
    } catch (error) {
      toast(error.message, true);
    }
  });
}

async function grantRoleFromAdmin(event) {
  event.preventDefault();
  if (!can("role:update")) {
    toast("缺少 role:update", true);
    return;
  }
  const form = event.currentTarget;
  const data = Object.fromEntries(new FormData(form).entries());
  const userIds = String(data.userIds || "").split(/[\s,，、]+/).map((value) => Number(value)).filter((value) => Number.isInteger(value) && value > 0);
  if (!userIds.length) {
    toast("请填写有效用户 ID", true);
    return;
  }
  try {
    const result = await apiPost("/api/admin/roles/grant", { userIds, roleCode: data.roleCode, reason: data.reason || "一期白名单开通" });
    const success = Number(result.success || 0);
    const alreadyActive = Number(result.alreadyActive || 0);
    const failed = Number(result.failed || 0);
    const failedItems = (result.items || []).filter((item) => item.status === "failed");
    const failureText = failedItems.map((item) => `${item.userId}：${item.reason || "开通失败"}`).join("；");
    const receipt = `批次 ${result.batchId || "-"}：成功 ${success}，已开通 ${alreadyActive}，失败 ${failed}`;
    const receiptNode = $("#role-grant-result");
    if (receiptNode) receiptNode.textContent = failureText ? `${receipt}。${failureText}` : receipt;
    toast(failed ? `${receipt}，请查看结果回执` : receipt);
    form.reset();
    await loadRoles();
  } catch (error) {
    toast(error.message, true);
  }
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
  renderPaginatedList("#identity-list", items, "identities", (item) => stackItem({
    title: userText(item.userId),
    badge: item.status,
    meta: [
      `手机号：${identityPhoneForAdmin(item)}`,
      `认证姓名：${identityNameForAdmin(item)}`,
      `身份证：${identityIDCardForAdmin(item)}`,
      `最近更新：${formatTime(item.updatedAt || item.createdAt)}`,
    ],
    action: identityActions(item),
  }), "暂无认证记录");
}

async function loadAvatarAudits() {
  if (!can("identity:read")) {
    renderNoAccess("#avatar-audit-list", "缺少 identity:read");
    return;
  }
  const data = await apiGet("/api/admin/avatar-audits");
  const items = data.items || [];
  state.avatarAudits = items;
  renderPaginatedList("#avatar-audit-list", items, "avatarAudits", (item) => stackItem({
    title: userText(item.userId),
    badge: item.status,
    meta: [
      `昵称：${item.nickname || item.userName || userText(item.userId)}`,
      `当前头像：${item.currentAvatarUrl ? "已设置" : "-"}`,
      `待审头像：${item.pendingAvatarUrl ? "待审核" : "-"}`,
      `审核说明：${item.statusText || "-"}`,
    ],
    action: avatarAuditActions(item),
  }), "暂无头像审核记录");
}

async function loadEnterpriseCertifications() {
  if (!can("identity:read")) {
    renderNoAccess("#enterprise-certification-list", "缺少 identity:read");
    return;
  }
  const data = await apiGet("/api/admin/enterprise-certifications");
  const items = data.items || [];
  state.enterpriseCertifications = items;
  renderPaginatedList("#enterprise-certification-list", items, "enterpriseCertifications", (item) => stackItem({
    title: userText(item.userId),
    badge: item.status,
    meta: [
      `企业：${item.companyName || "-"}`,
      `统一社会信用代码：${item.unifiedSocialCreditCode || "-"}`,
      `法定代表人：${item.legalPerson || "-"}`,
      `提交时间：${formatTime(item.createdAt)}`,
    ],
    action: enterpriseCertificationActions(item),
  }), "暂无企业认证记录");
}

function enterpriseCertificationActions(item) {
  const actions = [`<button class="ghost" data-action="enterprise-certification-detail" data-user-id="${escapeHTML(item.userId)}" type="button">详情</button>`];
  if (item.status === "pending" && can("identity:update")) {
    actions.push(`<button class="ghost" data-action="enterprise-certification-approve" data-user-id="${escapeHTML(item.userId)}" type="button">通过</button>`);
    actions.push(`<button class="ghost" data-action="enterprise-certification-reject" data-user-id="${escapeHTML(item.userId)}" type="button">驳回</button>`);
  }
  return actions.join("");
}

async function reviewEnterpriseCertification(userID, approve, remark) {
  return apiPost(`/api/admin/enterprise-certifications/${userID}/review`, { approve, remark });
}

async function showEnterpriseCertificationDetail(userID) {
  if (!userID) return;
  const item = await apiGet(`/api/admin/enterprise-certifications/${userID}`);
  openAdminDrawer({
    title: "企业认证详情",
    subtitle: userText(item.userId),
    body: `<div class="row-actions"><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></div><div class="detail-grid">
      ${detailCell("用户", userText(item.userId))}
      ${detailCell("企业名称", item.companyName || "-")}
      ${detailCell("统一社会信用代码", item.unifiedSocialCreditCode || "-")}
      ${detailCell("法定代表人", item.legalPerson || "-")}
      ${detailCell("营业执照材料", item.businessLicenseFileId ? `<button class="ghost" data-action="download-application-file" data-id="${escapeHTML(item.businessLicenseFileId)}" type="button">查看材料</button>` : "-")}
      ${detailCell("对公账户材料", item.publicAccountFileId ? `<button class="ghost" data-action="download-application-file" data-id="${escapeHTML(item.publicAccountFileId)}" type="button">查看材料</button>` : "-")}
      ${detailCell("审核备注", item.reviewRemark || item.rejectReason || "-")}
      ${detailCell("提交时间", formatTime(item.createdAt))}
    </div>`,
  });
}

function identityActions(item) {
  const actions = [`<button class="ghost" data-action="identity-detail" data-user-id="${item.userId}" type="button">详情</button>`];
  if (identityCanReview(item)) {
    actions.push(`<button class="ghost" data-action="identity-approve" data-user-id="${item.userId}" type="button">通过</button>`);
    actions.push(`<button class="ghost" data-action="identity-reject" data-user-id="${item.userId}" type="button">驳回</button>`);
  }
  return actions.join("");
}

function identityCanReview(item) {
  return item && item.status === "pending" && can("identity:update") && Boolean(item.realNameMasked || item.idCardMasked);
}

function avatarAuditActions(item) {
  const actions = [`<button class="ghost" data-action="avatar-detail" data-user-id="${escapeHTML(item.userId)}" type="button">详情</button>`];
  if (item.status === "pending" && can("identity:update")) {
    actions.push(`<button class="ghost" data-action="avatar-approve" data-user-id="${escapeHTML(item.userId)}" type="button">通过</button>`);
    actions.push(`<button class="ghost" data-action="avatar-reject" data-user-id="${escapeHTML(item.userId)}" type="button">驳回</button>`);
  }
  return actions.join("");
}

async function reviewAvatarAudit(userID, approve, reason) {
  return apiPost(`/api/admin/avatar-audits/${userID}/review`, { approve, reason });
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
  renderPaginatedList("#role-list", items, "roleApplications", (item) => stackItem({
    title: `${roleLabel(item.roleCode)}申请`,
    badge: item.status,
    meta: [`申请人：${userText(item.userId)}`, `申请说明：${item.reason || "-"}`, `提交时间：${formatTime(item.createdAt)}`],
    action: roleApplicationActions(item),
  }), "暂无角色申请");
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

async function reviewIdentityVerification(userID, approve, reason) {
  return apiPost(`/api/admin/identity-verifications/${userID}/review`, { approve, reason });
}

function identityNameForAdmin(item) {
  return (can("identity:sensitive:read") && item?.realNameFull) || item?.realNameMasked || item?.realname || "-";
}

function identityPhoneForAdmin(item) {
  if (can("identity:sensitive:read") && item?.phoneFull) return item.phoneFull;
  if (can("identity:sensitive:read") && item?.phone) return item.phone;
  if (item?.phoneMasked) return item.phoneMasked;
  return "-";
}

function identityIDCardForAdmin(item) {
  if (can("identity:sensitive:read") && item?.idCardFull) return item.idCardFull;
  if (item?.idCardMasked) return `${item.idCardMasked}（旧记录仅保留脱敏信息）`;
  return "-";
}

async function showIdentityDetail(userID) {
  if (!userID) return;
  const item = await apiGet(`/api/admin/identity-verifications/${userID}`);
  const method = item.faceVerified ? "人脸核身" : (item.realNameMasked || item.idCardMasked ? "后台人工审核" : "短信/登录记录");
  openAdminDrawer({
    title: "认证记录详情",
    subtitle: userText(item.userId),
    body: `
    <div class="row-actions">
      <span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("用户", userText(item.userId))}
      ${detailCell("认证方式", method)}
      ${detailCell("手机号", identityPhoneForAdmin(item))}
      ${detailCell("姓名", identityNameForAdmin(item))}
      ${detailCell("身份证号", identityIDCardForAdmin(item))}
      ${detailCell("认证状态", statusLabel(item.status))}
      ${detailCell("失败/驳回原因", item.failureReason || "-")}
      ${detailCell("创建时间", formatTime(item.createdAt))}
      ${detailCell("最近更新", formatTime(item.updatedAt))}
    </div>
  `,
  });
}

function avatarPreviewCell(label, url, fallback = "-") {
  const text = url || fallback;
  const image = url ? `<img class="avatar-audit-image" src="${escapeHTML(url)}" alt="${escapeHTML(label)}" />` : "";
  return `
    <div class="detail-cell">
      <span>${escapeHTML(label)}</span>
      ${image}
      <strong>${escapeHTML(text)}</strong>
    </div>
  `;
}

function showAvatarAuditDetail(userID) {
  const item = state.avatarAudits.find((audit) => Number(audit.userId) === Number(userID));
  if (!item) return;
  const panel = openAdminDrawer({
    title: "头像审核详情",
    subtitle: userText(item.userId),
    body: `
    <div class="row-actions">
      <span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("用户", userText(item.userId))}
      ${detailCell("昵称", item.nickname || item.userName || "-")}
      ${detailCell("审核状态", statusLabel(item.status))}
      ${detailCell("审核说明", item.statusText || item.avatarAuditReason || "-")}
      ${avatarPreviewCell("当前头像", item.currentAvatarUrl)}
      ${avatarPreviewCell("待审头像", item.pendingAvatarUrl)}
    </div>
    <div class="row-actions">
      ${item.status === "pending" && can("identity:update") ? `
        <button class="primary" data-action="avatar-approve" data-user-id="${escapeHTML(item.userId)}" type="button">通过头像</button>
        <button class="ghost danger" data-action="avatar-reject" data-user-id="${escapeHTML(item.userId)}" type="button">驳回头像</button>
      ` : ""}
    </div>
  `,
  });
  panel.querySelectorAll("button[data-action='avatar-approve'], button[data-action='avatar-reject']").forEach((button) => {
    button.addEventListener("click", async () => {
      const approve = button.dataset.action === "avatar-approve";
      try {
        const reason = approve ? "后台审核通过" : askRejectReason("请输入头像驳回原因");
        if (!reason) return;
        await reviewAvatarAudit(button.dataset.userId, approve, reason);
        toast(approve ? "头像已通过" : "头像已驳回");
        closeAdminDrawer();
        await loadAvatarAudits();
      } catch (error) {
        toast(error.message, true);
      }
    });
  });
}

function showRoleApplicationDetail(applicationID) {
  const item = state.roleApplications.find((app) => Number(app.id) === Number(applicationID));
  if (!item) return;
  const inviteRelation = item.inviteRelation || {};
  const inviter = item.inviter || {};
  const inviteCode = item.inviteCode || inviteRelation.inviteCode || (inviteRelation.inviteCodeId ? `邀请码 ${inviteRelation.inviteCodeId}` : "-");
  const snapshot = item.eligibilitySnapshot || {};
  const snapshotItems = Array.isArray(snapshot.requirements) ? snapshot.requirements : [];
  const snapshotText = snapshotItems.length
    ? snapshotItems.map((entry) => {
      const current = entry.current ?? "-";
      const required = entry.required ?? "-";
      const progress = entry.current !== undefined || entry.required !== undefined ? `（${current}/${required}）` : "";
      return `${entry.title || entry.key}: ${entry.met ? "已满足" : "未满足"}${progress}`;
    }).join("；")
    : "未记录（历史申请）";
  const snapshotSummary = snapshotItems.length
    ? `${snapshot.eligible ? "可提交" : "条件未全满足"}；基础条件${snapshot.baseEligible ? "已满足" : "未满足"}${snapshot.capturedAt ? `；采集时间：${formatTime(snapshot.capturedAt)}` : ""}`
    : "历史申请未保存资格快照";
  openAdminDrawer({
    title: "角色申请详情",
    subtitle: `${roleLabel(item.roleCode)}申请`,
    body: `
    <div class="row-actions">
      <span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("申请编号", item.id)}
      ${detailCell("申请人", userText(item.userId))}
      ${detailCell("申请角色", roleLabel(item.roleCode))}
      ${detailCell("邀请人", inviterLabel(inviter, inviteRelation))}
      ${detailCell("使用邀请码", inviteCode)}
      ${detailCell("邀请绑定来源", inviteRelation.bindSource || "-")}
      ${detailCell("处理状态", statusLabel(item.status))}
      ${detailCell("申请说明", item.reason || "-")}
      ${detailCell("能力说明", item.abilityDescription || "-")}
      ${detailCell("提交时资格快照", `${snapshotSummary}；${snapshotText}`)}
      ${detailCell("证明材料", (item.proofFileIds || []).length ? `${item.proofFileIds.length} 份材料` : "-")}
      ${detailCell("审核人", adminUserText(item.reviewAdminId))}
      ${detailCell("审核备注", item.reviewRemark || item.rejectReason || "-")}
      ${detailCell("提交时间", formatTime(item.createdAt))}
      ${detailCell("最近更新", formatTime(item.updatedAt))}
    </div>
  `,
  });
}

async function renderRevenue() {
  bindSectionTabs($("#view-root"));
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
    toast("当前账号没有维护分润方案权限", true);
    return;
  }
  const data = Object.fromEntries(new FormData(event.currentTarget).entries());
  const payload = {
    name: data.name.trim(),
    gameType: data.gameType,
    platformBps: percentToBps(data.platformBps),
    creatorBps: percentToBps(data.creatorBps),
    memberBps: percentToBps(data.memberBps),
  };
  const totalBps = payload.platformBps + payload.creatorBps + payload.memberBps;
  if (totalBps < 0 || totalBps > 10000) {
    toast("分成比例合计不能超过 100%", true);
    return;
  }
  try {
    await apiPost("/api/admin/revenue/templates", payload);
    toast("分润方案已保存");
    await loadRevenueTemplates();
  } catch (error) {
    toast(error.message, true);
  }
}

async function upsertRevenueRule(event) {
  event.preventDefault();
  if (!can("revenue:template:update")) {
    toast("当前账号没有维护分润规则权限", true);
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
    renderNoAccess("#revenue-template-list", "当前账号没有查看分润方案权限");
    return;
  }
  const data = await apiGet("/api/admin/revenue/templates");
  state.revenueTemplates = data.items || [];
  renderPaginatedList("#revenue-template-list", state.revenueTemplates, "revenueTemplates", (item) => stackItem({
    title: item.name || revenueTemplateName(item.id),
    badge: statusLabel(item.status),
    meta: [
      `局类型：${gameTypeLabel(item.gameType)}`,
      `平台分成：${bpsText(item.platformBps)}`,
      `发起人分成：${bpsText(item.creatorBps)}`,
      `参与人分成：${bpsText(item.memberBps)}`,
    ],
    action: `<button class="ghost" data-action="load-revenue-rules" data-template-id="${item.id}" type="button">查看规则</button>`,
  }), "暂无分润模板", 6);
  $("#revenue-template-list").addEventListener("click", async (event) => {
    const button = event.target.closest("button[data-action='load-revenue-rules']");
    if (!button) return;
    selectRevenueTemplate(button.dataset.templateId);
    await loadRevenueRules(Number(button.dataset.templateId));
  });
  if (state.revenueTemplates[0] && !getFormValue($("#revenue-rule-form"), "templateId")) {
    selectRevenueTemplate(state.revenueTemplates[0].id);
  }
}

async function loadRevenueRules(templateId = numberOrZero(getFormValue($("#revenue-rule-form"), "templateId"))) {
  if (!can("revenue:template:view")) {
    renderNoAccess("#revenue-rule-list", "当前账号没有查看分润规则权限");
    return;
  }
  const data = await apiGet(`/api/admin/revenue/rules${templateId ? `?templateId=${templateId}` : ""}`);
  renderPaginatedList("#revenue-rule-list", data.items || [], "revenueRules", (item) => stackItem({
    title: revenueRuleLabel(item.ruleCode),
    badge: "active",
    meta: [`所属方案：${revenueTemplateName(item.templateId)}`, `规则内容：${adminDisplayValue(item.ruleValue)}`, `创建时间：${formatTime(item.createdAt)}`],
  }), "暂无分润规则", 6);
}

async function onRevenueCalcClick(event) {
  const button = event.target.closest("button[data-action]");
  if (!button) return;
  const payload = revenueCalcPayload();
  try {
    if (button.dataset.action === "revenue-preview") {
      if (!can("revenue:simulate")) {
        toast("当前账号没有分润试算权限", true);
        return;
      }
      renderRevenuePreview(await apiPost("/api/admin/revenue/preview", payload));
    }
    if (button.dataset.action === "revenue-generate") {
      if (!can("revenue:generate")) {
        toast("当前账号没有生成分润记录权限", true);
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
      ${detailCell("关联组局", preview.gameId ? `局 ${preview.gameId}` : "-")}
      ${detailCell("分润方案", revenueTemplateName(preview.templateId))}
      ${detailCell("结算金额", yuanText(preview.amountCent))}
      ${detailCell("阻止生成原因", compactList((preview.blockReasons || []).map(revenueBlockReasonLabel)) || "-")}
    </div>
    <div class="table-wrap nested-table">
      <table><thead><tr><th>角色</th><th>用户</th><th>金额</th></tr></thead><tbody>
        ${(preview.items || []).map((item) => `<tr><td>${escapeHTML(roleLabel(item.role))}</td><td>${escapeHTML(userText(item.userId))}</td><td>${escapeHTML(yuanText(item.amountCent))}</td></tr>`).join("")}
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
  renderPaginatedTable("#revenue-records-table", state.revenueRecords, "revenueRecords", revenueRecordRow, 8, "暂无分润记录");
}

function showRevenueRecordDetail(recordID) {
  const item = state.revenueRecords.find((record) => Number(record.id) === Number(recordID));
  if (!item) return;
  const settlements = state.revenueSettlements.filter((settlement) => Number(settlement.recordId) === Number(recordID));
  openAdminDrawer({
    title: "分润记录详情",
    subtitle: item.recordNo || (item.id ? `分润记录 ${item.id}` : ""),
    body: `
    <div class="row-actions">
      <span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("记录", item.id ? `分润记录 ${item.id}` : "-")}
      ${detailCell("记录号", item.recordNo)}
      ${detailCell("关联组局", item.gameId ? `局 ${item.gameId}` : "-")}
      ${detailCell("分润方案", revenueTemplateName(item.templateId))}
      ${detailCell("结算金额", yuanText(item.amountCent))}
      ${detailCell("冻结原因", item.frozenReason || "-")}
      ${detailCell("结算时间", item.settledAt ? formatTime(item.settledAt) : "-")}
      ${detailCell("创建时间", formatTime(item.createdAt))}
      ${detailCell("分润明细", `${(item.items || []).length} 条`)}
      ${detailCell("结算记录", `${settlements.length} 条`)}
    </div>
    <div class="table-wrap nested-table">
      <table><thead><tr><th>角色</th><th>用户</th><th>金额</th></tr></thead><tbody>
        ${(item.items || []).map((child) => `<tr><td>${escapeHTML(roleLabel(child.role))}</td><td>${escapeHTML(userText(child.userId))}</td><td>${escapeHTML(yuanText(child.amountCent))}</td></tr>`).join("") || emptyRow(3, "暂无分润明细")}
      </tbody></table>
    </div>
  `,
  });
}

async function loadRevenueSettlements() {
  if (!can("settlement:offline:create")) {
    $("#revenue-settlements-table").innerHTML = emptyRow(6, "无权限查看线下结算记录");
    return;
  }
  const data = await apiGet("/api/admin/revenue/settlements");
  state.revenueSettlements = data.items || [];
  renderPaginatedTable("#revenue-settlements-table", state.revenueSettlements, "revenueSettlements", (item) => `
    <tr>
      <td>${escapeHTML(item.id ? `结算 ${item.id}` : "-")}</td>
      <td>${escapeHTML(item.recordId ? `分润记录 ${item.recordId}` : "-")}</td>
      <td>${escapeHTML(settlementMethodLabel(item.method))}</td>
      <td>${escapeHTML(item.proofNo || "-")}</td>
      <td>${escapeHTML(yuanText(item.amountCent))}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `, 6, "暂无结算记录");
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
        toast("当前账号没有冻结分润记录权限", true);
        return;
      }
      await apiPost(`/api/admin/revenue/records/${button.dataset.id}/freeze`, { reason: "admin_freeze" });
      toast("分润记录已冻结");
    }
    if (button.dataset.action === "revenue-settle") {
      if (!can("settlement:offline:create")) {
        toast("当前账号没有登记线下结算权限", true);
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
  bindSectionTabs($("#view-root"));
  $("#redemption-item-form").addEventListener("submit", createRedemptionItem);
  $("#redemption-items-table").addEventListener("click", onRedemptionItemClick);
  $("#redemption-orders-table").addEventListener("click", onRedemptionOrderClick);
  $("#redemption-orders-refresh").addEventListener("click", loadRedemptionOrders);
  $("#points-logs-refresh").addEventListener("click", loadPointsLogs);
  const tasks = [];
  if (can("redemption:manage")) {
    tasks.push(loadRedemptionItems(), loadRedemptionOrders());
  } else {
    $("#redemption-items-table").innerHTML = emptyRow(8, "无权限管理积分兑换商品");
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
    toast("当前账号没有管理积分兑换权限", true);
    return;
  }
  const form = event.currentTarget;
  const data = Object.fromEntries(new FormData(form).entries());
  const payload = {
    name: data.name.trim(),
    description: data.description.trim(),
    imageUrl: data.imageUrl.trim(),
    pointsCost: Number(data.pointsCost),
    stock: Number(data.stock),
  };
  try {
    await apiPost("/api/admin/redemption/items", payload);
    toast("兑换商品已新增");
    form.reset();
    setFormValue(form, "pointsCost", "10");
    setFormValue(form, "stock", "1");
    await loadRedemptionItems();
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadRedemptionItems() {
  if (!can("redemption:manage")) {
    $("#redemption-items-table").innerHTML = emptyRow(8, "无权限管理积分兑换商品");
    return;
  }
  const data = await apiGet("/api/admin/redemption/items");
  state.redemptionItems = data.items || [];
  renderPaginatedTable("#redemption-items-table", state.redemptionItems, "redemptionItems", redemptionItemRow, 8, "暂无兑换商品", 6);
}

async function loadRedemptionOrders() {
  if (!can("redemption:manage")) {
    $("#redemption-orders-table").innerHTML = emptyRow(8, "无权限管理积分兑换订单");
    return;
  }
  const data = await apiGet("/api/admin/redemption/orders");
  state.redemptionOrders = data.items || [];
  renderPaginatedTable("#redemption-orders-table", state.redemptionOrders, "redemptionOrders", redemptionOrderRow, 8, "暂无兑换订单");
}

async function loadPointsLogs() {
  if (!can("points:read")) {
    $("#points-logs-table").innerHTML = emptyRow(7, "无权限查看积分流水");
    return;
  }
  const data = await apiGet("/api/admin/points/logs");
  state.pointsLogs = data.items || [];
  renderPaginatedTable("#points-logs-table", state.pointsLogs, "pointsLogs", pointsLogRow, 7, "暂无积分流水");
}

async function onRedemptionItemClick(event) {
  const button = event.target.closest("button[data-action]");
  if (!button) return;
  if (!can("redemption:manage")) {
    toast("当前账号没有处理兑换订单权限", true);
    return;
  }
  const row = button.closest("tr");
  const payload = {
    name: row.querySelector("[data-field='name']").value.trim(),
    description: row.querySelector("[data-field='description']").value.trim(),
    imageUrl: row.querySelector("[data-field='imageUrl']").value.trim(),
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
    toast("当前账号没有处理兑换订单权限", true);
    return;
  }
  const approve = button.dataset.action === "redemption-order-approve";
  let payload;
  if (button.dataset.status) {
    payload = { status: button.dataset.status, reason: button.dataset.reason || "后台处理" };
  } else {
    const reason = approve ? "后台审核通过" : askRejectReason("请输入兑换订单驳回原因");
    if (!reason) return;
    payload = { approve, reason };
  }
  try {
    await apiPost(`/api/admin/redemption/orders/${button.dataset.id}/review`, payload);
    toast("兑换订单已处理");
    await Promise.all([loadRedemptionOrders(), loadPointsLogs()]);
  } catch (error) {
    toast(error.message, true);
  }
}

async function renderMembers() {
  bindSectionTabs($("#view-root"));
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
  renderPaginatedTable("#member-teams-table", state.memberTeams, "memberTeams", memberTeamRow, 6, "暂无会员团队");
}

async function loadMemberReports() {
  const data = await apiGet("/api/admin/member-reports");
  state.memberReports = data.items || [];
  renderPaginatedTable("#member-reports-table", state.memberReports, "memberReports", memberReportRow, 7, "暂无会员报表");
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
  openAdminDrawer({
    title: `${team.id || teamID ? `团队 ${team.id || teamID}` : "团队"}详情`,
    subtitle: "成员、层级和收益汇总来自团队后端服务",
    body: `
    <div class="row-actions">
      <span class="${badgeClass(team.status)}">${statusLabel(team.status)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("团队负责人", userText(team.leaderUserId))}
      ${detailCell("团队名称", team.name || "-")}
      ${detailCell("团队状态", statusLabel(team.status))}
      ${detailCell("成员数量", `${revenue.memberCount || members.length} 人`)}
      ${detailCell("累计收益", yuanText(revenue.totalCent))}
      ${detailCell("待结算收益", yuanText(revenue.pendingCent))}
      ${detailCell("已结算收益", yuanText(revenue.settledCent))}
      ${detailCell("创建时间", formatTime(team.createdAt))}
    </div>
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>用户</th>
            <th>层级</th>
            <th>来源</th>
            <th>状态</th>
            <th>加入时间</th>
          </tr>
        </thead>
        <tbody>
          ${members.map(memberTeamMemberRow).join("") || emptyRow(5, "暂无团队成员")}
        </tbody>
      </table>
    </div>
  `,
  });
}

async function renderProfiles() {
  bindSectionTabs($("#view-root"));
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
    await loadExpertProfile(new FormData(event.currentTarget).get("keyword"));
  });
  $("#guide-profile-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!can("profile:read")) {
      toast("当前角色没有画像查看权限", true);
      return;
    }
    await loadGuideProfile(new FormData(event.currentTarget).get("keyword"));
  });
  $("#expert-profile-panel").innerHTML = emptyBlock("输入手机号、昵称、身份证号或用户编号后查询技能画像");
  $("#guide-profile-panel").innerHTML = emptyBlock("输入手机号、昵称、身份证号或用户编号后查询资源画像");
  if (can("connection:read")) {
    await loadConnections();
  } else {
    $("#connections-table").innerHTML = emptyRow(7, "当前角色没有人脉关系查看权限");
  }
}

async function loadConnections() {
  const data = await apiGet("/api/admin/connections");
  state.connections = data.items || [];
  renderPaginatedTable("#connections-table", state.connections, "connections", connectionRow, 7, "暂无人脉关系");
}

async function loadExpertProfile(keyword) {
  const id = await resolveProfileUserID(keyword, "#expert-profile-panel");
  if (!id) return;
  const data = await apiGet(`/api/admin/experts/${encodeURIComponent(id)}/skills`);
  state.expertProfile = data;
  $("#expert-profile-panel").innerHTML = profileDetailBlock("行家技能", data, [
    ["用户", data.userId || id],
    ["技能领域", compactList(data.skillTree)],
    ["服务标签", compactList(data.serviceTags)],
    ["案例材料", compactList(data.caseFileIds)],
    ["资料完整度", `${data.completeness || 0}%`],
    ["更新时间", formatTime(data.updatedAt)],
  ]);
}

async function loadGuideProfile(keyword) {
  const id = await resolveProfileUserID(keyword, "#guide-profile-panel");
  if (!id) return;
  const data = await apiGet(`/api/admin/guides/${encodeURIComponent(id)}/resources`);
  state.guideProfile = data;
  $("#guide-profile-panel").innerHTML = profileDetailBlock("领路人资源", data, [
    ["用户", data.userId || id],
    ["资源标签", compactList(data.resourceTags)],
    ["行业标签", compactList(data.industryTags)],
    ["覆盖城市", compactList(data.cityCodes)],
    ["人脉规模", data.connectionScale || "-"],
    ["资料完整度", `${data.completeness || 0}%`],
    ["更新时间", formatTime(data.updatedAt)],
  ]);
}

async function resolveProfileUserID(keyword, panelSelector) {
  const text = String(keyword || "").trim();
  if (!text) return "";
  const data = await apiGet(`/api/admin/profile-users${querySuffixFromObject({ keyword: text })}`);
  const items = data.items || [];
  if (items.length === 1) {
    return String(items[0].id || "");
  }
  const panel = $(panelSelector);
  if (!panel) return "";
  if (items.length === 0) {
    panel.innerHTML = emptyBlock("未找到匹配用户，请换手机号、昵称、身份证号或用户编号重试");
    return "";
  }
  panel.innerHTML = profileUserCandidatesBlock(items);
  toast("匹配到多个用户，请输入更精确的关键词", true);
  return "";
}

function profileUserCandidatesBlock(items) {
  const rows = items.slice(0, 8).map((item) => {
    const fields = Array.isArray(item.matchFields) && item.matchFields.length ? item.matchFields.join("、") : "-";
    return `
      <tr>
        <td>${escapeHTML(item.id || "-")}</td>
        <td>${escapeHTML(item.nickname || userText(item.userId))}</td>
        <td>${escapeHTML(identityPhoneForAdmin(item))}</td>
        <td>${escapeHTML(identityIDCardForAdmin(item))}</td>
        <td>${escapeHTML(fields)}</td>
      </tr>
    `;
  }).join("");
  return `
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>用户编号</th>
            <th>昵称</th>
            <th>手机号</th>
            <th>身份证号</th>
            <th>匹配字段</th>
          </tr>
        </thead>
        <tbody>${rows || emptyRow(5, "暂无匹配用户")}</tbody>
      </table>
    </div>
  `;
}

async function renderGrowth() {
  bindSectionTabs($("#view-root"));
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
  $("#user-growth-panel").innerHTML = emptyBlock("输入用户编号后查询成长、评价和信用轨迹");
  $("#game-review-trace-panel").innerHTML = emptyBlock("输入组局编号后查询该局评价和信用证据");
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
  const head = $("#reports-table").closest(".panel").querySelector(".panel-head");
  head.insertAdjacentHTML("beforeend", `
    <div class="row-actions">
      <button id="reports-batch-handle-button" class="ghost" type="button">批量处理</button>
      ${can("feedback:view") ? `<button id="feedback-records-button" class="ghost" type="button">用户反馈</button>` : ""}
    </div>
  `);
  $("#reports-batch-handle-button").addEventListener("click", batchHandleReports);
  $("#feedback-records-button")?.addEventListener("click", loadFeedbackRecords);
  $("#reports-table").addEventListener("click", onReportsTableClick);
  await loadReports();
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
    const result = await apiPost("/api/admin/reports/batch-handle", { reportIds: ids, action: "handle", result: "举报属实，已完成平台处理", outcome: "confirmed", rewardPoints: 20, creditDeduct: 10 });
    toast(`批量处理 ${result.success || 0} 条举报，失败 ${result.failed || 0} 条`);
    await loadReports();
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadReports() {
  const data = await apiGet("/api/admin/reports");
  state.reports = data.items || [];
  renderPaginatedTable("#reports-table", state.reports, "reports", reportRow, 9, "暂无举报申诉");
}

async function loadFeedbackRecords() {
  const data = await apiGet("/api/admin/feedback-records");
  state.feedbackRecords = data.items || [];
  const panel = openAdminDrawer({
    title: "用户反馈",
    subtitle: "来自小程序个人中心的反馈记录，可在此回复并同步到用户端",
    body: `
    <div class="row-actions"><span class="badge">${escapeHTML(data.total || state.feedbackRecords.length || 0)} 条</span></div>
    <div class="table-wrap"><table>
      <thead>
        <tr>
          <th>编号</th>
          <th>用户</th>
          <th>类型</th>
          <th>状态</th>
          <th>内容</th>
          <th>时间</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody id="feedback-records-table"></tbody>
    </table></div>
  `,
  });
  renderPaginatedTable("#feedback-records-table", state.feedbackRecords, "feedbackRecords", feedbackRow, 7, "暂无用户反馈");
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
      await apiPost(`/api/admin/reports/${button.dataset.id}/assign`, { handlerAdminId: adminID() || 1 });
      toast("举报已分配处理人");
      await loadReports();
    }
    if (button.dataset.action === "report-handle") {
      await apiPost(`/api/admin/reports/${button.dataset.id}/handle`, { result: "后台已处理" });
      toast("举报已处理");
      await loadReports();
    }
    if (button.dataset.action === "credit-appeal-approve") {
      await apiPost(`/api/admin/reports/${button.dataset.id}/handle`, { result: "信用申诉通过，已补回对应信用分", outcome: "appeal_approved" });
      toast("信用申诉已通过");
      await loadReports();
    }
    if (button.dataset.action === "credit-appeal-reject") {
      const reason = askRejectReason("请输入信用申诉驳回原因");
      if (!reason) return;
      await apiPost(`/api/admin/reports/${button.dataset.id}/handle`, { result: reason, outcome: "appeal_rejected" });
      toast("信用申诉已驳回");
      await loadReports();
    }
    if (button.dataset.action === "report-close") {
      await apiPost(`/api/admin/reports/${button.dataset.id}/close`, { result: "后台关闭" });
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
  const panel = openAdminDrawer({
    title: `${report.id ? `举报 ${report.id}` : "举报"}详情`,
    subtitle: report.content || "",
    body: `
    <div class="row-actions">
      <span class="${badgeClass(report.status)}">${statusLabel(report.status)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("关联组局", report.gameId ? `局 ${report.gameId}` : "-")}
      ${detailCell("举报人", userText(report.reporterUserId))}
      ${detailCell("被举报人", userText(report.targetUserId))}
      ${detailCell("举报类型", reportTypeLabel(report.reportType))}
      ${detailCell("局标题", game.title || "-")}
      ${detailCell("关联评价", evidence.ids?.reviewId ? "有评价证据" : "-")}
      ${detailCell("评价分数", review.score || "-")}
      ${detailCell("收益处理", report.revenueFrozen ? "已冻结" : "未冻结")}
      ${detailCell("聊天证据", `${chatMessages.length} 条`)}
      ${detailCell("附件", evidence.ids?.fileId ? "有附件" : "-")}
      ${detailCell("文件名称", file.fileName || "-")}
      ${detailCell("分润记录", evidence.ids?.revenueRecordId ? "有分润记录" : "-")}
    </div>
    ${review.id ? `
      <div class="detail-block">
        <h3>评价证据</h3>
        <p>评分：${escapeHTML(review.score || "-")} / 评价人：${escapeHTML(userText(review.reviewerUserId))} / 被评价人：${escapeHTML(userText(review.targetUserId))}</p>
        <p>${escapeHTML(review.content || "无评价文字")}</p>
      </div>
    ` : ""}
    ${chatMessages.length ? `
      <div class="detail-block">
        <h3>聊天证据</h3>
        <div class="report-chat-evidence">
          ${chatMessages.map((message) => `
            <div class="report-chat-message">
              <div class="muted">${escapeHTML(message.createdAt || message.created_at || "-")} · ${escapeHTML(userText(message.senderUserId || message.sender_user_id))} · ${escapeHTML(messageTypeLabel(message.messageType || message.message_type || "text"))}</div>
              <div>${escapeHTML(message.content || "")}</div>
            </div>
          `).join("")}
        </div>
      </div>
    ` : ""}
    ${fileAction ? `<div class="actions">${fileAction}</div>` : ""}
  `,
  });
  panel.querySelector("button[data-action='admin-file-download']")?.addEventListener("click", (event) => {
    downloadAdminFile(event.currentTarget.dataset.fileId);
  });
}

async function renderExports() {
  bindSectionTabs($("#view-root"));
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
  renderPaginatedList("#export-template-list", state.exportTemplates, "exportTemplates", (item) => stackItem({
    title: exportTemplateTitle(item),
    badge: item.enabled ? "active" : "disabled",
    meta: [`导出内容：${exportTypeLabel(item.exportType)}`, item.enabled ? "可创建导出任务" : "已停用"],
    action: `<button class="ghost" data-action="select-export-template" data-code="${escapeHTML(item.code)}" type="button">选择</button>`,
  }), "暂无导出模板");
  $("#export-template-list").addEventListener("click", (event) => {
    const button = event.target.closest("button[data-action='select-export-template']");
    if (!button) return;
    const templateCodeInput = $("#export-create-form").querySelector("[name='templateCode']");
    if (templateCodeInput) templateCodeInput.value = button.dataset.code;
    const templateItem = state.exportTemplates.find((item) => item.code === button.dataset.code);
    $("#export-template-display").value = exportTemplateTitle(templateItem || { code: button.dataset.code });
  });
}

async function createExportTask(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const data = Object.fromEntries(new FormData(form).entries());
  try {
    await apiPost("/api/admin/reports/export", { templateCode: data.templateCode.trim(), filters: {} });
    toast("导出任务已创建");
    await loadExportTasks();
  } catch (error) {
    toast(error.message, true);
  }
}

async function runExportTasks() {
  try {
    await apiPost("/api/internal/reports/export-runner", {});
    toast("待处理报表已生成");
    await loadExportTasks();
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadExportTasks(formData) {
  if (formData) resetPagination("exportTasks");
  const data = await apiGet(`/api/admin/export-tasks${querySuffix(formData)}`);
  state.exportTasks = data.items || [];
  renderPaginatedTable("#export-tasks-table", state.exportTasks, "exportTasks", exportTaskRow, 8, "暂无导出任务");
}

function ensureExportFilters() {
  if ($("#export-filter-form")) return;
  $("#export-tasks-table").closest(".panel").querySelector(".panel-head").insertAdjacentHTML("beforeend", `
    <form id="export-filter-form" class="filters">
      <select name="exportType">
        <option value="">全部报表</option>
        <option value="reports">举报申诉报表</option>
        <option value="operation_logs">操作日志报表</option>
        <option value="reviews">评价报表</option>
      </select>
      <select name="status">
        <option value="">全部状态</option>
        <option value="pending">待处理</option>
        <option value="running">执行中</option>
        <option value="done">已完成</option>
        <option value="failed">失败</option>
      </select>
      <button class="ghost" type="submit">筛选</button>
    </form>
  `);
}

async function showExportTaskDetail(taskID) {
  const item = state.exportTasks.find((task) => Number(task.id) === Number(taskID));
  if (!item) return;
  let downloadURL = "";
  if (item.status === "done" && item.fileId) {
    try {
      const result = await apiGet(`/api/admin/export-tasks/${item.id}/download-url`);
      downloadURL = exportDownloadText(result);
    } catch (error) {
      downloadURL = error.message;
    }
  }
  openAdminDrawer({
    title: "导出任务详情",
    subtitle: item.id ? `导出任务 ${item.id}` : "",
    body: `
    <div class="row-actions">
      <span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("任务", item.id ? `导出任务 ${item.id}` : "-")}
      ${detailCell("报表类型", exportTemplateName(item.templateCode))}
      ${detailCell("导出内容", exportTypeLabel(item.exportType))}
      ${detailCell("创建人", adminUserText(item.createdBy))}
      ${detailCell("创建时间", formatTime(item.createdAt))}
      ${detailCell("完成时间", item.finishedAt ? formatTime(item.finishedAt) : "-")}
      ${detailCell("失败原因", item.failReason || "-")}
      ${detailCell("文件状态", downloadURL ? "已生成，可在列表中下载" : "待生成")}
    </div>
  `,
  });
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
      const text = exportDownloadText(url);
      if (navigator.clipboard?.writeText && text && text !== "-") {
        await navigator.clipboard.writeText(text).catch(() => {});
      }
      toast("导出文件已准备");
    }
  } catch (error) {
    toast(error.message, true);
  }
}

async function renderAnalytics() {
  bindSectionTabs($("#view-root"));
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
    $("#ai-snapshot-panel").innerHTML = detailCell("访问状态", "当前角色没有智能数据权限");
    $("#ai-readiness-list").innerHTML = emptyBlock("当前角色没有智能数据权限");
  }
  if (!can("ai:data:export")) {
    $("#ai-im-export-panel").innerHTML = emptyBlock("当前角色没有智能局内消息导出权限");
  }
  await Promise.all(tasks);
}

async function loadFunnel(formData) {
  const data = await apiGet(`/api/admin/analytics/funnel${querySuffix(formData)}`);
  state.funnel = data;
  renderPaginatedTable("#analytics-funnel-table", data.steps || [], "analyticsFunnel", funnelRow, 4, "暂无漏斗数据", 8);
}

async function loadRetention() {
  const data = await apiGet("/api/admin/analytics/retention");
  state.retention = data;
  renderPaginatedTable("#analytics-retention-table", data.buckets || [], "analyticsRetention", retentionRow, 5, "暂无留存数据", 8);
}

async function loadBehaviorEvents(formData) {
  if (formData) resetPagination("behaviorEvents");
  const data = await apiGet(`/api/admin/behavior/events${querySuffix(formData)}`);
  state.behaviorEvents = data.items || [];
  renderPaginatedTable("#behavior-events-table", state.behaviorEvents, "behaviorEvents", behaviorEventRow, 7, "暂无行为事件");
}

async function loadBehaviorLogs() {
  const data = await apiGet("/api/admin/behavior-logs");
  state.behaviorLogs = data.items || [];
  renderPaginatedList("#behavior-log-list", state.behaviorLogs, "behaviorLogs", behaviorLogItem, "暂无行为时间线");
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
    toast("当前角色没有智能样本准备权限", true);
    return;
  }
  try {
    const result = await apiPost("/api/admin/ai-data/acceptance-fixture", {});
    toast(`智能样本已准备：用户 ${result.usersVerified || 0} 个，组局 ${result.gamesCreated || 0} 个`);
    await loadAIDataSnapshot();
    if (can("data:behavior:read")) await loadBehaviorEvents();
  } catch (error) {
    toast(error.message, true);
  }
}

async function exportAIIMData() {
  if (!can("ai:data:export")) {
    toast("当前角色没有智能局内消息导出权限", true);
    return;
  }
  try {
    const result = await apiPost("/api/admin/ai-data/im-export", {});
    const items = result.items || [];
    renderAIIMExport(items);
    toast(`智能局内消息数据已导出：${items.length} 条`);
  } catch (error) {
    $("#ai-im-export-panel").innerHTML = emptyBlock(`导出失败：${error.message}`);
    toast(error.message, true);
  }
}

function renderAIIMExport(items) {
  const sample = items.slice(0, 8);
  $("#ai-im-export-panel").innerHTML = [
    stackItem({
      title: `局内消息导出明细：${items.length} 条`,
      badge: items.length ? "ready" : "empty",
      meta: ["仅展示前 8 条样例，完整文件以导出结果为准"],
    }),
    ...sample.map((item) => stackItem({
      title: `消息 ${item.messageId || item.id || "-"}`,
      badge: item.messageType || item.type || "message",
      meta: [
        `房间编号：${item.roomId || "-"}`,
        `组局：${item.gameId ? `局 ${item.gameId}` : "-"}`,
        `发送人：${userText(item.senderUserId || item.senderId)}`,
        `创建时间：${item.createdAt || "-"}`,
        `内容：${String(item.content || "").slice(0, 80) || "-"}`,
      ],
    })),
  ].join("");
}

function renderAIDataSnapshot(data) {
  const checks = data.acceptanceChecks || {};
  $("#ai-snapshot-panel").innerHTML = `
    ${detailCell("数据准备", data.dataReady ? "已准备" : "待补齐")}
    ${detailCell("智能功能基础", data.acceptanceReady ? "已具备" : "待完善")}
    ${detailCell("组局样本", `${data.gameCount || 0} 条`)}
    ${detailCell("收藏样本", `${data.favoriteCount || 0} 条`)}
    ${detailCell("人脉样本", `${data.connectionCount || 0} 条`)}
    ${detailCell("足迹样本", `${data.footprintCount || 0} 条`)}
    ${detailCell("局内消息样本", `${data.imMessageCount || 0} 条`)}
    ${detailCell("消息导出", data.imExportEnabled ? "已启用" : "未启用")}
    ${detailCell("用户数据", checkText(checks.users))}
    ${detailCell("组局数据", checkText(checks.games))}
    ${detailCell("行为日志", checkText(checks.behaviorLogs))}
    ${detailCell("收藏数据", checkText(checks.favorites))}
    ${detailCell("评价数据", checkText(checks.reviews))}
  `;
  const sections = data.dataReadinessSections || {};
  renderPaginatedList("#ai-readiness-list", Object.entries(sections), "aiReadiness", ([key, item]) => stackItem({
    title: aiReadinessSectionLabel(key),
    badge: item.ready ? "已准备" : "待补齐",
    meta: [`样本数量：${item.count || 0}`, `状态：${item.ready ? "可用" : "待补齐"}`],
  }), "暂无数据准备分区");
}

async function renderDelivery() {
  bindSectionTabs($("#view-root"));
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
    $("#test-cases-table").innerHTML = emptyRow(5, "当前角色没有验收用例权限");
    $("#test-runs-table").innerHTML = emptyRow(6, "当前角色没有验收记录权限");
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
  renderPaginatedTable("#wechat-tasks-table", state.wechatTasks, "wechatTasks", wechatTaskRow, 8, "暂无微信订阅任务");
}

async function loadWechatTemplates() {
  if (!can("notification:wechat:view")) {
    $("#wechat-template-list").innerHTML = emptyBlock("当前角色没有微信订阅消息权限");
    return;
  }
  const data = await apiGet("/api/admin/notifications/wechat-templates");
  state.wechatTemplates = data.items || [];
  renderPaginatedList("#wechat-template-list", state.wechatTemplates, "wechatTemplates", (item) => stackItem({
    title: item.title || notificationSceneLabel(item.scene),
    badge: statusLabel(item.status),
    meta: [`使用场景：${notificationSceneLabel(item.scene)}`, `模板状态：${statusLabel(item.status)}`],
  }), "暂无订阅模板");
}

async function sendPendingWechatTasks() {
  if (!can("notification:wechat:view")) {
    toast("当前角色没有微信订阅消息权限", true);
    return;
  }
  try {
    const result = await apiPost("/api/internal/notifications/wechat-tasks/send-pending", { limit: 20 });
    toast(`批量发送完成：成功 ${result.sent || 0} 条，失败 ${result.failed || 0} 条，跳过 ${result.skipped || 0} 条`);
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
  renderPaginatedTable("#delivery-documents-table", state.deliveryDocuments, "deliveryDocuments", deliveryDocumentRow, 6, "暂无交付文档");
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
    $("#test-cases-table").innerHTML = emptyRow(5, "当前角色没有验收用例权限");
    return;
  }
  const data = await apiGet("/api/admin/test-cases");
  state.testCases = data.items || [];
  renderPaginatedTable("#test-cases-table", state.testCases, "testCases", testCaseRow, 5, "暂无验收用例");
}

async function createTestCase(event) {
  event.preventDefault();
  if (!can("testcase:manage")) {
    toast("当前角色没有验收用例写入权限", true);
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
    toast("验收用例已新增");
    form.reset();
    await loadTestCases();
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadTestRuns() {
  if (!can("testcase:read")) {
    $("#test-runs-table").innerHTML = emptyRow(6, "当前角色没有验收记录权限");
    return;
  }
  const data = await apiGet("/api/admin/test-runs");
  state.testRuns = data.items || [];
  renderPaginatedTable("#test-runs-table", state.testRuns, "testRuns", testRunRow, 6, "暂无验收记录");
}

async function createTestRun(event) {
  event.preventDefault();
  if (!can("testcase:manage")) {
    toast("当前角色没有验收记录写入权限", true);
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
    toast("验收结果已记录");
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
    toast("当前角色没有验收记录权限", true);
    return;
  }
  try {
    await downloadAdminFile(button.dataset.fileId);
  } catch (error) {
    toast(error.message, true);
  }
}

async function renderAdmins() {
  bindSectionTabs($("#view-root"));
  $("#admin-create-form").addEventListener("submit", createAdminUser);
  $("#admin-users-table").addEventListener("click", onAdminUserTableClick);
  $("#admin-applications-refresh").addEventListener("click", loadAdminApplications);
  $("#admin-applications-table").addEventListener("click", onAdminApplicationTableClick);
  await Promise.all([loadAdminAccounts(), loadAdminRoles(), loadAdminApplications()]);
}

async function createAdminUser(event) {
  event.preventDefault();
  const form = event.currentTarget;
  if (!can("admin_user:create")) {
    toast("当前角色没有创建后台账号权限", true);
    return;
  }
  const data = Object.fromEntries(new FormData(form).entries());
  try {
    await apiPost("/api/admin/admin-users", {
      username: data.username.trim(),
      password: data.password,
      roles: String(data.roles || "").split(/[,，]/).map((item) => adminRoleCode(item.trim())).filter(Boolean),
      status: data.status,
    });
    toast("后台账号已创建");
    form.reset();
    const statusInput = form.querySelector("[name='status']");
    if (statusInput) statusInput.value = "active";
    await loadAdminAccounts();
  } catch (error) {
    toast(error.message, true);
  }
}

async function loadAdminAccounts() {
  const data = await apiGet("/api/admin/admin-users");
  state.adminUsers = data.items || [];
  renderPaginatedTable("#admin-users-table", state.adminUsers, "adminUsers", adminUserRow, 6, "暂无后台账号");
}

async function loadAdminApplications() {
  const data = await apiGet("/api/admin/admin-applications");
  state.adminApplications = data.items || [];
  renderPaginatedTable("#admin-applications-table", state.adminApplications, "adminApplications", adminApplicationRow, 7, "暂无管理员申请");
}

function onAdminApplicationTableClick(event) {
  const button = event.target.closest("button[data-action='copy-admin-application']");
  if (!button) return;
  const item = state.adminApplications.find((application) => String(application.id) === String(button.dataset.id));
  if (!item) {
    toast("申请记录不存在", true);
    return;
  }
  copyText(adminApplicationCopyText(item), "申请信息已复制");
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
      const status = row.querySelector("select[data-field='status']").value;
      await apiPut(`/api/admin/admin-users/${id}`, {
        roles: rolesFromRow(row),
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
  renderPaginatedList("#admin-role-list", state.adminRoles, "adminRoles", (item) => stackItem({
    title: adminRoleLabel(item.code || item.name),
    badge: item.adminCount > 0 ? "active" : "pending",
    meta: [
      `当前账号数：${item.adminCount || 0} 人`,
      `职责：${adminRoleDescription(item)}`,
    ],
  }), "暂无后台角色");
}

async function renderSystem() {
  ensureGuideRulePanel();
  bindSectionTabs($("#view-root"));
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
  $("#growth-rules-refresh").addEventListener("click", loadGrowthRewardRules);
  $("#growth-rules-form").addEventListener("submit", saveGrowthRewardRules);
  $("#achievement-config-refresh").addEventListener("click", loadAchievementConfig);
  $("#achievement-config-form").addEventListener("submit", saveAchievementConfig);
  $("#operation-rules-refresh").addEventListener("click", loadOperationRules);
  $("#operation-rules-form").addEventListener("submit", saveOperationRules);
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
    tasks.push(loadGameCategoryConfig());
    tasks.push(loadHomeDisplayConfig());
    tasks.push(loadGameApplicationConfig());
    tasks.push(loadGameAuditConfig());
    tasks.push(loadConditionRuleConfig());
    tasks.push(loadRoleBenefitConfig());
    tasks.push(loadReviewCompleteConfig());
    tasks.push(loadGrowthRewardRules());
    tasks.push(loadAchievementConfig());
    tasks.push(loadOperationRules());
    tasks.push(loadCreditDeductionRules());
    tasks.push(loadProfitTemplateConfig());
  } else {
    renderNoAccess("#game-category-config-panel", "缺少 system_config:read");
    renderNoAccess("#home-display-config-panel", "缺少 system_config:read");
    renderNoAccess("#game-application-config-panel", "缺少 system_config:read");
    renderNoAccess("#game-audit-config-panel", "缺少 system_config:read");
    renderNoAccess("#condition-rule-config-panel", "缺少 system_config:read");
    renderNoAccess("#role-benefit-config-panel", "缺少 system_config:read");
    renderNoAccess("#review-complete-config-panel", "缺少 system_config:read");
    renderNoAccess("#growth-rules-config-panel", "缺少 system_config:read");
    renderNoAccess("#achievement-config-panel", "缺少 system_config:read");
    renderNoAccess("#operation-rules-config-panel", "缺少 system_config:read");
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

async function loadAchievementConfig() {
  if (!can("system_config:read")) {
    renderNoAccess("#achievement-config-panel", "缺少 system_config:read");
    return;
  }
  const data = await apiGet("/api/admin/growth/achievement-config");
  state.achievementConfig = data.config || data;
  const textarea = $("#achievement-config-json");
  if (textarea) textarea.value = JSON.stringify(state.achievementConfig, null, 2);
  renderAchievementConfig();
}

function renderAchievementConfig() {
  const config = state.achievementConfig || {};
  const catalog = Array.isArray(config.catalog) ? config.catalog : [];
  const locked = Array.isArray(config.locked) ? config.locked : [];
  const roleCount = catalog.filter((item) => Array.isArray(item.roles) && item.roles.length).length;
  $("#achievement-config-panel").innerHTML = [
    detailCell("已解锁目录", `${catalog.length} 项`),
    detailCell("进行中目录", `${locked.length} 项`),
    detailCell("角色专属成就", `${roleCount} 项`),
    detailCell("当前版本", config.version || "-"),
  ].join("");
}

async function saveAchievementConfig(event) {
  event.preventDefault();
  await saveJSONSystemConfig({
    form: event.currentTarget,
    endpoint: "/api/admin/growth/achievement-config",
    stateKey: "achievementConfig",
    textareaSelector: "#achievement-config-json",
    render: renderAchievementConfig,
    successMessage: "成就配置已保存",
  });
}

function ensureGuideRulePanel() {
  if ($("#guide-rule-panel")) return;
  const anchor = $("#permission-tree-panel")?.closest(".split");
  if (!anchor) return;
  anchor.insertAdjacentHTML("afterend", `
    <section id="guide-rule-panel" class="panel hidden" data-tab-panel-group="system">
      <div class="panel-head">
        <div>
          <h2>领路人资格规则</h2>
          <p>后台控制领路人和行家的准入门槛。</p>
        </div>
        <button id="guide-rule-refresh" class="ghost" type="button">刷新</button>
      </div>
      <form id="guide-rule-form" class="form-grid compact-grid hidden">
        <label>规则<input name="ruleId" inputmode="numeric" required placeholder="从规则列表选择" /></label>
        <label>最低邀请数<input name="minInviteCount" inputmode="numeric" placeholder="不填则不变" /></label>
        <label>最低信用分<input name="minCreditScore" inputmode="numeric" placeholder="0-100，不填则不变" /></label>
        <label>最低完成局数<input name="minCompletedGames" inputmode="numeric" placeholder="不填则不变" /></label>
        <label>是否需要付费
          <select name="paymentRequired">
            <option value="">不变</option>
            <option value="true">是</option>
            <option value="false">否</option>
          </select>
        </label>
        <label>状态
          <select name="status">
            <option value="">不变</option>
            <option value="active">正常</option>
            <option value="disabled">已禁用</option>
          </select>
        </label>
        <button class="primary" type="submit">保存规则</button>
      </form>
      <div id="guide-rule-list" class="stack-list"></div>
      <form id="guide-qualification-form" class="form-grid compact-grid profile-form-gap hidden">
        <label>用户<input name="userId" inputmode="numeric" required placeholder="领路人/行家用户" /></label>
        <label>条件是否满足
          <select name="conditionMet">
            <option value="">不变</option>
            <option value="true">是</option>
            <option value="false">否</option>
          </select>
        </label>
        <label>付费是否满足
          <select name="paymentMet">
            <option value="">不变</option>
            <option value="true">是</option>
            <option value="false">否</option>
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

function environmentLabel(value) {
  return { production: "生产环境", staging: "预发布环境", development: "本地开发环境", test: "验收环境" }[value] || adminDisplayValue(value || "-");
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
    detailCell("首页分类", `${(config.primaryCategories || []).length} 类`),
    detailCell("局型筛选", mergeGameTypeOptions(config.typeFilters).map((item) => item.name).join("、")),
    detailCell("位置筛选", `${(config.locationFilters || []).length} 项`),
    detailCell("默认分类", categoryName(config.primaryCategories, config.defaultPrimaryCategory)),
    detailCell("默认局型", gameTypeLabel(config.defaultType)),
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
    toast("配置数据不能为空", true);
    return;
  }
  let payload;
  try {
    payload = JSON.parse(raw);
  } catch (error) {
    toast("配置数据格式不正确", true);
    return;
  }
  try {
    const data = await apiPut("/api/admin/games/category-config", payload);
    state.gameCategoryConfig = data.config || data;
    const textarea = $("#game-category-config-json");
    if (textarea) textarea.value = JSON.stringify(state.gameCategoryConfig, null, 2);
    renderGameCategoryConfig();
    toast("局类型与筛选规则已保存");
  } catch (error) {
    toast(error instanceof SyntaxError ? "高级筛选格式不正确，请清空后重试" : error.message, true);
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
    detailCell("在线人数基数", config.onlineBaseCount ?? 0),
    detailCell("在线人数后缀", config.onlineSuffix || "-"),
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
    toast("配置数据不能为空", true);
    return;
  }
  let payload;
  try {
    payload = JSON.parse(raw);
  } catch (error) {
    toast("配置数据格式不正确", true);
    return;
  }
  try {
    const data = await apiPut("/api/admin/home/display-config", payload);
    state.homeDisplayConfig = data.config || data;
    const textarea = $("#home-display-config-json");
    if (textarea) textarea.value = JSON.stringify(state.homeDisplayConfig, null, 2);
    renderHomeDisplayConfig();
    toast("首页展示数据已保存");
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
    detailCell("入局协议", config.agreementTitle || "-"),
    detailCell("要求实名", yesNo(config.requireRealname)),
    detailCell("要求确认协议", yesNo(config.requireAgreement)),
    detailCell("允许重复申请", yesNo(config.allowDuplicateApply)),
    detailCell("最多上传材料", `${config.maxUploadCount ?? 0} 份`),
    detailCell("小程序搜索", config.searchEnabled ? "已开放" : "已关闭"),
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
    successMessage: "入局申请规则已保存",
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
    detailCell("普通局自动审核", yesNo(config.autoApproveFreeGames)),
    detailCell("需人工审核局型", compactList((config.requireManualAuditTypes || []).map(gameTypeLabel)) || "无"),
    detailCell("驳回必须填写原因", yesNo(config.requiredRejectReason)),
    detailCell("入局审核模式", applicationAuditModeLabel(config.applicationAuditMode)),
    detailCell("批量审核上限", `${config.batchAuditMaxCount ?? 0} 条`),
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
    successMessage: "组局审核规则已保存",
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
    detailCell("条件局能力", config.enabled ? "已启用" : "已关闭"),
    detailCell("小程序可见", yesNo(config.visibleInMiniProgram)),
    detailCell("仅后台可创建", yesNo(config.adminOnlyCreate)),
    detailCell("条件项数量", `${(config.ruleItems || []).length} 项`),
    detailCell("默认可见性", visibilityLabel(config.defaultVisibility)),
    detailCell("需要支付能力", yesNo(config.paymentRequired)),
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
    successMessage: "条件局规则已保存",
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
    detailCell("入口提示", `${Object.keys(prompts).length} 条`),
    detailCell("角色数量", `${(comparison.roles || []).length} 个`),
    detailCell("权益数量", `${(comparison.benefits || []).length} 项`),
    detailCell("权益标题", comparison.title || "-"),
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
    successMessage: "角色权益展示已保存",
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
    detailCell("完成后权益", `${(config.benefits || []).length} 项`),
    detailCell("再玩选项", `${(config.playOptions || []).length} 个`),
    detailCell("默认意愿", againIntentLabel((config.playOptions || [])[0]?.intent)),
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
    successMessage: "评价完成展示已保存",
  });
}

async function loadGrowthRewardRules() {
  if (!can("system_config:read")) {
    renderNoAccess("#growth-rules-config-panel", "缺少 system_config:read");
    return;
  }
  const data = await apiGet("/api/admin/growth/reward-rules");
  state.growthRewardRules = data.config || data;
  const textarea = $("#growth-rules-config-json");
  if (textarea) textarea.value = JSON.stringify(state.growthRewardRules, null, 2);
  renderGrowthRewardRules();
}

function renderGrowthRewardRules() {
  const config = state.growthRewardRules || {};
  $("#growth-rules-config-panel").innerHTML = [
    detailCell("完成组局经验", config.completedGameExperience ?? "-"),
    detailCell("完成组局积分", config.completedGamePoints ?? "-"),
    detailCell("提交评价经验", config.submittedReviewExperience ?? "-"),
    detailCell("收到评价经验", config.receivedReviewExperience ?? "-"),
    detailCell("提交评价积分", config.submittedReviewPoints ?? "-"),
    detailCell("升级所需经验", config.experiencePerLevel ?? "-"),
    detailCell("信用初始分（固定）", 100),
    detailCell("信用分上限", config.creditScoreCap ?? "-"),
  ].join("");
}

async function saveGrowthRewardRules(event) {
  event.preventDefault();
  await saveJSONSystemConfig({
    form: event.currentTarget,
    endpoint: "/api/admin/growth/reward-rules",
    stateKey: "growthRewardRules",
    textareaSelector: "#growth-rules-config-json",
    render: renderGrowthRewardRules,
    successMessage: "成长等级与积分规则已保存",
  });
}

async function loadOperationRules() {
  if (!can("system_config:read")) {
    renderNoAccess("#operation-rules-config-panel", "缺少 system_config:read");
    return;
  }
  const data = await apiGet("/api/admin/operation-rules");
  state.operationRules = data.config || data;
  const textarea = $("#operation-rules-config-json");
  if (textarea) textarea.value = JSON.stringify(state.operationRules, null, 2);
  renderOperationRules();
}

function renderOperationRules() {
  const config = state.operationRules || {};
  const tasks = config.tasks?.items || [];
  const roles = config.roles || {};
  const game = config.game || {};
  const invite = config.invite || {};
  const stateRules = config.state || {};
  $("#operation-rules-config-panel").innerHTML = [
    detailCell("任务规则", `${tasks.length} 条`),
    detailCell("行家发起局门槛", `${roles.expertCreatedGames ?? "-"} 次`),
    detailCell("领路人参与局门槛", `${roles.guideParticipatedGames ?? "-"} 次`),
    detailCell("组局人数", `${game.minPlayers ?? "-"}-${game.maxPlayers ?? "-"}`),
    detailCell("每日创建上限", `${game.dailyCreateLimit ?? "-"} 局`),
    detailCell("邀请有效期", `${invite.timeoutMinutes ?? "-"} 分钟`),
    detailCell("状态审核超时", `${stateRules.auditTimeoutHours ?? "-"} 小时`),
  ].join("");
}

async function saveOperationRules(event) {
  event.preventDefault();
  await saveJSONSystemConfig({
    form: event.currentTarget,
    endpoint: "/api/admin/operation-rules",
    stateKey: "operationRules",
    textareaSelector: "#operation-rules-config-json",
    render: renderOperationRules,
    successMessage: "运营约束规则已保存",
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
    detailCell("规则总数", `${items.length} 条`),
    detailCell("已启用", `${enabledCount} 条`),
    detailCell("已停用", `${items.length - enabledCount} 条`),
    detailCell("低分评价扣分", items.find((item) => item.ruleCode === "low_review")?.changeValue ?? "-"),
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
    toast("配置数据不能为空", true);
    return;
  }
  let payload;
  try {
    payload = JSON.parse(raw);
  } catch (error) {
    toast("配置数据格式不正确", true);
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
    setFormValue(form, "depositRuleText", state.profitTemplateConfig.depositRuleText || "");
    setFormValue(form, "depositNoticeText", state.profitTemplateConfig.depositNoticeText || "");
    setFormValue(form, "version", state.profitTemplateConfig.version || "");
  }
  renderProfitTemplateConfig();
}

function renderProfitTemplateConfig() {
  const config = state.profitTemplateConfig || {};
  $("#profit-config-panel").innerHTML = [
    detailCell("押金局规则", config.depositRuleText || "-"),
    detailCell("押金提示", config.depositNoticeText || "-"),
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
    toast("押金局规则已保存");
  } catch (error) {
    toast(error.message, true);
  }
}

async function createSensitiveWord(event) {
  event.preventDefault();
  const form = event.currentTarget || (event.target && event.target.closest ? event.target.closest("form") : null);
  if (!can("content:sensitive_word:create")) {
    toast("缺少 content:sensitive_word:create", true);
    return;
  }
  const data = Object.fromEntries(new FormData(form).entries());
  try {
    await apiPost("/api/admin/sensitive-words", {
      word: data.word.trim(),
      level: data.level,
      action: data.action,
      status: "active",
    });
    toast("敏感词已新增");
    if (form && typeof form.reset === "function") form.reset();
    await loadSensitiveWords();
  } catch (error) {
    toast(error.message, true);
  }
}

async function importSensitiveWords(event) {
  event.preventDefault();
  const form = event.currentTarget || (event.target && event.target.closest ? event.target.closest("form") : null);
  if (!can("content:sensitive_word:import")) {
    toast("缺少 content:sensitive_word:import", true);
    return;
  }
  const rawWords = String(new FormData(form).get("words") || "");
  const words = Array.from(new Set(rawWords
    .split(/[\s,，、]+/)
    .map((item) => item.trim())
    .filter(Boolean)));
  if (words.length === 0) {
    toast("请先填写要导入的敏感词", true);
    return;
  }
  try {
    await apiPost("/api/admin/sensitive-words/import", { words });
    toast(`已导入 ${words.length} 个敏感词`);
    if (form && typeof form.reset === "function") form.reset();
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
  renderPaginatedTable("#sensitive-words-table", state.sensitiveWords, "sensitiveWords", sensitiveWordRow, 7, "暂无敏感词");
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
  renderPaginatedList("#risk-log-list", state.riskLogs, "riskLogs", (item) => stackItem({
    title: item.id ? `${sensitiveActionLabel(item.action)}：${item.word || "内容"}（记录 ${item.id}）` : `${sensitiveActionLabel(item.action)}：${item.word || "内容"}`,
    badge: statusLabel(item.status),
    meta: [`关联组局：${item.gameId ? `局 ${item.gameId}` : "-"}`, `聊天房间：${item.roomId || "-"}`, `消息记录：${item.messageId || "-"}`, `来源：${sourceLabel(item.source)}`, `创建时间：${formatTime(item.createdAt)}`],
  }), "暂无内容风险日志");
}

async function loadPermissionTree() {
  const data = await apiGet("/api/admin/permissions/tree");
  state.permissionTree = data;
  $("#permission-tree-panel").innerHTML = [
    stackItem({
      title: `${data.adminUser?.username || "管理员"} 的后台范围`,
      badge: statusLabel(data.adminUser?.status || "active"),
      meta: [
        `角色：${(data.roles || []).map(adminRoleLabel).join("、") || "-"}`,
        `职责：${adminRoleScopeText({ roles: data.roles || [] })}`,
      ],
    }),
    stackItem({
      title: "可访问菜单",
      badge: "active",
      meta: [compactList((data.menus || []).map((item) => item.title || item.name || item.label || item.code), 16)],
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
  renderPaginatedList("#guide-rule-list", state.guideQualificationRules, "guideQualificationRules", (item) => stackItem({
    title: item.id ? `${guideRuleLabel(item.ruleCode)}（规则 ${item.id}）` : guideRuleLabel(item.ruleCode),
    badge: statusLabel(item.status),
    meta: [
      `最低邀请人数：${item.minInviteCount || 0}`,
      `最低信用分：${item.minCreditScore || 0}`,
      `最低完成局数：${item.minCompletedGames || 0}`,
      `是否需要付费：${Boolean(item.paymentRequired) ? "是" : "否"}`,
      `更新时间：${formatTime(item.updatedAt)}`,
    ],
    action: `<button class="ghost" data-action="guide-rule-load" data-id="${item.id}" type="button">载入</button>`,
  }), "暂无领路人资格规则");
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
    toast("智能局内消息 导出开关已更新");
    await loadAIExportConfig();
  } catch (error) {
    toast(error.message, true);
  }
}

function fillGuideRuleForm(ruleID) {
  const rule = state.guideQualificationRules.find((item) => String(item.id) === String(ruleID));
  const form = $("#guide-rule-form");
  if (!rule || !form) return;
  setFormValue(form, "ruleId", rule.id || "");
  setFormValue(form, "minInviteCount", rule.minInviteCount ?? "");
  setFormValue(form, "minCreditScore", rule.minCreditScore ?? "");
  setFormValue(form, "minCompletedGames", rule.minCompletedGames ?? "");
  setFormValue(form, "paymentRequired", String(Boolean(rule.paymentRequired)));
  setFormValue(form, "status", rule.status || "");
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
    toast("请输入有效的规则编号", true);
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
    toast("请输入有效的用户", true);
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
    toast("请输入有效的用户", true);
    return;
  }
  if (!("conditionMet" in payload) && !("paymentMet" in payload)) {
    toast("请选择要更新的资格条件", true);
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
    detailCell("用户", userText(item.userId)),
    detailCell("条件已满足", yesNo(item.conditionMet)),
    detailCell("支付条件已满足", yesNo(item.paymentMet)),
    detailCell("领路人开通状态", statusLabel(item.guideOpenStatus)),
    detailCell("更新时间", formatTime(item.updatedAt)),
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
    detailCell("导出开关", config.enabled ? "已启用" : "已关闭"),
    detailCell("更新时间", formatTime(config.updatedAt)),
    detailCell("更新人", config.updatedBy ? `管理员 ${config.updatedBy}` : "-"),
    detailCell("风险说明", config.enabled ? "允许导出局内消息" : "禁止导出局内消息"),
  ].join("");
}

async function renderIM() {
  $("#im-filter-form").addEventListener("submit", (event) => {
    event.preventDefault();
    loadIMRooms(new FormData(event.currentTarget));
  });
  $("#im-room-list").addEventListener("click", onIMRoomListClick);
  await loadIMRooms();
}

async function loadIMRooms(formData) {
  const data = await apiGet(`/api/admin/im/rooms${querySuffix(formData)}`);
  const list = $("#im-room-list");
  state.imRooms = data.items || [];
  renderPaginatedList("#im-room-list", state.imRooms, "imRooms", (item) => stackItem({
    title: `房间 ${item.id || item.roomId} / 局 ${item.gameId}`,
    badge: item.status,
    meta: [`消息服务：${imEngineLabel(item.engine)}`, `消息数：${item.messageCount || 0}`, `文件数：${item.fileMessageCount || 0}`],
    action: imRoomActions(item),
  }), "暂无局内消息房间");
}

function imRoomActions(item) {
  const roomID = item.id || item.roomId;
  const actions = [`<button class="ghost" data-action="im-room-detail" data-id="${escapeHTML(roomID)}" type="button">查看完整消息</button>`];
  if (can("im:message:view_dispute")) actions.push(`<button class="ghost" data-action="im-dispute-messages" data-id="${escapeHTML(roomID)}" type="button">争议消息</button>`);
  const needsSync = ["sync_failed", "failed"].includes(item.status) || (item.engine === "openim" && !item.openIMGroupId);
  if (can("im:room:retry_create") && needsSync) actions.push(`<button class="ghost" data-action="im-room-retry" data-id="${escapeHTML(roomID)}" type="button">重试同步</button>`);
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
      toast("局内消息房间已重试同步");
      await loadIMRooms(new FormData($("#im-filter-form")));
    }
    if (button.dataset.action === "im-room-archive") {
      await apiPost(`/api/admin/im/rooms/${roomID}/archive`, { reason: "admin archived from web" });
      toast("局内消息房间已归档");
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
  const panel = openAdminDrawer({
    title: `${room.id ? `消息房间 ${room.id}` : "局内消息房间"}详情`,
    subtitle: disputeOnly ? "争议证据消息" : "房间消息、文件消息和消息服务同步状态",
    body: `
    <div class="row-actions">
      <span class="${badgeClass(room.status)}">${statusLabel(room.status)}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("关联组局", room.gameId ? `局 ${room.gameId}` : "-")}
      ${detailCell("消息服务", imEngineLabel(room.engine))}
      ${detailCell("成员", compactList((room.memberIds || []).map(userText)) || "-")}
      ${detailCell("文字消息", `${messages.length} 条`)}
      ${detailCell("文件消息", `${fileMessages.length} 条`)}
      ${detailCell("归档原因", room.archiveReason || "-")}
    </div>
    ${technicalDetails("消息服务技术信息", detailCell("服务群组", room.openIMGroupId || "-"))}
    ${gameOpsTable(disputeOnly ? "争议消息" : "全部消息", ["消息编号", "发送人", "类型", "完整内容", "状态", "文件", "已送达", "已读", "发送时间", "操作"], messages.map(imMessageRow).join("") || emptyRow(10, "暂无消息"))}
    ${disputeOnly ? "" : gameOpsTable("文件消息", ["消息编号", "发送人", "类型", "完整内容", "状态", "文件", "已送达", "已读", "发送时间", "操作"], fileMessages.map(imMessageRow).join("") || emptyRow(10, "暂无文件消息"))}
  `,
  });
  panel.addEventListener("click", onIMDetailClick);
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
    toast("局内消息已隐藏");
    await showIMRoomDetail(button.dataset.roomId);
  } catch (error) {
    toast(error.message, true);
  }
}

async function renderLogs() {
  ensureLogTools();
  $("#logs-table").addEventListener("click", onLogsTableClick);
  $("#view-root button[data-action='reload-logs']")?.addEventListener("click", () => {
    loadOperationLogs(new FormData($("#log-filter-form")));
  });
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
        <select name="action">
          <option value="">全部操作</option>
          <option value="invite_code:create">生成邀请码</option>
          <option value="invite_code:disable">禁用邀请码</option>
          <option value="game:create_admin">后台创建组局</option>
          <option value="game:update_status">更新组局状态</option>
          <option value="role:update">审核角色申请</option>
          <option value="report:handle">处理举报申诉</option>
          <option value="revenue:generate">生成分润记录</option>
          <option value="redemption:manage">处理积分兑换</option>
          <option value="feedback:view">查看用户反馈</option>
          <option value="im:room:read">查看局内消息</option>
        </select>
        <select name="targetType">
          <option value="">全部业务</option>
          <option value="invite_code">邀请码</option>
          <option value="game">组局</option>
          <option value="role">角色申请</option>
          <option value="report">举报申诉</option>
          <option value="revenue">分润记录</option>
          <option value="redemption">积分兑换</option>
          <option value="feedback">用户反馈</option>
          <option value="im_room">局内消息</option>
        </select>
        <input name="adminUserId" inputmode="numeric" placeholder="管理员" />
        <button class="ghost" type="submit">筛选</button>
      </form>
    `);
  }
}

async function loadOperationLogs(formData) {
  if (formData) resetPagination("operationLogs");
  const data = await apiGet(`/api/admin/operation-logs${querySuffix(formData)}`);
  state.operationLogs = data.items || [];
  renderPaginatedTable("#logs-table", state.operationLogs, "operationLogs", operationLogRow, 5, "暂无操作日志");
}

function onLogsTableClick(event) {
  const button = event.target.closest("button[data-action='log-detail']");
  if (!button) return;
  showOperationLogDetail(button.dataset.id);
}

function showOperationLogDetail(id) {
  const item = state.operationLogs.find((log) => String(log.id) === String(id));
  if (!item) return;
  openAdminDrawer({
    title: "操作日志详情",
    subtitle: operationActionLabel(item.action),
    body: `
    <div class="row-actions">
      <span class="badge neutral">${escapeHTML(operationActionLabel(item.action))}</span>
    </div>
    <div class="detail-grid">
      ${detailCell("管理员", adminUserText(item.adminUserId))}
      ${detailCell("操作", operationActionLabel(item.action))}
      ${detailCell("关联业务", operationTargetText(item))}
      ${detailCell("登录地址", item.ip || "-")}
      ${detailCell("时间", formatTime(item.createdAt))}
      ${detailCell("操作摘要", operationDetailText(item.detail))}
    </div>
  `,
  });
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
    toast("无有效文件编号", true);
    return;
  }
  const result = await apiGet(`/api/admin/files/${encodeURIComponent(fileID)}/download-url`);
  const url = typeof result === "string" ? result : result.downloadUrl;
  if (!url) {
    toast("文件暂时无法下载", true);
    return;
  }
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(url).catch(() => {});
  }
  window.open(url, "_blank", "noopener");
  toast("文件下载已准备");
}

async function copyText(text, successMessage = "已复制") {
  const value = String(text || "").trim();
  if (!value) {
    toast("没有可复制的内容", true);
    return;
  }
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(value);
      toast(successMessage);
      return;
    }
  } catch (error) {
    // Fall through to the textarea copy path below.
  }
  const textarea = document.createElement("textarea");
  textarea.value = value;
  textarea.setAttribute("readonly", "readonly");
  textarea.style.position = "fixed";
  textarea.style.left = "-9999px";
  document.body.appendChild(textarea);
  textarea.select();
  const copied = document.execCommand("copy");
  textarea.remove();
  toast(copied ? successMessage : "复制失败，请打开详情手动复制", !copied);
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
      <td>${escapeHTML(userText(user.id))}</td>
      <td>${escapeHTML(wechatBindingText(user.openId))}</td>
      <td>${escapeHTML(user.nickname || userText(user.id))}</td>
      <td>${escapeHTML(inviterLabel(user.inviter || { id: user.inviterUserId, nickname: user.inviterNickname }, user.inviteRelation))}</td>
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
          <h2>${escapeHTML(gameID ? `局 ${gameID}` : "组局")}运营明细</h2>
        </div>
        ${can("game:update_status") ? '<button class="ghost" data-action="ensure-game-im-room" type="button">创建/同步 IM 房间</button>' : ""}
      </div>
      <form data-role="game-milestone-form" class="form-grid compact-grid">
        <label>里程碑标题<input name="title" required maxlength="120" placeholder="确认场地 / 完成服务" /></label>
        <label>状态
          <select name="status">
            <option value="pending">待处理</option>
            <option value="done">已完成</option>
          </select>
        </label>
        <button class="ghost" type="submit">新增里程碑</button>
      </form>
      <div class="split profile-form-gap">
        ${gameOpsTable("里程碑", ["编号", "事项", "状态", "创建时间"], milestones.map(gameMilestoneRow).join("") || emptyRow(4, "暂无里程碑"))}
        ${gameOpsTable("打卡", ["编号", "用户", "里程碑", "类型", "状态", "时间", "操作"], checkins.map(gameCheckinRow).join("") || emptyRow(7, "暂无打卡"))}
      </div>
      <div class="split profile-form-gap">
        ${gameOpsTable("复盘", ["编号", "用户", "再玩意愿", "复盘内容", "时间"], retrospectives.map(gameRetrospectiveRow).join("") || emptyRow(5, "暂无复盘"))}
        ${gameOpsTable("续局草稿", ["原局", "新局", "创建人", "标题", "状态", "创建时间"], continueDrafts.map(gameContinueDraftRow).join("") || emptyRow(6, "暂无续局草稿"))}
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
          <thead><tr>${headers.map((item) => `<th>${escapeHTML(adminLabel(item))}</th>`).join("")}</tr></thead>
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
      <td>${escapeHTML(userText(item.userId))}</td>
      <td>${escapeHTML(item.gameId ? `局 ${item.gameId}` : "-")}</td>
      <td>${escapeHTML(game.title || "-")}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function inviteCodeRow(item) {
  const canDisable = item.status === "active" && can("invite_code:manage");
  const actions = [
    `<button class="ghost" data-action="invite-materials" data-code="${escapeHTML(item.code)}" type="button">生成物料</button>`,
    `<button class="ghost" data-action="invite-copy-usage" data-code="${escapeHTML(item.code)}" type="button">复制使用方式</button>`,
    `<button class="ghost" data-action="invite-detail" data-code="${escapeHTML(item.code)}" type="button">详情</button>`,
  ];
  if (canDisable) actions.push(`<button class="ghost" data-action="invite-disable" data-code="${escapeHTML(item.code)}" type="button">禁用</button>`);
  return `
    <tr>
      <td>${escapeHTML(item.id ? `邀请码 ${item.id}` : "-")}</td>
      <td>${escapeHTML(item.code)}</td>
      <td>${escapeHTML(inviteEntryLabel(item.entryType))}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${escapeHTML(inviteUseText(item))}</td>
      <td>${escapeHTML(boundUserText(item))}</td>
      <td><div class="row-actions">${actions.join("")}</div></td>
    </tr>
  `;
}

function inviteUseText(item) {
  if (!item) return "-";
  if (item.boundWechatUserId || Number(item.usedCount || 0) > 0) return "已绑定";
  if (item.status === "disabled") return "已禁用";
  if (item.status === "exhausted") return "已用完";
  return "未使用";
}

function inviteOwnerText(invite, owner = {}) {
  return "后台管理员";
}

function boundUserText(item) {
  if (!item) return "-";
  if (item.boundWechatNickname) return item.boundWechatNickname;
  if (item.boundWechatUserId) return userText(item.boundWechatUserId);
  return "-";
}

function inviteRelationRow(item) {
  return `
    <tr>
      <td>${escapeHTML(inviteCodeText(item))}</td>
      <td>${escapeHTML(userText(item.inviteeUserId))}</td>
      <td>${escapeHTML(sourceLabel(item.bindSource))}</td>
      <td>${escapeHTML(inviteEntryLabel(item.entryType))}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function inviteRelationDetailRow(item, entryType) {
  return `
    <tr>
      <td>${escapeHTML(inviteCodeText(item))}</td>
      <td>${escapeHTML(userText(item.inviteeUserId))}</td>
      <td>${escapeHTML(sourceLabel(item.bindSource))}</td>
      <td>${inviteEntryLabel(entryType)}</td>
    </tr>
  `;
}

function inviteCodeText(item) {
  if (!item) return "-";
  return item.code || item.inviteCode || (item.inviteCodeId ? `邀请码 ${item.inviteCodeId}` : "-");
}

function copyInviteUsage(code) {
  return copyInviteMaterialText(code);
}

async function copyInviteMaterialText(code) {
  if (!code) return;
  const result = await apiGet(`/api/admin/invite-codes/${encodeURIComponent(code)}/materials`);
  const text = inviteMaterialCopyText(result, code);
  return copyText(text, result.available ? "发送文案已复制" : "物料未就绪说明已复制");
}

async function showInviteMaterials(code) {
  if (!code) return;
  if (!$("#invite-materials-result")) {
    await showInviteCodeDetail(code);
  }
  const result = await apiGet(`/api/admin/invite-codes/${encodeURIComponent(code)}/materials`);
  const holder = $("#invite-materials-result");
  if (!holder) return;
  const urlLink = result.urlLink || "";
  const wxaCodeDataUrl = result.wxaCodeDataUrl || "";
  const copyValue = inviteMaterialCopyText(result, code);
  const statusText = urlLink ? "链接可发送" : wxaCodeDataUrl ? "二维码可发送" : "待配置";
  const linkText = urlLink || (result.urlLinkError ? "正式发布后可生成：" + result.urlLinkError : "未生成");
  holder.innerHTML = `
    <div class="sub-panel invite-material-panel">
      <div class="panel-head">
        <div>
          <h2>可发给用户的物料</h2>
          <p>${escapeHTML(result.message || "链接可直接发给微信用户点击；二维码图片可保存后用于海报或线下扫码。")}</p>
        </div>
        <div class="row-actions">
          <span class="${result.available ? "status-pill ok" : "status-pill warn"}">${escapeHTML(statusText)}</span>
          <button class="ghost" data-action="copy-invite-share-text" type="button">复制发送文案</button>
          ${urlLink ? `<button class="ghost" data-action="copy-url-link" type="button">复制链接</button>` : ""}
          ${wxaCodeDataUrl ? `<a class="ghost" href="${escapeHTML(wxaCodeDataUrl)}" download="invite-${escapeHTML(code)}.png">下载二维码</a>` : ""}
        </div>
      </div>
      <div class="invite-send-card">
        <h3>直接发给用户</h3>
        <textarea readonly>${escapeHTML(copyValue)}</textarea>
      </div>
      <div class="detail-grid profile-detail-grid">
        ${detailCell("可点击链接", linkText)}
        ${detailCell("二维码状态", wxaCodeDataUrl ? "已生成" : result.wxaCodeError || "未生成")}
        ${detailCell("入口类型", result.entryLabel || inviteEntryLabel(result.entryType))}
        ${detailCell("邀请码", result.inviteCode || code)}
      </div>
      ${wxaCodeDataUrl ? `<div class="invite-code-preview"><img alt="邀请码二维码" src="${escapeHTML(wxaCodeDataUrl)}" /></div>` : ""}
      <details class="dev-params">
        <summary>开发参数</summary>
        <div class="detail-grid profile-detail-grid">
          ${detailCell("小程序路径", result.path)}
          ${detailCell("小程序码参数", result.scene)}
          ${detailCell("链接参数", result.query)}
        </div>
      </details>
    </div>
  `;
  const copyShareButton = holder.querySelector("button[data-action='copy-invite-share-text']");
  if (copyShareButton) {
    copyShareButton.addEventListener("click", () => copyText(copyValue, result.available ? "发送文案已复制" : "物料未就绪说明已复制"));
  }
  const copyButton = holder.querySelector("button[data-action='copy-url-link']");
  if (copyButton && urlLink) {
    copyButton.addEventListener("click", () => copyText(urlLink, "小程序链接已复制"));
  }
}

function inviteMaterialCopyText(result, code) {
  if (result?.copyText) return result.copyText;
  const inviteCode = result?.inviteCode || code || "";
  const entryLabel = result?.entryLabel || inviteEntryLabel(result?.entryType);
  if (result?.urlLink) {
    return [
      "真好玩邀请入口",
      `邀请码：${inviteCode}`,
      `入口方式：${entryLabel}`,
      `点击进入小程序：${result.urlLink}`,
      "进入后请按提示登录，系统会自动绑定该邀请码。",
    ].join("\n");
  }
  return [
    "该邀请码暂不能直接发送给用户。",
    `邀请码：${inviteCode}`,
    `入口方式：${entryLabel}`,
    "原因：后台还没有生成微信可打开链接或小程序码。",
    "请先完成微信 AppID/AppSecret 配置，再重新生成物料。",
  ].join("\n");
}

function inviteUsageInfo(item) {
  const code = inviteCodeText(item);
  const entryType = ["poster", "qrcode", "link"].includes(item?.entryType) ? item.entryType : "link";
  const query = `inviteCode=${encodeURIComponent(code)}&entryType=${encodeURIComponent(entryType)}`;
  const path = `/pages/login/invite/index?${query}`;
  const tipMap = {
    poster: "用于小程序分享卡片或海报入口，入口参数必须带 inviteCode 和 entryType。",
    qrcode: "用于生成小程序码或二维码，scene 只填邀请码，入口类型由后台记录决定。",
    link: "用于小程序链接入口，链接打开后进入邀请登录页并自动带入邀请码。",
  };
  return {
    code,
    entryType,
    entryLabel: inviteEntryLabel(entryType),
    path,
    query,
    scene: code,
    tip: tipMap[entryType] || "用户进入小程序后自动校验邀请码并绑定微信用户。",
  };
}

function inviteEntryLabel(entryType) {
  const labels = {
    poster: "小程序卡片",
    qrcode: "二维码",
    link: "链接",
  };
  return labels[entryType] || entryType || "-";
}

function revenueRecordRow(item) {
  const canFreeze = item.status !== "frozen" && item.status !== "settled";
  const canSettle = item.status === "pending_settlement";
  return `
    <tr>
      <td>${escapeHTML(item.id ? `分润记录 ${item.id}` : "-")}</td>
      <td>${escapeHTML(item.recordNo)}</td>
      <td>${escapeHTML(item.gameId ? `局 ${item.gameId}` : "-")}</td>
      <td>${escapeHTML(revenueTemplateName(item.templateId))}</td>
      <td>${escapeHTML(yuanText(item.amountCent))}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${escapeHTML((item.items || []).map((child) => `${roleLabel(child.role)}：${userText(child.userId)}，${yuanText(child.amountCent)}`).join(" / "))}</td>
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
  const imageURL = item.imageUrl || "";
  return `
    <tr>
      <td>
        <div class="product-image-cell">
          ${imageURL ? `<img src="${escapeHTML(imageURL)}" alt="${escapeHTML(item.name || "兑换商品")}" />` : `<span>未设置图片</span>`}
        </div>
      </td>
      <td><input data-field="imageUrl" value="${escapeHTML(imageURL)}" placeholder="商品图片地址" /></td>
      <td><input data-field="name" value="${escapeHTML(item.name || "")}" /></td>
      <td><input data-field="pointsCost" inputmode="numeric" value="${escapeHTML(item.pointsCost || 0)}" /></td>
      <td><input data-field="stock" inputmode="numeric" value="${escapeHTML(item.stock || 0)}" /></td>
      <td>
        <select data-field="status">
          <option value="active" ${item.status === "active" ? "selected" : ""}>正常</option>
          <option value="inactive" ${item.status === "inactive" ? "selected" : ""}>已下架</option>
        </select>
        <input data-field="description" type="hidden" value="${escapeHTML(item.description || "")}" />
      </td>
      <td>${formatTime(item.updatedAt || item.createdAt)}</td>
      <td>
        <div class="row-actions">
          <button class="ghost" data-action="redemption-item-update" data-id="${item.id}" type="button">保存</button>
          <button class="ghost" data-action="redemption-item-status" data-id="${item.id}" data-status="${nextStatus}" type="button">${statusLabel(nextStatus)}</button>
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
    actions.push(`<button class="ghost" data-action="redemption-order-shipping" data-id="${item.id}" data-status="shipping" data-reason="后台已发货" type="button">发货</button>`);
  }
  if (item.status === "shipping") {
    actions.push(`<button class="ghost" data-action="redemption-order-fulfill" data-id="${item.id}" data-status="fulfilled" data-reason="后台已完成履约" type="button">完成履约</button>`);
  }
  return `
    <tr>
      <td>${escapeHTML(item.id ? `订单 ${item.id}` : "-")}</td>
      <td>${escapeHTML(item.orderNo || "-")}</td>
      <td>${escapeHTML(userText(item.userId))}</td>
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
      <td>${escapeHTML(item.id ? `流水 ${item.id}` : "-")}</td>
      <td>${escapeHTML(userText(item.userId))}</td>
      <td>${escapeHTML(item.changeValue)}</td>
      <td>${escapeHTML(item.beforePoints)} -> ${escapeHTML(item.afterPoints)}</td>
      <td>${escapeHTML(pointsBizText(item))}</td>
      <td>${escapeHTML(item.reason || "-")}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function memberTeamRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.id ? `团队 ${item.id}` : "-")}</td>
      <td>${escapeHTML(userText(item.leaderUserId))}</td>
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
      <td>${escapeHTML(userText(item.userId))}</td>
      <td>${escapeHTML(item.period || item.reportType || "-")}</td>
      <td>${escapeHTML(item.membershipPlan || "-")}</td>
      <td>${escapeHTML(item.invitedCount || 0)}</td>
      <td>${escapeHTML(item.participatedGames || 0)} / ${escapeHTML(item.completedGames || 0)}</td>
      <td>${escapeHTML(`${yuanText(income.totalCent || 0)} / 待结算 ${yuanText(income.pendingCent || 0)} / 已结算 ${yuanText(income.settledCent || 0)}`)}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function memberTeamMemberRow(item) {
  return `
    <tr>
      <td>${escapeHTML(userText(item.userId))}</td>
      <td>${escapeHTML(item.relationLevel || 0)}</td>
      <td>${escapeHTML(sourceLabel(item.source))}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${formatTime(item.joinedAt)}</td>
    </tr>
  `;
}

function connectionRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.id ? `关系 ${item.id}` : "-")}</td>
      <td>${escapeHTML(userText(item.userId))}</td>
      <td>${escapeHTML(userText(item.connectedUserId))}</td>
      <td>${escapeHTML(relationTypeLabel(item.relationType))}</td>
      <td>${escapeHTML(sourceLabel(item.sourceType))}${item.sourceId ? "（已关联）" : ""}</td>
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
      <td>${escapeHTML(item.id ? `举报 ${item.id}` : "-")}</td>
      <td>${escapeHTML(item.gameId ? `局 ${item.gameId}` : "-")}</td>
      <td>${escapeHTML(userText(item.reporterUserId))}</td>
      <td>${escapeHTML(userText(item.targetUserId))}</td>
      <td>${escapeHTML(reportTypeLabel(item.reportType))}</td>
      <td>${escapeHTML(item.reviewId ? "有评价证据" : "-")}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${escapeHTML(item.revenueFrozen ? "已冻结" : "未冻结")}</td>
      <td>
        <div class="row-actions">
          <button class="ghost" data-action="report-detail" data-id="${item.id}" type="button">详情</button>
          ${canAssign ? `<button class="ghost" data-action="report-assign" data-id="${item.id}" type="button">指派</button>` : ""}
          ${canHandle ? `<button class="ghost" data-action="report-handle" data-id="${item.id}" type="button">确认处理</button>` : ""}
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
      <td>${escapeHTML(item.id ? `反馈 ${item.id}` : "-")}</td>
      <td>${escapeHTML(userText(item.userId))}</td>
      <td>${escapeHTML(feedbackTypeLabel(item.typeKey || item.type))}</td>
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
  if (canDownload) actions.push(`<button class="ghost" data-action="export-download" data-id="${item.id}" type="button">下载</button>`);
  return `
    <tr>
      <td>${escapeHTML(item.id ? `导出任务 ${item.id}` : "-")}</td>
      <td>${escapeHTML(exportTemplateName(item.templateCode))}</td>
      <td>${escapeHTML(exportTypeLabel(item.exportType))}</td>
      <td>${escapeHTML(item.fileId ? "已生成" : "待生成")}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${escapeHTML(item.fileId ? "已生成" : "-")}</td>
      <td>${formatTime(item.createdAt)}</td>
      <td><div class="row-actions">${actions.join("")}</div></td>
    </tr>
  `;
}

function exportTemplateName(code) {
  const item = (state.exportTemplates || []).find((template) => template.code === code);
  return item ? exportTemplateTitle(item) : exportTypeLabel(code) || adminDisplayValue(code || "-");
}

function exportTemplateTitle(item = {}) {
  const raw = String(item.name || "").trim();
  if (raw && !/[/_]/.test(raw)) return raw;
  return exportTypeLabel(item.exportType || item.code);
}

function exportTypeLabel(value) {
  return {
    reports: "举报申诉报表",
    operation_logs: "操作日志报表",
    reviews: "评价报表",
    reports_default: "举报申诉报表",
    operation_logs_default: "操作日志报表",
    reviews_default: "评价报表",
  }[value] || adminDisplayValue(value);
}

function exportDownloadText(value) {
  if (!value) return "-";
  if (typeof value === "string") return value;
  return value.downloadUrl || value.downloadURL || value.url || "已生成";
}

function settlementMethodLabel(value) {
  return {
    offline: "线下结算",
    bank_transfer: "银行转账",
    wechat_pay: "微信支付",
    manual: "人工登记",
  }[value] || adminDisplayValue(value || "-");
}

function funnelRow(item) {
  return `
    <tr>
      <td>${escapeHTML(behaviorEventLabel(item.eventCode))}</td>
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
      <td>${escapeHTML(item.id ? `行为 ${item.id}` : "-")}</td>
      <td>${escapeHTML(userText(item.userId))}</td>
      <td>${escapeHTML(behaviorTypeLabel(item.eventType))}</td>
      <td>${escapeHTML(behaviorEventLabel(item.eventCode))}</td>
      <td>${escapeHTML(operationTargetText(item))}</td>
      <td>${escapeHTML(sourceLabel(item.source))}</td>
      <td>${formatTime(item.createdAt || item.occurredAt)}</td>
    </tr>
  `;
}

function behaviorLogItem(item) {
  return stackItem({
    title: item.id ? `${behaviorEventLabel(item.eventCode || item.eventType)} ${item.id}` : behaviorEventLabel(item.eventCode || item.eventType),
    badge: behaviorTypeLabel(item.eventType || "behavior"),
    meta: [
      `用户：${userText(item.userId)}`,
      `目标：${operationTargetText(item)}`,
      `页面：${pagePathLabel(item.pagePath)}`,
      `来源：${sourceLabel(item.source)}`,
      `设备：${item.device || "-"}`,
      `关键词：${item.keyword || "-"}`,
      `发生时间：${formatTime(item.occurredAt || item.createdAt)}`,
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
      <td>${escapeHTML(item.id ? `通知 ${item.id}` : "-")}</td>
      <td>${escapeHTML(item.notificationId || "-")}</td>
      <td>${escapeHTML(userText(item.userId))}</td>
      <td>${escapeHTML(notificationSceneLabel(item.scene))}</td>
      <td>${escapeHTML(notificationTemplateLabel(item.templateId))}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${escapeHTML(notificationResultText(item))}</td>
      <td><div class="row-actions">${actions.join("") || "-"}</div></td>
    </tr>
  `;
}

function deliveryDocumentRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.id ? `材料 ${item.id}` : "-")}</td>
      <td>${escapeHTML(deliveryDocTypeLabel(item.docType))}</td>
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
      <td>${escapeHTML(item.id ? `验收用例 ${item.id}` : "-")}</td>
      <td>${escapeHTML(testModuleLabel(item.module))}</td>
      <td>${escapeHTML(item.caseName)}</td>
      <td><span class="${badgeClass(item.priority)}">${escapeHTML(priorityLabel(item.priority))}</span></td>
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
      <td>${escapeHTML(item.id ? `验收记录 ${item.id}` : "-")}</td>
      <td>${escapeHTML(item.caseId ? `验收用例 ${item.caseId}` : "-")}</td>
      <td><span class="${badgeClass(item.result)}">${statusLabel(item.result)}</span></td>
      <td>${escapeHTML(item.requestId ? "已记录" : "-")}</td>
      <td>${escapeHTML(item.evidenceFileId ? "有附件" : "-")} ${evidenceAction}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function sensitiveWordRow(item) {
  const nextStatus = item.status === "active" ? "disabled" : "active";
  return `
    <tr>
      <td>${escapeHTML(item.id ? `敏感词 ${item.id}` : "-")}</td>
      <td>${escapeHTML(item.word)}</td>
      <td>${escapeHTML(sensitiveLevelLabel(item.level))}</td>
      <td>${escapeHTML(sensitiveActionLabel(item.action))}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${formatTime(item.createdAt)}</td>
      <td><button class="ghost" data-action="sensitive-status" data-id="${item.id}" data-status="${nextStatus}" type="button">${statusLabel(nextStatus)}</button></td>
    </tr>
  `;
}

function imMessageRow(item) {
  const canHide = item.status !== "hidden" && can("im:message:hide");
  const fileAction = item.fileId ? `<button class="ghost" data-action="admin-file-download" data-file-id="${escapeHTML(item.fileId)}" type="button">文件</button>` : "";
  const hideAction = canHide ? `<button class="ghost" data-action="im-message-hide" data-id="${escapeHTML(item.id)}" data-room-id="${escapeHTML(item.roomId)}" type="button">隐藏</button>` : "";
  const content = imMessageContentText(item);
  return `
    <tr>
      <td>${escapeHTML(item.id ? `消息 ${item.id}` : "-")}</td>
      <td>${escapeHTML(userText(item.senderUserId))}</td>
      <td>${escapeHTML(messageTypeLabel(item.messageType))}</td>
      <td class="message-content-cell"><div class="message-content-full">${escapeHTML(content)}</div></td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${escapeHTML(item.fileId ? "有附件" : "-")}</td>
      <td>${escapeHTML(compactList(item.ackedBy || [], 4))}</td>
      <td>${escapeHTML(compactList(item.readBy || [], 4))}</td>
      <td>${formatTime(item.createdAt)}</td>
      <td><div class="row-actions">${fileAction}${hideAction}</div></td>
    </tr>
  `;
}

function imMessageContentText(item) {
  const content = String(item.content || "").trim();
  if (content) return content;
  if (item.fileId) return `附件消息，文件 ID：${item.fileId}`;
  return "-";
}

function operationLogRow(item) {
  return `
    <tr>
      <td>${escapeHTML(operationActionLabel(item.action))}</td>
      <td>${escapeHTML(operationTargetText(item))}</td>
      <td>${escapeHTML(adminUserText(item.adminUserId))}</td>
      <td>${formatTime(item.createdAt)}</td>
      <td>
        <div class="row-actions">
          <button class="ghost" data-action="log-detail" data-id="${escapeHTML(item.id)}" type="button">查看摘要</button>
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
    <tr data-roles="${escapeHTML((item.roles || []).join(","))}">
      <td>${escapeHTML(item.id ? `账号 ${item.id}` : "-")}</td>
      <td data-raw-copy>${escapeHTML(item.username)}</td>
      <td>${escapeHTML((item.roles || []).map(adminRoleLabel).join("、") || "-")}</td>
      <td>
        <select data-field="status">
          <option value="active" ${item.status === "active" ? "selected" : ""}>正常</option>
          <option value="disabled" ${item.status === "disabled" ? "selected" : ""}>已禁用</option>
        </select>
      </td>
      <td>${escapeHTML(adminRoleScopeText(item))}</td>
      <td>
        <div class="row-actions">
          ${actions.join("") || "-"}
        </div>
      </td>
    </tr>
  `;
}

function adminApplicationRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.id ? `申请 ${item.id}` : "-")}</td>
      <td>${escapeHTML(item.name || "-")}</td>
      <td>${escapeHTML(item.contact || "-")}</td>
      <td>${escapeHTML(adminRoleLabel(item.desiredRole))}</td>
      <td>${escapeHTML(item.reason || "-")}</td>
      <td>${formatTime(item.createdAt)}</td>
      <td>
        <div class="row-actions">
          <button class="ghost" data-action="copy-admin-application" data-id="${escapeHTML(item.id)}" type="button">复制</button>
        </div>
      </td>
    </tr>
  `;
}

function adminApplicationCopyText(item = {}) {
  return [
    "后台管理员申请",
    `申请编号：${item.id || "-"}`,
    `申请人：${item.name || "-"}`,
    `联系方式：${item.contact || "-"}`,
    `申请角色：${adminRoleLabel(item.desiredRole)}`,
    `角色代码：${item.desiredRole || "-"}`,
    `申请说明：${item.reason || "-"}`,
    `提交时间：${formatTime(item.createdAt)}`,
  ].join("\n");
}

function gameRow(game) {
  const canAudit = game.status === "pending_audit" && can("game:update_status");
  const checked = state.selectedGameAuditIds.has(String(game.id)) ? "checked" : "";
  const actions = [
    `<button class="ghost" data-action="detail" data-id="${game.id}" type="button">详情</button>`,
  ];
  if (canAudit) {
    actions.unshift(`
      <label class="audit-select-inline">
        <input class="game-audit-checkbox" type="checkbox" value="${escapeHTML(game.id)}" aria-label="选择局 ${escapeHTML(game.id)}" ${checked} />
        <span>勾选</span>
      </label>
    `);
    actions.push(`<button class="ghost" data-action="audit" data-id="${game.id}" type="button">通过</button>`);
    actions.push(`<button class="ghost danger" data-action="reject-audit" data-id="${game.id}" type="button">驳回</button>`);
  }
  return `
    <tr>
      <td>${escapeHTML(game.id ? `局 ${game.id}` : "-")}</td>
      <td>${escapeHTML(game.title)}</td>
      <td>${gameTypeLabel(game.gameType)}</td>
      <td>${escapeHTML(sourceLabel(game.gameSource))}</td>
      <td><span class="${badgeClass(game.status)}">${statusLabel(game.status)}</span></td>
      <td>${game.currentPlayers}/${game.minPlayers}-${game.maxPlayers}</td>
      <td>${escapeHTML(game.address || game.cityName || game.cityCode || "-")}</td>
      <td>
        <div class="row-actions">
          ${actions.join("")}
        </div>
      </td>
    </tr>
  `;
}

function gameMilestoneRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.id ? `里程碑 ${item.id}` : "-")}</td>
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
      <td>${escapeHTML(item.id ? `打卡 ${item.id}` : "-")}</td>
      <td>${escapeHTML(userText(item.userId))}</td>
      <td>${escapeHTML(item.milestoneId ? `里程碑 ${item.milestoneId}` : "-")}</td>
      <td>${escapeHTML(checkinTypeLabel(item.checkinType))}</td>
      <td><span class="${badgeClass(item.status)}">${statusLabel(item.status)}</span></td>
      <td>${formatTime(item.createdAt)}</td>
      <td>${canMarkInvalid ? `<button class="ghost" data-action="checkin-invalid" data-id="${item.id}" type="button">标记异常</button>` : "-"}</td>
    </tr>
  `;
}

function gameRetrospectiveRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.id ? `复盘 ${item.id}` : "-")}</td>
      <td>${escapeHTML(userText(item.userId))}</td>
      <td>${escapeHTML(againIntentLabel(item.againIntent))}</td>
      <td>${escapeHTML(item.content || "-")}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function gameContinueDraftRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.originalGameId ? `局 ${item.originalGameId}` : "-")}</td>
      <td>${escapeHTML(item.draftGameId ? `局 ${item.draftGameId}` : "-")}</td>
      <td>${escapeHTML(userText(item.creatorUserId))}</td>
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

function renderDashboardStatusBars(games) {
  const target = $("#game-status-bars");
  if (!target) return;
  const counts = games.reduce((acc, item) => {
    acc[item.status] = (acc[item.status] || 0) + 1;
    return acc;
  }, {});
  const statuses = ["pending_audit", "recruiting", "in_progress", "pending_review", "completed", "canceled"];
  const max = Math.max(1, ...Object.values(counts));
  target.innerHTML = statuses.map((status) => {
    const count = counts[status] || 0;
    return `
      <div class="bar-row">
        <span>${statusLabel(status)}</span>
        <div class="bar-track"><div class="bar-fill" style="width:${Math.round((count / max) * 100)}%"></div></div>
        <strong>${count}</strong>
      </div>
    `;
  }).join("");
}

function renderDashboardWorkQueue(games, imRooms) {
  const target = $("#dashboard-work-queue");
  if (!target) return;
  const pendingGames = games.filter((item) => item.status === "pending_audit");
  const activeGames = games.filter((item) => item.status === "recruiting" || item.status === "in_progress");
  const rooms = Array.isArray(imRooms) ? imRooms : [];
  target.innerHTML = [
    stackItem({
      title: "待审核组局",
      badge: pendingGames.length ? "pending" : "completed",
      meta: [`${pendingGames.length} 个待处理`, "进入组局管理后可批量通过"],
      action: `<button class="ghost" data-view-shortcut="games" type="button">去处理</button>`,
    }),
    stackItem({
      title: "运行中组局",
      badge: activeGames.length ? "active" : "completed",
      meta: [`${activeGames.length} 个正在招募或服务`, "重点关注人数、进度和局内消息"],
      action: `<button class="ghost" data-view-shortcut="games" type="button">查看组局</button>`,
    }),
    stackItem({
      title: "局内消息证据",
      badge: rooms.length ? "active" : "completed",
      meta: [`${rooms.length} 个房间`, "争议处理时从这里进入证据链"],
      action: can("im:room:read") ? `<button class="ghost" data-view-shortcut="im" type="button">查看消息</button>` : "",
    }),
  ].join("");
  target.querySelectorAll("button[data-view-shortcut]").forEach((button) => {
    button.addEventListener("click", () => {
      state.view = button.dataset.viewShortcut;
      setActiveNav();
      loadView(state.view);
    });
  });
}

function renderDashboardPermissionScope() {
  const target = $("#dashboard-permission-scope");
  if (!target) return;
  const menus = state.permissionTree?.menus || [];
  target.innerHTML = menus.length
    ? menus.map((item) => stackItem({
      title: item.name || views[item.code]?.title || item.code,
      badge: "active",
      meta: [`菜单：${views[item.code]?.crumb || item.code}`, `权限点：${compactList(item.permissions || [], 5)}`],
    })).join("")
    : emptyBlock("当前账号暂无后台菜单权限");
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
        <strong>${escapeHTML(title)}：${escapeHTML(trace?.userId ? userText(trace.userId) : trace?.gameId ? `局 ${trace.gameId}` : "-")}</strong>
        <span class="${badgeClass(reviews.length ? "active" : "pending")}">${escapeHTML(reviews.length)} 条评价</span>
      </div>
      <div class="detail-grid profile-detail-grid">
        ${detailCell("成长等级", `Lv.${profile.level || 0}`)}
        ${detailCell("成长经验", profile.experience || 0)}
        ${detailCell("信用分", profile.creditScore || profile.todayCreditScore || 100)}
        ${detailCell("可用积分", profile.availablePoints || 0)}
        ${detailCell("评价数量", profile.reviewCount || reviews.length)}
        ${detailCell("成就数量", achievements.length || (profile.achievements || []).length || 0)}
      </div>
      <div class="sub-panel">
        <h3>评价记录</h3>
        <div class="table-wrap">
          <table>
            <thead><tr><th>评价</th><th>组局</th><th>评价人</th><th>对象</th><th>评分</th><th>再玩意愿</th><th>创建时间</th></tr></thead>
            <tbody>${reviews.map(reviewTraceRow).join("") || emptyRow(7, "暂无评价记录")}</tbody>
          </table>
        </div>
      </div>
      <div class="sub-panel">
        <h3>信用流水</h3>
        <div class="table-wrap">
          <table>
            <thead><tr><th>流水</th><th>用户</th><th>组局</th><th>变动</th><th>变动前/后</th><th>原因</th><th>创建时间</th></tr></thead>
            <tbody>${creditLogs.map(creditLogRow).join("") || emptyRow(7, "暂无信用流水")}</tbody>
          </table>
        </div>
      </div>
      <div class="sub-panel">
        <h3>足迹证据</h3>
        <div class="table-wrap">
          <table>
            <thead><tr><th>用户</th><th>组局</th><th>动作</th><th>创建时间</th></tr></thead>
            <tbody>${footprints.map(footprintRow).join("") || emptyRow(4, "暂无足迹")}</tbody>
          </table>
        </div>
      </div>
      <div class="sub-panel">
        <h3>成就</h3>
        <div class="kv">${achievements.map((item) => `<span>${escapeHTML(item.title || achievementLabel(item.code) || "-")}</span>`).join("") || "<span>暂无成就</span>"}</div>
      </div>
    </article>
  `;
}

function reviewTraceRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.id ? `评价 ${item.id}` : "-")}</td>
      <td>${escapeHTML(item.gameId ? `局 ${item.gameId}` : "-")}</td>
      <td>${escapeHTML(userText(item.reviewerUserId))}</td>
      <td>${escapeHTML(userText(item.targetUserId))} / ${escapeHTML(roleLabel(item.targetRole))}</td>
      <td>${escapeHTML(item.score || 0)}</td>
      <td>${escapeHTML(againIntentLabel(item.againIntent))}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function creditLogRow(item) {
  return `
    <tr>
      <td>${escapeHTML(item.id ? `流水 ${item.id}` : "-")}</td>
      <td>${escapeHTML(userText(item.userId))}</td>
      <td>${escapeHTML(item.gameId ? `局 ${item.gameId}` : "-")}</td>
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
      <td>${escapeHTML(userText(item.userId))}</td>
      <td>${escapeHTML(item.gameId ? `局 ${item.gameId}` : "-")}</td>
      <td>${escapeHTML(footprintActionLabel(item.action))}</td>
      <td>${formatTime(item.createdAt)}</td>
    </tr>
  `;
}

function technicalDetails(summary, body) {
  return "";
}

function detailCell(label, value) {
  return `<div class="detail-cell"><span>${escapeHTML(adminLabel(label))}</span><strong>${escapeHTML(adminDisplayValue(value))}</strong></div>`;
}

function adminLabel(label) {
  const labels = {
    id: "编号",
    ID: "编号",
    code: "邀请码",
    openId: "微信标识",
    nickname: "昵称",
    name: "名称",
    title: "标题",
    description: "说明",
    status: "状态",
    realnameStatus: "实名状态",
    "identity.status": "认证状态",
    inviteCode: "邀请码",
    roles: "角色",
    "growth.level": "成长等级",
    "growth.creditScore": "信用分",
    "points.available": "可用积分",
    "membership.status": "会员状态",
    favorites: "收藏数",
    connections: "人脉数",
    "income.totalCent": "累计收益",
    "income.pendingCent": "待结算收益",
    "income.settledCent": "已结算收益",
    userId: "用户",
    connectedUserId: "关联用户",
    relationType: "关系类型",
    source: "来源",
    sourceType: "来源类型",
    strengthScore: "关系强度",
    gameId: "组局",
    "game.title": "局标题",
    gameType: "局类型",
    gameSource: "创建来源",
    creatorUserId: "创建用户",
    mainGuideUserId: "主领路人",
    milestones: "里程碑数",
    checkins: "打卡数",
    retrospectives: "复盘数",
    continueDrafts: "续局草稿数",
    entryType: "入口类型",
    usedCount: "已使用次数",
    batchCount: "生成数量",
    boundWechatUserId: "绑定用户",
    "bound.nickname": "绑定用户昵称",
    "owner.nickname": "归属用户昵称",
    inviteCodeId: "邀请码",
    inviterUserId: "邀请人",
    inviteeUserId: "被邀请人",
    bindSource: "绑定来源",
    minPlayers: "最少人数",
    maxPlayers: "最多人数",
    currentPlayers: "当前人数",
    cityCode: "城市编码",
    cityName: "城市名称",
    address: "详细地址",
    longitude: "经度",
    latitude: "纬度",
    reason: "原因",
    fileIds: "材料",
    platformBps: "平台比例",
    creatorBps: "发起人比例",
    memberBps: "参与人比例",
    templateId: "方案",
    ruleCode: "规则类型",
    ruleValue: "规则内容",
    recordNo: "记录号",
    recordId: "记录",
    method: "方式",
    proofNo: "凭证号",
    amountCent: "金额",
    memberIds: "参与人",
    expertUserId: "行家",
    guideUserId: "领路人",
    pointsCost: "所需积分",
    stock: "库存",
    orderNo: "订单号",
    itemId: "商品",
    item: "商品",
    items: "明细",
    reviewAdminId: "审核管理员",
    change: "变动",
    changeValue: "变动值",
    "before/after": "变动前/后",
    beforePoints: "变动前积分",
    afterPoints: "变动后积分",
    biz: "业务",
    bizType: "业务类型",
    bizId: "业务对象",
    leaderUserId: "团长",
    membershipPlan: "会员方案",
    period: "周期",
    plan: "方案",
    invited: "邀请数",
    games: "组局数",
    income: "收益",
    invitedCount: "邀请人数",
    incomeSummary: "收益汇总",
    reportType: "举报类型",
    reporterUserId: "举报人",
    targetUserId: "被举报人",
    revenueFrozen: "收益冻结",
    eventCode: "事件编码",
    userCount: "用户数",
    conversionRate: "转化率",
    dropOffRate: "流失率",
    cohortDate: "留存日期",
    newUsers: "新增用户",
    day1: "次日留存",
    day7: "7日留存",
    day30: "30日留存",
    eventType: "事件类型",
    target: "对象",
    targetType: "业务对象",
    scene: "场景",
    templateId: "方案",
    resultCode: "结果码",
    docType: "文档类型",
    caseName: "用例名称",
    priority: "优先级",
    expectedResult: "预期结果",
    actualResult: "实际结果",
    requestId: "请求追踪",
    evidenceFileId: "证据附件",
    username: "账号",
    roles: "角色",
    module: "模块",
    action: "操作",
    adminUserId: "管理员",
    detail: "详情",
    ip: "IP",
    userAgent: "客户端",
    environment: "环境",
    key: "配置项",
    configJson: "规则内容",
    ready: "就绪",
    critical: "关键项",
    message: "说明",
    permissionCount: "职责范围",
    roomId: "房间",
    messageId: "消息",
    senderUserId: "发送人",
    messageType: "消息类型",
    content: "内容",
    payload: "消息内容",
    phone: "手机号",
    realname: "真实姓名",
    faceIdRequestId: "人脸核身请求编号",
    roleCode: "申请角色",
    abilityDescription: "能力说明",
    proofFileIds: "证明材料",
    reviewRemark: "审核备注",
    blockReasons: "阻断原因",
    frozenReason: "冻结原因",
    settledAt: "结算时间",
    settlements: "结算记录数",
    memberCount: "成员数",
    totalCent: "累计金额",
    pendingCent: "待结算金额",
    settledCent: "已结算金额",
    relationLevel: "层级",
    joinedAt: "加入时间",
    skillTree: "技能领域",
    serviceTags: "服务标签",
    caseFileIds: "案例材料",
    completeness: "完整度",
    resourceTags: "资源标签",
    industryTags: "行业标签",
    cityCodes: "覆盖城市",
    connectionScale: "人脉规模",
    engine: "消息引擎",
    openIMGroupId: "消息服务群组",
    messages: "消息数",
    fileMessages: "文件消息数",
    archiveReason: "归档原因",
    sender: "发送人",
    type: "类型",
    fileId: "文件编号",
    acked: "已确认",
    read: "已读",
    role: "角色",
    amountCent: "金额",
    reviewer: "评价人",
    target: "对象",
    score: "评分",
    againIntent: "再玩意愿",
    experience: "经验值",
    availablePoints: "可用积分",
    reviewCount: "评价数",
    achievements: "成就",
    rules: "规则数",
    disabled: "已停用",
    low_review: "低分评价扣分",
    word: "敏感词",
    level: "等级",
    createdAt: "创建时间",
    updatedAt: "更新时间",
  };
  return labels[label] || label;
}

function localizeAdminCopy(root = document) {
  const replacements = adminCopyReplacements();
  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, {
    acceptNode(node) {
      const parent = node.parentElement;
      if (!parent || ["SCRIPT", "STYLE", "TEXTAREA", "CODE"].includes(parent.tagName)) return NodeFilter.FILTER_REJECT;
      if (parent.closest("[data-raw-copy]")) return NodeFilter.FILTER_REJECT;
      return NodeFilter.FILTER_ACCEPT;
    },
  });
  const nodes = [];
  while (walker.nextNode()) nodes.push(walker.currentNode);
  nodes.forEach((node) => {
    let text = node.nodeValue;
    replacements.forEach(([from, to]) => {
      text = text.replace(from, to);
    });
    node.nodeValue = text;
  });
  $$("input[placeholder], textarea[placeholder]", root).forEach((node) => {
    let text = node.getAttribute("placeholder") || "";
    replacements.forEach(([from, to]) => {
      text = text.replace(from, to);
    });
    node.setAttribute("placeholder", text);
  });
}

function polishAdminFragment(root = document) {
  localizeAdminCopy(root);
}

function adminCopyReplacements() {
  return [
    [/\bLocal Player\s+(\d+)\b/g, "玩家 $1"],
    [/\bmock_openid_local-player-(\d+)\b/g, "玩家 $1 微信标识"],
    [/\bmock_openid_local-expert-guide\b/g, "行家领路人微信标识"],
    [/\bmock_openid_local-guide\b/g, "领路人微信标识"],
    [/\bmock_openid_local-expert\b/g, "行家微信标识"],
    [/\bmainGuideUserId\b/g, "主领路人"],
    [/\bcreatorUserId\b/g, "创建用户"],
    [/\bentryType\b/g, "邀请入口"],
    [/\bbatchCount\b/g, "批量生成数量"],
    [/\bboundWechatUserId\b/g, "绑定用户"],
    [/\binviteCodeId\b/g, "邀请码"],
    [/\binviterUserId\b/g, "邀请人"],
    [/\binviteeUserId\b/g, "被邀请人"],
    [/\bbindSource\b/g, "绑定来源"],
    [/\bgameType\b/g, "局类型"],
    [/\bgameSource\b/g, "创建来源"],
    [/\bminPlayers\b/g, "最少人数"],
    [/\bmaxPlayers\b/g, "最多人数"],
    [/\bplatformBps\b/g, "平台比例"],
    [/\bcreatorBps\b/g, "发起人比例"],
    [/\bmemberBps\b/g, "参与人比例"],
    [/\btemplateId\b/g, "方案"],
    [/\bruleCode\b/g, "规则类型"],
    [/\bruleValue\b/g, "规则内容"],
    [/\brecordNo\b/g, "记录号"],
    [/\brecordId\b/g, "记录"],
    [/\bproofNo\b/g, "凭证号"],
    [/\bamountCent\b/g, "金额"],
    [/\bmemberIds\b/g, "参与人"],
    [/\bexpertUserId\b/g, "行家"],
    [/\bguideUserId\b/g, "领路人"],
    [/\bpointsCost\b/g, "所需积分"],
    [/\borderNo\b/g, "订单号"],
    [/\bitemId\b/g, "商品"],
    [/\bitems\b/g, "明细"],
    [/\bitem\b/g, "商品"],
    [/\breviewAdminId\b/g, "审核管理员"],
    [/\buserId\b/g, "用户"],
    [/\bleaderUserId\b/g, "团长"],
    [/\bconnectedUserId\b/g, "关联用户"],
    [/\brelationType\b/g, "关系类型"],
    [/\bsourceType\b/g, "来源类型"],
    [/\bstrengthScore\b/g, "关系强度"],
    [/\bperiod\b/g, "周期"],
    [/\bplan\b/g, "方案"],
    [/\binvited\b/g, "邀请数"],
    [/\bgames\b/g, "组局数"],
    [/\bincome\b/g, "收益"],
    [/\breporterUserId\b/g, "举报人"],
    [/\btargetUserId\b/g, "被举报人"],
    [/\breportType\b/g, "举报类型"],
    [/\brevenueFrozen\b/g, "收益冻结"],
    [/\beventType\b/g, "事件类型"],
    [/\btarget\b/g, "对象"],
    [/\bday1\b/g, "次日留存"],
    [/\bday7\b/g, "7日留存"],
    [/\bday30\b/g, "30日留存"],
    [/\bcreatedAt\b/g, "创建时间"],
    [/\bupdatedAt\b/g, "更新时间"],
    [/\bsuper_admin\b/g, "超级管理员"],
    [/\badmin\b/g, "管理员"],
    [/\bplayer\b/g, "玩家"],
    [/\bexpert\b/g, "行家"],
    [/\bguide\b/g, "领路人"],
    [/\bposter\b/g, "小程序卡片"],
    [/\bqrcode\b/g, "二维码"],
    [/\blink\b/g, "链接"],
    [/\bapp\b/g, "小程序"],
    [/\bmanual\b/g, "手动"],
    [/\bmock\b/g, "初始化数据"],
    [/\bseed\b/g, "初始化数据"],
    [/\bpending_audit\b/g, "待审核"],
    [/\brecruiting\b/g, "招募中"],
    [/\bin_progress\b/g, "进行中"],
    [/\bpending_review\b/g, "待评价"],
    [/\bcompleted\b/g, "已完成"],
    [/\bcanceled\b/g, "已取消"],
    [/\bcancelled\b/g, "已取消"],
    [/\bactive\b/g, "正常"],
    [/\bdisabled\b/g, "已禁用"],
    [/\binactive\b/g, "已下架"],
    [/\binvalid\b/g, "无效"],
    [/\bhidden\b/g, "已隐藏"],
    [/\bfailed\b/g, "失败"],
    [/\brejected\b/g, "已驳回"],
    [/\bapproved\b/g, "已通过"],
    [/\bfulfilled\b/g, "已履约"],
    [/\barchived\b/g, "已归档"],
    [/\bfrozen\b/g, "已冻结"],
    [/\bsettled\b/g, "已结算"],
    [/\bexhausted\b/g, "已用尽"],
    [/\bstandard\b/g, "标准局"],
    [/\bpublic_welfare\b/g, "公益局"],
    [/\bcrowdfund\b/g, "众筹局"],
    [/\bdeposit\b/g, "押金局"],
    [/\bcondition\b/g, "条件局"],
    [/\bfree\b/g, "普通局"],
  ];
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
  if (!value) return "-";
  if (Array.isArray(value)) return value.map(adminDisplayValue).join("、") || "-";
  if (typeof value !== "object") return adminDisplayValue(value);
  return Object.entries(value)
    .filter(([, enabled]) => Boolean(enabled))
    .map(([role]) => adminDisplayValue(role))
    .join("、") || "-";
}

function compactJSON(value) {
  const text = JSON.stringify(value || {});
  return text.length > 72 ? `${text.slice(0, 69)}...` : text;
}

function compactList(values, limit = 8) {
  const list = Array.isArray(values) ? values : [];
  if (list.length <= limit) return list.join("、") || "-";
  return `${list.slice(0, limit).join("、")} 等 ${list.length} 项`;
}

function percentText(value) {
  const number = Number(value || 0);
  if (!Number.isFinite(number)) return "0%";
  return `${Math.round(number * 10000) / 100}%`;
}

function bpsText(value) {
  const number = Number(value || 0);
  if (!Number.isFinite(number)) return "0%";
  return `${Math.round(number) / 100}%`;
}

function percentToBps(value) {
  const number = Number(String(value || "0").replace("%", "").trim());
  if (!Number.isFinite(number) || number < 0) return 0;
  return Math.round(number * 100);
}

function yuanText(value) {
  const number = Number(value || 0);
  if (!Number.isFinite(number)) return "0 元";
  return `${(number / 100).toFixed(2)} 元`;
}

function revenueTemplateName(id) {
  const item = (state.revenueTemplates || []).find((template) => Number(template.id) === Number(id));
  return item ? `${item.name}（${gameTypeLabel(item.gameType)}）` : (id ? `分润方案 ${id}` : "-");
}

function selectRevenueTemplate(templateId) {
  const name = revenueTemplateName(templateId);
  setFormValue($("#revenue-rule-form"), "templateId", templateId);
  setFormValue($("#revenue-calc-form"), "templateId", templateId);
  const ruleDisplay = $("#revenue-rule-template-display");
  if (ruleDisplay) ruleDisplay.value = name;
  const calcDisplay = $("#revenue-calc-template-display");
  if (calcDisplay) calcDisplay.value = name;
}

function revenueRuleLabel(value) {
  return {
    expert_bps: "行家分润比例",
    guide_bps: "领路人分润比例",
    platform_bps: "平台服务比例",
    creator_bps: "发起人比例",
    member_bps: "参与人比例",
    min_amount: "最低结算金额",
    freeze_days: "冻结天数",
  }[value] || adminDisplayValue(value);
}

function guideRuleLabel(value) {
  return {
    default: "默认领路人资格",
    invite_count: "邀请人数要求",
    credit_score: "信用分要求",
    completed_games: "完成局数要求",
    payment_required: "付费门槛",
  }[value] || adminDisplayValue(value);
}

function sensitiveActionLabel(value) {
  return {
    block: "拦截",
    mark: "标记",
    hide: "隐藏",
    review: "待人工复核",
    warn: "提醒",
  }[value] || adminDisplayValue(value);
}

function sensitiveLevelLabel(value) {
  return {
    low: "低风险",
    medium: "中风险",
    high: "高风险",
    critical: "严重风险",
  }[value] || adminDisplayValue(value);
}

function messageTypeLabel(value) {
  return {
    text: "文字消息",
    image: "图片消息",
    file: "文件消息",
    voice: "语音消息",
    system: "系统消息",
  }[value] || adminDisplayValue(value);
}

function footprintActionLabel(value) {
  return {
    checkin: "打卡",
    join: "入局",
    complete: "完成组局",
    review: "提交评价",
    share: "分享",
  }[value] || adminDisplayValue(value);
}

function checkinTypeLabel(value) {
  return {
    location: "位置打卡",
    image: "图片打卡",
    text: "文字打卡",
    manual: "人工记录",
  }[value] || adminDisplayValue(value || "-");
}

function pointsBizText(item = {}) {
  const type = {
    review: "评价奖励",
    redemption: "积分兑换",
    credit_appeal: "信用申诉",
    game_complete: "完成组局",
    invite_bind: "邀请绑定",
    admin_adjust: "后台调整",
  }[item.bizType] || adminDisplayValue(item.bizType || "积分变动");
  return item.bizId ? `${type}（已关联业务）` : type;
}

function feedbackTypeLabel(value) {
  return {
    feature: "功能建议",
    bug: "问题反馈",
    complaint: "投诉",
    experience: "体验优化",
    game: "组局相关",
    address: "地址问题",
  }[value] || adminDisplayValue(value || "-");
}

function checkText(item) {
  if (!item) return "-";
  return `${item.current || 0}/${item.required || 0} ${item.ready ? "已满足" : "待满足"}`;
}

function rolesFromRow(row) {
  if (!row) return [];
  const raw = row.dataset.roles || "";
  return raw.split(/[,，]/).map((item) => adminRoleCode(item.trim())).filter(Boolean);
}

function openAdminDrawer({ title, subtitle = "", body = "" }) {
  if (typeof state.drawerRestore === "function") {
    state.drawerRestore();
    state.drawerRestore = null;
  }
  let overlay = $("#admin-drawer-root");
  if (!overlay) {
    document.body.insertAdjacentHTML("beforeend", `<div id="admin-drawer-root" class="admin-drawer-backdrop hidden"></div>`);
    overlay = $("#admin-drawer-root");
  }
  overlay.classList.remove("hidden");
  overlay.innerHTML = `
    <aside class="admin-drawer" role="dialog" aria-modal="true" aria-label="${escapeHTML(title)}">
      <header class="admin-drawer-head">
        <div>
          <h2>${escapeHTML(title)}</h2>
          ${subtitle ? `<p>${escapeHTML(subtitle)}</p>` : ""}
        </div>
        <button class="drawer-close" data-action="drawer-close" type="button" aria-label="关闭">×</button>
      </header>
      <div class="admin-drawer-body">${body}</div>
    </aside>
  `;
  return $(".admin-drawer-body", overlay);
}

function openEmbeddedFormDrawer(selector, title, subtitle = "") {
  const form = $(selector);
  if (!form || form.dataset.inDrawer === "1") return;
  const placeholder = document.createComment(`${selector}-placeholder`);
  const parent = form.parentNode;
  parent.insertBefore(placeholder, form);
  form.classList.remove("hidden");
  form.dataset.inDrawer = "1";
  const body = openAdminDrawer({ title, subtitle, body: `<div id="drawer-form-host"></div>` });
  $("#drawer-form-host", body).appendChild(form);
  state.drawerRestore = () => {
    if (form.dataset.inDrawer !== "1") return;
    placeholder.replaceWith(form);
    form.classList.add("hidden");
    delete form.dataset.inDrawer;
  };
}

function closeAdminDrawer() {
  const overlay = $("#admin-drawer-root");
  if (!overlay) return;
  if (typeof state.drawerRestore === "function") {
    state.drawerRestore();
    state.drawerRestore = null;
  }
  overlay.classList.add("hidden");
  overlay.innerHTML = "";
}

function resetPagination(key) {
  if (!state.pagination[key]) return;
  state.pagination[key].page = 1;
}

function renderPaginatedTable(selector, items, key, rowRenderer, colspan, emptyText, pageSize = DEFAULT_PAGE_SIZE) {
  const tbody = $(selector);
  if (!tbody) return;
  const list = Array.isArray(items) ? items : [];
  const pageState = state.pagination[key] || { page: 1, pageSize };
  pageState.pageSize = pageSize;
  const totalPages = Math.max(1, Math.ceil(list.length / pageSize));
  pageState.page = Math.min(Math.max(Number(pageState.page) || 1, 1), totalPages);
  state.pagination[key] = pageState;
  const start = (pageState.page - 1) * pageSize;
  const pageItems = list.slice(start, start + pageSize);
  tbody.innerHTML = pageItems.map(rowRenderer).join("") || emptyRow(colspan, emptyText);
  if (key === "games") updateGameAuditSelectionUI();
  renderPager(tbody.closest(".table-wrap"), key, list.length, pageState.page, totalPages, () => {
    renderPaginatedTable(selector, list, key, rowRenderer, colspan, emptyText, pageSize);
  });
}

function renderPaginatedList(selector, items, key, itemRenderer, emptyText, pageSize = 8) {
  const container = $(selector);
  if (!container) return;
  const list = Array.isArray(items) ? items : [];
  const pageState = state.pagination[key] || { page: 1, pageSize };
  pageState.pageSize = pageSize;
  const totalPages = Math.max(1, Math.ceil(list.length / pageSize));
  pageState.page = Math.min(Math.max(Number(pageState.page) || 1, 1), totalPages);
  state.pagination[key] = pageState;
  const start = (pageState.page - 1) * pageSize;
  const pageItems = list.slice(start, start + pageSize);
  container.innerHTML = pageItems.map(itemRenderer).join("") || emptyBlock(emptyText);
  renderPager(container, key, list.length, pageState.page, totalPages, () => {
    renderPaginatedList(selector, list, key, itemRenderer, emptyText, pageSize);
  });
}

function renderPager(anchor, key, total, page, totalPages, rerender) {
  if (!anchor) return;
  const existing = anchor.nextElementSibling;
  if (existing && existing.classList.contains("table-pager") && existing.dataset.pagerKey === key) {
    existing.remove();
  }
  if (total <= state.pagination[key].pageSize) return;
  anchor.insertAdjacentHTML("afterend", `
    <div class="table-pager" data-pager-key="${escapeHTML(key)}">
      <span>共 ${total} 条，当前 ${page} / ${totalPages} 页</span>
      <div class="pager-controls">
        <button type="button" data-page="prev" ${page <= 1 ? "disabled" : ""}>上一页</button>
        <button type="button" data-page="next" ${page >= totalPages ? "disabled" : ""}>下一页</button>
      </div>
    </div>
  `);
  const pager = anchor.nextElementSibling;
  pager.addEventListener("click", (event) => {
    const button = event.target.closest("button[data-page]");
    if (!button || button.disabled) return;
    state.pagination[key].page += button.dataset.page === "prev" ? -1 : 1;
    rerender();
  });
}

function bindSectionTabs(root = document) {
  root.querySelectorAll("[data-tab-target]").forEach((button) => {
    button.addEventListener("click", () => {
      const group = button.dataset.tabGroup || "default";
      root.querySelectorAll(`[data-tab-group="${group}"]`).forEach((item) => item.classList.remove("active"));
      root.querySelectorAll(`[data-tab-panel-group="${group}"]`).forEach((panel) => panel.classList.add("hidden"));
      button.classList.add("active");
      root.querySelector(button.dataset.tabTarget)?.classList.remove("hidden");
    });
  });
}

function setActiveNav() {
  $$("nav button").forEach((button) => button.classList.toggle("active", button.dataset.view === state.view));
}

function applyNavPermissions() {
  $$("nav button[data-view]").forEach((button) => {
    button.hidden = !isViewAllowed(button.dataset.view);
  });
  syncNavGroupVisibility();
}

function navCollapsedGroups() {
  try {
    const values = JSON.parse(localStorage.getItem(NAV_COLLAPSE_STORAGE_KEY) || "[]");
    return new Set(Array.isArray(values) ? values : []);
  } catch (error) {
    return new Set();
  }
}

function saveNavCollapsedGroups(values) {
  localStorage.setItem(NAV_COLLAPSE_STORAGE_KEY, JSON.stringify([...values]));
}

function restoreNavGroups() {
  const collapsed = navCollapsedGroups();
  $$(".nav-section").forEach((section) => {
    const key = section.querySelector("[data-nav-group]")?.dataset.navGroup;
    section.classList.toggle("collapsed", collapsed.has(key));
  });
  expandActiveNavGroup();
  syncNavGroupVisibility();
}

function toggleNavGroup(group) {
  const section = $(`.nav-section [data-nav-group="${group}"]`)?.closest(".nav-section");
  if (!section) return;
  section.classList.toggle("collapsed");
  const collapsed = navCollapsedGroups();
  if (section.classList.contains("collapsed")) {
    collapsed.add(group);
  } else {
    collapsed.delete(group);
  }
  saveNavCollapsedGroups(collapsed);
}

function expandActiveNavGroup() {
  const active = $(`#main-nav button[data-view="${state.view}"]`);
  const section = active?.closest(".nav-section");
  if (!section) return;
  section.classList.remove("collapsed");
  const group = section.querySelector("[data-nav-group]")?.dataset.navGroup;
  if (!group) return;
  const collapsed = navCollapsedGroups();
  collapsed.delete(group);
  saveNavCollapsedGroups(collapsed);
}

function syncNavGroupVisibility() {
  $$(".nav-section").forEach((section) => {
    const visibleItems = $$("button[data-view]", section).filter((button) => !button.hidden);
    section.hidden = visibleItems.length === 0;
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
  if (!ADMIN_VISIBLE_VIEWS.has(view)) return false;
  const menus = state.permissionTree?.menus || [];
  if (!menus.length) return view === "dashboard";
  return menus.some((item) => item.code === view);
}

function setAPIStatus(ok) {
  const el = $("#api-status");
  el.textContent = ok ? "服务正常" : "服务异常";
  el.className = `status-dot ${ok ? "ok" : "pending"}`;
}

function setField(name, value) {
  const el = document.querySelector(`[data-field="${name}"]`);
  if (el) el.textContent = value;
}

function getFormValue(form, name) {
  return form?.querySelector(`[name="${name}"]`)?.value || "";
}

function setFormValue(form, name, value) {
  const input = form?.querySelector(`[name="${name}"]`);
  if (input) input.value = value ?? "";
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
    element.innerHTML = emptyBlock(humanMessage(message || "无权限访问"));
  }
}

function askRejectReason(title = "请输入驳回原因") {
  const value = window.prompt(title);
  const reason = String(value || "").trim();
  if (!reason) {
    toast("拒绝必须填写原因", true);
    return "";
  }
  return reason;
}

function toast(message, isError = false) {
  const el = $("#toast");
  el.textContent = humanMessage(message);
  el.className = `toast ${isError ? "error" : ""}`;
  window.clearTimeout(toast.timer);
  toast.timer = window.setTimeout(clearToast, 3600);
}

function humanMessage(message) {
  const text = String(message || "");
  const noPermission = text.match(/^缺少\s+(.+)$/);
  if (noPermission) return `当前账号无权限：${permissionLabel(noPermission[1])}`;
  if (/^[a-z_]+:[a-z_]+/i.test(text)) return `当前账号无权限：${permissionLabel(text)}`;
  return text;
}

function permissionLabel(code) {
  const labels = {
    "identity:read": "查看实名记录",
    "identity:update": "审核实名记录",
    "role:view": "查看角色申请",
    "role:update": "审核角色申请",
    "game:read": "查看组局",
    "game:view": "查看组局",
    "game:create_admin": "后台创建组局",
    "game:update_status": "审核组局",
    "game:progress:manage": "维护组局进度",
    "user:view": "查看用户",
    "connection:read": "查看画像关系",
    "profile:read": "查看角色资料",
    "revenue:template:view": "查看分润模板",
    "revenue:template:update": "维护分润模板",
    "revenue:simulate": "分润试算",
    "revenue:generate": "生成分润记录",
    "revenue:freeze": "冻结分润记录",
    "settlement:offline:create": "线下结算",
    "redemption:manage": "管理积分兑换",
    "member_report:read": "查看会员报表",
    "team:read": "查看会员团队",
    "system_config:read": "查看运营规则",
    "system_config:update": "维护运营规则",
    "content:risk_log:view": "查看内容风险日志",
    "content:sensitive_word:view": "查看敏感词库",
    "content:sensitive_word:create": "新增敏感词",
    "content:sensitive_word:import": "导入敏感词",
    "content:sensitive_word:update": "维护敏感词",
    "feedback:view": "查看用户反馈",
    "feedback:reply": "回复用户反馈",
    "report:view": "查看举报申诉",
    "report:assign": "分配举报申诉",
    "report:handle": "处理举报申诉",
    "report:close": "关闭举报申诉",
    "notification:wechat:view": "查看微信通知",
    "notification:wechat:send": "发送微信通知",
    "delivery_document:read": "查看交付记录",
    "delivery_document:create": "新增交付记录",
    "test_case:read": "查看验收用例",
    "test_case:create": "新增验收用例",
    "test_run:read": "查看验收记录",
    "test_run:create": "新增验收记录",
    "im:room:read": "查看局内消息房间",
    "im:room:archive": "归档局内消息房间",
    "im:room:retry_create": "重试消息房间同步",
    "admin_user:create": "创建后台账号",
    "admin_user:update": "维护后台账号",
    "ai:data:read": "查看智能数据配置",
    "ai:data:seed": "准备推荐数据",
    "ai:data:export": "导出智能数据",
  };
  return labels[code] || "后台权限";
}

function adminRoleLabel(value) {
  return {
    super_admin: "超级管理员",
    admin: "管理员",
    user_manager: "用户管理员",
    game_manager: "组局管理员",
    operation_manager: "运营管理员",
    finance_manager: "财务管理员",
    customer_manager: "客服管理员",
    data_analyst: "数据分析员",
    audit_manager: "审核管理员",
  }[value] || adminDisplayValue(value || "-");
}

function adminRoleCode(value) {
  const map = {
    超级管理员: "super_admin",
    管理员: "admin",
    用户管理员: "user_manager",
    组局管理员: "game_manager",
    运营管理员: "operation_manager",
    财务管理员: "finance_manager",
    客服管理员: "customer_manager",
    数据分析员: "data_analyst",
    审核管理员: "audit_manager",
  };
  return map[value] || value;
}

function adminRoleDescription(item = {}) {
  const code = item.code || item.name;
  const fallback = {
    super_admin: "拥有全部后台能力，负责账号、配置和最终审计",
    admin: "负责日常运营管理",
    user_manager: "负责用户资料、认证状态、角色资料和邀请关系",
    game_manager: "负责组局审核、局状态、后台开局和入局管理",
    operation_manager: "负责举报申诉、通知、内容安全和日常运营处理",
    finance_manager: "负责分润模板、收益记录、冻结和线下结算登记",
    customer_manager: "负责用户反馈、客服回复和基础用户协助",
    data_analyst: "负责行为、漏斗、留存、报表和智能数据只读分析",
    audit_manager: "负责身份认证记录、角色申请和审核流转",
  }[code] || "按角色职责开放后台功能";
  const description = String(item.description || "").trim();
  if (!description) return fallback;
  if (/[=:,]/.test(description) || /adminCount|permissionCount|permissions|[a-z_]+:[a-z_]+/i.test(description)) {
    return fallback;
  }
  return description;
}

function adminRoleScopeText(item = {}) {
  const roles = Array.isArray(item.roles) ? item.roles : [item.code || item.name].filter(Boolean);
  const descriptions = roles.map((role) => adminRoleDescription({ code: role })).filter((text) => text && text !== "-");
  if (descriptions.length) return compactList(descriptions, 2);
  return "按角色职责开放后台功能";
}

function permissionModuleLabel(value) {
  return {
    admin_user: "管理员账号",
    ai: "智能数据",
    analytics: "数据分析",
    connection: "画像关系",
    content: "内容安全",
    dashboard: "总览",
    export: "导出中心",
    feedback: "用户反馈",
    game: "组局管理",
    identity: "认证记录",
    im: "局内消息",
    invite_code: "邀请管理",
    notification: "通知与验收",
    operation_log: "操作日志",
    redemption: "积分兑换",
    report: "举报申诉",
    revenue: "分润结算",
    role: "角色申请",
    system_config: "运营规则",
    user: "用户管理",
  }[value] || "后台业务";
}

function permissionActionLabel(value) {
  return {
    create: "新增",
    update: "维护",
    view: "查看",
    read: "读取",
    manage: "管理",
    delete: "删除",
    export: "导出",
    assign: "分配",
    handle: "处理",
    approve: "通过",
    reject: "驳回",
    freeze: "冻结",
    settle: "结算",
  }[value] || "操作";
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
    aa: "均摊局",
    crowdfund: "众筹局",
    deposit: "押金局",
    condition: "条件局",
  }[value] || adminDisplayValue(value || "-");
}

function roleLabel(value) {
  return { player: "玩家", expert: "行家", guide: "领路人" }[value] || adminDisplayValue(value || "角色");
}

function sourceLabel(value) {
  return {
    app: "小程序",
    admin: "后台",
    system: "系统",
    seed: "初始化数据",
    mock: "初始化数据",
  }[value] || adminDisplayValue(value || "-");
}

function yesNo(value) {
  return value ? "是" : "否";
}

function applicationAuditModeLabel(value) {
  return {
    auto: "自动审核",
    manual: "人工审核",
    mixed: "按规则审核",
    admin: "后台审核",
  }[value] || adminDisplayValue(value || "-");
}

function visibilityLabel(value) {
  return {
    public: "公开可见",
    hidden: "隐藏",
    admin_only: "仅后台可见",
    member_only: "成员可见",
  }[value] || adminDisplayValue(value || "-");
}

function revenueBlockReasonLabel(value) {
  return {
    no_template: "未选择分润方案",
    invalid_amount: "结算金额不正确",
    missing_creator: "缺少发起人",
    missing_members: "缺少参与人",
    frozen: "记录已冻结",
  }[value] || adminDisplayValue(value || "-");
}

function imEngineLabel(value) {
  return {
    local: "局内消息",
    openim: "消息服务",
    websocket: "实时消息",
  }[value] || adminDisplayValue(value || "-");
}

function aiReadinessSectionLabel(value) {
  return {
    users: "用户数据",
    games: "组局数据",
    behaviorLogs: "行为日志",
    favorites: "收藏数据",
    reviews: "评价数据",
    imMessages: "局内消息",
    footprints: "足迹数据",
  }[value] || adminDisplayValue(value || "-");
}

function statusLabel(value) {
  return {
    draft: "草稿",
    ready: "已就绪",
    empty: "暂无数据",
    pending: "待处理",
    phone_bound: "已绑手机",
    sms_verified: "短信已验证",
    phone_verified: "手机号已核验",
    faceid_processing: "人脸核身中",
    face_verified: "人脸已核验",
    wechat_logged_in: "微信已登录",
    running: "执行中",
    pending_audit: "待审核",
    verified: "已实名",
    rejected: "已驳回",
    recruiting: "招募中",
    full: "已满员",
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
    appealed: "申诉中",
    done: "已完成",
    sent: "已发送",
    failed: "失败",
    create_failed: "创建失败",
    canceled: "已取消",
    cancelled: "已取消",
    invalid: "无效",
    hidden: "已隐藏",
  }[value] || adminDisplayValue(value || "-");
}

function reportTypeLabel(value) {
  return {
    service_dispute: "服务争议",
    im_message: "聊天消息",
    user_complaint: "用户投诉",
    credit_appeal: "信用申诉",
    low_review: "低分评价",
  }[value] || adminDisplayValue(value);
}

function rememberUserDisplayItems(items = []) {
  (Array.isArray(items) ? items : [items]).forEach(rememberUserDisplayItem);
}

function rememberUserDisplayItem(item = {}) {
  const id = Number(item.id ?? item.userId ?? item.userID);
  if (!Number.isFinite(id) || id <= 0) return;
  const hasNickname =
    Object.prototype.hasOwnProperty.call(item, "nickname") ||
    Object.prototype.hasOwnProperty.call(item, "nickName") ||
    Object.prototype.hasOwnProperty.call(item, "displayName");
  const nickname = String(item.nickname || item.nickName || item.displayName || "").trim();
  const existing = state.userDisplayMap.get(id) || {};
  state.userDisplayMap.set(id, {
    id,
    nickname: hasNickname ? nickname : existing.nickname || "",
  });
}

function userText(userID) {
  if (userID === undefined || userID === null || userID === "") return "-";
  const id = Number(userID);
  const cached = Number.isFinite(id) ? state.userDisplayMap.get(id) : null;
  if (cached?.nickname) return cached.nickname;
  return `用户 ${userID}`;
}

function wechatBindingText(openID) {
  if (!openID) return "未绑定微信";
  return "微信用户";
}

function adminUserText(adminUserID) {
  if (adminUserID === undefined || adminUserID === null || adminUserID === "") return "-";
  return `管理员 ${adminUserID}`;
}

function inviterLabel(inviter, inviteRelation) {
  const inviterID = Number(inviter?.id || inviteRelation?.inviterUserId || inviteRelation?.inviter_user_id || 0);
  if (!inviterID) return "未设置";
  const nickname = String(inviter?.nickname || "").trim();
  return nickname ? `用户 ${inviterID} - ${nickname}` : `用户 ${inviterID}`;
}

function inviterPickerValue(inviter, inviteRelation) {
  const inviterID = Number(inviter?.id || inviteRelation?.inviterUserId || inviteRelation?.inviter_user_id || 0);
  if (!inviterID) return "";
  return userPickerOptionText({ id: inviterID, nickname: inviter?.nickname || "" });
}

function inviterPickerID(inviter, inviteRelation) {
  return Number(inviter?.id || inviteRelation?.inviterUserId || inviteRelation?.inviter_user_id || 0) || "";
}

function userPickerOptions(inviter, inviteRelation) {
  const options = Array.from(state.userDisplayMap.values());
  const inviterID = Number(inviter?.id || inviteRelation?.inviterUserId || inviteRelation?.inviter_user_id || 0);
  if (inviterID && !options.some((item) => Number(item.id) === inviterID)) {
    options.unshift({ id: inviterID, nickname: inviter?.nickname || "" });
  }
  return options
    .filter((item) => Number(item.id) > 0)
    .sort((a, b) => Number(a.id) - Number(b.id));
}

function userPickerOptionText(item) {
  const id = Number(item?.id) || 0;
  const nickname = String(item?.nickname || item?.nickName || item?.displayName || "").trim();
  if (id && nickname) return `${id} - ${nickname}`;
  if (id) return String(id);
  return nickname;
}

function userPickerOptionButton(item) {
  const text = userPickerOptionText(item);
  if (!text) return "";
  return `<button class="user-picker-option" type="button" data-user-picker-option data-user-id="${escapeHTML(Number(item?.id) || 0)}" data-user-text="${escapeHTML(text)}">
    <span>${escapeHTML(text)}</span>
  </button>`;
}

function bindUserInvitePicker(form) {
  const picker = form?.querySelector("[data-user-picker]");
  if (!picker) return;
  const input = picker.querySelector("[data-user-picker-input]");
  const valueInput = picker.querySelector("[data-user-picker-value]");
  const empty = picker.querySelector("[data-user-picker-empty]");
  const options = Array.from(picker.querySelectorAll("[data-user-picker-option]"));
  const normalize = (value) => String(value || "").trim().toLowerCase();
  const findMatch = (text) => {
    const keyword = normalize(text);
    if (!keyword) return null;
    return options.find((option) => normalize(option.dataset.userText) === keyword || normalize(option.dataset.userId) === keyword) || null;
  };
  const choose = (option) => {
    if (!option) return;
    input.value = option.dataset.userText || "";
    valueInput.value = option.dataset.userId || "";
    picker.classList.remove("is-open");
  };
  const filter = () => {
    const keyword = normalize(input.value);
    let visibleCount = 0;
    options.forEach((option) => {
      const matched = !keyword || normalize(option.dataset.userText).includes(keyword) || normalize(option.dataset.userId).includes(keyword);
      option.hidden = !matched;
      if (matched) visibleCount += 1;
    });
    if (empty) empty.hidden = visibleCount > 0;
    const exact = findMatch(input.value);
    valueInput.value = exact ? exact.dataset.userId || "" : "";
  };
  input.addEventListener("focus", () => {
    filter();
    picker.classList.add("is-open");
  });
  input.addEventListener("input", () => {
    filter();
    picker.classList.add("is-open");
  });
  input.addEventListener("keydown", (event) => {
    if (event.key !== "Enter") return;
    const firstVisible = options.find((option) => !option.hidden);
    if (!valueInput.value && firstVisible) {
      event.preventDefault();
      choose(firstVisible);
    }
  });
  input.addEventListener("blur", () => {
    setTimeout(() => picker.classList.remove("is-open"), 120);
  });
  options.forEach((option) => {
    option.addEventListener("mousedown", (event) => event.preventDefault());
    option.addEventListener("click", () => choose(option));
  });
  filter();
}

function bindUserInviteRelationForm(userID) {
  const form = $("#user-invite-relation-form");
  if (!form) return;
  bindUserInvitePicker(form);
  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    try {
      const formData = new FormData(form);
      const inviterUserId = Number(formData.get("inviterUserId")) || 0;
      if (!inviterUserId) {
        toast("请选择有效的邀请人", true);
        return;
      }
      await apiPut(`/api/admin/users/${userID}/invite-relation`, { inviterUserId });
      toast("邀请人已更新");
      await Promise.all([showUserDetail(userID), loadUsers()]);
    } catch (error) {
      toast(error.message, true);
    }
  }, { once: true });
}

function memberListText(memberIDs) {
  const ids = Array.isArray(memberIDs) ? memberIDs.filter(Boolean) : [];
  return ids.length ? `${ids.length} 人（${ids.map(userText).join("、")}）` : "-";
}

function relationTypeLabel(value) {
  return {
    invite: "邀请关系",
    friend: "好友关系",
    teammate: "同局关系",
    follow: "关注关系",
    member_team: "会员团队",
  }[value] || adminDisplayValue(value);
}

function againIntentLabel(value) {
  return {
    yes: "愿意再玩",
    maybe: "待考虑",
    no: "不再参与",
    replay: "再玩一局",
  }[value] || adminDisplayValue(value || "-");
}

function behaviorTypeLabel(value) {
  return {
    behavior: "行为记录",
    page_view: "页面访问",
    click: "点击操作",
    submit: "表单提交",
    search: "搜索行为",
    conversion: "转化行为",
  }[value] || adminDisplayValue(value);
}

function behaviorEventLabel(value) {
  return {
    invite_open: "打开邀请入口",
    login_success: "登录成功",
    game_create: "创建组局",
    game_apply: "申请入局",
    game_start: "开始组局",
    im_send: "发送聊天消息",
    review_submit: "提交评价",
    report_submit: "提交举报",
  }[value] || adminDisplayValue(value);
}

function notificationSceneLabel(value) {
  return {
    game_start: "组局开始提醒",
    game_approved: "组局审核通过",
    game_rejected: "组局审核驳回",
    application_approved: "入局申请通过",
    application_rejected: "入局申请驳回",
    review_reminder: "评价提醒",
    report_handled: "举报处理通知",
    credit_changed: "信用变动通知",
  }[value] || adminDisplayValue(value);
}

function notificationTemplateLabel(value) {
  return {
    game_start: "组局开始通知模板",
    game_review: "组局评价通知模板",
    report_result: "举报处理通知模板",
    credit_result: "信用申诉通知模板",
  }[value] || (value ? `通知模板 ${value}` : "-");
}

function notificationResultText(item = {}) {
  if (item.status === "sent" || item.status === "done") return "发送成功";
  if (item.status === "failed") return adminDisplayValue(item.resultMessage || "发送失败");
  if (item.resultCode === "admin" || /admin marked sent/i.test(String(item.resultMessage || ""))) return "后台已确认";
  if (item.resultMessage) return adminDisplayValue(item.resultMessage);
  if (item.resultCode) return statusLabel(item.resultCode);
  return "-";
}

function pagePathLabel(value) {
  const text = String(value || "");
  if (!text) return "-";
  const labels = {
    "pages/entry/index": "入口页",
    "pages/home/index": "首页",
    "pages/game/detail/index": "组局详情",
    "pages/game/create/index": "发起组局",
    "pages/im/room/index": "局内消息",
    "pages/profile/index": "我的",
    "pages/map/index": "地图",
  };
  return labels[text] || (text.startsWith("pages/") ? "小程序页面" : adminDisplayValue(text));
}

function achievementLabel(value) {
  return {
    first_game: "首次完成组局",
    trusted_player: "可信玩家",
    expert_recommended: "行家推荐",
    guide_helper: "领路协助",
  }[value] || adminDisplayValue(value || "");
}

function deliveryDocTypeLabel(value) {
  return {
    privacy: "隐私保护指引",
    user_agreement: "用户协议",
    service_agreement: "服务协议",
    test_report: "验收报告",
    "test-report": "验收报告",
    launch_material: "上线材料",
    "release-record": "发布记录",
    release_record: "发布记录",
    "rollback-plan": "回滚方案",
    rollback_plan: "回滚方案",
    prd: "需求说明归档",
    openapi: "接口资料归档",
    database: "数据说明归档",
    deployment: "部署说明归档",
  }[value] || "归档材料";
}

function testModuleLabel(value) {
  return {
    invite: "邀请注册",
    auth: "登录与短信验证",
    game: "组局流程",
    im: "局内消息",
    report: "举报申诉",
    revenue: "分润结算",
    admin: "后台管理",
    map: "地图能力",
  }[value] || adminDisplayValue(value || "-");
}

function priorityLabel(value) {
  return {
    p0: "阻断",
    p1: "高",
    p2: "中",
    p3: "低",
    high: "高",
    medium: "中",
    low: "低",
  }[String(value || "").toLowerCase()] || value || "-";
}

function operationActionLabel(value) {
  const text = String(value || "");
  const labels = {
    "feedback:view": "查看用户反馈",
    "im:room:read": "查看局内消息房间",
    "invite_code:create": "生成邀请码",
    "invite_code:disable": "禁用邀请码",
    "game:create_admin": "后台创建组局",
    "game:update_status": "更新组局状态",
    "role:update": "审核角色申请",
    "report:handle": "处理举报申诉",
    "revenue:generate": "生成分润记录",
    "revenue:freeze": "冻结分润记录",
    "redemption:manage": "处理积分兑换",
  };
  if (labels[text]) return labels[text];
  return "后台操作";
}

function operationTargetText(item = {}) {
  const target = operationTargetLabel(item.targetType || item.businessType || "对象");
  const id = item.targetId || item.businessId || "-";
  const idText = String(id);
  if (id === "-" || idText.includes("#") || idText.toLowerCase() === "list") return target;
  if (!/^\d+$/.test(idText)) return target;
  return `${target} ${id === "-" ? "" : id}`.trim();
}

function operationTargetLabel(value) {
  return {
    invite_code: "邀请码",
    im_room: "局内消息房间",
    feedback: "用户反馈",
    report: "举报申诉",
    game: "组局",
    user: "用户",
    role: "角色申请",
    revenue: "分润记录",
    redemption: "积分兑换",
    object: "对象",
  }[value] || "业务对象";
}

function operationDetailText(detail) {
  if (!detail || typeof detail !== "object") return "-";
  if (detail.code) return `邀请码：${detail.code}`;
  if (detail.status) return `状态：${statusLabel(detail.status)}`;
  if (detail.total !== undefined) return `数量：${detail.total}`;
  if (detail.count !== undefined) return `数量：${detail.count}`;
  return "已记录";
}

function badgeClass(value) {
  if (["verified", "recruiting", "completed", "approved", "fulfilled", "active", "settled", "handled", "closed", "done"].includes(value)) return "badge success";
  if (["pending", "running", "pending_audit", "pending_review", "phone_bound", "pending_settlement", "assigned", "appealed"].includes(value)) return "badge warning";
  if (["rejected", "failed", "disabled", "inactive", "frozen", "canceled", "cancelled", "invalid", "hidden"].includes(value)) return "badge danger";
  return "badge neutral";
}

function adminDisplayValue(value) {
  if (typeof value === "boolean") return value ? "是" : "否";
  const text = String(value ?? "");
  if (!text) return "-";

  const localPlayer = text.match(/^Local Player\s+(\d+)$/i);
  if (localPlayer) return `玩家 ${localPlayer[1]}`;

  const mockPlayerOpenID = text.match(/^mock_openid_local-player-(\d+)$/i);
  if (mockPlayerOpenID) return `玩家 ${mockPlayerOpenID[1]}微信标识`;

  const known = {
    "mock_openid_local-expert-guide": "行家领路人微信标识",
    "mock_openid_local-guide": "领路人微信标识",
    "mock_openid_local-expert": "行家微信标识",
    super_admin: "超级管理员",
    user_manager: "用户管理员",
    game_manager: "组局管理员",
    operation_manager: "运营管理员",
    finance_manager: "财务管理员",
    customer_manager: "客服管理员",
    admin: "管理员",
    app: "小程序",
    poster: "小程序卡片",
    qrcode: "二维码",
    link: "链接",
    free: "普通局",
    standard: "标准局",
    public_welfare: "公益局",
    aa: "均摊局",
    crowdfund: "众筹局",
    deposit: "押金局",
    condition: "条件局",
    canceled: "已取消",
    cancelled: "已取消",
    active: "正常",
    disabled: "已禁用",
    inactive: "已下架",
    pending: "待处理",
    pending_audit: "待审核",
    pending_review: "待评价",
    pending_settlement: "待结算",
    recruiting: "招募中",
    in_progress: "进行中",
    completed: "已完成",
    approved: "已通过",
    rejected: "已驳回",
    fulfilled: "已履约",
    archived: "已归档",
    frozen: "已冻结",
    settled: "已结算",
    assigned: "已分配",
    handled: "已处理",
    closed: "已关闭",
    appealed: "申诉中",
    failed: "失败",
    done: "已完成",
    invalid: "无效",
    hidden: "已隐藏",
    text: "文字",
    image: "图片",
    file: "文件",
    system: "系统",
    seed: "初始化数据",
    mock: "初始化数据",
    manual: "手动",
    invite: "邀请",
    player: "玩家",
    expert: "行家",
    guide: "领路人",
  };

  if (known[text]) return known[text];
  if (/^mock_|^[a-z0-9]+([_:.][a-z0-9]+)+$/i.test(text)) return "已记录";
  if (/^[a-z][a-z0-9-]*(\s*[/|]\s*[a-z0-9_-]+)+$/i.test(text)) return "已记录";
  return text;
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
