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
    roles: [
      { name: '玩家', active: true },
      { name: '行家', active: false },
      { name: '领路人', active: false }
    ],
    apartmentWindows: [1, 2, 3, 4, 5, 6, 7, 8, 9],
    actions: [
      { title: '获取新地块', desc: '解锁新的数字空间坐标', tone: 'blue' },
      { title: '开拓街区路线', desc: '创建可探索的城市路线', tone: 'pink' },
      { title: '资产管理中心', desc: '管理空间、地块和资产', tone: 'green' },
      { title: '商城', desc: '浏览装扮与扩展资产', tone: 'orange' }
    ]
  },

  handleActionTap() {
    wx.showToast({
      title: '功能正在开发中',
      icon: 'none'
    })
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}
    navigateShellKey(key, {
      currentRoute: 'pages/metaverse/manage/index'
    })
  }
})
