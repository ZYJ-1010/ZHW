const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const ASSET_BASE = '/pages/profile/system-management/report-center/assets'
const MAX_EVIDENCE_COUNT = 9

function getUploadSlots(imageCount) {
  const restCount = MAX_EVIDENCE_COUNT - imageCount

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
    evidenceImages: [],
    uploadSlots: getUploadSlots(0),
    icons: {
      plus: `${ASSET_BASE}/icon-plus-green.png`,
      warning: `${ASSET_BASE}/icon-warning.png`
    },
    reportTypeRows: [],
    tipLines: []
  },

  onLoad() {
    this.loadReportOptions()
  },

  async loadReportOptions() {
    try {
      const data = await profileService.getSystemReportOptions()
      const reportTypes = this.normalizeList(data.reportTypes || data.types)
        .map((item) => ({
          ...item,
          key: item.key || item.type || item.id
        }))

      this.setData({
        activeType: data.defaultType || reportTypes[0] && reportTypes[0].key || '',
        reportTypeRows: this.buildRows(reportTypes),
        tipLines: this.normalizeTipLines(data.tipLines || data.tips)
      })
    } catch (error) {
      this.setData({
        activeType: '',
        reportTypeRows: [],
        tipLines: []
      })
      toast.info(error.message || '举报配置加载失败')
    }
  },

  handleTabTap(event) {
    const { target } = event.currentTarget.dataset
    const routeMap = {
      appeals: '/pages/profile/system-management/report-appeals/index',
      records: '/pages/profile/system-management/report-records/index'
    }

    if (routeMap[target]) {
      wx.redirectTo({ url: routeMap[target] })
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
    const restCount = MAX_EVIDENCE_COUNT - this.data.evidenceImages.length

    if (restCount <= 0) {
      toast.info(`最多上传${MAX_EVIDENCE_COUNT}张证据`)
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
    ).slice(0, MAX_EVIDENCE_COUNT)

    this.setData({
      evidenceImages: nextImages,
      uploadSlots: getUploadSlots(nextImages.length)
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
      uploadSlots: getUploadSlots(nextImages.length)
    })
  },

  async handleSubmitTap() {
    if (!this.data.activeType) {
      toast.info('请选择举报类型')
      return
    }

    if (!this.data.reportedUser.trim()) {
      toast.info('请输入被举报人')
      return
    }

    if (!this.data.reason.trim()) {
      toast.info('请填写举报原因')
      return
    }

    try {
      const data = await profileService.submitSystemReport({
        type: this.data.activeType,
        reportedUser: this.data.reportedUser,
        reason: this.data.reason,
        evidenceFileIds: this.data.evidenceImages.map((item) => item.fileId).filter(Boolean)
      })
      const reportId = data.reportId || data.id || ''
      const query = reportId ? `?reportId=${encodeURIComponent(reportId)}` : ''

      wx.navigateTo({
        url: `/pages/profile/system-management/report-detail/index${query}`
      })
    } catch (error) {
      toast.info(error.message || '举报提交失败')
    }
  },

  buildRows(list) {
    const rows = []

    this.normalizeList(list).forEach((item, index) => {
      const rowIndex = Math.floor(index / 3)

      if (!rows[rowIndex]) {
        rows[rowIndex] = []
      }

      rows[rowIndex].push(item)
    })

    return rows
  },

  normalizeList(list) {
    return Array.isArray(list) ? list : []
  },

  normalizeTipLines(list) {
    return this.normalizeList(list).map((item) => {
      if (typeof item === 'string') {
        return {
          text: item
        }
      }

      return {
        text: item.text || item.label || item.content || ''
      }
    }).filter((item) => item.text)
  }
})
