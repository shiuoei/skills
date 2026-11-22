# 我的技能库

集中维护自己的技能，按场景查找，按目录独立使用。仓库维护无需安装额外依赖。

- [安装与更新](docs/getting-started.md)
- [编写与维护约定](docs/authoring.md)
- [导入记录](docs/import-notes.md)

通过下面的分类导航查找技能，或用浏览器、GitHub、编辑器搜索“日报”“托盘”“后台”等关键词。

## 技能导航

当前收录 4 个技能。

### 常用入口

- [daily-report](skills/daily-report/SKILL.md)：整理今天已完成的工作，生成按周归档的日报。

### 项目开发

| 技能 | 用途 | 关键词 | 状态 | 示例指令 |
| --- | --- | --- | --- | --- |
| [buildadmin-dev](skills/buildadmin-dev/SKILL.md) | BuildAdmin 后台二次开发：表格、表单、接口及菜单权限 | BuildAdmin、ThinkPHP、Vue、后台、CRUD、菜单权限 | 试用中 | 用 buildadmin-dev 新增订单管理页面，包含表格、表单和菜单权限 |
| [buildadmin-module-dev](skills/buildadmin-module-dev/SKILL.md) | BuildAdmin 可安装模块的开发、打包与验证 | BuildAdmin、模块开发、安装包、ZIP、Nuxt、install.sql | 试用中 | 用 buildadmin-module-dev 将当前项目的公告管理功能打包成可安装模块并验证 |
| [mygo-dev](skills/mygo-dev/SKILL.md) | MyGo 应用开发：原生 UI、托盘、打包及自动更新 | MyGo、Go、桌面应用、原生UI、托盘、打包、自动更新 | 试用中 | 用 mygo-dev 开发一个带托盘菜单的桌面工具 |

### 工作整理

| 技能 | 用途 | 关键词 | 状态 | 示例指令 |
| --- | --- | --- | --- | --- |
| [daily-report](skills/daily-report/SKILL.md) | 将已完成的工作写成通俗日报，按周归档 | 日报、工作总结、收工、按周归档 | 试用中 | 用 daily-report 整理今天已完成的工作 |

## 目录约定

- `skills/<name>/SKILL.md`：技能入口；参考资料、模板和示例随技能存放。
- `catalog.yaml`：分类、标签、状态、平台与使用示例；不重复技能入口的名称和描述。
- `docs/`：统一的安装与维护说明。

README 导航手工维护；新增技能或调整分类、状态时，同步更新本页和索引。
状态表示维护成熟度；平台标签表示预期使用平台，不代表已逐平台测试。

## 许可

仓库尚未选择整体开源许可证。原有来源说明保留在各技能内。
