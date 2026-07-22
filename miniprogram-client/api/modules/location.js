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

function getLocationFallback() {
  return request.get('/api/app/locations/fallback')
}

function searchMapPlaces(params) {
  return request.get('/api/app/map/search', params)
}

function reverseGeocode(params) {
  return request.get('/api/app/map/reverse-geocode', params)
}

module.exports = {
  saveCurrentLocation,
  saveManualLocation,
  getLocationFallback,
  getNearbyGames,
  searchMapPlaces,
  reverseGeocode
}
