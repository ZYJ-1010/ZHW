const { ROUTES } = require('../../config/routes')
const messageService = require('../../services/message')

const NAV_ITEMS = [
  { name: '我的', key: 'mine' },
  { name: '元宇宙', key: 'metaverse' },
  { name: '地图', key: 'map' },
  { name: '消息', key: 'message' },
  { name: '首页', key: 'home' }
]

const QUICK_ACTIONS = [
]

const TABS = [
]

const MESSAGE_SECTIONS = [
]

function normalizeQuickActions(actions) {
  const source = Array.isArray(actions) && actions.length ? actions : QUICK_ACTIONS

  return source.map((item) => {
    const unreadCount = Number(item.unreadCount || 0)

    return Object.assign({}, item, {
      unreadCount,
      unread: Boolean(item.unread) || unreadCount > 0
    })
  })
}

function normalizeTabs(tabs) {
  const source = Array.isArray(tabs) && tabs.length ? tabs : TABS

  return source.map((item) => {
    const unreadCount = Number(item.unreadCount || 0)

    return Object.assign({}, item, {
      unreadCount,
      unread: Boolean(item.unread) || unreadCount > 0
    })
  })
}

function normalizeSections(sections) {
  const source = Array.isArray(sections) && sections.length ? sections : MESSAGE_SECTIONS

  return source.map((section) => Object.assign({}, section, {
    items: Array.isArray(section.items) ? section.items : []
  }))
}

function normalizeMessageCenter(source = {}, fallbackActiveTab = 'all') {
  const sections = normalizeSections(source.sections)

  return {
    pageTitle: source.pageTitle || '消息中心',
    onlineText: source.onlineText || '',
    quickActions: normalizeQuickActions(source.quickActions),
    tabs: normalizeTabs(source.tabs),
    activeTab: source.activeTab || fallbackActiveTab || 'all',
    sections,
    hasMessages: sections.some((section) => section.items.length)
  }
}

Page({
  data: {
    pageTitle: '消息中心',
    onlineText: '',
    navItems: NAV_ITEMS,
    quickActions: normalizeQuickActions(QUICK_ACTIONS),
    tabs: normalizeTabs(TABS),
    activeTab: 'all',
    sections: MESSAGE_SECTIONS,
    hasMessages: false,
    loading: false,
    errorText: ''
  },

  onLoad(options = {}) {
    this.loadMessageCenter({
      tab: options.tab || 'all'
    })
  },

  async loadMessageCenter(params = {}) {
    const tab = params.tab || this.data.activeTab || 'all'

    this.setData({
      loading: true,
      errorText: ''
    })

    try {
      const center = await messageService.getMessageCenter({ tab })
      const normalized = normalizeMessageCenter(center, tab)

      this.setData(Object.assign({}, normalized, {
        loading: false,
        errorText: ''
      }))
    } catch (error) {
      const errorText = error && error.message ? error.message : '消息中心加载失败'

      this.setData({
        loading: false,
        errorText
      })
      this.showInfo(errorText)
    }
  },

  onTabTap(event) {
    const { key } = event.currentTarget.dataset

    if (!key || key === this.data.activeTab) {
      return
    }

    this.setData({
      activeTab: key
    })
    this.loadMessageCenter({
      tab: key
    })
  },

  onQuickTap(event) {
    const { key } = event.currentTarget.dataset
    const routeMap = {
      system: ROUTES.messageSystemDetail || 'pages/message/system-detail/index',
      warning: ROUTES.messageTradeWarning || 'pages/message/trade-warning/index',
      friend: ROUTES.messageMy || 'pages/message/my/index'
    }
    const route = routeMap[key]

    if (route) {
      wx.navigateTo({
        url: `/${route}`
      })
      return
    }

    this.showInfo('功能正在开发中')
  },

  onMessageTap(event) {
    const {
      routeKey,
      roomId,
      conversationId,
      name,
      avatarText
    } = event.currentTarget.dataset
    const routeMap = {
      system: ROUTES.messageSystemDetail || 'pages/message/system-detail/index',
      warning: ROUTES.messageTradeWarning || 'pages/message/trade-warning/index',
      friend: ROUTES.messageMy || 'pages/message/my/index'
    }
    const route = routeMap[routeKey]

    if (!route) {
      return
    }

    const query = routeKey === 'friend'
      ? this.buildQuery({
        roomId,
        conversationId,
        name,
        avatarText
      })
      : ''

    wx.navigateTo({
      url: `/${route}${query}`
    })
  },

  buildQuery(params = {}) {
    const pairs = Object.keys(params)
      .filter((key) => params[key] !== undefined && params[key] !== null && params[key] !== '')
      .map((key) => `${encodeURIComponent(key)}=${encodeURIComponent(params[key])}`)

    return pairs.length ? `?${pairs.join('&')}` : ''
  },

  onActionTap(event) {
    const { action } = event.currentTarget.dataset

    this.showInfo(`${action || '操作'}待接入`)
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}

    if (key === 'map') {
      wx.showToast({
        title: '地图功能开发中',
        icon: 'none'
      })
      return
    }
    const routeMap = {
      home: ROUTES.playerHome || ROUTES.home,
      map: '',
      message: ROUTES.message,
      mine: ROUTES.profile,
      avatar: ROUTES.profile,
      metaverse: ROUTES.metaverse
    }
    const route = routeMap[key]

    if (!route || route === ROUTES.message) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
