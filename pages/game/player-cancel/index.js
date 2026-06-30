const { ROUTES } = require('../../../config/routes')
const toast = require('../../../utils/toast')
const { getSurnameInitials } = require('../../../utils/avatar')

const CONTENT_LEFT_RPX = 0
const CONTENT_TOP_RPX = 160
const CONTENT_WIDTH_RPX = 750
const DESIGN_FRAME_HEIGHT_PT = 810
const DESIGN_BOTTOM_HEIGHT_PT = 78
const NAV_TITLE_HEIGHT_RPX = 50
const BACK_BUTTON_SIZE_RPX = 40
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = DESIGN_FRAME_HEIGHT_PT * 2

const REASON_OPTIONS = [
  { key: 'need_changed', text: '需求变更，不再需要服务' },
  { key: 'other_solution', text: '找到其他解决方案' },
  { key: 'service_unexpected', text: '业务主服务不符合预期' },
  { key: 'budget', text: '预算问题/资金紧张' }
]

const AGREEMENT_TEXT = '我已阅读并同意上述赔付协议，理解主动取消需承担行家的时间成本损失，并同意按设置比例从托管资金中赔付行家。'
const AGREEMENT_ITEMS = [
  '我理解主动取消需承担行家的时间成本损失',
  '我同意按设置比例赔付行家，金额从托管资金扣除',
  '剩余金额将在3个工作日内原路退回',
  '此取消记录将影响信用分（-3分）'
]
const EMPTY_AMOUNT_DETAIL = {
  contractAmount: 0,
  platformFeeRate: 0
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

  const amount = Number(String(value || '').replace(/[¥,\s]/g, ''))

  return Number.isFinite(amount) && amount >= 0 ? amount : fallback
}

function getRateNumber(value, fallback = 0) {
  if (value === undefined || value === null || value === '') {
    return fallback
  }

  const rate = Number(String(value || '').replace(/[%\s]/g, ''))

  return Number.isFinite(rate) && rate >= 0 ? rate : fallback
}

function getOptionText(options, keys, fallback = '') {
  for (const key of keys) {
    const text = decodeOption(options[key]).trim()

    if (text) {
      return text
    }
  }

  return fallback
}

function formatCurrency(value) {
  const amount = Number(value)

  if (!Number.isFinite(amount)) {
    return '¥0'
  }

  return `¥${amount.toLocaleString('zh-CN', {
    maximumFractionDigits: amount % 1 === 0 ? 0 : 2
  })}`
}

function clampRate(value, minRate, maxRate) {
  const rate = getRateNumber(value, NaN)

  if (!Number.isFinite(rate)) {
    return minRate
  }

  return Math.max(minRate, Math.min(maxRate, Math.round(rate)))
}

function normalizeRequiredText(value) {
  return String(value || '').trim()
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

function buildSmartSuggestion(detail) {
  return `行家已投入${detail.servedDurationText}服务，建议设置${detail.suggestionMinRate}%-${detail.suggestionMaxRate}%赔付比例，体现对行家时间成本的尊重。`
}

function getAvatarText(name, fallback) {
  return getSurnameInitials(name, fallback)
}

function hasRequiredCancelDetail(detail) {
  return Boolean(
    detail &&
    detail.serviceOrderId &&
    detail.expertName &&
    detail.serviceTitle &&
    detail.servedText &&
    Number.isFinite(detail.contractAmount) &&
    Number.isFinite(detail.minRate) &&
    Number.isFinite(detail.maxRate) &&
    detail.maxRate >= detail.minRate &&
    Number.isFinite(detail.suggestedRate) &&
    Number.isFinite(detail.platformFeeRate)
  )
}

function buildCancelDetail(options = {}) {
  const amount = getNumber(
    options.amount || options.contractAmount || options.amountText || options.contractAmountText,
    NaN
  )
  const playerName = decodeOption(options.playerName)
  const playerAvatarText = getAvatarText(playerName, decodeOption(options.playerAvatarText))
  const expertName = decodeOption(options.expertName)
  const avatarText = getAvatarText(expertName, decodeOption(options.expertAvatarText))
  const minRate = getRateNumber(options.minRate, NaN)
  const maxRate = getRateNumber(options.maxRate, NaN)
  const suggestedRate = clampRate(options.suggestedRate || options.rate, minRate, maxRate)
  const suggestionMinRate = clampRate(
    options.suggestionMinRate || options.recommendMinRate || suggestedRate,
    minRate,
    maxRate
  )
  const suggestionMaxRate = Math.max(suggestionMinRate, clampRate(
    options.suggestionMaxRate || options.recommendMaxRate || suggestedRate,
    minRate,
    maxRate
  ))
  const rawServedText = getOptionText(options, ['servedText', 'serviceDurationText'])
  const servedTextParts = splitServedText(rawServedText)
  const servedDurationText = getOptionText(
    options,
    ['servedDurationText', 'servedDuration', 'servedHoursText', 'servedTimeText'],
    servedTextParts.servedDurationText
  )
  const totalDurationText = getOptionText(
    options,
    ['totalDurationText', 'totalDuration', 'totalHoursText', 'serviceTotalDurationText'],
    servedTextParts.totalDurationText
  )
  const servedText = buildServedText(servedDurationText, totalDurationText) || rawServedText
  const detail = {
    serviceOrderId: decodeOption(options.serviceOrderId) || decodeOption(options.orderId),
    gameId: decodeOption(options.gameId),
    ref: decodeOption(options.ref),
    playerName,
    playerAvatarText,
    expertName,
    expertAvatarText: avatarText,
    serviceTitle: decodeOption(options.serviceTitle),
    roleLabel: '玩家',
    contractAmount: amount,
    contractAmountText: Number.isFinite(amount) ? formatCurrency(amount) : '',
    minRate,
    maxRate,
    suggestedRate,
    suggestionMinRate,
    suggestionMaxRate,
    platformFeeRate: getRateNumber(options.platformFeeRate, NaN),
    servedDurationText,
    totalDurationText,
    servedText,
    smartSuggestion: '',
    warningTitle: decodeOption(options.warningTitle) || '取消需承担赔付',
    warningDesc: decodeOption(options.warningDesc) || '作为玩家主动取消，需按约定比例赔付行家损失（补偿已投入的时间成本）。'
  }

  if (!hasRequiredCancelDetail(detail)) {
    return null
  }

  return {
    ...detail,
    smartSuggestion: buildSmartSuggestion(detail)
  }
}

function buildReasonOptions(activeKey = 'other_solution') {
  return REASON_OPTIONS.map((item) => ({
    ...item,
    active: item.key === activeKey
  }))
}

function buildAmountState(detail, rate) {
  const compensationAmount = Math.round(detail.contractAmount * rate) / 100
  const platformFee = Math.round(compensationAmount * detail.platformFeeRate) / 100
  const payAmount = Math.round((compensationAmount + platformFee) * 100) / 100
  const refundAmount = Math.max(0, Math.round((detail.contractAmount - payAmount) * 100) / 100)

  return {
    compensationRate: rate,
    compensationAmountText: formatCurrency(compensationAmount),
    platformFeeText: formatCurrency(platformFee),
    payAmountText: formatCurrency(payAmount),
    refundAmountText: formatCurrency(refundAmount)
  }
}

function buildActivityCard(detail) {
  return {
    title: '活动信息',
    avatarText: detail.expertAvatarText,
    avatarClass: 'expert',
    name: detail.expertName,
    roleLabel: '行家',
    roleClass: 'expert',
    serviceTitle: detail.serviceTitle,
    rows: [
      { label: '合同金额', value: detail.contractAmountText, tone: 'strong', divider: true },
      { label: '已服务时长', value: detail.servedText, tone: 'blue' }
    ]
  }
}

Page({
  data: {
    shellLayout: getWhiteShellLayoutStyles(),
    hasDetail: false,
    detail: {},
    activityCard: {},
    reasonOptions: buildReasonOptions(),
    selectedReasonKey: 'other_solution',
    reasonText: '',
    agreementChecked: true,
    agreementText: AGREEMENT_TEXT,
    agreementItems: AGREEMENT_ITEMS,
    submitting: false,
    ...buildAmountState(EMPTY_AMOUNT_DETAIL, 0)
  },

  onLoad(options = {}) {
    const detail = buildCancelDetail(options)

    if (!detail) {
      this.setData({
        shellLayout: getWhiteShellLayoutStyles(),
        hasDetail: false,
        detail: {},
        activityCard: {},
        ...buildAmountState(EMPTY_AMOUNT_DETAIL, 0)
      })
      return
    }

    const rate = clampRate(detail.suggestedRate, detail.minRate, detail.maxRate)

    this.setData({
      shellLayout: getWhiteShellLayoutStyles(),
      hasDetail: true,
      detail,
      activityCard: buildActivityCard(detail),
      ...buildAmountState(detail, rate)
    })
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

  onBackTap() {
    this.goBack()
  },

  onKeepServiceTap() {
    this.goBack()
  },

  goBack() {
    if (getCurrentPages().length > 1) {
      wx.navigateBack()
      return
    }

    wx.redirectTo({
      url: `/${ROUTES.gamePlayerManage}`
    })
  },

  onRateChanging(event) {
    if (!this.data.hasDetail) {
      return
    }

    this.updateCompensationRate(event.detail && event.detail.value)
  },

  onRateChange(event) {
    if (!this.data.hasDetail) {
      return
    }

    this.updateCompensationRate(event.detail && event.detail.value)
  },

  updateCompensationRate(value) {
    const detail = this.data.detail
    const rate = clampRate(value, detail.minRate, detail.maxRate)

    this.setData(buildAmountState(detail, rate))
  },

  onReasonChange(event) {
    const key = event.detail.selectedKey || event.detail.key

    this.setData({
      selectedReasonKey: key,
      reasonOptions: buildReasonOptions(key)
    })
  },

  onReasonInput(event) {
    this.setData({
      reasonText: event.detail.value
    })
  },

  onAgreementChange(event) {
    this.setData({
      agreementChecked: Boolean(event.detail.checked)
    })
  },

  onAgreementCardTap() {
    this.setData({
      agreementChecked: !this.data.agreementChecked
    })
  },

  onConfirmCancelTap() {
    if (!this.data.hasDetail) {
      toast.info('缺少服务信息，请从组局管理进入')
      return
    }

    const reason = this.data.reasonOptions.find((item) => item.active)
    const reasonText = normalizeRequiredText(this.data.reasonText)

    if (!reason) {
      toast.info('请选择取消原因')
      return
    }

    if (!reasonText) {
      toast.info('请填写详细说明')
      return
    }

    if (!this.data.agreementChecked) {
      toast.info('请先同意赔付协议')
      return
    }

    wx.showModal({
      title: '确认取消并赔付',
      content: `将按${this.data.compensationRate}%赔付行家，实际支付${this.data.payAmountText}。确认后本次服务将进入取消流程。`,
      cancelText: '再想想',
      confirmText: '确认取消',
      confirmColor: '#ff4d4f',
      success: (res) => {
        if (!res.confirm) {
          return
        }

        this.submitCancelRequest(reason)
      }
    })
  },

  submitCancelRequest(reason) {
    if (this.data.submitting) {
      return
    }

    this.setData({
      submitting: true
    })
    this.setData({
      submitting: false
    })
    toast.info('取消提交接口待接入')
  }
})
