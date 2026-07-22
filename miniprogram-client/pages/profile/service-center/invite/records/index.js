const profileService = require('../../../../../services/profile')
const { navigateShellRoute } = require('../../../../../utils/shell-nav')

Page({
  data: {
    activeStatus: 'all',
    filters: [
      { key: 'all', label: '全部' },
      { key: 'progress', label: '进行中' },
      { key: 'completed', label: '已完成' },
      { key: 'timeout', label: '超时' },
      { key: 'cancelled', label: '已取消' }
    ],
    records: [],
    allRecords: [],
    warning: null
  },

  onLoad() {
    this.loadRecords()
  },

  handleFilterTap(event) {
    const { key } = event.currentTarget.dataset

    this.setData({
      activeStatus: key
    }, () => this.loadRecords())
  },

  handleWarningTap() {
    this.setData({
      activeStatus: 'timeout'
    }, () => this.loadRecords())
  },

  handleRecordTap(event) {
    const { id } = event.currentTarget.dataset
    const record = (this.data.records || []).find((item) => String(item.id) === String(id))
    const route = record && (record.detailRoute || record.gameRoute || record.route)

    if (route) {
      navigateShellRoute(route.indexOf('/') === 0 ? route : `/${route}`)
      return
    }

    wx.showToast({
      title: id ? '暂无详情入口' : '暂无组局',
      icon: 'none'
    })
  },

  async loadRecords() {
    try {
      const result = await profileService.getInviteRecords({
        status: this.data.activeStatus
      })

      this.setData({
        ...(result || {}),
        warning: result && result.timeoutWarning && result.timeoutWarning.show
          ? { title: result.timeoutWarning.title, desc: result.timeoutWarning.text }
          : null
      })
    } catch (error) {
      console.warn('get invite records failed', error)
      this.setData({ records: [] })
    }
  },

  applyFilters() {
    const { activeStatus, allRecords } = this.data
    const records = allRecords.filter((item) => {
      const statusMatched = activeStatus === 'all' || item.statusKey === activeStatus

      return statusMatched
    })

    this.setData({ records })
  }
})
