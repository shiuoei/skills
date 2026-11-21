---
name: mygo-dev
description: "MyGo（github.com/egoist/mygo）桌面应用开发技能：用 Go 构建跨平台桌面/手机应用，窗口可显示原生 UI（ui 包，无 webview）或 Web 前端；覆盖窗口/托盘/菜单/对话框/通知、mygo.json 配置、mygo dev/build、图标与资源、安装器打包、签名自动更新与 GitHub Actions 发布。当任务涉及 MyGo、egoist/mygo、mygo.json、ui 原生控件、托盘应用、桌面打包或自动更新时使用。"
metadata:
  whenToUse: "用户要在 Go 里做桌面应用、把现有桌面项目迁到 MyGo、排查 mygo build/mygo dev/图标/打包/自动更新问题，或需要查 MyGo 的窗口、托盘、菜单、原生控件、分发与更新 API 时。"
---

# MyGo 应用开发

MyGo 是纯 Go、无需 cgo 的桌面应用框架：窗口可以显示系统 webview 里的网页，也可以显示由 MyGo 用 GPU 绘制的原生 UI（`ui` 包，没有 HTML、JavaScript 或 webview）。一个应用可以同时有这两类窗口。

## 怎么用这个技能

1. 先读本文件，按「任务路由」只打开需要的参考文件，不要一次读全部。
2. 参考文件是 [MyGo 官方中文文档](https://mygo.run/zh/docs) 的提炼，页面级来源见 `sources.md`；需要细节时打开对应页面。
3. 实际 API 以**安装的版本**为准：`go doc` 与模块自带的 `docs/` 永远优先于本技能的记忆。

## 开工三件事

**① 确认版本**（文档可能比安装版本新）：

```sh
go list -m github.com/egoist/mygo
```

**② 核对 API**：

```sh
go doc github.com/egoist/mygo
go doc github.com/egoist/mygo/ui
dir="$(go list -m -f '{{.Dir}}' github.com/egoist/mygo)"   # 模块目录
ls "$dir/docs"                                             # 与安装版本一致的英文文档
```

PowerShell：`$dir = go list -m -f '{{.Dir}}' github.com/egoist/mygo; ls "$dir/docs"`。
本技能的快照对应 MyGo v0.10.0 文档（2026-11-21，125 页）。**若 `go list -m` 显示的版本更低（例如 v0.9.0），先 `go doc` 核对再照抄参考文件里的 API**：除了 `mobile` 与 `plugins/render` 两章整页需要 ≥ v0.10.0 外，正文里还夹有个别 v0.10 才新增的控件/页面 API（如 `ui.PopoverMenuOptions`、`ui.VideoSeeker`、`ListState.NearEnd`/`EndReached`、`Router.SystemStack`、`ScrollBy`、`SetThumbBar` 等；完整清单见 `sources.md` 的「版本门限」）。冲突时以 `go doc` 与模块自带 `docs/` 为准。

**③ 选形态**：原生 UI / Web 前端 / 混合（见下）。

## 决策：界面形式与应用形态

| 选择 | 适合 | 依赖 |
| --- | --- | --- |
| **Web 前端** | 富文档、已有 HTML/CSS/JS 资产、只有浏览器才能做的事 | 系统 webview：Windows WebView2、Linux WebKitGTK；前端工具用 Bun |
| **原生 UI** | 工具、设置页、检查器、托盘常驻、要求秒开的应用 | 只依赖 Go；Windows 用 Direct3D 11、macOS 用 Metal、Linux 用 GTK 3/OpenGL |
| **混合** | 同一应用里主窗口用网页、设置用原生 UI | 上述两者 |

应用形态：普通窗口 / 托盘常驻（关闭窗口只隐藏）/ 文稿型（每个文件一个窗口）/ 启动器（全局快捷键唤出浮层）。分发形态：安装器（Windows NSIS、macOS 磁盘映像、Linux deb/rpm/AppImage）、便携包、手机（≥ v0.10.0）、自动更新（GitHub Releases / 自定义 URL / S3）。

## 最小骨架

```sh
# 创建原生 UI 项目（无前端，无需 Bun）
go run github.com/egoist/mygo/cmd/mygo@latest init -template native my-app

# 创建 Web 前端项目（Vite + TypeScript）
go run github.com/egoist/mygo/cmd/mygo@latest init my-app

cd my-app
go tool mygo doctor     # 检查环境
go tool mygo dev        # 开发：改代码自动重建重启
go tool mygo build -platform windows/amd64
```

原生 UI 模板的目录：`main.go`、`main_test.go`、`go.mod`（把 CLI 作为 tool）、`mygo.json`、`resources/icon.png`（1024×1024）。手写最小原生 UI 应用：

```go
package main

import (
	"fmt"
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

type app struct{ n int }

func (a *app) view(c *ui.Context) {
	ui.Column(c).Fill().Center().Gap(12).Children(func() {
		ui.Text(c, fmt.Sprint(a.n)).FontSize(40).Bold()
		if ui.PrimaryButton(c, "Increment").Clicked() {
			a.n++
		}
	})
}

func main() {
	s := &app{}
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:   "Counter",
			Width:   320,
			Height:  240,
			Content: ui.View(s.view),
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```

视图测试不需要窗口（`ui.NewTester` 在内存里渲染并模拟输入）：

```go
func TestCounter(t *testing.T) {
	s := &app{}
	tt := ui.NewTester(s.view, 320, 240)
	if err := tt.Click("Increment"); err != nil {
		t.Fatal(err)
	}
	if s.n != 1 || !tt.HasText("1") {
		t.Fatalf("n=%d texts=%q", s.n, tt.Texts())
	}
}
```

## 硬约束（违反时通常静默出错）

| 约束 | 说明与后果 |
| --- | --- |
| **打包必须用 `mygo build`** | 只有它会写 exe 图标资源、版本资源、Manifest、安装器和更新产物；`go build` 的产物没有应用图标，`App.Version()` 为空 |
| **原生 UI 不是网页** | `ui` 包没有 DOM/HTML/JS；要网页就用 Web 窗口或 `ui.WebView` |
| **图标是方形 PNG** | 默认 `resources/icon.png`（理想 1024×1024）；旁边的 `icon-dev.png` 是开发版图标；托盘图标的 PNG 宽度决定实际显示尺寸（用 32×32） |
| **更新配置要成对** | `updates.publicKey` 必需，`github`/`url` 二选一；私钥经 `MYGO_UPDATER_PRIVATE_KEY` 或 `updates.privateKey` 提供，绝不入库 |
| **发布说明要有对应小节** | 更新说明文件里必须有 `## <version>`，否则带私钥构建直接失败 |
| **自动更新只对打包构建生效** | `mygo dev`/`go run` 没有更新；安装目录必须可写（per-user 安装） |
| **主线程** | `App.Run` 必须在 `main()` 调用；其他 API 可从任意 goroutine 调用；事件回调在主线程执行，不要阻塞 |
| **插件先注册** | `mygo.Use` 必须在 `App.Run` 之前且只调用一次（重复/过晚注册会 panic） |
| **交叉编译边界** | `CGO_ENABLED=0` 可交叉编译桌面应用；macOS 签名和磁盘映像需 Mac；Linux 原生 UI 需 GTK 3、网页窗口需 WebKitGTK；Windows 网页窗口需 WebView2 |
| **资源目录语义** | `resources/` 顶层文件进 `App.Path(PathResources)`；`<goos>-<goarch>/` 目录是平台资源；`.mygo-files` 让更新能删除旧文件 |

## 高频陷阱

完整清单见 [references/14-lessons-experience.md](references/14-lessons-experience.md)，这里先记住最容易踩的：

1. 包装 `slog.Default().Handler()` 再接回 `slog.SetDefault` → 日志死锁（进程卡死）。
2. 单实例锁身份随构建方式变化：打包构建与 `go run` 的锁不是同一个，可能同时运行。
3. 只用 `go build` → 没有图标、没有版本，任务栏/资源管理器显示默认图标。
4. 把 1024 的 PNG 直接传给托盘 → 系统按 PNG 宽度建出巨大的 HICON。
5. 托盘应用关闭窗口时销毁窗口 → 无法再显示；应 `PreventDefault()+Hide()`。
6. 托盘创建失败没有日志 → 进程在后台失联。
7. 更新构建时发布说明缺 `## <version>` → 构建失败。
8. CI 没有注入私钥 → 产物没有更新清单（静默跳过）。
9. 增量更新需要仓库里已有上一个发布版本。
10. 在非目标平台做签名/打包（如 Windows 构建 macOS dmg）。
11. 把平台专用文件放到 `resources/` 顶层而不是 `<goos>-<goarch>/`。
12. Windows GUI 进程没有控制台 → 调试日志要写文件或内存缓冲。

## 任务路由

| 任务 | 先读 |
| --- | --- |
| 初始化、`mygo.json`/`mygo.config.ts`、CLI、版本与平台要求 | [references/01-getting-started.md](references/01-getting-started.md)、`templates/mygo.json`、`templates/mygo.config.ts` |
| 窗口/页面事件、生命周期、单实例、托盘、菜单、对话框、通知、剪贴板、Shell | [references/02-window-app-lifecycle.md](references/02-window-app-lifecycle.md)、[references/03-menus-tray-native-api.md](references/03-menus-tray-native-api.md)、`templates/main-tray-app.go` |
| 原生 UI 指南（视图/布局/主题/输入/动画/辅助功能/测试/渲染） | [references/04-ui-foundations.md](references/04-ui-foundations.md) |
| 原生 UI 控件：动作、选择、输入、编辑器、集合、容器、导航、对话框、状态、媒体 | [references/05-ui-actions-selection.md](references/05-ui-actions-selection.md)、[references/06-ui-inputs-editors.md](references/06-ui-inputs-editors.md)、[references/07-ui-collections-containers-nav.md](references/07-ui-collections-containers-nav.md)、[references/08-ui-dialogs-status-media.md](references/08-ui-dialogs-status-media.md)、`templates/view-page.go` |
| Web 前端、`mygo.Bind`、生成的 TypeScript 客户端、官方插件 | [references/09-web-frontend-bindings-plugins.md](references/09-web-frontend-bindings-plugins.md) |
| 多语言、剪贴板/拖放数据、音频、原生模块（Swift/Kotlin/Rust/Zig） | [references/10-i18n-data-audio-native-modules.md](references/10-i18n-data-audio-native-modules.md) |
| 手机/平板（≥ v0.10.0） | [references/11-mobile.md](references/11-mobile.md) |
| 打包、图标/版本、签名、安装器、自动更新、GitHub Actions | [references/12-distribution-updates-ci.md](references/12-distribution-updates-ci.md)、`templates/release.yml`、`templates/updater.go` |
| 查具体 API/组件名 | [references/13-api-quick-reference.md](references/13-api-quick-reference.md) |
| 踩坑、实际项目经验、验证方法 | [references/14-lessons-experience.md](references/14-lessons-experience.md) |

## 交付前检查清单

- [ ] `go build ./...`、`go vet ./...`、`go test ./...` 通过；原生 UI 的视图用 `ui.NewTester` 测试。
- [ ] `go tool mygo doctor` 没有阻断项。
- [ ] `go tool mygo build -platform <目标>` 成功；检查产物图标与版本（Windows 看文件属性）。
- [ ] 至少对目标平台交叉编译通过（`CGO_ENABLED=0 GOOS=... GOARCH=... go build ./...`）。
- [ ] 托盘/关闭行为/单实例/退出清理符合预期。
- [ ] 自动更新：`updates` 配置、发布说明小节、构建产物含更新清单；发布后用旧版本实机验证一次。
- [ ] 同步更新项目内的 `README`、变更记录和发布说明。

## 版本与维护

- 文档快照、来源与覆盖见 [sources.md](sources.md)；上游文档为 <https://mygo.run/zh/docs>。
- MyGo 升级后：重新核对 `sources.md` 的版本门限，以及 [references/14-lessons-experience.md](references/14-lessons-experience.md) 中可能随版本变化的结论。
- 本技能是社区经验整理，与 MyGo 项目无隶属关系；MyGo 采用 MIT 许可证。

## 使用示例

参见 [典型任务与预期结果](examples/basic.md)。
