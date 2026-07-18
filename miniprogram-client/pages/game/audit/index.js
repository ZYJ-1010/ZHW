const { ROUTES } = require('../../../config/routes')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')
const gameService = require('../../../services/game')
const toast = require('../../../utils/toast')
const { getSurnameInitials } = require('../../../utils/avatar')

const AUDIT_SCROLL_TAP_STEP_RPX = 360
const AUDIT_SCROLL_HOLD_STEP_RPX = 72
const AUDIT_SCROLL_HOLD_INTERVAL_MS = 80
const AUDIT_SCROLL_HOLD_SUPPRESS_TAP_MS = 120

const EMPTY_APPLICATIONS = []
const APPLICATION_ROLE_ERROR = '后台返回角色错误'
const VALID_APPLICATION_ROLES = ['guide', 'expert', 'player']
const EMPTY_AUDIT_PAGE = {
  pageTitle: '',
  filters: [],
  statusTexts: {},
  roleNames: {},
  texts: {}
}

function normalizeStatus(status) {
  const value = String(status || '').toLowerCase()

  if (value === 'approved' || value === 'pass' || value === 'passed') {
    return 'approved'
  }

  if (value === 'rejected' || value === 'reject') {
    return 'rejected'
  }

  return 'pending'
}

function applyTemplate(template, values = {}) {
  let text = String(template || '')

  Object.keys(values).forEach((key) => {
    text = text.replace(new RegExp(`\\{${key}\\}`, 'g'), String(values[key]))
  })

  return text
}

function normalizeAuditPage(source = {}) {
  const auditPage = source.auditPage || source || {}

  return {
    pageTitle: String(auditPage.pageTitle || ''),
    filters: Array.isArray(auditPage.filters) ? auditPage.filters : [],
    statusTexts: auditPage.statusTexts || {},
    roleNames: auditPage.roleNames || {},
    texts: auditPage.texts || {}
  }
}

function normalizeApplication(item = {}, auditPage = EMPTY_AUDIT_PAGE) {
  const statusKey = normalizeStatus(item.status || item.statusKey)
  const texts = auditPage.texts || {}
  const userId = item.userId || item.userID || ''
  const nickname = item.nickname || item.userName || item.userNickname || applyTemplate(texts.userFallbackTemplate, { userId })
  const roleKey = String(item.roleKey || '').trim().toLowerCase()

  if (!VALID_APPLICATION_ROLES.includes(roleKey)) {
    throw new Error(APPLICATION_ROLE_ERROR)
  }

  return {
    id: item.id || item.applicationId || '',
    gameId: item.gameId || '',
    initiator: {
      nickname,
      avatarUrl: item.avatarUrl || item.avatarURL || '',
      avatarText: item.avatarText || getSurnameInitials(nickname, texts.avatarFallback || ''),
      roleName: auditPage.roleNames[roleKey]
    },
    roleKey,
    applyTime: item.createdAt || item.applyTime || '-',
    statusKey,
    statusText: auditPage.statusTexts[statusKey] || auditPage.statusTexts.pending || '',
    reason: item.reason || '',
    rejectReason: item.rejectReason || item.reject_reason || ''
  }
}

function filterApplications(items, activeFilter) {
  if (activeFilter === 'all') {
    return items
  }

  return items.filter((item) => item.statusKey === activeFilter)
}

Page({
  data: {
    gameId: '',
    onlineText: '在线',
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
    auditPage: EMPTY_AUDIT_PAGE,
    filters: [],
    applications: EMPTY_APPLICATIONS,
    displayApplications: EMPTY_APPLICATIONS,
    selectedApplicationIds: [],
    loading: false,
    reviewing: false,
    rejectDialogVisible: false,
    rejectDialogTitle: '填写驳回理由',
    rejectDialogApplicationId: '',
    rejectDialogBatch: false,
    rejectReasonDraft: ''
  },

  onLoad(options = {}) {
    this.setData({
      gameId: options.gameId || options.id || ''
    })
    this.loadAuditPageConfig()
      .then(() => this.loadApplications())
  },

  async loadAuditPageConfig() {
    try {
      const config = await gameService.getApplicationConfig()
      const auditPage = normalizeAuditPage(config)
      this.setData({
        auditPage,
        filters: auditPage.filters
      })
    } catch (error) {
      toast.info(error.message || this.textOf('loadFailedText'))
      this.setData({
        auditPage: EMPTY_AUDIT_PAGE,
        filters: []
      })
    }
  },

  textOf(key, values = {}) {
    const texts = this.data.auditPage && this.data.auditPage.texts || {}

    return applyTemplate(texts[key], values)
  },

  async loadApplications() {
    this.setData({ loading: true })

    try {
      const data = await gameService.getReceivedApplications({
        gameId: this.data.gameId || undefined
      })
      const items = (data.items || []).map((item) => normalizeApplication(item, this.data.auditPage))

      this.setData({
        applications: items,
        displayApplications: filterApplications(items, this.data.activeFilter),
        selectedApplicationIds: [],
        allSelected: false
      })
    } catch (error) {
      toast.info(error.message || this.textOf('loadFailedText'))
      this.setData({
        displayApplications: filterApplications(this.data.applications, this.data.activeFilter)
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  handleFilterTap(event) {
    const key = event.currentTarget.dataset.key || 'all'
    const selected = new Set(this.data.selectedApplicationIds.map((item) => String(item)))
    const displayApplications = filterApplications(this.data.applications, key).map((item) => ({ ...item, selected: selected.has(String(item.id)) }))

    this.setData({
      activeFilter: key,
      displayApplications,
      allSelected: false
    })
  },

  handleSelectAllTap() {
    const pendingIds = this.data.displayApplications.filter((item) => item.statusKey === 'pending' && item.id).map((item) => String(item.id))
    const selected = this.data.allSelected ? [] : pendingIds
    const selectedSet = new Set(selected)
    this.setData({
      allSelected: selected.length > 0 && selected.length === pendingIds.length,
      selectedApplicationIds: selected,
      displayApplications: this.data.displayApplications.map((item) => ({ ...item, selected: selectedSet.has(String(item.id)) }))
    })
  },

  handleApplicationSelectTap(event) {
    const id = String(event.currentTarget.dataset.id || '')
    if (!id) {
      return
    }
    const selected = new Set(this.data.selectedApplicationIds.map((item) => String(item)))
    if (selected.has(id)) {
      selected.delete(id)
    } else {
      selected.add(id)
    }
    const pendingIds = this.data.displayApplications.filter((item) => item.statusKey === 'pending' && item.id).map((item) => String(item.id))
    this.setData({
      selectedApplicationIds: [...selected],
      allSelected: pendingIds.length > 0 && pendingIds.every((item) => selected.has(item)),
      displayApplications: this.data.displayApplications.map((item) => ({ ...item, selected: selected.has(String(item.id)) }))
    })
  },

  handleApproveTap(event) {
    this.reviewApplication(event.currentTarget.dataset.id, true)
  },

  handleRejectTap(event) {
    this.openRejectDialog({ applicationId: event.currentTarget.dataset.id })
  },

  handleDetailTap(event) {
    const id = event.currentTarget.dataset.id || ''
    const query = id ? `?auditId=${id}` : ''

    navigateShellRoute(`/${ROUTES.gameAuditDetail}${query}`)
  },

  handleBatchApproveTap() {
    this.reviewVisiblePending(true)
  },

  handleBatchRejectTap() {
    this.openRejectDialog({ batch: true })
  },

  openRejectDialog(options = {}) {
    this.setData({
      rejectDialogVisible: true,
      rejectDialogApplicationId: options.applicationId || '',
      rejectDialogBatch: options.batch === true,
      rejectReasonDraft: ''
    })
  },

  handleRejectReasonInput(event) {
    this.setData({ rejectReasonDraft: event.detail.value || '' })
  },

  closeRejectDialog() {
    this.setData({ rejectDialogVisible: false, rejectReasonDraft: '' })
  },

  noop() {},

  handleRejectReasonConfirm() {
    const reason = String(this.data.rejectReasonDraft || '').trim()
    if (!reason) {
      toast.info('请填写驳回理由')
      return
    }
    const applicationId = this.data.rejectDialogApplicationId
    const batch = this.data.rejectDialogBatch
    this.closeRejectDialog()
    if (batch) {
      this.reviewVisiblePending(false, reason)
      return
    }
    this.reviewApplication(applicationId, false, reason)
  },

  async reviewApplication(applicationId, approve, rejectReason = '') {
    if (!applicationId || this.data.reviewing) {
      return
    }

    this.setData({ reviewing: true })
    wx.showLoading({
      title: approve ? this.textOf('approvingText') : this.textOf('rejectingText'),
      mask: true
    })

    try {
      await gameService.reviewGameApplication(applicationId, approve, rejectReason)
      toast.info(approve ? this.textOf('approveSuccessText') : this.textOf('rejectSuccessText'))
      await this.loadApplications()
    } catch (error) {
      toast.info(error.message || this.textOf('reviewFailedText'))
    } finally {
      wx.hideLoading()
      this.setData({ reviewing: false })
    }
  },

  async reviewVisiblePending(approve, rejectReason = '') {
    const selected = new Set(this.data.selectedApplicationIds.map((item) => String(item)))
    const items = this.data.displayApplications.filter((item) => item.statusKey === 'pending' && item.id && selected.has(String(item.id)))

    if (!items.length || this.data.reviewing) {
      toast.info('请先勾选要审核的申请')
      return
    }

    this.setData({ reviewing: true })
    wx.showLoading({
      title: approve ? this.textOf('batchApprovingText') : this.textOf('batchRejectingText'),
      mask: true
    })

    try {
      for (const item of items) {
        await gameService.reviewGameApplication(item.id, approve, rejectReason)
      }
      toast.info(approve ? this.textOf('batchApproveSuccess') : this.textOf('batchRejectSuccess'))
      await this.loadApplications()
    } catch (error) {
      toast.info(error.message || this.textOf('batchReviewFailedText'))
    } finally {
      wx.hideLoading()
      this.setData({ reviewing: false })
    }
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (key === 'up' || key === 'down') {
      if (!this.suppressNextNavTap) {
        this.scrollAudit(key, AUDIT_SCROLL_TAP_STEP_RPX)
      }
      return
    }

    if (navigateShellKey(key, {
      currentRoute: ROUTES.gameAudit
    })) {
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
      this.navigateToRoute(ROUTES.playerHome || ROUTES.home)
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

    navigateShellRoute(route)
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
