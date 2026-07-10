#!/usr/bin/env node

const fs = require("fs");
const path = require("path");

const root = path.resolve(__dirname, "..");
const files = [
  path.join(root, "db", "seeds", "system_configs.sql"),
  path.join(root, "db", "seeds", "user_system_management_configs.sql"),
  path.join(root, "db", "seeds", "invite_codes.sql"),
  path.join(root, "db", "seeds", "admin_roles_permissions.sql"),
];

const required = [
  ["system_configs.sql", "home.display_config"],
  ["system_configs.sql", "game.category_config"],
  ["system_configs.sql", "game.application_config"],
  ["system_configs.sql", "game.condition_rule_config"],
  ["system_configs.sql", "game.profit_template_config"],
  ["system_configs.sql", "game.cancel_config"],
  ["system_configs.sql", "map.index_config"],
  ["system_configs.sql", "map.play_pages_config"],
  ["system_configs.sql", "message.center_config"],
  ["system_configs.sql", "report.center_config"],
  ["system_configs.sql", "review.complete_config"],
  ["system_configs.sql", "review.page_config"],
  ["system_configs.sql", "role.expert_apply_config"],
  ["system_configs.sql", "role.application_page_config"],
  ["system_configs.sql", "role.status_page_config"],
  ["user_system_management_configs.sql", "feedback-home"],
  ["user_system_management_configs.sql", "feedback-records"],
  ["user_system_management_configs.sql", "profile-info"],
  ["user_system_management_configs.sql", "skill-config"],
  ["invite_codes.sql", "ENJOY2026"],
  ["admin_roles_permissions.sql", "super_admin"],
  ["admin_roles_permissions.sql", "operation_log:view_self"],
  ["admin_roles_permissions.sql", "operation_log:view_full"],
  ["admin_roles_permissions.sql", "game:create_admin"],
  ["admin_roles_permissions.sql", "redemption:manage"],
  ["admin_roles_permissions.sql", "report:handle"],
];

function readSeed(file) {
  return fs.readFileSync(file, "utf8");
}

function main() {
  const contents = new Map();
  for (const file of files) {
    contents.set(path.basename(file), readSeed(file));
  }

  const missing = [];
  for (const [fileName, token] of required) {
    const source = contents.get(fileName) || "";
    if (!source.includes(token)) {
      missing.push(`${fileName}: ${token}`);
    }
  }

  console.log(`Checked seed/config tokens: ${required.length}`);
  console.log(`Missing tokens: ${missing.length}`);
  if (missing.length) {
    for (const item of missing) {
      console.log(`- ${item}`);
    }
    process.exitCode = 1;
  }
}

main();
