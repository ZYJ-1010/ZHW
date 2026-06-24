Page({
  data: {
    participants: [
      {
        name: '陆毅',
        avatarSrc: '/pages/home/player/assets/ranking-avatar-01.png',
        role: '玩家',
        roleClass: 'player',
        position: '总经理 | 上海创世界科技有限公司',
        topic: 'AI赋能与市场运营助力企业IP打造',
        primaryTag: '第一标签：上海TMT投资领军者',
        tags: ['数字化内容服务'],
        location: '上海市浦东新区沙新镇黄赵路310号',
        distance: '2.1 km'
      },
      {
        name: '林一',
        avatarSrc: '/pages/home/player/assets/ranking-avatar-02.png',
        role: '行家',
        roleClass: 'expert',
        position: '品牌创始人 | 杭州欣悦服装工作',
        topic: '企业家服务平台',
        primaryTag: '第一标签：女性高品质服装领先者',
        tags: ['品牌增长', '企业服务'],
        location: '杭州市上城区',
        distance: '2.1 km'
      },
      {
        name: '陈序',
        avatarSrc: '/pages/home/player/assets/ranking-avatar-03.png',
        role: '玩家',
        roleClass: 'player',
        position: '联合创始人 | 苏州智造咨询',
        topic: '制造业数字化流程重构',
        primaryTag: '第一标签：智能制造转型顾问',
        tags: ['流程管理', '资源对接'],
        location: '苏州市工业园区',
        distance: '4.6 km'
      },
      {
        name: '赵晴',
        avatarSrc: '/pages/home/player/assets/ranking-avatar-me.png',
        role: '行家',
        roleClass: 'expert',
        position: '运营负责人 | 上海云栖服务',
        topic: '企业服务增长策略',
        primaryTag: '第一标签：企业服务增长操盘手',
        tags: ['增长策略', '私域运营'],
        location: '上海市黄浦区',
        distance: '5.3 km'
      }
    ]
  },

  onParticipantTap(event) {
    const participant = event.detail && event.detail.participant
    const name = (participant && participant.name) || '参与者'

    wx.showToast({
      title: `${name}资料待接入`,
      icon: 'none'
    })
  }
})
