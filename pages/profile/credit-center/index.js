const toast = require('../../../utils/toast')
const profileService = require('../../../services/profile')

Page({
  data: {
    scoreCard: {
      label: '',
      score: '',
      badges: []
    },
    summary: [],
    records: [],
    bottomNote: ''
  },

  onLoad() {
    this.loadCredit()
  },

  async loadCredit() {
    try {
      const data = await profileService.getProfileCredit()
      const credit = data || {}
      const scoreSource = credit.scoreCard || credit.credit || credit.creditSummary || {}

      this.setData({
        scoreCard: this.normalizeScoreCard(scoreSource, credit),
        summary: this.normalizeList(credit.summary || credit.stats || credit.items),
        records: this.normalizeList(credit.records || credit.list),
        bottomNote: this.pickText(credit.bottomNote, credit.noteText, credit.noticeText, credit.rulesNote)
      })
    } catch (error) {
      this.setData({
        scoreCard: {
          label: '',
          score: '',
          badges: []
        },
        summary: [],
        records: [],
        bottomNote: ''
      })
      toast.info(error.message || '信用中心加载失败')
    }
  },

  normalizeScoreCard(card = {}, credit = {}) {
    return {
      label: this.pickText(card.label, card.title, card.scoreLabel, credit.scoreLabel),
      score: this.pickText(card.score, card.creditScore, card.value, credit.score, credit.creditScore),
      badges: this.normalizeBadges(card.badges || card.tags || credit.badges || credit.scoreBadges)
    }
  },

  normalizeBadges(list) {
    return this.normalizeList(list).map((item) => {
      if (typeof item === 'string') {
        return {
          text: item,
          tone: ''
        }
      }

      return {
        text: item.text || item.label || item.name || '',
        tone: item.tone || item.type || ''
      }
    }).filter((item) => item.text)
  },

  normalizeList(list) {
    return Array.isArray(list) ? list : []
  },

  pickText(...values) {
    const value = values.find((item) => item || item === 0)

    return value || value === 0 ? String(value) : ''
  }
})
