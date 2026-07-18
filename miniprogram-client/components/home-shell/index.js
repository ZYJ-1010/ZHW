const DPAD_SCROLL_STEP_RPX = 360
const { navigateShellBack, navigateShellForward, navigateShellKey } = require('../../utils/shell-nav')
const profileService = require('../../services/profile')
const toast = require('../../utils/toast')
const { getHomeShellFixedFrameLayout } = require('./layout')
const DEFAULT_ONLINE_COUNT = '0'

function formatOnlineText(value) {
  const match = String(value || '').match(/\d[\d,]*/)
  const count = match ? match[0].replace(/,/g, '') : DEFAULT_ONLINE_COUNT

  return `在线${count}人`
}

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
    shellPageTitleStyle: '',
    shellContentStyle: '',
    shellDockStyle: '',
    shellContentScrollTop: 0,
    resolvedTopbarTitle: '',
    resolvedOnlineText: formatOnlineText(''),
    resolvedBrandVisible: true,
    resolvedToolbarActionsVisible: true,
    resolvedToolbarAvatarVisible: true,
    shellAvatarUrl: '',
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
      value: '',
      observer(value) {
        this.setData({
          resolvedOnlineText: formatOnlineText(value)
        })
      }
    },
    topbarTitle: {
      type: String,
      value: '',
      observer() {
        this.updateShellVariant()
      }
    },
    topbarTitleVisible: {
      type: Boolean,
      value: true,
      observer() {
        this.updateShellVariant()
      }
    },
    shellVariant: {
      type: String,
      value: '',
      observer() {
        this.updateShellVariant()
      }
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
    contentBottomGap: {
      type: Number,
      value: 0,
      observer() {
        this.updateShellLayout()
      }
    },
    dockStyle: {
      type: String,
      value: '',
      observer() {
        this.updateShellLayout()
      }
    },
    shellClass: {
      type: String,
      value: ''
    },
    navToastEnabled: {
      type: Boolean,
      value: true
    },
    toolbarActionsVisible: {
      type: Boolean,
      value: true,
      observer() {
        this.updateShellVariant()
      }
    },
    toolbarAvatarVisible: {
      type: Boolean,
      value: true,
      observer() {
        this.updateShellVariant()
      }
    },
    dockVisible: {
      type: Boolean,
      value: true,
      observer() {
        this.updateShellLayout()
      }
    },
    contentScrollY: {
      type: Boolean,
      value: false
    },
    navAutoNavigate: {
      type: Boolean,
      value: false
    }
  },

  lifetimes: {
    attached() {
      this.setData({
        shellNavItems: this.mergeNavItems(this.properties.navItems)
      })
      this.updateShellVariant()
      this.updateShellLayout()
      this.loadShellProfile()
    },
    ready() {
      this.updateShellLayout()
    }
  },

  pageLifetimes: {
    show() {
      this.updateShellLayout()
      this.loadShellProfile()
    },
    resize() {
      this.updateShellLayout()
    }
  },

  methods: {
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

    updateShellVariant() {
      const isJoinApply = this.properties.shellVariant === 'joinApply'
      const isTopNoBrand = this.properties.shellVariant === 'topNoBrand'
      const titleVisible = this.properties.topbarTitleVisible !== false
      const toolbarAvatarVisible = this.properties.toolbarAvatarVisible !== false

      this.setData({
        resolvedTopbarTitle: titleVisible ? (this.properties.topbarTitle || (isJoinApply || isTopNoBrand ? '申请加入' : '')) : '',
        resolvedBrandVisible: !isTopNoBrand,
        resolvedToolbarActionsVisible: (isJoinApply || isTopNoBrand) ? false : this.properties.toolbarActionsVisible !== false,
        resolvedToolbarAvatarVisible: toolbarAvatarVisible && !isTopNoBrand
      })
    },

    async loadShellProfile() {
      try {
        const data = await profileService.getProfileHome()
        const user = data && data.user ? data.user : {}
        const avatarUrl = String(user.avatarUrl || '').trim()

        if (avatarUrl !== this.data.shellAvatarUrl) {
          this.setData({
            shellAvatarUrl: avatarUrl
          })
        }
      } catch (error) {
        if (this.data.shellAvatarUrl) {
          this.setData({
            shellAvatarUrl: ''
          })
        }
      }
    },

    updateShellLayout() {
      const canvasStyle = this.properties.canvasStyle || ''
      const toolbarStyle = this.properties.toolbarStyle || ''
      const contentStyle = this.properties.contentStyle || ''
      const dockStyle = this.properties.dockStyle || ''
      const dockVisible = this.properties.dockVisible !== false
      const contentBottomGap = Math.max(0, Number(this.properties.contentBottomGap || 0))

      const layout = getHomeShellFixedFrameLayout({
        dockVisible,
        contentBottomGapRpx: contentBottomGap
      })

      if (!layout) {
        this.setData({
          shellCanvasStyle: canvasStyle,
          shellToolbarStyle: toolbarStyle,
          shellTopbarStyle: '',
          shellStatusFillStyle: '',
          shellNavStyle: '',
          shellBrandStyle: '',
          shellPageTitleStyle: '',
          shellContentStyle: contentStyle,
          shellDockStyle: dockVisible ? dockStyle : `${this.normalizeStyle(dockStyle)} display: none;`
        })
        return
      }

      const normalizedCanvasStyle = this.normalizeStyle(canvasStyle)
      const normalizedToolbarStyle = this.normalizeStyle(toolbarStyle)
      const normalizedContentStyle = this.normalizeStyle(contentStyle)
      const normalizedDockStyle = this.normalizeStyle(dockStyle)

      this.setData({
        shellCanvasStyle: `${normalizedCanvasStyle} height: ${layout.canvasHeightRpx}rpx; min-height: ${layout.canvasHeightRpx}rpx;`,
        shellToolbarStyle: `${normalizedToolbarStyle} top: ${layout.toolbarTopRpx}rpx; right: ${layout.toolbarRightRpx}rpx;`,
        shellTopbarStyle: `height: ${layout.topbarHeightRpx}rpx;`,
        shellStatusFillStyle: `height: ${layout.statusBarHeightRpx}rpx;`,
        shellNavStyle: `height: ${layout.navigationHeightRpx}rpx;`,
        shellBrandStyle: `top: ${layout.brandTopRpx}rpx;`,
        shellPageTitleStyle: `top: ${layout.pageTitleTopRpx}rpx;`,
        shellContentStyle: `${normalizedContentStyle} top: ${layout.contentTopRpx}rpx; height: ${layout.contentHeightRpx}rpx;`,
        shellDockStyle: dockVisible
          ? `${normalizedDockStyle} top: ${layout.dockTopRpx}rpx; height: ${layout.dockHeightRpx}rpx;`
          : `${normalizedDockStyle} top: ${layout.dockTopRpx}rpx; height: 0; display: none;`
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

      if (key === 'up' || key === 'down') {
        this.triggerEvent('navtap', { key })
        if (this.properties.navAutoNavigate) {
          this.scrollShellBy(key)
        }
        return
      }

      if (key === 'left') {
        this.triggerEvent('navtap', { key })
        if (this.properties.navAutoNavigate) {
          navigateShellBack()
        }
        return
      }

      if (key === 'right') {
        this.triggerEvent('navtap', { key })
        if (this.properties.navAutoNavigate) {
          navigateShellForward()
        }
        return
      }

      this.triggerEvent('navtap', { key })

      if (key === 'map') {
        toast.info('地图玩法暂未开放')
        return
      }

      if (this.properties.navAutoNavigate) {
        if (navigateShellKey(key, {
          onSameRoute: (routeKey) => {
            if (routeKey === 'home') {
              this.scrollShellToTop()
            }
          }
        })) {
          return
        }
      }
    },

    handleContentScroll(event) {
      const scrollTop = event.detail && event.detail.scrollTop

      if (typeof scrollTop === 'number') {
        this.shellContentScrollTopValue = scrollTop
      }
    },

    scrollShellBy(direction) {
      if (!this.properties.contentScrollY) {
        return
      }

      const current = Number(this.shellContentScrollTopValue || this.data.shellContentScrollTop || 0)
      const next = direction === 'up'
        ? Math.max(0, current - this.rpxToPx(DPAD_SCROLL_STEP_RPX))
        : current + this.rpxToPx(DPAD_SCROLL_STEP_RPX)

      this.shellContentScrollTopValue = next
      this.setData({ shellContentScrollTop: next })
    },

    scrollShellToTop() {
      this.shellContentScrollTopValue = 0
      this.setData({
        shellContentScrollTop: 0
      })
    },

    rpxToPx(value) {
      const windowInfo = this.getWindowInfo()
      const width = windowInfo && windowInfo.windowWidth ? windowInfo.windowWidth : 375

      return Math.round((Number(value) || 0) * width / 750)
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
