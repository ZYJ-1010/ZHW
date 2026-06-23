const BENEFITS = [
  {
    iconText: '⭐',
    theme: 'blue',
    title: '优先推荐权益',
    desc: '下次组局时优先展示给行家'
  },
  {
    iconText: '👑',
    theme: 'purple',
    title: '评价达人徽章',
    desc: '累计评价3次可获得专属标识'
  },
  {
    iconText: '📈',
    theme: 'red',
    title: '信用分提升',
    desc: '活跃评价有助于提升账号权重'
  }
]

const PLAY_OPTIONS = [
  {
    id: 'again',
    theme: 'green',
    title: '再玩一局',
    desc: '随时可约'
  },
  {
    id: 'pause',
    theme: 'yellow',
    title: '暂停',
    desc: '想歇歇'
  },
  {
    id: 'stop',
    theme: 'red',
    title: '不玩了',
    desc: '不再参与'
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
    pageScrollTop: 0,
    benefits: BENEFITS,
    playOptions: PLAY_OPTIONS,
    selectedPlayIntent: 'again'
  },

  handlePageScroll(event) {
    this.lastScrollTop = event.detail.scrollTop || 0
  },

  handleShellNavTap(event) {
    const { key } = event.detail

    if (key === 'up') {
      this.scrollBy(-180)
      return
    }

    if (key === 'down') {
      this.scrollBy(180)
      return
    }

    this.showInfo('功能正在开发中')
  },

  handleShellNavLongPress(event) {
    const { key } = event.detail

    if (key === 'up') {
      this.scrollBy(-520)
      return
    }

    if (key === 'down') {
      this.scrollBy(520)
    }
  },

  handleShellNavTouchEnd() {},

  scrollBy(distance) {
    const nextTop = Math.max(0, Math.round((this.lastScrollTop || 0) + distance))

    this.setData({
      pageScrollTop: nextTop
    })
  },

  onPlayIntentTap(event) {
    const id = event.currentTarget.dataset.id

    if (!id) {
      return
    }

    this.setData({
      selectedPlayIntent: id
    })
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
