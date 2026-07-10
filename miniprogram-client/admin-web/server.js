const http = require("http");
const fs = require("fs");
const path = require("path");

const port = Number(process.env.ADMIN_WEB_PORT || 5173);
const publicDir = path.join(__dirname, "src");
const apiBase = process.env.ADMIN_API_BASE || "http://127.0.0.1:8080";

const server = http.createServer((req, res) => {
  if (req.url.startsWith("/api/")) {
    proxyAPI(req, res);
    return;
  }

  const filePath = req.url === "/" ? "index.html" : req.url.replace(/^\/+/, "");
  const abs = path.join(publicDir, filePath);

  if (!abs.startsWith(publicDir)) {
    res.writeHead(403);
    res.end("Forbidden");
    return;
  }

  fs.readFile(abs, (err, data) => {
    if (err) {
      res.writeHead(404);
      res.end("Not Found");
      return;
    }
    res.writeHead(200, { "Content-Type": contentType(abs) });
    res.end(data);
  });
});

server.listen(port, () => {
  console.log(`admin-web listening on http://127.0.0.1:${port}`);
});

function contentType(file) {
  if (file.endsWith(".html")) return "text/html; charset=utf-8";
  if (file.endsWith(".css")) return "text/css; charset=utf-8";
  if (file.endsWith(".js")) return "application/javascript; charset=utf-8";
  return "text/plain; charset=utf-8";
}

function proxyAPI(req, res) {
  const target = new URL(req.url, apiBase);
  const proxyReq = http.request(
    target,
    {
      method: req.method,
      headers: { ...req.headers, host: target.host },
    },
    (proxyRes) => {
      res.writeHead(proxyRes.statusCode || 502, proxyRes.headers);
      proxyRes.pipe(res);
    }
  );

  proxyReq.on("error", (error) => {
    res.writeHead(502, { "Content-Type": "application/json; charset=utf-8" });
    res.end(JSON.stringify({ code: 502, message: error.message }));
  });
  req.pipe(proxyReq);
}
