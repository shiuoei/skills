# MyGo 语言、数据、音频与原生模块

本页提炼 MyGo 的国际化消息、剪贴板与拖放、音频、原生模块与 C 库绑定的 API、最小示例与限制。

## 源页面

- [语言](https://mygo.run/zh/docs/i18n)(i18n.md)
- [剪贴板与拖放数据](https://mygo.run/zh/docs/data-transfer)(data-transfer.md)
- [音频](https://mygo.run/zh/docs/audio)(audio.md)
- [原生模块](https://mygo.run/zh/docs/native-modules)(native-modules.md)

## 消息目录、ICU 语法与生成代码

**用途**：让应用文本、MyGo 控件与页面按用户语言显示，并编写复数、选择与格式化消息。

**关键 API**
- 配置 `i18n: { "dir": "locales", "source": "en" }`：`dir` 每种语言一个 BCP 47 标记命名的 JSON（`en.json`、`zh-Hant.json`），`source` 为编写语言、含全部消息，其他语言回退到它。
- ARB 对象：`@@locale` 标语言，键即函数名（字母开头，后接字母、数字和 `_`），`@key` 描述消息，`placeholders.<name>.type` 取 `String`、`int`、`double`、`num`、`DateTime`、`Object`。
- ICU 消息：插值 `Hello, {name}!`；复数 `{count, plural, one {# file} other {# files}}`（`=0` 为精确值，`offset:1` 让类别与 `#` 按 `n - 1` 计）；序数 `{place, selectordinal, one {#st} other {#th}}`；枚举 `{gender, select, female {She} other {They}}`；数字 `{price, number}`、`{n, number, ::.00}`、`{price, number, ::currency/EUR}`；日期 `{day, date, short}`、`{day, date, ::yMMMd}`。
- 生成 `catalog_gen.go`：`locales.FilesSelected(3)`、`locales.In("de").FilesSelected(3)`；有前端时生成 `src/locales.ts`：`messages.filesSelected(3)`、`messagesIn("de")`。

**最小示例**
```json
{
  "@@locale": "en",
  "greeting": "Hello, {name}!",
  "filesSelected": "{count, plural, =0 {No files selected} one {# file selected} other {# files selected}}",
  "@filesSelected": { "placeholders": { "count": { "type": "int" } } }
}
```

**限制/常见错误**
- 每个 `plural`、`selectordinal`、`select` 都必需 `other`，分支可嵌套；复数类别用该语言的 CLDR 集合（中文、日语只有 `other`）；撇号转义花括号与 `#`（`'{'` 表示 `{`）。
- 键是扁平的，点号保留给 `mygo.<key>`；无参数消息格式化一次后缓存；导入消息目录包即注册应用语言，Go 代码不显示消息时也要 `import _ "example.com/app/locales"`，否则 `mygo dev`、`mygo build` 会警告；生成文件应随项目提交。
- 数字骨架中 Go 和页面读取 `percent`、`integer`、小数精度、`group-off`、`currency/EUR`。

## 应用语言与区域格式

**用途**：读取或覆盖应用语言，按用户区域格式输出数字与日期，并在切换后刷新界面。

**关键 API**
- `mygo.App.Locale()` 为最终显示语言，`App.PreferredLanguages()` 为用户偏好，`App.Languages()` 为应用包含的语言（无消息目录时用 `App.SetLanguages` 声明）。
- `App.SetLocale("de")` 切换并持久化，`App.SetLocale("")` 恢复跟随用户，`App.ChosenLocale()` 返回所选值（跟随系统为 `""`），`App.OnLocaleChanged(func(string))` 监听变化；`MYGO_LOCALE` 覆盖本次运行语言，`MYGO_DIRECTION`（`rtl`/`ltr`）覆盖布局方向。
- `App.Region()` 返回区域格式、周首日与 12/24 小时制，`App.OnRegionChanged` 监听变化，`Region.FormatLocale(language)` 把语言与区域合为格式化标记；`ui.Context.Locale()`、`ui.Context.Region()` 返回当前帧的语言与区域。

**最小示例**
```go
mygo.App.OnLocaleChanged(func(string) {
	win.SetTitle(locales.Title())
	mygo.App.SetMenu(app.menu()) // 用新语言的标签重建
})
formats := mygo.App.Region().FormatLocale(mygo.App.Locale())
```

**限制/常见错误**
- 原生 UI 视图每帧重建、按角色标注的菜单项自动换标签；窗口标题与自定义菜单需在回调中重设。
- 应用只显示一种语言：按用户偏好选应用包含语言中最匹配者（`zh-TW` → `zh-Hant`），否则用源语言；只随附英语的应用对中文读者也显示英语。
- `SetLocale` 保存选择，`MYGO_LOCALE` 只影响本次运行；Linux 从 `LANGUAGE`、`LC_ALL`、`LC_MESSAGES`、`LANG` 读取用户语言，只读一次。

## 检查、伪区域与 MyGo 字符串

**用途**：构建前发现消息问题、预览布局，并让 MyGo 自带文本跟随语言。

**关键 API**
- `mygo i18n check [dir]`、`mygo generate`、`mygo build` 都检查目录；`mygo dev -lang de` 用指定语言运行且不改保存的语言。
- 伪区域：`en-XA` 给源语言字母加变音符号、加长三分之一并加方括号；`ar-XB` 从右到左显示并把应用切成 RTL。
- 错误（阻止构建）：文件名或 `@@locale` 不规范、消息无法解析、源语言没有的键、参数不符、生成函数重名；警告（不阻止）：缺翻译、省略参数、复数分支缺失或多余、占位符未使用。
- MyGo 文本用消息目录中的 `mygo.<key>` 覆盖，如 `"mygo.undo"`、`"mygo.menuQuitApp"`；键见 [internal/l10n/en.json](https://github.com/egoist/mygo/blob/main/internal/l10n/en.json)。

**最小示例**
```json
{ "mygo.undo": "Rückgängig", "mygo.menuQuitApp": "{app} beenden" }
```

**限制/常见错误**
- 随附语言但未导入翻译包（如 `import _ "github.com/egoist/mygo/locales/zh-Hans"`）时启动会提示并回退英语；原生 UI 检查器保持英语。
- 缺翻译回退到源语言，缺失的复数分支用 `other` 代替；Windows 安装程序完成页用 `mygo.desktopShortcut`。

## 剪贴板与拖放：数据模型

**用途**：一次写入内容，每项提供多个备选格式，剪贴板与拖放共用同一模型。

**关键 API**
- `transfer.Format`（自定义格式用 MIME 名与明确编码）；`transfer.New(items...)`、`transfer.NewItem(opts...)`、`transfer.Bytes(format, b)`、`transfer.Lazy(format, fn)`。
- 常用格式名：`transfer.Text`、`transfer.HTML`、`transfer.PNG`、`transfer.URIList`、`transfer.FileList`；构造辅助 `TextData`、`URLData`、`FileData`。
- `mygo.Clipboard.Write(data, mygo.ClipboardOptions{Concealed, OnRelease})`；拖放源用 `row.DragData(data)`（独立提供方缓存）。

**最小示例**
```go
data := transfer.New(transfer.NewItem(
	transfer.Bytes(transfer.Text, []byte(title)),
	transfer.Bytes(transfer.HTML, []byte("<b>"+html.EscapeString(title)+"</b>")),
	transfer.Lazy(noteFormat, func() ([]byte, error) { return json.Marshal(note) }),
))
mygo.Clipboard.Write(data, mygo.ClipboardOptions{})
```

**限制/常见错误**
- `Bytes` 复制内容；`Data` 不含 Go 指针或进程内值；空字节是有效值；文件以绝对路径表示并编码为转义文件 URL，不读取也不删除文件，每个路径一个项。
- `Concealed` 标记机密内容：macOS 带 `org.nspasteboard.ConcealedType`、Windows 设 `ExcludeClipboardContentFromMonitorProcessing`、Linux 提供 `x-kde-passwordManagerHint=secret`；浏览器没有该标记。

## 读取、观察与持久化

**用途**：按接收方偏好读取格式、监听变化，并延迟生成或持久化剪贴板内容。

**关键 API**
- `mygo.Clipboard.Formats()` 返回可移植格式名且不请求表示形式；`transfer.PreferredFormat(formats, want...)`、`Offer.Preferred(want...)` 按偏好协商。
- `Clipboard.ReadFormat(fmt)` 读单一格式，`Clipboard.Read(want...)` 一次读可接受的备选（无参时复制所有公布格式）；`Data.Preferred(...)`、`Data.Read(fmt)`、`Data.Items()`；旧接口 `ReadText`/`WriteText`、`ReadHTML`/`WriteHTML`、`ReadImage`/`WriteImage`、`ReadFiles`/`WriteFiles`、`Clear` 保留。
- `Clipboard.OnChanged(fn)` 返回停止函数；每次写入或拖放得到新的 `Data.Snapshot`，惰性提供方按快照至多运行一次，成功、错误与恢复的 panic（`transfer.ErrProviderPanic`）都缓存。
- `Clipboard.Flush()` 先生成应用拥有的每种表示再交给原生存储；`ClipboardOptions.OnRelease` 在主线程运行一次，用于释放提供方资源。

**最小示例**
```go
if f, ok := transfer.PreferredFormat(mygo.Clipboard.Formats(), transfer.HTML, transfer.Text); ok {
	b, err := mygo.Clipboard.ReadFormat(f) // 按约定编码解码，处理 err
	_ = b
}
win.OnClosed(mygo.Clipboard.OnChanged(func() { /* 只读取需要的格式 */ }))
```

**限制/常见错误**
- 空剪贴板返回空数据；显式请求无匹配格式返回 `transfer.ErrFormat`；`ReadFormat` 合并项（文本换行连接、URI/文件列表用 CRLF、其他取第一个匹配项）。
- 读取结果已生成、不依赖原生提供方；所有权变化时返回 `mygo.ErrClipboardChanged`，从提供方访问剪贴板返回 `mygo.ErrClipboardReentrant`；回调不报告初始内容，多次通知可能合并；浏览器不提供系统剪贴板观察。
- 提供方在主线程执行且必须迅速返回，不要等待另一个主线程操作或递归读取；资源保留到 `OnRelease`；Linux 无剪贴板管理器时 `Flush` 返回 `mygo.ErrClipboardPersistence`；其他应用的每个表示形式限 64 MiB。

## 音频：声音、流与格式

**用途**：播放音效、流式音乐与应用合成声音，在 Go 中混音后送系统输出。

**关键 API**
- `audio.DecodeSound([]byte) (*audio.Sound, error)` 一次解码；`(*Sound).Play()` 每次新建播放器、可叠加，`(*Sound).NewPlayer()` 返回已暂停的播放器。
- `audio.OpenStream(io.ReadSeeker) (*audio.Player, error)` 边播边解码（提前约一秒）；`audio.NewPlayer(decoder)` 接受任意 `Decoder`；`audio.NewSound(samples)` 用一次性算好的采样。
- 格式按副作用导入：`_ "github.com/egoist/mygo/audio/vorbis"`、`audio/mp3`、`audio/flac`、`audio/aac`、`audio/mp4`；WAV 内置，`audio.RegisterFormat` 注册自定义格式。

**最小示例**
```go
hoe, err := audio.DecodeSound(data)
hoe.NewPlayer().SetPan(-0.4).SetVolume(0.8).Play()
f, _ := sounds.Open("music.ogg")
music, err := audio.OpenStream(f.(io.ReadSeeker))
music.SetLoop(true).Play()
```

**限制/常见错误**
- 流的播放器用完后必须关闭（关闭会关闭文件）；声音的播放器无需关闭。
- 循环音乐优先 Ogg Vorbis：MP3 首尾的编码器填充会带来几毫秒静音，除非用 `SetLoopRange` 排除；AAC 解码使用 LGPL 许可的第三方库，导入 `audio/aac` 须遵守其条款，HE-AAC（AAC+）不支持。
- 无可用的音频输出设备时播放器仍按时走但不发声，原因见 `audio.Err`。

## 播放器控制、合成与输出

**用途**：播放中调整音量、声像、速度与循环，选择输出延迟并接入系统媒体键。

**关键 API**
- `SetVolume(v)`、`SetPan(p)`（-1 左到 1 右，语义同 Web Audio `StereoPannerNode`）、`SetSpeed(s)`、`SetPreservesPitch(true)`、`SetLoop(true)`、`SetLoopRange(start, end)`；设置方法返回播放器，可链式调用。
- 播放控制：`Play`、`Pause`、`Stop`（回开头）、`Seek`、`Position`、`Duration`、`Playing`、`OnEnd`；`audio.SetVolume` 设置所有声音；`audio.SetLatency` 接受 `audio.LatencyInteractive` 或 `audio.LatencyPlayback`。
- 合成器实现 `Decoder`：`SampleRate`、`Channels`、`Len`（`-1` 表示无尽头）、`SetPosition`、`Read([]float32)`；`mygo.NowPlaying.Control(player)` 返回释放函数，配合 `mygo.NowPlaying.Set(mygo.NowPlayingInfo{...})` 与 `Sync()` 接入系统媒体键。

**最小示例**
```go
music.SetPreservesPitch(true).SetSpeed(1.5).Play() // 语音保持音高
release := mygo.NowPlaying.Control(music)
defer release()
```

**限制/常见错误**
- `SetSpeed` 通过重采样改变音高；需要不变音高时用 `SetPreservesPitch`（WSOLA，语音适用约 0.5×–4×），鼓声等突然起音会略模糊。
- 音量、暂停、继续在 5 ms 内渐变以免咔嗒声；`OnEnd` 在独立 goroutine 回调，循环播放不触发；`Position` 表示实际被听到的位置。
- 输出为 48 kHz 立体声、提前约 40 ms 混音，全部停止五秒后关闭设备；`LatencyPlayback` 约 250 ms，须在播放前设置；浏览器需用户先交互才能播放。

## 原生模块：接口与生成代码

**用途**：用 Swift、Kotlin、Rust 或 Zig 实现 Go 接口，无 cgo、无手写胶水（实验性）。

**关键 API**
- Go 接口前加 `//mygo:native` 成为入口点，包中得到 `NewT() (T, error)`；实现侧需提供同名工厂：Swift/Zig 为 `makeAnalyzer`，Rust 为 `make_analyzer`，可抛错；`NativeError.unsupported` 对应 `errors.ErrUnsupported`。
- `mygo native new [-lang swift,kotlin,rust,zig|none] <module name>` 在 `modules/<name>`（或给定路径）生成各语言目录；`mygo native add <language> [module dir]` 追加实现语言。
- 生成物：Go 侧 `<module>_native.go`，实现侧 `<Module>.mygo.swift`、`<Module>.mygo.kt`、`mygo.rs`、`mygo.zig`；必须提交，且不要编辑。
- 平台选择：iOS 用 Swift、Android 用 Kotlin；macOS 有 Swift 用 Swift，否则 Rust，再否则 Zig；无实现时构造函数返回 `errors.ErrUnsupported`。

**最小示例**
```go
//mygo:native
type Analyzer interface {
	// DetectLanguage returns the BCP 47 code of the text's language.
	DetectLanguage(text string) (string, error)
}
```

**限制/常见错误**
- 模块名必须是合法 Go 包名；模块是普通 Go 包，`mygo dev`、`mygo build` 自动发现并构建依赖中的模块。
- 入口点值像打开的文件一样只创建一次并长期持有；接口变化后所有实现必须同步，否则编译器指出缺少方法。
- 视图、窗口、纹理及其他模块接口只有 Swift（Android 上 Kotlin）能实现；`mygo native add` 会指出不能实现的方法。

## 原生模块：方法、类型与错误

**用途**：掌握跨语言签名映射与错误传递。

**关键 API**
- 方法可接收任意多个受支持类型与回调，返回无值、一个值、错误、值加错误、`iter.Seq[T]`、`iter.Seq2[T, error]` 或 `io.Closer`。
- 类型映射：`bool`/`int`/`int64`/`uint8`…、`float32`/`float64`、`string`、`[]byte`、`time.Time`/`time.Duration`、`[]T`/`map[string]T`、`*T`、包内结构体与枚举、模块接口；Swift 对应 `Bool`/`Int`/`Int64`、`Float`/`Double`、`String`、`Data`、`Date`/`Duration`、`[T]`/`[String: T]`、`T?`、`struct`、`enum`、`any P`。
- Swift 签名：`error` → `throws`，`(T, error)` → `throws -> T`，`context.Context` → `async`（在 `Task` 中运行），`iter.Seq2[T, error]` → `AsyncThrowingStream<T, Error>`，`io.Closer` → `any Closer`；`//mygo:mainthread` → `@MainActor`。
- 错误：`native.NewError("undetermined")` 声明包级错误，Swift 抛出 `<Module>Error.undetermined`，Go 用 `errors.Is` 比较；其他错误以 `*native.Error` 到达，带 domain、code 与 `localizedDescription`。

**最小示例**
```go
var ErrUndetermined = native.NewError("undetermined")
// Swift: throw LanguageError.undetermined
if errors.Is(err, language.ErrUndetermined) { /* … */ }
```

**限制/常见错误**
- 结构体与切片跨界复制；`[]byte` 参数只在调用期间借用，要保留就用 `Data(data)` 复制。
- 不支持通道、`any`、键非 string 的 map、泛型类型与其他包的类型，会在声明行报错；这类数据改用包内结构体。
- 方法名转为小驼峰（`DetectLanguage` → `detectLanguage`）；实现崩溃如 `fatalError`、越界会结束应用，Rust panic 在边界被捕获并转为错误。

## 原生模块：对象、视图与大数据

**用途**：返回有生命周期的对象、启动可停止的任务、嵌入平台视图，并零复制传递大块数据。

**关键 API**
- 方法返回包内另一接口即对象：Go 通过接口调用实现返回的实例；嵌入 `io.Closer` 或 `Close() error` 时显式关闭，之后调用返回 `native.ErrClosed`；返回 `io.Closer` 的方法表示订阅，关闭即停止并释放回调（Swift 用 `NativeCloser { ... }`）。
- 嵌入 `native.View` 的接口由平台视图实现：`ui.NativeView(c, v).Fill()` 显示，之后绘制的内容覆盖其上；窗口与锚点参数 `*mygo.Window` + `ui.Rect` 映射为 `NSWindow` + `CGRect`。
- `*native.Buffer`：`native.NewBuffer(n)` 创建、`Bytes()` 读写，双方共享内存；`*native.Texture`：`native.NewTexture()`，实现向其中绘制，原生 UI 用 `ui.Image(c, tex)` 显示，全程不经过 CPU 复制。

**最小示例**
```go
func (a *app) view(c *ui.Context) {
	ui.NativeView(c, a.player).Fill().Children(func() {
		ui.Text(c, captionAt(a.now)).Absolute().Left(24).Bottom(24)
	})
}
```

**限制/常见错误**
- 返回到 Go 的对象可再传回实现或另一个模块，得到同一实例而非副本；Go 值持有对实现对象的强引用。
- 原生视图在主线程运行，尺寸与裁剪由 MyGo 设置，不要自行移动、调整大小或移除；应在 `App.WhenReady` 后创建，回调中改状态需调用 `win.Invalidate()`。
- 流中产出的缓冲区在循环请求下一个前有效，要保留就复制；纹理每帧应从缓冲区池取，不要反复写同一个缓冲区；带上下文的方法取消时会取消相应任务并让 Go 收到 `ctx.Err()`。

## 原生模块：构建、测试与发布

**用途**：构建实现、离线检查接口调用，并向其他机器发布预编译产物。

**关键 API**
- `mygo native build [flags] [dir]` 按平台构建，产物写入 `.mygo/native/release`；`-target` 取 `GOOS-GOARCH`、`all` 或 `host`，另有 `-lang`、`-release`、`-out`、`-upload`、`-force`。
- `mygo native check [-lang <language>] [module dir]` 生成、构建并 vet，问题按 `file:line: message` 输出，有错误时退出码 1；`mygo native call [-timeout] [dir] <Interface>.<Method> '{"arg": …}'` 自行创建上下文，结果按 JSON 输出，流中每个值一行。
- `mygo native build -upload` 上传到当前仓库的 `native-<name>` 发布版本，并把名称、URL、SHA-256 与源码哈希写入 `mygo-plugin.json` 的 `libraries`；依赖与权限写在 `swift/Package.swift`、`rust/Cargo.toml`、`zig/build.zig` 与 `swift/Info.plist`、`swift/Entitlements.plist`。

**最小示例**
```sh
mygo native build -target host -lang rust ./modules/language && go test ./modules/language
mygo native call modules/language Analyzer.DetectLanguage '{"text": "Bonjour"}'
```

**限制/常见错误**
- 各语言需要对应工具链：`swift/` 需 Xcode（iOS 必须 Xcode）或 Command Line Tools 与 Swift 6，`kotlin/` 需 Android SDK 与 JDK，`rust/` 需 cargo，`zig/` 需 Zig 0.16；`mygo doctor` 可检查。
- 跨系统无法构建的目标会跳过并在末尾列出；Swift 库需在 Mac 上构建，Rust 需对应 target 与链接器，Zig 可交叉编译；已发布构建与源码哈希匹配时直接复用。
- 手机从源码随应用构建，从不使用已发布构建；`go test` 不在手机上运行，用依赖接口的假实现测试调用方代码。

## 直接绑定 C 库（ffi）

**用途**：绕过模块生成，直接调用现成 C 库或 Rust/Zig/Swift 导出的函数。

**关键 API**
- 库源码放在 Go 包旁（Rust `cdylib`、Zig 包、macOS Swift 动态库），并在 `mygo-plugin.json` 的 `libraries[].build` 指定 `{"cargo"|"zig"|"swift": "./dir"}`，可带 `args` 与 `macOS` 最低版本。
- `ffi.New(name, manifest)` 得到库；`lib.Bind(&fns)` 绑定结构体函数字段（字段名转 snake case，或用 `ffi:"adder_version"`，`,optional` 表示缺函数时为 nil）；另有 `BindFunc`、`Symbol`、`ffi.GoString`。
- 查找顺序：`MYGO_NATIVE_<NAME>` → 应用资源/`Frameworks` → 可执行文件旁 → 非打包程序的 `.mygo/native/<goos>-<goarch>` → 用户缓存（校验 SHA-256）；打包应用从不下载库。

**最小示例**
```go
var lib = ffi.New("adder", manifest)
var fns struct{ Add func(a, b int32) int32 }

func Add(a, b int32) (int32, error) {
	if err := lib.Bind(&fns); err != nil {
		return 0, err
	}
	return fns.Add(a, b), nil
}
```

**限制/常见错误**
- 和 cgo 一样，库不得在调用结束后保留 Go 内存；返回的字符串会被复制且不释放，应返回静态字符串，或返回指针并提供释放函数。
- `ffi` 只在 macOS、Linux、Windows 的 amd64 与 arm64 工作，其他平台 `Load` 返回包装 `errors.ErrUnsupported` 的错误；未知 `tag` 或缺少函数会使 `Bind` 返回错误。
- `mygo-plugin.json` 的 `files` 按 `GOOS-GOARCH`（或 `darwin-universal`）索引，含 `name`、`url`、`sha256`、`sourceHash`；`mygo native build` 写入 `release` 与 `files`。

## 速查

| 页面/API | 一句话 |
|---|---|
| i18n.md | 消息目录、ICU 消息、语言与区域格式、页面消息模块。 |
| `locales.<Message>()` | 生成的消息函数，按应用语言返回文本。 |
| `App.SetLocale` / `App.OnLocaleChanged` | 切换并持久化语言、在变化时刷新一次性文本。 |
| `i18n.FormatNumber` / `FormatDate` | 按 CLDR 格式化数字、货币、日期与范围。 |
| `i18n.FormatRelativeTime` | 按 `Intl.RelativeTimeFormat` 输出相对时间。 |
| `mygo i18n check` | 检查消息目录，错误时退出码 1。 |
| `mygo.<key>` | 覆盖或翻译 MyGo 自带字符串。 |
| data-transfer.md | 剪贴板与拖放共用的 `transfer.Data` 模型与格式协商。 |
| `Clipboard.Read` / `Clipboard.ReadFormat` | 读取可接受格式的备选或单一格式。 |
| `Clipboard.OnChanged` | 监听剪贴板变化并返回停止函数。 |
| `Clipboard.Flush` | 生成应用拥有的表示并持久化。 |
| audio.md | 音效、流式音乐、合成声音与输出延迟。 |
| `audio.DecodeSound` / `OpenStream` | 一次解码音效 / 边播边解码流。 |
| `Player.SetPreservesPitch` | 变速时保持音高（WSOLA）。 |
| `audio.SetLatency` | 选择 40 ms 交互或 250 ms 播放缓冲。 |
| native-modules.md | 用四种语言实现 Go 接口，实验性。 |
| `//mygo:native` / `mygo native new` | 声明入口点并创建模块。 |
| `mygo native check` / `call` | 离线检查实现与调用单个方法。 |
| `mygo native build -upload` | 构建并发布预编译产物。 |
| `ffi.New` / `lib.Bind` | 直接绑定 C 库函数。 |
