# 来源与适用范围

## 技能来源

本技能整理自 `wx-vue` 项目的 `.agents/skills/buildadmin-module-dev/SKILL.md`，整理日期 2026-11-22。原技能的 Nav 模块实例使用项目内 `modules/nav/tools/package.php` 生成 ZIP、`modules/nav/tools/verify-package.php` 检查包，并以 `modules/nav/docs/nav-nuxt-deployment.md` 说明部署。这些文件只属于来源项目，未随技能复制，也不是其他项目的依赖。

来源项目脚本可以作为理解打包机制的资料，不能当作跨项目验证结果。

## 上游文档整理

`references/` 下的模块开发说明整理自 BuildAdmin 官方文档「模块开发」栏目，对应文档仓库 Gitee `wonderful-code/build-admin-doc` 的 v2 分支（文档站页脚标注 Apache-2.0，具体许可以仓库 LICENSE 为准）。抓取日期 2026-11-23。

| 本地文件 | 上游页面 |
| --- | --- |
| [package-layout.md](package-layout.md) | 《开始》`/senior/module/start.html`；《模块目录结构》`/senior/module/directoryStructure.html` |
| [info-ini.md](info-ini.md) | 《模块基本信息文件》`/senior/module/basicInfo.html` |
| [config-json.md](config-json.md) | 《模块配置（依赖声明等）》`/senior/module/configJson.html` |
| [install-sql.md](install-sql.md) | 《模块安装 SQL》`/senior/module/installSql.html` |
| [core-controller.md](core-controller.md) | 《模块核心控制器》`/senior/module/coreController.html` |
| [events.md](events.md) | 《模块行为事件》`/senior/module/event.html` |
| [web-bootstrap.md](web-bootstrap.md) | 《向 main.ts 插入代码》`/senior/module/webBootstrap.html` |

页面路径相对 `https://doc.buildadmin.com`。

整理方式：

- 内容为压缩、重排和补充说明，不是逐字全文；代码示例保留上游的类名、字段名和标记格式，仅 SQL 示例的表名按模块标识前缀规则做了改写（原示例见 [install-sql.md](install-sql.md)）。
- 标注「技能补充」的部分是本技能基于实践的工程判断，不是官方原文。
- 上游文档会随 BuildAdmin 版本更新；模块安装器、事件 tag、目录行为出现差异时，以目标项目的实际代码为准。
- 已知上游不一致：行为事件示例的 `user_register_successed` 与事件表的 `userRegisterSuccess` 不同，见 [events.md](events.md)。
