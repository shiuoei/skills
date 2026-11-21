# Web 前端、绑定与插件

定位：网页窗口如何加载前端、页面与 Go 之间如何互相调用，以及把能力以插件形式提供给页面的方式。适用于 Web 前端项目；纯原生 UI 应用不需要这一层（视图直接调用 Go）。

## 源页面

- 前端 — <https://mygo.run/zh/docs/frontend>
- 从前端调用 Go（绑定） — <https://mygo.run/zh/docs/bindings>
- 插件 — <https://mygo.run/zh/docs/plugins>
- 13 个插件页：automation <https://mygo.run/zh/docs/plugins/automation>、fetch <https://mygo.run/zh/docs/plugins/fetch>、game <https://mygo.run/zh/docs/plugins/game>、glass <https://mygo.run/zh/docs/plugins/glass>、loading <https://mygo.run/zh/docs/plugins/loading>、permissionflow <https://mygo.run/zh/docs/plugins/permissionflow>、render <https://mygo.run/zh/docs/plugins/render>、sfsymbols <https://mygo.run/zh/docs/plugins/sfsymbols>、sqlite <https://mygo.run/zh/docs/plugins/sqlite>、terminal <https://mygo.run/zh/docs/plugins/terminal>、updater <https://mygo.run/zh/docs/plugins/updater>、watch <https://mygo.run/zh/docs/plugins/watch>、websocket <https://mygo.run/zh/docs/plugins/websocket>

## 加载页面与前端配置

**用途**：窗口用不带协议的 URL（`/`、`/settings?tab=general`）加载前端；同一份代码在开发时指向开发服务器，构建后指向内嵌文件。

```ts
export default defineConfig({
  devUrl: "http://localhost:5173",
  devCommand: "bun run dev:web",
  buildCommand: "bun run build:web",
  frontendDist: "dist",
});
```

- `mygo dev` 期间解析到 `devUrl`（由 `devCommand` 启动），热重载在应用里同样有效；构建后的应用解析到 `mygo://localhost/`，由 `mygo build` 嵌入的 `frontendDist` 提供。没有 `devUrl` 时 `mygo dev` 从磁盘提供 `frontendDist`。不带扩展名又匹配不到文件的路径返回 `index.html`，客户端路由可用。
- 安全上下文的源：macOS/Linux 是 `mygo://localhost`，Windows 是 `http://mygo.localhost`（WebView2 把自定义协议放在 `http://<scheme>.localhost` 下）。
- 页面也可加载别的内容：`LoadURL("https://example.com")`、`LoadHTML(html, baseURL)`、`LoadFile("docs/index.html")`（先相对工作目录，再相对可执行文件目录与 macOS 的 Resources）。不用 CLI 时用 `//go:embed all:dist` + `fs.Sub` + `mygo.SetFrontend(web)`。

**限制/常见错误**：`localStorage`、IndexedDB 属于各自的源，`mygo dev` 下（开发服务器的源）存的数据与构建后的应用不是同一份。重要数据放 Go（如 `App.Path(mygo.PathUserData)` 下的文件），再通过绑定方法交给页面。不属于应用自己的页面不能调用 Go。

## 自定义协议

**用途**：用 `http.Handler` 提供运行时生成或读取的内容（缩略图、用户文件、插件 UI），页面像普通 Web URL 一样 fetch、`<img>`、ES 模块加载。

```go
mygo.Protocol.Handle("thumbs", mygo.FileServer(os.DirFS(cache)))
mygo.Protocol.HandleFunc("api", func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"path": r.URL.Path})
})
```

- **关键规则**：在创建使用这些协议的窗口之前注册；`mygo.FileServer` 支持内容类型与范围请求并像前端一样回退到 `index.html`；处理 `mygo` 协议会替换前端。每个协议是自己的源，跨源请求需要 CORS——对应用自己的页面响应带 CORS 头且 MyGo 应答预检，其他站点没有，除非 `PageOptions.TrustedOrigins` 信任或处理器自设头。Windows 上访问用 `http://<scheme>.localhost/`，跨主机用 `http://<host>.<scheme>.localhost/`（主机名须为 DNS 名、小写）。

**限制/常见错误**：

- Windows 上处理器写入达到 1 MB 后 WebView2 才开始接收；WebView2 一次只读一个自定义协议响应、读完才交付页面，所以 SSE、长轮询这类保持连接的响应会卡住并阻塞其他响应——改用事件（`Emit`/`Broadcast`）。`<video>`/`<audio>` 的 `Range: bytes=X-` 一次最多 4 MB（206 部分响应）。
- Linux 上 WebKitGTK 媒体播放器无法播放自定义协议中的音频视频（`MEDIA_ERR_SRC_NOT_SUPPORTED`），改用 `127.0.0.1:0` 的环回 HTTP 服务器提供媒体。这样的服务器机器上任何程序都能访问：只提供它们有权读的内容，或在路径里放应用密钥。

## 信任：哪些页面能做什么

**用途**：决定页面能否调用 Go、控制窗口、获得权限、读取自定义协议响应、获知拖放与文稿路径。

**受信任的页面**：应用前端、自定义协议页面、`file:` 与 `about:` 页面（`LoadHTML` 未给基准 URL 时为 `about:blank`）、`PageOptions.TrustedOrigins` 列出的源、开发构建中由 localhost 提供的页面。信任取决于页面 URL 并随导航变化，窗口与 Web 视图一致。

**限制/常见错误**：

- 绝不要在受信任页面里通过自定义协议提供第三方或用户 HTML，也不要用未指定基准 URL 的 `LoadHTML` 加载这类内容，否则插件 UI 或用户 HTML 预览会获得全部能力。改用不受信任页面或从窗口不信任的源加载（`https`、`data:`）。
- `PageOptions.Untrusted: true` 明确标记不受信任：不信任任何源（`TrustedOrigins` 也不例外）；绑定方法与 `window.mygo.window` 调用被拒；拖动区域无效；页面消息超过 4 KiB 在解码前丢弃、其余在主线程应答；`Broadcast` 与窗口的 `Emit` 到不了它，要用 `EmitPage`；没有 `SetPermissionHandler` 就没有权限，没有 `SetWindowOpenHandler` 就打不开窗口/浏览器，没有 `OnWillDownload` 什么也下载不了，没有 `OnWillNavigate` 只能停在自身源。插件 UI 就是这样与应用通信，处理器用 `mygo.CallerPage(r.Context())` 判断来源页面。

## mygo-runtime 包

**用途**：MyGo 把 `window.mygo` 注入每个页面；`mygo-runtime` npm 包提供类型与导入方式（如 `import { currentWindow, isMyGo, runtime } from "mygo-runtime"`），生成的客户端建立在它之上。

| 导出 | 作用 |
|---|---|
| `call`, `on`, `once`, `event<T>` | 无类型调用与事件（不用生成的客户端时） |
| `Channel` | Go 方法的值流 |
| `isCallError` | 一次拒绝是否来自 Go 的 error |
| `isMyGo()`、`runtime()` | 是否在 MyGo 窗口里；`platform`（`darwin`/`linux`/`win32`）、`windowId`、`version`，MyGo 之外抛错 |
| `currentWindow` | `minimize`、`maximize`、`unmaximize`、`toggleMaximize`、`isMaximized`、`toggleFullScreen`、`close`、`setTitle` |
| `onFileDrop` | 拖放到窗口的文件及其路径 `{paths, x, y}` |
| `preferences`、`onPreferences` | 桌面强调色与辅助功能设置；MyGo 之外抛错 |
| `windowActive`、`onWindowActive` | 窗口是否被绘制为活动窗口 |
| `locale`、`direction`、`onLocaleChange` | 应用显示的语言、`ltr`/`rtl` 及其变化 |

`mygo-runtime/i18n` 格式化 ICU MessageFormat；`mygo generate` 生成的 `locales.ts` 建立在它之上。普通浏览器中打开时没有运行时：调用被拒，`isMyGo()` 为 false。

## 自定义标题栏与桌面外观

**用途**：无边框或隐藏标题栏的窗口由页面自绘标题栏，并跟随桌面外观与辅助功能设置。

```go
mygo.NewWindow(mygo.WindowOptions{
	URL: "/", TitleBarStyle: mygo.TitleBarHidden, TitleBarHeight: 40,
	Transparent: true, Vibrancy: mygo.VibrancySidebar, // Windows 11 上是 Mica
})
```

```css
.titlebar {
  --app-region: drag;
  height: var(--mygo-titlebar-height, 40px);
  padding-left: var(--mygo-titlebar-inset-left, 0px);
  padding-right: var(--mygo-titlebar-inset-right, 0px);
}
.titlebar .search { --app-region: no-drag; }
```

- `--app-region: drag` 标记拖动区域，属性会被继承，`no-drag` 把按钮、输入框等重新排除；区域内交互元素仍可点击，双击等价双击标题栏。
- 从第一次绘制起可读的 CSS 变量：`--mygo-titlebar-height`、`--mygo-titlebar-inset-left/right`、`--mygo-titlebar-controls-bottom`（全屏时 `0px`）；`--mygo-accent`（`#rrggbb`，桌面无强调色时未设置）、`--mygo-text-scale`（`1` 为通常大小）。
- 深色模式：`prefers-color-scheme` + `color-scheme: light dark`，窗口设 `BackgroundColor: "light-dark(#f5f5f7, #1e1e1e)"` 在页面绘制前显示；`mygo.Theme.SetSource(mygo.ThemeDark)` 强制外观。非活动窗口的 `<html>` 带 `data-mygo-window-inactive`；`preferences()` 返回 `accent`、`reduceMotion`、`reduceTransparency`、`highContrast`、`textScale`，Go 侧用 `mygo.Theme.Preferences()`。

**限制/常见错误**：Wayland 下 GTK 会自己装饰无边框窗口（不含标题栏，页面外缘可调大小）；X11 等环境下页面最外侧 5 px 充当边框，那里收不到点击。Linux 上无边框窗口仍保留 `App.SetMenu`/`Window.SetMenu` 的菜单栏，位于页面之上；要自绘菜单，给它空菜单 `win.SetMenu(mygo.NewMenu(nil))`，改用 `Menu.PopupAt`。

## 绑定服务：页面调用 Go

**用途**：把 Go 值以服务形式暴露给页面，方法随调用返回 JSON 值；`mygo generate` 生成类型安全的 TypeScript 客户端。

```go
type Notes struct {
	mu    sync.Mutex
	items []Note
}

// Add creates a note and returns it.
func (n *Notes) Add(title string) (Note, error) {
	if title == "" {
		return Note{}, errors.New("a note needs a title")
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	note := Note{ID: int64(len(n.items) + 1), Title: title}
	n.items = append(n.items, note)
	return note, nil
}

func main() { mygo.Bind(&Notes{}) } // Notes.add / Notes.list 在页面可调用
```

- 服务以类型命名（`Notes`），导出方法变成 `Notes.add`、`Notes.list`；`mygo.BindAs("Store", v)` 显式命名。名称重复或方法用了 JSON 无法编码的类型（通道、函数）时 `Bind` panic，让错误在启动时暴露。
- 方法可接受任意数量 JSON 可编码参数（含可变参数）、接受 `*mygo.Channel[T]` 流式发值、返回无值 / 一个值 / 一个 error / 值加 error。每次调用在独立 goroutine 运行，可阻塞；共享状态要加锁；panic 被捕获并记录后按 error 报告。
- 第一个参数是 `context.Context` 时，它会在发起调用的页面导航离开或窗口关闭时取消；`mygo.CallerWindow(ctx)`、`mygo.CallerPage(ctx)` 返回调用方。上下文不是 TypeScript 参数。

## 生成的客户端与类型

**用途**：让两端都受编译器检查；`mygo generate` 把客户端写到配置的 `bindings`（新项目 `src/mygo.ts`），`mygo dev` 与 `mygo build` 会自动重新生成。

```ts
// Code generated by `mygo generate`. DO NOT EDIT.
import { call } from "mygo-runtime";

export const Notes = {
  /** Add creates a note and returns it. */
  add(title: string): Promise<Note> {
    return call("Notes.Add", title);
  },
} as const;
```

- 方法名变小驼峰（`GetUserByID` → `getUserByID`），Go 文档注释变 TSDoc。两个方法/事件映射到同一属性名时生成失败并指出两处声明。
- 生成时 `mygo generate` 会构建应用并在 `MYGO_GENERATE` 下运行它：`App.Run` 写出客户端后返回，不开窗口。因此 `main` 要一直运行到 `App.Run`，不能在那时做的（如迁移数据库）放到 `App.WhenReady`。

| Go | TypeScript |
|---|---|
| `bool` / 数字 / `string` | `boolean` / `number` / `string` |
| `[]byte`；`time.Time`；`time.Duration` | `string`（base64）；`string`（RFC 3339）；`number`（纳秒） |
| `[]T`、`[N]T`；`map[string]T` | `T[]`（nil 切片到达为 `[]`）；`Record<string, T>`，数字键为 `Record<number, T>` |
| 结构体；`*T` | 同名 interface；`T \| null` |
| `any`、其他 interface、带 `MarshalJSON` 的类型 | `unknown` |
| 带 `MarshalText` 的类型 | `string` |
| 带常量的命名 string/number | 其值的联合类型 |

结构体字段按 `json` 标签命名，没有标签保留 Go 名称——要给字段加 `json:"title"`。`omitempty`/`omitzero` 变可选（`body?: string`），`-` 省略。JavaScript 数字在 2⁵³ 内精确，更大的 ID 用 `json:"id,string"`。嵌入的 `map[string]T` 或 `jsontext.Value`（`json:",embed"`）添加顶层成员。

**错误**：方法返回非 nil error 时 promise 以 `CallError` 拒绝，`err.message` 是 error 文本，`err.method` 是失败方法名（`"Notes.Add"`）。用 `isCallError(err)` 区分 Go 的 error 与其他异常。

## 事件与通道

**事件**（Go → 页面）：在包级别用载荷类型声明，`mygo generate` 才能找到。

```go
var Progress = mygo.NewEvent[ExportProgress]("progress")

Progress.Emit(win, ExportProgress{Done: 3, Total: 10})   // 一个窗口
Progress.Broadcast(ExportProgress{Done: 10, Total: 10})  // 每个窗口
```

页面 DOM 就绪之前发出的事件会在就绪后送达；窗口关闭后 `Emit` 返回 error，可用来结束发送循环。事件名必须唯一，以 `mygo:` 开头的是保留名。页面侧从 `events.progress.on(({ done, total }) => …)` 订阅（事件名 `notes:changed` → `events.notesChanged`），返回的 `off()` 取消订阅。

**通道**（一次调用内的值流）：方法接受 `*mygo.Channel[T]` 参数并用 `Send` 发值；页面创建 `Channel` 并在该位置传入。

```go
func (Shell) Tail(ctx context.Context, command string, lines *mygo.Channel[string]) error {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	out, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		return err
	}
	for sc := bufio.NewScanner(out); sc.Scan(); {
		if err := lines.Send(sc.Text()); err != nil {
			return err // the page stopped listening
		}
	}
	return cmd.Wait()
}
```

```ts
const lines = new Channel<string>();
const done = Shell.tail("ping -c 3 example.com", lines);
for await (const line of lines) output.append(line + "\n");
await done;
```

- 数据按顺序送达，且在本次调用的 promise 完成/失败之前全部送达；方法返回或 `Close` 时关闭。页面 `close()` 或 `break` 后 `Send` 返回 `mygo.ErrChannelClosed`，方法上下文被取消。页面未读数据超过 1 MiB 时 `Send` 等待。通道只服务一次调用、只能作参数。
- 选择：事件给所有监听器、页面存在期间都可收；通道只传一次调用的结果并随调用关闭。不用生成的客户端时用 `call("Notes.Add", …)`、`on("progress", …)` 或 `window.mygo.call` / `mygo.channel`；未传入的参数在 Go 中取零值。

## 谁能调用与从 Go 到页面

**谁能调用**：只有应用自己的页面——前端页面、自定义协议页面、`file:`/`about:` 页面、开发构建中 localhost 提供的页面。显示其他站点的窗口会被阻止，除非 `PageOptions.TrustedOrigins` 列出其源（`"*"` 信任所有来源，包括导航到的任何站点，不要用于显示无法控制的内容）。页面中的 iframe 永远不能调用 Go，只有窗口顶层页面可以。

**从 Go 到页面**：要在页面运行代码（自动化、测试）用 `Page.Eval`，返回表达式值、等待 promise 并按 JSON 解码；`mygo.EvalAs[T]` 解码为 Go 类型。

```go
title, err := mygo.EvalAs[string](win.Page(), "document.title")
count, err := mygo.EvalAs[int](win.Page(), "const items = document.querySelectorAll('li'); return items.length")
```

语句作为异步函数体运行，`return` 产生值；抛异常的脚本返回 `*mygo.EvalError`；页面导航或窗口关闭会返回错误。

## 插件机制

**用途**：插件分两部分——Go 服务与封装服务调用的 JavaScript 包。应用用 `mygo.Use` 启用 Go 部分（须在 `App.Run` 之前），前端从 npm 引入 JS 部分。

```go
mygo.Use(fetch.Plugin, websocket.Plugin)
```

```sh
bun add @mygo-plugins/fetch @mygo-plugins/websocket
```

- 官方插件的 `Plugin` 变量带默认选项，`New` 接受选项（如 fetch/websocket 的 `http.Client` 与决定页面能请求哪些 URL 的 `Allow`），各插件页逐一列出。只用 JavaScript 包而不启用 Go 部分会报错并提示调用 `mygo.Use`；只有应用受信任的页面能调用插件。
- **编写插件**：Go 部分用 `mygo.Plugin{Name, Service, Setup}`，`Service` 的导出方法按 `plugin:<name>` 绑定（规则同 `mygo.Bind`），且不进生成的客户端；JS 包用 `mygo-runtime` 的 `call` 调用，`mygo-runtime` 列为对等依赖（同时作为开发依赖）。名称被占用或方法用了无法传到 JavaScript 的类型时 `mygo.Use` panic。

```go
package clock

var Plugin = mygo.Plugin{Name: "clock", Service: service{}}

type service struct{}

func (service) Now() time.Time { return time.Now() }
```

```ts
import { call } from "mygo-runtime";

export const now = async () => new Date(await call<string>("plugin:clock.Now"));
```

## 官方插件

fetch、websocket、sqlite、watch 是「JS 包 + `mygo.Use`」（sqlite 也可纯 Go）；updater 是纯 Go 的 Sparkle 式更新窗口（另有 `native.Plugin`）；terminal、glass、game、loading、sfsymbols、render、permissionflow 是纯 Go 的原生 UI 能力；automation 是纯 Go 的可选能力，不含在发布构建中。

### fetch

`@mygo-plugins/fetch` + `mygo.Use(fetch.Plugin)`：由 Go 发起 HTTP，页面得到标准 `Response`；`fetch.New(fetch.Options{Client, Allow})` 自定义客户端与允许的 URL。

```ts
const res = await fetch("https://api.example.com/items", {
  method: "POST",
  headers: { authorization: `Bearer ${token}`, "content-type": "application/json" },
  body: JSON.stringify({ name: "Ada" }),
});
if (!res.ok) throw new Error(`HTTP ${res.status}`);
```

限制：无 CORS（除 `Set-Cookie` 外每个响应头可读）；`mode`/`credentials`/`cache`/`referrerPolicy`/`integrity`/`keepalive`/`priority` 被忽略；Cookie 属 Go 的 `http.Client`（除非有 `Jar`）；HTTP 错误码不使请求失败，检查 `res.ok`；网络失败以 `TypeError` 拒绝；正文是 `ReadableStream`，未读超过 1 MiB 时 Go 端等待；`signal` 在 Go 中取消；`redirect: "manual"` 返回响应本身、`"error"` 遇重定向拒绝、`"follow"` 最多 20 次；非 http(s) URL 交给 webview 的 `fetch`；不要把插件的 `fetch` 赋给 `globalThis.fetch`；`Allow` 只检查初始请求，不检查重定向。

### websocket

`@mygo-plugins/websocket` + `mygo.Use(websocket.Plugin)`：Go 建立连接的 `WebSocket`，握手可带任意请求头；`websocket.New(websocket.Options{Client, Allow, MaxMessageSize, CloseTimeout})`。示例：`new WebSocket("wss://example.com/live", ["v1"], { headers: { authorization: `Bearer ${token}` } })`。
限制：`ws:`/`wss:`（`http:`/`https:` 自动转换），`headers` 是第三个参数；`MaxMessageSize` 默认 64 MiB，更大的消息使连接失败；页面未处理消息超过 1 MiB 时留在 Go 中；`close(code, reason)` 的 code 限 1000 或 3000–4999、reason ≤123 字节，`CloseTimeout` 默认 5 秒后以 1006 断开；不协商任何扩展（`extensions` 恒为空）；页面导航或窗口关闭时以 1001 关闭；`Client.Timeout` 与 `CheckRedirect` 被忽略；`Allow` 接收每个连接握手。

### sqlite

`@mygo-plugins/sqlite`（网页）或直接 `sqlite.Open`（Go/原生 UI）。SQLite 是原生模块（桌面 Zig、iOS Swift、Android Kotlin），构建不需要 cgo/C 编译器/Zig。网页：`open(name, {readOnly})`、`db.execute(sql, params?)` → `{changes, lastInsertId}`、`db.query(sql, params?)` → `{columns, rows}`、`db.transaction([{sql, params}])`、`db.close()`；`sqlite.New(sqlite.Options{Directory, BusyTimeout})`；Go：`sqlite.Open(ctx, "app.db", sqlite.OpenOptions{})` + `Execute`/`Query`/`Transaction`/`Close`。

```ts
const db = await open("app.db");
await db.execute("CREATE TABLE IF NOT EXISTS notes (id INTEGER PRIMARY KEY, title TEXT)");
const r = await db.execute("INSERT INTO notes(title) VALUES (?)", ["Hello SQLite"]);
const { columns, rows } = await db.query("SELECT id, title FROM notes WHERE id = ?", [r.lastInsertId]);
```

限制：文件位于 `mygo.PathUserData`（`Directory` 可改），`BusyTimeout` 默认 5 秒；绝对路径、目录遍历、符号链接文件、Windows 设备名被拒绝；`execute` 只接受一条语句，`query` 结果全量缓冲（大表用 `LIMIT` 分页）；参数按 SQLite 索引绑定，缺/多参数失败，绝不拼接 SQL；事务控制语句、`ATTACH`/`DETACH`/`VACUUM INTO` 被拒绝，扩展无法加载；`transaction` 用 `BEGIN IMMEDIATE`，错误或取消回滚；`lastInsertId` 在 TS 恒为 `bigint`；连接归发起页面所有，页面导航/窗口关闭会关闭它；整数超出安全范围绑 `bigint`，超出 SQLite int64 失败；取消经 `sqlite3_interrupt` 中断。

### watch

`@mygo-plugins/watch` + `mygo.Use(watch.New(watch.Options{Roots, Resolve, Allow}))`，根为 `map[string]watch.Root{Path, Recursive, Ignore}`；页面 `watch(name, {recursive, debounceMs, signal})` 返回可迭代 watcher，事件 `{op, path, oldPath, isDir}`。示例：`const w = await watch("project", { recursive: true }); for await (const events of w) { /* … */ }`。Go/原生 UI 直接用 `mygo.WatchFiles`。
限制：页面按名称而非路径指定根，名称由小写字母/数字/短横线组成且以字母开头；路径须为绝对路径且在 `mygo.Use` 时存在，否则 panic；`Resolve` 在运行时授予新根；`Allow(ctx, watch.Request)` 在每次监视前询问；`op` 为 `create`/`write`/`remove`/`rename`，`.` 表示被监视内容本身；启动时已存在的内容不报告；最后一次更改后 50 ms 报告，持续更改最多 500 ms；只报告差异，把事件当作重新读取的提示；每页面 8 个、全部页面 128 个监视，`MaxEntries` 8192（macOS 65536）；`WatchError.code`：`denied`/`not-found`/`limit`/`unsupported`/`path-encoding`/`io`；不跟随符号链接或 Windows 目录联接。

### loading

原生 UI 的 loading.dev 29 种加载指示器，外观与动效同浏览器版本；纯 Go，无 npm 包，无需 `mygo.Use`。`loading.Arc(c)`、`Ring`、`Blocks`、`BouncingDots`、`Wave` 等 29 个函数，链式 `.Label("Saving")`；`loading.All()`；`loading.Options{Size, Color, Duration, Paused, Easing, Cap, Sweep, Direction, Origin}`。示例：`ui.Row(c).Gap(8).Children(func() { loading.Arc(c).Label("Saving") })`。
限制：`Size` 默认 20 DIP，`BouncingDots` 与 `LinearDots` 是点行（宽分别为高的 1.15、0.9375 倍）；`Color` 零值用主题 `Text`；`Duration` 各指示器不同；`Paused` 暂停在当前位置；`Easing` 只影响旋转类指示器；`Cap` 决定描边端点；桌面要求减少动效时停在 loading.dev 的减少动效姿态且不请求帧；辅助技术识别为进度不确定的进度指示器，用 `Label` 说明在加载什么。

### sfsymbols

macOS 原生 UI 绘制 Apple SF Symbols，符号从运行中的系统读取，应用不含 Apple 图形素材；纯 Go，无需 `mygo.Use`。`sfsymbols.Symbol(c, "gearshape")` + `.Size`/`.Color`/`.Weight`/`.Fill()`/`.Natural()`/`.Fallback(svg)`；`Names()`；`Image(name, px, col, opts)`；`Available()`。示例：`sfsymbols.Symbol(c, "trash").Size(20).Weight(sfsymbols.Semibold).Fill()`。
限制：其他操作系统上 `Image` 返回 false、`Names()` 返回 nil，要准备自己的图标回退；`Image` 的 `px` 限于 8–512（更大的 `Symbol` 从 512 放大）；不跟随所在元素文本颜色，也不与旁边文本按基线对齐；不支持多色/分层符号；`Names()` 读系统目录（非公开 API），某版本移动位置时返回 nil。

### permissionflow

检查与请求辅助功能、屏幕录制、麦克风等需用户授予的系统权限，并引导到系统设置授权（macOS 在窗口下方显示随窗口移动的拖放面板）；纯 Go，网页通过应用自己的绑定访问。`Check(perm)` → `Granted`/`Denied`/`NotDetermined`/`Unknown`，`Request(perm)`，`Wait(ctx, perm)`，`Guide(ctx, perm, GuideOptions{})`，`OpenSettings`，`Allowed(perm)`；权限常量 `Accessibility`、`ScreenRecording`、`InputMonitoring`、`FullDiskAccess`、`AppManagement`、`DeveloperTools`、`Microphone`、`Camera`、`Calendars`、`Reminders`、`Bluetooth`、`MediaAppleMusic`。示例：`st, err := permissionflow.Check(permissionflow.Accessibility)`，未获准时 `permissionflow.Guide(ctx, permissionflow.Accessibility, permissionflow.GuideOptions{})`。
限制：macOS 不提供辅助功能/屏幕录制/输入监控“是否请求过”的信息，这些权限在获准前返回 `Denied`，屏幕录制变化需重启应用才能检测；`Request` 的提示每个应用只显示一次，麦克风、相机、日历、提醒事项、蓝牙需要 `macos.infoPlist` 的用途说明，缺失时 `Request` 返回 `ErrUsageDescription`（系统会终止应用）；系统不维护该类权限时返回 `errors.ErrUnsupported`（Windows 只有麦克风与相机，Linux 沙箱外什么都没有，浏览器中 `Allowed` 为 false）；`Request`/`Wait`/`Guide` 会阻塞主线程，从 goroutine 调用；`OpenSettings` 与 `Guide` 需要 `App.Run` 已启动；一次只显示一个引导面板。

### render

把原生 UI 视图按帧号渲染成视频（MP4/MOV/WebM），GPU 绘制、系统编码器编码，可带声音；不运行浏览器，不需要窗口或 `mygo.App`。`render.Composition{Width, Height, FPS, Frames, Scale, View, Dark, Preferences, Locale, Background, Transparent, MotionBlur}`；`render.Video(ctx, comp, path, Options)`、`Images`、`Still`、`Frames`；`Frame` 的 `Index`/`Time`/`Seconds`/`Progress`/`Interpolate`/`InterpolateColor`/`Spring`/`Sequence`/`Shift`/`Loop`/`Exact`/`Audio`；`Options{From, To, Renderer, Codec, Bitrate, AudioBitrate, Quality, Progress, Encoder, FFmpeg}`；`render.NewPlayer(comp)` 预览。

```go
comp := render.Composition{
	Width: 1920, Height: 1080, FPS: 30, Frames: 90, Scale: 2,
	View: func(c *ui.Context, f render.Frame) {
		ui.Text(c, "Hello").Opacity(float32(f.Interpolate([]float64{0, 20}, []float64{0, 1})))
	},
}
err := render.Video(ctx, comp, "hello.mp4", render.Options{})
```

限制：`Scale` 是 1 DIP 对应多少像素（Scale 2 的 1920×1080 按 960×540 DIP 布局）；同一帧每次构建必须得到相同结果——先加载图片字体数据、从帧读时间而不是 `time.Now`，只渲染部分帧（`From`）时会先构建之前的帧但不绘制；默认 `H264`（macOS 与 Linux 为 MP4/MOV，Windows 为 MP4），`HEVC` 与 Linux 的 `VP9`（WebM）另有约束，`ProRes`/`ProRes4444` 仅 macOS 的 MOV；系统不能编码时返回包装 `errors.ErrUnsupported` 的错误并说明要装什么（Linux 需要 GStreamer 插件）；`Options{Encoder: render.FFmpeg}` 改用 ffmpeg 写任意容器（帧仍由 MyGo 绘制）；`MotionBlur` 每帧开销是 `Samples` 倍；透明合成只有 ProRes 4444、VP9 和 PNG 保留 alpha。

### glass

原生 UI 的 macOS 26+ Liquid Glass 材质、滚动边缘、背景模糊，着色器在 GPU 或 CPU 上绘制，三平台外观一致；纯 Go，无需 `mygo.Use`。`glass.Glass{Regular|Clear, Tint, Interactive, Clarity, Native}` 作为 `ui.Material`；`glass.NewGroup(c)`（`Radius`、`Gap`、`Pane`、`Merge` 默认 24，最多 8 个窗格）；`glass.ScrollEdge{Soft|Hard, Bottom, Background}`；`glass.Blur{Radius, Mask *ui.LinearGradient}`；`glass.Paint(p, r, radius, g)`。示例：`ui.Row(c).Padding(8, 16).Gap(8).Radius(22).Material(glass.Glass{})`。
限制：玻璃只透出同一窗口中此前绘制在下方的内容，子元素绘制在玻璃上；`Clarity` 0–1 只对超过 100 DIP 的窗格生效（100→200 DIP 渐入）；`Native: true` 仅在 macOS 26+ 用系统 `NSGlassEffectView`，只采样其下方表面（webview 页面或窗口材质），从不采样帧自身绘制的元素，否则用着色器；启用“降低透明度”时改为不透明表面（玻璃用主题 `Surface`，模糊与滚动边缘用 `Background`），提高对比度时增强边缘高光；`ScrollEdge` 应先绘制再用 `PassThrough` 让指针穿透，`Soft` 元素高度设为栏高加 10 DIP、`Hard` 等于栏高；`Group` 的 `Merge` 连接相距约 12 DIP 的窗格，0 表示不融合；无 GPU 时 CPU 绘制，下方内容变化会触发窗格与周围模糊区域重绘。

### game

原生 UI 的游戏循环、输入、摄像机、像素级缩放与游戏手柄；`game.Game` 接口的 `Update(in *game.Input)` 与 `Draw(p *ui.Painter, r ui.Rect, alpha float64)`，`game.View(c, g)`，`game.Rate`（60）；`game.Input` 的 `Down`/`Pressed`/`Released`/`Pointer`/`PointerDown`/`Wheel`/`Size`/`Focused`；`game.Camera{X, Y, Zoom, Rotation}` 的 `Draw`/`ToWorld`/`Visible`/`Transform`；`game.IntegerFit`；`gamepad.Pads()`/`Pad.State()`/`LeftStick()`/`JustPressed`/`Deadzone`（0.15）。

```go
func (f *farm) Update(in *game.Input) {
	f.px, f.py = f.x, f.y
	if in.Down(ui.KeyD) {
		f.x += 2
	}
}

func (f *farm) Draw(p *ui.Painter, r ui.Rect, alpha float64) {
	x := f.px + (f.x-f.px)*float32(alpha)
	p.ImageRegion(sheet, hero, ui.Rect{X: r.X + x, Y: r.Y + y, W: 48, H: 48}, ui.ImageOptions{})
}
```

限制：固定时间步长（每秒 60 步，可由 `Rate() int` 指定），与刷新率无关；暂停后最多补算 1/4 秒；`alpha` 是当前帧在下一时间步内的进度，用于插值；按键按位置识别（`KeyW` 总是 W 所在键）；视图获得焦点时接收键盘，失去焦点时松开所有按住的键并暂停；`Draw` 每帧运行，不限制帧率时（`MYGO_UNCAPPED=1`）步数不变；摄像机不旋转时按屏幕整像素移动，整数缩放让 `Pixelated` 位图每个像素占相同整数像素；手柄按 W3C/SDL 标准布局读取，拔插后是新 `Pad`；`Update` 是普通 Go 代码，可在 `go test` 中直接测试。

### automation

让另一个进程（端到端测试运行器或智能体）驱动运行中的应用：列窗口、读原生 UI 元素、操作元素、发指针/滚轮/按键/文本、选菜单项、在页面执行脚本、截图、等界面稳定。构建标签 + `MYGO_AUTOMATION=<socket path>` 启用；Unix 域套接字、每行一个 JSON；方法 `windows`、`tree`、`act`、`input`、`menu`、`capture`、`eval`、`idle`；错误码 `parse_error`/`invalid_request`/`unknown_method`/`no_window`/`no_element`/`unsupported`/`timeout`/`failed`；元素用 64 位数字字符串 `id` 或应用设置的 `key`；CLI `mygo-automation`（`open`/`snapshot -i`/`press @e2 --settle`/`fill`/`screenshot`/`close`）。

```go
//go:build e2e

package main

import _ "github.com/egoist/mygo/plugins/automation"
```

```sh
go build -tags e2e -o app . && MYGO_AUTOMATION=/tmp/app.sock ./app
```

限制：只有 `MYGO_AUTOMATION` 指定路径时才监听；不含原生 UI 检查器的构建（`-tags mygo_noinspector`、`mygo build` 发布构建，除非 `-debug`/`MYGO_INSPECTOR=1`）会忽略它并在标准错误说明；**任何能打开套接字的人都能像用户一样驱动应用并执行脚本，不要放进对外发布的构建**；同一连接按顺序逐个回答，要并行就开多个连接；元素 ID 可能跨帧变化，长期有效的定位用 `Key`；显示网页的窗口 `tree` 回答 `unsupported`，改用 `eval`；`menu` 不覆盖 Dock/托盘菜单与上下文菜单；`idle` 不等待 `Context.After` 安排的事，永不稳定的内容要传 `timeout`；窗口不可见时会自行构建帧，`tree`/`capture`/`idle` 报告当前显示内容；升级应用不会升级已安装的 `mygo-automation`。

### updater

三平台的 Sparkle 式更新窗口：后台检查、显示发布说明、安装 / 稍后提醒 / 跳过、下载进度、重启进新版本；纯 Go，无 npm 包，构建时必须启用更新。`mygo.Use(updater.Plugin)`；`updater.CheckForUpdates()`；`updater.MenuItem()`；`updater.New(updater.Options{Interval, DisableAutomaticChecks, Icon, Language, Strings})`；偏好 `AutomaticChecks`/`AutomaticDownloads`/`SetAutomaticChecks`/`SetAutomaticDownloads`/`LastCheck`/`AvailableVersion`/`OnChange`/`OnInstalled`；原生 UI 用 `plugins/updater/native.Plugin`。

```go
mygo.Use(updater.New(updater.Options{
	Interval:               12 * time.Hour, // 默认一天
	DisableAutomaticChecks: true,
	Icon:                   iconPNG, // 默认 resources 里的 icon.png
}))
```

限制：启动 10 秒后检查、运行期间每天一次；用户选择保存在 `PathUserData` 的 `updater.json`；无法自我更新的构建（开发构建、包管理器安装）从不后台检查，手动检查时窗口说明原因；跳过只对本次版本有效，会提供下一个版本；`AvailableVersion` 在检查前/失败/无更新/已跳过时为空；窗口语言取检查开始时的 `App.Locale`（内置英语、简繁中文、荷法德意日韩波兰葡萄牙俄西土乌，其他回退英语），打开期间不变；`Strings` 的格式字符串不匹配时 `mygo.Use` 失败；只使用 `updater.Plugin` 与 `native.Plugin` 之一；网页版窗口在 Linux 需要 WebKitGTK、Windows 10 需要 WebView2 Runtime，纯原生 UI 应用用 `native.Plugin`。

### terminal

原生 UI 的终端视图，在伪终端中运行 shell 或任意程序，使用 Ghostty 的 libghostty-vt（原生库，运行时加载，无 cgo，`mygo build`/`mygo dev` 打包进应用）。`terminal.New(terminal.Options{Command, Dir, Env, Font, Theme, Cursor, Scrollback, ShellIntegration, Keybinds, RightClick, Padding, PaddingColor, OptionAsAlt, CopyOnSelect, ClipboardRead, NoClipboardWrite, OnClipboardRequest, OnCopy, OnClear, OnTitle, OnExit, OnBell, OnNotify, OpenLink, OnPaste, OnPasteStatus, OnKey, OnScroll, OnZoom, Conn, ResizeByConn})`；`terminal.View(c, term).Fill().AutoFocus()`；方法 `Send`、`Paste`、`PasteData`、`Feed`、`Title`、`Dir`、`Size`、`Text`、`HoveredLink`、`SetFont`、`SetTheme`、`Resize`、`Snapshot`、`Close`；服务器/远程用 `Conn`、`NewMirror`、`Frame`。

```go
term, err := terminal.New(terminal.Options{})
win := mygo.NewWindow(mygo.WindowOptions{
	Title:   "Terminal",
	Content: ui.View(func(c *ui.Context) { terminal.View(c, term).Fill().AutoFocus() }),
})
win.OnClosed(func() { term.Close() })
```

限制：程序在视图第一次显示时启动（以便按视图大小绘制第一屏），没有视图时半秒后以 80×24 启动；一个终端一次只在一个视图中显示；`Command` 为空时运行登录 shell（`$SHELL`，否则 macOS `/bin/zsh`、其他 `/bin/sh`、Windows PowerShell），`Dir` 默认家目录；`Scrollback` 默认 10 MB、`Font` 默认 13 DIP 系统等宽字体；`ShellIntegration` 默认关闭且会写临时 `.zshenv`（不修改用户文件，退出时删除）；`ClipboardRead` 默认拒绝程序读取，`NoClipboardWrite` 阻止程序写剪贴板；`OnExit`/`OnTitle`/`OnBell`/`OnNotify` 可能在窗口创建前运行，需等待窗口；`OpenLink` 能识别 `path:line:col` 形式的路径；`Mirror` 用于在另一进程显示服务器终端，帧须在 `term.Resize` 之后处理；`OnClear` 应交给服务器执行清除，否则服务器后续帧会把行带回错误位置。

## 速查

| 主题 | API / 配置 | 要点 |
|---|---|---|
| 前端加载 | `devUrl`、`devCommand`、`buildCommand`、`frontendDist` | 无协议 URL 相对解析；构建后由 `mygo://localhost` 提供 |
| 源与存储 | `mygo://localhost`、`http://mygo.localhost` | `localStorage` 属于源，dev 与构建后不同；数据放 Go |
| 自定义协议 | `mygo.Protocol.Handle/HandleFunc`、`mygo.FileServer` | 窗口创建前注册；`mygo` 协议替换前端 |
| 跨源 | 应用页面自动获得 CORS 头；`TrustedOrigins` | Windows 协议地址是 `http://<scheme>.localhost` |
| 信任 | `PageOptions.TrustedOrigins`、`PageOptions.Untrusted` | 受信任页面能调 Go、控窗口、获权限；勿承载第三方 HTML |
| 运行时包 | `call`/`on`/`Channel`/`runtime()`/`currentWindow`/`preferences` | `isMyGo()` 判断是否在应用窗口 |
| 标题栏 | `Frameless`、`TitleBarHidden`、`--app-region: drag` | 变量 `--mygo-titlebar-*`；全屏为 `0px` |
| 外观 | `prefers-color-scheme`、`--mygo-accent`、`--mygo-text-scale` | `BackgroundColor: light-dark(…)` 避免首帧闪烁 |
| 绑定 | `mygo.Bind`、`mygo.BindAs` | 在 `App.Run` 之前；名称/类型冲突 panic |
| 生成客户端 | `mygo generate`、`bindings` | `mygo dev`/`mygo build` 自动生成；属性名冲突则失败 |
| 错误 | `CallError`、`isCallError`、`err.method` | 只有 Go 的 error 会拒绝 promise |
| 事件 | `mygo.NewEvent[T]`、`Emit`、`Broadcast` | 事件名唯一、`mygo:` 保留；窗口关闭时 `Emit` 返回 error |
| 通道 | `*mygo.Channel[T]`、`Channel.close`、`ErrChannelClosed` | 一次调用一条；未读超过 1 MiB 时 `Send` 等待 |
| 调用权限 | 受信任源、`TrustedOrigins` | iframe 永远不能调用；`"*"` 慎用 |
| 执行脚本 | `Page.Eval`、`mygo.EvalAs[T]` | 语句是异步函数体；抛异常为 `*mygo.EvalError` |
| 插件启用 | `mygo.Use(plugin.Plugin)` + `@mygo-plugins/<name>` | 须在 `App.Run` 前；只装 JS 不启用 Go 会报错 |
| 编写插件 | `mygo.Plugin{Name, Service, Setup}`、`plugin:<name>` | 不进生成的客户端；JS 用 `call` |
| 纯 Go 插件 | loading、sfsymbols、glass、game、render、permissionflow、terminal | 无需 `mygo.Use`，也没有 npm 包 |
| 网页插件 | fetch、websocket、sqlite、watch | 需 `mygo.Use` 与 npm 包 |

## 文档未说明

- `mygo-runtime` npm 包与 CLI 的版本兼容矩阵：文档未说明。
- 各插件在旧版 macOS/Windows 上的最低系统要求：除个别（glass 需 macOS 26、sqlite 需 iOS 16 等）外文档未逐一说明。
