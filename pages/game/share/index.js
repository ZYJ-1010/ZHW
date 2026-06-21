const { ROUTES } = require('../../../config/routes')

const SHARE_SCROLL_TAP_STEP_RPX = 360
const SHARE_SCROLL_HOLD_STEP_RPX = 72
const SHARE_SCROLL_HOLD_INTERVAL_MS = 80
const SHARE_SCROLL_HOLD_SUPPRESS_TAP_MS = 120

const DEFAULT_GAME_INFO = {
  id: 'current-game-share',
  bannerImage: '/pages/game/hall/assets/hall-featured-city.jpg',
  title: '企业数字化转型及技术服务沙龙会',
  startTime: '2月12日 08:30-11:30',
  location: '上海市浦东新区沙新镇黄赵路310号',
  joinedCount: 5,
  maxPlayers: 8,
  introText: '本场沙龙针对企业数字化转型及技术服务话题展开。我们将邀请多位成功创业者与资深技术专家倾情分享落地成效及痛点突破经验。活动包含三大主题演讲分享、自由沙龙讨论，以及精准资本与技术资源对接，帮助每位参与者深度拓展人脉、获取前沿产业资源、寻找优质合作伙伴。',
  organizerName: '陆毅',
  organizerTitle: '总经理 | 上海创世界科技有限公司',
  organizerAvatarText: '陆',
  groupCreatedCount: 88,
  recommendCount: 20
}

Page({
  data: {
    onlineText: '3999人在线',
    shareScrollTop: 0,
    isInterested: false,
    currentJoinedCount: DEFAULT_GAME_INFO.joinedCount,
    showShareWindow: false,
    gameInfo: DEFAULT_GAME_INFO,
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ]
  },

  onLoad() {
    if (wx.showShareMenu) {
      wx.showShareMenu({
        withShareTicket: true,
        menus: ['shareAppMessage', 'shareTimeline']
      })
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
      map: ROUTES.map
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
