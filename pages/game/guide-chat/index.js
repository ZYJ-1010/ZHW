const { ROUTES } = require('../../../config/routes')
const { getSurnameInitials } = require('../../../utils/avatar')

const CHAT_SCROLL_TAP_STEP_RPX = 360
const CHAT_SCROLL_HOLD_STEP_RPX = 72
const CHAT_SCROLL_HOLD_INTERVAL_MS = 80
const CHAT_SCROLL_HOLD_SUPPRESS_TAP_MS = 120

function decodeQueryText(value = '') {
  try {
    return decodeURIComponent(value)
  } catch (error) {
    return value
  }
}

function getAvatarText(name = '') {
  return getSurnameInitials(name, 'LI')
}

function buildGuideMessage(card) {
  const player = card.player || {}
  const demand = card.submittedDemand || {}

  return `"${player.name}是我认识的${player.role}，正在找${demand.targetRole}${demand.actionText}，我觉得你们很匹配，要不要聊聊？"`
}

Page({
  data: {
    onlineText: '3999人在线',
    pageTitle: '小程序通知',
    chatScrollTop: 0,
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '入局', active: false },
      { name: '创局', active: false },
      { name: '首页', active: true }
    ],
    sender: {
      avatarText: 'ME'
    },
    guideAssistantCard: {
      title: '收到组局邀请',
      guideLabel: '领路人',
      guideName: '我',
      player: {
        name: '李娜',
        avatarText: getSurnameInitials('李娜', 'LI'),
        role: '产品总监',
        desc: '寻找产品经理合作'
      },
      submittedDemand: {
        targetRole: '资深产品经理',
        actionText: '做咨询'
      },
      game: {
        dateText: '3月25日 14:00',
        location: '中关村创业大街'
      }
    },
    guideMessageText: '',
    sentMessages: []
  },

  onLoad(options = {}) {
    const playerName = decodeQueryText(options.playerName || options.name || '')
    const playerDesc = decodeQueryText(options.playerDesc || options.desc || '')
    const playerRole = decodeQueryText(options.playerRole || options.playerTitle || options.role || '')
    const demandTargetRole = decodeQueryText(options.demandTargetRole || options.expertRole || options.expertTitle || '')
    const demandAction = decodeQueryText(options.demandAction || options.demandText || options.requirement || '')
    const referrerName = decodeQueryText(options.referrerName || options.guideName || '')
    const dateText = decodeQueryText(options.dateText || '')
    const location = decodeQueryText(options.location || '')

    const nextData = {}
    const guideAssistantCard = {
      ...this.data.guideAssistantCard,
      player: {
        ...this.data.guideAssistantCard.player
      },
      submittedDemand: {
        ...this.data.guideAssistantCard.submittedDemand
      },
      game: {
        ...this.data.guideAssistantCard.game
      }
    }

    if (playerName) {
      guideAssistantCard.player.name = playerName
      guideAssistantCard.player.avatarText = getAvatarText(playerName)
    }

    if (playerDesc) {
      guideAssistantCard.player.desc = playerDesc
    }

    if (playerRole) {
      guideAssistantCard.player.role = playerRole
    }

    if (demandTargetRole) {
      guideAssistantCard.submittedDemand.targetRole = demandTargetRole
    }

    if (demandAction) {
      guideAssistantCard.submittedDemand.actionText = demandAction
    }

    if (referrerName) {
      guideAssistantCard.guideName = referrerName
    }

    if (dateText) {
      guideAssistantCard.game.dateText = dateText
    }

    if (location) {
      guideAssistantCard.game.location = location
    }

    nextData.guideAssistantCard = guideAssistantCard
    nextData.guideMessageText = buildGuideMessage(guideAssistantCard)

    this.setData(nextData)
  },

  onAcceptTap() {
    this.showInfo('已确认参加')
  },

  onDeclineTap() {
    this.showInfo('已婉拒')
  },

  onSendMessage(event) {
    const value = String(event.detail && event.detail.value || '').trim()

    if (!value) {
      return
    }

    const sentMessages = this.data.sentMessages.concat({
      id: `message-${Date.now()}`,
      text: value
    })

    this.setData({
      sentMessages,
      chatScrollTop: 999999
    })
    this.chatScrollTopValue = 999999
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
