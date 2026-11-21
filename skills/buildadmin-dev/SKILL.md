---
name: "buildadmin-dev"
description: "BuildAdmin 后台管理系统二次开发技能：提供后端(ThinkPHP8)/前端(Vue3+TS+Element Plus)开发规范、数据表设计、CRUD 生成、菜单权限等约定。当用户在 BuildAdmin 项目中进行 AI 辅助开发（写控制器/模型/前端表格表单/配菜单权限等）时使用。"
---

# BuildAdmin 二次开发技能

本技能面向在 **BuildAdmin** 项目（仓库内含 `buildadmin-v2` 服务端/前端源码与 `build-admin-doc-v2` 官方文档）中进行的 AI 辅助二次开发，帮助开发者遵循框架约定，快速编写符合规范的后端控制器/模型/验证器、前端表格/表单页面、数据表与菜单权限配置。

## 1. 技术栈与简介

- **后端**：ThinkPHP 8（多应用模式），应用划分为 `admin`（后台）、`api`（前台）、`common`（公共）。
- **前端**：Vue 3 + TypeScript + Vite + Pinia + Element Plus + vue-i18n + Axios。
- **架构**：前后端分离，但服务端与 WEB 端同仓（`buildadmin-v2` 根目录即服务端，`web/` 子目录为前端工程）。
- **典型页面流程**：后台页面 = `TableHeader`（顶部菜单）+ `Table`（表格）+ `PopupForm`（弹窗表单），由 `baTable` 表格管家类驱动数据与事件。
- **运行流程速览**：WEB 端 `main.ts` 注册 pinia/i18n/router/element-plus，加载语言包与指令；进入后台 `/admin`（静态路由，绑定 `/@/layouts/backend/index.vue`）时发起初始化请求，获取动态路由与权限节点并注册；Server 端控制器通常继承 `Api` 基类（IP 检查、时区、语言包加载），后台/前台控制器再继承 `Backend`/`Frontend`（初始化 `$this->auth` 鉴权实例）。

## 2. 目录结构

### Server 端（buildadmin-v2 根目录）

```
├─app                应用目录
│  ├─admin           后台应用（controller/model/validate/lang/library 等）
│  ├─api             前台应用
│  └─common          公共（controller 基类、library、model 等）
├─config             配置目录（数据库配置在 config/database.php）
├─database           Phinx 数据库迁移文件
├─extend             扩展目录
├─install            安装器源代码（不一定存在）
├─modules            模块市场模块存放目录（如 /modules/sms/，模块根目录含 info.ini）
├─public             运行目录（static 静态资源、storage 上传文件保存处）
├─runtime            运行时目录（缓存、日志）
├─web                WEB 端源代码（前端工程）
└─think              ThinkPHP 命令行入口
```

要点：
- `app/admin/controller/` 后台控制器，`app/admin/model/` 模型，`app/admin/validate/` 模型验证器，`app/admin/lang/<语言>/` 后端语言包，`app/admin/library/traits/Backend.php` 后台通用 Trait。
- 后台业务模块的模型如 `app\admin\model\User`（对应控制器 `app\admin\controller\user\User.php`），关联模型同样放在 `app\admin\model`。

### Web 端（buildadmin-v2/web）

```
web/src
├─api              所有接口相关
│  ├─common.ts     公共 Url 定义、baTableApi 类（快速生成控制器增删改查排序请求）
│  ├─backend       后台接口方法定义
│  └─frontend      前台接口方法定义
├─components       组件
│  ├─table         Table 表格组件、TableHeader 顶部菜单、defaultOptButtons
│  ├─formItem      表单项组件（el-form-item 内嵌 baInput）
│  ├─baInput       输入组件封装（含上传、数组、城市/图标选择、富文本等）
│  └─icon、mixins、clickCaptcha、terminal 等
├─lang             语言包（autoload.ts 按需加载映射、globs-zh-cn.ts/globs-en.ts 全局语言包）
│  ├─backend      后台页面语言包（zh-cn.ts、en.ts 及按目录组织的页面语言包）
│  ├─frontend      前台页面语言包
│  └─common        公共页面语言包
├─layouts          布局（backend 后台布局、frontend 前台布局、common 公共布局）
├─router           路由（static.ts 静态路由等）
├─stores           状态商店（adminInfo 管理员资料、userInfo 用户资料、config 布局、navTabs 等）
├─styles           样式表
├─utils            工具库（axios.ts 网络请求封装、baTable.ts 表格管家类、common.ts 公共方法含 auth 鉴权、directives.ts 指令、validate.ts 表单验证等）
└─views            视图
   ├─backend       后台视图（与后台菜单组件路径对应，如 backend/auth/group/index.vue）
   ├─frontend      前台视图
   └─common        公共视图
```

要点：
- 语言包按 `t('目录名.页面名.翻译名')` 的约定组织，如后台用户管理页面语言包位于 `web/src/lang/backend/zh-cn/user/user.ts`（英文在 `web/src/lang/backend/en/user/user.ts`）。
- 后台视图默认目录 `web/src/views/backend/<菜单目录>/<页面>/index.vue` 与 `popupForm.vue`。

## 3. 开发环境

1. **启动服务端**：在 `buildadmin-v2` 根目录执行 `php think run`（**不要**开启 Nginx/Apache；该命令能读取环境变量以实现 WEB 终端的命令执行功能，也是开发阶段的 API 服务）。
2. **启动前端**：在 `web/` 目录执行 `npm run dev`，浏览器访问 **http://localhost:1818/#/admin**（1818 是 Vite 开发服务端口，修改前端代码后热更新生效）。
3. **数据库**：安装时填写的数据库资料保存在 `config/database.php`，开发前需开启对应数据库服务。
4. **调试**：建议开启 ThinkPHP 调试模式（可参考官方文档"开启调试模式"），开发环境下接口会返回详细报错信息。
5. 开发环境修改代码无需重新发布；上线前才进行构建发布。

## 4. 数据表设计规范

设计数据表时遵循本规范，`可视化 CRUD` 会自动将字段映射为对应的前端组件。

### 字段类型 → 组件映射

| 类型 | CRUD 搭配点 |
| --- | --- |
| `enum` | 自动生成**单选框** |
| `set` | 自动生成**复选框** |
| `date` / `year` / `time` | 日期 / 年份 / 时间选择组件 |
| `datetime` / `timestamp` | 时间日期选择组件 |
| `decimal` / `double` / `float` | Number 输入框（步长按默认值自动计算） |
| `int` / `bigint` / `mediumint` / `smallint` / `tinyint` | Number 输入框（步长为 1） |
| `longtext` / `text` / `mediumtext` / `smalltext` / `tinytext` / `bigtext` | textarea 文本框 |

### 特殊字段名

| 字段名 | 类型要求 | 作用 |
| --- | --- | --- |
| `weigh` | int | 后台排序字段，存在该字段则表格出现拖拽排序按钮 |
| `create_time` / `update_time` | bigint | 记录添加/更新时间，系统自动维护，无需手动赋值 |

### 字段后缀 → 组件映射（名称结尾）

| 后缀 | 名称示例 | CRUD 搭配点 |
| --- | --- | --- |
| `array` | namearray | 数组输入组件 |
| `list` / `select` / `data` | namelist | select 单选 |
| `lists` / `selects` / `multi` | name_selects | select 多选 |
| `_id` | user_id | 关联表远程 select 单选 |
| `_ids` | user_ids | 关联表远程 select 多选 |
| `city` | city | 城市选择器（需安装省份城市信息模块） |
| `image` / `avatar` | desc_image | 图片上传（单图） |
| `images` / `avatars` | descimages | 图片上传（多图） |
| `file` / `files` | attachfile | 文件上传（单/多文件） |
| `icon` / `color` | icon | 图标选择器 / 颜色选择器 |
| `number` / `int` / `num` | add_num | Number 输入框 |
| `time` / `datetime` | refreshtime | 时间日期组件（类型要求 int） |
| `textarea` / `multiline` / `rows` | rows | Textarea 输入框（类型要求 varchar） |
| `status` / `state` / `type` | status | 单选框（类型要求 tinyint(1)/char(1)） |
| `switch` / `toggle` | log_switch | 开关组件（类型要求 tinyint(1)/int/enum/char(1)，1 为开、0 为关） |
| `content` / `editor` | content | 富文本编辑器（类型要求 text 系列） |

### 字段注释字典

CRUD 会将 `字段注释` 与 `表注释` 解析为字典（单选框/下拉选项等），格式为 `标题:值1=含义1,值2=含义2`：

```
状态:0=禁用,1=启用
菜单类型:tab=选项卡,link=链接,iframe=Iframe
扩展属性:none=无,add_rules_only=只添加为路由,add_menu_only=只添加为菜单
```

- 表注释示例：`会员组表` 将解析为 `会员组管理`。
- **不支持复合主键**，请为数据表设置单字段主键。

## 5. 后端开发规范

### 控制器基类

- 后台控制器继承 `app\common\controller\Backend`，前台控制器（`api` 应用）继承 `app\common\controller\Frontend`；两者均继承 `Api`（`Api` 继承 `app\BaseController`）。
- `Api` 基类按路由加载控制器语言包、检查数据库连接、IP 检查、时区设定。
- `Backend` 基类自动根据 token 完成管理员登录与鉴权，提供 `queryBuilder()` 构建查询 SQL（快速搜索、公共搜索、排序等）。

典型后台控制器结构（参考 `app\admin\controller\user\User.php`）：

```php
namespace app\admin\controller\user;

use Throwable;
use app\common\controller\Backend;
use app\admin\model\User as UserModel;

class User extends Backend
{
    protected object $model;

    // 关联查询方法名（方法定义在模型中），如 user 关联 userGroup
    protected array $withJoinTable = ['userGroup'];

    // 新增/编辑时排除（忽略不入库）的字段
    protected string|array $preExcludeFields = ['last_login_time', 'login_failure', 'password', 'salt'];

    // 快速搜索字段
    protected string|array $quickSearchField = ['username', 'nickname', 'id'];

    public function initialize(): void
    {
        parent::initialize();
        $this->model = new UserModel();
    }
}
```

`Backend` 基类可重写的常用属性（父类中已有默认值，按需覆盖）：

| 属性 | 默认值 | 说明 |
| --- | --- | --- |
| `noNeedLogin` | `[]` | 无需登录即可访问的方法 |
| `noNeedPermission` | `[]` | 无需鉴权的方法 |
| `preExcludeFields` | `[]` | 新增/编辑时排除不入库的字段 |
| `quickSearchField` | `'id'` | 快速搜索字段 |
| `modelValidate` | `true` | 是否开启模型验证 |
| `modelSceneValidate` | `false` | 是否开启模型场景验证 |
| `withJoinTable` | `[]` | 关联查询方法名（定义在模型中） |
| `withJoinType` | `'LEFT'` | 关联 JOIN 方式 |
| `dataLimit` / `dataLimitField` | `false` / `'admin_id'` | 数据权限控制（如 `personal`、`parent`），启用时表内需有 `admin_id` 字段 |
| `indexField` | `['*']` | 查看请求返回的主表字段 |

### Backend Trait（通用增删改查实现）

- `app\admin\library\traits\Backend` Trait 实现了 `index`（查看）、`add`（添加）、`edit`（编辑）、`del`（删除）、`sortable`（排序）等方法，已自动引入后台控制器基类。
- 需要重写某个方法时，**将 Trait 中的方法复制到控制器后修改**（如 `User::index()` 复制 Trait 的实现并增加 `withoutField('password,salt')`）。
- 新建控制器时建议直接继承 `Backend`，即可免费获得完整的 CRUD 能力。

### 快捷响应方法

定义于 `Api` 基类，只用于控制器内：

```php
// 操作成功：$msg 提示消息、$data 返回数据、$code 状态码
$this->success('操作成功~', ['title' => '标题']);
$this->success('操作成功~');

// 操作失败（仅 code 默认值与 success 不同）
$this->error('操作失败~');
```

- `success()` 默认 `code = 1`，`error()` 默认 `code = 0`，均基于 `result($msg, $data, $code, $type, $header, $options)` 封装。
- 内部抛出 `HttpResponseException`：调用后**无需 `return`**（流程立即终止）；注意它属于异常抛出，**不建议在 try 代码块内使用**（会跳出 try 流程）。
- 不影响直接使用 `return json();` 返回数据。

### 权限类实例

登录管理员还是会员，取决于控制器继承 `Backend`（管理员）还是 `Frontend`（会员）：

```php
$this->auth->isLogin();          // 是否登录
$this->auth->getUser();          // 会员数据模型
$this->auth->getAdmin();         // 管理员数据模型
$this->auth->getToken();         // 登录 token
$this->auth->getRefreshToken();  // 刷新 token
$this->auth->getInfo();          // 管理员信息（仅允许输出的字段）
$this->auth->getUserInfo();      // 会员信息（仅允许输出的字段）
$this->auth->isSuperAdmin();     // 是否超管（拥有后台所有权限节点）
$this->auth->isSuperUser();      // 是否拥有所有会员中心权限的会员
```

### 修改会员余额/积分（重要）

**不要直接修改会员表的余额/积分字段**，应通过日志模型写入一条日志（余额/积分自动增减），且**必须开启事务**：

```php
use think\facade\Db;
use app\admin\model\UserMoneyLog; // 余额日志模型
use app\admin\model\UserScoreLog; // 积分日志模型

Db::startTrans();
try {
    UserMoneyLog::create([
        'user_id' => 1,
        'money'   => 1,
        'memo'    => '测试添加 1 元',
    ]);
    Db::commit();
} catch (Throwable $e) {
    Db::rollback();
    $this->error($e->getMessage());
}
```

积分同理，将 `UserMoneyLog` 换为 `UserScoreLog`，字段 `money` 换为 `score`。

### 模型验证器

- 验证器放在 `app/admin/validate` 目录，命名与模型对应（如模型 `Admin` → 验证器 `app\admin\validate\Admin.php`；注意并非每个模型都有同名验证器，未配置时仅跳过模型验证）。
- 基类自动通过 `str_replace("\\model\\", "\\validate\\", get_class($this->model))` 查找验证器，配合 `modelValidate = true` 生效；场景验证（add/edit）需开启 `modelSceneValidate` 并在验证器中定义 `scene()`。

### 后端多语言

- 代码内使用 `__('key')` / `__('key', ['param'])` 进行翻译。
- 后端语言包位于 `app/admin/lang/` 下：目录化语言包按控制器路径组织（如 `app/admin/lang/zh-cn/auth/group.php`、`app/admin/lang/zh-cn/user/moneylog.php`），公共翻译位于聚合文件 `app/admin/lang/zh-cn.php`（英文对应 `app/admin/lang/en.php`）。

## 6. 前端开发规范

### 网络请求

- 所有请求函数建议放在 `web/src/api/` 目录。
- 使用 `createAxios()`（封装自 axios，位于 `web/src/utils/axios.ts`），自动携带 token、无感刷新 token、携带语言、取消重复请求、自动错误/成功提示：

```ts
import createAxios from '/@/utils/axios'

export function getMenuRules() {
    return createAxios({
        url: '/admin/auth.menu/index',
        method: 'get',
    })
}
```

- `options` 常用项：`showSuccessMessage`（成功提示）、`showCodeMessage`（code!=1 时提示）、`loading`（loading 层）等。
- 接口返回 `code=1` 为成功（`error()` 默认 `code=0` 为失败），前端响应拦截对 `code !== 1` 的响应自动处理提示/刷新。

### baTableApi（接口方法生成器）

`web/src/api/common.ts` 中的 `baTableApi` 类，传入控制器 URL 即可快速生成该控制器的 `增、删、改、查、排序` 五个请求方法（index/add/edit/del/sortable）：

```ts
import { baTableApi } from '/@/api/common'

// 传入后台控制器 url，如会员管理
let baApi = new baTableApi('/admin/user.user/')

// 可按需修改某个接口地址（改地址不能满足时，可在 /src/api 下自定义请求函数）
baApi.actionUrl.set('del', '/admin/user.user/delNew')
```

- 接口 URL 规则形如 `/admin/auth.menu/index`、`/admin/user.user/index`（`index.php` 前缀由请求封装的基础地址 `VITE_AXIOS_BASE_URL` 处理，官方示例中亦写作 `/index.php/admin/auth.menu/index`）。
- `baTableApi` 实例通常直接传给 `baTable` 类使用。

### baTable 表格管家类

表格页面以 `baTable` 类（`web/src/utils/baTable.ts`）为核心：它提供表格数据、响应表格事件、预设查看/编辑前后置钩子。典型用法（参考 `web/src/views/backend/user/user/index.vue`）：

```ts
import { provide } from 'vue'
import baTableClass from '/@/utils/baTable'
import { baTableApi } from '/@/api/common'
import { defaultOptButtons } from '/@/components/table'

const baTable = new baTableClass(
    new baTableApi('/admin/user.user/'),
    {
        column: [
            { type: 'selection', align: 'center', operator: false },
            { label: t('user.user.User name'), prop: 'username', align: 'center', operator: 'LIKE', operatorPlaceholder: t('Fuzzy query') },
            {
                label: t('user.user.Gender'),
                prop: 'gender',
                align: 'center',
                render: 'tag',
                replaceValue: { '0': t('Unknown'), '1': t('user.user.male'), '2': t('user.user.female') },
            },
            { label: t('Create time'), prop: 'create_time', align: 'center', render: 'datetime', sortable: 'custom', operator: 'RANGE', width: 160 },
            {
                label: t('Operate'),
                align: 'center',
                width: '100',
                render: 'buttons',
                buttons: defaultOptButtons(['edit', 'delete']),
                operator: false,
            },
        ],
        dblClickNotEditColumn: [undefined],
    },
    {
        defaultItems: { gender: 0, status: 'enable' }, // 添加表单默认值
    }
)

baTable.mount()   // 相当于表格 onMounted，初始化表格
baTable.getData() // 获取表格数据

provide('baTable', baTable) // 将实例提供给上下文，子组件（PopupForm 等）可 inject 使用
```

**表格列常用属性**：

- `prop`：字段名；`label`：列标题。
- `render`：单元格渲染器（位于 `src/components/table/fieldRender/`），可选 `'buttons' | 'color' | 'customRender' | 'customTemplate' | 'datetime' | 'icon' | 'image' | 'images' | 'switch' | 'tag' | 'tags' | 'url' | 'slot'`。
- `operator`：公共搜索操作符，如 `'LIKE'`、`'RANGE'`（范围）、`'IN'`、`'FIND_IN_SET'`、`'eq/ne/gt/egt/lt/elt'` 等；`operator: false` 关闭该字段公共搜索。
- `replaceValue`：值替换字典（同时用于单元格渲染和公共搜索下拉选项），如 `{ '0': '未知', '1': '女', '2': '男' }`。
- `comSearchRender`：公共搜索渲染方式（`'select' | 'remoteSelect' | 'time' | 'date' | 'datetime' | 'customRender' | 'slot'`）。
- `remote`：公共搜索为远程下拉时的配置 `{ pk, field, remoteUrl, params, multiple }`。
- `buttons`：`render: 'buttons'` 时的操作按钮数组，常用 `defaultOptButtons(['edit', 'delete'])` 生成，可在渲染前增删改。

**操作前后置钩子**（前置返回 `false` 可取消原操作）：

```ts
// 获取表格数据前 / 后
baTable.before.getData = () => {}
baTable.after.getData = ({ res }: { res: ApiResponse }) => {}

// 删除前 / 后
baTable.before.postDel = ({ ids }: { ids: string[] }) => {}
baTable.after.postDel = ({ res }: { res: ApiResponse }) => {}

// 获取编辑行数据后（此时 baTable.form.items 已赋值）
baTable.after.getEditData = ({ res }: { res: ApiResponse }) => {}

// 表单提交前 / 后
baTable.before.onSubmit = ({ formEl, operate, items }: { formEl: FormInstance | undefined; operate: string; items: anyObj }) => {}
baTable.after.onSubmit = ({ res }: { res: ApiResponse }) => {}

// 表格初始化前（getData 别名 getIndex、getEditData 别名 requestEdit 亦可用）
baTable.before.mount = () => {}
```

### 页面模板

```vue
<template>
    <div class="default-main ba-table-box">
        <el-alert class="ba-table-alert" v-if="baTable.table.remark" :title="baTable.table.remark" type="info" show-icon />

        <!-- 表格顶部菜单：buttons 传入要显示的按钮名 -->
        <TableHeader
            :buttons="['refresh', 'add', 'edit', 'delete', 'comSearch', 'quickSearch', 'columnDisplay']"
            :quick-search-placeholder="t('Quick search placeholder', { fields: t('user.user.User name') + '/' + t('user.user.nickname') })"
        />

        <!-- 表格（el-table 的原生属性可直接加在 Table 标签上） -->
        <Table />

        <!-- 弹窗表单 -->
        <PopupForm />
    </div>
</template>

<script setup lang="ts">
import Table from '/@/components/table/index.vue'
import TableHeader from '/@/components/table/header/index.vue'
import PopupForm from './popupForm.vue'
// ...baTable 实例化见上文
</script>
```

### 表单（FormItem 组件）

`PopupForm.vue` 中使用 `FormItem`（`/@/components/formItem/index.vue`）构建表单项，它继承 `el-form-item` 属性，常用 `:label` + `type` + `v-model` + `prop` + `:rules`：

```vue
<el-form
    ref="formRef"
    @submit.prevent=""
    @keyup.enter="baTable.onSubmit(formRef)"
    :model="baTable.form.items"
    :rules="rules"
>
    <FormItem
        :label="t('user.user.nickname')"
        type="string"
        v-model="baTable.form.items!.nickname"
        prop="nickname"
        :placeholder="t('Please input field', { field: t('user.user.nickname') })"
    />

    <FormItem
        :label="t('user.user.Gender')"
        type="radio"
        v-model="baTable.form.items!.gender"
        prop="gender"
        :data="[{ value: 0, label: t('Unknown') }, { value: 1, label: t('user.user.male') }, { value: 2, label: t('user.user.female') }]"
    />

    <!-- 远程下拉：关联表字段（如 user_id） -->
    <FormItem
        :label="t('user.user.group')"
        v-model="baTable.form.items!.group_id"
        type="remoteSelect"
        :input-attr="{
            pk: 'id',            // 关联表主键，作为选项 value
            field: 'name',       // 关联表字段，作为选项 label
            'remote-url': '/admin/user.group/index', // 关联表控制器的 index 接口 URL
        }"
    />
</el-form>
```

`type` 可选：`string | password | number | radio | checkbox | switch | textarea | array | datetime | year | date | time | select | selects | remoteSelect | editor | city | image | images | file | files | icon | color`。

### 前端多语言

- 页面文本一律使用 `t('目录名.页面名.翻译名')`（vue-i18n），不要硬编码中文。
- 语言包文件位置：后台 `web/src/lang/backend/zh-cn/<目录>/<页面>.ts`（中文）与 `web/src/lang/backend/en/<目录>/<页面>.ts`（英文），目录结构须与页面视图目录一致（如 `user/user.ts` 对应 `views/backend/user/user`）。
- 全局公共语言包为 `web/src/lang/globs-zh-cn.ts` / `globs-en.ts`（如 `t('Operate')`、`t('Create time')`）。

### 前端鉴权

- 函数式：`import { auth } from '/@/utils/common'`，`auth('add')` 会以当前路由 path 拼接（如当前 `/admin/auth/group`，则检查 `auth/group/add`）；也可传对象精确鉴权：`auth({ name: '/admin/auth/group', subNodeName: '/admin/auth/group/add' })`。
- 指令式：`<el-button v-auth="'add'">添加</el-button>`（拼接为 `auth/menu/add`）。
- 菜单规则的 `name` 会自动注册为路由 name；规则无权限的按钮自动隐藏。

## 7. 数据库迁移规范（Phinx）

涉及数据表结构变更、枚举修改、测试数据导入、菜单规则等数据库改动时，推荐以 **Phinx 迁移** 形式提交，而非直接执行 SQL。

### 基本规则

- 迁移文件位于 `database/migrations/`，命名 `YYYYMMDDHHMMSS_<动作_表名>.php`（如 `20260803000005_add_asset_menu_rules.php`），类名与文件名对应（下划线转驼峰：`AddAssetMenuRules`），`use think\migration\Migrator;` 并继承 `Migrator`。
- 执行方式：项目根目录 `php think migrate:run`（后台「开发 → WEB 终端」亦可）；迁移记录保存在 **`wx_migrations`** 表（数据库表前缀 + `migrations`），已执行过的迁移不会重复执行，重复运行命令即幂等。
- 迁移按文件时间戳升序执行（`version_order = creation_time`）；执行失败会中断且不写入记录，修正后重新 `migrate:run` 即可。
- 方法使用 `up()`（执行升级）；`down()` 可选（回滚）。

### 用 `$this->execute()` 执行原始 SQL

结构/数据变更一律直接执行 SQL，**不要用 `$this->table()` 链式写**（复杂 DDL、枚举、数据导入易产生歧义）：

```php
public function up(): void
{
    // 表不存在则跳过（幂等守卫）
    if (!$this->hasTable('admin_rule')) {
        return;
    }

    // 1) 简单 SQL：双引号字符串即可
    $this->execute("ALTER TABLE `wx_survey_question` MODIFY COLUMN `type` ENUM('radio','checkbox','text','textarea','rate','date','image','datetime','select') NOT NULL DEFAULT 'radio' COMMENT '类型';");

    // 2) 字段级守卫：不存在才新增
    if ($this->hasTable('survey') && !$this->table('survey')->hasColumn('deadline')) {
        $this->execute("ALTER TABLE `wx_survey` ADD COLUMN `deadline` bigint(16) UNSIGNED DEFAULT NULL COMMENT '截止时间' AFTER `status`;");
    }

    // 3) 数据导入：INSERT IGNORE 幂等 + 显式列名（兼容已存在表结构的额外列）
    $sql = <<<'SQL'
INSERT IGNORE INTO `wx_survey` (`id`, `title`, `description`, `image`, `status`, `weigh`, `create_time`, `update_time`) VALUES (1, '测试1', '这是一个测试', '/storage/default/xxx.png', '1', 1, 1768470853, 1768470853);
SQL;
    $this->execute($sql);
}
```

要点：

- **一条 `execute()` 只执行一条语句**；多条独立语句逐条执行，可放入数组 `foreach` 循环（如菜单权限的多条 INSERT）。
- 含单引号/反斜杠的 SQL（如 JSON 数据 `'{\"1\":[\"http:\/\/...\"]}'`）用 **nowdoc（`<<<'SQL'`）** 包裹，原样传给 MySQL 解析，避免 PHP 转义歧义；无特殊字符的简单 SQL 可用双引号字符串。
- 数据导入使用 **`INSERT IGNORE INTO`**（重复执行自动跳过已存在行），并写**显式列名**——表可能已存在额外列（如 `create_uid`），缺列名会导致"列数不匹配"报错。
- `$this->execute()` 直连数据库，**表名须带前缀 `wx_`**（如 `wx_survey`、`wx_admin_rule`）；而 `$this->hasTable('survey')` 等守卫用的是**不带前缀**的表名。
- 写 SQL 前先确认实际表结构（`SHOW COLUMNS FROM 表名`）：框架表字段以实际为准，如 `admin_rule` 用 `update_time`/`create_time` 而非 `updatetime`/`createtime`；菜单规则写入用 `INSERT ... SELECT ... WHERE NOT EXISTS` 防重复。

## 8. CRUD 代码生成

### 可视化 CRUD（推荐，系统内置）

- 后台「开发→可视化 CRUD」打开设计器，拖拽设计字段组件，一次生成 **控制器、模型、验证器、语言包、表格、表单**。
- 支持从已有数据表开始：点击「选择数据表」，系统按[数据表设计规范](#4-数据表设计规范)解析字段并生成。
- 「高级配置」中可设置代码生成位置（任意层级目录）；后台菜单规则与代码位置目录层级一致，已存在的菜单规则不会被覆盖。
- 生成后代码位置：后端 `app/admin/controller/`（含对应 model/validate/lang），前端 `web/src/views/backend/` 对应目录；开发环境下热更新，立即可在 `http://localhost:1818/#/admin` 看到效果。
- 修改设计：在「CRUD 记录」中使用「复制设计」再覆盖（表名不变时再次生成即为覆盖）。

### 命令行 CRUD（可选，需在模块市场安装命令行 CRUD 模块）

```bash
# 生成 test_build 表的 CRUD，同时生成后台菜单（权限节点）
php think crud -t test_build

# 删除指定表的 CRUD 代码
php think crud -t test_build -d 1

# 指定控制器目录（对应视图目录）
php think crud -t test_build -c mysite/build

# 指定模型名
php think crud -t test_build -m testmodel

# 关联 user 表：外键 user_id、主键 id（可多次 -r 关联多表）
php think crud -t test_build -r user -k user_id -p id

# 指定图片上传组件后缀
php think crud -t test_build --imagefieldsuffix=image --imagefieldsuffix=img
```

常用参数：`-t/--table` 表名（必填）、`-c/--controller` 控制器、`-m/--model` 模型、`-f/--fields` 显示字段、`-o/--commonmodel` 生成到 common 模块、`-r/--relation` 关联表、`-k` 外键、`-p` 主键、`-l` 关联显示字段、`-a` 关联模式（hasone/belongsto）、`-s` 远程 select 显示字段、`-d/--delete` 删除、`--quicksearchfield` 快速搜索字段、`--sortfield` 默认排序字段、`--ignorefields` 排除字段、`--force` 覆盖模式；后缀识别参数如 `--radiofieldsuffix`、`--checkboxfieldsuffix`、`--switchfieldsuffix`、`--editorfieldsuffix`、`--remoteselectfieldsuffix`、`--imagefieldsuffix`、`--imagesfieldsuffix` 等。

## 9. 菜单规则与权限

### 菜单规则管理

- 后台位置：**权限管理 → 菜单规则管理**；前台位置：**会员管理 → 菜单规则管理**。
- **规则名称**：使用英文，自动注册为 Vue Router 的**路由 name**；推荐直接使用 `/` 分隔目录层次，并以**小写开头的驼峰**命名多个单词（如 `auth/group`、`user/user`）。
- **路由路径**：建议与规则名称保持一致。
- **菜单类型**：`选项卡`（背后对应 vue 与控制器文件）、`链接`、`Iframe`（外站链接）。
- **组件路径**：填写实际组件路径，如 `/src/views/backend/auth/group/index.vue`。
- **规则权重**：数字越大排序越靠前（系统内置"控制台"权重为 999）。
- **扩展属性**：可配置只自动注册为路由或只注册为菜单，实现"同一路由下多个菜单"（如 `auth/menu` 只添加为路由，`auth/menu/:a`、`auth/menu/:b/:c` 只添加为菜单）。
- **缓存**：开启后使用 `keep-alive` 包裹标签页，切换标签保留状态（如表单输入）。
- **上级菜单规则**：上一级规则不能是按钮，表格中以树状结构展示上下级关系。

### 权限的实现方式

1. 管理员访问后台时，请求 `/admin/index/index`（文档示例为 `/index.php/admin/index/index`）获取后台**菜单规则**，请求从 `/src/layouts/backend/index.vue` 发出。
2. 服务端通过 `\ba\Auth` 类只返回该管理员拥有的菜单规则（权限节点）。
3. WEB 端将菜单规则解析为 Vue-Router 路由并自动注册，同时记录各页面的"页面按钮"权限。
4. 前端鉴权时，将当前路由与要验证的权限名拼接，与"管理员页面按钮权限"比对，未找到则隐藏按钮。

### 权限分配

- 为管理员分配权限：**权限管理 → 角色组管理**。
- 为会员分配权限：**会员管理 → 会员分组管理**。
- 前端验权展示请使用 `auth()` 函数（`/@/utils/common`）或 `v-auth` 指令。

## 10. 开发流程建议

### AI 辅助开发的推荐顺序

1. **设计数据表**：遵循[数据表设计规范](#4-数据表设计规范)（字段类型/后缀/注释字典），保证单字段主键、含 `weigh`（排序）、`create_time`/`update_time`（bigint）。
2. **生成后端代码**：优先使用后台「可视化 CRUD」拖拽生成（控制器/模型/验证器/语言包/菜单规则一次到位）；或手写控制器继承 `Backend`、模型与验证器放对应目录。
3. **编写前端页面**：`web/src/views/backend/<目录>/<页面>/index.vue`（TableHeader + Table + baTable 实例）与 `popupForm.vue`（FormItem 表单），文本使用 `t()` 语言包。
4. **配置菜单规则**：在「权限管理→菜单规则管理」添加规则（名称/路由路径/组件路径/上级规则），并为角色分配权限。
5. **补充语言包**：后端 `app/admin/lang/zh-cn/...`（`__()`）、前端 `web/src/lang/backend/zh-cn/<目录>/<页面>.ts`（`t()`）与英文包。

### 常见注意事项

- 开发环境：`php think run` + `web` 目录 `npm run dev`，页面访问 `http://localhost:1818/#/admin`（接口报错信息仅在开发环境详细显示）。
- 接口约定：`code=1` 成功、`code=0` 失败（`error()` 默认值）；前端对 `code !== 1` 自动提示错误。
- 表格双击行进入编辑；表单内按 `Enter` 触发保存（`@keyup.enter="baTable.onSubmit(formRef)"`）。
- 需要扩展表格行为时优先使用 `baTable.before.xxx` / `baTable.after.xxx` 钩子，避免直接改框架文件。
- 修改会员余额/积分必须通过 `UserMoneyLog` / `UserScoreLog` 日志模型并开启事务，禁止直接改 `user` 表字段。
- 快捷响应方法（`success`/`error`）抛出 `HttpResponseException`，调用后无需 `return`，避免在 try 代码块内使用。
- 一个 vue 文件中只能渲染一个表格（一个 baTable 实例），多表格场景需拆分为独立 vue 文件导入。
- 后端代码文件名/类名/URL 中 `user.user` 表示 `user` 应用（目录）下的 `User` 控制器；大小写驼峰转小写下划线是框架约定。

## 使用示例

参见 [典型任务与预期结果](examples/basic.md)。
