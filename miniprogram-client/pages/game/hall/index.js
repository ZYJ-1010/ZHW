const toast = require('../../../utils/toast')
const gameService = require('../../../services/game')
const locationService = require('../../../services/location')
const locationAccess = require('../../../utils/location-access')
const { ROUTES } = require('../../../config/routes')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')
const { toUserMessage } = require('../../../utils/user-message')

const HALL_SCROLL_TAP_STEP_RPX = 360
const DEFAULT_EVENT_ACTIONS = ['分享', '关注', '打招呼']
const HALL_SCROLL_HOLD_STEP_RPX = 72
const HALL_SCROLL_HOLD_INTERVAL_MS = 80
const HALL_SCROLL_HOLD_SUPPRESS_TAP_MS = 120
const TYPE_FILTERS = [
  { key: 'all', label: '全部' }
]
const ADVANCED_LOCATION_OPTIONS = [
  { key: 'all', name: '全国' }
]
const ADVANCED_CATEGORY_OPTIONS = [
  { key: 'all', name: '全部' }
]
const SORT_OPTIONS = [
  { key: 'latest', name: '最新发布', sortKey: 'time', sortOrder: 'desc' },
  { key: 'hot', name: '热度最高', sortKey: 'hot', sortOrder: 'desc' },
  { key: 'distance', name: '距离最近', sortKey: 'distance', sortOrder: 'asc' },
  { key: 'credit', name: '信用优先', sortKey: 'credit', sortOrder: 'desc' }
]
const NEARBY_RADIUS_OPTIONS = [1000, 3000, 5000, 10000]
const CALENDAR_WEEKDAYS = ['日', '一', '二', '三', '四', '五', '六']
const DEFAULT_ADVANCED_DRAFT = {
  locationScope: 'all',
  cityName: '',
  categoryKey: 'all',
  sortMode: '',
  selectedDate: ''
}
const CATEGORY_ICON_MAP = {
  social: 'https://static.haowan.net.cn/miniprogram/pages/game/hall/assets/category-social.png',
  task: 'https://static.haowan.net.cn/miniprogram/pages/game/hall/assets/category-task.png',
  growth: '/pages/game/hall/assets/category-growth.png',
  income: 'https://static.haowan.net.cn/miniprogram/pages/game/hall/assets/category-income.png',
  explore: 'https://static.haowan.net.cn/miniprogram/pages/game/hall/assets/category-income.png',
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

function normalizeHallTypeFilters(data = {}) {
  const options = normalizeAdvancedCategoryOptions(data)

  if (!options) {
    return null
  }

  return options.map((item) => ({
    key: item.key,
    label: item.name
  }))
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
  return '免费局'
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

function normalizePlayerAvatars(game = {}) {
  const source = Array.isArray(game.playerAvatars) ? game.playerAvatars : []

  return source.map((avatar) => ({
    userId: avatar && avatar.userId,
    name: String((avatar && (avatar.name || avatar.displayName)) || '').trim(),
    imageUrl: String((avatar && (avatar.avatarUrl || avatar.imageUrl)) || '').trim(),
    text: String((avatar && (avatar.avatarText || avatar.text)) || '').trim()
  })).filter((avatar) => avatar.imageUrl || avatar.text).slice(0, 3)
}

function normalizeEventActions(data = {}) {
  const actions = Array.isArray(data.eventActions) ? data.eventActions.filter(Boolean) : []

  const shareComponent = data.shareComponent || {}
  const shareEnabled = shareComponent.enabled !== false
  const shareLabel = shareComponent.label || '分享'
  const normalized = actions.filter((item) => !['引荐', '邀请（站内）', '站内邀请'].includes(String(item || '').trim()))
    .filter((item) => shareEnabled || String(item || '').trim() !== shareLabel)
  if (shareEnabled && !normalized.includes(shareLabel)) normalized.unshift(shareLabel)
  return normalized.length ? normalized : (shareEnabled ? ['分享', '关注', '打招呼'] : ['关注', '打招呼'])
}

function normalizeSortOptions(data = {}) {
  const source = Array.isArray(data.sortOptions) ? data.sortOptions : []

  return source.map((item) => ({
    key: String(item.key || '').trim(),
    name: String(item.name || item.label || item.key || '').trim(),
    sortKey: String(item.sortKey || '').trim(),
    sortOrder: String(item.sortOrder || 'asc').trim() || 'asc'
  })).filter((item) => item.key && item.name && item.key !== 'comprehensive' && item.sortKey)
}

function normalizeHallGame(game = {}, eventActions = [], shareComponent = {}) {
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
    coverSrc: game.coverImage || game.coverUrl || game.coverSrc || 'https://static.haowan.net.cn/miniprogram/components/game-card/assets/game-cover-default.png',
    tag: game.primaryCategoryText || game.secondaryCategoryText || typeText,
    price: '',
    title: game.title || '未命名组局',
    official: Boolean(game.official || game.isOfficial || game.featured || game.isFeatured),
    location: locationParts.join(' · '),
    time: game.createdAt ? `发布 ${formatHallDate(game.createdAt)}` : '',
    action: game.status === 'in_progress'
      ? '已开局'
      : (currentPlayers >= maxPlayers || game.status === 'full' ? '已满员' : '招募中'),
    joinedText: `${currentPlayers}位玩家已入局`,
    playerAvatars: normalizePlayerAvatars(game),
    actions: eventActions,
    shareActionLabel: shareComponent.label || '分享',
    shareComponentVariant: shareComponent.variant === 'icon_button' ? 'icon_button' : 'channel_sheet'
  }
}

function normalizeHallGames(data = {}, eventActions = [], shareComponent = {}) {
  const source = Array.isArray(data.items)
    ? data.items
    : Array.isArray(data.games)
      ? data.games
      : Array.isArray(data)
        ? data
        : []

  return source.map((item) => normalizeHallGame(item, eventActions, shareComponent)).filter((item) => item.id)
}

function getSortOptionByState(sortKey, sortOrder, sortOptions = SORT_OPTIONS) {
  const matched = sortOptions.find((item) => {
    return item.sortKey === sortKey && item.sortOrder === sortOrder
  })

  return matched ? matched.key : ''
}

function getSortOptionByKey(key, sortOptions = SORT_OPTIONS) {
  return sortOptions.find((item) => item.key === key) || null
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
    featuredCover: 'https://static.haowan.net.cn/miniprogram/assets/game/hall/hall-featured-city.jpg',
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
    eventActions: DEFAULT_EVENT_ACTIONS,
    shareComponent: { enabled: true, variant: 'channel_sheet', label: '分享' },
    activeLocationScope: 'all',
    nearbyRadiusMeters: 0,
    locationGuideVisible: false,
    locationStatusText: '',
    activeCityName: '',
    selectedDate: '',
    typeFilterText: getTypeFilterLabel('all'),
    typeFilters: TYPE_FILTERS,
    sortKey: '',
    sortOrder: 'asc',
    sortArrow: '▶',
    sortFilterVisible: false,
    typeFilterVisible: false,
    sortFilterText: '排序',
    sortFilterActive: false,
    activeSortMode: '',
    advancedLocationOptions: ADVANCED_LOCATION_OPTIONS,
    advancedCategoryOptions: ADVANCED_CATEGORY_OPTIONS,
    advancedSortOptions: SORT_OPTIONS,
    sortOptions: SORT_OPTIONS,
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
      // 局前大厅的“类型”筛选对应四大局类型，而不是免费/有偿等旧的 type
      // 字段。四类数据与创建页、卡片标签共用后台 primaryCategories 配置。
      const typeFilters = normalizeHallTypeFilters(data)
      const locationOptions = normalizeLocationOptions(data)
      const sortOptions = normalizeSortOptions(data)
      const eventActions = normalizeEventActions(data)
      const shareComponent = {
        enabled: data.shareComponent && data.shareComponent.enabled === false ? false : true,
        variant: data.shareComponent && data.shareComponent.variant === 'icon_button' ? 'icon_button' : 'channel_sheet',
        label: String(data.shareComponent && data.shareComponent.label || '分享').trim() || '分享'
      }
      const nextData = {}

      if (categories) {
        nextData.categories = categories
      }

      if (advancedCategoryOptions) {
        nextData.advancedCategoryOptions = advancedCategoryOptions
      }

      if (typeFilters) {
        nextData.typeFilters = typeFilters
        nextData.typeFilterText = getTypeFilterLabel(this.data.activeFilter, typeFilters)
      }

      if (locationOptions) {
        nextData.advancedLocationOptions = locationOptions
      }

      if (sortOptions.length) {
        nextData.advancedSortOptions = sortOptions
        nextData.sortOptions = sortOptions
      }

      if (eventActions.length) {
        nextData.eventActions = eventActions
      }
      nextData.shareComponent = shareComponent

      if (Object.keys(nextData).length) {
        this.setData(nextData)
        this.setData({
          eventsList: (this.data.eventsList || []).map((item) => Object.assign({}, item, {
            actions: eventActions.length ? eventActions : item.actions,
            shareActionLabel: shareComponent.label,
            shareComponentVariant: shareComponent.variant
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
      const events = normalizeHallGames(data, this.data.eventActions, this.data.shareComponent)

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
    const isTimeSort = this.data.sortKey === 'time'
    const nextOrder = !isTimeSort
      ? 'asc'
      : (this.data.sortOrder === 'asc' ? 'desc' : '')

    this.updateDisplayEvents({
      sortKey: nextOrder ? 'time' : '',
      sortOrder: nextOrder || 'asc',
      activeSortMode: ''
    })
  },

  toggleLocationFilter() {
    const options = [0].concat(NEARBY_RADIUS_OPTIONS)
    const current = options.indexOf(this.data.nearbyRadiusMeters)
    const nearbyRadiusMeters = options[(current + 1) % options.length]

    if (nearbyRadiusMeters === 0) {
      const clearsDistanceSort = this.data.sortKey === 'distance'
      this.setData({ nearbyRadiusMeters, locationStatusText: '' })
      this.updateDisplayEvents({
        sortKey: clearsDistanceSort ? '' : this.data.sortKey,
        sortOrder: clearsDistanceSort ? 'asc' : this.data.sortOrder,
        activeSortMode: clearsDistanceSort ? '' : this.data.activeSortMode
      })
      this.loadGames()
      return
    }

    this.setData({ nearbyRadiusMeters })
    this.loadNearbyGames()
  },

  async loadNearbyGames() {
    this.setData({ loading: true, locationStatusText: '正在获取附近局' })
    try {
      const location = await locationAccess.getPreciseLocation()
      await this.applyNearbyLocation(location, '附近局')
    } catch (error) {
      this.setData({ loading: false, locationGuideVisible: true, locationStatusText: '未获取定位，可选择其他方式' })
    }
  },

  async applyNearbyLocation(location, label) {
    try {
      const data = await locationService.getNearbyGames({
        latitude: location.latitude,
        longitude: location.longitude,
        radiusMeters: this.data.nearbyRadiusMeters
      })
      const events = normalizeHallGames(data, this.data.eventActions, this.data.shareComponent)
      this.setData({
        loading: false,
        locationGuideVisible: false,
        locationStatusText: `${label}${this.data.nearbyRadiusMeters / 1000}km`,
        eventsList: events,
        sortKey: 'distance',
        sortOrder: 'asc'
      })
      this.updateDisplayEvents({})
    } catch (error) {
      this.setData({ loading: false, locationStatusText: toUserMessage(error && error.message, '附近局加载失败') })
      toast.info(error.message || '附近局加载失败')
    }
  },

  async chooseHallManualLocation() {
    try {
      const location = await locationAccess.chooseManualLocation()
      await this.applyNearbyLocation(location, '手动位置附近')
    } catch (error) { toast.info(error.message || '未选择位置') }
  },

  async useHallCityFallback() {
    try {
      const fallback = await locationAccess.getFallbackLocation()
      if (fallback.cityCode) {
        const data = await gameService.getSameCityGames({ cityCode: fallback.cityCode })
        const events = normalizeHallGames(data, this.data.eventActions, this.data.shareComponent)
        this.setData({ loading: false, locationGuideVisible: false, locationStatusText: `${fallback.cityName || '同城'}推荐`, eventsList: events, sortKey: 'time', sortOrder: 'desc' })
        this.updateDisplayEvents({})
        return
      }
      await this.applyNearbyLocation(fallback, fallback.message || '默认城市推荐')
    } catch (error) {
      this.setData({ loading: false, locationGuideVisible: false, locationStatusText: '按发布时间展示' })
      this.loadGames()
    }
  },

  async handleHallLocationGuideTap(event) {
    const action = event.currentTarget.dataset.action
    if (action === 'setting') return locationAccess.showDeniedGuide({ onManual: () => this.chooseHallManualLocation(), onFallback: () => this.useHallCityFallback() })
    if (action === 'manual') return this.chooseHallManualLocation()
    if (action === 'city') return this.useHallCityFallback()
  },

  openTypeFilter() {
    this.setData({ typeFilterVisible: true })
  },

  closeTypeFilter() {
    this.setData({ typeFilterVisible: false })
  },

  selectTypeFilter(event) {
    const key = String(event.currentTarget.dataset.key || '').trim()
    const option = (this.data.typeFilters || TYPE_FILTERS).find((item) => item.key === key)

    if (!option) {
      return
    }

    this.updateDisplayEvents({ activeFilter: option.key })
    this.closeTypeFilter()
  },

  resetTypeFilter() {
    this.updateDisplayEvents({ activeFilter: 'all' })
    this.closeTypeFilter()
  },

  toggleSort() {
    const sortKey = this.data.sortKey || 'time'
    const sortOrder = this.data.sortKey && this.data.sortOrder === 'asc' ? 'desc' : 'asc'

    this.updateDisplayEvents({
      sortKey,
      sortOrder
    })
  },

  openSortFilter() {
    this.setData({
      sortFilterVisible: true
    })
  },

  closeSortFilter() {
    this.setData({
      sortFilterVisible: false
    })
  },

  selectSortFilter(event) {
    const option = getSortOptionByKey(event.currentTarget.dataset.key, this.data.sortOptions || SORT_OPTIONS)

    if (!option) {
      return
    }

    this.updateDisplayEvents({
      sortKey: option.sortKey,
      sortOrder: option.sortOrder,
      activeSortMode: option.key
    })
    this.closeSortFilter()
  },

  resetSortFilter() {
    this.updateDisplayEvents({
      sortKey: '',
      sortOrder: 'asc',
      activeSortMode: ''
    })
    this.closeSortFilter()
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
    const sortOption = getSortOptionByKey(draft.sortMode, this.data.advancedSortOptions)

    this.updateDisplayEvents({
      activeFilter: draft.categoryKey || 'all',
      activeLocationScope: draft.locationScope || 'all',
      activeCityName: draft.cityName || '',
      selectedDate: draft.selectedDate || '',
      sortKey: sortOption ? sortOption.sortKey : '',
      sortOrder: sortOption ? sortOption.sortOrder : 'asc'
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
    const sortOptions = this.data.sortOptions || SORT_OPTIONS
    const matchedSortOption = getSortOptionByState(sortKey, sortOrder, sortOptions)
    const activeSortMode = nextState.activeSortMode == null
      ? (matchedSortOption ? matchedSortOption.key : '')
      : nextState.activeSortMode

    this.setData({
      activeFilter,
      activeTypeFilter,
      activeLocationScope,
      activeCityName,
      selectedDate,
      typeFilterText: getTypeFilterLabel(activeFilter, this.data.typeFilters || TYPE_FILTERS),
      sortKey,
      sortOrder,
      sortArrow: sortKey ? (sortOrder === 'desc' ? '▼' : '▲') : '▶',
      activeSortMode,
      sortFilterActive: Boolean(activeSortMode),
      sortFilterText: matchedSortOption ? matchedSortOption.name : '排序',
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

  async onCardAction(event) {
    const detail = event.detail || {}
    const item = detail.item || {}
    const gameId = Number(item.id || 0)
    const action = String(detail.action || detail.label || '').trim()

    if (!gameId) {
      toast.info('局信息不存在，请刷新后重试')
      return
    }

    if (action === 'primary') {
      this.onViewDetail({ detail: { item } })
      return
    }

    if (action === 'share' || action === this.data.shareComponent.label) {
      navigateShellRoute(`${ROUTES.gameShare}?id=${gameId}`, {
        currentRoute: ROUTES.gameHall,
        reuseExisting: false
      })
      return
    }

    if (action === '关注') {
      try {
        await gameService.favoriteGame(gameId)
        toast.success('已关注该局')
      } catch (error) {
        toast.info(error.message || '关注失败，请稍后重试')
      }
      return
    }

    if (action === '引荐') {
      try {
        const permission = await gameService.getInvitePermission({ gameId })
        if (!permission.allowed) {
          toast.info(permission.reason || '仅局创建者或主领路人可发起引荐')
          return
        }
        navigateShellRoute(`${ROUTES.gameInvite}?gameId=${gameId}`, {
          currentRoute: ROUTES.gameHall,
          reuseExisting: false
        })
      } catch (error) {
        toast.info(error.message || '引荐权限校验失败')
      }
      return
    }

    if (action === '打招呼') {
      try {
        const gameDetail = await gameService.getGameDetail(gameId)
        const relation = gameDetail && gameDetail.myRelation ? gameDetail.myRelation : {}
        if (relation.canEnterIM !== true) {
          toast.info(relation.isMember === true ? '局还未开' : '仅局内玩家可用，请先报名')
          return
        }
        navigateShellRoute(`${ROUTES.gameGreet}?gameId=${gameId}`, {
          currentRoute: ROUTES.gameHall,
          reuseExisting: false
        })
      } catch (error) {
        toast.info(error.message || '局内消息权限校验失败')
      }
    }
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
