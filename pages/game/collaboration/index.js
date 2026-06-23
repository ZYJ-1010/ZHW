const WHITE_CONTENT_LEFT_RPX = 2
const WHITE_CONTENT_TOP_RPX = 160
const WHITE_CONTENT_WIDTH_RPX = 750
const WHITE_DESIGN_FRAME_HEIGHT_PT = 810
const WHITE_DESIGN_BOTTOM_HEIGHT_PT = 78
const WHITE_NAV_TITLE_HEIGHT_RPX = 50
const WHITE_BACK_BUTTON_SIZE_RPX = 40
const WHITE_DEFAULT_CAPSULE_BOTTOM_RPX = 142
const WHITE_DEFAULT_FRAME_HEIGHT_RPX = WHITE_DESIGN_FRAME_HEIGHT_PT * 2

function roundRpx(value) {
  return Math.round(value * 100) / 100
}

function getMenuCapsuleBottomRpx() {
  try {
    if (typeof wx !== 'undefined' && wx.getSystemInfoSync && wx.getMenuButtonBoundingClientRect) {
      const menuButton = wx.getMenuButtonBoundingClientRect()
      const systemInfo = wx.getSystemInfoSync()

      if (menuButton && systemInfo && systemInfo.windowWidth) {
        return roundRpx((menuButton.top + menuButton.height) * 750 / systemInfo.windowWidth)
      }
    }
  } catch (error) {
    return WHITE_DEFAULT_CAPSULE_BOTTOM_RPX
  }

  return WHITE_DEFAULT_CAPSULE_BOTTOM_RPX
}

function getLayoutStyles() {
  const capsuleBottom = getMenuCapsuleBottomRpx()
  const titleTop = Math.max(0, roundRpx(capsuleBottom - WHITE_NAV_TITLE_HEIGHT_RPX))
  const backTop = Math.max(0, roundRpx(capsuleBottom - WHITE_BACK_BUTTON_SIZE_RPX))
  let frameHeight = WHITE_DEFAULT_FRAME_HEIGHT_RPX

  try {
    if (typeof wx !== 'undefined' && wx.getSystemInfoSync) {
      const systemInfo = wx.getSystemInfoSync()

      if (systemInfo && systemInfo.windowWidth && systemInfo.windowHeight) {
        frameHeight = roundRpx(systemInfo.windowHeight * 750 / systemInfo.windowWidth)
      }
    }
  } catch (error) {
    frameHeight = WHITE_DEFAULT_FRAME_HEIGHT_RPX
  }

  const bottomHeight = roundRpx(frameHeight * WHITE_DESIGN_BOTTOM_HEIGHT_PT / WHITE_DESIGN_FRAME_HEIGHT_PT)
  const bottomTop = Math.max(WHITE_CONTENT_TOP_RPX, roundRpx(frameHeight - bottomHeight))
  const contentHeight = Math.max(0, roundRpx(bottomTop - WHITE_CONTENT_TOP_RPX))

  return {
    frameStyle: `height: ${frameHeight}rpx; min-height: ${frameHeight}rpx;`,
    contentStyle: [
      `left: ${WHITE_CONTENT_LEFT_RPX}rpx`,
      `top: ${WHITE_CONTENT_TOP_RPX}rpx`,
      `width: ${WHITE_CONTENT_WIDTH_RPX}rpx`,
      `height: ${contentHeight}rpx`
    ].join('; '),
    bottomStyle: `top: ${bottomTop}rpx; height: ${bottomHeight}rpx;`,
    titleStyle: `top: ${titleTop}rpx; height: ${WHITE_NAV_TITLE_HEIGHT_RPX}rpx; line-height: ${WHITE_NAV_TITLE_HEIGHT_RPX}rpx;`,
    backStyle: `top: ${backTop}rpx; width: ${WHITE_BACK_BUTTON_SIZE_RPX}rpx; height: ${WHITE_BACK_BUTTON_SIZE_RPX}rpx;`
  }
}

Page({
  data: {
    layout: getLayoutStyles(),
    progressPercent: 46,
    membersText: '佩娜 (发起人) · 张伟 · 李娜 · 王刚',
    tasks: [
      {
        title: '已完成',
        desc: '确认产品方向'
      },
      {
        title: '进行中',
        desc: '完成用户旅程与页面结构'
      },
      {
        title: '待完成',
        desc: '输出 PRD 与字段清单'
      }
    ],
    messages: [
      {
        name: '佩娜',
        content: '欢迎大家加入，我们先把页面结构和MVP优先级收住。'
      },
      {
        name: '张伟',
        content: '我建议先打穿开局、入局、支付、反馈四条链路。'
      },
      {
        name: '李娜',
        content: '地图打卡可以作为第二增长曲线。'
      }
    ]
  },

  onShow() {
    this.setData({
      layout: getLayoutStyles()
    })
  },

  handleBack() {
    if (getCurrentPages().length > 1) {
      wx.navigateBack()
      return
    }

    wx.navigateTo({
      url: '/pages/game/hall/index'
    })
  },

  handleMemberManage() {
    wx.showToast({
      title: '成员管理',
      icon: 'none'
    })
  },

  handleEndSession() {
    wx.showToast({
      title: '结束本局',
      icon: 'none'
    })
  }
})
