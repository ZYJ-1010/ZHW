const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')

function normalizeBoolean(value) {
  if (typeof value === 'boolean') {
    return value
  }

  if (typeof value === 'string') {
    return value === 'true' || value === '1'
  }

  return Boolean(value)
}

function normalizePaymentSplit(split = {}) {
  return {
    key: split.key || split.type || '',
    marker: split.marker || '',
    label: split.label || split.name || '',
    amount: split.amount,
    amountText: split.amountText || split.value || '',
    desc: split.desc || split.description || ''
  }
}

function buildFeeBreakdown(splits = []) {
  return splits.map((item) => ({
    marker: item.marker || '',
    label: item.label || '',
    value: item.desc ? `${item.amountText || item.amount || ''}（${item.desc}）` : (item.amountText || item.amount || '')
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
    agreementChecked: normalizeBoolean(data.agreementChecked),
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

function normalizeRule(rule = {}) {
  return {
    icon: rule.icon || '',
    type: rule.type || '',
    text: rule.text || rule.content || ''
  }
}

function normalizePaymentConfig(data = {}, fallbackGameId = '') {
  const payment = data.payment || data
  const paymentSplits = Array.isArray(data.paymentSplits || data.splits || data.feeSplits)
    ? (data.paymentSplits || data.splits || data.feeSplits).map(normalizePaymentSplit)
    : []
  const rules = Array.isArray(data.rules || data.ruleItems)
    ? (data.rules || data.ruleItems).map(normalizeRule)
    : []

  return {
    onlineText: data.onlineText || '3999人在线',
    agreementChecked: normalizeBoolean(data.agreementChecked),
    payment: {
      gameId: payment.gameId || data.gameId || fallbackGameId || '',
      amount: payment.amount || data.amount || '',
      currency: payment.currency || data.currency || 'CNY'
    },
    paymentSplits,
    feeBreakdown: buildFeeBreakdown(paymentSplits),
    rules,
    pointsDescription: data.pointsDescription || data.pointsText || ''
  }
}

Page({
  data: {
    onlineText: '3999人在线',
    agreementChecked: false,
    loading: false,
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
      amount: '',
      currency: 'CNY'
    },
    paymentSplits: [],
    feeBreakdown: [],
    rules: [],
    pointsDescription: ''
  },

  onLoad(options = {}) {
    this.setData({
      'payment.gameId': options.gameId || options.id || ''
    })
    this.loadPaymentConfig(options)
  },

  async loadPaymentConfig(options = {}) {
    const gameId = options.gameId || options.id || this.data.payment.gameId || ''

    this.setData({
      loading: true
    })

    try {
      const data = await gameService.getGamePaymentConfig({
        ...options,
        gameId
      })

      this.setData({
        ...normalizePaymentConfig(data, gameId),
        loading: false
      })
    } catch (error) {
      this.setData({
        ...normalizePaymentConfig({}, gameId),
        loading: false
      })
      wx.showToast({
        title: error.message || '支付配置加载失败',
        icon: 'none'
      })
    }
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

    if (key === 'map') {
      wx.showToast({
        title: '地图功能开发中',
        icon: 'none'
      })
      return
    }

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
      map: ''
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
