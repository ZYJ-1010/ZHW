const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

Page({
  data: {
    activeFilter: 'all',
    filters: [],
    appeals: [],
    visibleAppeals: []
  },

  onLoad() {
    this.loadAppeals()
  },

  async loadAppeals() {
    try {
      const data = await profileService.getSystemReportAppeals({
        status: this.data.activeFilter
      })
      const appeals = this.normalizeList(data.appeals || data.list || data.items)

      this.setData({
        filters: this.normalizeList(data.filters || data.tabs),
        appeals
      }, () => this.applyFilter(this.data.activeFilter))
    } catch (error) {
      this.setData({
        filters: [],
        appeals: [],
        visibleAppeals: []
      })
      toast.info(error.message || '申诉列表加载失败')
    }
  },

  handleTabTap(event) {
    const { target } = event.currentTarget.dataset
    const routeMap = {
      home: '/pages/profile/system-management/report-center/index',
      records: '/pages/profile/system-management/report-records/index'
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
      }, () => this.loadAppeals())
    }
  },

  applyFilter(key) {
    const visibleAppeals = key === 'all'
      ? this.data.appeals
      : this.data.appeals.filter((item) => this.getStatusKey(item) === key)

    this.setData({
      activeFilter: key,
      visibleAppeals
    })
  },

  handleDetailTap(event) {
    const { id } = event.currentTarget.dataset
    const query = id ? `?appealId=${id}` : ''

    wx.redirectTo({
      url: `/pages/profile/system-management/appeal-detail/index${query}`
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
