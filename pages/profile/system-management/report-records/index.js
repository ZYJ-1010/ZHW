Page({
  data: {
    activeFilter: 'all',
    stats: [
      { label: '总举报', value: '8' },
      { label: '举报成功', value: '5' },
      { label: '举报失败', value: '3', className: 'danger' },
      { label: '累计奖励', value: '¥140' }
    ],
    filters: [
      { key: 'all', label: '全部' },
      { key: 'success', label: '举报成功' },
      { key: 'failed', label: '举报失败' },
      { key: 'partial', label: '部分属实' }
    ],
    records: [
      {
        id: 'record-1',
        title: '举报「用户A」诱导私下交易',
        filter: 'success',
        status: '举报成功',
        statusClass: 'success',
        parts: [
          { text: '处理结果：' },
          { text: '举报属实', className: 'text-success' },
          { text: '，被举报人扣除信用分5分，平台奖励您' },
          { text: '25元', className: 'text-success' }
        ],
        time: '2024-06-12 10:20'
      },
      {
        id: 'record-2',
        title: '举报「用户B」言语骚扰',
        filter: 'success',
        status: '举报成功',
        statusClass: 'success',
        parts: [
          { text: '处理结果：' },
          { text: '举报属实', className: 'text-success' },
          { text: '，被举报人扣除信用分3分，平台奖励您' },
          { text: '15元', className: 'text-success' }
        ],
        time: '2024-06-09 18:30'
      },
      {
        id: 'record-3',
        title: '举报「用户C」虚假信息',
        filter: 'failed',
        status: '举报失败',
        statusClass: 'danger',
        parts: [
          { text: '处理结果：' },
          { text: '不属实', className: 'text-danger' },
          { text: '，扣除您的信用分2分，请注意核实后再举报' }
        ],
        time: '2024-06-01 11:00'
      },
      {
        id: 'record-4',
        title: '举报「用户D」恶意取消',
        filter: 'partial',
        status: '部分属实',
        statusClass: 'warning',
        parts: [
          { text: '处理结果：' },
          { text: '部分属实', className: 'text-warning' },
          { text: '，被举报人扣除信用分1分，平台奖励您' },
          { text: '5元', className: 'text-success' }
        ],
        time: '2024-05-25 14:10'
      },
      {
        id: 'record-5',
        title: '举报「用户E」诱导私下交易',
        filter: 'success',
        status: '举报成功',
        statusClass: 'success',
        parts: [
          { text: '处理结果：' },
          { text: '举报属实', className: 'text-success' },
          { text: '，被举报人扣除信用分5分，平台奖励您' },
          { text: '50元', className: 'text-success' }
        ],
        time: '2024-05-20 09:45'
      }
    ],
    visibleRecords: []
  },

  onLoad() {
    this.applyFilter(this.data.activeFilter)
  },

  handleTabTap(event) {
    const { target } = event.currentTarget.dataset
    const routeMap = {
      home: '/pages/profile/system-management/report-center/index',
      appeals: '/pages/profile/system-management/report-appeals/index'
    }

    if (routeMap[target]) {
      wx.redirectTo({ url: routeMap[target] })
    }
  },

  handleFilterTap(event) {
    const { key } = event.currentTarget.dataset

    if (key) {
      this.applyFilter(key)
    }
  },

  applyFilter(key) {
    const visibleRecords = key === 'all'
      ? this.data.records
      : this.data.records.filter((item) => item.filter === key)

    this.setData({
      activeFilter: key,
      visibleRecords
    })
  },

  handleDetailTap(event) {
    const { id } = event.currentTarget.dataset
    const query = id ? `?recordId=${id}` : ''

    wx.redirectTo({
      url: `/pages/profile/system-management/report-record-detail/index${query}`
    })
  }
})
