# 申论批改模块：泛化重构设计与执行清单

更新日期：2026-05-12

## 0. 目标

让申论批改模块在面对各类公考真题（省考、国考、事业单位、教师招聘等）时具备稳定的"题目接得上"能力：

1. **题目—材料—参考答案**三者必须有可信、可追溯的关联，而不是"全部材料关到全部题目"的兜底结果。
2. 支持**跨文档拼装**：题目卷 + 答案卷 + 解析卷（由 `source_group` 标识同一卷套）能合并组装成完整可批改的题目。
3. 解析对各种排版（题号格式、材料编号、参考答案分段方式）保持稳健。
4. 错误链路可观察，可人工修正。
5. 现有 schema、API、前端组件做最小破坏式演进。

---

## 1. 现状审计

### 1.1 数据流

```
PDF 上传
  └─ pdftotext / Go 库 提取
      └─ ParseDocumentWithBoundaryModel
          ├─ boundaryModelID > 0:
          │    PrepareNumberedLines → BuildBoundaryPromptFromLines
          │    → LLM 返回 BoundaryPlan(行号区间) → ApplyBoundaryPlanToLines → []Section
          └─ boundaryModelID == 0:
               PrepareString → 锚点 + 状态机 → []Section
          └─ 写入 essay_sections，自动 AssembleQuestions
              └─ 生成 essay_questions + essay_section_relations
                  └─ ReviewAnswer 根据 question_id 取 relations 中的 material/answer 喂给 LLM
```

### 1.2 已经做对的事

- 引入了"语义区域 Section（material/question/answer/analysis）"作为主模型，并保留 chunk 视图兼容旧前端。
- BoundaryPlan 让 LLM 输出题号、关联题号、引用材料编号。
- AssembleQuestions 写出了三层关联策略（精确 → 顺序 → 兜底）。
- 调试接口 `debug-boundary` 暴露了 LLM prompt / raw / 解析后 JSON / 应用后 sections，便于排错。
- ReviewAnswer 通过 `essay_section_relations` 读取上下文，避免把整篇 PDF 全塞给批改模型。

### 1.3 当前关键问题

| # | 问题 | 影响 |
|---|------|------|
| P1 | LLM 输出的题号未做标准化（中文/阿拉伯/全半角混用），`question_no` 拼写不一致就关联失败 | 题目和参考答案接不上 |
| P2 | `extractMaterialNo` 只取标题里第一个中文/数字字符，"给定资料三、四" 无法识别为 [3,4] | 材料编号关联易错 |
| P3 | 题目卷与答案卷分别上传时（不同 `essay_document.id`）没有跨文档 assemble 通道 | 只能上传"混合卷"才能凑齐批改上下文 |
| P4 | 前端 `EssayReview.jsx` 中 `relatedMaterials` 与 `relatedAnswers` 是从 sections 自己过滤的，**没有读取 `essay_section_relations` 表**——也就是说后端写的精确关联前端没用上 | 前端展示与批改使用的上下文不一致 |
| P5 | 当 LLM 不给 `related_question_nos` 时，AssembleQuestions 兜底是"所有答案关到所有题目"。对于阅卷 prompt 来说是噪声 | LLM 批改命中要点降低 |
| P6 | 题目元数据（分值、字数、题型）依赖正则 `\d+\s*分` `\d+\s*字`，会被材料编号或正文里的数字误中 | 题目元数据脏 |
| P7 | LLM 切分 prompt 没有 few-shot，没有"严格 JSON Schema 校验+重试" | 不同省份、不同排版稳定性差 |
| P8 | 切分模型选择是"可选"的，但旧规则路径（无 LLM）几乎无法处理真实试卷 | 用户不选模型就解析翻车 |
| P9 | `/classify` `/chunks` 端点与 `EssayChunk` 表仍存在，但已是 sections 的视图，造成代码与心智负担 | 维护成本 |
| P10 | 没有 sanity check：题号是否连续、material 是否占文本主体、answer 数与 question 数是否匹配；也没有人工修正入口 | 错了只能重新跑一遍 |
| P11 | 批改 / 切分 prompt 全部硬编码在 Go 中，无法在前端 Prompt 管理里调试 | 调优需要改代码 |
| P12 | 没有针对真实样卷的回归测试 fixtures | 改一个地方容易回归别处 |

---

## 2. 设计思路

### 2.1 分层

```
┌─────────────────────────────────────────────────────────────┐
│ 表现层 (frontend)                                            │
│  - EssayReview.jsx 改为读 question/:id/context              │
│  - 新增 SourceGroup 视图：跨文档拼装                          │
│  - 新增 关联编辑面板（拖拽 material/answer 到题目）           │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│ 服务层 (handlers + services/essay_service.go)                │
│  - QuestionContextService.Build(questionID) → 单题完整上下文 │
│  - SourceGroupService.Assemble(sourceGroup)  → 跨文档拼装    │
│  - RelationService.Update(...)               → 手动改关联    │
│  - QuestionService.Update(...)               → 改题号/分值   │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│ 解析层 (parser/)                                             │
│  - BoundaryPlanner (LLM 切分，主路径)                        │
│      ├─ Prompt 模板 + few-shot                              │
│      ├─ Schema 严格校验 + 一次重试                            │
│      └─ Normalize：question_no/material_no 全部转 int        │
│  - SanityChecker：检查覆盖率、连续性、数量匹配                 │
│  - Linker：question ↔ material ↔ answer 关联，三级优先级     │
│  - RuleEngine（回退保留，但仅用于调试）                       │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 核心抽象

新增 `QuestionContext` 作为对外协议（前端与 LLM 批改共用）：

```go
type QuestionContext struct {
    Question  EssayQuestion   `json:"question"`
    Materials []EssaySection  `json:"materials"`
    Answers   []EssaySection  `json:"answers"`
    Analyses  []EssaySection  `json:"analyses"`   // 解析（可选）
    Sources   []SourceRef     `json:"sources"`    // 来自哪几个 document
    Warnings  []string        `json:"warnings"`   // sanity check 输出
}
```

### 2.3 三级关联策略（明确化）

| 级别 | 策略 | 落地 |
|------|------|------|
| L1 精确 | LLM 输出 `question_no` / `related_question_nos` / `material_nos` + 我方归一化（中文转阿拉伯，全角转半角） | Linker.exactMatch |
| L2 启发式 | 题目内文里出现"根据材料 X"、答案内文里出现"第 X 题"——后处理正则补出关联 | Linker.heuristicMatch |
| L3 顺序 | 当且仅当 question 与 answer 数量相等、且都没匹配上时，按顺序对应 | Linker.sequenceMatch |
| L4 兜底 | 仅在材料部分启用"全部材料关到全部题目"（申论传统就是共享材料），**答案不再兜底全连** | Linker.fallback |

### 2.4 跨卷套（source_group）合并

* 一个 `source_group` 内可以有多个 `essay_document`，每个有 `document_role`。
* 新增 SourceGroupService.Assemble：
  1. 取该 group 下所有 documents 的 sections。
  2. 按角色 dedupe：题目卷只取 question/material，答案卷只取 answer/analysis。
  3. 进入 Linker。
  4. 关联结果落在 `essay_section_relations`，但 `essay_question` 仍归属创建它的 question_paper 文档。
* 前端在文档列表里支持"按卷套分组"展示。

### 2.5 综应 C 类（事业单位自然科学专技类）兼容

综应 C 类卷面与申论本质不同，需要扩展现有模型：

| 差异点 | 申论 | 综应 C 类 |
|--------|------|-----------|
| 材料 | 多则给定资料，跨题共享 | 仅一篇科技文献，绑定大题 |
| 题目 | 1–5 道独立题 | 3 大题（科技文献阅读 / 数理 / 策略选择），每大题下 1–5 子问 |
| 题型 | 全主观题 | 主观 + 判断 + 选择 + 数值 混合 |
| 答案 | 自由文字 | 含正确选项 / 判断结果 / 数值 / 评分要点 |
| 关联 | 材料对所有题目共享 | 材料仅对**所在大题**及其子问可见 |

**兼容策略**：

* 不再强绑"申论"——`EssayDocument` 新增 `paper_type` 字段（默认 `shenlun`，可选 `zongying_c`）。
* `EssayQuestion` 加 `parent_question_id` + `sub_index`，支持二级父子结构（综应 C 类用，申论默认 0）。
* `EssayQuestion` 加 `objective_type` (`subjective` / `judgment` / `choice` / `numeric`) + `reference_answer_json`（结构化参考答案）。
* Linker 按 `paper_type` 分支：综应 C 类材料只关联到所在大题及其子问，不做全量兜底。
* 批改路径按 `objective_type` 分流：客观题先做字符串/数值比对，再让 LLM 补理由；主观题继续走现有全文 LLM 批改。
* Prompt 模板按 `paper_type` 分别预置（见 T16）。

### 2.6 模型分配速查

| 模型 | 适合任务（推荐顺序内） | 共计 |
|------|----------------------|------|
| **Kimi K2.6** | T01 / T07 / T12 / T16 | 4 条 |
| **Sonnet 4.6** | T03 / T06 / T08 / T09 / T14 / T15 | 6 条 |
| **DeepSeek V4 Pro** | T02 / T04 / T05 / T10 / T11 / T13 | 6 条 |

**Kimi 本轮任务包（按依赖顺序）**：

1. **T01** 题号与材料编号归一化（基础工具，先做）
2. **T07** BoundaryPrompt 加 few-shot（依赖 T02 已上）
3. **T16** 综应 C 类专属 prompt 模板（依赖 T13、T11；可与 T07 并行起草）
4. **T12** 真实样卷回归测试 + 文档收尾（最后做，依赖前面所有）

派给 Kimi 时**必须**提醒：
- 代码标识符（变量、函数、文件名）一律英文；只有注释和中文 prompt 内容用中文。
- 不要主动改任务允许范围外的文件。
- 完成后必须给出验收命令实际输出（`go test ./...` / `npm run build`）。
- 提交信息（commit message）用英文 + 一句中文摘要。

---

## 3. 任务清单（可委派给编程子模型执行）

每条任务都给出：目标、允许修改文件、禁止修改、验收命令、风险点。
**主模型在合并前要审计：是否破坏现有解析路径、是否引入循环依赖、是否漏写迁移、是否暴露 Key/路径。**

> 推荐执行顺序：T01 → T02 → T03 → T05 → T06 → T07 → T04 → T08 → T09 → T10 → T11 → T12 → T13 → T14 → T15 → T16
>
> 综应 C 类增量（T13–T16）依赖 T01–T12 完成，可作为下一阶段独立 PR 批次。

---

### T01 题号与材料编号归一化（基础工具）

**推荐执行模型**：**Kimi K2.6**（中文数字/全角半角/题号格式归一是 Kimi 的强项；测试用例覆盖中文边界要求高）

**目标**：把"一""1""１""壹"全部归一为 "1"；把"给定资料一、二"拆为 ["1","2"]。

**允许修改**：
- 新增 `backend/internal/parser/normalize.go`
- 新增 `backend/internal/parser/normalize_test.go`

**禁止修改**：service / handler / models（这是纯函数）。

**验收**：
```bash
cd backend && go test ./internal/parser/...
```

**必须覆盖测试**：
- `NormalizeQuestionNo("一")` == `"1"`
- `NormalizeQuestionNo("（二）")` == `"2"`
- `NormalizeQuestionNo("１")` == `"1"`
- `ExtractMaterialNos("给定资料三、四")` == `["3","4"]`
- `ExtractMaterialNos("材料 1、材料 2")` == `["1","2"]`

**风险**：把多义文本（如 "第一" 出现在正文中）错误归一。函数只负责"给定字符串本身是不是题号/材料号"，不做内容判断。

---

### T02 BoundaryPlan 输出归一化与校验

**推荐执行模型**：DeepSeek V4 Pro（机械 schema 校验 + 数据归一化，Go 类型工作 DSV4 稳定）

**目标**：LLM 返回的 plan 立刻经过归一与 schema 校验；不合法字段记入 warnings 而非直接丢弃。

**允许修改**：
- `backend/internal/parser/boundary_plan.go`（新增 NormalizeBoundaryPlan、ValidateBoundaryPlan）
- `backend/internal/services/essay_service.go`（在 suggestBoundaryPlanFromLines 与 DebugBoundarySplit 调用归一化）

**禁止修改**：
- 数据库模型
- prompt 文本（T07 再改）

**验收**：
```bash
cd backend && go test ./...
```

**重点**：
- 不合法 `start_line/end_line` 不再静默 skip，而是写进 `BoundaryDebugResult.Warnings`。
- 题号经 T01 的 NormalizeQuestionNo 转换后写回 plan.Sections[i].QuestionNo。
- `related_question_nos` / `material_nos` 同样归一。

**风险**：归一化结果与现存数据不一致。新数据用归一化值；老数据等 T10 重跑解析。

---

### T03 Linker 抽出独立组件 + 三级策略

**推荐执行模型**：**Sonnet 4.6**（涉及状态机+多级优先级+"答案不兜底全连"反直觉规则，错改代价高，**不要**降级派给便宜模型）

**目标**：把 `AssembleQuestions` 中的 `question/material/answer` 关联逻辑迁出来，成为 `parser/linker.go`。

**允许修改**：
- 新增 `backend/internal/parser/linker.go`
- 新增 `backend/internal/parser/linker_test.go`
- `backend/internal/services/essay_service.go`（AssembleQuestions 调 linker，不再内嵌策略）

**禁止修改**：
- handler / routes
- models（写入的字段不变）

**关键签名**：

```go
type LinkInput struct {
    Questions []models.EssaySection
    Materials []models.EssaySection
    Answers   []models.EssaySection
}

type LinkOutput struct {
    QuestionMaterial map[uint][]uint // questionSectionID -> materialSectionIDs
    QuestionAnswer   map[uint][]uint // questionSectionID -> answerSectionIDs
    Warnings         []string
}

func Link(input LinkInput) LinkOutput
```

**策略实现顺序**（**关键变更**：答案不再"所有答案关到所有题目"）：

1. exactMatch：用 question_no / related_question_nos / material_nos 严格命中。
2. heuristicMatch：扫题目正文 `根据材料\s*[一二三四1-9]+` / 答案正文 `第\s*[一二三四1-9]+\s*题`。
3. sequenceMatch：仅在 `len(answers)==len(questions)` 且全部未匹配时启用。
4. materialFallback：未关联的 material 默认关到所有 question（申论习惯）。
5. answerFallback：答案不兜底全连。未关联的 answer 输出 warning，由人工面板修正。

**验收**：
```bash
cd backend && go test ./internal/parser/... ./internal/services/...
```

**风险**：旧文档 assemble 出来的题目可能丢失"全部答案关到全部题目"的连接。可接受——本来就是噪声。前端审计点：UI 必须显示 Warnings。

---

### T04 SanityChecker：解析后健全性检查

**推荐执行模型**：DeepSeek V4 Pro（阈值统计 + 字段计数，主模型最后校阈值）

**目标**：解析完输出 warnings 报表，包含：
- material 总文字数占比是否 ≥ 30%
- question 数量在 1~10 之间
- question_no 是否连续无跳号
- answer 是否有未关联到 question 的孤儿

**允许修改**：
- 新增 `backend/internal/parser/sanity.go`
- 新增 `backend/internal/parser/sanity_test.go`
- `backend/internal/services/essay_service.go`（解析后调用 SanityCheck，warnings 写入 document.Note）
- `backend/internal/models/essay.go`（在 EssayDocument 加 `ParseWarnings string`，需迁移）

**禁止修改**：
- handler 返回字段名
- 已有 section 写入逻辑

**验收**：
```bash
cd backend && go test ./...
```

**风险**：阈值过严会大量误报。阈值放保守，warnings 是提示不是错误，不阻塞流程。

---

### T05 QuestionContext API + Service

**推荐执行模型**：DeepSeek V4 Pro（标准后端 CRUD + handler 路由，**派任务时必须强调用 IN 查询避免 N+1**）

**目标**：新增"取一道题完整批改上下文"的 API，前端只调一次就拿到全部所需材料。

**允许修改**：
- 新增 `backend/internal/services/essay_context.go`
- `backend/internal/handlers/essay.go`（新增 GetQuestionContext）
- `backend/internal/routes/routes.go`（新增 `GET /api/essay/questions/:id/context`）

**禁止修改**：
- 已有 ReviewAnswer 行为（但 ReviewAnswer 内部应改为复用 QuestionContextService.Build）

**返回结构**：

```json
{
  "question": { ... EssayQuestion ... },
  "materials": [ { id, title, content, page_start, page_end, source_document_id } ],
  "answers":   [ ... ],
  "analyses":  [ ... ],
  "sources":   [ { document_id, role, title } ],
  "warnings":  [ "答案 3 未关联题目" ]
}
```

**验收**：
```bash
cd backend && go test ./...
curl -s -H "X-User-ID: 1" http://localhost:21080/api/essay/questions/1/context | jq
```

**风险**：N+1 查询。一次性 IN 查询取出所有 section。

---

### T06 跨文档（source_group）拼装

**推荐执行模型**：**Sonnet 4.6**（跨文档级联 + 删除路径 + relation 表跨 document_id 引用，错一处掉数据，**不要降级**）

**目标**：题目卷 + 答案卷分别上传也能正常批改。

**允许修改**：
- `backend/internal/services/essay_service.go`（新增 AssembleBySourceGroup）
- `backend/internal/handlers/essay.go`（新增 handler）
- `backend/internal/routes/routes.go`（`POST /api/essay/groups/:source_group/assemble`）

**禁止修改**：
- 模型字段（已有 `source_group` `document_role`）

**实现要点**：
1. 按 `user_id + source_group` 查所有 documents。
2. 找一个 `document_role=question_paper`（或 combined）作为题目宿主，其余作为补充。
3. 把所有文档的 sections 喂进 Linker。
4. 创建 EssayQuestion 时 `document_id` 指向题目宿主。
5. `EssaySectionRelation.SectionID` 可以跨文档指向其他文档下的 section（表里已有，只要 user_id 匹配即可）。

**验收**：
- 上传 A（题目卷）、B（答案卷），相同 `source_group`。
- 调 `POST /api/essay/groups/2025-省考A/assemble`。
- 调 `GET /api/essay/questions/:id/context` 应返回 B 中的 answer section。

**风险**：跨文档 section 删除时关联未清。删除 document 时要级联清掉它 section 的所有 relation。

---

### T07 BoundaryPrompt 加 few-shot + JSON Schema 提示

**推荐执行模型**：**Kimi K2.6**（中文 prompt 工程 + few-shot 中文样卷撰写是 Kimi 强项，比 DSV4/Sonnet 写出来的更"像申论"）

**目标**：让 LLM 在不同排版下输出更稳定的 BoundaryPlan。

**允许修改**：
- `backend/internal/parser/boundary_plan.go`（BuildBoundaryPromptFromLines、BuildBoundarySystemPrompt）
- 新增 `backend/internal/parser/boundary_prompt_examples.go`（保存 2~3 个 few-shot 样例的脱敏文本）

**禁止修改**：
- service 层调用方
- prompt 输入参数签名

**要点**：
- 系统 prompt 中加入"输出必须是合法 JSON，禁止任何 markdown"。
- 用户 prompt 中加入 2 个简短 few-shot：一份混合卷、一份纯题目卷。
- 在 prompt 末尾给出"如果你识别不出某段，请使用 unknown 类型而不是丢弃"。
- 行号范围必须覆盖 [1, N]——强调，不允许重叠或留空。

**验收**：
- 用 debug 接口在真实样卷上跑，sections 覆盖率 ≥ 95%。
- `npm run build` 不受影响。

**风险**：prompt 加长意味着 token 上涨。控制在 2k tokens 内（few-shot 用脱敏短样本）。

---

### T08 前端：QuestionContext 切换 + 关联展示

**推荐执行模型**：**Sonnet 4.6**（EssayReview.jsx 状态复杂、AntD 组件多，DSV4 容易漏 import 出空白页）

**目标**：前端"选中题目"时使用新 context API，明确显示与题目关联的材料/答案，不再"全部材料一锅炖"。

**允许修改**：
- `src/api/essay.js`（新增 `getEssayQuestionContext(id)`）
- `src/pages/EssayReview.jsx`（替换 relatedMaterials / relatedAnswers 计算逻辑）
- `src/styles/pages.css`（必要的小幅样式）

**禁止修改**：
- 文档列表、上传表单、批改结果卡片（避免噪声）
- 后端

**展示要求**：
- 在题目详情卡片顶部显示 warnings（来自 context.warnings）。
- 材料卡片标题显示"来自文档：A 卷（题目卷）"。
- 答案卡片同理。

**验收**：
```bash
npm run build
```
人工：上传一份混合卷，选中第一题，能看到只有"该题关联的材料/答案"，不是全部。

**风险**：旧文档 relation 表为空，需要给"该 document 没有 relation 时回退到全部 section"的兜底，避免页面空白。

---

### T09 前端：手动编辑题号/分值/字数/关联

**推荐执行模型**：**Sonnet 4.6**（前后端联动 + AntD 表单 + relation 增删，跨文件改动多）

**目标**：解析错位时用户能直接修正而不必重新上传。

**允许修改**：
- 后端：`backend/internal/handlers/essay.go` 新增 `PUT /api/essay/questions/:id` 与 `POST /api/essay/questions/:id/relations`、`DELETE /api/essay/questions/:id/relations/:relation_id`
- 前端：`src/pages/EssayReview.jsx`（在题目详情卡片加"编辑"折叠）
- `src/api/essay.js`

**禁止修改**：
- 解析流程
- LLM prompt

**API 行为**：
- PUT 只允许改 `question_no` `question_type` `max_score` `word_limit` `title` `question_text`。
- POST relations 增加一条 question_material 或 question_answer。
- DELETE relations 删一条。

**验收**：
```bash
cd backend && go test ./...
npm run build
```

**风险**：用户改坏会影响批改。无需做版本，但操作前给确认弹窗。

---

### T10 清理旧 chunk 路径

**推荐执行模型**：DeepSeek V4 Pro（机械删除清理，必须强调"不删表、不删模型"）

**目标**：移除已无意义的 EssayChunk 写入和 `/classify` 路由（视图保留但 deprecated）。

**允许修改**：
- `backend/internal/handlers/essay.go`：删除 ClassifyChunks handler 方法及对应路由
- `backend/internal/routes/routes.go`：删除 `/classify` 与 `/chunks` 路由
- `backend/internal/services/essay_service.go`：删除 ClassifyChunks 方法、删除 EssayChunk 视图函数（保留模型暂不删表）
- `src/api/essay.js`：删除 classifyEssayChunks / getEssayChunks
- `src/pages/EssayReview.jsx`：移除任何 chunks 相关引用

**禁止修改**：
- 不要删表，留待 T12 一并处理

**验收**：
```bash
cd backend && gofmt -w cmd internal && go test ./...
cd .. && npm run build
```

**风险**：还有 mock 代码引用 chunks。grep 一次确保没漏。

---

### T11 Prompt 接入前端 Prompt 模板管理

**推荐执行模型**：DeepSeek V4 Pro（DB + fallback 链路，**主模型必须审 fallback 是否完整**；DSV4 容易写漏"找不到模板时用代码内置"分支）

**目标**：边界切分 prompt 与批改 prompt 可在 `/ai?tab=prompts` 里编辑。

**允许修改**：
- `backend/internal/services/essay_service.go`：在调用 LLM 前优先读 `prompt_templates`（按 `name="essay_boundary"` 等），找不到再用代码内置 fallback。
- `src/pages/AISettings.jsx` 的 PromptsCapability（如已存在则只加预置项）。

**禁止修改**：
- prompt_templates 表结构（已存在）
- 已有 BuildBoundary* 函数签名（让它们成为 fallback）

**预置模板名**：
- `essay_boundary_system`
- `essay_boundary_user`
- `essay_review_system`
- `essay_review_user`

**验收**：
```bash
cd backend && go test ./...
npm run build
```

**风险**：用户写错 prompt 导致 LLM 不返回 JSON。要保留"恢复默认"按钮（前端）+ fallback（后端）。

---

### T12 真实样卷回归测试 + 文档收尾

**推荐执行模型**：**Kimi K2.6**（中文样卷 fixture 撰写 + 中文文档润色是 Kimi 最适合的活）

**目标**：解析路径有最小覆盖的 fixture 测试，并把整改后的能力同步到 PROJECT_CHECKLIST / DEVELOPMENT_OVERVIEW。

**允许修改**：
- 新增 `backend/internal/parser/testdata/essay/*.txt`（脱敏的 PDF 提取文本，3~5 份）
- 新增 `backend/internal/parser/integration_test.go`（不依赖真实 LLM，使用 mock plan 验证 Linker 行为）
- `docs/PROJECT_CHECKLIST.md`：更新 API 列表与"下一步建议"。
- `docs/DEVELOPMENT_OVERVIEW.md`：第 10 节"申论 PDF 流程"重写。

**禁止修改**：
- 引入真实 LLM 调用到测试

**验收**：
```bash
cd backend && go test ./...
```

**风险**：fixture 文本里如果含个人/单位敏感信息要做脱敏。

---

## ── 综应 C 类兼容增量（T13–T16）──

下列任务实现"申论 / 综应 C 类"双轨。**前置依赖**：T01–T12 完成（特别是 T03 Linker 与 T11 prompt 模板）。

---

### T13 paper_type 字段引入 + 默认值

**推荐执行模型**：DeepSeek V4 Pro（字段迁移 + 表单选项，机械任务）

**目标**：让 `EssayDocument` 携带卷面类型字段，但**不改动任何解析/批改行为**（行为分叉留给 T14–T16），仅做"字段贯通"。

**允许修改**：
- `backend/internal/models/essay.go`：`EssayDocument` 加 `PaperType string` 字段，size 30，默认 `shenlun`，加索引
- `backend/internal/database/migrate.go`：把字段纳入 AutoMigrate（如需要手动 SQL 补，列在文件注释里）
- `backend/internal/handlers/essay.go`：`CreateDocument` 接收 `paper_type` form 字段，校验只能是 `shenlun` 或 `zongying_c`，落库
- `src/api/essay.js`：`createEssayDocument` 透传 `paperType`
- `src/pages/EssayReview.jsx`：上传表单加 PaperType 下拉（默认申论），文档列表卡片显示 PaperType 标签

**禁止修改**：
- 任何 service 的解析 / linker / review 行为（继续按"shenlun"走，**T14–T16 才分支**）
- Prompt 文本

**校验项**：
- `paper_type` 只接受 `shenlun` 与 `zongying_c`，其他值返回 400。
- 已有 document 升级时 `paper_type` 留空，读取时 fallback 为 `shenlun`。

**验收**：
```bash
cd backend && gofmt -w cmd internal && go test ./...
cd .. && npm run build
```

**风险**：DSV4 容易"顺手"把 prompt 也按 paper_type 分支了——明确禁止越界。

---

### T14 子题层级 (parent_question_id) + 综应 C 类 Linker 分支

**推荐执行模型**：**Sonnet 4.6**（涉及二级父子结构、Linker 内部分支、跨 paper_type 的边界条件，错改污染数据）

**目标**：支持"大题—子问"二级结构，并为综应 C 类增加 Linker 分支。

**允许修改**：
- `backend/internal/models/essay.go`：`EssayQuestion` 加 `ParentQuestionID uint`（0 = 顶级）+ `SubIndex string`（如 "1.1" "(1)"）
- `backend/internal/database/migrate.go`：迁移
- `backend/internal/parser/linker.go`：在 T03 抽出的 linker 中加 `paper_type` 分支：
  - `shenlun`：保留 T03 的现有策略
  - `zongying_c`：
    1. 识别 question section 内"子问"模式：`(\d+)\.(\d+)` `（[一二三四1-9]+）` `第[一二三四1-9]+问`
    2. 把一个 question section 拆成多个 EssayQuestion，第一个作为 parent，其余 ParentQuestionID = parent.ID
    3. 材料只关联到 parent（子问通过 parent 继承材料）
    4. 答案严格按"父题号.子问号"匹配到子问
- `backend/internal/services/essay_service.go`：AssembleQuestions 按 `document.PaperType` 派分；`QuestionContextService.Build` 取 ParentQuestionID 不为 0 的题时，自动并入父题的材料
- `src/pages/EssayReview.jsx`：题目列表显示父子缩进（综应 C 类才显示）

**禁止修改**：
- 申论现有行为（默认走 shenlun 分支）
- Review 的批改 prompt（T16）

**测试要求**（必写）：
- shenlun 卷不引入 parent_question_id（全部为 0）
- zongying_c 卷的 3 大题 + 子问能正确拆出 parent/child
- QuestionContext 对子问返回父题的材料

**验收**：
```bash
cd backend && go test ./internal/parser/... ./internal/services/...
cd .. && npm run build
```

**风险**：
- Linker 的状态机一旦写错会把申论卷也拆成父子。务必单测覆盖 shenlun 路径未变。
- 删除 parent 时是否级联删除 children——本任务**不**实现级联（标记为后续工作），但禁止 children 残留 ParentQuestionID 指向已删除题。

---

### T15 客观题答案结构化 + 分流批改

**推荐执行模型**：**Sonnet 4.6**（涉及多态批改路径 + JSON schema + LLM 回退，容易写漏）

**目标**：批改路径按题型分流，客观题先比对再让 LLM 解释。

**允许修改**：
- `backend/internal/models/essay.go`：`EssayQuestion` 加：
  - `ObjectiveType string`（`subjective` / `judgment` / `choice` / `numeric`，默认 `subjective`）
  - `ReferenceAnswerJSON string`（type:text；结构化参考答案）
- `backend/internal/database/migrate.go`：迁移
- `backend/internal/services/essay_service.go`：
  - `ReviewAnswer` 按 `ObjectiveType` 分支：
    - `subjective`：现有 LLM 全文批改
    - `judgment`：从 `ReferenceAnswerJSON` 取 `correct_judgment`，对比考生作答关键词；命中给主分，让 LLM 补"理由是否充分"评分
    - `choice`：对比 `correct_options` 集合，全对/部分对/全错按 rubric 给分；让 LLM 补"思路解释"
    - `numeric`：对比 `correct_numeric` ± `tolerance`，命中给数值分；让 LLM 评过程
  - 客观题分支即使 LLM 不可用也能给出客观部分分数（不再回退到 mockEssayScore）
- `src/pages/EssayReview.jsx`：题目详情显示 ObjectiveType 标签；作答区按 ObjectiveType 显示不同输入（subjective=多行文本；judgment=单选+理由文本；choice=多选+理由；numeric=数字+过程文本）

**禁止修改**：
- Parser（题型识别由 T16 的 prompt 输出，本任务只读字段）
- 申论现有批改路径（ObjectiveType 默认 subjective 即维持原行为）

**参考答案 JSON 结构**：

```json
{
  "correct_judgment": "wrong",
  "correct_options": ["A", "C"],
  "correct_numeric": 42.5,
  "tolerance": 0.1,
  "scoring_rubric": [
    { "key": "审题准确", "score": 5 },
    { "key": "理由充分", "score": 5 }
  ]
}
```

**验收**：
```bash
cd backend && go test ./...
cd .. && npm run build
```

**风险**：客观题字符串比对要做归一（大小写、全半角、中英文标点）；LLM 不可用时仍要返回部分分。

---

### T16 综应 C 类专属 Boundary / Review Prompt 模板

**推荐执行模型**：**Kimi K2.6**（核心是中文 prompt 工程 + few-shot 样卷撰写 + 题型差异化叙述）

**目标**：为综应 C 类提供独立的边界识别与批改 prompt，并预置进 prompt_templates 表。

**允许修改**：
- 新增 `backend/internal/parser/boundary_prompt_zongying.go`：
  - `BuildBoundaryPromptZongyingC(lines)` — 强调"一篇科技文献 + 三大题（科技文献阅读 / 数理 / 策略选择） + 每大题有子问"
  - 输出 schema 与申论共用 BoundaryPlan（兼容现有 sections 表），但 question section 必须输出 `sub_index` 字段（写进 reason 字段或 plan 扩展字段，本任务选其一并写注释）
- `backend/internal/services/essay_service.go`：`suggestBoundaryPlanFromLines` 按 `document.PaperType` 选择 prompt 模板（先读 prompt_templates 表中名为 `zongying_c_boundary_system` / `zongying_c_boundary_user` 的模板，没有再用代码内置 fallback）
- `backend/internal/services/essay_service.go`：`callLLMReview` 同样按 paper_type 选 `zongying_c_review_system` / `zongying_c_review_user`，review prompt 要按 `ObjectiveType` 改变"评分要求"
- 数据初始化：在 `backend/internal/services/prompt_service.go`（或新建 `prompt_seed.go`）加预置模板写入函数，启动时（仅当 prompt_templates 中无同名模板时）插入：
  - `essay_boundary_system` / `essay_boundary_user`（把当前申论 prompt 抽出）
  - `essay_review_system` / `essay_review_user`
  - `zongying_c_boundary_system` / `zongying_c_boundary_user`
  - `zongying_c_review_system` / `zongying_c_review_user`

**禁止修改**：
- prompt_templates 表结构
- 申论 prompt 文字（除"抽出代码硬编码搬进数据库"这步无可避免的迁移，文字不动）
- Linker / batch parse 主流程

**Prompt 写作要点**（必须在 prompt 中体现）：

综应 C 类 Boundary prompt：
- 卷面只有一篇科技/工程类文献作为材料，**不会**有"给定资料一/二/三"格式
- 大题区开头通常是"题一/题二/题三"或"第一题/第二题/第三题"
- 每大题下用 "1." / "（一）" / "第一问" 等开启子问
- 子问下可能直接是"答案：B" 这类客观题选项
- 答案区不一定独立成段——题目内可能直接附带答案行

综应 C 类 Review prompt：
- 按 `objective_type` 给出不同评分指引
- 强调过程分（数理题）
- 客观题部分要求 LLM 不重复给客观分（已由后端比对），只评"理由 / 过程 / 解释"

**验收**：
```bash
cd backend && go test ./...
```
人工：上传一份综应 C 类样卷，能切出 1 篇 material + 3 个父题 + 若干子问；客观题字段填出。

**风险**：
- 预置模板的 idempotent 写入要严格按 name 去重，不能覆盖用户编辑过的版本。
- Kimi 喜欢把变量名中文化——明确要求模板 name 字段用英文 snake_case。

---

## 4. 主模型审计通用要点

每条任务合并前：

- 是否只改了允许范围的文件。
- `go test ./...` 与 `npm run build` 是否通过。
- 是否破坏现有的 `parse → assemble → review` 主流程。
- 数据库模型新增字段是否在 `database.AutoMigrate` 列表中。
- 是否暴露任何 LLM Key / 本地路径 / 用户 ID 到前端。
- 是否引入循环 import（parser ↔ services ↔ models）。
- 是否漏掉端口约定：前端 21073、后端 21080、PG 21432。

## 5. 不在本轮范围内（明确排除）

- PDF 扫描件转图片 + OCR pipeline（仍按现有 OCR 模块单独处理）。
- 题目搜索 / 题库导入。
- 多用户认证。
- 自动给学习计划生成训练任务。
- 综应 A/B/D/E 类、行测、判断推理等其他考卷类型——本轮只覆盖申论与综应 C 类，其他类型若需要走相同的 `paper_type` 扩展模式。
