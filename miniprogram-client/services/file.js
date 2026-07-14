const env = require('../config/env')
const fileApi = require('../api/modules/file')

function guessFileName(path = '', index = 0) {
  const clean = String(path || '').split('?')[0]
  const name = clean.split('/').filter(Boolean).pop()

  return name || `evidence-${Date.now()}-${index}.jpg`
}

function guessMimeType(fileName = '') {
  const lower = String(fileName).toLowerCase()

  if (lower.endsWith('.png')) {
    return 'image/png'
  }

  if (lower.endsWith('.webp')) {
    return 'image/webp'
  }

  if (lower.endsWith('.pdf')) {
    return 'application/pdf'
  }

  if (lower.endsWith('.txt')) {
    return 'text/plain'
  }

  if (lower.endsWith('.zip')) {
    return 'application/zip'
  }

  if (lower.endsWith('.mp3')) {
    return 'audio/mpeg'
  }

  if (lower.endsWith('.m4a') || lower.endsWith('.mp4')) {
    return 'audio/mp4'
  }

  if (lower.endsWith('.aac')) {
    return 'audio/aac'
  }

  if (lower.endsWith('.amr')) {
    return 'audio/amr'
  }

  if (lower.endsWith('.wav')) {
    return 'audio/wav'
  }

  return 'image/jpeg'
}

function normalizeUploadFile(file, index) {
  if (typeof file === 'string') {
    return {
      path: file,
      fileName: guessFileName(file, index),
      mimeType: '',
      size: 1
    }
  }

  const path = file && (file.path || file.tempFilePath || file.filePath) || ''
  const fileName = file && (file.fileName || file.name) || guessFileName(path, index)

  return {
    path,
    fileName,
    mimeType: file && file.mimeType || '',
    size: Number(file && file.size || 1) || 1
  }
}

function isLocalUploadURL(uploadURL = '') {
  const value = String(uploadURL || '').trim().toLowerCase()

  return value.startsWith('mock://') || value.startsWith('local://')
}

function uploadToSignedURL(filePath, upload) {
  const uploadURL = upload && (upload.uploadUrl || upload.uploadURL) || ''

  if (env.isMock || isLocalUploadURL(uploadURL) || typeof wx === 'undefined' || typeof wx.uploadFile !== 'function') {
    return Promise.resolve()
  }

  if (!uploadURL) {
    return Promise.reject(new Error('上传地址缺失'))
  }

  return new Promise((resolve, reject) => {
    wx.uploadFile({
      url: uploadURL,
      filePath,
      name: 'file',
      header: upload.headers || {},
      formData: upload.formData || {},
      success: (response) => {
        if (response.statusCode >= 400) {
          const detail = typeof response.data === 'string'
            ? response.data.replace(/\s+/g, ' ').slice(0, 160)
            : ''
          reject(new Error(`文件上传失败：COS ${response.statusCode}${detail ? ` ${detail}` : ''}`))
          return
        }
        resolve(response)
      },
      fail: reject
    })
  })
}

async function uploadEvidenceImages(paths = [], options = {}) {
  const fileIds = []
  const bizType = options.bizType || 'report_attachment'
  const objectId = Number(options.objectId || 0) || 0

  for (let index = 0; index < paths.length; index += 1) {
    const uploadFile = normalizeUploadFile(paths[index], index)

    if (!uploadFile.path) {
      continue
    }

    const result = await fileApi.createUploadToken({
      bizType,
      objectId,
      fileName: uploadFile.fileName,
      mimeType: uploadFile.mimeType || guessMimeType(uploadFile.fileName),
      size: uploadFile.size
    })

    if (result.code !== 0) {
      throw new Error(result.message || '获取上传凭证失败')
    }

    const upload = result.data && result.data.upload ? result.data.upload : {}
    const file = result.data && result.data.file ? result.data.file : {}

    await uploadToSignedURL(uploadFile.path, upload)
    fileIds.push(upload.fileId || file.fileId)
  }

  return fileIds.filter(Boolean)
}

async function uploadSingleFile(path, options = {}) {
  const ids = await uploadEvidenceImages([path], options)
  return ids[0] || 0
}

module.exports = {
  uploadEvidenceImages,
  uploadSingleFile
}
