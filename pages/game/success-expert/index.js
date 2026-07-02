const gameService = require('../../../services/game')
const { ROUTES } = require('../../../config/routes')
const { getSurnameInitials } = require('../../../utils/avatar')
const { navigateShellRoute } = require('../../../utils/shell-nav')

const CONTENT_LEFT_RPX = 2
const CONTENT_TOP_RPX = 160
const CONTENT_WIDTH_RPX = 750
const DESIGN_FRAME_HEIGHT_PT = 810
const DESIGN_BOTTOM_HEIGHT_PT = 78
const NAV_TITLE_HEIGHT_RPX = 50
const BACK_BUTTON_SIZE_RPX = 40
const MORE_BUTTON_SIZE_RPX = 44
const MORE_BUTTON_LEFT_RPX = 658
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = DESIGN_FRAME_HEIGHT_PT * 2

const DEFAULT_SUCCESS_DETAIL = {
  group: {
    title: '',
    roles: '',
    hint: ''
  },
  participants: [],
  activityRows: [],
  fund: {
    title: '',
    desc: '',
    amountText: '',
    progress: 0,
    stepLabels: []
  },
  nextSteps: []
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

function getBlankShellLayoutStyles() {
  const capsuleBottom = getCapsuleBottomRpx()
  const frameHeight = getFrameHeightRpx()
  const bottomHeight = roundRpx(frameHeight * DESIGN_BOTTOM_HEIGHT_PT / DESIGN_FRAME_HEIGHT_PT)
  const bottomTop = Math.max(CONTENT_TOP_RPX, roundRpx(frameHeight - bottomHeight))
  const contentHeight = Math.max(0, roundRpx(bottomTop - CONTENT_TOP_RPX))
  const titleTop = Math.max(0, roundRpx(capsuleBottom - NAV_TITLE_HEIGHT_RPX))
  const backTop = Math.max(0, roundRpx(capsuleBottom - BACK_BUTTON_SIZE_RPX))
  const moreTop = Math.max(0, roundRpx(capsuleBottom - MORE_BUTTON_SIZE_RPX - 3))

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
    backStyle: `top: ${backTop}rpx; width: ${BACK_BUTTON_SIZE_RPX}rpx; height: ${BACK_BUTTON_SIZE_RPX}rpx;`,
    moreStyle: `left: ${MORE_BUTTON_LEFT_RPX}rpx; top: ${moreTop}rpx; width: ${MORE_BUTTON_SIZE_RPX}rpx; height: ${MORE_BUTTON_SIZE_RPX}rpx;`
  }
}

function normalizeSuccessDetail(data = {}) {
  const source = data && typeof data === 'object' ? data : {}
  const fund = Object.assign({}, DEFAULT_SUCCESS_DETAIL.fund, source.fund || {})
  const participants = Array.isArray(source.participants) && source.participants.length
    ? source.participants.map((item, index) => ({
      id: item.id || item.userId || `participant-${index}`,
      avatar: item.avatar || item.avatarText || getSurnameInitials(item.name || '', item.roleLabel || '成员'),
      colorClass: item.colorClass || item.avatarClass || 'blue'
    }))
    : []

  return {
    viewer: source.viewer || {},
    group: Object.assign({}, DEFAULT_SUCCESS_DETAIL.group, source.group || {}),
    participants,
    participantCount: participants.length,
    activityRows: Array.isArray(source.activityRows) && source.activityRows.length ? source.activityRows : [],
    fund: Object.assign({}, fund, {
      progressStyle: `width: ${Math.max(0, Math.min(100, Number(fund.progress || 0)))}%;`,
      stepLabels: Array.isArray(fund.stepLabels) && fund.stepLabels.length ? fund.stepLabels : []
    }),
    nextSteps: Array.isArray(source.nextSteps) && source.nextSteps.length ? source.nextSteps : []
  }
}

Page({
  data: {
    shellLayout: getBlankShellLayoutStyles(),
    gameId: '',
    loading: false,
    errorText: '',
    ...normalizeSuccessDetail(DEFAULT_SUCCESS_DETAIL)
  },

  onLoad(options = {}) {
    const gameId = options.gameId || options.id || ''

    this.setData({ gameId })
    if (gameId) {
      this.loadSuccessDetail(gameId)
    }
  },

  onShow() {
    this.updateShellLayout()
  },

  onResize() {
    this.updateShellLayout()
  },

  updateShellLayout() {
    this.setData({
      shellLayout: getBlankShellLayoutStyles()
    })
  },

  async loadSuccessDetail(gameId) {
    this.setData({
      loading: true,
      errorText: ''
    })

    try {
      const data = await gameService.getGameSuccessDetail(gameId, { role: 'expert' })

      this.setData({
        loading: false,
        errorText: '',
        ...normalizeSuccessDetail(data)
      })
    } catch (error) {
      this.setData({
        loading: false,
        errorText: error.message || '组局成功详情加载失败，请稍后重试'
      })
    }
  },

  onBackTap() {
    const pages = getCurrentPages()

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    navigateShellRoute(ROUTES.gameGuideProgress)
  },

  onMoreTap() {
    this.onManageTap()
  },

  onEnterChatTap() {
    const gameId = this.data.gameId || this.data.group.gameId || ''

    navigateShellRoute(`/${ROUTES.imRoom}?gameId=${encodeURIComponent(gameId)}&role=${encodeURIComponent(this.data.viewer.role || 'expert')}&prefill=${encodeURIComponent('我已进入三方群，准备确认后续服务安排。')}`)
  },

  onStepTap(event) {
    const action = event.currentTarget.dataset.action

    if (action === 'contact_player') {
      this.onEnterChatTap()
      return
    }

    this.onManageTap()
  },

  onManageTap() {
    navigateShellRoute(`/${ROUTES.gameManage || 'pages/game/manage/index'}${this.data.gameId ? `?gameId=${encodeURIComponent(this.data.gameId)}` : ''}`)
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
