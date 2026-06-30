const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

Page({
  data: {
    activeTab: 'all',
    tabs: [
      { key: 'all', label: '全部' },
      { key: 'processing', label: '处理中' },
      { key: 'resolved', label: '已解决' }
    ],
    records: [],
    allRecords: []
  },

  onLoad() {
    this.loadRecords()
  },

  async loadRecords() {
    try {
      const data = await profileService.getSystemFeedbackRecords({
        status: this.data.activeTab
      })
      const records = this.normalizeList(data.records || data.list || data.items)

      this.setData({
        tabs: this.normalizeList(data.tabs).length ? data.tabs : this.data.tabs,
        allRecords: records
      }, () => this.applyRecordFilter(this.data.activeTab))
    } catch (error) {
      this.setData({
        records: [],
        allRecords: []
      })
      toast.info(error.message || '反馈记录加载失败')
    }
  },

  handleTabTap(event) {
    const { key } = event.currentTarget.dataset

    if (key) {
      this.setData({
        activeTab: key
      }, () => this.loadRecords())
    }
  },

  applyRecordFilter(activeTab) {
    const records = this.data.allRecords.filter((record) => {
      const statusKey = this.getStatusKey(record)

      if (activeTab === 'resolved') {
        return statusKey === 'resolved'
      }

      if (activeTab === 'processing') {
        return statusKey !== 'resolved'
      }

      return true
    })

    this.setData({
      activeTab,
      records
    })
  },

  normalizeList(list) {
    return Array.isArray(list) ? list : []
  },

  getStatusKey(record) {
    if (!record || typeof record !== 'object') {
      return ''
    }

    return record.statusClass || record.statusKey || record.statusType || record.status || ''
  },

  handleRecordTap() {
    wx.navigateTo({
      url: '/pages/profile/system-management/feedback-detail/index'
    })
  }
})
