const request = require('../request')

function getGames(params) {
  return request.get('/api/app/games', params)
}

function getGameDetail(gameId) {
  return request.get(`/api/app/games/${gameId}`)
}

function createGame(data) {
  return request.post('/api/app/games', data)
}

function applyGame(gameId, data) {
  return request.post(`/api/app/games/${gameId}/apply`, data)
}

module.exports = {
  getGames,
  getGameDetail,
  createGame,
  applyGame
}
