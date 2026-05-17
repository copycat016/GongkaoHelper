# 项目开发总览

这份文档给新接手的模型或开发者快速理解项目使用。它描述当前项目结构、启动方式、核心接口、主要业务流和已知注意点。

## 1. 项目定位

Gkweb 是一个个人学习辅助系统，面向公考/学习训练。当前重点功能包括：

- LLM Provider / Model 配置
- Prompt 配置
- OCR 识题
- 错题库
- 番茄钟
- 学习计划 / 学习日志 / 日历
- 申论 PDF 结构化与后续 AI 批改
- 音乐播放器
- 主题配置
- JSON 数据备份

当前仍是个人自用阶段，但后端表结构普遍预留 `user_id`，方便未来扩展多用户。

## 2. 技术栈

前端：

- Vite
- React
- Ant Design
- CSS 文件分层维护

后端：

- Go
- Gin
- GORM
- PostgreSQL
- Docker Compose 启动 PostgreSQL

PDF：

- 优先使用系统命令 `pdftotext`，来自 Poppler。
- 如果系统没有 `pdftotext`，回退到 Go 库 `github.com/ledongthuc/pdf`。
- 扫描件 PDF 需要 OCR，不应期待文本解析器能处理。

OCR：

- 当前主要接百度 OCR。
- OCR AK/SK 可在前端配置页填写，后端保存到本地配置文件。

## 3. 端口约定

尽量使用 21000 以后的端口。

- 前端 Vite：`21073`
- 后端 Gin：`21080`
- PostgreSQL：`21432`
- PaddleOCR / PDF 转图片等未来服务：建议继续使用 `21090+`

## 4. 启动方式

启动数据库：

```bash
cd backend
cp .env.example .env
docker compose --env-file .env up -d
```

启动后端：

```bash
cd backend
set -a
source .env
set +a
go run ./cmd/server
```

启动前端：

```bash
npm run dev -- --host 0.0.0.0 --port 21073
```

常用验证：

```bash
cd backend
go test ./...

cd ..
npm run build
```

## 5. 目录结构

前端重点目录：

```txt
src/
├─ api/                 # 前端 API 封装
├─ components/          # 布局、侧边栏、番茄钟、音乐 Provider 等
├─ pages/               # 页面
├─ styles/              # theme/layout/components/pages/pomodoro 样式
├─ App.jsx
└─ AppRoutes.jsx
```

后端重点目录：

```txt
backend/
├─ cmd/server/main.go
├─ internal/config
├─ internal/database
├─ internal/handlers
├─ internal/models
├─ internal/routes
├─ internal/services
├─ internal/response
├─ uploads/
├─ data/
└─ docker-compose.yml
```

不要把新功能都塞进 `App.jsx` 或 `main.go`。

## 6. 前端路由

主路由文件：

- `src/AppRoutes.jsx`

主要页面：

- `/`：首页总览
- `/ocr`：OCR 识题
- `/questions`：题库管理，当前仍偏 mock
- `/mistakes`：错题库
- `/essay`：申论 PDF 结构化与批改
- `/pomodoro`：番茄钟
- `/music`：音乐播放器
- `/study`：学习管理，包含计划/日志/日历 Tab
- `/ai`：配置，包含 LLM、Prompt、OCR、PDF 测试、备份
- `/pdf`：已重定向到 `/ai?tab=pdf`

配置页支持 URL Tab：

```txt
/ai?tab=llm
/ai?tab=prompts
/ai?tab=ocr
/ai?tab=pdf
/ai?tab=backup
```

## 7. 统一响应格式

后端统一返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {},
  "request_id": ""
}
```

错误时：

```json
{
  "code": 40075,
  "message": "具体错误",
  "request_id": ""
}
```

前端 `src/api/request.js` 会自动解包 `data`，并对 `code > 0` 抛错。

注意：文件上传接口目前有些直接使用 `fetch + FormData`，不要手动设置 `Content-Type`，否则 multipart boundary 会丢失。

## 8. 后端接口总览

基础：

```txt
GET /api/health
GET /api/db/ping
```

LLM Provider / Model：

```txt
GET    /api/llm/providers
POST   /api/llm/providers
GET    /api/llm/providers/:id/models
PUT    /api/llm/providers/:id
DELETE /api/llm/providers/:id

GET    /api/llm/models
POST   /api/llm/models
PUT    /api/llm/models/:id
DELETE /api/llm/models/:id
```

`GET /api/llm/providers/:id/models` 会按 Provider 的 `base_url + api_key` 请求 OpenAI-compatible `/models`。如果服务商不支持模型列表，前端仍允许手填模型名。

Prompt：

```txt
GET    /api/prompts
POST   /api/prompts
PUT    /api/prompts/:id
DELETE /api/prompts/:id
```

错题库：

```txt
GET    /api/mistakes
POST   /api/mistakes
GET    /api/mistakes/:id
PUT    /api/mistakes/:id
DELETE /api/mistakes/:id
POST   /api/mistakes/:id/review
```

番茄钟：

```txt
POST /api/pomodoro/sessions
GET  /api/pomodoro/stats/today
```

学习日志：

```txt
GET  /api/logs
GET  /api/logs/stats
POST /api/logs
```

学习计划：

```txt
GET    /api/plans
POST   /api/plans
PUT    /api/plans/:id
DELETE /api/plans/:id
POST   /api/plans/:id/complete
```

日历：

```txt
GET /api/calendar/events
```

音乐：

```txt
GET  /api/music/playlists
POST /api/music/playlists
GET  /api/music/tracks
POST /api/music/tracks
GET  /api/music/playlists/:playlist_id/tracks
POST /api/music/playlists/:playlist_id/tracks/:track_id
```

上传音乐保存到：

```txt
backend/uploads/music/
```

后端通过：

```txt
/uploads/music/xxx
```

对前端暴露静态文件。

OCR：

```txt
GET  /api/ocr/engines
GET  /api/ocr/scenes
GET  /api/ocr/config
PUT  /api/ocr/config
GET  /api/ocr/usage/month
POST /api/ocr/recognize
```

`POST /api/ocr/recognize` 使用 `multipart/form-data`：

```txt
file: 图片文件
scene: printed / printed_accurate / handwriting ...
engine: general_basic / accurate / handwriting ...
```

当前 OCR 会对“单字一行”的百度结果做归并，避免前端显示成一字一列。

PDF 测试：

```txt
POST /api/pdf/parse-test
```

`multipart/form-data`：

```txt
file: PDF 文件
```

该接口只测试 PDF 文本层，不入库、不切 chunk、不分类。返回每页文本和质量判断。

申论：

```txt
GET    /api/essay/documents
POST   /api/essay/documents
DELETE /api/essay/documents/:id
POST   /api/essay/documents/:id/parse
POST   /api/essay/documents/:id/debug-boundary
GET    /api/essay/documents/:id/sections
GET    /api/essay/documents/:id/chunks
POST   /api/essay/documents/:id/classify
POST   /api/essay/documents/:id/assemble
GET    /api/essay/documents/:id/questions
POST   /api/essay/questions/:id/review
```

备份：

```txt
GET /api/backup/export?include_secrets=false
```

默认不导出 LLM Provider API Key。

## 9. 主要数据模型

基础模型：

```go
type BaseModel struct {
    ID        uint
    UserID    uint
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

重要业务表：

- `llm_providers`
- `llm_models`
- `prompt_templates`
- `mistakes`
- `pomodoro_sessions`
- `study_logs`
- `study_plans`
- `music_playlists`
- `music_tracks`
- `music_playlist_tracks`
- `ocr_tasks`
- `essay_documents`
- `essay_sections`
- `essay_section_relations`
- `essay_chunks`
- `essay_questions`
- `essay_question_chunks`
- `essay_reviews`

申论文档关键字段：

- `document_role`
  - `combined`：混合卷
  - `question_paper`：题目卷
  - `answer_key`：答案卷
  - `explanation`：解析卷
- `source_group`：卷套标识，用于未来把题目卷和答案卷关联起来。
- `status`
  - `uploaded`
  - `parsing`
  - `parsed`
  - `parse_failed`
  - `classified`
  - `assembled`

## 10. 申论 PDF 流程

当前流程：

1. 用户上传 PDF。
2. 后端创建 `essay_document`。
3. 上传后自动尝试解析 PDF。
4. 文本型 PDF 解析成功后生成 `essay_sections`（语义区域）和 `essay_chunks`（兼容旧流程的固定长度 chunk）。
5. **新流程（LLM 边界识别）**：调用 `parse` 时如果指定了 `boundary_model_id`，使用 LLM 做边界识别，直接划分出 `material / question / answer / analysis` 区域，存入 `essay_sections`。
6. **旧流程（规则分类回退）**：无 LLM 时先生成 `essay_chunks`，再调用 `classify` 做规则分类。
7. 根据 section 类型组装 `essay_questions`。
8. 用户提交答案。
9. `review` 调用 LLM 做结构化批改，返回分项评分和亮点/问题分析；如果 LLM 不可用则回退到启发式评分。

section 类型：

- `material`
- `question`
- `answer`
- `analysis`
- `unknown`

chunk 类型（兼容旧流程）：

- `material`
- `question`
- `reference_answer`
- `scoring_rule`
- `explanation`
- `unknown`

sections 是当前主流程使用的模型；chunks 保留向后兼容。新功能优先使用 sections 相关接口。

批改结果：

LLM 批改返回结构化 JSON，包含总分、分项得分（如论点深度、逻辑组织、表达准确等）、亮点和问题列表。如果没有可用 LLM 则回退启发式评分。

## 11. PDF 解析策略

代码位置：

- `backend/internal/services/pdf_text.go`

当前策略：

1. 优先调用系统 `pdftotext`。
2. 如果没有 `pdftotext`，回退 Go 库 `github.com/ledongthuc/pdf`。
3. 抽取后做文本质量判断。
4. 如果中文比例低、乱码多或文本太少，会判定文本层不可用。

建议安装：

```bash
sudo apt install poppler-utils
```

验证：

```bash
pdftotext -v
```

注意：

- 有些中文 PDF 即使能选中文字，字体映射也可能不标准。
- 纯 Go 解析库对中文 PDF 不稳定。
- 扫描件 PDF 必须走 OCR。

## 12. OCR 配置与注意点

OCR 配置页：

```txt
/ai?tab=ocr
```

百度 AK/SK：

- 前端填写。
- 后端保存到本地配置文件。
- 不应提交到 GitHub。

当前已支持按 OCR engine 设置本地月上限。

百度 OCR 额度：

- 百度不同能力可能共用 AK/SK，但额度按能力/服务不同。
- 项目里用本地统计做保护，不依赖百度余额查询。

## 13. 音乐播放器

音乐播放器分两层：

- 页面播放器：`src/pages/MusicPlayer.jsx`
- 全局播放状态：`src/components/MusicProvider.jsx`
- 底部播放栏：`src/components/GlobalMusicBar.jsx`

重要要求：

- 切换到番茄钟或其它页面时，音乐不能停止。
- 不要随便把 audio ref 放回页面组件里。
- 上传文件由后端保存到 `backend/uploads/music`。

## 14. 主题配置

主题相关：

- `src/components/ThemePanel.jsx`
- `src/styles/theme.css`

主题配置主要保存在 localStorage。背景图也在前端本地保存，图片太大可能触发 localStorage 限制。

## 15. 已知 mock / 未完成点

仍偏 mock 或骨架的地方：

- 题库管理后端暂时不作为重点，前端全部使用 mock 数据。
- Prompt 测试接口仍是前端 mock（后端无对应路由）。
- OCR 的 AI 修正接口仍是前端 mock（后端无对应路由）。
- 错题库前端仍使用 mock 数据展示，后端接口已完整实现。
- PDF 扫描件转图片还未实现。
- 学习计划"AI 生成计划"按钮为 no-op。

## 16. 常见坑

1. AntD 组件未导入会导致页面空白。

例如 JSX 里用了 `Tag`，但没有从 `antd` import。`npm run build` 不一定能提前发现所有运行时空白。

2. multipart 上传不要手动设置 `Content-Type`。

浏览器需要自动设置 boundary。

3. PostgreSQL text 不能保存 `\x00`。

PDF/OCR 文本入库前要清洗空字节。

4. PDF 乱码不一定是前端问题。

优先用 `/api/pdf/parse-test` 看原始抽取文本。如果这里已经乱码，说明文本层或解析器有问题，应走 OCR。

5. 后端端口不要乱用。

已有约定：前端 21073，后端 21080，数据库 21432。

6. 不要随意删除旧字段。

当前数据库没有正式 migration 版本管理，GORM AutoMigrate 只会补字段，不负责复杂回滚。

## 17. 给其他模型的工作方式

推荐让其它模型遵守：

1. 只处理一个任务。
2. 只改指定文件。
3. 先阅读相关文件。
4. 修改后必须运行对应验证：
   - 前端：`npm run build`
   - 后端：`go test ./...`
5. 最后列出：
   - 修改文件
   - 修改内容
   - 验证结果
   - 剩余风险

可委派任务见：

```txt
docs/DELEGATED_TASKS.md
```
