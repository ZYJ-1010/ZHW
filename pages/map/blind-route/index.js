const { ROUTES } = require('../../../config/routes')
const mapService = require('../../../services/map')
const toast = require('../../../utils/toast')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')

const EMPTY_PAGE = {
  title: '',
  description: '',
  sectionTitle: '',
  selectToast: '',
  cards: [],
  recentRoutes: []
}

function applyTemplate(template, values = {}) {
  let text = String(template || '')

  Object.keys(values).forEach((key) => {
    text = text.replace(new RegExp(`\\{${key}\\}`, 'g'), String(values[key]))
  })

  return text
}

function markCards(cards, selectedId) {
  return (cards || []).map((item) => ({
    ...item,
    selected: item.id === selectedId,
    className: `blind-route-card tone-${item.tone}${item.id === selectedId ? ' selected' : ''}`
  }))
}

Page({
  data: {
    onlineText: '在线',
    mode: 'blindRout',
    pageConfig: EMPTY_PAGE,
    navItems: [
      { name: '我的', key: 'mine' },
      { name: '元宇宙', key: 'metaverse' },
      { name: '地图', key: 'map' },
      { name: '消息', key: 'message' },
      { name: '首页', key: 'home' }
    ],
    routeCards: [],
    recentRoutes: []
  },

  onLoad(options = {}) {
    this.setData({
      mode: options.mode || 'blindRout'
    })
    this.loadPageConfig()
  },

  async loadPageConfig() {
    try {
      const pageConfig = Object.assign({}, EMPTY_PAGE, await mapService.getPlayPage('blind-route'))
      const firstCard = pageConfig.cards && pageConfig.cards[0]

      this.setData({
        pageConfig,
        routeCards: markCards(pageConfig.cards, firstCard && firstCard.id),
        recentRoutes: pageConfig.recentRoutes || []
      })
    } catch (error) {
      toast.info(error.message || '盲盒路线加载失败')
      this.setData({
        pageConfig: EMPTY_PAGE,
        routeCards: [],
        recentRoutes: []
      })
    }
  },

  handleRouteTap(event) {
    const id = event.currentTarget.dataset.id
    const cards = this.data.pageConfig.cards || []
    const selected = cards.find((item) => item.id === id) || cards[0]

    if (!selected) {
      return
    }

    this.setData({
      routeCards: markCards(cards, selected.id)
    })
    wx.showModal({
      title: selected.title,
      content: [
        selected.desc,
        (selected.tags || []).join(' / '),
        '开启后会写入最近路线，可继续完成路线状态。'
      ].filter(Boolean).join('\n'),
      confirmText: '开启路线',
      cancelText: '先看看',
      success: async (res) => {
        if (!res.confirm) {
          toast.info(applyTemplate(this.data.pageConfig.selectToast, { title: selected.title }))
          return
        }

        try {
          const openedRoute = await mapService.createBlindRoute({
            cardId: selected.id,
            title: selected.title
          })
          this.setData({
            recentRoutes: [openedRoute].concat(this.data.recentRoutes || [])
          })
          toast.info('路线已开启')
        } catch (error) {
          toast.info(error.message || '开启路线失败')
        }
      }
    })
  },

  handleRecentRouteTap(event) {
    const id = event.currentTarget.dataset.id
    const route = (this.data.recentRoutes || []).find((item) => String(item.id || item.title) === String(id))

    if (!route) {
      return
    }

    wx.showModal({
      title: route.title,
      content: `当前状态：${route.statusText || '进行中'}\n${route.timeText || ''}`,
      confirmText: route.status === 'completed' ? '查看地图' : '标记完成',
      cancelText: '关闭',
      success: (res) => {
        if (!res.confirm) {
          return
        }

        if (route.status === 'completed') {
          navigateShellRoute(ROUTES.map, {
            currentRoute: ROUTES.mapBlindRoute
          })
          return
        }

        mapService.completeBlindRoute(route.id || id)
          .then((completedRoute) => {
            this.setData({
              recentRoutes: (this.data.recentRoutes || []).map((item) => {
                if (String(item.id || item.title) !== String(id)) {
                  return item
                }

                return Object.assign({}, item, completedRoute)
              })
            })
            toast.info('路线已完成')
          })
          .catch((error) => {
            toast.info(error.message || '完成路线失败')
          })
      }
    })
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}
    navigateShellKey(key, {
      currentRoute: ROUTES.mapBlindRoute
    })
  }
})
