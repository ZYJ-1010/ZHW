const gameApi = require('../api/modules/game')

async function getGameList(params) {
  const result = await gameApi.getGames(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取局列表失败')
  }

  return result.data
}

async function getGameDetail(gameId) {
  if (!gameId) {
    throw new Error('缺少局信息')
  }

  const result = await gameApi.getGameDetail(gameId)

  if (result.code !== 0) {
    throw new Error(result.message || '获取局详情失败')
  }

  return result.data
}

async function getGameMembers(gameId, params) {
  if (!gameId) {
    throw new Error('缺少局信息')
  }

  const result = await gameApi.getGameMembers(gameId, params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取局成员失败')
  }

  return result.data
}

async function getInvitePlayerConfig(params) {
  const result = await gameApi.getInvitePlayerConfig(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取邀请玩家配置失败')
  }

  return result.data
}

async function getGameInviteConfig(params = {}) {
  const result = await gameApi.getGameInviteConfig(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取邀请配置失败')
  }

  return result.data
}

async function getInviteRecentPlayers(params) {
  const result = await gameApi.getInviteRecentPlayers(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取最近联系玩家失败')
  }

  return result.data
}

async function getInvitePlayers(params) {
  const result = await gameApi.getInvitePlayers(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取玩家列表失败')
  }

  return result.data
}

async function getReplayConfirmContext(params) {
  const result = await gameApi.getReplayConfirmContext(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取上局信息失败')
  }

  return result.data
}

async function getSystemRecommendations(params) {
  const result = await gameApi.getSystemRecommendations(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取系统推荐适配局失败')
  }

  return result.data
}

async function getProfitTemplates(params) {
  const result = await gameApi.getProfitTemplates(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取分润模板失败')
  }

  return result.data
}

async function getGameApplyConfig(params = {}) {
  const gameId = String(params.gameId || params.id || '').trim()

  if (!gameId) {
    throw new Error('缺少局信息')
  }

  const result = await gameApi.getGameApplyConfig(encodeURIComponent(gameId), params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取入局申请配置失败')
  }

  return result.data
}

async function applyGame(params = {}) {
  const gameId = String(params.gameId || params.id || '').trim()

  if (!gameId) {
    throw new Error('缺少局信息')
  }

  const intro = String(params.intro || params.applyReason || '').trim()

  if (!intro) {
    throw new Error('请先填写自我介绍')
  }

  const result = await gameApi.applyGame(encodeURIComponent(gameId), {
    applyReason: intro,
    intro,
    message: params.message || '',
    imageFiles: params.imageFiles || [],
    attachmentFiles: params.attachmentFiles || [],
    fromGuideId: params.fromGuideId || ''
  })

  if (result.code !== 0) {
    throw new Error(result.message || '提交入局申请失败')
  }

  return result.data
}

async function createReplayInvitation(payload) {
  const result = await gameApi.createReplayInvitation(payload)

  if (result.code !== 0) {
    throw new Error(result.message || '再次组局发起失败')
  }

  return result.data
}

async function createGameInvite(payload = {}) {
  const result = await gameApi.createGameInvite(payload)

  if (result.code !== 0) {
    throw new Error(result.message || '发起邀请失败')
  }

  return result.data
}

async function getServiceDeliveryDetail(params = {}) {
  const serviceOrderId = String(params.serviceOrderId || params.orderId || params.id || '').trim()

  if (!serviceOrderId) {
    throw new Error('缺少服务订单信息')
  }

  const result = await gameApi.getServiceDeliveryDetail(encodeURIComponent(serviceOrderId), params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取服务交付详情失败')
  }

  return result.data
}

async function remindPlayerConfirm(params = {}) {
  const serviceOrderId = String(params.serviceOrderId || params.orderId || params.id || '').trim()

  if (!serviceOrderId) {
    throw new Error('缺少服务订单信息')
  }

  const result = await gameApi.remindPlayerConfirm(encodeURIComponent(serviceOrderId), params)

  if (result.code !== 0) {
    throw new Error(result.message || '提醒玩家确认失败')
  }

  return result.data
}

async function confirmServiceDelivery(params = {}) {
  const serviceOrderId = String(params.serviceOrderId || params.orderId || params.id || '').trim()
  const confirmedItems = Array.isArray(params.confirmedItems) ? params.confirmedItems.filter(Boolean) : []

  if (!serviceOrderId) {
    throw new Error('缺少服务订单信息')
  }

  if (!confirmedItems.length) {
    throw new Error('请先确认服务内容')
  }

  const result = await gameApi.confirmServiceDelivery(encodeURIComponent(serviceOrderId), {
    gameId: params.gameId || '',
    confirmedItems
  })

  if (result.code !== 0) {
    throw new Error(result.message || '确认服务完成失败')
  }

  return result.data
}

async function getExpertCancelPreview(params = {}) {
  const serviceOrderId = String(params.serviceOrderId || params.orderId || params.id || '').trim()

  if (!serviceOrderId) {
    throw new Error('缺少服务订单信息')
  }

  const result = await gameApi.getExpertCancelPreview(encodeURIComponent(serviceOrderId), params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取取消赔付预览失败')
  }

  return result.data
}

async function cancelServiceWithCompensation(params = {}) {
  const serviceOrderId = String(params.serviceOrderId || params.orderId || params.id || '').trim()
  const reasonCode = String(params.reasonCode || params.reasonKey || '').trim()
  const reasonRemark = String(params.reasonRemark || params.reasonDetail || '').trim()

  if (!serviceOrderId) {
    throw new Error('缺少服务订单信息')
  }

  if (!reasonCode) {
    throw new Error('请选择取消原因')
  }

  if (!reasonRemark) {
    throw new Error('请填写详细说明')
  }

  const result = await gameApi.cancelServiceWithCompensation(encodeURIComponent(serviceOrderId), params)

  if (result.code !== 0) {
    throw new Error(result.message || '取消赔付提交失败')
  }

  return result.data
}

async function getGuideProgress(params) {
  const result = await gameApi.getGuideProgress(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取组局进度失败')
  }

  return result.data
}

async function getGuideSuccess(params) {
  const result = await gameApi.getGuideSuccess(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取组局成功信息失败')
  }

  return result.data
}

async function getGuideCancelDetail(params) {
  const result = await gameApi.getGuideCancelDetail(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取组局取消详情失败')
  }

  return result.data
}

async function getGameManage(params) {
  const result = await gameApi.getGameManage(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取我的组局管理失败')
  }

  return result.data
}

async function getPlayerGameManage(params) {
  const result = await gameApi.getPlayerGameManage(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取玩家组局管理失败')
  }

  return result.data
}

async function getGameCollaboration(params = {}) {
  const gameId = String(params.gameId || params.id || '').trim()

  if (!gameId) {
    throw new Error('缺少局信息')
  }

  const result = await gameApi.getGameCollaboration(encodeURIComponent(gameId), params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取局内协作信息失败')
  }

  return result.data
}

async function endGameCollaboration(params = {}) {
  const gameId = String(params.gameId || params.id || '').trim()

  if (!gameId) {
    throw new Error('缺少局信息')
  }

  const result = await gameApi.endGameCollaboration(encodeURIComponent(gameId), params)

  if (result.code !== 0) {
    throw new Error(result.message || '结束本局失败')
  }

  return result.data
}

async function getGameAudits(params = {}) {
  const result = await gameApi.getGameAudits(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取审核申请失败')
  }

  return result.data
}

async function getGameAuditDetail(params = {}) {
  const auditId = String(params.auditId || params.id || '').trim()

  if (!auditId) {
    throw new Error('缺少审核信息')
  }

  const result = await gameApi.getGameAuditDetail(encodeURIComponent(auditId), params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取审核详情失败')
  }

  return result.data
}

async function respondGameAudit(params = {}) {
  const auditId = String(params.auditId || params.id || '').trim()
  const action = String(params.action || '').trim()

  if (!auditId) {
    throw new Error('缺少审核信息')
  }

  if (!action) {
    throw new Error('缺少审核动作')
  }

  const result = await gameApi.respondGameAudit(encodeURIComponent(auditId), {
    ...params,
    action
  })

  if (result.code !== 0) {
    throw new Error(result.message || '审核处理失败')
  }

  return result.data
}

async function batchRespondGameAudits(params = {}) {
  const auditIds = Array.isArray(params.auditIds) ? params.auditIds.filter(Boolean) : []
  const action = String(params.action || '').trim()

  if (!auditIds.length) {
    throw new Error('请选择审核申请')
  }

  if (!action) {
    throw new Error('缺少审核动作')
  }

  const result = await gameApi.batchRespondGameAudits({
    ...params,
    auditIds,
    action
  })

  if (result.code !== 0) {
    throw new Error(result.message || '批量审核处理失败')
  }

  return result.data
}

async function getPlayAgainOptions(params = {}) {
  const result = await gameApi.getPlayAgainOptions(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取再玩一局推荐失败')
  }

  return result.data
}

async function selectPlayAgainOption(params = {}) {
  const optionId = String(params.optionId || params.id || '').trim()

  if (!optionId) {
    throw new Error('请选择推荐方式')
  }

  const result = await gameApi.selectPlayAgainOption({
    ...params,
    optionId
  })

  if (result.code !== 0) {
    throw new Error(result.message || '推荐方式提交失败')
  }

  return result.data
}

async function getExpertSuccess(params = {}) {
  const result = await gameApi.getExpertSuccess(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取组局成功信息失败')
  }

  return result.data
}

async function getReferralRecords(params = {}) {
  const result = await gameApi.getReferralRecords(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取引荐记录失败')
  }

  return result.data
}

async function triggerReferralRecordAction(params = {}) {
  const recordId = String(params.recordId || params.id || '').trim()
  const actionKey = String(params.actionKey || params.action || params.key || '').trim()

  if (!recordId) {
    throw new Error('缺少引荐记录')
  }

  if (!actionKey) {
    throw new Error('缺少操作类型')
  }

  const result = await gameApi.triggerReferralRecordAction({
    ...params,
    recordId,
    actionKey
  })

  if (result.code !== 0) {
    throw new Error(result.message || '引荐记录操作失败')
  }

  return result.data
}

async function getGameReviewConfig(params = {}) {
  const result = await gameApi.getGameReviewConfig(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取评价配置失败')
  }

  return result.data
}

async function submitGameReview(params = {}) {
  const result = await gameApi.submitGameReview(params)

  if (result.code !== 0) {
    throw new Error(result.message || '提交评价失败')
  }

  return result.data
}

async function getReviewCompleteConfig(params = {}) {
  const result = await gameApi.getReviewCompleteConfig(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取评价完成信息失败')
  }

  return result.data
}

async function selectReviewCompleteIntent(params = {}) {
  const intentId = String(params.intentId || params.id || '').trim()

  if (!intentId) {
    throw new Error('请选择后续意向')
  }

  const result = await gameApi.selectReviewCompleteIntent({
    ...params,
    intentId
  })

  if (result.code !== 0) {
    throw new Error(result.message || '后续意向提交失败')
  }

  return result.data
}

async function respondGameInvitation(invitationId, action) {
  if (!invitationId) {
    throw new Error('缺少邀约信息，无法处理')
  }

  const result = await gameApi.respondGameInvitation(invitationId, {
    action
  })

  if (result.code !== 0) {
    throw new Error(result.message || '处理邀约失败')
  }

  return result.data
}

function hasAnyContext(params, keys) {
  return keys.some((key) => String(params[key] || '').trim())
}

async function getGuideChatContext(params = {}) {
  if (!hasAnyContext(params, ['invitationId', 'gameId', 'serviceOrderId', 'guideId'])) {
    throw new Error('缺少邀请信息')
  }

  const result = await gameApi.getGuideChatContext(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取领路人聊天信息失败')
  }

  return result.data
}

async function respondGuideChatInvitation(params = {}) {
  const invitationId = String(params.invitationId || '').trim()
  const action = String(params.action || '').trim()

  if (!action) {
    throw new Error('缺少处理动作')
  }

  const result = invitationId
    ? await gameApi.respondGameInvitation(encodeURIComponent(invitationId), { action })
    : await gameApi.respondGuideChatInvitation(params)

  if (result.code !== 0) {
    throw new Error(result.message || '处理邀请失败')
  }

  return result.data
}

async function sendGuideChatMessage(params = {}) {
  const message = String(params.message || params.text || '').trim()

  if (!message) {
    throw new Error('请输入消息内容')
  }

  const result = await gameApi.sendGuideChatMessage({
    ...params,
    message
  })

  if (result.code !== 0) {
    throw new Error(result.message || '消息发送失败')
  }

  return result.data
}

async function getGameGreetingContext(params = {}) {
  if (!hasAnyContext(params, ['greetingId', 'gameId', 'serviceOrderId', 'invitationId'])) {
    throw new Error('缺少打招呼信息')
  }

  const result = await gameApi.getGameGreetingContext(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取打招呼信息失败')
  }

  return result.data
}

async function sendGameGreetingMessage(params = {}) {
  const message = String(params.message || params.text || '').trim()

  if (!message) {
    throw new Error('请输入消息内容')
  }

  const result = await gameApi.sendGameGreetingMessage({
    ...params,
    message
  })

  if (result.code !== 0) {
    throw new Error(result.message || '消息发送失败')
  }

  return result.data
}

async function getHallGreetingContext(params = {}) {
  const result = await gameApi.getHallGreetingContext(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取大厅打招呼配置失败')
  }

  return result.data
}

async function sendHallGreetingMessage(params = {}) {
  const message = String(params.message || params.text || '').trim()

  if (!message) {
    throw new Error('请输入消息内容')
  }

  const result = await gameApi.sendHallGreetingMessage({
    ...params,
    message
  })

  if (result.code !== 0) {
    throw new Error(result.message || '消息发送失败')
  }

  return result.data
}

async function triggerHallGreetingAction(params = {}) {
  const actionKey = String(params.actionKey || params.key || '').trim()

  if (!actionKey) {
    throw new Error('缺少操作类型')
  }

  const result = await gameApi.triggerHallGreetingAction({
    ...params,
    actionKey
  })

  if (result.code !== 0) {
    throw new Error(result.message || '操作提交失败')
  }

  return result.data
}

async function getGamePaymentConfig(params = {}) {
  const gameId = String(params.gameId || params.id || '').trim()
  const result = await gameApi.getGamePaymentConfig(gameId ? encodeURIComponent(gameId) : '', params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取支付配置失败')
  }

  return result.data
}

async function createGamePayment(payload) {
  const result = await gameApi.createGamePayment(payload)

  if (result.code !== 0) {
    throw new Error(result.message || '创建支付订单失败')
  }

  return result.data
}

module.exports = {
  getGameList,
  getGameDetail,
  getGameMembers,
  respondGameInvitation,
  getInvitePlayerConfig,
  getGameInviteConfig,
  getInviteRecentPlayers,
  getInvitePlayers,
  getReplayConfirmContext,
  getSystemRecommendations,
  getProfitTemplates,
  getGameApplyConfig,
  applyGame,
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
  getGuideChatContext,
  respondGuideChatInvitation,
  sendGuideChatMessage,
  getGameGreetingContext,
  sendGameGreetingMessage,
  getHallGreetingContext,
  sendHallGreetingMessage,
  triggerHallGreetingAction,
  getGamePaymentConfig,
  getServiceDeliveryDetail,
  remindPlayerConfirm,
  confirmServiceDelivery,
  getExpertCancelPreview,
  cancelServiceWithCompensation,
  createGamePayment
}
