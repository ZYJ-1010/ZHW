const profileService = require('../../../../services/profile')
const { navigateShellRoute } = require('../../../../utils/shell-nav')

function parseNumber(value, fallback) {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value
  }

  const number = Number(String(value || '').replace(/[^\d.-]/g, ''))

  return Number.isFinite(number) ? number : fallback
}

function formatNumber(value) {
  const number = Math.max(parseNumber(value, 0), 0)

  return String(Math.round(number)).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
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

function buildFallbackMap(goods) {
  return (goods || []).reduce((map, item) => {
    if (item && item.id) {
      map[item.id] = item
    }

    return map
  }, {})
}

function normalizeGood(item, pointsAvailable, fallback) {
  const source = item || {}
  const base = fallback || {}
  const id = pickFirstValue(source.id, source.goodId, source.productId, base.id)
  const cost = parseNumber(pickFirstValue(source.cost, source.pointsCost, source.pointsValue, base.cost), 0)
  const stockLeft = parseNumber(pickFirstValue(source.stockLeft, source.stockCount, source.remainingStock, base.stockLeft), 0)
  const remaining = pickFirstValue(source.remaining, source.remainingPoints, formatNumber(Math.max(pointsAvailable - cost, 0)))

  return {
    id,
    iconText: pickFirstValue(source.iconText, source.icon, base.iconText, '🎁'),
    imageUrl: pickFirstValue(source.imageUrl, source.goodsImageUrl, source.productImageUrl, source.coverUrl, base.imageUrl),
    title: pickFirstValue(source.title, source.name, source.productName, base.title, '兑换商品'),
    cost,
    remaining,
    points: pickFirstValue(source.points, source.pointsText, `${formatNumber(cost)}积分`),
    stock: pickFirstValue(source.stock, source.stockText, `库存: 剩余${stockLeft}件`),
    stockLeft,
    canExchange: source.canExchange !== false && stockLeft > 0
  }
}

function normalizeMallData(data, fallbackGoods) {
  const mall = data || {}
  const fallback = fallbackGoods && fallbackGoods.length ? fallbackGoods : []
  const pointsAvailable = parseNumber(
    pickFirstValue(mall.pointsAvailable, mall.availablePoints, mall.pointsBalance, mall.points, mall.pointsText),
    0
  )
  const goodsSource = Array.isArray(mall.goods)
    ? mall.goods
    : (Array.isArray(mall.items) ? mall.items : fallback)
  const fallbackMap = buildFallbackMap(fallback)

  return {
    points: pickFirstValue(mall.pointsText, mall.points, formatNumber(pointsAvailable)),
    expireTip: pickFirstValue(mall.expireTip, ''),
    goods: goodsSource.map((item, index) => normalizeGood(item, pointsAvailable, fallbackMap[item && item.id] || fallback[index]))
  }
}

Page({
  data: {
    points: '0',
    expireTip: '',
    showExchangeModal: false,
    selectedGood: null,
    isExchanging: false,
    goods: []
  },

  onLoad() {
    this.refreshMallGoods()
  },

  async refreshMallGoods(options = {}) {
    try {
      const mall = await profileService.getPointsMall()

      this.applyMallData(mall)

      return mall
    } catch (error) {
      if (!options.silent) {
        console.warn('get points mall failed', error)
      }

      return null
    }
  },

  applyMallData(mall) {
    const normalized = normalizeMallData(mall, this.data.goods)

    this.setData(normalized)
  },

  handleDetailTap(event) {
    const { id } = event.currentTarget.dataset
    const selectedGood = this.data.goods.find((item) => item.id === id)

    if (!selectedGood) {
      return
    }

    this.setData({
      selectedGood,
      showExchangeModal: true
    })
  },

  handleCloseExchangeModal() {
    this.setData({
      showExchangeModal: false,
      selectedGood: null
    })
  },

  handleModalContentTap() {},

  async handleConfirmExchange() {
    const selectedGood = this.data.selectedGood

    if (this.data.isExchanging || !selectedGood) {
      return
    }

    this.setData({
      isExchanging: true
    })

    try {
      const result = await profileService.exchangePointsMallGood({
        goodId: selectedGood.id
      })
      const responseData = result.data || {}
      const mallSnapshot = responseData.mall || responseData.mallSnapshot || responseData.pointsMall

      if (mallSnapshot) {
        this.applyMallData(mallSnapshot)
      }

      await this.refreshMallGoods({
        silent: true
      })

      this.setData({
        showExchangeModal: false,
        selectedGood: null
      })

      wx.showModal({
        title: result.success ? '兑换成功' : '兑换失败',
        content: result.message || (result.success ? '兑换成功，订单已进入待发货' : '兑换失败，请稍后再试'),
        showCancel: false,
        confirmText: '知道了'
      })
    } catch (error) {
      await this.refreshMallGoods({
        silent: true
      })

      this.setData({
        showExchangeModal: false,
        selectedGood: null
      })

      wx.showModal({
        title: '兑换失败',
        content: error.message || '兑换失败，请稍后再试',
        showCancel: false,
        confirmText: '知道了'
      })
    } finally {
      this.setData({
        isExchanging: false
      })
    }
  },

  handlePointsTap() {
    navigateShellRoute('/pages/profile/asset-center/points/index')
  }
})
