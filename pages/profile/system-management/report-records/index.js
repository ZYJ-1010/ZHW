const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

Page({
  data: {
    activeFilter: 'all',
    stats: [],
    filters: [],
    records: [],
    visibleRecords: []
  },

  onLoad() {
    this.loadRecords()
  },

  async loadRecords() {
    try {
      const data = await profileService.getSystemReportRecords({
        status: this.data.activeFilter
      })
      const records = this.normalizeList(data.records || data.list || data.items)

      this.setData({
        stats: this.normalizeList(data.stats),
        filters: this.normalizeList(data.filters || data.tabs),
        records
      }, () => this.applyFilter(this.data.activeFilter))
    } catch (error) {
      this.setData({
        stats: [],
        filters: [],
        records: [],
        visibleRecords: []
      })
      toast.info(error.message || '举报记录加载失败')
    }
  },

  handleTabTap(event) {
    const { target } = event.currentTarget.dataset
    const routeMap = {
      home: '/pages/profile/system-management/report-center/index',
      appeals: '/pages/profile/system-management/report-appeals/index'
    }

    if (routeMap[target]) {
      wx.redirectTo({ url: routeMap[target] })
    }
  },

  handleFilterTap(event) {
    const { key } = event.currentTarget.dataset

    if (key) {
      this.setData({
        activeFilter: key
      }, () => this.loadRecords())
    }
  },

  applyFilter(key) {
    const visibleRecords = key === 'all'
      ? this.data.records
      : this.data.records.filter((item) => this.getStatusKey(item) === key)

    this.setData({
      activeFilter: key,
      visibleRecords
    })
  },

  handleDetailTap(event) {
    const { id } = event.currentTarget.dataset
    const query = id ? `?recordId=${id}` : ''

    wx.redirectTo({
      url: `/pages/profile/system-management/report-record-detail/index${query}`
    })
  },

  normalizeList(list) {
    return Array.isArray(list) ? list : []
  },

  getStatusKey(item) {
    if (!item || typeof item !== 'object') {
      return ''
    }

    return item.filter || item.statusKey || item.statusType || item.status || ''
  }
})
