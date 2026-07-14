const profileService = require('../../../../services/profile')
const toast = require('../../../../utils/toast')
const { navigateShellRoute } = require('../../../../utils/shell-nav')

const DEFAULT_CONFIRM = {
  title: '',
  content: '',
  confirmText: '',
  cancelText: '',
  reason: ''
}

function pickFirstValue() {
  const values = Array.prototype.slice.call(arguments)

  for (let index = 0; index < values.length; index += 1) {
    if (values[index] !== undefined && values[index] !== null && values[index] !== '') {
      return values[index]
    }
  }

  return ''
}

function normalizeKey(value) {
  return String(value || '').trim()
}

function normalizeTabs(tabs) {
  if (!Array.isArray(tabs)) {
    return []
  }

  return tabs
    .map((item) => {
      if (!item) {
        return null
      }

      if (typeof item === 'string') {
        return {
          key: normalizeKey(item),
          label: item
        }
      }

      const key = normalizeKey(pickFirstValue(item.key, item.status, item.value, item.type))
      const label = pickFirstValue(item.label, item.name, item.title)

      return {
        key,
        label,
        count: item.count
      }
    })
    .filter((item) => item && item.key && item.label)
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
        return null
      }

      const key = normalizeKey(pickFirstValue(item.key, item.actionKey, item.action, item.value, item.code, item.event))
      const label = pickFirstValue(item.label, item.text, item.name, item.title)

      return {
        key,
        label,
        type: pickFirstValue(item.buttonType, item.style, item.variant, item.type, 'ghost')
      }
    })
    .filter((item) => item && item.key && item.label)
}

function normalizeOrder(item) {
  const source = item || {}

  return {
    id: pickFirstValue(source.id, source.orderId, source.orderNo, source.orderSn),
    statusKey: pickFirstValue(source.statusKey, source.orderStatus, source.status, source.state),
    status: pickFirstValue(source.statusText, source.statusLabel, source.statusName, source.status),
    statusTone: pickFirstValue(source.statusTone, 'blue'),
    iconText: pickFirstValue(source.iconText, source.goodsIcon, source.icon),
    imageUrl: pickFirstValue(source.imageUrl, source.goodsImageUrl, source.productImageUrl, source.coverUrl),
    title: pickFirstValue(source.title, source.goodsName, source.productName, source.name, source.itemName),
    points: pickFirstValue(source.points, source.pointsText, source.costText, source.amountText),
    time: pickFirstValue(source.time, source.timeText, source.exchangedAtText, source.createdAtText, source.createdAt),
    actions: normalizeActions(source.actions || source.actionList)
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

function normalizeLogisticsData(data, pageConfig) {
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
      name: pickFirstValue(source.courierName, source.expressName, courier.name, courier.companyName),
      trackingNo
    },
    timeline: normalizeTimeline(source.timeline || source.traces || source.events),
    emptyText: pickFirstValue(source.emptyText, pageConfig.logisticsEmptyText)
  }
}

function normalizeOrderDetailData(data, pageConfig) {
  const source = data || {}
  const order = normalizeOrder(source.order || source)
  const rows = Array.isArray(source.detailRows) ? source.detailRows : []

  return {
    order,
    rows: rows.map((item) => ({
      label: pickFirstValue(item.label, item.name, item.title),
      value: pickFirstValue(item.value, item.text, item.content)
    })).filter((item) => item.label && item.value),
    actions: normalizeActions(order.actions || []).filter((item) => item.key !== 'detail'),
    emptyText: pickFirstValue(source.emptyText, pageConfig.detailEmptyText)
  }
}

function normalizePageConfig(source) {
  const config = source || {}

  return {
    emptyText: pickFirstValue(config.emptyText),
    logisticsEmptyText: pickFirstValue(config.logisticsEmptyText),
    detailEmptyText: pickFirstValue(config.detailEmptyText),
    cancelConfirm: Object.assign({}, DEFAULT_CONFIRM, config.cancelConfirm || {})
  }
}

function normalizeOrdersData(data) {
  const source = data || {}
  const pageConfig = normalizePageConfig(source.pageConfig || source)
  const ordersSource = Array.isArray(source.orders)
    ? source.orders
    : (Array.isArray(source.list) ? source.list : (Array.isArray(data) ? data : []))

  return {
    tabs: normalizeTabs(source.tabs),
    orders: ordersSource.map(normalizeOrder),
    emptyText: pickFirstValue(source.emptyText, pageConfig.emptyText),
    pageConfig
  }
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
    tabs: [],
    activeTab: 'all',
    orders: [],
    isLoading: false,
    emptyText: '',
    pageConfig: normalizePageConfig(),
    showLogisticsModal: false,
    isLoadingLogistics: false,
    logisticsOrderId: '',
    logisticsCourier: {
      name: '',
      trackingNo: ''
    },
    logisticsTimeline: [],
    logisticsEmptyText: '',
    showDetailModal: false,
    isLoadingDetail: false,
    detailOrderId: '',
    detailOrder: {},
    detailRows: [],
    detailEmptyText: '',
    detailActions: [],
    cancelingOrderId: ''
  },

  onLoad(options = {}) {
    const status = normalizeEventValue(options.status)
    const orderId = normalizeEventValue(options.orderId, options.id)

    if (status) {
      this.setData({
        activeTab: status
      })
    }

    this.refreshOrders({
      statusKey: status || this.data.activeTab,
      openOrderId: orderId
    })
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
      orders: []
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
        emptyText: normalized.emptyText,
        pageConfig: normalized.pageConfig,
        logisticsEmptyText: normalized.pageConfig.logisticsEmptyText,
        isLoading: false
      })

      if (options.openOrderId) {
        this.openOrderDetail(options.openOrderId)
      }
    } catch (error) {
      if (requestSeq !== this._ordersRequestSeq) {
        return
      }

      console.warn('get points orders failed', error)

      this.setData({
        orders: [],
        emptyText: error.message || this.data.emptyText,
        isLoading: false
      })
    } finally {
      if (typeof options.complete === 'function') {
        options.complete()
      }
    }
  },

  openPointsMall() {
    navigateShellRoute('/pages/profile/asset-center/mall/index')
  },

  async openLogistics(orderId) {
    const id = normalizeEventValue(orderId)

    if (!id) {
      toast.info('缺少订单信息，无法查看物流')
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
      logisticsTimeline: []
    })

    try {
      const result = await profileService.getPointsOrderLogistics({
        orderId: id
      })

      if (requestSeq !== this._logisticsRequestSeq) {
        return
      }

      const normalized = normalizeLogisticsData(result, this.data.pageConfig)

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
        logisticsEmptyText: error.message || this.data.pageConfig.logisticsEmptyText,
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
      toast.info(this.data.pageConfig.logisticsEmptyText || '暂无物流单号')
      return
    }

    wx.setClipboardData({
      data: trackingNo
    })
  },

  async openOrderDetail(orderId) {
    const id = normalizeEventValue(orderId)

    if (!id) {
      toast.info('缺少订单信息，无法查看详情')
      return
    }

    const requestSeq = (this._detailRequestSeq || 0) + 1

    this._detailRequestSeq = requestSeq
    this.setData({
      showDetailModal: true,
      isLoadingDetail: true,
      detailOrderId: id,
      detailOrder: {},
      detailRows: [],
      detailActions: [],
      detailEmptyText: this.data.pageConfig.detailEmptyText
    })

    try {
      const result = await profileService.getPointsOrderDetail({
        orderId: id
      })

      if (requestSeq !== this._detailRequestSeq) {
        return
      }

      const normalized = normalizeOrderDetailData(result, this.data.pageConfig)

      this.setData({
        detailOrder: normalized.order,
        detailRows: normalized.rows,
        detailActions: normalized.actions,
        detailEmptyText: normalized.emptyText,
        isLoadingDetail: false
      })
    } catch (error) {
      if (requestSeq !== this._detailRequestSeq) {
        return
      }

      console.warn('get points order detail failed', error)
      this.setData({
        detailRows: [],
        detailActions: [],
        detailEmptyText: error.message || this.data.pageConfig.detailEmptyText || '',
        isLoadingDetail: false
      })
    }
  },

  handleCloseDetailModal() {
    this.setData({
      showDetailModal: false
    })
  },

  handleDetailModalContentTap() {},

  cancelOrder(orderId) {
    const id = normalizeEventValue(orderId)

    if (!id) {
      toast.info('缺少订单信息，无法取消订单')
      return
    }

    if (this.data.cancelingOrderId === id) {
      return
    }

    const confirmConfig = Object.assign({}, DEFAULT_CONFIRM, this.data.pageConfig.cancelConfirm || {})

    wx.showModal({
      title: confirmConfig.title,
      content: confirmConfig.content,
      confirmText: confirmConfig.confirmText,
      cancelText: confirmConfig.cancelText,
      success: async (res) => {
        if (!res.confirm) {
          return
        }

        try {
          this.setData({
            cancelingOrderId: id
          })

          const result = await profileService.cancelPointsOrder({
            orderId: id,
            reason: confirmConfig.reason
          })
          const updatedOrder = normalizeOrder(result && result.order || {})
          const orders = this.data.orders.map((item) => {
            if (String(item.id) !== id || !updatedOrder.id) {
              return item
            }

            return updatedOrder
          })
          const detailOrder = String(this.data.detailOrderId) === id && updatedOrder.id
            ? updatedOrder
            : this.data.detailOrder
          const detailActions = String(this.data.detailOrderId) === id && updatedOrder.id
            ? normalizeActions(updatedOrder.actions || []).filter((item) => item.key !== 'detail')
            : this.data.detailActions

          toast.info(result && result.message || '订单已取消')
          this.setData({
            orders,
            detailOrder,
            detailActions
          })
          if (this.data.showDetailModal && String(this.data.detailOrderId) === id) {
            this.openOrderDetail(id)
          }
          this.refreshOrders({
            statusKey: this.data.activeTab
          })
        } catch (error) {
          console.warn('cancel points order failed', error)
          toast.info(error.message || '取消订单失败')
        } finally {
          this.setData({
            cancelingOrderId: ''
          })
        }
      }
    })
  },

  handleActionTap(event) {
    const dataset = event.currentTarget.dataset || {}
    const actionKey = normalizeEventValue(dataset.actionKey, dataset.actionkey, dataset.action)
    const actionLabel = normalizeEventValue(dataset.actionLabel, dataset.actionlabel, dataset.label)
    const orderId = normalizeEventValue(dataset.orderId, dataset.orderid)

    if (actionKey === 'again') {
      this.handleCloseDetailModal()
      this.openPointsMall()
      return
    }

    if (actionKey === 'logistics') {
      this.handleCloseDetailModal()
      this.openLogistics(orderId)
      return
    }

    if (actionKey === 'detail') {
      this.openOrderDetail(orderId)
      return
    }

    if (actionKey === 'cancel') {
      this.cancelOrder(orderId)
      return
    }

    toast.info(actionLabel || '暂无可用操作')
  }
})
