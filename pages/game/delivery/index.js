const { ROUTES } = require('../../../config/routes')
const { getSurnameInitials } = require('../../../utils/avatar')

const CONTENT_LEFT_RPX = 2
const CONTENT_TOP_RPX = 160
const CONTENT_WIDTH_RPX = 750
const DESIGN_FRAME_HEIGHT_PT = 810
const DESIGN_BOTTOM_HEIGHT_PT = 78
const NAV_TITLE_HEIGHT_RPX = 50
const BACK_BUTTON_SIZE_RPX = 40
const MORE_BUTTON_SIZE_RPX = 44
const MORE_BUTTON_LEFT_RPX = 658
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = DESIGN_FRAME_HEIGHT_PT * 2
const DEFAULT_DELIVERY_MODE = 'paid'

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
  const moreTop = Math.max(0, roundRpx(capsuleBottom - MORE_BUTTON_SIZE_RPX - 3))

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
    moreStyle: `left: ${MORE_BUTTON_LEFT_RPX}rpx; top: ${moreTop}rpx; width: ${MORE_BUTTON_SIZE_RPX}rpx; height: ${MORE_BUTTON_SIZE_RPX}rpx;`
  }
}

const DELIVERY_MODE_CONFIGS = {
  paid: {
    pageTitle: '确认服务完成',
    status: {
      theme: 'paid',
      title: '服务已完成!',
      desc: '双方确认后，资金将全额结算'
    },
    statePill: {
      theme: 'green',
      text: '待确认完成'
    },
    activity: {
      avatar: getSurnameInitials('李明', 'LI'),
      expertName: '李明',
      serviceName: '产品架构咨询',
      orderNo: 'REF-20260320-001'
    },
    activityRows: [
      { label: '合同金额', value: '¥800.00', strong: true },
      { label: '服务时长', value: '2小时 (已完成)' },
      { label: '开始时间', value: '03-20 14:30' },
      { label: '完成时间', value: '03-20 16:30' }
    ],
    notice: {},
    settlement: {
      actualAmount: '¥192.00'
    },
    settlementRows: [
      { label: '合同总金额', value: '¥800.00' },
      { label: '结算比例', value: '60% (扣除成本40%)', type: 'success' },
      { label: '结算金额', value: '¥480.00', type: 'amount' },
      { label: '平台服务费 (10%)', value: '-¥48.00', type: 'danger' },
      { label: '领路人奖励 (40%)', value: '-¥192.00', type: 'warning' },
      { label: '系统级领路人奖励 (10%)', value: '-¥48.00', type: 'purple' }
    ],
    settlementNote: '正常交付无扣减：服务按时完成，全额结算',
    timeline: [
      {
        key: 'completed',
        title: '服务已完成',
        desc: '行家标记服务已完成',
        timeText: '03-20 16:35',
        state: 'done',
        hasLine: true
      },
      {
        key: 'waiting-player',
        title: '等待玩家确认',
        desc: '需李明确认服务已达标',
        state: 'active',
        hasLine: true,
        actionText: '提醒确认',
        actionIconText: '🔔',
        actionPlacement: 'head'
      },
      {
        key: 'settlement',
        title: '资金结算',
        desc: '双方确认后自动到账',
        state: 'pending',
        hasLine: false
      }
    ],
    confirmItems: [
      {
        id: 'completed',
        title: '服务已全部完成',
        desc: '约定的2小时咨询服务已完整交付',
        checked: false
      },
      {
        id: 'qualified',
        title: '服务质量达标',
        desc: '需求方对服务内容和质量无异议',
        checked: false
      },
      {
        id: 'communicated',
        title: '双方已沟通确认',
        desc: '已与需求方确认服务完成，对方同意结算',
        checked: false
      }
    ],
    confirmNote: '正常交付无需扣减任何费用，只需双方确认服务已完成，资金将按全额结算。如服务未完全达标，请与玩家沟通后再确认。',
    security: {
      title: '资金安全保障',
      desc: '资金已托管，双方确认后自动结算，无需担心'
    },
    submitHints: {
      ready: '确认后将通知玩家进行最终确认',
      pending: '需勾选上方确认项后方可提交'
    },
    submitToast: '服务完成确认待接入'
  },
  free: {
    pageTitle: '确认服务完成',
    status: {
      theme: 'free',
      title: '服务已完成!',
      desc: '双方确认后，服务正式结束'
    },
    statePill: {
      theme: 'blue',
      text: '待确认完成'
    },
    activity: {
      avatar: getSurnameInitials('李明', 'LI'),
      expertName: '李明',
      serviceName: '产品架构咨询',
      orderNo: 'REF-20260320-001'
    },
    activityRows: [
      { label: '服务类型', value: '免费局', type: 'blue' },
      { label: '服务时长', value: '6小时（已完成）' },
      { label: '开始时间', value: '03-20 14:30' },
      { label: '完成时间', value: '03-20 20:30' }
    ],
    notice: {
      iconText: '🎁',
      title: '免费局说明',
      parts: [
        { text: '本局为' },
        { text: '免费体验局', strong: true },
        { text: '不涉及资金结算。双方确认完成后，行家将获得' },
        { text: '信用积分+5和免费局贡献徽章', strong: true },
        { text: '，玩家' },
        { text: '优先推荐权益', strong: true }
      ]
    },
    settlement: {},
    settlementRows: [],
    settlementNote: '',
    timeline: [
      {
        key: 'completed',
        title: '服务已完成',
        desc: '行家标记服务已完成',
        timeText: '03-20 16:35',
        state: 'done',
        hasLine: true
      },
      {
        key: 'waiting-player',
        title: '等待玩家确认',
        desc: '需李明确认服务已达标',
        state: 'active',
        hasLine: true,
        actionText: '提醒确认',
        actionIconText: '🔔',
        actionPlacement: 'head'
      },
      {
        key: 'archive',
        title: '服务归档',
        desc: '双方确认后自动归档',
        state: 'pending',
        hasLine: false
      }
    ],
    confirmItems: [
      {
        id: 'completed',
        title: '服务已全部完成',
        desc: '约定的2小时咨询服务已完整交付',
        checked: true,
        locked: true
      },
      {
        id: 'qualified',
        title: '服务质量达标',
        desc: '需求方对服务内容和质量无异议',
        checked: true,
        locked: true
      },
      {
        id: 'communicated',
        title: '双方已沟通确认',
        desc: '已与需求方确认服务完成，对方同意归档',
        checked: false
      }
    ],
    confirmNote: '免费局无需扣除任何费用，只需双方确认服务已完成，系统将自动归档。如服务未完全达标，请与玩家沟通后再次确认。',
    security: {
      title: '服务保障',
      desc: '免费局同样享受平台服务保障，评价真实有效'
    },
    submitHints: {
      ready: '确认后将通知玩家进行最终确认',
      pending: '需勾选上方确认项后方可提交'
    },
    submitToast: '免费局服务完成确认待接入'
  }
}

function cloneList(list) {
  return list.map((item) => ({ ...item }))
}

function cloneNotice(notice) {
  if (!notice || !notice.title) {
    return {}
  }

  return {
    ...notice,
    parts: cloneList(notice.parts || [])
  }
}

function normalizeDeliveryMode(mode) {
  return String(mode || '').toLowerCase() === 'free' ? 'free' : 'paid'
}

function createDeliveryState(mode) {
  const deliveryMode = normalizeDeliveryMode(mode)
  const config = DELIVERY_MODE_CONFIGS[deliveryMode]
  const confirmItems = cloneList(config.confirmItems)

  return {
    deliveryMode,
    pageTitle: config.pageTitle,
    status: { ...config.status },
    statePill: { ...config.statePill },
    activity: { ...config.activity },
    activityRows: cloneList(config.activityRows),
    notice: cloneNotice(config.notice),
    settlement: { ...config.settlement },
    settlementRows: cloneList(config.settlementRows),
    hasSettlement: config.settlementRows.length > 0,
    settlementNote: config.settlementNote,
    timeline: cloneList(config.timeline),
    confirmItems,
    allConfirmed: confirmItems.every((item) => item.checked),
    confirmNote: config.confirmNote,
    security: { ...config.security },
    submitHints: { ...config.submitHints },
    submitToast: config.submitToast,
    quickActions: [
      { title: '联系玩家', theme: 'blue', iconSrc: '/pages/game/delivery/assets/i18@3x.png' },
      { title: '联系领路人', theme: 'orange', iconText: '👬' }
    ]
  }
}

Page({
  data: {
    shellLayout: getWhiteShellLayoutStyles(),
    ...createDeliveryState(DEFAULT_DELIVERY_MODE)
  },

  onLoad(options = {}) {
    this.applyDeliveryMode(options.mode)
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

  applyDeliveryMode(mode) {
    this.setData(createDeliveryState(mode || DEFAULT_DELIVERY_MODE))
  },

  onBackTap() {
    const pages = getCurrentPages()

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    wx.navigateTo({
      url: `/${ROUTES.gamePlayerManage || 'pages/game/player-manage/index'}`
    })
  },

  onMoreTap() {
    this.showInfo('更多操作待接入')
  },

  onTimelineActionTap(event) {
    const item = event.detail && event.detail.item

    if (item && item.key === 'waiting-player') {
      this.showInfo('已提醒玩家确认')
      return
    }

    this.showInfo('操作待接入')
  },

  onQuickActionTap(event) {
    const { title } = event.currentTarget.dataset
    this.showInfo(`${title || '操作'}待接入`)
  },

  toggleConfirm(event) {
    const { id } = event.currentTarget.dataset
    const confirmItems = this.data.confirmItems.map((item) => {
      if (item.id !== id || item.locked) {
        return item
      }

      return {
        ...item,
        checked: !item.checked
      }
    })
    const allConfirmed = confirmItems.every((item) => item.checked)

    this.setData({
      confirmItems,
      allConfirmed
    })
  },

  onSubmitTap() {
    if (!this.data.allConfirmed) {
      this.showInfo('请先勾选全部确认项')
      return
    }

    this.showInfo(this.data.submitToast || '服务完成确认待接入')
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
