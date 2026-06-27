Page({
  data: {
    activePeriodIndex: 0,
    activeType: 'inviteCount',
    periods: [
      { key: 'week', label: '本周' },
      { key: 'month', label: '本月' },
      { key: 'quarter', label: '本季' },
      { key: 'year', label: '本年' },
      { key: 'all', label: '全部' }
    ],
    rankTypes: [
      { key: 'inviteCount', label: '邀约数排行' },
      { key: 'profitContribution', label: '分润贡献排行' }
    ],
    members: [
      { id: 'member-1', rank: 1, avatar: '👨‍💼', name: '张大山', level: '一级成员', activeDays: 12, inviteCount: 8, profitContribution: '¥4,260' },
      { id: 'member-2', rank: 2, avatar: '👩‍💻', name: '李小红', level: '一级成员', activeDays: 9, inviteCount: 6, profitContribution: '¥3,180' },
      { id: 'member-3', rank: 3, avatar: '👨‍🎨', name: '王建国', level: '一级成员', activeDays: 7, inviteCount: 4, profitContribution: '¥2,540' },
      { id: 'member-4', rank: 4, avatar: '👩‍🎤', name: '赵小美', level: '一级成员', activeDays: 5, inviteCount: 3, profitContribution: '¥1,860' },
      { id: 'member-5', rank: 5, avatar: '👨‍🔬', name: '陈博士', level: '一级成员', activeDays: 4, inviteCount: 3, profitContribution: '¥1,420' }
    ]
  },

  handlePeriodTap(event) {
    const { index } = event.currentTarget.dataset

    if (index === this.data.activePeriodIndex) {
      return
    }

    this.setData({
      activePeriodIndex: index
    })
  },

  handleTypeTap(event) {
    const { type } = event.currentTarget.dataset

    if (!type || type === this.data.activeType) {
      return
    }

    this.setData({
      activeType: type
    })
  }
})
