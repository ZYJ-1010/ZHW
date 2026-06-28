const toast = require('../../../../utils/toast')

const ASSET_BASE = '/pages/profile/system-management/feedback/assets'

Page({
  data: {
    icons: {
      attachment: `${ASSET_BASE}/icon-paperclip.svg`,
      send: `${ASSET_BASE}/icon-send-plane.svg`
    },
    feedbackImage: `${ASSET_BASE}/feedback-detail-rail.png`,
    messages: [
      {
        id: 'user-1',
        role: 'me',
        avatar: '我',
        time: '18:32',
        content: '建议在组局详情页增加「一键复制地址」功能'
      },
      {
        id: 'service-1',
        role: 'service',
        avatar: '客',
        time: '18:45',
        content: '您好，感谢您的建议！我们已记录该需求，产品团队正在评估中。请问您希望复制的是完整地址还是仅复制门店名称呢？'
      },
      {
        id: 'user-2',
        role: 'me',
        avatar: '我',
        time: '18:50',
        content: '完整地址比较好，另外最好能支持一键跳转地图导航'
      },
      {
        id: 'service-2',
        role: 'service',
        avatar: '客',
        time: '19:05',
        content: '收到！这个需求很有价值，我们已加入下个迭代的优先级队列。预计两周内上线，上线后会第一时间通知您。'
      }
    ]
  },

  handleAttachTap() {
    toast.developing('补充截图待接入文件接口')
  },

  handleSendTap() {
    toast.developing('补充说明提交待接入反馈接口')
  }
})
