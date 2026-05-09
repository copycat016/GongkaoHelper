# Gkweb Project Checklist

## 1. 启动顺序

```bash
cd /vol1/1000/Code/Gkweb/backend
cp .env.example .env
docker compose --env-file .env up -d

set -a
source .env
set +a
go run ./cmd/server
```

```bash
cd /vol1/1000/Code/Gkweb
npm run dev -- --host 0.0.0.0 --port 21073
```

默认端口：

- Frontend: `21073`
- Backend: `21080`
- PostgreSQL host port: `21432`

## 2. 验证命令

```bash
cd backend
go test ./...
```

```bash
cd ..
npm run build
```

接口冒烟：

```bash
curl http://localhost:21080/api/health
curl http://localhost:21080/api/db/ping
```

## 3. 当前主功能

- 首页总览：今日任务、工具入口、统计、已完成任务。
- OCR 识题：前端流程骨架，后端待接 PaddleOCR / VL 模型。
- 错题库：后端 CRUD 已有，前端仍需接真实接口。
- 申论批改：前端骨架，后端待实现。
- 番茄钟：可写入后端，并联动学习日志和每日任务。
- 音乐播放器：服务器上传、歌单、播放。
- 学习日志：番茄钟自动写入，支持基础统计。
- 学习计划：新增、完成、删除。
- 日历：聚合计划、番茄钟记录、错题复习。
- 配置：LLM、Prompt、OCR 配置、PDF 能力预留、数据备份导出。

## 4. 后端结构

```txt
backend/
├─ cmd/server/main.go
├─ internal/config
├─ internal/database
├─ internal/handlers
├─ internal/middleware
├─ internal/models
├─ internal/response
├─ internal/routes
├─ internal/services
└─ uploads
```

约定：

- `models` 只放 GORM 数据模型。
- `services` 放业务逻辑和数据库查询。
- `handlers` 只做参数解析、调用 service、返回响应。
- `routes` 只注册路由，不写业务逻辑。
- 所有业务表保留 `user_id`。
- 统一返回使用 `internal/response`。
- 临时用户通过 `X-User-ID`，默认 `1`，后续替换为认证中间件。

## 5. 前端结构

```txt
src/
├─ api
├─ components
├─ pages
├─ styles
├─ App.jsx
└─ AppRoutes.jsx
```

约定：

- 页面放在 `src/pages`。
- 可复用组件放在 `src/components`。
- 后端请求只从 `src/api` 发出。
- `App.jsx` 只保留应用入口，不再放业务页面。
- 主导航从 `src/components/Sidebar.jsx` 维护。
- 顶部页面标题从 `src/components/AppLayout.jsx` 的 `pageMeta` 维护。

## 6. API 模块状态

已接真实后端：

- `llm`
- `prompts`
- `mistakes` 后端已完成，前端待完整接入
- `pomodoro`
- `logs`
- `plans`
- `calendar`
- `music`

仍是 mock / 待实现：

- `ocr`
- `questions`
- `essay`
- `pdf`
- `theme` 目前使用 localStorage

## 7. 下一步建议

优先级建议：

1. 错题库前端接真实后端。
2. OCR 后端基础：上传图片、任务表、PaddleOCR endpoint 调用。
3. LLM 拉取模型列表：OpenAI-compatible `/models`。
4. PDF 文本解析接口：先支持非扫描件。
5. 音乐库补删除、排序、权限。
6. 引入认证和 root/admin 权限模型。

## 8. 开发前检查

每次新增功能前确认：

- 是否需要新表，是否包含 `user_id`。
- 是否需要迁移到 `database.AutoMigrate`。
- 是否需要新增 `api/*.js`。
- 是否需要在 `Sidebar.jsx` 增加主入口。
- 是否只是能力预留，避免放进主侧边栏。
- 是否会引入大文件上传，上传目录是否在 `backend/uploads`。
- 是否需要后端环境变量，不要硬编码密钥和地址。

## 9. 提交前检查

```bash
cd backend && gofmt -w cmd internal && go test ./...
cd .. && npm run build
```

确认：

- 没有旧端口 `8080`、`5173` 被写进启动文档。
- 没有 API Key、真实密钥、私有地址进入代码。
- 前端页面没有重复标题组。
- 不需要开发的能力不要放进侧边栏主栏目。
