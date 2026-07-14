const fs = require('fs')
const path = require('path')

const root = path.resolve(__dirname, '..')
const appJsonPath = path.join(root, 'miniprogram-client', 'app.json')
const routesPath = path.join(root, 'miniprogram-client', 'config', 'routes.js')

const appJson = JSON.parse(fs.readFileSync(appJsonPath, 'utf8'))
const registeredPages = new Set(appJson.pages || [])
for (const pkg of appJson.subPackages || appJson.subpackages || []) {
  const root = String(pkg.root || '').replace(/^\/+|\/+$/g, '')
  for (const page of pkg.pages || []) {
    registeredPages.add([root, page].filter(Boolean).join('/'))
  }
}
const { ROUTES } = require(routesPath)

const missing = Object.entries(ROUTES)
  .map(([key, route]) => [key, String(route || '').replace(/^\/+/, '').split('?')[0]])
  .filter(([, route]) => route && !registeredPages.has(route))

console.log(`configured_routes=${Object.keys(ROUTES).length}`)
console.log(`missing_registered_pages=${missing.length}`)

for (const [key, route] of missing) {
  console.log(`${key}: ${route}`)
}

if (missing.length > 0) {
  process.exit(1)
}
