const { ROUTES } = require('../../../config/routes')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')
const gameService = require('../../../services/game')

const DEFAULT_PAYMENT_AMOUNT = 0
const PAYMENT_SPLITS = []

function buildFeeBreakdown(splits = PAYMENT_SPLITS) {
  return splits.map((item) => ({
    marker: item.marker,
    label: item.label,
    value: item.desc ? `${item.amountText}（${item.desc}）` : item.amountText
  }))
}

function buildPaymentPayload(data = {}) {
  const payment = data.payment || {}
  const splits = data.paymentSplits || []

  return {
    gameId: payment.gameId || '',
    scene: 'deposit_game',
    payChannel: 'wechat',
    amount: Number(payment.amount || 0),
    currency: payment.currency || 'CNY',
    agreementChecked: Boolean(data.agreementChecked),
    splits: splits.map((item) => ({
      key: item.key,
      label: item.label,
      amount: item.amount,
      desc: item.desc || ''
    }))
  }
}

function normalizeWechatPaymentParams(result = {}) {
  const params = result.paymentParams || result.wechatPaymentParams || result.wxPayParams || result

  return {
    timeStamp: String(params.timeStamp || params.timestamp || ''),
    nonceStr: params.nonceStr || params.nonce || '',
    package: params.package || params.packageValue || '',
    signType: params.signType || 'RSA',
    paySign: params.paySign || params.sign || ''
  }
}

Page({
  data: {
    onlineText: '在线',
    agreementChecked: false,
    paymentSubmitting: false,
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    payment: {
      gameId: '',
      amount: DEFAULT_PAYMENT_AMOUNT,
      currency: 'CNY'
    },
    paymentSplits: PAYMENT_SPLITS,
    feeBreakdown: buildFeeBreakdown(PAYMENT_SPLITS),
    rules: [
      { icon: '✅', type: 'success', text: '一期免费局不发起真实微信支付' },
      { icon: '💡', type: 'warning', text: '收费、押金、分账能力由后端订单接口返回后展示' }
    ],
    pointsDescription: '免费局确认后进入组局流程，积分与成长由服务确认和评价链路沉淀。'
  },

  onLoad(options = {}) {
    const amount = Number(options.amount)

    this.setData({
      'payment.gameId': options.gameId || options.id || '',
      'payment.amount': Number.isFinite(amount) && amount >= 0 ? amount : DEFAULT_PAYMENT_AMOUNT
    })
  },

  toggleAgreement() {
    this.setData({
      agreementChecked: !this.data.agreementChecked
    })
  },

  handleCancel() {
    if (getCurrentPages().length > 1) {
      wx.navigateBack()
      return
    }

    if (this.data.payment.gameId) {
      navigateShellRoute(`/${ROUTES.gameDetail}?gameId=${this.data.payment.gameId}`)
      return
    }

    navigateShellRoute(ROUTES.gameHall)
  },

  async handleWechatPay() {
    if (!this.data.agreementChecked) {
      this.showAgreementRequired()
      return
    }

    if (this.data.paymentSubmitting) {
      return
    }

    this.setData({
      paymentSubmitting: true
    })

    wx.showLoading({
      title: '发起支付中',
      mask: true
    })

    try {
      const order = await gameService.createGamePayment(buildPaymentPayload(this.data))
      this.applyOrderPayment(order)

      if (order.needWechatPay === false) {
        this.handlePlaceholderPaymentSuccess(order)
        return
      }

      if (order.mockPayment) {
        this.handlePlaceholderPaymentSuccess(order)
        return
      }

      const paymentParams = normalizeWechatPaymentParams(order)

      await this.requestWechatPayment(paymentParams)
      this.showInfo('支付成功')
    } catch (error) {
      if (error && error.errMsg && error.errMsg.includes('cancel')) {
        this.showInfo('已取消支付')
      } else {
        this.showInfo(error.message || error.errMsg || '支付失败，请重试')
      }
    } finally {
      wx.hideLoading()
      this.setData({
        paymentSubmitting: false
      })
    }
  },

  showAgreementRequired() {
    wx.showModal({
      title: '请先勾选规则',
      content: '请勾选“我已了解并同意当前组局规则”后再确认订单。',
      showCancel: false,
      confirmText: '知道了'
    })
  },

  applyOrderPayment(order = {}) {
    const source = order.payment || order.order || order
    const amount = Number(source.amount || source.totalAmount || source.amountYuan || 0)
    const splits = Array.isArray(order.paymentSplits) ? order.paymentSplits : []

    this.setData({
      payment: {
        ...this.data.payment,
        gameId: source.gameId || this.data.payment.gameId,
        amount: Number.isFinite(amount) ? amount : 0,
        currency: source.currency || this.data.payment.currency || 'CNY'
      },
      paymentSplits: splits,
      feeBreakdown: buildFeeBreakdown(splits)
    })
  },

  handlePlaceholderPaymentSuccess(order = {}) {
    const gameId = order.gameId || (order.order && order.order.gameId) || this.data.payment.gameId
    this.showInfo('已确认，无需微信支付')

    if (!gameId) {
      return
    }

    setTimeout(() => {
      navigateShellRoute(`/${ROUTES.gameDetail}?gameId=${encodeURIComponent(gameId)}`)
    }, 500)
  },

  requestWechatPayment(paymentParams) {
    const requiredFields = ['timeStamp', 'nonceStr', 'package', 'signType', 'paySign']
    const missingField = requiredFields.find((key) => !paymentParams[key])

    if (missingField) {
      return Promise.reject(new Error('微信支付参数不完整'))
    }

    return new Promise((resolve, reject) => {
      wx.requestPayment({
        ...paymentParams,
        success: resolve,
        fail: reject
      })
    })
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (navigateShellKey(key, {
      currentRoute: ROUTES.gamePayment
    })) {
      return
    }

    if (key === 'search') {
      this.navigateToRoute(ROUTES.gameHall)
      return
    }

    if (key === 'comment' || key === 'message') {
      this.navigateToRoute(ROUTES.message)
      return
    }

    if (key === 'avatar' || key === 'mine') {
      this.navigateToRoute(ROUTES.profile)
      return
    }

    const routeMap = {
      home: ROUTES.playerHome || ROUTES.home,
      metaverse: ROUTES.metaverse,
      map: ROUTES.map
    }

    this.navigateToRoute(routeMap[key])
  },

  navigateToRoute(route) {
    if (!route || route === ROUTES.gamePayment) {
      return
    }

    navigateShellRoute(route)
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
