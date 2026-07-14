const toast = require('../../../../utils/toast')
const reportService = require('../../services/report')
const { navigateShellRoute } = require('../../../../utils/shell-nav')

function formatTime(value) {
  return value ? String(value).replace('T', ' ').replace(/:\d{2}(?:\.\d+)?Z?$/, '') : ''
}

Page({
  data: {
    basicInfo: [],
    appealReason: '',
    evidence: [],
    originalInfo: [],
    originalReason: '',
    timeline: [],
    reportId: 0,
    canWithdraw: false,
    loadError: ''
  },

  onLoad(options = {}) {
    const reportId = Number(options.reportId || options.appealId || 0) || 0

    if (reportId) {
      this.loadAppealDetail(reportId)
      return
    }

    this.setData({ loadError: '缺少申诉记录' })
  },

  async loadAppealDetail(reportId) {
    try {
      const report = await reportService.getReportDetail(reportId)
      const canWithdraw = report.status === 'appealed'

      this.setData({
        reportId,
        canWithdraw,
        basicInfo: [
          { label: '申诉编号', value: `AP${String(report.id).padStart(8, '0')}` },
          { label: '被举报类型', value: report.reportType || 'other' },
          { label: '举报人', value: `用户ID ${report.reporterUserId}` },
          { label: '申诉时间', value: formatTime(report.handledAt || report.createdAt) },
          { label: '当前状态', value: this.statusText(report.status), className: 'appeal-status-value' },
          { label: '处理结果', value: this.outcomeText(report.handleOutcome) }
        ],
        appealReason: report.handleResult || '',
        originalInfo: [
          { label: '举报编号', value: `RP${String(report.id).padStart(8, '0')}` },
          { label: '举报时间', value: formatTime(report.createdAt) }
        ],
        originalReason: report.content || '',
        timeline: [
          { time: formatTime(report.createdAt), title: '收到举报', desc: '平台收到举报申请' },
          { time: formatTime(report.handledAt || report.createdAt), title: '提交申诉', desc: '您提交了申诉申请及相关证据材料' },
          { time: canWithdraw ? '待定' : formatTime(report.handledAt || report.createdAt), title: canWithdraw ? '申诉审核中' : this.statusText(report.status), desc: canWithdraw ? '平台正在审核您的申诉材料' : '当前申诉状态已更新', state: 'current' }
        ],
        loadError: ''
      })
    } catch (error) {
      this.setData({
        loadError: error.message || '获取申诉详情失败',
        basicInfo: [],
        appealReason: '',
        evidence: [],
        originalInfo: [],
        originalReason: '',
        timeline: []
      })
      toast.info(error.message || '获取申诉详情失败')
    }
  },

  handleBackList() {
    navigateShellRoute('/pages/profile/system-management/report-appeals/index')
  },

  statusText(status) {
    if (status === 'appealed') {
      return '处理中'
    }
    if (status === 'handled') {
      return '已处理'
    }
    if (status === 'appeal_withdrawn') {
      return '已撤回'
    }
    return status || '未知'
  },

  outcomeText(outcome) {
    const map = {
      appeal_approved: '申诉通过',
      appeal_rejected: '申诉驳回',
      confirmed: '举报属实',
      malicious: '恶意举报',
      unconfirmed: '无法核实',
      processing: '处理中'
    }

    return map[outcome] || outcome || '待处理'
  },

  async handleWithdrawTap() {
    if (!this.data.canWithdraw || !this.data.reportId) {
      toast.info('当前申诉状态不可撤回')
      return
    }
    try {
      await reportService.withdrawAppeal(this.data.reportId)
      toast.success('申诉已撤回')
      navigateShellRoute('/pages/profile/system-management/report-appeals/index')
    } catch (error) {
      toast.info(error.message || '撤回申诉失败')
    }
  }
})
