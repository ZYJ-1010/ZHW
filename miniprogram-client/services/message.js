const messageApi = require('../api/modules/message')

async function getMessageCenter(params) {
  const result = await messageApi.getMessageCenter(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取消息中心失败')
  }

  return result.data
}

async function getMessageMyConfig() {
  const result = await messageApi.getMessageMyConfig()

  if (result.code !== 0) {
    throw new Error(result.message || '获取我的消息失败')
  }

  return result.data
}

async function getTradeWarningDetail(params) {
  const result = await messageApi.getTradeWarningDetail(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取交易提醒失败')
  }

  return result.data
}

async function handleTradeWarning(payload = {}) {
  const action = String(payload.action || '').trim()
  const warningId = String(payload.warningId || payload.id || '').trim()
  const orderId = String(payload.orderId || '').trim()

  if (action !== 'delay' && action !== 'deliver') {
    throw new Error('交易提醒操作无效')
  }

  const result = await messageApi.handleTradeWarning({
    action,
    warningId,
    orderId,
    gameId: payload.gameId || 0,
    deliveryMethod: payload.deliveryMethod || ''
  })

  if (result.code !== 0) {
    throw new Error(result.message || '处理交易提醒失败')
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

async function submitSystemNotificationFeedback(payload = {}) {
  const messageId = String(payload.messageId || payload.notificationId || payload.id || '').trim()
  const value = String(payload.value || payload.feedback || '').trim()

  if (!value) {
    throw new Error('请选择反馈结果')
  }

  const result = await messageApi.submitSystemNotificationFeedback({
    messageId,
    value
  })

  if (result.code !== 0) {
    throw new Error(result.message || '提交反馈失败')
  }

  return result.data
}

async function handleNotificationAction(payload = {}) {
  const notificationId = String(payload.notificationId || payload.id || '').trim()
  const action = String(payload.action || '').trim()

  if (!notificationId || !action) {
    throw new Error('缺少消息操作信息')
  }

  const result = await messageApi.handleNotificationAction(encodeURIComponent(notificationId), {
    action,
    reason: payload.reason || ''
  })

  if (result.code !== 0) {
    throw new Error(result.message || '处理消息失败')
  }

  return result.data
}

module.exports = {
  getMessageCenter,
  getMessageMyConfig,
  getTradeWarningDetail,
  handleTradeWarning,
  getSystemNotificationDetail,
  submitSystemNotificationFeedback,
  handleNotificationAction
}
