const fs = require("fs");
const path = require("path");

try {
  const root = path.join(__dirname, "..");
  const dist = path.join(root, "dist");
  fs.mkdirSync(dist, { recursive: true });

  for (const name of ["index.html", "styles.css", "main.js"]) {
    fs.copyFileSync(path.join(root, "src", name), path.join(dist, name));
  }

  console.log("admin-web build complete");
} catch (error) {
  console.error(error && error.stack ? error.stack : error);
  process.exit(1);
}
