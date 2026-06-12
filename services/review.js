const reviewApi = require('../api/modules/review')

async function getAvailableReviews(params) {
  const result = await reviewApi.getAvailableReviews(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取可评价局失败')
  }

  return result.data
}

module.exports = {
  getAvailableReviews
}
