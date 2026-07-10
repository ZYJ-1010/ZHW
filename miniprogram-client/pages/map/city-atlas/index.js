const { ROUTES } = require('../../../config/routes')
const mapService = require('../../../services/map')
const toast = require('../../../utils/toast')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')

const EMPTY_PAGE = {
  title: '',
  progressLabel: '',
  lockedText: '',
  hiddenBadge: '',
  routeSectionTitle: '',
  filters: [],
  unlockInfo: { unlockedCount: 0, totalCount: 0, progressPercent: 0 },
  unlockToast: '{name}',
  lockedToast: '{name}',
  unlockPoints: [],
  themeRoutes: []
}

function applyTemplate(template, values = {}) {
  let text = String(template || '')
  Object.keys(values).forEach((key) => {
    text = text.replace(new RegExp(`\\{${key}\\}`, 'g'), String(values[key]))
  })
  return text
}

function filterUnlockPoints(points, filter) {
  if (filter === '已解锁') {
    return points.filter((item) => item.statusType === 'unlocked')
  }
  if (filter === '未解锁') {
    return points.filter((item) => item.statusType === 'locked')
  }
  if (filter === '隐藏点') {
    return points.filter((item) => item.statusType === 'hidden')
  }
  return points
}

Page({
  data: {
    onlineText: '在线',
    navItems: [
      { name: '我的', key: 'mine' },
      { name: '元宇宙', key: 'metaverse' },
      { name: '地图', key: 'map' },
      { name: '消息', key: 'message' },
      { name: '首页', key: 'home' }
    ],
    pageConfig: EMPTY_PAGE,
    filters: [],
    activeFilter: '',
    unlockInfo: EMPTY_PAGE.unlockInfo,
    unlockPoints: [],
    themeRoutes: []
  },

  onLoad() {
    this.loadPageConfig()
  },

  async loadPageConfig() {
    try {
      const pageConfig = Object.assign({}, EMPTY_PAGE, await mapService.getPlayPage('city-atlas'))
      const filters = Array.isArray(pageConfig.filters) ? pageConfig.filters : []
      const activeFilter = filters[0] || ''
      this.setData({
        pageConfig,
        filters,
        activeFilter,
        unlockInfo: pageConfig.unlockInfo || EMPTY_PAGE.unlockInfo,
        unlockPoints: filterUnlockPoints(pageConfig.unlockPoints || [], activeFilter),
        themeRoutes: pageConfig.themeRoutes || []
      })
    } catch (error) {
      toast.info(error.message || '城市图鉴加载失败')
    }
  },

  handleBackTap() {
    wx.navigateBack({
      delta: 1,
      fail: () => {
        navigateShellRoute(ROUTES.map, {
          currentRoute: ROUTES.mapCityAtlas
        })
      }
    })
  },

  handleFilterTap(event) {
    const filter = event.currentTarget.dataset.filter || this.data.filters[0] || ''

    this.setData({
      activeFilter: filter,
      unlockPoints: filterUnlockPoints(this.data.pageConfig.unlockPoints || [], filter)
    })
  },

  handleUnlockPointTap(event) {
    const id = event.currentTarget.dataset.id
    const point = (this.data.pageConfig.unlockPoints || []).find((item) => item.id === id)

    if (!point) {
      return
    }

    const template = point.statusType === 'unlocked'
      ? this.data.pageConfig.unlockToast
      : this.data.pageConfig.lockedToast
    const content = [
      applyTemplate(template, { name: point.name }),
      point.unlockText,
      point.footprintValue
    ].filter(Boolean).join('\n')
    const confirmText = point.statusType === 'unlocked'
      ? '查看足迹'
      : (point.statusType === 'hidden' ? '去探索' : '去打卡')

    wx.showModal({
      title: point.name || '城市点位',
      content,
      confirmText,
      cancelText: '关闭',
      success: (res) => {
        if (!res.confirm) {
          return
        }

        const route = point.statusType === 'unlocked'
          ? ROUTES.mapMyCity
          : (point.statusType === 'hidden' ? ROUTES.mapBlindRoute : ROUTES.mapRealCheckin)
        navigateShellRoute(`${route}?pointId=${encodeURIComponent(point.id)}`, {
          currentRoute: ROUTES.mapCityAtlas
        })
      }
    })
  },

  handleThemeRouteTap(event) {
    const route = (this.data.themeRoutes || []).find((item) => item.id === event.currentTarget.dataset.id)

    navigateShellRoute(`${ROUTES.mapBlindRoute}?theme=${encodeURIComponent(route ? route.id : '')}`, {
      currentRoute: ROUTES.mapCityAtlas
    })
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}
    navigateShellKey(key, {
      currentRoute: ROUTES.mapCityAtlas
    })
  }
})
