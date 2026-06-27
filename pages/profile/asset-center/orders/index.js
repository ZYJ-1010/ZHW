const profileService = require('../../../../services/profile')
const toast = require('../../../../utils/toast')

const DEFAULT_TABS = [
  { key: 'all', label: '全部' },
  { key: 'pending_ship', label: '待发货' },
  { key: 'shipping', label: '配送中' },
  { key: 'completed', label: '已完成' }
]

const STATUS_LABEL_MAP = {
  all: '全部',
  pending_ship: '待发货',
  pendingShip: '待发货',
  shipping: '配送中',
  delivering: '配送中',
  completed: '已完成',
  complete: '已完成'
}

const STATUS_KEY_MAP = {
  '全部': 'all',
  '待发货': 'pending_ship',
  '配送中': 'shipping',
  '已完成': 'completed'
}

const STATUS_TONE_MAP = {
  pending_ship: 'orange',
  pendingShip: 'orange',
  shipping: 'blue',
  delivering: 'blue',
  completed: 'success',
  complete: 'success'
}

const ACTION_LABEL_KEY_MAP = {
  '查看物流': 'logistics',
  '再次兑换': 'again',
  '取消订单': 'cancel',
  '查看详情': 'detail'
}

const BUTTON_STYLE_TYPES = {
  ghost: true,
  primary: true,
  plain: true,
  default: true
}

const AGAIN_ACTION_KEYS = ['again', 'exchange', 'reexchange', 'redeemagain']
const LOGISTICS_ACTION_KEYS = ['logistics', 'viewlogistics', 'tracking', 'track', 'express', 'viewexpress', 'deliverytrace', 'shippingtrace']

function pickFirstValue() {
  const values = Array.prototype.slice.call(arguments)

  for (let index = 0; index < values.length; index += 1) {
    if (values[index] !== undefined && values[index] !== null && values[index] !== '') {
      return values[index]
    }
  }

  return ''
}

function normalizeStatusKey(value) {
  const text = String(value || '').trim()

  return STATUS_KEY_MAP[text] || text
}

function normalizeTabs(tabs) {
  const source = Array.isArray(tabs) && tabs.length ? tabs : DEFAULT_TABS

  return source.map((item) => {
    if (typeof item === 'string') {
      return {
        key: normalizeStatusKey(item),
        label: item
      }
    }

    const key = normalizeStatusKey(pickFirstValue(item.key, item.status, item.value, item.type))
    const label = pickFirstValue(item.label, item.name, item.title, STATUS_LABEL_MAP[key], key)

    return {
      key,
      label,
      count: item.count
    }
  })
}

function normalizeActionLookupKey(value) {
  return String(value || '').trim().replace(/[-_\s]/g, '').toLowerCase()
}

function normalizeActionKey(value, label) {
  const key = String(value || '').trim()
  const labelText = String(label || '').trim()

  if (key && !BUTTON_STYLE_TYPES[key]) {
    return key
  }

  return ACTION_LABEL_KEY_MAP[labelText] || labelText
}

function isAgainAction(actionKey, actionLabel) {
  return actionLabel === '再次兑换' || AGAIN_ACTION_KEYS.indexOf(normalizeActionLookupKey(actionKey)) !== -1
}

function isLogisticsAction(actionKey, actionLabel) {
  return actionLabel === '查看物流' || LOGISTICS_ACTION_KEYS.indexOf(normalizeActionLookupKey(actionKey)) !== -1
}

function normalizeActions(actions) {
  if (!Array.isArray(actions)) {
    return []
  }

  return actions
    .map((item) => {
      if (!item) {
        return null
      }

      if (typeof item === 'string') {
        const label = item.trim()

        return {
          key: normalizeActionKey('', label),
          label,
          type: 'ghost'
        }
      }

      const rawType = String(item.type || '').trim()
      const label = pickFirstValue(item.label, item.text, item.name, item.title)
      const buttonType = pickFirstValue(
        item.buttonType,
        item.style,
        item.variant,
        rawType === 'primary' || rawType === 'ghost' ? rawType : ''
      )
      const rawKey = pickFirstValue(item.key, item.actionKey, item.action, item.value, item.code, item.event, item.type)

      return {
        key: normalizeActionKey(rawKey, label),
        label,
        type: buttonType || 'ghost'
      }
    })
    .filter((item) => item && item.key && item.label)
}

function normalizeOrder(item) {
  const source = item || {}
  const statusKey = normalizeStatusKey(pickFirstValue(
    source.statusKey,
    source.orderStatus,
    source.status,
    source.state
  ))
  const statusLabel = pickFirstValue(
    source.statusText,
    source.statusLabel,
    source.statusName,
    STATUS_LABEL_MAP[statusKey],
    source.status
  )

  const actions = normalizeActions(source.actions || source.actionList)
  return {
    id: pickFirstValue(source.id, source.orderId, source.orderNo, source.orderSn),
    statusKey,
    status: statusLabel,
    statusTone: pickFirstValue(source.statusTone, STATUS_TONE_MAP[statusKey], 'blue'),
    iconText: pickFirstValue(source.iconText, source.goodsIcon, source.icon, '🎁'),
    imageUrl: pickFirstValue(source.imageUrl, source.goodsImageUrl, source.productImageUrl, source.coverUrl),
    title: pickFirstValue(source.title, source.goodsName, source.productName, source.name, '兑换商品'),
    points: pickFirstValue(source.points, source.pointsText, source.costText, source.amountText),
    time: pickFirstValue(
      source.time,
      source.timeText,
      source.exchangedAtText,
      source.createdAtText,
      source.exchangedAt ? `兑换时间: ${source.exchangedAt}` : '',
      source.createdAt ? `兑换时间: ${source.createdAt}` : ''
    ),
    actions
  }
}

function normalizeTimeline(list) {
  if (!Array.isArray(list)) {
    return []
  }

  return list.map((item, index) => ({
    id: pickFirstValue(item.id, item.nodeId, `timeline-${index}`),
    desc: pickFirstValue(item.desc, item.text, item.content, item.title),
    time: pickFirstValue(item.time, item.timeText, item.createdAtText, item.createdAt),
    active: item.active === true || item.current === true || index === 0
  })).filter((item) => item.desc)
}

function normalizeLogisticsData(data) {
  const source = data || {}
  const courier = source.courier || source.express || {}
  const trackingNo = pickFirstValue(
    source.trackingNo,
    source.trackingNumber,
    source.expressNo,
    courier.trackingNo,
    courier.trackingNumber,
    courier.expressNo
  )

  return {
    orderId: pickFirstValue(source.orderId, source.orderNo),
    courier: {
      name: pickFirstValue(source.courierName, source.expressName, courier.name, courier.companyName, '物流公司'),
      trackingNo
    },
    timeline: normalizeTimeline(source.timeline || source.traces || source.events),
    emptyText: pickFirstValue(source.emptyText, '暂无物流信息')
  }
}

function normalizeOrdersData(data) {
  const source = data || {}
  const ordersSource = Array.isArray(source.orders)
    ? source.orders
    : (Array.isArray(source.list) ? source.list : (Array.isArray(data) ? data : []))

  return {
    tabs: normalizeTabs(source.tabs),
    orders: ordersSource.map(normalizeOrder),
    emptyText: pickFirstValue(source.emptyText, '暂无订单')
  }
}

function getEmptyText(activeTab) {
  const label = STATUS_LABEL_MAP[activeTab] || ''

  return label && label !== '全部' ? `暂无${label}订单` : '暂无订单'
}

function normalizeEventValue() {
  const values = Array.prototype.slice.call(arguments)

  for (let index = 0; index < values.length; index += 1) {
    const value = values[index]

    if (value !== undefined && value !== null && value !== '') {
      return String(value).trim()
    }
  }

  return ''
}

Page({
  data: {
    tabs: DEFAULT_TABS,
    activeTab: 'all',
    orders: [],
    isLoading: false,
    emptyText: '暂无订单',
    showLogisticsModal: false,
    isLoadingLogistics: false,
    logisticsOrderId: '',
    logisticsCourier: {
      name: '',
      trackingNo: ''
    },
    logisticsTimeline: [],
    logisticsEmptyText: '暂无物流信息'
  },

  onLoad() {
    this.refreshOrders()
  },

  onPullDownRefresh() {
    this.refreshOrders({
      complete: () => {
        if (typeof wx !== 'undefined' && wx.stopPullDownRefresh) {
          wx.stopPullDownRefresh()
        }
      }
    })
  },

  handleTabTap(event) {
    const tab = event.currentTarget.dataset.key

    if (!tab || tab === this.data.activeTab) {
      return
    }

    this.setData({
      activeTab: tab
    })

    this.refreshOrders({
      statusKey: tab
    })
  },

  async refreshOrders(options = {}) {
    const activeTab = options.statusKey || this.data.activeTab || 'all'
    const requestSeq = (this._ordersRequestSeq || 0) + 1

    this._ordersRequestSeq = requestSeq

    this.setData({
      isLoading: true,
      orders: [],
      emptyText: getEmptyText(activeTab)
    })

    try {
      const result = await profileService.getPointsOrders({
        status: activeTab === 'all' ? '' : activeTab
      })

      if (requestSeq !== this._ordersRequestSeq) {
        return
      }

      const normalized = normalizeOrdersData(result)

      this.setData({
        tabs: normalized.tabs,
        orders: normalized.orders,
        emptyText: normalized.emptyText || getEmptyText(activeTab),
        isLoading: false
      })
    } catch (error) {
      if (requestSeq !== this._ordersRequestSeq) {
        return
      }

      console.warn('get points orders failed', error)

      this.setData({
        orders: [],
        emptyText: error.message || '订单加载失败，请稍后再试',
        isLoading: false
      })
    } finally {
      if (typeof options.complete === 'function') {
        options.complete()
      }
    }
  },

  openPointsMall() {
    wx.navigateTo({
      url: '/pages/profile/asset-center/mall/index',
      fail(error) {
        console.warn('navigate to points mall failed', error)
        toast.info('商城页面打开失败，请稍后再试')
      }
    })
  },

  async openLogistics(orderId) {
    const id = normalizeEventValue(orderId)

    if (!id) {
      toast.info('缺少订单号，暂不能查看物流')
      return
    }

    const requestSeq = (this._logisticsRequestSeq || 0) + 1

    this._logisticsRequestSeq = requestSeq
    this.setData({
      showLogisticsModal: true,
      isLoadingLogistics: true,
      logisticsOrderId: id,
      logisticsCourier: {
        name: '',
        trackingNo: ''
      },
      logisticsTimeline: [],
      logisticsEmptyText: '暂无物流信息'
    })

    try {
      const result = await profileService.getPointsOrderLogistics({
        orderId: id
      })

      if (requestSeq !== this._logisticsRequestSeq) {
        return
      }

      const normalized = normalizeLogisticsData(result)

      this.setData({
        logisticsOrderId: normalized.orderId || id,
        logisticsCourier: normalized.courier,
        logisticsTimeline: normalized.timeline,
        logisticsEmptyText: normalized.emptyText,
        isLoadingLogistics: false
      })
    } catch (error) {
      if (requestSeq !== this._logisticsRequestSeq) {
        return
      }

      console.warn('get points order logistics failed', error)

      this.setData({
        logisticsTimeline: [],
        logisticsEmptyText: error.message || '物流加载失败，请稍后再试',
        isLoadingLogistics: false
      })
    }
  },

  handleCloseLogisticsModal() {
    this.setData({
      showLogisticsModal: false
    })
  },

  handleLogisticsModalContentTap() {},

  handleCopyLogisticsNoTap() {
    const trackingNo = this.data.logisticsCourier.trackingNo

    if (!trackingNo) {
      toast.info('暂无可复制的运单号')
      return
    }

    wx.setClipboardData({
      data: trackingNo,
      success: () => {
        toast.info('运单号已复制')
      }
    })
  },

  handleActionTap(event) {
    const dataset = event.currentTarget.dataset || {}
    const actionKey = normalizeEventValue(dataset.actionKey, dataset.actionkey, dataset.action)
    const actionLabel = normalizeEventValue(dataset.actionLabel, dataset.actionlabel, dataset.label)
    const orderId = normalizeEventValue(dataset.orderId, dataset.orderid)

    if (isAgainAction(actionKey, actionLabel)) {
      this.openPointsMall()
      return
    }

    if (isLogisticsAction(actionKey, actionLabel)) {
      this.openLogistics(orderId)
      return
    }

    toast.developing()
  }
})
