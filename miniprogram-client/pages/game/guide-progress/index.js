const { getLegacyWhiteFrameLayoutStyles } = require('../../../utils/adaptive-shell-layout')
const { ROUTES } = require('../../../config/routes')
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
const CONFIRMED_ICON = '/pages/game/guide-progress/assets/status-confirmed.png'
const SUCCESS_ICON = '/pages/game/guide-progress/assets/result-success.png'
const CANCELED_ICON = '/pages/game/guide-progress/assets/result-canceled.png'
const AVATAR_CLASSES = ['pink', 'blue', 'orange', 'green', 'teal', 'purple']

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
  return getLegacyWhiteFrameLayoutStyles({
    contentLeftRpx: CONTENT_LEFT_RPX,
    contentWidthRpx: CONTENT_WIDTH_RPX,
    bottomHeightRatio: DESIGN_BOTTOM_HEIGHT_PT / DESIGN_FRAME_HEIGHT_PT,
    titleHeightRpx: NAV_TITLE_HEIGHT_RPX,
    backSizeRpx: BACK_BUTTON_SIZE_RPX
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

function normalizeRouteValue(route) {
  const text = normalizeText(route)

  if (!text) {
    return ''
  }

  return text.startsWith('/') ? text : `/${text}`
}

function formatTextTemplate(template = '', values = {}) {
  return String(template || '').replace(/\{(\w+)\}/g, (matched, key) => (
    values[key] === undefined || values[key] === null ? matched : String(values[key])
  ))
}

function clampPercent(value) {
  const percent = Number(value)

  if (!Number.isFinite(percent)) {
    return 0
  }

  return Math.max(0, Math.min(100, Math.round(percent)))
}

function parsePercent(value) {
  const match = String(value || '').match(/(\d+(?:\.\d+)?)%/)

  return match ? Number(match[1]) : null
}

function getMemberStateClass(status, statusText) {
  const value = `${status || ''} ${statusText || ''}`.toLowerCase()

  if (/confirm|accept|success|approved|已确认|已同意|成功/.test(value)) {
    return 'confirmed'
  }

  if (/reject|cancel|decline|fail|refuse|婉拒|拒绝|取消|失败/.test(value)) {
    return 'canceled'
  }

  if (/waiting|pending|待|等待|未/.test(value)) {
    return 'waiting'
  }

  return 'pending'
}

function getMemberStatusText(status, fallback = '') {
  return fallback
}

function normalizeMember(item, roleType, index) {
  const member = item || {}
  const name = normalizeText(member.name || member.nickname || member.realname)
  const stateText = normalizeText(member.statusText || member.stateText || member.state)
  const stateClass = normalizeText(member.stateClass, getMemberStateClass(member.confirmStatus || member.status, stateText))
  const avatarClass = normalizeText(member.avatarClass, AVATAR_CLASSES[index % AVATAR_CLASSES.length])

  return {
    id: member.id || member.userId || `${roleType}-${index}`,
    name,
    roleType,
    roleLabel: member.roleText || member.roleLabel || '',
    avatarUrl: member.avatarUrl || '',
    avatarText: member.avatarText || member.avatar || member.initials || '',
    avatarClass,
    state: stateText,
    stateClass,
    icon: member.icon || member.statusIconUrl || ''
  }
}

function splitParticipantsByRole(participants) {
  const players = []
  const experts = []

  asArray(participants).forEach((member) => {
    const role = normalizeText(member && (member.roleType || member.role || member.type)).toLowerCase()

    if (/expert|guide/.test(role)) {
      experts.push(member)
      return
    }

    if (/player/.test(role)) {
      players.push(member)
    }
  })

  return {
    players,
    experts
  }
}

function normalizeRoleMembers(item, roleType) {
  const source = roleType === 'expert'
    ? (item.experts || item.expertList || item.expert || item.expertUser)
    : (item.players || item.playerList || item.player || item.playerUser)

  return asArray(source).map((member, index) => normalizeMember(member, roleType, index))
}

function getFallbackMembers(item, roleType) {
  const participantGroups = splitParticipantsByRole(item.participants || item.members || item.users)
  const source = roleType === 'expert' ? participantGroups.experts : participantGroups.players

  if (source.length) {
    return source.map((member, index) => normalizeMember(member, roleType, index))
  }

  const participants = asArray(item.participants || item.members || item.users)
  const fallback = roleType === 'expert' ? participants.slice(1) : participants.slice(0, 1)

  return fallback.map((member, index) => normalizeMember(member, roleType, index))
}

function normalizeActiveParty(item, index) {
  const active = item || {}
  let players = normalizeRoleMembers(active, 'player')
  let experts = normalizeRoleMembers(active, 'expert')

  if (!players.length) {
    players = getFallbackMembers(active, 'player')
  }

  if (!experts.length) {
    experts = getFallbackMembers(active, 'expert')
  }

  const status = active.status || active.progressStatus || active.state || ''
  const statusText = normalizeText(active.statusText || active.tag)
  const progressPercent = clampPercent(
    active.progressPercent !== undefined
      ? active.progressPercent
      : parsePercent(active.progressStyle) !== null
        ? parsePercent(active.progressStyle)
        : 0
  )
  const player = players[0] || null
  const expert = experts[0] || null
  const notice = normalizeText(active.noticeText || active.notice)

  return {
    id: active.id || active.invitationId || `active-${index}`,
    gameId: active.gameId || active.sourceGameId || '',
    route: normalizeRouteValue(active.route),
    detailRoute: normalizeRouteValue(active.detailRoute || active.route),
    theme: active.theme || '',
    tag: normalizeText(active.tag || statusText),
    time: normalizeText(active.timeText || active.time || active.remainingText),
    progressText: normalizeText(active.progressText || active.stageText),
    progressStyle: `width: ${progressPercent}%;`,
    notice,
    player,
    expert
  }
}

function normalizeCompleteParty(item, index) {
  const complete = item || {}
  const resultStatus = normalizeText(complete.resultStatus || complete.status || complete.state).toLowerCase()
  const isCanceled = complete.isCanceled !== undefined
    ? Boolean(complete.isCanceled)
    : /cancel|reject|decline|refuse|fail/.test(resultStatus)

  return {
    id: complete.id || complete.invitationId || `complete-${index}`,
    gameId: complete.gameId || complete.sourceGameId || '',
    route: normalizeRouteValue(complete.route),
    detailRoute: normalizeRouteValue(complete.detailRoute || complete.route),
    successRoute: normalizeRouteValue(complete.successRoute),
    icon: complete.icon || complete.resultIconUrl || (isCanceled ? CANCELED_ICON : SUCCESS_ICON),
    title: normalizeText(complete.title || complete.statusText),
    time: normalizeText(complete.timeText || complete.time),
    muted: complete.muted !== undefined ? Boolean(complete.muted) : isCanceled,
    isCanceled,
    summaryPrefix: normalizeText(complete.summaryPrefix),
    memberText: normalizeText(complete.completedMemberText || complete.memberText || complete.membersText),
    summarySuffix: normalizeText(complete.summarySuffix),
    reward: normalizeText(complete.rewardText || complete.pointsText || complete.reward),
    gameTitle: normalizeText(complete.gameTitle || complete.topicText || complete.topic),
    rejectName: normalizeText(complete.rejectName || complete.refuserName),
    rejectRole: normalizeText(complete.rejectRoleText || complete.refuserRoleText || complete.roleText || complete.role),
    rejectText: normalizeText(complete.rejectText || complete.refuseText || complete.summarySuffix),
    reason: normalizeText(complete.reasonText || complete.reason)
  }
}

function normalizeGuideProgress(data) {
  const source = data || {}
  const pageTexts = Object.assign({}, source.pageTexts || {})
  const activeSource = source.activeParties || source.activeList || source.ongoingList || source.processingList || []
  const completedSource = source.completedParties || source.completedList || source.recentCompleted || source.historyList || []
  const activeParties = asArray(activeSource).map(normalizeActiveParty)
  const completedParties = asArray(completedSource).map(normalizeCompleteParty)
  const activeCount = Number.isFinite(Number(source.activeCount)) ? Number(source.activeCount) : activeParties.length

  return {
    pageTexts,
    activeCount,
    activeTitleText: formatTextTemplate(pageTexts.activeTitleTemplate, { count: activeCount }),
    activeParties,
    completedParties
  }
}

Page({
  data: {
    pageTitle: '',
    pageTexts: {},
    activeTitleText: '',
    shellLayout: getWhiteShellLayoutStyles(),
    queryParams: {},
    loading: false,
    errorText: '',
    activeCount: 0,
    activeParties: [],
    completedParties: []
  },

  onLoad(options) {
    this.setData({
      queryParams: options || {}
    })
    this.loadGuideProgress(options || {})
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

  async loadGuideProgress(params = {}) {
    this.setData({
      loading: true,
      errorText: ''
    })

    try {
      const data = await gameService.getGuideProgress(params)
      const normalized = normalizeGuideProgress(data)

      this.setData({
        loading: false,
        errorText: '',
        pageTexts: normalized.pageTexts,
        pageTitle: normalized.pageTexts.pageTitle || this.data.pageTitle,
        activeCount: normalized.activeCount,
        activeTitleText: normalized.activeTitleText,
        activeParties: normalized.activeParties,
        completedParties: normalized.completedParties
      })
    } catch (error) {
      this.setData({
        loading: false,
        errorText: error && error.message ? error.message : (this.data.pageTexts.loadFailedText || '')
      })
    }
  },

  onRetryTap() {
    this.loadGuideProgress(this.data.queryParams)
  },

  onProgressCardTap(event) {
    const item = event && event.detail ? event.detail.item || {} : {}

    if (!item.id) {
      return
    }

    const route = !item.isCanceled && item.successRoute
      ? item.successRoute
      : item.detailRoute || item.route || `/${ROUTES.gameGuideProgressDetail}?id=${encodeURIComponent(item.id)}`

    navigateShellRoute(route)
  },

  onBackTap() {
    const pages = getCurrentPages()

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    navigateShellRoute('/pages/game/hall/index')
  }
})
