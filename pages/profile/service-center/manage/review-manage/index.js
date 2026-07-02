const toast = require('../../../../../utils/toast')
const profileService = require('../../../../../services/profile')
const { navigateShellRoute } = require('../../../../../utils/shell-nav')

Page({
  data: {
    score: {
      overall: '0.0',
      total: '0',
      goodRate: '0%',
      stars: '☆☆☆☆☆',
      breakdown: [
        { label: '5星', percent: '0%', style: 'width: 0%;' },
        { label: '4星', percent: '0%', style: 'width: 0%;' },
        { label: '3星', percent: '0%', style: 'width: 0%;' }
      ]
    },
    pendingCount: 0,
    stats: [
      { value: '0', label: '近30天新增评价' },
      { value: '0%', label: '回复率' },
      { value: '0%', label: '好评率' },
      { value: '-', label: '平均响应' }
    ],
    reviews: [],
    allReviews: [],
    loadError: false,
    activeStatus: 'all',
    statusTabs: [
      { key: 'all', label: '全部', count: 0 },
      { key: 'pending', label: '待回复', count: 0 },
      { key: 'replied', label: '已回复', count: 0 },
      { key: 'critical', label: '需处理', count: 0 }
    ]
  },

  onLoad() {
    this.loadReviews()
  },

  onShow() {
    if (this.data.loaded) {
      this.loadReviews()
    }
  },

  async loadReviews() {
    try {
      const data = await profileService.getServiceReviews()

      this.setData({
        score: data.score || this.data.score,
        pendingCount: Number(data.pendingCount) || 0,
        stats: Array.isArray(data.stats) ? data.stats : this.data.stats,
        allReviews: Array.isArray(data.reviews) ? data.reviews.map((item) => ({
          ...item,
          tags: Array.isArray(item.tags) ? item.tags : []
        })) : [],
        templates: Array.isArray(data.templates) ? data.templates : [],
        loaded: true,
        loadError: false
      }, () => this.applyStatusFilter())
    } catch (error) {
      this.setData({
        loaded: true,
        loadError: true,
        reviews: [],
        allReviews: [],
        pendingCount: 0
      }, () => this.applyStatusFilter())
      toast.info(error.message || '评价列表加载失败')
    }
  },

  isPendingReview(item = {}) {
    return item.statusType === 'pending' || item.statusType === 'critical' || item.statusType === 'neutral'
  },

  buildStatusTabs(reviews = []) {
    const pendingCount = reviews.filter((item) => this.isPendingReview(item)).length
    const repliedCount = reviews.filter((item) => item.statusType === 'replied').length
    const criticalCount = reviews.filter((item) => item.statusType === 'critical').length

    return [
      { key: 'all', label: '全部', count: reviews.length },
      { key: 'pending', label: '待回复', count: pendingCount },
      { key: 'replied', label: '已回复', count: repliedCount },
      { key: 'critical', label: '需处理', count: criticalCount }
    ]
  },

  applyStatusFilter() {
    const source = this.data.allReviews
    const activeStatus = this.data.activeStatus
    const reviews = activeStatus === 'all'
      ? source
      : source.filter((item) => {
        if (activeStatus === 'pending') {
          return this.isPendingReview(item)
        }

        return item.statusType === activeStatus
      })

    this.setData({
      reviews,
      statusTabs: this.buildStatusTabs(source)
    })
  },

  onBackTap() {
    const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    navigateShellRoute('/pages/profile/index')
  },

  onReplyTap(event) {
    const id = event.currentTarget.dataset.id

    navigateShellRoute(`/pages/profile/service-center/manage/review-reply/index?id=${id || ''}`)
  },

  onStatusTap(event) {
    this.setData({
      activeStatus: event.currentTarget.dataset.key || 'all'
    }, () => this.applyStatusFilter())
  },

  onDetailTap(event) {
    const route = event.currentTarget.dataset.route
    const gameId = event.currentTarget.dataset.gameId

    navigateShellRoute(route || `/pages/game/detail/index?id=${gameId || ''}`)
  },

  onInterventionTap(event) {
    const route = event.currentTarget.dataset.route

    if (!route) {
      toast.info('缺少评价关联信息，无法发起平台介入')
      return
    }

    navigateShellRoute(route)
  },

  async onLikeTap(event) {
    const id = event.currentTarget.dataset.id

    if (!id) {
      toast.info('缺少评价信息，无法点赞')
      return
    }

    try {
      const data = await profileService.likeServiceReview(id)
      this.markReviewLiked(id, !!data.liked)
      toast.info(data.liked ? '已点赞' : '已取消点赞')
    } catch (error) {
      toast.info(error.message || '评价点赞失败')
    }
  },

  async onMenuTap(event) {
    const reviewId = event && event.currentTarget && event.currentTarget.dataset
      ? event.currentTarget.dataset.id
      : ''
    const target = (reviewId && this.data.reviews.find((item) => item.id === reviewId))
      || this.data.reviews.find((item) => this.isPendingReview(item))
      || this.data.reviews[0]

    if (!target) {
      toast.info('暂无可操作评价')
      return
    }

    try {
      const data = await profileService.getServiceReviewActions(target.id)
      const actions = Array.isArray(data.actions) ? data.actions : []

      if (!actions.length) {
        toast.info('暂无可用操作')
        return
      }

      wx.showActionSheet({
        itemList: actions.map((item) => item.text || item.key),
        success: (result) => {
          const action = actions[result.tapIndex]
          if (action && action.route) {
            navigateShellRoute(action.route)
          }
        }
      })
    } catch (error) {
      toast.info(error.message || '获取评价操作失败')
    }
  },

  markReviewLiked(reviewId, liked) {
    const mark = (items = []) => items.map((item) => (
      item.id === reviewId ? { ...item, liked } : item
    ))

    this.setData({
      reviews: mark(this.data.reviews),
      allReviews: mark(this.data.allReviews)
    })
  },

  onPendingTap() {
    const pendingReview = this.data.reviews.find((item) => (
      item.statusType === 'pending' ||
      item.statusType === 'critical' ||
      item.statusType === 'neutral'
    ))

    if (!pendingReview) {
      toast.info('暂无待回复评价')
      return
    }

    navigateShellRoute(`/pages/profile/service-center/manage/review-reply/index?id=${pendingReview.id}`)
  }
})
