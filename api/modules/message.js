const request = require('../request')

function getTradeWarningDetail(params) {
  return request.get('/api/app/messages/trade-warning', params)
}

function getSystemNotificationDetail(params) {
  return request.get('/api/app/messages/system-notification', params)
}

module.exports = {
  getTradeWarningDetail,
  getSystemNotificationDetail
}
