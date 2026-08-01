const { getLegacyWhiteFrameLayoutStyles } = require('../../../utils/adaptive-shell-layout')
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
  statusText: '',
  isFreeCancel: true,
  warningTitle: '取消本局提醒',
  warningDesc: '本局为免费局，取消后将按规则扣减信用分。'
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

  if (!Number.isFinite(amount) || amount <= 0) {
    return '免费'
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

  return {
    serviceOrderId: options.serviceOrderId || options.orderId || DEFAULT_ACTIVITY.serviceOrderId,
    gameId: options.gameId || DEFAULT_ACTIVITY.gameId,
    playerId: options.playerId || DEFAULT_ACTIVITY.playerId,
    playerName,
    avatarText: normalizeAvatarText(options.avatarText || '', playerName, DEFAULT_ACTIVITY.avatarText),
    serviceTitle,
    amount: 0,
    amountText: '免费',
    platformFeeRate: 0,
    platformFee: null,
    platformFeeText: '',
    statusText: decodeOption(options.statusText || DEFAULT_ACTIVITY.statusText),
    isFreeCancel: true,
    warningTitle: '取消本局提醒',
    warningDesc: '本局为免费局，取消后将按规则扣减信用分。'
  }
}

function buildPaymentState(activity, ratio) {
  if (activity.isFreeCancel) {
    return {
      compensationAmountText: '免费',
      serviceFeeText: '免费',
      totalDebitText: '免费'
    }
  }

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
    title: '组局信息',
    avatarText: activity.avatarText,
    avatarClass: 'player',
    name: activity.serviceTitle || '未命名组局',
    roleLabel: '组局',
    roleClass: 'player',
    serviceTitle: activity.serviceTitle,
    rows: [
      { label: '取消规则', value: '免费局取消将扣减信用分', tone: 'strong', divider: true },
      { label: '当前状态', value: activity.statusText, tone: 'active' }
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
    agreementText: '',
    submitting: false,
    detailLoading: true,
    detailError: ''
  },

  onLoad(options = {}) {
    const gameId = decodeOption(options.gameId || options.id)
    this.setData({ queryParams: { gameId }, detailLoading: true, detailError: '' })
    if (!gameId) {
      this.setData({ detailLoading: false, detailError: '后端未返回取消服务所需的局信息' })
      wx.showToast({ title: '后端未返回取消服务所需的局信息', icon: 'none' })
      return
    }
    this.loadCancelDetail(gameId)
    this.updateShellLayout()
  },

  async loadCancelDetail(gameId) {
    try {
      const data = await gameService.getExpertCancelDetail(gameId)
      const activity = normalizeActivity(data)
      this.setData({
        activity,
        activityCard: buildActivityCard(activity),
        compensationRatio: 0,
        detailLoading: false,
        detailError: '',
        ...buildPaymentState(activity, 0)
      })
      await this.loadCancelConfig()
    } catch (error) {
      const message = error.message || '后端未返回取消服务详情'
      this.setData({ activity: DEFAULT_ACTIVITY, detailLoading: false, detailError: message })
      wx.showToast({ title: message, icon: 'none' })
    }
  },

  async loadCancelConfig() {
    try {
      const config = await gameService.getCancelConfig({ role: 'expert' })
      const expert = config.expert || {}
      const reasonOptions = Array.isArray(expert.reasonOptions) ? expert.reasonOptions : []

      this.setData({
        reasonOptions,
        selectedReasonKey: expert.defaultReason || (reasonOptions[0] && reasonOptions[0].key) || '',
        agreementText: '我已了解本次取消将扣减信用分。'
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
    if (this.data.submitting) {
      return
    }

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
        title: '请先确认取消规则',
        icon: 'none'
      })
      return
    }

    wx.showModal({
      title: '确认取消本局',
      content: '本局为免费局，取消后将扣减信用分。确认后本局将进入取消流程。',
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

    this.setData({ submitting: true })
    try {
      await gameService.requestExpertCancel(activity.gameId, {
        serviceOrderId: activity.serviceOrderId,
        playerId: activity.playerId,
        reasonKey: this.data.selectedReasonKey,
        reasonText: reasonDetail,
        compensationRate: 0,
        compensationAmountText: '',
        platformFeeText: '',
        payAmountText: '',
        contractAmount: 0,
        statusText: activity.statusText
      })
      wx.showToast({
        title: '组局已取消',
        icon: 'none'
      })
      navigateShellRoute(ROUTES.gameManage)
    } catch (error) {
      wx.showToast({
        title: error && error.message ? error.message : '取消组局失败',
        icon: 'none'
      })
    } finally {
      this.setData({ submitting: false })
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
