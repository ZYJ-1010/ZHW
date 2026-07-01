const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')

const CONTENT_TOP_RPX = 160
const CONTENT_WIDTH_RPX = 750
const DESIGN_FRAME_HEIGHT_PT = 810
const DESIGN_BOTTOM_HEIGHT_PT = 34
const NAV_TITLE_HEIGHT_RPX = 50
const BACK_BUTTON_SIZE_RPX = 40
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = DESIGN_FRAME_HEIGHT_PT * 2
const ASSET_BASE = '/pages/game/referral-record/assets'

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

function getShellLayoutStyles() {
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
      `left: 0rpx`,
      `top: ${CONTENT_TOP_RPX}rpx`,
      `width: ${CONTENT_WIDTH_RPX}rpx`,
      `height: ${contentHeight}rpx`
    ].join('; '),
    bottomStyle: `top: ${bottomTop}rpx; height: ${bottomHeight}rpx;`,
    titleStyle: `top: ${titleTop}rpx; height: ${NAV_TITLE_HEIGHT_RPX}rpx; line-height: ${NAV_TITLE_HEIGHT_RPX}rpx;`,
    backStyle: `top: ${backTop}rpx; width: ${BACK_BUTTON_SIZE_RPX}rpx; height: ${BACK_BUTTON_SIZE_RPX}rpx;`
  }
}

function firstDefined() {
  const values = Array.prototype.slice.call(arguments)

  return values.find((value) => value !== undefined && value !== null && value !== '') || ''
}

function normalizeBooleanFlag(value) {
  if (value === undefined || value === null || value === '') {
    return undefined
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

  return Boolean(value)
}

function getRewardClass(record = {}) {
  if (record.rewardClass) {
    return record.rewardClass
  }

  if (record.state === 'completed') {
    return 'green'
  }

  if (record.state === 'canceled') {
    return 'gray'
  }

  return 'orange'
}

function getReviewState(record = {}) {
  const actionConfig = Array.isArray(record.actions) ? (record.actionConfig || {}) : (record.actions || {})
  const reviewStatus = firstDefined(record.reviewStatus, record.evaluateStatus, record.commentStatus)
  const reviewedByFlag = firstDefined(
    normalizeBooleanFlag(record.reviewed),
    normalizeBooleanFlag(record.hasReviewed),
    normalizeBooleanFlag(record.hasEvaluated),
    normalizeBooleanFlag(reviewStatus)
  )
  const canReview = firstDefined(
    normalizeBooleanFlag(record.canReviewBoth),
    normalizeBooleanFlag(record.canReview),
    normalizeBooleanFlag(actionConfig.canReviewBoth),
    normalizeBooleanFlag(actionConfig.canReview)
  )
  const reviewed = reviewedByFlag === true || canReview === false

  return {
    reviewed,
    tagText: reviewed
      ? firstDefined(record.reviewedTagText, record.extraTag)
      : firstDefined(record.extraTag, record.reviewTagText),
    actionText: reviewed
      ? firstDefined(record.reviewedActionText, record.reviewActionText)
      : firstDefined(record.reviewActionText)
  }
}

function normalizePerson(person = {}) {
  return {
    name: person.name || person.nickname || '',
    avatarText: person.avatarText || person.initials || '',
    avatarClass: person.avatarClass || ''
  }
}

function normalizeAction(action = {}) {
  return {
    key: action.key || action.actionKey || action.type || '',
    text: action.text || action.label || action.name || '',
    icon: action.icon || action.iconSrc || '',
    theme: action.theme || '',
    disabled: Boolean(action.disabled),
    route: action.route || action.path || '',
    message: action.message || action.toastText || ''
  }
}

function normalizeReward(value) {
  return String(value || '').replace(/^¥/, '')
}

function normalizeRecord(record = {}) {
  const actionList = Array.isArray(record.actions) ? record.actions.map(normalizeAction) : []
  const reviewState = getReviewState(record)

  return {
    ...record,
    id: record.id || record.recordId || '',
    state: record.state || record.status || '',
    stateText: record.stateText || record.statusText || '',
    stateIcon: record.stateIcon || record.statusIcon || `${ASSET_BASE}/status-processing.png`,
    stateClass: record.stateClass || '',
    timeText: record.timeText || record.createdAtText || '',
    expert: normalizePerson(record.expert || {}),
    player: normalizePerson(record.player || {}),
    matchIcon: record.matchIcon || `${ASSET_BASE}/handshake.png`,
    serviceTitle: record.serviceTitle || record.title || '',
    reward: normalizeReward(firstDefined(record.reward, record.rewardAmount, record.rewardText)),
    rewardClass: getRewardClass(record),
    reviewed: reviewState.reviewed,
    extraTag: firstDefined(record.extraTag, reviewState.tagText),
    noticeText: record.noticeText || '',
    actions: actionList.map((action) => action.key === 'review'
      ? {
        ...action,
        text: reviewState.actionText || action.text,
        disabled: reviewState.reviewed
      }
      : action)
  }
}

function normalizeSummary(summary = {}) {
  return {
    label: summary.label || summary.title || '',
    amount: summary.amount || summary.amountText || '',
    background: summary.background || '',
    iconSrc: summary.iconSrc || summary.icon || `${ASSET_BASE}/wallet.png`,
    stats: Array.isArray(summary.stats) ? summary.stats : []
  }
}

function normalizeTab(tab = {}) {
  return {
    key: tab.key || tab.status || '',
    label: tab.label || tab.name || '',
    count: Number(tab.count || 0)
  }
}

function normalizeReferralData(data = {}, activeTab = '') {
  const records = Array.isArray(data.records || data.list || data.items)
    ? (data.records || data.list || data.items).map(normalizeRecord).filter((item) => item.id)
    : []
  const tabs = Array.isArray(data.tabs)
    ? data.tabs.map(normalizeTab).filter((item) => item.key)
    : []
  const nextActiveTab = activeTab || data.activeTab || tabs[0] && tabs[0].key || ''

  return {
    pageTitle: data.pageTitle || data.title || '',
    summary: normalizeSummary(data.summary || {}),
    tabs,
    activeTab: nextActiveTab,
    records,
    visibleRecords: nextActiveTab ? records.filter((item) => item.state === nextActiveTab) : records
  }
}

Page({
  data: {
    pageTitle: '',
    shellLayout: getShellLayoutStyles(),
    summary: normalizeSummary({}),
    tabs: [],
    activeTab: '',
    visibleRecords: [],
    records: [],
    loading: false,
    queryParams: {}
  },

  onLoad(options = {}) {
    this.setData({
      queryParams: options
    })
    this.loadReferralRecords(options)
  },

  onShow() {
    this.updateShellLayout()
  },

  onResize() {
    this.updateShellLayout()
  },

  updateShellLayout() {
    this.setData({
      shellLayout: getShellLayoutStyles()
    })
  },

  async loadReferralRecords(params = {}) {
    this.setData({
      loading: true
    })

    try {
      const data = await gameService.getReferralRecords({
        ...this.data.queryParams,
        ...params,
        status: this.data.activeTab
      })

      this.setData({
        ...normalizeReferralData(data, this.data.activeTab),
        loading: false
      })
    } catch (error) {
      this.setData({
        ...normalizeReferralData({}, this.data.activeTab),
        loading: false
      })
      this.showToast(error.message || '引荐记录加载失败')
    }
  },

  onTabTap(event) {
    const key = event.currentTarget.dataset.key

    if (!key || key === this.data.activeTab) {
      return
    }

    this.setData({
      activeTab: key
    })
    this.loadReferralRecords({
      status: key
    })
  },

  onActionTap(event) {
    const action = event.currentTarget.dataset.action
    const disabled = event.currentTarget.dataset.disabled === true || event.currentTarget.dataset.disabled === 'true'

    if (disabled) {
      return
    }

    const matchedRecord = this.data.visibleRecords.find((record) => (
      record.actions || []
    ).some((item) => item.key === action))
    const matchedAction = matchedRecord && (matchedRecord.actions || []).find((item) => item.key === action)

    if (matchedAction && matchedAction.route) {
      wx.navigateTo({
        url: matchedAction.route
      })
      return
    }

    if (matchedRecord && matchedAction) {
      this.submitRecordAction(matchedRecord.id, matchedAction)
      return
    }

    this.showToast('操作待接入')
  },

  async submitRecordAction(recordId, action) {
    try {
      const result = await gameService.triggerReferralRecordAction({
        recordId,
        actionKey: action.key
      })

      this.showToast(result && (result.message || result.toastText) || action.message || '操作已提交')
      this.loadReferralRecords()
    } catch (error) {
      this.showToast(error.message || '操作提交失败')
    }
  },

  showToast(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  },

  onBackTap() {
    const pages = getCurrentPages()

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    wx.navigateTo({
      url: `/${ROUTES.guideHome}`
    })
  }
})
