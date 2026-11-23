# config.json 模块配置

> 来源：BuildAdmin 官方文档《模块配置（依赖声明等）》（`doc.buildadmin.com/senior/module/configJson.html`），抓取日期 2026-11-23。内容为压缩整理，不是原文照搬。

在模块根目录放 `config.json`，声明依赖、禁用时受保护的文件和 Composer 配置。系统在安装时按此文件安装依赖。JSON 不支持注释，写进包的必须是合法 JSON。

## 字段

| 字段 | 作用 |
| --- | --- |
| `require` | Composer 生产依赖 |
| `require-dev` | Composer 开发依赖 |
| `dependencies` | `web` 工程 NPM 生产依赖 |
| `devDependencies` | `web` 工程 NPM 开发依赖 |
| `nuxtDependencies` | `web-nuxt` 工程 NPM 生产依赖 |
| `nuxtDevDependencies` | `web-nuxt` 工程 NPM 开发依赖 |
| `protectedFiles` | 禁用模块时不能删除的文件 |
| `composerConfig` | 合并进项目 `composer.json` 的 `config` 字段 |

关键点：`web` 与 `web-nuxt` 是两套独立的前端工程和 NPM 依赖。只写 `dependencies` 不会把依赖装到 Nuxt 工程，反之亦然；模块用到哪个工程就写到对应字段。`web-nuxt/nuxt.config.ts` 之类共享文件被覆盖时，也要在交付说明里写明依赖和构建要求。

## 示例

以下为官方示例去掉注释后的形式，按实际依赖替换：

```json
{
    "require": {
        "nelexa/zip": "^3.3",
        "guzzlehttp/guzzle": "^6.3"
    },
    "require-dev": {
        "symfony/var-dumper": "^4.2",
        "topthink/think-trace": "^1.1"
    },
    "dependencies": {
        "vue-i18n": "~9.1.9",
        "vue-router": "~4.0.11"
    },
    "devDependencies": {
        "@types/lodash": "~4.14.180",
        "@types/node": "~17.0.9"
    },
    "nuxtDependencies": {
        "vue-i18n": "~9.1.9",
        "vue-router": "~4.0.11"
    },
    "nuxtDevDependencies": {
        "@types/lodash": "~4.14.180",
        "@types/node": "~17.0.9"
    },
    "protectedFiles": ["app/common.php", "app/admin/event.php"],
    "composerConfig": {
        "allow-plugins": {
            "easywechat-composer/easywechat-composer": true
        }
    }
}
```

## protectedFiles

模块禁用时，系统默认删除模块新增的文件，并把安装时备份的冲突文件还原到安装前状态。`protectedFiles` 用于列出**禁用时不能删除**的文件，例如模块往某个系统核心文件里追加过必要内容，删除会让站点无法运行。

- 只列真正需要的文件；越少越好，列表本身不解决冲突。
- 系统在安装模块前会备份冲突文件，禁用时自动恢复，因此被覆盖过的核心文件通常会回到安装前版本，`protectedFiles` 不是覆盖的免死金牌。
- 官方文档只描述**禁用**语义；卸载时这些文件是否保留没有说明。需要在卸载后保留的文件，清理逻辑要写进 `uninstall` 并实测。
- `composerConfig` 只用于合并 `composer.json` 的 `config` 字段，适合 `allow-plugins` 这类安装器设置。

## 不在 config.json 里的上架声明

以下三类信息**不在 `config.json` 中配置**（目前也不支持开发者自行维护）：模块小版本对 BuildAdmin 版本的要求、依赖市场上其他模块、与其他模块互斥。上架时把这三类关系同模块包一起发给 `hi@buildadmin.com`，由官方在官网后台维护。不要为它们自造 JSON 字段。

## 打包检查

- JSON 合法：无注释、无尾逗号，能被 `json_decode` 解析。
- 前端依赖写在正确字段：`web` 用 `dependencies`/`devDependencies`，`web-nuxt` 用 `nuxtDependencies`/`nuxtDevDependencies`。
- `require` 中的 Composer 包与 PHP 代码里的 `use` 对应，没有漏声明。
- `protectedFiles` 中每个路径在包内或项目中确实存在。
