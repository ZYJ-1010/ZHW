Page({
  data: {
    topTabs: [
      { label: '我的局', active: true },
      { label: '我参与的', active: false },
      { label: '我受邀的', active: false },
      { label: '我收藏的', active: false }
    ],
    filters: [
      { label: '全部', active: true },
      { label: '进行中(2)', active: false },
      { label: '已完成(5)', active: false },
      { label: '超时(0)', active: false },
      { label: '已取消(1)', active: false }
    ],
    games: [
      {
        title: '产品架构咨询',
        status: '进行中',
        statusClass: 'green',
        time: '3天前',
        expert: '行家：张专家',
        guide: '领路人：王引荐',
        escrow: '已托管',
        delivery: '预计交付：03-25 14:00',
        steps: [
          { title: '需求确认', desc: '你已确认服务需求', done: true },
          { title: '组局成功', desc: '三方连接已建立', done: true },
          { title: '服务进行中', desc: '业务主正在交付服务', done: false }
        ],
        actions: ['联系业务主', '查看群聊'],
        tail: '等待完成确认'
      },
      {
        title: 'UI设计服务',
        status: '已完成',
        statusClass: 'gray',
        time: '5天前',
        expert: '行家：陈设计师',
        guide: '',
        escrow: '已完成',
        delivery: '',
        steps: [],
        actions: ['评价', '再次购买'],
        tail: ''
      },
      {
        title: '技术咨询服务',
        status: '已取消',
        statusClass: 'orange',
        time: '10天前',
        expert: '行家：刘工',
        guide: '取消原因：时间冲突',
        escrow: '已退款',
        delivery: '',
        steps: [],
        actions: [],
        tail: ''
      }
    ]
  }
})
