const imApi = require('../api/modules/im')
const imSocket = require('./im-socket')

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

async function getPrivateMessages(targetUserId, params) {
  const result = await imApi.getPrivateMessages(targetUserId, params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取私聊失败')
  }

  return result.data
}

async function sendPrivateMessage(targetUserId, data) {
  const result = await imApi.sendPrivateMessage(targetUserId, data)

  if (result.code !== 0) {
    throw new Error(result.message || '发送私聊失败')
  }

  return result.data
}

module.exports = {
  getChatRoom,
  getRoomByGame,
  getMessages,
  sendMessage,
  getPrivateMessages,
  sendPrivateMessage,
  connectGameSocket: imSocket.connectGameSocket
}
