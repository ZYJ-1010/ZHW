const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')
const fileService = require('../../../../services/file')
const { navigateShellRoute } = require('../../../../utils/shell-nav')

const ASSET_BASE = '/pages/profile/system-management/feedback/assets'

Page({
  data: {
    activeType: 'feature',
    activeSession: 'general',
    quickMenuOpen: false,
    showQuickFeedback: false,
    quickType: 'problem',
    feedbackContent: '',
    contact: '',
    evidenceImages: [],
    voiceFiles: [],
    voiceFileIds: [],
    icons: {
      problem: `${ASSET_BASE}/icon-problem.svg`,
      pencil: `${ASSET_BASE}/icon-pencil.svg`,
      alert: `${ASSET_BASE}/icon-alert.svg`,
      image: `${ASSET_BASE}/icon-image.svg`,
      close: `${ASSET_BASE}/icon-close.svg`,
      closeWhite: `${ASSET_BASE}/icon-close-white.svg`,
      voice: `${ASSET_BASE}/icon-voice.png`,
      plus: `${ASSET_BASE}/icon-plus.svg`,
      notify: `${ASSET_BASE}/icon-notify.svg`,
      star: `${ASSET_BASE}/icon-star-outline.svg`
    },
    feedbackTypes: [],
    sessions: [],
    quickTypes: [],
    quickActions: [],
    feedbackLimits: {
      contentMaxLength: 500,
      fileMaxCount: 0,
      uploadNote: '',
      uploadFullText: '',
      uploadSelectedTemplate: ''
    }
  },

  onLoad(options = {}) {
    if (options.sheet === 'quick') {
      this.setData({
        showQuickFeedback: true
      })
    }
    this.loadFeedbackHome()
  },

  async loadFeedbackHome() {
    try {
      const data = await profileService.getSystemFeedbackHome()
      const feedbackTypes = this.withFeedbackIcons(data.feedbackTypes)
      const quickTypes = this.withFeedbackIcons(data.quickTypes)
      const quickActions = this.withFeedbackIcons(data.quickActions)
      this.setData({
        activeType: data.activeType || (feedbackTypes[0] && feedbackTypes[0].key) || '',
        activeSession: data.activeSession || this.data.activeSession,
        feedbackTypes,
        sessions: Array.isArray(data.sessions) && data.sessions.length ? data.sessions : this.data.sessions,
        quickType: (quickTypes[0] && quickTypes[0].key) || '',
        quickTypes,
        quickActions,
        feedbackLimits: this.normalizeFeedbackLimits(data.limits)
      })
    } catch (error) {
      toast.info(error.message || '反馈配置暂时不可用')
    }
  },

  normalizeFeedbackLimits(limits = {}) {
    const contentMaxLength = Number(limits.contentMaxLength || 500)
    const fileMaxCount = Math.max(0, Number(limits.fileMaxCount || 0))

    return {
      contentMaxLength,
      fileMaxCount,
      uploadNote: limits.uploadNote || '',
      uploadFullText: limits.uploadFullText || '',
      uploadSelectedTemplate: limits.uploadSelectedTemplate || ''
    }
  },

  withFeedbackIcons(items = []) {
    if (!Array.isArray(items)) {
      return []
    }

    return items
      .filter((item) => item && item.key && item.label)
      .map((item) => ({
        ...item,
        icon: item.icon || this.data.icons[item.iconKey] || this.data.icons.alert
      }))
  },

  handleTypeTap(event) {
    const { key } = event.currentTarget.dataset

    if (key) {
      this.setData({
        activeType: key
      })
    }
  },

  handleSessionTap(event) {
    const { key } = event.currentTarget.dataset

    if (key) {
      this.setData({
        activeSession: key
      })
    }
  },

  handleContentInput(event) {
    this.setData({
      feedbackContent: event.detail.value
    })
  },

  handleContactInput(event) {
    this.setData({
      contact: event.detail.value
    })
  },

  handleRecordTap() {
    navigateShellRoute('/pages/profile/system-management/feedback-records/index')
  },

  handleQuickTap() {
    this.setData({
      quickMenuOpen: true
    })
  },

  handleCloseQuickMenu() {
    this.setData({
      quickMenuOpen: false
    })
  },

  handleQuickActionTap(event) {
    const { key } = event.currentTarget.dataset

    this.setData({
      quickMenuOpen: false,
      showQuickFeedback: true,
      quickType: key === 'screenshot' ? 'problem' : key
    })
  },

  handleQuickTypeTap(event) {
    const { key } = event.currentTarget.dataset

    if (key) {
      this.setData({
        quickType: key
      })
    }
  },

  handleCloseQuick() {
    this.setData({
      showQuickFeedback: false
    })
  },

  handleUploadTap() {
    const restCount = this.remainingAttachmentCount()

    if (!restCount) {
      toast.info(this.data.feedbackLimits.uploadFullText || '附件数量已达上限')
      return
    }

    const appendImages = (paths = []) => {
      const nextImages = this.data.evidenceImages.concat(paths.map((path, index) => ({
        id: `${Date.now()}-${index}`,
        path
      }))).slice(0, this.data.evidenceImages.length + restCount)

      this.setData({
        evidenceImages: nextImages
      })
      toast.info(this.uploadSelectedText())
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

  handleVoiceTap() {
    if (!this.remainingAttachmentCount()) {
      toast.info(this.data.feedbackLimits.uploadFullText || '附件数量已达上限')
      return
    }

    if (typeof wx === 'undefined' || typeof wx.chooseMessageFile !== 'function') {
      toast.info('当前微信版本请通过截图或文字补充反馈')
      return
    }

    wx.chooseMessageFile({
      count: 1,
      type: 'file',
      success: async (res) => {
        const file = Array.isArray(res.tempFiles) ? res.tempFiles[0] : null

        if (!file || !file.path) {
          return
        }

        try {
          const fileId = await fileService.uploadSingleFile({
            path: file.path,
            fileName: file.name,
            size: file.size
          }, {
            bizType: 'report_attachment',
            objectId: 0
          })
          this.setData({
            voiceFiles: this.data.voiceFiles.concat({
              name: file.name || '语音反馈',
              fileId
            }),
            voiceFileIds: this.data.voiceFileIds.concat(fileId).filter(Boolean)
          })
          toast.success(this.uploadSelectedText())
        } catch (error) {
          toast.info(error.message || '语音上传失败')
        }
      }
    })
  },

  remainingAttachmentCount() {
    const maxCount = Number(this.data.feedbackLimits.fileMaxCount || 0)

    if (!maxCount) {
      return 0
    }

    return Math.max(0, maxCount - this.data.evidenceImages.length - this.data.voiceFileIds.length)
  },

  uploadSelectedText() {
    const selectedCount = this.data.evidenceImages.length + this.data.voiceFileIds.length
    const maxCount = Number(this.data.feedbackLimits.fileMaxCount || 0)
    const template = this.data.feedbackLimits.uploadSelectedTemplate || ''

    if (template) {
      return template.replace('{selected}', selectedCount).replace('{max}', maxCount)
    }

    return `已选择 ${selectedCount}/${maxCount} 个附件`
  },

  async handleSubmitTap() {
    try {
      const fileIds = await fileService.uploadEvidenceImages(
        this.data.evidenceImages.map((item) => item.path).filter(Boolean),
        { bizType: 'report_attachment', objectId: 0 }
      )
      const allFileIds = fileIds.concat(this.data.voiceFileIds).filter(Boolean)
      const result = await profileService.submitSystemFeedback({
        typeKey: this.data.activeType,
        sessionKey: this.data.activeSession,
        content: this.data.feedbackContent,
        contact: this.data.contact,
        fileIds: allFileIds,
        quick: false
      })
      const recordId = result && result.record ? result.record.id : ''
      toast.success(result.successTitle || '反馈已提交')
      if (typeof wx !== 'undefined' && wx.setStorageSync && result && result.successPage) {
        wx.setStorageSync('enjoy_feedback_success_page', result.successPage)
      }
      navigateShellRoute(`/pages/profile/system-management/feedback-success/index?id=${encodeURIComponent(recordId)}`)
    } catch (error) {
      toast.info(error.message || '反馈提交失败')
    }
  },

  noop() {}
})
