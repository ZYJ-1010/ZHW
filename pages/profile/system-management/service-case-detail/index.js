const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const ASSET_BASE = '/pages/profile/system-management/service-case-detail/assets'

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

function normalizeCaseDetail(detail = {}, iconOptions = {}) {
  const caseInfo = detail.caseInfo || detail.info || {}
  const rating = detail.rating || {}
  const rawScore = rating.score || caseInfo.score
  const score = Number(rawScore)
  const starCount = rawScore !== undefined && rawScore !== '' && Number.isFinite(score)
    ? Math.max(0, Math.min(5, Math.round(score)))
    : 0

  return {
    caseInfo: {
      ...caseInfo,
      iconText: iconOptions.iconText || caseInfo.iconText || '',
      tone: iconOptions.tone || caseInfo.tone || '',
      totalPlayers: caseInfo.totalPlayers || caseInfo.playerCount || 0,
      stars: Array.isArray(caseInfo.stars) ? caseInfo.stars : buildStars(starCount)
    },
    rating: {
      ...rating,
      score: rating.score || caseInfo.rating || '',
      tags: Array.isArray(rating.tags) ? rating.tags : [],
      stars: Array.isArray(rating.stars) ? rating.stars : buildStars(starCount)
    },
    detailSections: Array.isArray(detail.detailSections || detail.sections) ? (detail.detailSections || detail.sections) : [],
    players: Array.isArray(detail.players || detail.participants) ? (detail.players || detail.participants) : []
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
    caseId: '',
    ...normalizeCaseDetail()
  },

  onLoad(options = {}) {
    const caseId = options.caseId || options.id || ''

    this.setData({
      caseId,
      iconText: decodeOption(options.iconText),
      tone: decodeOption(options.tone)
    })

    if (caseId) {
      this.loadCaseDetail()
    }
  },

  async loadCaseDetail() {
    try {
      const data = await profileService.getSystemSkillCaseDetail({
        caseId: this.data.caseId
      })

      this.setData(normalizeCaseDetail(data.detail || data, {
        iconText: this.data.iconText,
        tone: this.data.tone
      }))
    } catch (error) {
      this.setData(normalizeCaseDetail({}, {
        iconText: this.data.iconText,
        tone: this.data.tone
      }))
      toast.info(error.message || '案例详情加载失败')
    }
  }
})
