const { ROUTES } = require('../../../config/routes')
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
const DEFAULT_ACTIVITY = {
  serviceOrderId: 'service_order_001',
  gameId: 'game_001',
  playerId: 'player_001',
  playerName: '李明',
  avatarText: 'LI',
  serviceTitle: '产品架构咨询',
  amount: 800,
  amountText: '¥800',
  platformFeeRate: 10,
  platformFee: null,
  platformFeeText: '',
  statusText: '进行中（第3天）'
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
    compensationAmountText: '¥160',
    serviceFeeText: '¥16',
    totalDebitText: '¥176',
    reasonOptions: [
      { key: 'schedule_conflict', text: '个人时间冲突，无法交付' },
      { key: 'requirement_mismatch', text: '需求与描述不符，无法完成' },
      { key: 'emergency', text: '身体原因/突发状况' },
      { key: 'other', text: '其他原因' }
    ],
    selectedReasonKey: 'schedule_conflict',
    reasonDetail: '',
    agreementChecked: true,
    agreementText: '我已阅读并同意《服务取消协议》，理解主动取消将对我的信用分产生影响（-5分），并同意按设置比例赔付玩家损失。'
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
      content: `本次将扣款 ${this.data.totalDebitText}，确认后服务取消接口待接入。`,
      confirmText: '确认取消',
      confirmColor: '#ef4444',
      success: (res) => {
        if (!res.confirm) {
          return
        }

        wx.showToast({
          title: '取消赔付接口待接入',
          icon: 'none'
        })
      }
    })
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
  }
})
