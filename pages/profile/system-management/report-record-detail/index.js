const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

Page({
  data: {
    icons: {
      check: '/pages/profile/system-management/report-center/assets/icon-check.svg'
    },
    recordId: '',
    basicInfo: [],
    reportReason: '',
    evidence: [],
    resultNote: '',
    timeline: []
  },

  onLoad(options = {}) {
    const recordId = options.recordId || options.id || ''

    this.setData({
      recordId
    })

    if (recordId) {
      this.loadRecordDetail()
    }
  },

  async loadRecordDetail() {
    try {
      const data = await profileService.getSystemReportRecordDetail({
        recordId: this.data.recordId
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
      toast.info(error.message || '处理详情加载失败')
    }
  },

  handleBackList() {
    wx.redirectTo({
      url: '/pages/profile/system-management/report-records/index'
    })
  },

  handleRateTap() {
    toast.developing('处理评价待接入评价接口')
  },

  normalizeList(list) {
    return Array.isArray(list) ? list : []
  }
})
