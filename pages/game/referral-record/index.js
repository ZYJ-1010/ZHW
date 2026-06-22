const { ROUTES } = require('../../../config/routes')

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

function firstDefined(...values) {
  return values.find((value) => value !== undefined && value !== null && value !== '')
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
      ? firstDefined(record.reviewedTagText, '已评价')
      : firstDefined(record.extraTag, record.reviewTagText, '待评价'),
    actionText: reviewed
      ? firstDefined(record.reviewedActionText, record.reviewActionText, '已评价')
      : firstDefined(record.reviewActionText, '评价双方')
  }
}

function normalizeRecord(record = {}) {
  const normalized = {
    ...record,
    rewardClass: getRewardClass(record)
  }
  const actionList = Array.isArray(record.actions) ? record.actions : (record.actionList || [])

  if (record.state !== 'completed') {
    return {
      ...normalized,
      actions: actionList
    }
  }

  const reviewState = getReviewState(record)

  return {
    ...normalized,
    reviewed: reviewState.reviewed,
    extraTag: reviewState.tagText,
    actions: actionList.map((action) => {
      if (action.key !== 'review') {
        return action
      }

      return {
        ...action,
        text: reviewState.actionText,
        disabled: reviewState.reviewed
      }
    })
  }
}

Page({
  data: {
    pageTitle: '我的引荐记录',
    shellLayout: getShellLayoutStyles(),
    summary: {
      label: '本月引荐收益',
      amount: '¥2,450',
      background: 'linear-gradient(135deg, #ffb347 0%, #ff7b00 100%)',
      iconSrc: `${ASSET_BASE}/wallet.png`,
      stats: [
        { key: 'success', text: '成功 12单' },
        { key: 'processing', text: '进行中 3单' },
        { key: 'review', text: '待评价 2单' }
      ]
    },
    tabs: [
      { key: 'processing', label: '进行中', count: 3 },
      { key: 'completed', label: '已完成', count: 12 },
      { key: 'canceled', label: '已取消', count: 2 }
    ],
    activeTab: 'processing',
    visibleRecords: [],
    records: [
      {
        id: 'REF-20260320-001',
        state: 'processing',
        stateText: '服务进行中',
        stateIcon: `${ASSET_BASE}/status-processing.png`,
        stateClass: 'blue',
        timeText: '3天前',
        expert: {
          name: '张专家',
          avatarText: 'ZH',
          avatarClass: 'blue'
        },
        player: {
          name: '李明',
          avatarText: 'LI',
          avatarClass: 'pink'
        },
        matchIcon: `${ASSET_BASE}/handshake.png`,
        serviceTitle: '产品架构咨询',
        reward: '80',
        rewardClass: 'orange',
        noticeText: '预计交付时间：2026-03-25（剩余2天）',
        actions: [
          { key: 'remind', text: '提醒交付', icon: `${ASSET_BASE}/bell.png`, theme: 'primary' },
          { key: 'chat', text: '查看群聊', icon: `${ASSET_BASE}/chat.png`, theme: 'plain' }
        ]
      },
      {
        id: 'REF-20260315-002',
        state: 'completed',
        stateText: '已完成',
        stateIcon: `${ASSET_BASE}/status-completed.png`,
        stateClass: 'green',
        extraTag: '待评价',
        reviewStatus: 'pending',
        cardClass: 'review-needed',
        expert: {
          name: '陈工',
          avatarText: 'CH',
          avatarClass: 'green'
        },
        player: {
          name: '王总',
          avatarText: 'WA',
          avatarClass: 'purple'
        },
        matchIcon: `${ASSET_BASE}/status-completed.png`,
        serviceTitle: 'UI设计服务',
        reward: '120',
        rewardClass: 'green',
        actions: [
          { key: 'review', text: '评价双方', theme: 'highlight' }
        ]
      },
      {
        id: 'REF-20260310-003',
        state: 'canceled',
        stateText: '已取消',
        stateIcon: `${ASSET_BASE}/status-canceled.png`,
        stateClass: 'gray',
        expert: {
          name: '刘设计师',
          avatarText: 'LI',
          avatarClass: 'gray'
        },
        player: {
          name: '赵客户',
          avatarText: 'ZH',
          avatarClass: 'gray'
        },
        matchIcon: `${ASSET_BASE}/status-canceled.png`,
        serviceTitle: '品牌视觉咨询',
        reward: '0',
        rewardClass: 'gray',
        avatarText: 'LI',
        cancelReason: '行家时间冲突',
        noticeText: '取消原因：行家时间冲突',
        actions: []
      }
    ]
  },

  onLoad() {
    this.updateVisibleRecords(this.data.activeTab)
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
    this.updateVisibleRecords(key)
  },

  updateVisibleRecords(activeTab) {
    this.setData({
      visibleRecords: this.data.records
        .filter((item) => item.state === activeTab)
        .map(normalizeRecord)
    })
  },

  onActionTap(event) {
    const action = event.currentTarget.dataset.action
    const disabled = event.currentTarget.dataset.disabled === true || event.currentTarget.dataset.disabled === 'true'

    if (disabled) {
      return
    }

    const actionTextMap = {
      remind: '已提醒交付',
      chat: '群聊页待接入',
      review: '评价页待接入'
    }

    this.showToast(actionTextMap[action] || '功能待接入')
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
