const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const toast = require('../../../utils/toast')

const CONTENT_LEFT_RPX = 2
const CONTENT_TOP_RPX = 160
const CONTENT_WIDTH_RPX = 750
const DESIGN_FRAME_HEIGHT_PT = 810
const DESIGN_BOTTOM_HEIGHT_PT = 78
const NAV_TITLE_HEIGHT_RPX = 50
const BACK_BUTTON_SIZE_RPX = 40
const SERVICE_BUTTON_SIZE_RPX = 44
const SERVICE_BUTTON_LEFT_RPX = 512
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = DESIGN_FRAME_HEIGHT_PT * 2
const DEFAULT_CURRENT_TIME = '2026-03-19T14:00:00+08:00'
const DAY_MS = 24 * 60 * 60 * 1000

const DEFAULT_SUMMARY = {
  label: '本月服务支出',
  amount: '¥0',
  activeCount: 0,
  completedCount: 0,
  canceledCount: 0
}

const DEFAULT_ORDERS = [
  {
    id: 'player-manage-active-001',
    statusType: 'active',
    statusText: '服务进行中',
    ref: '',
    avatar: '',
    name: '',
    title: '',
    amount: '',
    guideText: '',
    startedAt: '',
    expectedDeliveryAt: '',
    noticeText: '取消需赔付一定比例金额给行家',
    primaryActionText: '联系行家',
    secondaryActionText: '申请取消'
  },
  {
    id: 'player-manage-complete-001',
    statusType: 'complete',
    statusText: '已完成',
    ref: '',
    avatar: '',
    name: '',
    title: '',
    amount: '',
    guideText: '',
    completeSummary: '服务已完成',
    completedAt: '',
    resultText: '',
    reviewStatus: 'pending',
    reviewActionText: '评价双方'
  },
  {
    id: 'player-manage-canceled-001',
    statusType: 'canceled',
    statusText: '已取消（已赔付）',
    ref: '',
    avatar: '',
    name: '',
    title: '',
    reasonSummary: '',
    amount: '',
    reasonText: ''
  }
]

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
  const capsuleBottom = getCapsuleBottomRpx()
  const frameHeight = getFrameHeightRpx()
  const bottomHeight = roundRpx(frameHeight * DESIGN_BOTTOM_HEIGHT_PT / DESIGN_FRAME_HEIGHT_PT)
  const bottomTop = Math.max(CONTENT_TOP_RPX, roundRpx(frameHeight - bottomHeight))
  const contentHeight = Math.max(0, roundRpx(bottomTop - CONTENT_TOP_RPX))
  const titleTop = Math.max(0, roundRpx(capsuleBottom - NAV_TITLE_HEIGHT_RPX))
  const backTop = Math.max(0, roundRpx(capsuleBottom - BACK_BUTTON_SIZE_RPX))
  const serviceTop = Math.max(0, roundRpx(capsuleBottom - SERVICE_BUTTON_SIZE_RPX - 3))

  return {
    frameStyle: `height: ${frameHeight}rpx; min-height: ${frameHeight}rpx;`,
    contentStyle: [
      `left: ${CONTENT_LEFT_RPX}rpx`,
      `top: ${CONTENT_TOP_RPX}rpx`,
      `width: ${CONTENT_WIDTH_RPX}rpx`,
      `height: ${contentHeight}rpx`
    ].join('; '),
    bottomStyle: `top: ${bottomTop}rpx; height: ${bottomHeight}rpx;`,
    titleStyle: `top: ${titleTop}rpx; height: ${NAV_TITLE_HEIGHT_RPX}rpx; line-height: ${NAV_TITLE_HEIGHT_RPX}rpx;`,
    backStyle: `top: ${backTop}rpx; width: ${BACK_BUTTON_SIZE_RPX}rpx; height: ${BACK_BUTTON_SIZE_RPX}rpx;`,
    serviceStyle: `left: ${SERVICE_BUTTON_LEFT_RPX}rpx; top: ${serviceTop}rpx; width: ${SERVICE_BUTTON_SIZE_RPX}rpx; height: ${SERVICE_BUTTON_SIZE_RPX}rpx;`
  }
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
  const activeCount = getCount(firstDefined(source.activeCount, source.ongoingCount, source.processingCount), DEFAULT_SUMMARY.activeCount)
  const completedCount = getCount(firstDefined(source.completedCount, source.completeCount, source.doneCount), DEFAULT_SUMMARY.completedCount)
  const canceledCount = getCount(firstDefined(source.canceledCount, source.cancelledCount, source.refundCount, source.refundCancelCount), DEFAULT_SUMMARY.canceledCount)
  const amount = firstDefined(source.amountText, source.serviceExpenseText, source.monthlyServiceExpenseText)

  return {
    label: source.label || source.title || DEFAULT_SUMMARY.label,
    amount: amount || formatCurrency(
      firstDefined(source.amount, source.serviceExpense, source.monthlyServiceExpense),
      DEFAULT_SUMMARY.amount
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
  return DEFAULT_ORDERS.find((item) => item.statusType === statusType) || DEFAULT_ORDERS[0]
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
  const value = firstDefined(rawOrder.avatar, rawOrder.avatarText, rawOrder.initials, rawOrder.expertInitials)

  if (value) {
    return String(value).trim().slice(0, 2).toUpperCase()
  }

  return fallback.avatar
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
    normalizeBooleanFlag(rawOrder.canReviewBoth),
    normalizeBooleanFlag(rawOrder.canReview),
    normalizeBooleanFlag(fallback.canReviewBoth),
    normalizeBooleanFlag(fallback.canReview)
  )
  const reviewed = reviewedByFlag === true || canReview === false

  return {
    reviewed,
    actionText: reviewed
      ? firstDefined(rawOrder.reviewedActionText, rawOrder.reviewActionText, fallback.reviewedActionText, '已评价')
      : firstDefined(rawOrder.reviewActionText, fallback.reviewActionText, '评价双方')
  }
}

function normalizeOrder(rawOrder = {}, index = 0, context = {}) {
  const statusType = normalizeStatusType(firstDefined(rawOrder.statusType, rawOrder.status, rawOrder.state, rawOrder.statusText))
  const fallback = getDefaultOrder(statusType)
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
  const expert = rawOrder.expert || rawOrder.expertInfo || {}
  const guide = rawOrder.guide || rawOrder.guideInfo || rawOrder.referrer || {}

  return {
    id: rawOrder.id || rawOrder.orderId || rawOrder.gameId || `player-manage-order-${index}`,
    statusType,
    tabKey: getTabKeyByStatus(statusType),
    statusText: rawOrder.statusText || fallback.statusText,
    statusClass: statusType === 'active' ? 'active' : statusType === 'complete' ? 'complete' : 'canceled',
    cardClass: statusType === 'active' ? 'active-card' : statusType === 'complete' ? 'complete-card' : 'canceled-card',
    avatarClass: statusType === 'active' ? 'active' : statusType === 'complete' ? 'complete' : 'canceled',
    muted: statusType === 'canceled',
    ref: firstDefined(rawOrder.ref, rawOrder.refNo, rawOrder.orderNo, rawOrder.serviceNo, rawOrder.gameNo, rawOrder.groupNo, fallback.ref),
    avatar: getAvatarText({
      ...rawOrder,
      avatar: firstDefined(rawOrder.avatar, rawOrder.avatarText, expert.avatarText, expert.initials)
    }, fallback),
    name: firstDefined(rawOrder.name, rawOrder.expertName, expert.name, expert.nickname, getNestedValue(rawOrder, ['expertUser', 'nickname']), fallback.name),
    title: firstDefined(rawOrder.gameTitle, rawOrder.title, rawOrder.serviceTitle, rawOrder.serviceName, serviceParts.title, fallback.title),
    amount: typeof amount === 'number' ? formatCurrency(amount, fallback.amount) : amount,
    guideText: rawOrder.guideText || (firstDefined(rawOrder.guideName, guide.name, guide.nickname) ? `领路人：${firstDefined(rawOrder.guideName, guide.name, guide.nickname)}` : fallback.guideText),
    progressText: schedule.progressText,
    progressStyle: schedule.progressStyle,
    deliveryText: schedule.deliveryText,
    elapsedText: schedule.elapsedText,
    remainingText: schedule.remainingText,
    noticeText: rawOrder.noticeText || fallback.noticeText,
    primaryActionText: rawOrder.primaryActionText || fallback.primaryActionText,
    secondaryActionText: rawOrder.secondaryActionText || fallback.secondaryActionText,
    completeSummary: rawOrder.completeSummary || rawOrder.completedSummary || fallback.completeSummary,
    completedAt: rawOrder.completedAtText || rawOrder.completedAt || fallback.completedAt,
    resultText: rawOrder.resultText || fallback.resultText,
    reviewActionText: reviewState.actionText,
    reviewDisabled: reviewState.reviewed,
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

function normalizeOrders(rawOrders, context = {}, useDefault = false) {
  if (!Array.isArray(rawOrders)) {
    return useDefault ? DEFAULT_ORDERS.map((item, index) => normalizeOrder(item, index, context)) : []
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
    hasDisplayOrders: displayOrders.length > 0
  }
}

const INITIAL_SUMMARY = buildSummary(DEFAULT_SUMMARY)
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
    hasDisplayOrders: INITIAL_DISPLAY_STATE.hasDisplayOrders
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
      const data = await gameService.getPlayerGameManage(options)
      const summary = buildSummary(data && data.summary ? data.summary : data)
      const orders = normalizeOrders(getOrdersPayload(data), {
        currentTime: firstDefined(data && data.currentTime, data && data.serverTime)
      })
      const displayState = buildDisplayState(orders, this.data.activeTabKey)

      this.setData({
        summary,
        tabs: buildTabs(summary, this.data.activeTabKey),
        orders,
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

    wx.navigateTo({
      url: `/${ROUTES.gameHall}`
    })
  },

  onTabTap(event) {
    const key = event.currentTarget.dataset.key
    const displayState = buildDisplayState(this.data.orders, key)

    this.setData({
      activeTabKey: key,
      tabs: this.data.tabs.map((item) => ({
        ...item,
        active: item.key === key
      })),
      ...displayState
    })
  },

  onContactExpert() {
    toast.info('联系行家功能开发中')
  },

  onCancelOrder() {
    toast.info('取消申请功能开发中')
  },

  onReviewBoth() {
    toast.info('评价页待接入')
  }
})
