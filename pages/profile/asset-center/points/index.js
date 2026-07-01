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
    limitRules: [],
    earnExample: {
      title: '',
      subtitle: '',
      points: '',
      rows: [],
      result: ''
    },
    roleExampleTitle: '',
    roleExamples: [],
    noteText: '',
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
      const points = data || {}
      const roleExamplesSection = points.roleExamplesSection || points.rolePointExamplesSection || {}

      this.setData({
        summary: this.normalizeSummary(points.summary || points.pointsSummary || points),
        rules: this.normalizeList(points.rules),
        limitRules: this.normalizeList(points.limitRules || points.ruleLimits || points.limits || points.restrictions),
        earnExample: points.earnExample || points.example || this.data.earnExample,
        roleExampleTitle: points.roleExampleTitle || points.roleExamplesTitle || roleExamplesSection.title || '',
        roleExamples: this.normalizeList(points.roleExamples || points.rolePointExamples || roleExamplesSection.items),
        noteText: points.noteText || points.noticeText || points.rulesNote || '',
        records: this.normalizeList(points.records || points.list || points.items)
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
