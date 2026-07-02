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

module.exports = {
  getChatRoom,
  getRoomByGame,
  getMessages,
  sendMessage
}
