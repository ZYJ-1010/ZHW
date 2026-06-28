const ASSET_BASE = '/pages/profile/system-management/feedback/assets'

Page({
  data: {
    activeTab: 'all',
    tabs: [
      { key: 'all', label: '全部' },
      { key: 'processing', label: '处理中', count: 2 },
      { key: 'resolved', label: '已解决' }
    ],
    records: [],
    allRecords: [
      {
        id: 'feature-address-copy',
        type: '功能建议',
        typeTone: 'blue',
        status: '处理中',
        statusClass: 'processing',
        content: '建议在组局详情页增加「一键复制地址」功能，目前每次都要手动输入地址比较麻烦，希望能优化一下。',
        time: '2026-06-14 18:32',
        replyText: '客服已回复',
        replyClass: 'processing',
        images: []
      },
      {
        id: 'cashout-busy',
        type: '问题反馈',
        typeTone: 'red',
        status: '待处理',
        statusClass: 'pending',
        content: '积分提现时提示「系统繁忙」，已经持续两天了，请尽快修复。',
        time: '2026-06-13 09:15',
        images: [
          `${ASSET_BASE}/feedback-thumb-scenery.png`,
          `${ASSET_BASE}/feedback-thumb-building.png`
        ],
        extraImageCount: 1
      },
      {
        id: 'expert-home-speed',
        type: '体验优化',
        typeTone: 'gold',
        status: '已解决',
        statusClass: 'resolved',
        content: '行家主页加载速度有点慢，图片可以优化一下压缩策略。',
        time: '2026-06-10 14:22',
        replyText: '已解决',
        replyClass: 'resolved',
        images: []
      },
      {
        id: 'game-reminder',
        type: '组局相关',
        typeTone: 'gray',
        status: '已解决',
        statusClass: 'resolved',
        content: '组局开始前30分钟没有收到提醒，希望能增加推送通知功能。',
        time: '2026-06-08 11:05',
        images: []
      }
    ]
  },

  onLoad() {
    this.applyRecordFilter(this.data.activeTab)
  },

  handleTabTap(event) {
    const { key } = event.currentTarget.dataset

    if (key) {
      this.applyRecordFilter(key)
    }
  },

  applyRecordFilter(activeTab) {
    const records = this.data.allRecords.filter((record) => {
      if (activeTab === 'resolved') {
        return record.statusClass === 'resolved'
      }

      if (activeTab === 'processing') {
        return record.statusClass !== 'resolved'
      }

      return true
    })

    this.setData({
      activeTab,
      records
    })
  },

  handleRecordTap() {
    wx.navigateTo({
      url: '/pages/profile/system-management/feedback-detail/index'
    })
  }
})
