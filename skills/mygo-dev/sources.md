# 文档来源与覆盖矩阵

本技能把 [MyGo 官方中文文档](https://mygo.run/zh/docs) 的页面按主题提炼为 `references/` 下的文件。本文件记录来源、版本与逐页映射，便于核对覆盖和升级维护。

## 快照

| 项 | 值 |
| --- | --- |
| 来源 | <https://mygo.run/zh/docs> |
| 快照日期 | 2026-11-21 |
| 页面数 | 125 |
| 对应版本 | MyGo v0.10.0 文档（含 `mobile`、`plugins/render`） |
| 提取方式 | 中文文档页面逐页提炼；代码块保留 API 名称，正文重写 |

> 文档站可能比本机安装的 MyGo 新：先用 `go list -m github.com/egoist/mygo` 确认版本，再用 `go doc` 与模块自带 `docs/` 核对 API。本技能不绑定本机路径。

## 覆盖矩阵

每个文档页对应一个参考文件；`13-api-quick-reference.md` 由模块 API 与 ui 文档标题生成，`14-lessons-experience.md` 来自真实项目经验，二者不独占文档页。

### `references/01-getting-started.md`

| 页面 | 源路径 |
| --- | --- |
| [MyGo 文档](https://mygo.run/zh/docs) | `README.md` |
| [快速开始](https://mygo.run/zh/docs/getting-started) | `getting-started.md` |
| [配置](https://mygo.run/zh/docs/configuration) | `configuration.md` |
| [mygo CLI](https://mygo.run/zh/docs/cli) | `cli.md` |

共 4 页。

### `references/02-window-app-lifecycle.md`

| 页面 | 源路径 |
| --- | --- |
| [窗口](https://mygo.run/zh/docs/windows) | `windows.md` |
| [应用](https://mygo.run/zh/docs/app) | `app.md` |
| [文稿](https://mygo.run/zh/docs/documents) | `documents.md` |
| [启动器](https://mygo.run/zh/docs/launchers) | `launchers.md` |

共 4 页。

### `references/03-menus-tray-native-api.md`

| 页面 | 源路径 |
| --- | --- |
| [菜单与托盘](https://mygo.run/zh/docs/menus) | `menus.md` |
| [桌面 API](https://mygo.run/zh/docs/native) | `native.md` |

共 2 页。

### `references/04-ui-foundations.md`

| 页面 | 源路径 |
| --- | --- |
| [原生 UI](https://mygo.run/zh/docs/ui) | `ui/README.md` |
| [视图](https://mygo.run/zh/docs/ui/views) | `ui/views.md` |
| [布局](https://mygo.run/zh/docs/ui/layout) | `ui/layout.md` |
| [文本](https://mygo.run/zh/docs/ui/text) | `ui/text.md` |
| [样式与主题](https://mygo.run/zh/docs/ui/styling) | `ui/styling.md` |
| [输入](https://mygo.run/zh/docs/ui/input) | `ui/input.md` |
| [拖放](https://mygo.run/zh/docs/ui/drag-and-drop) | `ui/drag-and-drop.md` |
| [页面与导航](https://mygo.run/zh/docs/ui/navigation) | `ui/navigation.md` |
| [浮层](https://mygo.run/zh/docs/ui/overlays) | `ui/overlays.md` |
| [辅助功能](https://mygo.run/zh/docs/ui/accessibility) | `ui/accessibility.md` |
| [从右到左](https://mygo.run/zh/docs/ui/right-to-left) | `ui/right-to-left.md` |
| [绘制与动画](https://mygo.run/zh/docs/ui/drawing) | `ui/drawing.md` |
| [着色器](https://mygo.run/zh/docs/ui/shaders) | `ui/shaders.md` |
| [过渡](https://mygo.run/zh/docs/ui/transitions) | `ui/transitions.md` |
| [动效](https://mygo.run/zh/docs/ui/motion) | `ui/motion.md` |
| [变换](https://mygo.run/zh/docs/ui/transforms) | `ui/transforms.md` |
| [自定义控件](https://mygo.run/zh/docs/ui/custom-widgets) | `ui/custom-widgets.md` |
| [使用原生 UI 的窗口](https://mygo.run/zh/docs/ui/windows) | `ui/windows.md` |
| [测试](https://mygo.run/zh/docs/ui/testing) | `ui/testing.md` |
| [检查器](https://mygo.run/zh/docs/ui/inspector) | `ui/inspector.md` |
| [渲染](https://mygo.run/zh/docs/ui/rendering) | `ui/rendering.md` |
| [浏览器中的原生 UI](https://mygo.run/zh/docs/ui/browsers) | `ui/browsers.md` |
| [迁移到带校验的元素](https://mygo.run/zh/docs/ui/migration) | `ui/migration.md` |
| [原生 UI 性能](https://mygo.run/zh/docs/ui/performance) | `ui/performance.md` |

共 24 页。

### `references/05-ui-actions-selection.md`

| 页面 | 源路径 |
| --- | --- |
| [按钮](https://mygo.run/zh/docs/ui/button) | `ui/button.md` |
| [链接](https://mygo.run/zh/docs/ui/link) | `ui/link.md` |
| [菜单按钮](https://mygo.run/zh/docs/ui/menu-button) | `ui/menu-button.md` |
| [拆分按钮](https://mygo.run/zh/docs/ui/split-button) | `ui/split-button.md` |
| [上下文菜单](https://mygo.run/zh/docs/ui/context-menu) | `ui/context-menu.md` |
| [切换按钮](https://mygo.run/zh/docs/ui/toggle) | `ui/toggle.md` |
| [分段控件](https://mygo.run/zh/docs/ui/segmented) | `ui/segmented.md` |
| [工具栏](https://mygo.run/zh/docs/ui/toolbar) | `ui/toolbar.md` |
| [复选框](https://mygo.run/zh/docs/ui/checkbox) | `ui/checkbox.md` |
| [开关](https://mygo.run/zh/docs/ui/switch) | `ui/switch.md` |
| [单选按钮](https://mygo.run/zh/docs/ui/radio) | `ui/radio.md` |
| [选择器](https://mygo.run/zh/docs/ui/select) | `ui/select.md` |
| [滑块](https://mygo.run/zh/docs/ui/slider) | `ui/slider.md` |
| [范围滑块](https://mygo.run/zh/docs/ui/range-slider) | `ui/range-slider.md` |
| [步进器](https://mygo.run/zh/docs/ui/stepper) | `ui/stepper.md` |
| [评分](https://mygo.run/zh/docs/ui/rating) | `ui/rating.md` |

共 16 页。

### `references/06-ui-inputs-editors.md`

| 页面 | 源路径 |
| --- | --- |
| [文本输入](https://mygo.run/zh/docs/ui/text-input) | `ui/text-input.md` |
| [文本输入原语](https://mygo.run/zh/docs/ui/text-input-client) | `ui/text-input-client.md` |
| [数字输入](https://mygo.run/zh/docs/ui/number-input) | `ui/number-input.md` |
| [小数输入](https://mygo.run/zh/docs/ui/decimal-input) | `ui/decimal-input.md` |
| [搜索框](https://mygo.run/zh/docs/ui/search-field) | `ui/search-field.md` |
| [组合框](https://mygo.run/zh/docs/ui/combobox) | `ui/combobox.md` |
| [自动补全](https://mygo.run/zh/docs/ui/autocomplete) | `ui/autocomplete.md` |
| [标记字段](https://mygo.run/zh/docs/ui/token-field) | `ui/token-field.md` |
| [可编辑文本](https://mygo.run/zh/docs/ui/editable-text) | `ui/editable-text.md` |
| [查找栏](https://mygo.run/zh/docs/ui/find-bar) | `ui/find-bar.md` |
| [富文本编辑器](https://mygo.run/zh/docs/ui/rich-text-editor) | `ui/rich-text-editor.md` |
| [代码编辑器](https://mygo.run/zh/docs/ui/code-editor) | `ui/code-editor.md` |
| [差异查看器](https://mygo.run/zh/docs/ui/diff-viewer) | `ui/diff-viewer.md` |
| [Markdown](https://mygo.run/zh/docs/ui/markdown) | `ui/markdown.md` |
| [日历](https://mygo.run/zh/docs/ui/calendar) | `ui/calendar.md` |
| [日期输入](https://mygo.run/zh/docs/ui/date-input) | `ui/date-input.md` |
| [时间输入](https://mygo.run/zh/docs/ui/time-input) | `ui/time-input.md` |
| [颜色选择器](https://mygo.run/zh/docs/ui/color-picker) | `ui/color-picker.md` |

共 18 页。

### `references/07-ui-collections-containers-nav.md`

| 页面 | 源路径 |
| --- | --- |
| [列表](https://mygo.run/zh/docs/ui/list) | `ui/list.md` |
| [表格](https://mygo.run/zh/docs/ui/table) | `ui/table.md` |
| [树](https://mygo.run/zh/docs/ui/tree) | `ui/tree.md` |
| [大纲](https://mygo.run/zh/docs/ui/outline) | `ui/outline.md` |
| [网格视图](https://mygo.run/zh/docs/ui/grid-view) | `ui/grid-view.md` |
| [滚动视图](https://mygo.run/zh/docs/ui/scroll) | `ui/scroll.md` |
| [网格](https://mygo.run/zh/docs/ui/grid) | `ui/grid.md` |
| [分栏视图](https://mygo.run/zh/docs/ui/split) | `ui/split.md` |
| [可折叠](https://mygo.run/zh/docs/ui/collapsible) | `ui/collapsible.md` |
| [手风琴](https://mygo.run/zh/docs/ui/accordion) | `ui/accordion.md` |
| [表单](https://mygo.run/zh/docs/ui/form) | `ui/form.md` |
| [Web 视图](https://mygo.run/zh/docs/ui/web-view) | `ui/web-view.md` |
| [标签页](https://mygo.run/zh/docs/ui/tabs) | `ui/tabs.md` |
| [侧边栏](https://mygo.run/zh/docs/ui/sidebar) | `ui/sidebar.md` |
| [面包屑导航](https://mygo.run/zh/docs/ui/breadcrumbs) | `ui/breadcrumbs.md` |
| [后退和前进按钮](https://mygo.run/zh/docs/ui/history-buttons) | `ui/history-buttons.md` |

共 16 页。

### `references/08-ui-dialogs-status-media.md`

| 页面 | 源路径 |
| --- | --- |
| [对话框](https://mygo.run/zh/docs/ui/dialog) | `ui/dialog.md` |
| [警告对话框](https://mygo.run/zh/docs/ui/alert-dialog) | `ui/alert-dialog.md` |
| [浮层](https://mygo.run/zh/docs/ui/popover) | `ui/popover.md` |
| [工具提示](https://mygo.run/zh/docs/ui/tooltip) | `ui/tooltip.md` |
| [Toast](https://mygo.run/zh/docs/ui/toast) | `ui/toast.md` |
| [进度条](https://mygo.run/zh/docs/ui/progress) | `ui/progress.md` |
| [旋转指示器](https://mygo.run/zh/docs/ui/spinner) | `ui/spinner.md` |
| [量度计](https://mygo.run/zh/docs/ui/meter) | `ui/meter.md` |
| [徽标](https://mygo.run/zh/docs/ui/badge) | `ui/badge.md` |
| [图表](https://mygo.run/zh/docs/ui/chart) | `ui/chart.md` |
| [图像](https://mygo.run/zh/docs/ui/image) | `ui/image.md` |
| [视频](https://mygo.run/zh/docs/ui/video) | `ui/video.md` |
| [图标](https://mygo.run/zh/docs/ui/icon) | `ui/icon.md` |
| [头像](https://mygo.run/zh/docs/ui/avatar) | `ui/avatar.md` |

共 14 页。

### `references/09-web-frontend-bindings-plugins.md`

| 页面 | 源路径 |
| --- | --- |
| [前端](https://mygo.run/zh/docs/frontend) | `frontend.md` |
| [从前端调用 Go](https://mygo.run/zh/docs/bindings) | `bindings.md` |
| [插件](https://mygo.run/zh/docs/plugins) | `plugins.md` |
| [自动化](https://mygo.run/zh/docs/plugins/automation) | `plugins/automation.md` |
| [Fetch](https://mygo.run/zh/docs/plugins/fetch) | `plugins/fetch.md` |
| [游戏](https://mygo.run/zh/docs/plugins/game) | `plugins/game.md` |
| [玻璃](https://mygo.run/zh/docs/plugins/glass) | `plugins/glass.md` |
| [Loading](https://mygo.run/zh/docs/plugins/loading) | `plugins/loading.md` |
| [PermissionFlow](https://mygo.run/zh/docs/plugins/permissionflow) | `plugins/permissionflow.md` |
| [Render](https://mygo.run/zh/docs/plugins/render) | `plugins/render.md` |
| [SF Symbols](https://mygo.run/zh/docs/plugins/sfsymbols) | `plugins/sfsymbols.md` |
| [SQLite](https://mygo.run/zh/docs/plugins/sqlite) | `plugins/sqlite.md` |
| [终端](https://mygo.run/zh/docs/plugins/terminal) | `plugins/terminal.md` |
| [更新器](https://mygo.run/zh/docs/plugins/updater) | `plugins/updater.md` |
| [文件监视](https://mygo.run/zh/docs/plugins/watch) | `plugins/watch.md` |
| [WebSocket](https://mygo.run/zh/docs/plugins/websocket) | `plugins/websocket.md` |

共 16 页。

### `references/10-i18n-data-audio-native-modules.md`

| 页面 | 源路径 |
| --- | --- |
| [语言](https://mygo.run/zh/docs/i18n) | `i18n.md` |
| [剪贴板与拖放数据](https://mygo.run/zh/docs/data-transfer) | `data-transfer.md` |
| [音频](https://mygo.run/zh/docs/audio) | `audio.md` |
| [原生模块](https://mygo.run/zh/docs/native-modules) | `native-modules.md` |

共 4 页。

### `references/11-mobile.md`

| 页面 | 源路径 |
| --- | --- |
| [手机和平板](https://mygo.run/zh/docs/mobile) | `mobile/README.md` |
| [为手机构建和发布](https://mygo.run/zh/docs/mobile/building) | `mobile/building.md` |
| [手机上的界面](https://mygo.run/zh/docs/mobile/interface) | `mobile/interface.md` |
| [设备 API 和插件](https://mygo.run/zh/docs/mobile/plugins) | `mobile/plugins.md` |

共 4 页。

### `references/12-distribution-updates-ci.md`

| 页面 | 源路径 |
| --- | --- |
| [构建与分发](https://mygo.run/zh/docs/distribution) | `distribution.md` |
| [自动更新](https://mygo.run/zh/docs/updates) | `updates.md` |
| [GitHub Actions](https://mygo.run/zh/docs/github-actions) | `github-actions.md` |

共 3 页。

## 覆盖统计

| 参考文件 | 页数 |
| --- | --- |
| `01-getting-started.md` | 4 |
| `02-window-app-lifecycle.md` | 4 |
| `03-menus-tray-native-api.md` | 2 |
| `04-ui-foundations.md` | 24 |
| `05-ui-actions-selection.md` | 16 |
| `06-ui-inputs-editors.md` | 18 |
| `07-ui-collections-containers-nav.md` | 16 |
| `08-ui-dialogs-status-media.md` | 14 |
| `09-web-frontend-bindings-plugins.md` | 16 |
| `10-i18n-data-audio-native-modules.md` | 4 |
| `11-mobile.md` | 4 |
| `12-distribution-updates-ci.md` | 3 |
| **合计** | **125** |

## 版本门限

- 整页需要 mygo ≥ v0.10.0：`references/11-mobile.md`（手机/平板）与 `references/09-web-frontend-bindings-plugins.md` 中的 `plugins/render`。
- 其余页面按 v0.10 文档提炼，但正文夹有个别 v0.10 新增 API——**v0.9.0 的 `go doc` 里没有**。已核对清单（2026-11-21，对照 v0.9.0）：
  - `references/02-window-app-lifecycle.md`：`ConsumeActivationClick`、`SetThumbBar`、`ZoomControl`
  - `references/04-ui-foundations.md`：`FocusWith`、`Touches`/`TouchBegan`、`Router.SystemStack`、`DrawInstances`
  - `references/05-ui-actions-selection.md`：`ui.PopoverMenuOptions`（v0.9.0 为 `Element.PopoverMenu(fn)`）
  - `references/06-ui-inputs-editors.md`：`ScrollBy`、`ScrollToParagraph`、`SetBlocks`、`SetLineStyles`、`ShowPopup`、`ScrollMargin`、`OnTask`、`LinkColor`、`AnchorBlock`、`SourceLine`
  - `references/07-ui-collections-containers-nav.md`：`NearEnd`/`EndReached`、`Frozen`、`VirtualColumns`、`RowHeight`、`SplitGap`
  - `references/08-ui-dialogs-status-media.md`：`VideoSeeker`
- v0.9.0 项目遇到上述缺失项时：用 `go doc` 找替代写法，或升级 MyGo；不要直接照抄。
- `references/14-lessons-experience.md` 的结论标注了验证版本（v0.9.0 / Windows 11）；升级后复核。

## 刷新方法

1. 打开 <https://mygo.run/zh/docs>，核对页面总数（当前 125）与新增/删除的页面。
2. 更新上面的覆盖矩阵与对应参考文件顶部的「源页面」清单。
3. 若新增页面属于某个主题，补进对应参考文件；若跨主题，先在这里登记再决定归属。
4. 更新本文件的快照日期、对应版本与版本门限。
5. 重新用 `go doc` 抽查 10–15 个 API，确认与安装版本一致。
