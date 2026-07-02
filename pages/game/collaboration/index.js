const gameService = require('../../../services/game')
const { ROUTES } = require('../../../config/routes')
const { getSurnameInitials } = require('../../../utils/avatar')
const { navigateShellRoute } = require('../../../utils/shell-nav')

const WHITE_CONTENT_LEFT_RPX = 2
const WHITE_CONTENT_TOP_RPX = 160
const WHITE_CONTENT_WIDTH_RPX = 750
const WHITE_DESIGN_FRAME_HEIGHT_PT = 810
const WHITE_DESIGN_BOTTOM_HEIGHT_PT = 78
const WHITE_NAV_TITLE_HEIGHT_RPX = 50
const WHITE_BACK_BUTTON_SIZE_RPX = 40
const WHITE_DEFAULT_CAPSULE_BOTTOM_RPX = 142
const WHITE_DEFAULT_FRAME_HEIGHT_RPX = WHITE_DESIGN_FRAME_HEIGHT_PT * 2

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

function getLayoutStyles() {
  const capsuleBottom = getMenuCapsuleBottomRpx()
  const titleTop = Math.max(0, roundRpx(capsuleBottom - WHITE_NAV_TITLE_HEIGHT_RPX))
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
    titleStyle: `top: ${titleTop}rpx; height: ${WHITE_NAV_TITLE_HEIGHT_RPX}rpx; line-height: ${WHITE_NAV_TITLE_HEIGHT_RPX}rpx;`,
    backStyle: `top: ${backTop}rpx; width: ${WHITE_BACK_BUTTON_SIZE_RPX}rpx; height: ${WHITE_BACK_BUTTON_SIZE_RPX}rpx;`
  }
}

Page({
  data: {
    layout: getLayoutStyles(),
    gameId: '',
    loading: false,
    loadError: '',
    title: '局内协作',
    subtitle: '',
    progressPercent: 0,
    members: [],
    membersText: '',
    tasks: [],
    messages: [],
    actions: {
      canManageMembers: false,
      canEndGame: false,
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
        loadError: error && error.message ? error.message : '协作数据加载失败',
        subtitle: '',
        progressPercent: 0,
        members: [],
        membersText: '',
        tasks: [],
        messages: [],
        actions: {
          canManageMembers: false,
          canEndGame: false,
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

  handleEndSession() {
    const actions = this.data.actions || {}
    const route = actions.endConfirmRoute || `${ROUTES.gameDelivery}?gameId=${this.data.gameId || ''}`

    if (!actions.canEndGame) {
      wx.showToast({
        title: '当前状态不可结束',
        icon: 'none'
      })
      return
    }

    navigateShellRoute(route.startsWith('/') ? route : `/${route}`)
  }
})

function normalizeCollaboration(data = {}, fallbackGameId = '') {
  const source = data && typeof data === 'object' ? data : {}
  const progress = source.progress || {}
  const tasks = Array.isArray(progress.tasks) && progress.tasks.length
    ? progress.tasks.map((item, index) => ({
      title: item.title || `阶段 ${index + 1}`,
      desc: item.desc || item.description || '',
      state: item.state || ''
    }))
    : []
  const members = Array.isArray(source.members)
    ? source.members.map((item, index) => {
      const name = item.name || item.nickname || `成员${index + 1}`
      return {
        id: item.id || item.userId || `member-${index}`,
        name,
        roleText: item.roleText || item.roleLabel || item.role || '',
        avatarText: item.avatarText || getSurnameInitials(name, 'ME')
      }
    })
    : []
  const messages = Array.isArray(source.messages)
    ? source.messages.map((item, index) => ({
      id: item.id || item.messageId || `message-${index}`,
      name: item.senderName || item.name || item.nickname || '成员',
      content: item.content || ''
    })).filter((item) => item.content)
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
    actions: Object.assign({
      canManageMembers: true,
      canEndGame: false,
      manageRoute: `${ROUTES.gameParticipants}?gameId=${source.gameId || fallbackGameId}`,
      endConfirmRoute: `${ROUTES.gameDelivery}?gameId=${source.gameId || fallbackGameId}`
    }, source.actions || {})
  }
}

function clampPercent(value) {
  if (!Number.isFinite(value)) {
    return 0
  }
  return Math.max(0, Math.min(100, Math.round(value)))
}
