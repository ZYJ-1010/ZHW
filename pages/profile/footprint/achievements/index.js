const { ROUTES } = require('../../../../config/routes')
const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const NAV_ITEMS = [
  { name: '我的', key: 'mine' },
  { name: '元宇宙', key: 'metaverse' },
  { name: '地图', key: 'map' },
  { name: '消息', key: 'message' },
  { name: '首页', key: 'home' }
]

function filterAchievements(list, filterKey) {
  if (filterKey === 'all') {
    return list
  }

  return list.filter((item) => item.category === filterKey || item.categoryKey === filterKey)
}

function normalizeList(list) {
  return Array.isArray(list) ? list : []
}

function splitAchievements(data = {}) {
  const achieved = normalizeList(data.achieved || data.unlocked || data.completed)
  const locked = normalizeList(data.locked || data.uncompleted)

  if (achieved.length || locked.length) {
    return {
      achieved,
      locked
    }
  }

  const list = normalizeList(data.achievements || data.list || data.items)

  return list.reduce((result, item) => {
    const isLocked = item.locked === true || item.unlocked === false || item.status === 'locked'

    result[isLocked ? 'locked' : 'achieved'].push(item)
    return result
  }, {
    achieved: [],
    locked: []
  })
}

Page({
  data: {
    onlineText: '',
    navItems: NAV_ITEMS,
    activeFilter: 'all',
    filters: [],
    levelTitle: '',
    levelTip: '',
    progress: 0,
    achievedCount: 0,
    lockedCount: 0,
    achieved: [],
    locked: [],
    allAchieved: [],
    allLocked: [],
    season: {
      title: '',
      status: '',
      remain: '',
      achieved: '',
      locked: ''
    }
  },

  onLoad() {
    this.loadAchievements()
  },

  async loadAchievements() {
    try {
      const data = await profileService.getProfileAchievements({
        filter: this.data.activeFilter
      })
      const level = data.level || data.summary || {}
      const lists = splitAchievements(data)

      this.setData({
        onlineText: data.onlineText || '',
        filters: normalizeList(data.filters || data.categories),
        levelTitle: data.levelTitle || level.title || '',
        levelTip: data.levelTip || level.tip || level.nextLevelText || '',
        progress: Number(data.progress || data.progressPercent || level.progress || level.progressPercent) || 0,
        achievedCount: Number(data.achievedCount || data.counts && data.counts.achieved) || lists.achieved.length,
        lockedCount: Number(data.lockedCount || data.counts && data.counts.locked) || lists.locked.length,
        allAchieved: lists.achieved,
        allLocked: lists.locked,
        achieved: filterAchievements(lists.achieved, this.data.activeFilter),
        locked: filterAchievements(lists.locked, this.data.activeFilter),
        season: {
          ...this.data.season,
          ...(data.season || {})
        }
      })
    } catch (error) {
      this.setData({
        filters: [],
        achieved: [],
        locked: [],
        allAchieved: [],
        allLocked: []
      })
      toast.info(error.message || '成就墙加载失败')
    }
  },

  handleFilterTap(event) {
    const { key } = event.currentTarget.dataset

    this.setData({
      activeFilter: key,
      achieved: filterAchievements(this.data.allAchieved, key),
      locked: filterAchievements(this.data.allLocked, key)
    })
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}

    if (key === 'map') {
      wx.showToast({
        title: '地图功能开发中',
        icon: 'none'
      })
      return
    }
    const routeMap = {
      home: ROUTES.playerHome,
      map: '',
      message: ROUTES.message,
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
