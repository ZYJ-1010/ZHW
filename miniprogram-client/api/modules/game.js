const request = require('../request')

function getGames(params) {
  return request.get('/api/app/games', params)
}

function getGameDetail(gameId) {
  return request.get(`/api/app/games/${gameId}`)
}

function getGameMembers(gameId) {
  return request.get(`/api/app/games/${gameId}/members`)
}

function getGameSuccessDetail(gameId, params) {
  return request.get(`/api/app/games/${gameId}/success-detail`, params)
}

function getGameGuideSuccessDetail(gameId, params) {
  return request.get(`/api/app/games/${gameId}/guide-success-detail`, params)
}

function getGameCollaboration(gameId, params) {
  return request.get(`/api/app/games/${gameId}/collaboration`, params)
}

function requestGameCompletion(gameId) {
  return request.post(`/api/app/games/${gameId}/completion-request`, {})
}

function createGuideFollowUp(gameId, data) {
  return request.post(`/api/app/games/${gameId}/guide-follow-ups`, data)
}

function createGame(data) {
  return request.post('/api/app/games', data)
}

function startGame(gameId) {
  return request.post(`/api/app/games/${gameId}/manual-start`, {})
}

function createInviteEntry(data) {
  return request.post('/api/app/invites/entries', data)
}

function confirmService(gameId, data) {
  return request.post(`/api/app/games/${gameId}/service-confirm`, data)
}

function createRetrospective(gameId, data) {
  return request.post(`/api/app/games/${gameId}/retrospectives`, data)
}

function requestPlayerCancel(gameId, data) {
  return request.post(`/api/app/games/${gameId}/player-cancel`, data)
}

function getPlayerCancelDetail(gameId) {
  return request.get(`/api/app/games/${gameId}/player-cancel-detail`)
}

function requestExpertCancel(gameId, data) {
  return request.post(`/api/app/games/${gameId}/expert-cancel`, data)
}

function getExpertCancelDetail(gameId) {
  return request.get(`/api/app/games/${gameId}/expert-cancel-detail`)
}

function favoriteGame(gameId) {
  return request.post(`/api/app/games/${gameId}/favorite`, {})
}

function getProfitTemplates(params) {
  return request.get('/api/app/games/profit-templates', params)
}

function getCategoryConfig(params) {
  return request.get('/api/app/games/category-config', params)
}

function getApplicationConfig(params) {
  return request.get('/api/app/games/application-config', params)
}

function getConditionRuleConfig(params) {
  return request.get('/api/app/games/condition-rule-config', params)
}

function getCancelConfig(params) {
  return request.get('/api/app/games/cancel-config', params)
}

function applyGame(gameId, data) {
  return request.post(`/api/app/games/${gameId}/applications`, data)
}

function getReceivedApplications(params) {
  return request.get('/api/app/game-applications/received', params)
}

function reviewGameApplication(applicationId, data) {
  return request.post(`/api/app/game-applications/${applicationId}/audit`, data)
}

function respondGameInvitation(invitationId, data) {
  return request.post(`/api/app/game-invitations/${invitationId}/respond`, data)
}

function getInvitePlayerConfig(params) {
  return request.get('/api/app/game-invites/player-config', params)
}

function getInvitePermission(params) {
  return request.get('/api/app/game-invites/permission', params)
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

function createCurrentGameInvitation(data) {
  return request.post('/api/app/game-invites/current', data)
}

function sendGuideReminder(data) {
  return request.post('/api/app/game-invites/reminders', data)
}

function getGuideProgress(params) {
  return request.get('/api/app/game-invites/guide-progress', params)
}

function getGuideCancelDetail(params) {
  return request.get('/api/app/game-invites/guide-cancel-detail', params)
}

function getReferralRecords(params) {
  return request.get('/api/app/game-invites/referral-records', params)
}

function getGameManage(params) {
  return request.get('/api/app/games/my/manage', params)
}

function getPlayerGameManage(params) {
  return request.get('/api/app/games/player/manage', params)
}

function getMyFavoriteGames(params) {
  return request.get('/api/app/games/favorites/my', params)
}

function createGamePayment(data) {
  return request.post('/api/app/payment/precreate-placeholder', data)
}

function getGamePaymentPreview(gameId) {
  return request.get(`/api/app/games/${gameId}/payment-preview`)
}

module.exports = {
  getGames,
  getGameDetail,
  getGameMembers,
  getGameSuccessDetail,
  getGameGuideSuccessDetail,
  getGameCollaboration,
  requestGameCompletion,
  createGuideFollowUp,
  createGame,
  startGame,
  createInviteEntry,
  confirmService,
  createRetrospective,
  requestPlayerCancel,
  getPlayerCancelDetail,
  requestExpertCancel,
  getExpertCancelDetail,
  favoriteGame,
  getProfitTemplates,
  getCategoryConfig,
  getApplicationConfig,
  getConditionRuleConfig,
  getCancelConfig,
  applyGame,
  getReceivedApplications,
  reviewGameApplication,
  respondGameInvitation,
  getInvitePlayerConfig,
  getInvitePermission,
  getInviteRecentPlayers,
  getInvitePlayers,
  getReplayConfirmContext,
  getSystemRecommendations,
  createReplayInvitation,
  createCurrentGameInvitation,
  sendGuideReminder,
  getGuideProgress,
  getGuideCancelDetail,
  getReferralRecords,
  getGameManage,
  getPlayerGameManage,
  getMyFavoriteGames,
  createGamePayment,
  getGamePaymentPreview
}
