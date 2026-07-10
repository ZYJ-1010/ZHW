const imService = require('../../../services/im')
const { ROUTES } = require('../../../config/routes')
const toast = require('../../../utils/toast')
const { navigateShellRoute } = require('../../../utils/shell-nav')

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

function normalizeMessages(items) {
  return (Array.isArray(items) ? items : []).map((item) => ({
    id: item.id,
    roomId: item.roomId,
    gameId: item.gameId,
    senderUserId: item.senderUserId,
    messageType: item.messageType,
    content: item.content,
    fileId: item.fileId,
    status: item.status,
    createdAtText: formatTime(item.createdAt)
  }))
}

Page({
  data: {
    title: '局内 IM',
    desc: '加载中',
    gameId: 0,
    roomId: 0,
    roomStatus: '',
    roomEngine: '',
    openIMGroupId: '',
    loading: false,
    sending: false,
    inputText: '',
    messages: []
  },

  onLoad(options = {}) {
    this.gameId = toPositiveInt(options.gameId)
    this.prefillText = String(options.prefill || options.message || '').trim()
    this.imSocket = null
    this.loadRoom()
  },

  onUnload() {
    this.closeSocket()
  },

  async loadRoom() {
    if (!this.gameId) {
      this.setData({ desc: '缺少 gameId，无法进入会话' })
      return
    }

    this.setData({ loading: true })

    try {
      const [room, session, messagesResp] = await Promise.all([
        imService.getChatRoom(this.gameId),
        imService.getRoomByGame(this.gameId),
        imService.getMessages(this.gameId)
      ])
      const messages = normalizeMessages(messagesResp.items || messagesResp.messages || messagesResp.list || messagesResp.data || messagesResp)
      const roomId = toPositiveInt(room.id || session.roomId)

      this.setData({
        loading: false,
        title: session.engine === 'openim' ? 'OpenIM 会话' : '局内 IM',
        desc: session.openIMGroupId ? `群组 ${session.openIMGroupId}` : `房间 ${roomId || ''}`,
        roomId,
        roomStatus: room.status || '',
        roomEngine: session.engine || '',
        openIMGroupId: session.openIMGroupId || '',
        messages,
        inputText: this.prefillText || this.data.inputText
      })
      this.connectSocket()
    } catch (error) {
      this.setData({ loading: false })
      toast.info(error && error.message ? error.message : '加载会话失败')
    }
  },

  connectSocket() {
    if (!this.gameId || this.imSocket && this.imSocket.isOpen()) {
      return
    }

    this.closeSocket()
    this.imSocket = imService.connectGameSocket(this.gameId, {
      onConnected: (room) => {
        const roomId = toPositiveInt(room && room.id)
        if (roomId && roomId !== this.data.roomId) {
          this.setData({ roomId })
        }
      },
      onMessage: (message) => {
        this.appendMessage(message)
      },
      onError: (error) => {
        if (error && error.message) {
          toast.info(error.message)
        }
      }
    })
  },

  closeSocket() {
    if (this.imSocket && typeof this.imSocket.close === 'function') {
      this.imSocket.close()
    }
    this.imSocket = null
  },

  appendMessage(message) {
    const item = normalizeMessages([message])[0]
    if (!item || !item.id) {
      return
    }
    const exists = this.data.messages.some((current) => current.id === item.id)
    if (exists) {
      return
    }
    this.setData({
      messages: this.data.messages.concat(item)
    })
  },

  onInput(event) {
    this.setData({ inputText: event.detail.value || '' })
  },

  async onSendTap() {
    const content = String(this.data.inputText || '').trim()

    if (!content) {
      toast.info('请输入消息')
      return
    }

    if (!this.gameId) {
      toast.info('缺少局ID')
      return
    }

    this.setData({ sending: true })

    try {
      const payload = {
        messageType: 'text',
        content
      }
      if (this.imSocket && this.imSocket.isOpen()) {
        await this.imSocket.sendMessage(payload)
      } else {
        const message = await imService.sendMessage(this.gameId, payload)
        this.appendMessage(message)
      }
      this.setData({ sending: false, inputText: '' })
    } catch (error) {
      this.setData({ sending: false })
      toast.info(error && error.message ? error.message : '发送失败')
    }
  },

  onBackTap() {
    const pages = getCurrentPages()

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    navigateShellRoute(ROUTES.gameManage)
  },

  onReportTap() {
    if (!this.gameId) {
      toast.info('缺少局ID，无法发起举报')
      return
    }

    navigateShellRoute(`/pages/profile/system-management/report-center/index?gameId=${encodeURIComponent(this.gameId)}`)
  }
})
