const profileService = require('../../../services/profile')
const toast = require('../../../utils/toast')

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

function normalizeGrowth(data = {}) {
  const growth = data.growth || data
  const config = growth.achievementConfig || {}
  const unlocked = Array.isArray(growth.achievements) ? growth.achievements : []
  const locked = Array.isArray(config.locked) ? config.locked : []
  const achievements = unlocked.concat(locked)
  const footprints = Array.isArray(growth.footprints) ? growth.footprints : []
  const level = growth.level || growth.levelInfo || {}
  const credit = growth.credit || growth.creditInfo || {}
  const points = growth.points || growth.pointsInfo || {}

  return {
    levelName: toText(level.name || level.title || growth.levelName, '暂无等级'),
    levelValue: toText(level.value || level.level || growth.level || ''),
    xpText: toText(level.xpText || growth.xpText || growth.experienceText, '0 XP'),
    creditText: toText(credit.scoreText || credit.score || growth.creditScore, '0'),
    pointsText: toText(points.availableText || points.availablePoints || growth.availablePoints, '0'),
    achievementCount: achievements.length,
    footprintCount: footprints.length,
    achievements: achievements.map(normalizeAchievement),
    footprints: footprints.map(normalizeFootprint)
  }
}

Page({
  data: {
    loading: false,
    loadError: '',
    levelName: '暂无等级',
    levelValue: '',
    xpText: '0 XP',
    creditText: '0',
    pointsText: '0',
    achievementCount: 0,
    footprintCount: 0,
    achievements: [],
    footprints: []
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
      const data = await profileService.getGrowth()

      this.setData({
        loading: false,
        loadError: '',
        ...normalizeGrowth(data)
      })
    } catch (error) {
      this.setData({
        loading: false,
        loadError: error.message || '成就数据加载失败',
        achievements: [],
        footprints: []
      })
      toast.info(error.message || '成就数据加载失败')
    }
  }
})
