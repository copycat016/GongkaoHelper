# 文档索引

GongkaoHelper 的文档集中在本目录与 `backend/docs/`。按用途选读，不要把全部文档塞进同一个上下文。

## 按角色

- **第一次跑起来** → [QUICKSTART.md](QUICKSTART.md)
- **AI / 新人接手项目** → [AGENT_CONTEXT.md](AGENT_CONTEXT.md)（只贴一句「请先读 docs/AGENT_CONTEXT.md」即可）
- **改后端 / 加接口** → [MAINTENANCE.md](MAINTENANCE.md) + [../backend/docs/backend-dev.md](../backend/docs/backend-dev.md)
- **改前端样式 / 设计** → [design-system.md](design-system.md)

## 文档清单

| 文档 | 内容 | 更新频率 |
|------|------|----------|
| [AGENT_CONTEXT.md](AGENT_CONTEXT.md) | AI 接手上下文：当前状态、近期决策、待办焦点、快速入口 | 每次架构级改动后 |
| [QUICKSTART.md](QUICKSTART.md) | 快速启动：启动顺序、端口约定、验证命令、PDF/OCR 依赖 | 启动流程变化时 |
| [MAINTENANCE.md](MAINTENANCE.md) | 维护手册：目录结构、前后端约定、数据模型、API 状态、申论链路、常见坑 | 结构/约定变化时 |
| [design-system.md](design-system.md) | 前端设计系统：token 三层架构、双色语义、母题组件库、CSS 清洗准则 | 设计语言变化时 |
| [../backend/docs/backend-dev.md](../backend/docs/backend-dev.md) | 后端开发文档：数据库选择、环境变量、接口测试示例 | 后端约定变化时 |

> 项目总览见仓库根目录 [../README.md](../README.md)。
