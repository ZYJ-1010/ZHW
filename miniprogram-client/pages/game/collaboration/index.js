const { getLegacyWhiteFrameLayoutStyles } = require('../../../utils/adaptive-shell-layout')
const gameService = require('../../../services/game')
const { ROUTES } = require('../../../config/routes')
const { getSurnameInitials } = require('../../../utils/avatar')
const { navigateShellRoute } = require('../../../utils/shell-nav')
const { toUserMessage } = require('../../../utils/user-message')

const WHITE_CONTENT_LEFT_RPX = 2
const WHITE_CONTENT_TOP_RPX = 160
const WHITE_CONTENT_WIDTH_RPX = 750
const WHITE_DESIGN_FRAME_HEIGHT_PT = 810
const WHITE_DESIGN_BOTTOM_HEIGHT_PT = 78
const WHITE_NAV_TITLE_HEIGHT_RPX = 50
const WHITE_BACK_BUTTON_SIZE_RPX = 40

function getLayoutStyles() {
  const layout = getLegacyWhiteFrameLayoutStyles({
    contentLeftRpx: WHITE_CONTENT_LEFT_RPX,
    contentWidthRpx: WHITE_CONTENT_WIDTH_RPX,
    bottomHeightRatio: WHITE_DESIGN_BOTTOM_HEIGHT_PT / WHITE_DESIGN_FRAME_HEIGHT_PT,
    titleHeightRpx: WHITE_NAV_TITLE_HEIGHT_RPX,
    backSizeRpx: WHITE_BACK_BUTTON_SIZE_RPX
  })

  return Object.assign({}, layout, {
    footerStyle: `padding-bottom: ${34 + layout.safeBottomRpx}rpx;`
  })
}

Page({
  data: {
    layout: getLayoutStyles(),
    gameId: '',
    loading: false,
    ending: false,
    loadError: '',
    title: '局内协作',
    subtitle: '',
    progressPercent: 0,
    progressInput: '',
    progressNote: '',
    progressSaving: false,
    members: [],
    membersText: '',
    tasks: [],
    messages: [],
    actions: {
      canManageMembers: false,
      canEndGame: false,
      showFooter: false,
      manageText: '成员管理',
      endText: '结束本局',
      manageRoute: `${ROUTES.gameParticipants}?gameId=`,
      endConfirmRoute: `${ROUTES.gameDelivery}?gameId=`
    }
  },

  onLoad(options = {}) {
    const gameId = options.gameId || options.id || ''
    this.setData({ gameId })
    this.loadCollaboration(gameId)
  },

  onShow() {
    this.setData({
      layout: getLayoutStyles()
    })
  },

  onResize() {
    this.setData({
      layout: getLayoutStyles()
    })
  },

  async loadCollaboration(gameId = this.data.gameId) {
    if (!gameId) {
      return
    }

    this.setData({ loading: true, loadError: '' })

    try {
      const detail = await gameService.getGameCollaboration(gameId)
      const normalized = normalizeCollaboration(detail, gameId)
      this.setData(Object.assign({}, normalized, {
        loading: false,
        loadError: ''
      }))
    } catch (error) {
      this.setData({
        loading: false,
        loadError: toUserMessage(error && error.message, '协作数据加载失败'),
        subtitle: '',
        progressPercent: 0,
        members: [],
        membersText: '',
        tasks: [],
        messages: [],
        actions: {
          canManageMembers: false,
          canManageProgress: false,
          canEndGame: false,
          showFooter: false,
          manageText: '成员管理',
          endText: '结束本局',
          manageRoute: `${ROUTES.gameParticipants}?gameId=${gameId || ''}`,
          endConfirmRoute: `${ROUTES.gameDelivery}?gameId=${gameId || ''}`
        }
      })
    }
  },

  handleBack() {
    if (getCurrentPages().length > 1) {
      wx.navigateBack()
      return
    }

    navigateShellRoute('/pages/game/hall/index')
  },

  handleMemberManage() {
    const actions = this.data.actions || {}
    const route = actions.manageRoute || `${ROUTES.gameParticipants}?gameId=${this.data.gameId || ''}`

    if (!actions.canManageMembers) {
      wx.showToast({
        title: '当前账号无成员管理权限',
        icon: 'none'
      })
      return
    }

    if (route) {
      navigateShellRoute(route.startsWith('/') ? route : `/${route}`)
      return
    }

    const members = Array.isArray(this.data.members) ? this.data.members : []
    if (!members.length) {
      wx.showToast({
        title: '暂无成员',
        icon: 'none'
      })
      return
    }

    wx.showModal({
      title: '成员管理',
      content: members.map((item) => `${item.name}${item.roleText ? `（${item.roleText}）` : ''}`).join('\n'),
      showCancel: false,
      confirmText: '知道了'
    })
  },

  handleChatTap() {
    const gameId = this.data.gameId || ''

    if (!gameId) {
      wx.showToast({
        title: '缺少局信息，暂时无法进入聊天',
        icon: 'none'
      })
      return
    }

    navigateShellRoute(`/${ROUTES.imRoom}?gameId=${encodeURIComponent(gameId)}`)
  },

  onProgressInput(event) {
    this.setData({ progressInput: String(event.detail.value || '').replace(/[^0-9]/g, '').slice(0, 3) })
  },

  onProgressNoteInput(event) {
    this.setData({ progressNote: String(event.detail.value || '').slice(0, 500) })
  },

  async submitProgress() {
    const actions = this.data.actions || {}
    if (!actions.canManageProgress || !this.data.gameId || this.data.progressSaving) {
      return
    }
    const progress = Number(this.data.progressInput)
    const content = String(this.data.progressNote || '').trim()
    if (!Number.isInteger(progress) || progress < 0 || progress > 100) {
      wx.showToast({ title: '请输入 0 到 100 的进度', icon: 'none' })
      return
    }
    if (!content) {
      wx.showToast({ title: '请填写进度说明', icon: 'none' })
      return
    }
    if (progress < Number(this.data.progressPercent || 0)) {
      wx.showToast({ title: '进度不能低于当前进度', icon: 'none' })
      return
    }
    this.setData({ progressSaving: true })
    try {
      await gameService.createProgressFeedback(this.data.gameId, { progress, content })
      wx.showToast({ title: '进度已更新', icon: 'success' })
      this.setData({ progressInput: '', progressNote: '' })
      await this.loadCollaboration(this.data.gameId)
    } catch (error) {
      wx.showToast({ title: toUserMessage(error && error.message, '更新进度失败'), icon: 'none' })
    } finally {
      this.setData({ progressSaving: false })
    }
  },

  async handleEndSession() {
    const actions = this.data.actions || {}

    if (!actions.canEndGame) {
      wx.showToast({
        title: '当前状态不可结束',
        icon: 'none'
      })
      return
    }

    if (this.data.ending || !this.data.gameId) {
      return
    }

    const confirmed = await showEndConfirmation(actions)
    if (!confirmed) {
      return
    }

    this.setData({ ending: true })
    wx.showLoading({ title: '正在结束', mask: true })
    try {
      const result = await gameService.requestGameCompletion(this.data.gameId)
      wx.hideLoading()
      if (result.directReview) {
        wx.showToast({ title: '本局已结束，进入评价', icon: 'success' })
        const route = result.nextRoute || actions.reviewRoute || `${ROUTES.gameReview}?gameId=${this.data.gameId}`
        setTimeout(() => {
          navigateShellRoute(route.startsWith('/') ? route : `/${route}`)
        }, 300)
        return
      }
      wx.showToast({ title: '已通知成员确认完成', icon: 'success' })
      await this.loadCollaboration(this.data.gameId)
    } catch (error) {
      wx.hideLoading()
      wx.showToast({
        title: toUserMessage(error && error.message, '结束组局失败'),
        icon: 'none'
      })
    } finally {
      this.setData({ ending: false })
    }
  }
})

function normalizeCollaboration(data = {}, fallbackGameId = '') {
  const source = data && typeof data === 'object' ? data : {}
  const progress = source.progress || {}
  const tasks = Array.isArray(progress.tasks) && progress.tasks.length
    ? progress.tasks.map((item, index) => ({
      id: item.key || item.id || `task-${index}`,
      title: item.title || `阶段 ${index + 1}`,
      desc: item.desc || item.description || '',
      state: normalizeTaskState(item.state)
    }))
    : []
  const members = Array.isArray(source.members)
    ? source.members.map((item, index) => {
      const name = item.name || item.nickname || `成员${index + 1}`
      return {
        id: item.id || item.userId || `member-${index}`,
        name,
        roleText: item.roleText || item.roleLabel || item.role || '',
        roleClass: normalizeRoleClass(item.role),
        avatarText: item.avatarText || getSurnameInitials(name, 'ME')
      }
    })
    : []
  const messages = Array.isArray(source.messages)
    ? source.messages.map((item, index) => normalizeMessagePreview(item, index)).filter((item) => item.content)
    : []
  const progressPercent = clampPercent(Number(progress.percent || source.progressPercent || 0))
  const membersText = source.membersText || members.map((item) => `${item.name}${item.roleText ? ` (${item.roleText})` : ''}`).join(' · ')

  return {
    gameId: source.gameId || fallbackGameId,
    title: source.title || '局内协作',
    subtitle: [source.statusText, source.dayText].filter(Boolean).join(' · '),
    progressPercent,
    tasks,
    members,
    membersText,
    messages,
    actions: normalizeActions(Object.assign({
      canManageMembers: false,
      canManageProgress: false,
      canEndGame: false,
      showFooter: false,
      manageText: '成员管理',
      endText: '结束本局',
      manageRoute: `${ROUTES.gameParticipants}?gameId=${source.gameId || fallbackGameId}`,
      endConfirmRoute: `${ROUTES.gameDelivery}?gameId=${source.gameId || fallbackGameId}`
    }, source.actions || {}))
  }
}

function normalizeActions(actions = {}) {
  const normalized = Object.assign({}, actions, {
    canManageMembers: Boolean(actions.canManageMembers),
    canManageProgress: Boolean(actions.canManageProgress),
    canEndGame: Boolean(actions.canEndGame)
  })
  normalized.showFooter = Boolean(normalized.canManageMembers || normalized.canEndGame)
  normalized.manageText = normalized.manageText || '成员管理'
  normalized.endText = normalized.endText || '结束本局'
  return normalized
}

function normalizeTaskState(state) {
  const value = String(state || '').toLowerCase()
  return ['completed', 'active', 'pending'].indexOf(value) >= 0 ? value : 'pending'
}

function normalizeRoleClass(role) {
  const value = String(role || '').toLowerCase()
  if (value === 'expert') {
    return 'expert'
  }
  if (value === 'guide' || value === 'main_guide') {
    return 'guide'
  }
  return 'player'
}

function normalizeMessagePreview(item = {}, index) {
  const messageType = String(item.messageType || item.type || 'text').toLowerCase()
  const rawContent = String(item.content || '').trim()
  let name = item.senderName || item.name || item.nickname || '成员'
  let content = rawContent

  if (messageType === 'image') {
    content = '发来一张图片'
  } else if (messageType === 'file') {
    content = `发来文件：${item.fileName || rawContent || '未命名文件'}`
  } else if (messageType === 'service_confirm_remind' || messageType === 'system') {
    name = '组局助手'
    try {
      const payload = JSON.parse(rawContent)
      content = payload.subtitle || payload.title || '局内状态已更新'
    } catch (error) {
      content = rawContent || '局内状态已更新'
    }
  }

  return {
    id: item.id || item.messageId || `message-${index}`,
    name,
    content,
    messageType
  }
}

function showEndConfirmation(actions = {}) {
  const directReview = actions.completionMode === 'direct_review'
  const content = directReview
    ? '结束后本局将直接进入评价，是否确认结束？'
    : '结束后将先由行家确认完成，再由玩家确认。全部确认后进入评价，是否继续？'

  return new Promise((resolve) => {
    wx.showModal({
      title: '结束本局',
      content,
      cancelText: '再想想',
      confirmText: '确认结束',
      success: (result) => resolve(Boolean(result.confirm)),
      fail: () => resolve(false)
    })
  })
}

function clampPercent(value) {
  if (!Number.isFinite(value)) {
    return 0
  }
  return Math.max(0, Math.min(100, Math.round(value)))
}
