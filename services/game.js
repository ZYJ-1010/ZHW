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

module.exports = {
  getGameList,
  respondGameInvitation,
  getInvitePlayerConfig,
  getInviteRecentPlayers,
  getInvitePlayers,
  getGuideProgress
}
