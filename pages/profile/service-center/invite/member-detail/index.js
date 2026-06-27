Page({
  data: {
    member: {
      avatar: '👨‍💼',
      name: '张大山',
      level: '一级成员'
    },
    stats: [
      { value: '8', label: '总邀约' },
      { value: '3', label: '成功转化' },
      { value: '37.2%', label: '转化率' }
    ],
    income: [
      { label: '直接贡献收益', value: '¥2,400' },
      { label: '团队贡献收益', value: '¥1,860' },
      { label: '合计贡献', value: '¥4,260', highlight: true }
    ],
    activities: [
      { icon: '🎯', title: '桌游局 · 成功入局', time: '06-14 20:30', amount: '+' },
      { icon: '🎲', title: '剧本杀 · 成功入局', time: '06-13 19:00', amount: '+' }
    ]
  },

  onLoad(options) {
    this.setData({
      memberId: options.id || ''
    })
  }
})
