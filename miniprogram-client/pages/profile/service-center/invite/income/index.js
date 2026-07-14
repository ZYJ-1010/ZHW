const profileService = require('../../../../../services/profile')

const trendSeries = []

const TREND_PLOT_AXIS_RATIO = 186 / 508
const TREND_MIN_BOTTOM = 46
const TREND_MAX_BOTTOM = 78
const TREND_MID_BOTTOM = 62

function buildTrendPoints(series) {
  const list = Array.isArray(series) ? series : []
  const values = list.map((item) => Number(item.amount) || 0)
  const max = Math.max(...values, 0)
  const min = Math.min(...values, 0)
  const range = max - min
  const chartRange = TREND_MAX_BOTTOM - TREND_MIN_BOTTOM

  return list.map((item, index) => {
    const left = list.length === 1 ? 50 : Number((index * 100 / (list.length - 1)).toFixed(2))
    const amount = Number(item.amount) || 0
    const bottom = range === 0
      ? TREND_MID_BOTTOM
      : Number((TREND_MIN_BOTTOM + ((amount - min) / range * chartRange)).toFixed(2))

    return {
      ...item,
      left,
      bottom
    }
  })
}

function buildTrendSegments(points) {
  if (!Array.isArray(points) || points.length < 2) {
    return []
  }

  return points.slice(0, -1).map((point, index) => {
    const next = points[index + 1]
    const deltaX = next.left - point.left
    const deltaY = next.bottom - point.bottom
    const scaledDeltaY = deltaY * TREND_PLOT_AXIS_RATIO

    return {
      key: `${point.month}-${next.month}`,
      left: point.left,
      bottom: point.bottom,
      width: Number(Math.sqrt((deltaX * deltaX) + (scaledDeltaY * scaledDeltaY)).toFixed(2)),
      angle: Number((Math.atan2(-scaledDeltaY, deltaX) * 180 / Math.PI).toFixed(2))
    }
  })
}

function buildTrendShadow(points) {
  if (!Array.isArray(points) || points.length < 2) {
    return ''
  }

  const topLine = points
    .map((point) => `${point.left}% ${100 - point.bottom}%`)
    .join(', ')

  return `polygon(${topLine}, 100% 100%, 0 100%)`
}

const trendPoints = buildTrendPoints(trendSeries)

Page({
  data: {
    trendPoints,
    trendSegments: buildTrendSegments(trendPoints),
    trendShadowPath: buildTrendShadow(trendPoints),
    metrics: [
      { label: '本月分润', value: '¥0', desc: '已结算收益' },
      { label: '累计分润', value: '¥0', desc: '含待结算收益' },
      { label: '活跃成员', value: '0', desc: '当前关系数' },
      { label: '产生分润局数', value: '0', desc: '累计流水' }
    ],
    flows: [],
    loadError: ''
  },

  onLoad() {
    this.loadIncome()
  },

  async loadIncome() {
    try {
      const result = await profileService.getInviteIncome()
      const series = Array.isArray(result.trendSeries) && result.trendSeries.length
        ? result.trendSeries
        : trendSeries
      const points = buildTrendPoints(series)

      this.setData(Object.assign({}, result, {
        trendPoints: points,
        trendSegments: buildTrendSegments(points),
        trendShadowPath: buildTrendShadow(points)
      }))
    } catch (error) {
      console.warn('get invite income failed', error)
      this.setData({
        loadError: error.message || '收益明细加载失败'
      })
    }
  }
})
