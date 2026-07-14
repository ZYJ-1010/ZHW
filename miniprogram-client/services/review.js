const reviewApi = require('../api/modules/review')

async function getAvailableReviews(params) {
  const result = await reviewApi.getAvailableReviews(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取可评价局失败')
  }

  return result.data
}

async function getCompleteConfig(params) {
  const result = await reviewApi.getCompleteConfig(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取评价完成配置失败')
  }

  return result.data
}

async function submitReview(data) {
  const result = await reviewApi.submitReview(data)

  if (result.code !== 0) {
    throw new Error(result.message || '提交评价失败')
  }

  return result.data
}

module.exports = {
  getAvailableReviews,
  getCompleteConfig,
  submitReview
}
