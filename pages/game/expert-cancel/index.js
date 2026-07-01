const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const { getSurnameInitials } = require('../../../utils/avatar')

const CONTENT_LEFT_RPX = 0
const CONTENT_TOP_RPX = 160
const CONTENT_WIDTH_RPX = 750
const DESIGN_FRAME_HEIGHT_PT = 810
const DESIGN_BOTTOM_HEIGHT_PT = 78
const NAV_TITLE_HEIGHT_RPX = 50
const BACK_BUTTON_SIZE_RPX = 40
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = 1620
const EMPTY_ACTIVITY = {
  serviceOrderId: '',
  gameId: '',
  playerId: '',
  playerName: '',
  avatarText: '',
  serviceTitle: '',
  amount: 0,
  amountText: '',
  platformFeeRate: 0,
  platformFee: null,
  platformFeeText: '',
  statusText: ''
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

function pickFirstValue() {
  const values = Array.prototype.slice.call(arguments)

  for (let index = 0; index < values.length; index += 1) {
    if (values[index] !== undefined && values[index] !== null && values[index] !== '') {
      return values[index]
    }
  }

  return ''
}

function normalizeBoolean(value) {
  if (typeof value === 'boolean') {
    return value
  }

  if (typeof value === 'string') {
    return value === 'true' || value === '1'
  }

  return Boolean(value)
}

function normalizeList(list) {
  return Array.isArray(list) ? list : []
}

function parseAmountText(value) {
  const text = String(value || '').replace(/[^\d.]/g, '')
  const amount = Number(text)

  return Number.isFinite(amount) ? amount : 0
}

function decodeOption(value) {
  if (value === undefined || value === null) {
    return ''
  }

  try {
    return decodeURIComponent(String(value))
  } catch (error) {
    return String(value)
  }
}

function getNumber(value, fallback = 0) {
  if (value === undefined || value === null || value === '') {
    return fallback
  }

  const number = Number(String(value).replace(/[¥,\s]/g, ''))

  return Number.isFinite(number) ? number : fallback
}

function formatCurrency(value) {
  const amount = Number(value)

  if (!Number.isFinite(amount)) {
    return ''
  }

  return `¥${Math.round(amount).toLocaleString('zh-CN')}`
}

function normalizeAvatarText(value, name) {
  const text = decodeOption(value).trim()

  return text || (name ? getSurnameInitials(name, '') : '')
}

function normalizeActivity(source = {}) {
  const player = source.player || {}
  const service = source.service || {}
  const playerName = decodeOption(pickFirstValue(player.name, source.playerName))
  const amountText = decodeOption(pickFirstValue(service.amountText, service.contractAmountText, source.amountText, source.contractAmountText))
  const amount = getNumber(pickFirstValue(service.amount, service.contractAmount, source.amount), parseAmountText(amountText))

  return {
    serviceOrderId: pickFirstValue(source.serviceOrderId, source.orderId, source.id),
    gameId: pickFirstValue(source.gameId),
    playerId: pickFirstValue(player.id, player.userId, source.playerId),
    playerName,
    avatarText: normalizeAvatarText(pickFirstValue(player.avatarText, source.avatarText), playerName),
    serviceTitle: decodeOption(pickFirstValue(service.title, source.serviceTitle, source.title)),
    amount,
    amountText: amountText || (amount ? formatCurrency(amount) : ''),
    platformFeeRate: getNumber(pickFirstValue(source.platformFeeRate, source.serviceFeeRate), 0),
    platformFee: source.platformFee || source.serviceFee || null,
    platformFeeText: decodeOption(pickFirstValue(source.platformFeeText, source.serviceFeeText)),
    statusText: decodeOption(pickFirstValue(source.statusText, source.serviceStatusText))
  }
}

function buildActivityCard(activity) {
  return {
    title: '活动信息',
    avatarText: activity.avatarText,
    avatarClass: 'player',
    name: activity.playerName,
    roleLabel: '玩家',
    roleClass: 'player',
    serviceTitle: activity.serviceTitle,
    rows: [
      { label: '合同金额', value: activity.amountText, tone: 'strong', divider: true },
      { label: '服务状态', value: activity.statusText, tone: 'active' }
    ].filter((item) => item.value)
  }
}

function normalizeReasons(data) {
  return normalizeList(data.reasonOptions || data.cancelReasons || data.reasons)
    .map((item) => ({
      key: pickFirstValue(item.key, item.code, item.id),
      text: pickFirstValue(item.text, item.label, item.title)
    }))
    .filter((item) => item.key && item.text)
}

function normalizePreview(data = {}, query = {}) {
  const activity = normalizeActivity(Object.assign({}, query, data))
  const reasonOptions = normalizeReasons(data)
  const selectedReasonKey = pickFirstValue(data.selectedReasonKey, data.defaultReasonKey, reasonOptions[0] && reasonOptions[0].key)
  const compensationRatio = getNumber(
    pickFirstValue(data.compensationRatio, data.compensationRate, data.defaultCompensationRate, data.suggestedRate),
    0
  )

  return {
    activity,
    activityCard: buildActivityCard(activity),
    compensationRatio,
    compensationAmountText: pickFirstValue(data.compensationAmountText),
    serviceFeeText: pickFirstValue(data.serviceFeeText, data.platformFeeText),
    totalDebitText: pickFirstValue(data.totalDebitText, data.totalDebitAmountText),
    reasonOptions,
    selectedReasonKey,
    agreementChecked: normalizeBoolean(data.agreementChecked),
    agreementText: pickFirstValue(data.agreementText, data.cancelAgreementText)
  }
}

function normalizeRequiredText(value) {
  return String(value || '').trim()
}

Page({
  data: {
    shellLayout: getWhiteShellLayoutStyles(),
    queryParams: {},
    activity: EMPTY_ACTIVITY,
    activityCard: buildActivityCard(EMPTY_ACTIVITY),
    compensationRatio: 0,
    compensationAmountText: '',
    serviceFeeText: '',
    totalDebitText: '',
    reasonOptions: [],
    selectedReasonKey: '',
    reasonDetail: '',
    agreementChecked: false,
    agreementText: ''
  },

  onLoad(options = {}) {
    this.setData({
      queryParams: options
    })
    this.loadCancelPreview(options)
    this.updateShellLayout()
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

  async loadCancelPreview(params = {}) {
    const serviceOrderId = params.serviceOrderId || params.orderId || params.id || this.data.activity.serviceOrderId

    if (!serviceOrderId) {
      this.setData(normalizePreview({}, params))
      this.showInfo('缺少服务订单信息')
      return
    }

    try {
      const preview = await gameService.getExpertCancelPreview(Object.assign({}, params, {
        serviceOrderId
      }))

      this.setData(normalizePreview(preview, params))
    } catch (error) {
      this.setData(normalizePreview({}, params))
      this.showInfo(error.message || '取消赔付预览加载失败')
    }
  },

  onRatioChanging(event) {
    this.setData({
      compensationRatio: Number(event.detail.value)
    })
  },

  onRatioChange(event) {
    const compensationRatio = Number(event.detail.value)

    this.setData({
      compensationRatio
    })
    this.loadCancelPreview(Object.assign({}, this.data.queryParams, {
      serviceOrderId: this.data.activity.serviceOrderId,
      compensationRate: compensationRatio
    }))
  },

  onReasonChange(event) {
    const selectedReasonKey = event.detail.selectedKey || event.detail.key

    this.setData({
      selectedReasonKey
    })
  },

  onReasonInput(event) {
    const reasonDetail = event.detail.value

    this.setData({
      reasonDetail
    })
  },

  onAgreementChange(event) {
    const agreementChecked = normalizeBoolean(event.detail.checked)

    this.setData({
      agreementChecked
    })
  },

  async onConfirmCancelTap() {
    const reasonDetail = normalizeRequiredText(this.data.reasonDetail)

    if (!this.data.selectedReasonKey) {
      this.showInfo('请选择取消原因')
      return
    }

    if (!reasonDetail) {
      this.showInfo('请填写详细说明')
      return
    }

    if (!this.data.agreementChecked) {
      this.showInfo('请先同意服务取消协议')
      return
    }

    try {
      await gameService.cancelServiceWithCompensation({
        serviceOrderId: this.data.activity.serviceOrderId,
        gameId: this.data.activity.gameId,
        reasonCode: this.data.selectedReasonKey,
        reasonRemark: reasonDetail,
        compensationRate: this.data.compensationRatio
      })
      this.showInfo('取消赔付已提交')
      this.goBack()
    } catch (error) {
      this.showInfo(error.message || '取消赔付提交失败')
    }
  },

  onReconsiderTap() {
    this.goBack()
  },

  onBackTap() {
    this.goBack()
  },

  goBack() {
    const pages = getCurrentPages()

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    wx.navigateTo({
      url: `/${ROUTES.gameManage}`
    })
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
