const { ROUTES } = require('../../../../config/routes')
const profileService = require('../../../../services/profile')
const { navigateShellKey } = require('../../../../utils/shell-nav')

const NAV_ITEMS = [
  { name: '我的', key: 'mine' },
  { name: '元宇宙', key: 'metaverse' },
  { name: '地图', key: 'map' },
  { name: '消息', key: 'message' },
  { name: '首页', key: 'home' }
]

const FALLBACK_FILTERS = [
  { key: 'all', label: '全部' }
]

function filterAchievements(list, filterKey) {
  if (filterKey === 'all') {
    return list
  }

  return list.filter((item) => item.category === filterKey)
}

function normalizeAchievement(item) {
  const source = item && typeof item === 'object' ? item : { code: item }
  const id = String(source.id || source.code || '').trim()

  return {
    id,
    title: source.title || source.name || id || '成长成就',
    desc: source.desc || source.description || '来自后端成长记录',
    icon: source.icon || '/pages/profile/footprint/achievements/assets/icon-star.png',
    tone: source.tone || 'gold',
    category: source.category || 'city',
    progressPercent: Number(source.progressPercent || source.progress || 0) || 0
  }
}

function buildAchievementState(data = {}) {
  const profile = data.profile || {}
  const config = data.achievementConfig || {}
  const filters = Array.isArray(config.filters) && config.filters.length ? config.filters : FALLBACK_FILTERS
  const achievements = Array.isArray(data.achievements)
    ? data.achievements
    : (Array.isArray(profile.achievementItems) ? profile.achievementItems : (Array.isArray(profile.achievements) ? profile.achievements : []))
  const achieved = achievements.map(normalizeAchievement).filter((item) => item.id)
  const locked = Array.isArray(config.locked) ? config.locked.map(normalizeAchievement).filter((item) => item.id) : []
  const level = Number(profile.level) || 1
  const experience = Number(profile.experience) || 0
  const nextLevelExperience = Math.max(level * 100, 100)
  const progress = Math.max(0, Math.min(100, Math.round((experience % 100) / 100 * 100)))
  const remaining = Math.max(0, nextLevelExperience - experience)
  const footprintCount = Array.isArray(data.footprints) ? data.footprints.length : 0
  const season = config.season || {}
  const remainTpl = season.remainTpl || '已沉淀 {footprintCount} 条足迹'

  return {
    filters,
    onlineText: `${Math.max(1, achieved.length + locked.length)}${config.onlineSuffix || '人成长中'}`,
    levelTitle: `${config.levelTitlePrefix || 'Lv.'}${level}`,
    levelTip: remaining > 0 ? `再获得 ${remaining} 点升级` : '已达到当前等级目标',
    progress,
    achievedCount: achieved.length,
    lockedCount: locked.length,
    achieved,
    locked,
    season: {
      title: season.title || '赛季',
      status: season.status || '进行中',
      remain: remainTpl.replace('{footprintCount}', String(footprintCount)),
      achieved: achieved.length,
      locked: locked.length
    }
  }
}

Page({
  data: {
    onlineText: '加载中',
    navItems: NAV_ITEMS,
    activeFilter: 'all',
    filters: FALLBACK_FILTERS,
    levelTitle: 'Lv.1',
    levelTip: '成长数据加载中',
    progress: 0,
    achievedCount: 0,
    lockedCount: 0,
    achieved: [],
    locked: [],
    allAchieved: [],
    allLocked: [],
    season: {
      title: '赛季',
      status: '进行中',
      remain: '成长足迹加载中',
      achieved: 0,
      locked: 0
    }
  },

  onLoad() {
    this.loadAchievements()
  },

  async loadAchievements() {
    try {
      const growth = await profileService.getGrowth()
      const state = buildAchievementState(growth)
      this.setData({
        ...state,
        filters: state.filters,
        allAchieved: state.achieved,
        allLocked: state.locked,
        achieved: filterAchievements(state.achieved, this.data.activeFilter),
        locked: filterAchievements(state.locked, this.data.activeFilter)
      })
    } catch (error) {
      console.warn('get achievements failed', error)
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
    navigateShellKey(key, {
      currentRoute: ROUTES.profileFootprintAchievements
    })
  }
})
