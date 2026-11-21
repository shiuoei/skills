# 实战经验：真实 MyGo 应用踩过的坑

本章来自一个真实 Windows 桌面应用（PrintBridge：Go + MyGo 原生 UI，托盘常驻，签名自动更新，GitHub Actions 发布）的开发记录，环境为 Windows 11、Go 1.27.x、MyGo v0.9.0。公开仓库：<https://github.com/shiuoei/print-bridge>；下文文件路径均相对该仓库，命令可直接复现。少数结论会标注「v0.9.0 验证」，升级 MyGo 后请复核。

## 1. `slog` 日志死锁：不要包装默认 handler

**现象**：作业已经打印成功，进程却卡死在写成功日志的那一行，之后任何日志调用都不再输出，进程悬挂。

**根因**：`slog.SetDefault` 会把标准库 `log` 的输出接到新 logger 上，而 `slog` 的默认 handler 又通过 `log` 输出。把 `slog.Default().Handler()` 包进自己的 handler 再 `SetDefault`，就形成 `slog → 自己 → slog 默认 handler → log → 自己` 的闭环；`log` 的输出锁不可重入，第二次进入即死锁，且该 goroutine 永久持锁。

**做法**：日志装配只留一个入口（例如 `logging.Install`），安装自定义 handler 时绝不回接 `slog.Default().Handler()`；对包装默认 handler 的写法直接拒绝并加回归测试。新增日志入口也必须走同一入口。

**证据**：`internal/logging/logging.go`、`internal/logging` 的死锁回归测试、`docs/CHANGELOG.md` 中「日志死锁」一节。

## 2. 单实例锁的身份随构建方式变化

**现象**：已安装的应用正在运行，`go run`/`go build` 的开发版仍能再启动一个，两个实例同时操作同一份数据。

**根因**：单实例锁的 key 由应用身份决定；打包构建使用 `mygo.json` 的 `identifier` + 名称，而 `go run`/`go build` 没有打包元数据，只能用名称。两者不是同一把锁。

**做法**：在 `main` 一开始调用 `mygo.App.RequestSingleInstanceLock()`，拿不到就直接返回；再注册 `mygo.App.OnSecondInstance(func(args []string, wd string))`，把参数转给已运行实例并唤起窗口。开发时不要假设锁能拦住打包版本。

**证据**：`cmd/desktop/main.go`；MyGo 文档「应用」页的单实例一节。

## 3. 只用 `go build`：没有应用图标，也没有版本

**现象**：资源管理器、任务栏、窗口标题栏都是 Windows 默认空白图标；「关于」页版本显示 `dev`。

**根因**：`go build` 不会把图标资源、版本资源、Manifest 写进 exe。窗口类图标来自 `LoadIconW(instance, 1)`，资源不存在时得到 0，于是回落到系统默认图标；`App.Version()` 同理来自打包时写入的版本。

**做法**：发布与需要图标的验证一律用 `mygo build`；开发构建用 `Window.SetIcon(png)` 兜底，版本用环境变量/ldflags 兜底（如 `PRINTBRIDGE_VERSION`，再回退 `mygo.App.Version()`）。多入口应用（桌面/CLI/headless）要各自注入版本，否则 CLI 的 `doctor` 会显示 `dev`。

**证据**：`docs/CHANGELOG.md`「应用图标缺失」「版本显示为 dev」两节；`cmd/desktop/appicon.go`、`cmd/desktop/main.go`。

## 4. 图标体系与托盘尺寸

**要点**：
- 应用图标是方形 PNG（理想 1024×1024），默认 `resources/icon.png`；旁边的 `icon-dev.png` 是开发版图标。
- `mygo build` 用它生成 exe 的七档图标资源、NSIS 安装器的 `MUI_ICON`/`MUI_UNICON`、Linux 桌面图标、macOS 的 `.icns`。
- **托盘 PNG 的宽度就是 HICON 的尺寸**：直接把 1024 的 PNG 传给 `TrayOptions.Icon`，系统会按 1024 建图标，托盘里巨大且可能显示异常；先缩放到 32×32（Retina/DPI 场景仍按文档建议的 16×16 点即 32×32 像素）。
- 窗口图标用 256×256 交给系统按 DPI 缩放即可，省内存。
- 缩放要做面积平均，并按覆盖率加权颜色，否则透明边缘缩小后会出现黑边。

**证据**：`cmd/desktop/appicon.go`（`iconAt` 实现）、`docs/CHANGELOG.md`「应用图标缺失」一节。

## 5. 托盘应用：关闭窗口不等于销毁窗口

**现象**：点关闭后托盘图标还在，但再点「打开窗口」没有反应。

**根因**：`OnClose` 里让窗口真正关闭后，MyGo 会清掉原生句柄，之后再 `Show()` 无效。

**做法**：托盘应用在 `OnClose` 中 `e.PreventDefault()` 后 `Hide()`；用 `quitting` 标志区分「用户要退出」，退出时放行并 `mygo.App.Quit()`。退出前在 `OnWillQuit` 里清理数据库、后台任务和临时文件。

**证据**：`cmd/desktop/main.go` 的窗口关闭逻辑。

## 6. 托盘创建失败必须记录

**现象**：托盘创建失败（图标无效、系统限制等）后应用继续在后台运行，但没有任何入口能再打开窗口，用户只能杀进程。

**做法**：`mygo.NewTray` 的错误不能忽略；失败时至少写日志并考虑直接退出或改用窗口内操作入口。

## 7. 自动更新：配置、密钥与发布说明三件套

**要点**：
- `updates.publicKey` 必需，且必须与签名私钥配对；`github`（公开仓库 `owner/name`）或 `url` 二选一。
- `mygo keygen` 生成密钥，**私钥绝不入库**；构建时用 `MYGO_UPDATER_PRIVATE_KEY` 环境变量（或 `updates.privateKey` 指向文件）。
- 更新说明文件（`updates.changelog`，默认 `CHANGELOG.md`）里必须有与版本一致的 `## <version>` 节；缺失时带私钥构建会直接失败（这是有意的保护）。
- 没有私钥时构建会**静默跳过**更新产物：CI 忘记注入 Secret 时，产物能发布但客户端检查更新会 404。
- 更新窗口语言在插件注册时固定（`updater.Options.Language`），切换语言后需重启才生效；纯原生 UI 应用用 `plugins/updater/native`，不要用依赖 webview 的 `updater.Plugin`。

**证据**：`mygo.json`、`cmd/desktop/main.go`、`docs/DEVELOPMENT.md`「发布与自动更新」、`docs/RELEASES.md`。

## 8. 自动更新只对打包构建生效，且安装目录必须可写

**要点**：`mygo dev`/`go run` 的产物没有更新能力；安装在 `Program Files` 等只读目录时更新无法写入，`mygo.Updater.Enabled()` 为 false。安装器默认的 per-user 安装（`%LOCALAPPDATA%\Programs`）是可写的，这也是推荐的发布形态。升级 0.1.0 → 0.2.0 这类「旧版没有更新逻辑」的情况仍需手动安装一次。

**证据**：`docs/CHANGELOG.md`「检查更新与签名自动更新」、`docs/TODO.md` 更新器一行。

## 9. 增量更新依赖「已发布的上一版」

**要点**：`mygo build` 会尝试从更新源读取已发布版本来生成 delta。第一次发布仓库里没有任何发布版本，日志会出现「no version is published yet」，属于预期。之后每次发布都能生成增量（例：0.2.0 → 0.2.1 的增量约 124 KB，而整包约 19.6 MB）。`updates.deltas` 控制保留多少个旧版本可用增量。

**证据**：`docs/CHANGELOG.md` 中 0.2.0 与 0.2.1 的构建记录。

## 10. 发布：tag 必须等于配置版本，且分两步

**要点**：
- `mygo build -upload` 上传到配置版本对应的 tag；tag 与版本不一致会把产物挂到错误的 release，所以工作流第一步就要校验。
- `mygo build -upload` 创建的是**草稿** release；要再用 `gh release edit --draft=false` 发布成 latest。任何一步失败，release 留在草稿，客户端看不到。
- 已发布的资产不要删除/覆盖；修复版本就发新版本。

**证据**：`.github/workflows/release.yml`、`docs/DEVELOPMENT.md` 发布一节。

## 11. 交叉编译边界

**要点**：桌面构建 `CGO_ENABLED=0`，Go 可以直接交叉编译到其他桌面平台；但 **macOS 的签名与磁盘映像必须在 Mac 上完成**，Linux 上交叉构建 Windows 安装器需要 `makensis`（NSIS）。Linux 原生 UI 需要 GTK 3，网页窗口需要 WebKitGTK；Windows 网页窗口需要 WebView2；纯原生 UI 的 Windows/Linux 应用不需要 webview 运行时。开发机先用 `mygo doctor` 检查。

**证据**：`.github/workflows/release.yml` 的 NSIS 步骤；MyGo 文档「构建与分发」「GitHub Actions」。

## 12. 资源目录与分发

**要点**：
- `resources/` 顶层文件/目录会进入应用的资源目录（运行时 `mygo.App.Path(mygo.PathResources)` 可读）。
- 平台专用资源放到 `<goos>-<goarch>/` 子目录（例如 `windows-amd64/` 放 Windows 专用工具），构建时只带对应平台。
- `.mygo-files` 记录应用拥有的文件，更新时用来删除旧版本文件；不要手动破坏它。
- 需要随包分发的第三方二进制注意许可证与 SHA-256 清单（本项目的 SumatraPDF 见其 `THIRD_PARTY_NOTICES.md`）。

**证据**：`resources/`、`packaging/`、`THIRD_PARTY_NOTICES.md`。

## 13. 原生 UI 的开发与测试模式

**要点**：
- `ui.View` 每帧根据状态重建界面；状态放应用结构体，事件用 `Clicked()`/`Changed()` 查询返回值。
- `ui.NewTester(view, w, h)` 在内存里渲染并模拟点击、输入、滚动、键盘、菜单；按文本或 `Label` 找元素。测试里用 `tt.Advance(d)` 推进时钟，**不要 sleep**。
- 渲染函数里不要做阻塞 IO；耗时工作丢给 goroutine，完成后用 `Window.Invalidate()` 触发重绘。
- 列表/表格用惰性构建，避免一次性构造大量行；视图身份需要跨帧保持时用 `Key`/`Handle`。

**证据**：`cmd/desktop/*_test.go` 中的 `ui.NewTester` 用例；MyGo 文档「原生 UI」的 testing 页。

## 14. 桌面与 headless 拆分

**要点**：同一 Go 模块可以有多个入口（桌面、CLI、服务器）。headless 入口不要 import MyGo（避免引入 GTK/webview 运行时依赖）；把打印、协议、存储等逻辑放在 `internal/`，由桌面/CLI/服务器共用。Linux 服务器构建因此不需要图形环境。

**证据**：`cmd/server/` 与 `cmd/desktop/` 的拆分、`docs/DEVELOPMENT.md` 代码结构一节。

## 15. Windows GUI 进程没有控制台

**要点**：`-H=windowsgui` 的进程没有标准输出；崩溃与日志默认看不见。做法：内存环形缓冲 + 文件日志（便于用户导出诊断），日志装配统一入口（见第 1 条）。开发期可保留控制台构建或环境变量开关。

**证据**：`internal/logging/`、`cmd/desktop/diagnostics_ui.go`。

## 16. 外部工具/依赖的行为以源码与实测为准

**现象**：项目用 SumatraPDF 打印 PDF，按文档/注释给 `-print-settings` 传 `center`，居中没有生效。

**根因**：SumatraPDF 3.6.1 的 `-print-settings` 没有 `center` 令牌，未知令牌被静默忽略；`noscale` 的落点是物理纸张左上角，而居中只能由 `shrink`/`fit` 缩放模式表达（页面级居中，不能移动 PDF 页面内部内容）。

**做法**：依赖外部 CLI 时，优先读对应版本的源码/`--help` 输出并做最小复现；文档和注释都可能过期。把「版本 + 证据 + 复现命令」记进变更记录。

**证据**：`docs/CHANGELOG.md`「修复居中打印不生效」、`internal/printer/printer_windows.go`。

## 17. 发布闭环的验证清单

一次可复现的发布验证（本项目 0.2.0 → 0.2.1 实际流程）：

1. 推 tag 后确认 GitHub Actions 的 `Release` run 成功（不是草稿）。
2. 检查 release 资产：安装包、归档、增量、`update-windows-amd64.json`、`install.ps1`、`SHA256SUMS`。
3. 校验 `SHA256SUMS`；用系统文件属性/资源查看器确认 exe 版本资源等于 `mygo.json` 的 version。
4. 在安装好的**旧版本**客户端上验证：检查更新 → 增量下载 → 验签 → 重启生效。
5. 记录仍未验证的场景（真实设备、全新安装、断网/私钥错误等），不要写「通过」。

**证据**：`docs/CHANGELOG.md` 0.2.0/0.2.1 两节、`.github/workflows/release.yml`。

## 经验怎么记

每条经验至少写清：**现象**（用户能看到什么）、**根因**（代码/版本级证据）、**做法**（可直接照做）、**证据**（文件、命令、版本、环境）。只写结论而没有证据的「经验」，下次排查会被推翻；只记录一次验证的环境限制（Windows 11、MyGo v0.9.0、特定打印机），避免把单机结论当成跨平台事实。
