---
name: buildadmin-module-dev
description: Develop, package, and verify installable BuildAdmin modules, including PHP, web, web-nuxt, install.sql, and lifecycle hooks. Use for module ZIP creation or changes intended to install through BuildAdmin module management.
---

# BuildAdmin 模块开发与打包

适用于通过 BuildAdmin「模块管理」安装的模块。普通项目内功能开发使用 `buildadmin-dev`；涉及模块包、安装或升级行为时使用本技能。

## 确认安装边界

1. 查看当前项目的模块安装器、同类模块和目标 BuildAdmin 版本，以项目实际安装逻辑为准。官方参考：[开始](https://doc.buildadmin.com/senior/module/start.html)、[目录结构](https://doc.buildadmin.com/senior/module/directoryStructure.html)。
2. 列出模块用到的 PHP、`web`、`web-nuxt`、配置、静态资源和依赖。区分模块自身文件与必须覆盖的项目共享文件；检查共享文件的冲突和版本兼容性。
3. ZIP 根目录放 `info.ini`、`config.json`、需要数据库时的 `install.sql`，以及以模块标识命名的核心类。要复制到项目根目录的文件按原路径放在 ZIP 的 `app/`、`config/`、`extend/`、`public/`、`vendor/`、`web/`、`web-nuxt/` 等目录下。不要在 ZIP 内再套一层 `modules/<标识>/`。

## 元数据、数据库和生命周期

- `info.ini` 的 `uid` 是**模块标识**，不是管理员 ID；填写标题、作者、版本和状态。`config.json` 声明安装器支持的 Composer/NPM 依赖，并确认依赖会安装到哪个前端工程，不能假设 `web` 的 NPM 配置会自动处理 `web-nuxt`。参考 [基本信息](https://doc.buildadmin.com/senior/module/basicInfo.html)、[模块配置](https://doc.buildadmin.com/senior/module/configJson.html)。
- 新安装所需的完整表结构放在模块根目录 `install.sql`；表名使用 `__PREFIX__` 和模块标识前缀。首次发布前可以合并未发布的增量脚本；发布后保留已有安装的升级路径，不要把 `CREATE TABLE IF NOT EXISTS` 当作现有表的字段升级。参考 [安装 SQL](https://doc.buildadmin.com/senior/module/installSql.html)。
- 核心类实现任务所需的 `install`、`uninstall`、`update`、`enable`、`disable` 钩子。确认 SQL、文件复制与钩子的执行时序，避免重复执行非幂等操作；卸载只清理本模块拥有的资源。参考 [核心控制器](https://doc.buildadmin.com/senior/module/coreController.html)、[行为事件](https://doc.buildadmin.com/senior/module/event.html)。
- 需要修改 `web/src/main.ts` 时，先看 [WEB 引导程序](https://doc.buildadmin.com/senior/module/webBootstrap.html) 和项目已有机制，避免无条件覆盖启动文件。

## WEB 与 Nuxt 完整性

- 有后台或 SPA 页面时，将相应 `web` 源码、接口、语言包和资源一起放入包；菜单组件路径应与包内文件一致。
- 有 Nuxt SSR 页面时，`web-nuxt` 是 BuildAdmin 支持的模块覆盖目录。收集页面、组件、composables、语言包、资源、server 路由/中间件和确实需要的配置；检查导入依赖的文件闭包。不要打入 `node_modules`、`.nuxt`、`.output`、密钥或本地 `.env`。
- 文件安装不等于 SSR 服务已运行。交付说明写明 Nuxt 依赖安装、生产环境变量、构建、Node 进程和反向代理配置。若覆盖 `web-nuxt/nuxt.config.ts` 等共享文件，说明覆盖风险并验证目标项目。

## 打包与验证

1. 优先使用项目已有的打包器；没有时依据目标安装器的要求建立打包步骤。保留所需的 ZIP 根目录结构、归档元数据和 `.runtime` 文件清单。不要直接压缩整个项目目录。
2. 对照功能清单检查 ZIP 中每个必需文件，特别检查 `web-nuxt/` 是否实际入包，以及路径能否映射到安装目标。排除构建缓存、测试数据、凭据和无关文件。
3. 项目有包校验脚本时运行它；否则直接检查 ZIP 结构、清单和关键文件。有条件时在隔离环境验证全新安装、SQL 前缀替换、菜单注册和页面构建。对已有数据库的升级单独验证，避免用新装 SQL 覆盖数据。
4. 汇报安装包路径、包含范围、验证结果，以及安装后仍需执行的部署步骤。

具体调用方式见 [通用使用示例](examples/basic.md)。技能来源见 [来源记录](references/source.md)。
