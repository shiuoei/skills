# MyGo API 速查

本文件是索引，不是完整文档：具体行为、参数与限制见各参考文件与官方文档（<https://mygo.run/zh/docs>）；实际 API 以 `go doc` 为准：

```sh
go doc github.com/egoist/mygo
go doc github.com/egoist/mygo/ui
```

## mygo 顶层包

### App（`mygo.App`，应用单例）

| 类别 | API |
| --- | --- |
| 启动 | `mygo.App.Run() error`、`mygo.App.WhenReady(func())`、`mygo.App.IsReady() bool` |
| 生命周期 | `mygo.App.Quit()`、`mygo.App.Exit(code int)`、`mygo.App.Relaunch(args ...string)`、`mygo.App.IsPackaged() bool` |
| 事件 | `OnReady`、`OnWindowCreated`、`OnWindowAllClosed`、`OnBeforeQuit`、`OnWillQuit`、`OnQuit`、`OnQuitCanceled`、`OnActivate`、`OnSecondInstance`、`OnOpenURL`、`OnOpenFile`、`OnLocaleChanged`、`OnNotificationClick` |
| 单实例 | `mygo.App.RequestSingleInstanceLock() bool` |
| 名称/版本 | `mygo.App.Name()`、`mygo.App.SetName`、`mygo.App.Version()`、`mygo.App.SetVersion` |
| 语言 | `mygo.App.Locale()`、`mygo.App.Languages()`、`mygo.App.SetLocale`、`mygo.App.SetLanguages`、`mygo.App.OnLocaleChanged` |
| 目录 | `mygo.App.Path(name PathName) (string, error)`、`mygo.App.SetPath`；常量 `PathHome`、`PathAppData`、`PathUserData` 等 |
| 菜单 | `mygo.App.SetMenu(*Menu)`、`mygo.App.Menu()` |
| 启动项 | `mygo.App.OpenAtLogin() bool`、`mygo.App.SetOpenAtLogin(bool) error`、`mygo.App.WasOpenedAtLogin()` |
| 标准界面 | `mygo.App.Hide()`、`Show()`、`Focus()`、`SetBadgeCount`、`ShowAboutPanel`、`ActivationPolicy` |
| 最近文稿 | `AddRecentDocument`、`RecentDocuments`、`ClearRecentDocuments` |
| 其他 | `IsDev()`、`WebViewAvailable() bool`、`RunOnMain(func())`、`Version`（包级常量）、`ClearNotifications()` |

### 窗口（`mygo.NewWindow(opts WindowOptions) *Window`）

- 创建后：`Show/Hide/Focus/Close/Destroy`、`Center/Maximize/Minimize`、`SetIcon(png []byte)`、`SetSize`、`SetTitle`、`SetAlwaysOnTop`、`SetFullScreen`、`Invalidate()`。
- 事件（返回取消订阅函数）：`OnClose(fn func(*CloseEvent))`、`OnClosed`、`OnFocus/OnBlur`、`OnHide/OnShow`、`OnMove/OnResize`、`OnMinimize/OnMaximize`、`OnFileDrop`、`OnReadyToShow`。
- 查询：`ID()`、`IsDestroyed()`、`IsVisible()`、`IsFocused()`、`Bounds()`、`ContentBounds()`、`ContentSize()`、`NativeHandle()`。
- Web 窗口：`w.Page()`（导航、Cookie、下载、打印等）、`w.NewWebView`、`w.CapturePage()`。
- 原生 UI 窗口：`WindowOptions.Content: ui.View(view)`；窗口选项见 `references/04-ui-foundations.md` 的「原生 UI 的窗口」。
- 其他类型：`mygo.NewFlyout(opts)`（浮层窗口）、`mygo.Windows()`、`mygo.WindowByID(id)`、`mygo.FocusedWindow()`。

### 托盘、菜单、菜单项

```go
tray, err := mygo.NewTray(mygo.TrayOptions{ID: "main", Icon: png32, ToolTip: "Your App", Menu: menu})
tray.SetMenu(m); tray.SetToolTip("…"); tray.SetIcon(png, false); tray.OnClick(fn); tray.Destroy()
```

- `Menu`：`mygo.NewMenu([]*mygo.MenuItem)`；`mygo.Separator()`。
- `MenuItem` 字段：`ID`、`Label`、`Role`（`mygo.RoleAbout` 等预定义行为）、`Accelerator`（如 `"CmdOrCtrl+Shift+N"`）、`Disabled`、`Hidden`、`Checked`、`Submenu`、`Click func(item *MenuItem, win *Window)`。
- 状态更新：`it.SetEnabled/SetChecked/SetLabel`（加入菜单后用它，而不是改字段）。
- 托盘图标：PNG 宽度即 HICON 尺寸；建议 32×32 像素（`IconIsTemplate` 用于 macOS 模板图标）。

### 桌面 API（包级变量/构造函数）

| 名称 | 用途 |
| --- | --- |
| `mygo.Dialog` | 打开/保存文件、消息框等对话框 |
| `mygo.NewNotification(NotificationOptions)` | 系统通知（`Show`、`OnClick`）；未打包/无权限时有错误 |
| `mygo.Clipboard` | 文本/HTML/图片/文件列表/自定义格式读写 |
| `mygo.Shell` | 用系统默认程序打开 URL/文件、显示文件、打开终端等 |
| `mygo.Screen` | 显示器与光标位置 |
| `mygo.Theme` | 深/浅色主题来源（`SetSource`）与查询 |
| `mygo.Power` / `mygo.Network` | 电源状态、网络状态 |
| `mygo.GlobalShortcut` | 全局快捷键注册 |
| `mygo.Protocol` | 自定义 URL 协议（`http.Handler` 服务） |
| `mygo.Updater` | 底层更新 API（一般用 `plugins/updater`） |
| `mygo.Cookies` | 浏览数据/Cookie 清理 |

### 绑定与前端（Web 窗口）

- `mygo.Bind(services ...any)`、`mygo.BindAs(name, service)`：把 Go 值暴露给前端；导出方法生成 TypeScript 客户端。
- 事件：`mygo.NewEvent[T]()` 声明类型化事件，前端订阅。
- `mygo.SetFrontend(fs.FS)`、`mygo.FileServer(fs.FS)`、`mygo.GenerateTypeScript()` / `WriteTypeScript(path)`。
- `mygo.isdev` → `mygo.IsDev()`；调试 `win.Page().OpenDevTools()`（开发构建）。

## ui 包

### 入口

| API | 用途 |
| --- | --- |
| `ui.View(view func(*ui.Context))` | 把视图函数交给 `WindowOptions.Content` |
| `ui.Context` | 每帧传入；主题、语言、时钟、输入等上下文 |
| `ui.Element` | 16 字节的受检句柄；链式方法返回自身；跨帧要持久化身份用 `Handle` |
| `ui.NewTester(view, w, h)`、`ui.NewTesterWith(view, w, h, TesterOptions{})` | 无窗口视图测试 |
| `ui.Render(view, w, h, scale)` | 渲染一帧为 `*image.RGBA`（快照/导出） |

### 容器与布局

`ui.Column`、`ui.Row`、`ui.Box`、`ui.Grid`、`ui.GridView`、`ui.Scroll`、`ui.ScrollBoth`、`ui.ScrollHorizontal`、`ui.Split`、`ui.SplitPercent`、`ui.SplitVertical`、`ui.Tabs`、`ui.Sidebar`、`ui.SidebarSection`、`ui.SidebarItem`、`ui.Toolbar`、`ui.Accordion`、`ui.AccordionItem`、`ui.Collapsible`、`ui.Fieldset`、`ui.Form`、`ui.Divider`、`ui.Spacer`、`ui.FitContent`。

### 文本与展示

`ui.Text`、`ui.Textf`、`ui.RichText`、`ui.Link`、`ui.Image`、`ui.Icon`、`ui.Avatar`、`ui.Badge`、`ui.Breadcrumbs`、`ui.Video`。富文本编辑器、Markdown、代码编辑器、差异查看器和图表在独立子包里（见下）。

### 独立子包（富文本、编辑器、Markdown、图表、终端）

| 包 | 入口 | 说明 |
| --- | --- | --- |
| `ui/richtext` | `richtext.TextOf(...)`、`richtext.Editor(c, doc, opts...)`、`richtext.EditorBase(c, doc, opts...)` | 带属性的富文本与编辑器（格式、链接、段落样式、图片、Markdown 导出） |
| `ui/editor` | `editor.NewBuffer(src)`、`editor.View(c, *editor.State)`、`editor.DiffView(c, *editor.DiffState)` | 代码编辑器与差异查看器（高亮、多插入符、折叠、补全、诊断、修订操作） |
| `ui/markdown` | `markdown.View(c, src, *markdown.Options)` | 原生 UI 里显示 Markdown（CommonMark/GFM、脚注、公式、流式解析） |
| `ui/chart` | `chart.NewBand(...)`、`chart.NewLinear(...)`、`chart.Plot(c, chart.Axes{...}, fn)`、`chart.Extent(...)` | 无外观图表：标度、坐标轴、标记、折线/面积/饼图路径、堆叠、选择 |
| `plugins/terminal` | 官方插件 | 原生 UI 里的终端（运行 shell 或任意程序） |

### 动作与选择

`ui.Button`、`ui.PrimaryButton`、`ui.ButtonBase`、`ui.MenuButton`、`ui.SplitButton`、`ui.PrimarySplitButton`、`ui.Toggle`、`ui.ToggleGroup`、`ui.Segmented`、`ui.Checkbox`、`ui.CheckboxState`、`ui.CheckboxGroup`、`ui.Switch`、`ui.Radio`、`ui.RadioGroup`、`ui.Select`、`ui.Combobox`、`ui.Autocomplete`、`ui.Slider`、`ui.RangeSlider`、`ui.StepSlider`、`ui.Stepper`、`ui.Rating`、`ui.ColorPicker`、`ui.ColorWell`。

### 输入与编辑器

`ui.TextInput`、`ui.TextArea`、`ui.NumberInput`、`ui.DecimalInput`、`ui.SearchField`、`ui.TokenField`、`ui.EditableText`、`ui.FindBar`、`ui.Calendar`、`ui.DateInput`、`ui.TimeInput`（富文本编辑用 `richtext.Editor`，见上）。

### 集合与导航

`ui.List`、`ui.Table`、`ui.Tree`、`ui.TreeItem`、`ui.Outline`、`ui.OutlineTable`、`ui.Tabs`、`ui.Sidebar`、`ui.BackButton`、`ui.ForwardButton`、`ui.Router`（`ui.NewRouter(path)`；页面用 `r.Match`/`r.Param` 匹配）。

### 对话框、浮层与状态

`ui.DialogBase`、`ui.AlertDialog`、`ui.Popover`、`ui.PopoverMenuButton`、`ui.ContextMenuBase`、`ui.TooltipBase`、`ui.ToastBase`、`ui.ToastViewportBase`、`ui.Progress`、`ui.Spinner`、`ui.Meter`。

### 绘制、动画与主题

`ui.NewCanvas`、`ui.NewBitmap`、`ui.DecodeImage/DecodeBitmap`、`ui.Painter`（Fill/描边/渐变/路径/绘制对象）、`ui.LightTheme/DarkTheme`、`ui.Themed`、`ui.Animate*`、`ui.EaseIn/EaseOut/EaseInOut/Linear`、`ui.Rotation/Scaling/Translation`。

### Element 常用链式方法

外观与尺寸：`Fill()`、`FillWidth()`、`FillHeight()`、`Center()`、`Width()`、`Height()`、`Gap()`、`Padding()`、`Radius()`、`Background()`、`Border()`、`BorderWidth()`、`Opacity()`、`Visible()/Hidden()`。
布局：`AlignItems()`、`AlignContent()`、`AlignSelf()`、`Grow()`、`Shrink()`、`Basis()`、`MinWidth()/MaxWidth()`。
文本：`FontSize()`、`FontWeight()`、`Bold()`、`Italic()`、`TextColor()`、`LineHeight()`、`Ellipsis()`。
状态与输入：`Label()`、`Description()`、`Disabled()`、`Focusable()`、`AutoFocus()`、`Key()`、`ItemKey()`、`Clicked()`、`Changed()`、`Clicked()`/`DoubleClicked()`、`Submitted()`、`Checked()`、`Selected()`、`Hovered()`、`Focused()`、`OnChange()`、`OnClick()`、`OnSubmit()`、`OnShortcut()`。
容器：`Children(func())`、`Clip()`。
辅助功能：`Access*`、`Role`、`Name` 等（见 `references/04-ui-foundations.md` 的辅助功能一节）。

## 常用组合

```go
// 每帧重建：状态在结构体，事件查询返回是否发生
func (a *app) view(c *ui.Context) {
    ui.Column(c).Fill().Gap(12).Children(func() {
        ui.Text(c, "标题").FontSize(20).Bold()
        if ui.PrimaryButton(c, "保存").Clicked() { a.save() }
        if ui.Checkbox(c, &a.auto, "自动检查").Changed() { a.store() }
    })
}
```

```go
// 页面路由：router 只创建一次（应用状态的一部分），view 里用 switch 匹配路径
type notesApp struct{ router *ui.Router }

func (a *notesApp) view(c *ui.Context) {
	ui.Column(c).Children(func() {
		switch {
		case a.router.Match("/notes"):
			a.notes(c)
		case a.router.Match("/notes/{id}"):
			a.note(c, a.router.Param("id"))
		default:
			ui.Text(c, "Not found")
		}
	})
}

// main 中创建：a := &notesApp{router: ui.NewRouter("/notes")}
// 前往页面：a.router.Push("/notes/42")；后退/前进：a.router.Back()/Forward()
```

> 版本差异：`mygo.run` 文档可能比安装版本新；`mobile`、`plugins/render` 需要 mygo ≥ v0.10.0。见 `sources.md`。
