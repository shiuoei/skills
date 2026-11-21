# 窗口、应用生命周期与文稿

MyGo 桌面应用的主干：应用启动、事件循环与退出、单一实例、系统集成（深度链接、文件关联、最近文稿、登录启动、标准目录），窗口的创建、选项、事件与页面行为，以及每个窗口一个文件的文稿应用。窗口内容二选一：`URL` 显示系统 webview 中的网页，`Content` 显示 MyGo 用 `ui` 包绘制的原生 UI；除“页面”相关选项外两者规则相同。

## 源页面

| 源页面 | 本文件覆盖 |
| --- | --- |
| [`windows.md`](https://mygo.run/zh/docs/windows) | 窗口选项与事件、浮出窗口、面板、页面与打印 |
| [`app.md`](https://mygo.run/zh/docs/app) | 生命周期、单实例、系统集成、浏览数据 |
| [`documents.md`](https://mygo.run/zh/docs/documents) | 文稿应用 |
| [`launchers.md`](https://mygo.run/zh/docs/launchers) | 启动器 |

## 应用启动、线程与退出

**用途**：启动原生事件循环，在就绪后创建窗口，并界定退出流程。

**关键 API**
- `mygo.App.Run() error`：运行事件循环，应用退出后返回；必须从主 goroutine 调用（各平台 UI 工具包要求启动线程）。
- `mygo.App.WhenReady(fn)` / `OnReady`：启动完成后执行，窗口只在此后存在。`mygo.Bind(service)`、`mygo.BindAs(name, service)` 把服务暴露给前端；`mygo.RunOnMain(fn)` 在主线程运行原生 API 调用。
- `Run` 之前可做：绑定服务、注册监听器与协议，以及 `App.SetName`、`App.SetVersion`、`App.SetMenu`、`App.Dock.SetMenu`、`Theme.SetSource`、`Theme.SetGTKCSS`、`App.SetOpenAtLogin`、`App.RegisterURLScheme`；可读 `App.Path`、`App.Locale`、`Power.IsOnBattery`。
- 其余能力（剪贴板、显示器、对话框、通知、全局快捷键、托盘、Dock 徽标、`Theme.IsDark`、`Theme.Preferences`）依赖运行中的应用，在 `Run` 前调用会在每个平台 panic 并指出调用；从 `WhenReady` 起使用。
- 生命周期事件：`OnReady`、`OnActivate(func(hasVisibleWindows bool))`、`OnDidBecomeActive`、`OnDidResignActive`、`OnWindowCreated(func(w *mygo.Window))`、`OnWindowAllClosed`、`OnBeforeQuit(func(e *mygo.QuitEvent))`、`OnWillQuit`、`OnQuitCanceled`、`OnQuit`；每个 `On…` 返回移除监听器的函数。

**最小示例**
```go
func main() {
	mygo.Bind(&Notes{})         // 前端可调用的服务
	mygo.App.WhenReady(func() { // 应用启动后
		mygo.NewWindow(mygo.WindowOptions{URL: "/"})
	})
	if err := mygo.App.Run(); err != nil { // 事件循环，直到应用退出
		log.Fatal(err)
	}
}
```

**退出**
- 默认最后一个窗口关闭即退出，除非注册了 `OnWindowAllClosed`；macOS 惯例是保留 Dock 图标：在 `OnWindowAllClosed` 中仅当 `runtime.GOOS != "darwin"` 时 `Quit`，在 `OnActivate(hasVisibleWindows)` 中重新开窗。
- `App.Quit()` 先关闭每个窗口（`OnClose` 可拒绝）再退出，`OnBeforeQuit`/`OnWillQuit` 也可取消；Cmd+Q、退出菜单项、Ctrl+C（SIGINT/SIGTERM）走同一流程，第二次信号立即退出。取消后触发 `OnQuitCanceled`：要在这里清除退出状态，否则应用继续运行却保留该状态；在 `OnWillQuit` 阶段取消不会重开已关闭的窗口。
- `App.Exit(code)` 立即退出、不触发事件；`App.Relaunch()` / `Relaunch(args...)` 退出后以相同或替换的参数重启，参数会到达新实例（macOS、Linux、Windows 与 `mygo dev` 下有效，浏览器中只退出）。
- 操作系统结束会话（注销、重启、关机）：Windows 与 macOS 仍依次运行退出监听器、窗口 `OnClose` 和 `OnQuit`，但 `QuitEvent.SessionEnding` 已设置且不能取消，进程可能随时结束：须同步保存数据，不弹对话框、不等待 goroutine。Linux 不通知会话结束，关机发送 SIGTERM，按 `App.Quit()` 流程退出。

**限制/常见错误**：所有 MyGo 函数可从任意 goroutine 调用，调用会被送到主线程；事件监听器、菜单项点击与全局快捷键回调在主线程运行，应尽快返回，耗时工作另起 goroutine；绑定方法在自己的 goroutine 上运行、可以阻塞。未 recover 的 panic 由 `Run` 追加写入 `PathLogs` 下的 `crash.log`（超过 1 MB 时启动清空），`mygo dev`、`go run`、`go test` 只显示标准错误。

## 单一实例

**用途**：同一时间只运行一个实例；再次启动时把命令行交给已运行的实例，并让窗口到前台。

**关键 API**：`mygo.App.RequestSingleInstanceLock() bool`、`mygo.App.OnSecondInstance(func(args []string, workingDir string))`。

**最小示例**
```go
if !mygo.App.RequestSingleInstanceLock() {
	return // 由已运行的实例接管
}
mygo.App.OnSecondInstance(func(args []string, workingDir string) {
	focusMainWindow() // 例如 Restore + Focus
})
```

**限制/常见错误**：第二实例在锁返回 false 时应立即返回，不做初始化；锁持有到进程退出，应用在 `Run` 返回后的工作（停止子进程、写文件）不会与新实例并发；退出开始后新启动会等待旧实例退出（无论多久）再运行；崩溃或被终止时由系统释放锁。

## 深度链接与文件关联

**用途**：从浏览器、文件管理器、Dock 或跳转列表用应用打开 URL 或文件。

**关键 API**
- 配置（`mygo.config.ts`）：`urlSchemes: ["myapp"]`；`fileAssociations: [{ext: ["md", "markdown"], name: "Markdown Document", mimeType: "text/markdown"}]`；安装包安装时注册到系统。
- `mygo.App.OnOpenURL(func(rawURL string))`、`mygo.App.OnOpenFile(func(path string))`：都必须在 `Run` 之前注册，才能收到启动时携带的 URL 或文件。
- 便携版可在运行时注册协议：`App.RegisterURLScheme("myapp")`，配合 `UnregisterURLScheme`、`IsURLSchemeRegistered`；macOS 的协议必须在配置中列出。

**最小示例**
```go
mygo.App.OnOpenURL(openDeepLink) // 在 App.Run 之前注册
mygo.App.OnOpenFile(openDocument)
```

**限制/常见错误**：Windows 和 Linux 打开 URL 或文件会启动应用的新实例，需配合单一实例锁，让已运行实例收到回调；macOS 还会把拖到 Dock 图标的文件、命令行文件交给 `OnOpenFile`；已打开的文件再次打开会聚焦对应窗口。没有安装包的应用（如便携可执行文件）在 Windows/Linux 上需每次启动调用 `RegisterURLScheme`（开销很小）。

## 最近文稿、登录启动、标准目录与任务栏

**用途**：把打开的文稿提供给系统“最近使用”，随登录启动，规范读写位置，并在 Dock/任务栏/跳转列表呈现入口。

**关键 API**
- 最近文稿：`App.AddRecentDocument(path) error`、`App.RecentDocuments()`（最新在前、最多 10 个、含更早运行，已不存在的文件被跳过）、`App.ClearRecentDocuments()`；`RoleOpenRecent` 菜单列出它们，选中会用应用打开文件并触发 `OnOpenFile`。
- 登录启动：`App.SetOpenAtLogin(true) error`、`App.OpenAtLogin() bool`、`App.WasOpenedAtLogin() bool`（如 `Hidden: mygo.App.WasOpenedAtLogin()` 实现登录时后台启动）。
- 目录：`App.Path(mygo.PathUserData)` 等，`App.SetPath` 覆盖某个目录（如分开测试数据）。
- 任务栏：`App.SetBadgeCount(n)`（0 清除）、`App.Dock`（macOS：`SetBadge`、`Bounce`、`SetIcon`、`SetMenu`、`Hide`/`Show`）、`App.SetActivationPolicy(mygo.ActivationPolicyAccessory)`（无 Dock 图标）、`App.Focus`、`App.IsActive`、`App.Hide`/`App.Show`（macOS）、`App.ShowAboutPanel`；Windows 跳转列表 `App.JumpList.SetTasks([]mygo.JumpListTask{{Title, Description, Args, Path, WorkingDir, IconPath, IconIndex}})`。

| `App.Path` 名称 | 目录 |
| --- | --- |
| `PathUserData` | `~/Library/Application Support/<Name>`、`~/.config/<Name>`（`$XDG_CONFIG_HOME`）或 `%AppData%\<Name>`；首次使用时创建 |
| `PathCache` | `~/Library/Caches/<Name>`、`~/.cache/<Name>` 或 `%LocalAppData%\<Name>`；首次使用时创建 |
| `PathLogs` | `~/Library/Logs/<Name>`，否则用户数据目录中的 `logs`；首次使用时创建 |
| `PathResources`、`PathAppData` | 应用随附的文件；每用户配置目录（不带应用名） |
| `PathHome`、`PathTemp`、`PathExe` | 主目录、临时目录、可执行文件 |
| `PathDesktop`…`PathVideos` | 用户桌面、文稿、下载、音乐、图片、视频 |

**最小示例**
```go
dir, _ := mygo.App.Path(mygo.PathUserData)
db := filepath.Join(dir, "notes.db")
_ = mygo.App.SetOpenAtLogin(true)
_ = mygo.App.JumpList.SetTasks([]mygo.JumpListTask{
	{Title: "New note", Args: []string{"myapp://new-note"}},
})
```

**限制/常见错误**：登录启动在 macOS 13+ 是登录项（需打包应用）、macOS 12 是启动代理、Linux 是 XDG 自启动条目、Windows 是用户 Run 条目，系统设置可关闭；Windows 的 Run 条目以普通（未提权）令牌启动，需要提权的应用改用任务计划程序的登录任务（`HighestAvailable`，创建时需管理员），且 `WasOpenedAtLogin` 无法区分这种启动，需给任务命令加自定义参数。`RecentDocuments` 上限 10，Linux 的 `ClearRecentDocuments` 不影响用户其他文件的记录。跳转列表参数最好用已注册协议的 URL 或已关联类型的文件（分别到达 `OnOpenURL`/`OnOpenFile`），其他参数在应用未运行时进新实例的 `os.Args`、运行中则交给 `OnSecondInstance`；Windows 退出后保留任务，启动时设置一次即可，`App.JumpList` 在 macOS/Linux 上不做任何事。应用名决定用户数据、缓存、日志目录与单一实例锁：`go run` 用可执行文件名，`mygo dev` 的构建名为 `<Name> Dev`。

## 窗口选项

**用途**：`mygo.NewWindow(mygo.WindowOptions{...})` 创建窗口；零值是可见、可调整大小的 800×600 窗口，居中显示；尺寸与位置单位为 DIP。

**关键 API**

| 选项 | 说明 |
| --- | --- |
| `Title`、`URL` 或 `Content` | 标题（默认应用名，页面 `<title>` 会替换）；显示网页（`/` 等无协议前缀指向应用页面）或原生 UI |
| `Width`、`Height`、`X`、`Y`、`MinWidth`、`MinHeight`、`MaxWidth`、`MaxHeight` | 尺寸、左上角与调整限制；`X`、`Y` 都为 0 时居中；0 限制表示无 |
| `UseContentSize`、`FitContent` | 尺寸指页面/视图内容区域；按内容大小并在限制内跟随内容 |
| `StateKey`、`Hidden`、`Maximized`、`FullScreen` | 状态记忆键；初始可见性与窗口状态 |
| `Frameless`、`TitleBarStyle`、`TrafficLightPosition`、`TitleBarHeight`、`HideWindowButtons` | 自定义标题栏；隐藏按钮时页面自行绘制并调用 `Minimize`、`Maximize`、`Unmaximize`、`Close` |
| `Transparent`、`Vibrancy`、`VibrancyTint`、`CornerRadius`、`BackgroundColor`、`Opacity` | 透明与材质（`VibrancyGlass`、`VibrancyGlassClear` 为 macOS 26 Liquid Glass）；`BackgroundColor` 支持 `light-dark(...)` |
| `DisableResize`、`DisableMove`、`DisableMinimize`、`DisableMaximize`、`DisableClose`、`DisableFullScreen`、`DisableShadow` | 去掉对应能力 |
| `AlwaysOnTop`、`SkipTaskbar`、`AutoHideMenuBar`、`IgnoreMouseEvents`、`ConsumeActivationClick` | 层级、任务栏、菜单栏与鼠标行为 |
| `Parent`、`Modal`、`Panel`、`Page` | 子窗口、模态窗口、启动器面板；`mygo.PageOptions` 见“窗口的页面与页面行为” |

**最小示例**
```go
win := mygo.NewWindow(mygo.WindowOptions{
	Title: "Editor", URL: "/", Width: 1200, Height: 800,
	MinWidth: 600, MinHeight: 400, StateKey: "editor",
})
win.OnClosed(func() { log.Println("closed") })
```

**限制/常见错误**：只有网页窗口有页面，原生 UI 窗口的 `win.Page()` 为 nil。`ConsumeActivationClick`（macOS、Windows）让非活动窗口的首次点击只激活窗口；Windows 上未用 `Content` 的普通 WebView2 页面窗口在非激活时禁用子控件，激活前收不到滚轮、悬停和文件拖放，原生 UI 与嵌入的 WebView 则保留这些能力。透明窗口在 Windows 要求 11 22H2；带材质的窗口没有菜单栏，Alt/F10 以弹出菜单打开菜单；`VibrancyTint` 仅 macOS 26 Liquid Glass 生效。Wayland 下合成器决定位置并忽略应用设置的位置，装饰不属于 `Bounds`。

## 窗口状态、尺寸与显示

**用途**：避免白屏闪烁、恢复窗口与页面位置、读写窗口几何。

**关键 API**
- 无闪烁：网页窗口在页面绘制前填充 `BackgroundColor`（默认白色），跟随深色模式用 `light-dark(<light>, <dark>)`；只在页面就绪后显示用 `Hidden: true` + `win.OnReadyToShow(win.Show)`。原生 UI 窗口绘制首帧即显示（Windows 上首帧前一直不可见）。
- 显示时机：主线程调用 `NewWindow` 时窗口延迟到主循环下一轮、首帧绘制时显示，紧接注册的 `OnShow` 能收到，`Hide` 可阻止；从其他 goroutine 调用时先显示后返回，访问保存窗口的变量需同步或用 `Window.Update`。
- 状态记忆：`StateKey` 把位置、尺寸、最大化与全屏写入用户数据目录的 `window-state.json`（窗口关闭与应用退出时写）；同键新窗口取回并覆盖 `X`、`Y`、`Width`、`Height`、`UseContentSize`、`Maximized`、`FullScreen`；Windows 另存像素边界（多显示器缩放不同仍准确）。`App.SetWindowStateStore(store)` 在第一个带 `StateKey` 的窗口前设置，改用应用自己的存储（须保留 `WindowState` 全部字段）；`App.ClearWindowStates()` 永久遗忘状态。
- 页面恢复：`win.Page().SessionState()` 返回不透明字节（前进/后退列表、滚动位置、表单字段），`PageOptions{SessionState: state}` 在新窗口恢复并代替 `URL`。macOS 用 `WKWebView` 的 `interactionState`（12+，之前只有 URL）、Linux 序列化会话状态、Windows 只有页面 URL（WebView2 无前进/后退 API）。状态可能含用户输入，按用户数据保管；`LoadHTML` 的内容不在其中。
- 几何与状态：`SetSize`、`SetContentSize`、`SetPosition`、`SetBounds(mygo.Rectangle{...})`、`Center`，读回 `Size`、`ContentSize`、`Position`、`Bounds`、`ContentBounds`；`Maximize`/`Unmaximize`/`ToggleMaximize`/`IsMaximized`、`Minimize`/`Restore`/`IsMinimized`、`SetFullScreen`/`ToggleFullScreen`/`IsFullScreen`、`Hide`/`Show`/`ShowInactive`/`IsVisible`、`Focus`/`Blur`/`IsFocused`。
- 其他设置：`SetTitle`、`SetResizable`、`SetMovable`、`SetMinimizable`、`SetMaximizable`、`SetClosable`、`SetMinimumSize`、`SetMaximumSize`、`SetAlwaysOnTop`、`SetOpacity`、`SetHasShadow`、`SetBackgroundColor`、`SetVibrancy`、`SetIcon(png)`、`SetSkipTaskbar`、`SetAutoHideMenuBar`、`SetVisibleOnAllWorkspaces`、`SetContentProtection`、`SetIgnoreMouseEvents`、`SetActiveAppearance`、`SetWindowButtonsTheme`、`SetProgressBar`、`FlashFrame`、`SetThumbBar`。

**最小示例**
```go
win.OnClose(func(e *mygo.CloseEvent) {
	if state, err := win.Page().SessionState(); err == nil {
		os.WriteFile(sessionFile, state, 0o600) // 下次启动传给 PageOptions.SessionState
	}
})
```

**限制/常见错误**：`IsVisible` 表示被 `Show` 显示过且此后未 `Hide`——最小化、被覆盖、在其他桌面的窗口都算可见；`IsMinimized` 才表示已最小化。尺寸与位置是窗口的 DIP，不受页面缩放影响；`Bounds` 是整个窗口（含标题栏和边框），`ContentBounds` 是页面/原生 UI 区域。`ClearWindowStates` 之后被“重置设置”删除的状态会被写回；它不保存任何状态，对自定义存储写入空 map。`SetThumbBar` 仅 Windows 有效、最多 7 个按钮，首次调用确定数量和顺序，其他平台返回 `errors.ErrUnsupported`。macOS 的 `SetActiveAppearance` 使用私有 AppKit 调用，Mac App Store 应用应保持关闭。

## 窗口事件

**用途**：响应用户操作和窗口状态变化；监听器在主线程运行，应保持简短。

**关键 API**

| 事件 | 何时 |
| --- | --- |
| `OnClose(func(e *mygo.CloseEvent))` | 窗口即将关闭（用户或 `Close`）；`e.PreventDefault()` 保持打开 |
| `OnClosed` | 窗口已关闭 |
| `OnFocus`、`OnBlur` | 获得或失去键盘焦点 |
| `OnShow`、`OnHide` | 被显示或隐藏 |
| `OnResize`、`OnMove` | 被调整大小或移动 |
| `OnMaximize`、`OnUnmaximize`、`OnMinimize`、`OnRestore` | 状态改变 |
| `OnEnterFullScreen`、`OnLeaveFullScreen` | 进入或离开全屏 |
| `OnReadyToShow` | 第一个页面已可显示或加载失败，或原生 UI 绘制出第一帧 |
| `OnFileDrop` | 文件被拖放到页面或原生 UI |

**最小示例**
```go
win.OnClose(func(e *mygo.CloseEvent) {
	if !doc.Dirty() {
		return
	}
	e.PreventDefault()        // 先阻止关闭，再异步询问
	go askAndDestroy(win, doc) // 内部用 Dialog.Message 询问，选保存则 doc.Save() 后 win.Destroy()
})
```

**限制/常见错误**：`Close` 与用户关闭窗口一致，允许 `OnClose` 阻止；`Destroy` 立即关闭、不询问。能阻塞的 `Dialog` 必须在 goroutine 中调用，主线程只负责 `PreventDefault`，否则对话框等待期间 UI 与退出流程一起卡住。窗口的关闭按键、Cmd+W、Alt+F4 都走 `OnClose`。

## 子窗口、浮出窗口、面板与点击穿透

**用途**：模态对话框、紧贴控件的浮层、随时唤出的启动器面板，以及鼠标穿透的桌面宠物。

**关键 API**
- 子窗口与模态：`Parent` 让子窗口始终在父窗口之上并随父关闭，可移到任意显示器，父窗口移动时子窗口不跟随；macOS 上父窗口到前台子窗口排其上、父窗口全屏子窗口跟随、父窗口最小化或隐藏时隐藏；`Modal: true` 阻止操作父窗口，macOS 显示为附着式对话框（sheet）。父窗口也可以是浮出窗口。
- 浮出窗口：`mygo.NewFlyout(mygo.FlyoutOptions{...})`，无边框、无背景、无任务栏按钮，紧挨父窗口的锚点矩形放置（锚点以父内容区域 DIP 计）；字段 `Parent` 或 `Tray`、`Anchor`、`Placement`（如 `"bottom-start"`）、`Gap`、`Width`/`Height`、`URL` 或 `Content`、`Focusable`、`Popover`、`Shadow`、`FitContent`、`MinWidth`/`MaxWidth` 等；`SetAnchor`、`SetSize`、`IsFlyout`。
- `Focusable`：显示即获键盘焦点，点击外部、其他窗口获得焦点或应用失去活动时关闭（Wayland 合成器会直接关闭）；不设置时焦点留在父窗口，持续显示到主动关闭，适合工具提示、悬停卡片、输入建议；未处理的 Escape 会关闭它。
- 托盘浮出窗口：用 `Tray` 代替 `Parent`，紧挨图标、浮在其他应用窗口之上；`PlacementBottom` 在 macOS 菜单栏下居中、Windows 上翻到任务栏图标上方；`Popover`（macOS）改用 `NSPopover` 系统浮层外观。
- 浮出窗口上的对话框：另建以浮出窗口为 `Parent`、`Modal: true` 的窗口，无 `X`/`Y` 时在其上方居中并落在工作区内；期间浮出窗口不接受输入、不失焦关闭，关闭浮出窗口也关闭它。
- 面板：`Panel: true` 的窗口，用于启动器、命令面板、快速输入；显示在当前 Space/虚拟桌面、浮在所有应用窗口之上（含全屏）、不出现在任务栏/Dock 列表/应用切换器；不能最小化、最大化或全屏，Linux 和 Windows 没有菜单栏，macOS 上不激活所属应用。
- 点击穿透：`SetInputRegion([]mygo.Rectangle{...})` 只让这些 DIP 矩形接收鼠标，其他区域穿透；`nil` 恢复全窗口接收，空切片让全部穿透（等价 `SetIgnoreMouseEvents(true)`，后者优先）；内容移动或缩放后要重设区域。

**最小示例**
```go
fl := mygo.NewFlyout(mygo.FlyoutOptions{
	Parent: win, Anchor: mygo.Rectangle{X: 20, Y: 40, Width: 120, Height: 28},
	Placement: mygo.PlacementBottomStart, Gap: 4, Width: 240, Height: 300,
	Content: ui.View(app.suggestions),
})
panel := mygo.NewWindow(mygo.WindowOptions{Panel: true, Hidden: true, URL: "/launcher"})
mygo.App.OnWindowAllClosed(func() {}) // 无窗口也保持运行
mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) { quitting = true })
panel.OnClose(func(e *mygo.CloseEvent) {
	if !quitting {
		e.PreventDefault()
		panel.Hide() // 隐藏而不是关闭，下次立即显示
	}
})
```

**限制/常见错误**：工作区没有空间时浮出窗口翻到另一侧、再滑回工作区、最后缩小；Linux 和 Windows 的外部点击只关闭浮出窗口，macOS 还会把点击传给指针下方的内容（但不给锚点）。Linux 托盘图标不报告点击与位置，应从菜单项打开浮出窗口（显示在主显示器工作区中央）。保持打开的面板会取消退出，`OnClose` 要按退出标志放行；未处理的 Escape 会关闭面板。Wayland 上面板与点击穿透窗口需要合成器支持 layer shell（Sway、Hyprland、KDE Plasma、COSMIC 等）且安装 `libgtk-layer-shell.so.0`，否则位置和层级由合成器决定；`AlwaysOnTop` 的 layer surface 用 `SetAlwaysOnTop(false)` 无法退回普通窗口，需重新创建。原生 UI 的 `PassThrough` 是让指针穿过元素到窗口内的内容，不会穿到其他窗口。

## 窗口的页面与页面行为

**用途**：网页窗口显示一个带历史记录的页面；原生 UI 窗口的 `win.Page()` 返回 nil。页面 API 覆盖导航、事件、缩放、打印、下载、权限、认证与内容拦截。

**关键 API**
- 页面：`LoadURL`、`LoadHTML`、`LoadFile`、`Reload`、`ReloadIgnoringCache`、`Stop`、`GoBack`、`GoForward`、`CanGoBack`、`CanGoForward`、`URL()`、`SetZoomFactor`、`ZoomFactor`、`Print`、`PrintToPDF`、`CapturePage`、`FindInPage`、`ProcessID`。
- `WindowOptions.Page` 的 `PageOptions`：`PreloadScript`、`PreloadAllFrames`、`TrustedOrigins`、`DevTools`、`ZoomFactor`、`UserAgent`、`UserAgentSuffix`、`Session`、`SessionState`、`SwipeNavigation`、`PinchZoom`、`ZoomControl`、`PictureInPicture`、`SafariTabFocus`、`ContentRules`、`BackgroundThrottling`、`Autoplay`。
- 页面事件：`OnWillNavigate(func(e *mygo.NavigateEvent))`、`OnDidNavigate(func(url string))`、`OnDOMReady`、`OnDidFinishLoad`、`OnDidFailLoad(func(err *mygo.LoadError))`、`OnPageTitleUpdated(func(e *mygo.TitleEvent))`、`OnThemeColorChanged(func(color.NRGBA, ok bool))`、`OnRenderProcessGone(func(reason string))`、`OnAudibleChanged(func(bool))`、`OnCaptureStateChanged(func(kind mygo.CaptureKind, s mygo.CaptureState))`。
- 缩放与手势：`ZoomFactor`/`SetZoomFactor` 由应用控制，`Page.ZoomFactor` 报告用户缩放（WebView2 按站点记忆）；`ZoomControl` 只在 Windows 上让用户用 Ctrl+滚轮缩放（macOS、Linux 不因按键或滚轮缩放）；`SwipeNavigation`、`PinchZoom` 默认关闭（引擎默认开启，MyGo 关闭）；捏合只缩放显示、不重新布局。
- 打印与截图：`Page.Print()` 打开打印对话框；`PrintToPDF(mygo.PDFOptions{PageSize: mygo.PageA4, Landscape: false, Margins: &mygo.Margins{Top: 0.5, Right: 0.5, Bottom: 0.5, Left: 0.5}, Background: true})`，边距单位英寸，打印样式表生效；零值 `PDFOptions` 为 Letter、纵向、0.4 英寸边距、不打印背景；`win.CapturePage()` 返回窗口当前内容的 PNG。macOS 上 WebKit 的 PDF 把许多汉字记成康熙部首，`PrintToPDF` 会追加修正后的文本映射而不改变外观。
- 下载：默认进入“下载”目录并使用页面/服务器建议的名称；`Page.OnWillDownload(func(e *mygo.DownloadEvent))` 用 `e.SuggestedName` 判断、`e.Path` 改路径或 `e.PreventDefault()` 取消；`Page.OnDownloadDone(func(d *mygo.Download))` 用 `d.Path`、`d.Err` 判断结果。
- 权限：默认应用自己的页面（含自定义协议页面）获准，其他页面被拒绝；`Page.SetPermissionHandler(func(req mygo.PermissionRequest) bool)` 在主线程决定，`req.Origin`、`req.Permissions` 供判断，系统也可能询问用户。
- 主题色与声音：`Page.ThemeColor()`、`OnThemeColorChanged` 读取页面 `<meta name="theme-color">`（页面可用 `media` 为浅色/深色分别指定）；`IsAudible`、`OnAudibleChanged`、`SetAudioMuted`、`IsAudioMuted` 控制整页声音，静音状态在导航后保留。
- 后台与自动播放：`BackgroundThrottling: mygo.ThrottlingOff` 让页面在窗口隐藏、最小化或被覆盖时继续运行；`Autoplay: mygo.AutoplayAllow` 让隐藏窗口的页面也能在用户点击前播放可听媒体。Windows 上这两个选项属于 WebView2 环境，应用所有页面共享、由第一个显示页面的窗口创建，必须把选项给那个窗口，且同一应用的两个进程要一致。
- 捕获与画中画：`Page.CaptureState(kind)`（`CaptureNone`、`CaptureActive`、`CaptureMuted`）、`Page.SetCaptureState(kind, state)` 无需页面配合即可静音/恢复/停止捕获；Windows 上总是 `CaptureNone`。`PictureInPicture`（macOS 需开启，WebView2 总允许）、`TogglePictureInPicture`、`IsPictureInPictureActive`；Linux 无画中画。
- 认证与证书：`Page.SetAuthHandler(func(req mygo.AuthRequest) mygo.AuthResponse)` 回答 Basic/Digest/NTLM（`req.Host`、`req.IsProxy`；空响应取消），接受的凭据保留到本次会话结束，无处理器时 WebKitGTK/WebView2 弹自己的登录框、macOS 不弹。`SetCertificateErrorHandler(func(e mygo.CertificateError) bool)` 可放行无法验证的证书，但必须校验 `e.Certificate`（DER），不能仅凭主机名放行；`SetClientCertificateHandler(func(req mygo.ClientCertificateRequest) mygo.ClientCertificateResponse)` 回答双向 TLS，Windows 列出互信证书、macOS 列出钥匙串身份、WebKitGTK 不列出任何证书（Linux 只能拒绝，拒绝会让加载失败）。
- 内容规则：`mygo.CompileContentRules(id, json)` 编译 WebKit 内容拦截格式，页面经 `PageOptions.ContentRules` 或 `Page.SetContentRules` 应用；触发器 `url-filter`、`if-domain`/`unless-domain`、`if-top-url`/`unless-top-url`、`resource-type`、`load-type`（`first-party`/`third-party` 等），动作 `block`、`block-cookies`、`make-https`、`css-display-none`（带 `selector`）、`ignore-previous-rules`。
- 检查器与进程：`OpenDevTools`、`CloseDevTools`、`ToggleDevTools`、`IsDevToolsOpened`（`PageOptions.DevTools` 取 `DevToolsAuto`/`DevToolsEnabled`/`DevToolsDisabled`）；`ProcessID()` 与 `GPUProcessID()` 在 macOS/Windows 返回对应进程（Linux 为 0），页面导航或进程重启后会变，不要缓存；查找窗口用 `mygo.Windows()`、`FocusedWindow()`、`WindowByID(id)`、`CallerWindow(ctx)`、`CallerPage(ctx)`，`App.OnWindowCreated` 在每个新窗口创建时调用。

**最小示例**：让窗口停留在应用自己的页面，其他链接用浏览器打开；并配置打印与权限：
```go
win.Page().OnWillNavigate(func(e *mygo.NavigateEvent) {
	if u, err := url.Parse(e.URL); err == nil &&
		(u.Scheme == "http" || u.Scheme == "https") && e.UserInitiated {
		e.PreventDefault()
		go mygo.Shell.OpenExternal(e.URL)
	}
})
pdf, err := win.Page().PrintToPDF(mygo.PDFOptions{
	PageSize: mygo.PageA4, Background: true,
	Margins: &mygo.Margins{Top: 0.5, Right: 0.5, Bottom: 0.5, Left: 0.5},
})
win.Page().SetPermissionHandler(func(req mygo.PermissionRequest) bool {
	return req.Origin == "https://meet.example.com" &&
		!slices.Contains(req.Permissions, mygo.PermissionGeolocation)
})
```

**限制/常见错误**：`OnWillNavigate` 的 `e.UserInitiated` 区分用户手势与脚本发起的导航，`LoadURL` 及其同类不触发它；iframe 的导航也不触发（Linux 上 WebKitGTK 2.56 之前无法区分框架，会为 iframe 调用）。`OnPageTitleUpdated` 中阻止事件会保持窗口标题不变。开发期页面来自 `http` 开发服务器，按协议判断会误伤，应改查主机或用 `mygo.IsDev()` 区分。内容规则的正则是 WebKit 的：不支持交替（`|`）、计数重复或 `\d` 字符类，仅限 ASCII，域名是小写 punycode；`CompileContentRules` 长列表较慢，应在主线程之外调用，单列表受引擎规则数限制（约 150,000 条，旧版本 50,000）。Windows 没有 WebKit 内容拦截器：被阻止的页面显示 WebView2 的 403 错误页，WebSocket 不会被阻止，`resource-type: popup` 不匹配，`block-cookies` 移除请求的 cookie，页面内联 `!important` 样式仍保留。macOS 的摄像头/麦克风需要 `macos.infoPlist` 中的 `NSCameraUsageDescription`、`NSMicrophoneUsageDescription`，否则系统会结束应用；Linux 上证书处理器只询问页面本身、不询问子资源，为某主机放行的证书保持到应用退出。Linux 的 DMA-BUF 渲染器默认关闭（`Run` 设置 `WEBKIT_DISABLE_DMABUF_RENDERER=1`），可在 `Run` 前用 `App.SetDMABUFRenderer(mygo.DMABUFRendererAuto)`（或 `mygo.DMABUFRendererOn`、`mygo.DMABUFRendererOff`）调整，用户环境变量始终优先。

## Cookie、会话与浏览数据

**用途**：让页面登录、隔离多个账号或私密浏览，并清除页面数据。

**关键 API**
- `App.ClearBrowsingData()`：删除默认会话中页面存储的内容（Cookie、本地和会话存储、IndexedDB、Service Worker、缓存）；打开的页面把内存中的内容保留到重新加载。
- `mygo.Cookies`（默认会话）：`Set(mygo.Cookie{Name, Value, Domain, Secure, HTTPOnly, Expires, SameSite})`、`List(mygo.CookieFilter{URL, Name, Domain, Path})`、`Delete(cookie)`、`Clear()`。Cookie 由名称、域、路径标识：`Set` 替换同名同域同路径的，`Delete` 删除它；零值 `Expires` 持续到应用退出，过去的 `Expires` 删除它；空 `SameSite` 是 `Lax`，`None` 需 `Secure`；不带前导点的域只发给该主机，带前导点（`.example.com`）也发给子域；域是 ASCII（其他用 punycode），空路径为 `/`。
- 会话：`mygo.NewSession("work")`（同名同会话、写盘保留）、`mygo.NewEphemeralSession()`（内存、最后页面关闭即消失）、`PageOptions.Session`；`Session.Cookies`、`Session.ClearData(mygo.SiteDataCookies | mygo.SiteDataCaches)`、`mygo.DefaultSession()`；数据种类 `SiteDataCookies`、`SiteDataLocalStorage`（含会话存储）、`SiteDataIndexedDB`（含 WebSQL）、`SiteDataCaches`（HTTP 缓存与 Cache Storage）、`SiteDataServiceWorkers`、`SiteDataAll`。

**最小示例**
```go
work := mygo.NewSession("work") // 跨运行保留；NewEphemeralSession 只在内存
win := mygo.NewWindow(mygo.WindowOptions{URL: mailURL, Page: mygo.PageOptions{Session: work}})
err := work.Cookies.Set(mygo.Cookie{Name: "session", Value: token, Domain: "example.com"})
```

**限制/常见错误**：Cookie 可能携带凭据，与其把模块绑定给页面，不如绑定一个完成页面所需操作的后端方法。Windows 上 WebView2 只能通过 web 视图提供 Cookie：没有显示页面的窗口时这些调用会创建隐藏 web 视图并启动 WebView2 进程；更改在调用返回后才生效，2 秒内未生效（如引擎拒绝）则返回错误；仅会话期间存在的 Cookie 会让隐藏视图保持打开。Linux 需要 WebKitGTK 2.42+（`Set` 另需 libsoup 2.70+），Cookie 存在 `~/.config/<Name>/cookies.sqlite`。持久会话底层：macOS 14+ `WKWebsiteDataStore`、Linux `Sessions/<UUID>`、Windows WebView2 配置 `mygo-<UUID>`。用 `window.open` 或 `target="_blank"` 打开的窗口留在打开者的会话中，忽略打开选项里的 `Session`。

## 文稿

**用途**：每个窗口对应一个文件的应用；`mygo.NewDocuments` 在三个平台提供与 AppKit `NSDocument`/SwiftUI `DocumentGroup` 相同的功能：每文稿一窗、文件菜单、自动保存、崩溃恢复、系统打开与外部修改跟踪。

**关键 API**
- `mygo.NewDocuments(mygo.DocumentsOptions[T]{...})` 必须在 `App.Run` 之前调用；启动时打开随启动传入的文件或恢复有未保存更改的文稿，否则创建无标题文稿（`NoUntitled` 可关闭）。
- `DocumentsOptions[T]`：`Types`（`[]mygo.DocumentType{{Name, Extensions}}`，打开对话框显示它们，新文稿用第一个扩展名）、`Read`（goroutine 中读）、`Write`（goroutine 中写）、`Window`（主线程建窗）、`New`、`Equal`、`Snapshot`、`Autosave`（`AutosaveInPlace` 默认 / `AutosaveOff`）、`AutosaveDelay`（默认 5 秒）、`AskOnQuit`、`NoUntitled`。
- `Document[T]`：`Content()`、`Set(v)`、`Changed()`、`SetEdited(bool)`、`State()`（名称、路径、已编辑、只读、冲突、已删除）、`Window()`、`Save()`、`SaveAs(path)`、`SaveAnyway()`、`Duplicate()`、`Rename(name)`、`MoveTo(dir)`、`Revert()`、`Close()`、`OnStateChange`、`OnReload`。
- 保存按 `NSDocument`/GIO 方式原子替换（先写旁路新文件、刷盘、再替换）并保留元数据；写入前对照上次读取/写入的修改时间与大小检测外部修改。`AutosaveInPlace` 在停止编辑片刻后保存、窗口失焦和关闭窗口时也保存；无标题、只读、已删除或被其他应用修改而无法就地保存的文稿不自动保存，关闭时询问。
- 有未保存更改且不会自动保存的文稿在停止编辑一秒后写恢复副本到 `PathUserData` 的 `Recovery/<kind>`，崩溃后以已编辑状态恢复。文稿监视自己的文件并在窗口获得焦点时重新检查：无未保存更改则重读（`OnReload`）；有更改则进入冲突，保存时询问；文件被重命名会跟踪；被删除后文稿仍打开并标记 `Deleted`，保存会再次写入。
- 菜单角色：`RoleNew`、`RoleOpen`、`RoleOpenRecent`、`RoleSave`、`RoleSaveAs`、`RoleDuplicate`、`RoleRename`、`RoleMoveTo`、`RoleRevert`、`RoleFileMenu` 作用于焦点窗口的文稿。网页窗口可 `mygo.BindAs("Note", docs.Service())`，页面通过生成的客户端调用 `content`、`setContent`、`state`、`save`、`saveAs`、`duplicate`、`rename`、`moveTo`、`revert`，并用 `onDocumentReload`、`onDocumentState` 跟踪。

**最小示例**
```go
mygo.NewDocuments(mygo.DocumentsOptions[string]{
	Types: []mygo.DocumentType{{Name: "Markdown Document", Extensions: []string{"md"}}},
	Read:  func(r io.Reader) (string, error) { b, err := io.ReadAll(r); return string(b), err },
	Write: func(w io.Writer, s string) error { _, err := io.WriteString(w, s); return err },
	Window: func(doc *mygo.Document[string]) *mygo.Window {
		return mygo.NewWindow(mygo.WindowOptions{Content: ui.View(app.editor(doc))})
	},
})
```

**限制/常见错误**：内容最好是由应用通过 `Set` 替换的值（字符串、值结构体或持久快照）；用就地修改的可变内容时必须提供 `Snapshot` 并在修改后调用 `Changed()`。`SetEdited(false)` 只把当前内容视为与文件一致，不写文件；保存期间用户继续编辑时，保存开始后发生的更改仍未保存。文稿窗口标题由文稿名决定，页面 `<title>` 不改变它；未保存标记在 macOS 是关闭按钮中的圆点、Linux/Windows 是标题前的 `*`。`AskOnQuit` 在 macOS 用附着式对话框逐个询问，Linux/Windows 用一个对话框列出。Windows 和 macOS 结束会话时不询问：文稿立即就地保存或保留，下次启动以已编辑状态恢复；不经询问销毁窗口（`Window.Destroy`、`App.Exit`）会保留恢复副本。文件夹形式的文稿（包）尚不支持。

## 启动器

**用途**：Spotlight、Raycast、PowerToys Run 类的后台应用：快捷键唤出面板、搜索并启动已安装应用，无 Dock 图标、单实例。

**关键 API**
- 后台应用：`RequestSingleInstanceLock`、`App.SetActivationPolicy(mygo.ActivationPolicyAccessory)`（macOS 无 Dock 图标，打包应用也可用 `macos.infoPlist.LSUIElement`）、`OnWindowAllClosed(func(){})`、`OnSecondInstance` 中重新显示面板。
- 面板：见“子窗口、浮出窗口、面板与点击穿透”；隐藏创建、用完隐藏，按指针所在显示器的 `WorkArea` 定位（`Screen.DisplayNearestPoint(Screen.CursorScreenPoint())`），并让搜索框获得焦点。
- 快捷键：`mygo.GlobalShortcut.Register("Alt+Shift+Space", fn)`；换绑时先 `Unregister` 再 `Register`，失败提示用户被占用；Wayland 由桌面绑定。
- 搜索：`Shell.InstalledApps()`（按名称排序、读取磁盘、几十毫秒，在 goroutine 调用）、`Shell.FileIcon(app.Path, size)`（尺寸为像素 = DIP × 显示器缩放；按路径缓存、只画可见行）、`Shell.LaunchApp(app)`（启动或切到前台，先隐藏面板让焦点回到原应用）。

**最小示例**
```go
if !mygo.App.RequestSingleInstanceLock() {
	return // 已运行的启动器显示自己
}
mygo.App.SetActivationPolicy(mygo.ActivationPolicyAccessory)
mygo.App.OnWindowAllClosed(func() {})
mygo.App.OnSecondInstance(func([]string, string) { showPanel() })
mygo.App.WhenReady(ready)
if err := mygo.App.Run(); err != nil {
	log.Fatal(err)
}
```

**限制/常见错误**：面板保持打开会取消退出，`OnClose` 要按退出标志放行；Windows 上全局快捷键的按键允许面板切到前台，Wayland 上由桌面绑定快捷键。`InstalledApps` 在 macOS 列出各 Applications 文件夹与 CoreServices 的应用、Windows 列出开始菜单“所有应用”、Linux 列出桌面条目；图标绘制要占用主线程几毫秒，应只画可见项并缓存结果。

## 速查

| 目标 | 做法 |
| --- | --- |
| 启动事件循环 | 主 goroutine 调 `mygo.App.Run()`，窗口在 `WhenReady` 或之后创建 |
| 一个实例 + 再次启动显示 | `RequestSingleInstanceLock()` + `OnSecondInstance` |
| 退出前确认 | `OnBeforeQuit`/`OnWillQuit` 中 `e.PreventDefault()`，取消后在 `OnQuitCanceled` 清理状态 |
| 关闭窗口到托盘 | `OnClose` 中按退出标志 `PreventDefault` + `Hide`，托盘 `OnClick(win.Show)` |
| 打开链接或文件 | `OnOpenURL` / `OnOpenFile`（在 `Run` 前注册）+ 配置 `urlSchemes` / `fileAssociations` |
| 记录最近文稿 | `App.AddRecentDocument(path)`，菜单用 `RoleOpenRecent` |
| 登录时启动 | `SetOpenAtLogin(true)`，`WasOpenedAtLogin()` 决定隐藏启动 |
| 读写数据 | `App.Path(mygo.PathUserData)` / `PathCache` / `PathLogs` |
| 记住窗口与页面 | `WindowOptions.StateKey`；`Page.SessionState()` + `PageOptions.SessionState` |
| 无闪烁显示 | `BackgroundColor` + `light-dark(...)`，或 `Hidden` + `OnReadyToShow(win.Show)` |
| 关闭前询问保存 | `OnClose` 中 `PreventDefault`，另起 goroutine 调 `Dialog.Message`，再 `Destroy` |
| 启动器面板 | `Panel: true` 窗口 + `GlobalShortcut.Register` + `ActivationPolicyAccessory` |
| 浮层/下拉菜单 | `NewFlyout`（`Parent` 或 `Tray`、`Anchor`、`Placement`、`Focusable`） |
| 打印 / PDF / 截图 | `Page.Print()`、`Page.PrintToPDF(PDFOptions)`、`win.CapturePage()` |
| 下载到指定目录 | `Page.OnWillDownload` 设置 `e.Path` 或 `e.PreventDefault()` |
| 隔离账号/私密浏览 | `NewSession` / `NewEphemeralSession` + `PageOptions.Session` |
| 清除登录状态 | `App.ClearBrowsingData()` 或 `Session.ClearData(SiteDataAll)` |
| 每个文件一个窗口 | `mygo.NewDocuments(mygo.DocumentsOptions[T]{...})`（`Run` 之前） |
| 原生窗口句柄 | `win.NativeHandle()`（`NSWindow*`、`GtkWindow*`、`HWND`）+ `mygo.RunOnMain` |
