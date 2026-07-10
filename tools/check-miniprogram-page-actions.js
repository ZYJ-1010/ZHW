const fs = require('fs')
const path = require('path')

const root = path.resolve(__dirname, '..')
const miniRoot = path.join(root, 'miniprogram-client')

const EVENT_RE = /\b(?:bind|catch|mut-bind):?([a-zA-Z]+)?\s*=\s*["']([A-Za-z_$][\w$]*)["']/g
const ROUTE_RE = /(?:url|route|path)\s*:\s*['"`](\/?pages\/[^'"`?]+(?:\/index)?)(?:\?[^'"`]*)?['"`]/g

function read(file) {
  return fs.readFileSync(file, 'utf8')
}

function walk(dir, files = []) {
  if (!fs.existsSync(dir)) {
    return files
  }

  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name)

    if (entry.isDirectory()) {
      walk(full, files)
    } else {
      files.push(full)
    }
  }

  return files
}

function loadAppPages() {
  const app = JSON.parse(read(path.join(miniRoot, 'app.json')))
  const pages = new Set(app.pages || [])

  for (const subPackage of app.subPackages || app.subpackages || []) {
    const subRoot = String(subPackage.root || '').replace(/\/+$/, '')

    for (const page of subPackage.pages || []) {
      pages.add(`${subRoot}/${page}`.replace(/^\/+/, ''))
    }
  }

  return pages
}

function hasHandler(jsSource, handler) {
  const escaped = handler.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const patterns = [
    new RegExp(`\\b${escaped}\\s*\\(`),
    new RegExp(`\\b${escaped}\\s*:`),
    new RegExp(`\\b${escaped}\\s*=\\s*function\\b`)
  ]

  return patterns.some((pattern) => pattern.test(jsSource))
}

function checkEvents() {
  const missing = []
  const wxmlFiles = walk(path.join(miniRoot, 'pages'))
    .concat(walk(path.join(miniRoot, 'components')))
    .filter((file) => file.endsWith('.wxml'))

  for (const wxmlFile of wxmlFiles) {
    const dir = path.dirname(wxmlFile)
    const jsFile = path.join(dir, `${path.basename(wxmlFile, '.wxml')}.js`)

    if (!fs.existsSync(jsFile)) {
      continue
    }

    const wxml = read(wxmlFile)
    const js = read(jsFile)
    const seen = new Set()
    let match

    while ((match = EVENT_RE.exec(wxml)) !== null) {
      const handler = match[2]

      if (seen.has(handler)) {
        continue
      }

      seen.add(handler)

      if (!hasHandler(js, handler)) {
        missing.push({
          file: path.relative(root, wxmlFile),
          handler
        })
      }
    }
  }

  return missing
}

function normalizeRoute(route) {
  return String(route || '')
    .split('${')[0]
    .replace(/^\/+/, '')
    .replace(/\/$/, '')
}

function looksLikePageRoute(route) {
  return !/\.(?:png|jpe?g|gif|webp|svg|json|wxss|js)$/i.test(route)
    && !route.includes('/assets')
    && route.split('/').length >= 3
}

function checkRoutes() {
  const appPages = loadAppPages()
  const misses = []
  const files = walk(path.join(miniRoot, 'pages'))
    .concat(walk(path.join(miniRoot, 'components')))
    .concat(walk(path.join(miniRoot, 'utils')))
    .concat(walk(path.join(miniRoot, 'config')))
    .filter((file) => file.endsWith('.js'))

  for (const file of files) {
    const source = read(file)
    let match

    while ((match = ROUTE_RE.exec(source)) !== null) {
      const route = normalizeRoute(match[1])

      if (!looksLikePageRoute(route)) {
        continue
      }

      if (!appPages.has(route)) {
        misses.push({
          file: path.relative(root, file),
          route
        })
      }
    }
  }

  return misses
}

const missingHandlers = checkEvents()
const missingRoutes = checkRoutes()

console.log(`missing_event_handlers=${missingHandlers.length}`)
for (const item of missingHandlers) {
  console.log(`${item.file} -> ${item.handler}`)
}

console.log(`missing_registered_routes=${missingRoutes.length}`)
for (const item of missingRoutes) {
  console.log(`${item.file} -> ${item.route}`)
}

if (missingHandlers.length > 0 || missingRoutes.length > 0) {
  process.exitCode = 1
}
