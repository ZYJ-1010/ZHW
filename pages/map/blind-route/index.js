const { ROUTES } = require('../../../config/routes')

const BLIND_ROUTE_CARDS = [
  {
    id: 'tonight',
    title: '今晚去哪局',
    desc: '随机生成今晚的社交路线',
    tone: 'blue',
    icon: '/pages/map/blind-route/assets/i50.png',
    tags: ['2-4人', '3小时']
  },
  {
    id: 'couple',
    title: '情侣半日局',
    desc: '浪漫约会专属组局',
    tone: 'orange',
    icon: '/pages/map/blind-route/assets/i52.png',
    tags: ['2人', '浪漫']
  },
  {
    id: 'explore',
    title: '组队探索局',
    desc: '一场城市漫游局',
    tone: 'purple',
    icon: '/pages/map/blind-route/assets/i54.png',
    tags: ['3人', '治愈']
  },
  {
    id: 'social',
    title: '3人轻社交局',
    desc: '轻松认识新朋友',
    tone: 'green',
    icon: '/pages/map/blind-route/assets/i49.png',
    tags: ['3人', '社交']
  },
  {
    id: 'startup',
    title: '创业人脑暴局',
    desc: '灵感碰撞路线',
    tone: 'pink',
    icon: '/pages/map/blind-route/assets/i51.png',
    tags: ['2-5人', '创业']
  }
]
const RECENT_ROUTES = [
  { title: '今晚去哪局', timeText: '昨天 18:30', statusText: '已完成' },
  { title: '组队探索局', timeText: '周二 20:15', statusText: '已点亮' }
]

function markCards(selectedId) {
  return BLIND_ROUTE_CARDS.map((item) => ({
    ...item,
    selected: item.id === selectedId,
    className: `blind-route-card tone-${item.tone}${item.id === selectedId ? ' selected' : ''}`
  }))
}

Page({
  data: {
    onlineText: '3999人在线',
    mode: 'blindRout',
    navItems: [
      { name: '我的', key: 'mine' },
      { name: '元宇宙', key: 'metaverse' },
      { name: '地图', key: 'map' },
      { name: '消息', key: 'message' },
      { name: '首页', key: 'home' }
    ],
    routeCards: markCards(BLIND_ROUTE_CARDS[0].id),
    recentRoutes: RECENT_ROUTES
  },

  onLoad(options = {}) {
    this.setData({
      mode: options.mode || 'blindRout'
    })
  },

  handleRouteTap(event) {
    const id = event.currentTarget.dataset.id
    const hasRoute = BLIND_ROUTE_CARDS.some((item) => item.id === id)
    const selectedId = hasRoute ? id : BLIND_ROUTE_CARDS[0].id

    this.setData({
      routeCards: markCards(selectedId)
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

    if (!route || route === ROUTES.mapBlindRoute) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  }
})
