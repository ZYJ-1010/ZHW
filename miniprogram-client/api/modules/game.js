const request = require('../request')

function idempotencyOptions(key) {
  const normalized = String(key || '').trim()
  return normalized ? { header: { 'Idempotency-Key': normalized } } : {}
}

function getGames(params) {
  return request.get('/api/app/games', params)
}

function getSameCityGames(params) {
  return request.get('/api/app/games/city', params)
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

function createProgressFeedback(gameId, data) {
  return request.post(`/api/app/games/${gameId}/progress-feedbacks`, data)
}

function getMilestones(gameId) {
  return request.get(`/api/app/games/${gameId}/milestones`)
}

function createMilestone(gameId, data) {
  return request.post(`/api/app/games/${gameId}/milestones`, data)
}

function updateMilestone(gameId, milestoneId, data) {
  return request.put(`/api/app/games/${gameId}/milestones/${milestoneId}`, data)
}

function getGameCheckins(gameId) {
  return request.get(`/api/app/games/${gameId}/checkins`)
}

function createGameCheckin(gameId, data) {
  return request.post(`/api/app/games/${gameId}/checkins`, data)
}

function requestGameCompletion(gameId) {
  return request.post(`/api/app/games/${gameId}/completion-request`, {})
}

function createGuideFollowUp(gameId, data) {
  return request.post(`/api/app/games/${gameId}/guide-follow-ups`, data)
}

function createGame(data, idempotencyKey) {
  return request.post('/api/app/games', data, idempotencyOptions(idempotencyKey))
}

function getGameDrafts() {
  return request.get('/api/app/game-drafts')
}

function getGameDraft(draftId) {
  return request.get(`/api/app/game-drafts/${draftId}`)
}

function createGameDraft(data, idempotencyKey) {
  return request.post('/api/app/game-drafts', data, idempotencyOptions(idempotencyKey))
}

function updateGameDraft(draftId, data, idempotencyKey) {
  return request.put(`/api/app/game-drafts/${draftId}`, data, idempotencyOptions(idempotencyKey))
}

function deleteGameDraft(draftId, idempotencyKey) {
  return request.del(`/api/app/game-drafts/${draftId}`, undefined, idempotencyOptions(idempotencyKey))
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

function cancelGameApplication(applicationId) {
  return request.post(`/api/app/game-applications/${applicationId}/cancel`, {})
}

function exitGame(gameId) {
  return request.post(`/api/app/games/${gameId}/exit`, {})
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
  getSameCityGames,
  getGameDetail,
  getGameMembers,
  getGameSuccessDetail,
  getGameGuideSuccessDetail,
  getGameCollaboration,
  createProgressFeedback,
  getMilestones,
  createMilestone,
  updateMilestone,
  getGameCheckins,
  createGameCheckin,
  requestGameCompletion,
  createGuideFollowUp,
  createGame,
  getGameDrafts,
  getGameDraft,
  createGameDraft,
  updateGameDraft,
  deleteGameDraft,
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
  cancelGameApplication,
  exitGame,
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
