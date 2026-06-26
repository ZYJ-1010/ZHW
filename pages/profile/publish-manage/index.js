Page({
  data: {
    tabs: [
      { label: '展示中(5)', active: true },
      { label: '已下架(2)', active: false },
      { label: '审核中(1)', active: false }
    ],
    services: [
      {
        title: '产品架构梳理咨询',
        status: '展示中',
        statusClass: 'green',
        desc: '资深产品经理提供产品架构设计、MVP规划等服务',
        views: '浏览 256',
        deals: '成交 12',
        actions: ['编辑', '置顶', '下架', '数据']
      },
      {
        title: 'UI/UX设计服务',
        status: '展示中',
        statusClass: 'green',
        desc: 'APP界面设计、小程序设计、网页设计等服务',
        views: '浏览 89',
        deals: '成交 5',
        actions: ['编辑', '置顶', '下架', '数据']
      },
      {
        title: '产品需求文档撰写',
        status: '已下架',
        statusClass: 'gray',
        desc: 'PRD文档撰写、需求梳理、功能规划等服务',
        views: '浏览 45',
        deals: '成交 3',
        actions: ['重新上架', '编辑', '删除']
      }
    ]
  }
})
