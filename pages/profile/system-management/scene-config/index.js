const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const ASSET_BASE = '/pages/profile/system-management/block-settings/assets'

const MODE_MAP = {
  hard: {
    label: '硬保护',
    desc: '完全过滤同类行家内容'
  },
  soft: {
    label: '软保护',
    desc: '降权排序，推至第5页后'
  },
  none: {
    label: '不过滤',
    desc: '正常展示同类行家内容'
  }
}

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
    blockSettings: null,
    scenes: [],
    modeOptions: [],
    rules: []
  },

  onLoad(options) {
    if (options && options.sheet) {
      this.openSheet(options.sheet)
    }
    this.loadScenes()
  },

  async loadScenes() {
    try {
      const data = await profileService.getSystemBlockSettings()
      const modes = Array.isArray(data.protectionModes) ? data.protectionModes : []
      this.setData({
        blockSettings: data || null,
        scenes: Array.isArray(data.scenes) ? data.scenes : [],
        modeOptions: modes.map(toModeOption),
        rules: Array.isArray(data.sceneRules) ? data.sceneRules : []
      })
    } catch (error) {
      this.setData({ scenes: [], modeOptions: [], rules: [] })
      toast.info(error.message || '分场景配置暂时不可用')
    }
  },

  getSceneTitle() {
    const scene = this.data.scenes.find((item) => item.key === this.data.selectedScene)
    return scene ? scene.title : ''
  },

  openSheet(sceneKey) {
    if (!this.data.scenes.length) {
      return
    }
    const fallback = this.data.scenes[0].key
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
      await profileService.saveSystemBlockSettings({
        ...(this.data.blockSettings || {}),
        scenes: this.data.scenes
      })
      toast.success('分场景配置已保存')
    } catch (error) {
      toast.info(error.message || '保存失败')
    }
  },

  modeLabel(mode) {
    return MODE_MAP[mode] ? MODE_MAP[mode].label : ''
  }
})

function toModeOption(item) {
  return {
    key: item.key,
    title: item.title,
    desc: item.desc,
    iconText: item.key === 'none' ? '!' : '锁'
  }
}
