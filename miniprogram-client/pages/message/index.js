const { ROUTES } = require('../../config/routes')
const messageService = require('../../services/message')
const { navigateShellKey, navigateShellRoute } = require('../../utils/shell-nav')
const { toUserMessage } = require('../../utils/user-message')

const NAV_ITEMS = [
  { name: '我的', key: 'mine' },
  { name: '元宇宙', key: 'metaverse' },
  { name: '地图', key: 'map' },
  { name: '消息', key: 'message' },
  { name: '首页', key: 'home' }
]

const MESSAGE_SECTIONS = []

const QUICK_BUCKETS = {
  join: 'group',
  system: 'system',
  achievement: 'achievement',
  warning: 'warning',
  friend: 'friend'
}

const QUICK_ACTION_ICONS = {
  join: '/pages/message/assets/i53@3x.png',
  system: '/pages/message/assets/i54@3x.png',
  achievement: '/pages/message/assets/i55@3x.png',
  warning: '/pages/message/assets/i56@3x.png',
  friend: '/pages/message/assets/i57@3x.png'
}

function quickActionBucket(item = {}) {
  if (item.bucket) {
    return item.bucket
  }

  const map = {
    join: 'group',
    system: 'system',
    achievement: 'achievement',
    warning: 'warning',
    friend: 'friend'
  }

  return map[item.key] || ''
}

function quickActionIconSrc(item = {}) {
  return QUICK_ACTION_ICONS[item.key] || item.iconSrc || ''
}

function hasOwn(object, key) {
  return Object.prototype.hasOwnProperty.call(object || {}, key)
}

function messageUnreadCount(items) {
  return (Array.isArray(items) ? items : []).filter((item) => item && item.unread).length
}

function sectionStats(sections) {
  const stats = {}

  ;(Array.isArray(sections) ? sections : []).forEach((section) => {
    const key = section && section.key ? section.key : ''
    const items = Array.isArray(section && section.items) ? section.items : []
    const unreadCount = hasOwn(section, 'unreadCount') ? Number(section.unreadCount || 0) : messageUnreadCount(items)
    const totalCount = hasOwn(section, 'totalCount') ? Number(section.totalCount || 0) : items.length

    if (key) {
      stats[key] = { unreadCount, totalCount }
    }
  })

  return stats
}

function normalizeQuickActions(actions, stats = {}) {
  const source = Array.isArray(actions) ? actions : []

  return source.map((item) => {
    const bucket = quickActionBucket(item)
    const stat = stats[bucket] || {}
    const unreadCount = hasOwn(item, 'unreadCount') ? Number(item.unreadCount || 0) : Number(stat.unreadCount || 0)
    const totalCount = hasOwn(item, 'totalCount') ? Number(item.totalCount || 0) : Number(stat.totalCount || 0)
    const countText = item.countText || `${unreadCount}/${totalCount}`

    return Object.assign({}, item, {
      bucket,
      unreadCount,
      totalCount,
      countText,
      displayLabel: `${item.label || ''}（${countText}）`,
      iconSrc: quickActionIconSrc(item),
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

function normalizeSections(sections, showAll = false) {
  const source = Array.isArray(sections) ? sections : MESSAGE_SECTIONS

  return source.map((section) => {
    const sourceItems = Array.isArray(section.items) ? section.items : []
    const unreadCount = hasOwn(section, 'unreadCount') ? Number(section.unreadCount || 0) : messageUnreadCount(sourceItems)
    const totalCount = hasOwn(section, 'totalCount') ? Number(section.totalCount || 0) : sourceItems.length
    const items = sourceItems.map((card) => normalizeMessageCard(card, section))

    return Object.assign({}, section, {
      moreText: '查看全部',
      moreRoute: normalizeRouteValue(section.moreRoute),
      unreadCount,
      totalCount,
      countText: section.countText || `${unreadCount}/${totalCount}`,
      items: showAll ? items : items.slice(0, 3)
    })
  })
}

function filterSectionsByBucket(sections, bucket) {
  const source = Array.isArray(sections) ? sections : []
  const targetBucket = String(bucket || '').trim()

  if (!targetBucket) {
    return source
  }

  return source.filter((section) => section && section.key === targetBucket)
}

function normalizeMessageCard(card = {}) {
  const route = normalizeRouteValue(card.route)
  const detailRoute = normalizeRouteValue(card.detailRoute)

  return Object.assign({}, card, {
    route,
    detailRoute,
    infoRows: Array.isArray(card.infoRows) ? card.infoRows : [],
    participants: Array.isArray(card.participants) ? card.participants : [],
    actions: Array.isArray(card.actions) ? card.actions.map((action) => Object.assign({}, action, {
      route: normalizeRouteValue(action.route),
      detailRoute: normalizeRouteValue(action.detailRoute)
    })) : []
  })
}

function normalizeRouteValue(route) {
  const text = String(route || '').trim()

  if (!text) {
    return ''
  }

  return text.startsWith('/') ? text : `/${text}`
}

function withMessageSource(route) {
  const normalized = normalizeRouteValue(route)

  if (!normalized || /[?&]from=message(?:&|$)/.test(normalized)) {
    return normalized
  }

  return `${normalized}${normalized.indexOf('?') >= 0 ? '&' : '?'}from=message`
}

function normalizeMessageCenter(source = {}, fallbackActiveTab = 'all') {
  const showAll = Boolean(source.showAll)
  const stats = sectionStats(source.sections)

  return {
    pageTitle: source.pageTitle || '',
    onlineText: source.onlineText || '',
    quickActions: normalizeQuickActions(source.quickActions, stats),
    tabs: normalizeTabs(source.tabs),
    activeTab: source.activeTab || fallbackActiveTab || 'all',
    activeBucket: source.activeBucket || '',
    showAll,
    sections: normalizeSections(source.sections, showAll),
    texts: source.texts || {}
  }
}

function textOf(data, key) {
  const texts = data && data.texts ? data.texts : {}
  return texts[key] || ''
}

function findMessageCard(sections = [], id = '') {
  for (const section of sections || []) {
    const items = Array.isArray(section.items) ? section.items : []
    const found = items.find((item) => String(item.id) === String(id))

    if (found) {
      return found
    }
  }

  return null
}

Page({
  data: {
    pageTitle: '',
    onlineText: '',
    navItems: NAV_ITEMS,
    quickActions: [],
    tabs: [],
    activeTab: 'all',
    activeBucket: '',
    activeType: '',
    showAll: false,
    sections: MESSAGE_SECTIONS,
    texts: {},
    loading: false,
    errorText: ''
  },

  onLoad(options = {}) {
    this._skipNextShowRefresh = true
    this.loadMessageCenter({
      tab: options.tab || 'all',
      type: options.type || '',
      bucket: options.bucket || '',
      showAll: options.detail === '1' || options.showAll === '1'
    })
  },

  onShow() {
    if (this._skipNextShowRefresh) {
      this._skipNextShowRefresh = false
      return
    }

    this.loadMessageCenter({
      tab: this.data.activeTab,
      type: this.data.activeType,
      bucket: this.data.activeBucket,
      showAll: this.data.showAll
    })
  },

  async loadMessageCenter(params = {}) {
    const tab = params.tab || this.data.activeTab || 'all'
    const type = params.type !== undefined ? params.type : this.data.activeType
    const bucket = params.bucket !== undefined ? params.bucket : this.data.activeBucket
    const showAll = params.showAll !== undefined ? Boolean(params.showAll) : Boolean(this.data.showAll)

    this.setData({
      loading: true,
      errorText: ''
    })

    try {
      const center = await messageService.getMessageCenter({
        tab,
        type,
        bucket: '',
        detail: showAll ? '1' : ''
      })
      const normalized = normalizeMessageCenter(center, tab)
      const allSections = normalized.sections

      this._messageSections = allSections
      normalized.activeTab = tab
      normalized.activeBucket = bucket
      normalized.showAll = showAll
      normalized.sections = filterSectionsByBucket(allSections, bucket)

      this.setData(Object.assign({}, normalized, {
        loading: false,
        errorText: '',
        activeType: type
      }))
    } catch (error) {
      const errorText = toUserMessage(error && error.message, textOf(this.data, 'loadFailedText') || '消息中心加载失败')

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
      tab: key,
      type: this.data.activeType,
      bucket: this.data.activeBucket,
      showAll: this.data.showAll
    })
  },

  onQuickTap(event) {
    const { key, bucket } = event.currentTarget.dataset
    const targetBucket = bucket || QUICK_BUCKETS[key] || ''

    if (!targetBucket) {
      this.showInfo(textOf(this.data, 'entryMissingText'))
      return
    }

    const nextBucket = this.data.activeBucket === targetBucket ? '' : targetBucket
    const cachedSections = Array.isArray(this._messageSections) ? this._messageSections : this.data.sections

    this.setData({
      activeBucket: nextBucket,
      activeTab: 'all',
      activeType: '',
      showAll: false,
      sections: filterSectionsByBucket(cachedSections, nextBucket)
    })
    this.loadMessageCenter({
      tab: 'all',
      type: '',
      bucket: nextBucket,
      showAll: false
    })
  },

  async onMessageTap(event) {
    const { id } = event.currentTarget.dataset
    const card = findMessageCard(this.data.sections, id)
    const backendRoute = card && (card.route || card.detailRoute)

    if (!backendRoute) {
      this.showInfo(textOf(this.data, 'routeMissingText') || '后台未返回跳转地址')
      return
    }

    await this.markNotificationReadIfNeeded(card)
    navigateShellRoute(withMessageSource(backendRoute), {
      currentRoute: ROUTES.message
    })
  },

  async markNotificationReadIfNeeded(card = {}) {
    if (!card || !card.unread) {
      return
    }

    this.markMessageReadLocally(card.id)

    try {
      // 局 IM 卡片是按房间聚合生成的展示卡；其 id 不是通知表主键。
      // 后端返回 notificationId 后必须优先用它，才能真正清除未读红点。
      await messageService.markNotificationRead(card.notificationId || card.id)
    } catch (error) {
      console.warn('[message] mark notification read failed', error)
    }
  },

  markMessageReadLocally(id) {
    const targetId = String(id || '')
    if (!targetId) {
      return
    }

    let changedBucket = ''
    let didChange = false
    const patchSections = (sections = []) => (Array.isArray(sections) ? sections : []).map((section) => {
      let changed = false
      const items = (Array.isArray(section.items) ? section.items : []).map((message) => {
        if (String(message.id) !== targetId || !message.unread) {
          return message
        }
        changed = true
        return Object.assign({}, message, {
          unread: false,
          unreadCount: 0
        })
      })

      if (!changed) {
        return section
      }

      didChange = true
      changedBucket = section.key || ''
      const unreadCount = Math.max(0, Number(section.unreadCount || 0) - 1)
      const totalCount = Number(section.totalCount || items.length || 0)
      return Object.assign({}, section, {
        unreadCount,
        totalCount,
        countText: `${unreadCount}/${totalCount}`,
        items
      })
    })

    const sections = patchSections(this.data.sections)
    if (Array.isArray(this._messageSections)) {
      this._messageSections = patchSections(this._messageSections)
    }

    if (!didChange) {
      return
    }

    const quickActions = (this.data.quickActions || []).map((item) => {
      if (!changedBucket || item.bucket !== changedBucket) {
        return item
      }
      const unreadCount = Math.max(0, Number(item.unreadCount || 0) - 1)
      const totalCount = Number(item.totalCount || 0)
      const countText = `${unreadCount}/${totalCount}`
      return Object.assign({}, item, {
        unreadCount,
        countText,
        displayLabel: `${item.label || ''}（${countText}）`,
        unread: unreadCount > 0
      })
    })

    const tabs = (this.data.tabs || []).map((item) => {
      const key = item.key || ''
      if (key !== 'all' && key !== 'unread' && !(changedBucket === 'warning' && key === 'trade')) {
        return item
      }
      const unreadCount = Math.max(0, Number(item.unreadCount || 0) - 1)
      return Object.assign({}, item, {
        unreadCount,
        unread: unreadCount > 0
      })
    })

    this.setData({
      sections,
      quickActions,
      tabs
    })
  },

  onSectionMoreTap(event) {
    const { key, route } = event.currentTarget.dataset

    if (route) {
      navigateShellRoute(withMessageSource(route), {
        currentRoute: ROUTES.message,
        reuseExisting: false
      })
      return
    }

    if (key) {
      this.loadMessageCenter({
        tab: this.data.activeTab,
        type: this.data.activeType,
        bucket: key,
        showAll: true
      })
      return
    }

    this.showInfo(textOf(this.data, 'routeMissingText') || '后台未返回跳转地址')
  },

  async onActionTap(event) {
    const { action, id } = event.currentTarget.dataset

    if (!id || !action) {
      this.showInfo(textOf(this.data, 'actionMissingText'))
      return
    }

    const navigationAction = action === 'detail' || action === 'view' || action === 'open'

    try {
      const result = await messageService.handleNotificationAction({
        notificationId: id,
        action
      })

      const card = findMessageCard(this.data.sections, id)
      await this.markNotificationReadIfNeeded(card)

      if (this.navigateByActionResult(result)) {
        return
      }

      if (navigationAction) {
        this.showInfo('后台返回失败')
        return
      }

      this.showInfo(result && result.handled ? textOf(this.data, 'actionSuccessText') : textOf(this.data, 'actionHandledText'))
      this.loadMessageCenter({ tab: this.data.activeTab, type: this.data.activeType })
    } catch (error) {
      this.showInfo(navigationAction ? '后台返回失败' : (error.message || textOf(this.data, 'actionFailedText')))
    }
  },

  navigateByActionResult(result = {}) {
    const target = result.target || {}
    const route = target.route || target.url || ''

    if (!route) {
      return false
    }

    return navigateShellRoute(withMessageSource(route), {
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
