const { getLegacyWhiteFrameLayoutStyles } = require('../../../utils/adaptive-shell-layout')
const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const toast = require('../../../utils/toast')
const { getSurnameInitials } = require('../../../utils/avatar')
const { navigateShellRoute } = require('../../../utils/shell-nav')

const CONTENT_LEFT_RPX = 0
const CONTENT_TOP_RPX = 160
const CONTENT_WIDTH_RPX = 750
const DESIGN_FRAME_HEIGHT_PT = 810
const DESIGN_BOTTOM_HEIGHT_PT = 78
const NAV_TITLE_HEIGHT_RPX = 50
const BACK_BUTTON_SIZE_RPX = 40
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = DESIGN_FRAME_HEIGHT_PT * 2

const EMPTY_BUSINESS_SUMMARY = {
  label: '',
  amount: '',
  activeCount: 0,
  pendingSettlementCount: 0,
  completedCount: 0,
  disputeCount: 0
}

const EMPTY_TIMELINE = []
const EMPTY_BUSINESS_ORDER = {
  id: '',
  statusType: 'active',
  statusText: '',
  ref: '',
  avatarText: '',
  avatarClass: 'pink',
  name: '',
  roleTag: '',
  serviceTitle: '',
  amountText: '',
  guideText: '',
  timeline: EMPTY_TIMELINE,
  primaryActionText: '',
  secondaryActionText: '',
  playerActionText: '',
  guideActionText: ''
}

function navigateRoute(route) {
  const url = route ? `/${String(route).replace(/^\/+/, '')}` : ''
  if (!url) {
    return false
  }

  navigateShellRoute(url)
  return true
}

function roundRpx(value) {
  return Math.round(value * 100) / 100
}

function getCapsuleBottomRpx() {
  try {
    if (typeof wx !== 'undefined' && wx.getSystemInfoSync && wx.getMenuButtonBoundingClientRect) {
      const menuButton = wx.getMenuButtonBoundingClientRect()
      const systemInfo = wx.getSystemInfoSync()

      if (menuButton && systemInfo && systemInfo.windowWidth) {
        return roundRpx((menuButton.top + menuButton.height) * 750 / systemInfo.windowWidth)
      }
    }
  } catch (error) {
    return DEFAULT_CAPSULE_BOTTOM_RPX
  }

  return DEFAULT_CAPSULE_BOTTOM_RPX
}

function getFrameHeightRpx() {
  try {
    if (typeof wx !== 'undefined' && wx.getSystemInfoSync) {
      const systemInfo = wx.getSystemInfoSync()

      if (systemInfo && systemInfo.windowWidth && systemInfo.windowHeight) {
        return roundRpx(systemInfo.windowHeight * 750 / systemInfo.windowWidth)
      }
    }
  } catch (error) {
    return DEFAULT_FRAME_HEIGHT_RPX
  }

  return DEFAULT_FRAME_HEIGHT_RPX
}

function getWhiteShellLayoutStyles() {
  return getLegacyWhiteFrameLayoutStyles({
    contentLeftRpx: CONTENT_LEFT_RPX,
    contentWidthRpx: CONTENT_WIDTH_RPX,
    bottomHeightRatio: DESIGN_BOTTOM_HEIGHT_PT / DESIGN_FRAME_HEIGHT_PT,
    titleHeightRpx: NAV_TITLE_HEIGHT_RPX,
    backSizeRpx: BACK_BUTTON_SIZE_RPX
  })
}

function formatCurrency(value, fallback = '¥0') {
  if (value === undefined || value === null || value === '') {
    return fallback
  }

  if (typeof value === 'string' && value.includes('¥')) {
    return value
  }

  const amount = Number(value)

  if (!Number.isFinite(amount)) {
    return String(value)
  }

  return `¥${amount.toLocaleString('zh-CN')}`
}

function getCount(value, fallback = 0) {
  const count = Number(value)

  return Number.isFinite(count) ? count : fallback
}

function firstDefined(...values) {
  return values.find((value) => value !== undefined && value !== null && value !== '')
}

function firstArray(...values) {
  return values.find((value) => Array.isArray(value) && value.length)
}

function buildBusinessSummary(rawSummary = {}) {
  const source = rawSummary || {}
  const activeCount = getCount(firstDefined(source.activeCount, source.ongoingCount, source.processingCount), EMPTY_BUSINESS_SUMMARY.activeCount)
  const pendingSettlementCount = getCount(firstDefined(source.pendingSettlementCount, source.settlementCount, source.waitingSettlementCount), EMPTY_BUSINESS_SUMMARY.pendingSettlementCount)
  const completedCount = getCount(firstDefined(source.completedCount, source.completeCount, source.doneCount), EMPTY_BUSINESS_SUMMARY.completedCount)
  const disputeCount = getCount(firstDefined(source.disputeCount, source.canceledCount, source.cancelledCount, source.refundCount), EMPTY_BUSINESS_SUMMARY.disputeCount)
  const amount = firstDefined(source.amountText, source.serviceIncomeText, source.monthlyServiceIncomeText, source.monthlyIncomeText)

  return {
    label: source.label || source.title || EMPTY_BUSINESS_SUMMARY.label,
    amount: amount || formatCurrency(
      firstDefined(source.amount, source.serviceIncome, source.monthlyServiceIncome, source.monthlyIncome),
      EMPTY_BUSINESS_SUMMARY.amount
    ),
    activeCount,
    pendingSettlementCount,
    completedCount,
    disputeCount,
    stats: [
      { text: `进行中 ${activeCount}单` },
      { text: `待结算 ${pendingSettlementCount}单` },
      { text: `已完成 ${completedCount}单` }
    ]
  }
}

function buildTabs(summary, activeKey = 'active') {
  return [
    { key: 'active', text: `进行中(${summary.activeCount})`, active: activeKey === 'active' },
    { key: 'complete', text: `已完成(${summary.completedCount})`, active: activeKey === 'complete' },
    { key: 'dispute', text: `争议(${summary.disputeCount})`, active: activeKey === 'dispute' }
  ]
}

function normalizeStatusType(value) {
  const text = String(value || '').trim().toLowerCase()

  if (/cancel|refund|refuse|reject|dispute|取消|退款|赔付|争议/.test(text)) {
    return 'canceled'
  }

  if (/complete|completed|finish|finished|done|success|early|delivered|已完成|完成|成功|提前|已交付|提前交付/.test(text)) {
    return 'complete'
  }

  return 'active'
}

function getTabKeyByStatus(statusType) {
  if (statusType === 'complete') {
    return 'complete'
  }

  if (statusType === 'canceled') {
    return 'dispute'
  }

  return 'active'
}

function getDefaultOrder(statusType) {
  return {
    ...EMPTY_BUSINESS_ORDER,
    statusType
  }
}

function splitTitleAndAmount(value) {
  const text = String(value || '').trim()

  if (!text) {
    return {}
  }

  const parts = text.split(/\s*[·|]\s*/)
  const amountIndex = parts.findIndex((item) => /^¥/.test(item.trim()))

  if (amountIndex < 0) {
    return { title: text }
  }

  return {
    title: parts.filter((item, index) => index !== amountIndex).join(' · ').trim(),
    amount: parts[amountIndex].trim()
  }
}

function buildAvatarText(rawOrder, fallback) {
  const name = String(firstDefined(rawOrder.name, rawOrder.playerName, rawOrder.expertName, fallback.name) || '').trim()
  const fallbackText = firstDefined(
    rawOrder.avatar,
    rawOrder.avatarText,
    rawOrder.initials,
    rawOrder.playerInitials,
    rawOrder.expertInitials,
    fallback.avatarText,
    fallback.avatar,
    'EN'
  )

  return getSurnameInitials(name, fallbackText)
}

function normalizeTimeline(rawTimeline, fallbackTimeline) {
  const source = Array.isArray(rawTimeline) && rawTimeline.length ? rawTimeline : fallbackTimeline || EMPTY_TIMELINE

  return source.map((item, index) => {
    const state = normalizeStageState(item.state || item.status || item.statusType || item.type, index)

    return {
      id: item.id || `timeline-${index}`,
      title: item.title || item.text || '',
      time: item.time || item.timeText || item.desc || '',
      state,
      isDone: state === 'done',
      last: index === source.length - 1
    }
  })
}

function normalizeStageState(value, index) {
  const text = String(value || '').trim().toLowerCase()

  if (/done|complete|completed|finish|finished|success|已完成|完成|成功/.test(text)) {
    return 'done'
  }

  if (/current|active|processing|progress|ongoing|进行中|当前/.test(text)) {
    return 'current'
  }

  if (/future|pending|waiting|todo|wait|等待|待/.test(text)) {
    return 'future'
  }

  return index === 0 ? 'done' : index === 1 ? 'current' : 'future'
}

function normalizeSettlementRows(rawOrder, fallback) {
  const sourceRows = Array.isArray(rawOrder.settlementRows) && rawOrder.settlementRows.length
    ? rawOrder.settlementRows
    : fallback.settlementRows

  if (Array.isArray(sourceRows) && sourceRows.length) {
    return sourceRows.map((item) => ({
      label: item.label,
      value: item.value,
      highlight: Boolean(item.highlight)
    }))
  }

  const rows = []
  const durationText = firstDefined(rawOrder.actualDurationText, rawOrder.serviceDurationText)
  const incomeText = firstDefined(rawOrder.actualIncomeText, rawOrder.settlementAmountText)

  if (durationText) {
    rows.push({ label: '实际服务时长', value: durationText })
  }

  if (incomeText) {
    rows.push({ label: '实际收入', value: incomeText, highlight: true })
  }

  return rows
}

function hasSettlementRows(rawOrder, statusText) {
  if (Array.isArray(rawOrder.settlementRows) && rawOrder.settlementRows.length) {
    return true
  }

  if (firstDefined(rawOrder.actualDurationText, rawOrder.serviceDurationText, rawOrder.actualIncomeText, rawOrder.settlementAmountText)) {
    return true
  }

  return /early|提前|已交付|提前交付/.test(String(statusText || rawOrder.statusText || rawOrder.statusType || ''))
}

function getReviewState(rawOrder = {}, fallback = {}, actionConfig = {}) {
	const reviewed = Boolean(firstDefined(
    rawOrder.reviewed,
    rawOrder.hasReviewed,
    rawOrder.reviewStatus === 'reviewed',
    rawOrder.reviewStatus === 'completed',
    fallback.reviewed,
		false
	))
	const canReview = Boolean(firstDefined(
		actionConfig.canReviewAction,
		rawOrder.canReviewAction,
		actionConfig.canReview,
		rawOrder.canReview,
		fallback.canReview,
		false
	))

	return {
		reviewed,
		canReview,
    reviewActionText: reviewed
      ? firstDefined(rawOrder.reviewedActionText, rawOrder.reviewActionText, fallback.reviewedActionText, '已评价')
      : firstDefined(rawOrder.reviewActionText, fallback.reviewActionText, '评价双方')
  }
}

function normalizeBusinessOrder(rawOrder = {}, index = 0) {
  const statusType = normalizeStatusType(firstDefined(rawOrder.statusType, rawOrder.status, rawOrder.state, rawOrder.statusText))
  const fallback = getDefaultOrder(statusType)
  const player = rawOrder.player || rawOrder.playerInfo || rawOrder.client || {}
  const guide = rawOrder.guide || rawOrder.guideInfo || rawOrder.referrer || {}
  const service = rawOrder.service && typeof rawOrder.service === 'object' ? rawOrder.service : {}
  const actionConfig = rawOrder.actions || {}
  const serviceParts = splitTitleAndAmount(firstDefined(
    rawOrder.serviceText,
    service.serviceText,
    service.text,
    typeof rawOrder.service === 'string' ? rawOrder.service : '',
    fallback.serviceText
  ))
  const title = firstDefined(
    rawOrder.gameTitle,
    rawOrder.title,
    rawOrder.serviceTitle,
    rawOrder.serviceName,
    service.title,
    service.name,
    service.serviceTitle,
    service.serviceName,
    serviceParts.title,
    fallback.serviceTitle
  )
  const amount = firstDefined(
    rawOrder.amountText,
    rawOrder.serviceAmountText,
    rawOrder.priceText,
    rawOrder.incomeText,
    rawOrder.compensationAmountText,
    service.amountText,
    service.serviceAmountText,
    service.priceText,
    service.incomeText,
    rawOrder.amount,
    service.amount,
    serviceParts.amount,
    fallback.amountText
  )
  const playerName = firstDefined(player.name, player.nickname, player.displayName, rawOrder.playerName, rawOrder.name, rawOrder.expertName, fallback.name)
  const guideName = firstDefined(guide.name, guide.nickname, guide.displayName, rawOrder.guideName, rawOrder.referrerName)
  const avatarSource = {
    ...rawOrder,
    avatar: firstDefined(player.avatarText, player.avatar, player.initials, player.avatarFallback, rawOrder.avatar, rawOrder.avatarText),
    avatarText: firstDefined(player.avatarText, player.initials, rawOrder.avatarText),
    name: playerName
  }
  const settlementSource = {
    ...rawOrder,
    actualDurationText: firstDefined(rawOrder.actualDurationText, service.actualDurationText, service.durationText, service.actualDuration),
    serviceDurationText: firstDefined(rawOrder.serviceDurationText, service.serviceDurationText, service.plannedDurationText),
    actualIncomeText: firstDefined(rawOrder.actualIncomeText, service.actualIncomeText, service.incomeText),
    settlementAmountText: firstDefined(rawOrder.settlementAmountText, service.settlementAmountText)
  }
  const hasSettlement = hasSettlementRows(settlementSource, rawOrder.statusText || fallback.statusText)
	const reviewState = getReviewState(rawOrder, fallback, actionConfig)
  const isEarlyComplete = statusType === 'complete' && /early|提前|已交付|提前交付/.test(String(rawOrder.statusText || rawOrder.statusType || ''))
  const completeCardClass = hasSettlement || isEarlyComplete ? 'completed-card early-card' : 'completed-card'
  const shouldAutoShowGuide = statusType === 'active' || (statusType === 'complete' && !isEarlyComplete)

  return {
    id: rawOrder.id || rawOrder.orderId || rawOrder.gameId || '',
    serviceOrderId: rawOrder.serviceOrderId || rawOrder.orderId || rawOrder.id || '',
    gameId: rawOrder.gameId || service.gameId || '',
    playerId: player.id || rawOrder.playerId || '',
    guideId: guide.id || rawOrder.guideId || '',
    statusType,
    isActive: statusType === 'active',
    isComplete: statusType === 'complete',
    isCanceled: statusType === 'canceled',
    tabKey: getTabKeyByStatus(statusType),
    statusText: rawOrder.statusText || fallback.statusText,
    statusClass: statusType === 'active' ? 'active' : statusType === 'complete' ? 'complete' : 'canceled',
    cardClass: statusType === 'active' ? 'active-card' : statusType === 'complete' ? completeCardClass : 'canceled-card',
    avatarClass: rawOrder.avatarClass || fallback.avatarClass || (statusType === 'active' ? 'pink' : 'green'),
    compact: statusType !== 'active',
    muted: statusType === 'canceled',
    ref: firstDefined(rawOrder.ref, rawOrder.refNo, rawOrder.orderNo, rawOrder.serviceOrderNo, rawOrder.gameNo, rawOrder.groupNo, rawOrder.partyNo, service.ref, fallback.ref),
    avatar: buildAvatarText(avatarSource, fallback),
    name: playerName,
    roleTag: firstDefined(player.roleTag, player.roleText, rawOrder.roleTag, rawOrder.roleText, fallback.roleTag),
    title,
    amount: typeof amount === 'number' ? formatCurrency(amount, fallback.amountText) : amount,
    guideText: rawOrder.guideText || guide.text || (
      shouldAutoShowGuide && guideName ? `领路人：${guideName}` : fallback.guideText
    ),
    guideTone: rawOrder.guideTone || fallback.guideTone || '',
    timeline: normalizeTimeline(firstArray(
      rawOrder.timeline,
      service.timeline,
      rawOrder.serviceStages,
      service.serviceStages,
      service.stages,
      rawOrder.progressStages,
      service.progressStages,
      rawOrder.progressTimeline,
      rawOrder.statusTimeline
    ), fallback.timeline),
    primaryActionText: rawOrder.primaryActionText || fallback.primaryActionText,
    secondaryActionText: rawOrder.secondaryActionText || fallback.secondaryActionText,
    playerActionText: rawOrder.playerActionText || fallback.playerActionText,
    guideActionText: rawOrder.guideActionText || fallback.guideActionText,
    deliveryRoute: firstDefined(actionConfig.finishDeliveryRoute, actionConfig.deliveryRoute, rawOrder.finishDeliveryRoute, rawOrder.deliveryRoute),
    expertCancelRoute: firstDefined(actionConfig.expertCancelRoute, actionConfig.cancelRoute, rawOrder.expertCancelRoute, rawOrder.cancelRoute),
    contactPlayerRoute: firstDefined(actionConfig.contactPlayerRoute, rawOrder.contactPlayerRoute),
    contactGuideRoute: firstDefined(actionConfig.contactGuideRoute, rawOrder.contactGuideRoute),
    reviewRoute: firstDefined(actionConfig.reviewRoute, rawOrder.reviewRoute),
    canFinishDelivery: firstDefined(actionConfig.canFinishDelivery, rawOrder.canFinishDelivery, false),
    canCancelWithCompensation: firstDefined(actionConfig.canCancelWithCompensation, rawOrder.canCancelWithCompensation, false),
    canContactPlayer: firstDefined(actionConfig.canContactPlayer, rawOrder.canContactPlayer, false),
    canContactGuide: firstDefined(actionConfig.canContactGuide, rawOrder.canContactGuide, false),
    hasSettlementRows: hasSettlement,
    settlementRows: normalizeSettlementRows(settlementSource, fallback),
		reviewed: reviewState.reviewed,
		canReview: reviewState.canReview,
		reviewDisabled: reviewState.reviewed || !reviewState.canReview,
    reviewActionText: reviewState.reviewActionText,
    completeSummary: rawOrder.completeSummary || rawOrder.completedSummary || fallback.completeSummary,
    completedAt: rawOrder.completedAtText || rawOrder.completedAt || fallback.completedAtText || fallback.completedAt,
    resultText: rawOrder.resultText || fallback.resultText,
    reasonSummary: rawOrder.reasonSummary || rawOrder.cancelSummary || fallback.reasonSummary,
    reasonText: rawOrder.reasonText || rawOrder.cancelReasonText || rawOrder.reason || fallback.reasonText
  }
}

function normalizeBusinessOrders(rawOrders) {
  if (!Array.isArray(rawOrders)) {
    return []
  }

  return rawOrders.map(normalizeBusinessOrder)
}

function getDisplayOrders(orders, activeTabKey) {
  return (orders || []).filter((item) => item.tabKey === activeTabKey)
}

function getBusinessOrdersPayload(data = {}) {
  return firstDefined(
    data.orders,
    data.orderList,
    data.businessOrders,
    data.gameOrders,
    data.cards,
    data.list,
    data.records,
    data.rows,
    data.items
  )
}

function getEmptyText(activeTabKey) {
  if (activeTabKey === 'complete') {
    return '暂无已完成业务'
  }

  if (activeTabKey === 'dispute') {
    return '暂无争议订单'
  }

  return '暂无进行中的业务'
}

function buildDisplayState(orders, activeTabKey) {
  const displayOrders = getDisplayOrders(orders, activeTabKey)

  return {
    displayOrders,
    hasDisplayOrders: displayOrders.length > 0,
    emptyText: getEmptyText(activeTabKey)
  }
}

function findOrderByEvent(orders = [], event = {}) {
  const orderId = event.currentTarget && event.currentTarget.dataset ? event.currentTarget.dataset.id : ''

  if (!orderId) {
    return null
  }

  return orders.find((item) => String(item.id) === String(orderId)) || null
}

function isAllowed(value) {
  return value === true
}

const INITIAL_BUSINESS_SUMMARY = buildBusinessSummary(EMPTY_BUSINESS_SUMMARY)
const INITIAL_BUSINESS_ORDERS = []
const INITIAL_DISPLAY_STATE = buildDisplayState(INITIAL_BUSINESS_ORDERS, 'active')

Page({
  data: {
    shellLayout: getWhiteShellLayoutStyles(),
    roleType: 'expert',
    summary: INITIAL_BUSINESS_SUMMARY,
    tabs: buildTabs(INITIAL_BUSINESS_SUMMARY),
    activeTabKey: 'active',
    orders: INITIAL_BUSINESS_ORDERS,
    displayOrders: INITIAL_DISPLAY_STATE.displayOrders,
    hasDisplayOrders: INITIAL_DISPLAY_STATE.hasDisplayOrders,
    emptyText: INITIAL_DISPLAY_STATE.emptyText,
    page: 1, pageSize: 5, total: 0, hasPrevious: false, hasMore: false
  },

  onLoad(options = {}) {
    this.loadBusinessSummary(options)
  },

  onShow() {
    this.updateShellLayout()
  },

  onResize() {
    this.updateShellLayout()
  },

  updateShellLayout() {
    this.setData({
      shellLayout: getWhiteShellLayoutStyles()
    })
  },

  async loadBusinessSummary(options = {}) {
    try {
      const data = await gameService.getGameManage({ ...options, page: options.page || 1, pageSize: 5, status: options.status || this.data.activeTabKey })
      const summary = buildBusinessSummary(data && data.summary ? data.summary : data)
      const orders = normalizeBusinessOrders(getBusinessOrdersPayload(data))
      const displayState = buildDisplayState(orders, this.data.activeTabKey)

      this.setData({
        summary,
        tabs: buildTabs(summary, this.data.activeTabKey),
        orders,
        page: data.page || 1,
        total: data.total || 0,
        hasPrevious: Boolean(data.hasPrevious),
        hasMore: Boolean(data.hasMore),
        ...displayState
      })
    } catch (error) {
      toast.info(error && error.message ? error.message : '我的业务管理加载失败')
    }
  },

  onBackTap() {
    const pages = getCurrentPages()

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    navigateShellRoute(ROUTES.profile)
  },

  onTabTap(event) {
    const key = event.currentTarget.dataset.key
    this.setData({
      activeTabKey: key,
      tabs: this.data.tabs.map((item) => ({
        ...item,
        active: item.key === key
      }))
    })
    this.loadBusinessSummary({ page: 1, status: key })
  },

  onPreviousPage() { if (this.data.hasPrevious) this.loadBusinessSummary({ page: this.data.page - 1, status: this.data.activeTabKey }) },
  onNextPage() { if (this.data.hasMore) this.loadBusinessSummary({ page: this.data.page + 1, status: this.data.activeTabKey }) },

  onEndDelivery(event = {}) {
    const order = findOrderByEvent(this.data.orders, event)

    if (!order || !order.gameId) {
      toast.info('暂无可结束的业务')
      return
    }

    if (!isAllowed(order.canFinishDelivery)) {
      toast.info('当前状态不可结束交付')
      return
    }

    if (navigateRoute(order.deliveryRoute)) {
      return
    }

    const params = [
      `gameId=${encodeURIComponent(order.gameId)}`,
      `mode=${encodeURIComponent('paid')}`,
      `note=${encodeURIComponent('确认服务完成')}`
    ].join('&')

    navigateShellRoute(`${ROUTES.gameDelivery}?${params}`)
  },

  onCancelWithCompensation(event = {}) {
    const order = findOrderByEvent(this.data.orders, event)

    if (!order) {
      toast.info('暂无可取消的业务')
      return
    }

    if (!isAllowed(order.canCancelWithCompensation)) {
      toast.info('当前状态不可取消并赔付')
      return
    }

    if (navigateRoute(order.expertCancelRoute)) {
      return
    }

    const params = [
      `serviceOrderId=${encodeURIComponent(order.serviceOrderId || order.id || '')}`,
      `gameId=${encodeURIComponent(order.gameId || '')}`,
      `playerId=${encodeURIComponent(order.playerId || '')}`,
      `playerName=${encodeURIComponent(order.name || '')}`,
      `avatarText=${encodeURIComponent(order.avatar || '')}`,
      `serviceTitle=${encodeURIComponent(order.title || '')}`,
      `amountText=${encodeURIComponent(order.amount || '')}`,
      `statusText=${encodeURIComponent(order.statusText || '')}`
    ].join('&')

    navigateShellRoute(`${ROUTES.gameExpertCancel}?${params}`)
  },

  onContactPlayer(event = {}) {
    const order = findOrderByEvent(this.data.orders, event)

    if (!order || !order.gameId) {
      toast.info('暂无可联系的玩家')
      return
    }

    if (!isAllowed(order.canContactPlayer)) {
      toast.info('暂无可联系的玩家')
      return
    }

    if (navigateRoute(order.contactPlayerRoute)) {
      return
    }

    toast.info('暂无可联系的玩家')
  },

  onContactGuide(event = {}) {
    const order = findOrderByEvent(this.data.orders, event)

    if (!order || !order.gameId) {
      toast.info('暂无可联系的领路人')
      return
    }

    if (!isAllowed(order.canContactGuide)) {
      toast.info('暂无可联系的领路人')
      return
    }

    if (navigateRoute(order.contactGuideRoute)) {
      return
    }

    toast.info('暂无可联系的领路人')
  },

  onReviewBoth(event = {}) {
    const order = findOrderByEvent(this.data.orders, event)

    if (!order || !order.gameId) {
      toast.info('暂无可评价的业务')
      return
    }

    if (navigateRoute(order.reviewRoute)) {
      return
    }

    const params = [
      `gameId=${encodeURIComponent(order.gameId)}`,
      `targetUserId=${encodeURIComponent(order.playerId || order.guideId || '')}`,
      `targetRole=${encodeURIComponent('member')}`,
      `title=${encodeURIComponent(order.title || '')}`,
      `name=${encodeURIComponent(order.name || '')}`
    ].join('&')

    navigateShellRoute(`${ROUTES.gameReview}?${params}`)
  }
})
