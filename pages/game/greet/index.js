const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const { getSurnameInitials } = require('../../../utils/avatar')

const EMPTY_CONTACT = {
  name: '',
  realName: '',
  nickname: '',
  avatarText: ''
}

const EMPTY_GROUP_INFO = {
  headerTitle: '',
  title: '',
  guideLabel: '',
  guideName: '',
  expertLabel: '',
  miniProgramText: '',
  expert: {
    name: '',
    avatarText: '',
    desc: '',
    intro: '',
    tags: []
  },
  stats: [],
  confirmText: ''
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

function getAvatarText(name = '', fallback = '') {
  return getSurnameInitials(name, fallback)
}

function getDisplayName(contact = {}) {
  return contact.realName || contact.name || contact.nickname || ''
}

function normalizeContact(contact = {}) {
  const name = contact.name || contact.displayName || contact.nickname || ''
  const realName = contact.realName || ''
  const nickname = contact.nickname || ''
  const displayName = realName || name || nickname

  return {
    name: displayName,
    realName,
    nickname,
    avatarText: contact.avatarText || getAvatarText(displayName, '')
  }
}

function normalizeExpert(expert = {}) {
  const name = expert.name || expert.nickname || ''

  return {
    name,
    avatarText: expert.avatarText || getAvatarText(name, ''),
    desc: expert.desc || expert.description || '',
    intro: expert.intro || expert.bio || '',
    tags: Array.isArray(expert.tags) ? expert.tags : []
  }
}

function normalizeGroupInfo(info = {}) {
  const expert = normalizeExpert(info.expert || {})

  return {
    ...EMPTY_GROUP_INFO,
    ...info,
    expert,
    stats: Array.isArray(info.stats) ? info.stats : []
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

function normalizeGreetContext(data = {}) {
  const contact = normalizeContact(data.contact || data.referrer || data.guide || {})
  const groupInfo = normalizeGroupInfo(data.groupInfo || data.gameInfo || {})
  const messages = Array.isArray(data.sentMessages || data.messages)
    ? (data.sentMessages || data.messages).map((message, index) => normalizeSentMessage(message, '', `message-${index}`)).filter(Boolean)
    : []

  return {
    onlineText: data.onlineText || '3999人在线',
    pageTitle: data.pageTitle || data.title || getDisplayName(contact),
    greetingId: data.greetingId || '',
    gameId: data.gameId || '',
    serviceOrderId: data.serviceOrderId || '',
    invitationId: data.invitationId || '',
    detailRoute: data.detailRoute || '',
    contact,
    groupInfo,
    messageText: data.messageText || data.guideMessageText || '',
    sentMessages: messages,
    canvasMinHeight: data.canvasMinHeight || 1280
  }
}

Page({
  data: {
    onlineText: '3999人在线',
    pageTitle: '',
    greetingId: '',
    gameId: '',
    serviceOrderId: '',
    invitationId: '',
    detailRoute: '',
    contact: EMPTY_CONTACT,
    groupInfo: EMPTY_GROUP_INFO,
    messageText: '',
    sentMessages: [],
    canvasMinHeight: 1280,
    loading: false,
    sending: false
  },

  onLoad(options = {}) {
    this.localMessageSeq = 0
    this.loadGreetContext(options)
  },

  async loadGreetContext(options = {}) {
    const context = decodeOptions(options)

    this.setData({
      greetingId: context.greetingId || context.id || '',
      gameId: context.gameId || '',
      serviceOrderId: context.serviceOrderId || context.orderId || '',
      invitationId: context.invitationId || '',
      loading: true
    })

    try {
      const data = await gameService.getGameGreetingContext({
        ...context,
        greetingId: context.greetingId || context.id || '',
        gameId: context.gameId || '',
        serviceOrderId: context.serviceOrderId || context.orderId || '',
        invitationId: context.invitationId || ''
      })

      this.setData({
        ...normalizeGreetContext(data),
        loading: false
      })
    } catch (error) {
      this.setData({
        ...normalizeGreetContext({}),
        loading: false
      })
      wx.showToast({
        title: error.message || '打招呼信息加载失败',
        icon: 'none'
      })
    }
  },

  onConfirmTap() {
    if (this.data.detailRoute) {
      wx.navigateTo({
        url: this.data.detailRoute
      })
      return
    }

    if (this.data.gameId) {
      wx.navigateTo({
        url: `/${ROUTES.gameDetail}?gameId=${this.data.gameId}`
      })
      return
    }

    wx.showToast({
      title: '缺少局信息',
      icon: 'none'
    })
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
      const result = await gameService.sendGameGreetingMessage({
        greetingId: this.data.greetingId,
        gameId: this.data.gameId,
        serviceOrderId: this.data.serviceOrderId,
        invitationId: this.data.invitationId,
        message: value
      })
      const nextMessage = normalizeSentMessage(result && (result.message || result), value, `local-message-${++this.localMessageSeq}`)

      if (nextMessage) {
        const sentMessages = this.data.sentMessages.concat(nextMessage)

        this.setData({
          sentMessages,
          canvasMinHeight: 1280 + sentMessages.length * 120
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

    if (key === 'map') {
      wx.showToast({
        title: '地图功能开发中',
        icon: 'none'
      })
      return
    }

    if (key === 'avatar' || key === 'mine') {
      this.navigateToRoute(ROUTES.profile)
      return
    }

    if (key === 'comment' || key === 'message') {
      this.navigateToRoute(ROUTES.message)
      return
    }

    const routeMap = {
      home: ROUTES.home,
      metaverse: ROUTES.metaverse,
      map: ''
    }

    this.navigateToRoute(routeMap[key])
  },

  navigateToRoute(route) {
    if (!route || route === ROUTES.gameGreet) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
