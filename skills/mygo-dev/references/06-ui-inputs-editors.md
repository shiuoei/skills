# UI 输入与编辑器

MyGo 的输入类能力分三层：绑定值的高层控件（`ui.TextInput`、`ui.NumberInput`、`ui.Calendar` 等）；把缓冲区、选区和撤销交给应用的输入原语（`TextInputBase`、`TextInputClient`、`TextBuffer`、`ShapeText`）；面向长文档的编辑器包（`ui/richtext` 编辑带属性文本、`ui/editor` 提供代码编辑器与差异查看器、`ui/markdown` 显示与流式渲染）。

**共同约定**：控件在视图构建函数里调用，首参 `*ui.Context`，元素方法链式返回；绑定值与缓冲区放进跨帧保留的应用状态。`Changed()`/`Submitted()` 报告"本帧发生的"变化；名称用 `Label`，校验错误用 `Field` 的 `Error`（`form` 页）；`Placeholder` 在 `SearchField`、`TokenField`、`Combobox` 上会传给内部输入框。绑定字符串既是编辑目标也是显示来源：用户编辑直接改写它，应用改写它会替换控件中的文本；要把应用文本并入用户编辑流，改用 `ReplaceText`（见下）。耗时工作放 `Submitted`，或用 `c.After` 在停顿后做。

## 源页面

- 文本与数值：[文本输入](https://mygo.run/zh/docs/ui/text-input)、[文本输入原语](https://mygo.run/zh/docs/ui/text-input-client)、[数字输入](https://mygo.run/zh/docs/ui/number-input)、[小数输入](https://mygo.run/zh/docs/ui/decimal-input)
- 选择与建议：[搜索框](https://mygo.run/zh/docs/ui/search-field)、[组合框](https://mygo.run/zh/docs/ui/combobox)、[自动补全](https://mygo.run/zh/docs/ui/autocomplete)、[标记字段](https://mygo.run/zh/docs/ui/token-field)
- 编辑与查找：[可编辑文本](https://mygo.run/zh/docs/ui/editable-text)、[查找栏](https://mygo.run/zh/docs/ui/find-bar)
- 长文档：[富文本编辑器](https://mygo.run/zh/docs/ui/rich-text-editor)、[代码编辑器](https://mygo.run/zh/docs/ui/code-editor)、[差异查看器](https://mygo.run/zh/docs/ui/diff-viewer)、[Markdown](https://mygo.run/zh/docs/ui/markdown)
- 日期与颜色：[日历](https://mygo.run/zh/docs/ui/calendar)、[日期输入](https://mygo.run/zh/docs/ui/date-input)、[时间输入](https://mygo.run/zh/docs/ui/time-input)、[颜色选择器](https://mygo.run/zh/docs/ui/color-picker)

## 文本输入

### `ui.TextInput` / `ui.TextArea`（text-input）

**用途** 编辑单行/多行 `*string`；输入直接改写绑定值，跟进处理用 `Changed`，单行框按 Enter 由 `Submitted` 报告。
**关键 API** `Placeholder`/`PlaceholderColor`、`CaretWidth`（可大于 1 DIP，0 隐藏）、`Password`（仅单行）、`AutoFocus`、`ReadOnly`、`Disabled`、`MaxLength`（UTF-16 代码单元，应用设置的值不裁剪，`n <= 0` 不限）；`OnPaste(func(string) bool)` 返回 true 接管粘贴；`OnTextKey(func(k ui.TextKey) bool)` 先于非文本键处理；`TextRanges` 按 rune 偏移着色/设字重；文本区用 `Lines(min, max)` 伸展并在超过时滚动；光标与选区用 `TextSelection`、`SetTextSelection`、`TextAnchorCaret`；行旁绘制用 `TrackLines`+`TrackScroll`+`ui.TextLines`（`Top`、`At`、`Rows`、`Height`）。
**最小示例**
```go
if ui.TextInput(c, &app.query).Placeholder("Search").Label("Search").Changed() {
	app.results = search(app.query)
}
if ui.TextInput(c, &app.name).AutoFocus().Submitted() {
	app.rename()
}
ui.TextArea(c, &app.notes).Lines(1, 8)
```
**限制/常见错误**
- 单行框把粘贴文本中的换行换成空格，`MaxLength` 也不裁剪 `ReplaceText` 写入的内容；要拒绝多行内容，只能在 `OnPaste` 里返回 true。
- `OnTextKey` 的回调可能在两帧之间执行：只能用传入的 `ui.TextKey`（`k.Key`、`k.Mods`、`k.Text()`、`k.SetText()`）；回调返回后再使用视图里留存的 `Element` 会 panic。
- 输入法组合期间按下的键归输入法（Enter 选候选、Escape 放弃组合），既不提交也不触发快捷键；`Composing()` 报告是否正在组合，此时文本尚未写入字符串。组合文本以无样式显示，`TextRanges` 在周围保持样式；密码框不显示范围。
- 直接赋值会替换文本但保留最后一个撤销步骤（置 `""` 也能 Undo 回来）；把应用文本并入编辑流用 `ReplaceText(start, end, s)`：偏移自动钳制，`0, math.MaxInt` 替换全部，光标落到末尾，作为一步撤销并经 `Changed` 报告。
- 重开编辑会话要清空字符串并改 `c.Key(...)`（丢弃选区、组合状态、撤销历史）；用 `ui.TextInputBuffer`/`ui.TextAreaBuffer` 编辑 `*ui.TextBuffer` 时，`buffer.Set("")` 会在下次同步时丢弃撤销/重做与组合状态。两种方式都不保证旧文本从内存抹去。无外观版本是 `TextInputBase`/`TextAreaBase`；文本区只排版视口内的段落，数十万行日志也能流畅输入。

### 输入原语与 `TextBuffer`（text-input-client）

**用途** 控件自带缓冲区、选区模型、标记文本或编辑策略时，把 `TextInputClient` 挂到元素上，MyGo 只负责它与平台输入法和文本查询的对接；内置 `TextInput`/`TextArea` 也实现同一契约。
**关键 API** `ui.Box(c).HandleTextInput(app.input).HandleInput(app.input.HandleInput).Draw(app.input.Draw).Label("Document").Role(ui.RoleTextField)`。契约：`TextForRange`、`Selection`、`MarkedRange`、`ReplaceText`、`SetMarkedText`、`UnmarkText`、`BoundsForRange`、`IndexForPoint`。范围按 **UTF-16 代码单元**（互操作坐标，非存储格式），`ui.UTF16Len`/`ui.UTF16ByteOffset` 帮助换算，代理对须完整。按索引存储用 `ui.TextBuffer`：`NewTextBuffer`、`Replace`/`Slice`（按 rune）、`ReplaceUTF16`、`UTF16Offset`/`RuneOffset`、`Version`、`Snapshot`、`Restore`、`String`、`WriteTo`。自绘几何用 `ShapeText`、`ShapeRichText`、`ShapeRichTextWith(spans, font, TextShape{Width, Align, Justify, Indent, NoBreakWords})`，布局对象给出 `Caret`、`IndexAt`、`SelectionRects`、`PreviousBoundary`/`NextBoundary`、`Lines`、`MoveCaret`、`SelectionRanges`、`Objects`。
**最小示例**
```go
document := ui.NewTextBuffer(initialText) // 跨帧保留，零值可用
in := ui.TextAreaBuffer(c, document).Fill().Font("monospace")
if in.Changed() {
	_ = document.Version()
}
```
**限制/常见错误** 回调在 UI 线程同步运行且在下一帧之前，不能等待需要该线程的工作；跨 goroutine 同步由客户端负责，改应用状态用 `Window.Update`。原始契约不提供键盘导航、指针选择、剪贴板策略和编辑菜单命令，这些都要由 `HandleInput` 自己实现。每次提交编辑时字符串绑定都会生成整个字符串，大文稿应改用 `TextBuffer`；`String` 显式展开全文，导出用 `WriteTo` 流式写。`TextCaret` 适合终端等只需输入法候选位置、不提供文稿的控件；`InputMethod(false)` 关闭输入法但仍可用按键输入（浏览器无法关闭）。

## 数值

### `ui.NumberInput`（number-input）

**用途** 编辑 `*float64`，限制在范围内；Up/Down 与右侧按钮按步长调整。输入成为范围内有效数字后立即生效，失焦时按步长的小数位数舍入**显示**；`Changed` 报告新值。
**关键 API** `ui.NumberInput(c, &app.opacity, 0, 1, 0.05).Label("Opacity")`；`Width` 是含按钮的整体宽度，文本框相应收窄。
**最小示例** `ui.NumberInput(c, &app.copies, 1, 99, 1).Label("Copies")`
**限制/常见错误** `float64` 无法精确表示金额，金额一律用 `DecimalInput`；只需按步长增减用步进器（`stepper` 页），只需粗略选择用滑块（`slider` 页）。

### `ui.DecimalInput`（decimal-input）

**用途** 面向金额的"数字输入"：用 `*string` 保存精确小数，内部以 `math/big` 整数按步长末位小数的位值计数，绝不以二进制浮点保存；值的小数位数与步长相同（步长 `"0.01"` → `"1234.50"`），上下限同写法，`""` 不限。
**关键 API** `ui.DecimalInput(c, &app.amount, "0", "", "0.01")`；小数点 `.`、千位分隔符 `,`、可带 `-`；`Error` 标记无效输入。
**最小示例**
```go
ui.Field(c, "Amount", func() {
	if ui.DecimalInput(c, &app.amount, "0", "", "0.01").Changed() {
		app.save()
	}
})
```
**限制/常见错误**
- 什么都不舍入：步长 `0.01` 时 `1.500` 是 `1.50`，但 `1.005` 不是它接受的数字 —— 保持原值并用 `Error` 标记，`Field` 在下方显示错误直到文本改正；非数字或越界同样标记且失焦后保留，笔误不会被悄悄替换。
- 清空输入会让值保持原样，失焦后再把值重新显示；应用设置值会替换当前内容（包括错误内容），若小数位多于步长或越界则按原样显示并标错。
- `Changed` 只报告数值变化，`1.5` 与 `1.50` 的写法差异不算变化；`1,50` 在任何语言下都不是 1.5，而是无效文本。

## 搜索与建议

### `ui.SearchField`（search-field）

**用途** 搜索专用文本框：放大镜图标、文本、有内容时的清空按钮（此时 Escape 也清空）；`Changed` 报告文本变化，Enter 由 `Submitted` 报告。
**关键 API** `ui.SearchField(c, &app.query).Label("Search mail")`；`Placeholder` 说明搜索范围。自绘输入框可用 `ui.TextInput(...).Search()` 声明为搜索框，并把搜索框与执行按钮放进 `Role(ui.RoleSearch)` 的地标。
**最小示例**
```go
if ui.SearchField(c, &app.query).Label("Search mail").Changed() {
	app.filter() // 输入即筛选；耗时搜索改用 Submitted 或 c.After
}
```
**限制/常见错误** 控件本身不执行搜索，只提供查询文本与清空/提交语义。

### `ui.Combobox`（combobox）

**用途** 在选项中选定一个 `*string`：显示当前选择的文本框，输入时过滤下方选项（以输入内容开头的排前），单击或按 Down 显示全部，单击显示全部时会选中当前值以便直接输入替换；Up/Down 移动，Enter 或单击选择，Escape 关闭，失焦后文本恢复为当前选择，`Changed` 报告新选择。
**关键 API** `ui.Combobox(c, &app.font, fonts).Label("Font")`、`Placeholder`；无外观版 `ui.ComboboxBase` 用 `Input` 过滤 `Popup` 中的 `Item`，`Chosen` 确认，并可用 `cb.Input.AutoComplete(ui.AutoCompleteList)` 向辅助技术声明建议。
**最小示例** `ui.Combobox(c, &app.font, fonts).Label("Font")`
**限制/常见错误** 值必须是选项之一；文本可任意并需要建议时用 `Autocomplete`，短列表且无需输入时用 `Select`。没有选项的弹出层会隐藏，除非它含可读内容（如"没有匹配项"），此时显示在选项位置。

### `ui.Autocomplete`（autocomplete）

**用途** 编辑 `*string`，下方显示包含输入内容的建议（以输入内容开头的排前）；Up/Down 切换，Enter 或单击采用，未选中建议时 Enter 触发 `Submitted`；输入与采用建议都触发 `Changed`。
**关键 API** `ui.Autocomplete(c, &app.city, cities).Label("City")`。
**最小示例** `ui.Autocomplete(c, &app.city, cities).Label("City")`
**限制/常见错误** 与组合框的关键差别：自动补全保留用户输入的任意内容，值不必来自列表。

### `ui.TokenField`（token-field）

**用途** 多项输入，编辑 `*[]string`（标签、收件人）：每项是带移除按钮的小标签（chip），末尾是输入框；Enter 或逗号添加当前输入，输入框为空时 Backspace 移除最后一项；下方建议排除已添加项，Up/Down 选择，Enter 或单击添加；`Changed` 报告变化。
**关键 API** `ui.TokenField(c, &app.tags, allTags).Label("Tags")`、`Placeholder`、`Query(&app.query)`。`Query` 在字段处理完用户输入后写入文本、添加一项时清空它（输入法组合中的内容等提交后才写入）；在它之后调用 `Changed` 也会报告文本变化。字段按应用给出的顺序原样显示建议、只排除已添加项，匹配由应用负责，因此忽略重音的查找可以成立；写入该字符串会设置输入框文本但不报告变化。
**最小示例**
```go
to := ui.TokenField(c, &app.to, app.found).Label("To").Query(&app.query)
if to.Changed() {
	app.typedAt = c.Now() // 停顿后再查，不要每个字符都发请求
}
```
**限制/常见错误** 停顿后异步查找要在每一帧用 `c.After(wait)` 重新请求停顿结束的那一帧，结果回来后用 `win.Update` 写入；在新建议传入前会一直显示旧建议。

## 就地编辑与查找

### `ui.EditableText`（editable-text）

**用途** 把 `*string` 显示为可就地编辑的文本，像 Finder 的文件名：双击，或文本有焦点时按 Enter，显示文本框并选中扩展名之前的文本；Enter 或失焦保留输入，Escape 恢复原值，`Changed` 报告新值。
**关键 API** `ui.EditableText(c, &app.title).Changed()`。
**最小示例**
```go
if ui.EditableText(c, &app.title).Changed() {
	app.save()
}
```
**限制/常见错误** 在列表/表格行中焦点留在列表上，文本不单独获取焦点：macOS 用 Return、其他平台用 F2 编辑选中行的文本，双击行触发 `Submitted`；选中一行后要等双击判定间隔结束，再单击该行文本才会进入编辑。

### `ui.FindBar`（find-bar）

**用途** `*bool` 为 true 时显示查找栏：查询文本框（打开时获得焦点并选中上次查询）、匹配项数量、上一个/下一个按钮、完成按钮；应用负责统计与显示匹配，`Changed` 报告当前匹配项的切换。
**关键 API** `ui.FindBar(c, &app.finding, &app.query, len(matches), &app.match)`；关闭时返回不显示任何内容的元素。
**最小示例**
```go
if c.Shortcut(ui.Cmd, ui.KeyF) {
	app.finding = true
}
matches := app.find(app.query)
ui.FindBar(c, &app.finding, &app.query, len(matches), &app.match)
if app.finding && len(matches) > 0 {
	app.show(matches[app.match])
}
```
**限制/常见错误** Enter 下一个、Shift+Enter 上一个；macOS 的 Cmd+G/Shift+Cmd+G、其他平台的 F3/Shift+F3 等效，两端循环；Escape 或"完成"关闭。名称与计数用窗口语言，不随应用文案。

## 富文本、代码、差异与 Markdown

### `ui/richtext`（rich-text-editor）

**用途** 编辑带属性的文本：粗斜体、下划线、删除线、行内代码、字体字号颜色、链接、标题/列表/引用/代码块/对齐、行内图像；带文字处理器快捷键、格式菜单、撤销时格式随文本恢复、富格式粘贴、Markdown 导出。只对视口内段落排版与绘制。
**关键 API** `import "github.com/egoist/mygo/ui/richtext"`；`app.doc = richtext.NewDocument(richtext.NewText("Hello"))`；`richtext.Editor(c, app.doc).Fill().Changed()`（`EditorBase` 无外观）。
- `Text` 是值类型：`TextOf`、`Paragraphs`、`Runs`、`AttributesAt`、`ParagraphStyleAt`、`String`/`PlainText`、`Replace`/`ReplaceString`、`Format`、`FormatParagraphs`、`Len`、`Markdown`/`HTML`、`ParseMarkdown`/`ParseHTML`/`ParseRTF`、`MarshalJSON`/`UnmarshalJSON`。
- `Document`：`Formatting()`（`FormatOn`/`FormatOff`/`FormatMixed`）、`ToggleBold`/`ToggleItalic`/`ToggleUnderline`/`ToggleStrikethrough`/`ToggleCode`、`Format(fn)`、`SetLink(url)`、`SetHeading`、`ToggleList`、`ToggleQuote`、`ToggleCodeBlock`、`Indent`、`SetAlign`、`ToggleChecked(i)`、`InsertText`/`Insert`/`InsertAttachment`、`Edit(fn)`（合并为一步撤销）、`Undo`/`Redo`/`BreakUndoCoalescing`、`SetText`、`IdentifyParagraphs`、`VisibleParagraphs()`、`CaretRect()`、`ScrollBy(dy)`、`ScrollToParagraph(i, align)`。
- `EditorOptions`：`Label`、`Placeholder`、`Font`、`ReadOnly`、`NoMarkdown`、`OnLink`、`OnLinkCommand`、`TagStyle`、`Menu`、`OnKey`、`Spans`/`SpansVersion`、`Space`、`Theme`。
**最小示例**
```go
app.doc = richtext.NewDocument(richtext.NewText("Hello"))
if richtext.Editor(c, app.doc).Fill().Changed() {
	app.save(app.doc.Text())
}
```
**限制/常见错误**
- 偏移按 rune 计（换行也算），与 `TextSelection` 一致；属性为零值时采用编辑器的字体字号颜色；`Code`、`Link`、`Tag` 只作用于已有内容，在边缘输入不会并入。
- 段落样式属于段落本身（空段落也记住将输入的样式）；合并/拆分时每段取第一个 rune 原所在段落的样式。列表是扁平的，层级由 `Indent` 表示。
- `Changed` 只报告用户编辑，不报告应用通过 `Document` 做的修改；`Formatting` 的临时输入属性在移动插入符后清除（同 NSTextView 的 Cmd+B）。
- 链接只自动打开 `http`、`https`、`mailto`（`ui.SafeToOpen`），Windows 上其他协议/路径会启动程序，用 `OnLink` 自行接管。
- Markdown 导出保留标题、列表、任务、引用、围栏代码、粗斜体删除线代码链接图像，下划线写成 `<u>`；对齐、字体和颜色没有对应语法会被略去，HTML/RTF 粘贴时字体字号颜色也会略去。
- 从其他 goroutine 改文稿要用 `Window.Update`；`VisibleParagraphs`/`CaretRect` 在编辑器首次显示文稿前返回 nil / `ok == false`。

### `ui/editor`（code-editor）

**用途** 参考 Zed 的原生代码编辑器：语法高亮、多插入符与多选区、软换行、折叠、查找替换、补全、悬停、诊断与内联提示、替换区段、行间块、行样式、片段；只对视口内的行排版，百万行文件与十行文件一样滚动。
**关键 API** `import "github.com/egoist/mygo/ui/editor"`；`buf := editor.NewBuffer(source)`、`buf.SetLanguage(editor.LanguageFor("main.go"))`、`app.code = &editor.State{Buffer: buf, WrapGuides: []int{100}}`、`editor.View(c, app.code)`。
- 缓冲区：`Edit(editor.Edit{Start, End, Text})`、`Transact`（一步撤销）、`EditRemote`、`Set`（可撤销）、`Reset`（不留历史）、`Undo`/`Redo`、`Modified`、`MarkSaved`、`WriteTo`、`Line`、`Slice`、`String`、`Pos`/`Offset`、`LineEnding`/`SetLineEnding`、`Snapshot`、`Version`；偏移按 **UTF-8 字节**计，`Pos` 是行号加行内字节数。
- 状态方法：`Run`、`Bindings`（`{Mods, Key, Command}`，组合键再给 `ThenMods`/`ThenKey`，`NoCommand` 抢走按键）、`CommandNamed`、`Selections`/`SetSelections`/`Select`/`Cursor`/`CursorPos`、`Insert`、`GoTo`、`ScrollTo`、`Find`/`FindMatches`、`Fold`/`Unfold`/`Folded`/`FoldImports`、`Focused`、`VisibleLines`、`ScrollY`/`SetScrollY`、`ViewportHeight`、`LineTop`、`OnInput`、`NoInputMethod`、`CursorShape`。
- 信息与选项：`OnCompletion`+`ShowCompletions`、`OnHover`+`ShowHover`（`HoverDelay` 默认 600 ms）、`SetDiagnostics(buf.Version(), list)`、`SetInlays`、`SetSuggestion`、`OnDefinition`、`ShowPopup`、`Rect`、`SetReplacements`、`SetBlocks`、`SetLineStyles`；`SoftWrap`/`WrapColumn`、`LineNumbers`、`LineBackgrounds`、`TextBackgrounds`、`TabSize`/`Indent`、`Whitespace`、`WrapGuides`、`ReadOnly`、`FitContent`、`KeepScrollbars`/`NoScrollbars`、`Font`/`LineHeight`、`ScrollMargin`、`DiffBase`、`ContextMenu`。
**最小示例**
```go
buf := editor.NewBuffer(source)
buf.SetLanguage(editor.LanguageFor("main.go"))
app.code = &editor.State{Buffer: buf, WrapGuides: []int{100}}
editor.View(c, app.code).Grow(1).FillWidth()
```
**限制/常见错误**
- 多个编辑器与差异两侧可共享同一缓冲区，一处编辑处处可见；撤销撤回缓冲区最后一次编辑，不区分来源。`Snippet` 只认 TextMate/LSP 的 Tab 跳转子集（嵌套占位符、变量、候选项、文本变换会返回错误），补全项的 `Insert` 恒为纯文本。
- `OnInput` 能看到每个输入事件（`InputKeyDown`、`InputText`、编辑命令、指针），返回 true 即取走；但取走按键不会取走其对应文本，字符模式要从 `InputText` 读。输入法组合/提交只交给编辑器。
- `NoInputMethod` 在编辑器获得焦点时禁用系统输入法（Vim 普通模式），浏览器无法禁用；用块代替隐藏源文本时，替换项必须从行首开始、至少覆盖一个换行符、`Text` 为空且 `Content` 为 nil。
- 换行按 Zed 方式处理：带 `\r\n`/`\r` 的文本按 `\n` 行尾保留并记住每行原本行尾，`WriteTo` 原样写回；`LineBackgrounds`/`TextBackgrounds` 固定在原行号上，缓冲区 `Version` 变化后要重新计算。

### `editor.DiffView`（diff-viewer）

**用途** 显示两段文本的差异，可并排或统一列；行内改动的词高亮，未变行折叠，新侧可就地编辑；两侧都是代码编辑器，共享高亮、选区与命令。
**关键 API** `app.diff = &editor.DiffState{Old: before, New: after, Editable: true}`；`editor.DiffView(c, app.diff)`。`Layout`（`DiffSplit`/`DiffUnified`）、`SplitPercent`、`Inline`（`DiffInlineWordAlt` 默认/按词/按字符/关闭）、`Indicators`（`DiffBars`/`DiffClassic`）、`NoBackground`、`NoLineNumbers`、`ContextLines`（默认 3）、`ExpansionLines`（默认 100，Shift 全展开）、`ExpandAll`/`CollapseAll`、`Wrap`、`FitContent`、`Theme`。审查：`GoToHunk`、`HunkActions`、`Revert`/`Accept`、`Hunks`、`Stats`、`LineSelection`/`OnLineSelection`/`SelectedLines`/`SetSelectedLines`、`OnGutterButton`、`Annotations`/`RenderAnnotation`、`ActiveSide`、`Cursor`/`CursorPos`、`CurrentHunk`、`Run`、`Find`、`ScrollY`/`SetScrollY`/`LineTop(side, line)`。补丁：`editor.ParsePatch`、`FilePatch.Apply`、`SetPatch`、`SetPatchFiles(f, old, new)`。
**最小示例**
```go
lang := editor.LanguageFor("main.go")
before, after := editor.NewBuffer(original), editor.NewBuffer(edited)
before.SetLanguage(lang)
after.SetLanguage(lang)
app.diff = &editor.DiffState{Old: before, New: after, Editable: true}
editor.DiffView(c, app.diff).Grow(1).FillWidth()
```
**限制/常见错误**
- 空文本按零行算（同 `git diff`），空的一侧不显示行号、行号栏不能选行。未启用 `Editable` 时逐行连同行尾换行符比较，缺结尾换行会显示"文件末尾没有换行"；启用 `Editable` 时不计换行符，添加结尾换行会新增一个空行。
- 行比较用 git 的 histogram 算法、剩余部分再用 Myers，并按 git 的缩进启发式调整边界；行内高亮把改动行与最相似的行配对，而不是按行序。
- 分栏每侧至少保留行号栏和 12 列文本，视图过窄时暂时均分但仍保存比例；`SplitPercent` 存进设置即可恢复用户上次的分栏比例。
- `SetPatchFiles` 在补丁是二进制或无法从 `old` 得到 `new` 时返回错误且不改变任何内容，此时应退回 `SetPatch`。

### `ui/markdown`（markdown）

**用途** 原生显示 Markdown：CommonMark + GitHub 表格/任务列表/删除线/自动链接 + 脚注与数学公式；针对模型逐 token 输出设计，只解析新增文本，未闭合语法按预期最终格式显示，文本平滑淡入。
**关键 API** `import "github.com/egoist/mygo/ui/markdown"`；`app.md = markdown.Options{}`、`markdown.View(c, app.readme, &app.md)`、`markdown.Stream(c, m.Text, &app.md)`、`markdown.Revealed(elem)`、`Options.Reveal{Unit, CatchUp, Fade}`（`ByWord` 默认、`ByCharacter`、`BySentence`、`ByLine`、`ByParagraph`、`AtOnce`）。
- 链接与图像：`OnLink`、`LinkMenu`、`Anchor`/`AnchorVersion`/`AnchorBlock`、`LinkIcon`、`LinkColor`、`LoadImage`、`FileImages(dir)`；任务与代码：`OnTask(line, checked)`、`CodeBlock`、`Code`、`WrapCode`、`SourceLine()`；数学与样式：`ExtDollarMath`、`Math`、`InlineMath`、`Theme`、`Block`。
- 解析器在 `ui/markdown/parser`：`Parse`、`Append`、`Edit`、`New(ext)`、`SetStreaming`、`HTML()`。
**最小示例**
```go
app.md = markdown.Options{}
if m.Done {
	markdown.View(c, m.Text, &app.md)
} else {
	markdown.Stream(c, m.Text, &app.md) // 完成时在同一位置改用 View，状态延续
}
```
**限制/常见错误**
- `Stream` 只补全末尾未闭合的最外层块或新增行，已完成的块永不改变；`Revealed` 只表示**已收到的内容**显示完，流式时还需自己判断接收是否完成。视图自身不联网，图像在 `LoadImage` 提供前只显示描述文字。
- 原始 HTML 作为文本显示（无法隐藏内容），只有 `<b>`、`<i>`、`<u>`、`<s>`、`<code>`、`<kbd>`、`<mark>`、`<sub>`、`<sup>`、`<br>` 等行内标签生效；HTML 块用 `HTML` 回调。
- 复制同时给纯文本、HTML 与 Markdown，段落间空行、表格单元格用制表符；`ShareSelection` 把选区交给外层 `Selectable` 容器。`Code`/`InlineMath` 在流式期间可能每帧调用，应缓存结果。
- 对 CJK 放宽了强调规则：CJK 字符、全角标点或弯引号旁的 `*` 与 `~~` 可开启/闭合（`_` 与单个 `~` 仍守 CommonMark）。

## 日期时间与颜色

### `ui.Calendar`（calendar）

**用途** 按月显示、选择 `*time.Time` 的日期；单击或方向键选择，Page Up/Page Down 或按钮换月，Home/End 跳到当月首末日；`Changed` 报告新日期，保留原值的时分秒与时区。
**关键 API** `ui.Calendar(c, &app.day).Label("Due date")`；`Browse()` 只浏览不选择（翻页只换月、方向键只移焦点，单击或 Enter/Space 才选中，`Changed` 只报告确认的日期），适合选中后关闭的浮层日历。
**最小示例**
```go
ui.Popover(c, chip, &app.dueOpen, func() {
	if ui.Calendar(c, &task.Due).Browse().Changed() {
		app.dueOpen = false
	}
})
```
**限制/常见错误** 从 `Context.Region` 的每周第一天起显示六周，前后月日子置灰并标记今天；标题、星期名与辅助技术名称取自窗口语言的 CLDR。

### `ui.DateInput`（date-input）

**用途** 按用户地区的数字日期格式显示 `*time.Time`（美国 `1/5/2026`、英国 `05/01/2026`、中国 `2026/1/5`），点击或 Enter/Space 在下方打开日历修改；日历内 Enter 选择、Page Up/Down 换月、Escape 关闭；`Changed` 报告新日期，保留时分秒与时区。
**关键 API** `ui.DateInput(c, &app.birthday)`。
**最小示例**
```go
ui.Row(c).Gap(8).Children(func() {
	ui.DateInput(c, &app.meeting).Label("Date")
	ui.TimeInput(c, &app.meeting).Label("Time")
})
```
**限制/常见错误** 只编辑日期部分；与 `TimeInput` 指向同一个 `time.Time` 时才构成完整的日期时间选择。

### `ui.TimeInput`（time-input）

**用途** 按用户地区的小时制（`15:04` 或 `3:04 PM`）显示 `*time.Time` 的时刻：小时、分钟、上下午各自可获得焦点（类似 AppKit 日期选择器的分段），Up/Down 循环增减当前段，键入数字设置当前段、小时输完自动转到分钟，输入上下午名称首字母切换，Left/Right 在分段间移动。
**关键 API** `ui.TimeInput(c, &app.alarm).Label("Alarm")`。
**最小示例** `ui.TimeInput(c, &app.alarm).Label("Alarm")`
**限制/常见错误** 上下午的名称与位置随窗口语言（中文显示"下午3:04"）；`Changed` 只改时间，日期与时区保持原值。

### `ui.ColorWell` / `ui.ColorPicker`（color-picker）

**用途** `ui.ColorWell` 是绑定 `*ui.Color` 的色块控件，单击或 Enter/Space 在下方打开选择器，Escape 或点外部关闭，`Changed` 报告颜色变化。`ui.ColorPicker` 是选择器本身（可放进自定义面板）：方形区域横向调饱和度、纵向调亮度，下方色相滑块，另有不透明度滑块、可编辑十六进制值与常见色块；方形区域获得焦点后 Up/Down 调亮度、Left/Right 调饱和度。
**关键 API** `ui.ColorWell(c, &app.tint)`、`ui.ColorPicker(c, &app.background)`。
**最小示例**
```go
ui.Field(c, "Tint", func() {
	ui.ColorWell(c, &app.tint)
})
```
**限制/常见错误** 两者只处理颜色值，不与主题绑定；辅助技术看到名称为十六进制的颜色块，以及"Hex"文本框和各色块按钮。

## 速查

| 控件 | 绑定 | 变化报告 | 记忆点 |
|---|---|---|---|
| `ui.TextInput` / `ui.TextArea` | `*string` | `Changed`；`Submitted`（单行 Enter） | `ReplaceText` 单步撤销；换 `Key` 重开会话 |
| `ui.TextInputBuffer` / `ui.TextAreaBuffer` | `*ui.TextBuffer` | `Changed` | 大文稿不整串复制；`Set` 清历史 |
| `TextInputClient` | 应用自管 | 无（由 `HandleInput` 自管） | 范围按 UTF-16；回调在 UI 线程 |
| `ShapeText` / `ShapeRichText` | `TextLayout` | — | 几何查询可跨 goroutine，绘制在 UI 帧 |
| `ui.NumberInput` | `*float64` | `Changed` | 失焦按步长小数位舍入显示 |
| `ui.DecimalInput` | `*string` | `Changed` | 无舍入，超步长精度标 `Error` |
| `ui.SearchField` | `*string` | `Changed`；`Submitted` | Escape/清空按钮清空 |
| `ui.Combobox` | `*string`（限选项） | `Changed` | 输入过滤；失焦恢复当前选择 |
| `ui.Autocomplete` | `*string`（任意） | `Changed`；`Submitted` | 保留用户输入，不强制选项 |
| `ui.TokenField` | `*[]string` | `Changed` | Enter/逗号添加，空输入 Backspace 删末项 |
| `ui.EditableText` | `*string` | `Changed` | 双击/Enter 编辑；Escape 还原 |
| `ui.FindBar` | `*bool`、`*string`、`int`、`*int` | `Changed`（切换匹配） | 应用自己统计与显示匹配 |
| `ui/richtext` `Document` / `Editor` | `*richtext.Document` | `Changed`（仅用户编辑） | `Text` 值类型；`Edit` 合并一步撤销 |
| `editor.State` / `editor.View` | `*editor.Buffer` | 由命令/缓冲区查询 | 偏移按 UTF-8 字节；`Snapshot` 免锁 |
| `editor.DiffState` / `editor.DiffView` | 旧/新缓冲区 | — | git histogram；`Editable` 编辑新侧 |
| `ui/markdown` `View` / `Stream` | 源字符串 | `Revealed` | 增量解析；未闭合语法按最终格式显示 |
| `ui.Calendar` | `*time.Time` | `Changed` | 保留时分秒与时区；`Browse` 只浏览 |
| `ui.DateInput` | `*time.Time` | `Changed` | 地区数字格式，下方日历 |
| `ui.TimeInput` | `*time.Time` | `Changed` | 分段编辑，12/24 小时随地区 |
| `ui.ColorWell` / `ui.ColorPicker` | `*ui.Color` | `Changed` | 色块打开选择器；方向键微调 |
