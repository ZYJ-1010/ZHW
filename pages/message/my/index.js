const { ROUTES } = require('../../../config/routes')
const imService = require('../../../services/im')

const QUICK_ACTIONS = [
  { key: 'friend', label: '加好友' },
  { key: 'greet', label: '打招呼' },
  { key: 'card', label: '发名片' },
  { key: 'location', label: '发定位' }
]

function normalizeFriend(source = {}) {
  const name = source.name || source.nickname || source.displayName || ''

  return {
    initials: source.initials || source.avatarText || (name ? name.trim().slice(0, 2) : ''),
    name,
    status: source.status || source.statusText || ''
  }
}

function getMessageList(source = {}) {
  if (Array.isArray(source)) {
    return source
  }

  if (!source || typeof source !== 'object') {
    return []
  }

  return source.list || source.records || source.items || source.messages || []
}

function normalizeMessage(item = {}) {
  const senderType = item.type || item.senderType || item.role || ''
  const text = item.text || item.content || item.message || ''

  return {
    id: item.id || item.messageId || item.clientMessageId || `${senderType}-${item.createdAt || item.timeText || text}`,
    type: senderType === 'self' || senderType === 'me' || item.isMine ? 'self' : 'friend',
    text,
    timeText: item.timeText || item.sentAtText || item.createdAtText || item.createdAt || ''
  }
}

Page({
  data: {
    pageTitle: '好友消息',
    onlineText: '',
    navItems: [
      { name: '我的', key: 'mine' },
      { name: '元宇宙', key: 'metaverse' },
      { name: '地图', key: 'map' },
      { name: '消息', key: 'message' },
      { name: '首页', key: 'home' }
    ],
    roomId: '',
    friend: normalizeFriend(),
    quickActions: QUICK_ACTIONS,
    messages: [],
    chatScrollTop: 0,
    loading: false,
    errorText: ''
  },

  onLoad(options = {}) {
    const roomId = options.roomId || options.conversationId || options.id || ''

    this.setData({
      roomId,
      friend: normalizeFriend({
        name: options.name || '',
        initials: options.initials || options.avatarText || '',
        status: options.status || ''
      })
    })

    this.loadMessages()
  },

  async loadMessages() {
    if (!this.data.roomId) {
      this.setData({
        loading: false,
        errorText: '',
        messages: []
      })
      return
    }

    this.setData({
      loading: true,
      errorText: ''
    })

    try {
      const data = await imService.getMessages(this.data.roomId, {
        page: 1,
        pageSize: 50
      })
      const friend = normalizeFriend(data && (data.friend || data.targetUser || data.conversation || {}))
      const messages = getMessageList(data).map(normalizeMessage).filter((item) => item.text)

      this.setData({
        loading: false,
        errorText: '',
        friend: friend.name ? friend : this.data.friend,
        messages,
        chatScrollTop: 999999
      })
    } catch (error) {
      const errorText = error && error.message ? error.message : '好友消息加载失败'

      this.setData({
        loading: false,
        errorText,
        messages: []
      })
      this.showInfo(errorText)
    }
  },

  onQuickActionTap(event) {
    const action = QUICK_ACTIONS.find((item) => item.key === event.currentTarget.dataset.key)
    this.showInfo(action ? `${action.label}待接入` : '操作待接入')
  },

  async onSendMessage(event) {
    const value = String(event.detail && event.detail.value || '').trim()

    if (!value) {
      return
    }

    try {
      await imService.sendMessage(this.data.roomId, {
        type: 'text',
        content: value
      })
      await this.loadMessages()
    } catch (error) {
      this.showInfo(error && error.message ? error.message : '消息发送失败')
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
    const { key } = event.detail || {}

    if (key === 'map') {
      wx.showToast({
        title: '地图功能开发中',
        icon: 'none'
      })
      return
    }
    const routeMap = {
      home: ROUTES.playerHome || ROUTES.home,
      map: '',
      message: ROUTES.message,
      mine: ROUTES.profile,
      avatar: ROUTES.profile,
      metaverse: ROUTES.metaverse
    }
    const route = routeMap[key]

    if (!route || route === ROUTES.messageMy) {
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
