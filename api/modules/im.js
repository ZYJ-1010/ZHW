const request = require('../request')

function getRoomByGame(gameId) {
  return request.get(`/api/im/games/${gameId}/room`)
}

function getMessages(roomId, params) {
  return request.get(`/api/im/rooms/${roomId}/messages`, params)
}

function sendMessage(roomId, data) {
  return request.post(`/api/im/rooms/${roomId}/messages`, data)
}

module.exports = {
  getRoomByGame,
  getMessages,
  sendMessage
}
