const { ROUTES } = require('../../../../config/routes')

const NAV_ITEMS = [
  { name: '我的', key: 'mine' },
  { name: '元宇宙', key: 'metaverse' },
  { name: '地图', key: 'map' },
  { name: '消息', key: 'message' },
  { name: '首页', key: 'home' }
]

const FILTERS = [
  { key: 'all', label: '全部' },
  { key: 'city', label: '点亮城市' },
  { key: 'streak', label: '连续打卡' },
  { key: 'hidden', label: '隐藏成就' }
]

const ACHIEVED = [
  {
    id: 'city-pioneer',
    title: '城市先锋',
    desc: '点亮首个城市',
    icon: '/pages/profile/footprint/achievements/assets/icon-city.png',
    tone: 'gold',
    category: 'city'
  },
  {
    id: 'seven-days',
    title: '连续7天',
    desc: '坚持打卡',
    icon: '/pages/profile/footprint/achievements/assets/icon-footprint.png',
    tone: 'cyan',
    category: 'streak'
  },
  {
    id: 'social-rookie',
    title: '社交新手',
    desc: '参与10个局',
    icon: '/pages/profile/footprint/achievements/assets/icon-group.png',
    tone: 'purple',
    category: 'city'
  },
  {
    id: 'nature-walk',
    title: '自然探索',
    desc: '发现5个公园',
    icon: '/pages/profile/footprint/achievements/assets/icon-shop.png',
    tone: 'green',
    category: 'city'
  },
  {
    id: 'romantic-route',
    title: '浪漫足迹',
    desc: '情侣路线完成',
    icon: '/pages/profile/footprint/achievements/assets/icon-star.png',
    tone: 'pink',
    category: 'city'
  },
  {
    id: 'night-walker',
    title: '夜行者',
    desc: '3次夜间打卡',
    icon: '/pages/profile/footprint/achievements/assets/icon-points.png',
    tone: 'indigo',
    category: 'streak'
  }
]

const LOCKED = [
  {
    id: 'world-traveler',
    title: '环球旅行家',
    desc: '点亮10个城市',
    icon: '/pages/profile/footprint/achievements/assets/icon-cycle.png',
    tone: 'locked',
    category: 'hidden'
  },
  {
    id: 'city-champion',
    title: '城市冠军',
    desc: '排行榜第一',
    icon: '/pages/profile/footprint/achievements/assets/icon-crown.png',
    tone: 'locked',
    category: 'hidden'
  },
  {
    id: 'mystery-finder',
    title: '神秘发现者',
    desc: '找到隐藏点',
    icon: '/pages/profile/footprint/achievements/assets/icon-star.png',
    tone: 'locked',
    category: 'hidden'
  }
]

function filterAchievements(list, filterKey) {
  if (filterKey === 'all') {
    return list
  }

  return list.filter((item) => item.category === filterKey)
}

Page({
  data: {
    onlineText: '3999人在线',
    navItems: NAV_ITEMS,
    activeFilter: 'all',
    filters: FILTERS,
    levelTitle: '探索行家 Lv.5',
    levelTip: '再获得 150 点升级',
    progress: 70,
    achievedCount: 12,
    lockedCount: 8,
    achieved: ACHIEVED,
    locked: LOCKED,
    season: {
      title: '春季赛季',
      status: '进行中',
      remain: '本赛季剩余 15 天',
      achieved: 3,
      locked: 5
    }
  },

  handleFilterTap(event) {
    const { key } = event.currentTarget.dataset

    this.setData({
      activeFilter: key,
      achieved: filterAchievements(ACHIEVED, key),
      locked: filterAchievements(LOCKED, key)
    })
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}

    if (key === 'map') {
      wx.showToast({
        title: '地图功能开发中',
        icon: 'none'
      })
      return
    }
    const routeMap = {
      home: ROUTES.playerHome,
      map: '',
      message: ROUTES.message,
      mine: ROUTES.profile,
      metaverse: ROUTES.metaverse
    }
    const route = routeMap[key]

    if (!route) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  }
})
