# 模块包结构与安装映射

> 来源：BuildAdmin 官方文档《开始》《模块目录结构》（`doc.buildadmin.com/senior/module/start.html`、`.../directoryStructure.html`），整理自 Gitee `wonderful-code/build-admin-doc` v2 分支，抓取日期 2026-11-23。本文是压缩整理，不是原文照搬；安装顺序和可覆盖目录以目标项目的模块安装器代码为准。

## 模块能力

模块是 BuildAdmin 的扩展包，安装后可以：

- 在系统任意位置新增文件；
- 声明 Composer / NPM 依赖，由系统模块安装器在安装时自动安装；
- 覆盖系统已有文件（安装时自动备份；要上架市场时，覆盖核心文件的模块大概率无法通过审核）；
- 安装时自动导入 `install.sql`；
- 在安装、卸载、启用、禁用、更新时自动执行核心控制器里的对应方法。

## ZIP 根布局

`info.ini` 是识别模块包的依据，必须存在且在 ZIP 根目录。其余文件按需提供：

```text
notice.zip
├── info.ini            # 必须：模块基本信息（info-ini.md）
├── config.json         # 可选：依赖声明、受保护文件（config-json.md）
├── install.sql         # 可选：需要建表/改数据时提供（install-sql.md）
├── Notice              # 可选：核心控制器，模块标识首字母大写（core-controller.md）
├── webBootstrap.stub   # 可选：向 main.ts / App.vue 注入代码（web-bootstrap.md）
├── README.md           # 可选：功能与安装说明
├── app/
├── config/
├── extend/
├── public/
├── vendor/
├── web/
└── web-nuxt/
```

规则：

- 不要在 ZIP 内再套一层 `modules/<标识>/`。`modules/<标识>/` 是**安装后**模块在项目中的位置，不是包内路径。
- 上表 `app/` 到 `web-nuxt/` 的目录按原路径覆盖到项目根目录，包内路径必须与项目内目标路径逐级对应。
- 包内只放干净的模块代码，不要带 `__MACOSX/`、`.DS_Store`、`.idea/`、`node_modules/`、`.nuxt/`、`.output/`、测试数据、日志或凭据。

## 安装过程与顺序

包内允许出现 `app`、`config`、`extend`、`public`、`vendor`、`web`、`web-nuxt` 七个目录，安装时直接覆盖到项目目录。官方文档说，覆盖前系统会：

1. 导入安装 SQL；
2. 检测文件与依赖的冲突并自动备份；
3. 执行开发者预设的模块启用脚本。

然后把上述目录覆盖到项目。官方文档没有写明第 3 步调用核心控制器的哪个方法（`install` 还是 `enable`），也没有分别给出 `install`、`update` 与 SQL、文件复制的完整时序，这些只能在目标安装器代码里确认。能确定的是：覆盖发生在这些步骤之后，所以钩子里不要假设 `/modules/<uid>/` 下的包内文件在 `install` 时就已可用。

禁用模块时，系统会删除模块新增的文件，并把安装时备份的冲突文件还原（受保护文件除外，见 [config-json.md](config-json.md)）；卸载时清理哪些资源，由 `uninstall` 与安装器共同决定。**以目标项目的安装器实现为准**，动手前读一遍安装器代码。

## 打包

- 进入模块目录（`info.ini` 所在目录），全选内容打包为 ZIP；不要直接压缩整个项目目录。
- 升级包与首次发布包打包方式一致，是**完整覆盖包**，不是增量包。
- 有 `uniapp` 代码时单独打成 `uniapp.zip`（进入 `pages.json` 所在目录全选打包），作为附件提供。
- 上架市场的前置条件（官方文档）：模块需官方审核；覆盖核心文件大概率不能上架；包及其文档、截图、介绍内不得含侵权字体、图标、代码、图片、音视频；准备 `README.md`、300×200 的模块 LOGO、作者 QQ；资料发送到 `hi@buildadmin.com`；发送前用干净系统测试模块包。

## 打包前核对

1. 根目录有 `info.ini`，`uid` 与核心类文件名、命名空间、数据表前缀一致。
2. 需要数据库时有 `install.sql`，表名以 `__PREFIX__<uid>_` 开头。
3. 需要钩子时有 `Ucfirst(uid)` 命名的核心类，命名空间 `modules\<uid>`。
4. `web/`、`web-nuxt/` 内的页面、语言包、静态资源路径与菜单组件、接口地址对得上。
5. 包内没有构建缓存、依赖目录、密钥和本地 `.env`。
