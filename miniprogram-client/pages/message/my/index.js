const { ROUTES } = require('../../../config/routes')
const messageService = require('../../../services/message')
const imService = require('../../../services/im')
const chatMedia = require('../../../services/chat-media')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')

function toPositiveInt(value) {
  const number = Number(value)
  return Number.isInteger(number) && number > 0 ? number : 0
}

function formatTime(value) {
  if (!value) {
    return ''
  }

  return String(value).replace('T', ' ').replace(/:\d{2}(?:\.\d+)?Z?$/, '')
}

function initialsFromName(name) {
  const text = String(name || '').trim()
  return text ? text.slice(0, 2).toUpperCase() : 'IM'
}

function normalizeChatMessages(items, currentUserId) {
  return (Array.isArray(items) ? items : []).map((item) => ({
    id: item.id || `m-${Date.now()}`,
    type: Number(item.senderUserId) === Number(currentUserId) ? 'self' : 'friend',
    text: item.content || '',
    timeText: formatTime(item.createdAt),
    failed: false,
    failReason: ''
  })).filter((item) => item.text)
}

function isSensitiveReject(error) {
  const message = String(error && error.message || '')
  return message.indexOf('敏感词') >= 0 || message.indexOf('451') >= 0
}

function formatTemplate(template, values = {}) {
  return String(template || '').replace(/\{(\w+)\}/g, (_, key) => values[key] == null ? '' : String(values[key]))
}

function normalizeMyMessageConfig(source = {}) {
  const friend = source.friend || {}
  const texts = source.texts || {}

  return {
    pageTitle: source.pageTitle || '',
    onlineText: source.onlineText || '',
    friend: {
      defaultInitials: friend.defaultInitials || '',
      defaultName: friend.defaultName || '',
      defaultStatus: friend.defaultStatus || '',
      nameTemplate: friend.nameTemplate || ''
    },
    quickActions: Array.isArray(source.quickActions) ? source.quickActions : [],
    texts
  }
}

Page({
  data: {
    pageTitle: '',
    onlineText: '',
    navItems: [
      { name: '我的', key: 'mine' },
      { name: '元宇宙', key: 'metaverse' },
      { name: '地图', key: 'map' },
      { name: '消息', key: 'message' },
      { name: '首页', key: 'home' }
    ],
    friend: {
      initials: '',
      name: '',
      status: ''
    },
    quickActions: [],
    texts: {},
    pageConfig: normalizeMyMessageConfig(),
    messages: [],
    gameId: 0,
    sourceGameId: 0,
    targetUserId: 0,
    targetName: '',
    chatMode: '',
    currentUserId: 0,
    loading: false,
    chatScrollTop: 0
  },

  onLoad(options = {}) {
    const gameId = toPositiveInt(options.gameId)
    const sourceGameId = toPositiveInt(options.sourceGameId)
    const targetUserId = toPositiveInt(options.targetUserId)
    const chatMode = String(options.mode || '').trim().toLowerCase()
    this.setData({
      gameId,
      sourceGameId,
      targetUserId,
      targetName: String(options.targetName || '').trim(),
      chatMode
    })
    this.loadPageConfig()
    this.loadConversation()
  },

  async loadPageConfig() {
    try {
      const config = await this.resolvePageConfig(true)
      this.setData({
        pageTitle: config.pageTitle,
        onlineText: config.onlineText,
        quickActions: config.quickActions,
        texts: config.texts,
        pageConfig: config,
        friend: {
          initials: config.friend.defaultInitials,
          name: config.friend.defaultName,
          status: config.friend.defaultStatus
        }
      })
    } catch (error) {
      this.showInfo(error && error.message ? error.message : this.textOf('loadFailedText'))
    }
  },

  async resolvePageConfig(force = false) {
    const current = this.data.pageConfig || normalizeMyMessageConfig()
    if (!force && current.pageTitle) {
      return current
    }

    return normalizeMyMessageConfig(await messageService.getMessageMyConfig())
  },

  async loadConversation() {
    if (this.data.chatMode === 'private') {
      await this.loadPrivateConversation()
      return
    }

    if (!this.data.gameId) {
      return
    }

    this.setData({ loading: true })

    try {
      const [config, room, messagesResp] = await Promise.all([
        this.resolvePageConfig(),
        imService.getChatRoom(this.data.gameId),
        imService.getMessages(this.data.gameId)
      ])
      const memberIDs = Array.isArray(room.memberIds) ? room.memberIds : []
      const currentUserId = toPositiveInt(room.currentUserId || room.userId)
      const peerID = memberIDs.find((id) => Number(id) !== Number(currentUserId)) || memberIDs[0] || 0
      const friendName = peerID
        ? formatTemplate(config.friend.nameTemplate, { userId: peerID })
        : config.friend.defaultName
      const rawMessages = messagesResp.items || messagesResp.messages || messagesResp.list || []

      this.setData({
        loading: false,
        pageTitle: config.pageTitle,
        onlineText: config.onlineText,
        quickActions: config.quickActions,
        texts: config.texts,
        pageConfig: config,
        currentUserId,
        friend: {
          initials: initialsFromName(friendName || config.friend.defaultInitials),
          name: friendName,
          status: room.status || config.friend.defaultStatus
        },
        messages: normalizeChatMessages(rawMessages, currentUserId),
        chatScrollTop: 999999
      })
    } catch (error) {
      this.setData({ loading: false })
      this.showInfo(error && error.message ? error.message : this.textOf('loadFailedText'))
    }
  },

  async loadPrivateConversation() {
    if (!this.data.targetUserId) {
      this.showInfo('缺少私聊对象')
      return
    }

    this.setData({ loading: true })

    try {
      const [config, conversationResp] = await Promise.all([
        this.resolvePageConfig(),
        imService.getPrivateMessages(this.data.targetUserId, {
          sourceGameId: this.data.sourceGameId || ''
        })
      ])
      const targetUser = conversationResp.targetUser || {}
      const currentUserId = toPositiveInt(conversationResp.currentUserId)
      const targetName = this.data.targetName || targetUser.nickname || targetUser.name || (this.data.targetUserId
        ? formatTemplate(config.friend.nameTemplate, { userId: this.data.targetUserId })
        : config.friend.defaultName)
      const rawMessages = conversationResp.items || conversationResp.messages || []

      this.setData({
        loading: false,
        pageTitle: '打招呼',
        onlineText: config.onlineText,
        quickActions: config.quickActions,
        texts: config.texts,
        pageConfig: config,
        currentUserId,
        friend: {
          initials: initialsFromName(targetName || config.friend.defaultInitials),
          name: targetName,
          status: config.friend.defaultStatus
        },
        messages: normalizeChatMessages(rawMessages, currentUserId),
        chatScrollTop: 999999
      })
    } catch (error) {
      this.setData({ loading: false })
      this.showInfo(error && error.message ? error.message : this.textOf('loadFailedText'))
    }
  },

  onQuickActionTap(event) {
    const key = event.currentTarget.dataset.key

    if (key === 'friend') {
      navigateShellRoute(ROUTES.profile, {
        currentRoute: ROUTES.messageMy
      })
      return
    }

    if (key === 'location') {
      const suffix = this.data.gameId ? `?gameId=${encodeURIComponent(this.data.gameId)}&mode=route` : ''
      navigateShellRoute(`${ROUTES.map}${suffix}`, {
        currentRoute: ROUTES.messageMy
      })
      return
    }

    if (key === 'greet') {
      this.sendQuickText(this.quickActionMessage(key))
      return
    }

    if (key === 'card') {
      this.sendQuickText(this.quickActionMessage(key))
      return
    }

    this.showInfo(this.textOf('actionMissingText'))
  },

  quickActionMessage(key) {
    const action = this.data.quickActions.find((item) => item.key === key) || {}
    return action.messageText || ''
  },

  async sendQuickText(text) {
    if (!text) {
      this.showInfo(this.textOf('actionMissingText'))
      return
    }

    if (this.data.chatMode === 'private') {
      await this.sendPrivateText(text)
      return
    }

    if (this.data.gameId) {
      try {
        await imService.sendMessage(this.data.gameId, {
          messageType: 'text',
          content: text
        })
      } catch (error) {
        if (isSensitiveReject(error)) {
          this.appendSelfMessage(text, { failed: true })
          return
        }
        this.showInfo(error && error.message ? error.message : this.textOf('sendFailedText'))
        return
      }
    }

    this.appendSelfMessage(text)
  },

  async onSendMessage(event) {
    const value = String(event.detail && event.detail.value || '').trim()

    if (!value) {
      return
    }

    if (this.data.chatMode === 'private') {
      await this.sendPrivateText(value)
      return
    }

    if (this.data.gameId) {
      try {
        await imService.sendMessage(this.data.gameId, {
          messageType: 'text',
          content: value
        })
      } catch (error) {
        if (isSensitiveReject(error)) {
          this.appendSelfMessage(value, { failed: true })
          return
        }
        this.showInfo(error && error.message ? error.message : this.textOf('sendFailedText'))
        return
      }
    }

    this.appendSelfMessage(value)
  },

  onRecordStart() {
    this.showInfo(this.textOf('recordStartText'))
  },

  onRecordStop() {
    this.showInfo(this.textOf('recordStopText'))
  },

  onRecordError() {
    this.showInfo(this.textOf('recordErrorText'))
  },

  onChooseImage(event) {
    this.sendMediaMessage('image', event)
  },

  onChooseFile(event) {
    this.sendMediaMessage('file', event)
  },

  async sendMediaMessage(messageType, event) {
    if (this.data.chatMode === 'private') {
      this.showInfo('私人聊天暂时仅支持文字')
      return
    }

    if (!this.data.gameId) {
      this.showInfo(this.textOf('fileEntryMissingText'))
      return
    }

    try {
      const result = await chatMedia.sendChosenFile(this.data.gameId, event, messageType)

      this.appendSelfMessage(result.text)
    } catch (error) {
      this.showInfo(error && error.message ? error.message : this.textOf('sendFailedText'))
    }
  },

  async sendPrivateText(text) {
    if (!this.data.targetUserId) {
      this.showInfo('缺少私聊对象')
      return
    }

    try {
      const result = await imService.sendPrivateMessage(this.data.targetUserId, {
        sourceGameId: this.data.sourceGameId || 0,
        messageType: 'text',
        content: text
      })
      const message = result.message || {}
      this.appendSelfMessage(message.content || text)
    } catch (error) {
      if (isSensitiveReject(error)) {
        this.appendSelfMessage(text, { failed: true })
        return
      }
      this.showInfo(error && error.message ? error.message : this.textOf('sendFailedText'))
    }
  },

  appendSelfMessage(text, options = {}) {
    this.setData({
      messages: this.data.messages.concat({
        id: `m-${Date.now()}`,
        type: 'self',
        text,
        timeText: this.textOf('justNowText'),
        failed: Boolean(options.failed),
        failReason: options.failed ? '消息涉及敏感词，已拒绝发送' : ''
      }),
      chatScrollTop: 999999
    })
  },

  onFailedMessageTap(event) {
    this.showInfo(event.currentTarget.dataset.reason || '消息发送失败')
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}
    navigateShellKey(key, {
      currentRoute: ROUTES.messageMy
    })
  },

  textOf(key) {
    return this.data.texts[key] || ''
  },

  showInfo(title) {
    if (!title) {
      return
    }
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
