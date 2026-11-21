# UI 对话框、状态与媒体

覆盖 `ui` 包中 14 个页面：对话框与浮层（dialog、alert-dialog、popover、tooltip、toast）、状态指示（progress、spinner、meter、badge）和媒体（chart、image、video、icon、avatar）。

共同点：对话框与浮层由 `*bool` 控制显隐，关闭时该值被置为 false，并把焦点还给原来的元素；状态组件只读取数据、不处理输入；媒体对象（位图、SVG、视频播放器）只在初始化时创建一次，视图每帧复用，不要在视图函数里创建。选型速记：需要用户回答用警告对话框，自定义内容用对话框，锚定面板用浮层，简短说明用工具提示，操作结果与撤销用 Toast；确定进度用进度条，不确定用旋转指示器，非进度数值用量度计，计数标记用徽标。

## 源页面

| 页面 | 在线地址 | 页面 | 在线地址 |
| --- | --- | --- | --- |
| 对话框 | [dialog](https://mygo.run/zh/docs/ui/dialog) | 警告对话框 | [alert-dialog](https://mygo.run/zh/docs/ui/alert-dialog) |
| 浮层 | [popover](https://mygo.run/zh/docs/ui/popover) | 工具提示 | [tooltip](https://mygo.run/zh/docs/ui/tooltip) |
| Toast | [toast](https://mygo.run/zh/docs/ui/toast) | 进度条 | [progress](https://mygo.run/zh/docs/ui/progress) |
| 旋转指示器 | [spinner](https://mygo.run/zh/docs/ui/spinner) | 量度计 | [meter](https://mygo.run/zh/docs/ui/meter) |
| 徽标 | [badge](https://mygo.run/zh/docs/ui/badge) | 图表 | [chart](https://mygo.run/zh/docs/ui/chart) |
| 图像 | [image](https://mygo.run/zh/docs/ui/image) | 视频 | [video](https://mygo.run/zh/docs/ui/video) |
| 图标 | [icon](https://mygo.run/zh/docs/ui/icon) | 头像 | [avatar](https://mygo.run/zh/docs/ui/avatar) |

## 对话框与浮层

### 对话框 `ui.Modal`
**用途**：`*bool` 为 true 时，在变暗的窗口上方显示由函数构建的对话框；单击它外面或按 Escape 会把它设为 false。
**关键 API**：`ui.Modal(c, &open, func(){...})`、`Dismissed()`；`ui.DialogBase(c, &open, func(backdrop, panel ui.Element))` 是无外观版本，由函数为背板和面板设置样式。
**示例**：
```go
ui.Modal(c, &app.renaming, func() {
	ui.Text(c, "Rename").Bold()
	if ui.TextInput(c, &app.name).AutoFocus().Submitted() {
		app.renaming = false
	}
	ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
		if ui.Button(c, "Cancel").Clicked() {
			app.renaming = false
		}
		if ui.PrimaryButton(c, "Rename").Clicked() {
			app.rename()
			app.renaming = false
		}
	})
})
```
**限制**：打开后焦点限制在内部，默认第一个可聚焦元素，有 `AutoFocus` 时优先；Tab 循环，背后内容不可交互，在对话框外构建的窗口快捷键暂停，关闭后焦点恢复到原来的元素。`Dismissed` 在因 Escape 或外部单击关闭的那一帧报告；必须保持打开时把 `*bool` 设回 true。对话框绘制在所属窗口内并被裁剪，浮出窗口中的对话框宽度不会超出；要超出需另开以浮出窗口为父的模态窗口。

### 警告对话框 `ui.AlertDialog`
**用途**：确认重要操作，含标题、消息和多个按钮；最后一个按钮是默认按钮，用强调色、获得焦点、按 Enter 触发。
**关键 API**：`ui.AlertDialog(c, &open, title, message, buttons...) int`；单击按钮的那一帧返回其索引，按 Escape 关闭返回 `ui.AlertDismissed`，其余返回 -1。
**示例**：
```go
if ui.Button(c, "Delete").Clicked() {
	app.asking = true
}
switch ui.AlertDialog(c, &app.asking, "Delete “Notes”?", "You can't undo this.", "Cancel", "Delete") {
case 1:
	app.delete()
}
```
**限制**：单击警告框外部无效。必须得到回答时，在返回 `ui.AlertDismissed` 的那一帧把 `*bool` 设回 true。按钮通常留在同一行，窗口太窄时换到多行，长于警告框的标签以省略号结尾。用户可以撤销的选择应立即执行，并在 Toast 中提供撤销。

### 浮层 `ui.Popover`
**用途**：`*bool` 为 true 时，在锚点元素下方显示由函数构建的面板；在面板外按下指针或按 Escape 会把它设为 false。
**关键 API**：`ui.Popover(c, anchor, &open, func(){...})`、`Dismissed()`；`ui.PopoverBase(c, anchor, &open, func(panel ui.Element))` 是无外观版本，面板可用 `AttachTo` 放到锚点旁边。
**示例**：
```go
more := ui.Button(c, "More ▾")
if more.Clicked() {
	app.menu = !app.menu
}
ui.Popover(c, more, &app.menu, func() {
	if ui.Button(c, "Rename").Clicked() {
		app.menu, app.renaming = false, true
	}
})
```
**限制**：关闭浮层的指针按下会继续传给下面的内容，因此单击另一个按钮会关闭浮层并触发该按钮；单击锚点也会关闭浮层。下方空间不足时面板显示在上方，窗口滚动时它跟随锚点。Tab 顺序中面板元素排在锚点之后，关闭时若焦点在面板内则还给之前的元素；`Dismissed` 报告因外部按下或 Escape 关闭的那一帧。系统菜单用菜单按钮，元素菜单用 `PopoverMenu`，提示用工具提示。

### 工具提示 `Element.Tooltip`
**用途**：指针停留在元素上或键盘焦点移入后显示提示，并向辅助技术描述该元素。
**关键 API**：`Element.Tooltip(s)`；`ui.TooltipBase(c, el, func(tip ui.Element))` 自定义外观，显示时返回提示元素、否则返回 nil；`Element.Description(s)`；主题 `TooltipDelay` 设置等待时间（0 为默认的 0.6 秒，负值立即显示）。
**示例**：
```go
ui.Button(c, "").Label("Share").Tooltip("Share with others").Children(func() {
	ui.Icon(c, share)
})
```
**限制**：提示始终在元素上方居中，上方空间不足时改在下方；一个提示消失后的 0.4 秒内，下一个提示立即显示；在元素上按下或单击、或按 Escape 后隐藏，直到指针移出再进入或焦点移走再回来。嵌套时只显示最内层元素的提示，一次只显示一个。提示默认采用主题的 `Inverse`/`InverseText` 颜色。`TooltipBase` 的内容不会被朗读，需要给元素设 `Description`。

### Toast `c.Toast`
**用途**：在窗口底部附近显示一条消息几秒钟，用于说明用户刚才操作的结果。
**关键 API**：`c.Toast(msg)`；`c.ToastAction(title, action, fn)`（带按钮，显示更久）；`c.AddToast(ui.Toast{ID, Title, Description, Type, Timeout, Action, OnAction, OnClose, Data})` 返回 ID；`c.CloseToast(id)`，`c.CloseToast("")` 关闭全部；自绘用 `ui.ToastViewportBase(c, func(viewport ui.Element, toasts []ui.Toast))` 加 `ui.ToastBase(c, t)`（`Root`、`ActionButton()`、`CloseButton()`、`Left()`）。
**示例**：
```go
if ui.Button(c, "Save").Clicked() {
	app.save()
	c.Toast("Saved")
}
app.trash(note)
c.ToastAction("Note deleted", "Undo", func() { app.restore(note) })
```
**限制**：显示最新三条，较旧的消息隐藏等待但计时继续；相同消息再次显示。`Timeout` 为零时显示 4 秒，为负时直到关闭。指针停在 toast 上或焦点在其中时暂停计时，Escape 关闭焦点所在的 toast。窗口在后台时计时继续，回到前台时已到期的直接关闭且不显示。自绘时在视图中调用一次 `ToastViewportBase`，没有 toast 时也保留视口以播放退出过渡。

## 状态指示

### 进度条 `ui.Progress`
**用途**：用 0 到 1 之间的值表示完成比例；负值表示进度不确定，显示持续移动的色块。
**关键 API**：`ui.Progress(c, value)`、`Label(s)`、`Reverse()`、`TextColor(fill)`、`Background(track)`。
**示例**：
```go
ui.Progress(c, float64(app.done)/float64(app.total)).Label("Uploading")
ui.Progress(c, -1).Label("Connecting")
```
**限制**：横向铺满所在的列，在行中要给它设 `Width`。默认用主题强调色填充、`Border` 色轨道，`TextColor` 与 `Background` 可改填充色与轨道色，不确定状态的移动色块也用填充色。`Reverse` 让它从右侧填充。不确定时只在可见时播放动画，只重绘上一帧。更省空间的旋转指示器用 Spinner，非进度的数值用 Meter。

### 旋转指示器 `ui.Spinner`
**用途**：用旋转动画表示进度不确定的任务正在进行，类似 AppKit 的旋转进度指示器。
**关键 API**：`ui.Spinner(c)`、`Label(s)`、`Size(w, h)`。
**示例**：
```go
ui.Row(c).Gap(8).Children(func() {
	ui.Spinner(c).Label("Loading")
	ui.Text(c, "Loading messages…")
})
```
**限制**：辐条使用主题的次要文字颜色，整体高度略大于主题字号；空视图中可用 `Size(32, 32)`。即使系统设置了减少动效，它仍会继续转动。只在可见时转动，开销仅来自绘制，辐条每转一圈更新十二次。

### 量度计 `ui.Meter`
**用途**：条形指示器，显示给定上下限之间的数值，并按阈值切换颜色。
**关键 API**：`ui.Meter(c, value, lo, hi, *ui.MeterLevels)`（levels 传 nil 表示不设阈值）；`ui.MeterLevels{Warning, Critical, NormalColor, WarningColor, CriticalColor}`；`Label(s)`、`Background(track)`。
**示例**：
```go
ui.Meter(c, app.battery, 0, 100, &ui.MeterLevels{Warning: 20, Critical: 10}).Label("Battery")
```
**限制**：默认使用主题 `Success`，达到 `Warning` 用 `Warning` 色，达到 `Critical` 用 `Danger` 色；`Critical` 低于 `Warning` 时表示数值越低越需警示，如电量。三种颜色字段为零值时沿用主题颜色；没有阈值时全程用 `NormalColor`。自定义颜色不改变形状、数值范围和辅助技术读数。进行中的工作用进度条。

### 徽标 `ui.Badge`
**用途**：用胶囊形背景显示短文本，如侧边栏条目旁的未读数。
**关键 API**：`ui.Badge(c, text)`、`Background(color)`、`TextColor(color)`。
**示例**：
```go
ui.SidebarItem(c, "inbox", inboxIcon, "Inbox").Children(func() {
	ui.Badge(c, fmt.Sprint(app.unread))
})
```
**限制**：辅助技术在读取它所在的条目时一并读出它的文本；需要强调时用 `Background` 与 `TextColor` 设置强调色。

## 媒体

### 图表 `ui/chart`
**用途**：不预设外观的图表包，提供标度、带坐标轴的绘图区、数据标记、折线/面积/饼图路径、数据堆叠与选择；颜色、尺寸和字体由应用决定。
**关键 API**：
- 标度：`chart.NewLinear(lo, hi).Nice(n)`、`Ticks(n)`、`At(v)`、`Span(a, b)`、`Invert(f)`、`Clamp()`、`Format`；`chart.Extent(values...)`；`chart.NewBand(labels...).Padding(inner, outer)` 的 `Band(i)`、`At(i)`、`Width()`、`Index(f)`、`Ticks()`；`chart.NewLog`、`chart.NewSymLog`、`chart.NewTime`；刻度为 `chart.Tick{At, Value, Label}`。
- 绘图区：`chart.Plot(c, chart.Axes{Left, Right, Top, Bottom, Label}, func(area ui.Element){...})`；位置是 0 到 1 的比例（x 从左、y 从底向上），用 `chart.X(r, f)`、`chart.Y(r, f)` 换算成坐标。
- 标记：`chart.Rect(c, x, y)`、`chart.Point(c, x, y, w, h)`、`chart.RuleX(c, x, width)`、`chart.RuleY(c, y, height)`、`chart.Anchor(c, x, y, h, v, fn)`；标记是普通元素，可 `Background`、`Tooltip`、`Clicked`、`Transition`。
- 路径：`chart.Line(r, xs, ys, curve)`、`chart.Area(r, xs, lo, hi, curve)` 交给 `area.Draw`；曲线有 `CurveLinear`、`CurveMonotone`、`CurveStepBefore`、`CurveStepAfter`、`CurveStep`；实时数据用 `chart.MinMax`、`chart.LineInto`、`Painter.StrokePolyline`。
- 堆叠与饼图：`chart.Stack(series, chart.StackStandard|StackNormalized|StackCenter)` 返回 `Segment{Lo, Hi}`；`chart.Pie(values, pad)`、`chart.Sector(cx, cy, inner, outer, slice)`、`chart.SliceAt`、`chart.Polar`。
- 选择：`chart.Select(c, area, &selected, xs, describe)` 让绘图区成为键盘 Tab 停留点，方向键、Home、End 移动选择，Escape 清除；`chart.Nearest` 为自建选择找最近点。
**示例**：
```go
x := chart.NewBand(app.months...)
_, hi := chart.Extent(app.sales...)
y := chart.NewLinear(0, hi).Nice(4)
chart.Plot(c, chart.Axes{Left: y.Ticks(4), Bottom: x.Ticks()}, func(area ui.Element) {
	for i, v := range app.sales {
		chart.Rect(c, x.Band(i), y.Span(0, v)).Background(t.Accent).
			Tooltip(fmt.Sprintf("%s: %.0f", app.months[i], v))
	}
}).Label("Sales by month").Height(200)
```
**限制**：定义域必须有限；对数标度不能接触或跨过零，零附近或带符号数据用 SymLog，NaN 样本按缺失值断线。标记按百分比定位，构建时不等待布局，并随绘图区缩放。该包不提供缩放平移、图例和标签避让，图例需自行用一行元素构建。辅助技术上用 `Label` 命名绘图区、`Description` 说明内容，需要读取的标记也要命名。

### 图像 `ui.Image`
**用途**：显示位图，或按自身颜色显示 SVG；默认按自身 DIP 尺寸显示，指定其他尺寸时缩放以适应。
**关键 API**：`ui.Image(c, img)`；位图 `ui.DecodeBitmap(data)`（PNG、JPEG、GIF 首帧、WebP、BMP，按 EXIF 方向摆正）与 `ui.NewBitmap(image.Image)`；SVG `ui.ParseSVG`、`ui.MustParseSVG`；未知格式用 `ui.DecodeImage`，尺寸用 `ui.ImageSize`。方法：`Size(w, h)`、`Fit(ui.Contain|Cover|FillBox|ScaleDown|NaturalSize)`、`Grayscale()`、`Radius(px)`、`AspectRatio(r)`、`ObjectPosition(x, y)`、`Width(w)`；位图另有 `Pixelated()`、`Update(img)`、`UpdateYUV(*ui.YUVImage)`、`ReleasePixels()`。
**示例**：
```go
//go:embed logo.png
var logoPNG []byte

var logo, _ = ui.DecodeBitmap(logoPNG) // once, outside the view

ui.Image(c, logo).Size(64, 64).Fit(ui.Contain).Radius(12)
```
**限制**：位图与解码只在视图外做一次；`Update` 复用同尺寸纹理，在 goroutine 中解码的帧通过 `Window.Update` 交给 `Update`；`ReleasePixels` 释放 CPU 像素副本，之后需要时重新生成，必须在主线程调用。`UpdateYUV` 直接显示解码器的 Y′CbCr 帧：`ui.YUVImage{Format, Width, Height, Planes, Strides, Matrix}`，`Format` 为 `ui.NV12` 或 `ui.I420`，`Matrix` 为 `ui.BT601`、`ui.BT709`、`ui.BT2020`。`Fit` 默认 `Contain`，`ObjectPosition` 默认 0.5 居中。SVG 不绘制文本、嵌入图像、图案、标记和滤镜，导出前把文本转成路径；要按文本颜色显示 SVG 用 `ui.Icon`。

### 视频 `ui.Video`
**用途**：播放视频：系统媒体框架打开的文件或流，或 Go 解码器提供的帧（动画 GIF/PNG 背景动画）；布局、适配和圆角方式与图像相同。
**关键 API**：`ui.Video(c, player)`；系统来源 `ui/video.Open(path)`、`video.OpenURL(url)`、`video.OpenData(bytes)`、`video.ErrUnsupported`；Go 解码器 `ui.NewGIFDecoder(data)`、`ui.NewAPNGDecoder(data)`、`ui.NewVideoPlayer(decoder)`；播放器 `Play()`、`Pause()`、`Playing()`、`Ended()`、`SetLoop(bool)`、`SetMuted(bool)`、`SetVolume`、`SetRate`、`Seek(d)`、`Position()`、`Duration()`、`Ready()`、`Buffering()`、`Err()`、`Close()`；自定义解码器实现 `ui.VideoDecoder`（`Next() (ui.VideoFrame, error)`、`Rewind()`），可选 `Duration()` 与 `ui.VideoSeeker`。
**示例**：
```go
app.trailer, err = video.OpenData(trailerMP4) // 或 video.Open(path)、video.OpenURL(url)
if err != nil {
	log.Fatal(err)
}
ui.Video(c, app.trailer).Fill().Fit(ui.Contain)
```
**限制**：播放器只创建一次，不要在视图里创建；同一时间只在一个 `Video` 中显示它，方法在主线程调用。新的播放器在 `Video` 显示它时立即播放；系统视频在首帧到来前 `Ready` 为 false、`Size` 为 0×0，流等待数据时 `Buffering` 为 true。H.264 与 AAC 的 MP4/MOV 在各平台可用，其他格式取决于系统编解码器；Linux 需要 GStreamer 插件，缺少解码器时出现 `Err`。静音视频只在可见时播放；系统要求减少动效时它静止不动（解码器视频停在第一帧）。有声音的视频在后台继续播放。GIF/APNG 没有声音，GIF 是否循环由 `SetLoop` 决定，不采用文件中的循环次数。

### 图标 `ui.Icon`
**用途**：把 SVG 显示为图标，颜色取自文本颜色并随祖先继承，高度等于字体大小并保持宽高比。
**关键 API**：`ui.Icon(c, svg)`、`FontSize(n)`、`TextColor(c)`、`Size(w, h)`、`Rotate(angle)`、`Label(s)`；`ui.ParseSVG`、`ui.MustParseSVG`；`Painter.Icon` 在 `Draw` 回调中绘制。
**示例**：
```go
ui.Row(c).Gap(6).Children(func() {
	ui.Icon(c, save)
	ui.Text(c, "Save")
})
```
**限制**：图标颜色不理会 SVG 文件自身的颜色，也不会为填满列宽而拉伸；在按钮中放在标签之前，只显示图标的按钮应给按钮设 `Label`。`Rotate` 可用于旋转指示或表示展开：`spin.Rotate(spin.Loop("spin", time.Second, ui.Linear) * 360)`。

### 头像 `ui.Avatar`
**用途**：为人物或事物创建头像：有图像时裁成圆形；没有图像时，在由名称决定的背景色上显示名称的首字母。
**关键 API**：`ui.Avatar(c, name, photo)`，`photo` 为 `*ui.Bitmap` 或 `nil`。
**示例**：
```go
ui.Row(c).Gap(8).Children(func() {
	ui.Avatar(c, user.Name, user.Photo) // a *ui.Bitmap, or nil
	ui.Text(c, user.Name)
})
```
**限制**：头像高度略大于主题字号的两倍；同一个名字的颜色始终一致。辅助技术看到一张以名字命名的图像。

## 速查

| 组件 | 一句话 | 关键 API |
| --- | --- | --- |
| 对话框 `ui.Modal` | 变暗背板上显示自定义内容 | `Modal(c,&open,fn)`、`Dismissed()` |
| 警告对话框 `ui.AlertDialog` | 标题、消息和按钮组成的确认框 | `AlertDialog(c,&open,title,msg,...)`、`AlertDismissed` |
| 浮层 `ui.Popover` | 锚点旁显示面板 | `Popover(c,anchor,&open,fn)`、`PopoverBase` |
| 工具提示 | 悬停或聚焦时显示说明 | `Element.Tooltip(s)`、`TooltipBase`、`Description` |
| Toast | 底部短暂消息，可带操作按钮 | `c.Toast`、`c.ToastAction`、`c.AddToast`、`c.CloseToast` |
| 进度条 `ui.Progress` | 0–1 进度，负值表示不确定 | `Progress(c,v)`、`Reverse`、`TextColor` |
| 旋转指示器 `ui.Spinner` | 不确定进度的旋转动画 | `Spinner(c)`、`Size`、`Label` |
| 量度计 `ui.Meter` | 区间数值与阈值颜色 | `Meter(c,v,lo,hi,levels)`、`MeterLevels` |
| 徽标 `ui.Badge` | 胶囊形短文本标记 | `Badge(c,text)`、`Background`、`TextColor` |
| 图表 `ui/chart` | 标度、绘图区、标记、路径与选择 | `Plot`、`NewLinear`/`NewBand`、`Rect`、`Select` |
| 图像 `ui.Image` | 位图或 SVG，可裁剪、缩放、圆角 | `Image(c,img)`、`DecodeBitmap`、`Fit`、`ObjectPosition` |
| 视频 `ui.Video` | 系统视频或解码器帧的播放器 | `Video(c,player)`、`video.Open`、`Play`/`Pause`/`Seek` |
| 图标 `ui.Icon` | 随文本颜色和字号的 SVG | `Icon(c,svg)`、`FontSize`、`TextColor`、`Rotate` |
| 头像 `ui.Avatar` | 圆形头像或首字母占位 | `Avatar(c,name,photo)` |
