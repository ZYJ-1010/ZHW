const profileService = require('../../../../../services/profile')

Page({
  data: {
    activeRole: 'referred',
    activeStatus: 'all',
    roleTabs: [
      { key: 'referred', label: '我引荐的' },
      { key: 'created', label: '我发起的' }
    ],
    filters: [
      { key: 'all', label: '全部' },
      { key: 'progress', label: '进行中' },
      { key: 'completed', label: '已完成' },
      { key: 'timeout', label: '超时' },
      { key: 'cancelled', label: '已取消' }
    ],
    records: [],
    allRecords: []
  },

  onLoad() {
    this.loadRecords()
  },

  async loadRecords() {
    try {
      const data = await profileService.getInviteRecords({
        role: this.data.activeRole,
        status: this.data.activeStatus
      })

      this.setData({
        filters: Array.isArray(data.filters) && data.filters.length ? data.filters : this.data.filters,
        allRecords: Array.isArray(data.records || data.list || data.items) ? (data.records || data.list || data.items) : []
      }, () => this.applyFilters())
    } catch (error) {
      wx.showToast({
        title: error.message || '邀请记录加载失败',
        icon: 'none'
      })
    }
  },

  handleRoleTap(event) {
    const { key } = event.currentTarget.dataset

    this.setData({
      activeRole: key
    }, () => this.loadRecords())
  },

  handleFilterTap(event) {
    const { key } = event.currentTarget.dataset

    this.setData({
      activeStatus: key
    }, () => this.loadRecords())
  },

  handleWarningTap() {
    this.setData({
      activeRole: 'referred',
      activeStatus: 'timeout'
    }, () => this.applyFilters())
  },

  handleRecordTap(event) {
    const { id } = event.currentTarget.dataset

    wx.showToast({
      title: id ? '查看组局' : '暂无组局',
      icon: 'none'
    })
  },

  applyFilters() {
    const { activeRole, activeStatus, allRecords } = this.data
    const records = allRecords.filter((item) => {
      const roleMatched = item.role === activeRole
      const statusMatched = activeStatus === 'all' || item.statusKey === activeStatus

      return roleMatched && statusMatched
    })

    this.setData({ records })
  }
})
