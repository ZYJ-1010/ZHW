const request = require('../request')

function getNewbieTasks() {
  return request.get('/api/app/newbie-tasks')
}

module.exports = {
  getNewbieTasks
}
