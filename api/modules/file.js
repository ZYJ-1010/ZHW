const request = require('../request')

function createUploadToken(data) {
  return request.post('/api/app/files/upload-token', data)
}

function getDownloadURL(fileId) {
  return request.get(`/api/app/files/${fileId}/download-url`)
}

module.exports = {
  createUploadToken,
  getDownloadURL
}

