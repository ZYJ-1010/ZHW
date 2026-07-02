const profileService = require('../../../../services/profile')

function markActive(items = [], activeKey) {
  return items.map((item) => ({
    ...item,
    active: item.key === activeKey || item.label === activeKey
  }))
}

Page({
  data: {
    activeRole: 'referred',
    activeStatus: 'all',
    emptyText: '暂无邀约记录',
    roleTabs: [
      { key: 'referred', label: '我引荐的', active: true },
      { key: 'created', label: '我发起的', active: false }
    ],
    filters: [
      { key: 'all', label: '全部', active: true },
      { key: 'progress', label: '进行中', active: false },
      { key: 'completed', label: '已完成', active: false },
      { key: 'timeout', label: '超时', active: false },
      { key: 'cancelled', label: '已取消', active: false }
    ],
    records: []
  },

  onLoad() {
    this.loadRecords()
  },

  async loadRecords() {
    try {
      const result = await profileService.getInviteRecords({
        role: this.data.activeRole,
        status: this.data.activeStatus
      })

      this.setData({
        activeRole: result.activeRole || this.data.activeRole,
        activeStatus: result.activeStatus || this.data.activeStatus,
        roleTabs: markActive(result.roleTabs || this.data.roleTabs, result.activeRole || this.data.activeRole),
        filters: markActive(result.filters || this.data.filters, result.activeStatus || this.data.activeStatus),
        records: Array.isArray(result.records) ? result.records : [],
        emptyText: result.emptyText || this.data.emptyText
      })
    } catch (error) {
      this.setData({
        records: [],
        emptyText: error.message || '邀约记录加载失败'
      })
    }
  }
})
