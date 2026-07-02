const ASSET_BASE = '/pages/profile/system-management/service-case-detail/assets'
const profileService = require('../../../../services/profile')

function buildStars(activeCount = 5) {
  return Array.from({ length: 5 }, (_, index) => ({
    key: `star-${index}`,
    active: index < activeCount
  }))
}

function decodeOption(value) {
  if (!value) {
    return ''
  }

  try {
    return decodeURIComponent(value)
  } catch (error) {
    return value
  }
}

function buildEmptyCaseDetail(iconOptions = {}) {
  return {
    caseInfo: {
      title: '服务案例',
      date: '',
      playersText: '加载中',
      totalPlayers: 0,
      ratingText: '待评分',
      iconText: iconOptions.iconText || '★',
      tone: iconOptions.tone || 'blue',
      stars: buildStars(0)
    },
    rating: {
      score: '待评分',
      tags: [],
      stars: buildStars(0)
    },
    detailSections: [],
    players: []
  }
}

function normalizeRemoteCaseDetail(detail = {}, iconOptions = {}) {
  const caseInfo = detail.caseInfo || {}
  const rating = detail.rating || {}
  const score = Number(rating.score || String(caseInfo.ratingText || '').replace('分', ''))
  const starCount = Number.isFinite(score) ? Math.max(1, Math.min(5, Math.round(score))) : 5

  return {
    caseInfo: {
      title: caseInfo.title || '服务案例',
      date: caseInfo.date || '',
      playersText: caseInfo.playersText || '待绑定',
      totalPlayers: Number(caseInfo.totalPlayers) || 0,
      ratingText: caseInfo.ratingText || `${Number.isFinite(score) ? score.toFixed(1) : '5.0'}分`,
      iconText: iconOptions.iconText || caseInfo.iconText || '★',
      tone: iconOptions.tone || caseInfo.tone || 'blue',
      stars: buildStars(starCount)
    },
    rating: {
      score: rating.score || (Number.isFinite(score) ? score.toFixed(1) : '5.0'),
      tags: Array.isArray(rating.tags) ? rating.tags : [],
      stars: buildStars(starCount)
    },
    detailSections: Array.isArray(detail.detailSections) ? detail.detailSections : [],
    players: Array.isArray(detail.players) ? detail.players : []
  }
}

Page({
  data: {
    assets: {
      calendar: `${ASSET_BASE}/icon-calendar.svg`,
      users: `${ASSET_BASE}/icon-users.svg`,
      caseFile: `${ASSET_BASE}/icon-case-file.svg`,
      star: `${ASSET_BASE}/icon-star.svg`,
      avatar: `${ASSET_BASE}/avatar-default.png`
    },
    ...buildEmptyCaseDetail()
  },

  onLoad(options = {}) {
    const iconOptions = {
      iconText: decodeOption(options.iconText),
      tone: decodeOption(options.tone)
    }
    const caseId = options.caseId || 'case-stranger'

    this.setData(buildEmptyCaseDetail(iconOptions))
    this.loadRemoteCaseDetail(caseId, iconOptions)
  },

  async loadRemoteCaseDetail(caseId, iconOptions = {}) {
    try {
      const detail = await profileService.getSystemServiceCaseDetail(caseId)
      this.setData(normalizeRemoteCaseDetail(detail, iconOptions))
    } catch (error) {
      console.warn('get service case detail failed', error)
    }
  }
})
