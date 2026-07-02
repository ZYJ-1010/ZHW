const toast = require('../../../../utils/toast')
const reportService = require('../../../../services/report')
const fileService = require('../../../../services/file')
const { navigateShellRoute } = require('../../../../utils/shell-nav')

const LOCAL_ASSET_BASE = '/pages/profile/system-management/credit-appeal/assets'
const BLOCK_ASSET_BASE = '/pages/profile/system-management/block-settings/assets'

Page({
  data: {
    activeReason: '',
    appealContent: '',
    appealContentLength: 0,
    reportId: 0,
    creditLogId: 0,
    submitting: false,
    evidenceImages: [],
    submitClass: 'disabled',
    appealReasons: [],
    icons: {
      ban: `${BLOCK_ASSET_BASE}/icon-ban.svg`,
      chevron: `${BLOCK_ASSET_BASE}/icon-chevron-right.svg`,
      plus: `${BLOCK_ASSET_BASE}/icon-plus.svg`,
      clock: `${LOCAL_ASSET_BASE}/icon-clock.svg`
    },
    reasonOptions: [],
    relatedRecord: {
      title: '管理员处罚 -10分',
      desc: '违规行为 · 02-28'
    },
    appealPlaceholder: '',
    uploadRequirement: '',
    uploadFullText: '',
    uploadSelectedTemplate: '',
    appealFileMaxCount: 0,
    uploadSlots: [],
    reviewTitle: '',
    reviewRules: []
  },

  onLoad(options = {}) {
    const reportId = Number(options.reportId || 0) || 0
    const creditLogId = Number(options.creditLogId || 0) || 0

    this.loadAppealConfig()

    if (reportId) {
      this.setData({ reportId })
      this.loadReport(reportId)
      return
    }

    if (creditLogId) {
      this.setData({
        creditLogId,
        relatedRecord: {
          title: `信用记录 #${creditLogId}`,
          desc: '信用处罚 · 可提交申诉'
        }
      })
    }
  },

  async loadAppealConfig() {
    try {
      const config = await reportService.getReportConfig()
      const reasons = Array.isArray(config.appealReasons)
        ? config.appealReasons.filter((item) => item && item.visible !== false && item.key && item.label)
        : []
      const activeReason = reasons[0] ? reasons[0].key : ''

      this.setData({
        activeReason,
        appealReasons: reasons,
        reasonOptions: this.buildReasonOptions(reasons, activeReason),
        appealPlaceholder: config.appealPlaceholder || '',
        uploadRequirement: config.appealUploadNote || '',
        uploadFullText: config.appealUploadFullText || '',
        uploadSelectedTemplate: config.appealUploadSelectedTemplate || '',
        appealFileMaxCount: this.appealFileMaxCount(config),
        uploadSlots: this.buildUploadSlots(this.appealFileMaxCount(config)),
        reviewTitle: config.appealReviewTitle || '',
        reviewRules: (Array.isArray(config.appealReviewRules) ? config.appealReviewRules : []).map((text, index) => ({
          key: `rule-${index + 1}`,
          text
        }))
      })
    } catch (error) {
      this.setData({
        activeReason: '',
        appealReasons: [],
        reasonOptions: [],
        reviewRules: []
      })
      toast.info(error.message || '获取申诉配置失败')
    }
  },

  buildReasonOptions(reasons = [], activeKey = '') {
    return reasons.map((item) => ({
      key: item.key,
      label: item.label,
      className: [
        'reason-pill',
        item.key === activeKey ? 'active' : '',
        item.key === 'other' ? 'compact' : ''
      ].filter(Boolean).join(' ')
    }))
  },

  appealFileMaxCount(config = {}) {
    const count = Number(config.appealFileMaxCount || config.maxEvidenceCount || 0)

    return Math.max(0, count)
  },

  buildUploadSlots(count) {
    return Array.from({ length: count }, (_, index) => ({
      key: `slot-${index + 1}`
    }))
  },

  async loadReport(reportId) {
    try {
      const report = await reportService.getReportDetail(reportId)

      this.setData({
        relatedRecord: {
          title: `举报记录 RP${String(report.id).padStart(8, '0')}`,
          desc: `${report.reportType || 'other'} · 局ID ${report.gameId}`
        }
      })
    } catch (error) {
      toast.info(error.message || '获取举报记录失败')
    }
  },

  handleReasonTap(event) {
    const { key } = event.currentTarget.dataset

    if (key && key !== this.data.activeReason) {
      this.setData({
        activeReason: key,
        reasonOptions: this.buildReasonOptions(this.data.appealReasons, key)
      })
    }
  },

  handleRecordTap() {
    if (!this.data.reportId) {
      toast.info('暂无关联记录')
      return
    }

    navigateShellRoute(`/pages/profile/system-management/appeal-detail/index?reportId=${this.data.reportId}`)
  },

  handleContentInput(event) {
    const value = event.detail.value || ''

    this.setData({
      appealContent: value,
      appealContentLength: value.length,
      submitClass: value.trim() ? 'ready' : 'disabled'
    })
  },

  handleUploadTap() {
    const restCount = Math.max(0, Number(this.data.appealFileMaxCount || 0) - this.data.evidenceImages.length)

    if (!restCount) {
      toast.info(this.data.uploadFullText || '证明材料数量已达上限')
      return
    }

    const appendImages = (paths = []) => {
      const nextImages = this.data.evidenceImages.concat(paths.map((path, index) => ({
        id: `${Date.now()}-${index}`,
        path
      }))).slice(0, this.data.evidenceImages.length + restCount)

      this.setData({
        evidenceImages: nextImages,
        uploadRequirement: this.uploadSelectedText(nextImages.length)
      })
    }

    if (wx.chooseMedia) {
      wx.chooseMedia({
        count: restCount,
        mediaType: ['image'],
        sourceType: ['album', 'camera'],
        sizeType: ['compressed'],
        success: (res) => {
          appendImages((res.tempFiles || []).map((item) => item.tempFilePath).filter(Boolean))
        }
      })
      return
    }

    wx.chooseImage({
      count: restCount,
      sourceType: ['album', 'camera'],
      sizeType: ['compressed'],
      success: (res) => {
        appendImages(res.tempFilePaths || [])
      }
    })
  },

  uploadSelectedText(selectedCount) {
    const maxCount = Number(this.data.appealFileMaxCount || 0)
    const template = this.data.uploadSelectedTemplate || ''

    if (template) {
      return template.replace('{selected}', selectedCount).replace('{max}', maxCount)
    }

    return `已选择 ${selectedCount}/${maxCount} 张证明材料`
  },

  async handleSubmitTap() {
    if (this.data.submitting) {
      return
    }

    if (!this.data.reportId && !this.data.creditLogId) {
      toast.info('缺少关联记录，无法提交申诉')
      return
    }

    if (!this.data.appealContent.trim()) {
      toast.info('请先填写详细说明')
      return
    }

    if (!this.data.activeReason) {
      toast.info('请选择申诉原因')
      return
    }

    this.setData({ submitting: true })

    try {
      const fileIds = await fileService.uploadEvidenceImages(
        this.data.evidenceImages.map((item) => item.path).filter(Boolean),
        { bizType: 'report_attachment', objectId: this.data.reportId || this.data.creditLogId }
      )
      const payload = {
        reason: this.data.activeReason,
        content: this.data.appealContent.trim(),
        fileId: fileIds[0] || 0,
        fileIds
      }
      const report = this.data.reportId
        ? await reportService.submitAppeal(this.data.reportId, payload)
        : await reportService.submitCreditAppeal(Object.assign({}, payload, { creditLogId: this.data.creditLogId }))

      toast.success('申诉已提交')
      navigateShellRoute(`/pages/profile/system-management/appeal-detail/index?reportId=${report.id || this.data.reportId}`)
    } catch (error) {
      toast.info(error.message || '提交申诉失败')
    } finally {
      this.setData({ submitting: false })
    }
  }
})
