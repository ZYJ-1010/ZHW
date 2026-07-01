const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

Page({
  data: {
    icons: {
      check: '/pages/profile/system-management/report-center/assets/icon-check.svg'
    },
    reportId: '',
    statusCard: {
      main: '',
      sub: '',
      tone: ''
    },
    basicInfo: [],
    reportReason: '',
    evidence: [],
    evidenceTitle: '证据材料',
    resultRows: [],
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
      const evidence = this.normalizeList(detail.evidence || detail.attachments)

      this.setData({
        statusCard: this.normalizeStatusCard(detail),
        basicInfo: this.normalizeList(detail.basicInfo),
        reportReason: detail.reportReason || detail.reason || '',
        evidence,
        evidenceTitle: this.getEvidenceTitle(detail, evidence),
        resultRows: this.normalizeResultRows(detail.resultRows || detail.results || detail.resultItems),
        resultNote: detail.resultNote || detail.result || '',
        timeline: this.normalizeList(detail.timeline)
      })
    } catch (error) {
      this.setData({
        statusCard: {
          main: '',
          sub: '',
          tone: ''
        },
        basicInfo: [],
        reportReason: '',
        evidence: [],
        evidenceTitle: '证据材料',
        resultRows: [],
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
  },

  normalizeStatusCard(detail = {}) {
    const card = detail.statusCard || detail.status || {}

    return {
      main: this.pickText(card.main, card.title, card.text, detail.statusText),
      sub: this.pickText(card.sub, card.subtitle, card.desc, detail.statusDesc),
      tone: card.tone || card.type || card.statusTone || ''
    }
  },

  normalizeResultRows(list) {
    return this.normalizeList(list).map((item) => ({
      label: this.pickText(item.label, item.title, item.name),
      value: this.pickText(item.value, item.text, item.result),
      tone: this.normalizeTextTone(item.tone || item.type || item.status)
    })).filter((item) => item.label || item.value)
  },

  normalizeTextTone(tone = '') {
    return tone && tone.indexOf('text-') !== 0 ? `text-${tone}` : tone
  },

  getEvidenceTitle(detail = {}, evidence = []) {
    const title = this.pickText(detail.evidenceTitle, detail.evidenceSectionTitle)

    if (title) {
      return title
    }

    return evidence.length ? `证据材料 (${evidence.length}张)` : '证据材料'
  },

  pickText(...values) {
    const value = values.find((item) => item || item === 0)

    return value || value === 0 ? String(value) : ''
  }
})
