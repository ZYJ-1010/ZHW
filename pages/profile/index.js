const toast = require('../../utils/toast')

Page({
  data: {
    user: {
      nickname: '小明',
      memberLevel: '基础会员',
      roleLevel: 'V5 探险家',
      role: '玩家',
      avatarText: '小'
    },
    stats: [
      { label: '引荐数', value: '128' },
      { label: '成功数', value: '86' },
      { label: '成交总额', value: '¥45K' },
      { label: '信用度', value: '98' }
    ],
    assets: [
      { label: '总成交额', value: '¥12,580', tone: '' },
      { label: '可提现', value: '¥3,200', tone: 'green' },
      { label: '待结算', value: '¥800', tone: 'orange' }
    ],
    serviceSections: [
      {
        title: '服务中心',
        items: [
          { title: '我的局', iconSrc: '/pages/profile/assets/i66@3x.png', iconClass: 'purple-blue', badge: '2进行中', badgeClass: 'pink', route: '/pages/profile/service-center/my-games/index' },
          { title: '我的邀请', iconSrc: '/pages/profile/assets/i68@3x.png', iconClass: 'purple-blue', badge: '3待确认', badgeClass: 'orange', route: '/pages/profile/service-center/invite/overview/index' },
          { title: '评价管理', iconSrc: '/pages/profile/assets/i69@3x.png', iconClass: 'purple-blue', badge: '2待评价', badgeClass: 'pink', route: '/pages/profile/service-center/manage/review-manage/index' }
        ]
      },
      {
        title: '资产中心',
        items: [
          { title: '我的资产', iconSrc: '/pages/profile/assets/i70@3x.png', iconClass: 'orange', route: '/pages/profile/asset-center/manage/index' },
          { title: '我的押金', iconSrc: '/pages/profile/assets/i71@3x.png', iconClass: 'orange' },
          { title: '积分商城', iconSrc: '/pages/profile/assets/i72@3x.png', iconClass: 'orange', route: '/pages/profile/asset-center/mall/index' },
          { title: '我的积分', iconSrc: '/pages/profile/assets/i73@3x.png', iconClass: 'orange', route: '/pages/profile/asset-center/points/index' },
          { title: '开票中心', iconSrc: '/pages/profile/assets/i74@3x.png', iconClass: 'orange' }
        ]
      },
      {
        title: '足迹中心',
        items: [
          { title: '我的足迹', iconSrc: '/pages/profile/assets/i75@3x.png', iconClass: 'teal' },
          { title: '我的城市故事', iconSrc: '/pages/profile/assets/i76@3x.png', iconClass: 'teal' },
          { title: '我的成就墙', iconSrc: '/pages/profile/assets/i77@3x.png', iconClass: 'teal', route: '/pages/profile/footprint/achievements/index' }
        ]
      },
      {
        title: '系统管理',
        items: [
          { title: '我的资料', iconSrc: '/pages/profile/assets/i78@3x.png', iconClass: 'blue-purple', route: '/pages/profile/system-management/profile-info/index' },
          { title: '技能配置', iconSrc: '/pages/profile/assets/i79@3x.png', iconClass: 'blue-purple', route: '/pages/profile/system-management/skill-config/index' },
          { title: '屏蔽设置', iconSrc: '/pages/profile/assets/i80@3x.png', iconClass: 'blue-purple', route: '/pages/profile/system-management/block-settings/index' },
          { title: '信用中心', iconSrc: '/pages/profile/assets/i81@3x.png', iconClass: 'blue-purple', route: '/pages/profile/credit-center/index' },
          { title: '举报中心', iconSrc: '/pages/profile/assets/i82@3x.png', iconClass: 'blue-purple', route: '/pages/profile/system-management/report-center/index' },
          { title: '签署协议', iconSrc: '/pages/profile/assets/i83@3x.png', iconClass: 'blue-purple', route: '/pages/profile/system-management/agreement-sign/index' },
          { title: '建议反馈', iconSrc: '/pages/profile/assets/i84@3x.png', iconClass: 'blue-purple', route: '/pages/profile/system-management/feedback/index' },
          { title: '系统设置', iconSrc: '/pages/profile/assets/i85@3x.png', iconClass: 'blue-purple', route: '/pages/profile/settings/index' }
        ]
      }
    ]
  },

  handleMenuTap(event) {
    const { route } = event.currentTarget.dataset

    if (!route) {
      toast.developing()
      return
    }

    wx.navigateTo({ url: route })
  },

  handleAssetAllTap() {
    wx.navigateTo({
      url: '/pages/profile/asset-center/manage/index'
    })
  },

  handleUpgradeTap() {
    wx.navigateTo({
      url: '/pages/profile/member/index'
    })
  }
})
