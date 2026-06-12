const request = require('../request')

function saveCurrentLocation(data) {
  return request.post('/api/app/locations/current', data)
}

function saveManualLocation(data) {
  return request.post('/api/app/locations/manual', data)
}

function getNearbyGames(params) {
  return request.get('/api/app/games/nearby', params)
}

module.exports = {
  saveCurrentLocation,
  saveManualLocation,
  getNearbyGames
}
