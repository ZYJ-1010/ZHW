Page({
  data: {
    roleTabs: [
      { label: '我引荐的', active: true },
      { label: '我发起的', active: false }
    ],
    filters: [
      { label: '全部', active: true },
      { label: '进行中(3)', active: false },
      { label: '已完成(12)', active: false },
      { label: '超时(1)', active: false },
      { label: '已取消(2)', active: false }
    ],
    records: [
      {
        status: '进行中',
        statusClass: 'green',
        time: '3天前',
        expert: '张专家',
        player: '李明',
        title: '产品架构咨询',
        budget: '预算：¥800 | 你的奖励：¥80',
        steps: ['组局成功', '服务进行中', '预计交付：03-25'],
        tail: '等待完成确认',
        actions: ['提醒交付', '查看详情']
      },
      {
        status: '已完成',
        statusClass: 'gray',
        time: '5天前',
        expert: '陈工',
        player: '王总',
        title: 'UI设计服务',
        budget: '预算：¥600 | 你的奖励：¥60',
        steps: ['已到账'],
        tail: '',
        actions: []
      },
      {
        status: '超时',
        statusClass: 'red',
        time: '已超时15天',
        expert: '刘设计师',
        player: '赵客户',
        title: '技术咨询服务',
        budget: '系统已自动发送超时预警，建议联系双方确认状态',
        steps: [],
        tail: '',
        actions: ['查看详情']
      }
    ]
  }
})
