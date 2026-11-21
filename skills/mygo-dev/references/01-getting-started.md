# MyGo 快速上手与配置

MyGo 让开发者用 Go 构建桌面应用：窗口可显示系统 webview 中的 Web 前端，也可显示 MyGo 自绘的原生 UI；本页提炼安装、创建项目、开发、构建、配置与 CLI 的可操作事实。

## 源页面

- [MyGo 文档](https://mygo.run/zh/docs)(README.md)
- [快速开始](https://mygo.run/zh/docs/getting-started)(getting-started.md)
- [配置](https://mygo.run/zh/docs/configuration)(configuration.md)
- [mygo CLI](https://mygo.run/zh/docs/cli)(cli.md)

## 两种窗口与运行要求

**用途**：决定项目使用 Web 前端还是原生 UI；同一个应用也可以混用两类窗口。

**关键 API**
- Web 窗口：`mygo.NewWindow(mygo.WindowOptions{Title, URL: "/"})`，页面运行在 macOS 的 WKWebView、Linux 的 WebKitGTK 或 Windows 的 WebView2 中。
- 原生 UI 窗口：`mygo.WindowOptions{Content: ui.View(fn)}`，用 Go 和 `ui` 包按应用状态逐帧构建，没有 HTML/JavaScript，也没有 webview。
- 两类窗口共用窗口选项、事件、菜单与桌面 API；`mygo.WebViewAvailable()` 可在 `App.Run` 前检查系统 webview。
- 应用是单个几 MB 的 Go 程序，不使用 cgo，可在任意机器交叉编译。

**最小示例**
```go
mygo.App.WhenReady(func() {
	mygo.NewWindow(mygo.WindowOptions{Title: "Hello", URL: "/"})
})
if err := mygo.App.Run(); err != nil {
	log.Fatal(err)
}
```

**限制/常见错误**
- Web 窗口依赖系统 webview：Windows 10 需 WebView2 Runtime（Windows 11 已包含）；Linux 需 GTK 3 与 WebKitGTK 4.1（或 4.0），Fedora 包名为 `webkit2gtk4.1`；macOS 12 起无需安装。
- 纯原生 UI 应用既不需要 WebKitGTK，也不需要 WebView2；Linux 上仍需要 GTK 3，托盘图标还需要 `libayatana-appindicator3`。

## 环境与安装

**用途**：准备工具链并自检开发环境。

**关键 API**
- Go 1.27 或更高版本；带 Web 前端的项目另需 Bun，原生 UI 项目只需 Go；`mygo-cli` 也可用 npm 安装。
- CLI 获取方式：`go run github.com/egoist/mygo/cmd/mygo@latest`、`bunx mygo-cli` / `npx mygo-cli`、`go install github.com/egoist/mygo/cmd/mygo@latest`。
- 原生 UI 项目的 `go.mod` 把 CLI 作为工具，`go tool mygo` 按 `go.mod` 固定版本运行。
- `mygo doctor` 检查 Go、Bun、平台 webview、NSIS、签名工具、`gh` 等。

**最小示例**
```sh
go install github.com/egoist/mygo/cmd/mygo@latest
mygo doctor
```

**限制/常见错误**
- 构建应用始终需要 Go，因为应用由 Go 编译；CLI 的 npm 包名是 `mygo-cli`，npm 上另有一个无关的 `mygo` 包，未安装 `mygo-cli` 时 `bunx mygo` 会运行那个包。

## 创建项目

**用途**：生成 Web 前端或原生 UI 模板。

**关键 API**
- `mygo init [flags] <dir>`：默认 `-template web`，生成并排的 Go 模块与 Vite/TypeScript 前端。
- `mygo init -template native <dir>`：只生成 Go 模块，窗口显示原生 UI。
- 标志：`-name`（默认目录名）、`-module`（Go 模块路径，默认目录名）、`-mygo`（用本地 MyGo 检出，通过 `replace` 指令替代已发布模块）。

**最小示例**
```sh
go run github.com/egoist/mygo/cmd/mygo@latest init my-app
mygo init -template native my-native-app
```

**限制/常见错误**
- Web 模板含 `main.go`、`mygo.config.ts`、`package.json`、`index.html`、`src/`、`resources/icon.png`；产物写入 `dist/`、`build/`、`.mygo/`，`.gitignore` 已排除。
- 原生模板含 `main.go`、`main_test.go`、`go.mod`、`mygo.json` 与图标；两个模板都附带 `.agents/skills/mygo-maintenance/` 技能，讲解元素生命周期、稳定标识、线程与测试。

## 开发循环

**用途**：改代码时持续重建并运行开发版应用。

**关键 API**
- `mygo dev [flags] [dir]`：写出 TypeScript 客户端，运行配置中的 `devCommand`，等 `devUrl` 有响应后构建并启动开发版；应用加载 `devUrl`，无 `devUrl` 时从磁盘提供 `frontendDist`。
- 标志：`-skip-dev-command`、`-sign`（macOS 签名身份）、`-lang`（如 `de`、`zh-Hant`、伪区域 `en-XA`/`ar-XB`）、`-background`、`-web`、`-port`（配合 `-web`，默认 5174）、`-open`（配合 `-web`）。
- 终端按键：`r` 重建并重启、`c` 清屏、`q` 退出、`h` 帮助；`NO_COLOR` 关闭颜色。
- Web 项目用 `bun run dev`，原生 UI 项目用 `go tool mygo dev`。

**最小示例**
```sh
cd my-app
bun run dev
```

**限制/常见错误**
- 改动 Go 代码、配置、图标、资源或消息目录会重建并重启；编译失败时保留上一版运行。窗口在原位置、原大小重开。
- 开发版名为 `<name> Dev`，标识符为 `<identifier>.dev`，数据、偏好设置与单实例锁和已安装应用隔离。

## 从前端调用 Go

**用途**：把 Go 值暴露为带类型的 TypeScript 客户端。

**关键 API**
- `mygo.Bind(value)` 绑定导出的方法；`mygo.NewEvent[T]("name")` 声明可订阅事件。
- `mygo generate [-o file] [dir]` 写出客户端到配置的 `bindings`（默认随项目结构推导），`mygo dev` 与 `mygo build` 会代跑。
- 页面侧 `import { Greeter, events } from "./mygo"`；方法返回 Promise，事件用 `events.tick.on(...)`。

**最小示例**
```go
type Greeter struct{}

func (Greeter) Greet(name string) string { return "Hello, " + name }

func main() {
	mygo.Bind(Greeter{})
	mygo.App.WhenReady(func() { mygo.NewWindow(mygo.WindowOptions{Title: "Hello", URL: "/"}) })
	_ = mygo.App.Run()
}
```

**限制/常见错误**
- 没有前端的应用不会生成客户端；方法可接收、返回的类型以及错误、事件与安全语义见[从前端调用 Go](https://mygo.run/zh/docs/bindings)。
- 生成文件由 CLI 在接口变化后重新生成。

## 原生 UI 开发与视图测试

**用途**：用 Go 写界面，并用普通单元测试驱动视图。

**关键 API**
- 视图是 `func(*ui.Context)`，MyGo 每帧按应用状态重建；交互通过查询上一帧以来的元素状态获得。
- 常用构造：`ui.Column(c).Fill().Center().Gap(12)`、`Children(func(){...})`、`ui.Text(c, s)`、`ui.TextInput(c, &s)`、`ui.PrimaryButton(c, "…")`、`.Clicked()`。
- `go test` 可在不开窗口的情况下点击、输入视图。

**最小示例**
```go
func (a *app) view(c *ui.Context) {
	ui.Column(c).Fill().Center().Gap(16).Children(func() {
		ui.TextInput(c, &a.name).Placeholder("Your name")
		if ui.PrimaryButton(c, "Greet").Clicked() {
			a.greeting = greet(a.name)
		}
		ui.Text(c, a.greeting)
	})
}
```

**限制/常见错误**
- 视图元素的生命周期、稳定标识与线程约束由 `mygo-maintenance` 技能说明；从其他 goroutine 修改状态需按原生 UI 文档处理。

## 构建与打包

**用途**：生成当前或指定平台的可交付产物。

**关键 API**
- `mygo build [flags] [dir]`：运行 `buildCommand`，编译应用并嵌入前端，输出到 `<out>/<os>-<arch>`（`out` 默认 `dist`，新项目常设为 `build`）。
- `-platform`：逗号分隔的 `GOOS/GOARCH`，如 `darwin/universal,windows/amd64,linux/amd64`；`web`（或 `js/wasm`）构建浏览器静态站点。
- 其他标志：`-debug`、`-o`、`-sign`、`-skip-build-command`、`-skip-dmg`、`-skip-notarize`、`-skip-installer`、`-upload`。
- 构建参数：`MYGO_LDFLAGS`、`GOFLAGS` 中的 `-tags=…` 会被追加。

**最小示例**
```sh
go tool mygo build -platform darwin/universal,windows/amd64,linux/amd64
```

**限制/常见错误**
- 产物：macOS 为 `My App.app` 与 `.dmg`；Windows 为 `My App.exe` 与安装程序；Linux 为可执行文件、桌面条目与图标、含 `install.sh` 的归档，以及 `.deb`/`.rpm`。
- 无 cgo 即可跨平台编译，但 macOS 应用的签名与磁盘映像制作仍需在 Mac 上完成；交付前应签名并公证。

## 配置

**用途**：用一个文件描述应用元数据、开发/构建方式、打包与签名。

**关键 API**
- 项目根目录二选一：`mygo.config.ts`（默认导出 `defineConfig({...})`，类型来自 `mygo-cli`）或 `mygo.json`，不能同时存在。
- 应用字段：`name`、`slug`、`identifier`、`version`、`copyright`、`publisher`、`license`、`icon`、`main`。
- 开发/前端字段：`devUrl`、`devCommand`、`buildCommand`、`frontendDist`、`bindings`、`out`。
- 集成字段：`resources`、`resourcesExclude`、`urlSchemes`、`fileAssociations`（`ext`/`name`/`mimeType`/`role`）。
- 配置段：`i18n`、`updates`、`macos`、`windows`、`linux`；配置可为同步或异步函数，参数含当前命令 `"dev"`/`"build"`/`"generate"`/`"init"`。

**最小示例**
```ts
import { defineConfig } from "mygo-cli";

export default defineConfig({
  name: "My App",
  slug: "my-app",
  identifier: "com.example.myapp",
  version: "1.2.0",
  devUrl: "http://localhost:5173",
});
```

**限制/常见错误**
- `mygo.config.ts` 由 Bun 求值，没有 Bun 时用 Node.js 22.6 或更高版本（只移除类型，不能用 `enum` 等生成代码的语法）。
- `mygo dev` 只在该文件本身变化时重载，其导入的文件变化不会。
- `windows.installDirName` 不能含路径分隔符与 `: * ? " < > |`，不能是 `.` 或 `..`；`updates` 必须有 `publicKey`，且 `github` 与 `url` 至少一个。

## CLI 命令一览

**用途**：查常用子命令与发布辅助命令。

**关键 API**
- 项目命令：`mygo init`、`mygo dev`、`mygo build`、`mygo generate`、`mygo i18n check`、`mygo native ...`。
- `mygo install-skills [dir]`：在 `.agents/skills/` 安装或更新随 CLI 捆绑的智能体技能；覆盖同名的捆绑文件，保留其他技能和自定义文件。
- `mygo keygen [-o dir] [-force]`：生成更新签名密钥 `mygo-update.key` 与 `mygo-update.pub`。
- `mygo doctor`、`mygo version`。

**最小示例**
```sh
mygo install-skills
mygo keygen
mygo i18n check
```

**限制/常见错误**
- `mygo native` 系列是实验性的：命令、标志与产物可能在任何发布版本中改变。
- `mygo i18n check` 有错误时以状态 1 退出；语言相关内容见另一份参考 `10-i18n-data-audio-native-modules.md`。

## 环境变量

**用途**：不修改配置的情况下调整构建与运行行为。

**关键 API**
- `CGO_ENABLED`：CLI 在未设置或为空时按 `0` 处理；依赖需要 cgo 时设为 `1`（对 `dev`、`build`、`generate` 生效）。
- `MYGO_LDFLAGS`：追加的 `-ldflags`，排在 CLI 自身参数之后；`GOFLAGS` 中的 `-tags=…` 会并入构建标签。
- `MYGO_INSPECTOR=1`：生产构建保留原生 UI 检查器；`MYGO_ENV=production`：让应用按生产构建行为运行。
- `MYGO_UPDATER_PRIVATE_KEY`、`MYGO_WINDOWS_CERTIFICATE_PASSWORD`：更新签名私钥与 Windows 证书密码。
- `MYGO_CLI_BINARY`：供 `mygo-cli` 包运行的 CLI 构建；`MYGO_NATIVE_<NAME>`：指定原生库文件；`MYGO_LOCALE`：覆盖运行语言。

**最小示例**
```sh
CGO_ENABLED=1 go tool mygo build
```

**限制/常见错误**
- `MYGO_LDFLAGS` 按引号外空格拆分，与 `go build -ldflags` 的规则一致；`MYGO_NATIVE_<NAME>` 中的名称转大写，字母和数字之外的字符替换为 `_`。

## 速查

| 页面/API | 一句话 |
|---|---|
| README | Web 窗口与原生 UI 两种形态，共用窗口、事件、菜单与桌面 API。 |
| `mygo init` | 创建 Web 或原生 UI 模板项目，默认 Web 模板。 |
| `mygo dev` | 实时重载开发，Go/配置/资源变化重建并重启应用。 |
| `mygo build` | 构建并打包到 `<out>/<os>-<arch>`，可多平台。 |
| `mygo generate` | 生成 TypeScript 客户端；有 `i18n` 时先生成消息函数。 |
| `mygo doctor` | 检查 Go、Bun、webview 与可选工具。 |
| `mygo install-skills` | 安装/更新随 CLI 捆绑的智能体技能。 |
| `mygo keygen` | 生成更新签名密钥对并打印配置片段。 |
| `mygo.config.ts` | 应用元数据、开发/构建命令、打包与签名配置。 |
| `mygo.json` | 同上，JSON 形式，不需要 JavaScript 运行时。 |
| `mygo.Bind` | 把 Go 值暴露给页面，经生成的客户端调用。 |
| `mygo.NewEvent` | 声明可订阅的类型化事件。 |
| `ui.View` / `ui.Context` | 原生 UI 视图函数与逐帧构建上下文。 |
| `-platform` | 逗号分隔的 `GOOS/GOARCH`，或 `web`。 |
| `CGO_ENABLED` | CLI 默认 `0`；需要 cgo 依赖时设为 `1`。 |
| `MYGO_LOCALE` | 覆盖应用本次运行使用的语言。 |
