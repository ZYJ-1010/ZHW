const { ROUTES } = require('../../../config/routes')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')
const gameService = require('../../../services/game')

const SHARE_SCROLL_TAP_STEP_RPX = 360
const SHARE_SCROLL_HOLD_STEP_RPX = 72
const SHARE_SCROLL_HOLD_INTERVAL_MS = 80
const SHARE_SCROLL_HOLD_SUPPRESS_TAP_MS = 120

function emptyGameInfo() {
  return {
    id: '',
    bannerImage: '',
    title: '',
    startTime: '',
    location: '',
    joinedCount: 0,
    maxPlayers: 8,
    introText: '',
    organizerName: '',
    organizerTitle: '',
    organizerAvatarText: '',
    groupCreatedCount: 0,
    recommendCount: 0,
    hasData: false
  }
}

function gameStatusText(status) {
  const map = {
    draft: '草稿',
    pending: '待审核',
    approved: '已通过',
    recruiting: '招募中',
    full: '已满员',
    in_progress: '进行中',
    completed: '已完成',
    canceled: '已取消',
    rejected: '已驳回'
  }

  return map[status] || status || ''
}

function normalizeShareGameInfo(data = {}, fallback = emptyGameInfo()) {
  const game = data.game || data
  const detailDisplay = data.detailDisplay || game.detailDisplay || {}
  const creatorId = game.creatorUserId || ''
  const title = String(game.title || '').trim()
  const location = String(detailDisplay.locationText || game.locationText || game.address || game.cityName || '').trim()
  const statusText = String(detailDisplay.statusText || game.statusText || gameStatusText(game.status)).trim()
  const maxPlayers = Number(game.maxPlayers || fallback.maxPlayers || 8)
  const joinedCount = Number(game.currentPlayers || (Array.isArray(data.memberIds) ? data.memberIds.length : 0))

  return {
    id: game.id || fallback.id,
    bannerImage: game.coverUrl || game.bannerImage || fallback.bannerImage,
    title,
    startTime: statusText ? `状态：${statusText}` : fallback.startTime,
    location,
    joinedCount,
    maxPlayers,
    introText: game.introText || game.description || (title ? `${title}。${location ? `地点：${location}。` : ''}` : ''),
    organizerName: game.organizerName || game.creatorName || game.creatorNickname || (creatorId ? `发起人 ${creatorId}` : ''),
    organizerTitle: game.organizerTitle || game.creatorTitle || (game.mainGuideUserId ? `主行家 ${game.mainGuideUserId}` : ''),
    organizerAvatarText: game.organizerAvatarText || game.creatorAvatarText || (creatorId ? String(creatorId).slice(-2) : ''),
    groupCreatedCount: Number(game.groupCreatedCount || game.createdGameCount || 0),
    recommendCount: Number(game.recommendCount || game.referralCount || 0),
    hasData: Boolean(game.id || title)
  }
}

const DEFAULT_GAME_INFO = emptyGameInfo()

Page({
  data: {
    gameId: '',
    onlineText: '在线',
    shareScrollTop: 0,
    isInterested: false,
    currentJoinedCount: DEFAULT_GAME_INFO.joinedCount,
    showShareWindow: false,
    gameInfo: DEFAULT_GAME_INFO,
    shareEntry: null,
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ]
  },

  onLoad(options = {}) {
    const gameId = options.gameId || options.id || ''

    this.setData({
      gameId,
      gameInfo: Object.assign({}, DEFAULT_GAME_INFO, { id: gameId }),
      currentJoinedCount: 0
    })
    this.loadShareGame(gameId)

    if (wx.showShareMenu) {
      wx.showShareMenu({
        withShareTicket: true,
        menus: ['shareAppMessage', 'shareTimeline']
      })
    }
  },

  async loadShareGame(gameId) {
    if (!gameId) {
      return
    }

    try {
      const detail = await gameService.getGameDetail(gameId)
      const gameInfo = normalizeShareGameInfo(detail, this.data.gameInfo)

      this.setData({
        gameInfo,
        currentJoinedCount: gameInfo.joinedCount
      })
      this.ensureShareEntry(gameInfo)
    } catch (error) {
      this.showToast(error.message || '组局信息加载失败')
    }
  },

  async ensureShareEntry(gameInfo = this.data.gameInfo) {
    const gameId = gameInfo.id || this.data.gameId

    if (!gameId || this.data.shareEntry) {
      return this.data.shareEntry
    }

    try {
      const shareEntry = await gameService.createInviteEntry({
        entryType: 'link',
        gameId: Number(gameId),
        title: gameInfo.title || '真好玩组局邀请'
      })

      this.setData({ shareEntry })
      return shareEntry
    } catch (error) {
      this.showToast(error.message || '邀请链接生成失败')
      return null
    }
  },

  onToggleInterest() {
    const isInterested = !this.data.isInterested

    if (!this.data.gameInfo.hasData) {
      this.showToast('组局信息加载后才可以操作')
      return
    }

    this.setData({
      isInterested
    })

    this.showToast(isInterested ? '已标记感兴趣' : '已取消感兴趣')
  },

  onCopyLocation() {
    const address = this.data.gameInfo.location

    if (!address) {
      this.showToast('暂无可复制地址')
      return
    }

    if (!wx.setClipboardData) {
      this.showToast(address)
      return
    }

    wx.setClipboardData({
      data: address,
      success: () => {
        this.showToast('地址已复制')
      }
    })
  },

  onOpenShare() {
    if (!this.data.gameInfo.hasData) {
      this.showToast('组局信息加载后才可以分享')
      return
    }

    this.setData({
      showShareWindow: true
    })
  },

  onCloseShare() {
    this.setData({
      showShareWindow: false
    })
  },

  onNativeShareTap() {
    this.onCloseShare()
  },

  onShareDirect() {
    this.onCloseShare()
    navigateShellRoute(`/${ROUTES.message}?from=gameShare${this.data.gameId ? `&gameId=${encodeURIComponent(this.data.gameId)}` : ''}`)
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (key === 'up' || key === 'down') {
      if (!this.suppressNextNavTap) {
        this.scrollShare(key, SHARE_SCROLL_TAP_STEP_RPX)
      }
      return
    }

    if (navigateShellKey(key, {
      currentRoute: ROUTES.gameShare
    })) {
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
    this.stopShareScrollHold(false)
    this.scrollShare(key, SHARE_SCROLL_HOLD_STEP_RPX)

    this.shareScrollHoldTimer = setInterval(() => {
      this.scrollShare(key, SHARE_SCROLL_HOLD_STEP_RPX)
    }, SHARE_SCROLL_HOLD_INTERVAL_MS)
  },

  handleShellNavTouchEnd() {
    this.stopShareScrollHold(true)
  },

  handleShellAction(key) {
    if (key === 'home') {
      this.scrollShareToTop()
      return
    }

    if (key === 'search') {
      this.navigateToRoute(ROUTES.gameHall)
      return
    }

    if (key === 'comment' || key === 'message') {
      this.navigateToRoute(ROUTES.message)
      return
    }

    if (key === 'avatar' || key === 'mine') {
      this.navigateToRoute(ROUTES.profile)
      return
    }

    const routeMap = {
      metaverse: ROUTES.metaverse,
      map: ROUTES.map
    }

    this.navigateToRoute(routeMap[key])
  },

  navigateToRoute(route) {
    if (!route || route === ROUTES.gameShare) {
      return
    }

    navigateShellRoute(route)
  },

  handleShareScroll(event) {
    const scrollTop = event.detail && event.detail.scrollTop

    if (typeof scrollTop === 'number') {
      this.shareScrollTopValue = scrollTop
    }
  },

  scrollShare(direction, stepRpx = SHARE_SCROLL_TAP_STEP_RPX) {
    const current = Number(this.shareScrollTopValue || this.data.shareScrollTop || 0)
    const distance = this.rpxToPx(stepRpx)
    const nextTop = direction === 'up'
      ? Math.max(0, current - distance)
      : current + distance

    this.shareScrollTopValue = nextTop
    this.setData({
      shareScrollTop: nextTop
    })
  },

  scrollShareToTop() {
    this.shareScrollTopValue = 0
    this.setData({
      shareScrollTop: 0
    })
  },

  stopShareScrollHold(resetTapSuppress) {
    if (this.shareScrollHoldTimer) {
      clearInterval(this.shareScrollHoldTimer)
      this.shareScrollHoldTimer = null
    }

    if (resetTapSuppress && this.suppressNextNavTap) {
      if (this.shareScrollSuppressTimer) {
        clearTimeout(this.shareScrollSuppressTimer)
      }

      this.shareScrollSuppressTimer = setTimeout(() => {
        this.suppressNextNavTap = false
        this.shareScrollSuppressTimer = null
      }, SHARE_SCROLL_HOLD_SUPPRESS_TAP_MS)
    }
  },

  clearShareScrollTimers() {
    this.stopShareScrollHold(false)

    if (this.shareScrollSuppressTimer) {
      clearTimeout(this.shareScrollSuppressTimer)
      this.shareScrollSuppressTimer = null
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

  showToast(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  },

  onShareAppMessage() {
    if (!this.data.gameInfo.hasData) {
      return {
        title: '真好玩',
        path: `/${ROUTES.gameHall}`
      }
    }

    const entry = this.data.shareEntry || {}
    const inviteCode = entry.inviteCode || ''
    const entryType = entry.entryType || 'link'
    const gameId = this.data.gameInfo.id || this.data.gameId || ''
    const fallbackPath = `/${ROUTES.gameShare}?id=${encodeURIComponent(gameId)}`
    const path = entry.path
      ? entry.path.replace(/^\/+/, '/')
      : `${fallbackPath}${inviteCode ? `&inviteCode=${encodeURIComponent(inviteCode)}&entryType=${encodeURIComponent(entryType)}` : ''}`

    return {
      title: entry.title || this.data.gameInfo.title || '真好玩组局邀请',
      path,
      imageUrl: this.data.gameInfo.bannerImage || undefined
    }
  },

  onShareTimeline() {
    if (!this.data.gameInfo.hasData) {
      return {
        title: '真好玩',
        query: ''
      }
    }

    const entry = this.data.shareEntry || {}
    const inviteCode = entry.inviteCode || ''
    const entryType = entry.entryType || 'link'
    const gameId = this.data.gameInfo.id || this.data.gameId || ''
    const query = [
      gameId ? `id=${encodeURIComponent(gameId)}` : '',
      inviteCode ? `inviteCode=${encodeURIComponent(inviteCode)}` : '',
      entryType ? `entryType=${encodeURIComponent(entryType)}` : ''
    ].filter(Boolean).join('&')

    return {
      title: entry.title || this.data.gameInfo.title || '真好玩组局邀请',
      query,
      imageUrl: this.data.gameInfo.bannerImage || undefined
    }
  },

  onPreventTouch() {},

  onPreventBubble() {},

  onUnload() {
    this.clearShareScrollTimers()
  }
})
