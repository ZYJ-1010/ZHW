const messageApi = require('../api/modules/message')

async function getTradeWarningDetail(params) {
  const result = await messageApi.getTradeWarningDetail(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取交易预警失败')
  }

  return result.data
}

async function getSystemNotificationDetail(params) {
  const result = await messageApi.getSystemNotificationDetail(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取系统通知失败')
  }

  return result.data
}

module.exports = {
  getTradeWarningDetail,
  getSystemNotificationDetail
}
