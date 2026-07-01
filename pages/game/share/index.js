const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')

const SHARE_SCROLL_TAP_STEP_RPX = 360
const SHARE_SCROLL_HOLD_STEP_RPX = 72
const SHARE_SCROLL_HOLD_INTERVAL_MS = 80
const SHARE_SCROLL_HOLD_SUPPRESS_TAP_MS = 120

function formatDateTimeText(value) {
  if (!value) {
    return ''
  }

  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return String(value)
  }

  const month = date.getMonth() + 1
  const day = date.getDate()
  const hour = String(date.getHours()).padStart(2, '0')
  const minute = String(date.getMinutes()).padStart(2, '0')

  return `${date.getFullYear()}年${month}月${day}日 ${hour}:${minute}`
}

function normalizeGameInfo(data = {}) {
  const creator = data.creator || data.organizer || {}
  const organizerName = creator.name || creator.nickname || ''

  return {
    onlineText: data.onlineText || '3999人在线',
    id: data.id || data.gameId || '',
    bannerImage: data.bannerImage || data.coverSrc || data.coverUrl || data.coverFileUrl || '',
    title: data.title || '',
    startTime: data.timeText || data.startTimeText || formatDateTimeText(data.startAt || data.startTime),
    location: data.addressName || data.address || data.locationName || '',
    joinedCount: Number(data.approvedMemberCount || data.memberCount || data.joinedCount || 0),
    maxPlayers: Number(data.maxParticipants || data.maxPlayers || 0),
    introText: data.introText || data.introduction || data.description || '',
    organizerName,
    organizerTitle: creator.title || creator.role || creator.company || '',
    organizerAvatarText: creator.avatarText || organizerName.slice(0, 1),
    groupCreatedCount: Number(creator.groupCreatedCount || creator.createdGameCount || 0),
    recommendCount: Number(creator.recommendCount || creator.referralCount || 0)
  }
}

Page({
  data: {
    onlineText: '3999人在线',
    shareScrollTop: 0,
    isInterested: false,
    currentJoinedCount: 0,
    showShareWindow: false,
    loading: false,
    loadErrorText: '',
    gameInfo: normalizeGameInfo(),
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

    if (gameId) {
      this.loadGameInfo(gameId)
    }

    if (wx.showShareMenu) {
      wx.showShareMenu({
        withShareTicket: true,
        menus: ['shareAppMessage', 'shareTimeline']
      })
    }
  },

  async loadGameInfo(gameId) {
    this.setData({
      loading: true,
      loadErrorText: ''
    })

    try {
      const gameInfo = normalizeGameInfo(await gameService.getGameDetail(gameId))

      this.setData({
        loading: false,
        onlineText: gameInfo.onlineText,
        gameInfo,
        currentJoinedCount: gameInfo.joinedCount
      })
    } catch (error) {
      this.setData({
        loading: false,
        loadErrorText: error.message || '活动信息加载失败'
      })
      this.showToast(error.message || '活动信息加载失败')
    }
  },

  onToggleInterest() {
    const isInterested = !this.data.isInterested
    const joinedCount = this.data.gameInfo.joinedCount + (isInterested ? 1 : 0)

    this.setData({
      isInterested,
      currentJoinedCount: joinedCount
    })

    this.showToast(isInterested ? '已加入感兴趣列表' : '已取消感兴趣')
  },

  onCopyLocation() {
    const address = this.data.gameInfo.location

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
    this.showToast('私信分享待接入')
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

    if (key === 'up' || key === 'down') {
      if (!this.suppressNextNavTap) {
        this.scrollShare(key, SHARE_SCROLL_TAP_STEP_RPX)
      }
      return
    }

    if (key === 'left' || key === 'right') {
      this.showToast('功能正在开发中')
      return
    }

    this.handleShellAction(key)
  },

  handleShellNavLongPress(event) {
    const key = event.detail && event.detail.key

    if (key === 'map') {
      wx.showToast({
        title: '地图功能开发中',
        icon: 'none'
      })
      return
    }

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
      this.showToast('搜索功能开发中')
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
      map: ''
    }

    this.navigateToRoute(routeMap[key])
  },

  navigateToRoute(route) {
    if (!route || route === ROUTES.gameShare) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
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
    return {
      title: this.data.gameInfo.title,
      path: `/${ROUTES.gameShare}?id=${this.data.gameInfo.id}`,
      imageUrl: this.data.gameInfo.bannerImage
    }
  },

  onShareTimeline() {
    return {
      title: this.data.gameInfo.title,
      query: `id=${this.data.gameInfo.id}`,
      imageUrl: this.data.gameInfo.bannerImage
    }
  },

  onPreventTouch() {},

  onPreventBubble() {},

  onUnload() {
    this.clearShareScrollTimers()
  }
})
