const { ROUTES } = require('../../../../config/routes')
const profileService = require('../../../../services/profile')
const { navigateShellKey } = require('../../../../utils/shell-nav')
const { getActiveRole, normalizeActiveRole } = require('../../../../utils/active-role')

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

function normalizeRoleGrowth(item = {}, index = 0, activeRoleCode = 'player') {
  const active = item.active === true
  const roleCode = normalizeActiveRole(item.roleCode)
  return {
    id: item.roleCode || `role-${index}`,
    roleCode,
    roleName: item.roleName || '角色',
    levelTitle: item.levelTitle || (active ? '等级计算中' : '身份未开通'),
    scoreText: active ? `${Number(item.score || 0)} ${item.scoreLabel || ''}`.trim() : '待开通',
    description: item.description || (active ? '数据实时计算' : '开通后开始累计'),
    active,
    selected: roleCode === activeRoleCode
  }
}

function buildAchievementState(data = {}, requestedRole = 'player') {
  const profile = data.profile || {}
  const activeRoleCode = normalizeActiveRole(data.activeRoleCode || requestedRole)
  const rawRoleGrowth = Array.isArray(data.roleGrowth) ? data.roleGrowth : []
  const roleGrowth = rawRoleGrowth.map((item, index) => normalizeRoleGrowth(item, index, activeRoleCode))
  const rawActiveRole = data.activeRoleGrowth && data.activeRoleGrowth.roleCode === activeRoleCode
    ? data.activeRoleGrowth
    : (rawRoleGrowth.find((item) => item && item.roleCode === activeRoleCode) || null)
  const activeRole = rawActiveRole ? normalizeRoleGrowth(rawActiveRole, 0, activeRoleCode) : null
  const activeMetric = rawActiveRole && Array.isArray(rawActiveRole.metrics) ? rawActiveRole.metrics[0] : null
  const config = data.achievementConfig || {}
  const filters = Array.isArray(config.filters) && config.filters.length ? config.filters : FALLBACK_FILTERS
  const achievements = Array.isArray(data.achievements)
    ? data.achievements
    : (Array.isArray(profile.achievementItems) ? profile.achievementItems : (Array.isArray(profile.achievements) ? profile.achievements : []))
  const achieved = achievements.map(normalizeAchievement).filter((item) => item.id)
  const locked = Array.isArray(config.locked) ? config.locked.map(normalizeAchievement).filter((item) => item.id) : []
  const level = Number(activeRole ? rawActiveRole.levelNo : profile.level) || 0
  const experience = Number(profile.experience) || 0
  const progressSource = rawActiveRole && rawActiveRole.progressPercent !== undefined
    ? rawActiveRole.progressPercent
    : (activeRoleCode === 'player' ? (activeMetric && activeMetric.score) : (rawActiveRole && rawActiveRole.score))
  const progress = Math.max(0, Math.min(100, Math.round(Number(progressSource || 0))))
  const footprintCount = Array.isArray(data.footprints) ? data.footprints.length : 0
  const season = config.season || {}
  const remainTpl = season.remainTpl || '已沉淀 {footprintCount} 条足迹'

  return {
    filters,
    onlineText: data.onlineText || '成长中心',
    activeRoleCode,
    activeRoleName: data.activeRoleName || (activeRole && activeRole.roleName) || '玩家',
    levelTitle: activeRole ? activeRole.levelTitle : `${config.levelTitlePrefix || 'Lv.'}${level}`,
    levelTip: activeRole ? activeRole.description : `已累计 ${experience} 经验`,
    progress,
    roleGrowth,
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
    roleGrowth: [],
    activeRoleCode: 'player',
    activeRoleName: '玩家',
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
      const activeRole = getActiveRole()
      const growth = await profileService.getGrowth({ roleType: activeRole })
      const state = buildAchievementState(growth, activeRole)
      this.setData({
        ...state,
        filters: state.filters,
        allAchieved: state.achieved,
        allLocked: state.locked,
        roleGrowth: state.roleGrowth,
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
