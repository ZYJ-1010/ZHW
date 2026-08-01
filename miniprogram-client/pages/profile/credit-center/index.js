const toast = require('../../../utils/toast')
const profileService = require('../../../services/profile')
const { navigateShellRoute } = require('../../../utils/shell-nav')

function normalizeRecord(item = {}, index = 0) {
  const score = item.score === null || item.score === undefined || item.score === '' ? '0' : String(item.score)
  const numericScore = Number(score)
  let tone = item.tone
  if (!tone) {
    tone = numericScore > 0 ? 'plus' : (numericScore < 0 ? 'minus' : 'neutral')
  }

  return {
    id: item.id || item.creditLogId || `credit-record-${index}`,
    creditLogId: Number(item.creditLogId || item.id || 0) || 0,
    reportId: Number(item.reportId || 0) || 0,
    title: item.title || '信用变更',
    desc: item.desc || item.description || '信用账户变更记录',
    score,
    tone,
    canAppeal: item.canAppeal === true,
    appealId: Number(item.appealId || 0) || 0,
    appealStatusText: item.appealStatusText || '',
    appealRoute: item.appealRoute || ''
  }
}

function normalizeAppealEntry(entry = {}, fallback = {}) {
  const enabled = entry.enabled === true
  const count = Math.max(0, Number(entry.count || 0) || 0)
  const hasRoute = Object.prototype.hasOwnProperty.call(entry, 'route')

  return {
    enabled,
    count,
    text: entry.text || (enabled ? `信用申诉（${count}）` : '暂无可申诉记录'),
    route: hasRoute ? (entry.route || '') : (fallback.route || '')
  }
}

function creditStatusTone(status) {
  const value = String(status || '').trim().toLowerCase()
  if (value === 'frozen') {
    return 'red'
  }
  if (value === 'restricted_join' || value === 'restricted_create') {
    return 'orange'
  }
  return 'green'
}

Page({
  data: {
    loading: false,
    scoreLabel: '信用分',
    score: 100,
    level: '正常',
    statusTone: 'green',
    monthlyDelta: '+0',
    accountTip: '永久信用账户，不按天重置',
    bottomNote: '信用为永久账户。低于60分不能发局，低于40分不能报名或接受邀请，低于20分仅可查看和申诉。',
    appealEntry: {
      enabled: false,
      text: '信用申诉',
      route: '/pages/profile/system-management/credit-appeal/index'
    },
    summary: [
      { label: '当前状态', value: '正常' },
      { label: '正向记录', value: '0条' },
      { label: '扣分记录', value: '0条' }
    ],
    records: []
  },

  onShow() {
    if (this.data.loading) {
      return
    }
    this.loadCreditCenter()
  },

  async loadCreditCenter() {
    this.setData({ loading: true })

    try {
      const data = await profileService.getCreditCenter()
      this.setData({
        // 0 分是有效信用分，不能被逻辑或误显示为初始 100 分。
        score: data.score === null || data.score === undefined ? 100 : data.score,
        scoreLabel: data.scoreLabel || '信用分',
        level: data.level || '正常',
        statusTone: creditStatusTone(data.status),
        monthlyDelta: data.monthlyDelta === null || data.monthlyDelta === undefined ? '+0' : String(data.monthlyDelta),
        summary: Array.isArray(data.summary) && data.summary.length ? data.summary : this.data.summary,
        records: Array.isArray(data.records) ? data.records.map(normalizeRecord) : [],
        bottomNote: data.bottomNote || this.data.bottomNote,
        accountTip: data.isPermanent ? '永久信用账户，不按天重置' : this.data.accountTip,
        appealEntry: normalizeAppealEntry(data.appealEntry, this.data.appealEntry)
      })
    } catch (error) {
      toast.info(error.message || '获取信用中心失败')
    } finally {
      this.setData({ loading: false })
    }
  },

  handleRecordTap(event) {
    const { reportId, creditLogId, appealRoute, canAppeal } = event.currentTarget.dataset

    if (appealRoute) {
      navigateShellRoute(appealRoute)
      return
    }

    if (reportId) {
      navigateShellRoute(`/pages/profile/system-management/credit-appeal/index?reportId=${reportId}`)
      return
    }

    if (creditLogId && (canAppeal === true || canAppeal === 'true')) {
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

    if (!this.data.appealEntry.route) {
      toast.info('请点击需要申诉的扣分记录')
      return
    }

    navigateShellRoute(this.data.appealEntry.route)
  }
})
