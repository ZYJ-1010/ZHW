const gameApi = require('../api/modules/game')

async function getGameList(params) {
  const result = await gameApi.getGames(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取局列表失败')
  }

  return result.data
}

async function getGameDetail(gameId) {
  const result = await gameApi.getGameDetail(gameId)

  if (result.code !== 0) {
    throw new Error(result.message || '获取局详情失败')
  }

  return result.data
}

async function getGameMembers(gameId) {
  const result = await gameApi.getGameMembers(gameId)

  if (result.code !== 0) {
    throw new Error(result.message || '获取参与者失败')
  }

  return result.data
}

async function getGameSuccessDetail(gameId, params) {
  const result = await gameApi.getGameSuccessDetail(gameId, params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取组局成功详情失败')
  }

  return result.data
}

async function getGameGuideSuccessDetail(gameId, params) {
  const result = await gameApi.getGameGuideSuccessDetail(gameId, params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取领路人成功详情失败')
  }

  return result.data
}

async function getGameCollaboration(gameId, params) {
  const result = await gameApi.getGameCollaboration(gameId, params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取局内协作失败')
  }

  return result.data
}

async function requestGameCompletion(gameId) {
  const result = await gameApi.requestGameCompletion(gameId)

  if (result.code !== 0) {
    throw new Error(result.message || '结束组局失败')
  }

  return result.data
}

async function createGuideFollowUp(gameId, payload) {
  const result = await gameApi.createGuideFollowUp(gameId, payload)

  if (result.code !== 0) {
    throw new Error(result.message || '记录领路人跟进失败')
  }

  return result.data
}

async function createGame(payload) {
  const result = await gameApi.createGame(payload)

  if (result.code !== 0) {
    throw new Error(result.message || '创建局失败')
  }

  return result.data
}

async function getGameDrafts() {
  const result = await gameApi.getGameDrafts()
  if (result.code !== 0) {
    throw new Error(result.message || '获取草稿箱失败')
  }
  return result.data
}

async function getGameDraft(draftId) {
  const result = await gameApi.getGameDraft(draftId)
  if (result.code !== 0) {
    throw new Error(result.message || '获取草稿失败')
  }
  return result.data
}

async function saveGameDraft(payload = {}) {
  const draftId = Number(payload.id || 0)
  const body = { title: payload.title || '', payload: payload.payload || {} }
  const result = draftId > 0
    ? await gameApi.updateGameDraft(draftId, body)
    : await gameApi.createGameDraft(body)
  if (result.code !== 0) {
    throw new Error(result.message || '保存草稿失败')
  }
  return result.data
}

async function deleteGameDraft(draftId) {
  const result = await gameApi.deleteGameDraft(draftId)
  if (result.code !== 0) {
    throw new Error(result.message || '删除草稿失败')
  }
  return result.data
}

async function startGame(gameId) {
  const result = await gameApi.startGame(gameId)

  if (result.code !== 0) {
    throw new Error(result.message || '开始组局失败')
  }

  return result.data
}

async function createInviteEntry(payload) {
  const result = await gameApi.createInviteEntry(payload)

  if (result.code !== 0) {
    throw new Error(result.message || '生成邀请入口失败')
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

async function getInvitePermission(params) {
  const result = await gameApi.getInvitePermission(params)

  if (result.code !== 0) {
    throw new Error(result.message || '引荐权限校验失败')
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

async function getCategoryConfig(params) {
  const result = await gameApi.getCategoryConfig(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取组局分类配置失败')
  }

  return result.data
}

async function getApplicationConfig(params) {
  const result = await gameApi.getApplicationConfig(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取入局申请配置失败')
  }

  return result.data
}

async function getConditionRuleConfig(params) {
  const result = await gameApi.getConditionRuleConfig(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取条件局规则失败')
  }

  return result.data
}

async function applyGame(gameId, payload) {
  const result = await gameApi.applyGame(gameId, payload)

  if (result.code !== 0) {
    throw new Error(result.message || '提交报名申请失败')
  }

  return result.data
}

async function getReceivedApplications(params) {
  const result = await gameApi.getReceivedApplications(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取申请列表失败')
  }

  return result.data
}

async function reviewGameApplication(applicationId, approve, rejectReason = '') {
  const result = await gameApi.reviewGameApplication(applicationId, {
    approve: Boolean(approve),
    rejectReason: String(rejectReason || '').trim()
  })

  if (result.code !== 0) {
    throw new Error(result.message || '审核申请失败')
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

async function createCurrentGameInvitation(payload) {
  const result = await gameApi.createCurrentGameInvitation(payload)

  if (result.code !== 0) {
    throw new Error(result.message || '邀请进入组局失败')
  }

  return result.data
}

async function sendGuideReminder(payload) {
  const result = await gameApi.sendGuideReminder(payload)

  if (result.code !== 0) {
    throw new Error(result.message || '发送提醒失败')
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

async function getGuideCancelDetail(params) {
  const result = await gameApi.getGuideCancelDetail(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取组局取消详情失败')
  }

  return result.data
}

async function getCancelConfig(params) {
  const result = await gameApi.getCancelConfig(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取取消配置失败')
  }

  return result.data
}

async function getReferralRecords(params) {
  const result = await gameApi.getReferralRecords(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取引荐记录失败')
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

async function getMyFavoriteGames(params) {
  const result = await gameApi.getMyFavoriteGames(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取收藏组局失败')
  }

  return result.data
}

async function respondGameInvitation(invitationId, action) {
  if (!invitationId) {
    throw new Error('缺少邀约信息，无法处理')
  }

  const normalizedAction = String(action || '').trim().toLowerCase()
  const acceptActions = ['accept', 'accepted', 'approve', 'approved', 'yes', 'confirm', 'confirmed', 'join', 'agree', '确认', '确认参加', '接受', '同意', '加入']
  const accept = acceptActions.indexOf(normalizedAction) !== -1

  const result = await gameApi.respondGameInvitation(invitationId, {
    accept,
    action
  })

  if (result.code !== 0) {
    throw new Error(result.message || '处理邀约失败')
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

async function getGamePaymentPreview(gameId) {
  const result = await gameApi.getGamePaymentPreview(gameId)
  if (result.code !== 0) {
    throw new Error(result.message || '后端未返回支付信息')
  }
  const data = result.data || {}
  if (!data.payment || Number(data.payment.gameId || 0) !== Number(gameId)) {
    throw new Error('后端未返回完整的支付信息')
  }
  return data
}

async function confirmService(gameId, payload) {
  const result = await gameApi.confirmService(gameId, payload)

  if (result.code !== 0) {
    throw new Error(result.message || '服务确认失败')
  }

  return result.data
}

async function createRetrospective(gameId, payload) {
  const result = await gameApi.createRetrospective(gameId, payload)

  if (result.code !== 0) {
    throw new Error(result.message || '复盘意愿提交失败')
  }

  return result.data
}

async function requestPlayerCancel(gameId, payload) {
  const result = await gameApi.requestPlayerCancel(gameId, payload)

  if (result.code !== 0) {
    throw new Error(result.message || '取消申请提交失败')
  }

  return result.data
}

async function getPlayerCancelDetail(gameId) {
  const result = await gameApi.getPlayerCancelDetail(gameId)
  if (result.code !== 0) {
    throw new Error(result.message || '后端未返回玩家取消详情')
  }
  if (!result.data || !result.data.gameId || !result.data.serviceOrderId) {
    throw new Error('后端未返回完整的玩家取消详情')
  }
  return result.data
}

async function requestExpertCancel(gameId, payload) {
  const result = await gameApi.requestExpertCancel(gameId, payload)

  if (result.code !== 0) {
    throw new Error(result.message || '专家取消赔付提交失败')
  }

  return result.data
}

async function getExpertCancelDetail(gameId) {
  const result = await gameApi.getExpertCancelDetail(gameId)
  if (result.code !== 0) {
    throw new Error(result.message || '后端未返回行家取消详情')
  }
  if (!result.data || !result.data.gameId || !result.data.serviceOrderId) {
    throw new Error('后端未返回完整的行家取消详情')
  }
  return result.data
}

async function favoriteGame(gameId) {
  const result = await gameApi.favoriteGame(gameId)

  if (result.code !== 0) {
    throw new Error(result.message || '收藏组局失败')
  }

  return result.data
}

module.exports = {
  getGameList,
  getGameDetail,
  getGameMembers,
  getGameSuccessDetail,
  getGameGuideSuccessDetail,
  getGameCollaboration,
  requestGameCompletion,
  createGuideFollowUp,
  createGame,
  getGameDrafts,
  getGameDraft,
  saveGameDraft,
  deleteGameDraft,
  startGame,
  createInviteEntry,
  respondGameInvitation,
  getInvitePlayerConfig,
  getInvitePermission,
  getInviteRecentPlayers,
  getInvitePlayers,
  getReplayConfirmContext,
  getSystemRecommendations,
  getProfitTemplates,
  getCategoryConfig,
  getApplicationConfig,
  getConditionRuleConfig,
  applyGame,
  getReceivedApplications,
  reviewGameApplication,
  createReplayInvitation,
  createCurrentGameInvitation,
  sendGuideReminder,
  getGuideProgress,
  getGuideCancelDetail,
  getCancelConfig,
  getReferralRecords,
  getGameManage,
  getPlayerGameManage,
  getMyFavoriteGames,
  createGamePayment,
  getGamePaymentPreview,
  confirmService,
  createRetrospective,
  requestPlayerCancel,
  getPlayerCancelDetail,
  requestExpertCancel,
  getExpertCancelDetail,
  favoriteGame
}
