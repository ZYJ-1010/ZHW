const { ROUTES } = require('../../../config/routes')

const NAV_ITEMS = [
  { name: '我的', key: 'mine' },
  { name: '元宇宙', key: 'metaverse' },
  { name: '地图', key: 'map' },
  { name: '消息', key: 'message' },
  { name: '首页', key: 'home' }
]

const CHECKIN_DETAIL = {
  distanceText: '距离目标 15米',
  spotName: '外滩观景台',
  statusTitle: '地点已解锁',
  statusDesc: '完成打卡任务获得足迹值',
  storyTitle: '留下你的故事',
  storyPlaceholder: '用20个字记录此刻的心情...',
  storyMinLength: 20,
  storyMaxLength: 120,
  rewards: ['+20 足迹值', '+1 成就点']
}

const CHECKIN_TASKS = [
  {
    id: 'photo',
    title: '拍摄地标合影',
    desc: '与标志性建筑合影',
    scoreText: '+10分'
  },
  {
    id: 'angle',
    title: '发现隐藏角度',
    desc: '拍摄独特的视角',
    scoreText: '+20分'
  }
]

Page({
  data: {
    onlineText: '3999人在线',
    navItems: NAV_ITEMS,
    checkinDetail: CHECKIN_DETAIL,
    checkinTasks: CHECKIN_TASKS,
    storyText: '',
    storyCountText: `0/${CHECKIN_DETAIL.storyMaxLength}`
  },

  handleCloseTap() {
    wx.navigateBack({
      delta: 1,
      fail: () => {
        wx.navigateTo({
          url: `/${ROUTES.map}`
        })
      }
    })
  },

  handleShootTap() {
    wx.showToast({
      title: '拍照打卡待接入',
      icon: 'none'
    })
  },

  handleStoryTap() {
    wx.showToast({
      title: '打卡文字待接入',
      icon: 'none'
    })
  },

  handleStoryInput(event) {
    const value = event.detail.value || ''

    this.setData({
      storyText: value,
      storyCountText: `${value.length}/${CHECKIN_DETAIL.storyMaxLength}`
    })
  },

  handleTaskTap(event) {
    const task = CHECKIN_TASKS.find((item) => item.id === event.currentTarget.dataset.id)

    wx.showToast({
      title: task ? `${task.title}待接入` : '打卡任务待接入',
      icon: 'none'
    })
  },

  handleFragmentTap() {
    wx.showToast({
      title: '足迹碎片待生成',
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

    if (!route || route === ROUTES.mapRealCheckin) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  }
})
