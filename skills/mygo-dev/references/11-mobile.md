# 手机与平板

> 需要 mygo ≥ v0.10.0

定位：同一个 MyGo 应用如何运行在 iPhone、iPad 与 Android 手机/平板上——开发与设备调试、手机上的界面与手势、设备 API 与插件、构建签名与商店发布。Go 代码、绑定的服务与事件不变；网页窗口用系统 webview（iOS 的 WKWebView、Android 的 WebView），原生 UI 用 GPU 绘制（iOS 的 Metal、Android 的 OpenGL ES）。

## 源页面

- 手机和平板 — <https://mygo.run/zh/docs/mobile>
- 为手机构建和发布 — <https://mygo.run/zh/docs/mobile/building>
- 手机上的界面 — <https://mygo.run/zh/docs/mobile/interface>
- 设备 API 和插件 — <https://mygo.run/zh/docs/mobile/plugins>

## 准备与工具链

**用途**：确认目标平台缺什么，以及 `-device` 能接受什么。

```sh
mygo doctor                    # what is there, and what is missing
mygo devices                   # what -device takes
mygo dev -platform ios         # the booted iOS simulator
mygo dev -platform android     # the device or emulator adb reaches
mygo build -platform ios       # a signed .ipa
mygo publish -platform ios     # the .ipa to App Store Connect
```

- **iOS**：一台装有 Xcode 的 Mac。Go 代码用 Xcode 的 clang 链接（Go 为 iOS 链接需要 cgo），CLI 编译 Swift 宿主、图标与应用包并签名，不需要 Xcode 项目。在设备上运行与上架需要 Apple 开发者团队，以及与 `identifier` 匹配的描述文件（Xcode 设置 › 账户，或 developer.apple.com）。
- **Android**：任意系统上装有 SDK 与 NDK 的 Android Studio（SDK Manager › SDK Tools › NDK）。Go 代码为每个 ABI 构建成共享库、由 NDK 的 clang 链接，Gradle 构建应用；CLI 在 SDK 常用位置或 `ANDROID_HOME` 找 SDK，在 SDK 中或 `ANDROID_NDK_HOME` 找 NDK，用 Android Studio 的 JDK 或 `JAVA_HOME`（JDK 17+）。
- 桌面构建仍然不需要 cgo，也不需要这些工具。`mygo doctor` 缺少项目时只给警告、不算失败；只面向桌面的应用一样都不需要。
- `mygo devices` 列出 iOS 模拟器与运行时、iOS 设备、Android 设备/模拟器、尚未运行的 AVD，`-json` 输出脚本可读形式。它询问 `xcrun simctl`、`xcrun devicectl`（Xcode 15+）、`adb`、`emulator`，问不到的就不列出。

## 开发循环

**用途**：改代码后几秒内在模拟器或设备上看到结果，并保持应用状态。

- `mygo dev -platform ios` 为模拟器构建、安装到已启动的模拟器（或 `-device` 指定的装置，按名称或 UDID，会启动它）并运行，应用输出显示在控制台。`-platform android` 对 `adb` 能连的设备/模拟器做同样的事（或 `-device` 指定序列号），通过 `adb reverse` 访问开发服务器。`-device` 也接受 AVD 名称：未运行的 AVD 在 `mygo dev` 之外启动，流程等其系统启动完成，结束后它继续运行。
- 前端的 `devUrl` 与桌面一样可用：模拟器能访问 Mac 上的开发服务器；设备不在 Mac 的网络栈上，`localhost` 上的 `devUrl` 改用 Mac 在网络中的地址，开发服务器须监听该地址（`vite --host`）。没有 `devUrl` 时应用嵌入 `frontendDist`，因为设备读不到电脑的磁盘。
- Go 改动约两秒生效：宿主 Swift 保持已编译，iOS 把新可执行文件放进已安装应用，Android 不经 Gradle 直接把新库传给可调试应用；宿主、插件、资源或配置改动会重新构建并安装整个应用。Go 无法替换运行中进程的代码，所以每次改动重启应用并回到原来的位置。
- 滚动位置、`ui.Router` 所在页面、控件用 `ui.Local` 保存的内容自动恢复。应用自己的值（含未导出字段）用一行恢复：`a := &app{tab: 0}; mygo.KeepState(a)`。它保留数据：函数、通道、接口与指向其他包类型的指针保持程序创建时的样子，类型被本次改动修改过的字段保持新建时的值。需要更多控制用 `App.OnSaveState` 与 `App.RestoredState`（桌面 `mygo dev` 同样可用）。开发版应用名为 `<name> Dev`、标识符 `<identifier>.dev`，可与正式版并存。
- **iOS 真机**：需要开启开发者模式（设置 › 隐私与安全性）。应用为设备构建并用开发描述文件签名——从已安装描述文件中选最匹配 `<identifier>.dev` 的那个（通常是 Xcode 维护的通配符描述文件，证书在钥匙串中且列出该设备）；`ios.provisioningProfile` 用于 `ios.distribution`，只有它是 `development` 时才用于开发版。`xcrun devicectl` 负责安装与启动，其 `--console` 不可用时应用照常运行但看不到输出。通配符描述文件不能带关联域，使用时会省略 `ios.associatedDomains` 并提示。

**限制/常见错误**：iOS 上 Go 链接必须 cgo，因此必须有 Xcode；`mygo doctor` 只提示缺什么，不会替你安装。

## 应用在手机上的行为

**用途**：了解与桌面不同的窗口、生命周期与系统集成语义，避免把桌面假设带过来。

- **一个屏幕**：最后显示的窗口占满屏幕、位于之前窗口之上；隐藏或关闭它显示下面的窗口。窗口的大小、位置与标题栏不起作用；`Window.Bounds` 是屏幕范围。**安全区**：原生 UI 在安全区域内布局（避开系统栏、挖孔、键盘），窗口背景铺满边缘；网页在系统栏之内布局，Android 上还在键盘之上。`Context.SafeArea` 给出边缘被什么遮挡，供 `Overlay` 使用。
- **返回**：Android 的返回会弹出最后推入的窗口、在网页中后退、关闭原生 UI 中 Escape 会关闭的东西、在 `ui.Router` 中后退，否则让应用在后台继续运行；系统提前知道这些以支持预测性返回（`ui.Router` 页面跟随手势移动，没有可返回内容时显示应用离开动画）。
- **生命周期**：应用从不自己退出；进入后台是电源事件 `"suspend"`（`Power.OnSuspend`），回前台是 `"resume"` 与 `App.OnActivate`，两者之间帧停止。系统可能在后台结束应用：`App.OnSaveState` 保存、下次启动在 `App.RestoredState` 中取回。内存不足时 `App.OnMemoryWarning`：原生 UI 释放渲染器纹理与文本缓存，下一帧重建。
- **系统集成**：深度链接（`urlSchemes`）经 `App.OnOpenURL` 打开；消息框是系统警告框，`Dialog.Open` 显示系统文稿选择器并在应用临时文件中提供副本，`Dialog.Save` 提供写入用临时文件，`Dialog.SaveCopy` 让用户选保存位置；`WatchFiles` 与桌面一致；密钥在 iOS 存入钥匙串、Android 存入应用存储并用 Android 密钥库中的密钥加密；通知第一次使用时请求许可。
- **没有的功能什么都不做**：菜单栏、托盘、程序坞、全局快捷键、更新器（手机应用由应用商店更新）。桌面 API 在手机上要么无操作，要么返回 `errors.ErrUnsupported`。
- **Android webview 不加载自定义协议**：前端（`mygo://localhost`）与其他协议在 `https://<scheme>.localhost` 上提供，网页除 `location.href` 外察觉不到。对应用协议的 `fetch` 与 `XMLHttpRequest` 保留请求体；提交到应用协议的表单不带请求体到达（Android 不提供），表单请用 `fetch` 或绑定方法提交。

## 构建与签名

**用途**：产出 App Store / Google Play 接受的产物，并为正式版签名。

```sh
mygo build -platform ios             # dist/ios-arm64/<slug>-<version>.ipa, signed
mygo build -platform ios/simulator   # dist/ios-simulator/<name>.app
mygo build -platform android         # dist/android/<slug>-<version>.apk and .aab
mygo build -platform android/arm64   # one ABI: arm64, arm or amd64
```

- `.ipa` 用已安装描述文件中最匹配应用标识符与 `ios.distribution`（默认 `app-store`，TestFlight 也用它）的那个签名，证书须在钥匙串中；它带 App Store Connect 要求的隐私清单，按应用代码用到的内容生成，除非项目自带 `ios/PrivacyInfo.xcprivacy`。引用了 Apple 私有方法时 CLI 拒绝构建（App 审核会拒）。
- `.aab` 是 Google Play 接受的格式，`.apk` 供其他商店与设备安装。正式版用 `android.keystore` 签名，密码来自 `MYGO_ANDROID_KEYSTORE_PASSWORD` 与 `MYGO_ANDROID_KEY_PASSWORD`；否则用调试密钥，而 Google Play 不接受调试密钥。

```ts
export default defineConfig({
  name: "Todo",
  identifier: "dev.example.todo",
  ios: { team: "ABCDE12345", infoPlist: { NSCameraUsageDescription: "Scan receipts." } },
  android: { permissions: ["CAMERA"], keystore: "upload.jks", keyAlias: "upload" },
});
```

**应用扩展**：应用扩展（如 VPN 的 Network Extension / packet tunnel provider）是 `PlugIns/` 中独立、带自己描述文件与权限的包。用 Xcode 单独构建后在 `ios.extensions` 中列出 `{path, provisioningProfile}`；CLI 把它复制进 `PlugIns/`、在签名应用前用应用的证书签名，描述文件作为 `embedded.mobileprovision` 嵌入。扩展标识符须以应用标识符加一个点开头，否则 iOS 拒绝安装；版本号与构建号应与应用相同（App Store Connect 会检查）。模拟器构建以 ad hoc 签名；真机 `mygo dev` 用开发描述文件，开发应用 `com.example.vpn.dev` 以 `com.example.vpn.dev.tunnel` 携带扩展——扩展标识符应由 `Bundle.main.bundleIdentifier` 推出。

**限制/常见错误**：正式版必须有 `android.keystore` 与其密码环境变量，否则只能得到调试签名、商店拒收；iOS 描述文件与 `identifier`、团队、能力不匹配时签名失败。

## 发布

**用途**：把构建好的产物上传到 App Store Connect / Google Play。

```sh
mygo build -platform ios && mygo publish -platform ios
mygo build -platform android && mygo publish -platform android -track beta
```

- `mygo publish` 上传 `mygo build` 为当前版本生成的产物，或 `-file` 指定的文件。
- **iOS**：用 `xcrun altool` 把 `.ipa` 传到 App Store Connect，处理完成后出现在 TestFlight。需要团队 API 密钥（App Store Connect › 用户和访问 › 集成），密钥从不出现在命令行：`MYGO_ASC_KEY_ID`（密钥 ID）、`MYGO_ASC_ISSUER_ID`（签发者 ID）、`MYGO_ASC_KEY`（`.p8` 路径）。
- **Android**：通过 Google Play 的 Android Publisher API 上传 `.aab`——创建编辑、上传应用包、在轨道创建发布、提交；任何一步失败都会删除这次编辑。`MYGO_PLAY_SERVICE_ACCOUNT` 是服务账号 JSON 密钥的路径（也可直接是 JSON，便于放 CI 机密），该账号需在 Play 管理中心有应用访问权限。`-track` 默认 `internal`，也可 `alpha`、`beta`、`production` 或封闭测试轨道名。上传的目标包就是应用的 `identifier`。

**限制/常见错误**：应用的第一次上传只能在 Google Play 管理中心完成；尚未经审核的应用只接受 `-status draft` 的发布。iOS 上传依赖 Xcode 的 `altool`，换机或 CI 要准备同一团队的密钥。

## 图标、启动画面、链接与后台模式

**用途**：手机平台专属的图标/启动/链接/后台能力都在 `ios` 与 `android` 配置段。

```ts
export default defineConfig({
  ios: {
    associatedDomains: ["example.com"],
    launchScreen: { background: "#102030", image: "resources/logo.png" },
    backgroundModes: ["audio", "remote-notification"],
  },
  android: {
    appLinks: ["example.com"],
    adaptiveIcon: { foreground: "resources/icon-foreground.png", background: "#102030" },
    splash: "#102030",
  },
});
```

- **链接**：域名下的 https URL 打开应用并传给 `App.OnOpenURL`。`ios.associatedDomains` 让应用带 `applinks:example.com` 签名（`webcredentials:` 等自带服务的值保持原样），应用 ID 需要“关联域”能力，主机需提供 `apple-app-site-association`；`android.appLinks` 为每个主机加一个验证过的 intent filter（`*.example.com` 含子域名），主机需提供列出应用及签名密钥的 `/.well-known/assetlinks.json`。
- **Android 图标**：`android.adaptiveIcon` 是 Android 8.0+ 的启动器图标，432×432 或更大的 PNG 叠在一种颜色上；启动器裁成中间 72/108 并可能移动它，重要内容放在那里；旧系统仍显示 `icon`。
- **启动画面**：`ios.launchScreen` 是一种颜色、居中图片或两者（不需要 storyboard，CLI 编译进资源目录并在 `UILaunchScreen` 引用），图片按 3 倍屏尺寸，600×600 像素占 200×200 点；`android.splash` 是启动画面颜色，应用图标显示在其上。`ios.backgroundModes` 写入 `UIBackgroundModes`，并加上插件要求的那些。

**限制/常见错误**：可执行程序（如 VPN 的代理核心）在 Android 上要作为原生库随原生模块发布到 `jniLibs/<abi>/lib<name>.so`；Android 只允许执行 `nativeLibraryDir` 中的文件（W^X，targetSdk 29+），需设 `android.extractNativeLibs: true` 让安装器解压，再从 `context.applicationInfo.nativeLibraryDir + "/lib<name>.so"` 使用。此类可执行文件须按 16 KB 页链接（`-Wl,-z,max-page-size=16384`），否则在 16 KB 页设备上一启动就崩溃。

## 界面：标签页与栈

**用途**：用底部的系统标签栏与屏幕栈组织手机导航。

```go
mygo.App.WhenReady(func() {
	list := mygo.NewWindow(mygo.WindowOptions{Title: "Fruits", Content: ui.View(app.list), Hidden: true})
	web := mygo.NewWindow(mygo.WindowOptions{Title: "Web", URL: "/", Hidden: true})
	tabs, _ := mygo.NewTabs(
		mygo.TabItem{Title: "Fruits", Symbol: "leaf", Icon: leafPNG, Window: list},
		mygo.TabItem{Title: "Web", Symbol: "globe", Icon: globePNG, Window: web},
	)
	tabs.SetBadge(1, "3")
	list.Push(mygo.NewWindow(mygo.WindowOptions{Title: "Apple", Content: ui.View(app.detail("Apple")), Hidden: true}))
})
```

- `mygo.NewTabs` 显示系统标签栏（iOS 的 UIKit、Android 的 Material 底部导航）；`Window.Push` 把窗口放进所在标签页的栈；返回按钮、iOS 左缘轻扫、Android 返回、`Window.Pop` 都会弹出并关闭它。标签项在 iOS 指定 SF Symbol，在 Android（iOS 缺符号时）提供 PNG，按栏颜色作模板绘制。iOS 上栏浮在原生 UI 之上（内容在玻璃下继续滚动），Android 用窗口背景色。桌面上这些就是普通窗口：`Tabs.Select` 显示其一，推入的窗口也显示为一个窗口。
- `ui.Router` 设置 `SystemStack` 后把从第一页到当前页的历史显示为系统栈：顶部系统导航栏显示 `Route.Title` 与返回按钮；iOS 由 UIKit 导航控制器移动页面并支持左缘轻扫，Android 是 Material 工具栏、页面像 Android 应用那样滑动。`InSystemStack` 为真时不要再画自己的栏。`Router` 的 View 必须占满窗口；iOS 上窗口只有一个绘制表面，它跟随当前页移动，下面的屏幕保留各自最后显示时的快照（返回时先露快照、回到该页再实时显示）。
- iPad：`mygo.json` 设 `"ios": {"multipleWindows": true}` 后窗口可成为系统窗口并与应用其他窗口并排（`mygo.WindowOptions{OwnScene: true}`）；其他设备上仍是覆盖前一个窗口。Android 平板自行与其他应用分屏。

## 手指、文本与手机控件

**用途**：原生 UI 按系统工具包的方式处理触摸与输入；`ui/mobile` 提供手机特有控件。

- **手指**：轻点即点击，命中范围 14 DIP；移动的手指锁定到移动方向与在该方向上滚动的最内层容器（纵向列表中的横向照片只收横向移动），抬起后惯性滑动，落下即停且不会点到东西；过界后内容跟随变少再弹回（iOS 弹得远、Android 近）并显示滚动指示器；沿某方向拖动的元素（如横向 `Draggable`）把纵向移动留给列表；按住半秒是长按（`Element.LongPressed`、上下文菜单、选中词、`Reorder` 拖动），伴随触感；两指可捏合与旋转（`InputEvent` 的 `InputZoom`、`Rotation`）；iOS 上左缘轻扫让 `ui.Router` 返回、轻点状态栏让可滚动内容回到顶部。
- **下拉刷新**：`Element.PullToRefresh` 让手指把滚动容器拉过顶部——`list.PullToRefresh(&app.refresh)` 后由 `win.Update` 更新数据与 `Refreshing`。
- **文本**：键盘通过输入元素的 `TextInputClient` 编辑，带系统输入建议与输入法。元素说明装的内容以选键盘：`ui.TextInput(c, &app.email).Keyboard(ui.KeyboardEmail).ReturnKey(ui.ReturnNext).Capitalize(ui.CapitalizeNone).Autocorrect(false).Autofill(ui.AutofillEmail)`；`ReturnNext` 移到下一个元素，`Password` 是安全输入；键盘弹出时让正在使用的输入元素保持可见；按住词选中并显示系统操作菜单，选区手柄带放大镜。
- **`ui/mobile` 控件**（不导入不计代价）：`Sheet`、`ActionSheet`（底部面板，可拖走）、`Drawer`、`PageView`/`PageDots`、`SwipeRow`、`Wheel`/`DatePicker`/`TimePicker`（带触感）、`TopBar`/`BottomBar`/`FloatingButton`（应用自绘栏，窗口内路由用）、`ListRow`/`Section`、`Ink`（Material 水波纹）、`Switch`、`Zoomable`。`mobile.SetLook` 可不论系统选择一种设计。
- **网络图片**用 `ui/netimage` 代替 `ui.Image`：`netimage.Image(c, url, netimage.Options{MaxSide: 240}).Size(80, 80).Radius(40)`；占位符先占位、加载后淡入，按缓存头跨运行存盘，离线可见；屏幕上的图片优先下载，划过的根本不下载。桌面与浏览器同样可用。

## 设备 API 与官方插件

**用途**：手机独有能力按「设备 API」与「插件」两层提供。

| 设备 API | 作用 |
|---|---|
| `mygo.Haptics` | 选择、碰撞或任务结束方式的触感 |
| `mygo.SystemBars` | 系统栏图标颜色、隐藏系统栏 |
| `mygo.SetOrientations` | 屏幕可如何旋转 |
| `mygo.Share` | 系统共享面板（文本、链接、文件） |
| `mygo.Network` | 是否在线、流量是否计费 |
| `mygo.Power.Battery` | 电池电量、低电量模式 |
| `mygo.SetBarItems` | 窗口系统导航栏中的按钮 |
| `Context.SafeArea` | 窗口边缘被什么遮挡 |

需要权限或较大依赖的功能是插件，只有导入才拥有，其用途说明与权限会加入应用：`plugins/maps`（MapKit 与 osmdroid，无需密钥）、`location`、`imagepicker`、`camera`（AVCaptureSession / CameraX）、`videoplayer`（AVPlayer / MediaPlayer）、`biometrics`、`sensors`、`push`（APNs / Firebase，iOS 需描述文件带推送能力、Android 需 Firebase 项目）、`purchases`（StoreKit 2 / Google Play 结算）、`permissionflow`、`preferences`、`sqlite`（带全文搜索，编译进应用）、`background`（BGTaskScheduler / WorkManager）。

## 编写手机插件

**用途**：把 Swift/Kotlin 原生代码以两种方式接入应用：原生模块（推荐）或 JSON 调用的插件。

- **原生模块**（推荐）：一个 Go 接口 + iOS 的 Swift / Android 的 Kotlin 实现，Go 像调用普通 Go 值一样调用它（带类型、错误、流、回调），`mygo dev` 与 `mygo build` 随应用从源码构建。`mygo native new -lang swift,kotlin torch` 生成模块骨架：

```go
//mygo:native
//mygo:mainthread
type Torch interface {
	// Set lights the torch, or puts it out.
	Set(on bool) error
}
```

  模块的 `Info.plist`、权限与 Gradle 依赖会加入应用。SQLite 插件就是这样的模块（iOS Swift、Android 带 C 的 Kotlin、桌面 Zig）。
- **JSON 插件**：一个 Go 包旁边放 Swift 与 Kotlin，由 `mygo-plugin.json` 的 `mobile` 段列出（`id`、`ios.sources/class/frameworks/infoPlist`、`android.sources/class/permissions/dependencies/application`）；`mygo build` 在应用导入的包中找到它，把代码编译进宿主并注册。Go 用 `mygo.Native("torch").Call(ctx, "set", map[string]bool{"on": on}, nil)` 调用，用 `mygo.Native("torch").On` 订阅事件；Swift 放在宿主的模块中，Kotlin 放在包 `dev.mygo` 中、在 UI 线程运行，各自实现 `MyGoPlugin` 的 `call` 并在任意时刻回复一次，`PluginHost.emit` 发事件。`PluginHost` 提供宿主能力：iOS 的 `presenter`、`applicationEvent`，Android 的 `activity`、`startForResult`、`requestPermissions`；错误以 `*mygo.NativeError` 到达 Go。
- **C 代码**：通过原生模块进入手机——Kotlin 用 CMake（`externalNativeBuild`、NDK）为应用的 ABI 构建，Swift 包可带 C target 或以二进制 framework 引入库。
- **原生视图**：`PluginHost.registerView` 在插件初始化方法中注册视图，`mygo.NewPluginView` 为原生 UI 窗口创建它、由 `ui.WebView` 放置；视图显示在内容之下、内容帧留出的空洞里，之后绘制的内容在它之上，落在这些内容上的手指仍归内容所有；视图的 `ID` 在相关调用中指代它（`plugins/videoplayer` 是简短示例）。

## 测试与性能

**用途**：在开发机与真机上验证界面、帧时间与端到端流程。

- 单元测试用 `ui.Tester` 运行视图、推进时钟，并用 `FingerDown`/`FingerMove`/`FingerUp`/`SetInsets`；`ui/golden` 把绘制结果与测试旁的图片比较：`golden.Match(t, ui.NewTester(card, 320, 200), "testdata/card.png")`，`MYGO_UPDATE_GOLDEN=1` 写入图片（须在运行测试的平台上写，各平台文本渲染不同）。
- `MYGO_FRAME_STATS=1` 记录慢于 8 毫秒的帧及耗时所在；手机上模拟器用 `SIMCTL_CHILD_MYGO_FRAME_STATS=1`，Android 用 `adb shell am start ... --es mygo.env MYGO_FRAME_STATS=1`。
- 端到端测试从另一进程驱动运行中的应用（模拟器与真机一致），用自动化插件读元素、操作、发手指与文本、截屏、等画面稳定。开发版中四根手指轻点可打开原生 UI 检查器（显示元素、样式、帧时间与问题），效果同桌面按 F12。

## 速查

| 主题 | 命令 / API | 要点 |
|---|---|---|
| 环境检查 | `mygo doctor`、`mygo devices` | iOS 需 Xcode 的 Mac；Android 需 SDK+NDK+JDK 17+ |
| 开发 | `mygo dev -platform ios\|android [-device …]` | 模拟器/设备/AVD；Android 经 `adb reverse` 访问 `devUrl` |
| 状态保持 | `mygo.KeepState(&app{…})`、`App.OnSaveState`/`RestoredState` | 改动重启后回到原位置；后台结束也能恢复 |
| 构建 | `mygo build -platform ios[/simulator]\|android[/arm64]` | `.ipa` / `.app` / `.apk`+`.aab`；ABI：arm64、arm、amd64 |
| iOS 签名 | 已安装描述文件 + `ios.distribution`（默认 `app-store`） | 团队 API 密钥用 `MYGO_ASC_*` 上传 |
| Android 签名 | `android.keystore`、`MYGO_ANDROID_KEYSTORE_PASSWORD`、`MYGO_ANDROID_KEY_PASSWORD` | 无 keystore 时为调试签名，商店拒收 |
| 发布 | `mygo publish -platform ios\|android [-track beta]` | iOS：`altool` + `MYGO_ASC_KEY_ID/ISSUER_ID/KEY`；Android：`MYGO_PLAY_SERVICE_ACCOUNT` |
| 导航 | `mygo.NewTabs`、`TabItem`、`Window.Push/Pop`、`Tabs.Select`、`SetBadge` | 系统标签栏 + 屏幕栈；桌面退化为普通窗口 |
| 栈路由 | `Router.SystemStack`、`Route.Title`、`InSystemStack` | iOS UIKit 导航、Android Material 工具栏 |
| 多窗口 | `ios.multipleWindows`、`WindowOptions.OwnScene` | 仅 iPad 成为系统窗口 |
| 手势 | `PullToRefresh`、`InputZoom`、`FingerDown/Move/Up` | 命中 14 DIP、方向锁定、长按 0.5 秒 |
| 文本 | `TextInputClient`、`Keyboard`、`ReturnKey`、`Autofill` | 键盘选择与焦点移动按元素声明 |
| 控件 | `ui/mobile`、`mobile.SetLook`、`ui/netimage` | Sheet/ActionSheet/Drawer/PageView/SwipeRow/Zoomable 等 |
| 链接 | `ios.associatedDomains`、`android.appLinks` | 需 AASA / assetlinks.json |
| 图标启动 | `android.adaptiveIcon`、`ios.launchScreen`、`android.splash` | adaptive icon 432×432，内容放中间 72/108 |
| 插件 | `plugins/maps\|location\|camera\|purchases\|push\|…` | 导入才获得，用途说明与权限随应用 |
| 原生代码 | `mygo native new -lang swift,kotlin`、`mygo-plugin.json` 的 `mobile` | 模块优先；JSON 插件用 `mygo.Native(...).Call/On` |

## 文档未说明

- iOS/Android 最低支持的系统版本：文档只给出单项要求（如 sqlite 插件需 iOS 16），未列出整体下限。
- 手机端 webview 的远程调试步骤：文档未说明。
