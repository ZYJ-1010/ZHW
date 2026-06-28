const toast = require('../../../../utils/toast')

const ASSET_BASE = '/pages/profile/system-management/feedback/assets'

Page({
  data: {
    activeType: 'feature',
    activeSession: 'werewolf',
    quickMenuOpen: false,
    showQuickFeedback: false,
    quickType: 'problem',
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
    feedbackTypes: [
      { key: 'feature', label: '功能建议' },
      { key: 'problem', label: '问题反馈' },
      { key: 'experience', label: '体验优化' },
      { key: 'game', label: '组局相关' },
      { key: 'expert', label: '行家相关' },
      { key: 'points', label: '积分/提现' },
      { key: 'other', label: '其他' }
    ],
    sessions: [
      { key: 'werewolf', title: '周末狼人杀局', meta: '06-14 14:00 · 8人' },
      { key: 'script', title: '剧本杀新手局', meta: '06-12 19:30 · 6人' },
      { key: 'boardgame', title: '桌游社交局', meta: '06-10 15:00 · 4人' }
    ],
    quickTypes: [
      { key: 'problem', label: '遇到问题', icon: `${ASSET_BASE}/icon-problem.svg`, tone: 'red' },
      { key: 'feature', label: '功能建议', icon: `${ASSET_BASE}/icon-pencil.svg`, tone: 'cyan' },
      { key: 'experience', label: '体验优化', icon: `${ASSET_BASE}/icon-star-outline.svg`, tone: 'gold' },
      { key: 'other', label: '其他', icon: `${ASSET_BASE}/icon-alert.svg`, tone: 'gray' }
    ],
    quickActions: [
      { key: 'feature', label: '功能建议', icon: `${ASSET_BASE}/icon-pencil.svg` },
      { key: 'problem', label: '问题反馈', icon: `${ASSET_BASE}/icon-alert.svg` },
      { key: 'screenshot', label: '截图反馈', icon: `${ASSET_BASE}/icon-image.svg` }
    ]
  },

  onLoad(options = {}) {
    if (options.sheet === 'quick') {
      this.setData({
        showQuickFeedback: true
      })
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

  handleSubmitTap() {
    wx.navigateTo({
      url: '/pages/profile/system-management/feedback-success/index'
    })
  },

  noop() {}
})
