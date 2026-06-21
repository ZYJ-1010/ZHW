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

function getInvitePlayerConfig(params) {
  return request.get('/api/app/game-invites/player-config', params)
}

function getInviteRecentPlayers(params) {
  return request.get('/api/app/game-invites/recent-players', params)
}

function getInvitePlayers(params) {
  return request.get('/api/app/game-invites/players', params)
}

module.exports = {
  getGames,
  getGameDetail,
  createGame,
  applyGame,
  getInvitePlayerConfig,
  getInviteRecentPlayers,
  getInvitePlayers
}
