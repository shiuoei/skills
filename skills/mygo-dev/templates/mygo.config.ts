import { defineConfig } from "mygo-cli";

// Web 前端项目的配置。原生 UI 项目改用 mygo.json，并省略前端字段。
export default defineConfig({
  name: "Your App",
  slug: "your-app",
  identifier: "com.example.your-app",
  version: "0.1.0",
  copyright: "© 2026 Your Name",
  license: "MIT",

  // 开发时 mygo dev 运行的命令与页面地址。
  devUrl: "http://localhost:5173",
  devCommand: "bun run dev:web",
  buildCommand: "bun run build:web",
  frontendDist: "dist",
  bindings: "src/mygo.ts",

  // 构建产物目录。
  out: "build",

  // 额外资源与集成。
  resources: ["third_party/licenses"],
  urlSchemes: ["your-app"],
  fileAssociations: [{ ext: ["md"], name: "Markdown Document", mimeType: "text/markdown" }],

  // 消息目录与签名自动更新。
  i18n: { dir: "locales", source: "en" },
  updates: {
    publicKey: "REPLACE_WITH_CONTENT_OF_mygo-update.pub",
    github: "OWNER/REPO",
    changelog: "RELEASES.md",
  },

  // 各平台打包/签名（按需填写；不填则生成未签名产物）。
  macos: { minimumSystemVersion: "12.0" },
  linux: { maintainer: "Your Name <you@example.com>" },
});
