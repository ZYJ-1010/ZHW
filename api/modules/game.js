const request = require('../request')

function getGames(params) {
  return request.get('/api/app/games', params)
}

function getGameDetail(gameId) {
  return request.get(`/api/app/games/${gameId}`)
}

function getGameMembers(gameId, params) {
  return request.get(`/api/app/games/${gameId}/members`, params)
}

function createGame(data) {
  return request.post('/api/app/games', data)
}

function getProfitTemplates(params) {
  return request.get('/api/app/games/profit-templates', params)
}

function applyGame(gameId, data) {
  return request.post(`/api/app/games/${gameId}/apply`, data)
}

function getGameApplyConfig(gameId, params) {
  return request.get(`/api/app/games/${gameId}/apply-config`, params)
}

function respondGameInvitation(invitationId, data) {
  return request.post(`/api/app/game-invitations/${invitationId}/respond`, data)
}

function getInvitePlayerConfig(params) {
  return request.get('/api/app/game-invites/player-config', params)
}

function getGameInviteConfig(params) {
  return request.get('/api/app/game-invites/config', params)
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

function createGameInvite(data) {
  return request.post('/api/app/game-invites', data)
}

function getGuideProgress(params) {
  return request.get('/api/app/game-invites/guide-progress', params)
}

function getGuideSuccess(params) {
  return request.get('/api/app/game-invites/guide-success', params)
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

function getGameCollaboration(gameId, params) {
  return request.get(`/api/app/games/${gameId}/collaboration`, params)
}

function endGameCollaboration(gameId, data) {
  return request.post(`/api/app/games/${gameId}/end`, data)
}

function getGameAudits(params) {
  return request.get('/api/app/game-audits', params)
}

function getGameAuditDetail(auditId, params) {
  return request.get(`/api/app/game-audits/${auditId}`, params)
}

function respondGameAudit(auditId, data) {
  return request.post(`/api/app/game-audits/${auditId}/respond`, data)
}

function batchRespondGameAudits(data) {
  return request.post('/api/app/game-audits/batch-respond', data)
}

function getPlayAgainOptions(params) {
  return request.get('/api/app/game-replays/options', params)
}

function selectPlayAgainOption(data) {
  return request.post('/api/app/game-replays/options/select', data)
}

function getExpertSuccess(params) {
  return request.get('/api/app/game-invites/expert-success', params)
}

function getReferralRecords(params) {
  return request.get('/api/app/game-referrals/records', params)
}

function triggerReferralRecordAction(data) {
  return request.post('/api/app/game-referrals/actions', data)
}

function getGameReviewConfig(params) {
  return request.get('/api/app/game-reviews/config', params)
}

function submitGameReview(data) {
  return request.post('/api/app/game-reviews', data)
}

function getReviewCompleteConfig(params) {
  return request.get('/api/app/game-reviews/complete-config', params)
}

function selectReviewCompleteIntent(data) {
  return request.post('/api/app/game-reviews/complete-intent', data)
}

function getServiceDeliveryDetail(serviceOrderId, params) {
  return request.get(`/api/app/game-services/${serviceOrderId}/delivery-detail`, params)
}

function remindPlayerConfirm(serviceOrderId, data) {
  return request.post(`/api/app/game-services/${serviceOrderId}/remind-player-confirm`, data)
}

function confirmServiceDelivery(serviceOrderId, data) {
  return request.post(`/api/app/game-services/${serviceOrderId}/confirm-delivery`, data)
}

function getExpertCancelPreview(serviceOrderId, params) {
  return request.get(`/api/app/game-services/${serviceOrderId}/expert-cancel-preview`, params)
}

function cancelServiceWithCompensation(serviceOrderId, data) {
  return request.post(`/api/app/game-services/${serviceOrderId}/cancel-with-compensation`, data)
}

function getGuideChatContext(params) {
  return request.get('/api/app/game-invites/guide-chat-context', params)
}

function respondGuideChatInvitation(data) {
  return request.post('/api/app/game-invites/guide-chat/respond', data)
}

function sendGuideChatMessage(data) {
  return request.post('/api/app/game-invites/guide-chat/messages', data)
}

function getGameGreetingContext(params) {
  return request.get('/api/app/game-greetings/context', params)
}

function sendGameGreetingMessage(data) {
  return request.post('/api/app/game-greetings/messages', data)
}

function getHallGreetingContext(params) {
  return request.get('/api/app/game-hall/greeting-context', params)
}

function sendHallGreetingMessage(data) {
  return request.post('/api/app/game-hall/greetings/messages', data)
}

function triggerHallGreetingAction(data) {
  return request.post('/api/app/game-hall/greetings/actions', data)
}

function getGamePaymentConfig(gameId, params) {
  if (gameId) {
    return request.get(`/api/app/games/${gameId}/payment-config`, params)
  }

  return request.get('/api/app/game-payments/config', params)
}

function createGamePayment(data) {
  return request.post('/api/app/game-payments/wechat', data)
}

module.exports = {
  getGames,
  getGameDetail,
  getGameMembers,
  createGame,
  getProfitTemplates,
  applyGame,
  getGameApplyConfig,
  respondGameInvitation,
  getInvitePlayerConfig,
  getGameInviteConfig,
  getInviteRecentPlayers,
  getInvitePlayers,
  getReplayConfirmContext,
  getSystemRecommendations,
  createReplayInvitation,
  createGameInvite,
  getGuideProgress,
  getGuideSuccess,
  getGuideCancelDetail,
  getGameManage,
  getPlayerGameManage,
  getGameCollaboration,
  endGameCollaboration,
  getGameAudits,
  getGameAuditDetail,
  respondGameAudit,
  batchRespondGameAudits,
  getPlayAgainOptions,
  selectPlayAgainOption,
  getExpertSuccess,
  getReferralRecords,
  triggerReferralRecordAction,
  getGameReviewConfig,
  submitGameReview,
  getReviewCompleteConfig,
  selectReviewCompleteIntent,
  getServiceDeliveryDetail,
  remindPlayerConfirm,
  confirmServiceDelivery,
  getExpertCancelPreview,
  cancelServiceWithCompensation,
  getGuideChatContext,
  respondGuideChatInvitation,
  sendGuideChatMessage,
  getGameGreetingContext,
  sendGameGreetingMessage,
  getHallGreetingContext,
  sendHallGreetingMessage,
  triggerHallGreetingAction,
  getGamePaymentConfig,
  createGamePayment
}
