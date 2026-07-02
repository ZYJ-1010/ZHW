const reportService = require('../../../../services/report')
const { navigateShellRoute } = require('../../../../utils/shell-nav')

function statusMeta(report) {
  if (report.handleOutcome === 'confirmed') {
    return { filter: 'success', status: '举报成功', statusClass: 'success' }
  }
  if (report.handleOutcome === 'malicious' || report.status === 'closed') {
    return { filter: 'failed', status: report.handleOutcome === 'malicious' ? '恶意举报' : '已关闭', statusClass: 'danger' }
  }
  if (report.status === 'handled') {
    return { filter: 'partial', status: '无法核实', statusClass: 'warning' }
  }
  if (report.status === 'assigned' || report.status === 'appealed') {
    return { filter: 'partial', status: '处理中', statusClass: 'warning' }
  }
  return { filter: 'partial', status: '待处理', statusClass: 'warning' }
}

function formatReportTime(value) {
  return value ? String(value).replace('T', ' ').replace(/:\d{2}(?:\.\d+)?Z?$/, '') : ''
}

function normalizeReportRecord(report) {
  const meta = statusMeta(report)
  const parts = [
    { text: '处理状态：' },
    { text: meta.status, className: meta.statusClass === 'danger' ? 'text-danger' : meta.statusClass === 'success' ? 'text-success' : 'text-warning' }
  ]

  if (report.creditChange) {
    parts.push({ text: `，信用变更 ${report.creditChange}`, className: report.creditChange < 0 ? 'text-danger' : 'text-success' })
  }

  if (report.rewardPoints) {
    parts.push({ text: `，平台奖励 +${report.rewardPoints}积分`, className: 'text-success' })
  }

  return Object.assign({
    id: report.id,
    title: `举报「局 ${report.gameId}」${report.reportType || 'other'}`,
    parts,
    time: formatReportTime(report.createdAt),
    raw: report
  }, meta)
}

Page({
  data: {
    activeFilter: 'all',
    stats: [
      { label: '总举报', value: '0' },
      { label: '举报成功', value: '0' },
      { label: '举报失败', value: '0', className: 'danger' },
      { label: '累计奖励', value: '0积分' }
    ],
    filters: [
      { key: 'all', label: '全部' },
      { key: 'success', label: '举报成功' },
      { key: 'failed', label: '举报失败' },
      { key: 'partial', label: '处理中' }
    ],
    records: [],
    visibleRecords: [],
    page: 1,
    pageSize: 50,
    total: 0,
    hasMore: false
  },

  onLoad() {
    this.loadReports()
  },

  onShow() {
    this.loadReports()
  },

  async loadReports() {
    try {
      const data = await reportService.getMyReports({
        page: this.data.page,
        pageSize: this.data.pageSize
      })
      const items = Array.isArray(data.items) ? data.items.map(normalizeReportRecord) : []
      const totalReward = items.reduce((sum, item) => sum + Number(item.raw.rewardPoints || 0), 0)

      this.setData({
        records: items,
        total: Number(data.total || items.length),
        hasMore: Boolean(data.hasMore),
        stats: [
          { label: '总举报', value: String(data.total || items.length) },
          { label: '举报成功', value: String(items.filter((item) => item.filter === 'success').length) },
          { label: '举报失败', value: String(items.filter((item) => item.filter === 'failed').length), className: 'danger' },
          { label: '累计奖励', value: `${totalReward}积分` }
        ]
      })
      this.applyFilter(this.data.activeFilter)
    } catch (error) {
      wx.showToast({
        title: error.message || '获取举报记录失败',
        icon: 'none'
      })
    }
  },

  handleTabTap(event) {
    const { target } = event.currentTarget.dataset
    const routeMap = {
      home: '/pages/profile/system-management/report-center/index',
      appeals: '/pages/profile/system-management/report-appeals/index'
    }

    if (routeMap[target]) {
      navigateShellRoute(routeMap[target])
    }
  },

  handleFilterTap(event) {
    const { key } = event.currentTarget.dataset

    if (key) {
      this.applyFilter(key)
    }
  },

  applyFilter(key) {
    const visibleRecords = key === 'all'
      ? this.data.records
      : this.data.records.filter((item) => item.filter === key)

    this.setData({
      activeFilter: key,
      visibleRecords
    })
  },

  handleDetailTap(event) {
    const { id } = event.currentTarget.dataset
    const query = id ? `?reportId=${id}` : ''

    navigateShellRoute(`/pages/profile/system-management/report-detail/index${query}`)
  }
})
