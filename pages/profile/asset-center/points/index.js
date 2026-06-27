const toast = require('../../../../utils/toast')

const ASSET_BASE = '/pages/profile/asset-center/points/assets'

Page({
  data: {
    icons: {
      back: `${ASSET_BASE}/icon-chevron-left.svg`,
      more: `${ASSET_BASE}/icon-ellipsis-vertical.svg`
    },
    summary: {
      available: '2,580',
      stats: [
        { key: 'total', label: '累计积分', value: '5,000' },
        { key: 'redeemed', label: '已兑换', value: '2,420' },
        { key: 'expired', label: '过期积分', value: '0' }
      ]
    },
    rules: [
      {
        text: '每笔交易完成后，按各角色获得交易',
        strong: '分润金额 × 10%',
        suffix: '生成积分'
      },
      {
        text: '积分有效期',
        strong: '12个月',
        suffix: '，过期自动清零'
      },
      {
        text: '积分仅可兑换',
        strong: '平台限定商品',
        suffix: ''
      }
    ],
    earnExample: {
      title: '交易分润积分',
      subtitle: '唯一积分来源',
      points: '+48',
      rows: [
        { label: '行家分润金额', value: '¥480' },
        { label: '积分比例', value: '× 10%' }
      ],
      result: '= 积分 +48'
    },
    roleExamples: [
      { key: 'expert', role: '行家（分润60%）', amount: '¥480', points: '+48', iconText: '🧑' },
      { key: 'guide', role: '领路人（分润30%）', amount: '¥240', points: '+24', iconText: '👬' },
      { key: 'platform', role: '平台（分润10%）', amount: '¥80', points: '+8', iconText: '🌐' }
    ],
    filters: ['全部', '收入', '支出'],
    activeFilter: '全部',
    records: [
      {
        id: 'point-001',
        title: '行家交易分润积分',
        desc: '分润 ¥480 × 10%',
        time: '2026-03-20 14:30 · 订单 REF-20260320-001',
        points: '+48',
        tone: 'plus',
        iconText: '💰',
        iconTone: 'green'
      },
      {
        id: 'point-002',
        title: '领路人交易分润积分',
        desc: '分润 ¥240 × 10%',
        time: '2026-03-20 14:30 · 订单 REF-...',
        points: '+24',
        tone: 'plus',
        iconText: '👬',
        iconTone: 'green'
      },
      {
        id: 'point-003',
        title: '行家交易分润积分',
        desc: '分润 ¥600 × 10%',
        time: '2026-03-15 11:20',
        points: '+60',
        tone: 'plus',
        iconText: '💰',
        iconTone: 'green'
      },
      {
        id: 'point-004',
        title: '兑换平台限定商品',
        desc: '优先推荐权益包 x 1',
        time: '2026-03-10 16:45',
        points: '-500',
        tone: 'minus',
        iconText: '🎁',
        iconTone: 'pink'
      },
      {
        id: 'point-005',
        title: '领路人交易分润积分',
        desc: '分润 ¥150 × 10%',
        time: '2026-03-08 09:10',
        points: '+15',
        tone: 'plus',
        iconText: '👬',
        iconTone: 'green'
      },
      {
        id: 'point-006',
        title: '行家交易分润积分',
        desc: '分润 ¥320 × 10%',
        time: '2026-03-01 10:01 · 订单 REF-20260301-001',
        points: '+32',
        tone: 'plus',
        iconText: '💰',
        iconTone: 'green'
      }
    ]
  },

  handleFilterTap(event) {
    const filter = event.currentTarget.dataset.filter

    this.setData({
      activeFilter: filter
    })
  },

  handleDeveloping() {
    toast.developing()
  }
})
