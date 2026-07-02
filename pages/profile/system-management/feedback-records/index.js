const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')
const { navigateShellRoute } = require('../../../../utils/shell-nav')

const ASSET_BASE = '/pages/profile/system-management/feedback/assets'

Page({
  data: {
    activeTab: 'all',
    tabs: [
      { key: 'all', label: '全部' },
      { key: 'processing', label: '处理中', count: 0 },
      { key: 'resolved', label: '已解决' }
    ],
    records: [],
    allRecords: []
  },

  onLoad() {
    this.loadRecords(this.data.activeTab)
  },

  handleTabTap(event) {
    const { key } = event.currentTarget.dataset

    if (key) {
      this.loadRecords(key)
    }
  },

  async loadRecords(activeTab) {
    try {
      const data = await profileService.getSystemFeedbackRecords({ tab: activeTab })
      this.setData({
        activeTab: data.activeTab || activeTab,
        tabs: Array.isArray(data.tabs) && data.tabs.length ? data.tabs : this.data.tabs,
        records: Array.isArray(data.records) ? data.records : [],
        allRecords: Array.isArray(data.allRecords) ? data.allRecords : this.data.allRecords
      })
    } catch (error) {
      toast.info(error.message || '反馈记录暂时不可用')
      this.setData({
        activeTab,
        records: [],
        allRecords: []
      })
    }
  },

  applyRecordFilter(activeTab) {
    const records = this.data.allRecords.filter((record) => {
      if (activeTab === 'resolved') {
        return record.statusClass === 'resolved'
      }

      if (activeTab === 'processing') {
        return record.statusClass !== 'resolved'
      }

      return true
    })

    this.setData({
      activeTab,
      records
    })
  },

  handleRecordTap(event) {
    const { id } = event.currentTarget.dataset
    const query = id ? `?id=${encodeURIComponent(id)}` : ''

    navigateShellRoute(`/pages/profile/system-management/feedback-detail/index${query}`)
  }
})
