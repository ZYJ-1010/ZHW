const request = require('../request')

function createReport(data) {
  return request.post('/api/app/reports', data)
}

function getReportConfig() {
  return request.get('/api/app/reports/config')
}

function getMyReports(params) {
  return request.get('/api/app/reports/my', params)
}

function getReportDetail(reportId) {
  return request.get(`/api/app/reports/${reportId}`)
}

function getMyAppeals(params) {
  return request.get('/api/app/reports/appeals/my', params)
}

function submitAppeal(reportId, data) {
  return request.post(`/api/app/reports/${reportId}/appeal`, data)
}

function submitCreditAppeal(data) {
  return request.post('/api/app/profile/credit-appeals', data)
}

function withdrawAppeal(reportId) {
  return request.post(`/api/app/reports/${reportId}/appeal/withdraw`, {})
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
