# WEB 引导程序（webBootstrap.stub）

> 来源：BuildAdmin 官方文档《向 main.ts 插入代码》（`doc.buildadmin.com/senior/module/webBootstrap.html`），抓取日期 2026-11-23。内容为压缩整理，代码为官方示例。

自 v1.1.0 起，模块可以向系统 `main.ts`、`App.vue` 的指定位置插入代码；自 v2.3.2 起支持 `web-nuxt/app.vue`。支持的能力：

1. 向 `main.ts`、`App.vue`、`web-nuxt/app.vue` 插入自定义 `import`；
2. 向 `main.ts` 的 `start` 函数结尾插入代码；
3. 向 `App.vue` 的 `onMounted` 函数结尾插入代码；
4. 向 `web-nuxt/app.vue` 底部插入代码；
5. 安装模块时自动写入，禁用模块时自动清理。

## 写法

把要插入的代码按块写进模块根目录的 `webBootstrap.stub`。一共 6 个代码块，只写需要的；任何块都可以是单行或多行。文件换行符必须是 LF（`\n`），否则可能出现换行符相关问题。

```text
#main.ts import code start#
console.log('main.ts import 1')
console.log('main.ts import 2')
#main.ts import code end#

#main.ts start code start#
    console.log('main.ts start 1')
    console.log('main.ts start 2')
    console.log('main.ts start 3')
#main.ts start code end#

#App.vue import code start#
console.log('App.vue import 1')
console.log('App.vue import 2')
#App.vue import code end#

#App.vue onMounted code start#
    console.log('App.vue onMounted1')
    console.log('App.vue onMounted2')
    console.log('App.vue onMounted3')
#App.vue onMounted code end#

#web-nuxt/app.vue import code start#
console.log('web-nuxt/app.vue import 1')
console.log('web-nuxt/app.vue import 2')
#web-nuxt/app.vue import code end#

#web-nuxt/app.vue start code start#
console.log('web-nuxt/app.vue start 1')
console.log('web-nuxt/app.vue start 2')
console.log('web-nuxt/app.vue start 3')
#web-nuxt/app.vue start code end#
```

标记行必须一字不差（大小写、空格、`#` 个数），否则安装器匹配不到，代码不会写入。块内代码按原样插入对应的函数或文件位置。第 4 条能力说的是"`web-nuxt/app.vue` 底部"，对应标记名却是 `start`；实际插入位置以安装器实现为准，插入后看一眼生成的文件确认位置。

## 与直接覆盖文件的关系

需要改动 `main.ts` 时优先用 `webBootstrap.stub` 注入，而不是在包里放 `web/src/main.ts` 直接覆盖：

- 覆盖会替换整个启动文件，与系统和其他模块的修改冲突，禁用时只能靠备份还原；
- `.stub` 由安装器按标记插入、按标记清理，冲突面小，多个模块可以各自注入。

需要插入的逻辑超出"加一段代码"（例如改写启动流程）时，先看目标项目是否已有其他注入机制或钩子，再决定是否覆盖，并在交付说明里写明覆盖了哪个文件、影响范围和验证方式。

## 版本限制

- `main.ts`、`App.vue` 的注入需要 BuildAdmin ≥ v1.1.0；
- `web-nuxt/app.vue` 的注入需要 BuildAdmin ≥ v2.3.2。

目标项目低于对应版本时，不要假定该块会生效；以目标项目安装器代码为准，必要时改用覆盖文件或其他接入方式，并在交付说明中写明。

## 打包检查

- `webBootstrap.stub` 位于 ZIP 根目录，换行符为 LF。
- 6 个块的开闭标记没有拼写错误，也没有写在别的块内部。
- 插入的 `import` 所引用的 NPM 包，已写进 `config.json` 的对应依赖字段（`web` 工程用 `dependencies`，`web-nuxt` 工程用 `nuxtDependencies`，见 [config-json.md](config-json.md)）。
- 插入的代码不依赖尚未安装的模块或环境变量；交付说明写清启用后需要重启/重新构建前端。
