const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')
const fileService = require('../../../../services/file')

const ASSET_BASE = '/pages/profile/system-management/feedback/assets'

Page({
  data: {
    icons: {
      attachment: `${ASSET_BASE}/icon-paperclip.svg`,
      send: `${ASSET_BASE}/icon-send-plane.svg`
    },
    recordId: '',
    record: {
      type: '',
      typeTone: 'blue',
      status: '',
      content: '',
      time: ''
    },
    replyContent: '',
    evidenceImages: [],
    sending: false,
    feedbackImage: `${ASSET_BASE}/feedback-detail-rail.png`,
    messages: [],
    loadError: '',
    feedbackLimits: {
      fileMaxCount: 0,
      uploadFullText: '',
      uploadSelectedTemplate: ''
    }
  },

  onLoad(options = {}) {
    const recordId = String(options.id || options.recordId || '').trim()

    if (recordId) {
      this.setData({ recordId })
      this.loadFeedbackConfig()
      this.loadFeedbackDetail(recordId)
      return
    }

    this.setData({ loadError: '缺少反馈记录' })
  },

  async loadFeedbackConfig() {
    try {
      const data = await profileService.getSystemFeedbackHome()
      const limits = data.limits || {}

      this.setData({
        feedbackLimits: {
          fileMaxCount: Math.max(0, Number(limits.fileMaxCount || 0)),
          uploadFullText: limits.uploadFullText || '',
          uploadSelectedTemplate: limits.uploadSelectedTemplate || ''
        }
      })
    } catch (error) {
      toast.info(error.message || '反馈配置暂时不可用')
    }
  },

  async loadFeedbackDetail(recordId) {
    try {
      const data = await profileService.getSystemFeedbackDetail(recordId)

      this.setData({
        record: data.record || this.data.record,
        messages: Array.isArray(data.messages) ? data.messages : [],
        loadError: ''
      })
    } catch (error) {
      this.setData({
        loadError: error.message || '获取反馈详情失败',
        messages: []
      })
      toast.info(error.message || '获取反馈详情失败')
    }
  },

  handleReplyInput(event) {
    this.setData({
      replyContent: event.detail.value || ''
    })
  },

  handleAttachTap() {
    const remain = this.remainingAttachmentCount()

    if (!remain) {
      toast.info(this.data.feedbackLimits.uploadFullText || '附件数量已达上限')
      return
    }

    const onSuccess = (result = {}) => {
      const files = Array.isArray(result.tempFiles) ? result.tempFiles : []
      const paths = files
        .map((item) => item.tempFilePath || item.path)
        .filter(Boolean)
        .slice(0, remain)
        .map((path) => ({ path }))

      if (!paths.length) {
        return
      }

      const evidenceImages = this.data.evidenceImages.concat(paths).slice(0, this.data.evidenceImages.length + remain)
      this.setData({ evidenceImages })
      toast.success(this.uploadSelectedText())
    }

    if (wx.chooseMedia) {
      wx.chooseMedia({
        count: remain,
        mediaType: ['image'],
        sourceType: ['album', 'camera'],
        success: onSuccess
      })
      return
    }

    wx.chooseImage({
      count: remain,
      sourceType: ['album', 'camera'],
      success: onSuccess
    })
  },

  remainingAttachmentCount() {
    const maxCount = Number(this.data.feedbackLimits.fileMaxCount || 0)

    if (!maxCount) {
      return 0
    }

    return Math.max(0, maxCount - this.data.evidenceImages.length)
  },

  uploadSelectedText() {
    const selectedCount = this.data.evidenceImages.length
    const maxCount = Number(this.data.feedbackLimits.fileMaxCount || 0)
    const template = this.data.feedbackLimits.uploadSelectedTemplate || ''

    if (template) {
      return template.replace('{selected}', selectedCount).replace('{max}', maxCount)
    }

    return `已选择 ${selectedCount}/${maxCount} 个附件`
  },

  async handleSendTap() {
    if (this.data.sending) {
      return
    }

    if (!this.data.recordId) {
      toast.info('缺少反馈记录')
      return
    }

    const content = String(this.data.replyContent || '').trim()

    if (!content) {
      toast.info('请输入补充说明')
      return
    }

    this.setData({ sending: true })

    try {
      const fileIds = await fileService.uploadEvidenceImages(
        this.data.evidenceImages.map((item) => item.path).filter(Boolean),
        {
          bizType: 'report_attachment',
          objectId: 0
        }
      )

      const data = await profileService.appendSystemFeedbackMessage(this.data.recordId, {
        content,
        fileIds
      })

      this.setData({
        record: data.record || this.data.record,
        messages: Array.isArray(data.messages) ? data.messages : this.data.messages,
        replyContent: '',
        evidenceImages: [],
        sending: false
      })
      toast.success('已补充说明')
    } catch (error) {
      this.setData({ sending: false })
      toast.info(error.message || '补充说明提交失败')
    }
  }
})
