const { getLegacyWhiteFrameLayoutStyles } = require('../../../utils/adaptive-shell-layout')
const gameService = require('../../../services/game')
const { navigateShellRoute } = require('../../../utils/shell-nav')

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
const REMIND_BELL_ICON = 'https://static.haowan.net.cn/miniprogram/pages/game/referral-record/assets/action-bell-blue.svg'
const PARTICIPANT_CONFIRMED_ICON = 'https://static.haowan.net.cn/miniprogram/components/game-detail/party-card/icon-confirmed.svg'
const PARTICIPANT_WAITING_ICON = 'https://static.haowan.net.cn/miniprogram/pages/game/guide-progress-detail/assets/participant-waiting.svg'
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
    roleLabel: '',
    roleClass: '',
    desc: '',
    state: '',
    stateClass: '',
    cardClass: '',
    badgeClass: '',
    badgeIcon: '',
    avatarText: '',
    avatarClass: ''
  },
  expert: {
    name: '',
    roleLabel: '',
    roleClass: '',
    desc: '',
    state: '',
    stateClass: '',
    cardClass: '',
    badgeClass: '',
    badgeIcon: '',
    avatarText: '',
    avatarClass: ''
  },
  game: {
    topic: '',
    time: '',
    location: ''
  }
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
  const layout = getLegacyWhiteFrameLayoutStyles({
    contentLeftRpx: CONTENT_LEFT_RPX,
    contentWidthRpx: CONTENT_WIDTH_RPX,
    bottomHeightRatio: DESIGN_BOTTOM_HEIGHT_PT / DESIGN_FRAME_HEIGHT_PT,
    titleHeightRpx: NAV_TITLE_HEIGHT_RPX,
    backSizeRpx: BACK_BUTTON_SIZE_RPX
  })
  const actionTop = Math.max(layout.contentTopRpx, roundRpx(layout.bottomTopRpx - ACTION_BAR_HEIGHT_RPX))

  return Object.assign({}, layout, {
    actionStyle: `top: ${actionTop}rpx; min-height: ${ACTION_BAR_HEIGHT_RPX}rpx;`
  })
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

function getParticipantVisualState(statusSignal) {
  const value = normalizeText(statusSignal).toLowerCase()

  if (/waiting|pending|unconfirmed|not_confirmed|unaccepted|待确认|待处理|未确认|等待|未同意/.test(value)) {
    return 'waiting'
  }

  if (/confirmed|accepted|accept|approved|success|done|已确认|已同意|成功/.test(value)) {
    return 'confirmed'
  }

  if (/rejected|reject|declined|refused|canceled|cancel/.test(value)) {
    return 'canceled'
  }

  return 'waiting'
}

function hasParticipantStatusSignal(statusSignal) {
  return /waiting|pending|unconfirmed|not_confirmed|unaccepted|confirmed|accepted|accept|approved|success|done|待确认|待处理|未确认|等待|未同意|已确认|已同意|成功/.test(normalizeText(statusSignal).toLowerCase())
}

function normalizeMember(member = {}, roleType, fallbackMember) {
  const fallback = fallbackMember || {}
  const name = normalizeText(member.name || member.nickname || member.realname, fallback.name || '')
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
  const visualState = normalizeText(member.stateClass || member.cardClass, getParticipantVisualState(statusSignal))
  const badgeIcon = member.badgeIcon || member.badgeUrl || member.statusIcon || member.icon || (
    visualState === 'confirmed'
      ? PARTICIPANT_CONFIRMED_ICON
      : (visualState === 'waiting' ? PARTICIPANT_WAITING_ICON : '')
  )

  return {
    id: member.id || member.userId || fallback.id || roleType,
    name,
    roleLabel: member.roleLabel || member.roleText || fallback.roleLabel || '',
    roleClass: member.roleClass || fallback.roleClass || roleType,
    desc,
    state: normalizeText(ownStateText || fallbackStateText),
    stateClass: visualState,
    cardClass: visualState,
    badgeClass: visualState,
    badgeIcon,
    avatarUrl: member.avatarUrl || fallback.avatarUrl || '',
    avatarText: member.avatarText || member.avatar || member.initials || fallback.avatarText || '',
    avatarType: member.avatarType || fallback.avatarType || '',
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
  return normalizeText(fallback)
}

function getCountdownProgressStyle(item) {
  const percent = Number(item.countdownProgressPercent || item.timeoutPercent || item.countdownPercent)

  if (!Number.isFinite(percent)) {
    return item.countdownProgressStyle || DEFAULT_DETAIL.countdownProgressStyle
  }

  return `width: ${Math.max(0, Math.min(100, Math.round(percent)))}%;`
}

function finiteNumber(value, fallback = 0) {
  const number = Number(value)

  return Number.isFinite(number) ? number : fallback
}

function parseTimeValue(value) {
  const text = normalizeText(value)

  if (!text) {
    return 0
  }

  const parsed = Date.parse(text)

  return Number.isFinite(parsed) ? parsed : 0
}

function twoDigit(value) {
  const number = Math.max(0, Math.floor(value))

  return number < 10 ? `0${number}` : String(number)
}

function formatCountdownText(seconds) {
  const total = Math.max(0, Math.ceil(finiteNumber(seconds, 0)))
  const hours = Math.floor(total / 3600)
  const minutes = Math.floor((total % 3600) / 60)
  const remainSeconds = total % 60

  return `${twoDigit(hours)}:${twoDigit(minutes)}:${twoDigit(remainSeconds)}`
}

function getLiveCountdown(detail = {}, now = Date.now()) {
  if (normalizeText(detail.status).toLowerCase() !== 'pending') {
    return null
  }

  const timeoutAt = normalizeText(detail.timeoutAt || detail.expiresAt)
  const timeoutAtMs = parseTimeValue(timeoutAt)
  const timeoutSeconds = finiteNumber(detail.timeoutSeconds || detail.timeLimitSeconds, 0)

  if (!timeoutAtMs || timeoutSeconds <= 0) {
    return null
  }

  const remainingSeconds = Math.max(0, Math.ceil((timeoutAtMs - now) / 1000))
  const percent = Math.max(0, Math.min(100, Math.round((remainingSeconds * 100) / timeoutSeconds)))

  return {
    timeoutAt,
    timeoutSeconds,
    remainingSeconds,
    countdownProgressPercent: percent,
    countdownText: formatCountdownText(remainingSeconds),
    countdownProgressStyle: `width: ${percent}%;`
  }
}

function normalizeGameInfo(source = {}) {
  const game = source.gameInfo || source.game || {}

  return {
    topic: normalizeText(game.topic || game.title || game.gameTitle || source.gameTitle || source.topic),
    time: normalizeText(game.time || game.timeText || game.scheduleText || source.gameTimeText || source.gameTime || source.timeText),
    location: normalizeText(game.location || game.locationText || game.address || source.locationText || source.location)
  }
}

function normalizeInfoRows(source = {}) {
  return asArray(source.infoRows || source.gameInfoRows || source.sessionInfoRows)
    .map((row) => ({
      label: normalizeText(row && (row.label || row.title || row.key)),
      value: normalizeText(row && (row.value || row.text || row.content))
    }))
    .filter((row) => row.label)
}

function normalizeBackendSteps(source = {}) {
  const rawSteps = asArray(source.steps || source.timeline || source.progressSteps)

  return rawSteps
    .map((step, index) => ({
      key: normalizeText(step && (step.key || step.id), `step-${index}`),
      title: normalizeText(step && (step.title || step.name)),
      desc: normalizeText(step && (step.desc || step.description || step.text)),
      timeText: normalizeText(step && (step.timeText || step.time)),
      state: normalizeText(step && (step.state || step.status), 'pending'),
      emphasis: Boolean(step && step.emphasis),
      actionText: normalizeText(step && (step.actionText || step.action)),
      actionKey: normalizeText(step && (step.actionKey || step.key || step.id)),
      iconText: normalizeText(step && step.iconText),
      iconSrc: normalizeText(step && (step.iconSrc || step.icon), step && step.actionText ? REMIND_BELL_ICON : ''),
      hasLine: step && step.hasLine !== undefined ? Boolean(step.hasLine) : index < rawSteps.length - 1,
      lineState: normalizeText(step && (step.lineState || step.lineStatus || step.state || step.status), 'pending')
    }))
    .filter((step) => step.title)
}

function normalizeDetail(item = {}) {
  const source = Object.assign({}, DEFAULT_DETAIL, item)
  const player = normalizeMember(getFirstMember(source, 'player') || {}, 'player', DEFAULT_DETAIL.player)
  const expert = normalizeMember(getFirstMember(source, 'expert') || {}, 'expert', DEFAULT_DETAIL.expert)
  const game = normalizeGameInfo(source)
  const status = source.status || source.progressStatus || source.state || DEFAULT_DETAIL.status
  const statusTitle = getStatusTitle(status, source.statusTitle || source.title || source.progressText)
  const countdownText = normalizeText(source.countdownText || source.remainingText, DEFAULT_DETAIL.countdownText)
  const timeoutSeconds = finiteNumber(source.timeoutSeconds || source.timeLimitSeconds, 0)
  const timeoutAt = normalizeText(source.timeoutAt || source.expiresAt || source.deadlineAt)
  const remainingSeconds = finiteNumber(source.remainingSeconds, 0)
  const liveCountdown = getLiveCountdown({ status, timeoutAt, timeoutSeconds })
  const resolvedCountdownText = liveCountdown ? liveCountdown.countdownText : countdownText
  const resolvedProgressStyle = liveCountdown ? liveCountdown.countdownProgressStyle : getCountdownProgressStyle(source)
  const detail = {
    id: source.id || source.invitationId || DEFAULT_DETAIL.id,
    gameId: source.gameId || (source.game && source.game.id) || (source.gameInfo && source.gameInfo.id) || '',
    status,
    statusTitle,
    timeoutAt,
    expiresAt: timeoutAt,
    timeoutSeconds,
    timeLimitSeconds: timeoutSeconds,
    remainingSeconds: liveCountdown ? liveCountdown.remainingSeconds : remainingSeconds,
    countdownProgressPercent: liveCountdown ? liveCountdown.countdownProgressPercent : finiteNumber(source.countdownProgressPercent || source.timeoutPercent || source.countdownPercent, 0),
    countdownText: resolvedCountdownText,
    statusCard: {
      title: statusTitle,
      countdown: resolvedCountdownText,
      progressStyle: resolvedProgressStyle
    },
    primaryActionText: source.primaryActionText || '',
    primaryActionRoute: source.primaryActionRoute || '',
    remindTarget: source.remindTarget || source.targetRole || '',
    cancelRoute: normalizeText(source.cancelRoute),
    player,
    expert,
    inviter: source.inviter || {},
    game,
    playerConfirmedText: source.playerConfirmedText || DEFAULT_DETAIL.playerConfirmedText,
    playerConfirmedAt: source.playerConfirmedAt || DEFAULT_DETAIL.playerConfirmedAt,
    expertConfirmedText: source.expertConfirmedText || DEFAULT_DETAIL.expertConfirmedText,
    startedAt: source.startedAt || source.createdAt || DEFAULT_DETAIL.startedAt,
    infoRows: normalizeInfoRows(source),
    chatPageTitle: source.chatPageTitle || '',
    chatCardTitle: source.chatCardTitle || '',
    chatGuideLabel: source.chatGuideLabel || '',
    chatAssistantName: source.chatAssistantName || '',
    chatPlayerInfoLabel: source.chatPlayerInfoLabel || '',
    chatAcceptButtonText: source.chatAcceptButtonText || '',
    chatDeclineButtonText: source.chatDeclineButtonText || '',
    chatMessageText: source.chatMessageText || '',
    chatDemandActionText: source.chatDemandActionText || ''
  }

  detail.steps = normalizeBackendSteps(source)

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
    pageTitle: '',
    pageTexts: {},
    shellLayout: getWhiteShellLayoutStyles(),
    queryParams: {},
    detailScrollTop: 0,
    loading: false,
    errorText: '',
    detail: normalizeDetail(DEFAULT_DETAIL)
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
    this.startCountdownTimer()
  },

  onHide() {
    this.stopCountdownTimer()
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
    this.stopCountdownTimer()
    this.setData({
      loading: true,
      errorText: ''
    })

    try {
      const data = await gameService.getGuideProgress(params)
      const pageTexts = Object.assign({}, data.pageTexts || {})
      const item = findProgressItem(data, params.id || params.invitationId)

      if (!item) {
        this.setData({
          loading: false,
          pageTexts,
          pageTitle: pageTexts.progressDetailTitle || this.data.pageTitle,
          errorText: pageTexts.detailEmptyText || '',
          detail: normalizeDetail({})
        })
        return
      }

      this.setData({
        loading: false,
        errorText: '',
        pageTexts,
        pageTitle: pageTexts.progressDetailTitle || this.data.pageTitle,
        detail: normalizeDetail(item)
      }, () => {
        this.startCountdownTimer()
      })
    } catch (error) {
      this.setData({
        loading: false,
        errorText: error && error.message ? error.message : (this.data.pageTexts.loadFailedText || '')
      })
    }
  },

  onRetryTap() {
    this.loadProgressDetail(this.data.queryParams)
  },

  startCountdownTimer() {
    this.stopCountdownTimer()
    const hasCountdown = this.refreshCountdownDisplay()

    if (!hasCountdown) {
      return
    }

    this.countdownTimer = setInterval(() => {
      const keepRunning = this.refreshCountdownDisplay()

      if (!keepRunning) {
        this.stopCountdownTimer()
      }
    }, 1000)
  },

  stopCountdownTimer() {
    if (this.countdownTimer) {
      clearInterval(this.countdownTimer)
      this.countdownTimer = null
    }
  },

  refreshCountdownDisplay() {
    const countdown = getLiveCountdown(this.data.detail || {})

    if (!countdown) {
      return false
    }

    this.setData({
      'detail.remainingSeconds': countdown.remainingSeconds,
      'detail.countdownText': countdown.countdownText,
      'detail.countdownProgressPercent': countdown.countdownProgressPercent,
      'detail.statusCard.countdown': countdown.countdownText,
      'detail.statusCard.progressStyle': countdown.countdownProgressStyle
    })

    return countdown.remainingSeconds > 0
  },

  onMapTap() {
    const detail = this.data.detail || {}
    const game = detail.game || {}
    const route = `/pages/map/index?gameId=${encodeURIComponent(this.data.queryParams.sourceGameId || detail.id || '')}&mode=route&title=${encodeURIComponent(game.topic || '')}`

    navigateShellRoute(route)
  },

  onCancelTap() {
    const detail = this.data.detail || {}
    const route = detail.cancelRoute || `/pages/game/guide-cancel/index?gameId=${encodeURIComponent(detail.gameId || '')}`

    navigateShellRoute(route)
  },

  async onRemindTap() {
    const primaryActionRoute = (this.data.detail && this.data.detail.primaryActionRoute) || ''

    if (primaryActionRoute) {
      navigateShellRoute(primaryActionRoute)
      return
    }

    await this.sendReminder()
    const detail = this.data.detail || {}
    const player = detail.player || {}
    const game = detail.game || {}
    const query = [
      `playerName=${encodeURIComponent(player.name || '')}`,
      `playerDesc=${encodeURIComponent(player.desc || '')}`,
      `playerRole=${encodeURIComponent(player.roleLabel || '')}`,
      `demandTargetRole=${encodeURIComponent((detail.expert && detail.expert.roleLabel) || '')}`,
      `demandAction=${encodeURIComponent(detail.chatDemandActionText || '')}`,
      `referrerName=${encodeURIComponent((detail.inviter && detail.inviter.name) || '')}`,
      `pageTitle=${encodeURIComponent(detail.chatPageTitle || '')}`,
      `cardTitle=${encodeURIComponent(detail.chatCardTitle || '')}`,
      `guideLabel=${encodeURIComponent(detail.chatGuideLabel || '')}`,
      `assistantName=${encodeURIComponent(detail.chatAssistantName || '')}`,
      `playerInfoLabel=${encodeURIComponent(detail.chatPlayerInfoLabel || '')}`,
      `acceptButtonText=${encodeURIComponent(detail.chatAcceptButtonText || '')}`,
      `declineButtonText=${encodeURIComponent(detail.chatDeclineButtonText || '')}`,
      `guideMessageText=${encodeURIComponent(detail.chatMessageText || '')}`,
      `gameId=${encodeURIComponent(detail.gameId || '')}`,
      `dateText=${encodeURIComponent(game.time || '')}`,
      `location=${encodeURIComponent(game.location || '')}`
    ].join('&')

    navigateShellRoute(`/pages/game/guide-chat/index?${query}`)
  },

  async sendReminder() {
    const detail = this.data.detail || {}
    const query = this.data.queryParams || {}
    const invitationId = query.invitationId || query.id || detail.invitationId || detail.id
    const gameId = query.sourceGameId || query.gameId || detail.gameId || ''

    try {
      await gameService.sendGuideReminder({
        invitationId,
        gameId,
        remindTarget: detail.remindTarget || '',
        message: `${detail.primaryActionText || ''}${detail.primaryActionText ? '：' : ''}${(detail.game && detail.game.topic) || ''}`
      })
    } catch (error) {
      wx.showToast({
        title: error.message || this.data.pageTexts.remindFailedText || '',
        icon: 'none'
      })
    }
  },

  onTimelineActionTap(event) {
    const key = event.currentTarget.dataset.key

    if (key === 'remindExpert') {
      this.onRemindTap()
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

    navigateShellRoute('/pages/game/guide-progress/index')
  },

  onUnload() {
    this.stopCountdownTimer()
    this.clearDetailScrollTimers()
  }
})
