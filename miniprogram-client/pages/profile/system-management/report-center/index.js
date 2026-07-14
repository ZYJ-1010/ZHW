const toast = require('../../../../utils/toast')
const reportService = require('../../services/report')
const fileService = require('../../../../services/file')
const { navigateShellRoute } = require('../../../../utils/shell-nav')

const ASSET_BASE = '/pages/profile/system-management/report-center/assets'
const DEFAULT_MAX_EVIDENCE_COUNT = 9

function chunkRows(items, size) {
  const rows = []
  for (let index = 0; index < items.length; index += size) {
    rows.push(items.slice(index, index + size))
  }
  return rows
}

function normalizeReportTypes(items = []) {
  return items
    .filter((item) => item && item.key && item.label)
    .map((item, index) => ({
      key: item.key,
      label: item.label,
      reportType: item.reportType || 'other',
      order: Number(item.order || index + 1),
      visible: item.visible !== false
    }))
    .filter((item) => item.visible)
    .sort((left, right) => left.order - right.order)
}

function getUploadSlots(imageCount, maxCount = DEFAULT_MAX_EVIDENCE_COUNT) {
  const limit = Math.max(0, Number(maxCount) || DEFAULT_MAX_EVIDENCE_COUNT)
  const restCount = limit - imageCount

  if (restCount <= 0) {
    return []
  }

  if (imageCount === 0) {
    return [0, 1, 2, 3]
  }

  const fillCount = 4 - imageCount % 4
  const slotCount = Math.min(restCount, fillCount || 4)

  return Array.from({ length: slotCount }, (_, index) => index)
}

Page({
  data: {
    activeType: '',
    reportedUser: '',
    reason: '',
    gameId: 0,
    targetUserId: 0,
    reviewId: 0,
    submitting: false,
    reportTypes: [],
    reportTypeMap: {},
    maxEvidenceCount: DEFAULT_MAX_EVIDENCE_COUNT,
    tips: [],
    evidenceImages: [],
    uploadSlots: getUploadSlots(0),
    icons: {
      plus: `${ASSET_BASE}/icon-plus-green.png`,
      warning: `${ASSET_BASE}/icon-warning.png`
    },
    reportTypeRows: []
  },

  onLoad(options = {}) {
    this.setData({
      gameId: Number(options.gameId || 0) || 0,
      targetUserId: Number(options.targetUserId || 0) || 0,
      reviewId: Number(options.reviewId || 0) || 0,
      reportedUser: options.targetName || options.nickname || ''
    })
    this.loadReportConfig()
  },

  async loadReportConfig() {
    try {
      const config = await reportService.getReportConfig()
      const reportTypes = normalizeReportTypes(config.types)
      const reportTypeMap = reportTypes.reduce((result, item) => {
        result[item.key] = item.reportType || 'other'
        return result
      }, {})
      const activeType = reportTypeMap[this.data.activeType]
        ? this.data.activeType
        : (config.defaultType && reportTypeMap[config.defaultType] ? config.defaultType : (reportTypes[0] && reportTypes[0].key) || '')
      const maxEvidenceCount = Math.min(Math.max(Number(config.maxEvidenceCount || DEFAULT_MAX_EVIDENCE_COUNT), 1), DEFAULT_MAX_EVIDENCE_COUNT)

      this.setData({
        activeType,
        reportTypes,
        reportTypeMap,
        reportTypeRows: chunkRows(reportTypes, 3),
        maxEvidenceCount,
        tips: Array.isArray(config.tips) ? config.tips : [],
        uploadSlots: getUploadSlots(this.data.evidenceImages.length, maxEvidenceCount)
      })
    } catch (error) {
      toast.info(error.message || '获取举报配置失败')
    }
  },

  handleTabTap(event) {
    const { target } = event.currentTarget.dataset
    const routeMap = {
      appeals: '/pages/profile/system-management/report-appeals/index',
      records: '/pages/profile/system-management/report-records/index'
    }

    if (routeMap[target]) {
      navigateShellRoute(routeMap[target])
    }
  },

  handleTypeTap(event) {
    const { key } = event.currentTarget.dataset

    if (key) {
      this.setData({
        activeType: key
      })
    }
  },

  handleUserInput(event) {
    this.setData({
      reportedUser: event.detail.value
    })
  },

  handleReasonInput(event) {
    this.setData({
      reason: event.detail.value
    })
  },

  handleUploadTap() {
    const restCount = this.data.maxEvidenceCount - this.data.evidenceImages.length

    if (restCount <= 0) {
      toast.info(`最多上传${this.data.maxEvidenceCount}张证据`)
      return
    }

    if (wx.chooseMedia) {
      wx.chooseMedia({
        count: restCount,
        mediaType: ['image'],
        sourceType: ['album', 'camera'],
        sizeType: ['compressed'],
        success: (res) => {
          const paths = (res.tempFiles || []).map((item) => item.tempFilePath).filter(Boolean)
          this.appendEvidenceImages(paths)
        }
      })
      return
    }

    wx.chooseImage({
      count: restCount,
      sourceType: ['album', 'camera'],
      sizeType: ['compressed'],
      success: (res) => {
        this.appendEvidenceImages(res.tempFilePaths || [])
      }
    })
  },

  appendEvidenceImages(paths = []) {
    if (!paths.length) {
      return
    }

    const nextImages = this.data.evidenceImages.concat(
      paths.map((path, index) => ({
        id: `${Date.now()}-${index}`,
        path
      }))
    ).slice(0, this.data.maxEvidenceCount)

    this.setData({
      evidenceImages: nextImages,
      uploadSlots: getUploadSlots(nextImages.length, this.data.maxEvidenceCount)
    })
  },

  handlePreviewImage(event) {
    const { index } = event.currentTarget.dataset
    const urls = this.data.evidenceImages.map((item) => item.path)

    if (!urls[index]) {
      return
    }

    wx.previewImage({
      urls,
      current: urls[index]
    })
  },

  handleRemoveImage(event) {
    const { index } = event.currentTarget.dataset
    const nextImages = this.data.evidenceImages.filter((_, imageIndex) => imageIndex !== index)

    this.setData({
      evidenceImages: nextImages,
      uploadSlots: getUploadSlots(nextImages.length, this.data.maxEvidenceCount)
    })
  },

  async handleSubmitTap() {
    if (this.data.submitting) {
      return
    }

    if (!this.data.activeType) {
      toast.info('请选择举报类型')
      return
    }

    if (!this.data.gameId) {
      toast.info('缺少局ID，请从局详情、聊天或评价入口发起举报')
      return
    }

    if (!this.data.targetUserId || !this.data.reportedUser.trim()) {
      toast.info('请输入被举报人')
      return
    }

    if (!this.data.reason.trim()) {
      toast.info('请填写举报原因')
      return
    }

    const targetUserId = this.data.targetUserId
    const reportType = this.data.reportTypeMap[this.data.activeType] || 'other'
    const type = this.data.reportTypes.find((item) => item.key === this.data.activeType)
    this.setData({ submitting: true })

    try {
      const fileIds = await fileService.uploadEvidenceImages(
        this.data.evidenceImages.map((item) => item.path).filter(Boolean),
        { bizType: 'report_attachment', objectId: this.data.gameId }
      )
      const content = [
        type ? `举报类型：${type.label}` : '',
        `被举报人：${this.data.reportedUser.trim()}`,
        `举报原因：${this.data.reason.trim()}`,
        fileIds.length ? `证据文件ID：${fileIds.join(',')}` : ''
      ].filter(Boolean).join('\n')
      const result = await reportService.createReport({
        gameId: this.data.gameId,
        targetUserId,
        reportType,
        content,
        fileId: fileIds[0] || 0,
        reviewId: this.data.reviewId || 0
      })
      const report = result.report || result

      toast.success('举报已提交')
      navigateShellRoute(`/pages/profile/system-management/feedback-success/index?source=report&reportId=${report.id || ''}`)
    } catch (error) {
      toast.info(error.message || '提交举报失败')
    } finally {
      this.setData({ submitting: false })
    }
  }
})
