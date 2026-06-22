# 维护手册

## 目录结构

```
GongkaoHelper/
├─ README.md                 # 项目概述、快速启动、Docker 部署
├─ docs/
│   ├─ README.md             # 文档索引
│   ├─ AGENT_CONTEXT.md      # AI 接手上下文，避免重复加载长上下文
│   ├─ QUICKSTART.md         # 快速启动
│   ├─ MAINTENANCE.md        # 维护手册（本文件）
│   └─ design-system.md      # 前端设计系统规范（token/母题组件/清洗准则）
├─ index.html                # Vite 入口 HTML
├─ package.json              # 前端依赖
├─ vite.config.js            # Vite 配置（端口 21073，代理 /api -> 21080）
│
├─ src/                      # 前端源码
│   ├─ main.jsx              # React 入口
│   ├─ App.jsx / App.css     # 根组件 / CSS 导入顺序
│   ├─ AppRoutes.jsx         # 路由定义（lazy loading）
│   ├─ index.css             # 全局样式
│   ├─ api/                  # 后端 API 封装（禁止页面直接写 fetch）
│   │   ├─ request.js        # 统一 fetch 封装、错误处理、response 解包、Bearer Token
│   │   ├─ auth.js           # 登录 / 当前用户
│   │   ├─ llm.js            # LLM Provider / Model
│   │   ├─ prompts.js        # Prompt 模板
│   │   ├─ pomodoro.js       # 番茄钟
│   │   ├─ logs.js           # 学习日志
│   │   ├─ tasks.js          # 首页今日任务兼容 API（/api/tasks）
│   │   ├─ planning.js       # 三层规划 API（阶段目标/子项/周任务/日任务）
│   │   ├─ intake.js         # 统一录入器
│   │   ├─ music.js          # 音乐
│   │   ├─ ocr.js            # OCR
│   │   ├─ pdf.js            # PDF 解析工具/测试
│   │   ├─ essay.js          # 申论
│   │   ├─ backup.js         # 备份导出
│   │   ├─ theme.js          # 主题（后端 /api/theme + localStorage fallback）
│   │   └─ mockData.js       # 残留 mock 数据（待清理）
│   ├─ components/           # 可复用组件（非页面）
│   │   ├─ AppLayout.jsx     # 页面布局框架（含 pageMeta 标题表）
│   │   ├─ Sidebar.jsx / sidebarItems.jsx  # 主导航菜单 + 菜单项配置
│   │   ├─ ProtectedRoute.jsx # 登录态路由守卫
│   │   ├─ PreviewBanner.jsx # 预览环境提示条
│   │   ├─ Pomodoro.jsx      # 番茄钟计时器
│   │   ├─ MusicProvider.jsx / musicContext.js  # 全局音乐状态
│   │   ├─ GlobalMusicBar.jsx
│   │   ├─ ThemePanel.jsx    # 主题配置面板
│   │   └─ ui/               # 母题组件库（见 design-system.md）
│   │       ├─ Page / PageHeader / AppCard / SectionHeader
│   │       ├─ StatCard / Toolbar / FormGrid / EmptyState
│   │       ├─ index.js      # 统一导出
│   │       └─ ui.css        # 母题组件样式（CSS 导入最后，权重最高）
│   ├─ pages/                # 页面组件（与路由一一对应）
│   │   ├─ Login.jsx         # /login 登录页
│   │   ├─ Dashboard.jsx     # / 首页容器：今日 / 本周 / 阶段规划
│   │   ├─ dashboard/        # 首页拆分：展示区块、列表/弹窗组件、工具函数
│   │   ├─ IntakePage.jsx    # /intake 统一录入器（图片 OCR / PDF / 粘贴文本）
│   │   ├─ EssayReview.jsx   # /essay
│   │   ├─ PomodoroPage.jsx  # /pomodoro
│   │   ├─ MusicPlayer.jsx   # /music
│   │   ├─ StudyCenter.jsx / StudyLogs.jsx  # /study 学习日志
│   │   ├─ AISettings.jsx    # /ai (LLM/Prompt/OCR/PDF/备份 Tab)
│   │   ├─ LLMSettings.jsx
│   │   └─ PromptSettings.jsx
│   ├─ theme/                # 主题系统
│   │   ├─ ThemeProvider.jsx / themeContext.js  # 主题上下文
│   │   ├─ palettes.js       # 4 套配色（aozora/sakura/matcha/sumi）
│   │   ├─ antdTheme.js      # antd ConfigProvider 主题（buildAntdTheme）
│   │   ├─ applyThemeConfig.js  # 注入 :root 变量
│   │   └─ tokens.css        # 基础层 token（间距/圆角/阴影/字号/字重）
│   └─ styles/               # CSS 样式（语义层 + 页面样式）
│       ├─ theme.css         # 语义层 token
│       ├─ layout.css / components.css / pages.css
│       ├─ dashboard.css / pomodoro.css
│
└─ backend/
    ├─ cmd/server/main.go    # 唯一入口
    ├─ cmd/server/web/       # 构建时嵌入的前端 dist（仅保留 .gitkeep）
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
        ├─ auth/             # JWT token 签发/校验
        ├─ middleware/       # CORS、RequestID、Recovery、AuthRequired
        ├─ response/         # 统一响应格式
        ├─ routes/           # 路由注册
        ├─ models/           # GORM 数据模型
        ├─ handlers/         # HTTP Handler（参数解析 → service → 响应）
        ├─ services/         # 业务逻辑与 DB 查询
        ├─ parser/           # 申论解析层（纯函数，不依赖 HTTP）
        └─ version/          # 版本信息
```

## 前后端约定

### 前端

- **页面** 放 `src/pages/`，**可复用组件** 放 `src/components/`
- **所有后端请求** 只从 `src/api/*.js` 发出，禁止页面内直接 `fetch`
- **路由** 统一在 `AppRoutes.jsx` 维护
- **导航菜单** 在 `Sidebar.jsx` 维护
- **页面标题** 在 `AppLayout.jsx` 的 `pageMeta` 维护
- **新页面用母题组件** 拼装（`src/components/ui` 的 Page/PageHeader/AppCard/...），样式只引用 token，详见 [`design-system.md`](design-system.md)
- `App.jsx` 只保留应用入口

### 后端

- `models/` 只放 GORM 模型，所有业务表保留 `user_id`
- `services/` 放业务逻辑和 DB 查询
- `handlers/` 只做参数解析 → 调用 service → 返回响应
- `routes/` 只注册路由，不写业务逻辑
- `parser/` 是纯解析层，不依赖 HTTP/Gin
- 统一响应使用 `internal/response`
- 业务路由通过 JWT 鉴权中间件注入 `user_id`；开发空库默认单人管理员 ID 为 `1`

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
| `users` | 用户（单人模式默认 owner，ID=1） |
| `theme_configs` | 用户主题配置 |
| `llm_providers` | LLM 服务商配置 |
| `llm_models` | 模型列表 |
| `prompt_templates` | Prompt 模板 |
| `pomodoro_sessions` | 番茄钟 |
| `study_logs` | 学习日志 |
| `study_plans` | 学习计划 |
| `stage_goals` | 阶段目标（考前/阶段规划、起止日期、目标指标） |
| `stage_items` | 阶段子项（阶段目标下的中间层） |
| `weekly_tasks` | 推进任务（标准周 / 专项项目；归属阶段目标/子项、DDL、状态、进度聚合） |
| `daily_tasks` | 日任务（归属周任务/阶段子项/目标、计划执行日、DDL、来源 origin、完成状态） |
| `music_playlists` / `music_playlist_tracks` | 歌单 / 歌单-音轨关联 |
| `music_tracks` | 音轨 |
| `ocr_tasks` | OCR 调用记录 |
| `essay_documents` | 申论文档 |
| `essay_chunks` / `essay_question_chunks` | 文本块 / 题目-文本块关联 |
| `essay_sections` | 语义区域（material/question/answer/analysis） |
| `essay_section_relations` | 题目-材料-答案关联 |
| `essay_questions` | 组装后的题目 |
| `essay_reviews` | 批改结果（保存题面/材料/答案/Prompt/模型快照） |
| `mistakes` | 错题记录（**遗留表**，UI 已下线，无注册路由） |

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

接口注册见 [`routes/routes.go`](../backend/internal/routes/routes.go)。除 `/api/health`、`/api/db/ping`、`/api/auth/login` 公开外，其余全部走 `AuthRequired` JWT 中间件。

### 已接真实后端

Auth（`/api/auth/login`、`/me`）、LLM、Prompts、Pomodoro、Study Logs（`/api/logs`）、Music、OCR（百度 OCR 已接入）、Essay、Backup（仅导出）、Theme、Planning（三层规划 `/api/planning/*`；`/api/tasks/*` 保留为日任务兼容入口）

### 遗留 / 未实现

- `models.Mistake`（`mistakes` 表）：错题库 UI 与 handler/service 已删除，但模型仍被 `backup_service.go`（导出）和 `migrate.go` 引用而保留；彻底移除需同步改备份格式与迁移。
- `StudyService.CalendarEvents` / `handlers/study.go` 的 `CalendarEvents`：无注册路由，属未接线代码。
- `src/api/mockData.js`：残留 mock 数据，待清理。
- 备份导入恢复：仅有导出，导入未实现。

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
