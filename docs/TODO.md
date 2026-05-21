# ToDo 清单：整体实用化重构

## 执行原则

- 先做减法，再做核心增强：先清掉 mock 和黑盒入口，再加强申论维护与批改逻辑。
- 每个 Phase 必须能独立验收，完成后运行检查命令。
- 不做题库和错题库；知识点沉淀优先单向归档到思源笔记，不在本项目内重复维护一套知识库。
- PDF、OCR、粘贴文本统一视为“录入来源”，最终都产出可预览、可修正、可下游流转的纯文本。
- 申论批改是核心功能，优先保证题目可维护、关联可修正、Prompt 可按题型和变体切换。

## 默认产品决策

- 思源联动：先做单向归档，不做双向同步。
- 录入器形态：单页三 Tab，包含图片 OCR、PDF 文本层、粘贴文本。
- 申论维护范围：支持编辑题面、题号、分值、字数、题型、Section 内容和题目-材料-答案关联。
- Prompt 策略：按题型分发模板，支持单题自定义模板/评分细则。
- 用户模型：默认单人使用，暂不引入认证，继续沿用现有 `X-User-ID` 临时方案。

## 当前进度

- 已完成：里程碑 1 清场与统一录入器。
- 暂缓：里程碑 2 学习管理瘦身。用户确认该部分先推掉，不纳入当前后续推进。
- 已完成：里程碑 3 申论维护能力。
- 下一步：从里程碑 4 申论 Prompt 分发与变体继续。

---

## 里程碑 1：清场与统一录入器（已完成）

目标：删除无价值 mock 模块，合并 OCR/PDF/文本入口，让“录入”成为一个可检查、可修正、可流转的入口。

### Phase 1.1 新建统一录入器页面（已完成）

涉及文件：
- `src/pages/IntakePage.jsx`：新增页面。
- `src/api/intake.js`：新增统一录入 API 封装。
- `src/api/ocr.js`：保留 OCR 调用，必要时被 `intake.js` 复用。
- `src/api/pdf.js`：保留 PDF 测试调用，必要时被 `intake.js` 复用。
- `src/AppRoutes.jsx`：新增 `/intake` 路由；`/ocr` 可重定向到 `/intake`。
- `src/components/Sidebar.jsx`：菜单“录入器”。
- `src/components/AppLayout.jsx`：补充页面标题元信息。
- `src/styles/pages.css` 或相关样式文件：补录入器布局样式。

修改内容：
- 新页面包含三个 Tab：图片 OCR、PDF 文件、粘贴文本。
- 三种来源最后统一进入同一个预览区：来源、字符数、行数、质量状态、原始文本、可编辑修正文稿。
- 预览区提供操作：复制文本、发送到申论、写入思源（按钮可先 disabled，Phase 4 接通）、清空。
- 图片 OCR 继续复用现有 `runOcr`；PDF 复用现有 `parsePdfTest`；粘贴文本前端本地生成结果即可。
- “发送到申论”通过 query/localStorage/sessionStorage 传递文本到 `/essay`，例如写入 `sessionStorage.intake_text` 后跳转 `/essay?from=intake`。

检查节点：
- 上传图片后可以看到 OCR 原文。
- 上传 PDF 后可以看到每页文本和质量标签。
- 粘贴文本后可以直接进入统一预览区。
- 点击“发送到申论”后，`EssayReview` 上传区的原始文本能自动填充。

验收命令：

```bash
npm run build
```

### Phase 1.2 后端统一录入结果结构（已完成）

涉及文件：
- `backend/internal/services/document_parse_tool.go`
- `backend/internal/services/pdf_text.go`
- `backend/internal/services/ocr_service.go`
- `backend/internal/handlers/pdf.go`
- `backend/internal/handlers/ocr.go`
- `backend/internal/routes/routes.go`

修改内容：
- 复用现有 `ParseDocumentSource` 作为统一归一化入口。
- 给 PDF 测试返回补充 `source_engine` 字段：`poppler` / `go_pdf` / `unknown`，让前端不再黑盒。
- OCR 返回结构补充统一字段：`source`、`text`、`quality`、`line_count`。
- 如新增 `/api/intake/*` 接口，handler 只做薄封装，不复制 PDF/OCR 逻辑。

检查节点：
- PDF 文本层可用时返回 `quality.ok=true`。
- 扫描件或乱码 PDF 返回清晰 `quality.reason`。
- OCR 结果和 PDF 结果前端能用同一套预览组件渲染。

验收命令：

```bash
cd backend && go test ./...
npm run build
```

### Phase 1.3 删除题库与错题库入口（已完成）

涉及文件：
- `src/AppRoutes.jsx`
- `src/components/Sidebar.jsx`
- `src/pages/QuestionBank.jsx`
- `src/pages/QuestionDetail.jsx`
- `src/pages/MistakeBook.jsx`
- `src/api/questions.js`
- `src/api/mistakes.js`
- `src/api/mockData.js`
- `backend/internal/routes/routes.go`
- `backend/internal/handlers/question_bank.go`
- `backend/internal/services/question_bank_service.go`
- `backend/internal/handlers/mistakes.go`
- `backend/internal/services/mistake_service.go`
- `backend/internal/models/mistake.go`
- `backend/internal/database/migrate.go`
- `backend/internal/services/backup_service.go`

修改内容：
- 先从前端菜单和路由移除 `/questions`、`/mistakes`。
- 旧路径重定向到 `/intake` 或 `/`，不要出现空白页。
- 前端删除 mock 题库/错题页面和 API 引用。
- 后端暂不强删表；如果相关 handler/service 无调用，可先保留代码但从路由解绑，避免一次性破坏 backup/migrate。
- 如确定无依赖，再在后续清理中移除 model/migrate/backup 中的 mistake/question 相关项。

检查节点：
- 侧栏不再出现“题库管理”“错题库”。
- 访问 `/questions`、`/mistakes` 不报错，能跳到替代页。
- `npm run build` 无 dead import。

验收命令：

```bash
npm run build
cd backend && go test ./...
```

---

## 里程碑 2：学习管理瘦身（暂缓，不用管）

目标：删除花架子计划/日历，只保留能真实记录学习行为的日志能力。

状态：暂缓。该里程碑先推掉，不作为当前 TODO 的继续执行项。

### Phase 2.1 学习管理只保留日志（暂缓）

涉及文件：
- `src/pages/StudyCenter.jsx`
- `src/pages/StudyLogs.jsx`
- `src/pages/StudyPlans.jsx`
- `src/pages/CalendarPage.jsx`
- `src/api/plans.js`
- `src/api/calendar.js`
- `src/AppRoutes.jsx`
- `src/components/Sidebar.jsx`
- `src/components/AppLayout.jsx`
- `backend/internal/handlers/study.go`
- `backend/internal/services/study_service.go`
- `backend/internal/routes/routes.go`

修改内容：
- `/study` 直接展示 `StudyLogs`，删除 plans/calendar Tab。
- `/plans`、`/calendar` 重定向到 `/study`。
- 前端删除 `StudyPlans.jsx`、`CalendarPage.jsx`、`api/plans.js`、`api/calendar.js` 的引用。
- 后端可保留计划/日历 service 代码，但路由先解绑或标记废弃；避免 UI 再触达。
- Dashboard 改为三层规划首页：`stage_goals` 阶段目标 → `weekly_tasks` 周任务 → `daily_tasks` 日任务。首页提供「今日 / 本周 / 阶段规划」Tab；日任务同时支持 `plan_date`（计划哪天做）与 `due_date`（DDL），未设置 `plan_date` 的 DDL 任务进入「待安排」池；专注数据仍来自 `pomodoro_sessions` 与 `study_logs`，不再依赖 mock。

检查节点：
- `/study` 不再有 Tab，只显示日志。
- `/plans`、`/calendar` 不报错。
- Dashboard 不再请求 `getPlans`。

验收命令：

```bash
npm run build
cd backend && go test ./...
```

### Phase 2.2 学习日志补登（暂缓）

涉及文件：
- `src/pages/StudyLogs.jsx`
- `src/api/logs.js`
- `backend/internal/handlers/study.go`
- `backend/internal/services/study_service.go`
- `backend/internal/models/study.go`

修改内容：
- 在 `StudyLogs` 顶部增加“补登日志”按钮和 Modal。
- 字段：开始时间、结束时间或持续分钟、学习类型、科目、备注。
- 调用现有 `saveLog` 接口。
- 保存后刷新列表和统计。
- 移除永远为 0 的“中断次数”统计卡。

检查节点：
- 手动补登 30 分钟后，日志表格出现记录。
- 总时长统计立即变化。
- 番茄钟生成的日志仍然正常显示。

验收命令：

```bash
npm run build
cd backend && go test ./...
```

---

## 里程碑 3：申论维护能力（已完成）

目标：解决“解析结果需要人工修正”的核心问题，让题目、分段和关联都可编辑。

### Phase 3.1 后端新增题目与分段编辑接口（已完成）

涉及文件：
- `backend/internal/models/essay.go`
- `backend/internal/services/essay_service.go`
- `backend/internal/handlers/essay.go`
- `backend/internal/routes/routes.go`
- `backend/internal/database/migrate.go`
- `backend/docs/backend-dev.md`

修改内容：
- `EssayQuestion` 增加字段：`ManuallyEdited bool`、`CustomPromptID *uint`、`ScoringRubric string`。
- 新增 service 方法：
  - `UpdateQuestion(userID, questionID, payload)`：更新题号、标题、题型、题面、满分、字数、评分细则、prompt。
  - `CreateQuestion(userID, payload)`：手动新建申论题。
  - `DeleteQuestion(userID, questionID)`：删除题目及关联/批改记录。
  - `UpdateSection(userID, sectionID, payload)`：更新 section 类型、标题、内容、题号、关联题号。
  - `ReplaceQuestionRelations(userID, questionID, materialIDs, answerIDs)`：重置材料/答案关联。
- 新增 handler 和 route：
  - `POST /api/essay/questions`
  - `PUT /api/essay/questions/:id`
  - `DELETE /api/essay/questions/:id`
  - `PUT /api/essay/sections/:id`
  - `POST /api/essay/questions/:id/relations`
- 重新解析前要提示会覆盖；如实现保护，则保留 `ManuallyEdited=true` 的题目不被自动删除。

检查节点：
- PUT 修改题面后，刷新仍然保存。
- 修改题型后，下次批改读取新题型。
- 替换材料/答案关联后，`ReviewAnswer` 的 context 只包含新关联内容。

验收命令：

```bash
cd backend && gofmt -w internal cmd && go test ./...
```

### Phase 3.2 前端申论编辑抽屉（已完成）

涉及文件：
- `src/pages/EssayReview.jsx`
- `src/api/essay.js`
- `src/styles/pages.css`

修改内容：
- `api/essay.js` 增加上述新增接口封装。
- 题目列表每项增加“编辑”按钮。
- 新增题目编辑 Drawer：题号、标题、题型、题面、满分、字数、评分细则、Prompt 模板。
- 题目详情中材料/答案区域改为可勾选关联；保存后调用 relations 接口。
- Section 列表每项增加“编辑分段”按钮，可修改类型、标题、内容、题号。
- “重新解析”增加确认弹窗：提示可能覆盖手动修正。

检查节点：
- 前端可编辑题面并保存。
- 前端可调整题目关联材料/答案。
- 编辑后无需重新上传 PDF 即可重新批改。

验收命令：

```bash
npm run build
```

---

## 里程碑 4：申论 Prompt 分发与变体

目标：让概括、对策、分析、应用文、大作文等不同题型使用不同评分维度，并允许单题定制。

### Phase 4.1 Prompt 模板模型升级

涉及文件：
- `backend/internal/models/prompt.go`
- `backend/internal/services/prompt_service.go`
- `backend/internal/handlers/prompts.go`
- `backend/internal/database/migrate.go`
- `backend/internal/services/essay_service.go`
- `src/pages/PromptSettings.jsx`
- `src/api/prompts.js`

修改内容：
- `PromptTemplate` 增加或复用字段：`scenario`、`question_type_match`、`version`、`enabled`。
- 增加默认场景：
  - `essay_review_default`
  - `essay_review_summary`
  - `essay_review_strategy`
  - `essay_review_analysis`
  - `essay_review_application`
  - `essay_review_big_essay`
- `PromptService` 新增 `ResolveEssayReviewPrompt(userID, questionType, customPromptID)`。
- 初始化默认模板时写入上述场景。
- 前端 Prompt 设置页支持按 scenario 筛选/编辑。

检查节点：
- 数据库迁移后默认模板存在。
- 禁用某模板后不会被自动选中。
- 前端能编辑并保存模板内容。

验收命令：

```bash
cd backend && gofmt -w internal cmd && go test ./...
npm run build
```

### Phase 4.2 批改调用改为模板渲染

涉及文件：
- `backend/internal/services/essay_service.go`
- `backend/internal/models/essay.go`
- `backend/internal/database/migrate.go`
- `src/pages/EssayReview.jsx`

修改内容：
- `callLLMReview` 不再使用硬编码 `buildReviewSystemPrompt` / `buildReviewUserPrompt`。
- 使用解析后的模板渲染变量：题目信息、材料、参考答案、考生答案、评分细则。
- `EssayReview` 记录 `prompt_id`、`prompt_version`、`raw_request`、`raw_response`，方便排错。
- LLM 输出 JSON 增加 `score_breakdown`，后端校验维度分之和和总分一致。
- 如果分项和总分不一致，优先以后端分项求和修正总分，或标记 `validation_warning`。

检查节点：
- 概括题和对策题批改结果的维度名称不同。
- 单题设置 `CustomPromptID` 后，批改使用该模板。
- 批改结果保存 raw response，方便回放。

验收命令：

```bash
cd backend && gofmt -w internal cmd && go test ./...
npm run build
```

### Phase 4.3 题型识别与人工覆盖

涉及文件：
- `backend/internal/services/essay_service.go`
- `src/pages/EssayReview.jsx`

修改内容：
- 优化 `guessQuestionType`，至少覆盖：概括题、对策题、分析题、应用文、大作文、综合题。
- 题型识别只做默认值，前端编辑题型后必须以后端保存值为准。
- 题型改变后，批改 Prompt 自动跟随改变。

检查节点：
- “请概括……”识别为概括题。
- “请提出对策/建议……”识别为对策题。
- 前端手动改题型后刷新仍然生效。

验收命令：

```bash
cd backend && go test ./...
npm run build
```

---

## 里程碑 5：思源单向归档

目标：不在本项目重复做知识库，而是把申论批改和录入文本归档到思源笔记。

### Phase 5.1 验证思源 API

涉及文件：
- 无代码文件必须修改；先查文档和本地接口。
- 后续可能新增 `backend/internal/services/siyuan_service.go`。

修改内容：
- 先通过官方文档确认当前思源接口：连接测试、列出笔记本、用 Markdown 创建文档。
- 预期接口包括：`/api/notebook/lsNotebooks`、`/api/filetree/createDocWithMd`，但必须以实际文档和本地测试为准。

检查节点：
- 能用 token 调通思源本地服务。
- 能列出笔记本。
- 能创建一篇 Markdown 文档。

验收命令：

```bash
# 示例，具体按实际思源配置替换
curl -H "Authorization: Token <token>" http://127.0.0.1:6806/api/notebook/lsNotebooks
```

### Phase 5.2 后端思源配置与归档接口

涉及文件：
- `backend/internal/models/`：如需要持久化配置，新增 `siyuan.go` 或复用配置表。
- `backend/internal/services/siyuan_service.go`：新增。
- `backend/internal/handlers/siyuan.go`：新增。
- `backend/internal/routes/routes.go`
- `backend/internal/database/migrate.go`
- `backend/internal/services/essay_service.go`

修改内容：
- 新增配置：endpoint、token、默认 notebook_id、默认归档路径模板。
- 新增接口：
  - `POST /api/siyuan/test`
  - `GET /api/siyuan/notebooks`
  - `POST /api/siyuan/archive`
- `archive` 支持 `type=essay_review`：根据 `EssayReview.id` 生成 Markdown，包含题面、我的答案、得分、维度点评、亮点/问题、改进建议。
- `archive` 支持 `type=intake_text`：将录入器文本归档。
- token 不写入 Git；如使用环境变量或本地 DB，文档中说明。

检查节点：
- 配置错误时返回清晰错误。
- 思源服务未启动时不影响申论批改主流程。
- 归档成功返回 doc id 或可访问路径。

验收命令：

```bash
cd backend && gofmt -w internal cmd && go test ./...
```

### Phase 5.3 前端归档入口

涉及文件：
- `src/pages/AISettings.jsx`
- `src/pages/EssayReview.jsx`
- `src/pages/IntakePage.jsx`
- `src/api/siyuan.js`

修改内容：
- 配置页增加“思源”Tab：endpoint、token、测试连接、选择默认笔记本。
- 申论批改结果卡片增加“归档到思源”按钮。
- 录入器预览区“写入思源”按钮从 disabled 变为可用。
- 归档弹窗支持选择笔记本、文档路径、标题。

检查节点：
- 申论批改结果可以一键归档到思源。
- 录入文本可以归档到思源。
- 思源异常时前端显示错误，不清空当前内容。

验收命令：

```bash
npm run build
```

---

## 里程碑 6：收尾与文档

目标：让代码和文档反映新产品结构，避免后续 AI 继续按旧模块开发。

### Phase 6.1 更新维护文档

涉及文件：
- `README.md`
- `docs/QUICKSTART.md`
- `docs/MAINTENANCE.md`
- `backend/docs/backend-dev.md`

修改内容：
- 更新主要页面：去掉题库/错题库，加入录入器和思源归档。
- 更新 API 模块状态：标记题库/错题库废弃，说明申论维护和 Prompt 分发。
- 更新目录结构，避免文档继续指向删除页面。
- 更新验证命令。

检查节点：
- 文档中不再把 `/questions`、`/mistakes` 描述为主要功能。
- 文档说明 PDF/OCR 统一入口。
- 文档说明思源 token 不提交。

验收命令：

```bash
npm run build
cd backend && go test ./...
```

### Phase 6.2 全量检查

涉及文件：
- 全项目。

检查内容：
- 搜索旧入口和 mock 引用。
- 搜索死路由、死 import。
- 确保没有硬编码密钥。
- 确保没有旧端口 `8080` / `5173`。

建议命令：

```bash
npm run build
cd backend && gofmt -w internal cmd && go test ./...
```

最终验收：
- 侧栏只保留：首页、录入器、申论批改、番茄钟、学习日志、配置。
- OCR/PDF/粘贴文本都能进入同一预览和申论流转。
- 申论题目可编辑，材料/答案关联可改。
- 批改 Prompt 按题型选择，支持单题变体。
- 批改结果可归档到思源。
- `npm run build` 和 `go test ./...` 均通过。
