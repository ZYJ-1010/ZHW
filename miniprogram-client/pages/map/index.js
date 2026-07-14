const locationService = require('../../services/location')
const mapService = require('../../services/map')
const { ROUTES } = require('../../config/routes')
const { navigateShellKey, navigateShellRoute } = require('../../utils/shell-nav')

const FALLBACK_MAP_CONFIG = {
  onlineText: '在线',
  defaultLocation: {
  latitude: 31.2304,
  longitude: 121.4737
  },
  defaultRadiusMeters: 3000,
  radiusOptions: [1000, 3000, 5000],
  mapFilters: ['附近组局', '组局路线', '热力图', '好友分布', '解锁图鉴', 'AR'],
  texts: {
    searchPlaceholder: '搜局、搜人、搜地块...',
    loadingNearbyText: '正在获取附近数据...',
    locateToolText: '定',
    refreshToolText: '刷',
    dateRangeText: '2025.09.23 - 2026.03.30',
    detailActionText: '查看详情',
    joinActionText: '去组队',
    playerDetailActionText: '查看资料',
    currentInteractText: '当前可互动',
    offlineInteractText: '暂未在线，可查看轨迹',
    nearbySectionTitle: '附近玩法点',
    nearbyEmptyText: '当前位置附近暂无玩法点'
  }
}
const PLAYER_MARKER_BASE_ID = 1000
const MARKER_ICON = 'https://static.haowan.net.cn/miniprogram/components/participant-card/assets/icon-location.png'
const ONLINE_PLAYER_ICON = 'https://static.haowan.net.cn/miniprogram/pages/home/player/assets/ranking-avatar-01.png'
const OFFLINE_PLAYER_ICON = 'https://static.haowan.net.cn/miniprogram/pages/home/player/assets/ranking-avatar-03.png'
const FALLBACK_OFFSETS = [
  { latitude: 0.0022, longitude: 0.0016 },
  { latitude: -0.0018, longitude: 0.0026 },
  { latitude: 0.001, longitude: -0.0028 },
  { latitude: -0.0025, longitude: -0.0014 }
]
function toNumber(value) {
  const number = Number(value)

  return Number.isFinite(number) ? number : null
}

function getCoordinate(item, center, index) {
  const rawLocation = item.location || item.coordinate || item.coords || {}
  const latitude = toNumber(item.latitude || item.lat || rawLocation.latitude || rawLocation.lat)
  const longitude = toNumber(item.longitude || item.lng || rawLocation.longitude || rawLocation.lng)

  if (latitude !== null && longitude !== null) {
    return { latitude, longitude }
  }

  const offset = FALLBACK_OFFSETS[index % FALLBACK_OFFSETS.length]

  return {
    latitude: center.latitude + offset.latitude,
    longitude: center.longitude + offset.longitude
  }
}

function extractNearbyList(data) {
  if (Array.isArray(data)) {
    return data
  }

  if (!data || typeof data !== 'object') {
    return []
  }

  return data.list || data.items || data.nearbyGames || data.games || []
}

function extractList(data, keys) {
  if (!data || typeof data !== 'object') {
    return []
  }

  for (let index = 0; index < keys.length; index += 1) {
    const value = data[keys[index]]

    if (Array.isArray(value)) {
      return value
    }
  }

  return []
}

function formatRadius(radiusMeters) {
  if (radiusMeters >= 1000) {
    const value = radiusMeters / 1000

    return `${Number.isInteger(value) ? value : value.toFixed(1)}km范围`
  }

  return `${radiusMeters}m范围`
}

function formatDistance(distance) {
  if (!Number.isFinite(distance)) {
    return '--'
  }

  if (distance >= 1000) {
    const value = distance / 1000

    return `${value >= 10 ? Math.round(value) : value.toFixed(1)}km`
  }

  return `${Math.max(1, Math.round(distance))}m`
}

function distanceMeters(pointA, pointB) {
  const toRadians = (value) => value * Math.PI / 180
  const earthRadius = 6371000
  const latitudeDelta = toRadians(pointB.latitude - pointA.latitude)
  const longitudeDelta = toRadians(pointB.longitude - pointA.longitude)
  const latitudeA = toRadians(pointA.latitude)
  const latitudeB = toRadians(pointB.latitude)
  const sinLatitude = Math.sin(latitudeDelta / 2)
  const sinLongitude = Math.sin(longitudeDelta / 2)
  const angle = 2 * Math.atan2(
    Math.sqrt(sinLatitude * sinLatitude + Math.cos(latitudeA) * Math.cos(latitudeB) * sinLongitude * sinLongitude),
    Math.sqrt(1 - (sinLatitude * sinLatitude + Math.cos(latitudeA) * Math.cos(latitudeB) * sinLongitude * sinLongitude))
  )

  return Math.round(earthRadius * angle)
}

function normalizeNearbyPlayers(data, center) {
  const onlinePlayers = extractList(data, ['onlinePlayers', 'onlineUsers', 'activePlayers'])
  const offlinePlayers = extractList(data, ['offlinePlayers', 'offlineUsers', 'inactivePlayers'])
  const hasPlayerData = onlinePlayers.length > 0 || offlinePlayers.length > 0

  if (!hasPlayerData) {
    return {
      items: [],
      onlineCount: 0,
      offlineCount: 0
    }
  }

  const rawPlayers = [
    ...onlinePlayers.map((item) => ({ ...item, status: 'online' })),
    ...offlinePlayers.map((item) => ({ ...item, status: 'offline' }))
  ]
  const items = rawPlayers.map((item, index) => {
    const coordinate = getCoordinate(item, center, index + FALLBACK_OFFSETS.length)
    const status = item.status === 'offline' ? 'offline' : 'online'
    const distanceText = item.distanceText || item.distance || formatDistance(distanceMeters(center, coordinate))

    return {
      id: item.id || `player-${index + 1}`,
      markerId: PLAYER_MARKER_BASE_ID + index + 1,
      status,
      name: item.name || item.nickname || item.title || `附近玩家 ${index + 1}`,
      statusText: item.statusText || (status === 'online' ? '在线玩家' : '离线玩家'),
      distanceText: String(distanceText),
      activeText: item.activeText || item.lastActiveText || (status === 'online' ? '正在附近' : '刚刚离线'),
      latitude: coordinate.latitude,
      longitude: coordinate.longitude,
      route: item.route || item.detailRoute || buildPlayerDetailRoute(item)
    }
  })
  const onlineCount = toNumber(data.onlinePlayerCount || data.onlineCount || data.activePlayerCount)
  const offlineCount = toNumber(data.offlinePlayerCount || data.offlineCount || data.inactivePlayerCount)

  return {
    items,
    onlineCount: onlineCount === null ? onlinePlayers.length : onlineCount,
    offlineCount: offlineCount === null ? offlinePlayers.length : offlineCount
  }
}

function buildMapLegendItems(groupCount, onlineCount, offlineCount) {
  return [
    { key: 'groups', label: `全部组局 ${groupCount}`, value: String(groupCount) },
    { key: 'location', label: '当前位置', value: '' },
    { key: 'online', label: `在线 ${onlineCount}`, value: String(onlineCount) },
    { key: 'offline', label: `离线 ${offlineCount}`, value: String(offlineCount) }
  ]
}

function isSameQueryArea(previousCenter, nextCenter, previousRadius, nextRadius) {
  if (!previousCenter || !nextCenter) {
    return false
  }

  return distanceMeters(previousCenter, nextCenter) < 80 && Math.abs(previousRadius - nextRadius) < 120
}

function normalizeNearbyItems(data, center) {
  const source = extractNearbyList(data)

  return source.map((item, index) => {
    const coordinate = getCoordinate(item, center, index)
    const markerId = index + 1
    const title = item.title || item.name || item.placeName || `附近组局 ${markerId}`
    const cityName = item.cityName || item.address || item.locationName || item.poiName || '当前位置附近'
    const distanceText = item.distanceText || item.distance || '附近'
    const memberText = item.memberText || item.membersText || item.capacityText || '待加入'
    const timeText = item.timeText || item.dateText || item.startTimeText || '时间待定'
    const statusText = item.statusText || item.typeText || '组局'
    const priceText = item.priceText || item.feeText || '¥0/人'

    return {
      id: item.id || `nearby-${markerId}`,
      markerId,
      title,
      addressText: cityName,
      distanceText: String(distanceText),
      memberText: String(memberText),
      timeText: String(timeText),
      statusText: String(statusText),
      priceText: String(priceText),
      latitude: coordinate.latitude,
      longitude: coordinate.longitude,
      route: item.route || item.detailRoute || buildGameDetailRoute(item.id)
    }
  })
}

function normalizeMapConfig(config) {
  const source = config || {}
  const fallback = FALLBACK_MAP_CONFIG
  const defaultLocation = source.defaultLocation || source.location || {}
  const latitude = toNumber(defaultLocation.latitude || defaultLocation.lat)
  const longitude = toNumber(defaultLocation.longitude || defaultLocation.lng)
  const radius = Number(source.defaultRadiusMeters || source.defaultRadius || 0)
  const radiusOptions = Array.isArray(source.radiusOptions)
    ? source.radiusOptions.map((item) => Number(item)).filter((item) => Number.isFinite(item) && item > 0)
    : []
  const mapFilters = Array.isArray(source.mapFilters)
    ? source.mapFilters.map((item) => String(item || '').trim()).filter(Boolean)
    : []

  return {
    onlineText: source.onlineText || fallback.onlineText,
    defaultLocation: {
      latitude: latitude === null ? fallback.defaultLocation.latitude : latitude,
      longitude: longitude === null ? fallback.defaultLocation.longitude : longitude
    },
    defaultRadiusMeters: radius > 0 ? radius : fallback.defaultRadiusMeters,
    radiusOptions: radiusOptions.length ? radiusOptions : fallback.radiusOptions,
    mapFilters: mapFilters.length ? mapFilters : fallback.mapFilters,
    texts: Object.assign({}, fallback.texts, source.texts || {})
  }
}

function buildGameDetailRoute(gameId) {
  if (!gameId) {
    return ''
  }

  return `pages/game/detail/index?id=${encodeURIComponent(gameId)}`
}

function appendRouteQuery(route, params) {
  if (!route) {
    return ''
  }

  const query = Object.keys(params || {})
    .filter((key) => params[key] !== undefined && params[key] !== null && params[key] !== '')
    .map((key) => `${encodeURIComponent(key)}=${encodeURIComponent(params[key])}`)
    .join('&')

  return query ? `${route}${route.indexOf('?') === -1 ? '?' : '&'}${query}` : route
}

function normalizeRoute(route) {
  const value = String(route || '').trim()

  if (!value) {
    return ''
  }

  return value.indexOf('/') === 0 ? value : `/${value}`
}

function routeWithoutQuery(route) {
  return String(route || '').split('?')[0].replace(/^\/+/, '')
}

function isConcreteGameDetailRoute(route) {
  const value = String(route || '')

  return routeWithoutQuery(value) === ROUTES.gameDetail && value.indexOf('?') !== -1
}

function buildPlayerDetailRoute(player) {
  const playerId = player && String(player.userId || player.memberId || player.id || '').trim()

  if (!playerId) {
    return ''
  }

  return `pages/profile/service-center/invite/member-detail/index?id=${encodeURIComponent(playerId)}`
}

function buildPlayerMarkers(items, selectedMarkerId = 0) {
  return items.map((item) => ({
    id: item.markerId,
    latitude: item.latitude,
    longitude: item.longitude,
    iconPath: item.status === 'online' ? ONLINE_PLAYER_ICON : OFFLINE_PLAYER_ICON,
    width: selectedMarkerId === item.markerId ? 30 : 24,
    height: selectedMarkerId === item.markerId ? 30 : 24,
    alpha: item.status === 'online' ? 0.98 : 0.36,
    callout: {
      content: item.name,
      color: item.status === 'online' ? '#065f46' : '#475569',
      fontSize: 11,
      borderRadius: 6,
      bgColor: item.status === 'online' ? '#dcfce7' : '#e2e8f0',
      padding: 6,
      display: selectedMarkerId === item.markerId ? 'ALWAYS' : 'BYCLICK'
    }
  }))
}

function buildMarkers(items, selectedMarkerId, playerItems = [], selectedPlayerMarkerId = 0) {
  const groupMarkers = items.map((item) => ({
    id: item.markerId,
    latitude: item.latitude,
    longitude: item.longitude,
    iconPath: MARKER_ICON,
    width: selectedMarkerId === item.markerId ? 38 : 32,
    height: selectedMarkerId === item.markerId ? 38 : 32,
    callout: {
      content: item.title,
      color: '#052e2b',
      fontSize: 13,
      borderRadius: 6,
      bgColor: '#d1fae5',
      padding: 8,
      display: selectedMarkerId === item.markerId ? 'ALWAYS' : 'BYCLICK'
    }
  }))

  return groupMarkers.concat(buildPlayerMarkers(playerItems, selectedPlayerMarkerId))
}

function buildCircles(center, radiusMeters) {
  return [
    {
      latitude: center.latitude,
      longitude: center.longitude,
      color: '#10b98166',
      fillColor: '#10b9811f',
      radius: radiusMeters,
      strokeWidth: 2
    }
  ]
}

Page({
  data: {
    onlineText: FALLBACK_MAP_CONFIG.onlineText,
    userId: '1769915682',
    latitude: FALLBACK_MAP_CONFIG.defaultLocation.latitude,
    longitude: FALLBACK_MAP_CONFIG.defaultLocation.longitude,
    scale: 15,
    radiusMeters: FALLBACK_MAP_CONFIG.defaultRadiusMeters,
    radiusButtonText: '3km',
    rangeText: formatRadius(FALLBACK_MAP_CONFIG.defaultRadiusMeters),
    locationStatusText: '正在定位',
    nearbyCountText: '附近信息点 0 个',
    mapLegendItems: buildMapLegendItems(0, 0, 0),
    activeFilterIndex: 0,
    mapFilters: FALLBACK_MAP_CONFIG.mapFilters,
    mapTexts: FALLBACK_MAP_CONFIG.texts,
    loadingNearby: false,
    nearbyItems: [],
    nearbyPlayers: [],
    markers: [],
    circles: buildCircles(FALLBACK_MAP_CONFIG.defaultLocation, FALLBACK_MAP_CONFIG.defaultRadiusMeters),
    selectedMarkerId: 0,
    selectedPlace: null,
    selectedPlayer: null,
    navItems: [
      { name: '我的', key: 'mine' },
      { name: '元宇宙', key: 'metaverse' },
      { name: '地图', key: 'map' },
      { name: '消息', key: 'message' },
      { name: '首页', key: 'home' }
    ]
  },

  async onLoad() {
    await this.loadMapIndexConfig()
    this.loadCurrentLocation()
  },

  async loadMapIndexConfig() {
    try {
      const config = normalizeMapConfig(await mapService.getIndexConfig())
      this.mapIndexConfig = config
      this.radiusOptions = config.radiusOptions
      this.locationCenter = config.defaultLocation
      this.userLocationCenter = config.defaultLocation

      this.setData({
        onlineText: config.onlineText,
        latitude: config.defaultLocation.latitude,
        longitude: config.defaultLocation.longitude,
        radiusMeters: config.defaultRadiusMeters,
        radiusButtonText: config.defaultRadiusMeters >= 1000 ? `${config.defaultRadiusMeters / 1000}km` : `${config.defaultRadiusMeters}m`,
        rangeText: formatRadius(config.defaultRadiusMeters),
        mapFilters: config.mapFilters,
        mapTexts: config.texts,
        circles: buildCircles(config.defaultLocation, config.defaultRadiusMeters)
      })
    } catch (error) {
      console.warn('get map index config failed', error)
      const config = normalizeMapConfig()
      this.mapIndexConfig = config
      this.radiusOptions = config.radiusOptions
      this.locationCenter = config.defaultLocation
      this.userLocationCenter = config.defaultLocation
    }
  },

  onReady() {
    this.mapContext = wx.createMapContext ? wx.createMapContext('nearbyMap', this) : null
  },

  loadCurrentLocation() {
    if (typeof wx === 'undefined' || typeof wx.getLocation !== 'function') {
      this.useFallbackLocation('当前环境无法定位')
      return
    }

    this.setData({
      locationStatusText: '正在定位'
    })

    wx.getLocation({
      type: 'gcj02',
      success: (result) => {
        const latitude = toNumber(result.latitude)
        const longitude = toNumber(result.longitude)

        if (latitude === null || longitude === null) {
          this.useFallbackLocation('定位结果异常')
          return
        }

        this.locationCenter = { latitude, longitude }
        this.userLocationCenter = this.locationCenter
        this.setMapCenter(this.locationCenter, this.data.radiusMeters, {
          locationStatusText: '已获取当前位置'
        })
        this.loadNearbyGames(this.locationCenter, this.data.radiusMeters)
      },
      fail: () => {
        this.useFallbackLocation('未授权定位，显示默认位置')
      }
    })
  },

  useFallbackLocation(statusText) {
    const fallbackLocation = (this.mapIndexConfig && this.mapIndexConfig.defaultLocation) || FALLBACK_MAP_CONFIG.defaultLocation
    this.locationCenter = fallbackLocation
    this.userLocationCenter = fallbackLocation
    this.setMapCenter(fallbackLocation, this.data.radiusMeters, {
      locationStatusText: statusText
    })
    this.loadNearbyGames(fallbackLocation, this.data.radiusMeters)
  },

  setMapCenter(center, radiusMeters, extraData = {}) {
    this.setData({
      latitude: center.latitude,
      longitude: center.longitude,
      radiusMeters,
      radiusButtonText: radiusMeters >= 1000 ? `${radiusMeters / 1000}km` : `${radiusMeters}m`,
      rangeText: formatRadius(radiusMeters),
      circles: buildCircles(center, radiusMeters),
      ...extraData
    })
  },

  async loadNearbyGames(center, radiusMeters) {
    this.setData({
      loadingNearby: true
    })

    try {
      const data = await locationService.getNearbyGames({
        latitude: center.latitude,
        longitude: center.longitude,
        radiusMeters,
        radius: radiusMeters,
        pageSize: 50
      })
      const nearbyItems = normalizeNearbyItems(data, center)
      const selectedGroupMarkerId = this.data.selectedMarkerId < PLAYER_MARKER_BASE_ID ? this.data.selectedMarkerId : 0
      const selectedPlayerMarkerId = this.data.selectedMarkerId >= PLAYER_MARKER_BASE_ID ? this.data.selectedMarkerId : 0
      const selectedPlace = selectedGroupMarkerId
        ? nearbyItems.find((item) => item.markerId === this.data.selectedMarkerId) || null
        : null
      const nearbyPlayers = normalizeNearbyPlayers(data, center)
      const selectedPlayer = selectedPlayerMarkerId
        ? nearbyPlayers.items.find((item) => item.markerId === this.data.selectedMarkerId) || null
        : null
      const responseGroupCount = toNumber(data.total)
      const responseNearbyGameCount = toNumber(data.nearbyGameCount)
      const responseNearbyCount = toNumber(data.nearbyCount)
      const nearbyGroupCount = responseGroupCount !== null
        ? responseGroupCount
        : responseNearbyGameCount !== null
          ? responseNearbyGameCount
          : responseNearbyCount !== null
            ? responseNearbyCount
            : nearbyItems.length

      this.setData({
        nearbyItems,
        nearbyPlayers: nearbyPlayers.items,
        markers: buildMarkers(
          nearbyItems,
          selectedPlace ? selectedPlace.markerId : 0,
          nearbyPlayers.items,
          selectedPlayer ? selectedPlayer.markerId : 0
        ),
        selectedPlace,
        selectedPlayer,
        nearbyCountText: `附近信息点 ${nearbyItems.length} 个`,
        mapLegendItems: buildMapLegendItems(nearbyGroupCount, nearbyPlayers.onlineCount, nearbyPlayers.offlineCount),
        loadingNearby: false
      })
    } catch (error) {
      const nearbyItems = []
      const nearbyPlayers = {
        items: [],
        onlineCount: 0,
        offlineCount: 0
      }

      this.setData({
        nearbyItems,
        nearbyPlayers: nearbyPlayers.items,
        markers: buildMarkers(nearbyItems, 0, nearbyPlayers.items),
        selectedMarkerId: 0,
        selectedPlace: null,
        selectedPlayer: null,
        nearbyCountText: `附近信息点 ${nearbyItems.length} 个`,
        mapLegendItems: buildMapLegendItems(nearbyItems.length, nearbyPlayers.onlineCount, nearbyPlayers.offlineCount),
        locationStatusText: '附近数据加载失败，请稍后重试',
        loadingNearby: false
      })
    }
  },

  handleMarkerTap(event) {
    const markerId = Number(event.markerId || event.detail && event.detail.markerId)
    this.suppressNextMapTapUntil = Date.now() + 450

    if (markerId >= PLAYER_MARKER_BASE_ID) {
      this.selectNearbyPlayer(markerId)
      return
    }

    this.selectNearbyPlace(markerId)
  },

  selectNearbyPlace(markerId) {
    const selectedPlace = this.data.nearbyItems.find((item) => item.markerId === markerId)

    if (!selectedPlace) {
      return null
    }

    this.setData({
      selectedMarkerId: markerId,
      selectedPlace,
      selectedPlayer: null,
      latitude: selectedPlace.latitude,
      longitude: selectedPlace.longitude,
      markers: buildMarkers(this.data.nearbyItems, markerId, this.data.nearbyPlayers)
    })

    return selectedPlace
  },

  selectNearbyPlayer(markerId) {
    const selectedPlayer = this.data.nearbyPlayers.find((item) => item.markerId === markerId)

    if (!selectedPlayer) {
      return null
    }

    this.setData({
      selectedMarkerId: markerId,
      selectedPlace: null,
      selectedPlayer,
      latitude: selectedPlayer.latitude,
      longitude: selectedPlayer.longitude,
      markers: buildMarkers(this.data.nearbyItems, 0, this.data.nearbyPlayers, markerId)
    })

    return selectedPlayer
  },

  handleMapTap() {
    if (this.suppressNextMapTapUntil && Date.now() < this.suppressNextMapTapUntil) {
      return
    }

    if (!this.data.selectedPlace && !this.data.selectedPlayer) {
      return
    }

    this.setData({
      selectedMarkerId: 0,
      selectedPlace: null,
      selectedPlayer: null,
      markers: buildMarkers(this.data.nearbyItems, 0, this.data.nearbyPlayers)
    })
  },

  handleCloseInfoTap() {
    this.suppressNextMapTapUntil = 0
    this.setData({
      selectedMarkerId: 0,
      selectedPlace: null,
      selectedPlayer: null,
      markers: buildMarkers(this.data.nearbyItems, 0, this.data.nearbyPlayers)
    })
  },

  handleRegionChange(event) {
    if (event.type !== 'end') {
      return
    }

    const causedBy = event.causedBy || (event.detail && event.detail.causedBy)

    if (causedBy !== 'drag' && causedBy !== 'scale') {
      return
    }

    clearTimeout(this.regionChangeTimer)
    this.regionChangeTimer = setTimeout(() => {
      this.refreshByMapRegion()
    }, 500)
  },

  refreshByMapRegion() {
    if (!this.mapContext || typeof this.mapContext.getRegion !== 'function') {
      this.loadNearbyGames(this.locationCenter || this.mapIndexConfig.defaultLocation, this.data.radiusMeters)
      return
    }

    this.mapContext.getRegion({
      success: (region) => {
        const southwest = region.southwest || {}
        const northeast = region.northeast || {}
        const latitude = (Number(southwest.latitude) + Number(northeast.latitude)) / 2
        const longitude = (Number(southwest.longitude) + Number(northeast.longitude)) / 2

        if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) {
          this.loadNearbyGames(this.locationCenter || this.mapIndexConfig.defaultLocation, this.data.radiusMeters)
          return
        }

        const center = { latitude, longitude }
        const viewportRadius = distanceMeters(center, {
          latitude: Number(northeast.latitude),
          longitude: Number(northeast.longitude)
        })
        const radiusMeters = Math.max(800, Math.min(10000, viewportRadius || this.data.radiusMeters))

        if (isSameQueryArea(this.locationCenter, center, this.data.radiusMeters, radiusMeters)) {
          return
        }

        this.locationCenter = center
        this.setMapCenter(center, radiusMeters, {
          locationStatusText: '已按地图视野更新'
        })
        this.loadNearbyGames(center, radiusMeters)
      },
      fail: () => {
        this.loadNearbyGames(this.locationCenter || this.mapIndexConfig.defaultLocation, this.data.radiusMeters)
      }
    })
  },

  handleMapToolTap(event) {
    const action = event.currentTarget.dataset.action

    if (action === 'locate') {
      this.loadCurrentLocation()
      return
    }

    if (action === 'refresh') {
      this.refreshByMapRegion()
      return
    }

    if (action === 'radius') {
      const radiusOptions = this.radiusOptions || FALLBACK_MAP_CONFIG.radiusOptions
      const currentIndex = radiusOptions.indexOf(this.data.radiusMeters)
      const nextRadius = radiusOptions[(currentIndex + 1) % radiusOptions.length]
      const center = this.locationCenter || {
        latitude: this.data.latitude,
        longitude: this.data.longitude
      }

      this.setMapCenter(center, nextRadius, {
        locationStatusText: '已切换查询范围'
      })
      this.loadNearbyGames(center, nextRadius)
    }
  },

  handleDetailTap() {
    const route = this.data.selectedPlace && this.data.selectedPlace.route
    const targetRoute = isConcreteGameDetailRoute(route) ? route : ROUTES.gameHall

    navigateShellRoute(targetRoute, {
      currentRoute: ROUTES.map
    })
  },

  handlePlayerDetailTap() {
    const route = this.data.selectedPlayer && this.data.selectedPlayer.route
    const targetRoute = route || ROUTES.profile

    navigateShellRoute(targetRoute, {
      currentRoute: ROUTES.map
    })
  },

  handleMapFilterTap(event) {
    const index = Number(event.currentTarget.dataset.index)

    if (!Number.isNaN(index)) {
      this.setData({
        activeFilterIndex: index
      })
    }
  },

  handleNearbyItemTap(event) {
    const markerId = Number(event.currentTarget.dataset.markerId)

    if (!Number.isNaN(markerId)) {
      this.selectNearbyPlace(markerId)
    }
  },

  handleNearbyActionTap(event) {
    const markerId = Number(event.currentTarget.dataset.markerId)
    const action = event.currentTarget.dataset.action
    const selectedPlace = this.selectNearbyPlace(markerId)

    if (!selectedPlace) {
      return
    }

    if (action === 'detail') {
      navigateShellRoute(isConcreteGameDetailRoute(selectedPlace.route) ? selectedPlace.route : ROUTES.gameHall, {
        currentRoute: ROUTES.map
      })
      return
    }

    const route = appendRouteQuery(
      isConcreteGameDetailRoute(selectedPlace.route) ? selectedPlace.route : ROUTES.gameHall,
      {
      intent: 'join',
      from: 'nearbyMap'
      }
    )

    if (route) {
      navigateShellRoute(route, {
        currentRoute: ROUTES.map
      })
      return
    }
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}

    navigateShellKey(key, {
      currentRoute: ROUTES.map
    })
  }
})
