const request = require('../request')

function getNetworkHome(params) {
  return request.get('/api/app/connections/my', params || {})
}

module.exports = {
  getNetworkHome
}
