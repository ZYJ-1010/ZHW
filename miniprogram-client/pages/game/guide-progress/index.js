const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const { getSurnameInitials } = require('../../../utils/avatar')
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
  const capsuleBottom = getCapsuleBottomRpx()
  const frameHeight = getFrameHeightRpx()
  const bottomHeight = roundRpx(frameHeight * DESIGN_BOTTOM_HEIGHT_PT / DESIGN_FRAME_HEIGHT_PT)
  const bottomTop = Math.max(CONTENT_TOP_RPX, roundRpx(frameHeight - bottomHeight))
  const contentHeight = Math.max(0, roundRpx(bottomTop - CONTENT_TOP_RPX))
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

function getInitials(name, fallback) {
  return getSurnameInitials(name, fallback)
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

function getMemberStatusText(status, fallback = '待确认') {
  const value = normalizeText(status).toLowerCase()

  if (/confirm|accept|success|approved/.test(value)) {
    return '已确认'
  }

  if (/reject|decline|refuse/.test(value)) {
    return '已婉拒'
  }

  if (/cancel/.test(value)) {
    return '已取消'
  }

  return fallback
}

function normalizeMember(item, roleType, index) {
  const member = item || {}
  const name = normalizeText(member.name || member.nickname || member.realname, roleType === 'expert' ? '行家' : '玩家')
  const stateText = normalizeText(member.statusText || member.stateText || member.state, getMemberStatusText(member.confirmStatus || member.status))
  const stateClass = normalizeText(member.stateClass, getMemberStateClass(member.confirmStatus || member.status, stateText))
  const avatarClass = normalizeText(member.avatarClass, AVATAR_CLASSES[index % AVATAR_CLASSES.length])

  return {
    id: member.id || member.userId || `${roleType}-${index}`,
    name,
    roleType,
    roleLabel: member.roleText || member.roleLabel || (roleType === 'expert' ? '行家' : '玩家'),
    avatarUrl: member.avatarUrl || '',
    avatarText: getInitials(name, member.avatarText || member.initials || (roleType === 'expert' ? 'EX' : 'PL')),
    avatarClass,
    state: stateText,
    stateClass,
    icon: member.icon || member.statusIconUrl || (stateClass === 'confirmed' ? CONFIRMED_ICON : '')
  }
}

function splitParticipantsByRole(participants) {
  const players = []
  const experts = []

  asArray(participants).forEach((member) => {
    const role = normalizeText(member && (member.roleType || member.role || member.type)).toLowerCase()

    if (/expert|guide|行家/.test(role)) {
      experts.push(member)
      return
    }

    if (/player|玩家/.test(role)) {
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

function inferTheme(status, statusText) {
  const value = `${status || ''} ${statusText || ''}`.toLowerCase()

  if (/both|all|双方|全部/.test(value)) {
    return 'blue'
  }

  return 'orange'
}

function getActiveTheme(status, statusText, players, experts) {
  const members = players.concat(experts)

  if (members.some((member) => member.stateClass === 'confirmed')) {
    return 'orange'
  }

  if (members.length && members.every((member) => member.stateClass !== 'confirmed')) {
    return 'blue'
  }

  return inferTheme(status, statusText)
}

function inferProgressPercent(status, players, experts) {
  const value = normalizeText(status).toLowerCase()

  if (/success|completed|done/.test(value)) {
    return 100
  }

  if (/both|all|双方|waiting_all/.test(value)) {
    return 0
  }

  const members = players.concat(experts)
  const confirmedCount = members.filter((member) => member.stateClass === 'confirmed').length

  return members.length ? Math.round(confirmedCount * 100 / members.length) : 0
}

function getActiveStatusText(status, fallback = '进行中') {
  const value = normalizeText(status).toLowerCase()

  if (/both|all|双方|waiting_all/.test(value)) {
    return '待双方确认'
  }

  if (/expert/.test(value)) {
    return '进行中'
  }

  return fallback
}

function getProgressText(status, fallback) {
  if (fallback) {
    return fallback
  }

  const value = normalizeText(status).toLowerCase()

  if (/both|all|双方|waiting_all/.test(value)) {
    return '等待双方确认'
  }

  if (/expert/.test(value)) {
    return '等待行家确认'
  }

  if (/player/.test(value)) {
    return '等待玩家确认'
  }

  return '等待确认'
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
  const statusText = normalizeText(active.statusText || active.tag, getActiveStatusText(status))
  const progressPercent = clampPercent(
    active.progressPercent !== undefined
      ? active.progressPercent
      : parsePercent(active.progressStyle) !== null
        ? parsePercent(active.progressStyle)
        : inferProgressPercent(status, players, experts)
  )
  const player = players[0] || null
  const expert = experts[0] || null
  const notice = normalizeText(active.noticeText || active.notice)

  return {
    id: active.id || active.invitationId || `active-${index}`,
    theme: getActiveTheme(status, statusText, players, experts),
    tag: statusText,
    time: normalizeText(active.timeText || active.time || active.remainingText),
    progressText: getProgressText(status, active.progressText || active.stageText),
    progressStyle: `width: ${progressPercent}%;`,
    notice,
    player,
    expert
  }
}

function getFirstNameText(list) {
  const item = asArray(list)[0]

  return normalizeText(item && (item.name || item.nickname || item.realname))
}

function getCompleteMemberText(players, experts) {
  const playerName = getFirstNameText(players)
  const expertName = getFirstNameText(experts)

  if (playerName && expertName) {
    return `${playerName} 与 ${expertName}`
  }

  return playerName || expertName
}

function normalizeCompleteParty(item, index) {
  const complete = item || {}
  const resultStatus = normalizeText(complete.resultStatus || complete.status || complete.state).toLowerCase()
  const players = asArray(complete.players || complete.playerList || complete.player || complete.playerUser)
  const experts = asArray(complete.experts || complete.expertList || complete.expert || complete.expertUser)
  const isCanceled = /cancel|reject|decline|refuse|fail|取消|拒绝|婉拒/.test(resultStatus)
  const rejectRoleType = normalizeText(complete.rejectRoleType || complete.refuseRoleType).toLowerCase()
  const rejectFallbackName = rejectRoleType === 'expert' ? getFirstNameText(experts) : getFirstNameText(players)

  return {
    id: complete.id || complete.invitationId || `complete-${index}`,
    icon: complete.icon || complete.resultIconUrl || (isCanceled ? CANCELED_ICON : SUCCESS_ICON),
    title: normalizeText(complete.title || complete.statusText, isCanceled ? '组局已取消' : '组局成功'),
    time: normalizeText(complete.timeText || complete.time),
    muted: complete.muted !== undefined ? Boolean(complete.muted) : isCanceled,
    isCanceled,
    summaryPrefix: normalizeText(complete.summaryPrefix),
    memberText: normalizeText(
      complete.completedMemberText || complete.memberText || complete.membersText,
      getCompleteMemberText(players, experts)
    ),
    summarySuffix: normalizeText(complete.summarySuffix),
    reward: normalizeText(complete.rewardText || complete.pointsText || complete.reward),
    gameTitle: normalizeText(complete.gameTitle || complete.topicText || complete.topic),
    rejectName: normalizeText(complete.rejectName || complete.refuserName, rejectFallbackName),
    rejectRole: normalizeText(complete.rejectRoleText || complete.refuserRoleText || complete.roleText || complete.role),
    rejectText: normalizeText(complete.rejectText || complete.refuseText || complete.summarySuffix),
    reason: normalizeText(complete.reasonText || complete.reason)
  }
}

function normalizeGuideProgress(data) {
  const source = data || {}
  const activeSource = source.activeParties || source.activeList || source.ongoingList || source.processingList || []
  const completedSource = source.completedParties || source.completedList || source.recentCompleted || source.historyList || []
  const activeParties = asArray(activeSource).map(normalizeActiveParty)
  const completedParties = asArray(completedSource).map(normalizeCompleteParty)

  return {
    activeCount: Number.isFinite(Number(source.activeCount)) ? Number(source.activeCount) : activeParties.length,
    activeParties,
    completedParties
  }
}

Page({
  data: {
    pageTitle: '组局消息',
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
        activeCount: normalized.activeCount,
        activeParties: normalized.activeParties,
        completedParties: normalized.completedParties
      })
    } catch (error) {
      this.setData({
        loading: false,
        errorText: error && error.message ? error.message : '组局消息加载失败'
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

    navigateShellRoute(`/${ROUTES.gameGuideProgressDetail}?id=${encodeURIComponent(item.id)}`)
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
