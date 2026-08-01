const profileService = require('../../../services/profile')
const toast = require('../../../utils/toast')
const { getActiveRole, normalizeActiveRole } = require('../../../utils/active-role')

function toText(value, fallback = '') {
  if (value === undefined || value === null || value === '') {
    return fallback
  }

  return String(value)
}

function toNumber(value, fallback = 0) {
  const number = Number(value)

  return Number.isFinite(number) ? number : fallback
}

function clampPercent(value) {
  return Math.max(0, Math.min(100, toNumber(value, 0)))
}

function normalizeAchievement(item = {}, index = 0) {
  const progress = clampPercent(item.progressPercent || item.progress || item.percent)

  return {
    id: item.id || item.code || `achievement-${index}`,
    title: toText(item.title || item.name, '未命名成就'),
    desc: toText(item.desc || item.description || item.summary),
    statusText: toText(item.statusText || item.statusLabel || item.status),
    unlocked: item.unlocked === true || item.status === 'unlocked',
    progressText: `${progress}%`,
    progressStyle: `width: ${progress}%;`
  }
}

function normalizeFootprint(item = {}, index = 0) {
  return {
    id: item.id || `footprint-${index}`,
    title: toText(item.title || item.gameTitle || item.name, '未命名足迹'),
    desc: toText(item.desc || item.description || item.cityName || item.locationName),
    time: toText(item.time || item.createdAt || item.completedAt)
  }
}

function normalizeRoleGrowth(item = {}, index = 0, activeRoleCode = 'player') {
  const metrics = Array.isArray(item.metrics) ? item.metrics : []
  const active = item.active === true
  const roleCode = normalizeActiveRole(item.roleCode)

  return {
    id: item.roleCode || `role-growth-${index}`,
    roleCode,
    roleName: toText(item.roleName, '角色'),
    levelTitle: toText(item.levelTitle, active ? '等级计算中' : '身份未开通'),
    active,
    selected: roleCode === activeRoleCode,
    scoreText: active ? `${toText(item.score, '0')} ${toText(item.scoreLabel, '')}`.trim() : '开通后开始累计',
    description: toText(item.description, active ? '数据将在业务完成后同步更新' : '开通对应身份后可查看'),
    metrics: metrics.map((metric, metricIndex) => {
      const score = clampPercent(metric.score)
      const rawValue = toNumber(metric.rawValue, 0)
      const targetValue = toNumber(metric.targetValue, 0)
      const unit = toText(metric.unit)
      return {
        id: metric.metricCode || `metric-${metricIndex}`,
        title: toText(metric.title, '指标'),
        valueText: targetValue > 0 ? `${rawValue}${unit} / ${targetValue}${unit}` : `${rawValue}${unit}`,
        progressText: `${score}%`,
        progressStyle: `width: ${score}%;`
      }
    })
  }
}

function firstValue(...values) {
  return values.find((value) => value !== undefined && value !== null && value !== '')
}

function roleName(roleCode) {
  return { player: '玩家', expert: '行家', guide: '领路人' }[roleCode] || '玩家'
}

function normalizeGrowth(data = {}, requestedRole = 'player') {
  const growth = data.growth || data
  const config = growth.achievementConfig || {}
  const unlocked = Array.isArray(growth.achievements) ? growth.achievements : []
  const locked = Array.isArray(config.locked) ? config.locked : []
  const achievements = unlocked.concat(locked)
  const footprints = Array.isArray(growth.footprints) ? growth.footprints : []
  const level = growth.level || growth.levelInfo || {}
  const credit = growth.credit || growth.creditInfo || {}
  const points = growth.points || growth.pointsInfo || {}
  const roleProfiles = Array.isArray(growth.roleGrowth) ? growth.roleGrowth : []
  const activeRoleCode = normalizeActiveRole(growth.activeRoleCode || requestedRole)
  const activeRole = growth.activeRoleGrowth && growth.activeRoleGrowth.roleCode === activeRoleCode
    ? growth.activeRoleGrowth
    : (roleProfiles.find((item) => item && item.roleCode === activeRoleCode) || {})
  const score = toNumber(firstValue(activeRole.score, growth.profile?.experience, growth.experience), 0)
  const scoreUnit = activeRoleCode === 'player' ? '经验' : '分'

  return {
    activeRoleCode,
    activeRoleName: toText(growth.activeRoleName || activeRole.roleName, roleName(activeRoleCode)),
    levelName: toText(activeRole.levelTitle || level.name || level.title || growth.levelName, '暂无等级'),
    levelValue: toText(firstValue(activeRole.levelNo, level.value, level.level)),
    xpText: `${score} ${scoreUnit}`,
    xpLabel: toText(activeRole.scoreLabel, activeRoleCode === 'player' ? '累计经验' : '综合得分'),
    creditText: toText(firstValue(credit.scoreText, credit.score, growth.creditScore), '0'),
    pointsText: toText(firstValue(points.availableText, points.availablePoints, growth.availablePoints), '0'),
    achievementCount: achievements.length,
    footprintCount: footprints.length,
    achievements: achievements.map(normalizeAchievement),
    footprints: footprints.map(normalizeFootprint),
    roleProfiles: roleProfiles.map((item, index) => normalizeRoleGrowth(item, index, activeRoleCode))
  }
}

Page({
  data: {
    loading: false,
    loadError: '',
    levelName: '暂无等级',
    levelValue: '',
    activeRoleCode: 'player',
    activeRoleName: '玩家',
    xpText: '0 经验',
    xpLabel: '累计经验',
    creditText: '0',
    pointsText: '0',
    achievementCount: 0,
    footprintCount: 0,
    achievements: [],
    footprints: [],
    roleProfiles: []
  },

  onLoad() {
    this.loadGrowth()
  },

  async loadGrowth() {
    this.setData({
      loading: true,
      loadError: ''
    })

    try {
      const activeRole = getActiveRole()
      const data = await profileService.getGrowth({ roleType: activeRole })

      this.setData({
        loading: false,
        loadError: '',
        ...normalizeGrowth(data, activeRole)
      })
    } catch (error) {
      this.setData({
        loading: false,
        loadError: error.message || '成就数据加载失败',
        achievements: [],
        footprints: [],
        roleProfiles: []
      })
      toast.info(error.message || '成就数据加载失败')
    }
  }
})
