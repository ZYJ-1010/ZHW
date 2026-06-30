const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const ASSET_BASE = '/pages/profile/system-management/block-settings/assets'

Page({
  data: {
    icons: {
      lock: `${ASSET_BASE}/icon-lock.svg`,
      ban: `${ASSET_BASE}/icon-ban.svg`,
      check: `${ASSET_BASE}/icon-check.svg`,
      chevron: `${ASSET_BASE}/icon-chevron-right.svg`
    },
    selectedScene: '',
    selectedMode: '',
    selectedSceneTitle: '',
    showModeSheet: false,
    scenes: [],
    modeOptions: [],
    rules: []
  },

  onLoad(options) {
    this.loadSceneConfig(options && options.sheet)
  },

  async loadSceneConfig(sheet) {
    try {
      const data = await profileService.getSystemBlockSettings()
      const sceneConfig = data.sceneConfig || data.scenesConfig || {}

      this.setData({
        scenes: this.normalizeList(sceneConfig.scenes || data.scenes),
        modeOptions: this.normalizeList(sceneConfig.modeOptions || data.modeOptions),
        rules: this.normalizeList(sceneConfig.rules || data.rules)
      }, () => {
        if (sheet) {
          this.openSheet(sheet)
        }
      })
    } catch (error) {
      this.setData({
        scenes: [],
        modeOptions: [],
        rules: []
      })
      toast.info(error.message || '分场景配置加载失败')
    }
  },

  getSceneTitle() {
    const scene = this.data.scenes.find((item) => item.key === this.data.selectedScene)
    return scene ? scene.title : ''
  },

  openSheet(sceneKey) {
    const fallback = this.data.scenes[0] && this.data.scenes[0].key

    if (!fallback) {
      toast.info('暂无可配置场景')
      return
    }

    const selectedScene = this.data.scenes.some((item) => item.key === sceneKey) ? sceneKey : fallback
    const scene = this.data.scenes.find((item) => item.key === selectedScene)

    this.setData({
      selectedScene,
      selectedMode: scene ? scene.mode : '',
      selectedSceneTitle: scene ? scene.title : '',
      showModeSheet: true
    })
  },

  handleSceneTap(event) {
    this.openSheet(event.currentTarget.dataset.key)
  },

  handleCloseSheet() {
    this.setData({
      showModeSheet: false
    })
  },

  handleModeChoice(event) {
    const { mode } = event.currentTarget.dataset
    const scenes = this.data.scenes.map((item) => (
      item.key === this.data.selectedScene ? { ...item, mode } : item
    ))

    this.setData({
      scenes,
      selectedMode: mode,
      showModeSheet: false
    })
  },

  async handleSaveTap() {
    try {
      await profileService.saveSystemBlockScenes({
        scenes: this.data.scenes
      })
      toast.success('分场景配置已保存')
    } catch (error) {
      toast.info(error.message || '分场景配置保存失败')
    }
  },

  normalizeList(list) {
    return Array.isArray(list) ? list : []
  }
})
