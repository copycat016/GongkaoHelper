# GongkaoHelper

个人公考学习辅助系统，核心功能：统一录入器（图片 OCR / PDF / 粘贴文本）、申论 PDF 结构化与 AI 批改、三层学习规划、番茄钟、学习日志、本地音乐播放器，以及 LLM / OCR / PDF / 备份 / 主题配置。

## 技术栈

- **前端**：Vite + React 19 + Ant Design 6 + React Router 7
- **后端**：Go + Gin + GORM
- **数据库**：SQLite（默认）/ PostgreSQL
- **PDF 解析**：Poppler `pdftotext`（优先）+ Go `pdf` 库（回退）
- **OCR**：百度 OCR

## 快速启动

### 1. 启动数据库（可选）

默认使用 SQLite，无需额外配置：

```bash
cd backend
cp .env.example .env
# 保持 DB_DRIVER=sqlite 即可
# 开发模式首次登录：admin / 123456
```

如需 PostgreSQL：

```bash
cd backend
cp .env.example .env
# 修改 .env：DB_DRIVER=postgres
docker compose --env-file .env up -d
```

### 2. 启动后端

```bash
cd backend
set -a && source .env && set +a
go run ./cmd/server
```

### 3. 启动前端

```bash
npm run dev -- --host 0.0.0.0 --port 21073
```

访问：`http://localhost:21073`

## 端口约定

| 服务 | 端口 |
|------|------|
| 前端 Vite | `21073` |
| 后端 Gin | `21080` |
| PostgreSQL | `21432` |

## 验证命令

```bash
# 后端测试
cd backend && go test ./...

# 前端构建
cd .. && npm run build

# 接口冒烟
curl http://localhost:21080/api/health
curl http://localhost:21080/api/db/ping
```

## 文档索引

完整索引见 [`docs/README.md`](docs/README.md)。

| 文档 | 内容 |
|------|------|
| [`docs/AGENT_CONTEXT.md`](docs/AGENT_CONTEXT.md) | **AI 接手上下文** — 用于新会话快速接手、降低 token 消耗 |
| [`docs/QUICKSTART.md`](docs/QUICKSTART.md) | **快速启动** — 启动顺序、端口、验证命令、依赖说明 |
| [`docs/MAINTENANCE.md`](docs/MAINTENANCE.md) | **维护手册** — 目录结构、约定、API 状态、申论流程、常见坑 |
| [`docs/design-system.md`](docs/design-system.md) | **前端设计系统** — token 三层架构、母题组件、清洗准则 |
| [`backend/docs/backend-dev.md`](backend/docs/backend-dev.md) | **后端开发文档** — 数据库选择、环境变量、接口测试示例 |

## 提交前检查

```bash
cd backend && gofmt -w cmd internal && go test ./...
cd .. && npm run build
```
