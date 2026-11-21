# 原生 UI 基础

> 定位：MyGo 原生 UI 用 Go 的 `ui` 包直接绘制窗口内容，没有 HTML、JavaScript 或前端构建，窗口不启动 webview，因此启动快、内存少；macOS 上用 Metal、Windows 上用 Direct3D 11、Linux 上用 OpenGL 绘制，浏览器中为 WebGL 2。本文件供编写、审查和排查原生 UI 页面时查阅，按主题合并 `ui/` 下 24 页指南，只保留可操作事实与取舍。

示例约定沿用官方文档：`c` 是视图的 `*ui.Context`，`app` 是保存状态的自定义类型实例，`t := c.Theme()`。`app.volume`、`app.save()` 这类字段和方法由应用自行声明。

## 源页面

- [`ui/README.md`](https://mygo.run/zh/docs/ui) — 原生 UI 总览：无 webview 的 GPU 绘制、示例约定与全部控件索引。
- [`ui/views.md`](https://mygo.run/zh/docs/ui/views) — 视图生命周期、事件查询与回调、键，以及从其他 goroutine 更新状态。
- [`ui/layout.md`](https://mygo.run/zh/docs/ui/layout) — 行与列、尺寸、间距、对齐、定位与绘制顺序。
- [`ui/text.md`](https://mygo.run/zh/docs/ui/text) — 文本样式、富文本、句子中的链接和字体。
- [`ui/styling.md`](https://mygo.run/zh/docs/ui/styling) — 背景、边框、阴影、渐变、遮罩，以及主题与桌面偏好。
- [`ui/input.md`](https://mygo.run/zh/docs/ui/input) — 指针、键盘焦点、快捷键、按住的键、输入法、拖入文件与逐键输入。
- [`ui/drag-and-drop.md`](https://mygo.run/zh/docs/ui/drag-and-drop) — 窗口内拖动的值、重新排序，以及与系统交换的传输格式。
- [`ui/navigation.md`](https://mygo.run/zh/docs/ui/navigation) — 页面路由器、历史、嵌套布局、页面过渡与焦点。
- [`ui/overlays.md`](https://mygo.run/zh/docs/ui/overlays) — 对话框与浮层共用的键盘、焦点和关闭行为。
- [`ui/accessibility.md`](https://mygo.run/zh/docs/ui/accessibility) — 名称、角色、状态、实时区域、播报与各平台映射。
- [`ui/right-to-left.md`](https://mygo.run/zh/docs/ui/right-to-left) — 从右到左布局、逻辑边缘、镜像规则与试用开关。
- [`ui/drawing.md`](https://mygo.run/zh/docs/ui/drawing) — `Draw`/`DrawOver` 绘制、路径、画布、自绘文本与逐帧动画。
- [`ui/shaders.md`](https://mygo.run/zh/docs/ui/shaders) — 用 Go 语法编写着色器，在 GPU 上计算像素。
- [`ui/transitions.md`](https://mygo.run/zh/docs/ui/transitions) — 布局变化、出现与离开、共享元素与飞行。
- [`ui/motion.md`](https://mygo.run/zh/docs/ui/motion) — 弹簧、曲线、延迟、拖动与惯性、关键帧与时间线。
- [`ui/transforms.md`](https://mygo.run/zh/docs/ui/transforms) — 平移、缩放、旋转、3D、透视，以及输入与浮层的跟随。
- [`ui/custom-widgets.md`](https://mygo.run/zh/docs/ui/custom-widgets) — 各控件的无外观基础，以及自建控件的状态与清理。
- [`ui/windows.md`](https://mygo.run/zh/docs/ui/windows) — 窗口选项、自绘标题栏、内容缩放、材质与活动状态。
- [`ui/testing.md`](https://mygo.run/zh/docs/ui/testing) — `ui.NewTester` 与假时钟，断言包括辅助技术看到的内容。
- [`ui/inspector.md`](https://mygo.run/zh/docs/ui/inspector) — 元素树、盒模型、可编辑样式、帧耗时与问题列表。
- [`ui/rendering.md`](https://mygo.run/zh/docs/ui/rendering) — GPU/CPU 渲染路径、帧节奏与帧耗时统计。
- [`ui/browsers.md`](https://mygo.run/zh/docs/ui/browsers) — 编译为 WebAssembly 在浏览器运行，以及功能差异。
- [`ui/migration.md`](https://mygo.run/zh/docs/ui/migration) — 元素值、迁移命令与 `mygo vet` 检查。
- [`ui/performance.md`](https://mygo.run/zh/docs/ui/performance) — 实测帧开销与常驻内存，以及长集合的取舍。

## 逐帧重建与元素身份（README · views · migration）

**用途**：视图把应用状态映射成界面：收到输入、模型变化或动画运行时，MyGo 在主线程调用视图构建帧；一帧可以构建多次，因此绘制时已呈现操作后的结果。窗口有固定的 `*ui.Context`，子构建器共享它，「元素 → 链式调用 → 子构建回调」就是全部 API 形态。

**关键 API**：`ui.View(fn)` 作为 `mygo.WindowOptions.Content`（同一 Content 可给多个窗口，各自元素状态）；`ui.Element` 只属于一次构建，`Valid()` 判空、无元素用 `ui.Element{}`；持久控件身份用 `ui.Handle`，跨构建回调用 `c.Services()`；键：`c.Key(v)` 给下一个控件，容器的 `e.Key(v)` 须在读取元素状态前调用；`Window.Update(fn)` 在主线程改状态并重画，`Window.Invalidate` 只重画，视图内用 `c.Invalidate()`/`c.After(d)`；`ui.Local(e, name, init)` 保存元素局部状态，`OnUnmount(fn)` 在元素不再构建后运行一次；`mygo.RunOnMain` 在主线程比较而不绘制。

**最小示例**
```go
func (app *todoList) view(c *ui.Context) {
	ui.Column(c).Fill().Padding(16).Gap(8).Children(func() {
		for i := range app.todos {
			ui.Checkbox(c.Key(app.todos[i].ID), &app.todos[i].Done, app.todos[i].Title)
		}
		if ui.TextInput(c, &app.draft).Placeholder("New to-do").Submitted() {
			app.todos = append(app.todos, Todo{Title: app.draft})
			app.draft = ""
		}
	})
}
```

**限制/常见错误** 不要把构建值（元素、`*ui.Context`、`Local` 指针）存进应用状态或带进后台任务；只在视图里创建一次的状态（路由器、`ListState`）必须放到应用字段。已处理的输入在重新构建前被消费，同一输入只触发一次操作。过期元素在开发构建和 `Tester` 中报错或 panic，生产构建返回空结果或忽略修改。从其他 goroutine 改状态用 `Window.Update`。`mygo vet` 检查存放在结构体、包变量或 goroutine 中的构建值；源码迁移用 `go tool mygo migrate-ui`（先预览，再 `-write`），之后仍需人工检查派生值与绑定。

## 布局与尺寸（layout）

**用途**：元素用 flexbox（如同 CSS）或网格排布子元素，单位是设备无关像素（DIP）。

**关键 API**：容器 `ui.Row`、`ui.Column`、`ui.Box`（也是列）、`ui.Spacer`、`ui.Divider`、`ui.Dividers(width, color)`、`ui.Grid`；尺寸 `Width`、`Height`、`Size`、`WidthPercent`、`HeightPercent`、`Min*`/`Max*`（含 Percent 形式）、`Fill`/`FillWidth`/`FillHeight`、`Grow`/`Shrink`/`Basis`/`AspectRatio`；间距 `Gap`/`GapX`/`GapY`、`Padding`/`Margin`（1、2 或 4 个值，同 CSS 顺序）、`PaddingX`/`Y`、`MarginX`/`Y`、逻辑边缘 `PaddingStart`/`End`、`MarginStart`/`End`、`ui.Auto` 外边距；对齐 `Justify`、`AlignItems`、`AlignSelf`、`Center`、`Wrap`、`AlignContent`、`Baseline`；定位 `Absolute` + `Top`/`Right`/`Bottom`/`Left`（含 Percent、`Start`/`End`）、`Attach(at, self)`、`Sticky`；层叠 `ZIndex`；裁剪 `Clip`/`ClipX`/`ClipY`、`Invisible`、`Hidden`、`Debug`。

**最小示例**
```go
ui.Row(c).Gap(8).Padding(12).Children(func() {
	ui.Icon(c, folder)
	ui.Text(c, "Documents").Grow(1)
	ui.Text(c, "12 items").TextColor(t.TextMuted)
})
ui.Box(c).Size(40, 40).Children(func() {
	ui.Icon(c, bell).FontSize(24)
	ui.Badge(c, "3").Absolute().Top(-4).Right(-6)
})
```

**限制/常见错误** 绝对定位距离从父元素边框内侧量起（边框占元素尺寸内的空间）；`Top` 等用在文档流元素上只做相对偏移，不移动兄弟；`ZIndex` 只在兄弟之间生效，每个元素像 CSS 堆叠上下文一样整体移动，提高卡片要让所在列提高；`Invisible` 保留占位但不收指针和焦点，`Hidden` 收回空间并保留内部状态（滚动偏移、编辑中的文本、动画），焦点会离开；同时设置 min/max 取较大的最小值和较小的最大值，例如 `MaxWidthPercent(78).MaxWidth(544)`。

## 文本与富文本（text）

**用途**：文本在可用宽度内换行，文本样式可从任何祖先继承；富文本在一个段落内混排样式 span，句子中还能嵌入可点击、可聚焦的元素。

**关键 API**：`ui.Text`/`ui.Textf`；`FontSize`、`FontWeight`、`Bold`、`Italic`、`Font`、`LineHeight`、`FixedLineHeight`、`TextColor`、`TextAlign`、`Underline`/`WavyUnderline`/`DottedUnderline`、`Strikethrough`、`DecorationColor`/`DecorationThickness`/`UnderlineOffset`、`TextBackground`、`LetterSpacing`、`FontFeatures("tnum")`；`TextStroke(w, c)`、`TextShadow(x, y, blur, c)`、`TabSize(n)`；换行 `SingleLine`、`MaxLines`、`Ellipsis`、`NoWrap`、`Balance`、`Pretty`；选择 `Selectable`（容器上共享选区）、`CopyData`、`SelectedTexts`、`SelectionColor`、`CaretColor`；`ui.RichText` + `ui.Span{Text: …, Weight: …, Color: …, Background: …, Radius: …, PaddingX: …, BaselineShift: …, ObjectWidth: …}`、`DrawObjects`；字体 `ui.RegisterFont(data, "Inter")`、`Font("Inter, Noto Sans JP")`、`ui.FontFamilies()`；自绘排版 `ui.Shape`/`ui.ShapeRichTextWith`。

**最小示例**
```go
ui.Text(c, "Settings").FontSize(20).Bold()
ui.RichText(c,
	ui.Span{Text: "Saved "},
	ui.Span{Text: "report.pdf", Weight: 600},
	ui.Span{Text: " to "},
	ui.Span{Text: "Documents", Color: t.Accent, Underline: true},
)
```

**限制/常见错误** 文本元素里只能放 `Text`、`Link`、`RichText`，它们的大小、内边距、边框、圆角不起作用，`Background` 只高亮文字；`TextShadow` 只支持一个阴影（不是列表），且下划线、删除线不投影；`Balance` 在文本排版时计算一次，与 `Pretty` 一样适合标题、说明等短文本；span 的 `PaddingX` 占行内空间并影响换行，`PaddingY` 只影响绘制；`Selectable` 在虚拟化列表中只含已构建段落，选整篇文稿应用普通滚动容器；`ui.FontFamilies()` 首次调用可能耗几十毫秒，浏览器中只列出注册过的家族。字形回退、bidi 与 emoji 由系统文本引擎（DirectWrite、Core Text、Pango）处理。

## 样式、主题与桌面偏好（styling）

**用途**：没有样式表和状态选择器：元素外观跟随视图代码和主题；主题默认随系统浅色/深色并跟随桌面偏好，可整体或局部覆盖。

**关键 API**：盒子 `Background`、`Border(w, c)`/`BorderWidth`/`BorderColor`/`BorderStyle`、`Radius`、`Shadow(x, y, blur, spread, c)`/`InsetShadow(...)`/`NoShadow`、`Opacity`、`Blur`、`ClipInset`/`ClipInsetPercent`、`Cache`、`Material`；渐变 `Gradient`、`LinearGradient`/`ConicGradient`/`RadialGradient`（`Oklab`、`Stops`）、`Stripes`、`BorderLinearGradient`/`BorderConicGradient`；遮罩 `MaskLinearGradient`/`MaskRadialGradient`/`MaskConicGradient`、`Masks` + `ui.Mask{Image, Box, Composite}`；颜色 `ui.RGB`/`RGBA`/`Hex`/`Oklch`、`Mix`/`MixOklab`；指针 `Cursor` 系列；主题 `t := *c.Theme().WithAccent(ui.Hex("#7c3aed"))`、`c.SetTheme(&t)`、`ui.Themed(c, other)`、`ui.LightTheme()`/`ui.DarkTheme()`；桌面偏好 `c.Preferences()`。

**最小示例**
```go
t := *c.Theme().WithAccent(ui.Hex("#7c3aed"))
t.Radius, t.Spacing = 8, 3 // 默认 Spacing 为 4，3 更紧凑
c.SetTheme(&t)

ui.Column(c).Padding(18).Gap(12).Radius(10).
	Background(t.Background).Border(1, t.Border).
	Shadow(0, 1, 3, 0, ui.RGBA(0, 0, 0, 0.06))
```

**限制/常见错误** `SetTheme` 影响整个窗口（含根元素背景与文本），`ui.Themed` 只改变局部子树，子树内打开的浮层/菜单/对话框沿用该主题，toast 属于窗口例外；自定义主题不会自动跟随桌面设置，需 `WithPreferences(c.Preferences())`，`WithAccent` 只接收强调色并派生悬停、按下、选区、焦点色；控件留白由 `Spacing` 决定，可用 `t.Space(n)`/`t.Rem(n)` 让自绘元素同步；`Oklch` 颜色同时保留真实值与 sRGB 回退值，不要直接给它的 `R/G/B` 赋值；`Masks` 多图层在 CPU 上生成遮罩（每个显示尺寸一次），`Cache` 只在内容、尺寸、缩放比例或主题变化时重绘；桌面偏好包括 `Accent`、`HighContrast`、`TextScale`、`ReduceMotion`、`Pointer`/`Hover()`。

## 输入、焦点与快捷键（input）

**用途**：元素在你构建它时报告自上一帧以来的用户操作；控件在构造函数返回前就通过绑定指针写入输入，所以下一条语句能读到新值，内联处理可以照常用普通控制流。

**关键 API**：指针 `Hovered`、`Pressed`、`Clicked`、`DoubleClicked`、`RightClicked`、`Dragged`、`PointerPosition`、`PressedInside`、`PressedButton`、`PressedPointer`、`PassThrough`、`HitTest`、`IgnoreInput`、`PressCanceled`；修饰键 `c.Modifiers()`、`ClickModifiers()`；焦点 `Focusable`、`AutoFocus`、`Focus`、`Focused`、`FocusVisible`、`FocusWithin`、`FocusLeft`、`TabStop(false)`、`FocusGroup(ui.Horizontal)`、`c.FocusTakesText()`；持久焦点 `ui.Handle` + `Bind`/`Focus`/`FocusWith`/`SelectAll`/`Bounds`/`Resolve`、`FocusBind(&field, v)` + `ui.FocusedValue`；快捷键 `c.Shortcut(mods, key)`、`e.Shortcut`、`ShortcutRepeated()`、`OnShortcut`、`c.OnWindowShortcut`、`KeyShortcuts`；键状态 `c.KeyDown`/`KeysDown`；逐键 `HandleInput(func(ui.InputEvent) bool)`、`ObserveInput`、`TextCaret`、`InputMethod(false)`；滚动/捏合 `InputScroll`（`DX`、`DY`、`Precise`）、`InputZoom`；触摸 `Touches()` 与 `TouchBegan/Moved/Stationary/Ended/Canceled`；拖入文件 `DroppedFiles`/`FileDragOver`/`OnFileDrop`；剪贴板与 URL `c.ReadClipboard`/`WriteClipboard`/`WriteClipboardThen`/`ReadClipboardData`/`ClipboardFormats`/`OpenURL`/`OpenURLThen`。

**最小示例**
```go
card := ui.Column(c).Padding(12).Radius(8).Focusable()
if card.Hovered() {
	card.Background(t.SurfaceHover)
}
if c.Shortcut(ui.Cmd, ui.KeyK) && !c.FocusTakesText() {
	app.palette = true
}
ui.TextInput(c, &app.query).Bind(&app.search) // app.search 为 ui.Handle
```

**限制/常见错误** `Pressed`/`Clicked`/`Dragged` 等只跟随主按钮，右键或中键拖动只产生右键点击；系统取消的按下不触发点击、放下、拖动结束或上下文菜单，用 `PressCanceled` 撤销并回滚；按下的元素在指针按住期间保持 `Hovered`；只有在视图构建时读取 `Hovered` 才会因悬停重建视图，仅影响绘制时应在 `Draw` 里读；`Disabled` 移出焦点与 Tab 顺序，`AccessDisabled` 只告诉辅助技术；对话框外构建的窗口快捷键在对话框显示期间不响应；`HandleInput` 接下的键不再传给快捷键、Tab、菜单和滚动容器，不用的键返回 false，并用 `Key != ui.KeyUnknown` 过滤只有 `Code` 的键；`ui.Cmd` 在 macOS/iOS 是 Command，其他平台是 Ctrl。

## 拖放（drag-and-drop）

**用途**：在窗口内拖动 Go 值，通过拖动重排列表/表格/网格的行，或以系统传输格式与其他应用交换数据。

**关键 API**：窗口内 `e.Drag(value)`、`ui.Drop[T](e)`、`ui.DragOver[T](e)`、`Dragging()`、`DragMissed()`；重排 `ListState.Reorder`、`GridState.Reorder`、`ListState.DragRows`；系统传输 `e.DragData(data)`、`e.DragDataFrom(func() transfer.Data)`、`ui.DropData(e, opts)`、`ui.DataDragOver(e, opts)`、`transfer.TextData`/`URLData`/`FileData`/`New`+`NewItem`+`Bytes`+`Lazy`，`transfer.DragOptions{Operations: transfer.Copy|transfer.Move, Done: func(transfer.Result)}`；`Window.CancelDataDrag`、`Window.StartDataDrag`。

**最小示例**
```go
for _, task := range app.todo {
	ui.Row(c).Key(task.ID).Drag(task).Children(func() { ui.Text(c, task.Name) })
}
done := ui.Column(c).Grow(1).Padding(12).Border(1, t.Border)
if _, over := ui.DragOver[*Task](done); over {
	done.Border(2, t.Accent)
}
if task, ok := ui.Drop[*Task](done); ok {
	task.Done = true
}
```

**限制/常见错误** 能接受该值类型的最内层元素接收放下，仅接受单击的控件（按钮）在按下后移动会转为拖动元素，接受指针移动的控件（文本输入、滑块）保留移动；窗口内拖动默认留在本窗口，与 `DragData` 结合才把 Go 值提供给同进程其他窗口；外部表示每项限制 64 MiB，`Lazy` 提供者在主线程运行、必须迅速返回且不能等待主线程；目标默认复制，双方都允许 `Move` 时才移动，MyGo 从不删除文件，移动成功后由应用移除记录；`Done` 在主线程恰好运行一次（取消、失败、源被移除、窗口销毁也会运行）。

## 页面路由与导航（navigation）

**用途**：`ui.Router` 为窗口或窗口中的区域保存页面历史；路由器是应用状态的一部分，只创建一次，视图每帧构建当前页面。

**关键 API**：`ui.NewRouter("/notes")`；`router.View(c, func(r *ui.Route))`、`r.Match("/notes/{id}")`、`r.Match("/files/{path...}")`、`r.Param`、`r.Title`、`r.View`（嵌套布局）；`Push`/`Replace`/`Back`/`Forward`/`Go`，相对路径 `Push("edit")`、`Push("?tab=info")`；`Path`/`Query`/`Location`；`ui.Link(c, "Open note", "/notes/42")` 在所属路由器中跳转；`router.Transition = ui.TransitionNone|ui.TransitionMatch`、`router.Animation`、`SystemStack`；`mygo.App.OnOpenURL` 深度链接。

**最小示例**
```go
app.router.View(c, func(r *ui.Route) {
	switch {
	case r.Match("/notes"):
		r.Title("Notes")
		app.notes(c)
	case r.Match("/notes/{id}"):
		r.Title("Note")
		app.note(c, r.Param("id"))
	}
})
```

**限制/常见错误** 路由器必须在启动时创建并保存为字段，不能在视图里创建；页面在历史中保留元素状态（前后各十个页面），同一路径带不同查询是同一页面的不同历史条目；更深的路径从右侧滑入、返回向左滑回，其他页面淡入，减少动效时全部淡入且 `TransitionNone` 立即显示；离开的页面不接收指针与键盘；切换页面把键盘焦点移到该页上次获得焦点的元素，没有则移到页面本身，屏幕阅读器朗读 `Title`；浏览器中第一个路由器跟随标签页历史（`#/notes/42`）。

## 浮层（overlays）

**用途**：浮层显示在窗口内容之上：对话框、警告对话框、弹出面板、下拉选择框与组合框的列表、窗口内绘制的菜单、工具提示、toast。它们共享焦点、Escape 与背景下压的规则。

**关键 API**：`ui.Overlay(c, func())`（用 `Absolute` 按窗口 DIP 定位）、`ui.Popover(c, anchor, &open, fn)`、`ui.Modal(c, &open, fn)`；`AttachTo(target, at, self)`、`AttachEdgeMargin`、`AttachedAnchors`、`AttachedOrigin`；`PressedOutside`/`PressedInside`、`OverlayShortcut`、`Modal()`；无外观 `PopoverBase`/`DialogBase`；`Tooltip`、`c.Toast`、`AutoFocus`。

**最小示例**
```go
ui.Overlay(c, func() {
	back := ui.Row(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).
		Justify(ui.End).Background(ui.RGBA(0, 0, 0, 0.3)).Modal()
	back.Children(func() {
		panel := ui.Column(c).Width(320).FillHeight().Background(t.Background)
		if panel.PressedOutside() || back.OverlayShortcut(0, ui.KeyEscape) {
			app.drawer = false
		}
	})
})
```

**限制/常见错误** 对话框把焦点限制在内部（默认落在第一个可聚焦元素或 `AutoFocus`），Tab 只在内部循环，外部快捷键不响应；Escape 关闭最上层浮层，关闭时焦点还给打开前的元素；在弹出面板外按下会关闭并继续传给下面内容，下拉选择框的弹出列表则拦截这次按下；浮层沿用创建位置的主题，toast 用窗口主题；`AttachTo` 在空间不足时翻到另一侧或沿另一方向放置，`AttachedAnchors` 读上一帧结果，`Flip` 让进入/离开动效随之镜像；退出中的副本不接收指针、焦点和辅助技术。

## 辅助功能（accessibility）

**用途**：屏幕阅读器把原生 UI 当作其他应用读取：控件及其角色、名称、值、状态、文本和焦点变化；辅助技术也能按下、勾选、滑动、编辑和移动焦点。

**关键 API**：名称 `Label`、`LabelledBy`、`Description`；角色 `Role(ui.RoleButton)`（常用 `ui.RoleHeading`、`ui.RoleNavigation`、`ui.RoleButton`、`ui.RoleTabList`/`ui.RoleTab`、`ui.RoleMenu*`、`ui.RoleListBox`/`ui.RoleOption`、`ui.RoleGrid`/`ui.RoleTreeGrid`/`ui.RoleTable`/`ui.RoleRow`/`ui.RoleCell`/`ui.RoleColumnHeader`）、`RoleNone`，标题用 `RoleHeading` + `Level(n)`；状态 `Checked`/`Mixed`、`AccessSelected`、`Expanded`、`Value`、`Range`/`Step`、`Level`、`PositionInSet`、`ActiveDescendant`、`Busy`、`Current`、`AutoComplete`、`Controls`、`HasPopup`、`Sort`、`Orientation`、`DescribedBy`/`Invalid`/`ErrorMessage`、`ReadOnly`、`Multiselectable`、`RowCount`/`ColumnCount`/`RowIndex`/`ColumnIndex`/`ColumnSpan`/`RowSpan`；可用性 `AccessDisabled`、`IgnoreInput`、`AccessHidden`；播报 `c.Announce(text)`，实时区域 `RoleLog`/`RoleAlert`/`RoleStatus`/`RoleTimer`。

**最小示例**
```go
ui.Slider(c, &app.volume, 0, 100).Label("Volume")
ui.Text(c, "Appearance").Bold().Role(ui.RoleHeading).Level(2)
del := ui.Button(c, "Delete").AccessDisabled(app.locked)
if app.locked {
	del.Description("Unlock the project to delete it")
}
```

**限制/常见错误** 只在悬停时构建的按钮对辅助技术不存在，改用始终构建 + `IgnoreInput(!hovered)` + `Opacity(0)`；`Disabled` 会让元素离开焦点和 Tab 顺序并变暗，`AccessDisabled` 只报告禁用而保留焦点、Tab 顺序、外观与点击，视图必须自己忽略这些操作；`AccessHidden` 仍绘制仍收指针，应避免键盘焦点进入；`IgnoreInput` 不出现在 Tab 顺序、不收点击，但辅助技术仍能找到并按下；`Hidden`/`Invisible`/`AccessHidden` 才是移出树；虚拟化列表用 `PositionInSet(pos, size)`，自建网格用计数与索引；`c.Announce` 只播报、不移动焦点，`RoleStatus` 等实时区域按 ARIA 规则朗读；映射遵循 WAI-ARIA Core Accessibility API Mappings。

## 从右到左（right-to-left）

**用途**：应用语言是阿拉伯语、希伯来语、波斯语或乌尔都语时，原生 UI 整体镜像布局；只有布局镜像，坐标系不变（`Rect`、`Draw` 与指针的 X 仍向右增大）。

**关键 API**：`ui.Direction`（`ui.LeftToRight`/`ui.RightToLeft`）；元素 `Direction(…)`、`c.Direction()`；逻辑边缘 `PaddingStart`/`End`、`MarginStart`/`End`、`BorderStartWidth`/`EndWidth`、定位 `Start`/`End`、锚点 `AnchorTopStart`/`AnchorTopEnd` 等；环境变量 `MYGO_DIRECTION=rtl` 试用；`Tester.SetDirection(ui.RightToLeft)`；检查器工具栏的 RTL 切换。

**最小示例**
```go
ui.Row(c).Direction(ui.LeftToRight).Gap(8).Children(func() { // 播放器控件保持 LTR
	ui.Button(c, "⏮")
	ui.Button(c, "⏯")
})
forward := "M5 2 11 8 5 14"
if c.Direction() == ui.RightToLeft {
	forward = "M11 2 5 8 11 14" // 自绘箭头需要自己镜像
}
```

**限制/常见错误** 物理方向属性不镜像：`Left`/`Right`、`Padding(top, right, bottom, left)`、`BorderWidth`、`AnchorTopLeft`、`Radius` 各角、渐变角度、变换；文本、图像、图标与 `Draw` 内容不镜像，需要指向另一边的图形要在读取 `c.Direction()` 后绘制；代码编辑器、差异查看器、Markdown 代码块、终端与检查器保持从左到右；`TextAlign` 的 `Start` 指元素右侧（不受文本自身阅读方向影响），段落阅读方向仍由 Unicode 双向算法决定；浮层、选择器、菜单在打开元素的起始侧打开。

## 绘制与着色器（drawing · shaders）

**用途**：`Draw` 在元素背景之后绘制，`DrawOver` 在子元素之后绘制，都以窗口 DIP 为单位的 `*ui.Painter` 操作；着色器在 GPU 上按像素计算矩形颜色，可读取 uniform、图像和窗口背后的内容。

**关键 API**：`*ui.Painter`：`Fill`、`FillPath`、`Stroke`、`StrokePath`、`StrokeDashed`、`Line`、`Shadow`、`Image`/`ImageRegion`（`ui.ImageOptions{FlipX, FlipY, Tint}`）、`Clip`/`ClipPath`、`Text`/`RichText`/`Glyphs`/`MeasureText`/`TextLayout`、`Transform(t, ox, oy, fn)`、`FocusRing`；`ui.Path`（`MoveTo`/`LineTo`/`Circle`/`Trim`/`Cap`/`Join`/`Dash`/`FillRule(ui.EvenOdd)`/`Contains`）、`ui.ParsePath`/`MustParsePath`；`ui.Canvas`（`Draw`/`Bitmap`）；`ui.Shape`；`Draw` 中动画 `p.AnimationFrame()`、`p.AnimationFrameRate(fps)`、`p.After(d)`；着色器 `shader.MustCompile`/`Compile`、`Fragment(dst, pos)`、`shader.Options{Uniforms, Values, Images, Radius, BackdropBlur, MaxSampleOffset}`、`Draw`/`FillPath`/`Text`/`Fill`/`DrawInstances`、`Effect(...)`（`ColorEffect`/`DistortionEffect`/`LayerEffect`）、`p.Accelerated()`。

**最小示例**
```go
ui.Box(c).Height(120).Draw(func(p *ui.Painter, r ui.Rect) {
	var wave ui.Path
	wave.MoveTo(r.X, r.Y+r.H)
	wave.LineTo(r.X+r.W/3, r.Y+r.H/3)
	wave.LineTo(r.X+2*r.W/3, r.Y+2*r.H/3)
	wave.LineTo(r.X+r.W, r.Y)
	p.StrokePath(&wave, 2, t.Accent)
})
```

**限制/常见错误** 绘制函数只负责绘制，一帧内可能被多次调用，不能有副作用，也不能在这里改状态；`Draw` 读的是视图上次运行时的状态，外部改了数据要 `Window.Update`；着色器是 `package main` 的 Go 文件（派生自 Ebitengine 的 Kage），`Fragment` 必须返回预乘颜色，`Compile` 约 1 ms 且应作为包变量只编译一次；`DrawInstances` 一次最多 65536 个四边形、每实例最多 16 个浮点，数据长度不符或读取背景的着色器会 panic；CPU 渲染（含 `Tester`）上着色器按解释执行、每像素几微秒，可用 `p.Accelerated()` 跳过重效果；uniform 设为非数字时 `Draw` panic。

## 过渡、动效与变换（transitions · motion · transforms）

**用途**：过渡让元素从上一帧状态平滑到当前布局状态（位置、尺寸、背景色、边框色、阴影）；动效把数值插值到目标并持续请求帧；变换在视觉上平移、缩放、旋转元素而布局保留原空间。

**关键 API**：过渡 `e.Transition(ui.ElementTransition{...})`：`Duration`/`Ease`/`Animation`、`Position`/`Size`/`Width`/`Height`/`Colors`/`Shadows`、`Enter`/`Exit`（`ui.Motion{X, Y, XPercent, YPercent, Opacity, Scale, Rotate, Blur, Collapse, Flip}`）、`ExitDuration`/`ExitEase`/`ExitDelay`、`ChildExit`、`AfterExits`、`Match`、`Fly`、`ScaleContent`、`ContentTransition{Morph, Numeric}`、`Transitioning()`；动效 `e.AnimateTo(key, target, anim)`、`Animate`/`AnimateWith`、`AnimateColor`、`ui.Smooth`/`Snappy`/`Bouncy`、`ui.Animation{Duration, Ease, Spring: ui.SpringOptions{Duration, Bounce, Stiffness, Damping, Mass}, Oklab, NonMotion, Delay}`、`AnimatePose(ui.Pose{...}, anim)`、`Draggable(&point, ui.DraggableOptions{Axis, Bounds, Snap, Animation})`、`Jump`/`Dragged`/`DragEnded`/`DragVelocity`/`ui.Project`/`ui.Rubberband`、`Settled`/`Arrived`、`Keyframes`、`Timeline`（`Tween`、`Label`、`AfterPrevious`/`WithPrevious`/`AtLabel`/`AtTime`、`Paused`/`Reverse`/`Seek`/`Progress`）、`InView`/`InViewMargin`/`InViewAmount`、`c.DisableAnimations()`/`c.Instant()`；变换 `Translate`/`Scale`/`Rotate`/`Skew`/`RotateX`/`RotateY`/`TranslateZ`/`Perspective`/`BackfaceHidden`/`TransformOrigin`/`Transform(ui.Transform)`/`TranslatePercent`。

**最小示例**
```go
card := ui.Box(c).Size(120, 80).Radius(12).Background(t.Accent)
if app.open {
	card.Translate(card.AnimateTo("x", 200, ui.Snappy), 0)
}
row := ui.Row(c).Key(app.activeID).Padding(8)
row.Transition(ui.ElementTransition{Enter: &ui.Motion{Y: 8}, Exit: &ui.Motion{Y: -8}})
```

**限制/常见错误** 过渡和动画要求元素跨帧保持同一身份（相同 `Key` 或兄弟位置）；受影响而移动的兄弟元素和容器也要各自设置过渡；改变窗口大小不做动画；`ReduceMotion` 下值和姿态立即到达、序列停在末值，颜色和淡入淡出仍平滑，`Loop` 继续，不移动的动画可设 `NonMotion`；退出副本不占空间、不收输入与辅助技术、保留原 `ZIndex`；`AnimatePose`、内容过渡与 `Draw` 中计算的动画只重绘不运行视图（开销最低）；`Draggable` 的偏移保存在 `ui.Point`，`DragOffset()` 会读取并在移动的每一帧重建视图；变换值必须有限，0 缩放不可见也不参与命中测试，`Transform` 会替换已有变换而便捷方法会追加，指针/输入/辅助技术跟随变换后的外框，浮层附着到变换后的边界。

## 自定义控件（custom-widgets）

**用途**：每个内置控件都建立在一个无外观基础（`*Base`）上：基础处理指针、键盘、焦点和辅助功能信息，外观由你给返回元素设置样式。

**关键 API**：`ButtonBase`、`CheckboxBase`/`SwitchBase`/`ToggleBase`、`CheckboxStateBase`（`*ui.CheckState`）、`RadioBase`、`SliderBase`、`TabsBase`、`SegmentedBase`、`CollapsibleBase`、`AccordionBase`、`TreeBase`、`SelectBase`、`ComboboxBase`、`PopoverBase`/`DialogBase`、`MenuBase`/`ContextMenuBase`、`TooltipBase`、`ToastViewportBase`/`ToastBase`、`TextInputBase`/`TextAreaBase`；有状态基础在构造前用 `c.Key(...)`；局部状态 `ui.Local(e, name, init)`、`OnUnmount`、`SetChanged`；焦点环 `FocusRing(false)`/`Painter.FocusRing`；禁用表现 `DimDisabled(false)`、`IsDisabled()`。

**最小示例**
```go
view := ui.SegmentedBase(c.Key("period"), &app.view, 3)
view.Track.Padding(3).Radius(999).Background(t.Surface).Children(func() {
	for i, name := range []string{"Day", "Week", "Month"} {
		seg := view.Segment(i).Padding(5, 14).Radius(999)
		if i == app.view {
			seg.Background(t.Background).Shadow(0, 1, 2, 0, ui.RGBA(0, 0, 0, 0.15))
		}
		seg.Children(func() { ui.Text(c, name) })
	}
})
```

**限制/常见错误** 有状态基础必须在构造前用 `Context.Key`，元素 `Key` 来不及；`ui.Local` 的值若实现 `io.Closer` 会在元素不再构建时关闭（列表行滚出视口、页面移出历史、窗口关闭都会发生），`Close` 在主线程运行一次、应取消工作后立即返回，不要等待；应用共享的数据应放应用状态而非 `Local`；元素带退出过渡离开时，副本消失后才关闭值；禁用元素默认半透明，`DimDisabled(false)` 后自行表现；自建控件用 `Role`/状态方法提供辅助功能信息，用 `Shortcut`/`HandleInput` 收键，用 `Draw` 绘制，浮层用 `AttachTo` + `PressedOutside`/`OverlayShortcut` 关闭。

## 原生窗口（windows）

**用途**：带 `Content` 的窗口显示原生 UI；标题栏、内容缩放、窗口材质和活动/隐藏状态需要视图读取对应 API 自行处理。

**关键 API**：`mygo.NewWindow(mygo.WindowOptions{Title, Width, Height, StateKey, Frameless, Parent, Content: ui.View(app.view)})`；`Page()` 为 nil、`URL` 与 `Page` 被忽略，`CapturePage` 返回 PNG；标题栏 `TitleBarStyle: mygo.TitleBarHidden`、`c.TitleBar()`（`Height`/`Right`/`Left`/`Bottom`）、`DragWindow()`、`WindowDoubleClicked()`、`HideWindowButtons: true`、`Minimize`/`Maximize`/`Unmaximize`/`Close`；缩放 `c.SetZoom(0.25…5)`/`c.Zoom()`，视图尺寸用 `c.Size()`；材质 `Vibrancy`、`VibrancyGlass`/`VibrancyGlassClear`、`VibrancyTint`、`c.Vibrancy()`、`SetActiveAppearance`、`FitContent`；状态 `c.WindowActive()`、`c.WindowHidden()`、`Destroy`/`Close`。

**最小示例**
```go
bar := c.TitleBar()
ui.Row(c).Height(max(bar.Height, 32)).
	Padding(0, bar.Right+12, 0, bar.Left+12).
	DragWindow().Children(func() {
		ui.Text(c, "Inbox").Bold()
	})

if c.Vibrancy() {
	c.Root().Background(ui.Transparent)
}
```

**限制/常见错误** `c.TitleBar()` 全屏时为零；`SetZoom` 不是放大截图，绘制像素密度随 `z` 提高、文本与矢量保持清晰，但 `WindowOptions.Width/Height` 和 `SetContentSize` 仍是窗口 DIP（是视图尺寸的 `z` 倍）；`Vibrancy` 只在 macOS 与 Windows 11 22H2+ 且创建时启用材质时生效，其他平台透明处不显示材质，需自绘背景；`VibrancyGlass` 通常配合 `Frameless` + `Transparent`，透明玻璃只在活动外观下保持，`SetActiveAppearance` 使用 AppKit 私有调用（Mac App Store 应用不应启用）；`WindowHidden` 在 Linux X11 与 Windows 只报告最小化；关闭窗口后不保留视图和内容，应先打开接替窗口再关最后一个；`mygo_nowebview` 构建标签移除 WebView2（无 `Content` 的窗口会让 `NewWindow` panic）。

## 测试与检查器（testing · inspector）

**用途**：`ui.NewTester` 无需窗口即可运行视图，在内存中渲染帧、按文本或 `Label` 查找元素并执行操作；检查器在窗口中按元素查看树、盒模型、样式和帧耗时，与 Chrome 开发者工具类似。

**关键 API**：`ui.NewTester(app.view, 320, 240)`、`NewTesterWith(view, w, h, ui.TesterOptions{Preferences, Dark})`；查找与交互 `Click`/`ClickAt`/`ClickWith`/`RightClick`/`Find`/`Focused`/`Texts`/`HasText`/`Cursor`/`Clipboard`/`OpenedURLs`/`Announcements`；指针 `Press`/`Move`/`Release`/`PressButton`/`ReleaseButton`/`CancelPress`/`Scroll`/`Wheel`/`UsePointer(ui.PointerPen, 0.6)`；键盘 `Key`/`KeyRepeat`/`TypeKey`/`Type`/`Compose`/`Command`/`TextCaret`/`CodeDown`/`CodeUp`；菜单 `RightClick`/`Menu`/`ChooseMenuItem`/`CloseMenu`；时钟 `tt.Now()`/`tt.Advance(d)`；帧 `Frame`/`Image`/`FramePending`/`ui.Render`；环境 `SetSize`/`SetScale`/`SetZoom`/`SetDark`/`SetPreferences`/`SetDirection`/`SetFocused`/`SetWindowActive`/`SetWindowHidden`；辅助技术 `Access`/`AccessByRole`/`AccessAll`、`UseScreenReader`/`Alerts`；检查器 `F12`/`Ctrl+Shift+I`（macOS `Alt+Cmd+I`）、挑选元素 `Ctrl+Shift+C`、`PageOptions.DevTools`、Styles/Computed/Properties 标签、`:hov` 强制状态、Copy as Go、`MYGO_EDITOR`、Performance/Issues。

**最小示例**
```go
func TestCounter(t *testing.T) { // tt 是测试器，这里的 t 是 testing.T
	app := &counter{}
	tt := ui.NewTester(app.view, 320, 240)
	if err := tt.Click("Increment"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(100 * time.Millisecond)
	if app.n != 1 || !tt.HasText("1") {
		t.Errorf("count %d, texts %q", app.n, tt.Texts())
	}
}
```

**限制/常见错误** `Click` 只检查第一个匹配元素，元素不可见或中心被其他控件/空白命中时返回错误，且不会自动滚动或换位置；`Find` 返回裁剪后的边界框，宽高可能为零；同一父元素下重复 `Key` 会让 `Tester` 在第二次赋键处 panic；所有依赖时间的行为使用测试器时钟（动画、`After`、双击、撤销分组、插入符闪烁、提示延迟），不要用 sleep 等待，`Advance` 只运行视图请求的帧，外部修改内容后先 `tt.Frame`；输入和 `tt.Frame` 最多运行 20 帧，用 `FramePending` 判断视图是否仍未稳定；检查器约占二进制 400 KB，生产构建默认去掉（`-debug` 或 `MYGO_INSPECTOR=1` 保留），`-trimpath` 构建没有可打开的源码链接，显示布局时会关闭布局缓存。

## 渲染、性能与浏览器（rendering · performance · browsers）

**用途**：MyGo 只在有变化时绘制一帧：GPU 路径为 Metal/Direct3D 11/OpenGL（Linux 经 GTK 的 GtkGLArea），无 GPU 时用 CPU 渲染；同一应用可编译为 WebAssembly 在浏览器用 WebGL 2 运行，代码无需改动。

**关键 API**：后端开关 `MYGO_GPU=0`（CPU 渲染）、Linux `MYGO_GPU=1`（强制 OpenGL）、帧统计 `MYGO_FRAME_STATS=1`（阈值毫秒或 `all`）、不限帧率 `MYGO_UNCAPPED=1`；浏览器 `mygo dev -web`（`http://localhost:5174`，改动后自动重载）、`mygo build -platform web`（输出静态站点）、手写 `GOOS=js GOARCH=wasm go build` + `wasm_exec.js`；浏览器主题 `data-mygo-theme="dark"|"light"`。

**最小示例**
```sh
mygo dev -web
mygo build -platform web
```

**限制/常见错误** 大多数帧会完整运行视图并比较、布局、定位它构建的每个元素，开销随构建的元素数量增长而不是随变化量增长：长集合放进 `List`/`Table`（只构建可见行，文档在 Apple M5 上实测 3,000 行中一个标签变化约 0.2 ms），不要放进会构建每行的 `Scroll`（同场景约 11 ms）；只在视图构建时读 `Hovered` 会因悬停重建整个视图，仅绘制变化的内容在 `Draw` 里读；输入会在同一帧构建两次；窗口不可见时不绘制、视图不运行，后台工作应在 goroutine 中用 `Window.Update` 更新界面；CPU 回退（含 WARP、llvmpipe、SwiftShader）可用 `p.Accelerated()` 检测；浏览器中不提供网页窗口、应用菜单、托盘、通知、全局快捷键、文件对话框和文件之外的数据拖放，辅助技术还看不到内容，`Dialog.Message` 用浏览器 `alert`/`confirm`，浏览器可能拒绝用户未主动触发的剪贴板写入；counter 的浏览器构建为 8.2 MB（gzip 2.2 MB）。

## 速查

| 页面 / API | 一句话 |
|---|---|
| `ui/README.md` | 原生 UI 总览：无 webview、GPU 绘制、可与网页窗口共存。 |
| `ui/views.md` | 逐帧重建；`ui.View`、`*ui.Context`、`Valid`、`Handle`、`Window.Update`。 |
| `ui/layout.md` | flexbox/网格、DIP 尺寸、间距对齐、绝对定位、`ZIndex`、裁剪。 |
| `ui/text.md` | `Text`/`RichText`、span 样式、选择、字体注册与系统文本引擎。 |
| `ui/styling.md` | 盒子、渐变、遮罩、缓存；`c.Theme()` 与桌面偏好。 |
| `ui/input.md` | 指针、焦点、快捷键、按住的键、输入法、`HandleInput`、触摸、剪贴板。 |
| `ui/drag-and-drop.md` | `Drag`/`Drop[T]` 窗口内拖动；`DragData`/`DropData` 系统传输。 |
| `ui/navigation.md` | `ui.Router` 页面历史、嵌套布局、页面过渡与焦点恢复。 |
| `ui/overlays.md` | 对话框/弹出面板/菜单/提示共用的焦点与 Escape 规则，`ui.Overlay`。 |
| `ui/accessibility.md` | `Label`/`Role`/状态方法、`AccessDisabled`、`AccessHidden`、`c.Announce`。 |
| `ui/right-to-left.md` | `ui.RightToLeft`、逻辑边缘、`MYGO_DIRECTION=rtl`。 |
| `ui/drawing.md` | `Draw`/`DrawOver` + `ui.Painter`、路径、画布、自绘文本。 |
| `ui/shaders.md` | `ui/shader` 的 Go 语法着色器、uniform、图像、背景、效果。 |
| `ui/transitions.md` | `ElementTransition` 的位置/尺寸/颜色/阴影、Enter/Exit、`Match`/`Fly`。 |
| `ui/motion.md` | `AnimateTo`、弹簧、`AnimatePose`、`Draggable`、`Keyframes`、`Timeline`。 |
| `ui/transforms.md` | `Translate`/`Scale`/`Rotate` 与 3D、透视，输入与浮层跟随。 |
| `ui/custom-widgets.md` | 无外观 `*Base` 控件、`ui.Local` 生命周期、自建控件。 |
| `ui/windows.md` | 标题栏、`c.SetZoom`、材质、活动/隐藏窗口、关闭。 |
| `ui/testing.md` | `ui.NewTester` 内存渲染、假时钟、菜单与辅助技术断言。 |
| `ui/inspector.md` | 元素树、盒模型、可编辑样式、`:hov`、帧耗时与问题。 |
| `ui/rendering.md` | Metal/D3D11/OpenGL/WebGL 2、CPU 回退、帧节奏与统计。 |
| `ui/browsers.md` | WebAssembly + WebGL 2、`mygo dev -web`、浏览器功能差异。 |
| `ui/migration.md` | `go tool mygo migrate-ui`、元素值、`Handle`、`FocusBind`、`mygo vet`。 |
| `ui/performance.md` | 实测帧开销与常驻内存；长集合用 `List`/`Table`。 |
| `ui.Context` / `ui.Element` | 只在一次构建内有效；`Valid`、`Key`、`Local`、`OnUnmount`。 |
| `ui.Handle` | 跨构建的持久控件身份与焦点：`Bind`/`Focus`/`Bounds`/`Resolve`。 |
| `c.Preferences()` | 强调色、高对比度、文本缩放、减少动效、指针精度与悬停能力。 |
| `c.Shortcut` / `c.OnWindowShortcut` | 窗口级快捷键；元素的 `Shortcut` 在有焦点时优先。 |
| `ListState.Reorder` | 拖动重排列表行；`DragRows` 让行可拖出列表。 |
| `AccessDisabled` / `IgnoreInput` / `AccessHidden` | 三种不同的「可见但不可用 / 不可点 / 不可读」。 |
| `c.Announce` | 不移动焦点的屏幕阅读器播报；`RoleLog`/`RoleStatus`/`RoleAlert` 是实时区域。 |
| `ui.ElementTransition` | 进入/离开、共享元素（`Match`/`Fly`）与内容过渡的配置。 |
| `ui.NewTester` | 无窗口测试视图，假时钟与辅助功能断言。 |
| `c.SetZoom` | 内容缩放 0.25–5，按缩放后的像素密度绘制而非放大截图。 |
| `mygo.WindowOptions` | 原生窗口选项：`Content`、大小、`TitleBarStyle`、`Frameless`。 |
| `mygo vet` / `migrate-ui` | 检查构建值逃逸，并迁移到带校验的元素 API。 |
| `mygo dev -web` | 浏览器中开发原生 UI，编辑后自动重建并重载。 |
| `MYGO_FRAME_STATS` | 记录慢帧的构建、布局、绘制、呈现耗时与分配量。 |
