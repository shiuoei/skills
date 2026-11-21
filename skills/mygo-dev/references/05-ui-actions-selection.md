# UI 动作与选择控件

覆盖 `ui` 包中 16 个触发动作或做选择的控件。动作控件只报告“被触发”，选择控件把结果写入绑定值：`*bool`（复选框、开关、切换按钮）、`*int`（分段控件、评分）、`*float64`（滑块、范围滑块、步进器）、`*string`（选择器），单选按钮可绑定任意可比较类型。

通用约定：构造函数接收 `*ui.Context` 与绑定值指针；`Clicked`/`Changed`/`Chosen` 查询本帧输入，已处理的输入在视图重建前被消耗；外观用 `ui.Element` 的链式方法（`Padding`、`Radius`、`Background`、`TextColor`、`Disabled`）；空标签加 `Children(fn)` 可以放图标，纯图标控件要给控件本身设 `Label` 与 `Tooltip`。

选型速记：即时生效的布尔设置用开关，随表单保存的用复选框，工具栏状态用切换按钮；少量枚举横排用分段控件或单选按钮，多项用选择器；数值用滑块或步进器，评分用星星。

## 源页面

| 页面 | 在线地址 | 页面 | 在线地址 |
| --- | --- | --- | --- |
| 按钮 | [button](https://mygo.run/zh/docs/ui/button) | 链接 | [link](https://mygo.run/zh/docs/ui/link) |
| 菜单按钮 | [menu-button](https://mygo.run/zh/docs/ui/menu-button) | 拆分按钮 | [split-button](https://mygo.run/zh/docs/ui/split-button) |
| 上下文菜单 | [context-menu](https://mygo.run/zh/docs/ui/context-menu) | 切换按钮 | [toggle](https://mygo.run/zh/docs/ui/toggle) |
| 分段控件 | [segmented](https://mygo.run/zh/docs/ui/segmented) | 工具栏 | [toolbar](https://mygo.run/zh/docs/ui/toolbar) |
| 复选框 | [checkbox](https://mygo.run/zh/docs/ui/checkbox) | 开关 | [switch](https://mygo.run/zh/docs/ui/switch) |
| 单选按钮 | [radio](https://mygo.run/zh/docs/ui/radio) | 选择器 | [select](https://mygo.run/zh/docs/ui/select) |
| 滑块 | [slider](https://mygo.run/zh/docs/ui/slider) | 范围滑块 | [range-slider](https://mygo.run/zh/docs/ui/range-slider) |
| 步进器 | [stepper](https://mygo.run/zh/docs/ui/stepper) | 评分 | [rating](https://mygo.run/zh/docs/ui/rating) |

## 动作控件

### 按钮 `ui.Button`
**用途**：显示标签并在单击时报告；`ui.PrimaryButton` 是强调色的主要操作按钮。
**关键 API**：`ui.Button(c, label)`、`ui.PrimaryButton(c, label)`；`Clicked()`、`OnClick(fn)`、`Disabled(bool)`、`AccessDisabled(bool)`、`Label(s)`、`Tooltip(s)`、`Children(fn)`；`ui.ButtonBase` 是无预设外观的可聚焦按钮行，供自定义外观。
**示例**：
```go
if ui.Button(c, "Cancel").Clicked() {
	app.editing = false
}
ui.PrimaryButton(c, "Save").Disabled(app.saving).OnClick(app.save)
```
**限制**：`Disabled(true)` 同时阻止单击与焦点；只让辅助技术知道不可用、但仍可到达，用 `AccessDisabled(true)`。按钮有焦点时自己处理 Enter 和 Space；要在焦点别处时也用 Enter 提交，用 `c.Shortcut(0, ui.KeyEnter)`。纯图标按钮用空标签加 `Children`，并设 `Label` 与 `Tooltip`。

### 链接 `ui.Link`
**用途**：文本链接，单击或聚焦后按 Enter 在浏览器或对应协议的应用中打开；在路由器页面中可跳转内部路径。
**关键 API**：`ui.Link(c, label, url)`、`Clicked()`、`Children(fn)`；URL 留空时只报告单击，用于应用内动作。
**示例**：
```go
ui.Link(c, "Privacy policy", "https://example.com/privacy")
if ui.Link(c, "Show details", "").Clicked() {
	app.details = true
}
```
**限制**：Space 滚动外围容器而不是打开链接。不带协议前缀的路径（`/notes/42`、`edit`、`?tab=info`）通过路由器跳转，效果同 `Push`，相对路径以当前页面为基准。链接使用强调色，悬停显示下划线；在 `ui.RichText` 中作为段落的一部分随文字换行。

### 菜单按钮 `ui.MenuButton`
**用途**：带标签和箭头的按钮，按下时在下方打开系统菜单；具有焦点时 Enter、Space、Up、Down 也能打开。
**关键 API**：`ui.MenuButton(c, label, func(m *ui.Menu))`；菜单项与上下文菜单相同（`Item`、`Checked`、`Radio`、`Disabled`、`Shortcut`、`Chosen`、`Separator`、`Submenu`）；任意按钮可用 `Element.Menu(fn)` 获得同样的菜单。窗口内自绘：`ui.PopoverMenuButton` 或 `Element.PopoverMenuButton(fn)`，选项 `ui.PopoverMenuButtonOptions{At, Self, Open, Panel}` 放在函数之后。
**示例**：
```go
ui.MenuButton(c, "Sort by", func(m *ui.Menu) {
	for _, by := range []string{"Name", "Date", "Size"} {
		if m.Item(by).Checked(app.sort == by).Chosen() {
			app.sort = by
		}
	}
})
```
**限制**：菜单函数在打开时构建菜单项，并在选中某项后的那一帧再次运行，由 `Chosen` 报告结果；若该选择会改动正在构建的集合（如删除一行），先记录，循环结束后再改。`Open` 是 `*bool`，可由按钮更新，也可由应用设置；一个窗口一次只显示一个自绘菜单。`Panel` 在主题样式之后设置菜单面板外观，`NoShadow` 可移除主题阴影。

### 拆分按钮 `ui.SplitButton`
**用途**：主操作区段与箭头菜单区段连在一起，如 Save 与 Save As…。
**关键 API**：`ui.SplitButton(c, label)`、`ui.PrimarySplitButton(c, label)`；返回值含 `Group`、`Button`、`Arrow`；`Arrow.Menu(fn)`、`Arrow.PopoverMenuButton(fn, opts)`；`Group.Disabled(bool)` 禁用整个控件，`Button.Disabled(bool)` 只禁用主操作、箭头仍可用。
**示例**：
```go
save := ui.SplitButton(c, "Save")
save.Arrow.Menu(func(m *ui.Menu) {
	if m.Item("Save As…").Chosen() {
		app.saveAs()
	}
})
if save.Button.Clicked() {
	app.save()
}
```
**限制**：整个控件只占一个 Tab 停留点，Left/Right 在区段间移动焦点（从右到左时方向相反）；按钮区段有焦点时 Enter/Space 触发，箭头区段有焦点时 Enter/Space/Up/Down 打开菜单，也可用 Alt+Down。外边距、对齐与尺寸加在 `Group` 上。区段可用 `Button.Label` 与 `Arrow.Label` 命名，测试中按名称点击。

### 上下文菜单 `ContextMenu` 与 `PopoverMenu`
**用途**：右键单击（macOS 也可 Control+单击）、菜单键或 Shift+F10 打开的菜单。
**关键 API**：`Element.ContextMenu(fn)` 用系统菜单；`Element.PopoverMenu(fn)` 在窗口内自绘（v0.9.0 的形态；v0.10 起还可传 `ui.PopoverMenuOptions{Panel, Item}` 自定义面板与菜单项，v0.10+）。菜单项：`m.Item(label)` 返回项，支持 `Checked(on)`、`Radio(on)`、`Disabled(d)`、`Shortcut(mods, key)`、`Tooltip(s)`、`Subtitle(s)`、`Icon(icon)`、`IconColor(color)`、`ID(id)`、`Chosen()`；`m.Separator()`、`m.Submenu(label, build)`、`m.EditItems()`；自绘菜单另有 `Danger()`、`Trailing(s)`、`Content(fn)`、`Highlighted()`、`m.MaxWidth(w)`。无外观版本：`ui.ContextMenuBase(c, el, &open)` 加 `Popup(fn)`；`Context.SetMenuPresentation(ui.PopoverMenus)` 让窗口内所有 MyGo 菜单自绘（默认 `ui.SystemMenus`）。
**示例**：
```go
row.ContextMenu(func(m *ui.Menu) {
	if m.Item("Rename").Shortcut(0, ui.KeyF2).Chosen() {
		app.renaming = true
	}
	m.Separator()
	if m.Item("Delete").Disabled(note.locked).Chosen() {
		deleted = i
	}
})
```
**限制**：单击由最内层带菜单的元素处理（行与行内按钮各有菜单时打开按钮的）。文本输入与可选文本自带编辑菜单，`ContextMenu` 会替换它，`EditItems` 把这些命令加入自己的菜单。若某个选择会把该行移出集合，先记下，循环结束后再改。系统菜单对这些扩展项只显示标签，`IconColor`、`Danger`、`Trailing`、`Content`、`MaxWidth` 只在自绘菜单生效（`Icon` 系统菜单也显示，浏览器除外）。自绘菜单会获得焦点并支持方向键、Home/End、子菜单左右键、Enter/Space、Escape/Tab。

### 工具栏 `ui.Toolbar`
**用途**：把按钮、切换按钮、切换按钮组、分段控件、拆分按钮和菜单按钮横向排成工具栏，通常放在窗口顶部。
**关键 API**：`ui.Toolbar(c, func(){...})`、`Label(s)`；`ui.Spacer(c)` 把后面的控件推到末端。键盘上整个工具栏只占一个 Tab 停留点，Left/Right 在控件间移动，Home/End 到首尾。
**示例**：
```go
ui.Toolbar(c, func() {
	ui.ToggleGroup(c, func() {
		ui.Toggle(c, &app.bold, "Bold")
		ui.Toggle(c, &app.italic, "Italic")
	})
	ui.Spacer(c)
	if ui.Button(c, "Share").Clicked() {
		app.share()
	}
}).Label("Format")
```
**限制**：工具栏中的控件在指针悬停前不显示独立的按钮外观。放不下的控件从末尾开始收进工具栏末端的溢出菜单，选中项可触发按钮、切换状态或打开菜单按钮的菜单。

### 切换按钮 `ui.Toggle`
**用途**：在 `*bool` 为 true 时保持按下状态的按钮，如编辑器工具栏中的 Bold。
**关键 API**：`ui.Toggle(c, &v, label)`、`Changed()`、`Label(s)`、`Children(fn)`；`ui.ToggleGroup(c, func(){...})` 把切换按钮和普通按钮连成一个控件，可 `Label(s)`、`FillEqually()`；`ui.ToggleBase` 是无外观版本。
**示例**：
```go
ui.Toggle(c, &app.bold, "").Label("Bold").Children(func() {
	ui.Icon(c, boldIcon)
})
```
**限制**：单击或聚焦后按 Space 切换并写入 `*bool`；`FillEqually` 让按钮组填满父元素宽度并等分。要在几个选项中选一个，用分段控件。

### 分段控件 `ui.Segmented`
**用途**：用 `*int` 保存选中分段索引的单选式分段控件，如视图切换。
**关键 API**：`ui.Segmented(c, &index, labels...)`、`Changed()`、`Label(s)`、`FillEqually()`；`ui.SegmentedBase(c, &index, n)` 是无外观版本，`Track` 设置轨道样式、`Segment(i)` 返回各段，构建后对 `Track` 调 `FillEqually`。
**示例**：
```go
ui.Segmented(c, &app.view, "List", "Grid").Label("View")
switch app.view {
case 0:
	app.list(c)
case 1:
	app.grid(c)
}
```
**限制**：一个 Tab 停留点，Tab 跳到选中段，方向键改选择。`FillEqually` 让控件与父元素同宽并在分段间平分，之后设 `Width` 指定整个控件宽度。图标或自定义外观用 `SegmentedBase`。

## 选择控件

### 复选框 `ui.Checkbox`
**用途**：带标签的复选框，切换 `*bool`；适合随表单一起保存的选择。
**关键 API**：`ui.Checkbox(c, &v, label)`、`Changed()`；`ui.CheckboxGroup(c, label, func(){...})` 生成带总复选框的组（全部勾选/混合/全不勾）；`ui.CheckboxState(c, *ui.CheckState, label)` 支持混合态，取值为 `ui.CheckOff`、`ui.CheckOn`、`ui.CheckMixed`；`ui.CheckboxBase`、`ui.CheckboxStateBase` 无外观。
**示例**：
```go
if ui.Checkbox(c, &app.showHidden, "Show hidden files").Changed() {
	app.reload()
}
all := ui.CheckOff
if n := app.selected(); n == len(app.rows) {
	all = ui.CheckOn
} else if n > 0 {
	all = ui.CheckMixed
}
if ui.CheckboxState(c, &all, "").Label("Select all").Changed() {
	app.selectAll(all == ui.CheckOn)
}
```
**限制**：单击复选框或标签就立即写值，`Changed` 只判断本帧是否发生了操作。混合态单击后变为勾选，已勾选时取消。`CheckboxGroup` 的总复选框部分勾选时显示混合状态并向内缩进其子复选框。

### 开关 `ui.Switch`
**用途**：切换 `*bool` 的开关，外观上不显示文本，适合立即生效的设置。
**关键 API**：`ui.Switch(c, &v)`、`Label(s)`、`Changed()`；`ui.SwitchBase` 是无外观版本。
**示例**：
```go
ui.Row(c).Gap(12).Children(func() {
	ui.Text(c, "Wi-Fi").Grow(1)
	ui.Switch(c, &app.wifi).Label("Wi-Fi")
})
```
**限制**：开关不显示文本，必须设 `Label`，或放进 `ui.Field` 让字段标签为它命名（单击标签会切换）。单击或聚焦后按 Space 切换。要在表单保存时才生效，用复选框。

### 单选按钮 `ui.Radio`
**用途**：把选中的值写入 `*T`，一组中选一个；值可以是任何可比较类型。
**关键 API**：`ui.Radio(c, &v, value, label)`、`Changed()`；`ui.RadioGroup(c, func(){...})` 把按钮组作为一个 Tab 停留点，`Row()` 改为横排，组可 `Label(s)`；`ui.RadioBase` 是无外观版本。
**示例**：
```go
ui.RadioGroup(c, func() {
	for _, size := range []string{"Small", "Medium", "Large"} {
		ui.Radio(c, &app.size, size, size)
	}
}).Label("Size")
```
**限制**：组默认一列；Tab 跳到选中项，方向键以及 Home/End 改选择。要横排加 `.Row().Gap(16)`。许多选项用选择器或组合框，视图切换用分段控件。无外观版本可用 `DimDisabled(false)` 让禁用项仍清晰可读。

### 选择器 `ui.Select`
**用途**：下拉选择器，从字符串列表中选一个值写入 `*string`。
**关键 API**：`ui.Select(c, &s, []string{...})`、`Label(s)`、`Changed()`；`ui.SelectBase(c, &v)` 可绑定任意可比较类型，用 `Trigger`、`Popup(fn)`、`Item(v)`、`Highlight(v)`、`Open()`、`SetOpen(bool)` 自行绘制；`ui.SelectBaseOpen(c, &v, *bool)` 由应用提供打开状态。
**示例**：
```go
ui.Select(c, &app.plan, []string{"Free", "Pro", "Team"}).Label("Plan")
```
**限制**：单击或聚焦后 Enter/Space/Up/Down 打开选项，方向键与 Home/End 移动，Enter/Space/单击选中，Escape 关闭；下方无空间时显示在上方，选项多到窗口容纳不下时高度等于可用空间并可滚动。`SelectBase` 的选项在弹出层打开时构建，`SetOpen(true)` 后若选项尚未构建，`Highlight` 要等构建之后再调用。长列表需要输入筛选用组合框；选项少且需要同时显示用单选按钮。

### 滑块 `ui.Slider`
**用途**：在上下限之间设置 `*float64`。
**关键 API**：`ui.Slider(c, &v, lo, hi)`、`Label(s)`、`Changed()`；`ui.StepSlider(c, &v, lo, hi, step)` 吸附到刻度并画刻度线；`ui.SliderBase(c, &v, lo, hi)` 用 `Draw` 自绘，另有 `Step`、`Vertical`、`PaddingX`/`PaddingY`。
**示例**：
```go
ui.Slider(c, &app.volume, 0, 100).Label("Volume")
ui.StepSlider(c, &app.quality, 0, 100, 25).Label("Quality")
```
**限制**：键盘有方向键、Page Up/Page Down（范围的十分之一）、Home 和 End。`SliderBase` 在扣除内边距后的内容区内拖动，按 `(value-lo)/(hi-lo)` 计算位置，内边距设为滑块柄宽度的一半以免越界；从右到左布局中低值在右侧，应按 `c.Direction() == ui.RightToLeft` 镜像绘制。同时设置上下界用范围滑块，精确数值用数字输入或步进器。

### 范围滑块 `ui.RangeSlider`
**用途**：两个旋钮在指定上下界内设置最小值和最大值，如价格范围。
**关键 API**：`ui.RangeSlider(c, &min, &max, lo, hi, step)`、`Label(s)`、`Changed()`。
**示例**：
```go
ui.RangeSlider(c, &app.minPrice, &app.maxPrice, 0, 1000, 10).Label("Price")
ui.Textf(c, "$%.0f – $%.0f", app.minPrice, app.maxPrice)
```
**限制**：每个旋钮都可获得焦点，Tab 在两者间切换，方向键和 Home/End 移动聚焦的旋钮；按下轨道移动离按下点最近的旋钮；两个旋钮不会互相越过。步长大于 0 时吸附到下界或下界加步长整数倍并显示刻度线，步长为 0 时连续调整。辅助技术看到两个滑块，名称是 `Label` 加“minimum”和“maximum”。

### 步进器 `ui.Stepper`
**用途**：一对箭头，在上下限之间按步长改变 `*float64`。
**关键 API**：`ui.Stepper(c, &v, lo, hi, step)`、`Label(s)`、`Changed()`。
**示例**：
```go
ui.Row(c).Gap(6).Children(func() {
	ui.Textf(c, "%.0f copies", app.copies)
	ui.Stepper(c, &app.copies, 1, 99, 1).Label("Copies")
})
```
**限制**：单击箭头或在获得焦点后按 Up/Down 增减；按住箭头持续调整，按得越久速度越快。用户还会键入数字时用数字输入，它自带步进箭头。

### 评分 `ui.Rating`
**用途**：一行星星，填充数量由 `*int` 指定，类似 AppKit 评分等级指示器。
**关键 API**：`ui.Rating(c, &n, stars)`、`Label(s)`、`Changed()`。
**示例**：
```go
ui.Rating(c, &app.stars, 5).Label("Rating")
```
**限制**：单击某颗星设置评分，再次单击当前分值对应的星星清除评分；聚焦后可用方向键、Home 和 End 调整。星星尺寸取自主题字号，指针停下时预示单击会设置的评分。辅助技术看到从 0 到星星数量的滑块。

## 速查

| 组件 | 一句话 | 关键 API |
| --- | --- | --- |
| 按钮 `ui.Button` | 标签按钮，报告单击 | `Button(c,l)`、`PrimaryButton`、`Clicked()`、`OnClick` |
| 链接 `ui.Link` | 文本链接，打开 URL 或内部路径 | `Link(c,l,url)`、`Clicked()`、空 URL 做动作 |
| 菜单按钮 `ui.MenuButton` | 按下列出菜单项 | `MenuButton(c,l,fn)`、`Item().Chosen()`、`PopoverMenuButton` |
| 拆分按钮 `ui.SplitButton` | 主操作加箭头菜单 | `SplitButton(c,l)`、`Button`/`Arrow`/`Group`、`Arrow.Menu` |
| 上下文菜单 | 右键或菜单键打开的菜单 | `ContextMenu(fn)`、`PopoverMenu(fn,opts)`、`EditItems()` |
| 工具栏 `ui.Toolbar` | 横向排列控件并按需收进溢出菜单 | `Toolbar(c,fn)`、`Spacer(c)`、`Label` |
| 切换按钮 `ui.Toggle` | 按下保持状态的布尔按钮 | `Toggle(c,&v,l)`、`Changed()`、`ToggleGroup` |
| 分段控件 `ui.Segmented` | 按索引选择的横排分段 | `Segmented(c,&i,labels...)`、`FillEqually` |
| 复选框 `ui.Checkbox` | 切换布尔值，可分组或混合态 | `Checkbox(c,&v,l)`、`CheckboxGroup`、`CheckboxState` |
| 开关 `ui.Switch` | 无可见文本的即时开关 | `Switch(c,&v)`、`Label`、`Changed()` |
| 单选按钮 `ui.Radio` | 一组中选一个值 | `Radio(c,&v,val,l)`、`RadioGroup` |
| 选择器 `ui.Select` | 字符串下拉选择 | `Select(c,&s,options)`、`SelectBase` |
| 滑块 `ui.Slider` | 在上下限内设置数值 | `Slider(c,&v,lo,hi)`、`StepSlider`、`SliderBase` |
| 范围滑块 `ui.RangeSlider` | 两个旋钮设置区间 | `RangeSlider(c,&lo,&hi,lo,hi,step)` |
| 步进器 `ui.Stepper` | 箭头按步长增减 | `Stepper(c,&v,lo,hi,step)` |
| 评分 `ui.Rating` | 星星表示的分值 | `Rating(c,&n,stars)`、`Changed()` |
