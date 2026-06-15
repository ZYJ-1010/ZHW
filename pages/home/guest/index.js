const homeService = require('../../../services/home')
const toast = require('../../../utils/toast')

Page({
  data: {
    onlineText: '',
    nearbyGames: [
      {
        id: 'guest-1',
        title: '苏州河记忆碎片采集',
        place: '静安区 · 2.1km',
        time: '2026.05.15 20:00-22:00',
        memberText: '0/8人'
      }
    ]
  },

  onLoad() {
    this.loadGuestHome()
  },

  async loadGuestHome() {
    try {
      const homeData = await homeService.getHome()
      this.setData({
        onlineText: homeData.hero ? homeData.hero.onlineText : ''
      })
    } catch (error) {
      this.setData({
        onlineText: '在线人数获取中'
      })
    }
  },

  goLogin() {
    toast.developing()
  },

  goInvite() {
    toast.developing()
  }
})
