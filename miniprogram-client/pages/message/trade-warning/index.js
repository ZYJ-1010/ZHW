const { ROUTES } = require('../../../config/routes')
const messageService = require('../../../services/message')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')
const { toUserMessage } = require('../../../utils/user-message')

const EMPTY_TRADE_WARNING_DETAIL = {
  pageTitle: '',
  onlineText: '',
  warning: {
    title: '',
    prefixText: '',
    highlightText: '',
    suffixText: ''
  },
  countdown: [],
  order: {
    orderNo: '',
    statusText: '',
    customerAvatarText: '',
    customerTitle: '',
    customerDesc: '',
    detailRows: []
  },
  deliveryMethods: [],
  actions: {
    delayText: '',
    deliverText: ''
  },
  texts: {}
}

function textOf(config, key) {
  const texts = config && config.texts ? config.texts : {}
  return texts[key] || ''
}

function normalizeTradeWarningDetail(data = {}) {
  const source = Object.assign({}, EMPTY_TRADE_WARNING_DETAIL, data)
  const order = Object.assign({}, EMPTY_TRADE_WARNING_DETAIL.order, data.order || {})
  const warning = Object.assign({}, EMPTY_TRADE_WARNING_DETAIL.warning, data.warning || {})
  const actions = Object.assign({}, EMPTY_TRADE_WARNING_DETAIL.actions, data.actions || {})
  const countdown = Array.isArray(data.countdown) && data.countdown.length
    ? data.countdown
    : EMPTY_TRADE_WARNING_DETAIL.countdown
  const deliveryMethods = normalizeDeliveryMethods(data.deliveryMethods)

  return {
    pageTitle: source.pageTitle || '',
    onlineText: source.onlineText || '',
    warning,
    countdown,
    order,
    detailRows: Array.isArray(order.detailRows) ? order.detailRows : EMPTY_TRADE_WARNING_DETAIL.order.detailRows,
    deliveryMethods,
    actions,
    texts: source.texts || {},
    warningId: source.warningId || source.id || '',
    gameId: source.gameId || source.bizId || '',
    loading: false,
    errorText: ''
  }
}

function normalizeDeliveryMethods(methods) {
  const source = Array.isArray(methods) && methods.length
    ? methods
    : EMPTY_TRADE_WARNING_DETAIL.deliveryMethods
  const hasActive = source.some((item) => item.active)

  return source.map((item, index) => Object.assign({}, item, {
    active: hasActive ? Boolean(item.active) : index === 0
  }))
}

Page({
  data: {
    pageTitle: '',
    onlineText: '',
    navItems: [
      { name: '我的', key: 'mine' },
      { name: '元宇宙', key: 'metaverse' },
      { name: '地图', key: 'map' },
      { name: '消息', key: 'message' },
      { name: '首页', key: 'home' }
    ],
    warning: EMPTY_TRADE_WARNING_DETAIL.warning,
    countdown: EMPTY_TRADE_WARNING_DETAIL.countdown,
    order: EMPTY_TRADE_WARNING_DETAIL.order,
    detailRows: EMPTY_TRADE_WARNING_DETAIL.order.detailRows,
    deliveryMethods: EMPTY_TRADE_WARNING_DETAIL.deliveryMethods,
    actions: EMPTY_TRADE_WARNING_DETAIL.actions,
    texts: EMPTY_TRADE_WARNING_DETAIL.texts,
    warningId: '',
    gameId: '',
    submittingAction: '',
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
        orderId: options.orderId || '',
        gameId: options.gameId || ''
      })

      this.setData(normalizeTradeWarningDetail(detail))
    } catch (error) {
      const errorText = toUserMessage(error && error.message, textOf(this.data, 'loadFailedText') || '交易提醒加载失败')
      this.setData(Object.assign({}, normalizeTradeWarningDetail(EMPTY_TRADE_WARNING_DETAIL), {
        errorText
      }))
      this.showInfo(errorText)
    }
  },

  async onActionTap(event) {
    const { action } = event.currentTarget.dataset
    const selectedMethod = this.data.deliveryMethods.find((item) => item.active) || {}

    if (this.data.submittingAction) {
      return
    }

    if (action !== 'delay' && action !== 'deliver') {
      return
    }

    this.setData({ submittingAction: action })

    try {
      const data = await messageService.handleTradeWarning({
        action,
        warningId: this.data.warningId,
        orderId: this.data.order.orderNo,
        gameId: this.data.gameId,
        deliveryMethod: selectedMethod.id || 'online'
      })

      if (action === 'deliver') {
        const target = data.target || {}
        navigateShellRoute(target.route || ROUTES.gameDelivery || 'pages/game/delivery/index', {
          currentRoute: ROUTES.messageTradeWarning
        })
        return
      }

      this.setData({
        order: Object.assign({}, this.data.order, {
          statusText: data.statusText || textOf(this.data, 'delayStatusText')
        })
      })
      this.showInfo(data.message || textOf(this.data, 'delaySuccessText'))
    } catch (error) {
      this.showInfo(error.message || textOf(this.data, 'actionFailedText'))
    } finally {
      this.setData({ submittingAction: '' })
    }
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
    navigateShellKey(key, {
      currentRoute: ROUTES.messageTradeWarning
    })
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
