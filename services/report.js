const reportApi = require('../api/modules/report')

async function getReportConfig() {
  const result = await reportApi.getReportConfig()

  if (result.code !== 0) {
    throw new Error(result.message || '获取举报配置失败')
  }

  return result.data
}

async function createReport(data) {
  const result = await reportApi.createReport(data)

  if (result.code !== 0) {
    throw new Error(result.message || '提交举报失败')
  }

  return result.data
}

async function getMyReports(params) {
  const result = await reportApi.getMyReports(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取举报记录失败')
  }

  return result.data
}

async function getReportDetail(reportId) {
  const result = await reportApi.getReportDetail(reportId)

  if (result.code !== 0) {
    throw new Error(result.message || '获取举报详情失败')
  }

  return result.data
}

async function getMyAppeals(params) {
  const result = await reportApi.getMyAppeals(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取申诉记录失败')
  }

  return result.data
}

async function submitAppeal(reportId, data) {
  const result = await reportApi.submitAppeal(reportId, data)

  if (result.code !== 0) {
    throw new Error(result.message || '提交申诉失败')
  }

  return result.data
}

async function submitCreditAppeal(data) {
  const result = await reportApi.submitCreditAppeal(data)

  if (result.code !== 0) {
    throw new Error(result.message || '提交信用申诉失败')
  }

  return result.data
}

async function withdrawAppeal(reportId) {
  const result = await reportApi.withdrawAppeal(reportId)

  if (result.code !== 0) {
    throw new Error(result.message || '撤回申诉失败')
  }

  return result.data
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
