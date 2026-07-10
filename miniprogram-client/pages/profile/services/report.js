const request = require('../../../api/request')

function unwrap(result, fallbackMessage) {
  if (result.code !== 0) {
    throw new Error(result.message || fallbackMessage)
  }

  return result.data
}

async function getReportConfig() {
  return unwrap(await request.get('/api/app/reports/config'), '获取举报配置失败')
}

async function createReport(data) {
  return unwrap(await request.post('/api/app/reports', data), '提交举报失败')
}

async function getMyReports(params) {
  return unwrap(await request.get('/api/app/reports/my', params), '获取举报记录失败')
}

async function getReportDetail(reportId) {
  return unwrap(await request.get(`/api/app/reports/${reportId}`), '获取举报详情失败')
}

async function getMyAppeals(params) {
  return unwrap(await request.get('/api/app/reports/appeals/my', params), '获取申诉记录失败')
}

async function submitAppeal(reportId, data) {
  return unwrap(await request.post(`/api/app/reports/${reportId}/appeal`, data), '提交申诉失败')
}

async function submitCreditAppeal(data) {
  return unwrap(await request.post('/api/app/profile/credit-appeals', data), '提交信用申诉失败')
}

async function withdrawAppeal(reportId) {
  return unwrap(await request.post(`/api/app/reports/${reportId}/appeal/withdraw`, {}), '撤回申诉失败')
}

module.exports = {
  getReportConfig,
  createReport,
  getMyReports,
  getReportDetail,
  getMyAppeals,
  submitAppeal,
  submitCreditAppeal,
  withdrawAppeal
}
