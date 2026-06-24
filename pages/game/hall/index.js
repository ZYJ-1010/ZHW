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
const ADVANCED_LOCATION_OPTIONS = [
  { key: 'all', name: '全国' },
  { key: 'nearby', name: '附近(50km)' }
]
const ADVANCED_CATEGORY_OPTIONS = [
  { key: 'all', name: '全部' },
  { key: 'social', name: '社交局' },
  { key: 'explore', name: '探索局' },
  { key: 'task', name: '任务局' },
  { key: 'growth', name: '成长局' }
]
const ADVANCED_SORT_OPTIONS = [
  { key: 'comprehensive', name: '综合排序', sortKey: '', sortOrder: 'asc' },
  { key: 'latest', name: '最新发布', sortKey: 'time', sortOrder: 'desc' },
  { key: 'hot', name: '热度最高', sortKey: 'hot', sortOrder: 'desc' },
  { key: 'distance', name: '距离最近', sortKey: 'distance', sortOrder: 'asc' },
  { key: 'credit', name: '信用优先', sortKey: 'credit', sortOrder: 'desc' }
]
const CALENDAR_WEEKDAYS = ['日', '一', '二', '三', '四', '五', '六']
const DEFAULT_ADVANCED_DRAFT = {
  locationScope: 'all',
  cityName: '',
  categoryKey: 'all',
  sortMode: 'comprehensive',
  selectedDate: ''
}

const eventsList = [
  {
    id: 'deposit-morning',
    type: 'deposit',
    typeText: '押金局',
    categoryKey: 'growth',
    startAt: '2026-05-01T07:00:00+08:00',
    distanceKm: 8.2,
    cityName: '上海',
    heatScore: 86,
    creditScore: 92,
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
    cityName: '上海',
    heatScore: 78,
    creditScore: 95,
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
    cityName: '上海',
    heatScore: 92,
    creditScore: 88,
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

function getAdvancedDraft(defaults = {}) {
  return Object.assign({}, DEFAULT_ADVANCED_DRAFT, defaults)
}

function getAdvancedSortByState(sortKey, sortOrder) {
  const matched = ADVANCED_SORT_OPTIONS.find((item) => {
    return item.sortKey === sortKey && item.sortOrder === sortOrder
  })

  return matched ? matched.key : 'comprehensive'
}

function getAdvancedSortByKey(key) {
  return ADVANCED_SORT_OPTIONS.find((item) => item.key === key) || ADVANCED_SORT_OPTIONS[0]
}

function getCurrentDateInfo() {
  const now = new Date()

  return {
    year: now.getFullYear(),
    month: now.getMonth() + 1,
    date: [
      now.getFullYear(),
      String(now.getMonth() + 1).padStart(2, '0'),
      String(now.getDate()).padStart(2, '0')
    ].join('-')
  }
}

function formatCalendarTitle(year, month) {
  return `${year}年 ${month}月`
}

function formatCalendarDate(year, month, day) {
  return [
    year,
    String(month).padStart(2, '0'),
    String(day).padStart(2, '0')
  ].join('-')
}

function buildCalendarDays(year, month) {
  const firstDate = new Date(year, month - 1, 1)
  const firstDay = firstDate.getDay()
  const daysInMonth = new Date(year, month, 0).getDate()
  const today = getCurrentDateInfo().date
  const days = []

  for (let index = 0; index < firstDay; index += 1) {
    days.push({
      id: `empty-${index}`,
      empty: true
    })
  }

  for (let day = 1; day <= daysInMonth; day += 1) {
    const date = formatCalendarDate(year, month, day)

    days.push({
      id: date,
      day,
      date,
      today: date === today
    })
  }

  return days
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

function getEventSortValue(item = {}, sortKey) {
  if (sortKey === 'distance') {
    return getEventDistance(item)
  }

  if (sortKey === 'hot') {
    return Number(item.heatScore || item.hotScore || item.joinedCount || 0)
  }

  if (sortKey === 'credit') {
    return Number(item.creditScore || 0)
  }

  return getEventStartTime(item)
}

function isSameEventDate(item = {}, selectedDate) {
  if (!selectedDate) {
    return true
  }

  const startTime = getEventStartTime(item)

  if (!Number.isFinite(startTime)) {
    return false
  }

  const date = new Date(startTime)
  const eventDate = formatCalendarDate(date.getFullYear(), date.getMonth() + 1, date.getDate())

  return eventDate === selectedDate
}

function getSortedEvents(list, sortKey, sortOrder) {
  if (!sortKey) {
    return list
  }

  const direction = sortOrder === 'desc' ? -1 : 1

  return list.slice().sort((prev, next) => {
    const prevValue = getEventSortValue(prev, sortKey)
    const nextValue = getEventSortValue(next, sortKey)
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
  const activeLocationScope = options.activeLocationScope || 'all'
  const activeCityName = options.activeCityName || ''
  const selectedDate = options.selectedDate || ''
  let list = eventsList.slice()

  if (activeFilter !== 'all') {
    list = list.filter((item) => item.categoryKey === activeFilter || item.type === activeFilter)
  }

  if (activeTypeFilter !== 'all') {
    list = list.filter((item) => item.type === activeTypeFilter)
  }

  if (activeLocationScope === 'nearby') {
    list = list.filter((item) => getEventDistance(item) <= 50)
  }

  if (activeLocationScope === 'city' && activeCityName) {
    list = list.filter((item) => item.cityName === activeCityName)
  }

  if (selectedDate) {
    list = list.filter((item) => isSameEventDate(item, selectedDate))
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
    activeLocationScope: 'all',
    activeCityName: '',
    selectedDate: '',
    typeFilterText: getTypeFilterLabel('all'),
    sortKey: '',
    sortOrder: 'asc',
    sortArrow: '▶',
    advancedFilterVisible: false,
    advancedLocationOptions: ADVANCED_LOCATION_OPTIONS,
    advancedCategoryOptions: ADVANCED_CATEGORY_OPTIONS,
    advancedSortOptions: ADVANCED_SORT_OPTIONS,
    advancedDraft: getAdvancedDraft(),
    calendarWeekdays: CALENDAR_WEEKDAYS,
    calendarYear: getCurrentDateInfo().year,
    calendarMonth: getCurrentDateInfo().month,
    calendarTitle: formatCalendarTitle(getCurrentDateInfo().year, getCurrentDateInfo().month),
    calendarDays: buildCalendarDays(getCurrentDateInfo().year, getCurrentDateInfo().month),
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

  openAdvancedFilter() {
    const categoryKey = ADVANCED_CATEGORY_OPTIONS.some((item) => item.key === this.data.activeFilter)
      ? this.data.activeFilter
      : 'all'
    const draft = getAdvancedDraft({
      locationScope: this.data.activeLocationScope,
      cityName: this.data.activeCityName,
      categoryKey,
      sortMode: getAdvancedSortByState(this.data.sortKey, this.data.sortOrder),
      selectedDate: this.data.selectedDate
    })

    this.setData({
      advancedFilterVisible: true,
      advancedDraft: draft
    })
  },

  closeAdvancedFilter() {
    this.setData({
      advancedFilterVisible: false
    })
  },

  noop() {},

  selectAdvancedOption(event) {
    const field = event.currentTarget.dataset.field
    const key = event.currentTarget.dataset.key

    if (!field) {
      return
    }

    const nextDraft = Object.assign({}, this.data.advancedDraft, {
      [field]: key
    })

    if (field === 'locationScope' && key !== 'city') {
      nextDraft.cityName = ''
    }

    this.setData({
      advancedDraft: nextDraft
    })
  },

  selectAdvancedCity() {
    this.setData({
      advancedDraft: Object.assign({}, this.data.advancedDraft, {
        locationScope: 'city',
        cityName: this.data.advancedDraft.cityName || '上海'
      })
    })
  },

  changeAdvancedMonth(event) {
    const direction = event.currentTarget.dataset.direction
    const offset = direction === 'prev' ? -1 : 1
    const nextDate = new Date(this.data.calendarYear, this.data.calendarMonth - 1 + offset, 1)
    const calendarYear = nextDate.getFullYear()
    const calendarMonth = nextDate.getMonth() + 1

    this.setData({
      calendarYear,
      calendarMonth,
      calendarTitle: formatCalendarTitle(calendarYear, calendarMonth),
      calendarDays: buildCalendarDays(calendarYear, calendarMonth)
    })
  },

  selectAdvancedDate(event) {
    const selectedDate = event.currentTarget.dataset.date

    if (!selectedDate) {
      return
    }

    this.setData({
      advancedDraft: Object.assign({}, this.data.advancedDraft, {
        selectedDate
      })
    })
  },

  resetAdvancedFilter() {
    this.setData({
      advancedDraft: getAdvancedDraft()
    })
  },

  confirmAdvancedFilter() {
    const draft = this.data.advancedDraft
    const sortOption = getAdvancedSortByKey(draft.sortMode)

    this.updateDisplayEvents({
      activeFilter: draft.categoryKey || 'all',
      activeLocationScope: draft.locationScope || 'all',
      activeCityName: draft.cityName || '',
      selectedDate: draft.selectedDate || '',
      sortKey: sortOption.sortKey,
      sortOrder: sortOption.sortOrder
    })

    this.setData({
      advancedFilterVisible: false
    })
  },

  updateDisplayEvents(nextState = {}) {
    const activeFilter = nextState.activeFilter || this.data.activeFilter
    const activeTypeFilter = nextState.activeTypeFilter || this.data.activeTypeFilter
    const activeLocationScope = nextState.activeLocationScope == null ? this.data.activeLocationScope : nextState.activeLocationScope
    const activeCityName = nextState.activeCityName == null ? this.data.activeCityName : nextState.activeCityName
    const selectedDate = nextState.selectedDate == null ? this.data.selectedDate : nextState.selectedDate
    const sortKey = nextState.sortKey == null ? this.data.sortKey : nextState.sortKey
    const sortOrder = nextState.sortOrder || this.data.sortOrder

    this.setData({
      activeFilter,
      activeTypeFilter,
      activeLocationScope,
      activeCityName,
      selectedDate,
      typeFilterText: getTypeFilterLabel(activeTypeFilter),
      sortKey,
      sortOrder,
      sortArrow: sortKey ? (sortOrder === 'desc' ? '▼' : '▲') : '▶',
      displayEventsList: getDisplayEvents({
        activeFilter,
        activeTypeFilter,
        activeLocationScope,
        activeCityName,
        selectedDate,
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
