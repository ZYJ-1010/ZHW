Page({
  data: {
    brand: '真好玩',
    onlineText: '3999人在线',
    toolbarStyle: '',
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ]
  },

  onLoad() {
    this.alignToolbarToCapsule()
  },

  handleShellNavTap(event) {
    const { key } = event.detail

    if (!key) {
      return
    }
  },

  alignToolbarToCapsule() {
    if (!wx.getMenuButtonBoundingClientRect || !wx.getSystemInfoSync) {
      return
    }

    const menuButton = wx.getMenuButtonBoundingClientRect()
    const system = wx.getSystemInfoSync()
    const ratio = 750 / system.windowWidth
    const iconCenterOffset = 29
    const capsuleCenterTop = (menuButton.top + menuButton.height / 2) * ratio
    const toolbarTop = capsuleCenterTop - iconCenterOffset
    const toolbarRight = (system.windowWidth - menuButton.left + 10) * ratio

    this.setData({
      toolbarStyle: `top: ${toolbarTop}rpx; right: ${toolbarRight}rpx;`
    })
  }
})
