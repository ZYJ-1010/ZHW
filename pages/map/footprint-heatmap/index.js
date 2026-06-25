const { ROUTES } = require('../../../config/routes')

const NAV_ITEMS = [
  { name: '我的', key: 'mine' },
  { name: '元宇宙', key: 'metaverse' },
  { name: '地图', key: 'map' },
  { name: '消息', key: 'message' },
  { name: '首页', key: 'home' }
]

const RANGE_TABS = ['今日', '本周', '本月', '全部']
const RANGE_STATS = {
  今日: [
    { value: '12', label: '打卡城市' },
    { value: '1.8k', label: '玩家足迹' },
    { value: '3', label: '热门城市' }
  ],
  本周: [
    { value: '38', label: '打卡城市' },
    { value: '8.5k', label: '玩家足迹' },
    { value: '9', label: '热门城市' }
  ],
  本月: [
    { value: '76', label: '打卡城市' },
    { value: '26k', label: '玩家足迹' },
    { value: '18', label: '热门城市' }
  ],
  全部: [
    { value: '126', label: '打卡城市' },
    { value: '92k', label: '玩家足迹' },
    { value: '31', label: '热门城市' }
  ]
}

const CITY_HEAT_POINTS = [
  { id: 'beijing', city: '北京', level: 'mid', className: 'footprint-city-point beijing level-mid' },
  { id: 'shanghai', city: '上海', level: 'hot', className: 'footprint-city-point shanghai level-hot' },
  { id: 'chengdu', city: '成都', level: 'hot', className: 'footprint-city-point chengdu level-hot' },
  { id: 'guangzhou', city: '广州', level: 'mid', className: 'footprint-city-point guangzhou level-mid' },
  { id: 'shenzhen', city: '深圳', level: 'high', className: 'footprint-city-point shenzhen level-high' },
  { id: 'xian', city: '西安', level: 'low', className: 'footprint-city-point xian level-low' },
  { id: 'hangzhou', city: '杭州', level: 'high', className: 'footprint-city-point hangzhou level-high' }
]

const FRIEND_UPDATES = [
  {
    id: 'alex',
    avatarText: 'AL',
    name: 'Alex',
    desc: '刚刚在成都宽窄巷子打卡',
    online: true
  },
  {
    id: 'sarah',
    avatarText: 'SA',
    name: 'Sarah',
    desc: '25分钟前在西安城墙打卡',
    online: false
  }
]

const HOT_CITIES = [
  { id: 'shanghai', rank: 1, city: '上海市中心', desc: '2456人在这里打卡', progress: 86, level: 'hot' },
  { id: 'chengdu', rank: 2, city: '成都市', desc: '1892人在这里打卡', progress: 72, level: 'warm' },
  { id: 'shenzhen', rank: 3, city: '深圳湾', desc: '1567人在这里打卡', progress: 58, level: 'active' }
]

Page({
  data: {
    onlineText: '3999人在线',
    navItems: NAV_ITEMS,
    rangeTabs: RANGE_TABS,
    activeRange: RANGE_TABS[0],
    heatStats: RANGE_STATS[RANGE_TABS[0]],
    cityHeatPoints: CITY_HEAT_POINTS,
    friendUpdates: FRIEND_UPDATES,
    hotCities: HOT_CITIES
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

  handleRangeTap(event) {
    const range = event.currentTarget.dataset.range || RANGE_TABS[0]

    this.setData({
      activeRange: range,
      heatStats: RANGE_STATS[range] || RANGE_STATS[RANGE_TABS[0]]
    })
  },

  handleCityTap(event) {
    const city = CITY_HEAT_POINTS.find((item) => item.id === event.currentTarget.dataset.id)

    wx.showToast({
      title: city ? `${city.city}打卡热力待接入` : '城市热力待接入',
      icon: 'none'
    })
  },

  handleFriendTap() {
    wx.showToast({
      title: '好友城市动态待接入',
      icon: 'none'
    })
  },

  handleHotCityTap(event) {
    const city = HOT_CITIES.find((item) => item.id === event.currentTarget.dataset.id)

    wx.showToast({
      title: city ? `${city.city}详情待接入` : '城市详情待接入',
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

    if (!route || route === ROUTES.mapFootprintHeatmap) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  }
})
