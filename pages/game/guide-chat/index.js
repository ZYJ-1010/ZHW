const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const { getSurnameInitials } = require('../../../utils/avatar')

const CHAT_SCROLL_TAP_STEP_RPX = 360
const CHAT_SCROLL_HOLD_STEP_RPX = 72
const CHAT_SCROLL_HOLD_INTERVAL_MS = 80
const CHAT_SCROLL_HOLD_SUPPRESS_TAP_MS = 120

const EMPTY_GUIDE_ASSISTANT_CARD = {
  title: '',
  guideLabel: '',
  guideName: '',
  player: {
    name: '',
    avatarText: '',
    role: '',
    desc: ''
  },
  submittedDemand: {
    targetRole: '',
    actionText: ''
  },
  game: {
    dateText: '',
    location: ''
  }
}

function decodeQueryText(value = '') {
  try {
    return decodeURIComponent(value)
  } catch (error) {
    return value
  }
}

function decodeOptions(options = {}) {
  return Object.keys(options).reduce((result, key) => {
    result[key] = decodeQueryText(options[key])
    return result
  }, {})
}

function normalizePlayer(player = {}) {
  const name = player.name || player.nickname || ''

  return {
    name,
    avatarText: player.avatarText || getSurnameInitials(name, ''),
    role: player.role || player.roleName || '',
    desc: player.desc || player.description || ''
  }
}

function normalizeGuideAssistantCard(card = {}) {
  const player = normalizePlayer(card.player || card.targetUser || {})
  const demand = card.submittedDemand || card.demand || {}
  const game = card.game || card.session || {}

  return {
    title: card.title || '',
    guideLabel: card.guideLabel || card.referrerLabel || '',
    guideName: card.guideName || card.referrerName || '',
    player,
    submittedDemand: {
      targetRole: demand.targetRole || demand.role || '',
      actionText: demand.actionText || demand.text || ''
    },
    game: {
      dateText: game.dateText || game.timeText || game.startTimeText || '',
      location: game.location || game.locationText || ''
    }
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

function normalizeGuideChatContext(data = {}) {
  const card = normalizeGuideAssistantCard(data.guideAssistantCard || data.assistantCard || data.card || {})
  const messages = Array.isArray(data.sentMessages || data.messages)
    ? (data.sentMessages || data.messages).map((message, index) => normalizeSentMessage(message, '', `message-${index}`)).filter(Boolean)
    : []

  return {
    onlineText: data.onlineText || '3999人在线',
    pageTitle: data.pageTitle || data.title || '',
    invitationId: data.invitationId || data.gameInvitationId || '',
    gameId: data.gameId || '',
    serviceOrderId: data.serviceOrderId || '',
    guideId: data.guideId || data.referrerId || '',
    sender: {
      avatarText: data.sender && data.sender.avatarText || ''
    },
    guideAssistantCard: card,
    guideMessageText: data.guideMessageText || data.messageText || '',
    sentMessages: messages
  }
}

Page({
  data: {
    onlineText: '3999人在线',
    pageTitle: '',
    invitationId: '',
    gameId: '',
    serviceOrderId: '',
    guideId: '',
    chatScrollTop: 0,
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '入局', active: false },
      { name: '创局', active: false },
      { name: '首页', active: true }
    ],
    sender: {
      avatarText: ''
    },
    guideAssistantCard: EMPTY_GUIDE_ASSISTANT_CARD,
    guideMessageText: '',
    sentMessages: [],
    loading: false,
    responding: false,
    sending: false
  },

  onLoad(options = {}) {
    this.localMessageSeq = 0
    this.loadGuideChatContext(options)
  },

  async loadGuideChatContext(options = {}) {
    const context = decodeOptions(options)

    this.setData({
      invitationId: context.invitationId || context.gameInvitationId || '',
      gameId: context.gameId || context.id || '',
      serviceOrderId: context.serviceOrderId || context.orderId || '',
      guideId: context.guideId || context.referrerId || '',
      loading: true
    })

    try {
      const data = await gameService.getGuideChatContext({
        ...context,
        invitationId: context.invitationId || context.gameInvitationId || '',
        gameId: context.gameId || context.id || '',
        serviceOrderId: context.serviceOrderId || context.orderId || '',
        guideId: context.guideId || context.referrerId || ''
      })

      this.setData({
        ...normalizeGuideChatContext(data),
        loading: false
      })
    } catch (error) {
      this.setData({
        ...normalizeGuideChatContext({}),
        loading: false
      })
      wx.showToast({
        title: error.message || '聊天信息加载失败',
        icon: 'none'
      })
    }
  },

  onAcceptTap() {
    this.handleInvitationResponse('accept')
  },

  onDeclineTap() {
    this.handleInvitationResponse('decline')
  },

  async handleInvitationResponse(action) {
    if (this.data.responding) {
      return
    }

    this.setData({
      responding: true
    })

    try {
      await gameService.respondGuideChatInvitation({
        invitationId: this.data.invitationId,
        gameId: this.data.gameId,
        serviceOrderId: this.data.serviceOrderId,
        guideId: this.data.guideId,
        action
      })
      wx.showToast({
        title: '已提交',
        icon: 'none'
      })
    } catch (error) {
      wx.showToast({
        title: error.message || '处理失败',
        icon: 'none'
      })
    } finally {
      this.setData({
        responding: false
      })
    }
  },

  async onSendMessage(event) {
    const value = String(event.detail && event.detail.value || '').trim()

    if (!value || this.data.sending) {
      return
    }

    this.setData({
      sending: true
    })

    try {
      const result = await gameService.sendGuideChatMessage({
        invitationId: this.data.invitationId,
        gameId: this.data.gameId,
        serviceOrderId: this.data.serviceOrderId,
        guideId: this.data.guideId,
        message: value
      })
      const nextMessage = normalizeSentMessage(result && (result.message || result), value, `local-message-${++this.localMessageSeq}`)

      if (nextMessage) {
        this.setData({
          sentMessages: this.data.sentMessages.concat(nextMessage),
          chatScrollTop: 999999
        })
        this.chatScrollTopValue = 999999
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

  onRecordStart() {
    this.showInfo('开始录音')
  },

  onRecordStop() {
    this.showInfo('录音发送功能待接入')
  },

  onRecordError() {
    this.showInfo('录音失败')
  },

  onChooseImage() {
    this.showInfo('图片发送功能待接入')
  },

  onChooseFile() {
    this.showInfo('文件发送功能待接入')
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (key === 'up' || key === 'down') {
      if (!this.suppressNextNavTap) {
        this.scrollChat(key, CHAT_SCROLL_TAP_STEP_RPX)
      }
      return
    }

    if (key === 'left' || key === 'right') {
      this.showInfo('功能正在开发中')
      return
    }

    this.handleShellAction(key)
  },

  handleShellNavLongPress(event) {
    const key = event.detail && event.detail.key

    if (key !== 'up' && key !== 'down') {
      return
    }

    this.suppressNextNavTap = true
    this.stopChatScrollHold(false)
    this.scrollChat(key, CHAT_SCROLL_HOLD_STEP_RPX)

    this.chatScrollHoldTimer = setInterval(() => {
      this.scrollChat(key, CHAT_SCROLL_HOLD_STEP_RPX)
    }, CHAT_SCROLL_HOLD_INTERVAL_MS)
  },

  handleShellNavTouchEnd() {
    this.stopChatScrollHold(true)
  },

  handleShellAction(key) {
    if (key === 'home') {
      this.scrollChatToTop()
      return
    }

    if (key === 'avatar' || key === 'mine') {
      this.navigateToRoute(ROUTES.profile)
      return
    }

    const routeMap = {
      metaverse: ROUTES.metaverse,
      map: ROUTES.gameHall,
      message: ROUTES.gameCreate
    }

    this.navigateToRoute(routeMap[key])
  },

  navigateToRoute(route) {
    if (!route || route === ROUTES.gameGuideChat) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  },

  handleChatScroll(event) {
    const scrollTop = event.detail && event.detail.scrollTop

    if (typeof scrollTop === 'number') {
      this.chatScrollTopValue = scrollTop
    }
  },

  scrollChat(direction, stepRpx = CHAT_SCROLL_TAP_STEP_RPX) {
    const current = Number(this.chatScrollTopValue || this.data.chatScrollTop || 0)
    const distance = this.rpxToPx(stepRpx)
    const nextTop = direction === 'up'
      ? Math.max(0, current - distance)
      : current + distance

    this.chatScrollTopValue = nextTop
    this.setData({
      chatScrollTop: nextTop
    })
  },

  scrollChatToTop() {
    this.chatScrollTopValue = 0
    this.setData({
      chatScrollTop: 0
    })
  },

  stopChatScrollHold(resetTapSuppress) {
    if (this.chatScrollHoldTimer) {
      clearInterval(this.chatScrollHoldTimer)
      this.chatScrollHoldTimer = null
    }

    if (resetTapSuppress && this.suppressNextNavTap) {
      if (this.chatScrollSuppressTimer) {
        clearTimeout(this.chatScrollSuppressTimer)
      }

      this.chatScrollSuppressTimer = setTimeout(() => {
        this.suppressNextNavTap = false
        this.chatScrollSuppressTimer = null
      }, CHAT_SCROLL_HOLD_SUPPRESS_TAP_MS)
    }
  },

  clearChatScrollTimers() {
    this.stopChatScrollHold(false)

    if (this.chatScrollSuppressTimer) {
      clearTimeout(this.chatScrollSuppressTimer)
      this.chatScrollSuppressTimer = null
    }

    this.suppressNextNavTap = false
  },

  rpxToPx(value) {
    if (!wx.getSystemInfoSync) {
      return value / 2
    }

    const system = wx.getSystemInfoSync()
    const windowWidth = system && system.windowWidth ? system.windowWidth : 375

    return Math.round((value * windowWidth) / 750)
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  },

  onUnload() {
    this.clearChatScrollTimers()
  }
})
