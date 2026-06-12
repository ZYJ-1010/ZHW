const { PAGE_GROUPS } = require('../../config/page-map')

Page({
  data: {
    group: PAGE_GROUPS.find((item) => item.code === 'map')
  }
})
