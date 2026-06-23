const { ROUTES } = require('../../../config/routes')
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

Page({
  data: {
    shellLayout: getBlankShellLayoutStyles(),
    group: {
      title: '产品架构咨询 - 三方群',
      roles: '行家、领路人、玩家',
      hint: '领路人王引荐将持续跟进活动进度，确保双方顺利对接'
    },
    participants: [
      { id: 'me', avatar: 'ME', colorClass: 'blue' },
      { id: 'wa', avatar: getSurnameInitials('王引荐', 'WA'), colorClass: 'orange' },
      { id: 'lm', avatar: getSurnameInitials('李明', 'LI'), colorClass: 'pink' }
    ],
    activityRows: [
      { label: '活动编号', value: 'REF-20260323-001' },
      { label: '创建时间', value: '2026-03-23 10:23' },
      { label: '组局时间', value: '2026-03-23 14:30' },
      { label: '当前阶段', value: '待交付服务', highlight: true }
    ],
    nextSteps: [
      { index: 1, title: '联系玩家确认具体时间', desc: '建议24小时内完成', active: true },
      { index: 2, title: '按时交付服务', desc: '等待确认时间' },
      { index: 3, title: '确认完成并收款', desc: '等待服务完成' }
    ]
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
    this.showInfo('群聊入口待接入')
  },

  onManageTap() {
    wx.navigateTo({
      url: `/${ROUTES.gameManage || 'pages/game/manage/index'}`,
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
