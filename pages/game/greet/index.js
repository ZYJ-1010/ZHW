const { ROUTES } = require('../../../config/routes')

const GREET_SCROLL_TAP_STEP_RPX = 360
const GREET_SCROLL_HOLD_STEP_RPX = 72
const GREET_SCROLL_HOLD_INTERVAL_MS = 80
const GREET_SCROLL_HOLD_SUPPRESS_TAP_MS = 120

Page({
  data: {
    onlineText: '3999人在线',
    greetScrollTop: 0,
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    sender: {
      avatarText: 'ME'
    },
    referrer: {
      name: '我',
      message: '李娜是我认识的产品总监，正在找资深产品经理做咨询，我觉得你们很匹配，要不要聊聊？'
    },
    player: {
      name: '玩家的姓名',
      avatarText: '李',
      desc: '寻找产品经理合作'
    },
    gameInfo: {
      dateText: '3月5日',
      location: '中关村创'
    }
  },

  onLoad(options = {}) {
    const playerName = this.decodeQueryText(options.playerName || options.name)

    if (playerName) {
      this.setData({
        'player.name': playerName,
        'player.avatarText': playerName.slice(0, 1)
      })
    }
  },

  decodeQueryText(value = '') {
    try {
      return decodeURIComponent(value)
    } catch (error) {
      return value
    }
  },

  onAcceptTap() {
    this.showInfo('已确认参加')
  },

  onDeclineTap() {
    this.showInfo('已婉拒')
  },

  onInputToolTap() {
    this.showInfo('功能正在开发中')
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (key === 'up' || key === 'down') {
      if (!this.suppressNextNavTap) {
        this.scrollGreet(key, GREET_SCROLL_TAP_STEP_RPX)
      }
      return
    }

    if (key === 'left' || key === 'right') {
      this.showInfo('功能正在开发中')
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
    this.stopGreetScrollHold(false)
    this.scrollGreet(key, GREET_SCROLL_HOLD_STEP_RPX)

    this.greetScrollHoldTimer = setInterval(() => {
      this.scrollGreet(key, GREET_SCROLL_HOLD_STEP_RPX)
    }, GREET_SCROLL_HOLD_INTERVAL_MS)
  },

  handleShellNavTouchEnd() {
    this.stopGreetScrollHold(true)
  },

  handleShellAction(key) {
    if (key === 'home') {
      this.scrollGreetToTop()
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
      metaverse: ROUTES.metaverse,
      map: ROUTES.map
    }

    this.navigateToRoute(routeMap[key])
  },

  navigateToRoute(route) {
    if (!route || route === ROUTES.gameGreet) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  },

  handleGreetScroll(event) {
    const scrollTop = event.detail && event.detail.scrollTop

    if (typeof scrollTop === 'number') {
      this.greetScrollTopValue = scrollTop
    }
  },

  scrollGreet(direction, stepRpx = GREET_SCROLL_TAP_STEP_RPX) {
    const current = Number(this.greetScrollTopValue || this.data.greetScrollTop || 0)
    const distance = this.rpxToPx(stepRpx)
    const nextTop = direction === 'up'
      ? Math.max(0, current - distance)
      : current + distance

    this.greetScrollTopValue = nextTop
    this.setData({
      greetScrollTop: nextTop
    })
  },

  scrollGreetToTop() {
    this.greetScrollTopValue = 0
    this.setData({
      greetScrollTop: 0
    })
  },

  stopGreetScrollHold(resetTapSuppress) {
    if (this.greetScrollHoldTimer) {
      clearInterval(this.greetScrollHoldTimer)
      this.greetScrollHoldTimer = null
    }

    if (resetTapSuppress && this.suppressNextNavTap) {
      if (this.greetScrollSuppressTimer) {
        clearTimeout(this.greetScrollSuppressTimer)
      }

      this.greetScrollSuppressTimer = setTimeout(() => {
        this.suppressNextNavTap = false
        this.greetScrollSuppressTimer = null
      }, GREET_SCROLL_HOLD_SUPPRESS_TAP_MS)
    }
  },

  clearGreetScrollTimers() {
    this.stopGreetScrollHold(false)

    if (this.greetScrollSuppressTimer) {
      clearTimeout(this.greetScrollSuppressTimer)
      this.greetScrollSuppressTimer = null
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

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  },

  onUnload() {
    this.clearGreetScrollTimers()
  }
})
