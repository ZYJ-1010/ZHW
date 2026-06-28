Page({
  data: {
    activeFilter: 'all',
    filters: [
      { key: 'all', label: '全部' },
      { key: 'pending', label: '待处理' },
      { key: 'processing', label: '处理中' },
      { key: 'resolved', label: '已处理' }
    ],
    appeals: [
      {
        id: 'appeal-1',
        title: '被举报「诱导私下交易」',
        filter: 'pending',
        status: '待处理',
        statusClass: 'warning',
        reason: '交易为平台正常订单，并非私下交易，可提供订单截图作为证明',
        time: '2024-06-10 14:30'
      },
      {
        id: 'appeal-2',
        title: '被举报「言语骚扰」',
        filter: 'processing',
        status: '处理中',
        statusClass: 'success',
        reason: '沟通为正常服务交流，不存在骚扰行为，聊天记录可查证',
        time: '2024-06-08 09:15'
      },
      {
        id: 'appeal-3',
        title: '被举报「虚假信息」',
        filter: 'resolved',
        status: '已处理',
        statusClass: 'success',
        reason: '个人资料真实有效，可提供身份证明及学历证明',
        time: '2024-06-05 16:45'
      },
      {
        id: 'appeal-4',
        title: '被举报「恶意取消」',
        filter: 'resolved',
        status: '申诉驳回',
        statusClass: 'danger',
        reason: '因不可抗力因素取消，非主观恶意',
        time: '2024-05-28 11:20'
      }
    ],
    visibleAppeals: []
  },

  onLoad() {
    this.applyFilter(this.data.activeFilter)
  },

  handleTabTap(event) {
    const { target } = event.currentTarget.dataset
    const routeMap = {
      home: '/pages/profile/system-management/report-center/index',
      records: '/pages/profile/system-management/report-records/index'
    }

    if (routeMap[target]) {
      wx.redirectTo({ url: routeMap[target] })
    }
  },

  handleFilterTap(event) {
    const { key } = event.currentTarget.dataset

    if (key) {
      this.applyFilter(key)
    }
  },

  applyFilter(key) {
    const visibleAppeals = key === 'all'
      ? this.data.appeals
      : this.data.appeals.filter((item) => item.filter === key)

    this.setData({
      activeFilter: key,
      visibleAppeals
    })
  },

  handleDetailTap(event) {
    const { id } = event.currentTarget.dataset
    const query = id ? `?appealId=${id}` : ''

    wx.redirectTo({
      url: `/pages/profile/system-management/appeal-detail/index${query}`
    })
  }
})
