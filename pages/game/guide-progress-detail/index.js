const gameService = require('../../../services/game')
const { getSurnameInitials } = require('../../../utils/avatar')

const CONTENT_LEFT_RPX = 2
const CONTENT_TOP_RPX = 160
const CONTENT_WIDTH_RPX = 750
const DESIGN_FRAME_HEIGHT_PT = 810
const DESIGN_BOTTOM_HEIGHT_PT = 78
const NAV_TITLE_HEIGHT_RPX = 50
const BACK_BUTTON_SIZE_RPX = 40
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = DESIGN_FRAME_HEIGHT_PT * 2
const ACTION_BAR_HEIGHT_RPX = 142
const REMIND_BELL_ICON = '/pages/game/guide-progress-detail/assets/remind-bell.png'
const PARTICIPANT_CONFIRMED_ICON = '/pages/game/guide-progress-detail/assets/participant-confirmed.png'
const PARTICIPANT_WAITING_ICON = '/pages/game/guide-progress-detail/assets/participant-waiting.png'
const DETAIL_SCROLL_TAP_STEP_RPX = 360
const DETAIL_SCROLL_HOLD_STEP_RPX = 72
const DETAIL_SCROLL_HOLD_INTERVAL_MS = 80
const DETAIL_SCROLL_HOLD_SUPPRESS_TAP_MS = 120
const DEFAULT_DETAIL = {
  id: '',
  status: '',
  statusTitle: '',
  countdownText: '',
  countdownProgressStyle: 'width: 0%;',
  startedAt: '',
  playerConfirmedAt: '',
  playerConfirmedText: '',
  expertConfirmedText: '',
  primaryActionText: '',
  player: {
    name: '',
    roleLabel: '玩家',
    roleClass: 'player',
    desc: '',
    state: '',
    stateClass: '',
    cardClass: '',
    badgeClass: '',
    badgeIcon: '',
    avatarText: '',
    avatarClass: 'pink'
  },
  expert: {
    name: '',
    roleLabel: '行家',
    roleClass: 'expert',
    desc: '',
    state: '',
    stateClass: '',
    cardClass: '',
    badgeClass: '',
    badgeIcon: '',
    avatarText: '',
    avatarClass: 'blue'
  },
  game: {
    topic: '',
    time: '',
    location: ''
  }
}

function createEmptyDetail() {
  return Object.assign({}, DEFAULT_DETAIL, {
    player: Object.assign({}, DEFAULT_DETAIL.player),
    expert: Object.assign({}, DEFAULT_DETAIL.expert),
    game: Object.assign({}, DEFAULT_DETAIL.game),
    statusCard: {
      title: '',
      countdown: '',
      progressStyle: DEFAULT_DETAIL.countdownProgressStyle
    },
    infoRows: [],
    steps: []
  })
}

function roundRpx(value) {
  return Math.round(value * 100) / 100
}

function getCapsuleBottomRpx() {
  try {
    if (typeof wx !== 'undefined' && wx.getSystemInfoSync && wx.getMenuButtonBoundingClientRect) {
      const menuButton = wx.getMenuButtonBoundingClientRect()
      const systemInfo = wx.getSystemInfoSync()

      if (menuButton && systemInfo && systemInfo.windowWidth) {
        return roundRpx((menuButton.top + menuButton.height) * 750 / systemInfo.windowWidth)
      }
    }
  } catch (error) {
    return DEFAULT_CAPSULE_BOTTOM_RPX
  }

  return DEFAULT_CAPSULE_BOTTOM_RPX
}

function getFrameHeightRpx() {
  try {
    if (typeof wx !== 'undefined' && wx.getSystemInfoSync) {
      const systemInfo = wx.getSystemInfoSync()

      if (systemInfo && systemInfo.windowWidth && systemInfo.windowHeight) {
        return roundRpx(systemInfo.windowHeight * 750 / systemInfo.windowWidth)
      }
    }
  } catch (error) {
    return DEFAULT_FRAME_HEIGHT_RPX
  }

  return DEFAULT_FRAME_HEIGHT_RPX
}

function getWhiteShellLayoutStyles() {
  const capsuleBottom = getCapsuleBottomRpx()
  const frameHeight = getFrameHeightRpx()
  const bottomHeight = roundRpx(frameHeight * DESIGN_BOTTOM_HEIGHT_PT / DESIGN_FRAME_HEIGHT_PT)
  const bottomTop = Math.max(CONTENT_TOP_RPX, roundRpx(frameHeight - bottomHeight))
  const contentHeight = Math.max(0, roundRpx(bottomTop - CONTENT_TOP_RPX))
  const actionTop = Math.max(CONTENT_TOP_RPX, roundRpx(bottomTop - ACTION_BAR_HEIGHT_RPX))
  const titleTop = Math.max(0, roundRpx(capsuleBottom - NAV_TITLE_HEIGHT_RPX))
  const backTop = Math.max(0, roundRpx(capsuleBottom - BACK_BUTTON_SIZE_RPX))

  return {
    frameStyle: `height: ${frameHeight}rpx; min-height: ${frameHeight}rpx;`,
    contentStyle: [
      `left: ${CONTENT_LEFT_RPX}rpx`,
      `top: ${CONTENT_TOP_RPX}rpx`,
      `width: ${CONTENT_WIDTH_RPX}rpx`,
      `height: ${contentHeight}rpx`
    ].join('; '),
    bottomStyle: `top: ${bottomTop}rpx; height: ${bottomHeight}rpx;`,
    actionStyle: `top: ${actionTop}rpx; min-height: ${ACTION_BAR_HEIGHT_RPX}rpx;`,
    titleStyle: `top: ${titleTop}rpx; height: ${NAV_TITLE_HEIGHT_RPX}rpx; line-height: ${NAV_TITLE_HEIGHT_RPX}rpx;`,
    backStyle: `top: ${backTop}rpx; width: ${BACK_BUTTON_SIZE_RPX}rpx; height: ${BACK_BUTTON_SIZE_RPX}rpx;`
  }
}

function asArray(value) {
  if (Array.isArray(value)) {
    return value
  }

  if (value) {
    return [value]
  }

  return []
}

function normalizeText(value, fallback = '') {
  if (value === undefined || value === null) {
    return fallback
  }

  const text = String(value).trim()

  return text || fallback
}

function getInitials(name, fallback) {
  return getSurnameInitials(name, fallback)
}

function getParticipantVisualState(statusSignal) {
  const value = normalizeText(statusSignal).toLowerCase()

  if (/waiting|pending|unconfirmed|not_confirmed|unaccepted|待确认|待处理|未确认|等待|未同意/.test(value)) {
    return 'waiting'
  }

  if (/confirmed|accepted|accept|approved|success|done|已确认|已同意|成功/.test(value)) {
    return 'confirmed'
  }

  return 'waiting'
}

function hasParticipantStatusSignal(statusSignal) {
  return /waiting|pending|unconfirmed|not_confirmed|unaccepted|confirmed|accepted|accept|approved|success|done|待确认|待处理|未确认|等待|未同意|已确认|已同意|成功/.test(normalizeText(statusSignal).toLowerCase())
}

function normalizeMember(member = {}, roleType, fallbackMember) {
  const fallback = fallbackMember || {}
  const name = normalizeText(member.name || member.nickname || member.realname, fallback.name || (roleType === 'expert' ? '行家' : '玩家'))
  const desc = normalizeText(member.desc || member.description || member.title || member.subtitle, fallback.desc || '')
  const ownStateText = normalizeText(member.statusText || member.stateText || member.state)
  const fallbackStateText = normalizeText(fallback.statusText || fallback.state)
  const descStateText = hasParticipantStatusSignal(desc) ? desc : ''
  const explicitStatusSignal = [
    member.confirmStatus,
    member.status,
    ownStateText
  ].filter(Boolean).join(' ')
  const statusSignal = hasParticipantStatusSignal(explicitStatusSignal)
    ? explicitStatusSignal
    : [descStateText, fallbackStateText].filter(Boolean).join(' ')
  const visualState = getParticipantVisualState(statusSignal)
  const badgeIcon = member.badgeIcon || member.badgeUrl || member.statusIcon || (visualState === 'confirmed' ? PARTICIPANT_CONFIRMED_ICON : PARTICIPANT_WAITING_ICON)

  return {
    id: member.id || member.userId || fallback.id || roleType,
    name,
    roleLabel: member.roleLabel || member.roleText || fallback.roleLabel || (roleType === 'expert' ? '行家' : '玩家'),
    roleClass: member.roleClass || fallback.roleClass || roleType,
    desc,
    state: normalizeText(ownStateText || fallbackStateText, visualState === 'confirmed' ? '已确认' : '待确认'),
    stateClass: visualState,
    cardClass: visualState,
    badgeClass: visualState,
    badgeIcon,
    avatarUrl: member.avatarUrl || fallback.avatarUrl || '',
    avatarText: getInitials(name, member.avatarText || member.initials || fallback.avatarText || ''),
    avatarClass: member.avatarClass || fallback.avatarClass || (roleType === 'expert' ? 'pink' : 'blue'),
    statusText: normalizeText(ownStateText || fallbackStateText)
  }
}

function getFirstMember(item, roleType) {
  const source = roleType === 'expert'
    ? (item.experts || item.expertList || item.expert || item.expertUser)
    : (item.players || item.playerList || item.player || item.playerUser)

  const directMember = asArray(source)[0]

  if (directMember) {
    return directMember
  }

  const participants = asArray(item.participants || item.members || item.users)

  if (!participants.length) {
    return null
  }

  const matched = participants.find((member) => {
    const role = normalizeText(member && (member.roleType || member.role || member.type)).toLowerCase()

    return roleType === 'expert' ? /expert|guide|行家/.test(role) : /player|玩家/.test(role)
  })

  if (matched) {
    return matched
  }

  return roleType === 'expert' ? participants[1] : participants[0]
}

function getStatusTitle(status, fallback) {
  const value = normalizeText(status).toLowerCase()

  if (/success|completed|done/.test(value)) {
    return '组局成功'
  }

  if (/cancel|reject|decline|refuse|fail|取消|拒绝|婉拒/.test(value)) {
    return '组局已取消'
  }

  if (/player/.test(value)) {
    return '等待玩家确认'
  }

  if (/both|all|双方|waiting_all/.test(value)) {
    return '等待双方确认'
  }

  return fallback || '等待行家确认'
}

function getCountdownProgressStyle(item) {
  const percent = Number(item.countdownProgressPercent || item.timeoutPercent || item.countdownPercent)

  if (!Number.isFinite(percent)) {
    return item.countdownProgressStyle || DEFAULT_DETAIL.countdownProgressStyle
  }

  return `width: ${Math.max(0, Math.min(100, Math.round(percent)))}%;`
}

function getPrimaryActionText(status) {
  const value = normalizeText(status).toLowerCase()

  if (/player/.test(value)) {
    return '提醒玩家'
  }

  if (/success|completed|done/.test(value)) {
    return '查看消息'
  }

  if (/cancel|reject|decline|refuse|fail|取消|拒绝|婉拒/.test(value)) {
    return '重新发起'
  }

  return '提醒行家'
}

function getStepState(index, progressPercent, status) {
  const value = normalizeText(status).toLowerCase()

  if (/success|completed|done/.test(value)) {
    return 'done'
  }

  if (/cancel|reject|decline|refuse|fail|取消|拒绝|婉拒/.test(value)) {
    return index === 0 ? 'done' : 'pending'
  }

  if (/expert/.test(value)) {
    if (index === 0) {
      return 'done'
    }

    return index === 1 ? 'active' : 'pending'
  }

  if (/player/.test(value)) {
    return index === 0 ? 'active' : 'pending'
  }

  if (index === 0 || progressPercent >= 34 && index === 1) {
    return 'done'
  }

  if (index === 1) {
    return 'active'
  }

  return 'pending'
}

function normalizeSteps(item, detail) {
  const percent = Number(item.progressPercent)
  const progressPercent = Number.isFinite(percent) ? percent : 0
  const status = item.status || item.progressStatus || item.state || detail.status
  const isExpertWaiting = /expert/.test(normalizeText(status).toLowerCase())

  return [
    {
      key: 'launch',
      title: '发起引荐',
      desc: normalizeText(item.launchDesc || item.inviteDesc, '你向双方发送了组局邀请'),
      timeText: normalizeText(item.startedAt || item.createdAt || detail.startedAt),
      state: getStepState(0, progressPercent, status),
      hasLine: true,
      lineState: 'confirmed'
    },
    {
      key: 'player',
      title: '玩家已确认',
      desc: normalizeText(item.playerConfirmedText || detail.player.statusText || detail.playerConfirmedText),
      timeText: normalizeText(item.playerConfirmedAt || detail.playerConfirmedAt),
      state: 'confirmed',
      emphasis: true,
      hasLine: true,
      lineState: isExpertWaiting ? 'pending' : 'confirmed'
    },
    {
      key: 'expert',
      title: '等待行家确认',
      desc: normalizeText(item.expertWaitingText || item.expertConfirmedText || detail.expert.statusText || detail.expertConfirmedText, '已发送提醒'),
      timeText: isExpertWaiting ? '待处理' : normalizeText(item.expertConfirmedAt),
      state: isExpertWaiting ? 'active' : getStepState(2, progressPercent, status),
      actionText: isExpertWaiting ? '再次提醒' : '',
      actionKey: 'remindExpert',
      iconSrc: item.remindIconSrc || REMIND_BELL_ICON,
      hasLine: true,
      lineState: /success|completed|done/.test(normalizeText(status).toLowerCase()) ? 'confirmed' : 'pending'
    },
    {
      key: 'success',
      title: '组局成功',
      desc: '双方确认后自动成局',
      timeText: '',
      state: /success|completed|done/.test(normalizeText(status).toLowerCase()) ? 'confirmed' : 'pending',
      hasLine: false,
      lineState: 'pending'
    }
  ]
}

function normalizeDetail(item = {}) {
  if (!item || Object.keys(item).length === 0) {
    return createEmptyDetail()
  }

  const source = Object.assign({}, DEFAULT_DETAIL, item || {})
  const player = normalizeMember(getFirstMember(source, 'player') || {}, 'player', DEFAULT_DETAIL.player)
  const expert = normalizeMember(getFirstMember(source, 'expert') || {}, 'expert', DEFAULT_DETAIL.expert)
  const game = Object.assign({}, DEFAULT_DETAIL.game, source.game || source.gameInfo || {
    topic: source.topic || source.title || source.gameTitle,
    time: source.gameTime || source.appointmentTime || source.startTime,
    location: source.location || source.address
  })
  const status = source.status || source.progressStatus || source.state || DEFAULT_DETAIL.status
  const statusTitle = getStatusTitle(status, source.statusTitle || source.progressText)
  const countdownText = normalizeText(source.countdownText || source.remainingText, DEFAULT_DETAIL.countdownText)
  const detail = {
    id: source.id || source.invitationId || DEFAULT_DETAIL.id,
    status,
    statusTitle,
    countdownText,
    statusCard: {
      title: statusTitle,
      countdown: countdownText,
      progressStyle: getCountdownProgressStyle(source)
    },
    primaryActionText: source.primaryActionText || getPrimaryActionText(status),
    player,
    expert,
    game,
    playerConfirmedText: normalizeText(source.playerConfirmedText),
    playerConfirmedAt: normalizeText(source.playerConfirmedAt),
    expertConfirmedText: normalizeText(source.expertConfirmedText),
    startedAt: normalizeText(source.startedAt),
    infoRows: [
      { label: '主题', value: normalizeText(game.topic, DEFAULT_DETAIL.game.topic) },
      { label: '时间', value: normalizeText(game.time, DEFAULT_DETAIL.game.time) },
      { label: '地点', value: normalizeText(game.location, DEFAULT_DETAIL.game.location) }
    ]
  }

  detail.steps = normalizeSteps(source, detail)

  return detail
}

function findProgressItem(data = {}, id) {
  const activeList = asArray(data.activeParties || data.activeList || data.ongoingList || data.processingList)
  const completedList = asArray(data.completedParties || data.completedList || data.recentCompleted || data.historyList)
  const list = activeList.concat(completedList)

  if (!id) {
    return list[0]
  }

  return list.find((item) => String(item.id || item.invitationId || '') === String(id)) || list[0]
}

Page({
  data: {
    pageTitle: '组局进度详情',
    shellLayout: getWhiteShellLayoutStyles(),
    queryParams: {},
    detailScrollTop: 0,
    loading: false,
    errorText: '',
    detail: createEmptyDetail()
  },

  onLoad(options = {}) {
    this.setData({
      queryParams: options,
      shellLayout: getWhiteShellLayoutStyles()
    })
    this.loadProgressDetail(options)
  },

  onShow() {
    this.updateShellLayout()
  },

  onResize() {
    this.updateShellLayout()
  },

  updateShellLayout() {
    this.setData({
      shellLayout: getWhiteShellLayoutStyles()
    })
  },

  async loadProgressDetail(params = {}) {
    this.setData({
      loading: true,
      errorText: ''
    })

    try {
      const data = await gameService.getGuideProgress(params)
      const item = findProgressItem(data, params.id || params.invitationId)

      this.setData({
        loading: false,
        errorText: '',
        detail: item ? normalizeDetail(item) : createEmptyDetail()
      })
    } catch (error) {
      this.setData({
        loading: false,
        errorText: error && error.message ? error.message : '组局进度详情加载失败'
      })
    }
  },

  onRetryTap() {
    this.loadProgressDetail(this.data.queryParams)
  },

  onMapTap() {
    this.showToast('地图位置待接入')
  },

  onCancelTap() {
    this.showToast('取消组局页待接入')
  },

  onRemindTap() {
    this.showToast('提醒行家待接入')
  },

  onTimelineActionTap(event) {
    const key = event.currentTarget.dataset.key

    if (key === 'remindExpert') {
      this.showToast('提醒行家待接入')
      return
    }

    this.onRemindTap()
  },

  onDetailScroll(event) {
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

  showToast(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  },

  onBackTap() {
    const pages = getCurrentPages()

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    wx.navigateTo({
      url: '/pages/game/guide-progress/index'
    })
  },

  onUnload() {
    this.clearDetailScrollTimers()
  }
})
