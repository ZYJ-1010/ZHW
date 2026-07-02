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

async function reviewGameApplication(applicationId, approve) {
  const result = await gameApi.reviewGameApplication(applicationId, {
    approve: Boolean(approve)
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

async function requestExpertCancel(gameId, payload) {
  const result = await gameApi.requestExpertCancel(gameId, payload)

  if (result.code !== 0) {
    throw new Error(result.message || '专家取消赔付提交失败')
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
  createGuideFollowUp,
  createGame,
  createInviteEntry,
  respondGameInvitation,
  getInvitePlayerConfig,
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
  sendGuideReminder,
  getGuideProgress,
  getGuideCancelDetail,
  getCancelConfig,
  getReferralRecords,
  getGameManage,
  getPlayerGameManage,
  getMyFavoriteGames,
  createGamePayment,
  confirmService,
  createRetrospective,
  requestPlayerCancel,
  requestExpertCancel,
  favoriteGame
}
