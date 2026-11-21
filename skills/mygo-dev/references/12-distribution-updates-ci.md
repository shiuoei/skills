# 构建、分发与自动更新

定位：把 MyGo 项目从源码变成用户可安装、可自我更新的应用。覆盖 `mygo build` 的产物与标志、三平台打包与签名、安装器与安装脚本、`updates` 配置与签名密钥、增量更新与更新窗口、GitHub / S3 / 自建服务器发布，以及 tag 触发的 GitHub Actions 发布工作流。

术语沿用 MyGo 文档：产物写到配置里的 `out`（新项目为 `build`），每个平台一个子目录 `build/<os>-<arch>/`；需要纯 ASCII 的应用名是 `slug`；更新签名密钥对是 `mygo-update.key` / `mygo-update.pub`。

## 源页面

- 构建与分发 — <https://mygo.run/zh/docs/distribution>
- 自动更新 — <https://mygo.run/zh/docs/updates>
- GitHub Actions — <https://mygo.run/zh/docs/github-actions>

## `mygo build` 做什么

**用途**：一次构建完成编译、打包、签名和更新文件生成，输出到 `<out>/<os>-<arch>/`。

**关键行为**（按顺序）：① 有 Web 前端时写出 TypeScript 客户端、运行 `buildCommand` 把前端构建到 `frontendDist`；② 编译生产版本并嵌入前端（不带 Web 检查器，把配置里的名称、`identifier`、`version` 链接进去）；③ 为每个 `-platform` 目标打包到 `build/<os>-<arch>/`；④ 在支持签名的平台签名，配置了 `updates` 时生成已签名更新归档、各平台清单和基于已发布版本的增量更新。

```sh
# Web 前端项目
bun run build -- -platform darwin/universal,windows/amd64,linux/amd64
# 原生 UI 项目（CLI 由 go.mod 以工具方式固定，不需要 Bun）
go tool mygo build -platform windows/amd64
```

**关键 flag**：

| 标志 | 作用 |
|---|---|
| `-platform` | 逗号分隔的 `GOOS/GOARCH`，如 `darwin/universal,windows/amd64,windows/arm64,linux/amd64,linux/arm64`，或 `web`（`js/wasm`）；默认本机 |
| `-upload` | 用 `gh` 上传安装程序与更新到该版本的 GitHub 草稿发布，或上传到 `updates.s3` 的存储桶 |
| `-debug` | 保留开发功能，如 Web 检查器、原生 UI 检查器 |
| `-o` | 覆盖输出目录（默认配置中的 `out`） |
| `-sign` | 覆盖 macOS 签名身份（默认 `macos.signingIdentity`） |
| `-skip-build-command` | 不运行 `buildCommand` |
| `-skip-dmg` | 不制作 macOS 磁盘映像 |
| `-skip-notarize` | 不公证，即使设置了 `macos.notarize` |
| `-skip-installer` | 不制作 Windows 安装程序，也不下载 NSIS，也不写 `install.ps1` |

**各平台产物**（`build/<os>-<arch>/`）：macOS 为已签名的 `My App.app` 与 `my-app-0.1.0-arm64.dmg`（`darwin/universal` 为 `my-app-0.1.0.dmg`）；Windows 为 `My App.exe`、应用的文件与安装程序 `my-app-setup-0.1.0-amd64.exe`；Linux 为 `my-app`、`my-app.desktop`、`my-app.png`、文件类型对应的 `my-app.xml`、应用的文件、`my-app-0.1.0-linux-amd64.tar.gz`（含 `install.sh`），以及 `linux.formats` 选出的 `my-app_0.1.0_amd64.deb`、`my-app-0.1.0-1.x86_64.rpm`、`PKGBUILD` 与 `my-app-256x256.png`/`my-app-512x512.png`。配置了 `updates` 时还分别有 `install.sh` / `install.ps1`。

**限制/常见错误**：`darwin/universal` 把 arm64 与 amd64 合到一个应用里；MyGo 不需要 cgo，任何机器都能为每个平台编译，但 macOS 应用只在 macOS 上签名并放进磁盘映像，Windows 安装程序在其他系统上需要已安装 NSIS。CLI 默认 `CGO_ENABLED=0`，依赖需要 cgo 时须为 `mygo dev`、`mygo build`、`mygo generate` 设 `CGO_ENABLED=1` 并准备目标平台的 C 编译器与 `CC`。`-upload` 在配置了 `updates` 但缺少签名私钥时会在构建前停下。

## 应用名称、图标与版本（`mygo.json` 字段）

**用途**：这些字段进入应用包、可执行文件、菜单、安装程序、系统注册信息和更新清单，是分发的身份来源。原生 UI 项目用 `mygo.json`；Web 前端项目用 `mygo.config.ts`，字段相同，一个项目只能有其中一个。

**关键字段**：

```json
{
  "name": "My App",
  "slug": "my-app",
  "identifier": "com.example.myapp",
  "version": "1.2.0",
  "icon": "resources/icon.png",
  "resources": ["third_party/licenses", "bin/helper"],
  "resourcesExclude": ["**/*.psd", "design"],
  "urlSchemes": ["myapp"],
  "fileAssociations": [
    { "ext": ["md"], "name": "Markdown Document", "mimeType": "text/markdown" }
  ]
}
```

| 字段 | 默认值 | 说明 |
|---|---|---|
| `name` | 目录名 | 用户看到的名字：应用、可执行文件、菜单、安装程序；Windows 显示名可用 `windows.displayName` 单列 |
| `slug` | `name` 的小写连字符形式（`My App` → `my-app`；中文按普通话读音，`云音乐` → `yun-yin-le`） | 应用文件用的纯 ASCII 名：macOS 磁盘映像、Windows 安装程序、Linux 可执行文件/桌面条目/包名与 `/opt` 目录、更新归档、安装脚本变量。发布后不要改 |
| `identifier` | `com.mygo.` + 名称字母数字 | 反向 DNS 唯一名，系统据此索引偏好设置、权限和注册信息；初始化后不要改 |
| `version` | `0.1.0` | `App.Version()` 返回它，更新也用它比较 |
| `copyright` / `publisher` / `license` | — | 版权写入 `Info.plist` 与 Windows 文件属性；`publisher` 省略 © 与年份，是“设置 > 应用”的发布者；`license` 是 SPDX 表达式 |
| `icon` | 存在时为 `resources/icon.png` | 正方形 PNG，理想 1024×1024；旁边带 `-dev` 的同名文件（`icon-dev.png`）是开发版图标，不随应用发布 |

**图标转换**：`mygo build` 把 PNG 做成 macOS 的 `.icns`、Windows 可执行文件的图标资源和 Linux 桌面条目的图标。macOS 还可指定 `macos.icon`：`Assets.car`（actool 编译出的资源目录，图标名 `AppIcon`，或 `macos.infoPlist` 里的 `CFBundleIconName`）或 Icon Composer 的 `.icon`（需要 Xcode 26+ 的 `actool`；其他系统或编译失败时跳过并提示，仍用 PNG 生成的 `.icns`）；`mygo dev` 显示 `.icns`，不用 `.icon`。没有 Xcode 的 CI 上，把 `actool` 产出的 `Assets.car` 提交进仓库：

```sh
mkdir -p icon
xcrun actool AppIcon.icon --compile icon --platform macosx --target-device mac \
  --minimum-deployment-target 12.0 --app-icon AppIcon \
  --output-partial-info-plist icon/partial.plist   # 写出 icon/Assets.car
```

```ts
export default defineConfig({ macos: { icon: "icon/Assets.car" } });
```

**限制/常见错误**：`.icon` 必须命名为 `AppIcon.icon`，`actool` 按文件名命名图标，`--app-icon` 不一致时不会编译任何内容；平台目标名用 `darwin`/`windows`/`linux` 加架构，写成 `darwin-x64`、`macos` 之类的平台目录会报错。

## 资源与平台目录

**用途**：让运行时需要读取的文件（数据库初始数据、辅助程序、托盘图标、原生库）随应用一起发布。

**关键规则**：

- `resources/` 中的内容全部复制进应用，Go 源文件除外；配置里 `resources` 列出的文件和目录也复制，保留各自的基名。运行时用 `App.Path(mygo.PathResources)` 取目录：macOS 应用是 `Contents/Resources`，其他地方是可执行文件所在目录，`go run` / `go test` 下是项目的 `resources/`。
- `resourcesExclude` 按 Go `path.Match` 的段规则匹配：`**` 匹配任意层目录，`*.psd` 只匹配顶层；`resources` 里显式列出的路径不受排除影响。`resources/` 顶层以点开头的条目与隐藏目录（`.git`、`.cache`）不发布，`resources` 内部的点文件会发布。
- 平台目录布局与 `resources` 相同，目录名取 `-platform` 目标：`darwin/`、`darwin-arm64/`、`windows-amd64/`、`linux-arm64/` 等。构建把当前目标的平台资源与通用资源合并；两个目录安装到同一路径会失败。
- `darwin/universal` 带上 `darwin` 与 `darwin-universal`，并合并 `darwin-arm64` 与 `darwin-amd64`：一边有的文件另一边也必须有，程序和库合并为通用二进制（如 `lipo` 产物），其他文件必须完全一致。
- `mygo dev` 只带本机平台目录；`go run` / `go test` 下 `PathResources` 就是 `resources/` 本身。

```go
dir, _ := mygo.App.Path(mygo.PathResources)
seed := filepath.Join(dir, "seed.db")
cmd := exec.Command(filepath.Join(dir, "bin", "server"))
```

**辅助可执行文件**：放到对应平台目录，从 `PathResources` 运行；Windows 上同一路径 `exec.Command` 能找到 `.exe`。构建会为它们签名（macOS 每个 Mach-O 与包含它的包由内向外签；Windows 无签名的可执行文件与 DLL 在配置了 `certificate` 或 `signCommand` 时签名，已由发行者签名的保留原签名）。强化运行时下需要授权的辅助程序用 `macos.helperEntitlements` 按资源路径指定，如 `{ "bin/server": "server.entitlements.plist" }`。

**包的原生库**：用 purego 加载原生库的包在源码旁 `mygo-plugin.json` 的 `libraries[].files` 中按平台写明 `name`、`url`、`sha256`；`mygo build` 与 `mygo dev` 下载到 `<cache>/mygo/natives/<sha256>/`，校验后安装到应用资源顶层并签名。为没有对应文件的平台构建会失败。

## macOS：应用包、磁盘映像、签名与公证

**用途**：产出一个可双击安装、能通过 Gatekeeper 的应用。

**关键配置**：

```ts
export default defineConfig({
  macos: {
    minimumSystemVersion: "13.0",
    signingIdentity: "Developer ID Application: Jane Doe (TEAMID)",
    notarize: { keychainProfile: "notary", keychain: process.env.NOTARY_KEYCHAIN },
    infoPlist: { NSCameraUsageDescription: "Scan documents with the camera." },
    entitlements: "entitlements.plist",
    helperEntitlements: { "bin/server": "server.entitlements.plist" },
    dmgTitle: "My App Installer",
  },
});
```

- **签名**：默认 ad hoc（`-`），只够在构建它的 Mac 上运行，分发会被 Gatekeeper 拦截；分发用 Apple Developer Program 的 Developer ID 身份。
- **公证**：`notarize.keychainProfile` 指向 `xcrun notarytool store-credentials notary` 存入钥匙串的凭据，`notarize.keychain` 指定存放它的钥匙串（CI 用临时钥匙串）。构建启用强化运行时，先提交应用到公证服务并附票据，再对磁盘映像做同样的事；更新归档里的应用也带票据。
- **描述文件**：`provisioningProfile` 在签名前复制到 `Contents/embedded.provisionprofile`。必须与签名证书、包标识符、授权（含应用与团队标识符）匹配；不能与 ad hoc 签名一起用；MyGo 不创建描述文件、也不从它推导授权。
- `minimumSystemVersion` 默认 `12.0`；`infoPlist` 添加或替换键；`entitlements` 给应用签名，`helperEntitlements` 给辅助程序签名。

```sh
xcrun notarytool store-credentials notary
security find-identity -v -p codesigning   # 列出签名身份
mygo doctor                                # 统计 Developer ID 数量
```

**限制/常见错误**：

- `codesign` 缺什么就在本机屏幕上弹框要（锁屏钥匙串密码、使用该身份密钥的许可），无人值守构建（锁屏后、SSH、CI）会停在那里；15 秒后 `mygo build` / `mygo dev` 会提示 `codesign` 可能正在等待。先在构建前解锁钥匙串并允许使用密钥：
  ```sh
  security unlock-keychain ~/Library/Keychains/login.keychain-db
  security set-key-partition-list -S apple-tool:,apple: -s ~/Library/Keychains/login.keychain-db
  ```
- `mygo dev` 默认用 Apple Development 身份签名，没有则 ad hoc；开发版应用标识符是 `<identifier>.dev`。描述文件通常不授权这个标识符，用计算配置把 `provisioningProfile` 与受限授权排除在 `dev` 之外：
  ```ts
  export default defineConfig(({ command }) => ({
    macos: {
      signingIdentity: process.env.MACOS_SIGNING_IDENTITY,
      ...(command === "build" && {
        provisioningProfile: process.env.MACOS_PROVISIONING_PROFILE,
        entitlements: "entitlements.plist",
      }),
    },
  }));
  ```
- 描述文件不要进版本控制，通过构建环境提供路径。

## Windows：可执行文件、NSIS 安装程序与签名

**用途**：产出带图标、版本信息和应用清单的可执行文件，以及当前用户级、免管理员、能被应用自我更新的安装程序。

**关键配置与行为**：

```ts
export default defineConfig({
  name: "weather",
  windows: {
    displayName: "天气助手",
    installDirName: "ZhiLian",
    certificate: "certs/code-signing.pfx",
    signCommand: "signtool sign /fd sha256 /tr http://timestamp.digicert.com /td sha256 /a %1",
    installerInclude: "build/installer.nsh",
  },
});
```

- 可执行文件内嵌图标、版本信息与应用清单（逐显示器 DPI 感知、现代控件），运行时不显示控制台窗口；可在主包放自己的 `.syso` 替换。
- 安装程序名 `my-app-setup-1.2.0-amd64.exe`，取自 `slug` 并带架构。安装到 `%LOCALAPPDATA%\Programs\<installDirName>`，添加开始菜单快捷方式与“设置 > 应用”卸载项，注册 URL 协议和文件关联；完成页可运行应用、创建桌面快捷方式（默认勾选），静默安装 `/S` 不创建桌面快捷方式。
- `windows.displayName` 控制快捷方式、应用条目、安装程序标题与产品名；`windows.installDirName` 只控制安装文件夹名（可为只认 ASCII 路径的工具用 ASCII 目录）。改 `installDirName` 只影响新安装，已安装的应用留在原文件夹。
- 安装程序用 NSIS 制作：Windows 上发现已安装就用，否则第一次需要时下载官方 NSIS 3.13（GitHub releases 副本，缺失时取 SourceForge），按 SHA-256 校验后存到 `%LOCALAPPDATA%\mygo`；macOS / Linux 上需先装好（`brew install makensis`、`apt install nsis`），否则跳过安装程序。
- 配置了 `certificate` 或 `signCommand` 时，可执行文件、资源中未签名的可执行文件与 DLL、安装程序及其卸载程序都被签名。`.pfx` 密码来自 `MYGO_WINDOWS_CERTIFICATE_PASSWORD`；Windows 上用 SDK 的 `signtool`，其他系统用 `osslsigncode`；`timestampUrl` 默认 `http://timestamp.digicert.com`。私钥在硬件令牌或云服务（如 Azure Trusted Signing）时用 `signCommand`，`%1` 是文件路径。

**安装程序钩子**：`windows.installerInclude` 指向一个 NSIS 脚本，安装程序在 MUI2.nsh 及其语言之后包含它。`customInit` 在 `.onInit` 运行（安装任何东西之前），`customInstall` 在安装结束、文件已在 `$INSTDIR` 时运行，`customUnInstall` 在卸载时、移除 `$INSTDIR` 之前运行。脚本未定义的宏被跳过；脚本不存在时构建停止。

```nsis
; build/installer.nsh
!macro customInstall
  CreateDirectory "$LOCALAPPDATA\My App\bin"
  CopyFiles /SILENT "$INSTDIR\bin\helper.exe" "$LOCALAPPDATA\My App\bin"
!macroend

!macro customUnInstall
  Delete "$LOCALAPPDATA\My App\bin\helper.exe"
  RMDir "$LOCALAPPDATA\My App\bin"
!macroend
```

**限制/常见错误**：钩子在静默安装/卸载（`/S`）中也会运行，`IfSilent` 可区分；消息框需要 `/SD` 才能自动应答；`customUnInstall` 调用的函数属于卸载程序，以 `un.` 命名。安装程序只为当前用户安装，钩子也只应操作用户可写位置（`$LOCALAPPDATA`、`HKCU`）；卸载程序只移除应用文件夹，保留应用数据；添加钩子不影响安装程序与卸载程序的签名。未签名的应用会让 SmartScreen 警告；`SHA256SUMS` 未签名，只能检测下载损坏，不能防恶意替换。`-skip-installer` 同时不写 `install.ps1`，而它负责运行安装程序，一行安装需要发布版本里同时提供两者。

## Linux：可执行文件、包与 PKGBUILD

**用途**：产出桌面条目齐全、能通过 Debian/RPM 包管理器安装、也能用户级安装并自我更新的应用。

```ts
export default defineConfig({
  linux: {
    maintainer: "Jane Doe <jane@example.com>",
    comment: "Take notes",
    categories: ["Office"],
    command: ["my-app", "ma"],
    webview: false,
    depends: { deb: ["libayatana-appindicator3-1"], rpm: ["libayatana-appindicator-gtk3"], arch: ["libayatana-appindicator"] },
    formats: ["deb", "rpm", "pkgbuild"],
  },
});
```

- `formats` 取 `deb`、`rpm`、`pkgbuild`，默认三种全有；不认识的名称报错。带 `install.sh` 的归档总会写出（`install.sh` 从它更新应用，PKGBUILD 也打包它），`formats: []` 只写归档。`SHA256SUMS` 与 `-upload` 只处理构建实际写出的内容。
- Debian 与 RPM 包安装到 `/opt/<slug>`，安装桌面条目、图标、注册 URL 协议与文件类型，依赖 GTK 3 与 WebKitGTK；`depends` 按 `deb`/`rpm`/`arch` 追加依赖，或给一个 Debian 包列表由 MyGo 映射。`webview: false` 时不再依赖 WebKitGTK，`install.sh` 与 `mygo doctor` 也不再检查它。
- `command` 给应用命令：包内 `/usr/bin/<command>`、`install.sh` 的 `~/.local/bin/<command>`；可给多个名字。不设置时不加命令，把这个名字留给同名命令行工具。桌面条目的 `StartupWMClass` 用 `command` 的可执行文件名。
- `maintainer` 默认取 package.json 的 `author`，否则取应用名；`comment` 描述应用；`categories` 默认 `["Utility"]`。

**Arch Linux**：每次 Linux 构建还得到 `PKGBUILD`（除非 `formats` 排除），用它把旁边的归档打成包：

```sh
cd build/linux-amd64
makepkg -si
```

PKGBUILD 字段来源：`pkgname` = `slug`；`pkgver` 去掉连字符（`1.2.0-beta.1` → `1.2.0beta.1`，`1.2.0-1` → `1.2.0pre1`，`1.2.0+build.5` → `1.2.0_build.5`）；`pkgdesc` = `linux.comment`，否则名称；`arch` 为 `x86_64` / `aarch64`；`url` 取 package.json `homepage`，否则 `updates.github` 仓库；`license` 取 `mygo.json` 的 `license`，否则 package.json 的，否则 `LicenseRef-custom`；`depends` 为 `gtk3`，除非 `webview: false` 还有 `webkit2gtk-4.1`，再加 `linux.depends.arch`。图标用 PKGBUILD 旁的 `my-app-256x256.png`、`my-app-512x512.png`。`-upload` 不带 PKGBUILD（每个架构各一份，打包的是旁边的归档）。

**限制/常见错误**：已发布的应用不要改 `slug`。Linux 文件以 slug 命名、更新替换同名文件，换 slug 会把旧可执行文件留在原处，`install.sh` 注册的桌面条目仍会运行它，而新 slug 的 Debian/RPM 包是另一个包、与旧的并排安装。首次发布前就选好 ASCII slug。

## 安装脚本

**用途**：在 macOS、Windows、Linux 上各用一行命令安装应用，并让 Linux 用户级安装能自我更新。

- **Linux**：每次构建都生成 `my-app-<version>-linux-amd64.tar.gz` 与 `install.sh`；配置了 `updates` 时，`install.sh` 从当前机器对应更新清单下载最新版本，并在安装前像更新器一样校验大小与 Ed25519 签名（`openssl pkeyutl -rawin`，需 OpenSSL 3.0+；没有该命令时脚本不下载任何东西并说明，此时可把归档直接交给脚本、不做校验）。脚本安装到 `~/.local/<slug>.app`，安装桌面条目、图标，注册 URL 协议与文件类型；`linux.command` 存在时在 `~/.local/bin` 建命令链接，但不会占用其他程序已有的命令名。**macOS**：配置了 `updates` 时同一个 `install.sh` 下载 Mac 架构（无则 `darwin-universal`；Apple silicon 上即使 shell 跑在 Rosetta 下也取 `darwin-arm64`）清单指定的归档，用 `openssl` + `shasum` + JavaScript for Automation 完成 Ed25519 校验；若应用在运行先退出，替换 `/Applications`（不可写或原本就在 `~/Applications` 则替换那里）后打开应用，从不请求 root，同名但包标识符不同的应用不会被碰，`curl` 不设隔离标记所以没有 Gatekeeper 对下载应用的提示。**Windows**：构建得到 `install.ps1`，在 Windows PowerShell 5.1 与 7 中运行；下载本机架构（`amd64` / `arm64`）的最新安装程序，对照 `SHA256SUMS` 校验 SHA-256，应用在运行则先退出，在已安装版本的文件夹中静默运行安装程序（`/S`），再打开应用；卸载仍从“设置 > 应用”操作。

```sh
curl -fsSL https://github.com/me/my-app/releases/latest/download/install.sh | sh
```

```powershell
irm https://github.com/me/my-app/releases/latest/download/install.ps1 | iex
```

```sh
sh install.sh my-app-1.2.0-linux-amd64.tar.gz   # 安装指定归档
sh install.sh --uninstall                        # 移除应用，保留设置和数据
```

**限制/常见错误**：

- `tagPrefix` 不是 `v` 时，任何发布版本中的脚本（如 `…/releases/download/desktop-v1.2.0/install.sh`）都会安装带该前缀标签的最新发布版本；再次运行 `install.sh` 会覆盖安装它找到的版本；指向旧名称的链接（slug 除外）会保留，`--uninstall` 移除 `~/.local/bin` 中所有指向应用的链接。
- `install.sh` 不安装 WebKitGTK；系统库缓存里没有时，安装后说明并给出 Debian/Ubuntu/Fedora/Arch/openSUSE 的安装命令（`linux.webview: false` 时除外）。
- 带更新的构建会写出 `SHA256SUMS`（安装程序、归档和脚本的 SHA-256，`sha256sum` 格式）；`-upload` 随它们一起发布，并合并同一发布版本 / 同一 `updates.url` 服务器里其他平台早先上传的行。GitHub 会把上传文件名中的空格变成点，`install.ps1` 遵循这一点。
- 测试用环境变量：`MY_APP_DOWNLOAD_URL`（应用 `slug` 的大写形式）、`MY_APP_INSTALL_DIR`（macOS 上应用的文件夹）、`MY_APP_NO_OPEN=1`（保持应用关闭）。

## 自动更新：密钥、配置与产物

**用途**：应用从 GitHub 或任意 HTTPS 服务器下载已签名的新版本，用编译进应用的公钥验证签名后替换自己。

**关键步骤**：

```sh
mygo keygen            # 私钥 mygo-update.key、公钥 mygo-update.pub
mygo keygen -o ~/keys  # 另一个目录；-force 替换现有密钥
```

默认写到用户配置目录中的 `mygo/update-keys`，并打印位置与要加入配置的 `updates` 部分。私钥放在仓库之外，存进密码管理器和 CI 机密：已安装的应用只接受用它签名的更新，丢失它会让这些应用无法更新，而持有它的人能向用户发布代码。

```ts
export default defineConfig({
  version: "1.2.0",
  updates: {
    publicKey: "…the contents of mygo-update.pub…",
    github: "you/my-app",
    tagPrefix: "desktop-v",
    privateKey: "~/keys/mygo-update.key",
    changelog: "CHANGELOG.md",
    deltas: 3,
  },
});
```

`mygo.json` 中是同样字段的 JSON 写法。

| 字段 | 默认值 | 说明 |
|---|---|---|
| `publicKey` | 必需 | `mygo-update.pub` 的内容，编译进应用用于验签 |
| `github` | 与 `url` 二选一 | 公开仓库 `owner/name`，更新放在它的发布版本里，标签如 `v1.2.0` |
| `tagPrefix` | `v` | 标签中版本前的部分；用 `desktop-v` 时应用查带此前缀的最新正式版本，跳过草稿与预发布。每次检查一个请求，受 API 每 IP 每小时 60 次限制 |
| `url` | 与 `github` 二选一 | 存放更新的目录的 HTTPS URL（自建服务器或 S3/CDN） |
| `s3` | — | 为 `url` 提供内容的存储桶：`bucket`，可选 `prefix`、`region`、`endpoint`、`pathStyle` |
| `privateKey` | — | `mygo-update.key` 的路径（可 `~/` 开头）；环境变量 `MYGO_UPDATER_PRIVATE_KEY` 优先 |
| `changelog` | 存在时为 `CHANGELOG.md` | 发布说明来源 |
| `deltas` | `3` | 多少个较早版本获得增量更新；`0` 表示不生成 |

```sh
MYGO_UPDATER_PRIVATE_KEY="$(cat path/to/mygo-update.key)" bun run build
MYGO_UPDATER_PRIVATE_KEY="$(cat path/to/mygo-update.key)" go tool mygo build
```

**生成的更新文件**（每个平台一套）：归档 `my-app-1.2.0-darwin-arm64.tar.gz`（与安装后相同的应用文件，以 `slug` 命名）；增量 `my-app-1.1.0-to-1.2.0-darwin-arm64.delta`；清单 `update-darwin-arm64.json`，含版本、发布说明、日期，以及归档和每个增量更新的 URL、大小与签名。

**限制/常见错误**：没有私钥时 `mygo build` 只构建应用、不写更新文件（Linux 归档例外，`install.sh` 要从它安装）；`mygo build -upload` 需要密钥并在构建前停下，因为已安装的应用无法更新到没有清单的发布版本。每个平台（`darwin-arm64`、`darwin-universal`、`windows-amd64`…）各有清单，应用只读自己平台的那份。`mygo build` 会优先使用环境中已有的 `GH_TOKEN` 或 `GITHUB_TOKEN`。

## 发布说明、预发布与增量更新

**发布说明**：取 `CHANGELOG.md`（或 `updates.changelog`）中该版本的章节，在 `## 1.2.0` 标题下，或 Keep a Changelog 风格的 `## [1.2.0] - 2026-09-19`。文件存在时必须包含该章节，否则构建失败——这是最常见的发布阻塞点。

**预发布**：GitHub 上标记为预发布的版本不会成为“最新发布版本”，已安装应用默认跳过。提供 beta 通道的应用调用 `mygo.Updater.SetPrereleases(settings.Beta)`，之后按 GitHub API 找非草稿中版本号最高的版本，语义化版本比较（`1.3.0-beta.2` < `1.3.0-beta.10` < `1.3.0`）。关闭预发布不会降级；选项只在本进程有效，应用要自己持久化并在启动时恢复。发布测试版本务必标为预发布，如 `gh release edit v1.3.0-beta.1 --draft=false --prerelease`。

**增量更新**：`mygo build` 为更新签名时读取每个已发布平台的清单，下载最近 3 个版本的归档（`updates.deltas` 改数量，`0` 不生成），校验签名后用 [bsdiff](https://www.daemonology.net/bsdiff/) 生成到新版本的增量：未更改的文件从已安装应用复制，更改的打补丁，其余完整包含。macOS 公证票据（`Contents/CodeResources`）总是完整包含；不小于归档的增量被省略；下载不到已发布更新（如第一个版本）就不生成并说明原因。最新 GitHub 发布版本缺少正在构建平台的清单时，构建按 `tagPrefix` 在更早发布版本中查找（跳过草稿与预发布）。应用侧校验签名后以已安装文件为基础在应用目录旁生成新版本，逐个校验文件大小与 SHA-256（含代码签名），任何一步失败（如安装后文件被改过）则改下载完整归档。

## 更新窗口与 `mygo.Updater` API

**用途**：给用户一个 Mac 风格的更新界面，或在自己的 UI 里做更新入口。

**插件方式**（推荐，三平台都有）：`mygo.Use(updater.Plugin)`（`github.com/egoist/mygo/plugins/updater`）。插件提供后台检查、显示发布说明、安装 / 跳过 / 稍后提醒、“检查更新…”菜单项、保存的偏好设置与多语言支持；不显示网页的原生 UI 应用用 `native.Plugin` 得到同样窗口。

**自行实现**：直接使用 `mygo.Updater`。

```go
var UpdateProgress = mygo.NewEvent[float64]("update-progress")

func checkForUpdates(ctx context.Context) {
	if !mygo.Updater.Enabled() {
		return
	}
	up, err := mygo.Updater.Check(ctx)
	if err != nil || up == nil { // offline, or up to date
		return
	}
	res, _ := mygo.Dialog.Message(mygo.MessageOptions{
		Message: "Version " + up.Version + " is available.",
		Detail:  up.Notes, // Markdown
		Buttons: []string{"Install and Restart", "Later"},
	})
	if res.Button != 0 {
		return
	}
	err = up.Install(ctx, func(downloaded, total int64) {
		UpdateProgress.Broadcast(float64(downloaded) / float64(total))
	})
	if err == nil {
		mygo.App.Relaunch()
	}
}
```

- `Check` 返回较新版本（`Version`、`Notes`、`Date`），已最新时返回 nil；`Install` 下载归档或增量、验签并替换应用，进度回调先给增量字节，增量失败改从头用归档字节。运行中的应用不受影响，新版本在 `App.Relaunch()` 之后或下次启动时运行。
- macOS 在应用退出后替换应用包（Dock 保留图标）；Linux 与 Windows 替换应用文件夹中的文件，并移除运行中版本装过、新版本不再有的文件（清单是文件夹里的 `.mygo-files`；更早构建的应用没有该列表，第一次更新不会移除任何东西）。任何版本都未安装的文件（如便携副本旁边的数据）与应用正在运行的可执行文件会保留。
- `Enabled` 表示应用能否自我更新：构建时必须启用更新，且安装位置可写。开发构建不支持自我更新，`Check` 返回 `mygo.ErrUpdatesDisabled`。`Check` 与 `Install` 走系统代理；设置了 `HTTPS_PROXY` 时走该环境变量指定的代理。
- 自定义更新服务用 `mygo.Updater.SetSource`，实现 `Manifest(ctx, prereleases)` 与 `Open(ctx, url)`；更新器仍用 `updates.publicKey` 校验签名，数据源提供的仍是 `mygo build` 的产物，构建时依然要配置 `updates`（含公钥与 `github`/`url`）。`SetSource(nil)` 恢复默认。
- **应用能自我更新的位置**：应用要能写自己的安装位置——macOS 应用包（`/Applications` 或 `~/Applications`）、Windows 安装程序的当前用户安装 `%LOCALAPPDATA%\Programs`、Linux 用户目录（`~/.local/my-app.app` 或解压后的构建目录）。由包管理器安装的应用（如 `/opt` 的 Debian 包）由包管理器更新，`Enabled` 为 false。

## 发布位置：GitHub、S3、自建服务器

**GitHub**：配置 `updates.github` 后，`mygo build -upload` 把磁盘映像、安装程序、包、Linux 归档、安装脚本与更新文件上传到该版本的草稿发布（标签 `v1.2.0`），检查草稿后手动发布；需要已登录的 [GitHub CLI](https://cli.github.com)（`gh`）。正式发布后应用从 `https://github.com/you/my-app/releases/latest/download/` 读取清单。

**S3 / Cloudflare R2**：

```ts
export default defineConfig({
  updates: {
    publicKey: "…",
    url: "https://downloads.example.com/my-app",
    s3: { bucket: "downloads", prefix: "my-app",
          endpoint: "https://<account>.r2.cloudflarestorage.com", pathStyle: false },
  },
});
```

`-upload` 把安装程序、归档、增量与安装脚本上传到该存储桶，最后上传清单。凭据来自 `AWS_ACCESS_KEY_ID`、`AWS_SECRET_ACCESS_KEY`，临时凭据还有 `AWS_SESSION_TOKEN`；区域取 `s3.region`，否则 `AWS_REGION`（或 `AWS_DEFAULT_REGION`），否则 `us-east-1`。清单、安装脚本和 `SHA256SUMS` 名称跨版本不变，以 `Cache-Control: no-cache` 上传，让 CDN 重新获取。Amazon S3 把桶名放主机名，用 `endpoint` 时放路径；只接受主机名桶名的服务把 `pathStyle` 设为 `false`。

**自建服务器**：把归档、增量、清单上传到 `url` 目录，连同 `install.sh`、`install.ps1`、`SHA256SUMS`（每平台一行），保留较早版本文件；**最后上传清单**，这样应用不会看到文件缺失的清单。带 `tagPrefix` 时保留其他产品的“最新发布版本”标记：用 `gh release edit desktop-v1.2.0 --draft=false --latest=false` 发布，或取消勾选“设为最新发布版”。

## GitHub Actions 发布工作流

**用途**：推送 tag 时在每个平台构建、签名，把产物与更新上传到同一个草稿发布，检查后发布。

**关键步骤**：`push.tags: ["v*"]` 触发；`draft-release` 作业先建草稿 `v1.2.0`；`build` 作业按平台矩阵并行、各自 `-upload` 到该草稿；`publish` 作业在所有构建后取消草稿。MyGo 不需要 cgo，Ubuntu 运行器既能构建 Linux 也能构建 Windows；macOS 应用需要 macOS 运行器（签名、公证、磁盘映像）。

```yaml
name: Release
on:
  push:
    tags: ["v*"]
permissions:
  contents: write            # 草稿发布及其上传
jobs:
  draft-release:
    runs-on: ubuntu-latest
    steps:
      - env: { GH_TOKEN: "${{ github.token }}" }
        run: |
          gh release view "$GITHUB_REF_NAME" >/dev/null 2>&1 ||
            gh release create "$GITHUB_REF_NAME" --draft --verify-tag \
              --title "$GITHUB_REF_NAME" --generate-notes
  build:
    needs: draft-release
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false
      matrix:
        include:
          - { name: macOS,   os: macos-latest, platform: "darwin/arm64,darwin/amd64" }
          - { name: Windows, os: ubuntu-latest, platform: "windows/amd64,windows/arm64" }
          - { name: Linux,   os: ubuntu-latest, platform: "linux/amd64,linux/arm64" }
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7
        with: { go-version-file: go.mod }
      - uses: oven-sh/setup-bun@v2
      - run: bun install --frozen-lockfile
      - name: Install NSIS
        if: matrix.name == 'Windows' && runner.os == 'Linux'
        run: sudo apt-get update && sudo apt-get install -y nsis
      - name: Build and upload
        env:
          GH_TOKEN: ${{ github.token }}
          MYGO_UPDATER_PRIVATE_KEY: ${{ secrets.MYGO_UPDATER_PRIVATE_KEY }}
        run: bun run build -- -platform "${{ matrix.platform }}" -upload
```

原生 UI 项目没有前端、也就没有 Bun：省略 `setup-bun` 与 `bun install`，改用 `go tool mygo build -platform "${{ matrix.platform }}" -upload`。把 `mygo-update.key` 的内容存进仓库机密：`gh secret set MYGO_UPDATER_PRIVATE_KEY < path/to/mygo-update.key`；发布前把配置里的 `version` 改成新版本、提交，再 `git tag v1.2.0 && git push origin v1.2.0`。

**要点与限制**：

- `mygo build` 按配置里的 `version` 与 `tagPrefix` 上传，与触发工作流的标签无关；用一步校验一致，否则会以另一个版本号发布：`run: test "$GITHUB_REF_NAME" = "v$(jq -r .version package.json)"`（原生 UI 项目改为 `jq -r .version mygo.json`）。
- 草稿在构建前就存在，所以各次构建都上传到它；没有草稿时每个 `mygo build -upload` 会各建一个草稿。再次运行工作流保留草稿并替换文件。
- `fail-fast: false` 让一个平台失败时其他平台继续完成；“重新运行失败的作业”只重建那一个，仍上传到同一草稿。
- 一行 `platform` 可列多个目标，`mygo build` 依次构建；每个目标单独一行会用更多运行器并行。
- 工作流的 `GITHUB_TOKEN` 让运行器自带的 `gh` 创建发布并上传；有 `tagPrefix` 时也让 `mygo build` 找到增量更新所基于的更早发布版本。
- 上述工作流可运行但签名不满足分发：macOS 是 ad hoc（其他 Mac 被 Gatekeeper 阻止），Windows 未签名（SmartScreen 警告）。

### macOS 签名与公证

需要 Developer ID 证书与 Apple 公证凭据。从“钥匙串访问”导出带密码的 `.p12`，添加机密 `MACOS_CERTIFICATE`（base64）、`MACOS_CERTIFICATE_PASSWORD`、`APPLE_ID`、`APPLE_TEAM_ID`、`APPLE_APP_PASSWORD`（[account.apple.com](https://account.apple.com) 的 App 专用密码），再在 `Build and upload` 之前导入到作业专用钥匙串（作业结束即消失），并把 `NOTARY_KEYCHAIN` 写进 `GITHUB_ENV`：

```yaml
- name: Import the certificate and the notary credentials
  if: runner.os == 'macOS'
  env:
    MACOS_CERTIFICATE: ${{ secrets.MACOS_CERTIFICATE }}
    APPLE_ID: ${{ secrets.APPLE_ID }}
    APPLE_TEAM_ID: ${{ secrets.APPLE_TEAM_ID }}
    APPLE_APP_PASSWORD: ${{ secrets.APPLE_APP_PASSWORD }}
  run: |
    keychain="$RUNNER_TEMP/signing.keychain-db"
    password="$(openssl rand -base64 24)"
    security create-keychain -p "$password" "$keychain"
    security set-keychain-settings -lut 21600 "$keychain"
    security unlock-keychain -p "$password" "$keychain"
    echo "$MACOS_CERTIFICATE" | base64 --decode > "$RUNNER_TEMP/certificate.p12"
    security import "$RUNNER_TEMP/certificate.p12" -k "$keychain" \
      -P "$MACOS_CERTIFICATE_PASSWORD" -T /usr/bin/codesign
    rm "$RUNNER_TEMP/certificate.p12"
    security set-key-partition-list -S apple-tool:,apple: -s -k "$password" "$keychain"
    security list-keychains -d user -s "$keychain" login.keychain
    xcrun notarytool store-credentials notary --keychain "$keychain" \
      --apple-id "$APPLE_ID" --team-id "$APPLE_TEAM_ID" --password "$APPLE_APP_PASSWORD"
    echo "NOTARY_KEYCHAIN=$keychain" >> "$GITHUB_ENV"
```

配置里用该钥匙串；自己机器上不设 `NOTARY_KEYCHAIN`，用登录钥匙串：

```ts
export default defineConfig({
  macos: {
    signingIdentity: "Developer ID Application: Jane Doe (TEAMID)",
    notarize: { keychainProfile: "notary", keychain: process.env.NOTARY_KEYCHAIN },
  },
});
```

也可用 App Store Connect API 密钥代替 Apple ID：用 `--key`、`--key-id`、`--issuer` 把密钥文件、ID、颁发者提供给 `notarytool store-credentials`。

### Windows 签名

Ubuntu 运行器上用 `osslsigncode` 给 `.pfx` 签名，机密为 `WINDOWS_CERTIFICATE`（base64）与 `WINDOWS_CERTIFICATE_PASSWORD`：

```yaml
- name: Install NSIS and osslsigncode
  if: matrix.name == 'Windows' && runner.os == 'Linux'
  run: sudo apt-get update && sudo apt-get install -y nsis osslsigncode
- name: Write the certificate
  if: matrix.name == 'Windows'
  run: |
    echo "${{ secrets.WINDOWS_CERTIFICATE }}" | base64 --decode > "$RUNNER_TEMP/code-signing.pfx"
    echo "WINDOWS_CERTIFICATE_FILE=$RUNNER_TEMP/code-signing.pfx" >> "$GITHUB_ENV"
```

构建步骤的环境里加入 `MYGO_WINDOWS_CERTIFICATE_PASSWORD: ${{ secrets.WINDOWS_CERTIFICATE_PASSWORD }}`，配置里用 `windows: { certificate: process.env.WINDOWS_CERTIFICATE_FILE }`。

私钥在硬件令牌或云服务（如 Azure Trusted Signing）时改用 `windows.signCommand`（通常需要 `signtool`），那就把 Windows 行换成 `os: windows-latest`（自带 `signtool`，`mygo build` 会自己下载 NSIS），并在 `Build and upload` 前按签名服务文档登录；Windows 运行器上的步骤默认在 PowerShell 中运行，除非设 `shell: bash`。

### 发布、S3 与无发布版本

所有作业上传完后自动发布（也可看过草稿后手动发布）：

```yaml
  publish:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - env: { GH_TOKEN: "${{ github.token }}" }
        run: gh release edit "$GITHUB_REF_NAME" --draft=false
```

- 用 `desktop-v` 之类前缀时：`tags: ["desktop-v*"]`，校验标签为 `desktop-v` + 版本号，并用 `--latest=false` 发布。预发布标签（如 `v1.3.0-beta.1`）要加 `--prerelease`。
- 用 `updates.s3` 时没有草稿发布：省略创建草稿的作业与构建作业的 `needs`，把 `AWS_ACCESS_KEY_ID`、`AWS_SECRET_ACCESS_KEY` 代替 `GH_TOKEN` 提供给作业，也不需要 `contents: write`。
- 没有 `updates` 配置时 `-upload` 无处可传：构建不加 `-upload`，用 `actions/upload-artifact` 保留 `build/*/*.dmg`、`build/*/*-setup-*.exe`、`build/*/*.deb`、`build/*/*.rpm`、`build/*/*.tar.gz`；每次推送都可这样检查应用能否构建。
- Windows 运行器每次都要下载 NSIS，可用 `actions/cache` 缓存 `%LOCALAPPDATA%\mygo\nsis-*`，键用 `hashFiles('bun.lock')`（原生 UI 项目用 `hashFiles('go.sum')`）。
- 相关环境变量：`CGO_ENABLED`（默认 `0`，依赖需要 cgo 时设 `1`）、`MYGO_LDFLAGS`（追加 `-ldflags`，如 `-X pkg.Var=…`）、`GOFLAGS`（其中 `-tags=…` 会并入构建标签，如 `-tags=mygo_nowebview`）、`MYGO_INSPECTOR=1`（生产构建保留原生 UI 检查器）、`MYGO_ENV=production`。

## 速查

| 主题 | 命令 / 字段 | 要点 |
|---|---|---|
| 构建 | `bun run build -- -platform …` / `go tool mygo build` | 产物在 `<out>/<os>-<arch>/`，新项目 `out` 为 `build` |
| 平台目标 | `-platform` | `darwin/universal`、`windows/amd64`、`linux/arm64`、`web`；默认本机 |
| 跳过步骤 | `-skip-dmg` / `-skip-notarize` / `-skip-installer` / `-skip-build-command` | `-skip-installer` 也不写 `install.ps1` |
| 身份 | `name` / `slug` / `identifier` / `version` | `identifier` 与已发布 `slug` 不要再改 |
| 图标 | `resources/icon.png`、`icon-dev.png`、`macos.icon` | 正方形 PNG，理想 1024×1024 |
| 资源 | `resources/`、`resources`、`resourcesExclude`、`App.Path(mygo.PathResources)` | 平台目录按 `-platform` 命名；Go 源文件不打包 |
| macOS 分发 | `macos.signingIdentity`、`notarize`、`provisioningProfile` | ad hoc 只够本机；公证票据与描述文件随包 |
| Windows 安装 | `windows.displayName`、`installDirName`、`certificate`、`installerInclude` | 装到 `%LOCALAPPDATA%\Programs`，当前用户、免管理员 |
| Windows 签名 | `certificate` + `MYGO_WINDOWS_CERTIFICATE_PASSWORD`、`signCommand`、`timestampUrl` | `%1` 为文件；默认时间戳 `http://timestamp.digicert.com` |
| Linux 包 | `linux.formats`、`depends`、`command`、`webview` | 包安装到 `/opt/<slug>`；归档 + `install.sh` 用于自我更新 |
| 更新密钥 | `mygo keygen`、`MYGO_UPDATER_PRIVATE_KEY`、`updates.privateKey` | 私钥不进仓库；丢失后已安装应用无法更新 |
| 更新配置 | `updates.publicKey` +（`github` 或 `url`） | `tagPrefix`、`s3`、`changelog`、`deltas` |
| 更新产物 | `<slug>-<ver>-<platform>.tar.gz`、`*.delta`、`update-<platform>.json` | 每平台一份清单；清单最后上传 |
| 发布说明 | `CHANGELOG.md` 的 `## <version>` 章节 | 文件存在但缺该章节则构建失败 |
| 发布 | `-upload`、`gh release edit … --draft=false` | GitHub 草稿；S3 直接上传；`--prerelease` 标记测试版本 |
| 应用 API | `mygo.Updater.Check/Install/Enabled/SetPrereleases/SetSource` | 开发构建返回 `mygo.ErrUpdatesDisabled` |
| 更新窗口 | `updater.Plugin` / `native.Plugin` | Sparkle 式窗口，三平台可用 |
| CI | tag 触发 + 草稿作业 + 平台矩阵 + `-upload` | Ubuntu 可构建 Windows/Linux；macOS 需 macOS 运行器 |
| CI 机密 | `MYGO_UPDATER_PRIVATE_KEY`、`MACOS_*`、`APPLE_*`、`WINDOWS_CERTIFICATE*` | 证书 base64 存机密，运行时写入临时文件/钥匙串 |

## 文档未说明

- `updates.s3` 在非 AWS 兼容服务（如 R2）上的 `region` 推荐取值：文档只说明默认 `us-east-1`；增量更新的实际压缩率、构建时下载旧归档的耗时上限、安装脚本被防火墙或代理拦截时的行为，文档也未给出。
