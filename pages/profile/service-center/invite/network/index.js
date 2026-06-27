Page({
  data: {
    summary: [
      { value: '156', label: '已服务\n位玩家' },
      { value: '¥1,240', label: '本周收益' }
    ],
    networkNodes: ['1', '2', '3', '4', '5', '6', '7', '8'],
    avatars: ['👨‍💼', '👩‍💻', '👨‍🎨', '👩‍🎤'],
    members: [
      { id: 'member-001', avatar: '👨‍💼', name: '张大山', desc: '邀约 86 · 转化 32 · 活跃 12天', direct: '+', team: '利用 +¥1,860' },
      { id: 'member-002', avatar: '👩‍💻', name: '李小红', desc: '邀约 64 · 转化 28 · 活跃 9天', direct: '+', team: '利用 +¥1,200' },
      { id: 'member-003', avatar: '👨‍🎨', name: '王建国', desc: '邀约 45 · 转化 19 · 活跃 7天', direct: '+', team: '利用 +¥680' },
      { id: 'member-004', avatar: '👩‍🎤', name: '赵小美', desc: '邀约 38 · 转化 15 · 活跃 5天', direct: '+', team: '利用 +¥520' },
      { id: 'member-005', avatar: '👨‍🔬', name: '陈博士', desc: '邀约 31 · 转化 12 · 活跃 4天', direct: '+', team: '利用 +¥340' }
    ]
  },

  handleMemberTap(event) {
    const { id } = event.currentTarget.dataset

    wx.navigateTo({
      url: `/pages/profile/service-center/invite/member-detail/index?id=${id}`
    })
  }
})
