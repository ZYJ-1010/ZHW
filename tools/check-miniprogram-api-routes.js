const fs = require('fs')
const path = require('path')

const root = path.resolve(__dirname, '..')
const miniRoot = path.join(root, 'miniprogram-client')
const goRoot = path.join(root, 'services', 'go-api')

const REQUEST_RE = /request\.(get|post|put|delete|patch)\(\s*(`([^`]+)`|'([^']+)'|"([^"]+)")/g
const HANDLE_RE = /handle\(\s*"([A-Z]+)\s+([^"]+)"/g

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

function normalizeFrontendPath(raw) {
  return String(raw || '')
    .replace(/\$\{[^}]+\}/g, ':id')
    .replace(/\/+/g, '/')
}

function routeToPattern(route) {
  const escaped = route
    .replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    .replace(/\\:id/g, '[^/]+')

  return new RegExp(`^${escaped}$`)
}

function loadFrontendRequests() {
  const files = walk(path.join(miniRoot, 'api'))
    .concat(walk(path.join(miniRoot, 'services')))
    .filter((file) => file.endsWith('.js'))
  const requests = []

  for (const file of files) {
    const source = read(file)
    let match

    while ((match = REQUEST_RE.exec(source)) !== null) {
      const method = match[1].toUpperCase()
      const rawPath = match[3] || match[4] || match[5] || ''

      if (!rawPath.startsWith('/api/')) {
        continue
      }

      requests.push({
        method,
        path: normalizeFrontendPath(rawPath),
        file: path.relative(root, file)
      })
    }
  }

  return requests
}

function loadBackendRoutes() {
  const files = walk(path.join(goRoot, 'internal', 'appapi'))
    .concat(walk(path.join(goRoot, 'cmd', 'server')))
    .filter((file) => file.endsWith('.go'))
  const exact = []
  const prefixes = []

  for (const file of files) {
    const source = read(file)
    let match

    while ((match = HANDLE_RE.exec(source)) !== null) {
      const method = match[1].toUpperCase()
      const route = match[2]
      const item = { method, route, file: path.relative(root, file) }

      if (route.endsWith('/')) {
        prefixes.push(item)
      } else {
        exact.push(Object.assign(item, { pattern: routeToPattern(route) }))
      }
    }
  }

  return { exact, prefixes }
}

function isCovered(request, routes) {
  if (routes.exact.some((route) => route.method === request.method && route.pattern.test(request.path))) {
    return true
  }

  return routes.prefixes.some((route) => (
    route.method === request.method
      && request.path.startsWith(route.route)
  ))
}

const frontendRequests = loadFrontendRequests()
const backendRoutes = loadBackendRoutes()
const missing = frontendRequests.filter((request) => !isCovered(request, backendRoutes))

console.log(`frontend_api_calls=${frontendRequests.length}`)
console.log(`missing_backend_routes=${missing.length}`)

for (const item of missing) {
  console.log(`${item.method} ${item.path} <- ${item.file}`)
}

if (missing.length > 0) {
  process.exitCode = 1
}
