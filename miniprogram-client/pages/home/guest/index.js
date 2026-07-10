const homeService = require('../../../services/home')
const toast = require('../../../utils/toast')
const { ROUTES } = require('../../../config/routes')
const { navigateShellRoute } = require('../../../utils/shell-nav')

Page({
  data: {
    onlineText: '',
    nearbyGames: []
  },

  onLoad() {
    this.loadGuestHome()
  },

  async loadGuestHome() {
    try {
      const homeData = await homeService.getHome()
      this.setData({
        onlineText: homeData.hero ? homeData.hero.onlineText : '',
        nearbyGames: this.normalizeNearbyGames(homeData.nearbyGames)
      })
    } catch (error) {
      this.setData({
        onlineText: '在线人数获取中',
        nearbyGames: []
      })
    }
  },

  normalizeNearbyGames(items) {
    return (Array.isArray(items) ? items : []).map((item) => ({
      id: item.id || item.gameId || '',
      title: item.title || item.name || '附近组局',
      place: item.place || item.locationText || item.address || '附近',
      time: item.time || item.timeText || item.startTimeText || '',
      memberText: item.memberText || item.membersText || `${item.currentPlayers || item.joinedCount || 0}/${item.maxPlayers || 8}人`
    }))
  },

  goLogin() {
    navigateShellRoute(ROUTES.login)
  },

  goInvite() {
    navigateShellRoute(ROUTES.login)
  }
})
