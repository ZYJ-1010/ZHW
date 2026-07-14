const { getLegacyWhiteFrameLayoutStyles } = require('../../../utils/adaptive-shell-layout')
const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const toast = require('../../../utils/toast')
const { getSurnameInitials } = require('../../../utils/avatar')
const { navigateShellRoute } = require('../../../utils/shell-nav')

const CONTENT_LEFT_RPX = 2
const CONTENT_TOP_RPX = 160
const CONTENT_WIDTH_RPX = 750
const DESIGN_FRAME_HEIGHT_PT = 810
const DESIGN_BOTTOM_HEIGHT_PT = 78
const NAV_TITLE_HEIGHT_RPX = 50
const BACK_BUTTON_SIZE_RPX = 40
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = DESIGN_FRAME_HEIGHT_PT * 2
const DAY_MS = 24 * 60 * 60 * 1000

const EMPTY_SUMMARY = {
  label: '',
  amount: '',
  activeCount: 0,
  completedCount: 0,
  canceledCount: 0
}

const EMPTY_ORDER = {
  id: '',
  statusType: 'active',
  statusText: '',
  ref: '',
  avatar: '',
  name: '',
  title: '',
  amount: '',
  guideText: '',
  startedAt: '',
  expectedDeliveryAt: '',
  noticeText: '',
  primaryActionText: '',
  secondaryActionText: ''
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

function firstDefined(...values) {
  return values.find((value) => value !== undefined && value !== null && value !== '')
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

function parseCurrencyAmount(value) {
  if (value === undefined || value === null || value === '') {
    return undefined
  }

  if (typeof value === 'number') {
    return Number.isFinite(value) ? value : undefined
  }

  const amount = Number(String(value).replace(/[¥,\s]/g, ''))

  return Number.isFinite(amount) ? amount : undefined
}

function splitServedText(servedText) {
  const parts = String(servedText || '').split(/[\/／]/).map((item) => item.trim()).filter(Boolean)

  return {
    servedDurationText: parts[0] || '',
    totalDurationText: parts[1] || ''
  }
}

function buildServedText(servedDurationText, totalDurationText) {
  if (servedDurationText && totalDurationText) {
    return `${servedDurationText} / ${totalDurationText}`
  }

  return servedDurationText || totalDurationText || ''
}

function getCount(value, fallback = 0) {
  const count = Number(value)

  return Number.isFinite(count) ? count : fallback
}

function parseDateTime(value) {
  if (value === undefined || value === null || value === '') {
    return null
  }

  if (value instanceof Date) {
    return Number.isNaN(value.getTime()) ? null : value
  }

  if (typeof value === 'number') {
    const date = new Date(value)

    return Number.isNaN(date.getTime()) ? null : date
  }

  const text = String(value).trim()

  if (!text) {
    return null
  }

  const normalized = text.includes('T') ? text : text.replace(/-/g, '/')
  const date = new Date(normalized)

  return Number.isNaN(date.getTime()) ? null : date
}

function padTime(value) {
  return String(value).padStart(2, '0')
}

function formatDateTime(value) {
  const date = parseDateTime(value)

  if (!date) {
    return ''
  }

  return [
    date.getFullYear(),
    padTime(date.getMonth() + 1),
    padTime(date.getDate())
  ].join('-') + ` ${padTime(date.getHours())}:${padTime(date.getMinutes())}`
}

function buildSummary(rawSummary = {}) {
  const source = rawSummary || {}
  const activeCount = getCount(firstDefined(source.activeCount, source.ongoingCount, source.processingCount), EMPTY_SUMMARY.activeCount)
  const completedCount = getCount(firstDefined(source.completedCount, source.completeCount, source.doneCount), EMPTY_SUMMARY.completedCount)
  const canceledCount = getCount(firstDefined(source.canceledCount, source.cancelledCount, source.refundCount, source.refundCancelCount), EMPTY_SUMMARY.canceledCount)
  const amount = firstDefined(source.amountText, source.serviceExpenseText, source.monthlyServiceExpenseText)

  return {
    label: source.label || source.title || EMPTY_SUMMARY.label,
    amount: amount || formatCurrency(
      firstDefined(source.amount, source.serviceExpense, source.monthlyServiceExpense),
      EMPTY_SUMMARY.amount
    ),
    activeCount,
    completedCount,
    canceledCount,
    stats: [
      { text: `进行中 ${activeCount}单` },
      { text: `已完成 ${completedCount}单` },
      { text: `已取消 ${canceledCount}单` }
    ]
  }
}

function buildTabs(summary, activeKey = 'active') {
  return [
    { key: 'active', text: `进行中(${summary.activeCount})`, active: activeKey === 'active' },
    { key: 'complete', text: `已完成(${summary.completedCount})`, active: activeKey === 'complete' },
    { key: 'refund', text: `退款/取消(${summary.canceledCount})`, active: activeKey === 'refund' }
  ]
}

function normalizeStatusType(value) {
  const text = String(value || '').trim().toLowerCase()

  if (/cancel|refund|refuse|reject|取消|退款|赔付/.test(text)) {
    return 'canceled'
  }

  if (/complete|completed|finish|finished|done|success|已完成|完成|成功/.test(text)) {
    return 'complete'
  }

  return 'active'
}

function getTabKeyByStatus(statusType) {
  if (statusType === 'complete') {
    return 'complete'
  }

  if (statusType === 'canceled') {
    return 'refund'
  }

  return 'active'
}

function clampPercent(value, fallback = 0) {
  const percent = Number(value)

  if (!Number.isFinite(percent)) {
    return fallback
  }

  return Math.max(0, Math.min(100, Math.round(percent)))
}

function normalizeBooleanFlag(value) {
  if (typeof value === 'boolean') {
    return value
  }

  if (typeof value === 'number') {
    return value === 1
  }

  if (typeof value === 'string') {
    const text = value.trim().toLowerCase()

    if (['true', '1', 'yes', 'y', 'reviewed', 'completed', 'done', 'evaluated', '已评价', '评价完成'].includes(text)) {
      return true
    }

    if (['false', '0', 'no', 'n', 'pending', 'unreviewed', '待评价', '未评价'].includes(text)) {
      return false
    }
  }

  return undefined
}

function getDefaultOrder(statusType) {
  return {
    ...EMPTY_ORDER,
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

function getAvatarText(rawOrder, fallback) {
  const name = firstDefined(rawOrder.name, rawOrder.expertName, rawOrder.playerName, fallback.name)
  const fallbackText = firstDefined(rawOrder.avatar, rawOrder.avatarText, rawOrder.initials, rawOrder.expertInitials, fallback.avatar)

  return getSurnameInitials(name, fallbackText)
}

function getNestedValue(source, path) {
  return path.reduce((value, key) => (value && value[key] !== undefined ? value[key] : undefined), source)
}

function buildSchedule(rawOrder, fallback, currentTime) {
  const startAt = firstDefined(
    rawOrder.startedAt,
    rawOrder.startAt,
    rawOrder.serviceStartedAt,
    rawOrder.confirmedAt,
    rawOrder.createdAt,
    fallback.startedAt
  )
  const expectedDeliveryAt = firstDefined(
    rawOrder.expectedDeliveryAt,
    rawOrder.deliveryDeadlineAt,
    rawOrder.deliveryAt,
    rawOrder.dueAt,
    fallback.expectedDeliveryAt
  )
  const startDate = parseDateTime(startAt)
  const deliveryDate = parseDateTime(expectedDeliveryAt)
  const nowDate = parseDateTime(firstDefined(rawOrder.currentTime, rawOrder.serverTime, currentTime)) || new Date()
  const fallbackPercent = clampPercent(firstDefined(rawOrder.progressPercent, rawOrder.progress), fallback.progressPercent || 0)
  const fallbackDeliveryText = rawOrder.deliveryText || rawOrder.expectedDeliveryText || fallback.deliveryText

  if (!startDate || !deliveryDate || deliveryDate <= startDate) {
    return {
      progressPercent: fallbackPercent,
      progressText: `${fallbackPercent}%`,
      progressStyle: `width: ${fallbackPercent}%;`,
      deliveryText: fallbackDeliveryText || '',
      elapsedText: rawOrder.elapsedText || fallback.elapsedText || '',
      remainingText: rawOrder.remainingText || fallback.remainingText || ''
    }
  }

  const totalMs = deliveryDate.getTime() - startDate.getTime()
  const elapsedMs = Math.max(0, Math.min(nowDate.getTime() - startDate.getTime(), totalMs))
  const progressPercent = clampPercent(Math.round((elapsedMs / totalMs) * 100), fallbackPercent)
  const totalDays = Math.max(1, Math.ceil(totalMs / DAY_MS))
  const elapsedDays = Math.max(0, Math.min(totalDays, Math.floor(elapsedMs / DAY_MS)))
  const calculatedRemainingDays = Math.max(0, Math.ceil((deliveryDate.getTime() - nowDate.getTime()) / DAY_MS))
  const remainingDays = getCount(firstDefined(rawOrder.remainingDays, rawOrder.leftDays), calculatedRemainingDays)
  const deliveryTimeText = formatDateTime(expectedDeliveryAt)

  return {
    progressPercent,
    progressText: `${progressPercent}%`,
    progressStyle: `width: ${progressPercent}%;`,
    deliveryText: fallbackDeliveryText || (deliveryTimeText ? `预计交付：${deliveryTimeText}` : ''),
    elapsedText: rawOrder.elapsedText || `已进行 ${elapsedDays}/${totalDays} 天`,
    remainingText: rawOrder.remainingText || `剩余 ${remainingDays} 天`
  }
}

function getReviewState(rawOrder = {}, fallback = {}) {
  const reviewStatus = firstDefined(rawOrder.reviewStatus, rawOrder.evaluateStatus, rawOrder.commentStatus, fallback.reviewStatus)
  const reviewedByFlag = firstDefined(
    normalizeBooleanFlag(rawOrder.reviewed),
    normalizeBooleanFlag(rawOrder.hasReviewed),
    normalizeBooleanFlag(rawOrder.hasEvaluated),
    normalizeBooleanFlag(reviewStatus)
  )
  const canReview = firstDefined(
		normalizeBooleanFlag(rawOrder.canReviewAction),
		normalizeBooleanFlag(rawOrder.actions && rawOrder.actions.canReviewAction),
    normalizeBooleanFlag(rawOrder.canReviewBoth),
    normalizeBooleanFlag(rawOrder.canReview),
    normalizeBooleanFlag(fallback.canReviewBoth),
    normalizeBooleanFlag(fallback.canReview)
  )
  const reviewed = reviewedByFlag === true

  return {
    reviewed,
		canReview: canReview === true,
    actionText: reviewed
      ? firstDefined(rawOrder.reviewedActionText, rawOrder.reviewActionText, fallback.reviewedActionText, '已评价')
      : firstDefined(rawOrder.reviewActionText, fallback.reviewActionText, '评价双方')
  }
}

function normalizeOrder(rawOrder = {}, index = 0, context = {}) {
  const statusType = normalizeStatusType(firstDefined(rawOrder.statusType, rawOrder.status, rawOrder.state, rawOrder.statusText))
  const fallback = getDefaultOrder(statusType)
  const actionConfig = rawOrder.actions || {}
  const serviceParts = splitTitleAndAmount(firstDefined(rawOrder.serviceText, rawOrder.service, fallback.title))
  const schedule = buildSchedule(rawOrder, fallback, context.currentTime)
  const reviewState = getReviewState(rawOrder, fallback)
  const amount = firstDefined(
    rawOrder.amountText,
    rawOrder.serviceAmountText,
    rawOrder.fundAmountText,
    rawOrder.expertAmountText,
    rawOrder.serviceFeeText,
    rawOrder.priceText,
    rawOrder.compensationAmountText,
    rawOrder.amount,
    rawOrder.fundAmount,
    rawOrder.expertAmount,
    rawOrder.serviceFee,
    serviceParts.amount,
    fallback.amount
  )
  const amountText = typeof amount === 'number' ? formatCurrency(amount, fallback.amount) : amount
  const completedAtRaw = firstDefined(
    rawOrder.completedAtText,
    rawOrder.completedAt,
    rawOrder.finishedAtText,
    rawOrder.finishedAt,
    rawOrder.deliveredAtText,
    rawOrder.deliveredAt
  )
  const completedAtText = completedAtRaw && String(completedAtRaw).includes('T')
    ? formatDateTime(completedAtRaw)
    : String(completedAtRaw || '')
  const expert = rawOrder.expert || rawOrder.expertInfo || {}
  const expertId = expert.id || rawOrder.expertId || ''
  const guide = rawOrder.guide || rawOrder.guideInfo || rawOrder.referrer || {}
  const cancelPreview = rawOrder.cancelPreview || rawOrder.playerCancelPreview || rawOrder.cancelInfo || {}
  const rawServedText = firstDefined(
    rawOrder.servedText,
    rawOrder.serviceDurationText,
    cancelPreview.servedText,
    cancelPreview.serviceDurationText,
    fallback.servedText
  )
  const servedTextParts = splitServedText(rawServedText)
  const servedDurationText = firstDefined(
    rawOrder.servedDurationText,
    rawOrder.servedDuration,
    rawOrder.servedHoursText,
    rawOrder.servedTimeText,
    cancelPreview.servedDurationText,
    cancelPreview.servedDuration,
    cancelPreview.servedHoursText,
    cancelPreview.servedTimeText,
    servedTextParts.servedDurationText
  )
  const totalDurationText = firstDefined(
    rawOrder.totalDurationText,
    rawOrder.totalDuration,
    rawOrder.totalHoursText,
    rawOrder.serviceTotalDurationText,
    cancelPreview.totalDurationText,
    cancelPreview.totalDuration,
    cancelPreview.totalHoursText,
    cancelPreview.serviceTotalDurationText,
    servedTextParts.totalDurationText
  )
  const servedText = buildServedText(servedDurationText, totalDurationText) || rawServedText || ''

  return {
    id: rawOrder.id || rawOrder.orderId || rawOrder.gameId || '',
    serviceOrderId: rawOrder.serviceOrderId || rawOrder.orderId || rawOrder.id || '',
    gameId: rawOrder.gameId || '',
    expertId,
    statusType,
    tabKey: getTabKeyByStatus(statusType),
    statusText: rawOrder.statusText || fallback.statusText,
    statusClass: statusType === 'active' ? 'active' : statusType === 'complete' ? 'complete' : 'canceled',
    cardClass: statusType === 'active' ? 'active-card' : statusType === 'complete' ? 'complete-card' : 'canceled-card',
    avatarClass: statusType === 'active' ? 'active' : statusType === 'complete' ? 'complete' : 'canceled',
    muted: statusType === 'canceled',
    ref: firstDefined(rawOrder.ref, rawOrder.refNo, rawOrder.orderNo, rawOrder.serviceNo, rawOrder.gameNo, rawOrder.groupNo, fallback.ref),
    name: firstDefined(rawOrder.name, rawOrder.expertName, expert.name, expert.nickname, getNestedValue(rawOrder, ['expertUser', 'nickname']), fallback.name),
    avatar: getAvatarText({
      ...rawOrder,
      avatar: firstDefined(rawOrder.avatar, rawOrder.avatarText, expert.avatarText, expert.initials),
      name: firstDefined(rawOrder.name, rawOrder.expertName, expert.name, expert.nickname, getNestedValue(rawOrder, ['expertUser', 'nickname']), fallback.name)
    }, fallback),
    title: firstDefined(rawOrder.gameTitle, rawOrder.title, rawOrder.serviceTitle, rawOrder.serviceName, serviceParts.title, fallback.title),
    servedDurationText,
    totalDurationText,
    servedText,
    minRate: firstDefined(rawOrder.minRate, rawOrder.cancelMinRate, rawOrder.playerCancelMinRate, cancelPreview.minRate),
    maxRate: firstDefined(rawOrder.maxRate, rawOrder.cancelMaxRate, rawOrder.playerCancelMaxRate, cancelPreview.maxRate),
    suggestedRate: firstDefined(rawOrder.suggestedRate, rawOrder.cancelSuggestedRate, rawOrder.playerCancelSuggestedRate, cancelPreview.suggestedRate),
    suggestionMinRate: firstDefined(rawOrder.suggestionMinRate, rawOrder.recommendMinRate, rawOrder.cancelRecommendMinRate, cancelPreview.suggestionMinRate, cancelPreview.recommendMinRate),
    suggestionMaxRate: firstDefined(rawOrder.suggestionMaxRate, rawOrder.recommendMaxRate, rawOrder.cancelRecommendMaxRate, cancelPreview.suggestionMaxRate, cancelPreview.recommendMaxRate),
    contractAmount: firstDefined(
      parseCurrencyAmount(rawOrder.fundAmount),
      parseCurrencyAmount(rawOrder.contractAmount),
      parseCurrencyAmount(rawOrder.amount),
      parseCurrencyAmount(rawOrder.amountText),
      parseCurrencyAmount(rawOrder.serviceAmountText),
      parseCurrencyAmount(rawOrder.fundAmountText),
      parseCurrencyAmount(rawOrder.serviceFee),
      parseCurrencyAmount(rawOrder.serviceFeeText),
      parseCurrencyAmount(rawOrder.price),
      parseCurrencyAmount(rawOrder.priceText),
      parseCurrencyAmount(serviceParts.amount),
      0
    ),
    amount: amountText,
    guideText: rawOrder.guideText || (firstDefined(rawOrder.guideName, guide.name, guide.nickname) ? `领路人：${firstDefined(rawOrder.guideName, guide.name, guide.nickname)}` : fallback.guideText),
    progressText: schedule.progressText,
    progressStyle: schedule.progressStyle,
    deliveryText: schedule.deliveryText,
    elapsedText: schedule.elapsedText,
    remainingText: schedule.remainingText,
    noticeText: rawOrder.noticeText || actionConfig.noticeText || '',
    primaryActionText: rawOrder.primaryActionText || actionConfig.primaryActionText || '',
    secondaryActionText: rawOrder.secondaryActionText || actionConfig.secondaryActionText || '',
    contactExpertRoute: firstDefined(actionConfig.contactExpertRoute, rawOrder.contactExpertRoute),
    playerCancelRoute: firstDefined(actionConfig.playerCancelRoute, actionConfig.cancelRoute, rawOrder.playerCancelRoute, rawOrder.cancelRoute),
    reviewRoute: firstDefined(actionConfig.reviewRoute, rawOrder.reviewRoute),
    canContactExpert: firstDefined(actionConfig.canContactExpert, rawOrder.canContactExpert, Boolean(expertId)),
    canCancelOrder: firstDefined(actionConfig.canCancelOrder, rawOrder.canCancelOrder, false),
    completeSummary: rawOrder.completeSummary || rawOrder.completedSummary || fallback.completeSummary,
    completedAt: completedAtText || fallback.completedAt,
    completedRows: [
      { label: '局信息', value: firstDefined(rawOrder.gameInfoText, rawOrder.serviceTitle, rawOrder.title, serviceParts.title) },
      { label: '完成时间', value: completedAtText },
      { label: '费用', value: amountText }
    ].filter((item) => item.value),
    resultText: rawOrder.resultText || fallback.resultText,
    reviewActionText: reviewState.actionText,
		reviewDisabled: reviewState.reviewed || !reviewState.canReview,
    reasonSummary: rawOrder.reasonSummary || rawOrder.cancelSummary || fallback.reasonSummary,
    reasonText: rawOrder.reasonText || rawOrder.cancelReasonText || rawOrder.reason || fallback.reasonText
  }
}

function getOrdersPayload(data = {}) {
  return firstDefined(
    data.orders,
    data.orderList,
    data.gameOrders,
    data.cards,
    data.list,
    data.records,
    data.rows,
    data.items
  )
}

function normalizeOrders(rawOrders, context = {}) {
  if (!Array.isArray(rawOrders)) {
    return []
  }

  return rawOrders.map((item, index) => normalizeOrder(item, index, context))
}

function getDisplayOrders(orders, activeTabKey) {
  return (orders || []).filter((item) => item.tabKey === activeTabKey)
}

function buildDisplayState(orders, activeTabKey) {
  const displayOrders = getDisplayOrders(orders, activeTabKey)

  return {
    displayOrders,
    hasDisplayOrders: displayOrders.length > 0,
    emptyText: getEmptyText(activeTabKey)
  }
}

function getEmptyText(activeTabKey) {
  if (activeTabKey === 'complete') {
    return '暂无已完成业务'
  }

  if (activeTabKey === 'refund') {
    return '暂无退款/取消订单'
  }

  return '暂无进行中的业务'
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

function buildCancelQuery(order = {}) {
  const params = {
    orderId: order.id,
    serviceOrderId: order.serviceOrderId,
    gameId: order.gameId,
    ref: order.ref,
    expertName: order.name,
    expertAvatarText: order.avatar,
    serviceTitle: order.title,
    amount: order.contractAmount,
    servedDurationText: order.servedDurationText,
    totalDurationText: order.totalDurationText,
    servedText: order.servedText,
    minRate: order.minRate,
    maxRate: order.maxRate,
    suggestedRate: order.suggestedRate,
    suggestionMinRate: order.suggestionMinRate,
    suggestionMaxRate: order.suggestionMaxRate
  }

  return Object.keys(params)
    .filter((key) => params[key] !== undefined && params[key] !== null && params[key] !== '')
    .map((key) => `${encodeURIComponent(key)}=${encodeURIComponent(params[key])}`)
    .join('&')
}

const INITIAL_SUMMARY = buildSummary(EMPTY_SUMMARY)
const INITIAL_ORDERS = []
const INITIAL_DISPLAY_STATE = buildDisplayState(INITIAL_ORDERS, 'active')

Page({
  data: {
    shellLayout: getWhiteShellLayoutStyles(),
    summary: INITIAL_SUMMARY,
    tabs: buildTabs(INITIAL_SUMMARY),
    activeTabKey: 'active',
    orders: INITIAL_ORDERS,
    displayOrders: INITIAL_DISPLAY_STATE.displayOrders,
    hasDisplayOrders: INITIAL_DISPLAY_STATE.hasDisplayOrders,
    emptyText: INITIAL_DISPLAY_STATE.emptyText,
    page: 1, pageSize: 5, total: 0, hasPrevious: false, hasMore: false
  },

  onLoad(options = {}) {
    this.loadManageData(options)
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

  async loadManageData(options = {}) {
    try {
      const data = await gameService.getPlayerGameManage({ ...options, page: options.page || 1, pageSize: 5, status: options.status || this.data.activeTabKey })
      const summary = buildSummary(data && data.summary ? data.summary : data)
      const orders = normalizeOrders(getOrdersPayload(data), {
        currentTime: firstDefined(data && data.currentTime, data && data.serverTime)
      })
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
      toast.info(error && error.message ? error.message : '组局管理加载失败')
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
    this.loadManageData({ page: 1, status: key })
  },

  onPreviousPage() { if (this.data.hasPrevious) this.loadManageData({ page: this.data.page - 1, status: this.data.activeTabKey }) },
  onNextPage() { if (this.data.hasMore) this.loadManageData({ page: this.data.page + 1, status: this.data.activeTabKey }) },

  onContactExpert(event = {}) {
    const order = findOrderByEvent(this.data.orders, event)

    if (!order || !order.gameId) {
      toast.info('暂无可联系的行家')
      return
    }

    if (!isAllowed(order.canContactExpert)) {
      toast.info('暂无可联系的行家')
      return
    }

    if (navigateRoute(order.contactExpertRoute)) {
      return
    }

    toast.info('暂无可联系的行家')
  },

  onCancelOrder(event) {
    const order = findOrderByEvent(this.data.orders, event)

    if (!order) {
      toast.info('未找到服务信息')
      return
    }

    if (!isAllowed(order.canCancelOrder)) {
      toast.info('当前状态不可申请取消')
      return
    }

    if (navigateRoute(order.playerCancelRoute)) {
      return
    }

    navigateShellRoute(`${ROUTES.gamePlayerCancel}?${buildCancelQuery(order)}`)
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
      `targetUserId=${encodeURIComponent(order.expertId || '')}`,
      `targetRole=${encodeURIComponent('member')}`,
      `title=${encodeURIComponent(order.title || '')}`,
      `name=${encodeURIComponent(order.name || '')}`
    ].join('&')

    navigateShellRoute(`${ROUTES.gameReview}?${params}`)
  }
})
