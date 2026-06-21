const toast = require('../../../utils/toast')
const { ROUTES } = require('../../../config/routes')

const HALL_SCROLL_TAP_STEP_RPX = 360
const HALL_SCROLL_HOLD_STEP_RPX = 72
const HALL_SCROLL_HOLD_INTERVAL_MS = 80
const HALL_SCROLL_HOLD_SUPPRESS_TAP_MS = 120
const DEFAULT_EVENT_ACTIONS = ['分享', '关注', '引荐', '打招呼']
const TYPE_FILTERS = [
  { key: 'all', label: '类型' },
  { key: 'deposit', label: '押金局' },
  { key: 'task', label: '任务局' },
  { key: 'social', label: '社交局' }
]

const eventsList = [
  {
    id: 'deposit-morning',
    type: 'deposit',
    typeText: '押金局',
    categoryKey: 'growth',
    startAt: '2026-05-01T07:00:00+08:00',
    distanceKm: 8.2,
    coverSrc: '/components/game-card/assets/game-cover-default.png',
    tag: '押金局',
    price: '￥100/人',
    title: '早起星人挑战：连续7天打卡',
    location: '📍黄浦区 · 8.2km · 3/8人',
    time: '⏰2026年5月1日-7日 7:00',
    action: '加入',
    joinedText: '+14位玩家已入局',
    avatarFallbacks: ['早', '局', '玩'],
    actions: DEFAULT_EVENT_ACTIONS
  },
  {
    id: 'ai-build',
    type: 'task',
    typeText: '任务局',
    categoryKey: 'task',
    startAt: '2026-05-01T14:00:00+08:00',
    distanceKm: 8.2,
    coverSrc: '/components/game-card/assets/game-cover-default.png',
    tag: '任务局',
    tagTone: 'task',
    price: '￥0/人',
    title: 'AI赋能系统搭建交流局',
    location: '📍黄浦区 · 8.2km · 3/8人',
    time: '⏰2026年5月1日 14:00--16:00',
    action: '加入',
    joinedText: '+3位玩家已入局',
    avatarFallbacks: ['AI', '搭', '局'],
    actions: DEFAULT_EVENT_ACTIONS
  },
  {
    id: 'coffee-social',
    type: 'social',
    typeText: '社交局',
    categoryKey: 'social',
    startAt: '2026-05-02T10:00:00+08:00',
    distanceKm: 2.4,
    coverSrc: '/components/game-card/assets/game-cover-default.png',
    tag: '社交局',
    typeTone: 'explore',
    price: '￥29/人',
    title: '周末咖啡创业交流局',
    location: '📍静安区 · 2.4km · 5/8人',
    time: '⏰2026年5月2日 10:00--12:00',
    action: '加入',
    joinedText: '+5位玩家已入局',
    avatarFallbacks: ['咖', '创', '聊'],
    actions: DEFAULT_EVENT_ACTIONS
  }
]

function getTypeFilterLabel(key) {
  const matched = TYPE_FILTERS.find((item) => item.key === key)

  return matched ? matched.label : TYPE_FILTERS[0].label
}

function getNextTypeFilterKey(currentKey) {
  const currentIndex = TYPE_FILTERS.findIndex((item) => item.key === currentKey)
  const nextIndex = currentIndex > -1 ? currentIndex + 1 : 1

  return TYPE_FILTERS[nextIndex % TYPE_FILTERS.length].key
}

function parseChineseStartTime(text = '') {
  const matched = text.match(/(\d{4})年(\d{1,2})月(\d{1,2})日(?:-\d{1,2}日)?\s*(\d{1,2}):(\d{2})/)

  if (!matched) {
    return Number.MAX_SAFE_INTEGER
  }

  const [, year, month, day, hour, minute] = matched
  const normalized = `${year}-${month.padStart(2, '0')}-${day.padStart(2, '0')}T${hour.padStart(2, '0')}:${minute}:00+08:00`
  const timestamp = Date.parse(normalized)

  return Number.isNaN(timestamp) ? Number.MAX_SAFE_INTEGER : timestamp
}

function getEventStartTime(item = {}) {
  if (typeof item.startAt === 'number') {
    return item.startAt
  }

  if (item.startAt) {
    const timestamp = Date.parse(item.startAt)

    if (!Number.isNaN(timestamp)) {
      return timestamp
    }
  }

  return parseChineseStartTime(item.time || item.timeText || '')
}

function getEventDistance(item = {}) {
  if (typeof item.distanceKm === 'number') {
    return item.distanceKm
  }

  const text = item.distanceText || item.location || ''
  const matched = text.match(/(\d+(?:\.\d+)?)\s*km/i)

  return matched ? Number(matched[1]) : Number.MAX_SAFE_INTEGER
}

function getSortedEvents(list, sortKey, sortOrder) {
  if (!sortKey) {
    return list
  }

  const direction = sortOrder === 'desc' ? -1 : 1

  return list.slice().sort((prev, next) => {
    const prevValue = sortKey === 'distance' ? getEventDistance(prev) : getEventStartTime(prev)
    const nextValue = sortKey === 'distance' ? getEventDistance(next) : getEventStartTime(next)
    const diff = prevValue - nextValue

    if (diff !== 0) {
      return diff * direction
    }

    return String(prev.id).localeCompare(String(next.id))
  })
}

function getDisplayEvents(options = {}) {
  const activeFilter = options.activeFilter || 'all'
  const activeTypeFilter = options.activeTypeFilter || 'all'
  let list = eventsList.slice()

  if (activeFilter !== 'all') {
    list = list.filter((item) => item.categoryKey === activeFilter || item.type === activeFilter)
  }

  if (activeTypeFilter !== 'all') {
    list = list.filter((item) => item.type === activeTypeFilter)
  }

  return getSortedEvents(list, options.sortKey, options.sortOrder)
}

Page({
  data: {
    onlineText: '3999人在线',
    keyword: '',
    hallScrollTop: 0,
    featuredCover: '/pages/game/hall/assets/hall-featured-city.jpg',
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    categories: [
      { key: 'social', name: '社交局', iconSrc: '/pages/game/hall/assets/category-social.png', className: 'social' },
      { key: 'task', name: '任务局', iconSrc: '/pages/game/hall/assets/category-task.png', className: 'task' },
      { key: 'growth', name: '成长局', iconSrc: '/pages/game/hall/assets/category-growth.png', className: 'growth' },
      { key: 'income', name: '变现局', iconSrc: '/pages/game/hall/assets/category-income.png', className: 'income' },
      { key: 'more', name: '更多', iconSrc: '/pages/game/hall/assets/category-more.png', className: 'more' }
    ],
    activeFilter: 'all',
    activeTypeFilter: 'all',
    typeFilterText: getTypeFilterLabel('all'),
    sortKey: '',
    sortOrder: 'asc',
    sortArrow: '▶',
    eventsList,
    displayEventsList: eventsList
  },

  onSearchInput(event) {
    this.setData({
      keyword: event.detail.value || ''
    })
  },

  onSearch() {
    toast.info('搜索功能开发中')
  },

  onBannerTap() {
    toast.info('官方局详情开发中')
  },

  onCategoryTap(event) {
    const key = event.currentTarget.dataset.type || 'all'

    if (key === 'more') {
      toast.info('更多分类开发中')
      return
    }

    this.updateDisplayEvents({
      activeFilter: key
    })
  },

  toggleTimeFilter() {
    this.updateDisplayEvents({
      sortKey: 'time',
      sortOrder: 'asc'
    })
  },

  toggleLocationFilter() {
    this.updateDisplayEvents({
      sortKey: 'distance',
      sortOrder: 'asc'
    })
  },

  toggleTypeFilter() {
    this.updateDisplayEvents({
      activeTypeFilter: getNextTypeFilterKey(this.data.activeTypeFilter)
    })
  },

  toggleSort() {
    const sortKey = this.data.sortKey || 'time'
    const sortOrder = this.data.sortKey && this.data.sortOrder === 'asc' ? 'desc' : 'asc'

    this.updateDisplayEvents({
      sortKey,
      sortOrder
    })
  },

  updateDisplayEvents(nextState = {}) {
    const activeFilter = nextState.activeFilter || this.data.activeFilter
    const activeTypeFilter = nextState.activeTypeFilter || this.data.activeTypeFilter
    const sortKey = nextState.sortKey == null ? this.data.sortKey : nextState.sortKey
    const sortOrder = nextState.sortOrder || this.data.sortOrder

    this.setData({
      activeFilter,
      activeTypeFilter,
      typeFilterText: getTypeFilterLabel(activeTypeFilter),
      sortKey,
      sortOrder,
      sortArrow: sortKey ? (sortOrder === 'desc' ? '▼' : '▲') : '▶',
      displayEventsList: getDisplayEvents({
        activeFilter,
        activeTypeFilter,
        sortKey,
        sortOrder
      })
    })
  },

  onViewDetail(event) {
    const item = event.detail && event.detail.item
    const id = (item && item.id) || event.currentTarget.dataset.id || ''

    wx.navigateTo({
      url: `/${ROUTES.gameDetail}?id=${id}`
    })
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (key === 'up' || key === 'down') {
      if (!this.suppressNextNavTap) {
        this.scrollHall(key, HALL_SCROLL_TAP_STEP_RPX)
      }
      return
    }

    if (key === 'left' || key === 'right') {
      toast.info('功能正在开发中')
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
    this.stopHallScrollHold(false)
    this.scrollHall(key, HALL_SCROLL_HOLD_STEP_RPX)

    this.hallScrollHoldTimer = setInterval(() => {
      this.scrollHall(key, HALL_SCROLL_HOLD_STEP_RPX)
    }, HALL_SCROLL_HOLD_INTERVAL_MS)
  },

  handleShellNavTouchEnd() {
    this.stopHallScrollHold(true)
  },

  handleShellAction(key) {
    if (key === 'home') {
      this.scrollHallToTop()
      return
    }

    if (key === 'search') {
      this.onSearch()
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
    if (!route || route === ROUTES.gameHall) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  },

  handleHallScroll(event) {
    const scrollTop = event.detail && event.detail.scrollTop

    if (typeof scrollTop === 'number') {
      this.hallScrollTopValue = scrollTop
    }
  },

  scrollHall(direction, stepRpx = HALL_SCROLL_TAP_STEP_RPX) {
    const current = Number(this.hallScrollTopValue || this.data.hallScrollTop || 0)
    const distance = this.rpxToPx(stepRpx)
    const nextTop = direction === 'up'
      ? Math.max(0, current - distance)
      : current + distance

    this.hallScrollTopValue = nextTop
    this.setData({
      hallScrollTop: nextTop
    })
  },

  scrollHallToTop() {
    this.hallScrollTopValue = 0
    this.setData({
      hallScrollTop: 0
    })
  },

  stopHallScrollHold(resetTapSuppress) {
    if (this.hallScrollHoldTimer) {
      clearInterval(this.hallScrollHoldTimer)
      this.hallScrollHoldTimer = null
    }

    if (resetTapSuppress && this.suppressNextNavTap) {
      if (this.hallScrollSuppressTimer) {
        clearTimeout(this.hallScrollSuppressTimer)
      }

      this.hallScrollSuppressTimer = setTimeout(() => {
        this.suppressNextNavTap = false
        this.hallScrollSuppressTimer = null
      }, HALL_SCROLL_HOLD_SUPPRESS_TAP_MS)
    }
  },

  clearHallScrollTimers() {
    this.stopHallScrollHold(false)

    if (this.hallScrollSuppressTimer) {
      clearTimeout(this.hallScrollSuppressTimer)
      this.hallScrollSuppressTimer = null
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

  onUnload() {
    this.clearHallScrollTimers()
  }
})
