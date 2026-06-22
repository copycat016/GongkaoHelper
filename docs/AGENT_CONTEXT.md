# AI 接手上下文

最后更新：2026-06-16

这个文件用于降低后续对话的 token 消耗。新的 AI 会话或人工接手时，优先阅读本文件，再按任务读取具体源码；不要把 README、TODO、维护手册和长 diff 全量塞进对话。

## 使用方式

1. 开新会话时只贴一句：`请先读 docs/AGENT_CONTEXT.md，再处理当前任务。`
2. 每完成一个有架构影响的任务，更新本文件的「当前状态」「近期决策」「待办焦点」三块。
3. 单次更新保持精简：只写稳定结论、入口文件、未解决风险；不要复制大段代码、日志、完整接口返回。
4. 如果信息已经在 README / QUICKSTART / MAINTENANCE / backend-dev 中长期维护，这里只放链接和差异。

## 项目概况

- 前端：Vite + React 19 + Ant Design 6 + React Router 7。
- 后端：Go + Gin + GORM。
- 数据库：默认 SQLite，可选 PostgreSQL。
- 运行形态：开发态前端 `21073` 代理后端 `21080`；生产态 Docker 单容器提供 Web + API。
- 核心功能：统一录入器、申论 PDF 结构化与批改、番茄钟、学习日志、三层规划、本地音乐、LLM/OCR/PDF/备份配置。

## 快速入口

- 前端路由：[src/AppRoutes.jsx](../src/AppRoutes.jsx)
- 前端请求封装：[src/api/request.js](../src/api/request.js)
- Dashboard 首页容器：[src/pages/Dashboard.jsx](../src/pages/Dashboard.jsx)
- Dashboard 展示拆分：[src/pages/dashboard/](../src/pages/dashboard/)
- 后端入口：[backend/cmd/server/main.go](../backend/cmd/server/main.go)
- 路由注册：[backend/internal/routes/routes.go](../backend/internal/routes/routes.go)
- 配置读取：[backend/internal/config/config.go](../backend/internal/config/config.go)
- 数据迁移：[backend/internal/database/migrate.go](../backend/internal/database/migrate.go)
- 申论主服务：[backend/internal/services/essay_service.go](../backend/internal/services/essay_service.go)
- 三层规划服务：[backend/internal/services/daily_task_service.go](../backend/internal/services/daily_task_service.go)
- 鉴权服务：[backend/internal/services/auth_service.go](../backend/internal/services/auth_service.go)

## 当前状态

- 鉴权已切到 JWT：`/api/auth/login` 公开，其余业务接口走 `AuthRequired`。
- 单人模式仍是默认：空库会创建 owner 用户。
- 开发模式空库默认 `admin / 123456`，仅用于本地。
- release 模式必须配置 `JWT_SECRET` 和强 `AUTH_BOOTSTRAP_PASSWORD`；发布安装文档会生成随机值。
- 主题已接后端 `/api/theme`，前端仍保留 localStorage fallback。
- Dashboard 已拆为容器页、展示区块、列表/弹窗组件、工具函数，样式迁入 `src/styles/dashboard.css`。
- 备份导出已覆盖当前迁移表，并增加上传文件 manifest；仍未实现导入恢复。
- 三层规划更新接口已具备 PATCH 语义：未传字段保留，显式传空值/null 才清空。
- 周任务拆日任务已收口到后端事务接口 `POST /api/planning/weekly-tasks/:id/materialize`；新生成日任务标记 `origin=weekly_materialized`，保存执行方式时可清理过期未完成物化任务。
- 长期规划已新增阶段子项层 `stage_items`：推荐按阶段目标 -> 阶段子项 -> 推进任务 -> 日任务拆解。
- 推进任务支持 `task_kind=standard_week` 和 `task_kind=special_project`：标准周按周一到周日归档，专项项目按 `start_date/end_date` 跨周重叠查询。
- 周推进任务和日任务均可挂 `stage_item_id`；后端会校验子项归属并自动推导 `stage_goal_id`，物化日任务会继承目标/子项关联。
- Dashboard 日任务列表已显示来源标签：手动、周任务关联、周任务生成；编辑物化日任务时会提示其来源关系。
- 申论批改记录已保存当次题面、材料、参考答案、Prompt、模型和模板快照；自定义 Prompt 模板会参与批改 Prompt 构建。
- 前端路由已改为 lazy loading，主包不再静态包含所有大页面。
- 题库/错题库 UI 已从主流程移除，部分后端旧代码和文档残留仍需清理。
- `npm run build` 可通过；路由懒加载后无 500KB chunk 警告。
- 后端测试可通过，但覆盖主要集中在 parser 和 auth bootstrap。

## 近期决策

- 生产发布优先保证可启动和基本安全：禁止 release 空库使用固定默认密码。
- `AUTH_BOOTSTRAP_PASSWORD` 只用于空库初始化或轮换旧开发默认密码，不是长期密码管理系统。
- 上传文件目录和数据库目录在 release 安装中改为 `chmod 700`。
- 业务接口示例需要携带 Bearer Token；旧的 `X-User-ID` 口径作废。
- 前端大页面先按低风险边界拆：容器负责取数/事件，展示组件只接 props，页面专属样式从 `pages.css` 外移。
- 三层规划的拆解/安排属于业务动作，放后端事务化；前端只提交执行方式和日期。
- 任务类型 `task_kind` 与执行方式 `execute_mode` 分离：前者回答“这是标准周还是中短期专项”，后者回答“是否进入日计划、按单日还是多日拆解”。

## 待办焦点

1. 业务初始化页：首次打开 Web 时设置管理员密码、展示备份目录、确认部署模式。
2. 上传安全：给 OCR/PDF/音乐上传加大小限制、MIME 校验、用户配额、失败文件清理。
3. 备份恢复：从“完整导出 JSON + 文件 manifest”升级为可导入恢复，覆盖数据库、上传文件、版本兼容检查和冲突策略。
4. 申论业务闭环：下一步补批改历史列表/回放 UI，并把 raw request/raw response 与模型错误状态暴露给用户。
5. 统一录入数据模型：把 OCR、PDF、粘贴文本统一落成可版本化的 `DocumentSource`，减少 sessionStorage 大文本流转。
6. 三层规划精细化：下一步补阶段子项 UI、标准周/专项项目切换 UI、批量改期、拖拽排序、阶段目标进度策略配置，以及更完整的生成关系回放。
7. 文档清理：删除或标记 `/api/mistakes`、`/api/plans`、`/api/calendar` 等旧接口示例。
8. 前端性能：对 AntD/页面做按路由 code splitting。
9. 继续拆大文件：优先 `EssayReview.jsx`、`MusicPlayer.jsx`、`backend/internal/services/essay_service.go`，按业务阶段和纯函数边界拆。

## 常用验证

```bash
npm run build
GOCACHE=/tmp/gocache go test ./...
GOCACHE=/tmp/gocache go vet ./...
```

如果 Go 命令报 `/root/.cache/go-build` 只读，使用 `GOCACHE=/tmp/gocache`。

## 上下文节流规则

- 查文件先用 `rg` / `rg --files`，只打开命中的小范围。
- 长文件优先读函数入口和调用链，不整文件粘贴。
- 让 AI 报告问题时要求输出：严重级别、文件位置、影响、建议，不要输出完整 diff。
- 每次大改后把“真正改变系统行为”的结论写回本文件，避免下次从 git diff 重新推理。
