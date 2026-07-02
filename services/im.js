const imApi = require('../api/modules/im')

async function getChatRoom(gameId) {
  const result = await imApi.getChatRoom(gameId)

  if (result.code !== 0) {
    throw new Error(result.message || '获取房间失败')
  }

  return result.data
}

async function getRoomByGame(gameId) {
  const result = await imApi.getRoomByGame(gameId)

  if (result.code !== 0) {
    throw new Error(result.message || '获取会话失败')
  }

  return result.data
}

async function getMessages(gameId, params) {
  const result = await imApi.getMessages(gameId, params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取消息失败')
  }

  return result.data
}

async function sendMessage(gameId, data) {
  const result = await imApi.sendMessage(gameId, data)

  if (result.code !== 0) {
    throw new Error(result.message || '发送消息失败')
  }

  return result.data
}

module.exports = {
  getChatRoom,
  getRoomByGame,
  getMessages,
  sendMessage
}
