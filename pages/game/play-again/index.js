const RECOMMEND_OPTIONS = [
  {
    id: 'same-friends',
    theme: 'green',
    iconText: '👫',
    title: '同局好友再玩一局',
    desc: '立即邀请王总、张专家等3人'
  },
  {
    id: 'smart-match',
    theme: 'blue',
    iconText: '🤖',
    title: '系统推荐适配组局',
    desc: '基于你的偏好，已找到5个高匹配度新局'
  },
  {
    id: 'create-new',
    theme: 'pink',
    iconType: 'plus',
    title: '玩家创建新局',
    desc: '自定义需求，开启全新体验'
  }
]

Page({
  data: {
    onlineText: '3999人在线',
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    recommendOptions: RECOMMEND_OPTIONS,
    selectedOption: 'same-friends'
  },

  onOptionTap(event) {
    const id = event.currentTarget.dataset.id
    const option = RECOMMEND_OPTIONS.find((item) => item.id === id)

    if (!option) {
      return
    }

    this.setData({
      selectedOption: id
    })
    this.showInfo(`${option.title}待接入`)
  },

  onCloseTap() {
    this.showInfo('关闭推荐面板待接入')
  },

  handleShellNavTap() {
    this.showInfo('功能正在开发中')
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
