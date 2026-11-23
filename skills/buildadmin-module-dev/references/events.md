# 行为事件

> 来源：BuildAdmin 官方文档《模块行为事件》（`doc.buildadmin.com/senior/module/event.html`），抓取日期 2026-11-23。内容为压缩整理，代码为官方示例节选。

核心控制器除了生命周期钩子（见 [core-controller.md](core-controller.md)），还可以监听 ThinkPHP 与 BuildAdmin 的事件，在应用运行期间扩展行为。

## 入口：AppInit

在核心控制器里定义 `AppInit()` 方法即可监听应用初始化。`AppInit` 是第一个执行的事件，适合在这里注册其他事件的监听。

```php
<?php

namespace modules\test;

class Test
{
    /**
     * 监听 AppInit
     */
    public function AppInit()
    {
        // AppInit 时执行的代码
    }

    public function install()
    {
        // ...
    }
}
```

## 注册其他事件

其他事件在 `AppInit` 内用 `Event::listen()` 注册，此时注册正合适，因为它们还没触发：

```php
<?php

namespace modules\test;

use think\facade\Event;
use app\admin\model\UserScoreLog;

class Test
{
    public function AppInit()
    {
        // 监听会员注册成功事件（ba 定义，具体 tag 见下方说明）
        Event::listen('user_register_successed', function ($user) {
            UserScoreLog::create([
                'user_id' => $user->id,
                'score'   => 10,
                'memo'    => '感谢注册，赠送10个积分',
            ]);
        });

        // 监听应用开始事件
        Event::listen('HttpRun', function () {
            // 应用开始了
        });
    }

    public function install()
    {
        // ...
    }
}
```

## ThinkPHP 内置事件

| 事件 | 描述 | 参数 |
| --- | --- | --- |
| AppInit | 应用初始化 | 无 |
| HttpRun | 应用开始 | 无 |
| HttpEnd | 应用结束 | 当前响应对象实例 |
| LogWrite | 日志 write 方法标签位 | 当前写入的日志信息 |
| RouteLoaded | 路由加载完成 | 无 |
| LogRecord | 日志记录 | 无 |

## BuildAdmin 内置事件

| 事件 | 描述 | 参数 |
| --- | --- | --- |
| backendInit | 管理员验权标签位 | 管理员的鉴权类实例 |
| frontendInit | 会员验权标签位 | 会员的鉴权类实例 |
| userRegisterSuccess | 会员注册成功标签位 | 会员数据模型 |
| uploadConfigInit | 上传信息配置 | `App $app` |
| cacheClearAfter | 清理系统缓存后 | `App $app` |
| AttachmentDel | 附件删除时 | `AttachmentModel $attachment` |

**注意上游文档不一致**：注册示例用的是 `user_register_successed`，而事件表列的是 `userRegisterSuccess`。接入前在项目代码里搜索实际的 `Event::listen` / `Event::trigger` tag，以代码为准，不要照抄示例字符串。

## 使用建议

- 只在 `AppInit` 里做注册，不在里面写耗时业务；事件回调在对应时机随应用执行。
- 监听前确认事件在目标 BuildAdmin 版本存在；不同版本的事件集合会变化，以项目代码为准。
- 行为事件不替代钩子：安装、卸载、启用、禁用、升级用核心控制器钩子；运行期扩展用事件。
- 事件回调里访问的模型、服务类必须已经存在，跨模块依赖要在交付说明里写明。
