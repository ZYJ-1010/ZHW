const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const ASSET_BASE = '/pages/profile/asset-center/points/assets'

Page({
  data: {
    icons: {
      back: `${ASSET_BASE}/icon-chevron-left.svg`,
      more: `${ASSET_BASE}/icon-ellipsis-vertical.svg`
    },
    summary: {
      available: '',
      stats: []
    },
    rules: [],
    earnExample: {
      title: '',
      subtitle: '',
      points: '',
      rows: [],
      result: ''
    },
    roleExamples: [],
    filters: ['全部', '收入', '支出'],
    activeFilter: '全部',
    records: []
  },

  onLoad() {
    this.loadPoints()
  },

  async loadPoints() {
    try {
      const data = await profileService.getProfilePoints({
        filter: this.data.activeFilter
      })

      this.setData({
        summary: this.normalizeSummary(data.summary || data.pointsSummary || data),
        rules: this.normalizeList(data.rules),
        earnExample: data.earnExample || data.example || this.data.earnExample,
        roleExamples: this.normalizeList(data.roleExamples || data.rolePointExamples),
        records: this.normalizeList(data.records || data.list || data.items)
      })
    } catch (error) {
      toast.info(error.message || '积分信息加载失败')
    }
  },

  normalizeSummary(summary = {}) {
    return {
      available: summary.available || summary.availablePoints || summary.points || '',
      stats: this.normalizeList(summary.stats || summary.items)
    }
  },

  normalizeList(list) {
    return Array.isArray(list) ? list : []
  },

  handleFilterTap(event) {
    const filter = event.currentTarget.dataset.filter

    this.setData({
      activeFilter: filter
    })
    this.loadPoints()
  },

  handleDeveloping() {
    toast.developing()
  }
})
