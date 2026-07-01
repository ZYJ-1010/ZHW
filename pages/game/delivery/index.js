const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
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
const PLAYER_ICON = '/pages/game/delivery/assets/i18@3x.png'

const EMPTY_DELIVERY_STATE = {
  deliveryMode: '',
  pageTitle: '确认服务完成',
  status: {},
  statePill: {},
  activity: {},
  activityRows: [],
  notice: {},
  settlement: {},
  settlementRows: [],
  hasSettlement: false,
  settlementNote: '',
  timeline: [],
  confirmItems: [],
  allConfirmed: false,
  confirmNote: '',
  security: {},
  submitHints: {},
  submitToast: '',
  quickActions: []
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

function pickFirstValue() {
  const values = Array.prototype.slice.call(arguments)

  for (let index = 0; index < values.length; index += 1) {
    if (values[index] !== undefined && values[index] !== null && values[index] !== '') {
      return values[index]
    }
  }

  return ''
}

function normalizeList(list) {
  return Array.isArray(list) ? list : []
}

function normalizeDeliveryMode(value) {
  const mode = String(value || '').toLowerCase()

  return mode === 'free' ? 'free' : mode === 'paid' ? 'paid' : ''
}

function normalizeActivity(data) {
  const source = data || {}
  const expert = source.expert || {}
  const service = source.service || {}
  const expertName = pickFirstValue(expert.name, source.expertName)

  return {
    avatar: pickFirstValue(expert.avatarText, source.avatarText, expertName ? getSurnameInitials(expertName, '') : ''),
    expertName,
    serviceName: pickFirstValue(service.title, source.serviceName, source.serviceTitle),
    orderNo: pickFirstValue(source.orderNo, source.orderNoText, source.serviceOrderNo)
  }
}

function normalizeActivityRows(data) {
  const source = data || {}
  const service = source.service || {}
  const rows = normalizeList(source.activityRows || source.infoRows)

  if (rows.length) {
    return rows.map((item) => ({
      label: pickFirstValue(item.label, item.title),
      value: pickFirstValue(item.value, item.text),
      strong: Boolean(item.strong),
      type: pickFirstValue(item.type, item.tone)
    })).filter((item) => item.label || item.value)
  }

  return [
    { label: '合同金额', value: pickFirstValue(service.contractAmountText, source.contractAmountText), strong: true },
    { label: '服务时长', value: pickFirstValue(service.durationText, source.durationText) },
    { label: '开始时间', value: pickFirstValue(service.startedAtText, source.startedAtText) },
    { label: '完成时间', value: pickFirstValue(service.completedAtText, source.completedAtText) }
  ].filter((item) => item.value)
}

function normalizeSettlementRows(settlement) {
  const source = settlement || {}
  const rows = normalizeList(source.rows || source.items)

  if (rows.length) {
    return rows.map((item) => ({
      label: pickFirstValue(item.label, item.title),
      value: pickFirstValue(item.value, item.text),
      type: pickFirstValue(item.type, item.tone)
    })).filter((item) => item.label || item.value)
  }

  return [
    { label: '合同总金额', value: source.contractAmountText },
    { label: '结算比例', value: source.settlementRatioText, type: 'success' },
    { label: '结算金额', value: source.settlementAmountText, type: 'amount' },
    { label: '平台服务费', value: source.platformFeeText, type: 'danger' },
    { label: '领路人奖励', value: source.guideRewardText, type: 'warning' },
    { label: '系统级领路人奖励', value: source.systemGuideRewardText, type: 'purple' }
  ].filter((item) => item.value)
}

function normalizeNotice(notice) {
  const source = notice || {}

  return {
    iconText: pickFirstValue(source.iconText),
    title: pickFirstValue(source.title),
    parts: normalizeList(source.parts || source.contents).map((item) => ({
      text: pickFirstValue(item.text, item.value),
      strong: Boolean(item.strong)
    })).filter((item) => item.text)
  }
}

function normalizeTimeline(data) {
  return normalizeList(data).map((item) => ({
    key: pickFirstValue(item.key, item.id),
    title: pickFirstValue(item.title, item.name),
    desc: pickFirstValue(item.desc, item.description),
    timeText: pickFirstValue(item.timeText, item.time, item.createdAtText),
    state: pickFirstValue(item.state, item.status),
    hasLine: item.hasLine !== false,
    actionText: pickFirstValue(item.actionText, item.actionTitle),
    actionIconText: pickFirstValue(item.actionIconText),
    actionPlacement: pickFirstValue(item.actionPlacement),
    actionKey: pickFirstValue(item.actionKey, item.key, item.id)
  })).filter((item) => item.title || item.desc)
}

function normalizeConfirmItems(data) {
  return normalizeList(data).map((item) => ({
    id: pickFirstValue(item.id, item.key),
    title: pickFirstValue(item.title, item.name),
    desc: pickFirstValue(item.desc, item.description),
    checked: Boolean(item.checked),
    locked: Boolean(item.locked || item.disabled)
  })).filter((item) => item.id || item.title)
}

function normalizeQuickActions(data) {
  return normalizeList(data).map((item) => {
    const key = pickFirstValue(item.key, item.id, item.action)
    const targetRole = pickFirstValue(item.targetRole, item.role)

    return {
      key,
      targetRole,
      title: pickFirstValue(item.title, item.name),
      theme: pickFirstValue(item.theme, targetRole === 'guide' ? 'orange' : 'blue'),
      iconSrc: pickFirstValue(item.iconSrc, item.localIcon, targetRole === 'player' ? PLAYER_ICON : ''),
      iconText: pickFirstValue(item.iconText),
      route: pickFirstValue(item.route, item.path),
      message: pickFirstValue(item.message, item.toastText)
    }
  }).filter((item) => item.key || item.title)
}

function normalizeDeliveryData(data = {}) {
  const settlement = data.settlement || {}
  const confirmItems = normalizeConfirmItems(data.confirmItems || data.confirmations)
  const settlementRows = normalizeSettlementRows(settlement)

  return Object.assign({}, EMPTY_DELIVERY_STATE, {
    deliveryMode: normalizeDeliveryMode(data.serviceType || data.deliveryMode || data.mode),
    pageTitle: pickFirstValue(data.pageTitle, data.title, EMPTY_DELIVERY_STATE.pageTitle),
    status: {
      theme: pickFirstValue(data.statusTheme, data.status && data.status.theme, data.serviceType),
      title: pickFirstValue(data.statusTitle, data.status && data.status.title),
      desc: pickFirstValue(data.statusDesc, data.status && data.status.desc)
    },
    statePill: {
      theme: pickFirstValue(data.stateTheme, data.statePill && data.statePill.theme),
      text: pickFirstValue(data.stateText, data.statusText, data.statePill && data.statePill.text)
    },
    activity: normalizeActivity(data),
    activityRows: normalizeActivityRows(data),
    notice: normalizeNotice(data.notice),
    settlement: {
      actualAmount: pickFirstValue(settlement.actualAmountText, data.actualAmountText)
    },
    settlementRows,
    hasSettlement: settlementRows.length > 0,
    settlementNote: pickFirstValue(settlement.note, data.settlementNote),
    timeline: normalizeTimeline(data.timeline || data.steps),
    confirmItems,
    allConfirmed: confirmItems.length > 0 && confirmItems.every((item) => item.checked),
    confirmNote: pickFirstValue(data.confirmNote),
    security: data.security || {},
    submitHints: data.submitHints || {},
    submitToast: pickFirstValue(data.submitToast, data.submitMessage),
    quickActions: normalizeQuickActions(data.quickActions || data.actionsList || data.contactActions)
  })
}

Page({
  data: Object.assign({
    shellLayout: getWhiteShellLayoutStyles(),
    queryParams: {},
    serviceOrderId: '',
    gameId: ''
  }, EMPTY_DELIVERY_STATE),

  onLoad(options = {}) {
    const serviceOrderId = options.serviceOrderId || options.orderId || options.id || ''

    this.setData({
      queryParams: options,
      serviceOrderId,
      gameId: options.gameId || ''
    })
    this.loadDeliveryDetail(options)
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

  async loadDeliveryDetail(params = {}) {
    const serviceOrderId = params.serviceOrderId || params.orderId || params.id || this.data.serviceOrderId

    if (!serviceOrderId) {
      this.showInfo('缺少服务订单信息')
      this.setData(EMPTY_DELIVERY_STATE)
      return
    }

    try {
      const detail = await gameService.getServiceDeliveryDetail(Object.assign({}, params, {
        serviceOrderId
      }))

      this.setData(Object.assign(normalizeDeliveryData(detail), {
        serviceOrderId,
        gameId: detail && detail.gameId || this.data.gameId
      }))
    } catch (error) {
      this.setData(EMPTY_DELIVERY_STATE)
      this.showInfo(error.message || '服务交付详情加载失败')
    }
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

  async onTimelineActionTap(event) {
    const item = event.detail && event.detail.item

    if (!item) {
      return
    }

    if (item.route) {
      wx.navigateTo({ url: item.route })
      return
    }

    if (item.actionKey === 'waiting-player' || item.actionKey === 'remindPlayer' || item.key === 'waiting-player') {
      await this.remindPlayer()
      return
    }

    this.showInfo(item.message || '操作待接入')
  },

  async remindPlayer() {
    try {
      await gameService.remindPlayerConfirm({
        serviceOrderId: this.data.serviceOrderId,
        gameId: this.data.gameId
      })
      this.showInfo('已提醒玩家确认')
      this.loadDeliveryDetail(this.data.queryParams)
    } catch (error) {
      this.showInfo(error.message || '提醒玩家确认失败')
    }
  },

  onQuickActionTap(event) {
    const { title } = event.currentTarget.dataset
    const action = this.data.quickActions.find((item) => item.title === title)

    if (action && action.route) {
      wx.navigateTo({ url: action.route })
      return
    }

    this.showInfo(action && action.message || `${title || '操作'}待接入`)
  },

  toggleConfirm(event) {
    const { id } = event.currentTarget.dataset
    const confirmItems = this.data.confirmItems.map((item) => {
      if (item.id !== id || item.locked) {
        return item
      }

      return Object.assign({}, item, {
        checked: !item.checked
      })
    })
    const allConfirmed = confirmItems.length > 0 && confirmItems.every((item) => item.checked)

    this.setData({
      confirmItems,
      allConfirmed
    })
  },

  async onSubmitTap() {
    if (!this.data.allConfirmed) {
      this.showInfo('请先勾选全部确认项')
      return
    }

    const confirmedItems = this.data.confirmItems
      .filter((item) => item.checked)
      .map((item) => item.id)

    try {
      const result = await gameService.confirmServiceDelivery({
        serviceOrderId: this.data.serviceOrderId,
        gameId: this.data.gameId,
        confirmedItems
      })

      if (result && (result.timeline || result.status || result.service)) {
        this.setData(Object.assign(normalizeDeliveryData(result), {
          serviceOrderId: this.data.serviceOrderId,
          gameId: result.gameId || this.data.gameId
        }))
      }

      this.showInfo(this.data.submitToast || '服务完成确认已提交')
    } catch (error) {
      this.showInfo(error.message || '确认服务完成失败')
    }
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
