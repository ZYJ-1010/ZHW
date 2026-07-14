const request = require('../request')

function getHome(params) {
  return request.get('/api/app/home', params || {})
}

module.exports = {
  getHome
}
