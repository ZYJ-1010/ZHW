const reportService = require('../../services/report')
const { navigateShellRoute } = require('../../../../utils/shell-nav')

function formatReportTime(value) {
  return value ? String(value).replace('T', ' ').replace(/:\d{2}(?:\.\d+)?Z?$/, '') : ''
}

function statusText(status) {
  const map = {
    pending: '待处理',
    assigned: '处理中',
    appealed: '申诉中',
    handled: '已处理',
    closed: '已关闭'
  }

  return map[status] || status || '待处理'
}

function outcomeText(outcome, status) {
  const map = {
    confirmed: '举报属实',
    malicious: '恶意举报',
    unconfirmed: '无法核实'
  }

  return map[outcome] || statusText(status)
}

function buildTimeline(report) {
  const items = [
    { time: formatReportTime(report.createdAt), title: '提交举报', desc: '您提交了举报申请，等待平台审核' }
  ]

  if (report.status === 'assigned') {
    items.push({ time: formatReportTime(report.createdAt), title: '平台受理', desc: '平台已分配处理人员', state: 'current' })
  }

  if (report.status === 'handled' || report.status === 'closed') {
    items.push({
      time: report.handledAt ? formatReportTime(report.handledAt) : formatReportTime(report.createdAt),
      title: '处理完成',
      desc: report.handleResult || outcomeText(report.handleOutcome, report.status),
      state: 'current'
    })
  }

  if (items.length === 1) {
    items[0].state = 'current'
  }

  return items
}

function buildResultRows(report) {
  const rows = [
    {
      label: '核实结果',
      value: outcomeText(report.handleOutcome, report.status),
      className: report.handleOutcome === 'malicious' || report.status === 'closed' ? 'text-danger' : report.status === 'handled' ? 'text-success' : 'text-warning'
    },
  ]

  if (report.creditChange) {
    rows.push({
      label: '信用变更',
      value: `用户ID ${report.creditTargetUserId || '-'} ${report.creditChange}`,
      className: report.creditChange < 0 ? 'text-danger' : 'text-success'
    })
  }

  if (report.rewardPoints) {
    rows.push({
      label: '平台奖励',
      value: `+${report.rewardPoints}积分`,
      className: 'text-success'
    })
  }

  rows.push({
    label: '处理人员',
    value: report.handlerAdminId ? `管理员ID ${report.handlerAdminId}` : '待分配',
    className: ''
  })

  return rows
}

Page({
  data: {
    icons: {
      check: '/pages/profile/system-management/report-center/assets/icon-check.svg'
    },
    statusMain: '待处理',
    statusSub: '平台已收到举报，等待审核处理',
    reportId: 0,
    basicInfo: [],
    reportReason: '',
    evidence: [],
    evidenceTitle: '证据材料 (0条)',
    resultRows: [],
    resultNote: '',
    timeline: [],
    canAppeal: false
  },

  onLoad(options = {}) {
    const reportId = Number(options.reportId || options.recordId || 0) || 0

    if (reportId) {
      this.setData({ reportId })
      this.loadReportDetail(reportId)
    }
  },

  async loadReportDetail(reportId) {
    try {
      const report = await reportService.getReportDetail(reportId)
      const evidence = [
        report.chatMessageId ? { label: '聊天证据', index: String(report.chatMessageId), tone: 'teal' } : null,
        report.fileId ? { label: '文件证据', index: String(report.fileId), tone: 'red' } : null,
        report.reviewId ? { label: '评价证据', index: String(report.reviewId), tone: 'purple' } : null
      ].filter(Boolean)
      const resultText = outcomeText(report.handleOutcome, report.status)

      this.setData({
        statusMain: resultText,
        statusSub: report.status === 'pending' ? '平台已收到举报，等待审核处理' : '平台已更新举报处理状态',
        basicInfo: [
          { label: '举报编号', value: `RP${String(report.id).padStart(8, '0')}` },
          { label: '举报类型', value: report.reportType || 'other' },
          { label: '被举报人', value: report.targetUserId ? `用户ID ${report.targetUserId}` : '未指定' },
          { label: '举报时间', value: formatReportTime(report.createdAt) },
          { label: '处理状态', value: statusText(report.status) },
          { label: '关联局ID', value: String(report.gameId || '') }
        ],
        reportReason: report.content || '',
        evidence,
        evidenceTitle: `证据材料 (${evidence.length}条)`,
        resultRows: buildResultRows(report),
        resultNote: report.handleResult || (report.status === 'pending' ? '平台已收到举报，正在等待处理。' : statusText(report.status)),
        timeline: buildTimeline(report),
        canAppeal: Boolean(report.canAppeal)
      })
    } catch (error) {
      wx.showToast({
        title: error.message || '获取举报详情失败',
        icon: 'none'
      })
    }
  },

  handleBackList() {
    navigateShellRoute('/pages/profile/system-management/report-records/index')
  },

  handleAppealTap() {
    navigateShellRoute(`/pages/profile/system-management/credit-appeal/index${this.data.reportId ? `?reportId=${this.data.reportId}` : ''}`)
  }
})
