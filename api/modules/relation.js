const request = require('../request')

function getNetworkHome(params) {
  return request.get('/api/app/relations/network-home', params || {})
}

module.exports = {
  getNetworkHome
}
