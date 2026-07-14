const { ROUTES } = require('../../../config/routes')
const { navigateShellKey } = require('../../../utils/shell-nav')

Page({
  data: {
    onlineText: '在线0人',
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: true },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: false }
    ],
    entries: [
      {
        key: 'avatar',
        title: '进入虚拟形象工坊',
        tone: 'cyan',
        icon: 'avatar'
      },
      {
        key: 'space',
        title: '我的元宇宙空间',
        tone: 'violet',
        icon: 'home'
      },
      {
        key: 'hall',
        title: '全球玩家大厅',
        tone: 'green',
        icon: 'globe'
      }
    ]
  },

  handleEntryTap() {
    wx.showToast({
      title: '功能正在开发中',
      icon: 'none'
    })
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}

    navigateShellKey(key, {
      currentRoute: ROUTES.metaverse
    })
  }
})
