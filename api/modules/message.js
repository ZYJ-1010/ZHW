const request = require('../request')

function getMessageCenter(params) {
  return request.get('/api/app/notifications', params)
}

function getMessageMyConfig() {
  return request.get('/api/app/messages/my-config')
}

function getTradeWarningDetail(params) {
  return request.get('/api/app/messages/trade-warning', params || {})
}

function handleTradeWarning(data) {
  return request.post('/api/app/messages/trade-warning/actions', data)
}

function getSystemNotificationDetail(params) {
  return request.get('/api/app/messages/system-notification', params || {})
}

function submitSystemNotificationFeedback(data) {
  return request.post('/api/app/messages/system-notification/feedback', data)
}

function handleNotificationAction(notificationId, data) {
  return request.post(`/api/app/notifications/${notificationId}/actions`, data)
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
