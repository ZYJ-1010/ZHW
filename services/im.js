const imApi = require('../api/modules/im')

async function getMessages(roomId, params) {
  const result = await imApi.getMessages(roomId, params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取消息失败')
  }

  return result.data
}

module.exports = {
  getMessages
}
