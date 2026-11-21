# 菜单、托盘与桌面 API

MyGo 的菜单与操作系统集成：菜单模板与角色、菜单栏、上下文菜单（含 webview 标准菜单）、Dock 菜单、托盘图标、加速键与全局快捷键，以及对话框、通知、剪贴板、Shell、显示器、深色模式与辅助功能、电源、媒体键、网络与代理、机密、文件监视。这些 API 与窗口/页面 API 一样可从任意 goroutine 调用，回调在主线程运行。

## 源页面

| 源页面 | 本文件覆盖 |
| --- | --- |
| [`menus.md`](https://mygo.run/zh/docs/menus) | 菜单模板与角色、菜单栏、上下文菜单、Dock 菜单、托盘、快捷键 |
| [`native.md`](https://mygo.run/zh/docs/native) | 对话框、通知、剪贴板、Shell、显示器、深色模式、电源、媒体、网络、机密、监视文件、全局快捷键 |

## 菜单模板与角色

**用途**：用 `*mygo.MenuItem` 模板构建菜单栏、上下文菜单、Dock 菜单与托盘菜单；同一份模板用于三个平台，macOS 专用角色在其他平台被略去。

**关键 API**
- 菜单项字段：`Label`、`Accelerator`；`Click func(item *mygo.MenuItem, win *mygo.Window)`（主线程运行；Linux/Windows 的菜单栏与上下文菜单传所属窗口，其他菜单传当前焦点窗口，没有则为 nil）；`Type`（`MenuItemNormal`、`MenuItemCheckbox`、`MenuItemRadio`、`MenuItemSeparator`、`MenuItemSubmenu`；相邻单选按钮组成一组，复选框与单选按钮在 `Click` 运行前切换 `Checked`）；`Disabled`、`Hidden`、`ToolTip`；`Icon`；`Subtitle`（macOS 14 和浏览器显示为标签下方第二行，旧平台显示为 `Label — Subtitle`）；`ID`（`Menu.ItemByID` 查找）；`Role`。
- `mygo.NewMenu([]*mygo.MenuItem{...})`；`mygo.Separator()`；`Menu.Append`、`Menu.Insert`。
- 角色带常规标签与快捷键，`Label`/`Accelerator` 可覆盖：编辑 `RoleUndo`/`RoleRedo`/`RoleCut`/`RoleCopy`/`RolePaste`/`RoleDelete`/`RoleSelectAll`（macOS 还有 `RolePasteAndMatchStyle`）；格式 `RoleBold`/`RoleItalic`/`RoleUnderline`/`RoleStrikethrough`/`RoleAlignLeft`/`RoleAlignCenter`/`RoleAlignRight`；页面 `RoleReload`/`RoleForceReload`/`RoleToggleDevTools`/`RoleResetZoom`/`RoleZoomIn`/`RoleZoomOut`；窗口 `RoleToggleFullScreen`/`RoleMinimize`/`RoleZoom`/`RoleClose`；文稿 `RoleNew`/`RoleOpen`/`RoleOpenRecent`/`RoleSave`/`RoleSaveAs`/`RoleDuplicate`/`RoleRename`/`RoleMoveTo`/`RoleRevert`；应用 `RoleQuit` 与 macOS 的 `RoleAbout`/`RoleHide`/`RoleHideOthers`/`RoleUnhide`/`RoleFront`/`RoleServices`/`RoleStartSpeaking`/`RoleStopSpeaking`；整个菜单 `RoleFileMenu`/`RoleEditMenu`/`RoleFormatMenu`/`RoleViewMenu`/`RoleWindowMenu` 与 macOS 的 `RoleAppMenu`/`RoleWindow`/`RoleHelp`。

**最小示例**
```go
menu := mygo.NewMenu([]*mygo.MenuItem{
	{Role: mygo.RoleAppMenu},
	{Label: "File", Submenu: []*mygo.MenuItem{
		{Label: "New Note", Accelerator: "CmdOrCtrl+N", Click: newNote},
		mygo.Separator(),
		{Role: mygo.RoleClose},
	}},
	{Role: mygo.RoleEditMenu},
	{Label: "View", Submenu: []*mygo.MenuItem{
		{ID: "sidebar", Label: "Show Sidebar", Type: mygo.MenuItemCheckbox, Checked: true, Click: toggleSidebar},
		{Label: "List", Type: mygo.MenuItemRadio, Checked: true},
		{Label: "Grid", Type: mygo.MenuItemRadio},
	}},
	{Role: mygo.RoleWindowMenu},
})
```

**限制/常见错误**：编辑角色只在焦点支持相应操作时启用（原生 UI 有选中文本可剪切复制、有可撤销内容可撤销、文本输入可粘贴）；应用明确 `Disabled` 的菜单项始终禁用；macOS 按焦点验证菜单项，Linux/Windows 上页面的菜单项保持应用设置的状态。Windows 页面中"粘贴"会把焦点还给页面并通过 DevTools 运行 Chromium 的粘贴命令，页面会看到一个没有修饰键的 `Unidentified` 键事件，它的 Ctrl+V 处理器不会被调用；"粘贴并匹配样式"插入纯文本；给编辑角色指定页面收不到的按键时仍执行该角色的命令。

## 修改菜单项与图标

**用途**：菜单使用期间原地更新，并给菜单项加图标。

**关键 API**
- 菜单项方法：`SetLabel`、`SetEnabled`、`SetVisible`、`SetChecked`、`SetAccelerator`、`SetSubtitle`、`SetIcon`（nil 移除）。
- `mygo.NewMenuIcon(png []byte, template bool)`：按图像自身颜色制作图标，非 PNG 数据返回错误；图像等比放入 16×16 点的方框，至少提供 32×32 或 48×48 像素以在高分屏与 150%/200% 缩放下清晰；`template` 让黑白透明图像采用菜单文字颜色（Linux 按原样显示）。
- `mygo.NewMenuIconSVG(svg *ui.SVG, template bool)`：用原生 UI 的渲染器把 SVG 画成长边 64 像素的图像，最高 400% 缩放仍清晰，与 `ui.Icon` 用同一渲染器；模板模式只看形状。
- 同一图像只解码、缩放一次，供所有显示它的菜单项与重复打开的菜单复用，不再被引用后释放。

**最小示例**
```go
//go:embed icons/trash.svg
var trashSVG []byte
trashIcon, err := mygo.NewMenuIconSVG(ui.MustParseSVG(trashSVG), true)
menu.ItemByID("delete").SetIcon(trashIcon)
menu.ItemByID("sidebar").SetChecked(false)
```

**限制/常见错误**：平台差异——macOS 在勾选标记列之后显示图标；Windows 11 的菜单由 MyGo 绘制，图标有独立一列；Windows 10 把图标画在勾选列中，已勾选项显示带勾选边框的图标；Linux 在菜单栏菜单与 GTK 上下文菜单中显示图标，菜单栏本身的项不显示；Linux 托盘菜单的图标经 DBusMenu 提供，GNOME Shell 的 AppIndicator 扩展与 KDE Plasma 显示，其他面板可能不显示。原生 UI 绘制的菜单用菜单项的 `IconColor` 给模板图标单独上色。

## 菜单栏

**用途**：应用菜单栏；macOS 上属于应用，Linux/Windows 上属于窗口。

**关键 API**：`mygo.App.SetMenu(menu)`；`Window.SetMenu(menu)`（Linux/Windows 单窗口菜单）；`App.SetMenu(nil)` 移除。macOS 应用默认有应用、文件、编辑、视图、窗口菜单，直到设置自己的菜单栏——保留 `RoleAppMenu` 与 `RoleEditMenu`，Cmd+Q 与复制粘贴才能工作；Linux/Windows 的 `App.SetMenu` 给每个没有自己菜单的窗口，带文稿的应用默认得到文件与编辑菜单。

**最小示例**
```go
mygo.App.SetMenu(menu) // 在 main 中、App.Run 之前也可调用
```

**限制/常见错误**：`AutoHideMenuBar` 的窗口把菜单栏藏起来，单独按 Alt 或 F10 显示（Windows 上焦点落在菜单栏，Linux 上打开第一个菜单），选中一项或离开后再次隐藏，快捷键始终有效。没有标题栏的窗口（无边框或隐藏标题栏）没有菜单栏的位置：Alt/F10 仍从页面左上角以弹出菜单打开。全屏窗口隐藏菜单栏，Alt/F10 仍可打开，菜单打开期间菜单栏会显示。Windows 上菜单栏跟随浅色/深色外观；Windows 11 菜单看起来像系统菜单（更宽敞的项、圆角高亮、Acrylic 背景，透明效果关闭/节能/对比度主题/远程会话时退化为纯色背景加轮廓线）。

## 上下文菜单与 webview 菜单

**用途**：在鼠标处弹出自己的菜单；定制网页右键菜单。

**关键 API**
- `Menu.Popup(win)` 在窗口中鼠标位置显示，菜单关闭后返回；`Menu.PopupAt(win, position)` 在页面中的指定位置显示。从绑定方法里显示用 `mygo.CallerWindow(ctx)`；页面在 `contextmenu` 事件中 `preventDefault()` 后调用该方法。
- 页面没有自己的处理器时显示 webview 标准菜单；`Page.OnContextMenu(func(e *mygo.ContextMenuEvent))` 在标准菜单即将打开时修改它：`e.LinkURL`、`LinkText`、`ImageURL`、`MediaURL`、`SelectionText`、`Editable`、`FrameURL`；`e.Menu` 是 webview 菜单项的菜单，可隐藏、重命名、禁用、移除或追加应用自己的项；`e.PreventDefault()` 不显示菜单；把 `e.Menu` 换成应用的菜单则显示那个。webview 菜单项保留其行为，除非应用给其中一项设置 `Click`；页面自己的 `contextmenu` 处理器调用 `preventDefault` 时不显示任何菜单，监听器也不会被调用。
- webview 菜单项 ID（各支持平台相同；同组内后续条目省略 `ContextMenu` 前缀，实际常量如 `ContextMenuPasteAndMatchStyle`）：导航 `ContextMenuBack`/`Forward`/`Reload`/`Stop`；`ContextMenuInspect`；编辑 `ContextMenuUndo`/`Redo`/`Cut`/`Copy`/`Paste`/`PasteAndMatchStyle`/`Delete`/`SelectAll`；拼写 `ContextMenuSpellingGuess`；链接 `ContextMenuOpenLink`/`OpenLinkInNewWindow`/`DownloadLink`/`CopyLink`；图像 `ContextMenuOpenImageInNewWindow`/`DownloadImage`/`CopyImage`/`CopyImageURL`；媒体 `ContextMenuOpenMediaInNewWindow`/`DownloadMedia`/`CopyMediaURL`/`MediaControls`/`MediaLoop`/`MediaFullScreen`/`PictureInPicture`；其余 `ContextMenuOpenFrameInNewWindow`/`LookUp`/`SearchWeb`/`Share`/`Emoji`/`Print`。其他项以平台自己的名称作 ID（macOS 是 WebKit 标识符、Linux 是标准操作昵称、Windows 是 WebView2 名称），分隔符与未命名项没有 ID。
- 原生 UI 的窗口用 `ui.ContextMenu` 为元素提供上下文菜单。

**最小示例**
```go
page.OnContextMenu(func(e *mygo.ContextMenuEvent) {
	e.Menu.Remove(mygo.ContextMenuReload, mygo.ContextMenuInspect)
	if e.LinkURL != "" {
		e.Menu.Append(mygo.Separator(), &mygo.MenuItem{
			Label: "Open in Browser",
			Click: func(*mygo.MenuItem, *mygo.Window) { go mygo.Shell.OpenExternal(e.LinkURL) },
		})
	}
})
```

**限制/常见错误**：macOS 就地修改 WebKit 的菜单（查询、翻译、服务、拼写建议继续工作），但 WebKit 不告知右键单击了什么：MyGo 的运行时从页面自身的 `contextmenu` 事件获取，因此在 iframe 中单击不会提供信息、`FrameURL` 为空。Linux 就地修改 WebKitGTK 菜单，应用菜单项不显示快捷键或图标、单选项显示为复选框，只有 WebKitGTK 标准操作的项可重命名，`FrameURL` 为空。Windows 从 WebView2（1.0.1185+，更旧运行时不调用监听器）获得信息，被修改的菜单由 MyGo 绘制，WebView2 不提供拼写建议、`data:` 图像 URL 由页面事件提供。应用修改过的菜单会隐藏开头、结尾和连续重复的分隔符。

## Dock 菜单

**用途**：macOS 上给 Dock 图标的右键菜单加项。

**关键 API**：`App.Dock.SetMenu(menu)`，与 `App.SetMenu` 一样可在 `main` 中、`App.Run` 之前调用；菜单项加在右键 Dock 图标所打开菜单的标准项上方。

**最小示例**
```go
mygo.App.Dock.SetMenu(mygo.NewMenu([]*mygo.MenuItem{
	{Label: "New Window", Click: func(*mygo.MenuItem, *mygo.Window) { openMainWindow() }},
}))
```

**限制/常见错误**：仅 macOS 有效；其他平台上 `App.Dock` 的方法不做任何事（`SetMenu` 不报错）。

## 托盘图标

**用途**：macOS 菜单栏、Windows/Linux 通知区域里的常驻图标与菜单。

**关键 API**
- `mygo.NewTray(mygo.TrayOptions{Icon []byte, IconIsTemplate bool, ToolTip string, Menu *mygo.Menu})`，在应用就绪后创建；图标是 PNG，理想为 16×16 点（高分辨率用 32×32 像素的图像）；模板图标为透明背景上的黑色，macOS 会为浅色/深色菜单栏着色。
- 带菜单的托盘单击时显示菜单；没有菜单时单击进入 `OnClick`，右键进入 `OnRightClick`（macOS 上 Control 单击等同右键）。修改与移除：`SetIcon`、`SetMenu`、`SetToolTip`、`SetTitle`（macOS，图标旁文字）、`Destroy`。托盘菜单的项可显示图标，状态在菜单下次打开时生效。
- `Tray.PopUpMenu` 以编程方式打开 `SetMenu` 设置的菜单；只在右键显示菜单而保留主单击的写法：`OnRightClick` 中 `SetMenu`、调用 `PopUpMenu`、调用返回后 `SetMenu(nil)`。
- 多个图标：macOS 上新图标出现在已有图标左侧，用户按住 ⌘ 拖动排序，给唯一 `ID` 后重建（下次启动或 `Destroy` 之后）会回到用户放置的位置；Windows 上带 `ID` 的图标保持用户"显示在任务栏"的选择，与创建顺序无关。
- 彩色图标用 `Tray.Dark()` 判断所在栏是否深色（系统外观或壁纸影响的 macOS 菜单栏、Windows 任务栏），并在 `OnAppearanceChange` 中重设图标；模板图标自动适应。显示变化数字的标题要保持宽度：用数字空格（U+2007，与数字等宽；普通空格更窄）填充到固定字符数；字母仍不等宽，切换单位（如 `KB/s` 与 `MB/s`）仍会变宽。

**最小示例**
```go
tray, err := mygo.NewTray(mygo.TrayOptions{
	Icon: iconPNG, IconIsTemplate: true, ToolTip: "Notes",
	Menu: mygo.NewMenu([]*mygo.MenuItem{
		{Label: "Open Notes", Click: focusMainWindow},
		mygo.Separator(),
		{Role: mygo.RoleQuit},
	}),
})
if err != nil {
	log.Println("tray:", err)
}
```

**限制/常见错误**：托盘图标不会让应用保持运行——最后一个窗口关闭时应用退出，要"关闭到托盘"应隐藏窗口：`OnBeforeQuit` 先于窗口关闭运行，用标志区分退出与用户关闭，在 `OnClose` 中 `PreventDefault()` + `Hide`，托盘 `OnClick(win.Show)` 恢复。macOS 常驻菜单栏的应用用 `App.SetActivationPolicy(mygo.ActivationPolicyAccessory)` 去掉 Dock 图标。Linux 用 AppIndicator，需要 `libayatana-appindicator3`，没有该库时 `NewTray` 返回错误；带 `OnClick` 的图标响应 `activate`（旧库版本不报告单击），GNOME 的 AppIndicator 扩展在双击时发送激活，超过双击间隔的单击打开菜单；桌面偏好深色时模板图标为白色，`Tray.Dark` 读该桌面偏好（独立于 `Theme.SetSource`，没有门户时为 false），AppIndicator 无法报告面板自己的主题。

## 加速键与全局快捷键

**用途**：定义菜单快捷键，并在应用位于后台时接收按键。

**关键 API**
- 语法：修饰键与按键用 `+` 连接，不区分大小写。修饰键 `CmdOrCtrl`（`CommandOrControl`；macOS/iOS 为 Command，其他为 Control）、`Cmd`/`Command`/`Super`/`Meta`/`Win`、`Ctrl`/`Control`、`Alt`/`Option`/`AltGr`、`Shift`。按键：一个字符（字母、数字或 `,`、`/` 等符号，`+` 写作 `Plus`）、`F1`–`F24`、`Enter`/`Return`、`Tab`、`Space`、`Backspace`、`Delete`/`Del`、`Insert`/`Ins`、`Escape`/`Esc`、方向键与 `Home`/`End`/`PageUp`/`PageDown`、`Num0`–`Num9`/`NumDec`/`NumAdd`/`NumSub`/`NumMult`/`NumDiv`、`VolumeUp`/`VolumeDown`/`VolumeMute`/`MediaPlayPause`/`MediaNextTrack`/`MediaPreviousTrack`/`MediaStop`/`PrintScreen`/`CapsLock`/`NumLock`/`ScrollLock`。
- `mygo.GlobalShortcut.Register(accel, fn) error`，回调在主线程；`Unregister`、`UnregisterAll`、`IsRegistered`；快捷键已被占用（可能是应用自己，也可能是另一个应用）时 `Register` 失败。

**最小示例**
```go
err := mygo.GlobalShortcut.Register("CmdOrCtrl+Shift+Space", func() {
	win.Show()
	win.Focus()
})
```

**限制/常见错误**：X11 下应用自己抓取按键；Wayland 下应用无法自己抓取，桌面通过 XDG 桌面门户绑定（KDE Plasma 6、GNOME 48+），需要应用已安装（按桌面条目识别），桌面可能要求用户确认新快捷键并允许改选，快捷键在 `Register` 返回后不久生效；在 Wayland 会话中通过 XWayland（`GDK_BACKEND=x11`）运行的应用，自己抓取的按键只在它拥有焦点时收到。函数显示或聚焦的窗口，在桌面提供激活令牌时（如 GNOME）会获得焦点。启动器换绑快捷键时先 `Unregister` 再 `Register`，失败要提示用户已被占用。

## 对话框

**用途**：系统的文件选择、保存与消息对话框。

**关键 API**
- `mygo.Dialog.Open(mygo.OpenDialogOptions{Parent, Title, Filters []mygo.FileFilter{{Name, Extensions, MIMETypes}}, Multiple, Directory})`：取文件（取消返回 nil）、选目录。`Save(mygo.SaveDialogOptions{Parent, DefaultPath})`：选择保存位置，取消返回 `""`。
- `mygo.Dialog.Message(mygo.MessageOptions{Parent, Type, Message, Detail, Buttons, DefaultButton, CancelButton, DestructiveButtons, CheckboxLabel})`：返回 `res.Button`（按钮下标）。`mygo.Dialog.Error(title, detail)` 报告错误。
- 调用阻塞到用户作出选择：绑定方法在独立 goroutine 中可直接调用；事件监听器和菜单项回调要尽快返回，应另起 goroutine。`Parent` 指定对话框所属窗口，macOS 上显示为附着式对话框（sheet）。
- 文件过滤器可按类型匹配（类似 HTML 的 `accept`），如 `MIMETypes: []string{"image/*"}` 或 `"application/pdf"`；Windows 上在选择目录的对话框中选择库（如"文档"）会得到该库保存到的文件夹。

**最小示例**
```go
res, err := mygo.Dialog.Message(mygo.MessageOptions{
	Parent: win, Type: mygo.MessageWarning,
	Message: "Delete 3 notes?", Detail: "You cannot undo this.",
	Buttons: []string{"Cancel", "Delete"}, DestructiveButtons: []int{1},
})
if res.Button == 1 {
	deleteNotes()
}
```

**限制/常见错误**：消息框中 Enter 激活 `DefaultButton`、Escape 激活 `CancelButton`（即使两者是同一个按钮）；macOS 上 Cmd+. 也取消，`DefaultButton` 为 -1 时 Enter 不激活任何按钮。`CancelButton` 默认为第一个标有 Cancel 或 No 的按钮，否则为第一个按钮；Delete/Keep 场景设 `CancelButton: 1` 以让 Escape 选 Keep。文件过滤器在 macOS 11+ 映射为 `UTType`（`image/*` 等大类映射为 `public.image` 等），更早的 macOS 只接受这些类型；GTK 匹配类型与通配符；Windows 按注册表扩展名过滤。

## 通知

**用途**：系统通知，支持点击回跳、更新、分组与图像。

**关键 API**
- `mygo.NewNotification(mygo.NotificationOptions{Title, Body, ID, Group, Image, Silent, Sender})`；`n.Show() error`、`n.Close()`；`mygo.NotificationsSupported()` 报告应用能否显示通知；`mygo.ClearNotifications()` 移除应用的所有通知。
- `App.OnNotificationClick(func(id string))`：带 `ID` 的通知被点击时收到 ID，也包括应用因这次点击而启动的那一次（处理器要在 `Run` 之前注册）；用仍在显示的通知的 `ID` 再显示会替换原通知；`Group` 把通知归入通知中心同一个堆叠（macOS）。
- `Image` 为 PNG 或 JPEG 字节（如曲目封面）：macOS 作为缩略图，Linux 交给桌面通知服务并缩到 256 像素以内，Windows 作为气泡大图标；`NewNotification` 复制图像，MyGo 不获取远程图像。`Sender` 见限制。
- 页面也可以请求通知权限（由页面的权限处理器决定，见窗口页面相关文档）。

**最小示例**
```go
n := mygo.NewNotification(mygo.NotificationOptions{
	ID: "now-playing", Title: track.Title, Body: track.Artist, Image: track.Cover, Silent: true,
})
n.OnClick(func() { mainWindow().Focus(); n.Close() })
if err := n.Show(); err != nil {
	log.Println(err)
}
```

**限制/常见错误**：macOS 只显示打包应用的通知（`mygo dev`、`mygo build` 的产物是打包应用），第一次 `Show` 会询问用户是否允许并等待回答；未获允许时返回 `mygo.ErrNotificationsDenied`，`ShowContext(ctx)` 可在 ctx 结束时停止等待并撤掉之后才显示的通知。macOS 通知被点击后仍留在通知中心，直到应用 `Close` 它；可在 `OnDidBecomeActive` 中调用 `ClearNotifications()`。macOS 12+ 的发送者头像需要 `Sender: &mygo.NotificationSender{ID, Name, Avatar}`，并启用 Communication Notifications、把 `INSendMessageIntent` 加入 `macos.infoPlist.NSUserActivityTypes`、用 `com.apple.developer.usernotifications.communication` 签名并嵌入匹配描述文件（`macos.provisioningProfile`）；缺权限或数据无效时回退为普通通知，仅用于收到的私信。

## 剪贴板与 Shell

**用途**：读写剪贴板，用系统默认程序打开/定位/回收文件。

**关键 API**
- 剪贴板：`Clipboard.WriteText`/`ReadText`、`WriteHTML`、`WriteImage(png)`/`ReadImage`、`AvailableFormats`、`Clear`；`Clipboard.Write(transfer.Data)`、`Read(formats...)`、`Formats`、`ReadFormat`、`ReadFiles`、`WriteFiles`、`Flush` 使用与原生拖放相同的数据模型（有序数据项、同一数据的多种表示、文件列表、自定义 MIME 格式、按需提供数据的提供者）。页面可用 `navigator.clipboard`。
- Shell：`Shell.OpenExternal(url)`（默认浏览器）、`OpenPath(path)`（默认应用）、`ShowItemInFolder(path)`（Finder/资源管理器/文件管理器）、`TrashItem(path)`（移到废纸篓）、`PreviewItem(path)`（Quick Look；仅 macOS，其他平台返回 `ErrUnsupported`）、`Beep()`、`FileIcon(path, size)`、`TypeIcon(ext, size)`、`InstalledApps()`、`LaunchApp(app)`。
- `FileIcon` 返回文件管理器显示的图标（PNG），尺寸就是请求的像素数（DIP × 显示器缩放）；`TypeIcon` 按扩展名返回类型图标，磁盘上不需要存在该类型的文件；`InstalledApps` 返回 `Name`、`ID`、`Path` 三个字段，按名称排序。

**最小示例**
```go
mygo.Clipboard.WriteText("hello")
png, err := mygo.Shell.FileIcon(path, 64)
mygo.Shell.ShowItemInFolder(path)
```

**限制/常见错误**：`TrashItem` 从不永久删除——无法移入废纸篓时失败并留下原文件；在 Windows 上，大到回收站放不下的项、位于没有回收站的驱动器（网络共享、许多可移动驱动器）上的项属于这种情况，否则 Windows 会不经询问直接删除。`PreviewItem` 的面板不是模态的，应用继续运行，再次请求会替换它当前显示的项。`FileIcon` 每次绘制要占用主线程几毫秒，只画可见项并缓存；Windows 上该路径也可指向 shell 项（如 `shell:AppsFolder\<AppUserModelID>`）。`InstalledApps` 读取磁盘、几十毫秒，应在 goroutine 中调用；`LaunchApp` 启动应用或把它带到前台。

## 显示器、深色模式与辅助功能

**用途**：读取多显示器布局、跟随系统外观与无障碍设置。

**关键 API**
- `mygo.Screen`：`PrimaryDisplay()`、`Displays()`、`CursorScreenPoint()`、`DisplayNearestPoint(point)`、`DisplayMatching(rect)`（矩形重叠最多的显示器）、`OnDisplaysChanged(fn)`；显示器有 `Bounds`、`WorkArea`（不含菜单栏、Dock 和任务栏）、`ScaleFactor`；坐标以 DIP 计、原点在主显示器左上角。
- 深色模式：`Theme.IsDark()`、`Theme.OnUpdated(fn)`、`Theme.SetSource(mygo.ThemeDark/ThemeLight/ThemeSystem)`；`SetSource` 可在 `App.Run` 之前调用，应用以该外观启动；页面通过 `prefers-color-scheme` 跟随。
- 强调色与辅助功能：`Theme.Preferences()` 一次返回 `Accent`（`color.NRGBA`，A 为 0 表示桌面没有强调色）、`ReduceMotion`、`ReduceTransparency`、`HighContrast`、`TextScale`（1 为常规大小）；`Theme.OnUpdated` 在这些值变化或调用 `SetSource` 时触发；MyGo 只读取一次并保留到下次变化。

**最小示例**
```go
primary := mygo.Screen.PrimaryDisplay()
fmt.Println(primary.Bounds, primary.WorkArea, primary.ScaleFactor)
mygo.Theme.SetSource(mygo.ThemeDark) // 或 ThemeLight、ThemeSystem
p := mygo.Theme.Preferences()
toolbarColor := ui.RGB(p.Accent.R, p.Accent.G, p.Accent.B)
```

**限制/常见错误**：Windows 多显示器可有不同缩放，每个显示器的坐标从其左上角开始，起点是像素除以所有显示器中的最小缩放比，再按自身缩放延伸，因此缩放不同的显示器之间会留下空隙；窗口边界采用它大部分所在显示器的缩放比。`Theme.Preferences` 需要应用正在运行（在 `App.Run` 之前的 `main` 中调用会 panic），从其他 goroutine 调用会等待应用启动。Linux 桌面可能没有降低透明度与强调色（对应值为 false / A 为 0），`TextScale` 在 macOS 固定为 1；`ui` 的默认主题已跟随强调色、对比度与文本缩放，`Animate` 跟随减弱动态效果。页面通过 `mygo-runtime` 的 `preferences()` 及 `--mygo-accent`、`--mygo-text-scale` CSS 变量获取同一份数据。

## 电源、媒体键、网络与代理

**用途**：长任务保持唤醒、响应系统电源事件、在系统媒体控件中显示播放状态，并感知网络与代理。

**关键 API**
- 电源：`Power.KeepAwake(reason string, keepDisplayOn bool)` 返回 release 函数；`Power.KeepRunning(reason string)` 返回 stop 函数；`Power.OnSuspend`、`OnResume`、`OnLockScreen`、`OnUnlockScreen`；`Power.IsOnBattery()`；`Power.IdleTime()`（自上次键盘或鼠标输入）。传的原因会显示在系统列出保持唤醒项的地方。
- 媒体：`NowPlaying.Set(NowPlayingInfo{Title, Artist, Album, Artwork, Duration, Commands})`、`NowPlaying.OnCommand(func(cmd mygo.MediaCommand, pos time.Duration))`、`NowPlaying.SetPlayback(mygo.PlaybackPlaying/PlaybackPaused, pos)`、`NowPlaying.Clear()`；`NowPlaying.Control(player *audio.Player)` 返回 release；浏览器中就是页面的 Media Session。
- 网络：`Network.Status()` 返回 `{Online, Metered}`；`Network.OnChanged(func(mygo.NetworkStatus))`（主线程运行，返回移除函数）；页面已有 `navigator.onLine`。
- 代理：`github.com/egoist/mygo/netproxy` 的 `netproxy.Transport()` 作为 `http.Client` 传输层，或设置 `t.Proxy = netproxy.Proxy`，让 Go 请求像页面一样遵循系统代理。

**最小示例**
```go
release := mygo.Power.KeepAwake("Exporting the video", false)
defer release()
mygo.Network.OnChanged(func(s mygo.NetworkStatus) {
	if s.Online && !s.Metered {
		syncNow()
	}
})
client := &http.Client{Transport: netproxy.Transport(), Timeout: time.Minute}
```

**限制/常见错误**：`KeepAwake` 让计算机保持唤醒；`KeepRunning` 只阻止 macOS 的 App Nap（窗口全被隐藏、最小化、遮挡或锁屏时应用被降优先级、合并计时器、限制 I/O），但允许计算机空闲休眠，其他平台上 `KeepRunning` 不做任何事；macOS 上只要托盘图标还显示在菜单栏，就算有窗口显示。Linux 上屏幕保护程序启动时触发 `OnLockScreen`，不上报空闲时间的桌面 `IdleTime` 为 0。`NowPlaying` 只在 `Commands` 中提供的命令会到达 `OnCommand`（`Commands` 为 0 时提供播放、暂停和停止）；`PlaybackPlaying` 时系统推进位置，应用只在跳转时重设；`Control` 监听播放器播放结束，应用自己改变状态后用 `NowPlaying.Sync` 同步。`Network.Online` 表示系统有到计算机之外某个网络的连接，不证明服务器会应答（强制门户或路径中断也算在线）；`Metered` 在 macOS 是高费用路径或低数据模式、Linux 按流量计费、Windows 费用固定/可变或漫游或接近数据上限。`netproxy` 的顺序是回环直连、环境变量（`HTTP_PROXY` 等）优先、否则系统设置；系统按 URL 的答案缓存 5 秒；支持 HTTP、HTTPS、SOCKS5 代理，不支持 SOCKS4；只在系统返回的代理设置含密码时才发送密码。

## 机密与文件监视

**用途**：把令牌放进系统凭据存储；监视文件与目录变化并得到一批差异事件。

**关键 API**
- 机密：`mygo.Secrets.Set(key string, value []byte) error`、`Get(key) ([]byte, error)`、`Delete(key) error`；找不到时为 `mygo.ErrSecretNotFound`；一个机密最多 `mygo.MaxSecretSize`（2560）字节，键最多 256 字节。服务名为 `mygo.config.ts` 的 `identifier`，`mygo dev` 用 `<identifier>.dev`，没有标识符的应用用其名称；`Secrets.Service(name)` 可选择另一个服务，与 CLI 等程序共享机密或读取 Electron 应用用 keytar 保存的机密。方法可在 `Run` 之前调用、可从任意 goroutine 调用，系统询问（如解锁钥匙串）期间主线程继续处理事件。
- 文件监视：`mygo.WatchFiles(ctx, path, mygo.FileWatchOptions{Recursive, Ignore func(path string, isDir bool) bool, Debounce, MaxDelay, MaxEntries, HintsOnly})` 在所有内容进入监视后返回；`w.Next(ctx)` 等待下一批事件，应放在单独的 goroutine；`w.Close()`、`w.Root()`、`w.Done`。
- 事件：`e.Op` 为 `mygo.FileCreate`、`FileWrite`、`FileRemove` 或 `FileRename`；`e.Path` 相对被监视内容（`w.Root()`，解析符号链接后的路径）、用斜杠分隔、自身为 `.`；重命名时 `e.OldPath` 给出原位置。把事件当作"需要重新读取文件"的提示，而不是造成差异的每一步。

**最小示例**
```go
w, err := mygo.WatchFiles(ctx, dir, mygo.FileWatchOptions{
	Recursive: true, HintsOnly: false,
	Ignore: func(p string, isDir bool) bool { return isDir && path.Base(p) == ".git" },
})
if err != nil {
	return err
}
go func() {
	defer w.Close()
	for {
		events, err := w.Next(ctx)
		if err != nil {
			return // 关闭或目录消失后为 io.EOF
		}
		for _, e := range events {
			log.Println(e.Op, e.Path)
		}
	}
}()
```

**限制/常见错误**：机密能防的是其他用户和被盗磁盘，不防同一用户的其他应用：macOS 在另一个应用读取前询问用户（以 ad hoc 签名或 `go run` 的应用每次重新构建都会再问），Linux 和 Windows 允许它们读取；没有凭据存储时方法失败并返回满足 `errors.Is(err, errors.ErrUnsupported)` 的错误（浏览器中，以及 Linux 上没有 libsecret、没有会话总线或会话总线上没有 Secret Service 时），应用可退化为每次询问用户。文件监视只报告大小/修改时间变化或系统报告写入、或被替换；权限与其他元数据（索引器、杀毒软件会改）和目录写入不报告；存在的内容不报告；50 ms 无动静后报告（`Debounce` 为负数时立即），最长 500 ms（`MaxDelay`）；`Next` 取走上一批前不会再次查看，期间变化进入下一批，无权限的目录被跳过；目录（或文件所在目录）被移除/重命名/替换后最后一批有 `.` 的 `FileRemove`，随后 `Next` 返回 `io.EOF`；条目超过 `MaxEntries`（macOS 65536，其他地方 8192）时以 `mygo.ErrWatchLimit` 停止。`HintsOnly` 只报告 `FileHint`（只给目录，不 stat、无快照，适合开发重载器），`Ignore` 只对目录调用，`MaxEntries` 统计纳入的目录数；需要知道哪个文件变化时绝不要用它。

## 速查

| 目标 | 做法 |
| --- | --- |
| 应用菜单栏 | `App.SetMenu(mygo.NewMenu(...))`，保留 `RoleAppMenu`、`RoleEditMenu` |
| 单窗口菜单栏 | Linux/Windows 用 `Window.SetMenu` |
| 右键菜单 | `Menu.Popup(win)` / `PopupAt(win, pos)`，页面 `contextmenu` 中 `preventDefault` 后调用 |
| 改 webview 右键菜单 | `Page.OnContextMenu` 修改 `e.Menu` 或 `e.PreventDefault()` |
| Dock 菜单 | macOS `App.Dock.SetMenu` |
| 托盘图标 | `NewTray(TrayOptions{Icon, IconIsTemplate, Menu})` + `OnClick`/`OnRightClick` |
| 托盘点击弹菜单 | `Tray.PopUpMenu`，或 `OnRightClick` 中临时 `SetMenu` 后清除 |
| 关闭窗口到托盘 | `OnBeforeQuit` 设标志，`OnClose` 中 `PreventDefault` + `Hide` |
| 全局快捷键 | `GlobalShortcut.Register("CmdOrCtrl+Shift+Space", fn)`，失败时快捷键被占用 |
| 文件/消息对话框 | `Dialog.Open`、`Dialog.Save`、`Dialog.Message`、`Dialog.Error`（阻塞，放 goroutine） |
| 系统通知 | `NewNotification(...).Show()` + `OnClick`/`App.OnNotificationClick` |
| 剪贴板 | `Clipboard.WriteText/ReadText`、`WriteImage`、`AvailableFormats` |
| 打开文件/网页 | `Shell.OpenExternal`、`OpenPath`、`ShowItemInFolder`、`TrashItem` |
| 多显示器布局 | `Screen.PrimaryDisplay()`、`WorkArea`、`DisplayNearestPoint` |
| 跟随深色模式 | `Theme.SetSource(ThemeSystem)`、`Theme.IsDark()`、`OnUpdated` |
| 强调色与无障碍 | `Theme.Preferences()` 的 `Accent`、`ReduceMotion`、`HighContrast`、`TextScale` |
| 长任务保持唤醒 | `Power.KeepAwake(reason, displayOn)` / `KeepRunning` |
| 媒体键与系统媒体控件 | `NowPlaying.Set` + `OnCommand` + `SetPlayback` |
| 网络状态 | `Network.Status()` 的 `Online`/`Metered` + `OnChanged` |
| Go 请求走系统代理 | `netproxy.Transport()` |
| 保存令牌 | `Secrets.Set/Get/Delete`（`ErrSecretNotFound`、`MaxSecretSize`） |
| 监视文件变化 | `WatchFiles(...)` + `w.Next(ctx)` 的 `FileCreate`/`Write`/`Remove`/`Rename` |
