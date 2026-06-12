const ENV = {
  MOCK: 'mock',
  PROD: 'prod'
}

const currentEnv = ENV.MOCK

const serverMap = {
  [ENV.MOCK]: '',
  [ENV.PROD]: 'https://api.example.com'
}

module.exports = {
  ENV,
  currentEnv,
  isMock: currentEnv === ENV.MOCK,
  baseUrl: serverMap[currentEnv]
}
