Component({
  options: {
    multipleSlots: true,
    addGlobalClass: true,
    styleIsolation: 'shared'
  },

  data: {
    shellNavItems: [
      { name: '我的', key: 'mine' },
      { name: '元宇宙', key: 'metaverse' },
      { name: '地图', key: 'map' },
      { name: '消息', key: 'message' },
      { name: '首页', key: 'home' }
    ]
  },

  properties: {
    onlineText: {
      type: String,
      value: ''
    },
    navItems: {
      type: Array,
      value: [],
      observer(value) {
        this.setData({
          shellNavItems: this.mergeNavItems(value)
        })
      }
    },
    toolbarStyle: {
      type: String,
      value: ''
    },
    canvasStyle: {
      type: String,
      value: ''
    },
    contentStyle: {
      type: String,
      value: ''
    },
    dockStyle: {
      type: String,
      value: ''
    },
    shellClass: {
      type: String,
      value: ''
    },
    navToastEnabled: {
      type: Boolean,
      value: true
    }
  },

  lifetimes: {
    attached() {
      this.setData({
        shellNavItems: this.mergeNavItems(this.properties.navItems)
      })
    }
  },

  methods: {
    mergeNavItems(navItems) {
      const defaults = [
        { name: '我的', key: 'mine' },
        { name: '元宇宙', key: 'metaverse' },
        { name: '地图', key: 'map' },
        { name: '消息', key: 'message' },
        { name: '首页', key: 'home' }
      ]

      if (!Array.isArray(navItems) || navItems.length === 0) {
        return defaults
      }

      return defaults.map((item, index) => ({
        ...item,
        ...(navItems[index] || {})
      }))
    },

    handleNavTap(event) {
      const { key } = event.currentTarget.dataset

      if (this.properties.navToastEnabled) {
        wx.showToast({
          title: '功能正在开发中',
          icon: 'none'
        })
      }

      this.triggerEvent('navtap', { key })
    },

    handleNavLongPress(event) {
      const { key } = event.currentTarget.dataset

      this.triggerEvent('navlongpress', { key })
    },

    handleNavTouchEnd(event) {
      const { key } = event.currentTarget.dataset

      this.triggerEvent('navtouchend', { key })
    }
  }
})
