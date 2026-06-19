const DEFAULT_CONTENT_TOP_RPX = 181
const DEFAULT_DOCK_TOP_RPX = 1408
const DEFAULT_TOPBAR_HEIGHT_RPX = 182
const TOOLBAR_HEIGHT_RPX = 58
const TOOLBAR_CAPSULE_GAP_RPX = 14
const NAV_BOTTOM_GAP_RPX = 13
const BRAND_ONLINE_BOTTOM_OFFSET_RPX = 82

Component({
  options: {
    multipleSlots: true,
    addGlobalClass: true,
    styleIsolation: 'shared'
  },

  data: {
    shellCanvasStyle: '',
    shellToolbarStyle: '',
    shellTopbarStyle: '',
    shellStatusFillStyle: '',
    shellNavStyle: '',
    shellBrandStyle: '',
    shellContentStyle: '',
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
      value: '',
      observer() {
        this.updateShellLayout()
      }
    },
    canvasStyle: {
      type: String,
      value: '',
      observer() {
        this.updateShellLayout()
      }
    },
    contentStyle: {
      type: String,
      value: '',
      observer() {
        this.updateShellLayout()
      }
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
      this.updateShellLayout()
    },
    ready() {
      this.updateShellLayout()
    }
  },

  pageLifetimes: {
    show() {
      this.updateShellLayout()
    },
    resize() {
      this.updateShellLayout()
    }
  },

  methods: {
    roundRpx(value) {
      return Math.round(value * 100) / 100
    },

    normalizeStyle(style) {
      const value = style || ''

      if (!value || value.trim().endsWith(';')) {
        return value
      }

      return `${value};`
    },

    getWindowInfo() {
      if (wx.getWindowInfo) {
        return wx.getWindowInfo()
      }

      if (wx.getSystemInfoSync) {
        return wx.getSystemInfoSync()
      }

      return null
    },

    updateShellLayout() {
      const canvasStyle = this.properties.canvasStyle || ''
      const toolbarStyle = this.properties.toolbarStyle || ''
      const contentStyle = this.properties.contentStyle || ''

      if (!wx.getMenuButtonBoundingClientRect) {
        this.setData({
          shellCanvasStyle: canvasStyle,
          shellToolbarStyle: toolbarStyle,
          shellTopbarStyle: '',
          shellStatusFillStyle: '',
          shellNavStyle: '',
          shellBrandStyle: '',
          shellContentStyle: contentStyle
        })
        return
      }

      const menuButton = wx.getMenuButtonBoundingClientRect()
      const windowInfo = this.getWindowInfo()

      if (!menuButton || !menuButton.width || !menuButton.height || !windowInfo || !windowInfo.windowWidth) {
        this.setData({
          shellCanvasStyle: canvasStyle,
          shellToolbarStyle: toolbarStyle,
          shellTopbarStyle: '',
          shellStatusFillStyle: '',
          shellNavStyle: '',
          shellBrandStyle: '',
          shellContentStyle: contentStyle
        })
        return
      }

      const ratio = 750 / windowInfo.windowWidth
      const statusBarHeight = Number(windowInfo.statusBarHeight) || Math.max(0, menuButton.top - 4)
      const capsuleTopGap = Math.max(0, menuButton.top - statusBarHeight)
      const navHeightPx = capsuleTopGap * 2 + menuButton.height
      const navBottomRpx = this.roundRpx((statusBarHeight + navHeightPx) * ratio)
      const statusFillHeight = this.roundRpx(statusBarHeight * ratio)
      const navHeight = this.roundRpx(navHeightPx * ratio)
      const contentTop = Math.max(DEFAULT_CONTENT_TOP_RPX, this.roundRpx(navBottomRpx + NAV_BOTTOM_GAP_RPX))
      const topbarHeight = Math.max(DEFAULT_TOPBAR_HEIGHT_RPX, this.roundRpx(contentTop + 1))
      const contentHeight = Math.max(0, this.roundRpx(DEFAULT_DOCK_TOP_RPX - contentTop))
      const toolbarTop = this.roundRpx((menuButton.top + menuButton.height / 2) * ratio - TOOLBAR_HEIGHT_RPX / 2)
      const toolbarRight = this.roundRpx((windowInfo.windowWidth - menuButton.left) * ratio + TOOLBAR_CAPSULE_GAP_RPX)
      const capsuleBottom = this.roundRpx((menuButton.top + menuButton.height) * ratio)
      const brandTop = this.roundRpx(capsuleBottom - statusFillHeight - BRAND_ONLINE_BOTTOM_OFFSET_RPX)
      const normalizedCanvasStyle = this.normalizeStyle(canvasStyle)
      const normalizedToolbarStyle = this.normalizeStyle(toolbarStyle)
      const normalizedContentStyle = this.normalizeStyle(contentStyle)

      this.setData({
        shellCanvasStyle: normalizedCanvasStyle,
        shellToolbarStyle: `${normalizedToolbarStyle} top: ${toolbarTop}rpx; right: ${toolbarRight}rpx;`,
        shellTopbarStyle: `height: ${topbarHeight}rpx;`,
        shellStatusFillStyle: `height: ${statusFillHeight}rpx;`,
        shellNavStyle: `height: ${navHeight}rpx;`,
        shellBrandStyle: `top: ${brandTop}rpx;`,
        shellContentStyle: `${normalizedContentStyle} top: ${contentTop}rpx; height: ${contentHeight}rpx;`
      })
    },

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
