const toast = require('../../../../../utils/toast')
const profileService = require('../../../../../services/profile')
const { navigateShellRoute } = require('../../../../../utils/shell-nav')

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
    submitting: false,
    hasReview: false,
    loadError: false
  },

  onLoad(options = {}) {
    const reviewId = options.id || options.reviewId || ''

    this.setData({
      reviewId,
      templateRows: this.buildTemplateRows(this.data.templates)
    })

    if (!reviewId) {
      this.setData({
        loadError: true
      })
      toast.info('缺少评价信息')
      return
    }

    this.loadReviewDetail(reviewId)
  },

  async loadReviewDetail(reviewId) {
    try {
      const data = await profileService.getServiceReviewDetail(reviewId)
      const templates = Array.isArray(data.templates) ? data.templates : this.data.templates
      const review = {
        ...this.data.review,
        ...(data.review || {}),
        tags: Array.isArray(data.review && data.review.tags) ? data.review.tags : []
      }

      this.setData({
        review,
        templates,
        templateRows: Array.isArray(data.templateRows) ? data.templateRows : this.buildTemplateRows(templates),
        history: Array.isArray(data.history) ? data.history : this.data.history,
        replyText: review.reply || '',
        hasReview: Boolean(data.review),
        loadError: false
      })
    } catch (error) {
      this.setData({
        hasReview: false,
        loadError: true,
        history: [],
        templates: [],
        templateRows: []
      })
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

    navigateShellRoute('/pages/profile/service-center/manage/review-manage/index')
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

    if (!this.data.hasReview) {
      toast.info('评价详情未加载，暂不能回复')
      return
    }

    if (!content) {
      toast.info('请输入回复内容')
      return
    }

    this.setData({
      submitting: true
    })

    try {
      const result = await profileService.replyServiceReview({
        reviewId: this.data.reviewId,
        content
      })

      this.updateManagePageReply(
        this.data.reviewId,
        result && result.reply && result.reply.content || content,
        result && result.statusType || 'replied'
      )
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

  updateManagePageReply(reviewId, content, statusType) {
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
        statusType: statusType || 'replied'
      }
    })

    previousPage.setData({
      reviews,
      pendingCount: wasPending ? Math.max(0, previousPage.data.pendingCount - 1) : previousPage.data.pendingCount
    })
  }
})
