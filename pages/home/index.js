const homeService = require('../../services/home')
const toast = require('../../utils/toast')

Page({
  data: {
    loading: true,
    user: {
      nickname: '',
      todayCreditScore: '',
      level: '',
      member: {
        planName: ''
      },
      needRealname: true
    },
    hero: {
      roleName: '',
      dateLabel: '',
      subtitle: '',
      onlineText: ''
    },
    notices: [],
    quickActions: [],
    recommendedGames: [],
    playerSummary: {
      displayName: '',
      roleLabel: '',
      title: '',
      xpText: '',
      progressPercent: 0,
      nextLevelText: '',
      stats: []
    },
    rankingList: [],
    achievementList: [],
    friendGames: [],
    metaverseEntry: {
      title: '',
      desc: '',
      actionText: '',
      route: ''
    },
    nearbySummary: {
      cityName: '',
      count: 0
    }
  },

  onLoad() {
    this.loadHome()
  },

  async loadHome() {
    try {
      const home = await homeService.getHome()
      this.setData({
        loading: false,
        user: home.user,
        hero: home.hero,
        notices: home.notices,
        quickActions: home.quickActions,
        recommendedGames: home.recommendedGames,
        playerSummary: home.playerSummary,
        rankingList: home.rankingList,
        achievementList: home.achievementList,
        friendGames: home.friendGames,
        metaverseEntry: home.metaverseEntry,
        nearbySummary: home.nearbySummary
      })
    } catch (error) {
      this.setData({
        loading: false
      })
      toast.info(error.message || '首页加载失败')
    }
  },

  goAction(event) {
    const route = event.currentTarget.dataset.route

    if (!route) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  }
})
