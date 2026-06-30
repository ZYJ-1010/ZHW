const toast = require('../../../utils/toast')
const profileService = require('../../../services/profile')

Page({
  data: {
    summary: [],
    records: []
  },

  onLoad() {
    this.loadCredit()
  },

  async loadCredit() {
    try {
      const data = await profileService.getProfileCredit()

      this.setData({
        summary: this.normalizeList(data.summary || data.stats || data.items),
        records: this.normalizeList(data.records || data.list)
      })
    } catch (error) {
      this.setData({
        summary: [],
        records: []
      })
      toast.info(error.message || '信用中心加载失败')
    }
  },

  normalizeList(list) {
    return Array.isArray(list) ? list : []
  }
})
