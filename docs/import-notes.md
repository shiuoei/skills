# 首次导入记录

日期：2026-11-21。

用户确认收录 buildadmin-dev、daily-report、mygo-dev。来源为当前机器已安装的同名技能，未覆盖原始安装，也未导入其他技能。

- buildadmin-dev：保留原有开发规范，新增使用示例。
- daily-report：去掉本机固定路径，改为用户指定目录优先、环境变量 DAILY_REPORT_DIR 次之，均无配置时询问；配置不再写回 SKILL.md，新增使用示例。
- mygo-dev：保留全部参考文档、模板和来源记录；把额外的 whenToUse 字段移入 metadata，保留内容；新增使用示例。

三个技能初始状态均为 experimental，表示本仓库尚未完成实际任务验收，并不表示原技能不可用。本次整理未运行 BuildAdmin 或 MyGo 项目、未生成真实日报。
