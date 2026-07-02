const toast = require('../../../utils/toast')
const gameService = require('../../../services/game')
const { ROUTES } = require('../../../config/routes')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')

const HALL_SCROLL_TAP_STEP_RPX = 360
const HALL_SCROLL_HOLD_STEP_RPX = 72
const HALL_SCROLL_HOLD_INTERVAL_MS = 80
const HALL_SCROLL_HOLD_SUPPRESS_TAP_MS = 120
const TYPE_FILTERS = [
  { key: 'all', label: '类型' }
]
const ADVANCED_LOCATION_OPTIONS = [
  { key: 'all', name: '全国' }
]
const ADVANCED_CATEGORY_OPTIONS = [
  { key: 'all', name: '全部' }
]
const ADVANCED_SORT_OPTIONS = []
const CALENDAR_WEEKDAYS = ['日', '一', '二', '三', '四', '五', '六']
const DEFAULT_ADVANCED_DRAFT = {
  locationScope: 'all',
  cityName: '',
  categoryKey: 'all',
  sortMode: 'comprehensive',
  selectedDate: ''
}
const CATEGORY_ICON_MAP = {
  social: '/pages/game/hall/assets/category-social.png',
  task: '/pages/game/hall/assets/category-task.png',
  growth: '/pages/game/hall/assets/category-growth.png',
  income: '/pages/game/hall/assets/category-income.png',
  explore: '/pages/game/hall/assets/category-income.png',
  more: '/pages/game/hall/assets/category-more.png'
}
const CATEGORY_CLASS_MAP = {
  social: 'social',
  task: 'task',
  growth: 'growth',
  income: 'income',
  explore: 'income',
  more: 'more'
}

const eventsList = []

function getTypeFilterLabel(key, filters = TYPE_FILTERS) {
  const matched = filters.find((item) => item.key === key)

  return matched ? matched.label : filters[0].label
}

function getNextTypeFilterKey(currentKey, filters = TYPE_FILTERS) {
  const currentIndex = filters.findIndex((item) => item.key === currentKey)
  const nextIndex = currentIndex > -1 ? currentIndex + 1 : 1

  return filters[nextIndex % filters.length].key
}

function getAdvancedDraft(defaults = {}) {
  return Object.assign({}, DEFAULT_ADVANCED_DRAFT, defaults)
}

function normalizeCategoryList(data = {}) {
  const source = Array.isArray(data.primaryCategories) ? data.primaryCategories : []
  const list = source
    .filter((item) => item && item.visible !== false && item.key)
    .sort((left, right) => Number(left.order || 0) - Number(right.order || 0))
    .map((item) => {
      const key = String(item.key || '').trim()
      const iconKey = String(item.icon || key).replace(/^category-/, '')

      return {
        key,
        name: String(item.name || key).trim(),
        iconSrc: CATEGORY_ICON_MAP[key] || CATEGORY_ICON_MAP[iconKey] || CATEGORY_ICON_MAP.more,
        className: CATEGORY_CLASS_MAP[key] || CATEGORY_CLASS_MAP[iconKey] || 'more'
      }
    })

  return list.length ? list : null
}

function normalizeAdvancedCategoryOptions(data = {}) {
  const source = Array.isArray(data.primaryCategories) ? data.primaryCategories : []
  const list = source
    .filter((item) => item && item.visible !== false && item.key)
    .sort((left, right) => Number(left.order || 0) - Number(right.order || 0))
    .map((item) => ({
      key: String(item.key || '').trim(),
      name: String(item.name || item.key || '').trim()
    }))
    .filter((item) => item.key && item.name)

  return list.length ? [{ key: 'all', name: '全部' }].concat(list) : null
}

function normalizeTypeFilters(data = {}) {
  const source = Array.isArray(data.typeFilters) ? data.typeFilters : []
  const list = source
    .filter((item) => item && item.visible !== false && item.key)
    .sort((left, right) => Number(left.order || 0) - Number(right.order || 0))
    .map((item) => ({
      key: String(item.key || '').trim(),
      label: String(item.label || item.name || item.key || '').trim()
    }))
    .filter((item) => item.key && item.label)

  return list.length ? list : null
}

function normalizeLocationOptions(data = {}) {
  const source = Array.isArray(data.locationFilters) ? data.locationFilters : []
  const list = source
    .filter((item) => item && item.visible !== false && item.key)
    .sort((left, right) => Number(left.order || 0) - Number(right.order || 0))
    .map((item) => ({
      key: String(item.key || '').trim(),
      name: String(item.name || item.label || item.key || '').trim()
    }))
    .filter((item) => item.key && item.name)

  return list.length ? list : null
}

function gameTypeLabel(type = '') {
  const map = {
    free: '免费局',
    standard: '普通局',
    public_welfare: '公益局',
    aa: 'AA局',
    crowdfund: '众筹局',
    deposit: '押金局',
    condition: '条件局'
  }

  return map[type] || type || '组局'
}

function formatHallDate(value) {
  if (!value) {
    return ''
  }

  return String(value).replace('T', ' ').replace(/:\d{2}(?:\.\d+)?Z?$/, '')
}

function formatHallDistance(game = {}) {
  if (game.distanceLabel) {
    return game.distanceLabel
  }

  const meter = Number(game.distanceMeter || 0)
  if (meter > 0) {
    return meter >= 1000 ? `${(meter / 1000).toFixed(1)}km` : `${Math.round(meter)}m`
  }

  return ''
}

function normalizeAvatarFallbacks(title = '') {
  const chars = Array.from(String(title || '').replace(/\s/g, '')).slice(0, 3)

  return chars.length ? chars : ['局']
}

function normalizeEventActions(data = {}) {
  return Array.isArray(data.eventActions) ? data.eventActions.filter(Boolean) : []
}

function normalizeSortOptions(data = {}) {
  const source = Array.isArray(data.sortOptions) ? data.sortOptions : []

  return source.map((item) => ({
    key: String(item.key || '').trim(),
    name: String(item.name || item.label || item.key || '').trim(),
    sortKey: String(item.sortKey || '').trim(),
    sortOrder: String(item.sortOrder || 'asc').trim() || 'asc'
  })).filter((item) => item.key && item.name)
}

function normalizeHallGame(game = {}, eventActions = []) {
  const currentPlayers = Number(game.currentPlayers || 0)
  const maxPlayers = Number(game.maxPlayers || 8)
  const gameType = String(game.gameType || game.type || 'free')
  const typeText = gameTypeLabel(gameType)
  const categoryKey = String(game.primaryCategory || game.categoryKey || gameType || 'all')
  const locationParts = [
    game.address || game.cityName || '地点待定',
    formatHallDistance(game),
    `${currentPlayers}/${maxPlayers}人`
  ].filter(Boolean)

  return {
    id: game.id,
    type: gameType,
    typeText,
    categoryKey,
    startAt: game.createdAt || '',
    distanceKm: Number(game.distanceMeter || 0) / 1000,
    cityName: game.cityName || '',
    heatScore: currentPlayers,
    creditScore: 0,
    coverSrc: game.coverSrc || '/components/game-card/assets/game-cover-default.png',
    tag: game.primaryCategoryText || game.secondaryCategoryText || typeText,
    price: gameType === 'free' ? '0元/人' : '',
    title: game.title || '未命名组局',
    official: Boolean(game.official || game.isOfficial || game.featured || game.isFeatured),
    location: locationParts.join(' · '),
    time: game.createdAt ? `发布 ${formatHallDate(game.createdAt)}` : '',
    action: currentPlayers >= maxPlayers ? '已满员' : '加入',
    joinedText: `${currentPlayers}位玩家已入局`,
    avatarFallbacks: normalizeAvatarFallbacks(game.title),
    actions: eventActions
  }
}

function normalizeHallGames(data = {}, eventActions = []) {
  const source = Array.isArray(data.items)
    ? data.items
    : Array.isArray(data.games)
      ? data.games
      : Array.isArray(data)
        ? data
        : []

  return source.map((item) => normalizeHallGame(item, eventActions)).filter((item) => item.id)
}

function getAdvancedSortByState(sortKey, sortOrder, sortOptions = ADVANCED_SORT_OPTIONS) {
  const matched = sortOptions.find((item) => {
    return item.sortKey === sortKey && item.sortOrder === sortOrder
  })

  return matched ? matched.key : ''
}

function getAdvancedSortByKey(key, sortOptions = ADVANCED_SORT_OPTIONS) {
  return sortOptions.find((item) => item.key === key) || sortOptions[0] || { sortKey: '', sortOrder: 'asc' }
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
  const keyword = String(options.keyword || '').trim().toLowerCase()
  let list = (Array.isArray(options.eventsList) ? options.eventsList : eventsList).slice()

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

  if (keyword) {
    list = list.filter((item) => {
      return [item.title, item.location, item.tag, item.typeText, item.cityName].some((value) => {
        return String(value || '').toLowerCase().indexOf(keyword) > -1
      })
    })
  }

  return getSortedEvents(list, options.sortKey, options.sortOrder)
}

Page({
  data: {
    onlineText: '在线',
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
    categories: [],
    activeFilter: 'all',
    activeTypeFilter: 'all',
    eventActions: [],
    activeLocationScope: 'all',
    activeCityName: '',
    selectedDate: '',
    typeFilterText: getTypeFilterLabel('all'),
    typeFilters: TYPE_FILTERS,
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
    displayEventsList: eventsList,
    loading: false
  },

  onLoad() {
    this.loadCategoryConfig()
    this.loadGames()
  },

  loadCategoryConfig() {
    gameService.getCategoryConfig().then((data) => {
      const categories = normalizeCategoryList(data)
      const advancedCategoryOptions = normalizeAdvancedCategoryOptions(data)
      const typeFilters = normalizeTypeFilters(data)
      const locationOptions = normalizeLocationOptions(data)
      const sortOptions = normalizeSortOptions(data)
      const eventActions = normalizeEventActions(data)
      const nextData = {}

      if (categories) {
        nextData.categories = categories
      }

      if (advancedCategoryOptions) {
        nextData.advancedCategoryOptions = advancedCategoryOptions
      }

      if (typeFilters) {
        nextData.typeFilters = typeFilters
        nextData.typeFilterText = getTypeFilterLabel(this.data.activeTypeFilter, typeFilters)
      }

      if (locationOptions) {
        nextData.advancedLocationOptions = locationOptions
      }

      if (sortOptions.length) {
        nextData.advancedSortOptions = sortOptions
      }

      if (eventActions.length) {
        nextData.eventActions = eventActions
      }

      if (Object.keys(nextData).length) {
        this.setData(nextData)
        this.setData({
          eventsList: (this.data.eventsList || []).map((item) => Object.assign({}, item, {
            actions: eventActions.length ? eventActions : item.actions
          }))
        })
        this.updateDisplayEvents({})
      }
    }).catch(() => {})
  },

  async loadGames() {
    this.setData({ loading: true })

    try {
      const data = await gameService.getGameList()
      const events = normalizeHallGames(data, this.data.eventActions)

      this.setData({
        loading: false,
        eventsList: events
      })
      this.updateDisplayEvents({})
    } catch (error) {
      this.setData({
        loading: false,
        eventsList: [],
        displayEventsList: []
      })
      toast.info(error.message || '组局列表加载失败')
    }
  },

  onSearchInput(event) {
    this.setData({
      keyword: event.detail.value || ''
    })
  },

  onSearch() {
    this.updateDisplayEvents({})
  },

  onBannerTap() {
    const featured = this.data.eventsList.find((item) => (
      item.official ||
      item.tag === '官方局' ||
      String(item.title || '').indexOf('官方') > -1 ||
      String(item.title || '').indexOf('首发') > -1
    )) || this.data.eventsList[0]

    if (featured && featured.id) {
      navigateShellRoute(`${ROUTES.gameDetail}?id=${featured.id}`, {
        currentRoute: ROUTES.gameHall
      })
      return
    }

    this.scrollHallToTop()
  },

  onCategoryTap(event) {
    const key = event.currentTarget.dataset.type || 'all'

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
      activeTypeFilter: getNextTypeFilterKey(this.data.activeTypeFilter, this.data.typeFilters || TYPE_FILTERS)
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
    const categoryOptions = this.data.advancedCategoryOptions || ADVANCED_CATEGORY_OPTIONS
    const categoryKey = categoryOptions.some((item) => item.key === this.data.activeFilter)
      ? this.data.activeFilter
      : 'all'
    const draft = getAdvancedDraft({
      locationScope: this.data.activeLocationScope,
      cityName: this.data.activeCityName,
      categoryKey,
      sortMode: getAdvancedSortByState(this.data.sortKey, this.data.sortOrder, this.data.advancedSortOptions),
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
        cityName: this.data.advancedDraft.cityName || this.data.activeCityName || ''
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
    const sortOption = getAdvancedSortByKey(draft.sortMode, this.data.advancedSortOptions)

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
      typeFilterText: getTypeFilterLabel(activeTypeFilter, this.data.typeFilters || TYPE_FILTERS),
      sortKey,
      sortOrder,
      sortArrow: sortKey ? (sortOrder === 'desc' ? '▼' : '▲') : '▶',
      displayEventsList: getDisplayEvents({
        eventsList: this.data.eventsList,
        keyword: this.data.keyword,
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

    navigateShellRoute(`${ROUTES.gameDetail}?id=${id}`, {
      currentRoute: ROUTES.gameHall
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

    if (navigateShellKey(key, {
      currentRoute: ROUTES.gameHall,
      onSameRoute: () => this.scrollHallToTop()
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

    navigateShellRoute(route, {
      currentRoute: ROUTES.gameHall
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
