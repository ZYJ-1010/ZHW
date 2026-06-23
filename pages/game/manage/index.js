const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const toast = require('../../../utils/toast')

const CONTENT_LEFT_RPX = 0
const CONTENT_TOP_RPX = 160
const CONTENT_WIDTH_RPX = 750
const DESIGN_FRAME_HEIGHT_PT = 810
const DESIGN_BOTTOM_HEIGHT_PT = 78
const NAV_TITLE_HEIGHT_RPX = 50
const BACK_BUTTON_SIZE_RPX = 40
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = DESIGN_FRAME_HEIGHT_PT * 2

const DEFAULT_BUSINESS_SUMMARY = {
  label: '本月服务收入',
  amount: '¥5,280',
  activeCount: 1,
  pendingSettlementCount: 1,
  completedCount: 8,
  disputeCount: 0
}

const DEFAULT_TIMELINE = [
  { id: 'group-success', title: '组局成功', time: '03-20 14:30', state: 'done' },
  { id: 'service-active', title: '服务进行中', time: '预计交付：03-25', state: 'current' },
  { id: 'waiting-confirm', title: '等待确认完成', time: '', state: 'future' }
]

const DEFAULT_BUSINESS_ORDERS = [
  {
    id: 'business-active-001',
    statusType: 'active',
    statusText: '服务进行中',
    ref: 'REF-20260320-001',
    avatarText: 'LI',
    avatarClass: 'pink',
    name: '李明',
    roleTag: '玩家',
    serviceTitle: '产品架构咨询',
    amountText: '¥800',
    guideText: '领路人：王引荐',
    timeline: DEFAULT_TIMELINE,
    primaryActionText: '提前结束交付',
    secondaryActionText: '取消并赔付',
    playerActionText: '联系玩家',
    guideActionText: '联系领路人'
  },
  {
    id: 'business-early-001',
    statusType: 'complete',
    statusText: '已提前交付',
    ref: 'REF-20260318-004',
    avatarText: 'ZH',
    avatarClass: 'purple',
    name: '赵经理',
    serviceTitle: '技术咨询',
    amountText: '¥600',
    guideText: '提前2天完成',
    guideTone: 'success',
    settlementRows: [
      { label: '实际服务时长', value: '1.5小时 (原定2小时)' },
      { label: '实际收入', value: '¥450 (按比例结算)', highlight: true }
    ],
    reviewActionText: '评价双方'
  },
  {
    id: 'business-complete-001',
    statusType: 'complete',
    statusText: '已完成',
    ref: 'REF-20260312-006',
    avatarText: 'WA',
    avatarClass: 'green',
    name: '王同学',
    serviceTitle: '品牌定位咨询',
    amountText: '¥1,200',
    completeSummary: '服务已完成',
    completedAtText: '完成时间：03-15 18:30',
    resultText: '双方已确认，收入已进入结算',
    settlementRows: [
      { label: '实际服务时长', value: '2小时' },
      { label: '实际收入', value: '¥1,200', highlight: true }
    ]
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
    backStyle: `top: ${backTop}rpx; width: ${BACK_BUTTON_SIZE_RPX}rpx; height: ${BACK_BUTTON_SIZE_RPX}rpx;`
  }
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
  const activeCount = getCount(firstDefined(source.activeCount, source.ongoingCount, source.processingCount), DEFAULT_BUSINESS_SUMMARY.activeCount)
  const pendingSettlementCount = getCount(firstDefined(source.pendingSettlementCount, source.settlementCount, source.waitingSettlementCount), DEFAULT_BUSINESS_SUMMARY.pendingSettlementCount)
  const completedCount = getCount(firstDefined(source.completedCount, source.completeCount, source.doneCount), DEFAULT_BUSINESS_SUMMARY.completedCount)
  const disputeCount = getCount(firstDefined(source.disputeCount, source.canceledCount, source.cancelledCount, source.refundCount), DEFAULT_BUSINESS_SUMMARY.disputeCount)
  const amount = firstDefined(source.amountText, source.serviceIncomeText, source.monthlyServiceIncomeText, source.monthlyIncomeText)

  return {
    label: source.label || source.title || DEFAULT_BUSINESS_SUMMARY.label,
    amount: amount || formatCurrency(
      firstDefined(source.amount, source.serviceIncome, source.monthlyServiceIncome, source.monthlyIncome),
      DEFAULT_BUSINESS_SUMMARY.amount
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
  if (statusType === 'complete') {
    return DEFAULT_BUSINESS_ORDERS.find((item) => item.id === 'business-complete-001') || DEFAULT_BUSINESS_ORDERS[0]
  }

  return DEFAULT_BUSINESS_ORDERS.find((item) => item.statusType === statusType) || DEFAULT_BUSINESS_ORDERS[0]
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

function getChineseSurnameInitials(name) {
  const surname = String(name || '').trim().charAt(0)
  const initialsMap = {
    张: 'ZH',
    王: 'WA',
    李: 'LI',
    刘: 'LI',
    陈: 'CH',
    杨: 'YA',
    黄: 'HU',
    赵: 'ZH',
    吴: 'WU',
    周: 'ZH',
    徐: 'XU',
    孙: 'SU',
    马: 'MA',
    朱: 'ZH',
    胡: 'HU',
    郭: 'GU',
    何: 'HE',
    林: 'LI',
    高: 'GA',
    罗: 'LU',
    郑: 'ZH'
  }

  return initialsMap[surname] || ''
}

function buildAvatarText(rawOrder, fallback) {
  const value = firstDefined(
    rawOrder.avatar,
    rawOrder.avatarText,
    rawOrder.initials,
    rawOrder.playerInitials,
    rawOrder.expertInitials
  )

  if (value) {
    return String(value).trim().slice(0, 2).toUpperCase()
  }

  const name = String(firstDefined(rawOrder.name, rawOrder.playerName, rawOrder.expertName, fallback.name) || '').trim()
  const chineseInitials = getChineseSurnameInitials(name)

  if (chineseInitials) {
    return chineseInitials
  }

  const letters = name.match(/[A-Za-z]/g)

  if (letters && letters.length) {
    return letters.slice(0, 2).join('').toUpperCase()
  }

  return fallback.avatarText || fallback.avatar || 'EN'
}

function normalizeTimeline(rawTimeline, fallbackTimeline) {
  const source = Array.isArray(rawTimeline) && rawTimeline.length ? rawTimeline : fallbackTimeline || DEFAULT_TIMELINE

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

  return [
    {
      label: '实际服务时长',
      value: firstDefined(rawOrder.actualDurationText, rawOrder.serviceDurationText, '1.5小时 (原定2小时)')
    },
    {
      label: '实际收入',
      value: firstDefined(rawOrder.actualIncomeText, rawOrder.settlementAmountText, '¥450 (按比例结算)'),
      highlight: true
    }
  ]
}

function hasSettlementRows(rawOrder, statusText) {
  if (Array.isArray(rawOrder.settlementRows) && rawOrder.settlementRows.length) {
    return true
  }

  if (normalizeStatusType(firstDefined(rawOrder.statusType, rawOrder.status, rawOrder.state, statusText)) === 'complete') {
    return true
  }

  if (firstDefined(rawOrder.actualDurationText, rawOrder.serviceDurationText, rawOrder.actualIncomeText, rawOrder.settlementAmountText)) {
    return true
  }

  return /early|提前|已交付|提前交付/.test(String(statusText || rawOrder.statusText || rawOrder.statusType || ''))
}

function getReviewState(rawOrder = {}, fallback = {}) {
  const reviewed = Boolean(firstDefined(
    rawOrder.reviewed,
    rawOrder.hasReviewed,
    rawOrder.reviewStatus === 'reviewed',
    rawOrder.reviewStatus === 'completed',
    fallback.reviewed,
    false
  ))

  return {
    reviewed,
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
  const reviewState = getReviewState(rawOrder, fallback)
  const isEarlyComplete = statusType === 'complete' && /early|提前|已交付|提前交付/.test(String(rawOrder.statusText || rawOrder.statusType || ''))
  const shouldAutoShowGuide = statusType === 'active' || (statusType === 'complete' && !isEarlyComplete)

  return {
    id: rawOrder.id || rawOrder.orderId || rawOrder.gameId || `business-order-${index}`,
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
    cardClass: statusType === 'active' ? 'active-card' : statusType === 'complete' ? 'completed-card' : 'canceled-card',
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
    primaryActionText: rawOrder.primaryActionText || fallback.primaryActionText || '提前结束交付',
    secondaryActionText: rawOrder.secondaryActionText || fallback.secondaryActionText || '取消并赔付',
    playerActionText: rawOrder.playerActionText || fallback.playerActionText || '联系玩家',
    guideActionText: rawOrder.guideActionText || fallback.guideActionText || '联系领路人',
    canFinishDelivery: firstDefined(actionConfig.canFinishDelivery, rawOrder.canFinishDelivery, true),
    canCancelWithCompensation: firstDefined(actionConfig.canCancelWithCompensation, rawOrder.canCancelWithCompensation, true),
    canContactPlayer: firstDefined(actionConfig.canContactPlayer, rawOrder.canContactPlayer, true),
    canContactGuide: firstDefined(actionConfig.canContactGuide, rawOrder.canContactGuide, true),
    hasSettlementRows: hasSettlement,
    settlementRows: normalizeSettlementRows(settlementSource, fallback),
    reviewed: reviewState.reviewed,
    reviewDisabled: reviewState.reviewed,
    reviewActionText: reviewState.reviewActionText,
    completeSummary: rawOrder.completeSummary || rawOrder.completedSummary || fallback.completeSummary || '服务已完成',
    completedAt: rawOrder.completedAtText || rawOrder.completedAt || fallback.completedAtText || fallback.completedAt,
    resultText: rawOrder.resultText || fallback.resultText || '双方已确认，收入已进入结算',
    reasonSummary: rawOrder.reasonSummary || rawOrder.cancelSummary || fallback.reasonSummary || '争议处理中',
    reasonText: rawOrder.reasonText || rawOrder.cancelReasonText || rawOrder.reason || fallback.reasonText || '请关注双方协商与平台处理结果'
  }
}

function normalizeBusinessOrders(rawOrders) {
  if (!Array.isArray(rawOrders)) {
    return DEFAULT_BUSINESS_ORDERS.map(normalizeBusinessOrder)
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

const INITIAL_BUSINESS_SUMMARY = buildBusinessSummary(DEFAULT_BUSINESS_SUMMARY)
const INITIAL_BUSINESS_ORDERS = DEFAULT_BUSINESS_ORDERS.map(normalizeBusinessOrder)
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
    emptyText: INITIAL_DISPLAY_STATE.emptyText
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
      const data = await gameService.getGameManage(options)
      const summary = buildBusinessSummary(data && data.summary ? data.summary : data)
      const orders = normalizeBusinessOrders(getBusinessOrdersPayload(data))
      const displayState = buildDisplayState(orders, this.data.activeTabKey)

      this.setData({
        summary,
        tabs: buildTabs(summary, this.data.activeTabKey),
        orders,
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

  onEndDelivery() {
    toast.info('提前结束交付功能开发中')
  },

  onCancelWithCompensation(event = {}) {
    const orderId = event.currentTarget && event.currentTarget.dataset ? event.currentTarget.dataset.id : ''
    const order = this.data.orders.find((item) => item.id === orderId) || this.data.displayOrders[0]

    if (!order) {
      toast.info('取消服务确认页待接入')
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

    wx.navigateTo({
      url: `/${ROUTES.gameExpertCancel}?${params}`
    })
  },

  onContactPlayer() {
    toast.info('联系玩家功能开发中')
  },

  onContactGuide() {
    toast.info('联系领路人功能开发中')
  },

  onReviewBoth() {
    toast.info('评价双方功能开发中')
  }
})
