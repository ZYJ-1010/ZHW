const imApi = require('../api/modules/im')

async function getMessages(roomId, params) {
  const result = await imApi.getMessages(roomId, params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取消息失败')
  }

  return result.data
}

async function sendMessage(roomId, payload = {}) {
  const content = String(payload.content || payload.text || '').trim()

  if (!roomId) {
    throw new Error('缺少会话信息，无法发送消息')
  }

  if (!content) {
    throw new Error('请输入消息内容')
  }

  const result = await imApi.sendMessage(roomId, {
    type: payload.type || 'text',
    content
  })

  if (result.code !== 0) {
    throw new Error(result.message || '消息发送失败')
  }

  return result.data
}

module.exports = {
  getMessages,
  sendMessage
}
