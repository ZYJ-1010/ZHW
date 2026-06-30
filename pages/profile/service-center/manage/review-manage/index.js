const toast = require('../../../../../utils/toast')
const profileService = require('../../../../../services/profile')

Page({
  data: {
    score: {
      overall: '',
      total: '',
      goodRate: '',
      stars: '',
      breakdown: []
    },
    pendingCount: 0,
    stats: [],
    reviews: []
  },

  onLoad() {
    this.loadReviews()
  },

  async loadReviews() {
    try {
      const data = await profileService.getServiceReviews()

      this.setData({
        score: data.score || this.data.score,
        pendingCount: Number(data.pendingCount || 0),
        stats: Array.isArray(data.stats) ? data.stats : [],
        reviews: Array.isArray(data.reviews || data.list || data.items) ? (data.reviews || data.list || data.items) : []
      })
    } catch (error) {
      toast.info(error.message || '评价列表加载失败')
    }
  },

  onBackTap() {
    const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    wx.redirectTo({
      url: '/pages/profile/index'
    })
  },

  onReplyTap(event) {
    const id = event.currentTarget.dataset.id

    wx.navigateTo({
      url: `/pages/profile/service-center/manage/review-reply/index?id=${id || ''}`
    })
  },

  onLikeTap() {
    toast.info('点赞功能待接入')
  },

  onMenuTap() {
    toast.info('评价管理菜单待接入')
  },

  onPendingTap() {
    const pendingReview = this.data.reviews.find((item) => item.statusType === 'pending')

    if (!pendingReview) {
      toast.info('暂无待回复评价')
      return
    }

    wx.navigateTo({
      url: `/pages/profile/service-center/manage/review-reply/index?id=${pendingReview.id}`
    })
  }
})
