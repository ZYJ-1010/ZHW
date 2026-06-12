const request = require('../request')

function getAvailableReviews(params) {
  return request.get('/api/app/reviews/available', params)
}

function submitReview(data) {
  return request.post('/api/app/reviews', data)
}

module.exports = {
  getAvailableReviews,
  submitReview
}
