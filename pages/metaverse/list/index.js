Page({
  data: {
    onlineText: '3999人在线',
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: true },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: false }
    ],
    leftWindows: [1, 2, 3, 4, 5],
    rightWindows: [1, 2, 3, 4, 5, 6],
    towerWindows: [1, 2, 3, 4, 5, 6, 7, 8],
    players: [
      { name: '喵小七', avatarText: '喵', tone: 'pink', active: true },
      { name: '陆家嘴车神', avatarText: '陆', tone: 'cyan', active: false },
      { name: '老猫 (我)', avatarText: '老', tone: 'orange', active: true }
    ]
  },

  handleActionTap() {
    wx.showToast({
      title: '功能正在开发中',
      icon: 'none'
    })
  },

  handleShellNavTap() {}
})
