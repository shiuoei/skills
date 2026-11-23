# install.sql 安装 SQL

> 来源：BuildAdmin 官方文档《模块安装 SQL》（`doc.buildadmin.com/senior/module/installSql.html`），抓取日期 2026-11-23。内容为压缩整理，不是原文照搬。

模块根目录的 `install.sql` 会在安装时自动在项目数据库中执行。可以写建表、改表结构、更新配置等 SQL 命令，一个文件内可以有多条语句。

## 规则

- 用 `__PREFIX__` 代表当前项目的数据表前缀，不要写死 `ba_` 或其他前缀。
- 模块新增的数据表以**模块唯一标识**开头，例如模块 `notice` 的表名为 `__PREFIX__notice_*`。
- 卸载时是否需要删表由核心控制器的 `uninstall` 决定（见 [core-controller.md](core-controller.md)）；`install.sql` 只负责安装/升级时的结构。

## 示例

```sql
-- 修改系统配置
UPDATE __PREFIX__config SET value='测试1' WHERE name='site_name';

-- 创建一个数据表
CREATE TABLE IF NOT EXISTS `__PREFIX__notice_notice` (
    `id`  int(10) UNSIGNED NULL AUTO_INCREMENT COMMENT 'ID' ,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模块添加的数据表';

-- 假设模块在新版本修改了表结构，可以直接在此文件内写好改变结构的SQL
ALTER TABLE `__PREFIX__notice_notice` ADD COLUMN `welcome_msg` text COMMENT '欢迎语' AFTER `id`;
```

表名按命名规则改成了 `notice` 模块的写法；官方示例用的是 `__PREFIX__module_NewTable`，仅作示意，照抄会违反"表名以模块标识开头"的要求。

## 升级与重复执行

官方文档的示例把 `ALTER TABLE` 直接写在 `install.sql` 里，对应"改了表结构就重发完整包，安装器执行该文件"的做法。落地时注意：

- 文件会被**整份执行**。重复执行同一条 `ALTER TABLE ... ADD COLUMN` 会报错并中断安装/升级，语句要能判断或避免重复添加（实际执行范围以目标安装器为准）。
- 不要把字段直接并进 `CREATE TABLE`，又保留一条 `ADD COLUMN`：全新安装会先建出带该列的表，再 ALTER 一次，导致失败。
- `CREATE TABLE IF NOT EXISTS` 只保证表存在，**不会**给已有表补字段；已有安装的字段升级要单独写可执行的变更。
- 发布后的升级路径要保留：首次发布前可以合并未发布的增量脚本，已经发布的版本不要再回改历史语义。

（以上"升级与重复执行"为技能补充的工程判断，不是官方原文；实际执行时机以目标项目安装器为准。）

## 验证

1. 在隔离库中导入项目当前数据，跑一遍新装流程，检查前缀替换、表结构、默认数据。
2. 用已有数据的旧版本库模拟升级，确认语句可重复执行、不覆盖业务数据。
3. 检查表名、索引名、字段名带模块标识前缀，避免与系统表或其他模块冲突。
