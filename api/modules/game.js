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

function respondGameInvitation(invitationId, data) {
  return request.post(`/api/app/game-invitations/${invitationId}/respond`, data)
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

function getReplayConfirmContext(params) {
  return request.get('/api/app/game-invites/replay-context', params)
}

function getSystemRecommendations(params) {
  return request.get('/api/app/game-invites/system-recommendations', params)
}

function createReplayInvitation(data) {
  return request.post('/api/app/game-invites/replay', data)
}

function getGuideProgress(params) {
  return request.get('/api/app/game-invites/guide-progress', params)
}

function getGuideCancelDetail(params) {
  return request.get('/api/app/game-invites/guide-cancel-detail', params)
}

function getGameManage(params) {
  return request.get('/api/app/games/my/manage', params)
}

function getPlayerGameManage(params) {
  return request.get('/api/app/games/player/manage', params)
}

function createGamePayment(data) {
  return request.post('/api/app/game-payments/wechat', data)
}

module.exports = {
  getGames,
  getGameDetail,
  createGame,
  applyGame,
  respondGameInvitation,
  getInvitePlayerConfig,
  getInviteRecentPlayers,
  getInvitePlayers,
  getReplayConfirmContext,
  getSystemRecommendations,
  createReplayInvitation,
  getGuideProgress,
  getGuideCancelDetail,
  getGameManage,
  getPlayerGameManage,
  createGamePayment
}
