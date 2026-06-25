const request = require('../request')

function getTradeWarningDetail(params) {
  return request.get('/api/app/messages/trade-warning', params)
}

module.exports = {
  getTradeWarningDetail
}
