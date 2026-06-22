const { ROUTES } = require('../../../config/routes')

const CONTENT_LEFT_RPX = 2
const CONTENT_TOP_RPX = 160
const CONTENT_WIDTH_RPX = 750
const DESIGN_FRAME_HEIGHT_PT = 810
const DESIGN_BOTTOM_HEIGHT_PT = 78
const NAV_TITLE_HEIGHT_RPX = 50
const BACK_BUTTON_SIZE_RPX = 40
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = DESIGN_FRAME_HEIGHT_PT * 2
const CTA_BAR_HEIGHT_RPX = 158

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
  const ctaTop = Math.max(CONTENT_TOP_RPX, roundRpx(frameHeight - CTA_BAR_HEIGHT_RPX))

  return {
    frameStyle: `height: ${frameHeight}rpx; min-height: ${frameHeight}rpx;`,
    contentStyle: [
      `left: ${CONTENT_LEFT_RPX}rpx`,
      `top: ${CONTENT_TOP_RPX}rpx`,
      `width: ${CONTENT_WIDTH_RPX}rpx`,
      `height: ${contentHeight}rpx`
    ].join('; '),
    bottomStyle: `top: ${bottomTop}rpx; height: ${bottomHeight}rpx;`,
    ctaStyle: `top: ${ctaTop}rpx; height: ${CTA_BAR_HEIGHT_RPX}rpx;`,
    titleStyle: `top: ${titleTop}rpx; height: ${NAV_TITLE_HEIGHT_RPX}rpx; line-height: ${NAV_TITLE_HEIGHT_RPX}rpx;`,
    backStyle: `top: ${backTop}rpx; width: ${BACK_BUTTON_SIZE_RPX}rpx; height: ${BACK_BUTTON_SIZE_RPX}rpx;`
  }
}

Page({
  data: {
    shellLayout: getWhiteShellLayoutStyles(),
    timeline: [
      {
        title: '发起引荐',
        desc: '你向双方发送了组局邀请',
        time: '03-21 10:23'
      },
      {
        title: '玩家确认',
        desc: '李娜确认参加组局',
        time: '03-21 11:05'
      },
      {
        title: '行家确认',
        desc: '王强确认参加组局',
        time: '03-21 14:30'
      },
      {
        title: '组局成功！',
        desc: '双方已建立连接，进入交付阶段',
        time: '03-21 14:30',
        active: true
      }
    ],
    party: {
      confirmedText: '',
      cardClass: 'success-guide-party-card',
      cardStyle: 'width: 684rpx; height: 446rpx; margin: 32rpx auto 0; box-shadow: none;',
      titleClass: 'regular',
      player: {
        avatarText: 'LN',
        avatarClass: 'pink',
        name: '李娜',
        role: '玩家',
        state: '已确认',
        stateClass: 'confirmed'
      },
      expert: {
        avatarText: 'WQ',
        avatarClass: 'blue',
        name: '王强',
        role: '行家',
        state: '已确认',
        stateClass: 'confirmed'
      }
    },
    followUps: [
      {
        key: 'schedule',
        title: '查看组局日程',
        desc: '活动将于3月25日举行',
        iconSrc: './assets/follow-schedule.png',
        iconClass: 'schedule',
        theme: 'blue'
      },
      {
        key: 'feedback',
        title: '询问双方反馈',
        desc: '了解交流情况，促成深度合作',
        iconSrc: './assets/follow-feedback.png',
        iconClass: 'feedback',
        theme: 'purple',
        reward: '+20积分'
      },
      {
        key: 'deal',
        title: '促成交易',
        desc: '协助双方达成合作意向',
        iconSrc: './assets/follow-deal.png',
        iconClass: 'deal',
        theme: 'orange',
        reward: '+100积分'
      }
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
      shellLayout: getWhiteShellLayoutStyles()
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

  onFollowTap(event) {
    const item = this.data.followUps.find((entry) => entry.key === event.currentTarget.dataset.key)

    wx.showToast({
      title: item ? item.title : '后续跟进待接入',
      icon: 'none'
    })
  },

  onShareAppMessage() {
    return {
      title: '组局成功',
      path: `/${ROUTES.gameSuccessGuide}`
    }
  }
})
