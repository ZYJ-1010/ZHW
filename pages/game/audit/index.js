const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const toast = require('../../../utils/toast')
const { getSurnameInitials } = require('../../../utils/avatar')

const AUDIT_SCROLL_TAP_STEP_RPX = 360
const AUDIT_SCROLL_HOLD_STEP_RPX = 72
const AUDIT_SCROLL_HOLD_INTERVAL_MS = 80
const AUDIT_SCROLL_HOLD_SUPPRESS_TAP_MS = 120

function normalizeApplication(item = {}) {
  const initiator = item.initiator || item.applicant || item.user || {}
  const nickname = initiator.nickname || initiator.name || item.nickname || ''

  return {
    id: item.id || item.auditId || item.applicationId || '',
    initiator: {
      nickname,
      avatarText: initiator.avatarText || getSurnameInitials(nickname, ''),
      roleName: initiator.roleName || initiator.roleText || item.roleName || ''
    },
    roleKey: item.roleKey || item.roleType || '',
    applyTime: item.applyTime || item.applyTimeText || item.createdAtText || item.createdAt || '',
    statusKey: item.statusKey || item.status || '',
    statusText: item.statusText || ''
  }
}

function normalizeAuditList(data = {}) {
  const list = Array.isArray(data.list || data.records || data.items)
    ? (data.list || data.records || data.items).map(normalizeApplication).filter((item) => item.id)
    : []

  return {
    onlineText: data.onlineText || '3999人在线',
    applications: list,
    displayApplications: list
  }
}

Page({
  data: {
    onlineText: '3999人在线',
    auditScrollTop: 0,
    activeFilter: 'all',
    allSelected: false,
    loading: false,
    actionLoading: false,
    applications: [],
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
    displayApplications: []
  },

  onLoad(options = {}) {
    this.loadAudits(options)
  },

  async loadAudits(extraParams = {}) {
    this.setData({
      loading: true
    })

    try {
      const data = await gameService.getGameAudits({
        ...extraParams,
        status: this.data.activeFilter === 'all' ? '' : this.data.activeFilter
      })

      this.setData({
        ...normalizeAuditList(data),
        loading: false,
        allSelected: false
      })
    } catch (error) {
      this.setData({
        ...normalizeAuditList({}),
        loading: false,
        allSelected: false
      })
      toast.info(error.message || '审核申请加载失败')
    }
  },

  handleFilterTap(event) {
    const key = event.currentTarget.dataset.key || 'all'

    this.setData({
      activeFilter: key
    })
    this.loadAudits()
  },

  handleSelectAllTap() {
    this.setData({
      allSelected: !this.data.allSelected
    })
  },

  handleApproveTap(event) {
    this.respondAudit(event.currentTarget.dataset.id, 'approve')
  },

  handleRejectTap(event) {
    this.respondAudit(event.currentTarget.dataset.id, 'reject')
  },

  async respondAudit(auditId, action) {
    if (this.data.actionLoading) {
      return
    }

    this.setData({
      actionLoading: true
    })

    try {
      await gameService.respondGameAudit({
        auditId,
        action
      })
      toast.success('已提交审核结果')
      this.loadAudits()
    } catch (error) {
      toast.info(error.message || '审核处理失败')
    } finally {
      this.setData({
        actionLoading: false
      })
    }
  },

  handleDetailTap(event) {
    const id = event.currentTarget.dataset.id || ''
    const query = id ? `?auditId=${id}` : ''

    wx.navigateTo({
      url: `/${ROUTES.gameAuditDetail}${query}`
    })
  },

  handleBatchApproveTap() {
    this.batchRespondAudits('approve')
  },

  handleBatchRejectTap() {
    this.batchRespondAudits('reject')
  },

  async batchRespondAudits(action) {
    if (this.data.actionLoading) {
      return
    }

    const auditIds = this.data.allSelected
      ? this.data.displayApplications.map((item) => item.id).filter(Boolean)
      : []

    this.setData({
      actionLoading: true
    })

    try {
      await gameService.batchRespondGameAudits({
        auditIds,
        action
      })
      toast.success('已提交批量审核结果')
      this.loadAudits()
    } catch (error) {
      toast.info(error.message || '批量审核处理失败')
    } finally {
      this.setData({
        actionLoading: false
      })
    }
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (key === 'map') {
      wx.showToast({
        title: '地图功能开发中',
        icon: 'none'
      })
      return
    }

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

    if (key === 'map') {
      wx.showToast({
        title: '地图功能开发中',
        icon: 'none'
      })
      return
    }

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
      map: ''
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
