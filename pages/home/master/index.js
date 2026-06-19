Page({
  data: {
    masterMode: 'default',
    brand: '真好玩',
    onlineText: '3999人在线',
    dockVisible: true,
    contentScrollY: false,
    shellClass: '',
    applyMaster: {
      caption: '申请页母版',
      navTitle: '标题'
    },
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ]
  },

  onLoad(options = {}) {
    if (options.mode === 'apply') {
      this.setData({
        masterMode: 'apply',
        dockVisible: false,
        contentScrollY: false,
        shellClass: ''
      })
      return
    }

    if (options.mode === 'contentOnly') {
      this.setData({
        masterMode: 'contentOnly',
        dockVisible: false,
        contentScrollY: true,
        shellClass: 'home-shell--content-only'
      })
    }
  },

  handleShellNavTap(event) {
    const { key } = event.detail

    if (!key) {
      return
    }
  }
})
