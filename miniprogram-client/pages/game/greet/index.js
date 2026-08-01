const { ROUTES } = require('../../../config/routes')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')
const { getSurnameInitials } = require('../../../utils/avatar')
const gameService = require('../../../services/game')
const imService = require('../../../services/im')
const chatMedia = require('../../../services/chat-media')

const DEFAULT_CONTACT = {
  name: '',
  realName: '',
  nickname: '',
  avatarText: 'WA'
}

function decodeQueryText(value = '') {
  try {
    return decodeURIComponent(value)
  } catch (error) {
    return value
  }
}

function getAvatarText(name = '') {
  return getSurnameInitials(name, DEFAULT_CONTACT.avatarText)
}

function getDisplayName(contact) {
  return contact.realName || contact.name || contact.nickname || DEFAULT_CONTACT.name || '联系人'
}

Page({
  data: {
    onlineText: '在线',
    pageTitle: '打招呼',
    contact: DEFAULT_CONTACT,
    groupInfo: {
      headerTitle: '组局信息',
      title: '',
      guideLabel: '领路人',
      guideName: '',
      expertLabel: '行家',
      miniProgramText: '小程序 · 真好玩',
      expert: {
        name: '',
        avatarText: '',
        desc: '',
        intro: '',
        tags: []
      },
      stats: [],
      confirmText: '查看详情并确认'
    },
    messageText: '',
    sentMessages: [],
    gameId: 0,
    canSendIM: false,
    canvasMinHeight: 1280
  },

  async onLoad(options = {}) {
    const realName = decodeQueryText(options.realName || '')
    const name = decodeQueryText(options.playerName || options.name || '')
    const nickname = decodeQueryText(options.nickname || '')
    const contactName = realName || name || nickname
    const gameId = Number(options.gameId || options.sourceGameId || 0)
    this.setData({
      gameId: Number.isInteger(gameId) && gameId > 0 ? gameId : 0
    })
    await this.loadGameAccess()

    if (!contactName) {
      return
    }

    const contact = {
      ...this.data.contact,
      realName,
      name: contactName,
      nickname,
      avatarText: getAvatarText(contactName)
    }

    this.setData({
      contact,
      pageTitle: getDisplayName(contact),
      'groupInfo.guideName': getDisplayName(contact)
    })

    if (options.message) {
      this.setData({
        messageText: decodeQueryText(options.message)
      })
    }
  },

  async loadGameAccess() {
    if (!this.data.gameId) {
      this.setData({ canSendIM: false })
      return
    }

    try {
      const detail = await gameService.getGameDetail(this.data.gameId)
      const relation = detail && detail.myRelation ? detail.myRelation : {}

      this.setData({
        canSendIM: relation.canEnterIM === true && relation.isCreator !== true
      })
    } catch (error) {
      this.setData({ canSendIM: false })
    }
  },

  onConfirmTap() {
    wx.showToast({
      title: '已确认组局信息',
      icon: 'none'
    })
  },

  async onSendMessage(event) {
    const value = String(event.detail && event.detail.value || '').trim()

    if (!value) {
      return
    }

    if (!this.data.gameId || !this.data.canSendIM) {
      this.showInfo('当前身份不可打招呼')
      return
    }

    try {
      await imService.sendMessage(this.data.gameId, {
        messageType: 'text',
        content: value
      })
    } catch (error) {
      this.showInfo(error && error.message ? error.message : '发送失败')
      return
    }

    const sentMessages = this.data.sentMessages.concat({
      id: `message-${Date.now()}`,
      text: value
    })

    this.setData({
      sentMessages,
      canvasMinHeight: 1280 + sentMessages.length * 120
    })
  },

  onRecordStart() {
    this.showInfo('开始录音')
  },

  onRecordStop(event) {
    this.sendMediaMessage('voice', event)
  },

  onRecordError() {
    this.showInfo('录音失败')
  },

  onChooseImage(event) {
    this.sendMediaMessage('image', event)
  },

  onChooseFile(event) {
    this.sendMediaMessage('file', event)
  },

  async sendMediaMessage(messageType, event) {
    if (!this.data.gameId || !this.data.canSendIM) {
      this.showInfo('请从局内消息入口发送文件')
      return
    }

    try {
      const result = await chatMedia.sendChosenFile(this.data.gameId, event, messageType)
      const sentMessages = this.data.sentMessages.concat({
        id: `message-${Date.now()}`,
        text: result.text
      })

      this.setData({
        sentMessages,
        canvasMinHeight: 1280 + sentMessages.length * 120
      })
    } catch (error) {
      this.showInfo(error && error.message ? error.message : '发送失败')
    }
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (navigateShellKey(key, {
      currentRoute: ROUTES.gameGreet
    })) {
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
      home: ROUTES.playerHome || ROUTES.home,
      metaverse: ROUTES.metaverse,
      map: ROUTES.map
    }

    this.navigateToRoute(routeMap[key])
  },

  navigateToRoute(route) {
    if (!route || route === ROUTES.gameGreet) {
      return
    }

    navigateShellRoute(route)
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
