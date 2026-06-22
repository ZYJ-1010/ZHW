const { ROUTES } = require('../../../config/routes')
const toast = require('../../../utils/toast')

const AUDIT_SCROLL_TAP_STEP_RPX = 360
const AUDIT_SCROLL_HOLD_STEP_RPX = 72
const AUDIT_SCROLL_HOLD_INTERVAL_MS = 80
const AUDIT_SCROLL_HOLD_SUPPRESS_TAP_MS = 120

const applications = [
  {
    id: 'audit-001',
    initiator: {
      avatarText: '强',
      nickname: '赛博导游阿强',
      roleName: '行家'
    },
    roleKey: 'expert',
    applyTime: '2026-03-30 10:45',
    statusKey: 'pending',
    statusText: '待审核'
  },
  {
    id: 'audit-002',
    initiator: {
      avatarText: '夏',
      nickname: '极客少女小夏',
      roleName: '玩家'
    },
    roleKey: 'player',
    applyTime: '2026-03-30 11:20',
    statusKey: 'pending',
    statusText: '待审核'
  },
  {
    id: 'audit-003',
    initiator: {
      avatarText: 'G',
      nickname: '老顽童G',
      roleName: '领路人'
    },
    roleKey: 'guide',
    applyTime: '2026-03-29 18:30',
    statusKey: 'approved',
    statusText: '已通过'
  },
  {
    id: 'audit-004',
    initiator: {
      avatarText: '林',
      nickname: '城市玩家小林',
      roleName: '玩家'
    },
    roleKey: 'player',
    applyTime: '2026-03-29 16:10',
    statusKey: 'rejected',
    statusText: '已拒绝'
  }
]

function getDisplayApplications(activeFilter) {
  if (activeFilter === 'all') {
    return applications
  }

  return applications.filter((item) => item.statusKey === activeFilter)
}

Page({
  data: {
    onlineText: '3999人在线',
    auditScrollTop: 0,
    activeFilter: 'all',
    allSelected: false,
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    filters: [
      { key: 'all', name: '全部' },
      { key: 'pending', name: '待审核' },
      { key: 'approved', name: '已通过' },
      { key: 'rejected', name: '已拒绝' }
    ],
    displayApplications: applications
  },

  handleFilterTap(event) {
    const key = event.currentTarget.dataset.key || 'all'

    this.setData({
      activeFilter: key,
      displayApplications: getDisplayApplications(key)
    })
  },

  handleSelectAllTap() {
    this.setData({
      allSelected: !this.data.allSelected
    })
  },

  handleApproveTap() {
    toast.info('通过申请功能开发中')
  },

  handleRejectTap() {
    toast.info('拒绝申请功能开发中')
  },

  handleDetailTap(event) {
    const id = event.currentTarget.dataset.id || ''
    const query = id ? `?auditId=${id}` : ''

    wx.navigateTo({
      url: `/${ROUTES.gameAuditDetail}${query}`
    })
  },

  handleBatchApproveTap() {
    toast.info('批量通过功能开发中')
  },

  handleBatchRejectTap() {
    toast.info('批量拒绝功能开发中')
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (key === 'up' || key === 'down') {
      if (!this.suppressNextNavTap) {
        this.scrollAudit(key, AUDIT_SCROLL_TAP_STEP_RPX)
      }
      return
    }

    if (key === 'left' || key === 'right') {
      toast.info('功能正在开发中')
      return
    }

    this.handleShellAction(key)
  },

  handleShellNavLongPress(event) {
    const key = event.detail && event.detail.key

    if (key !== 'up' && key !== 'down') {
      return
    }

    this.suppressNextNavTap = true
    this.stopAuditScrollHold(false)
    this.scrollAudit(key, AUDIT_SCROLL_HOLD_STEP_RPX)

    this.auditScrollHoldTimer = setInterval(() => {
      this.scrollAudit(key, AUDIT_SCROLL_HOLD_STEP_RPX)
    }, AUDIT_SCROLL_HOLD_INTERVAL_MS)
  },

  handleShellNavTouchEnd() {
    this.stopAuditScrollHold(true)
  },

  handleShellAction(key) {
    if (key === 'home') {
      this.scrollAuditToTop()
      return
    }

    if (key === 'avatar' || key === 'mine') {
      this.navigateToRoute(ROUTES.profile)
      return
    }

    if (key === 'message') {
      this.navigateToRoute(ROUTES.message)
      return
    }

    const routeMap = {
      metaverse: ROUTES.metaverse,
      map: ROUTES.map
    }

    this.navigateToRoute(routeMap[key])
  },

  navigateToRoute(route) {
    if (!route || route === ROUTES.gameAudit) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  },

  handleAuditScroll(event) {
    const scrollTop = event.detail && event.detail.scrollTop

    if (typeof scrollTop === 'number') {
      this.auditScrollTopValue = scrollTop
    }
  },

  scrollAudit(direction, stepRpx = AUDIT_SCROLL_TAP_STEP_RPX) {
    const current = Number(this.auditScrollTopValue || this.data.auditScrollTop || 0)
    const distance = this.rpxToPx(stepRpx)
    const nextTop = direction === 'up'
      ? Math.max(0, current - distance)
      : current + distance

    this.auditScrollTopValue = nextTop
    this.setData({
      auditScrollTop: nextTop
    })
  },

  scrollAuditToTop() {
    this.auditScrollTopValue = 0
    this.setData({
      auditScrollTop: 0
    })
  },

  stopAuditScrollHold(resetTapSuppress) {
    if (this.auditScrollHoldTimer) {
      clearInterval(this.auditScrollHoldTimer)
      this.auditScrollHoldTimer = null
    }

    if (resetTapSuppress && this.suppressNextNavTap) {
      if (this.auditScrollSuppressTimer) {
        clearTimeout(this.auditScrollSuppressTimer)
      }

      this.auditScrollSuppressTimer = setTimeout(() => {
        this.suppressNextNavTap = false
        this.auditScrollSuppressTimer = null
      }, AUDIT_SCROLL_HOLD_SUPPRESS_TAP_MS)
    }
  },

  clearAuditScrollTimers() {
    this.stopAuditScrollHold(false)

    if (this.auditScrollSuppressTimer) {
      clearTimeout(this.auditScrollSuppressTimer)
      this.auditScrollSuppressTimer = null
    }

    this.suppressNextNavTap = false
  },

  rpxToPx(value) {
    if (!wx.getSystemInfoSync) {
      return value / 2
    }

    const system = wx.getSystemInfoSync()
    const windowWidth = system && system.windowWidth ? system.windowWidth : 375

    return Math.round((value * windowWidth) / 750)
  },

  onUnload() {
    this.clearAuditScrollTimers()
  }
})
