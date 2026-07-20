const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const ASSET_BASE = '/pages/profile/asset-center/points/assets'

function formatNumber(value) {
  const number = Number(value) || 0

  return number.toLocaleString('en-US')
}

function formatTime(value) {
  const date = value ? new Date(value) : null

  if (!date || Number.isNaN(date.getTime())) {
    return ''
  }

  const pad = (input) => String(input).padStart(2, '0')

  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function pointBizTitle(log) {
  const reason = String(log.reason || '').trim()
  const bizType = String(log.bizType || '').trim()

  if (reason) {
    return reason
  }

  const titleMap = {
    report_reward: '举报核实奖励',
    redemption_order: '积分商城兑换',
    redemption_refund: '兑换订单退回',
    service_profit: '服务分润积分',
    guide_reward: '领路人引荐奖励',
    admin_adjust: '后台积分调整'
  }

  return titleMap[bizType] || '积分变动'
}

function normalizePointLog(log = {}) {
  const changeValue = Number(log.changeValue || 0)
  const isIncome = changeValue >= 0
  const bizType = String(log.bizType || '').trim()
  const bizId = log.bizId || log.bizID || ''
  const descParts = []

  if (bizType) {
    descParts.push(`类型：${bizType}`)
  }

  if (bizId) {
    descParts.push(`业务ID：${bizId}`)
  }

  return {
    id: String(log.id || `${bizType || 'points'}-${log.createdAt || Date.now()}`),
    title: pointBizTitle(log),
    desc: descParts.length ? descParts.join(' · ') : '平台积分流水',
    time: formatTime(log.createdAt),
    points: `${isIncome ? '+' : ''}${formatNumber(changeValue)}`,
    changeValue,
    tone: isIncome ? 'plus' : 'minus',
    iconText: isIncome ? '奖' : '兑',
    iconTone: isIncome ? 'green' : 'pink'
  }
}

function statValueByKey(account, logs, key) {
  const normalizedKey = String(key || '').trim()

  if (normalizedKey === 'total') {
    return Number(account.totalEarnedPoints || 0)
  }

  if (normalizedKey === 'redeemed') {
    return Number(account.redeemedPoints || 0) || logs.reduce((sum, item) => {
      const changeValue = Number(item.changeValue || 0)

      return changeValue < 0 ? sum + Math.abs(changeValue) : sum
    }, 0)
  }

  if (normalizedKey === 'expired') {
    return Number(account.expiredPoints || 0)
  }

  return Number(account[normalizedKey] || 0)
}

function buildSummary(account = {}, logs = []) {
  const available = Number(account.availablePoints || 0)
  const statConfig = Array.isArray(account.stats) ? account.stats : []

  return {
    available: formatNumber(available),
    stats: statConfig.map((item) => Object.assign({}, item, {
      value: formatNumber(statValueByKey(account, logs, item.key))
    }))
  }
}

function normalizeFilters(filters) {
  return (Array.isArray(filters) ? filters : [])
    .map((item, index) => {
      if (typeof item === 'string') {
        return { key: index === 0 ? 'all' : `filter_${index}`, label: item, tone: index === 0 ? 'all' : '' }
      }

      return {
        key: String(item && item.key || `filter_${index}`),
        label: String(item && item.label || ''),
        tone: String(item && item.tone || item && item.key || '')
      }
    })
    .filter((item) => item.label)
}

Page({
  data: {
    icons: {
      back: `${ASSET_BASE}/icon-chevron-left.svg`,
      more: `${ASSET_BASE}/icon-ellipsis-vertical.svg`
    },
    loading: false,
    summary: buildSummary(),
    rules: [],
    earnExample: {},
    roleExamples: [],
    filters: [],
    activeFilter: 'all',
    noteText: '',
    records: [],
    visibleRecords: []
  },

  onLoad() {
    this.loadPointsCenter()
  },

  async loadPointsCenter() {
    this.setData({ loading: true })

    try {
      const [summary, logResult] = await Promise.all([
        profileService.getPointsSummary(),
        profileService.getPointsLogs()
      ])
      const rawLogs = Array.isArray(logResult && logResult.items) ? logResult.items : []
      const records = rawLogs.map(normalizePointLog)
      const filters = normalizeFilters(summary && summary.filters)
      const activeFilter = filters.some((item) => item.key === this.data.activeFilter)
        ? this.data.activeFilter
        : (filters[0] && filters[0].key || 'all')

      this.setData({
        summary: buildSummary(summary, rawLogs),
        rules: Array.isArray(summary && summary.rules) ? summary.rules : [],
        earnExample: summary && summary.earnExample || {},
        roleExamples: Array.isArray(summary && summary.roleExamples) ? summary.roleExamples : [],
        filters,
        activeFilter,
        noteText: summary && summary.noteText || '',
        records,
        loading: false
      })
      this.applyFilter(activeFilter, records)
    } catch (error) {
      this.setData({ loading: false })
      toast.info(error.message || '积分中心加载失败')
    }
  },

  applyFilter(filter, sourceRecords) {
    const records = Array.isArray(sourceRecords) ? sourceRecords : this.data.records
    const filterItem = this.data.filters.find((item) => item.key === filter) || {}
    const tone = filterItem.tone || filter
    let visibleRecords = records

    if (tone === 'income') {
      visibleRecords = records.filter((item) => Number(item.changeValue || 0) > 0)
    } else if (tone === 'expense') {
      visibleRecords = records.filter((item) => Number(item.changeValue || 0) < 0)
    }

    this.setData({ visibleRecords })
  },

  handleFilterTap(event) {
    const filter = event.currentTarget.dataset.filter || 'all'

    this.setData({ activeFilter: filter })
    this.applyFilter(filter)
  },

  handleDeveloping() {
    toast.info('积分明细以后台流水为准')
  }
})
