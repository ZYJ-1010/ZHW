const imService = require('../../../services/im')
const fileService = require('../../../services/file')
const gameService = require('../../../services/game')
const { ROUTES } = require('../../../config/routes')
const toast = require('../../../utils/toast')
const { navigateShellBack, navigateShellRoute } = require('../../../utils/shell-nav')

const READONLY_STATUSES = ['archived', 'readonly', 'ended', 'finished', 'completed', 'canceled']
const ROOM_REFRESH_INTERVAL_MS = 3000
const SOCKET_RECONNECT_DELAY_MS = 1500

function toPositiveInt(value) {
  const number = Number(value)
  return Number.isInteger(number) && number > 0 ? number : 0
}

function safeText(value, fallback = '') {
  const text = String(value || '').trim()
  return text || fallback
}

function pad2(value) {
  return String(value).padStart(2, '0')
}

function formatDateMinute(date) {
  return [
    date.getFullYear(),
    pad2(date.getMonth() + 1),
    pad2(date.getDate())
  ].join('-') + ' ' + [
    pad2(date.getHours()),
    pad2(date.getMinutes())
  ].join(':')
}

function formatTime(value) {
  if (!value) {
    return ''
  }

  const text = String(value).trim()
  if (!text) {
    return ''
  }

  if (/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}/.test(text)) {
    return text.slice(0, 16)
  }

  const parsed = new Date(text)
  if (!Number.isNaN(parsed.getTime())) {
    return formatDateMinute(parsed)
  }

  return text.replace('T', ' ').replace(/:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?$/, '')
}

function avatarTextFromName(name, userId) {
  const text = safeText(name)
  if (text) {
    return text.slice(0, 1)
  }

  const id = toPositiveInt(userId)
  return id ? `U${id}`.slice(-2) : '局'
}

function roomStatusText(status) {
  const map = {
    active: '进行中',
    archived: '已归档',
    readonly: '只读',
    ended: '已结束',
    finished: '已结束',
    completed: '已结束',
    canceled: '已取消'
  }

  return map[status] || '进行中'
}

function isReadOnlyRoom(room = {}) {
  return READONLY_STATUSES.indexOf(safeText(room.status || room.gameStatus)) >= 0
}

function isSystemMessage(type) {
  return [
    'system',
    'invite',
    'invitation',
    'game_invite',
    'group_success',
    'member_join',
    'time_location_changed',
    'service_confirm_remind'
  ].indexOf(type) >= 0
}

function tryParseObject(content) {
  if (!content || typeof content !== 'string') {
    return null
  }

  try {
    const parsed = JSON.parse(content)
    return parsed && typeof parsed === 'object' ? parsed : null
  } catch (error) {
    return null
  }
}

function systemTitle(type, payload) {
  if (payload && payload.title) {
    return payload.title
  }

  const map = {
    invite: '收到组局邀请',
    invitation: '收到组局邀请',
    game_invite: '收到组局邀请',
    group_success: '组局成功',
    member_join: '成员加入',
    time_location_changed: '时间地点变更'
  }

  return map[type] || '局内通知'
}

function pickGameTitle(gameDetail) {
  if (!gameDetail || typeof gameDetail !== 'object') {
    return ''
  }

  const detailDisplay = gameDetail.detailDisplay || {}
  const header = detailDisplay.header || {}
  const game = gameDetail.game || gameDetail.detail || {}

  return safeText(gameDetail.title || game.title || detailDisplay.title || header.title)
}

function normalizeMembers(room = {}) {
  const members = Array.isArray(room.members) ? room.members : []
  const memberIds = Array.isArray(room.memberIds) ? room.memberIds : []
  const map = {}

  members.forEach((member) => {
    const userId = toPositiveInt(member.userId || member.id)
    if (!userId) {
      return
    }

    map[userId] = {
      userId,
      name: safeText(member.name || member.nickname, `成员${userId}`),
      roleText: safeText(member.roleText || member.roleName, '成员'),
      avatarText: safeText(member.avatarText, avatarTextFromName(member.name || member.nickname, userId)),
      avatarSrc: safeText(member.avatarSrc || member.avatarUrl)
    }
  })

  memberIds.forEach((userId) => {
    const id = toPositiveInt(userId)
    if (!id || map[id]) {
      return
    }

    map[id] = {
      userId: id,
      name: `成员${id}`,
      roleText: '成员',
      avatarText: avatarTextFromName('', id),
      avatarSrc: ''
    }
  })

  return map
}

function buildMemberPreview(room = {}, memberMap = {}) {
  const memberIds = Array.isArray(room.memberIds) ? room.memberIds : Object.keys(memberMap)

  return memberIds.map((userId) => memberMap[toPositiveInt(userId)]).filter(Boolean).slice(0, 3)
}

function memberForMessage(item, memberMap) {
  const userId = toPositiveInt(item.senderUserId || item.senderId)
  const member = memberMap[userId] || {}

  if (!userId) {
    return {
      userId: 0,
      name: '组局助手',
      roleText: '系统',
      avatarText: '助',
      avatarSrc: ''
    }
  }

  return {
    userId,
    name: safeText(item.senderName, safeText(member.name, `成员${userId}`)),
    roleText: safeText(item.senderRoleText || item.roleText, safeText(member.roleText, '成员')),
    avatarText: safeText(item.avatarText || item.senderAvatarText, safeText(member.avatarText, avatarTextFromName(member.name, userId))),
    avatarSrc: safeText(
      item.avatarSrc || item.avatarUrl || item.senderAvatarSrc || item.senderAvatarUrl,
      safeText(member.avatarSrc)
    )
  }
}

function buildSystemCard(message, type, payload = {}) {
  const member = payload.player || payload.member || {}
  const game = payload.game || {}

  return {
    assistantName: safeText(payload.assistantName, '组局助手'),
    title: systemTitle(type, payload),
    subtitle: safeText(payload.subtitle, safeText(message.content, '局内状态已更新')),
    guideLabel: safeText(payload.guideLabel, '领路人'),
    guideName: safeText(payload.guideName, '待同步'),
    playerInfoLabel: safeText(payload.playerInfoLabel, '成员信息'),
    player: {
      name: safeText(member.name, safeText(payload.memberName, '局内成员')),
      desc: safeText(member.desc, safeText(payload.memberDesc, '成员资料以后端群资料为准')),
      avatarText: safeText(member.avatarText, '员'),
      avatarSrc: safeText(member.avatarSrc)
    },
    game: {
      dateText: safeText(game.dateText, safeText(payload.dateText, '时间以后端同步为准')),
      location: safeText(game.location, safeText(payload.location, '地点以后端同步为准'))
    },
    acceptButtonText: safeText(payload.acceptButtonText),
    declineButtonText: safeText(payload.declineButtonText),
    actionRoute: safeText(payload.actionRoute)
  }
}

function buildRoomCard(room, session, readOnly, titleText) {
  const memberCount = Array.isArray(room.memberIds) ? room.memberIds.length : 0
  const title = safeText(titleText || room.title || session.title, '局')
  const statusText = readOnly ? '已结束' : '进行中'

  return {
    id: `room-card-${room.id || room.gameId || 'current'}`,
    kind: 'system',
    messageType: 'group_success',
    createdAtText: '',
    card: {
      assistantName: '组局助手',
      title: `${title}群聊`,
      subtitle: readOnly ? '本局已结束，群聊切换为只读归档' : '局内消息仅成员可见',
      guideLabel: '局状态',
      guideName: statusText,
      playerInfoLabel: '群聊成员',
      player: {
        name: memberCount ? `${memberCount} 位成员` : '成员待同步',
        desc: '成员消息会显示姓名、角色和头像标识',
        avatarText: '局',
        avatarSrc: ''
      },
      game: {
        dateText: readOnly ? '成员仅可查看历史消息' : '时间变更会通过系统卡片同步',
        location: '地点变更会通过系统卡片同步'
      },
      acceptButtonText: '',
      declineButtonText: ''
    }
  }
}

function normalizeMessage(item, currentUserId, memberMap) {
  const messageType = safeText(item.messageType || item.type, 'text')
  const sender = memberForMessage(item, memberMap)
  const isSelf = currentUserId > 0 && sender.userId === currentUserId
  const payload = tryParseObject(item.content)

  if (isSystemMessage(messageType)) {
    return {
      id: item.id || `system-${Date.now()}`,
      kind: 'system',
      messageType,
      createdAtText: formatTime(item.createdAtText || item.createdAt),
      card: buildSystemCard(item, messageType, payload || {})
    }
  }

  const fileName = safeText(item.fileName || item.content, messageType === 'image' ? '图片消息' : (messageType === 'voice' ? '语音消息' : '局内文件'))

  return {
    id: item.id || `message-${Date.now()}`,
    kind: messageType === 'image' || messageType === 'file' || messageType === 'voice' ? 'media' : 'text',
    messageType,
    isSelf,
    rowClass: isSelf ? 'self' : 'other',
    senderUserId: sender.userId,
    senderName: sender.name,
    senderRoleText: sender.roleText,
    avatarText: sender.avatarText,
    avatarSrc: sender.avatarSrc,
    contentText: safeText(item.content),
    fileName,
    fileId: toPositiveInt(item.fileId),
    createdAtText: formatTime(item.createdAtText || item.createdAt),
    statusText: item.status === 'risk_flagged' ? '待审核' : ''
  }
}

function messageTimeValue(item = {}) {
  const value = item.createdAt || item.created_at || item.createdAtText || ''
  const timestamp = Date.parse(String(value).replace(' ', 'T'))
  return Number.isFinite(timestamp) ? timestamp : 0
}

function messageDateLabel(item = {}) {
  const timestamp = messageTimeValue(item)
  if (!timestamp) {
    return ''
  }
  const date = new Date(timestamp)
  return `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日`
}

function normalizeMessages(items, currentUserId, memberMap) {
  const source = (Array.isArray(items) ? items : []).slice().sort((left, right) => {
    const timeDiff = messageTimeValue(left) - messageTimeValue(right)
    if (timeDiff !== 0) {
      return timeDiff
    }
    return Number(left.id || 0) - Number(right.id || 0)
  })
  let previousDate = ''
  return source.map((item) => {
    const normalized = normalizeMessage(item, currentUserId, memberMap)
    const dateText = messageDateLabel(item)
    normalized.showDateDivider = Boolean(dateText && dateText !== previousDate)
    normalized.dateDividerText = dateText
    previousDate = dateText || previousDate
    return normalized
  })
}

function firstChosenFile(result = {}) {
  const files = result.tempFiles || result.files || []
  const paths = result.tempFilePaths || []
  const file = files[0] || {}
  const path = file.tempFilePath || file.path || paths[0] || ''
  const name = file.name || file.fileName || String(path).split('/').filter(Boolean).pop() || ''
  return {
    path,
    name,
    size: Number(file.size || 1) || 1,
    width: Number(file.width || 0) || 0,
    height: Number(file.height || 0) || 0,
    durationMs: 0
  }
}

function recordedFile(result = {}) {
  const path = result.tempFilePath || result.filePath || ''
  return {
    path,
    name: path ? (path.split('/').filter(Boolean).pop() || `voice-${Date.now()}.mp3`) : '',
    size: Number(result.fileSize || result.size || 1) || 1,
    durationMs: Number(result.duration || 0) || 0
  }
}

function isSensitiveReject(error) {
  const message = safeText(error && error.message)
  return message.indexOf('敏感词') >= 0 || message.indexOf('451') >= 0
}

function buildFailedTextMessage(content) {
  return {
    id: `failed-${Date.now()}`,
    kind: 'text',
    messageType: 'text',
    isSelf: true,
    rowClass: 'self failed',
    senderUserId: 0,
    senderName: '我',
    senderRoleText: '玩家',
    avatarText: '我',
    avatarSrc: '',
    contentText: content,
    createdAtText: '刚刚',
    statusText: '已拒绝发送',
    failed: true,
    failReason: '消息涉及敏感词，已拒绝发送'
  }
}

Page({
  data: {
    pageTitle: '局内群聊',
    title: '局',
    desc: '正在加载群聊',
    memberText: '成员',
    gameId: 0,
    roomId: 0,
    roomStatus: '',
    roomEngine: '',
    openIMGroupId: '',
    loading: false,
    sending: false,
    uploading: false,
    readOnly: false,
    readOnlyText: '',
    inputText: '',
    messages: [],
    memberPreview: [],
    collaborationEntry: null,
    scrollAnchor: ''
  },

  onLoad(options = {}) {
    this.gameId = toPositiveInt(options.gameId)
    this.prefillText = safeText(options.prefill || options.message)
    this.hasShown = false
    this.hasLoadedRoom = false
    this.shouldStickToLatest = true
    this.lastMessageScrollTop = 0
    this.socketReconnectAttempts = 0
    this.pageVisible = false
    this.initialRoomLoad = this.loadRoom({ forceScroll: true })
  },

  async onShow() {
    this.pageVisible = true
    const loaded = this.hasShown
      ? await this.loadRoom()
      : await this.initialRoomLoad
    this.hasShown = true

    if (!this.pageVisible) {
      return
    }

    if (loaded) {
      this.startSocket()
    }
    this.startRoomRefresh()
  },

  onHide() {
    this.pageVisible = false
    this.stopRoomRefresh()
    this.stopSocket()
  },

  onUnload() {
    this.pageVisible = false
    this.stopRoomRefresh()
    this.stopSocket()
    if (this.audioContext) {
      this.audioContext.stop()
      this.audioContext.destroy()
      this.audioContext = null
    }
  },

  startRoomRefresh() {
    this.stopRoomRefresh()
    if (!this.pageVisible) {
      return
    }
    this.roomRefreshTimer = setInterval(() => {
      if (!this.imSocket && !this.data.loading && !this.data.sending && !this.data.uploading) {
        this.loadRoom().then((loaded) => {
          if (loaded) {
            this.startSocket()
          }
        })
      }
    }, ROOM_REFRESH_INTERVAL_MS)
  },

  stopRoomRefresh() {
    if (this.roomRefreshTimer) {
      clearInterval(this.roomRefreshTimer)
      this.roomRefreshTimer = null
    }
  },

  startSocket() {
    if (!this.pageVisible || !this.gameId || !this.hasLoadedRoom || this.imSocket || typeof imService.connectGameSocket !== 'function') {
      return
    }
    this.socketStopping = false
    const socket = imService.connectGameSocket(this.gameId, {
      onOpen: () => {
        this.socketReconnectAttempts = 0
        this.setData({ socketConnected: true })
        this.stopRoomRefresh()
      },
      onConnected: () => {
        this.setData({ socketConnected: true })
      },
      onMessage: (message) => this.appendSocketMessage(message),
      onError: (error) => {
        // 敏感词拒绝是单条消息业务错误，连接本身仍然有效；发送 Promise
        // 会把 45101 返回给 onSendTap，由失败消息气泡展示具体原因。
        if (error && Number(error.code) === 45101) {
          return
        }
        this.setData({ socketConnected: false })
      },
      onClose: () => {
        this.imSocket = null
        this.setData({ socketConnected: false })
        if (!this.socketStopping && this.pageVisible) {
          this.startRoomRefresh()
          this.scheduleSocketReconnect()
        }
      }
    })
    if (socket) {
      this.imSocket = socket
      this.socketStopping = false
    }
  },

  scheduleSocketReconnect() {
    if (this.socketReconnectTimer || this.socketStopping || !this.pageVisible || !this.hasLoadedRoom || !this.gameId) {
      return
    }
    this.socketReconnectAttempts = Math.min((this.socketReconnectAttempts || 0) + 1, 5)
    const delay = SOCKET_RECONNECT_DELAY_MS * this.socketReconnectAttempts
    this.socketReconnectTimer = setTimeout(() => {
      this.socketReconnectTimer = null
      this.startSocket()
    }, delay)
  },

  stopSocket() {
    this.socketStopping = true
    if (this.socketReconnectTimer) {
      clearTimeout(this.socketReconnectTimer)
      this.socketReconnectTimer = null
    }
    if (this.imSocket) {
      this.imSocket.close()
      this.imSocket = null
    }
    this.setData({ socketConnected: false })
  },

  appendSocketMessage(message) {
    if (!message || Number(message.gameId || this.gameId) !== this.gameId) {
      return
    }
    const messageID = String(message.id || message.messageId || '')
    const current = Array.isArray(this.data.messages) ? this.data.messages : []
    if (messageID && current.some((item) => String(item.id) === messageID)) {
      return
    }
    const normalized = normalizeMessage(message, this.currentUserId || 0, this.memberMap || {})
    const previous = current[current.length - 1]
    const dateText = messageDateLabel(message)
    normalized.showDateDivider = Boolean(dateText && (!previous || previous.dateDividerText !== dateText))
    normalized.dateDividerText = dateText
    const messages = current.concat(normalized)
    this.setData({
      messages,
      scrollAnchor: `message-${messages.length - 1}`
    })
    this.shouldStickToLatest = true
  },

  onMessageScroll(event) {
    const detail = event.detail || {}
    const scrollTop = Number(detail.scrollTop || 0)
    if (!Number.isFinite(scrollTop)) {
      return
    }

    if (scrollTop + 2 < (this.lastMessageScrollTop || 0)) {
      this.shouldStickToLatest = false
    }
    this.lastMessageScrollTop = scrollTop
  },

  onMessageScrollToLower() {
    this.shouldStickToLatest = true
  },

  onBackTap() {
    if (navigateShellBack()) {
      return
    }

    if (this.gameId) {
      navigateShellRoute(`${ROUTES.gameDetail}?gameId=${encodeURIComponent(this.gameId)}`, { reuseExisting: false })
      return
    }

    navigateShellRoute(ROUTES.message, { reuseExisting: false })
  },

  async loadRoom(options = {}) {
    if (!this.gameId) {
      this.setData({
        title: '局',
        pageTitle: '局内群聊',
        desc: '缺少局信息，无法进入群聊'
      })
      return false
    }

    this.setData({ loading: true })

    try {
      const [room, session, messagesResp, gameDetail] = await Promise.all([
        imService.getChatRoom(this.gameId),
        imService.getRoomByGame(this.gameId),
        imService.getMessages(this.gameId),
        gameService.getGameDetail(this.gameId).catch(() => null)
      ])
      const list = messagesResp.items || messagesResp.messages || messagesResp.list || messagesResp.data || messagesResp
      const roomId = toPositiveInt(room.id || session.roomId)
      const currentUserId = toPositiveInt(room.currentUserId || session.currentUserId)
      const memberCount = Array.isArray(room.memberIds) ? room.memberIds.length : 0
      const readOnly = isReadOnlyRoom(room)
      const memberMap = normalizeMembers(room)
      this.memberMap = memberMap
      this.currentUserId = currentUserId
      const memberPreview = buildMemberPreview(room, memberMap)
      const roomTitle = safeText(session.title || room.title || pickGameTitle(gameDetail), '局')
      const messages = [
        buildRoomCard(room, session, readOnly, roomTitle),
        ...normalizeMessages(list, currentUserId, memberMap)
      ]
      const shouldScrollToLatest = Boolean(options.forceScroll) || !this.hasLoadedRoom || this.shouldStickToLatest

      const nextData = {
        loading: false,
        title: roomTitle,
        pageTitle: `${roomTitle}群聊`,
        desc: readOnly ? '本局已结束，只能查看历史消息' : '局内消息仅成员可见',
        memberText: memberCount ? `成员${memberCount}人` : '成员',
        memberPreview,
        gameId: this.gameId,
        roomId,
        roomStatus: room.status || '',
        collaborationEntry: room.collaborationEntry || null,
        readOnly,
        readOnlyText: readOnly ? '本局已结束，成员只能查看历史消息，不能再发送消息' : '',
        roomEngine: session.engine || room.engine || '',
        openIMGroupId: session.openIMGroupId || room.openIMGroupId || '',
        messages,
        inputText: this.prefillText || this.data.inputText
      }
      if (shouldScrollToLatest) {
        nextData.scrollAnchor = messages.length ? `message-${messages.length - 1}` : ''
      }
      this.setData(nextData)
      this.hasLoadedRoom = true
      if (shouldScrollToLatest) {
        this.shouldStickToLatest = true
      }
      return true
    } catch (error) {
      this.setData({ loading: false })
      toast.info(error && error.message ? error.message : '加载局内群聊失败')
      return false
    }
  },

  onRecordStart() {
    toast.info('再次点击语音按钮结束录音')
  },

  onRecordStop(event) {
    const file = recordedFile(event.detail || {})
    if (!file.path) {
      toast.info('未获取到录音文件')
      return
    }
    this.sendPickedFile(file, 'voice')
  },

  onRecordError() {
    toast.info('录音失败')
  },

  onSendMessage(event) {
    const value = event.detail && event.detail.value
    return this.onSendTap(value || '')
  },

  onChooseImage(event) {
    this.sendPickedFile(firstChosenFile(event.detail || {}), 'image')
  },

  onChooseFile(event) {
    this.sendPickedFile(firstChosenFile(event.detail || {}), 'file')
  },

  async onMediaTap(event) {
    const dataset = event.currentTarget.dataset || {}
    const fileId = toPositiveInt(dataset.fileId)
    const messageType = safeText(dataset.messageType)
    const fileName = safeText(dataset.fileName, '局内文件')
    if (!fileId) {
      toast.info('文件暂未同步完成')
      return
    }
    try {
      const url = await fileService.getDownloadURL(fileId)
      if (!url || typeof url !== 'string') {
        throw new Error('文件地址无效')
      }
      if (messageType === 'voice') {
        if (this.audioContext) {
          this.audioContext.stop()
          this.audioContext.destroy()
        }
        this.audioContext = wx.createInnerAudioContext()
        this.audioContext.src = url
        this.audioContext.onError(() => toast.info('语音播放失败'))
        this.audioContext.play()
        return
      }
      if (messageType === 'image') {
        wx.previewImage({ urls: [url], fail: () => toast.info('图片打开失败') })
        return
      }
      wx.downloadFile({
        url,
        success: (result) => {
          if (result.statusCode !== 200) {
            toast.info('文件下载失败')
            return
          }
          wx.openDocument({
            filePath: result.tempFilePath,
            fileType: fileName.includes('.') ? fileName.split('.').pop().toLowerCase() : '',
            showMenu: true,
            fail: () => toast.info('当前文件暂不支持预览，请稍后重试')
          })
        },
        fail: () => toast.info('文件下载失败')
      })
    } catch (error) {
      toast.info(error && error.message ? error.message : '获取文件失败')
    }
  },

  onSystemActionTap(event) {
    const action = event.currentTarget.dataset.action || ''
    const route = safeText(event.currentTarget.dataset.route)

    if (route) {
      navigateShellRoute(route.startsWith('/') ? route : `/${route}`, {
        reuseExisting: false
      })
      return
    }

    toast.info(action === 'accept' ? '已记录确认参加意向' : '已记录婉拒意向')
  },

  async onSendTap(value) {
    const source = typeof value === 'string' ? value : this.data.inputText
    const content = String(source || '').trim()

    if (!content) {
      toast.info('请输入消息')
      return
    }

    if (this.data.readOnly) {
      toast.info('本局已结束，只能查看历史消息')
      return
    }

    if (!this.gameId) {
      toast.info('缺少局 ID')
      return
    }

    this.setData({ sending: true })

    try {
      const payload = {
        messageType: 'text',
        content
      }
      const sentViaSocket = Boolean(this.imSocket && this.imSocket.isOpen())
      if (sentViaSocket) {
        await this.imSocket.sendMessage(payload)
      } else {
        await imService.sendMessage(this.gameId, payload)
      }
      this.setData({ sending: false, inputText: '' })
      this.shouldStickToLatest = true
      if (!sentViaSocket) {
        await this.loadRoom({ forceScroll: true })
      }
    } catch (error) {
      if (isSensitiveReject(error)) {
        const messages = this.data.messages.concat(buildFailedTextMessage(content))
        this.shouldStickToLatest = true
        this.setData({
          sending: false,
          inputText: '',
          messages,
          scrollAnchor: `message-${messages.length - 1}`
        })
        return
      }
      this.setData({ sending: false })
      toast.info(error && error.message ? error.message : '发送失败')
    }
  },

  onFailedMessageTap(event) {
    const reason = event.currentTarget.dataset.reason || '消息发送失败'
    toast.info(reason)
  },

  async sendPickedFile(file, messageType) {
    if (!this.gameId) {
      toast.info('缺少局 ID')
      return
    }
    if (this.data.readOnly) {
      toast.info('本局已结束，只能查看历史消息')
      return
    }
    if (!file.path) {
      toast.info('未选择文件')
      return
    }
    this.setData({ uploading: true })
    try {
      const uploaded = await fileService.uploadSingleFileDetail(file, { bizType: 'chat_file', objectId: this.gameId })
      const fileId = uploaded && uploaded.fileId
      if (!fileId) {
        throw new Error('文件上传失败')
      }
      const payload = {
        messageType,
        content: file.name || (messageType === 'image' ? '图片消息' : (messageType === 'voice' ? '语音消息' : '局内文件')),
        fileId,
        width: Number(uploaded.width || 0) || 0,
        height: Number(uploaded.height || 0) || 0,
        durationMs: Math.max(0, Math.round(Number(uploaded.durationMs || 0) || 0))
      }
      const sentViaSocket = Boolean(this.imSocket && this.imSocket.isOpen())
      if (sentViaSocket) {
        await this.imSocket.sendMessage(payload)
      } else {
        await imService.sendMessage(this.gameId, payload)
      }
      this.setData({ uploading: false })
      this.shouldStickToLatest = true
      if (!sentViaSocket) {
        await this.loadRoom({ forceScroll: true })
      }
    } catch (error) {
      this.setData({ uploading: false })
      toast.info(error && error.message ? error.message : '文件发送失败')
    }
  },

  onMembersTap() {
    if (!this.gameId) {
      toast.info('缺少局 ID')
      return
    }

    navigateShellRoute(`${ROUTES.gameParticipants}?gameId=${encodeURIComponent(this.gameId)}`)
  },

  onCollaborationTap() {
    const entry = this.data.collaborationEntry || {}
    const route = safeText(entry.route, `${ROUTES.gameCollaboration}?gameId=${encodeURIComponent(this.gameId || 0)}`)

    if (!this.gameId) {
      toast.info('缺少局 ID')
      return
    }

    navigateShellRoute(route.startsWith('/') ? route : `/${route}`)
  },

  onGameInfoTap() {
    if (!this.gameId) {
      toast.info('缺少局 ID')
      return
    }

    navigateShellRoute(`${ROUTES.gameDetail}?gameId=${encodeURIComponent(this.gameId)}`)
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
      toast.info('缺少局 ID，无法发起举报')
      return
    }

    navigateShellRoute(`/pages/profile/system-management/report-center/index?gameId=${encodeURIComponent(this.gameId)}`)
  }
})
