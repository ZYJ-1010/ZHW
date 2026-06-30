const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const ASSET_BASE = '/pages/profile/system-management/feedback/assets'

Page({
  data: {
    icons: {
      attachment: `${ASSET_BASE}/icon-paperclip.svg`,
      send: `${ASSET_BASE}/icon-send-plane.svg`
    },
    feedbackImage: '',
    messages: [],
    feedbackId: ''
  },

  onLoad(options = {}) {
    const feedbackId = options.feedbackId || options.id || ''

    this.setData({
      feedbackId
    })

    if (feedbackId) {
      this.loadFeedbackDetail()
    }
  },

  async loadFeedbackDetail() {
    try {
      const data = await profileService.getSystemFeedbackDetail({
        feedbackId: this.data.feedbackId
      })
      const detail = data.detail || data.feedback || data

      this.setData({
        feedbackImage: detail.imageUrl || detail.feedbackImage || this.data.feedbackImage,
        messages: this.normalizeList(data.messages || detail.messages)
      })
    } catch (error) {
      this.setData({
        messages: []
      })
      toast.info(error.message || '反馈详情加载失败')
    }
  },

  normalizeList(list) {
    return Array.isArray(list) ? list : []
  },

  handleAttachTap() {
    toast.developing('补充截图待接入文件接口')
  },

  handleSendTap() {
    toast.developing('补充说明提交待接入反馈接口')
  }
})
