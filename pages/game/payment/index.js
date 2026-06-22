const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')

const DEFAULT_PAYMENT_AMOUNT = 100
const PAYMENT_SPLITS = [
  { key: 'serviceFee', marker: '├─', label: '服务费', amount: 10, amountText: '10元' },
  { key: 'platformServiceFee', marker: '├─', label: '平台服务费', amount: 2.5, amountText: '2.5元' },
  { key: 'inviterReward', marker: '├─', label: '邀请人奖励', amount: 5, amountText: '5元' },
  { key: 'partnerReward', marker: '├─', label: '合伙人奖励', amount: 2.5, amountText: '2.5元' },
  { key: 'depositPool', marker: '└─', label: '押金池', amount: 90, amountText: '90元', desc: '完成任务后返还' }
]

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
    onlineText: '3999人在线',
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
      { icon: '✅', type: 'success', text: '完成任务，拿回90元押金' },
      { icon: '❌', type: 'danger', text: '未完成任务，90元押金由完成者平分' },
      { icon: '💡', type: 'warning', text: '服务费10元不退还' }
    ],
    pointsDescription: '按模板规则，支付完成并结算后，相关角色将自动获得积分。'
  },

  onLoad(options = {}) {
    const amount = Number(options.amount)

    this.setData({
      'payment.gameId': options.gameId || options.id || '',
      'payment.amount': Number.isFinite(amount) && amount > 0 ? amount : DEFAULT_PAYMENT_AMOUNT
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
      wx.redirectTo({
        url: `/${ROUTES.gameDetail}?gameId=${this.data.payment.gameId}`
      })
      return
    }

    wx.redirectTo({
      url: `/${ROUTES.gameHall}`
    })
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

      if (order.mockPayment) {
        this.showInfo('支付请求已提交（mock）')
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
      content: '请勾选“我已了解并同意押金局规则”后再继续微信支付。',
      showCancel: false,
      confirmText: '知道了'
    })
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

    if (key === 'search') {
      this.showInfo('搜索功能开发中')
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
      home: ROUTES.home,
      metaverse: ROUTES.metaverse,
      map: ROUTES.map
    }

    this.navigateToRoute(routeMap[key])
  },

  navigateToRoute(route) {
    if (!route || route === ROUTES.gamePayment) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
