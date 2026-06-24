const WHITE_CONTENT_LEFT_RPX = 2
const WHITE_CONTENT_TOP_RPX = 160
const WHITE_CONTENT_WIDTH_RPX = 750
const WHITE_DESIGN_FRAME_HEIGHT_PT = 810
const WHITE_DESIGN_BOTTOM_HEIGHT_PT = 78
const WHITE_BACK_BUTTON_SIZE_RPX = 40
const WHITE_DEFAULT_CAPSULE_BOTTOM_RPX = 142
const WHITE_DEFAULT_FRAME_HEIGHT_RPX = WHITE_DESIGN_FRAME_HEIGHT_PT * 2

const QUICK_ACTION_TIPS = {
  addFriend: '加好友功能待接入',
  sayHi: '已发送招呼',
  card: '发名片功能待接入',
  location: '发定位功能待接入'
}

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

function getWhiteShellLayoutStyles() {
  const capsuleBottom = getMenuCapsuleBottomRpx()
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
    backStyle: `top: ${backTop}rpx; width: ${WHITE_BACK_BUTTON_SIZE_RPX}rpx; height: ${WHITE_BACK_BUTTON_SIZE_RPX}rpx;`
  }
}

Page({
  data: {
    whiteShellLayout: getWhiteShellLayoutStyles(),
    messageValue: '',
    sentMessages: [],
    scrollIntoView: '',
    quickActions: [
      { key: 'addFriend', label: '加好友', iconSrc: '/pages/game/hall-greet/assets/action-add-friend.png' },
      { key: 'sayHi', label: '打招呼', iconSrc: '/pages/game/hall-greet/assets/action-say-hi.png' },
      { key: 'card', label: '发名片', iconSrc: '/pages/game/hall-greet/assets/action-card.png' },
      { key: 'location', label: '发定位', iconSrc: '/pages/game/hall-greet/assets/action-location.png' }
    ]
  },

  onLoad() {
    this.setData({
      whiteShellLayout: getWhiteShellLayoutStyles()
    })
  },

  onResize() {
    this.setData({
      whiteShellLayout: getWhiteShellLayoutStyles()
    })
  },

  onShareAppMessage() {
    return {
      title: '组局大厅-打招呼',
      path: '/pages/game/hall-greet/index'
    }
  },

  handleBack() {
    const pages = getCurrentPages()

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    wx.navigateTo({
      url: '/pages/game/hall/index'
    })
  },

  handleInput(event) {
    this.setData({
      messageValue: event.detail.value || ''
    })
  },

  handleSend() {
    const value = String(this.data.messageValue || '').trim()

    if (!value) {
      return
    }

    const nextMessage = {
      id: `message-${Date.now()}`,
      text: value
    }

    this.setData({
      messageValue: '',
      sentMessages: this.data.sentMessages.concat(nextMessage),
      scrollIntoView: nextMessage.id
    })
  },

  handleAddOrSend() {
    if (this.data.messageValue) {
      this.handleSend()
      return
    }

    this.showInfo('更多功能待接入')
  },

  handleVoiceTap() {
    this.showInfo('语音功能待接入')
  },

  handleEmojiTap() {
    this.showInfo('表情功能待接入')
  },

  handleQuickAction(event) {
    const key = event.currentTarget.dataset.key

    this.showInfo(QUICK_ACTION_TIPS[key] || '功能待接入')
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
