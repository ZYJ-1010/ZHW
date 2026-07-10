const toast = require('../../../utils/toast')
const profileService = require('../../../services/profile')
const { navigateShellRoute } = require('../../../utils/shell-nav')

Page({
  data: {
    loading: false,
    scoreLabel: '信用分',
    score: 100,
    level: '优秀',
    monthlyDelta: '+0',
    bottomNote: '信用分低于80分将限制部分功能，低于60分将暂停服务资格。',
    appealEntry: {
      enabled: false,
      text: '信用申诉',
      route: '/pages/profile/system-management/credit-appeal/index'
    },
    summary: [
      { label: '信用等级', value: '优秀' },
      { label: '奖励中心', value: '0条待查看' },
      { label: '惩罚中心', value: '0条记录' }
    ],
    records: []
  },

  onLoad() {
    this.loadCreditCenter()
  },

  async loadCreditCenter() {
    this.setData({ loading: true })

    try {
      const data = await profileService.getCreditCenter()
      this.setData({
        score: data.score || 100,
        scoreLabel: data.scoreLabel || '信用分',
        level: data.level || '优秀',
        monthlyDelta: data.monthlyDelta || '+0',
        summary: Array.isArray(data.summary) && data.summary.length ? data.summary : this.data.summary,
        records: Array.isArray(data.records) ? data.records : [],
        bottomNote: data.bottomNote || this.data.bottomNote,
        appealEntry: data.appealEntry || this.data.appealEntry
      })
    } catch (error) {
      toast.info(error.message || '获取信用中心失败')
    } finally {
      this.setData({ loading: false })
    }
  },

  handleRecordTap(event) {
    const { reportId, creditLogId, appealRoute } = event.currentTarget.dataset

    if (appealRoute) {
      navigateShellRoute(appealRoute)
      return
    }

    if (reportId) {
      navigateShellRoute(`/pages/profile/system-management/credit-appeal/index?reportId=${reportId}`)
      return
    }

    if (creditLogId) {
      navigateShellRoute(`/pages/profile/system-management/credit-appeal/index?creditLogId=${creditLogId}`)
      return
    }

    toast.info('该信用记录暂无可申诉入口')
  },

  handleAppealTap() {
    if (!this.data.appealEntry.enabled) {
      toast.info('暂无可申诉的信用记录')
      return
    }

    navigateShellRoute(this.data.appealEntry.route || '/pages/profile/system-management/credit-appeal/index')
  }
})
