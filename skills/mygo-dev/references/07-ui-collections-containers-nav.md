# UI 集合、容器与导航

本参考提炼 MyGo 原生 UI 中与集合、容器、导航相关的 16 个页面，按“集合控件 → 布局容器 → 导航控件”组织。`List`、`Table`、`Outline`、`GridView` 只构建可见部分，状态放在应用持有的状态结构中；`Scroll`、`Grid`、`Split` 等容器只负责布局，会构建自己的全部子元素。

## 源页面

在线地址：`https://mygo.run/zh/docs/ui/<页面名>`；相对页面路径：`ui/<页面名>.md`。

- [列表](https://mygo.run/zh/docs/ui/list) · [表格](https://mygo.run/zh/docs/ui/table)
- [树](https://mygo.run/zh/docs/ui/tree) · [大纲](https://mygo.run/zh/docs/ui/outline)
- [网格视图](https://mygo.run/zh/docs/ui/grid-view) · [滚动视图](https://mygo.run/zh/docs/ui/scroll)
- [网格](https://mygo.run/zh/docs/ui/grid) · [分栏视图](https://mygo.run/zh/docs/ui/split)
- [可折叠](https://mygo.run/zh/docs/ui/collapsible) · [手风琴](https://mygo.run/zh/docs/ui/accordion)
- [表单](https://mygo.run/zh/docs/ui/form) · [Web 视图](https://mygo.run/zh/docs/ui/web-view)
- [标签页](https://mygo.run/zh/docs/ui/tabs) · [侧边栏](https://mygo.run/zh/docs/ui/sidebar)
- [面包屑导航](https://mygo.run/zh/docs/ui/breadcrumbs) · [后退和前进按钮](https://mygo.run/zh/docs/ui/history-buttons)

## 列表

`ui.List` 显示任意长度集合的行，只构建视口内的行和视口外的少量行；行高由内容决定并逐行测量，因此百万行列表与十行列表一样快。

**关键 API**：`ui.List(c, *ui.ListState, n, func(i int))`；`ListState` 的 `Key`（条目身份）、`Label`、`Header`（分区标题）、`KeepOffset`、`FollowEnd`/`FollowAnimation`、`HoldPlace`、`Selected`/`Selection`、`Handle`；`Changed`/`Submitted` 报告交互；`ScrollTo`/`ScrollIntoView`/`ScrollToEnd`/`Visible`/`AtEnd` 控制滚动；`NearEnd` + `EndReached` 分页；`Reorder`/`DragRows`/`DragSelect`；`Rows` 构建自定义行（`ListRow.Hovered`/`Pressed`/`PointerPosition`）。

```go
app.list.Key = func(i int) any { return app.chats[i].ID }
app.list.NearEnd = 20
app.list.EndReached = func(last int) {
	if app.more && !app.loading {
		app.loading = true
		go app.loadPage(len(app.chats)) // appends on the UI thread
	}
}
ui.List(c, &app.list, len(app.chats), func(i int) { chatRow(c, app.chats[i]) }).Grow(1)
```

**限制**：每个列表使用自己的 `ListState`，构建前设置字段，换数据源时重置为 `ui.ListState{}`；`EndReached` 只在用户滚动进入末尾附近时触发，程序化滚动、首次显示、行增删/重排都不触发，第一页应高于列表；构建后读取所选行的代码要处理未选中时的 `-1`；`Reorder`、`DragRows`、`DragSelect` 不能同时启用；分区标题用 `Header`，不要用粘性元素。

## 表格

`ui.Table` 把列表的行排到 `ui.TableColumn` 表头下的列中，同样只构建视口行，行高取最高单元格。

**关键 API**：`ui.Table(c, *ui.ListState, []ui.TableColumn, n, func(row, col int))`，选择、`Key`、`Reorder` 等与列表相同；`TableColumn` 有 `Title`、`Width`（0 则分摊剩余空间）、`Align`、`Sortable`、`RowHeader`、`MinWidth`/`MaxWidth`、`Fixed`、`Frozen`。列可拖动调整、移动、双击自适应，`VirtualColumns` 只构建视野内的列（要求每列有 `Width`），`Columns` 保存/恢复顺序与宽度；`CellSelect` + `Cells`/`CellSelected`/`CellsStyle` 做单元格块选择；`Sort`/`SortBy` 排序，排序本身由应用完成；`HeaderMenu`/`HeaderCell`/`CellMenu` 挂菜单；`RowHeight` 固定行高，`RoundRows` 给行加圆角；`Changed` 在列序、列宽或排序变化时报告一次。

```go
cols := []ui.TableColumn{{Title: "Name", Sortable: true}, {Title: "Size", Width: 90, Align: ui.End, Sortable: true}}
app.files.Sort = &app.sort
sorted := sortFiles(app.files, app.sort) // by app.sort.Column and Descending
app.files.Key = func(i int) any { return sorted[i].ID }
if ui.Table(c, &app.files, cols, len(sorted), func(row, col int) {
	ui.Text(c, sorted[row].Cell(col)).SingleLine()
}).Grow(1).Changed() {
	app.saveSettings() // columns and sort
}
```

**限制**：表格按 `cell` 收到行的顺序显示，排序由应用完成，必须用 `Key` 让选择跟随条目；冻结列须从第一列起连续、每列都要有宽度，冻结列占满表格时其余列不可见且不会自动取消冻结；`RowHeader` 列不参与排序；就地重命名用 `EditableText`（macOS 用 Return，其他平台用 F2）；单元格内元素的 `DoubleClicked` 会拦截行点击，应改用 `RowDoubleClicked`。

## 树

`ui.Tree` 构建层级项目，每个项目构建在所属项目内部，适合节点不多的树。

**关键 API**：`ui.Tree(c, func())` + `ui.TreeItem(c, label, *bool, children func())`；`children` 或 `*bool` 为 nil 即叶节点；`Selected` 显示选中，`Clicked` 报告单击非箭头区域或 Enter。整棵树只占一个 Tab 停留点，Up/Down 移动焦点，Home/End 到首尾，Right 展开或进入第一个子项目、Left 收起或移到父项目（从右向左布局时反向），输入首字母定位。`ui.TreeBase` 是无外观版本：`Root`、`Item(*bool)`、`item.Row`、`item.Group`、`Depth()`，在 `Item` 之前用 `c.Key` 设置键，展开收起后焦点才能留在同一项目。

```go
ui.Tree(c, func() {
	item := func(path, label string, open *bool, children func()) {
		if ui.TreeItem(c, label, open, children).Selected(app.chosen == path).Clicked() {
			app.chosen = path
		}
	}
	item("src", "src", &app.srcOpen, func() { item("src/main.go", "main.go", nil, nil) })
	item("go.mod", "go.mod", nil, nil)
})
```

**限制**：`Tree` 构建它显示的每个项目，数千项或带列的树改用大纲；展开状态由 `*bool` 持有，需应用保存；`TreeBase` 不自带主题外观、缩进和箭头。

## 大纲

`ui.Outline` 以树形结构显示项目，只构建可见行，类似 AppKit 的大纲视图和 Finder 的列表视图。

**关键 API**：`ui.Outline(c, *ui.OutlineState[T], []T, func(T) []T, func(T))`；`children` 返回 nil 表示叶节点。`OutlineState` 的 `List` 提供列表的选择与滚动，`Open` 保存已展开项目，`Rows`/`Item`/`Depth` 读取当前行，`Indent` 设缩进（默认主题 `Space(4)`，不能为负），`Prefix` 自定义行首，`Generation` 缓存行。单击箭头展开收起，Option/Alt 单击递归展开收起，Right 展开或进入第一个子项目，Left 收起或移到父项目；`OutlineTable` 在表格列中提供同样能力。

```go
app.files.List.Selected = &app.row
ui.Outline(c, &app.files, []string{"/"}, func(dir string) []string {
	return app.children[dir] // nil for a file
}, func(path string) {
	ui.Text(c, filepath.Base(path))
}).Grow(1)
```

**限制**：每帧遍历所有已展开项目，开销与展开数量成正比；用 `Generation` 缓存行时，`children` 结果变化必须递增，否则显示旧行；收起项目时若其子项目被选中，选择移到被收起的项目；`OutlineTable` 中 Left/Right 仍用于展开收起，列只能用指针选择。

## 网格视图

`ui.GridView` 以固定行高、自适应列数展示大量条目，只构建可见条目，类似照片图库或 Finder 的图标视图。

**关键 API**：`ui.GridView(c, *ui.GridState, n, minWidth, rowHeight float32, func(i int))`（示例为 `140, 120`）；`Gap` 设间距；`GridState` 的 `Selected`、`Selection`（以 `Key` 区分）、`Label`、`Reorder`、`Submitted`；方向键在行列间移动，Home/End 跳到首尾。

```go
app.photos.Selected = &app.photo
ui.GridView(c, &app.photos, len(photos), 140, 120, func(i int) {
	ui.Image(c, thumbs[i]).Fit(ui.Contain).Grow(1)
	ui.Text(c, photos[i].Name).SingleLine()
}).Grow(1)
```

**限制**：行高由参数统一指定、列数按最小列宽自适应，条目不能单独设置尺寸；只支持规则网格，需要跨行跨列或自定义轨道时用 `ui.Grid`；`ui.Grid` 会构建全部子元素，不要用它显示大量条目。

## 滚动视图

`ui.Scroll`、`ui.ScrollHorizontal`、`ui.ScrollBoth` 用滚轮、触摸板、滚动条和键盘滚动子元素，滚轮带短暂缓出；应用通过状态或方法滚动容器后由应用接管。容器需要尺寸或 `Grow`。

**关键 API**：`TrackScroll(&ui.ScrollState)` 读写 `X`/`Y`/`MaxX`/`MaxY`（`Y = 0` 为顶部，`math.MaxFloat32` 为末尾）；`ScrollTo(x, y, animation)`、`Scrolling()`；`Sticky().Top(0)`/`Start(0)` 等把元素固定在最近的滚动容器内；`ScrollSnap(ui.SnapMandatory|ui.SnapProximity)` 配合 `SnapAlign(ui.Start|Center|End)` 做吸附；`ScrollIntoView`/`ScrollIntoViewOptions{Block, Inline, Animation}` 在布局后滚动到元素；`ScrollbarStyle`、`ScrollbarInsets`、`Scrollbars(false)`、`ScrollbarAxes(h, v)`、`KeepScrollbars(true)` 控制滚动条，这组设置对 `List` 和 `Table` 同样有效。

```go
ui.Scroll(c).Fill().Children(func() {
	for _, day := range app.days {
		ui.Text(c, day.Title).Bold().Padding(8, 16).Background(t.Surface).Sticky().Top(0)
		for _, ev := range day.Events { eventRow(c, ev) }
	}
})
```

**限制**：容器构建全部子元素，长数据集合应使用 `List`；`Sticky` 对 `Absolute` 元素和文本中的元素无效，元素留在布局流内、需要背景，两端内缩冲突时纵向以 `Top` 为准、横向以 `Left` 为准；`ScrollIntoView` 只在添加元素的那一帧调用，否则用户无法把元素滚走；列表分区标题用 `ListState.Header`，不要用 `Sticky`。

## 网格

`ui.Grid` 像 CSS grid 一样把子元素逐行填入列，按需增加行，适合元素不多的二维布局。

**关键 API**：`Columns(n)`、`ColumnTracks(ui.Fixed(220), ui.Fr(1), ui.FitContent())`、`RowTracks`、`Rows`、`Gap`/`GapX`/`GapY`。轨道尺寸：`Fr` 分摊剩余空间且不会因内容变宽（`minmax(0, 1fr)`），`.Min(w)` 设最小尺寸，`.MinContent()` 不小于内容的 min-content 宽。放置用 `ColumnSpan`/`RowSpan`（负值到末列/末行）、`ColumnStart`/`RowStart`（从 1 计数）；子元素默认拉伸，用 `JustifyItems`/`AlignItems`、`JustifySelf`/`AlignSelf` 或自动外边距改变，`AlignItems(ui.Baseline)` 按行对齐首行文本；子元素百分比相对其单元格。

```go
ui.Grid(c).ColumnTracks(ui.FitContent(), ui.Fr(1)).GapX(16).GapY(8).Children(func() {
	ui.Text(c, "Name").Bold().ColumnSpan(-1) // across every column
	ui.Text(c, "Email")
	ui.TextInput(c, &app.email)
	ui.Text(c, "Bio")
	ui.TextArea(c, &app.bio).RowSpan(2)
})
```

**限制**：网格构建全部子元素，条目很多时用 `GridView`；`.Min` 轨道宁可让网格溢出盒子也不挤压内容，溢出交给外层滚动容器；纯 `Fr` 轨道按 `minmax(0, 1fr)` 分配空间，内容过宽时溢出单元格；控件旁的标签用 `Form`，不要手写两列网格。

## 分栏

`ui.Split` 并排显示两个窗格，分隔条可拖动，获得焦点后也可用方向键调整。

**关键 API**：`ui.Split(c, &w float32, first, second func())`、`SplitVertical`（上下，`*float32` 为高度）、`SplitPercent`/`SplitVerticalPercent`（0–100）。分隔条与两侧边缘至少相距 40 DIP；空间不足时第一窗格缩小、为第二窗格保留 40 DIP，空间恢复后回到用户设置的宽度，`*float32` 始终保存用户值。`Changed()` 报告位置变化，`SplitReset(dip 或百分比)` 让双击分隔条恢复该尺寸，`SplitGap(dip)` 用间距代替 1 DIP 分隔线（0 为无缝）。

```go
ui.Split(c, &app.sidebarWidth, func() { app.files(c) }, func() { app.editor(c) }).
	SplitReset(220).Fill()
```

**限制**：分栏不持久化尺寸，需应用保存 `*float32`；未设置 `SplitReset` 时双击分隔条无效；窗口过小时保留设定值，空间恢复后才生效；文档未说明 40 DIP 的边距可配置。

## 可折叠

`ui.Collapsible` 是展开控件，标签旁有箭头，单击或按 Enter、Space 展开折叠，类似 SwiftUI 的 `DisclosureGroup`。

**关键 API**：`ui.Collapsible(c, label, *bool, func())`；`*bool` 为 true 时构建内容，`Changed` 报告单击；展开时箭头旋转、内容逐渐展开，系统要求减少动效时立即显示。`ui.CollapsibleBase(c, *bool)` 是无外观版本：`Trigger` 显示标签，`Panel(func())` 构建内容，`Progress()` 返回 0（关闭）到 1（打开）的展开进度。

```go
ui.Collapsible(c, "Advanced", &app.advanced, func() {
	ui.Checkbox(c, &app.verbose, "Verbose logging")
	ui.Checkbox(c, &app.experimental, "Experimental features")
})
```

**限制**：一个容器里放多个可折叠分区时用 `Accordion`；`CollapsibleBase` 不自带标签、箭头、间距和主题外观；`Progress` 是动画进度，不是目标状态。

## 手风琴

`ui.Accordion` 在带边框的容器中纵向排列分区，每个 `ui.AccordionItem` 独立开合。

**关键 API**：`ui.Accordion(c, func())` + `ui.AccordionItem(c, title, *bool, func())`；单击标题栏或按 Enter、Space 切换，Up/Down 在标题间移动焦点，Home/End 跳到首尾，`Changed`/`OnChange` 报告切换；分区展开时按 Tab 从标题进入内容。`ui.AccordionBase` 是无外观版本：`Root` 列，`Item(&open)` 返回 `Trigger`、`Panel`、`Progress()`，按键行为相同；`Accordion` 把焦点环画在标题内部（`FocusRingOffset`）。

```go
for i, s := range sections {
	open := app.section == i
	ui.AccordionItem(c, s.Title, &open, s.Build).OnChange(func() {
		app.section = -1
		if open { app.section = i }
	})
}
```

**限制**：分区默认各自独立，一次只开一个需在 `OnChange` 中收起其余分区；`AccordionBase` 不绘制边框、标题栏和焦点环；展开高度由内容决定，文档未说明最大高度或内部滚动。

## 表单

`ui.Field` 给控件加标签，标签为辅助技术命名控件、单击时聚焦或切换控件；`Description` 和 `Error` 在控件下方显示说明与错误。

**关键 API**：`ui.Field(c, label, func())`；`Error` 同时把控件的值标记为无效，文本框用主题的 `Danger` 画边框，`Error` 为空则有效。`ui.Form(c, func())` 把标签放在控件左侧独立列，右对齐、与最宽标签同宽、与控件首行文本对齐，无文本控件居中；表单外的 `Field` 标签在控件上方。`ui.Fieldset(c, title, func())` 分组并命名，`Disabled` 禁用整组；`LabelledBy`、`DescribedBy`、`ErrorMessage`、`Invalid` 用于自行绘制的标签、提示和错误。

```go
ui.Field(c, "Email", func() {
	ui.TextInput(c, &app.email)
}).Description("For receipts.").Error(invalid)
```

**限制**：字段的控件是回调中第一个可获得焦点的元素或一组元素（如单选按钮组）；自带名称的控件（如带文字的复选框）保留自身名称，字段标签改为命名它所在的组；没有错误消息时用 `Invalid` 而不是 `Error`；提示和错误作为独立元素绘制时，用 `DescribedBy`/`ErrorMessage` 与控件关联。

## Web 视图

`ui.WebView` 在原生窗口中显示由系统 webview（WKWebView、WebView2、WebKitGTK）渲染的页面，例如应用前端、文稿或 HTML 预览。

**关键 API**：用 `win.NewWebView(mygo.WebViewOptions{URL, Page, BackgroundColor, Transparent, ScrollToContainer})` 创建，`WebView.Page` 提供 `LoadURL`/`LoadHTML`/`Eval`/`GoBack`/`OnDidNavigate`/`OnWillNavigate`；`WebView` 提供 `Focus`/`SetBackgroundColor`/`CapturePage`/`Destroy`。`ui.WebView(c, wv)` 本身无固有尺寸，靠拉伸、`Grow` 或 `Height`；页面裁剪到内容盒与 `Radius`，只有圆角内的指针事件传给页面，之后绘制的内容覆盖在页面上。`ScrollToContainer` 把滚轮和触摸板交给周围原生滚动容器，页面自身不滚；页面未处理的按键交给原生 UI；页面绑定与窗口页面相同（`mygo.Bind`、`mygo.NewEvent`、`Event.Emit`/`EmitPage`/`Broadcast`）。

```go
app.docs, _ = win.NewWebView(mygo.WebViewOptions{URL: "/docs"})
// in view:
if app.tab == "docs" {
	ui.WebView(c, app.docs).Grow(1).Radius(8)
} else {
	ui.WebView(c, app.changelog).Grow(1)
}
```

**限制**：web 视图只创建一次，视图只负责显示，不要在视图里创建；未构建的 web 视图隐藏但保留页面，标签页切回来时页面还在原处；第三方或用户 HTML 必须放在不受信任页面（`PageOptions.Untrusted`）或不受信任来源，非应用页面不能调用 Go；MyGo 不绘制页面本身，窗口捕获在页面处是洞，glass/模糊等效果看不到页面内容；`GOOS=js` 目标没有 web view。

## 标签页

`ui.Tabs` 创建一行标题，`*int` 保存所选索引，在标签栏下方构建所选页面。

**关键 API**：`ui.Tabs(c, &index, titles...)`；单击、方向键、Home/End 选择，焦点随选择移动，`Changed` 报告变化；整行标签页只占一个 Tab 停留点。`ui.TabsBase(c, &index, n)` 是无外观版本：`List` 容纳标签页，`Tab(i)` 构建一页，`Disabled(true)` 的标签页会被单击、方向键和 Home/End 跳过。

```go
ui.Tabs(c, &app.tab, "General", "Appearance", "Advanced")
switch app.tab {
case 0: app.general(c)
case 1: app.appearance(c)
case 2: app.advanced(c)
}
```

**限制**：`Tabs` 只画标签栏，页面内容由应用按索引构建；需要历史记录和后退时用路由器（`ui/navigation`），在工具栏切换视图用分段控件；关闭按钮、可滚动标签栏与溢出行为文档未说明。

## 侧边栏

`ui.Sidebar` 创建源列表，类似 Finder 和 Mail 的侧边栏：图标、分区、可折叠节，选中项 ID 写入 `*string`，内容放不下时可滚动。

**关键 API**：`ui.Sidebar(c, &id, func()).Width(220)`；`Changed` 报告选择。`ui.SidebarItem(c, id, icon, label)`，图标可为 nil，`Children` 添加徽章等内容；`ui.SidebarSection(c, title, *bool, func())`，`*bool` 为 nil 时该节始终展开，单击标题或箭头展开收起，指针悬停在标题上才显示箭头。单击或 Up/Down/Home/End、输入首字母选择，整个侧边栏只占一个 Tab 停留点；有焦点时选中项用强调色，否则显示为灰色。

```go
ui.Sidebar(c, &app.mailbox, func() {
	ui.SidebarSection(c, "Mailboxes", &app.mailboxesShown, func() {
		ui.SidebarItem(c, "inbox", inboxIcon, "Inbox").Children(func() {
			ui.Badge(c, fmt.Sprint(app.unread))
		})
		ui.SidebarItem(c, "sent", sentIcon, "Sent")
	})
}).Width(220)
```

**限制**：需要显式设置宽度；文档只描述单选，未说明多选；与路由器配合时把项 ID 设为页面路径，在 `Changed` 时 `router.Push(page)`（见 `ui/navigation`）。

## 面包屑导航

`ui.Breadcrumbs` 创建路径导航，各项用尖角分隔，除最后一项外都是链接。

**关键 API**：`ui.Breadcrumbs(c, path []string, &index int).Label("Path").Changed()`；单击或 Enter 选中某项后写入索引，`Changed` 报告变化；空间不足时较长条目以省略号截断；最后一项是当前位置，不是链接。辅助技术把它识别为导航地标中的链接，`Label` 命名地标，最后一项读作当前页（`ui.CurrentPage`）；自定义路径导航用 `Role(ui.RoleNavigation)` 加 `Current(ui.CurrentPage)`。

```go
path := []string{"Macintosh HD", "Users", "ada", "Documents"}
if ui.Breadcrumbs(c, path, &app.chosen).Label("Path").Changed() {
	app.open(path[:app.chosen+1])
}
```

**限制**：组件只写索引、不做导航，应用需按索引打开路径；最后一项不可选；分隔符号样式、自定义图标和横向溢出滚动文档未说明。

## 后退和前进按钮

`ui.BackButton` 和 `ui.ForwardButton` 在路由器的历史中后退和前进，显示折线箭头，到达历史端点时禁用。

**关键 API**：`ui.BackButton(c, *ui.Router)`、`ui.ForwardButton(c, *ui.Router)`，通常放在 `ui.Toolbar` 中；右键单击按路由的 `Title` 列出可后退或可前进的页面（无标题时按路径），最近的在前，选中立即跳转。快捷键与按钮一致：macOS 为 `Cmd+[` 和 `Cmd+]`，其他平台为 `Alt+Left` 和 `Alt+Right`，鼠标侧键也可后退和前进。

```go
ui.Toolbar(c, func() {
	ui.BackButton(c, app.router)
	ui.ForwardButton(c, app.router)
	ui.Text(c, app.router.Title()).Bold()
})
```

**限制**：按钮不创建路由器，`*ui.Router` 必须已在应用状态中创建；页面切换与标题仍由 `router.View` 中的 `Title` 决定；按钮外观自定义文档未说明。

## 速查

| 控件 | 入口 | 关键状态/成员 | 惰性构建 |
| --- | --- | --- | --- |
| 列表 | `ui.List` | `ListState`：`Key`、`Selected`/`Selection`、`NearEnd`/`EndReached` | 是 |
| 表格 | `ui.Table` | `ListState` + `TableColumn`：`Width`、`Sortable`、`Frozen` | 是（行；列用 `VirtualColumns`） |
| 树 | `ui.Tree` / `TreeBase` | `*bool` 展开、`Selected`/`Clicked` | 否 |
| 大纲 | `ui.Outline` / `OutlineTable` | `OutlineState`：`List`、`Open`、`Generation` | 是 |
| 网格视图 | `ui.GridView` | `GridState`：`Selected`、`Selection`、`Reorder` | 是 |
| 滚动视图 | `ui.Scroll` / `ScrollHorizontal` / `ScrollBoth` | `ScrollState`：`X`/`Y`/`MaxX`/`MaxY` | 否 |
| 网格 | `ui.Grid` | 轨道、`ColumnSpan`/`RowSpan`、对齐 | 否 |
| 分栏 | `ui.Split` / `SplitVertical`（及百分比变体） | `*float32`、`SplitReset`、`SplitGap` | 否 |
| 可折叠 | `ui.Collapsible` / `CollapsibleBase` | `*bool`、`Trigger`/`Panel`/`Progress` | 否 |
| 手风琴 | `ui.Accordion` / `AccordionBase` | 每个分区一个 `*bool` | 否 |
| 表单 | `ui.Form` / `Field` / `Fieldset` | 标签、`Description`/`Error`、`Invalid` | 否 |
| Web 视图 | `ui.WebView` | `mygo.WebView`、`WebViewOptions` | 否 |
| 标签页 | `ui.Tabs` / `TabsBase` | `*int` 所选索引 | 否 |
| 侧边栏 | `ui.Sidebar` | `*string` 选中 ID、`Changed` | 文档未说明 |
| 面包屑 | `ui.Breadcrumbs` | `[]string` + `*int` | 否 |
| 历史按钮 | `ui.BackButton` / `ForwardButton` | `*ui.Router` | 否 |
