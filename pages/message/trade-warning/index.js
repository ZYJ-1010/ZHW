const { ROUTES } = require('../../../config/routes')
const messageService = require('../../../services/message')

const DEFAULT_TRADE_WARNING_DETAIL = {
  pageTitle: '交易预警',
  onlineText: '3999人在线',
  warning: {
    title: '即将超时',
    prefixText: '该订单将于',
    highlightText: '1小时30分钟',
    suffixText: '后自动标记为逾期，请立即处理'
  },
  countdown: [
    { value: '01', label: '小时' },
    { value: '30', label: '分钟' },
    { value: '45', label: '秒' }
  ],
  order: {
    orderNo: 'GD2024032201',
    statusText: '待交付',
    customerAvatarText: 'CL',
    customerTitle: '客户需求',
    customerDesc: '寻找资深产品经理进行业务咨询',
    detailRows: [
      { label: '约定交付时间', value: '今天 16:00' },
      { label: '服务费用', value: '¥500', strong: true }
    ]
  },
  deliveryMethods: [
    {
      id: 'online',
      title: '线上确认',
      desc: '双方在线确认服务完成',
      active: true
    },
    {
      id: 'upload',
      title: '上传凭证',
      desc: '上传服务完成截图或文件',
      active: false
    }
  ],
  actions: {
    delayText: '申请延期',
    deliverText: '立即交付'
  }
}

function normalizeTradeWarningDetail(data = {}) {
  const source = Object.assign({}, DEFAULT_TRADE_WARNING_DETAIL, data)
  const order = Object.assign({}, DEFAULT_TRADE_WARNING_DETAIL.order, data.order || {})
  const warning = Object.assign({}, DEFAULT_TRADE_WARNING_DETAIL.warning, data.warning || {})
  const actions = Object.assign({}, DEFAULT_TRADE_WARNING_DETAIL.actions, data.actions || {})
  const countdown = Array.isArray(data.countdown) && data.countdown.length
    ? data.countdown
    : DEFAULT_TRADE_WARNING_DETAIL.countdown
  const deliveryMethods = normalizeDeliveryMethods(data.deliveryMethods)

  return {
    pageTitle: source.pageTitle || DEFAULT_TRADE_WARNING_DETAIL.pageTitle,
    onlineText: source.onlineText || DEFAULT_TRADE_WARNING_DETAIL.onlineText,
    warning,
    countdown,
    order,
    detailRows: Array.isArray(order.detailRows) ? order.detailRows : DEFAULT_TRADE_WARNING_DETAIL.order.detailRows,
    deliveryMethods,
    actions,
    warningId: source.warningId || source.id || '',
    loading: false,
    errorText: ''
  }
}

function normalizeDeliveryMethods(methods) {
  const source = Array.isArray(methods) && methods.length
    ? methods
    : DEFAULT_TRADE_WARNING_DETAIL.deliveryMethods
  const hasActive = source.some((item) => item.active)

  return source.map((item, index) => Object.assign({}, item, {
    active: hasActive ? Boolean(item.active) : index === 0
  }))
}

Page({
  data: {
    pageTitle: '交易预警',
    onlineText: '3999人在线',
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
    detailRows: DEFAULT_TRADE_WARNING_DETAIL.order.detailRows,
    deliveryMethods: DEFAULT_TRADE_WARNING_DETAIL.deliveryMethods,
    actions: DEFAULT_TRADE_WARNING_DETAIL.actions,
    warningId: '',
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
      this.setData(Object.assign({}, normalizeTradeWarningDetail(DEFAULT_TRADE_WARNING_DETAIL), {
        errorText: error.message || '获取交易预警失败'
      }))
      this.showInfo(error.message || '获取交易预警失败')
    }
  },

  onActionTap(event) {
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
    const routeMap = {
      home: ROUTES.playerHome || ROUTES.home,
      map: ROUTES.map,
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
