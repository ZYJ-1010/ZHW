const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const { getSurnameInitials } = require('../../../utils/avatar')

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

function normalizeParticipant(participant = {}, index = 0) {
  const name = participant.name || participant.nickname || ''

  return {
    id: participant.id || participant.userId || `participant-${index}`,
    avatar: participant.avatar || participant.avatarText || getSurnameInitials(name, ''),
    colorClass: participant.colorClass || participant.avatarClass || ''
  }
}

function normalizeExpertSuccess(data = {}) {
  const group = data.group || data.chatGroup || {}
  const participants = Array.isArray(data.participants || data.members)
    ? (data.participants || data.members).map(normalizeParticipant)
    : []
  const activityRows = Array.isArray(data.activityRows || data.infoRows)
    ? (data.activityRows || data.infoRows).map((item) => ({
      label: item.label || item.name || '',
      value: item.value || item.text || '',
      highlight: Boolean(item.highlight)
    })).filter((item) => item.label || item.value)
    : []
  const nextSteps = Array.isArray(data.nextSteps || data.steps)
    ? (data.nextSteps || data.steps).map((item, index) => ({
      index: item.index || index + 1,
      title: item.title || item.name || '',
      desc: item.desc || item.description || '',
      active: Boolean(item.active || item.current)
    })).filter((item) => item.title || item.desc)
    : []

  return {
    group: {
      title: group.title || '',
      roles: group.roles || group.rolesText || '',
      hint: group.hint || group.hintText || '',
      chatRoute: group.chatRoute || data.chatRoute || '',
      manageRoute: group.manageRoute || data.manageRoute || ''
    },
    participants,
    activityRows,
    nextSteps
  }
}

Page({
  data: {
    shellLayout: getBlankShellLayoutStyles(),
    queryParams: {},
    loading: false,
    group: {
      title: '',
      roles: '',
      hint: '',
      chatRoute: '',
      manageRoute: ''
    },
    participants: [],
    activityRows: [],
    nextSteps: []
  },

  onLoad(options = {}) {
    this.setData({
      queryParams: options
    })
    this.loadExpertSuccess(options)
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

  async loadExpertSuccess(options = {}) {
    this.setData({
      loading: true
    })

    try {
      const data = await gameService.getExpertSuccess(options)

      this.setData({
        ...normalizeExpertSuccess(data),
        loading: false
      })
    } catch (error) {
      this.setData({
        ...normalizeExpertSuccess({}),
        loading: false
      })
      this.showInfo(error.message || '组局成功信息加载失败')
    }
  },

  onBackTap() {
    const pages = getCurrentPages()

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    wx.navigateTo({
      url: `/${ROUTES.gameGuideProgress}`
    })
  },

  onMoreTap() {
    this.showInfo('更多操作待接入')
  },

  onEnterChatTap() {
    if (this.data.group.chatRoute) {
      wx.navigateTo({
        url: this.data.group.chatRoute
      })
      return
    }

    this.showInfo('群聊入口待接入')
  },

  onManageTap() {
    wx.navigateTo({
      url: this.data.group.manageRoute || `/${ROUTES.gameManage || 'pages/game/manage/index'}`,
      fail: () => {
        this.showInfo('局管理页待接入')
      }
    })
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
