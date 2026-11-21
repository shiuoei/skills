# 安装与更新

## 查找

打开仓库首页的分类导航，使用浏览器页面查找搜索“日报”“托盘”等关键词；也可以使用 GitHub 仓库搜索或编辑器的全文搜索。

查找和维护本仓库无需安装 Node.js 或 npm 依赖。

## 安装技能

本次整理只建立仓库副本，没有修改已经安装的技能。

把所需的整个技能目录复制到目标工具配置的技能目录，保留 SKILL.md、参考资料和模板的相对位置。以当前机器正在使用的共享技能目录为例，PowerShell 安装到尚不存在的目录：

```powershell
$skillName = 'daily-report'
$skillSource = Join-Path (Get-Location) "skills/$skillName"
$skillRoot = Join-Path $env:USERPROFILE '.agents/skills'
$skillTarget = Join-Path $skillRoot $skillName
if (Test-Path -LiteralPath $skillTarget) { throw '目标已存在，请先按下方更新步骤备份。' }
New-Item -ItemType Directory -Path $skillRoot -Force | Out-Null
Copy-Item -LiteralPath $skillSource -Destination $skillTarget -Recurse
```

目标工具使用其他目录时，修改 skillRoot。技能被加载后，可以使用首页的示例指令调用。

## 更新已安装版本

1. 在仓库内修改并检查；以后将仓库作为唯一维护源。
2. 对比已安装目录，先将需要保留的改动合并回仓库。
3. 将旧目录完整备份到技能发现目录之外。
4. 移走旧目录后复制整个仓库版本，避免合并复制残留旧文件。
5. 在目标工具中确认新版本已加载，再用一个典型任务验证。

本次不自动覆盖现有安装。尤其是 daily-report 的目录配置：仓库版通过环境变量 DAILY_REPORT_DIR 或用户提供的路径确定，不再写回技能文件。

## 日报目录

可在会话中说明“日报保存到某目录”，也可以在启动使用技能的工具前设置 DAILY_REPORT_DIR 环境变量。用户本次明确给出的目录优先。没有配置时，技能会询问保存位置。

## GitHub 托管

当前仓库没有远程地址，也尚未创建提交。选择 GitHub 仓库及公开/私有可见性后，在本目录执行：

```sh
git add .
git commit -m "chore: initialize personal skills catalog"
git branch -M main
git remote add origin <你的GitHub仓库地址>
git push -u origin main
```

上述命令以新建的空远程仓库为前提；若远程已有内容，先获取并对齐历史。
整体许可证尚未指定，发布前根据自己的复用意愿选择。
