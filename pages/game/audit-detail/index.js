const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const toast = require('../../../utils/toast')
const { getSurnameInitials } = require('../../../utils/avatar')

const DETAIL_SCROLL_TAP_STEP_RPX = 360
const DETAIL_SCROLL_HOLD_STEP_RPX = 72
const DETAIL_SCROLL_HOLD_INTERVAL_MS = 80
const DETAIL_SCROLL_HOLD_SUPPRESS_TAP_MS = 120
const GUIDE_ICON_SRC = '/pages/game/guide-chat/assets/icon-invite.png'
const OPTION_ICON_MAP = {
  time: '/pages/game/audit-detail/assets/option-time.png',
  chat: '/pages/game/audit-detail/assets/option-chat.png'
}

const EMPTY_PLAYER = {
  requirementConfirmed: false,
  requirementStatusText: '',
  confirmed: false,
  statusText: '',
  name: '',
  avatarText: '',
  desc: '',
  tags: [],
  needText: '',
  expectedTime: '',
  remark: ''
}

const EMPTY_GUIDE = {
  iconSrc: GUIDE_ICON_SRC,
  online: false,
  avatarText: '',
  name: '',
  recommendation: ''
}

const EMPTY_GAME_INFO = {
  topic: '',
  time: '',
  location: '',
  activityType: '',
  serviceDuration: '',
  clientBudget: ''
}

function pickFirstValue() {
  const values = Array.prototype.slice.call(arguments)

  for (let index = 0; index < values.length; index += 1) {
    if (values[index] !== undefined && values[index] !== null && values[index] !== '') {
      return values[index]
    }
  }

  return ''
}

function normalizeBoolean(value) {
  if (typeof value === 'boolean') {
    return value
  }

  if (typeof value === 'string') {
    return value === 'true' || value === '1'
  }

  return Boolean(value)
}

function normalizeList(list) {
  return Array.isArray(list) ? list : []
}

function normalizePlayer(player = {}) {
  const name = pickFirstValue(player.name, player.nickname)

  return {
    requirementConfirmed: normalizeBoolean(player.requirementConfirmed),
    requirementStatusText: pickFirstValue(player.requirementStatusText, player.requirementStatus),
    confirmed: normalizeBoolean(player.confirmed),
    statusText: pickFirstValue(player.statusText, player.stateText),
    name,
    avatarText: pickFirstValue(player.avatarText, name ? getSurnameInitials(name, '') : ''),
    desc: pickFirstValue(player.desc, player.description, player.title),
    tags: normalizeList(player.tags),
    needText: pickFirstValue(player.needText, player.requirementText, player.demandText),
    expectedTime: pickFirstValue(player.expectedTime, player.expectedTimeText),
    remark: pickFirstValue(player.remark, player.note)
  }
}

function normalizeGuide(guide = {}) {
  const name = pickFirstValue(guide.name, guide.nickname)

  return {
    iconSrc: pickFirstValue(guide.iconSrc, guide.icon, GUIDE_ICON_SRC),
    online: normalizeBoolean(guide.online),
    avatarText: pickFirstValue(guide.avatarText, name ? getSurnameInitials(name, '') : ''),
    name,
    recommendation: pickFirstValue(guide.recommendation, guide.recommendationText, guide.comment)
  }
}

function normalizeExpert(expert = {}) {
  const name = pickFirstValue(expert.name, expert.nickname)

  return {
    avatarText: pickFirstValue(expert.avatarText, name ? getSurnameInitials(name, '') : ''),
    name,
    roleText: pickFirstValue(expert.roleText, expert.roleName)
  }
}

function normalizeGameInfo(gameInfo = {}) {
  return {
    topic: pickFirstValue(gameInfo.topic, gameInfo.title),
    time: pickFirstValue(gameInfo.time, gameInfo.timeText, gameInfo.startTimeText),
    location: pickFirstValue(gameInfo.location, gameInfo.address, gameInfo.locationText),
    activityType: pickFirstValue(gameInfo.activityType, gameInfo.typeText),
    serviceDuration: pickFirstValue(gameInfo.serviceDuration, gameInfo.durationText),
    clientBudget: pickFirstValue(gameInfo.clientBudget, gameInfo.budgetText, gameInfo.priceText)
  }
}

function buildSessionInfo(gameInfo) {
  return [
    {
      label: '组局主题',
      value: gameInfo.topic,
      iconText: 'H',
      iconClass: 'topic'
    },
    {
      label: '时间',
      value: gameInfo.time,
      iconSrc: '/pages/game/detail/assets/icon-clock.png',
      iconText: '',
      iconClass: 'time'
    },
    {
      label: '地点',
      value: gameInfo.location,
      actionText: gameInfo.location ? '地图位置' : '',
      iconSrc: '/pages/game/detail/assets/icon-location.png',
      iconText: '',
      iconClass: 'place'
    }
  ].filter((item) => item.value)
}

function buildConfirmRows(gameInfo, settlement = {}) {
  return [
    { label: '活动类型', value: gameInfo.activityType },
    { label: '服务时长', value: gameInfo.serviceDuration },
    { label: '客户预算', value: gameInfo.clientBudget, highlight: true, divider: true },
    { label: '平台', value: pickFirstValue(settlement.platformFee, settlement.platformFeeText) },
    { label: '领路人', value: pickFirstValue(settlement.guideReward, settlement.guideRewardText) },
    { label: '生态合伙人', value: pickFirstValue(settlement.partnerReward, settlement.partnerRewardText), divider: true },
    { label: '你的收益', value: pickFirstValue(settlement.expertIncome, settlement.expertIncomeText), success: true, total: true }
  ].filter((item) => item.value)
}

function normalizeStatus(data = {}, guide = {}) {
  const status = data.status || data.statusCard || {}

  return {
    title: pickFirstValue(status.title, data.statusTitle),
    quote: pickFirstValue(status.quote, data.quote),
    guideName: pickFirstValue(status.guideName, guide.name),
    countdown: pickFirstValue(status.countdown, data.countdownText)
  }
}

function normalizeRelation(data = {}, player, guide, expert) {
  const relation = data.relation || {}

  return {
    title: pickFirstValue(relation.title),
    totalCount: Number(pickFirstValue(relation.totalCount, data.totalCount, 0)),
    confirmedCount: Number(pickFirstValue(relation.confirmedCount, data.confirmedCount, 0)),
    confirmedText: pickFirstValue(relation.confirmedText),
    noticeVisible: typeof relation.noticeVisible === 'boolean' ? relation.noticeVisible : undefined,
    noticeText: pickFirstValue(relation.noticeText),
    expert: {
      avatarText: pickFirstValue(relation.expert && relation.expert.avatarText, expert.avatarText),
      name: pickFirstValue(relation.expert && relation.expert.name, expert.name),
      roleText: pickFirstValue(relation.expert && relation.expert.roleText, expert.roleText)
    },
    guide: {
      iconSrc: pickFirstValue(relation.guide && relation.guide.iconSrc, guide.iconSrc),
      online: normalizeBoolean(relation.guide && relation.guide.online || guide.online),
      avatarText: pickFirstValue(relation.guide && relation.guide.avatarText, guide.avatarText),
      name: pickFirstValue(relation.guide && relation.guide.name, guide.name)
    },
    player: {
      avatarText: pickFirstValue(relation.player && relation.player.avatarText, player.avatarText),
      name: pickFirstValue(relation.player && relation.player.name, player.name),
      confirmed: normalizeBoolean(relation.player && relation.player.confirmed || player.confirmed),
      statusText: pickFirstValue(relation.player && relation.player.statusText, player.statusText)
    }
  }
}

function normalizeOptionalActions(data = {}) {
  return normalizeList(data.optionalActions || data.actions)
    .map((item) => {
      const key = pickFirstValue(item.key, item.action)

      return {
        key,
        name: pickFirstValue(item.name, item.title),
        route: pickFirstValue(item.route, item.path),
        message: pickFirstValue(item.message, item.toastText),
        iconSrc: pickFirstValue(item.iconSrc, item.icon, OPTION_ICON_MAP[key])
      }
    })
    .filter((item) => item.key && item.name)
}

function normalizeAuditDetail(data = {}) {
  const player = normalizePlayer(data.player || data.game && data.game.player || {})
  const guide = normalizeGuide(data.guide || data.game && data.game.guide || {})
  const expert = normalizeExpert(data.expert || data.game && data.game.expert || {})
  const gameInfo = normalizeGameInfo(data.info || data.gameInfo || data.game && data.game.info || {})
  const settlement = data.backendSettlement || data.settlement || {}

  return {
    onlineText: data.onlineText || '3999人在线',
    status: normalizeStatus(data, guide),
    relation: normalizeRelation(data, player, guide, expert),
    game: {
      player,
      info: gameInfo,
      guide
    },
    backendSettlement: settlement,
    sessionInfo: buildSessionInfo(gameInfo),
    confirmRows: buildConfirmRows(gameInfo, settlement),
    optionalActions: normalizeOptionalActions(data),
    noticeBullets: normalizeList(data.noticeBullets || data.notices).map((item) => typeof item === 'string' ? item : pickFirstValue(item.text, item.content)).filter(Boolean)
  }
}

Page({
  data: {
    auditId: '',
    actionLoading: false,
    loading: false,
    onlineText: '3999人在线',
    detailScrollTop: 0,
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    status: normalizeStatus({}),
    relation: normalizeRelation({}, EMPTY_PLAYER, EMPTY_GUIDE, {}),
    game: {
      player: EMPTY_PLAYER,
      info: EMPTY_GAME_INFO,
      guide: EMPTY_GUIDE
    },
    backendSettlement: {},
    sessionInfo: [],
    confirmRows: [],
    optionalActions: [],
    noticeBullets: []
  },

  onLoad(options = {}) {
    const auditId = options.auditId || options.id || ''

    this.setData({
      auditId
    })
    this.loadAuditDetail({
      ...options,
      auditId
    })
  },

  async loadAuditDetail(params = {}) {
    this.setData({
      loading: true
    })

    try {
      const data = await gameService.getGameAuditDetail(params)

      this.setData({
        ...normalizeAuditDetail(data),
        loading: false
      })
    } catch (error) {
      this.setData({
        ...normalizeAuditDetail({}),
        loading: false
      })
      toast.info(error.message || '审核详情加载失败')
    }
  },

  handleMapTap() {
    toast.info('地图位置待接入')
  },

  handleUploadImageTap() {
    toast.info('上传图片功能开发中')
  },

  handleUploadFileTap() {
    toast.info('上传文件功能开发中')
  },

  handleOptionalActionTap(event) {
    const key = event.currentTarget.dataset.key
    const item = this.data.optionalActions.find((action) => action.key === key)

    if (item && item.route) {
      wx.navigateTo({
        url: item.route
      })
      return
    }

    toast.info(item && (item.message || item.name) || '操作待接入')
  },

  handleDeclineTap() {
    this.respondAudit('reject')
  },

  handleApproveTap() {
    this.respondAudit('approve')
  },

  async respondAudit(action) {
    if (this.data.actionLoading) {
      return
    }

    this.setData({
      actionLoading: true
    })

    try {
      await gameService.respondGameAudit({
        auditId: this.data.auditId,
        action
      })
      toast.success('已提交审核结果')
      this.loadAuditDetail({
        auditId: this.data.auditId
      })
    } catch (error) {
      toast.info(error.message || '审核处理失败')
    } finally {
      this.setData({
        actionLoading: false
      })
    }
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (key === 'map') {
      wx.showToast({
        title: '地图功能开发中',
        icon: 'none'
      })
      return
    }

    if (key === 'up' || key === 'down') {
      if (!this.suppressNextNavTap) {
        this.scrollDetail(key, DETAIL_SCROLL_TAP_STEP_RPX)
      }
      return
    }

    if (key === 'left' || key === 'right') {
      toast.info('功能正在开发中')
      return
    }

    this.handleShellAction(key)
  },

  handleShellNavLongPress(event) {
    const key = event.detail && event.detail.key

    if (key === 'map') {
      wx.showToast({
        title: '地图功能开发中',
        icon: 'none'
      })
      return
    }

    if (key !== 'up' && key !== 'down') {
      return
    }

    this.suppressNextNavTap = true
    this.stopDetailScrollHold(false)
    this.scrollDetail(key, DETAIL_SCROLL_HOLD_STEP_RPX)

    this.detailScrollHoldTimer = setInterval(() => {
      this.scrollDetail(key, DETAIL_SCROLL_HOLD_STEP_RPX)
    }, DETAIL_SCROLL_HOLD_INTERVAL_MS)
  },

  handleShellNavTouchEnd() {
    this.stopDetailScrollHold(true)
  },

  handleShellAction(key) {
    if (key === 'home') {
      this.scrollDetailToTop()
      return
    }

    if (key === 'search') {
      toast.info('搜索功能开发中')
      return
    }

    if (key === 'comment' || key === 'message') {
      this.navigateToRoute(ROUTES.message)
      return
    }

    if (key === 'avatar' || key === 'mine') {
      this.navigateToRoute(ROUTES.profile)
      return
    }

    const routeMap = {
      metaverse: ROUTES.metaverse,
      map: ''
    }

    this.navigateToRoute(routeMap[key])
  },

  navigateToRoute(route) {
    if (!route || route === ROUTES.gameAuditDetail) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  },

  handleDetailScroll(event) {
    const scrollTop = event.detail && event.detail.scrollTop

    if (typeof scrollTop === 'number') {
      this.detailScrollTopValue = scrollTop
    }
  },

  scrollDetail(direction, stepRpx = DETAIL_SCROLL_TAP_STEP_RPX) {
    const current = Number(this.detailScrollTopValue || this.data.detailScrollTop || 0)
    const distance = this.rpxToPx(stepRpx)
    const nextTop = direction === 'up'
      ? Math.max(0, current - distance)
      : current + distance

    this.detailScrollTopValue = nextTop
    this.setData({
      detailScrollTop: nextTop
    })
  },

  scrollDetailToTop() {
    this.detailScrollTopValue = 0
    this.setData({
      detailScrollTop: 0
    })
  },

  stopDetailScrollHold(resetTapSuppress) {
    if (this.detailScrollHoldTimer) {
      clearInterval(this.detailScrollHoldTimer)
      this.detailScrollHoldTimer = null
    }

    if (resetTapSuppress && this.suppressNextNavTap) {
      if (this.detailScrollSuppressTimer) {
        clearTimeout(this.detailScrollSuppressTimer)
      }

      this.detailScrollSuppressTimer = setTimeout(() => {
        this.suppressNextNavTap = false
        this.detailScrollSuppressTimer = null
      }, DETAIL_SCROLL_HOLD_SUPPRESS_TAP_MS)
    }
  },

  clearDetailScrollTimers() {
    this.stopDetailScrollHold(false)

    if (this.detailScrollSuppressTimer) {
      clearTimeout(this.detailScrollSuppressTimer)
      this.detailScrollSuppressTimer = null
    }

    this.suppressNextNavTap = false
  },

  rpxToPx(value) {
    if (!wx.getSystemInfoSync) {
      return value / 2
    }

    const system = wx.getSystemInfoSync()
    const windowWidth = system && system.windowWidth ? system.windowWidth : 375

    return Math.round((value * windowWidth) / 750)
  },

  onUnload() {
    this.clearDetailScrollTimers()
  }
})
