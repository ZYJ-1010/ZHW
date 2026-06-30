const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const ASSET_BASE = '/pages/profile/system-management/feedback/assets'

Page({
  data: {
    activeType: '',
    activeSession: '',
    quickMenuOpen: false,
    showQuickFeedback: false,
    quickType: '',
    feedbackContent: '',
    contact: '',
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
    quickActions: []
  },

  onLoad(options = {}) {
    this.loadFeedbackOptions()
    this.loadFeedbackGames()

    if (options.sheet === 'quick') {
      this.setData({
        showQuickFeedback: true
      })
    }
  },

  async loadFeedbackOptions() {
    try {
      const data = await profileService.getSystemFeedbackOptions()
      const feedbackTypes = this.normalizeList(data.feedbackTypes || data.types)
        .map((item) => ({
          ...item,
          key: item.key || item.type || item.id
        }))
      const quickTypes = this.normalizeList(data.quickTypes)
        .map((item) => this.withIcon({
          ...item,
          key: item.key || item.type || item.id
        }))
      const quickActions = this.normalizeList(data.quickActions)
        .map((item) => this.withIcon({
          ...item,
          key: item.key || item.type || item.id
        }))

      this.setData({
        feedbackTypes,
        quickTypes,
        quickActions,
        activeType: this.data.activeType || feedbackTypes[0] && feedbackTypes[0].key || '',
        quickType: this.data.quickType || quickTypes[0] && quickTypes[0].key || ''
      })
    } catch (error) {
      this.setData({
        feedbackTypes: [],
        quickTypes: [],
        quickActions: [],
        activeType: '',
        quickType: ''
      })
      toast.info(error.message || '反馈配置加载失败')
    }
  },

  async loadFeedbackGames() {
    try {
      const data = await profileService.getSystemFeedbackGames()
      const sessions = this.normalizeList(data.sessions || data.games || data.list || data.items)
        .map((item) => ({
          ...item,
          key: item.key || item.id || item.gameId
        }))

      this.setData({
        sessions,
        activeSession: sessions[0] && (sessions[0].key || sessions[0].id || sessions[0].gameId) || ''
      })
    } catch (error) {
      this.setData({
        sessions: [],
        activeSession: ''
      })
      toast.info(error.message || '可反馈组局加载失败')
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
    wx.navigateTo({
      url: '/pages/profile/system-management/feedback-records/index'
    })
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
    toast.developing('截图上传待接入文件接口')
  },

  handleVoiceTap() {
    toast.developing('语音反馈待接入录音能力')
  },

  async handleSubmitTap() {
    try {
      await profileService.submitSystemFeedback({
        type: this.data.showQuickFeedback ? this.data.quickType : this.data.activeType,
        gameId: this.data.activeSession,
        content: this.data.feedbackContent,
        contact: this.data.contact,
        source: this.data.showQuickFeedback ? 'quick' : 'form'
      })

      wx.navigateTo({
        url: '/pages/profile/system-management/feedback-success/index'
      })
    } catch (error) {
      toast.info(error.message || '反馈提交失败')
    }
  },

  withIcon(item = {}) {
    if (item.icon) {
      return item
    }

    const iconMap = {
      problem: this.data.icons.problem,
      feature: this.data.icons.pencil,
      experience: this.data.icons.star,
      screenshot: this.data.icons.image,
      other: this.data.icons.alert
    }

    return {
      ...item,
      icon: iconMap[item.key] || item.iconSrc || this.data.icons.alert
    }
  },

  normalizeList(list) {
    return Array.isArray(list) ? list : []
  },

  noop() {}
})
