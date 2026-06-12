const request = require('../request')

function getHome() {
  return request.get('/api/app/home')
}

module.exports = {
  getHome
}
