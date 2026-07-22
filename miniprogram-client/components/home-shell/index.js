const DPAD_SCROLL_STEP_RPX = 360
const { navigateShellBack, navigateShellForward, navigateShellKey } = require('../../utils/shell-nav')
const profileService = require('../../services/profile')
const toast = require('../../utils/toast')
const { getHomeShellFixedFrameLayout } = require('./layout')
const DEFAULT_ONLINE_COUNT = '0'
const FLOATING_TASK_POSITION_STORAGE_KEY = 'enjoy_home_task_float_position_v1'
const FLOATING_TASK_WIDTH_RPX = 112
const FLOATING_TASK_HEIGHT_RPX = 112
const FLOATING_TASK_MARGIN_RPX = 24
const FLOATING_TASK_DRAG_THRESHOLD_RPX = 6

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
    floatingTaskStyle: '',
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
    },
    floatingTaskVisible: {
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
      this.restoreFloatingTaskPosition()
      this.loadShellProfile()
    },
    ready() {
      this.updateShellLayout()
    },
    detached() {
      if (this.taskFloatTapSuppressTimer) {
        clearTimeout(this.taskFloatTapSuppressTimer)
        this.taskFloatTapSuppressTimer = null
      }
    }
  },

  pageLifetimes: {
    show() {
      this.updateShellLayout()
      this.loadShellProfile()
    },
    resize() {
      this.updateShellLayout()
      this.restoreFloatingTaskPosition()
    }
  },

  methods: {
    getFloatingTaskBounds() {
      const layout = getHomeShellFixedFrameLayout({
        dockVisible: this.properties.dockVisible !== false,
        contentBottomGapRpx: Math.max(0, Number(this.properties.contentBottomGap || 0))
      })
      const canvasHeight = Number(layout && layout.canvasHeightRpx) || 1626
      const contentTop = Number(layout && layout.contentTopRpx) || 182
      const dockTop = Number(layout && layout.dockTopRpx) || (canvasHeight - 218)
      const minX = FLOATING_TASK_MARGIN_RPX
      const maxX = Math.max(minX, 750 - FLOATING_TASK_WIDTH_RPX - FLOATING_TASK_MARGIN_RPX)
      const minY = contentTop + FLOATING_TASK_MARGIN_RPX
      const maxY = Math.max(minY, dockTop - FLOATING_TASK_HEIGHT_RPX - FLOATING_TASK_MARGIN_RPX)

      return { minX, maxX, minY, maxY }
    },

    clampFloatingTaskPosition(position = {}) {
      const bounds = this.getFloatingTaskBounds()
      const x = Number(position.x)
      const y = Number(position.y)
      const fallback = { x: bounds.maxX, y: bounds.maxY }

      return {
        x: Math.round(Math.min(bounds.maxX, Math.max(bounds.minX, Number.isFinite(x) ? x : fallback.x))),
        y: Math.round(Math.min(bounds.maxY, Math.max(bounds.minY, Number.isFinite(y) ? y : fallback.y)))
      }
    },

    floatingTaskStyleFor(position = {}) {
      return `left: ${position.x}rpx; top: ${position.y}rpx; right: auto; bottom: auto;`
    },

    restoreFloatingTaskPosition() {
      let savedPosition = null
      try {
        savedPosition = wx.getStorageSync(FLOATING_TASK_POSITION_STORAGE_KEY)
      } catch (error) {
        savedPosition = null
      }

      const position = this.clampFloatingTaskPosition(savedPosition || this.floatingTaskPosition || {})
      this.floatingTaskPosition = position
      this.setData({
        floatingTaskStyle: this.floatingTaskStyleFor(position)
      })
    },

    taskTouchPoint(event) {
      const touch = (event && event.touches && event.touches[0]) || (event && event.changedTouches && event.changedTouches[0])
      if (!touch) {
        return null
      }

      const windowInfo = this.getWindowInfo() || {}
      const windowWidth = Number(windowInfo.windowWidth || windowInfo.screenWidth || 0)
      if (!Number.isFinite(windowWidth) || windowWidth <= 0) {
        return null
      }

      return {
        x: Number(touch.clientX) * 750 / windowWidth,
        y: Number(touch.clientY) * 750 / windowWidth
      }
    },

    handleTaskTouchStart(event) {
      const point = this.taskTouchPoint(event)
      if (!point) {
        return
      }

      if (!this.floatingTaskPosition) {
        this.restoreFloatingTaskPosition()
      }

      this.taskFloatDidDrag = false
      this.taskFloatDragState = {
        startX: point.x,
        startY: point.y,
        originX: this.floatingTaskPosition.x,
        originY: this.floatingTaskPosition.y
      }
    },

    handleTaskTouchMove(event) {
      const point = this.taskTouchPoint(event)
      const drag = this.taskFloatDragState
      if (!point || !drag) {
        return
      }

      const nextPosition = this.clampFloatingTaskPosition({
        x: drag.originX + point.x - drag.startX,
        y: drag.originY + point.y - drag.startY
      })
      if (Math.abs(nextPosition.x - drag.originX) >= FLOATING_TASK_DRAG_THRESHOLD_RPX || Math.abs(nextPosition.y - drag.originY) >= FLOATING_TASK_DRAG_THRESHOLD_RPX) {
        this.taskFloatDidDrag = true
      }

      this.floatingTaskPosition = nextPosition
      this.setData({
        floatingTaskStyle: this.floatingTaskStyleFor(nextPosition)
      })
    },

    handleTaskTouchEnd() {
      if (!this.taskFloatDragState) {
        return
      }

      this.taskFloatDragState = null
      if (!this.taskFloatDidDrag) {
        return
      }

      try {
        wx.setStorageSync(FLOATING_TASK_POSITION_STORAGE_KEY, this.floatingTaskPosition)
      } catch (error) {
        // 位置保存失败不影响本次拖动和按钮使用。
      }

      if (this.taskFloatTapSuppressTimer) {
        clearTimeout(this.taskFloatTapSuppressTimer)
      }
      this.taskFloatTapSuppressTimer = setTimeout(() => {
        this.taskFloatDidDrag = false
        this.taskFloatTapSuppressTimer = null
      }, 250)
    },

    handleTaskTap() {
      if (this.taskFloatDidDrag) {
        return
      }
      this.triggerEvent('tasktap')
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

      // 地图与元宇宙属于二期能力。一期所有底部导航页统一只提示，
      // 不再把点击事件交给各页面，避免部分页面仍跳转到预留页面。
      if (key === 'map' || key === 'metaverse') {
        toast.info(key === 'map' ? '地图玩法暂未开放' : '元宇宙玩法暂未开放')
        return
      }

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
