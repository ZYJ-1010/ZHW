#!/usr/bin/env node

const fs = require("fs");
const path = require("path");

const root = path.resolve(__dirname, "..");
const serverFile = path.join(root, "services", "go-api", "internal", "appapi", "server.go");
const openapiFiles = [
  path.join(root, "docs", "openapi", "app.openapi.yaml"),
  path.join(root, "docs", "openapi", "admin.openapi.yaml"),
];

function read(file) {
  return fs.readFileSync(file, "utf8");
}

function collectGoRoutes() {
  const source = read(serverFile);
  const routes = [];
  const matcher = /handle\("([A-Z]+)\s+([^"]+)",/g;
  let match;
  while ((match = matcher.exec(source))) {
    const method = match[1];
    const route = match[2];
    if (!route.startsWith("/api/app/") && !route.startsWith("/api/admin/")) {
      continue;
    }
    routes.push({
      method,
      route,
      dynamicPrefix: route.endsWith("/"),
    });
  }
  return routes;
}

function collectOpenAPIPaths() {
  const paths = new Set();
  for (const file of openapiFiles) {
    const source = read(file);
    for (const line of source.split(/\r?\n/)) {
      const match = line.match(/^ {2}(\/api\/[^:]+):\s*$/);
      if (match) {
        paths.add(match[1]);
      }
    }
  }
  return paths;
}

function main() {
  const routes = collectGoRoutes();
  const openapiPaths = collectOpenAPIPaths();
  const staticRoutes = routes.filter((item) => !item.dynamicPrefix);
  const dynamicPrefixes = routes.filter((item) => item.dynamicPrefix);
  const missing = staticRoutes.filter((item) => !openapiPaths.has(item.route));

  console.log(`Go static routes: ${staticRoutes.length}`);
  console.log(`Go dynamic route prefixes: ${dynamicPrefixes.length}`);
  console.log(`OpenAPI paths: ${openapiPaths.size}`);
  console.log(`Missing static routes in OpenAPI: ${missing.length}`);

  if (missing.length) {
    console.log("\nMissing static routes:");
    for (const item of missing) {
      console.log(`- ${item.method} ${item.route}`);
    }
  }

  if (process.argv.includes("--show-prefixes") && dynamicPrefixes.length) {
    console.log("\nDynamic route prefixes needing manual OpenAPI coverage:");
    for (const item of dynamicPrefixes) {
      console.log(`- ${item.method} ${item.route}`);
    }
  }
}

main();
