const { ROUTES } = require('../../../config/routes')
const mapService = require('../../../services/map')
const toast = require('../../../utils/toast')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')

const NAV_ITEMS = [
  { name: '我的', key: 'mine' },
  { name: '元宇宙', key: 'metaverse' },
  { name: '入局', key: 'map' },
  { name: '创局', key: 'message' },
  { name: '首页', key: 'home' }
]

const EMPTY_PAGE = {
  title: '',
  challengeTitle: '',
  rankingTitle: '',
  nationalRankText: '',
  challengeToast: '{title}',
  emptyChallengeText: '',
  rankingToast: '{rank}',
  emptyRankingText: '',
  nationalToast: '',
  duel: {},
  challenges: [],
  rankings: []
}

function applyTemplate(template, values = {}) {
  let text = String(template || '')
  Object.keys(values).forEach((key) => {
    text = text.replace(new RegExp(`\\{${key}\\}`, 'g'), String(values[key]))
  })
  return text
}

Page({
  data: {
    onlineText: '在线',
    navItems: NAV_ITEMS,
    pageConfig: EMPTY_PAGE,
    duel: {},
    challenges: [],
    rankings: []
  },

  onLoad() {
    this.loadPageConfig()
  },

  async loadPageConfig() {
    try {
      const pageConfig = Object.assign({}, EMPTY_PAGE, await mapService.getPlayPage('friend-city'))
      this.setData({
        pageConfig,
        duel: pageConfig.duel || {},
        challenges: pageConfig.challenges || [],
        rankings: pageConfig.rankings || []
      })
    } catch (error) {
      toast.info(error.message || '好友城市加载失败')
    }
  },

  handleBackTap() {
    wx.navigateBack({
      delta: 1,
      fail: () => {
        navigateShellRoute(ROUTES.map, {
          currentRoute: ROUTES.mapFriendCity
        })
      }
    })
  },

  handleStartChallenge() {
    const duel = this.data.duel || {}

    wx.showModal({
      title: duel.startButtonText || '发起挑战',
      content: '发起后会写入城市挑战记录，可继续进入打卡。',
      confirmText: '发起',
      cancelText: '取消',
      success: async (res) => {
        if (!res.confirm) {
          return
        }

        try {
          const challenge = await mapService.createChallenge({
            title: '城市挑战',
            total: 3
          })
          this.setData({
            challenges: [challenge].concat(this.data.challenges || [])
          })
          toast.info('挑战已发起')
        } catch (error) {
          toast.info(error.message || '发起挑战失败')
        }
      }
    })
  },

  handleRecordTap() {
    navigateShellRoute(ROUTES.profileFootprintAchievements || ROUTES.profile, {
      currentRoute: ROUTES.mapFriendCity
    })
  },

  handleChallengeTap(event) {
    const challenge = (this.data.challenges || []).find((item) => item.id === event.currentTarget.dataset.id)

    if (!challenge) {
      toast.info(this.data.pageConfig.emptyChallengeText || '暂无可打开的挑战')
      return
    }

    wx.showModal({
      title: challenge.title,
      content: [
        `状态：${challenge.statusText || '进行中'}`,
        `${this.data.duel.selfName || '我'}：${challenge.selfValue || 0}/${challenge.total || 0}`,
        `${this.data.duel.rivalName || '好友'}：${challenge.rivalValue || 0}/${challenge.total || 0}`,
        `剩余时间：${challenge.timeLeft || '-'}`
      ].join('\n'),
      confirmText: '去打卡',
      cancelText: '关闭',
      success: (res) => {
        if (!res.confirm) {
          return
        }

        navigateShellRoute(`${ROUTES.mapRealCheckin}?challengeId=${encodeURIComponent(challenge.id)}`, {
          currentRoute: ROUTES.mapFriendCity
        })
      }
    })
  },

  handleRankingTap(event) {
    const rank = event.currentTarget.dataset.rank

    wx.showToast({
      title: rank ? applyTemplate(this.data.pageConfig.rankingToast, { rank }) : this.data.pageConfig.emptyRankingText,
      icon: 'none'
    })
  },

  handleNationalRankTap() {
    wx.showToast({
      title: this.data.pageConfig.nationalToast || '全国排行加载中',
      icon: 'none'
    })
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}
    navigateShellKey(key, {
      currentRoute: ROUTES.mapFriendCity
    })
  }
})
