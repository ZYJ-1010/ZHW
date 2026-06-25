const { ROUTES } = require('../../../config/routes')

const FILTERS = ['全部', '已解锁', '未解锁', '隐藏点']
const UNLOCK_POINTS = [
  {
    id: 'bund-night',
    name: '外滩夜景',
    statusType: 'unlocked',
    unlockText: '2024.01.15 解锁',
    footprintValue: '+50 足迹值',
    tone: 'blue',
    iconType: 'building',
    checked: true
  },
  {
    id: 'tianzifang',
    name: '田子坊',
    statusType: 'unlocked',
    unlockText: '2024.02.03 解锁',
    footprintValue: '+30 足迹值',
    tone: 'green',
    iconType: 'lantern',
    checked: true
  },
  {
    id: 'wukang-road',
    name: '武康路街角',
    statusType: 'locked',
    unlockText: '完成 2 次附近打卡后解锁',
    footprintValue: '+40 足迹值',
    tone: 'locked',
    iconType: 'lock',
    checked: false
  },
  {
    id: 'hidden-rooftop',
    name: '城市天台',
    statusType: 'hidden',
    unlockText: '隐藏点待发现',
    footprintValue: '+80 足迹值',
    tone: 'purple',
    iconType: 'hidden',
    checked: false
  }
]
const THEME_ROUTES = [
  {
    id: 'couple-walk',
    name: '情侣漫步',
    meta: '6个地点 · 预计3小',
    progressText: '已解锁 2/6',
    tone: 'sunset',
    iconText: '💕'
  },
  {
    id: 'coffee-shop',
    name: '咖啡探店',
    meta: '8个地点 · 预计4小',
    progressText: '已解锁 0/8',
    tone: 'cyan',
    iconText: '☕'
  }
]

function filterUnlockPoints(filter) {
  if (filter === '已解锁') {
    return UNLOCK_POINTS.filter((item) => item.statusType === 'unlocked')
  }

  if (filter === '未解锁') {
    return UNLOCK_POINTS.filter((item) => item.statusType === 'locked')
  }

  if (filter === '隐藏点') {
    return UNLOCK_POINTS.filter((item) => item.statusType === 'hidden')
  }

  return UNLOCK_POINTS
}

Page({
  data: {
    onlineText: '3999人在线',
    navItems: [
      { name: '我的', key: 'mine' },
      { name: '元宇宙', key: 'metaverse' },
      { name: '地图', key: 'map' },
      { name: '消息', key: 'message' },
      { name: '首页', key: 'home' }
    ],
    filters: FILTERS,
    activeFilter: FILTERS[0],
    unlockInfo: {
      unlockedCount: 1,
      totalCount: 36,
      progressPercent: 33
    },
    unlockPoints: filterUnlockPoints(FILTERS[0]),
    themeRoutes: THEME_ROUTES
  },

  handleBackTap() {
    wx.navigateBack({
      delta: 1,
      fail: () => {
        wx.navigateTo({
          url: `/${ROUTES.map}`
        })
      }
    })
  },

  handleFilterTap(event) {
    const filter = event.currentTarget.dataset.filter || FILTERS[0]

    this.setData({
      activeFilter: filter,
      unlockPoints: filterUnlockPoints(filter)
    })
  },

  handleUnlockPointTap(event) {
    const id = event.currentTarget.dataset.id
    const point = UNLOCK_POINTS.find((item) => item.id === id)

    if (!point) {
      return
    }

    wx.showToast({
      title: point.statusType === 'unlocked' ? `${point.name}已解锁` : `${point.name}待解锁`,
      icon: 'none'
    })
  },

  handleThemeRouteTap(event) {
    const route = THEME_ROUTES.find((item) => item.id === event.currentTarget.dataset.id)

    wx.showToast({
      title: route ? `${route.name}待接入` : '主题路线待接入',
      icon: 'none'
    })
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}
    const routeMap = {
      home: ROUTES.playerHome,
      map: ROUTES.map,
      message: ROUTES.message,
      mine: ROUTES.profile,
      metaverse: ROUTES.metaverse
    }
    const route = routeMap[key]

    if (!route || route === ROUTES.mapCityAtlas) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  }
})
