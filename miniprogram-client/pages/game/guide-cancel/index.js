const { getLegacyWhiteFrameLayoutStyles } = require('../../../utils/adaptive-shell-layout')
const gameService = require('../../../services/game')
const { ROUTES } = require('../../../config/routes')
const { getSurnameInitials } = require('../../../utils/avatar')
const { navigateShellRoute } = require('../../../utils/shell-nav')

const CONTENT_LEFT_RPX = 0
const CONTENT_TOP_RPX = 160
const CONTENT_WIDTH_RPX = 750
const DESIGN_FRAME_HEIGHT_PT = 810
const DESIGN_BOTTOM_HEIGHT_PT = 30
const NAV_TITLE_HEIGHT_RPX = 50
const BACK_BUTTON_SIZE_RPX = 40
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = DESIGN_FRAME_HEIGHT_PT * 2
const EMPTY_CANCEL_DETAIL = {
  statusTitle: '',
  statusDesc: '',
  cancelRole: '',
  canceledBy: {
    name: '',
    roleType: '',
    roleLabel: '',
    avatarText: ''
  },
  reason: {
    title: '',
    desc: ''
  },
  message: '',
  messageTimeText: '',
  timeline: []
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
  return getLegacyWhiteFrameLayoutStyles({
    contentLeftRpx: CONTENT_LEFT_RPX,
    contentWidthRpx: CONTENT_WIDTH_RPX,
    bottomHeightRatio: DESIGN_BOTTOM_HEIGHT_PT / DESIGN_FRAME_HEIGHT_PT,
    titleHeightRpx: NAV_TITLE_HEIGHT_RPX,
    backSizeRpx: BACK_BUTTON_SIZE_RPX
  })
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
  return getSurnameInitials(name, fallback)
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
  ], EMPTY_CANCEL_DETAIL.cancelRole)
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
  ], EMPTY_CANCEL_DETAIL.canceledBy.name)
  const roleClass = getRoleClass(user.roleType || user.role || cancelRole || roleLabel)

  return {
    id: user.id || user.userId || source.cancelUserId || source.rejectUserId || '',
    name,
    roleLabel,
    roleClass,
    avatarUrl: user.avatarUrl || user.avatar || source.cancelAvatarUrl || '',
    avatarText: getInitials(name, firstText([
      user.avatarText,
      user.initials,
      source.cancelAvatarText
    ], roleLabel)),
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
  ], EMPTY_CANCEL_DETAIL.reason.title)
  const reasonDesc = firstText([
    reason.desc,
    reason.description,
    reason.text,
    reason.reasonText,
    source.reasonDesc,
    source.cancelReasonText,
    source.reasonSummary
  ], EMPTY_CANCEL_DETAIL.reason.desc)
  const statusDesc = firstText([
    source.statusDesc,
    source.cancelSummary,
    source.rejectText,
    source.cancelText
  ], `${cancelUser.roleLabel}取消了此次组局邀请`)

  return {
    id: source.id || source.invitationId || source.gameInviteId || '',
    statusTitle: firstText([source.statusTitle, source.title], EMPTY_CANCEL_DETAIL.statusTitle),
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
    ], EMPTY_CANCEL_DETAIL.message),
    messageTimeText: firstText([
      source.messageTimeText,
      source.cancelTimeText,
      source.rejectedAtText,
      source.canceledAtText,
      source.timeText
    ], EMPTY_CANCEL_DETAIL.messageTimeText),
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

  return EMPTY_CANCEL_DETAIL.timeline.map((item) => {
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
    cancelInfo: normalizeCancelInfo(EMPTY_CANCEL_DETAIL),
    timeline: EMPTY_CANCEL_DETAIL.timeline
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
    navigateShellRoute(this.buildUrl(ROUTES.gameCreate, {
        source: 'guideCancel',
        gameId: this.data.queryParams.gameId || '',
        sourceGameId: this.data.queryParams.gameId || '',
        invitationId: this.data.queryParams.invitationId || this.data.queryParams.id || this.data.cancelInfo.id || ''
      }))
  },

  onRecommendTap() {
    navigateShellRoute(this.buildUrl(ROUTES.gameInvite, {
        source: 'guideCancel',
        gameId: this.data.queryParams.gameId || '',
        invitationId: this.data.queryParams.invitationId || this.data.queryParams.id || this.data.cancelInfo.id || ''
      }))
  },

  buildUrl(route, params = {}) {
    const query = Object.keys(params)
      .filter((key) => params[key])
      .map((key) => `${key}=${encodeURIComponent(params[key])}`)
      .join('&')

    return `/${route}${query ? `?${query}` : ''}`
  },

  onBackTap() {
    const pages = getCurrentPages()

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    navigateShellRoute('/pages/game/guide-progress/index')
  }
})
