const request = require('../request')

function getChatRoom(gameId) {
  return request.get(`/api/app/games/${gameId}/chat-room`)
}

function getRoomByGame(gameId) {
  return request.get(`/api/app/games/${gameId}/chat-session`)
}

function getMessages(gameId, params) {
  return request.get(`/api/app/games/${gameId}/chat/messages`, params)
}

function sendMessage(gameId, data) {
  return request.post(`/api/app/games/${gameId}/chat/messages`, data)
}

function getPrivateMessages(targetUserId, params) {
  return request.get('/api/app/private-chat/messages', Object.assign({
    targetUserId
  }, params || {}))
}

function sendPrivateMessage(targetUserId, data) {
  return request.post('/api/app/private-chat/messages', Object.assign({}, data || {}, {
    targetUserId
  }))
}

module.exports = {
  getChatRoom,
  getRoomByGame,
  getMessages,
  sendMessage,
  getPrivateMessages,
  sendPrivateMessage
}
