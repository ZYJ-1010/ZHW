const { ROUTES } = require('../../../config/routes')
const mapService = require('../../../services/map')
const fileService = require('../../../services/file')
const toast = require('../../../utils/toast')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')

const NAV_ITEMS = [
  { name: '我的', key: 'mine' },
  { name: '元宇宙', key: 'metaverse' },
  { name: '地图', key: 'map' },
  { name: '消息', key: 'message' },
  { name: '首页', key: 'home' }
]

const EMPTY_DETAIL = {
  distanceText: '',
  spotName: '',
  statusTitle: '',
  statusDesc: '',
  storyTitle: '',
  storyPlaceholder: '',
  storyMinLength: 0,
  storyMaxLength: 120,
  rewards: []
}

const EMPTY_PAGE = {
  taskSectionTitle: '',
  generateButtonText: '',
  texts: {},
  checkinDetail: EMPTY_DETAIL,
  checkinTasks: []
}

function applyTemplate(template, values = {}) {
  let text = String(template || '')
  Object.keys(values).forEach((key) => {
    text = text.replace(new RegExp(`\\{${key}\\}`, 'g'), String(values[key]))
  })
  return text
}

Page({
  data: {
    onlineText: '在线',
    navItems: NAV_ITEMS,
    pageConfig: EMPTY_PAGE,
    checkinDetail: EMPTY_DETAIL,
    checkinTasks: [],
    storyText: '',
    storyCountText: `0/${EMPTY_DETAIL.storyMaxLength}`,
    selectedFileIds: [],
    selectedPhotoCount: 0,
    isSubmitting: false,
    pointId: '',
    challengeId: '',
    gameId: ''
  },

  onLoad(options = {}) {
    this.setData({
      pointId: options.pointId || '',
      challengeId: options.challengeId || '',
      gameId: options.gameId || ''
    })
    this.loadPageConfig()
  },

  async loadPageConfig() {
    try {
      const pageConfig = Object.assign({}, EMPTY_PAGE, await mapService.getPlayPage('real-checkin'))
      const checkinDetail = Object.assign({}, EMPTY_DETAIL, pageConfig.checkinDetail || {})
      this.setData({
        pageConfig,
        checkinDetail,
        checkinTasks: pageConfig.checkinTasks || [],
        storyCountText: `0/${checkinDetail.storyMaxLength}`
      })
    } catch (error) {
      toast.info(error.message || '真实打卡加载失败')
    }
  },

  handleCloseTap() {
    wx.navigateBack({
      delta: 1,
      fail: () => {
        navigateShellRoute(ROUTES.map, {
          currentRoute: ROUTES.mapRealCheckin
        })
      }
    })
  },

  handleShootTap() {
    const texts = this.data.pageConfig.texts || {}
    if (!wx.chooseMedia && !wx.chooseImage) {
      wx.showToast({ title: texts.unsupportedCamera || '', icon: 'none' })
      return
    }

    const onSuccess = async (res) => {
      const tempFile = res.tempFiles && res.tempFiles[0]
      const path = tempFile && tempFile.tempFilePath || res.tempFilePaths && res.tempFilePaths[0]

      if (!path) {
        toast.info(texts.photoRequired || '')
        return
      }

      try {
        const fileId = await fileService.uploadSingleFile(path, {
          bizType: 'map_checkin',
          fileName: 'map-checkin.jpg'
        })
        this.setData({
          selectedFileIds: [fileId],
          selectedPhotoCount: 1,
          checkinDetail: Object.assign({}, this.data.checkinDetail, {
            statusTitle: '照片已选择',
            statusDesc: '可填写故事并提交审核'
          })
        })
        toast.info(texts.photoSelected || '照片已选择')
      } catch (error) {
        console.warn('upload map checkin photo failed', error)
        toast.info(error.message || texts.unsupportedCamera || '')
      }
    }

    if (wx.chooseMedia) {
      wx.chooseMedia({
        count: 1,
        mediaType: ['image'],
        sourceType: ['camera', 'album'],
        sizeType: ['compressed'],
        success: onSuccess
      })
      return
    }

    wx.chooseImage({
      count: 1,
      sourceType: ['camera', 'album'],
      sizeType: ['compressed'],
      success: onSuccess
    })
  },

  handleStoryTap() {
    const texts = this.data.pageConfig.texts || {}
    wx.showToast({
      title: this.data.storyText ? texts.storySaved : texts.storyRequired,
      icon: 'none'
    })
  },

  handleStoryInput(event) {
    const value = event.detail.value || ''
    const maxLength = this.data.checkinDetail.storyMaxLength || EMPTY_DETAIL.storyMaxLength

    this.setData({
      storyText: value,
      storyCountText: `${value.length}/${maxLength}`
    })
  },

  handleTaskTap(event) {
    const task = (this.data.checkinTasks || []).find((item) => item.id === event.currentTarget.dataset.id)
    const texts = this.data.pageConfig.texts || {}

    wx.showToast({
      title: task ? applyTemplate(texts.taskSelected, { title: task.title }) : texts.taskRequired,
      icon: 'none'
    })
  },

  handleFragmentTap() {
    const texts = this.data.pageConfig.texts || {}
    const minLength = this.data.checkinDetail.storyMinLength || 0

    if (this.data.isSubmitting) {
      return
    }

    if (!this.data.selectedFileIds.length) {
      toast.info(texts.photoRequired || '请先上传打卡照片')
      return
    }

    if (this.data.storyText.length < minLength) {
      toast.info(texts.storyRequired || `请至少填写${minLength}字故事`)
      return
    }

    wx.showModal({
      title: '提交打卡审核',
      content: '提交后进入后台审核，审核通过后生成足迹碎片。',
      confirmText: '提交',
      cancelText: '取消',
      success: async (res) => {
        if (!res.confirm) {
          return
        }

        this.setData({
          isSubmitting: true
        })
        try {
          const checkin = await mapService.submitCheckin({
            pointId: this.data.pointId,
            challengeId: this.data.challengeId,
            gameId: this.data.gameId,
            story: this.data.storyText,
            fileIds: this.data.selectedFileIds
          })
          this.setData({
            isSubmitting: false,
            checkinDetail: Object.assign({}, this.data.checkinDetail, {
              statusTitle: checkin.statusText || '待审核',
              statusDesc: checkin.statusDesc || '打卡材料已提交，等待后台审核'
            })
          })
          toast.info(texts.fragmentPending || '已提交审核')
        } catch (error) {
          this.setData({
            isSubmitting: false
          })
          toast.info(error.message || '提交打卡失败')
        }
      }
    })
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}
    navigateShellKey(key, {
      currentRoute: ROUTES.mapRealCheckin
    })
  }
})
