const toast = require('../../../../../utils/toast')

Page({
  data: {
    score: {
      overall: '4.9',
      total: '48',
      goodRate: '98%',
      stars: '★★★★☆',
      breakdown: [
        { label: '5星', percent: '90%', style: 'width: 90%;' },
        { label: '4星', percent: '8%', style: 'width: 8%;' },
        { label: '3星', percent: '2%', style: 'width: 2%;' }
      ]
    },
    pendingCount: 3,
    stats: [
      { value: '12', label: '近30天新增评价' },
      { value: '94%', label: '回复率' },
      { value: '96%', label: '好评率' },
      { value: '2小时', label: '平均响应' }
    ],
    reviews: [
      {
        id: 'review-001',
        user: '李明',
        avatar: 'LM',
        avatarClass: 'pink',
        time: '03-25',
        rating: '★★★★★',
        title: '产品架构咨询',
        content: '张专家非常专业，帮我们把产品架构梳理得很清晰，解决了很多历史遗留问题。沟通顺畅，响应及时，强烈推荐！',
        tags: ['专业能力强', '交付及时', '沟通顺畅'],
        reply: '感谢李总的认可，期待下次合作！',
        statusType: 'replied'
      },
      {
        id: 'review-002',
        user: '王华',
        avatar: 'WH',
        avatarClass: 'green',
        time: '03-24',
        rating: '★★★★★',
        title: 'UI设计服务',
        content: '设计质量很高，完全符合我们的品牌调性。修改响应也很快，整体体验非常好！',
        tags: ['超出预期', '性价比高'],
        reply: '',
        statusType: 'pending'
      },
      {
        id: 'review-003',
        user: '赵四',
        avatar: 'ZS',
        avatarClass: 'indigo',
        time: '03-20',
        rating: '★★★☆☆',
        title: '技术架构咨询',
        content: '整体还可以，但是沟通效率有待提升，回复有时候比较慢。建议改善响应速度。',
        tags: [],
        reply: '感谢您的反馈，我会加强响应速度，提升服务体验！',
        statusType: 'neutral'
      },
      {
        id: 'review-004',
        user: '陈五',
        avatar: 'CW',
        avatarClass: 'red',
        time: '03-15',
        rating: '★☆☆☆☆',
        title: '产品需求梳理',
        content: '服务与描述不符，没有解决实际问题。希望能退款。',
        tags: [],
        reply: '',
        alert: '已申请平台介入处理',
        statusType: 'critical'
      }
    ]
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
