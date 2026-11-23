# info.ini 模块基本信息

> 来源：BuildAdmin 官方文档《模块基本信息文件》（`doc.buildadmin.com/senior/module/basicInfo.html`），抓取日期 2026-11-23。内容为压缩整理，不是原文照搬。

BuildAdmin 通过 `info.ini` 判断一个 ZIP 是否为合法模块包：文件**必须存在，且必须位于 ZIP 根目录**。读不到时，模块管理无法识别该包。

## 字段

| 字段 | 含义 | 注意 |
| --- | --- | --- |
| `uid` | 模块唯一标识 | 不是管理员 ID。决定安装目录 `modules/<uid>/`、核心类文件名与命名空间、数据表前缀；改动等于换一个模块。 |
| `title` | 模块标题 | 展示名称。 |
| `intro` | 模块介绍 | 一句话说明功能。 |
| `author` | 模块作者 | |
| `authorid` | 模块作者官网 ID | 文档示例值为 `1`。 |
| `website` | 模块主页 | 完整 `http(s)://` 地址。 |
| `version` | 模块版本号 | 如 `1.0.0`；升级判断依赖它，发布新版本必须递增。 |
| `state` | 模块状态 | 系统自动维护，文件中"有此字段即可"，示例为 `0`；不要手工写成已安装状态。 |

## 示例

```ini
# 模块唯一标识，若不知您新建的模块标识是否唯一，暂时的，可主动询问群聊管理员
uid = test1
# 模块标题
title = 测试模块
# 模块介绍
intro = 测试模块的介绍
# 模块作者
author = 码上开花
# 模块作者官网ID
authorid = 1
# 模块主页
website = https://www.buildadmin.com
# 模块版本号
version = 1.0.0
# 模块状态，系统自动维护，有此字段即可
state = 0
```

要上架市场时，`info.ini` 内**不要带注释**，只保留 `键 = 值` 行。

## 读取模块信息

系统通过 `Server::getIni()` 读取模块目录，`state == 1` 表示已安装：

```php
$dir  = root_path() . 'modules/sms/'; // 模块所在目录，即 /modules/$uid/
$info = Server::getIni(Filesystem::fsFit($dir));
if ($info && $info['state'] == 1) {
    echo '短信模块存在，且状态为已安装';
}
```

示例省略了 `use` 语句，`Server`、`Filesystem` 的命名空间以目标项目上下文为准；`modules/sms/` 里的 `sms` 是模块标识，换成自己的 `uid`。

## 打包检查

- `uid` 与核心控制器文件名、命名空间、`install.sql` 表前缀三者一致。
- `version` 比上一版递增，不要用同一版本号重复发布。
- `state` 保持系统维护的值，不手工改。
- 上架包的 `info.ini` 无注释、无空字段。
