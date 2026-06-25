const { ROUTES } = require('../../../config/routes')

const NAV_ITEMS = [
  { name: '我的', key: 'mine' },
  { name: '元宇宙', key: 'metaverse' },
  { name: '入局', key: 'map' },
  { name: '创局', key: 'message' },
  { name: '首页', key: 'home' }
]

const CHALLENGES = [
  {
    id: 'jingan-first',
    title: '率先点亮静安区',
    timeLeft: '2天',
    statusText: '进行中',
    statusTone: 'pending',
    selfValue: 3,
    rivalValue: 2,
    total: 5,
    selfPercent: 60,
    rivalPercent: 40
  },
  {
    id: 'landmark-speed',
    title: '10个地标速通',
    timeLeft: '5天',
    statusText: '领先中',
    statusTone: 'leading',
    selfValue: 7,
    rivalValue: 4,
    total: 10,
    selfPercent: 70,
    rivalPercent: 40
  }
]

const RANKINGS = [
  { rank: 1, name: 'Mike', desc: '已点亮 28 区', score: '2,450', tone: 'gold' },
  { rank: 2, name: 'Sarah', desc: '已点亮 24 区', score: '2,180', tone: 'silver' },
  { rank: 3, name: 'David', desc: '已点亮 22 区', score: '1,950', tone: 'bronze' },
  { rank: 4, name: '我', desc: '已点亮 12 区', score: '1,240', tone: 'normal' }
]

Page({
  data: {
    onlineText: '3999人在线',
    navItems: NAV_ITEMS,
    challenges: CHALLENGES,
    rankings: RANKINGS
  },

  handleBackTap() {
    wx.navigateBack({
      delta: 1,
      fail: () => {
        wx.navigateTo({
          url: `/${ROUTES.map}`
        })
      }
    })
  },

  handleStartChallenge() {
    wx.showToast({
      title: '发起好友挑战待接入',
      icon: 'none'
    })
  },

  handleRecordTap() {
    wx.showToast({
      title: '挑战记录待接入',
      icon: 'none'
    })
  },

  handleChallengeTap(event) {
    const challenge = CHALLENGES.find((item) => item.id === event.currentTarget.dataset.id)

    wx.showToast({
      title: challenge ? `${challenge.title}待接入` : '挑战详情待接入',
      icon: 'none'
    })
  },

  handleRankingTap(event) {
    const rank = event.currentTarget.dataset.rank

    wx.showToast({
      title: rank ? `第${rank}名详情待接入` : '城市榜详情待接入',
      icon: 'none'
    })
  },

  handleNationalRankTap() {
    wx.showToast({
      title: '全国榜待接入',
      icon: 'none'
    })
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}
    const routeMap = {
      home: ROUTES.playerHome,
      map: ROUTES.gameHall,
      message: ROUTES.gameCreate,
      mine: ROUTES.profile,
      metaverse: ROUTES.metaverse
    }
    const route = routeMap[key]

    if (!route) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  }
})
