# 维护手册

## 目录结构

```
GongkaoHelper/
├─ README.md                 # 项目概述、快速启动、端口约定
├─ docs/
│   ├─ QUICKSTART.md         # 快速启动
│   ├─ MAINTENANCE.md        # 维护手册（本文件）
│   └─ TODO.md               # ToDo 清单
├─ index.html                # Vite 入口 HTML
├─ package.json              # 前端依赖
├─ vite.config.js            # Vite 配置（端口 21073，代理 /api -> 21080）
│
├─ src/                      # 前端源码
│   ├─ main.jsx              # React 入口
│   ├─ App.jsx               # 根组件
│   ├─ AppRoutes.jsx         # 路由定义
│   ├─ index.css             # 全局样式
│   ├─ api/                  # 后端 API 封装（禁止页面直接写 fetch）
│   │   ├─ request.js        # 统一 fetch 封装、错误处理、response 解包
│   │   ├─ llm.js            # LLM Provider / Model
│   │   ├─ prompts.js        # Prompt 模板
│   │   ├─ mistakes.js       # 错题库
│   │   ├─ pomodoro.js       # 番茄钟
│   │   ├─ logs.js           # 学习日志
│   │   ├─ plans.js          # 学习计划
│   │   ├─ calendar.js       # 日历
│   │   ├─ tasks.js          # 首页今日任务兼容 API
│   │   ├─ planning.js       # 三层规划 API（阶段目标/周任务/日任务）
│   │   ├─ music.js          # 音乐
│   │   ├─ ocr.js            # OCR
│   │   ├─ pdf.js            # PDF 测试
│   │   ├─ essay.js          # 申论
│   │   ├─ backup.js         # 备份
│   │   ├─ theme.js          # 主题（localStorage）
│   │   ├─ questions.js      # 题库（mock）
│   │   └─ mockData.js       # Mock 数据
│   ├─ components/           # 可复用组件（非页面）
│   │   ├─ AppLayout.jsx     # 页面布局框架
│   │   ├─ Sidebar.jsx       # 主导航菜单
│   │   ├─ PageHeader.jsx    # 页面标题
│   │   ├─ StatCard.jsx      # 统计卡片
│   │   ├─ Pomodoro.jsx      # 番茄钟计时器
│   │   ├─ MusicProvider.jsx # 全局音乐状态
│   │   ├─ GlobalMusicBar.jsx
│   │   └─ ThemePanel.jsx    # 主题配置面板
│   ├─ pages/                # 页面组件（与路由一一对应）
│   │   ├─ Dashboard.jsx     # / 首页：今日 / 本周 / 阶段规划
│   │   ├─ IntakePage.jsx    # /intake 统一录入器（图片 OCR / PDF / 粘贴文本）
│   │   ├─ EssayReview.jsx   # /essay
│   │   ├─ PomodoroPage.jsx  # /pomodoro
│   │   ├─ MusicPlayer.jsx   # /music
│   │   ├─ StudyCenter.jsx   # /study 学习日志
│   │   ├─ StudyLogs.jsx     # 学习日志组件
│   │   ├─ AISettings.jsx    # /ai (LLM/Prompt/OCR/PDF/备份 Tab)
│   │   ├─ LLMSettings.jsx
│   │   ├─ PromptSettings.jsx
│   │   ├─ OCRQuestion.jsx   # 旧 /ocr 重定向到 /intake
│   │   └─ PDFParser.jsx     # 重定向到 /ai?tab=pdf
│   └─ styles/               # CSS 样式
│       ├─ theme.css
│       ├─ layout.css
│       ├─ components.css
│       ├─ pages.css
│       └─ pomodoro.css
│
└─ backend/
    ├─ cmd/server/main.go    # 唯一入口
    ├─ go.mod / go.sum
    ├─ docker-compose.yml    # PostgreSQL 容器
    ├─ .env.example          # 环境变量模板
    ├─ data/                 # SQLite 数据库文件
    ├─ uploads/              # 上传文件
    ├─ docs/
    │   └─ backend-dev.md    # 后端开发文档
    └─ internal/
        ├─ config/           # 配置读取
        ├─ database/         # 连接、AutoMigrate
        ├─ middleware/       # CORS、RequestID、Recovery
        ├─ response/         # 统一响应格式
        ├─ routes/           # 路由注册
        ├─ models/           # GORM 数据模型
        ├─ handlers/         # HTTP Handler（参数解析 → service → 响应）
        ├─ services/         # 业务逻辑与 DB 查询
        └─ parser/           # 申论解析层（纯函数，不依赖 HTTP）
```

## 前后端约定

### 前端

- **页面** 放 `src/pages/`，**可复用组件** 放 `src/components/`
- **所有后端请求** 只从 `src/api/*.js` 发出，禁止页面内直接 `fetch`
- **路由** 统一在 `AppRoutes.jsx` 维护
- **导航菜单** 在 `Sidebar.jsx` 维护
- **页面标题** 在 `AppLayout.jsx` 的 `pageMeta` 维护
- `App.jsx` 只保留应用入口

### 后端

- `models/` 只放 GORM 模型，所有业务表保留 `user_id`
- `services/` 放业务逻辑和 DB 查询
- `handlers/` 只做参数解析 → 调用 service → 返回响应
- `routes/` 只注册路由，不写业务逻辑
- `parser/` 是纯解析层，不依赖 HTTP/Gin
- 统一响应使用 `internal/response`
- 临时用户通过 `X-User-ID` 头传递，默认 `1`

## 端口约定

| 服务 | 端口 |
|------|------|
| 前端 Vite | `21073` |
| 后端 Gin | `21080` |
| PostgreSQL | `21432` |

## 数据模型速查

所有模型继承 `BaseModel`（ID, UserID, CreatedAt, UpdatedAt）。

| 表名 | 说明 |
|------|------|
| `llm_providers` | LLM 服务商配置 |
| `llm_models` | 模型列表 |
| `prompt_templates` | Prompt 模板 |
| `mistakes` | 错题记录 |
| `pomodoro_sessions` | 番茄钟 |
| `study_logs` | 学习日志 |
| `study_plans` | 学习计划 |
| `stage_goals` | 阶段目标（考前/阶段规划、起止日期、目标指标） |
| `weekly_tasks` | 周任务（归属阶段目标、周起始日、DDL、状态、进度聚合） |
| `daily_tasks` | 日任务（归属周任务/阶段目标、计划执行日、DDL、完成状态） |
| `music_playlists` | 歌单 |
| `music_tracks` | 音轨 |
| `ocr_tasks` | OCR 调用记录 |
| `essay_documents` | 申论文档 |
| `essay_sections` | 语义区域（material/question/answer/analysis） |
| `essay_section_relations` | 题目-材料-答案关联 |
| `essay_questions` | 组装后的题目 |
| `essay_reviews` | 批改结果 |

## 申论流程核心链路

```
PDF 上传 → pdftotext/Go库提取文本
  → ParseDocumentWithBoundaryModel
    ├─ LLM 边界识别 (BoundaryPlan) → ApplyBoundaryPlanToLines → Sections
    └─ 规则回退 (锚点+状态机) → Sections
  → AssembleQuestions → essay_questions + essay_section_relations
  → ReviewAnswer (取关联材料/答案 → LLM 批改 → 结构化评分)
```

核心文件：`essay_service.go:ParseDocumentWithBoundaryModel` → `parser/boundary_plan.go` → `parser/linker.go` → `essay_service.go:AssembleQuestions` → `essay_service.go:ReviewAnswer`

## API 模块状态

### 已接真实后端

LLM、Prompts、Mistakes（后端完成，前端部分 mock）、Pomodoro、Study Logs、Study Plans、Calendar、Music、OCR（百度 OCR 已接入）、Essay、Backup、Planning（首页三层规划，`/api/planning/*`；`/api/tasks/*` 保留为日任务兼容入口）

### 仍是 mock / 待实现

- 题库管理（`questions`）：前端全部 mock
- 主题配置（`theme`）：使用 localStorage
- Prompt 测试接口（`POST /api/prompts/test`）：前端 mock
- OCR AI 修正（`POST /api/ocr/correct`）：前端 mock

## 开发快速定位

### 新增页面

1. 在 `src/pages/` 新建页面组件
2. 在 `src/AppRoutes.jsx` 添加路由
3. 在 `src/components/Sidebar.jsx` 添加导航
4. 在 `src/components/AppLayout.jsx` 的 `pageMeta` 添加标题
5. 如需后端接口，新增 `src/api/xxx.js` 和对应后端 handler/service/model

### 新增后端接口

1. 在 `internal/models/` 添加/修改模型
2. 在 `internal/database/migrate.go` 加入 AutoMigrate
3. 在 `internal/services/` 添加业务逻辑
4. 在 `internal/handlers/` 添加 HTTP handler
5. 在 `internal/routes/routes.go` 注册路由

## 常见坑

1. **AntD 组件未导入** 导致页面空白，`npm run build` 不一定能发现
2. **multipart 上传** 不要手动设置 `Content-Type`，浏览器需要自动设 boundary
3. **PostgreSQL** 文本不能保存 `\x00`，入库前需清洗
4. **PDF 乱码** 优先用 `/api/pdf/parse-test` 看原始文本
5. **不要随意删除旧字段**——GORM AutoMigrate 只补字段，不回滚

## 验证命令

```bash
# 后端测试
cd backend && go test ./...

# 前端构建
npm run build

# 接口冒烟
curl http://localhost:21080/api/health
curl http://localhost:21080/api/db/ping
```

## 提交前检查

```bash
cd backend && gofmt -w cmd internal && go test ./...
cd .. && npm run build
```

确认：无旧端口 `8080`/`5173`、无密钥泄漏、无重复页面标题。
