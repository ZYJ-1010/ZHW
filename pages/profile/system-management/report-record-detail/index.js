const reportService = require('../../../../services/report')
const toast = require('../../../../utils/toast')
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

function outcomeClass(report) {
  if (report.handleOutcome === 'malicious' || report.status === 'closed') {
    return 'text-danger'
  }

  if (report.handleOutcome === 'confirmed' || report.status === 'handled') {
    return 'text-success'
  }

  return 'text-warning'
}

function buildEvidence(report) {
  return [
    report.chatMessageId ? { label: '聊天证据', index: String(report.chatMessageId), tone: 'teal' } : null,
    report.fileId ? { label: '文件证据', index: String(report.fileId), tone: 'red' } : null,
    report.reviewId ? { label: '评价证据', index: String(report.reviewId), tone: 'purple' } : null,
    report.revenueRecordId ? { label: '分账证据', index: String(report.revenueRecordId), tone: 'purple' } : null
  ].filter(Boolean)
}

function buildTimeline(report) {
  const items = [
    { time: formatReportTime(report.createdAt), title: '提交举报', desc: '您提交了举报申请，等待平台审核' }
  ]

  if (report.status === 'assigned' || report.status === 'appealed') {
    items.push({
      time: formatReportTime(report.handledAt || report.createdAt),
      title: report.status === 'appealed' ? '申诉处理中' : '平台受理',
      desc: report.status === 'appealed' ? '相关用户已提交申诉，平台将重新核实' : '平台已分配处理人员',
      state: 'current'
    })
  }

  if (report.status === 'handled' || report.status === 'closed') {
    items.push({
      time: formatReportTime(report.handledAt || report.createdAt),
      title: report.status === 'closed' ? '处理关闭' : '处理完成',
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
      className: outcomeClass(report)
    },
    {
      label: '资金冻结',
      value: report.revenueFrozen ? '已冻结相关分账' : '未冻结',
      className: report.revenueFrozen ? 'text-danger' : ''
    }
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
      className: 'text-success',
      reward: true
    })
  }

  return rows
}

Page({
  data: {
    icons: {
      check: '/pages/profile/system-management/report-center/assets/icon-check.svg'
    },
    reportId: 0,
    statusMain: '待处理',
    statusSub: '平台已收到举报，等待审核处理',
    rewardValue: '+0积分',
    rewardDesc: '处理完成后如有奖励会在此展示',
    basicInfo: [],
    reportReason: '',
    evidence: [],
    evidenceTitle: '证据材料 (0条)',
    resultRows: [],
    resultNote: '',
    timeline: []
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
      const evidence = buildEvidence(report)
      const resultRows = buildResultRows(report)
      const resultText = outcomeText(report.handleOutcome, report.status)

      this.setData({
        statusMain: resultText,
        statusSub: report.status === 'pending' ? '平台已收到举报，等待审核处理' : '平台已更新举报处理状态',
        rewardValue: report.rewardPoints ? `+${report.rewardPoints}积分` : '+0积分',
        rewardDesc: report.rewardPoints ? '已发放至您的积分账户' : '处理完成后如有奖励会在此展示',
        basicInfo: [
          { label: '举报编号', value: `RP${String(report.id).padStart(8, '0')}` },
          { label: '举报类型', value: report.reportType || 'other' },
          { label: '被举报人', value: report.targetUserId ? `用户ID ${report.targetUserId}` : '未指定' },
          { label: '举报时间', value: formatReportTime(report.createdAt) },
          { label: '完成时间', value: report.handledAt ? formatReportTime(report.handledAt) : '待处理' },
          { label: '处理人', value: report.handlerAdminId ? `管理员ID ${report.handlerAdminId}` : '待分配' }
        ],
        reportReason: report.content || '',
        evidence,
        evidenceTitle: `证据材料 (${evidence.length}条)`,
        resultRows,
        resultNote: report.handleResult || (report.status === 'pending' ? '平台已收到举报，正在等待处理。' : statusText(report.status)),
        timeline: buildTimeline(report)
      })
    } catch (error) {
      wx.showToast({
        title: error.message || '获取处理记录失败',
        icon: 'none'
      })
    }
  },

  handleBackList() {
    navigateShellRoute('/pages/profile/system-management/report-records/index')
  },

  handleRateTap() {
    if (!this.data.reportId) {
      toast.info('暂无可评价的处理记录')
      return
    }

    navigateShellRoute(`/pages/profile/system-management/feedback/index?bizType=report&bizId=${encodeURIComponent(this.data.reportId)}`)
  }
})
