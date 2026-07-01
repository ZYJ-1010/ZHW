const gameService = require('../../../services/game')

const WHITE_CONTENT_LEFT_RPX = 2
const WHITE_CONTENT_TOP_RPX = 160
const WHITE_CONTENT_WIDTH_RPX = 750
const WHITE_DESIGN_FRAME_HEIGHT_PT = 810
const WHITE_DESIGN_BOTTOM_HEIGHT_PT = 78
const WHITE_BACK_BUTTON_SIZE_RPX = 40
const WHITE_DEFAULT_CAPSULE_BOTTOM_RPX = 142
const WHITE_DEFAULT_FRAME_HEIGHT_RPX = WHITE_DESIGN_FRAME_HEIGHT_PT * 2

const QUICK_ACTION_ICON_MAP = {
  addFriend: '/pages/game/hall-greet/assets/action-add-friend.png',
  sayHi: '/pages/game/hall-greet/assets/action-say-hi.png',
  card: '/pages/game/hall-greet/assets/action-card.png',
  location: '/pages/game/hall-greet/assets/action-location.png'
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

function normalizeQuickAction(action = {}) {
  const key = action.key || action.actionKey || action.type || ''

  return {
    key,
    label: action.label || action.title || '',
    iconSrc: action.iconSrc || action.icon || QUICK_ACTION_ICON_MAP[key] || ''
  }
}

function normalizeSentMessage(message = {}, fallbackText = '', fallbackId = '') {
  const text = message.text || message.content || message.message || fallbackText

  if (!text) {
    return null
  }

  return {
    id: String(message.id || message.messageId || fallbackId),
    text
  }
}

function normalizeHallGreetContext(data = {}) {
  const messages = Array.isArray(data.sentMessages || data.messages)
    ? (data.sentMessages || data.messages).map((message, index) => normalizeSentMessage(message, '', `message-${index}`)).filter(Boolean)
    : []

  return {
    greetingId: data.greetingId || '',
    gameId: data.gameId || '',
    targetUserId: data.targetUserId || '',
    shareTitle: data.shareTitle || '',
    quickActions: Array.isArray(data.quickActions) ? data.quickActions.map(normalizeQuickAction).filter((item) => item.key) : [],
    sentMessages: messages
  }
}

Page({
  data: {
    whiteShellLayout: getWhiteShellLayoutStyles(),
    greetingId: '',
    gameId: '',
    targetUserId: '',
    shareTitle: '',
    messageValue: '',
    sentMessages: [],
    scrollIntoView: '',
    quickActions: [],
    loading: false,
    sending: false,
    actionSubmitting: false
  },

  onLoad(options = {}) {
    this.localMessageSeq = 0
    this.setData({
      whiteShellLayout: getWhiteShellLayoutStyles()
    })
    this.loadHallGreetContext(options)
  },

  onResize() {
    this.setData({
      whiteShellLayout: getWhiteShellLayoutStyles()
    })
  },

  onShareAppMessage() {
    return {
      title: this.data.shareTitle || '组局大厅-打招呼',
      path: '/pages/game/hall-greet/index'
    }
  },

  async loadHallGreetContext(options = {}) {
    this.setData({
      greetingId: options.greetingId || options.id || '',
      gameId: options.gameId || '',
      targetUserId: options.targetUserId || options.userId || '',
      loading: true
    })

    try {
      const data = await gameService.getHallGreetingContext({
        ...options,
        greetingId: options.greetingId || options.id || '',
        gameId: options.gameId || '',
        targetUserId: options.targetUserId || options.userId || ''
      })

      this.setData({
        ...normalizeHallGreetContext(data),
        loading: false
      })
    } catch (error) {
      this.setData({
        ...normalizeHallGreetContext({}),
        loading: false
      })
      wx.showToast({
        title: error.message || '打招呼配置加载失败',
        icon: 'none'
      })
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

  async handleSend() {
    const value = String(this.data.messageValue || '').trim()

    if (!value || this.data.sending) {
      return
    }

    this.setData({
      sending: true
    })

    try {
      const result = await gameService.sendHallGreetingMessage({
        greetingId: this.data.greetingId,
        gameId: this.data.gameId,
        targetUserId: this.data.targetUserId,
        message: value
      })
      const nextMessage = normalizeSentMessage(result && (result.message || result), value, `local-message-${++this.localMessageSeq}`)

      if (nextMessage) {
        this.setData({
          messageValue: '',
          sentMessages: this.data.sentMessages.concat(nextMessage),
          scrollIntoView: nextMessage.id
        })
      }
    } catch (error) {
      wx.showToast({
        title: error.message || '消息发送失败',
        icon: 'none'
      })
    } finally {
      this.setData({
        sending: false
      })
    }
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

  async handleQuickAction(event) {
    const key = event.currentTarget.dataset.key

    if (!key || this.data.actionSubmitting) {
      return
    }

    this.setData({
      actionSubmitting: true
    })

    try {
      const result = await gameService.triggerHallGreetingAction({
        greetingId: this.data.greetingId,
        gameId: this.data.gameId,
        targetUserId: this.data.targetUserId,
        actionKey: key
      })

      this.showInfo(result && (result.message || result.toastText) || '操作已提交')
    } catch (error) {
      this.showInfo(error.message || '操作提交失败')
    } finally {
      this.setData({
        actionSubmitting: false
      })
    }
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
