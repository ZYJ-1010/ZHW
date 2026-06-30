const toast = require('../../../../../utils/toast')
const profileService = require('../../../../../services/profile')

Page({
  data: {
    review: {
      user: '',
      avatar: '',
      time: '',
      rating: '',
      title: '',
      content: '',
      tags: [],
      orderNo: '',
      amount: ''
    },
    templates: [],
    templateRows: [],
    history: [],
    replyText: '',
    reviewId: '',
    submitting: false
  },

  onLoad(options = {}) {
    this.setData({
      reviewId: options.id || options.reviewId || '',
      templateRows: this.buildTemplateRows(this.data.templates)
    })
    this.loadReviewDetail()
  },

  async loadReviewDetail() {
    if (!this.data.reviewId) {
      return
    }

    try {
      const data = await profileService.getServiceReviewDetail({
        reviewId: this.data.reviewId
      })
      const templates = Array.isArray(data.templates) ? data.templates : []

      this.setData({
        review: data.review || data.detail || this.data.review,
        templates,
        templateRows: this.buildTemplateRows(templates),
        history: Array.isArray(data.history || data.messages) ? (data.history || data.messages) : []
      })
    } catch (error) {
      toast.info(error.message || '评价详情加载失败')
    }
  },

  buildTemplateRows(templates = []) {
    const rows = []

    templates.forEach((template, index) => {
      if (index % 2 === 0) {
        rows.push([template])
        return
      }

      rows[rows.length - 1].push(template)
    })

    return rows
  },

  onBackTap() {
    const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    wx.redirectTo({
      url: '/pages/profile/service-center/manage/review-manage/index'
    })
  },

  onReplyInput(event) {
    this.setData({
      replyText: event.detail.value
    })
  },

  onTemplateTap(event) {
    const text = event.currentTarget.dataset.text

    this.setData({
      replyText: text
    })
  },

  async onSubmitTap() {
    if (this.data.submitting) {
      return
    }

    const content = this.data.replyText.trim()

    if (!content) {
      toast.info('请输入回复内容')
      return
    }

    this.setData({
      submitting: true
    })

    try {
      await profileService.replyServiceReview({
        reviewId: this.data.reviewId,
        content
      })

      this.updateManagePageReply(this.data.reviewId, content)
      toast.info('回复已提交')
      wx.navigateBack()
    } catch (error) {
      toast.info(error.message || '回复评价提交失败')
    } finally {
      this.setData({
        submitting: false
      })
    }
  },

  updateManagePageReply(reviewId, content) {
    const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []
    const previousPage = pages.length > 1 ? pages[pages.length - 2] : null

    if (!previousPage || !previousPage.data || !Array.isArray(previousPage.data.reviews)) {
      return
    }

    const wasPending = previousPage.data.reviews.some((item) => item.id === reviewId && item.statusType === 'pending')
    const reviews = previousPage.data.reviews.map((item) => {
      if (item.id !== reviewId) {
        return item
      }

      return {
        ...item,
        reply: content,
        statusType: 'replied'
      }
    })

    previousPage.setData({
      reviews,
      pendingCount: wasPending ? Math.max(0, previousPage.data.pendingCount - 1) : previousPage.data.pendingCount
    })
  }
})
