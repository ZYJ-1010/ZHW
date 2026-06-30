const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

Page({
  data: {
    icons: {
      check: '/pages/profile/system-management/report-center/assets/icon-check.svg'
    },
    reportId: '',
    basicInfo: [],
    reportReason: '',
    evidence: [],
    resultNote: '',
    timeline: []
  },

  onLoad(options = {}) {
    const reportId = options.reportId || options.id || ''

    this.setData({
      reportId
    })

    if (reportId) {
      this.loadReportDetail()
    }
  },

  async loadReportDetail() {
    try {
      const data = await profileService.getSystemReportDetail({
        reportId: this.data.reportId
      })
      const detail = data.detail || data

      this.setData({
        basicInfo: this.normalizeList(detail.basicInfo),
        reportReason: detail.reportReason || detail.reason || '',
        evidence: this.normalizeList(detail.evidence || detail.attachments),
        resultNote: detail.resultNote || detail.result || '',
        timeline: this.normalizeList(detail.timeline)
      })
    } catch (error) {
      this.setData({
        basicInfo: [],
        reportReason: '',
        evidence: [],
        resultNote: '',
        timeline: []
      })
      toast.info(error.message || '举报详情加载失败')
    }
  },

  handleBackList() {
    wx.redirectTo({
      url: '/pages/profile/system-management/report-records/index'
    })
  },

  handleAppealTap() {
    wx.navigateTo({
      url: '/pages/profile/system-management/report-appeals/index'
    })
  },

  normalizeList(list) {
    return Array.isArray(list) ? list : []
  }
})
