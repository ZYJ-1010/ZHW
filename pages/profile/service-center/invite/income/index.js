const trendSeries = [
  { month: '1月', amount: 620 },
  { month: '2月', amount: 760 },
  { month: '3月', amount: 690 },
  { month: '4月', amount: 920 },
  { month: '5月', amount: 1120 },
  { month: '6月', amount: 980 }
]

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
      { label: '本月分润', value: '¥3,280', desc: '较上月 +¥420' },
      { label: '累计分润', value: '¥12,580', desc: '共 342 笔' },
      { label: '活跃成员', value: '86', desc: '本月有收益' },
      { label: '产生分润局数', value: '156', desc: '本月累计' }
    ],
    flows: [
      { icon: '🎯', title: '张大山 · 桌游局分润', time: '06-14 20:30', amount: '+¥80' },
      { icon: '🎲', title: '李小红 · 剧本杀分润', time: '06-14 18:15', amount: '+¥65' },
      { icon: '🃏', title: '王建国 · 狼人杀分润', time: '06-14 15:00', amount: '+¥45' },
      { icon: '🎮', title: '赵小美 · 电竞局分润', time: '06-13 21:45', amount: '+¥38' },
      { icon: '🎭', title: '陈博士 · 密室逃脱分润', time: '06-13 14:20', amount: '+¥52' }
    ]
  }
})
