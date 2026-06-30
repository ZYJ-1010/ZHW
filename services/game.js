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

async function createReplayInvitation(payload) {
  const result = await gameApi.createReplayInvitation(payload)

  if (result.code !== 0) {
    throw new Error(result.message || '再次组局发起失败')
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
  getInviteRecentPlayers,
  getInvitePlayers,
  getReplayConfirmContext,
  getSystemRecommendations,
  getProfitTemplates,
  createReplayInvitation,
  getGuideProgress,
  getGuideSuccess,
  getGuideCancelDetail,
  getGameManage,
  getPlayerGameManage,
  createGamePayment
}
