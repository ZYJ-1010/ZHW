const request = require('../request')

function getMessageCenter(params) {
  return request.get('/api/app/messages/center', params)
}

function getTradeWarningDetail(params) {
  return request.get('/api/app/messages/trade-warning', params)
}

function getSystemNotificationDetail(params) {
  return request.get('/api/app/messages/system-notification', params)
}

module.exports = {
  getMessageCenter,
  getTradeWarningDetail,
  getSystemNotificationDetail
}
