Page({
  data: {
    activeRole: 'referred',
    activeStatus: 'all',
    roleTabs: [
      { key: 'referred', label: '我引荐的' },
      { key: 'created', label: '我发起的' }
    ],
    filters: [
      { key: 'all', label: '全部' },
      { key: 'progress', label: '进行中(3)' },
      { key: 'completed', label: '已完成(12)' },
      { key: 'timeout', label: '超时(1)' },
      { key: 'cancelled', label: '已取消(2)' }
    ],
    records: [],
    allRecords: [
      {
        role: 'referred',
        statusKey: 'progress',
        status: '进行中',
        statusClass: 'blue',
        id: 'REF-20260320-001',
        time: '3天前',
        expertAvatar: 'ZH',
        expert: '张专家',
        playerAvatar: 'LI',
        player: '李明',
        title: '产品架构咨询',
        budget: '预算：¥800 | 你的奖励：¥80',
        steps: [
          { label: '组局成功', time: '03-20 14:30' },
          { label: '服务进行中', time: '预计交付：03-25' },
          { label: '等待完成确认', time: '' }
        ],
        actions: ['提醒交付', '查看详情']
      },
      {
        role: 'referred',
        statusKey: 'completed',
        status: '已完成',
        statusClass: 'green',
        id: 'REF-20260315-002',
        time: '5天前',
        expertAvatar: 'CH',
        expert: '陈工',
        playerAvatar: 'WA',
        player: '王总',
        title: 'UI设计服务',
        budget: '预算：¥600 | 你的奖励：¥60',
        income: '已到账',
        steps: [],
        actions: []
      },
      {
        role: 'referred',
        statusKey: 'timeout',
        status: '超时',
        statusClass: 'red',
        id: 'REF-20260301-003',
        time: '已超时15天',
        compact: true,
        expertAvatar: 'LI',
        expert: '刘设计师 ↔ 赵客户',
        title: '技术咨询服务',
        warning: '系统已自动发送超时预警，建议联系双方确认状态',
        steps: [],
        actions: []
      },
      {
        role: 'created',
        statusKey: 'progress',
        status: '进行中',
        statusClass: 'blue',
        id: 'INV-20260322-006',
        time: '1天前',
        expertAvatar: 'ME',
        expert: '我',
        playerAvatar: 'ZH',
        player: '赵客户',
        title: '品牌增长咨询',
        budget: '预算：¥1,200 | 预计奖励：¥120',
        steps: [
          { label: '邀约已发出', time: '03-22 10:00' },
          { label: '等待对方确认', time: '' }
        ],
        actions: ['查看详情']
      },
      {
        role: 'created',
        statusKey: 'cancelled',
        status: '已取消',
        statusClass: 'red',
        id: 'INV-20260311-004',
        time: '12天前',
        compact: true,
        expertAvatar: 'ME',
        expert: '我 ↔ 王产品',
        title: '商业计划书梳理',
        warning: '该邀约已取消，可重新发起邀约',
        steps: [],
        actions: []
      }
    ]
  },

  onLoad() {
    this.applyFilters()
  },

  handleRoleTap(event) {
    const { key } = event.currentTarget.dataset

    this.setData({
      activeRole: key
    }, () => this.applyFilters())
  },

  handleFilterTap(event) {
    const { key } = event.currentTarget.dataset

    this.setData({
      activeStatus: key
    }, () => this.applyFilters())
  },

  handleWarningTap() {
    this.setData({
      activeRole: 'referred',
      activeStatus: 'timeout'
    }, () => this.applyFilters())
  },

  handleRecordTap(event) {
    const { id } = event.currentTarget.dataset

    wx.showToast({
      title: id ? '查看组局' : '暂无组局',
      icon: 'none'
    })
  },

  applyFilters() {
    const { activeRole, activeStatus, allRecords } = this.data
    const records = allRecords.filter((item) => {
      const roleMatched = item.role === activeRole
      const statusMatched = activeStatus === 'all' || item.statusKey === activeStatus

      return roleMatched && statusMatched
    })

    this.setData({ records })
  }
})
