const gameService = require('../../../services/game')
const userService = require('../../../services/user')
const { navigateShellRoute } = require('../../../utils/shell-nav')
const { getViewportMetrics } = require('../../../utils/viewport-layout')

const NAV_BOTTOM_GAP_RPX = 16
const DEFAULT_NAV_HEIGHT_RPX = 160

function getNavigationLayout() {
  const metrics = getViewportMetrics()
  if (!metrics) {
    return {
      navStyle: `height:${DEFAULT_NAV_HEIGHT_RPX}rpx`,
      titleStyle: 'top:80rpx;height:64rpx;line-height:64rpx',
      backStyle: 'top:86rpx;width:52rpx;height:52rpx'
    }
  }
  const navHeight = metrics.capsuleBottomRpx + NAV_BOTTOM_GAP_RPX
  const backSize = Math.min(52, metrics.capsuleHeightRpx)
  const backTop = metrics.capsuleTopRpx + (metrics.capsuleHeightRpx - backSize) / 2
  return {
    navStyle: `height:${navHeight}rpx`,
    titleStyle: `top:${metrics.capsuleTopRpx}rpx;height:${metrics.capsuleHeightRpx}rpx;line-height:${metrics.capsuleHeightRpx}rpx`,
    backStyle: `top:${backTop}rpx;width:${backSize}rpx;height:${backSize}rpx`
  }
}

const ROLE_TABS = [
  { key: 'player', label: '玩家' },
  { key: 'guide', label: '领路人' },
  { key: 'expert', label: '行家' }
]
const STATUS_TABS_BY_ROLE = {
  player: [
    { key: 'active', label: '进行中' },
    { key: 'complete', label: '已完成' },
    { key: 'cancelled', label: '退款/取消' }
  ],
  guide: [
    { key: 'active', label: '进行中' },
    { key: 'complete', label: '已完成' },
    { key: 'cancelled', label: '已取消' }
  ],
  expert: [
    { key: 'active', label: '进行中' },
    { key: 'complete', label: '已完成' },
    { key: 'dispute', label: '争议' }
  ]
}

function listOf(data, keys) {
  for (const key of keys) {
    if (data && Array.isArray(data[key])) return data[key]
  }
  return []
}

function statusKey(item = {}) {
  const value = String(item.statusType || item.state || item.status || '').toLowerCase()
  if (/review|评价/.test(value)) return 'complete'
  if (/complete|finished|done|已完成/.test(value)) return 'complete'
  if (/dispute|争议/.test(value)) return 'dispute'
  if (/cancel|refund|退款|取消/.test(value)) return 'cancelled'
  return 'active'
}

function normalize(item, role, index) {
  const gameId = item.gameId || item.gameID || item.sourceGameId || ''
  const id = item.id || item.orderId || item.recordId || `${role}-${index}`
  const roleMeta = {
    expert: { label: '行家', tone: 'blue', route: '/pages/game/manage/index' },
    guide: { label: '领路人', tone: 'orange', route: '/pages/game/referral-record/index' },
    player: { label: '玩家', tone: 'purple', route: '/pages/game/player-manage/index' }
  }[role]
  const status = statusKey(item)
  const expert = item.expert || {}
  const player = item.player || {}
  const guide = item.guide || {}
  return {
    id: `${role}-${id}`,
    rawId: id,
    gameId,
    role,
    roleLabel: roleMeta.label,
    roleTone: roleMeta.tone,
    title: item.title || item.serviceTitle || item.gameTitle || item.name || '组局服务',
    person: item.name || item.playerName || item.expertName || item.targetName || '',
    expertName: expert.name || item.expertName || '',
    expertAvatar: expert.avatarText || '行',
    playerName: player.name || item.playerName || '',
    playerAvatar: player.avatarText || '玩',
    guideName: guide.name || item.guideName || '',
    guideAvatar: guide.avatarText || '领',
    amount: item.amount || item.amountText || item.priceText || '',
    status,
    reviewPending: item.reviewStatus === 'pending' || Boolean(item.canReview),
    statusText: item.statusText || ({ active: '进行中', complete: '已完成', cancelled: '已取消', review: '待评价' }[status]),
    progressText: item.progressText || item.currentStageText || item.timeText || '',
    noticeText: item.noticeText || (item.expectedDeliveryAt ? `预计交付：${String(item.expectedDeliveryAt).slice(0, 10)}` : ''),
    primaryText: role === 'expert' ? '管理交付' : (role === 'guide' ? '跟进双方' : '查看服务'),
    secondaryText: role === 'expert' ? '联系玩家' : (role === 'guide' ? '查看群聊' : '联系行家'),
    route: roleMeta.route
  }
}

function availableRoleTabs(user = {}) {
  const roles = Array.isArray(user.roles) ? user.roles : []
  const statusMap = user.roleStatusMap || {}
  const enabled = (key) => key === 'player' || roles.includes(key) || statusMap[key] === 'active' || statusMap[key] === 'approved'

  return ROLE_TABS.filter((item) => enabled(item.key))
}

Page({
  data: {
    navigationLayout: getNavigationLayout(),
    loading: true,
    errorText: '',
    roleTabs: [ROLE_TABS[0]],
    statusTabs: STATUS_TABS_BY_ROLE.player,
    activeRole: 'player',
    activeStatus: 'active',
    allItems: [],
    visibleItems: [],
    roleSummaries: {},
    currentSummary: {}
  },

  onLoad() { this.loadData() },
  onShow() {
    this.setData({ navigationLayout: getNavigationLayout() })
    if (this.loaded) this.loadData()
  },
  onResize() { this.setData({ navigationLayout: getNavigationLayout() }) },

  onBackTap() {
    const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []
    if (pages.length > 1) {
      wx.navigateBack()
      return
    }
    navigateShellRoute('/pages/profile/index')
  },

  async loadData() {
    this.setData({ loading: true, errorText: '' })
    try {
      const user = await userService.getCurrentUser()
      const roleTabs = availableRoleTabs(user)
      const enabledRoles = roleTabs.map((item) => item.key)
      const [playerData, guideData, expertData] = await Promise.all([
        gameService.getPlayerGameManage(),
        enabledRoles.includes('guide') ? gameService.getReferralRecords() : Promise.resolve({ records: [] }),
        enabledRoles.includes('expert') ? gameService.getGameManage() : Promise.resolve({ orders: [] })
      ])
      const items = []
        .concat(listOf(playerData, ['orders', 'items']).map((item, index) => normalize(item, 'player', index)))
        .concat(listOf(guideData, ['records', 'items']).map((item, index) => normalize(item, 'guide', index)))
        .concat(listOf(expertData, ['orders', 'items']).map((item, index) => normalize(item, 'expert', index)))
      this.loaded = true
      this.setData({
        loading: false,
        roleTabs,
        activeRole: roleTabs.some((item) => item.key === this.data.activeRole) ? this.data.activeRole : roleTabs[0].key,
        allItems: items,
        roleSummaries: {
          player: playerData.summary || {},
          guide: guideData.summary || {},
          expert: expertData.summary || {}
        }
      })
      this.applyFilters()
    } catch (error) {
      this.setData({ loading: false, errorText: error.message || '组局管理加载失败' })
    }
  },

  onRoleTap(event) {
    const activeRole = event.currentTarget.dataset.key
    this.setData({
      activeRole,
      activeStatus: 'active',
      statusTabs: STATUS_TABS_BY_ROLE[activeRole] || [],
      currentSummary: this.data.roleSummaries[activeRole] || {}
    })
    this.applyFilters()
  },
  onStatusTap(event) {
    this.setData({ activeStatus: event.currentTarget.dataset.key })
    this.applyFilters()
  },
  applyFilters() {
    const { allItems, activeRole, activeStatus } = this.data
    const roleItems = allItems.filter((item) => item.role === activeRole)
    const statusTabs = (STATUS_TABS_BY_ROLE[activeRole] || []).map((tab) => ({
      ...tab,
      count: roleItems.filter((item) => item.status === tab.key).length,
      displayLabel: `${tab.label}(${roleItems.filter((item) => item.status === tab.key).length})`
    }))
    this.setData({
      statusTabs,
      currentSummary: this.data.roleSummaries[activeRole] || {},
      visibleItems: roleItems.filter((item) => item.status === activeStatus)
    })
  },
  onItemTap(event) {
    const item = this.data.allItems.find((row) => row.id === event.currentTarget.dataset.id)
    if (!item) return
    const query = [item.gameId ? `gameId=${encodeURIComponent(item.gameId)}` : '', item.rawId ? `id=${encodeURIComponent(item.rawId)}` : ''].filter(Boolean).join('&')
    navigateShellRoute(`${item.route}${query ? `?${query}` : ''}`)
  }
})
