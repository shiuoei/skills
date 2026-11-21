# mygo-dev 技能包

MyGo 桌面应用开发技能：把 [MyGo 官方中文文档](https://mygo.run/zh/docs) 的 125 页内容提炼成按需加载的参考文件，并补充实际项目验证过的经验（原生 UI、托盘、图标、打包、签名自动更新、GitHub Actions 发布）。

## 目录结构

```
mygo-dev/
├── SKILL.md                          # 枢纽：触发条件、开工步骤、决策树、路由表、硬约束、检查清单
├── README.md                         # 本文件：安装、更新、来源
├── sources.md                        # 文档快照、125 页覆盖矩阵、版本门限
├── references/
│   ├── 01-getting-started.md         # 安装、项目结构、配置、CLI
│   ├── 02-window-app-lifecycle.md    # 窗口、应用生命周期、单实例、文稿、启动器
│   ├── 03-menus-tray-native-api.md   # 菜单、托盘、对话框、通知、剪贴板等桌面 API
│   ├── 04-ui-foundations.md          # 原生 UI 指南（视图/布局/主题/动画/测试/渲染）
│   ├── 05-ui-actions-selection.md    # 动作与选择控件
│   ├── 06-ui-inputs-editors.md       # 输入与编辑器
│   ├── 07-ui-collections-containers-nav.md # 集合、容器、导航
│   ├── 08-ui-dialogs-status-media.md # 对话框、状态、图表、媒体
│   ├── 09-web-frontend-bindings-plugins.md # Web 前端、绑定、官方插件
│   ├── 10-i18n-data-audio-native-modules.md # 多语言、数据、音频、原生模块
│   ├── 11-mobile.md                  # 手机/平板（需要 mygo ≥ v0.10.0）
│   ├── 12-distribution-updates-ci.md # 打包、签名、安装器、自动更新、CI
│   ├── 13-api-quick-reference.md     # API/组件速查
│   └── 14-lessons-experience.md      # 实战经验与踩坑
└── templates/
    ├── mygo.json                     # 原生 UI 项目配置
    ├── mygo.config.ts                # Web 前端项目配置
    ├── main-tray-app.go              # 托盘常驻应用骨架
    ├── view-page.go                  # 原生 UI 页面写法
    ├── view-page_test.go             # ui.NewTester 视图测试
    ├── updater.go                    # 签名自动更新接入
    └── release.yml                   # GitHub Actions 发布工作流
```

## 安装位置

技能是「目录 + `SKILL.md`」的通用格式；把整个 `mygo-dev/` 目录放到任意 agent 工具的技能目录即可。

| 场景 | 放置位置 |
| --- | --- |
| DSH（推荐，跨会话共享） | `~/.agents/skills/mygo-dev/`（DSH 的共享 agent 根；`DSH_AGENTS_HOME` 可覆盖 `~/.agents`） |
| DSH 用户目录 | `~/.dsh/skills/mygo-dev/` |
| DSH 项目级 | `<项目>/.agents/skills/mygo-dev/` 或 `<项目>/.dsh/skills/mygo-dev/` |
| Claude Code | `~/.claude/skills/mygo-dev/` 或 `<项目>/.claude/skills/mygo-dev/` |
| Cursor | `<项目>/.cursor/skills/mygo-dev/` |
| Codex | `~/.codex/skills/mygo-dev/` 或 `<项目>/.codex/skills/mygo-dev/` |

DSH 会监听技能目录：新增或修改后通常无需重启；如果当前会话的目录没有刷新，新开一个会话即可。同名技能以项目级/更高优先级的目录为准。

## 使用

对 agent 说「用 mygo-dev 技能帮我做/查 …」，或直接提出 MyGo 相关任务（初始化项目、原生 UI 页面、托盘、打包、自动更新等），技能会按 `SKILL.md` 的路由表加载对应参考文件。

典型任务与入口：

- 新建应用、配置 `mygo.json`、了解 CLI → `references/01-getting-started.md`
- 窗口/生命周期/托盘/菜单/对话框/通知 → `references/02`、`03`
- 原生 UI 页面与控件 → `references/04`–`08`，模板 `view-page*.go`
- Web 前端、绑定、插件 → `references/09`
- 打包、图标、签名、自动更新、CI → `references/12`、`templates/release.yml`、`templates/updater.go`
- 查 API 名 → `references/13`；踩坑 → `references/14`

## 来源、版本与许可

- 文档来源：MyGo 官方中文文档 <https://mygo.run/zh/docs>，快照日期 2026-11-21，125 页，对应 MyGo **v0.10.0** 文档。
- 每个参考文件都按页面列出源链接；逐页映射见 `sources.md`。
- 参考文件是提炼与重写，不是文档原文的逐页复制；代码示例保持 API 名称不变。
- 经验章来自公开项目的实际开发记录（MyGo v0.9.0 的 Windows 原生 UI 应用与 GitHub Actions 发布），只引用公开可复现的事实。
- MyGo 采用 MIT 许可证；本技能与 MyGo 项目无隶属关系。

## 版本与可移植性

- 本技能正文**不包含本机绝对路径、模块缓存位置或本机文档目录**；用户名只出现在经验章的公开仓库 URL 中。文档位置在运行时解析：
  - 在线：<https://mygo.run/zh/docs>
  - 本地（与安装版本一致）：`go list -m -f '{{.Dir}}' github.com/egoist/mygo` 下的 `docs/`
- 实际 API 以 `go doc github.com/egoist/mygo`（及 `.../ui`）为准；文档快照比安装版本新时，以安装版本为准，并注意 `sources.md` 的版本门限（`mobile`、`plugins/render` 整页，以及若干 v0.10 新增 API 的清单）。

## 维护与更新

1. MyGo 升级后：`go list -m github.com/egoist/mygo` 对比版本 → 复核 `sources.md` 的版本门限与 `references/14` 中受版本影响的结论。
2. 文档更新：以 <https://mygo.run/zh/docs> 为准，核对页数（当前 125），更新 `sources.md` 的映射与版本信息。
3. 修改参考文件时保持既有约定：中文直述、术语与代码名不变、每文件顶部列源页面、不要在正文写入本机路径、控制篇幅（提炼而非复制）。
