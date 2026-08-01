const { getLegacyWhiteFrameLayoutStyles } = require('../../../utils/adaptive-shell-layout')
const gameService = require('../../../services/game')
const { ROUTES } = require('../../../config/routes')
const { navigateShellRoute } = require('../../../utils/shell-nav')
const { toUserMessage } = require('../../../utils/user-message')

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
  return getLegacyWhiteFrameLayoutStyles({
    contentLeftRpx: 0,
    contentWidthRpx: CONTENT_WIDTH_RPX,
    bottomHeightRatio: DESIGN_BOTTOM_HEIGHT_PT / DESIGN_FRAME_HEIGHT_PT,
    titleHeightRpx: NAV_TITLE_HEIGHT_RPX,
    backSizeRpx: BACK_BUTTON_SIZE_RPX
  })
}

function firstDefined(...values) {
  return values.find((value) => value !== undefined && value !== null && value !== '')
}

function applyTemplate(template, values = {}) {
  return String(template || '').replace(/\{(\w+)\}/g, (_, key) => values[key] == null ? '' : values[key])
}

function normalizePageConfig(config = {}) {
  return {
    pageTitle: config.pageTitle || '',
    summary: config.summary || {},
    texts: config.texts || {}
  }
}

function textOf(config, key) {
  const texts = config && config.texts ? config.texts : {}
  return texts[key] || ''
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

function getStateIcon(state) {
  const icons = {
    processing: `${ASSET_BASE}/status-processing-dot.svg`,
    completed: `${ASSET_BASE}/status-complete-check.svg`,
    canceled: `${ASSET_BASE}/status-cancel-x.svg`
  }

  return icons[state] || icons.processing
}

function getMatchIcon(state) {
  return state === 'canceled'
    ? `${ASSET_BASE}/status-cancel-x.svg`
    : `${ASSET_BASE}/deal-handshake-yellow.svg`
}

function normalizeActionIcon(action = {}) {
  if (action.icon) {
    return action.icon
  }

  if (action.key === 'chat') {
    return `${ASSET_BASE}/action-chat.svg`
  }

  if (action.key === 'remind') {
    return `${ASSET_BASE}/action-bell-blue.svg`
  }

  return ''
}

function getReviewState(record = {}, pageConfig = {}) {
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
  const reviewed = reviewedByFlag === true

  return {
    reviewed,
    tagText: reviewed
      ? firstDefined(record.reviewedTagText, textOf(pageConfig, 'reviewedTagText'))
      : firstDefined(record.extraTag, record.reviewTagText, textOf(pageConfig, 'pendingReviewTagText')),
    actionText: reviewed
      ? firstDefined(record.reviewedActionText, record.reviewActionText, textOf(pageConfig, 'reviewedActionText'))
      : firstDefined(record.reviewActionText, textOf(pageConfig, 'reviewActionText'))
  }
}

function normalizeRecord(record = {}, pageConfig = {}) {
  const recordId = firstDefined(record.id, record.recordId, record.referralId)
  const gameId = firstDefined(record.gameId, record.serviceOrderId, record.orderId, recordId)
  const expertUserId = firstDefined(record.expertUserId, record.expert && record.expert.userId)
  const playerUserId = firstDefined(record.playerUserId, record.player && record.player.userId)
  const reviewTargetUserId = firstDefined(record.reviewTargetUserId, expertUserId, playerUserId)
  const state = firstDefined(record.state, record.statusType, 'processing')
  const expertName = firstDefined(record.expert && record.expert.name, record.expertName, '')
  const playerName = firstDefined(record.player && record.player.name, record.playerName, '')
  const cancelReason = firstDefined(record.cancelReasonText, record.reasonText, record.cancelReason, '')
  const normalized = {
    ...record,
    id: recordId,
    gameId,
    state,
    partyText: firstDefined(record.partyText, expertName && playerName ? `${expertName} ↔ ${playerName}` : expertName || playerName),
    cancelReasonText: cancelReason ? (/取消原因/.test(cancelReason) ? cancelReason : `取消原因：${cancelReason}`) : '',
    stateIcon: firstDefined(record.stateIcon, getStateIcon(state)),
    matchIcon: firstDefined(record.matchIcon, getMatchIcon(state)),
    expertUserId,
    playerUserId,
    reviewTargetUserId
  }
  const actionList = Array.isArray(record.actions) ? record.actions : (record.actionList || [])

  if (record.state !== 'completed') {
    return {
      ...normalized,
      actions: actionList.map((action) => ({
        ...action,
        icon: normalizeActionIcon(action)
      }))
    }
  }

  const reviewState = getReviewState(record, pageConfig)

  return {
    ...normalized,
    reviewed: reviewState.reviewed,
    extraTag: reviewState.tagText,
    actions: actionList.map((action) => {
      if (action.key !== 'review') {
        return {
          ...action,
          icon: normalizeActionIcon(action)
        }
      }

      return {
        ...action,
        icon: normalizeActionIcon(action),
        text: reviewState.actionText,
        disabled: reviewState.reviewed
      }
    })
  }
}

function formatSummaryAmount(value, fallback, prefix = '') {
  if (fallback) {
    return fallback
  }
  if (typeof value === 'number') {
    return `${prefix}${value.toLocaleString()}`
  }

  return firstDefined(value, fallback)
}

Page({
  data: {
    pageTitle: '',
    pageConfig: normalizePageConfig(),
    texts: {},
    shellLayout: getShellLayoutStyles(),
    loading: false,
    errorText: '',
    summary: {
      label: '',
      amount: '',
      background: '',
      iconSrc: '',
      stats: []
    },
    tabs: [],
    activeTab: 'processing',
    visibleRecords: [],
    records: [],
    page: 1, pageSize: 5, total: 0, hasPrevious: false, hasMore: false
  },

  onLoad() {
    this.loadReferralRecords()
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

  onTabTap(event) {
    const key = event.currentTarget.dataset.key

    if (!key || key === this.data.activeTab) {
      return
    }

    this.setData({
      activeTab: key
    })
    this.loadReferralRecords({ page: 1, status: key })
  },

  updateVisibleRecords(activeTab) {
    this.applyRecords(this.data.records, activeTab)
  },

  async loadReferralRecords(options = {}) {
    this.setData({
      loading: true,
      errorText: ''
    })

    try {
      const data = await gameService.getReferralRecords({ page: options.page || 1, pageSize: 5, status: options.status || this.data.activeTab })
      const pageConfig = normalizePageConfig(data.pageConfig)
      const records = Array.isArray(data.records) ? data.records : []
      const tabs = Array.isArray(data.tabs) && data.tabs.length ? data.tabs : this.data.tabs
      const summary = data.summary ? {
        ...this.data.summary,
        ...data.summary,
        amount: formatSummaryAmount(
          data.summary.amount,
          data.summary.amountText || this.data.summary.amount,
          ''
        )
      } : this.data.summary

      this.setData({
        pageTitle: pageConfig.pageTitle || '',
        pageConfig,
        texts: pageConfig.texts,
        loading: false,
        errorText: '',
        summary,
        tabs,
        records,
        page: data.page || 1,
        total: data.total || 0,
        hasPrevious: Boolean(data.hasPrevious),
        hasMore: Boolean(data.hasMore)
      })
      this.applyRecords(records, this.data.activeTab)
    } catch (error) {
      this.setData({
        loading: false,
        errorText: toUserMessage(error && error.message, textOf(this.data.pageConfig, 'loadFailedText') || '引荐记录加载失败')
      })
    }
  },

  onPreviousPage() { if (this.data.hasPrevious) this.loadReferralRecords({ page: this.data.page - 1, status: this.data.activeTab }) },
  onNextPage() { if (this.data.hasMore) this.loadReferralRecords({ page: this.data.page + 1, status: this.data.activeTab }) },

  applyRecords(records, activeTab) {
    this.setData({
      visibleRecords: (records || [])
        .filter((item) => item.state === activeTab)
        .map((item) => normalizeRecord(item, this.data.pageConfig))
    })
  },

  onActionTap(event) {
    const action = event.currentTarget.dataset.action
    const recordId = event.currentTarget.dataset.recordId
    const disabled = event.currentTarget.dataset.disabled === true || event.currentTarget.dataset.disabled === 'true'

    if (disabled) {
      return
    }

    const record = this.data.visibleRecords.find((item) => String(item.id) === String(recordId)) || {}
    const gameId = firstDefined(record.gameId, record.id)

    if (action === 'chat') {
      navigateShellRoute(`/${ROUTES.imRoom}?gameId=${encodeURIComponent(gameId || '')}&prefill=${encodeURIComponent(textOf(this.data.pageConfig, 'chatPrefill'))}`)
      return
    }

    if (action === 'review') {
      const targetUserId = firstDefined(record.reviewTargetUserId, record.expertUserId, record.playerUserId)
      const params = [
        gameId ? `gameId=${encodeURIComponent(gameId)}` : '',
        targetUserId ? `targetUserId=${encodeURIComponent(targetUserId)}` : '',
        'role=guide'
      ].filter(Boolean).join('&')

      navigateShellRoute(`/${ROUTES.gameReview}${params ? `?${params}` : ''}`)
      return
    }

    if (action === 'remind') {
      this.remindDelivery(record)
      return
    }

    this.showToast(textOf(this.data.pageConfig, 'unavailableText'))
  },

  async remindDelivery(record = {}) {
    try {
      await gameService.sendGuideReminder({
        invitationId: firstDefined(record.invitationId, record.referralId, record.id),
        gameId: firstDefined(record.gameId, record.id),
        remindTarget: 'delivery',
        message: applyTemplate(textOf(this.data.pageConfig, 'remindMessageTemplate'), {
          serviceTitle: record.serviceTitle || textOf(this.data.pageConfig, 'remindServiceFallback')
        })
      })
      this.showToast(textOf(this.data.pageConfig, 'remindSuccessText'))
    } catch (error) {
      this.showToast(error.message || textOf(this.data.pageConfig, 'remindFailedText'))
    }
  },

  showToast(title) {
    if (!title) {
      return
    }

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

    navigateShellRoute(ROUTES.profile)
  }
})
