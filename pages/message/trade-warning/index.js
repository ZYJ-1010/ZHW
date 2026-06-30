const { ROUTES } = require('../../../config/routes')
const messageService = require('../../../services/message')

const DEFAULT_TRADE_WARNING_DETAIL = {
  pageTitle: '交易预警',
  onlineText: '',
  warning: {},
  countdown: [],
  order: {},
  deliveryMethods: [],
  actions: {
    delayText: '',
    deliverText: ''
  }
}

function normalizeTradeWarningDetail(data = {}) {
  const source = Object.assign({}, DEFAULT_TRADE_WARNING_DETAIL, data)
  const order = Object.assign({}, DEFAULT_TRADE_WARNING_DETAIL.order, data.order || {})
  const warning = Object.assign({}, DEFAULT_TRADE_WARNING_DETAIL.warning, data.warning || {})
  const actions = Object.assign({}, DEFAULT_TRADE_WARNING_DETAIL.actions, data.actions || {})
  const countdown = Array.isArray(data.countdown) && data.countdown.length
    ? data.countdown
    : []
  const deliveryMethods = normalizeDeliveryMethods(data.deliveryMethods)
  const detailRows = Array.isArray(order.detailRows) ? order.detailRows : []

  return {
    pageTitle: source.pageTitle || DEFAULT_TRADE_WARNING_DETAIL.pageTitle,
    onlineText: source.onlineText || DEFAULT_TRADE_WARNING_DETAIL.onlineText,
    warning,
    countdown,
    order,
    detailRows,
    deliveryMethods,
    actions,
    warningId: source.warningId || source.id || '',
    hasWarning: Boolean(warning.title || warning.prefixText || warning.highlightText || warning.suffixText),
    hasCountdown: Boolean(countdown.length),
    hasOrder: Boolean(order.orderNo || order.customerTitle || detailRows.length),
    hasDeliveryMethods: Boolean(deliveryMethods.length),
    hasActions: Boolean(actions.delayText || actions.deliverText),
    loading: false,
    errorText: ''
  }
}

function normalizeDeliveryMethods(methods) {
  const source = Array.isArray(methods) && methods.length
    ? methods
    : []
  const hasActive = source.some((item) => item.active)

  return source.map((item, index) => Object.assign({}, item, {
    active: hasActive ? Boolean(item.active) : index === 0
  }))
}

Page({
  data: {
    pageTitle: '交易预警',
    onlineText: '',
    navItems: [
      { name: '我的', key: 'mine' },
      { name: '元宇宙', key: 'metaverse' },
      { name: '地图', key: 'map' },
      { name: '消息', key: 'message' },
      { name: '首页', key: 'home' }
    ],
    warning: DEFAULT_TRADE_WARNING_DETAIL.warning,
    countdown: DEFAULT_TRADE_WARNING_DETAIL.countdown,
    order: DEFAULT_TRADE_WARNING_DETAIL.order,
    detailRows: [],
    deliveryMethods: DEFAULT_TRADE_WARNING_DETAIL.deliveryMethods,
    actions: DEFAULT_TRADE_WARNING_DETAIL.actions,
    warningId: '',
    hasWarning: false,
    hasCountdown: false,
    hasOrder: false,
    hasDeliveryMethods: false,
    hasActions: false,
    loading: false,
    errorText: ''
  },

  onLoad(options = {}) {
    this.loadTradeWarningDetail(options)
  },

  async loadTradeWarningDetail(options = {}) {
    this.setData({
      loading: true,
      errorText: ''
    })

    try {
      const detail = await messageService.getTradeWarningDetail({
        warningId: options.warningId || options.id || '',
        orderId: options.orderId || ''
      })

      this.setData(normalizeTradeWarningDetail(detail))
    } catch (error) {
      const errorText = error && error.message ? error.message : '获取交易预警失败'

      this.setData(Object.assign({}, normalizeTradeWarningDetail(), {
        errorText
      }))
      this.showInfo(errorText)
    }
  },

  onActionTap(event) {
    if (!this.data.hasActions) {
      return
    }

    const { action } = event.currentTarget.dataset
    const actionText = action === 'deliver' ? this.data.actions.deliverText : this.data.actions.delayText
    const text = `${actionText}待接入`
    this.showInfo(text)
  },

  onMethodTap(event) {
    const { id } = event.currentTarget.dataset
    const deliveryMethods = this.data.deliveryMethods.map((item) => ({
      ...item,
      active: item.id === id
    }))

    this.setData({ deliveryMethods })
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}

    if (key === 'map') {
      wx.showToast({
        title: '地图功能开发中',
        icon: 'none'
      })
      return
    }
    const routeMap = {
      home: ROUTES.playerHome || ROUTES.home,
      map: '',
      message: ROUTES.message,
      mine: ROUTES.profile,
      avatar: ROUTES.profile,
      metaverse: ROUTES.metaverse
    }
    const route = routeMap[key]

    if (!route || route === ROUTES.messageTradeWarning) {
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
