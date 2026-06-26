Page({
  data: {
    profile: {
      level: 'V5 探险家',
      name: '小明',
      desc: '唯一ID: 00189 · 已加入 128 天'
    },
    metrics: [
      { value: '1.28', label: '总邀约数', trend: '▲', tone: 'up' },
      { value: '34', label: '成功转化', trend: '▲', tone: 'up' },
      { value: '26.6', label: '转化率', trend: '▼ 2.1%', tone: 'down' },
      { value: '¥12.58', label: '分润收益', trend: '▲', tone: 'up' }
    ],
    actions: [
      { icon: '🔗', label: '分享邀请码' },
      { icon: '▦', label: '二维码', iconClass: 'white' },
      { icon: '▧', label: '生成海报', iconClass: 'white' }
    ],
    tabs: ['数据概览', '关系网络', '邀约记录', '贡献排行', '收益明细'],
    trends: [
      { label: '本周新增邀约', value: '+86', tone: 'cyan' },
      { label: '本周新增转化', value: '+24', tone: 'green' },
      { label: '本周分润', value: '¥1,240', tone: 'cyan' }
    ]
  }
})
