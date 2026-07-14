const { getLegacyWhiteFrameLayoutStyles } = require('../../../utils/adaptive-shell-layout')
const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const toast = require('../../../utils/toast')
const { getSurnameInitials } = require('../../../utils/avatar')
const { navigateShellRoute } = require('../../../utils/shell-nav')

const CONTENT_LEFT_RPX = 0
const CONTENT_TOP_RPX = 160
const CONTENT_WIDTH_RPX = 750
const DESIGN_FRAME_HEIGHT_PT = 810
const DESIGN_BOTTOM_HEIGHT_PT = 78
const NAV_TITLE_HEIGHT_RPX = 50
const BACK_BUTTON_SIZE_RPX = 40
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = DESIGN_FRAME_HEIGHT_PT * 2

const EMPTY_CANCEL_DETAIL = {
  serviceOrderId: '',
  gameId: '',
  ref: '',
  playerName: '',
  playerAvatarText: '',
  expertName: '',
  expertAvatarText: '',
  serviceTitle: '',
  roleLabel: '',
  contractAmount: 0,
  contractAmountText: '',
  servedDurationText: '',
  totalDurationText: '',
  servedText: '',
  minRate: 0,
  maxRate: 0,
  suggestedRate: 0,
  suggestionMinRate: 0,
  suggestionMaxRate: 0,
  platformFeeRate: 0,
  smartSuggestion: '',
  warningTitle: '免费局取消无需赔付',
  warningDesc: '当前没有收费局，本次取消不会产生赔付金额，但会扣减信用分。',
  isFreeCancel: true,
  agreementTitle: '取消确认',
  cancelTipText: '我理解免费局取消无需赔付，但会扣减信用分。'
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
  if (detail.isFreeCancel) {
    return '免费局无需设置赔付比例，取消后仅扣减信用分。'
  }

  if (!detail.servedDurationText) {
    return ''
  }

  return `行家已投入${detail.servedDurationText}服务，建议设置${detail.suggestionMinRate}%-${detail.suggestionMaxRate}%赔付比例，体现对行家时间成本的尊重。`
}

function getAvatarText(name, fallback) {
  return getSurnameInitials(name, fallback)
}

function buildCancelDetail(options = {}) {
  const amount = getNumber(
    options.amount || options.contractAmount || options.amountText || options.contractAmountText,
    EMPTY_CANCEL_DETAIL.contractAmount
  )
  const gameType = decodeOption(options.gameType || options.type || options.mode)
  const isFreeCancel = gameType === 'free' || decodeOption(options.freeCancel) === '1' || amount <= 0
  const playerName = decodeOption(options.playerName) || EMPTY_CANCEL_DETAIL.playerName
  const playerAvatarText = getAvatarText(playerName, decodeOption(options.playerAvatarText) || EMPTY_CANCEL_DETAIL.playerAvatarText)
  const expertName = decodeOption(options.expertName) || EMPTY_CANCEL_DETAIL.expertName
  const avatarText = getAvatarText(expertName, decodeOption(options.expertAvatarText) || EMPTY_CANCEL_DETAIL.expertAvatarText)
  const minRate = getRateNumber(options.minRate, EMPTY_CANCEL_DETAIL.minRate)
  const maxRate = getRateNumber(options.maxRate, EMPTY_CANCEL_DETAIL.maxRate)
  const suggestedRate = clampRate(options.suggestedRate || options.rate, minRate, maxRate)
  const suggestionMinRate = clampRate(
    options.suggestionMinRate || options.recommendMinRate || EMPTY_CANCEL_DETAIL.suggestionMinRate,
    minRate,
    maxRate
  )
  const suggestionMaxRate = Math.max(suggestionMinRate, clampRate(
    options.suggestionMaxRate || options.recommendMaxRate || EMPTY_CANCEL_DETAIL.suggestionMaxRate,
    minRate,
    maxRate
  ))
  const rawServedText = getOptionText(options, ['servedText', 'serviceDurationText'])
  const servedTextParts = splitServedText(rawServedText || EMPTY_CANCEL_DETAIL.servedText)
  const servedDurationText = getOptionText(
    options,
    ['servedDurationText', 'servedDuration', 'servedHoursText', 'servedTimeText'],
    servedTextParts.servedDurationText || EMPTY_CANCEL_DETAIL.servedDurationText
  )
  const totalDurationText = getOptionText(
    options,
    ['totalDurationText', 'totalDuration', 'totalHoursText', 'serviceTotalDurationText'],
    servedTextParts.totalDurationText || EMPTY_CANCEL_DETAIL.totalDurationText
  )
  const servedText = buildServedText(servedDurationText, totalDurationText) || rawServedText || EMPTY_CANCEL_DETAIL.servedText
  const detail = {
    ...EMPTY_CANCEL_DETAIL,
    minRate,
    maxRate,
    suggestedRate,
    suggestionMinRate,
    suggestionMaxRate,
    servedDurationText,
    totalDurationText,
    servedText,
    serviceOrderId: decodeOption(options.serviceOrderId) || decodeOption(options.orderId) || EMPTY_CANCEL_DETAIL.serviceOrderId,
    gameId: decodeOption(options.gameId) || '',
    ref: decodeOption(options.ref) || EMPTY_CANCEL_DETAIL.ref,
    playerName,
    playerAvatarText,
    expertName,
    expertAvatarText: avatarText,
    serviceTitle: decodeOption(options.serviceTitle) || EMPTY_CANCEL_DETAIL.serviceTitle,
    contractAmount: amount,
    contractAmountText: amount <= 0 ? '免费' : formatCurrency(amount),
    isFreeCancel,
    warningTitle: decodeOption(options.warningTitle) || (isFreeCancel ? '免费局取消无需赔付' : '取消需承担赔付'),
    warningDesc: decodeOption(options.warningDesc) || (isFreeCancel
      ? '当前没有收费局，本次取消不会产生赔付金额，但会扣减信用分。'
      : '作为玩家主动取消，需要按约定比例赔付行家已投入的时间成本。'),
    agreementTitle: isFreeCancel ? '取消确认' : '赔付协议确认',
    cancelTipText: isFreeCancel
      ? '我理解免费局取消无需赔付，但会扣减信用分。'
      : '我已阅读并同意上述赔付协议，确认主动取消服务并承担相应赔付责任。'
  }

  return {
    ...detail,
    smartSuggestion: buildSmartSuggestion(detail)
  }
}

function buildReasonOptions(options = [], activeKey = '') {
  return options.map((item) => ({
    ...item,
    active: item.key === activeKey
  }))
}

function buildAmountState(detail, rate) {
  if (detail.isFreeCancel) {
    return {
      compensationRate: 0,
      compensationAmountText: '免费',
      platformFeeText: '免费',
      payAmountText: '免费',
      refundAmountText: '免费'
    }
  }

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

const EMPTY_PAGE_DETAIL = buildCancelDetail()

Page({
  data: {
    shellLayout: getWhiteShellLayoutStyles(),
    detail: EMPTY_PAGE_DETAIL,
    activityCard: buildActivityCard(EMPTY_PAGE_DETAIL),
    reasonOptions: [],
    selectedReasonKey: '',
    reasonText: '',
    agreementChecked: true,
    agreementText: '',
    agreementItems: [],
    submitting: false,
    detailLoading: true,
    detailError: '',
    ...buildAmountState(EMPTY_PAGE_DETAIL, EMPTY_PAGE_DETAIL.suggestedRate)
  },

  onLoad(options = {}) {
    const gameId = decodeOption(options.gameId || options.id)
    this.setData({ shellLayout: getWhiteShellLayoutStyles(), detailLoading: true, detailError: '' })
    if (!gameId) {
      this.setData({ detailLoading: false, detailError: '后端未返回取消服务所需的局信息' })
      toast.info('后端未返回取消服务所需的局信息')
      return
    }
    this.loadCancelDetail(gameId)
  },

  async loadCancelDetail(gameId) {
    try {
      const data = await gameService.getPlayerCancelDetail(gameId)
      const detail = buildCancelDetail(data)
      const rate = clampRate(detail.suggestedRate, detail.minRate, detail.maxRate)
      this.setData({
        detail,
        activityCard: buildActivityCard(detail),
        detailLoading: false,
        detailError: '',
        ...buildAmountState(detail, rate)
      })
      await this.loadCancelConfig()
    } catch (error) {
      const message = error.message || '后端未返回取消服务详情'
      this.setData({ detail: EMPTY_CANCEL_DETAIL, detailLoading: false, detailError: message })
      toast.info(message)
    }
  },

  async loadCancelConfig() {
    try {
      const config = await gameService.getCancelConfig({ role: 'player' })
      const player = config.player || {}
      const reasonOptions = Array.isArray(player.reasonOptions) ? player.reasonOptions : []
      const selectedReasonKey = player.defaultReason || (reasonOptions[0] && reasonOptions[0].key) || ''

      this.setData({
        reasonOptions: buildReasonOptions(reasonOptions, selectedReasonKey),
        selectedReasonKey,
        agreementText: player.agreementText || '',
        agreementItems: this.data.detail.isFreeCancel
          ? ['免费局取消无需赔付', '我理解本次取消将扣减信用分']
          : Array.isArray(player.agreementItems) ? player.agreementItems : []
      })
    } catch (error) {
      toast.info(error && error.message ? error.message : '取消配置加载失败')
    }
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

    navigateShellRoute(ROUTES.gamePlayerManage)
  },

  onRateChanging(event) {
    this.updateCompensationRate(event.detail && event.detail.value)
  },

  onRateChange(event) {
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
      reasonOptions: buildReasonOptions(this.data.reasonOptions, key)
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
      toast.info(this.data.detail.isFreeCancel ? '请先确认取消规则' : '请先同意赔付协议')
      return
    }

    wx.showModal({
      title: '确认取消服务',
      content: this.data.detail.isFreeCancel
        ? '免费局取消无需赔付，但会扣减信用分。确认后本次服务将取消。'
        : `将按${this.data.compensationRate}%赔付行家，实际支付${this.data.payAmountText}。确认后本次服务将进入取消流程。`,
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

  async submitCancelRequest(reason) {
    if (this.data.submitting) {
      return
    }

    const detail = this.data.detail
    const gameId = detail && detail.gameId

    if (!gameId) {
      toast.info('缺少组局信息，无法提交取消申请')
      return
    }

    this.setData({
      submitting: true
    })

    try {
      const result = await gameService.requestPlayerCancel(gameId, {
        serviceOrderId: detail.serviceOrderId,
        ref: detail.ref,
        reasonKey: reason.key,
        reasonText: this.data.reasonText,
        compensationRate: detail.isFreeCancel ? 0 : this.data.compensationRate,
        compensationAmountText: this.data.compensationAmountText,
        platformFeeText: this.data.platformFeeText,
        payAmountText: this.data.payAmountText,
        refundAmountText: this.data.refundAmountText,
        contractAmount: detail.contractAmount,
        servedDurationText: detail.servedDurationText,
        totalDurationText: detail.totalDurationText
      })

      this.setData({
        submitting: false
      })
      toast.info(result && result.status === 'exited' ? '已退出组局' : '组局已取消')
      navigateShellRoute(ROUTES.gamePlayerManage)
    } catch (error) {
      this.setData({
        submitting: false
      })
      toast.info(error && error.message ? error.message : '取消组局失败')
    }
  }
})
