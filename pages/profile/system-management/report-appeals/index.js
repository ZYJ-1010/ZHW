const reportService = require('../../../../services/report')
const { navigateShellRoute } = require('../../../../utils/shell-nav')

function formatTime(value) {
  return value ? String(value).replace('T', ' ').replace(/:\d{2}(?:\.\d+)?Z?$/, '') : ''
}

function normalizeAppeal(report) {
  return {
    id: report.id,
    title: `被举报「${report.reportType || '申诉'}」`,
    filter: report.status === 'appealed' ? 'processing' : 'resolved',
    status: report.status === 'appealed' ? '处理中' : '已处理',
    statusClass: report.status === 'closed' ? 'danger' : 'success',
    reason: report.handleResult || report.content || '',
    time: formatTime(report.handledAt || report.createdAt),
    raw: report
  }
}

Page({
  data: {
    activeFilter: 'all',
    filters: [
      { key: 'all', label: '全部' },
      { key: 'pending', label: '待处理' },
      { key: 'processing', label: '处理中' },
      { key: 'resolved', label: '已处理' }
    ],
    appeals: [],
    visibleAppeals: [],
    emptyText: '暂无申诉记录',
    page: 1,
    pageSize: 50,
    total: 0,
    hasMore: false
  },

  onLoad() {
    this.loadAppeals()
  },

  onShow() {
    this.loadAppeals()
  },

  async loadAppeals() {
    try {
      const data = await reportService.getMyAppeals({
        page: this.data.page,
        pageSize: this.data.pageSize
      })
      const appeals = Array.isArray(data.items) ? data.items.map(normalizeAppeal) : []

      this.setData({
        appeals,
        total: Number(data.total || appeals.length),
        hasMore: Boolean(data.hasMore)
      })
      this.applyFilter(this.data.activeFilter)
    } catch (error) {
      wx.showToast({
        title: error.message || '获取申诉记录失败',
        icon: 'none'
      })
    }
  },

  handleTabTap(event) {
    const { target } = event.currentTarget.dataset
    const routeMap = {
      home: '/pages/profile/system-management/report-center/index',
      records: '/pages/profile/system-management/report-records/index'
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
    const visibleAppeals = key === 'all'
      ? this.data.appeals
      : this.data.appeals.filter((item) => item.filter === key)

    this.setData({
      activeFilter: key,
      visibleAppeals
    })
  },

  handleDetailTap(event) {
    const { id } = event.currentTarget.dataset
    const query = id ? `?reportId=${id}` : ''

    navigateShellRoute(`/pages/profile/system-management/appeal-detail/index${query}`)
  }
})
