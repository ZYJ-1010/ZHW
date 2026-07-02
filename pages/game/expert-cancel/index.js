const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
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
const DEFAULT_FRAME_HEIGHT_RPX = 1620
const DEFAULT_ACTIVITY = {
  serviceOrderId: '',
  gameId: '',
  playerId: '',
  playerName: '',
  avatarText: '',
  serviceTitle: '',
  amount: 0,
  amountText: '¥0',
  platformFeeRate: 10,
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

function parseAmountText(value) {
  const text = String(value || '').replace(/[^\d.]/g, '')
  const amount = Number(text)

  return Number.isFinite(amount) ? amount : DEFAULT_ACTIVITY.amount
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
    return '¥0'
  }

  return `¥${Math.round(amount).toLocaleString('zh-CN')}`
}

function getAvatarText(name, fallback = 'LI') {
  return getSurnameInitials(name, fallback)
}

function normalizeAvatarText(value, name, fallback = 'LI') {
  const text = decodeOption(value).trim()

  return getAvatarText(name, text || fallback)
}

function normalizeActivity(options = {}) {
  const playerName = decodeOption(options.playerName || DEFAULT_ACTIVITY.playerName)
  const serviceTitle = decodeOption(options.serviceTitle || DEFAULT_ACTIVITY.serviceTitle)
  const amount = options.amount
    ? Number(options.amount)
    : parseAmountText(options.amountText || DEFAULT_ACTIVITY.amountText)
  const platformFeeRate = getNumber(options.platformFeeRate, DEFAULT_ACTIVITY.platformFeeRate)
  const platformFeeText = decodeOption(options.platformFeeText || '')
  const platformFee = platformFeeText
    ? parseAmountText(platformFeeText)
    : options.platformFee || options.serviceFee
      ? getNumber(options.platformFee || options.serviceFee)
      : null

  return {
    serviceOrderId: options.serviceOrderId || options.orderId || DEFAULT_ACTIVITY.serviceOrderId,
    gameId: options.gameId || DEFAULT_ACTIVITY.gameId,
    playerId: options.playerId || DEFAULT_ACTIVITY.playerId,
    playerName,
    avatarText: normalizeAvatarText(options.avatarText || '', playerName, DEFAULT_ACTIVITY.avatarText),
    serviceTitle,
    amount,
    amountText: decodeOption(options.amountText || '') || formatCurrency(amount),
    platformFeeRate,
    platformFee,
    platformFeeText,
    statusText: decodeOption(options.statusText || DEFAULT_ACTIVITY.statusText)
  }
}

function buildPaymentState(activity, ratio) {
  const compensationAmount = activity.amount * ratio / 100
  const serviceFee = activity.platformFee !== null && activity.platformFee !== undefined
    ? activity.platformFee
    : compensationAmount * activity.platformFeeRate / 100
  const totalDebit = compensationAmount + serviceFee

  return {
    compensationAmountText: formatCurrency(compensationAmount),
    serviceFeeText: activity.platformFeeText || formatCurrency(serviceFee),
    totalDebitText: formatCurrency(totalDebit)
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
    ]
  }
}

function normalizeRequiredText(value) {
  return String(value || '').trim()
}

Page({
  data: {
    shellLayout: getWhiteShellLayoutStyles(),
    queryParams: {},
    activity: DEFAULT_ACTIVITY,
    activityCard: buildActivityCard(DEFAULT_ACTIVITY),
    compensationRatio: 20,
    compensationAmountText: '¥0',
    serviceFeeText: '¥0',
    totalDebitText: '¥0',
    reasonOptions: [],
    selectedReasonKey: '',
    reasonDetail: '',
    agreementChecked: true,
    agreementText: ''
  },

  onLoad(options = {}) {
    const activity = normalizeActivity(options)
    const paymentState = buildPaymentState(activity, this.data.compensationRatio)

    this.setData({
      queryParams: options,
      activity,
      activityCard: buildActivityCard(activity),
      ...paymentState
    })
    this.loadCancelConfig()
    this.updateShellLayout()
  },

  async loadCancelConfig() {
    try {
      const config = await gameService.getCancelConfig({ role: 'expert' })
      const expert = config.expert || {}
      const reasonOptions = Array.isArray(expert.reasonOptions) ? expert.reasonOptions : []

      this.setData({
        reasonOptions,
        selectedReasonKey: expert.defaultReason || (reasonOptions[0] && reasonOptions[0].key) || '',
        agreementText: expert.agreementText || ''
      })
    } catch (error) {
      wx.showToast({
        title: error && error.message ? error.message : '取消配置加载失败',
        icon: 'none'
      })
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

  updatePaymentState(ratio) {
    this.setData({
      compensationRatio: ratio,
      ...buildPaymentState(this.data.activity, ratio)
    })
  },

  onRatioChanging(event) {
    this.updatePaymentState(Number(event.detail.value))
  },

  onRatioChange(event) {
    this.updatePaymentState(Number(event.detail.value))
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
    const agreementChecked = Boolean(event.detail.checked)

    this.setData({
      agreementChecked
    })
  },

  onConfirmCancelTap() {
    const reasonDetail = normalizeRequiredText(this.data.reasonDetail)

    if (!this.data.selectedReasonKey) {
      wx.showToast({
        title: '请选择取消原因',
        icon: 'none'
      })
      return
    }

    if (!reasonDetail) {
      wx.showToast({
        title: '请填写详细说明',
        icon: 'none'
      })
      return
    }

    if (!this.data.agreementChecked) {
      wx.showToast({
        title: '请先同意服务取消协议',
        icon: 'none'
      })
      return
    }

    wx.showModal({
      title: '确认取消服务',
      content: `本次将扣款 ${this.data.totalDebitText}，确认后会提交取消赔付申请。`,
      confirmText: '确认取消',
      confirmColor: '#ef4444',
      success: (res) => {
        if (!res.confirm) {
          return
        }

        this.submitExpertCancel(reasonDetail)
      }
    })
  },

  async submitExpertCancel(reasonDetail) {
    const activity = this.data.activity

    if (!activity || !activity.gameId) {
      wx.showToast({
        title: '缺少组局信息，无法提交取消',
        icon: 'none'
      })
      return
    }

    try {
      await gameService.requestExpertCancel(activity.gameId, {
        serviceOrderId: activity.serviceOrderId,
        playerId: activity.playerId,
        reasonKey: this.data.selectedReasonKey,
        reasonText: reasonDetail,
        compensationRate: this.data.compensationRatio,
        compensationAmountText: this.data.compensationAmountText,
        platformFeeText: this.data.serviceFeeText,
        payAmountText: this.data.totalDebitText,
        contractAmount: activity.amount,
        statusText: activity.statusText
      })
      wx.showToast({
        title: '取消赔付申请已提交',
        icon: 'none'
      })
      navigateShellRoute(ROUTES.gameManage)
    } catch (error) {
      wx.showToast({
        title: error && error.message ? error.message : '取消赔付提交失败',
        icon: 'none'
      })
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

    navigateShellRoute(ROUTES.gameManage)
  }
})
