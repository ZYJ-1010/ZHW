const fileService = require('./file')
const imService = require('./im')

function toPositiveInt(value) {
  const number = Number(value)
  return Number.isInteger(number) && number > 0 ? number : 0
}

function fileNameFromPath(path = '', fallback = 'chat-file') {
  const clean = String(path || '').split('?')[0].split('#')[0]
  const name = clean.split('/').filter(Boolean).pop()
  return name || `${fallback}-${Date.now()}`
}

function firstChosenFile(event) {
  const detail = event && event.detail ? event.detail : event || {}
  const files = detail.tempFiles || detail.files || []
  const paths = detail.tempFilePaths || []
  const file = files[0] || {}
  const path = file.tempFilePath || file.path || paths[0] || ''
  const fileName = file.name || file.fileName || fileNameFromPath(path)

  return {
    path,
    fileName,
    size: Number(file.size || 1) || 1
  }
}

function displayText(messageType, fileName) {
  if (messageType === 'image') {
    return '图片已发送'
  }
  return fileName ? `文件：${fileName}` : '文件已发送'
}

async function sendChosenFile(gameId, event, messageType) {
  const normalizedGameId = toPositiveInt(gameId)
  if (!normalizedGameId) {
    throw new Error('缺少局 ID，无法发送到局内 IM')
  }
  const chosenFile = firstChosenFile(event)
  if (!chosenFile.path) {
    throw new Error('未选择文件')
  }

  const fileIds = await fileService.uploadEvidenceImages([chosenFile.path], {
    bizType: 'chat_file',
    objectId: normalizedGameId
  })
  const fileId = fileIds[0]
  if (!fileId) {
    throw new Error('文件上传失败')
  }

  const message = await imService.sendMessage(normalizedGameId, {
    messageType,
    content: chosenFile.fileName,
    fileId
  })

  return {
    message,
    fileId,
    fileName: chosenFile.fileName,
    text: displayText(messageType, chosenFile.fileName)
  }
}

module.exports = {
  sendChosenFile
}
