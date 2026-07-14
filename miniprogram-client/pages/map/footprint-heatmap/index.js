const { ROUTES } = require('../../../config/routes')
const mapService = require('../../../services/map')
const toast = require('../../../utils/toast')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')

const NAV_ITEMS = [
  { name: '我的', key: 'mine' },
  { name: '元宇宙', key: 'metaverse' },
  { name: '地图', key: 'map' },
  { name: '消息', key: 'message' },
  { name: '首页', key: 'home' }
]

const EMPTY_PAGE = {
  title: '',
  heatTitle: '',
  legendLabel: '',
  friendTitle: '',
  friendMoreText: '',
  hotTitle: '',
  rangeTabs: [],
  rangeStats: {},
  cityHeatPoints: [],
  friendUpdates: [],
  hotCities: []
}

Page({
  data: {
    onlineText: '在线',
    navItems: NAV_ITEMS,
    pageConfig: EMPTY_PAGE,
    rangeTabs: [],
    activeRange: '',
    heatStats: [],
    cityHeatPoints: [],
    friendUpdates: [],
    hotCities: []
  },

  onLoad() {
    this.loadPageConfig()
  },

  async loadPageConfig() {
    try {
      const pageConfig = Object.assign({}, EMPTY_PAGE, await mapService.getPlayPage('footprint-heatmap'))
      const rangeTabs = Array.isArray(pageConfig.rangeTabs) ? pageConfig.rangeTabs : []
      const activeRange = rangeTabs[0] || ''
      const rangeStats = pageConfig.rangeStats || {}
      this.setData({
        pageConfig,
        rangeTabs,
        activeRange,
        heatStats: rangeStats[activeRange] || [],
        cityHeatPoints: pageConfig.cityHeatPoints || [],
        friendUpdates: pageConfig.friendUpdates || [],
        hotCities: pageConfig.hotCities || []
      })
    } catch (error) {
      toast.info(error.message || '足迹热力加载失败')
    }
  },

  handleBackTap() {
    wx.navigateBack({
      delta: 1,
      fail: () => {
        navigateShellRoute(ROUTES.map, {
          currentRoute: ROUTES.mapFootprintHeatmap
        })
      }
    })
  },

  handleRangeTap(event) {
    const range = event.currentTarget.dataset.range || this.data.rangeTabs[0] || ''
    const rangeStats = this.data.pageConfig.rangeStats || {}

    this.setData({
      activeRange: range,
      heatStats: rangeStats[range] || []
    })
  },

  handleCityTap(event) {
    const city = (this.data.cityHeatPoints || []).find((item) => item.id === event.currentTarget.dataset.id)

    if (!city) {
      return
    }

    wx.showModal({
      title: `${city.city}足迹报告`,
      content: [
        `时间范围：${this.data.activeRange || '全部'}`,
        `热力等级：${city.level || '未分级'}`,
        '可进入地图查看附近局，也可在我的城市页继续查看足迹故事。'
      ].join('\n'),
      confirmText: '查看地图',
      cancelText: '关闭',
      success: (res) => {
        if (!res.confirm) {
          return
        }
        navigateShellRoute(`${ROUTES.map}?city=${encodeURIComponent(city.city || '')}`, {
          currentRoute: ROUTES.mapFootprintHeatmap
        })
      }
    })
  },

  handleFriendTap() {
    navigateShellRoute(ROUTES.relationNetwork || ROUTES.profile, {
      currentRoute: ROUTES.mapFootprintHeatmap
    })
  },

  handleHotCityTap(event) {
    const city = (this.data.hotCities || []).find((item) => item.id === event.currentTarget.dataset.id)

    if (!city) {
      return
    }

    wx.showModal({
      title: `${city.city}城市热榜`,
      content: [
        city.desc,
        `热度进度：${city.progress || 0}%`,
        '进入地图后可查看该城市附近局和点位。'
      ].filter(Boolean).join('\n'),
      confirmText: '查看地图',
      cancelText: '关闭',
      success: (res) => {
        if (!res.confirm) {
          return
        }

        navigateShellRoute(`${ROUTES.map}?city=${encodeURIComponent(city.city || '')}`, {
          currentRoute: ROUTES.mapFootprintHeatmap
        })
      }
    })
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}
    navigateShellKey(key, {
      currentRoute: ROUTES.mapFootprintHeatmap
    })
  }
})
