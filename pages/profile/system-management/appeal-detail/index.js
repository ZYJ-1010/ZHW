const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

Page({
  data: {
    appealId: '',
    basicInfo: [],
    appealReason: '',
    evidence: [],
    originalInfo: [],
    originalReason: '',
    timeline: []
  },

  onLoad(options = {}) {
    const appealId = options.appealId || options.id || ''

    this.setData({
      appealId
    })

    if (appealId) {
      this.loadAppealDetail()
    }
  },

  async loadAppealDetail() {
    try {
      const data = await profileService.getSystemReportAppealDetail({
        appealId: this.data.appealId
      })
      const detail = data.detail || data

      this.setData({
        basicInfo: this.normalizeList(detail.basicInfo),
        appealReason: detail.appealReason || detail.reason || '',
        evidence: this.normalizeList(detail.evidence || detail.attachments),
        originalInfo: this.normalizeList(detail.originalInfo),
        originalReason: detail.originalReason || '',
        timeline: this.normalizeList(detail.timeline)
      })
    } catch (error) {
      this.setData({
        basicInfo: [],
        appealReason: '',
        evidence: [],
        originalInfo: [],
        originalReason: '',
        timeline: []
      })
      toast.info(error.message || '申诉详情加载失败')
    }
  },

  handleBackList() {
    wx.redirectTo({
      url: '/pages/profile/system-management/report-appeals/index'
    })
  },

  async handleWithdrawTap() {
    try {
      await profileService.withdrawSystemReportAppeal({
        appealId: this.data.appealId
      })
      toast.success('申诉已撤回')
      this.loadAppealDetail()
    } catch (error) {
      toast.info(error.message || '撤回申诉失败')
    }
  },

  normalizeList(list) {
    return Array.isArray(list) ? list : []
  }
})
