#!/usr/bin/env node

/**
 * 从墨刀 HTML 导出包中读取 flat-tree 数据，导出全部画布及其元素的 PT 标注。
 * 仅处理本地导出文件，不请求网络，也不写入业务代码目录。
 */
const fs = require('fs')
const path = require('path')
const zlib = require('zlib')
const crypto = require('crypto')

const defaultSource = '/Users/yancey_chou/Desktop/pm2mqf45v78zqi7hs-mqf45uoo'
const source = process.argv[2] || defaultSource
const outDir = process.argv[3] || path.resolve(__dirname, '../docs/prototype-pt-20260719')
const dataFile = path.join(source, 'extra/data.1.js')
const powers86 = [1, 86, 7396, 636056, 54700816, 4704270176, 404567235136, 34792782221696, 0xaa15f068e6100]

function decode86(value) {
  let result = 0
  for (let index = 0; index < value.length; index += 1) {
    const code = value.charCodeAt(index)
    result += (code <= 91 ? code - 40 : code - 41) * powers86[value.length - index - 1]
  }
  return result
}

function decodePointPair(value) {
  if (typeof value !== 'string' || value.length !== 8) return null
  return {
    x: (decode86(value.slice(0, 4)) - 10000000) / 100,
    y: (decode86(value.slice(4, 8)) - 10000000) / 100
  }
}

function parseTree(text) {
  const sourceMatch = text.match(/window\["hzv5"\]\["flpk"\]\s*=\s*\[\s*\[\s*\d+\s*,\s*\d+\s*,\s*"([^"]+)"/)
  if (!sourceMatch) throw new Error('未找到墨刀 flat-tree 数据')
  const lines = zlib.gunzipSync(Buffer.from(sourceMatch[1], 'base64')).toString('utf8').split('\n')
  let cursor = 0

  function readNode() {
    const line = lines[cursor++]
    if (!line) throw new Error(`flat-tree 在第 ${cursor} 行意外结束`)
    const tabIndex = line.indexOf('\t')
    const header = line.slice(0, tabIndex)
    const spaceIndex = header.indexOf(' ')
    const key = spaceIndex < 0 ? header : header.slice(0, spaceIndex)
    const childrenCount = spaceIndex < 0 ? 0 : decode86(header.slice(spaceIndex + 1))
    const node = { key, raw: JSON.parse(line.slice(tabIndex + 1)), children: [] }
    for (let index = 0; index < childrenCount; index += 1) node.children.push(readNode())
    return node
  }

  const root = readNode()
  if (cursor !== lines.length) throw new Error(`flat-tree 未完全解析：${cursor}/${lines.length}`)
  return root
}

function compactNode(node, parentKey) {
  const style = { ...node.raw }
  const iconSvg = style.icP || null
  const iconViewBox = style.icVB || null
  delete style.icP
  delete style.icVB
  const point = decodePointPair(style.xy)
  const size = decodePointPair(style.wh)
  return {
    key: node.key,
    parentKey,
    name: style.N || '',
    type: style.T || '',
    pt: point && size ? { x: point.x, y: point.y, width: size.x, height: size.y } : null,
    textRuns: style['b/#000000'] || null,
    imageRef: style.imgR || null,
    iconViewBox,
    iconSvg,
    style
  }
}

function main() {
  const tree = parseTree(fs.readFileSync(dataFile, 'utf8'))
  const screens = []
  const icons = new Map()
  const imageReferences = new Map()

  function collect(node, ancestors = []) {
    const nextAncestors = [...ancestors, node.raw.N || node.key]
    if (node.raw.ic === 'page-0') {
      const elements = []
      function collectElements(item, parentKey) {
        // 子画布由自身条目完整记录，避免把同一界面重复收录到父画布。
        if (item.raw.ic === 'page-0') return
        const itemData = compactNode(item, parentKey)
        elements.push(itemData)
        if (itemData.imageRef) {
          const reference = imageReferences.get(itemData.imageRef) || { ref: itemData.imageRef, count: 0, usage: [] }
          reference.count += 1
          reference.usage.push({ screenId: node.key, screenName: node.raw.N || node.key, elementKey: itemData.key, elementName: itemData.name })
          imageReferences.set(itemData.imageRef, reference)
        }
        if (itemData.iconSvg) icons.set(itemData.key, { key: itemData.key, name: itemData.name, viewBox: itemData.iconViewBox, svg: itemData.iconSvg })
        item.children.forEach((child) => collectElements(child, item.key))
      }
      node.children.forEach((child) => collectElements(child, node.key))
      screens.push({
        id: node.key,
        name: node.raw.N || node.key,
        path: ancestors.filter((item) => !['@@R', '@@M', 'B@main', 'B@trash', '@@T'].includes(item)),
        canvas: compactNode(node, null),
        elements
      })
    }
    node.children.forEach((child) => collect(child, nextAncestors))
  }
  collect(tree)

  fs.mkdirSync(outDir, { recursive: true })
  const iconDir = path.join(outDir, 'reusable-svg-icons')
  fs.mkdirSync(iconDir, { recursive: true })
  const uniqueIcons = new Map()
  for (const icon of icons.values()) {
    const hash = crypto.createHash('sha1').update(`${icon.viewBox || ''}\n${icon.svg}`).digest('hex').slice(0, 12)
    if (!uniqueIcons.has(hash)) {
      const file = `icon-${hash}.svg`
      const viewBox = icon.viewBox ? ` viewBox="${icon.viewBox}"` : ''
      const normalizedSvg = icon.svg.replace(/[ \t]+$/gm, '')
      fs.writeFileSync(path.join(iconDir, file), `<svg xmlns="http://www.w3.org/2000/svg"${viewBox}>${normalizedSvg}</svg>\n`)
      uniqueIcons.set(hash, { file: `reusable-svg-icons/${file}`, viewBox: icon.viewBox || '', sourceName: icon.name || '', sourceKey: icon.key })
    }
  }
  const inventory = {
    source: { path: source, canvasWidthPt: 375, canvasHeightPt: 812, exportedAt: '2026-07-19' },
    screensCount: screens.length,
    screens
  }
  fs.writeFileSync(path.join(outDir, 'all-screens-pt.json'), JSON.stringify(inventory, null, 2))
  fs.writeFileSync(path.join(outDir, 'reusable-svg-icons.json'), JSON.stringify({
    totalInstances: icons.size,
    uniqueFiles: [...uniqueIcons.values()]
  }, null, 2))
  fs.writeFileSync(path.join(outDir, 'image-reference-manifest.json'), JSON.stringify({
    sourceAssetDirectories: ['uploads6', 'uploads7', 'res-img'],
    note: '只记录图片引用关系；头像、用户图片、榜单和局封面等演示数据不复制到业务项目。',
    images: [...imageReferences.values()]
  }, null, 2))
  const lines = [
    '# 墨刀原型 PT 标注总目录',
    '',
    `- 原型：小程序开发（对外）副本`,
    `- 画布基准：375 × 812pt`,
    `- 已解析页面画布：${screens.length} 个（与导出包记录的 106 个画布相比，另有一个无页面标识的元节点）`,
    `- 完整元素 PT、文字样式、原始填充/边框/阴影、图片和 SVG 引用：\`all-screens-pt.json\``,
    `- 可复用 SVG 图标：\`reusable-svg-icons/\`（${uniqueIcons.size} 个去重文件）及 \`reusable-svg-icons.json\``,
    `- 图片引用清单：\`image-reference-manifest.json\`；演示头像、封面和业务数据不复制。`,
    '',
    '## 使用规则',
    '',
    '- 只复刻布局、样式和交互结构；用户、榜单、金额、头像、局内容等演示数据不进入业务前端。',
    '- 页面坐标和尺寸均为 PT；转换至 750rpx 小程序宽度时，数值乘以 2。',
    '- 系统状态栏和设备壳仅为原型环境，不直接实现。',
    '',
    '## 页面清单',
    ''
  ]
  screens.forEach((screen, index) => lines.push(`${index + 1}. ${screen.path.join(' / ') || '顶层'} / ${screen.name}（\`${screen.id}\`；元素 ${screen.elements.length}）`))
  fs.writeFileSync(path.join(outDir, 'README.md'), `${lines.join('\n')}\n`)
}

main()
