const gameService = require('../../../services/game')
const { ROUTES } = require('../../../config/routes')

const CONTENT_LEFT_RPX = 0
const CONTENT_TOP_RPX = 160
const CONTENT_WIDTH_RPX = 750
const DESIGN_FRAME_HEIGHT_PT = 810
const DESIGN_BOTTOM_HEIGHT_PT = 30
const NAV_TITLE_HEIGHT_RPX = 50
const BACK_BUTTON_SIZE_RPX = 40
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = DESIGN_FRAME_HEIGHT_PT * 2
const DEFAULT_CANCEL_DETAIL = {
  statusTitle: '组局已取消',
  statusDesc: '本次组局邀请已取消',
  cancelRole: 'member',
  canceledBy: {
    name: '取消方',
    roleType: 'member',
    roleLabel: '成员',
    avatarText: '取'
  },
  reason: {
    title: '时间冲突',
    desc: '对方临时有事，无法按时参加'
  },
  message: '抱歉，时间上有冲突，希望下次有机会再合作。',
  messageTimeText: '刚刚',
  timeline: [
    {
      key: 'invite',
      title: '发起邀请',
      desc: '你向双方发送了组局邀请',
      timeText: '03-21 10:23',
      state: 'active',
      hasLine: true
    },
    {
      key: 'cancel',
      title: '成员取消',
      desc: '对方取消了本次组局邀请',
      timeText: '03-21 16:45',
      state: 'error',
      hasLine: true
    },
    {
      key: 'canceled',
      title: '组局取消',
      desc: '因一方取消，组局自动取消',
      timeText: '',
      state: 'pending',
      hasLine: false
    }
  ]
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

function normalizeText(value, fallback = '') {
  if (value === undefined || value === null) {
    return fallback
  }

  const text = String(value).trim()

  return text || fallback
}

function firstText(values, fallback = '') {
  const matched = values.find((value) => normalizeText(value))

  return normalizeText(matched, fallback)
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

function getRoleLabel(role) {
  const value = normalizeText(role).toLowerCase()

  if (/expert|master|行家|专家/.test(value)) {
    return '行家'
  }

  if (/player|玩家/.test(value)) {
    return '玩家'
  }

  if (/guide|leader|领路人/.test(value)) {
    return '领路人'
  }

  return '成员'
}

function getRoleClass(role) {
  const label = getRoleLabel(role)

  if (label === '行家') {
    return 'expert'
  }

  if (label === '玩家') {
    return 'player'
  }

  if (label === '领路人') {
    return 'guide'
  }

  return 'member'
}

function getInitials(name, fallback) {
  const text = normalizeText(name, fallback)

  if (!text) {
    return ''
  }

  if (/^[A-Za-z\s]+$/.test(text)) {
    return text
      .split(/\s+/)
      .filter(Boolean)
      .map((part) => part.charAt(0).toUpperCase())
      .join('')
      .slice(0, 2)
  }

  return text.slice(0, 2)
}

function getSourceData(data = {}) {
  return data.detail || data.cancelDetail || data.cancelInfo || data
}

function normalizeCancelUser(source = {}) {
  const cancelRole = firstText([
    source.cancelRole,
    source.cancelRoleType,
    source.rejectRole,
    source.rejectRoleType,
    source.canceledByRole,
    source.canceledByRoleType
  ], DEFAULT_CANCEL_DETAIL.cancelRole)
  const user = source.canceledBy || source.cancelUser || source.rejectUser || source.actor || {}
  const roleLabel = firstText([
    user.roleLabel,
    user.roleText,
    source.cancelRoleLabel,
    source.rejectRoleText,
    getRoleLabel(user.roleType || user.role || cancelRole)
  ], '成员')
  const name = firstText([
    user.name,
    user.nickname,
    user.realname,
    source.cancelName,
    source.rejectName
  ], DEFAULT_CANCEL_DETAIL.canceledBy.name)
  const roleClass = getRoleClass(user.roleType || user.role || cancelRole || roleLabel)

  return {
    id: user.id || user.userId || source.cancelUserId || source.rejectUserId || '',
    name,
    roleLabel,
    roleClass,
    avatarUrl: user.avatarUrl || user.avatar || source.cancelAvatarUrl || '',
    avatarText: firstText([
      user.avatarText,
      user.initials,
      source.cancelAvatarText
    ], getInitials(name, roleLabel)),
    avatarClass: user.avatarClass || roleClass
  }
}

function normalizeCancelInfo(data = {}) {
  const source = getSourceData(data)
  const reason = source.reason || source.cancelReason || {}
  const cancelUser = normalizeCancelUser(source)
  const reasonTitle = firstText([
    reason.title,
    reason.reasonTitle,
    source.reasonTitle,
    source.cancelReasonTitle,
    source.reasonCodeText
  ], DEFAULT_CANCEL_DETAIL.reason.title)
  const reasonDesc = firstText([
    reason.desc,
    reason.description,
    reason.text,
    reason.reasonText,
    source.reasonDesc,
    source.cancelReasonText,
    source.reasonSummary
  ], DEFAULT_CANCEL_DETAIL.reason.desc)
  const statusDesc = firstText([
    source.statusDesc,
    source.cancelSummary,
    source.rejectText,
    source.cancelText
  ], `${cancelUser.roleLabel}取消了此次组局邀请`)

  return {
    id: source.id || source.invitationId || source.gameInviteId || '',
    statusTitle: firstText([source.statusTitle, source.title], DEFAULT_CANCEL_DETAIL.statusTitle),
    statusDesc,
    reasonTitle,
    reasonDesc,
    message: firstText([
      source.cancelExplanation,
      source.explanation,
      source.message,
      source.cancelMessage,
      source.rejectMessage,
      source.remark,
      source.comment
    ], DEFAULT_CANCEL_DETAIL.message),
    messageTimeText: firstText([
      source.messageTimeText,
      source.cancelTimeText,
      source.rejectedAtText,
      source.canceledAtText,
      source.timeText
    ], DEFAULT_CANCEL_DETAIL.messageTimeText),
    cancelUserName: cancelUser.name,
    cancelRoleLabel: cancelUser.roleLabel,
    cancelRoleClass: cancelUser.roleClass,
    cancelAvatarUrl: cancelUser.avatarUrl,
    cancelAvatarText: cancelUser.avatarText,
    cancelAvatarClass: cancelUser.avatarClass
  }
}

function normalizeTimeline(data = {}, cancelInfo) {
  const source = getSourceData(data)
  const steps = asArray(source.timeline || source.steps || source.cancelTimeline)

  if (steps.length) {
    return steps.map((item, index) => ({
      key: item.key || item.id || `step-${index}`,
      title: firstText([item.title, item.name], index === 1 ? `${cancelInfo.cancelRoleLabel}取消` : ''),
      desc: firstText([item.desc, item.description, item.content], ''),
      timeText: firstText([item.timeText, item.time, item.createdAtText], ''),
      state: item.state || item.status || (index === 1 ? 'error' : index === steps.length - 1 ? 'pending' : 'active'),
      hasLine: index < steps.length - 1
    }))
  }

  return DEFAULT_CANCEL_DETAIL.timeline.map((item) => {
    if (item.key !== 'cancel') {
      return item
    }

    return Object.assign({}, item, {
      title: `${cancelInfo.cancelRoleLabel}取消`,
      desc: `${cancelInfo.cancelUserName}${cancelInfo.reasonTitle ? `因${cancelInfo.reasonTitle}取消本次组局` : '取消了本次组局邀请'}`,
      timeText: cancelInfo.messageTimeText
    })
  })
}

function normalizeCancelDetail(data = {}) {
  const cancelInfo = normalizeCancelInfo(data)

  return {
    cancelInfo,
    timeline: normalizeTimeline(data, cancelInfo)
  }
}

Page({
  data: {
    pageTitle: '组局取消',
    shellLayout: getWhiteShellLayoutStyles(),
    queryParams: {},
    loading: false,
    errorText: '',
    cancelInfo: normalizeCancelInfo(DEFAULT_CANCEL_DETAIL),
    timeline: DEFAULT_CANCEL_DETAIL.timeline
  },

  onLoad(options = {}) {
    this.setData({
      queryParams: options
    })
    this.updateShellLayout()
    this.loadCancelDetail(options)
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

  async loadCancelDetail(params = {}) {
    this.setData({
      loading: true,
      errorText: ''
    })

    try {
      const data = await gameService.getGuideCancelDetail(params)
      const detail = normalizeCancelDetail(data)

      this.setData({
        loading: false,
        errorText: '',
        cancelInfo: detail.cancelInfo,
        timeline: detail.timeline
      })
    } catch (error) {
      const errorText = error && error.message ? error.message : '取消信息加载失败'

      this.setData({
        loading: false,
        errorText
      })
      wx.showToast({
        title: errorText,
        icon: 'none'
      })
    }
  },

  onRetryTap() {
    this.loadCancelDetail(this.data.queryParams)
  },

  onRestartTap() {
    wx.navigateTo({
      url: `/${ROUTES.gameCreate}`
    })
  },

  onRecommendTap() {
    wx.showToast({
      title: '推荐他人待接入',
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
  }
})
