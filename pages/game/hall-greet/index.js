const WHITE_CONTENT_LEFT_RPX = 2
const WHITE_CONTENT_TOP_RPX = 160
const WHITE_CONTENT_WIDTH_RPX = 750
const WHITE_DESIGN_FRAME_HEIGHT_PT = 810
const WHITE_DESIGN_BOTTOM_HEIGHT_PT = 78
const WHITE_BACK_BUTTON_SIZE_RPX = 40
const WHITE_DEFAULT_CAPSULE_BOTTOM_RPX = 142
const WHITE_DEFAULT_FRAME_HEIGHT_RPX = WHITE_DESIGN_FRAME_HEIGHT_PT * 2

const { ROUTES } = require('../../../config/routes')
const imService = require('../../../services/im')
const { navigateShellRoute } = require('../../../utils/shell-nav')

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
    gameId: 0,
    sentMessages: [],
    scrollIntoView: '',
    quickActions: [
      { key: 'addFriend', label: '加好友', iconSrc: '/pages/game/hall-greet/assets/action-add-friend.png' },
      { key: 'sayHi', label: '打招呼', iconSrc: '/pages/game/hall-greet/assets/action-say-hi.png' },
      { key: 'card', label: '发名片', iconSrc: '/pages/game/hall-greet/assets/action-card.png' },
      { key: 'location', label: '发定位', iconSrc: '/pages/game/hall-greet/assets/action-location.png' }
    ]
  },

  onLoad(options = {}) {
    const gameId = Number(options.gameId || options.sourceGameId || 0)

    this.setData({
      whiteShellLayout: getWhiteShellLayoutStyles(),
      gameId: Number.isInteger(gameId) && gameId > 0 ? gameId : 0
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

    navigateShellRoute('/pages/game/hall/index')
  },

  handleInput(event) {
    this.setData({
      messageValue: event.detail.value || ''
    })
  },

  async handleSend() {
    const value = String(this.data.messageValue || '').trim()

    if (!value) {
      return
    }

    if (await this.sendIMText(value)) {
      this.setData({ messageValue: '' })
      this.appendMessage(value)
    }
  },

  handleAddOrSend() {
    if (this.data.messageValue) {
      this.handleSend()
      return
    }

    this.showInfo('可使用上方快捷动作')
  },

  handleVoiceTap() {
    this.showInfo('当前支持文字消息')
  },

  handleEmojiTap() {
    this.showInfo('当前支持文字消息')
  },

  handleQuickAction(event) {
    const key = event.currentTarget.dataset.key

    if (key === 'addFriend') {
      navigateShellRoute(ROUTES.profile)
      return
    }

    if (key === 'location') {
      const suffix = this.data.gameId ? `?gameId=${encodeURIComponent(this.data.gameId)}&mode=route` : ''
      navigateShellRoute(`/${ROUTES.map}${suffix}`)
      return
    }

    if (key === 'sayHi') {
      this.sendQuickText('你好，我想和你打个招呼。')
      return
    }

    if (key === 'card') {
      this.sendQuickText('这是我的名片，可以先了解一下。')
      return
    }

    this.showInfo('操作信息不完整')
  },

  async sendQuickText(text) {
    if (await this.sendIMText(text)) {
      this.appendMessage(text)
      this.showInfo('已发送')
    }
  },

  async sendIMText(text) {
    if (!this.data.gameId) {
      return true
    }

    try {
      await imService.sendMessage(this.data.gameId, {
        messageType: 'text',
        content: text
      })
      return true
    } catch (error) {
      this.showInfo(error && error.message ? error.message : '发送失败')
      return false
    }
  },

  appendMessage(text) {
    const nextMessage = {
      id: `message-${Date.now()}`,
      text
    }

    this.setData({
      sentMessages: this.data.sentMessages.concat(nextMessage),
      scrollIntoView: nextMessage.id
    })
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
