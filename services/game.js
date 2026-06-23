const gameApi = require('../api/modules/game')

async function getGameList(params) {
  const result = await gameApi.getGames(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取局列表失败')
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
  respondGameInvitation,
  getInvitePlayerConfig,
  getInviteRecentPlayers,
  getInvitePlayers,
  getGuideProgress,
  getGuideCancelDetail,
  getGameManage,
  getPlayerGameManage,
  createGamePayment
}
