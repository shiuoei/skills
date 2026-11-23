# 核心控制器与生命周期钩子

> 来源：BuildAdmin 官方文档《模块核心控制器》（`doc.buildadmin.com/senior/module/coreController.html`），抓取日期 2026-11-23。示例为官方代码节选，说明文字是压缩整理。

## 命名

核心控制器以模块标识**首字母大写**命名，文件名和类名相同，命名空间为 `modules\<uid>`：

| 模块唯一标识 | 文件名 | 类名 |
| --- | --- | --- |
| test | Test | Test |
| userlogin | Userlogin | Userlogin |
| login1 | Login1 | Login1 |

文件放在 ZIP 根目录，安装后位于 `modules/<uid>/<Ucfirst(uid)>.php`。

## 钩子方法

| 方法 | 触发时机 | 典型用途 |
| --- | --- | --- |
| `install()` | 安装模块 | 创建菜单、写默认配置、初始化数据 |
| `uninstall()` | 卸载模块 | 删除本模块添加的菜单、表、文件、配置 |
| `enable()` | 启用模块 | 启用菜单、写回配置、注册配置入口 |
| `disable()` | 禁用模块 | 禁用菜单、移除入口、备份配置 |
| `update()` | 升级模块 | 数据迁移、修正菜单、补齐配置 |

只实现需要的钩子，方法体可以为空。钩子里的操作要能判断是否已完成或可安全重复执行，避免安装、升级重跑时报错或重复建菜单。卸载只清理本模块拥有的资源；被覆盖的系统文件由系统按安装时的备份还原。

## 示例：菜单、配置与配置缓存

```php
<?php

namespace modules\test;

use think\facade\Db;
use app\common\library\Menu;
use app\admin\model\AdminRule;

class Test
{
    /**
     * 安装模块时执行的方法
     */
    public function install()
    {
        // 往后台常规管理内添加一个菜单
        $pMenu = AdminRule::where('name', 'routine')->value('id');
        $menu  = [
            [
                'type'      => 'menu',
                'title'     => '通知公告管理',
                'name'      => 'routine/notice',
                'path'      => 'routine/notice',
                'icon'      => 'el-icon-ChatLineRound',
                'menu_type' => 'tab',
                'component' => '/src/views/backend/routine/notice/index.vue',
                'keepalive' => '1',
                'pid'       => $pMenu ? $pMenu : 0,
                'children'  => [
                    ['type' => 'button', 'title' => '查看', 'name' => 'routine/notice/index'],
                    ['type' => 'button', 'title' => '添加', 'name' => 'routine/notice/add'],
                    ['type' => 'button', 'title' => '编辑', 'name' => 'routine/notice/edit'],
                    ['type' => 'button', 'title' => '删除', 'name' => 'routine/notice/del'],
                    ['type' => 'button', 'title' => '快速排序', 'name' => 'routine/notice/sortable'],
                ],
            ],
        ];
        Menu::create($menu);
    }

    /**
     * 卸载模块时执行的方法
     */
    public function uninstall()
    {
        // 删除添加的菜单（第二个参数 true：同时删除子节点）
        Menu::delete('routine/notice', true);
    }

    /**
     * 启用模块时执行的方法
     */
    public function enable()
    {
        // 修改系统配置
        Db::name('config')
            ->where('name', 'record_number')
            ->update(['value' => '备案号-测试1']);

        // 启用模块添加的菜单
        Menu::enable('routine/notice');

        // 添加一个 系统配置 -> 快速配置 入口
        Config::addQuickEntrance('测试模块配置', '/admin/test/config');

        // 禁用时缓存的配置写回配置文件
        $config = Cache::pull('test-module-config');
        if ($config) {
            @file_put_contents(config_path() . 'test.php', $config);
        }
    }

    /**
     * 禁用模块时执行的方法
     */
    public function disable()
    {
        // 禁用模块添加的菜单
        Menu::disable('routine/notice');

        // 删除 系统配置 -> 快速配置 入口
        Config::removeQuickEntrance('测试模块配置');

        // 模块新增的文件在禁用时会被清理，先把配置文件内容存进缓存，启用时再写回
        $config = @file_get_contents(config_path() . 'test.php');
        if ($config) {
            Cache::set('test-module-config', $config, 3600);
        }
    }

    /**
     * 升级模块时执行的方法
     */
    public function update()
    {
        // 数据迁移、菜单修正、补齐配置
    }
}
```

说明：

- 示例中 `Cache`、`Config` 没有对应的 `use` 语句，按目标项目的实际命名空间引入。
- `Menu`、`AdminRule` 的具体行为以项目实现为准（`name` 为规则名，`component` 为页面路径）；新增菜单的 `component` 必须能在包内或项目内找到对应文件。
- "配置文件内容进缓存、启用时写回"用于模块配置没有存数据库时的备份，避免禁用清理文件后配置丢失；如果配置已存数据库，这一步可以省略。

## 与打包的对应关系

- 核心控制器必须放在 ZIP 根目录，且文件名与 `uid` 对得上，否则安装器找不到钩子。
- `install` 里注册的菜单、`enable` 里打开的菜单、`disable` 里关闭的菜单要成对出现；`uninstall` 删除的资源与 `install` 创建的资源对应。
- `install`/`update` 与 `install.sql` 的分工：结构变更写在 SQL，菜单、配置、数据初始化写在钩子。官方文档只说明"覆盖前"依次导入 SQL、检测备份并执行启用脚本，没有分别给出 `install`、`update` 与 SQL 的时点，以目标安装器代码为准（见 [package-layout.md](package-layout.md) 的「安装过程与顺序」）。
