const { ROUTES } = require('../../config/routes')
const messageService = require('../../services/message')
const { navigateShellKey, navigateShellRoute } = require('../../utils/shell-nav')

const NAV_ITEMS = [
  { name: '我的', key: 'mine' },
  { name: '元宇宙', key: 'metaverse' },
  { name: '地图', key: 'map' },
  { name: '消息', key: 'message' },
  { name: '首页', key: 'home' }
]

const MESSAGE_SECTIONS = []

function normalizeQuickActions(actions) {
  const source = Array.isArray(actions) ? actions : []

  return source.map((item) => {
    const unreadCount = Number(item.unreadCount || 0)

    return Object.assign({}, item, {
      unreadCount,
      unread: Boolean(item.unread) || unreadCount > 0
    })
  })
}

function normalizeTabs(tabs) {
  const source = Array.isArray(tabs) ? tabs : []

  return source.map((item) => {
    const unreadCount = Number(item.unreadCount || 0)

    return Object.assign({}, item, {
      unreadCount,
      unread: Boolean(item.unread) || unreadCount > 0
    })
  })
}

function normalizeSections(sections) {
  const source = Array.isArray(sections) ? sections : MESSAGE_SECTIONS

  return source.map((section) => Object.assign({}, section, {
    items: Array.isArray(section.items) ? section.items : []
  }))
}

function normalizeMessageCenter(source = {}, fallbackActiveTab = 'all') {
  return {
    pageTitle: source.pageTitle || '',
    onlineText: source.onlineText || '',
    quickActions: normalizeQuickActions(source.quickActions),
    tabs: normalizeTabs(source.tabs),
    activeTab: source.activeTab || fallbackActiveTab || 'all',
    sections: normalizeSections(source.sections),
    texts: source.texts || {}
  }
}

function textOf(data, key) {
  const texts = data && data.texts ? data.texts : {}
  return texts[key] || ''
}

Page({
  data: {
    pageTitle: '',
    onlineText: '',
    navItems: NAV_ITEMS,
    quickActions: [],
    tabs: [],
    activeTab: 'all',
    sections: MESSAGE_SECTIONS,
    texts: {},
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
      const errorText = error && error.message ? error.message : textOf(this.data, 'loadFailedText')

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
      join: ROUTES.message,
      system: ROUTES.messageSystemDetail || 'pages/message/system-detail/index',
      achievement: ROUTES.profileAchievements || 'pages/profile/achievements/index',
      warning: ROUTES.messageTradeWarning || 'pages/message/trade-warning/index',
      friend: ROUTES.messageMy || 'pages/message/my/index'
    }
    const route = routeMap[key]

    if (key === 'join') {
      this.loadMessageCenter({ tab: 'all' })
      return
    }

    if (route) {
      navigateShellRoute(route, {
        currentRoute: ROUTES.message
      })
      return
    }

    this.showInfo(textOf(this.data, 'entryMissingText'))
  },

  async onMessageTap(event) {
    const { id, routeKey } = event.currentTarget.dataset

    if (id) {
      try {
        const result = await messageService.handleNotificationAction({
          notificationId: id,
          action: 'detail'
        })

        if (this.navigateByActionResult(result)) {
          return
        }
      } catch (error) {
        this.showInfo(error.message || textOf(this.data, 'openFailedText'))
        return
      }
    }

    const routeMap = {
      system: ROUTES.messageSystemDetail || 'pages/message/system-detail/index',
      warning: ROUTES.messageTradeWarning || 'pages/message/trade-warning/index',
      friend: ROUTES.messageMy || 'pages/message/my/index'
    }
    const route = routeMap[routeKey]

    if (!route) {
      return
    }

    navigateShellRoute(route, {
      currentRoute: ROUTES.message
    })
  },

  onSectionMoreTap(event) {
    const { key } = event.currentTarget.dataset
    const tabMap = {
      group: 'all',
      system: 'all',
      warning: 'trade'
    }

    this.loadMessageCenter({
      tab: tabMap[key] || key || this.data.activeTab
    })
  },

  async onActionTap(event) {
    const { action, id } = event.currentTarget.dataset

    if (!id || !action) {
      this.showInfo(textOf(this.data, 'actionMissingText'))
      return
    }

    try {
      const result = await messageService.handleNotificationAction({
        notificationId: id,
        action
      })

      if (this.navigateByActionResult(result)) {
        return
      }

      this.showInfo(result && result.handled ? textOf(this.data, 'actionSuccessText') : textOf(this.data, 'actionHandledText'))
      this.loadMessageCenter({ tab: this.data.activeTab })
    } catch (error) {
      this.showInfo(error.message || textOf(this.data, 'actionFailedText'))
    }
  },

  navigateByActionResult(result = {}) {
    const target = result.target || {}
    const route = target.route || target.url || ''

    if (!route) {
      return false
    }

    return navigateShellRoute(route, {
      currentRoute: ROUTES.message
    })
  },
  handleShellNavTap(event) {
    const { key } = event.detail || {}

    navigateShellKey(key, {
      currentRoute: ROUTES.message
    })
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
