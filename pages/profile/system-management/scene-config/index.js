const toast = require('../../../../utils/toast')

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
    scenes: [
      { key: 'recommend', title: '推荐列表', desc: '首页/发现页推荐', mode: 'hard' },
      { key: 'search', title: '搜索结果', desc: '搜索同类行家过滤', mode: 'hard' },
      { key: 'nearby', title: '附近推荐', desc: 'LBS地理位置推荐', mode: 'hard' },
      { key: 'list', title: '列表浏览', desc: '局列表/行家列表', mode: 'none' }
    ],
    modeOptions: [
      { key: 'hard', title: '硬保护', desc: '完全过滤同类行家内容', iconText: '🔒' },
      { key: 'soft', title: '软保护', desc: '降权排序，推至第5页后', iconText: '🔒' },
      { key: 'none', title: '不过滤', desc: '正常展示同类行家内容', iconText: '🚫' }
    ],
    rules: [
      { strong: '硬保护', suffix: ' = 完全过滤' },
      { strong: '软保护', suffix: ' = 降权至第5页后' },
      { strong: '不过滤', suffix: ' = 正常展示' }
    ]
  },

  onLoad(options) {
    if (options && options.sheet) {
      this.openSheet(options.sheet)
    }
  },

  getSceneTitle() {
    const scene = this.data.scenes.find((item) => item.key === this.data.selectedScene)
    return scene ? scene.title : ''
  },

  openSheet(sceneKey) {
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

  handleSaveTap() {
    toast.success('分场景配置已保存')
  },

  modeLabel(mode) {
    return MODE_MAP[mode] ? MODE_MAP[mode].label : ''
  }
})
