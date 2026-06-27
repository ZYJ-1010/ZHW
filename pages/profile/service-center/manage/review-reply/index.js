const toast = require('../../../../../utils/toast')
const profileService = require('../../../../../services/profile')

Page({
  data: {
    review: {
      user: '李明',
      avatar: 'LM',
      time: '03-25 14:30',
      rating: '★★★★★',
      title: '产品架构咨询',
      content: '张专家非常专业，帮我们把产品架构梳理得很清晰，解决了很多历史遗留问题。沟通顺畅，响应及时，强烈推荐！',
      tags: ['专业能力强', '交付及时', '沟通顺畅'],
      orderNo: 'ORD-20260325-001',
      amount: '¥800'
    },
    templates: [
      '感谢您的认可！',
      '期待下次合作',
      '有问题随时联系',
      '我们会继续努力',
      '感谢反馈，已改进'
    ],
    templateRows: [
      ['感谢您的认可！', '期待下次合作'],
      ['有问题随时联系', '我们会继续努力'],
      ['感谢反馈，已改进']
    ],
    history: [
      {
        role: '李明（玩家）',
        avatar: '👤',
        time: '03-25 14:30',
        content: '张专家非常专业，帮我们把产品架构梳理得很清晰，解决了很多历史遗留问题。沟通顺畅，响应及时，强烈推荐！',
        side: 'user'
      },
      {
        role: '张专家（行家）',
        avatar: '🧑',
        time: '03-25 16:00',
        content: '感谢李总的认可，期待下次合作！',
        side: 'expert'
      }
    ],
    replyText: '',
    reviewId: '',
    submitting: false
  },

  onLoad(options = {}) {
    this.setData({
      reviewId: options.id || options.reviewId || 'review-002',
      templateRows: this.buildTemplateRows(this.data.templates)
    })
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
