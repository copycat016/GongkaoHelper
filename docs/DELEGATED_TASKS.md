# 可委派任务清单

这份清单用于把低风险、边界清楚的工作交给廉价模型完成，再由主模型负责审计、整合和最终把关。

原则：

- 每次只交一个任务，不要让廉价模型跨多个模块大改。
- 只允许它改任务中列出的文件或目录。
- 让它先读相关文件，再改代码。
- 让它最后必须写清楚：改了什么、如何验证、还有什么风险。
- 主模型审计时重点看：是否破坏现有流程、是否引入重复逻辑、是否误删用户代码、是否缺少错误处理。

## 任务 1：清理前端 lint 中的简单问题

适合模型：廉价代码模型。

目标：

修复不影响业务逻辑的 lint 问题，例如未使用变量、导出位置不合理、简单 hook 依赖提示。

允许修改：

- `src/api/theme.js`
- `src/components/MusicProvider.jsx`
- `src/components/Sidebar.jsx`
- `src/components/ThemePanel.jsx`

禁止修改：

- 不要重写组件结构。
- 不要改路由。
- 不要改 API 行为。
- 不要为了消除 lint 大幅重构状态逻辑。

具体步骤：

1. 运行 `npm run lint`，记录当前报错。
2. 只处理明显安全的问题：
   - 删除未使用变量。
   - 把非组件导出移动到独立文件，或用更小改动消除 fast-refresh 报错。
   - 对无风险的 hook 依赖做补齐。
3. 每改完一组文件，运行 `npm run build`。
4. 最后再次运行 `npm run lint`，列出剩余未处理问题。

验收标准：

- `npm run build` 通过。
- 不改变界面功能。
- 最终说明哪些 lint 还没修，为什么暂时不修。

主模型审计点：

- 是否引入循环依赖。
- `MusicProvider` 是否仍能跨页面后台播放。
- `Sidebar` 菜单是否仍正常选中。

## 任务 2：申论 PDF 测试输出体验优化

适合模型：廉价前端模型。

目标：

优化 `配置 -> PDF 解析` 中的 PDF 测试输出，让用户更容易判断 PDF 是正常文本、乱码还是扫描件。

允许修改：

- `src/pages/AISettings.jsx`
- `src/styles/pages.css`
- `src/api/pdf.js`

禁止修改：

- 不要改后端接口。
- 不要改申论主流程页面。
- 不要引入新依赖。

具体步骤：

1. 阅读 `PDFCapability` 组件。
2. 在测试结果区域增加：
   - 每页字符数。
   - 一键复制某页文本。
   - 文本层异常时的明显提示。
3. 输出区域要支持长文本滚动，不要撑爆页面。
4. 保持当前清新蓝色风格。
5. 运行 `npm run build`。

验收标准：

- 上传 PDF 后能看到质量判断。
- 每页文本不再挤成难读的一坨。
- 页面在平板宽度下不横向溢出。

主模型审计点：

- 是否把 PDF 配置页又写得过于复杂。
- 是否和 OCR 配置重复解释太多。
- 是否有未导入组件导致空白页。

## 任务 3：OCR 结果展示优化

适合模型：廉价前端模型。

目标：

让 OCR 原文和 AI 修正文稿更适合阅读、复制和对照。

允许修改：

- `src/pages/OCRQuestion.jsx`
- `src/styles/pages.css`

禁止修改：

- 不要改 OCR 后端。
- 不要改百度 AK/SK 配置。
- 不要改 OCR API 请求路径。

具体步骤：

1. 阅读 `OCRQuestion.jsx`。
2. 给 OCR 原文区和 AI 修正文稿区增加：
   - 复制按钮。
   - 清空按钮。
   - 字数统计。
   - 等宽/正常字体切换可以做，但不是必须。
3. 确保 TextArea 不出现“一字一列”的视觉问题。
4. 运行 `npm run build`。

验收标准：

- OCR 文本显示自然换行。
- 用户可以快速复制识别结果。
- 不影响现有“开始 OCR / AI 修正”按钮。

主模型审计点：

- OCR 状态是否仍会显示用量。
- 上传文件状态是否仍正常。
- 是否误改结构化题目表单。

## 任务 4：申论 section / chunk 卡片展示细化

适合模型：廉价前端模型。

目标：

让申论 PDF 结构化页面中的 section（语义区域）和 chunk（固定长度片段）更容易人工检查。sections 是当前主流程使用的模型，chunks 保留向后兼容。

允许修改：

- `src/pages/EssayReview.jsx`
- `src/styles/pages.css`

禁止修改：

- 不要改后端模型。
- 不要改分类接口。
- 不要改 PDF 解析接口。

具体步骤：

1. 阅读 `EssayReview.jsx` 中 sections 和 chunks 展示区域。
2. 增加筛选控件：
   - 按 section/chunk 类型筛选。
   - 按页码筛选或快速跳页。
3. 每个 section 卡片保留：
   - 页码范围（page_start / page_end）。
   - 类型（material / question / answer / analysis / unknown）。
   - 置信度。
   - 分类依据。
   - 原文内容。
4. 每个 chunk 卡片保留：
   - 页码。
   - 顺序。
   - 分类。
   - 置信度。
   - 分类依据。
   - 原文。
5. 不要把卡片做得太花。
6. 运行 `npm run build`。

验收标准：

- 用户可以只看题目、材料、参考答案或评分规则。
- 长文本不会撑坏布局。
- 平板下仍可读。

主模型审计点：

- 筛选状态是否和分页冲突。
- 是否错误隐藏 unknown section/chunk。
- 是否影响后续组装题目。

## 任务 5：后端 PDF 解析测试接口补充单元测试

适合模型：廉价 Go 模型。

目标：

给 PDF 文本清洗和质量判断函数补测试，不要求测试真实 PDF。

允许修改：

- `backend/internal/services/pdf_text_test.go`
- 必要时小幅调整 `backend/internal/services/pdf_text.go`

禁止修改：

- 不要改路由。
- 不要改 handler。
- 不要引入大型 PDF fixture。
- 不要依赖本机必须安装 `pdftotext`。

具体步骤：

1. 给 `sanitizePostgresText` 写测试：
   - 移除 `\x00`。
   - 非法 UTF-8 转为空。
2. 给 `assessPDFTextQuality` 写测试：
   - 正常中文文本应 OK。
   - 大量乱码/符号应 not OK。
   - 文本太少应 not OK。
3. 运行：
   - `cd backend`
   - `go test ./...`

验收标准：

- `go test ./...` 通过。
- 不需要真实 PDF 文件。
- 不需要外部命令。

主模型审计点：

- 测试是否过度贴合当前实现。
- 质量阈值是否太死。
- 是否影响真实 PDF 解析流程。

## 任务 6：Poppler 安装说明补文档

适合模型：廉价文档模型。

目标：

补充 PDF 解析依赖说明，让部署时知道如何安装 `pdftotext`。

允许修改：

- `README.md`
- `backend/docs/backend-dev.md`

禁止修改：

- 不要改代码。
- 不要写不确定的命令为唯一方案。

具体步骤：

1. 在后端启动说明中增加 PDF 解析说明。
2. 写清楚当前策略：
   - 优先使用系统 `pdftotext`。
   - 找不到时回退 Go PDF 库。
   - 扫描件需要 OCR。
3. 给出常见安装命令：
   - Debian/Ubuntu：`sudo apt install poppler-utils`
   - Fedora：`sudo dnf install poppler-utils`
   - Arch：`sudo pacman -S poppler`
4. 写验证命令：`pdftotext -v`

验收标准：

- 文档明确但不过度承诺。
- 用户能知道为什么有些 PDF 仍然需要 OCR。

主模型审计点：

- 是否把 Poppler 写成强制依赖。
- 是否说明了 fallback。
- 是否和当前代码行为一致。

## 任务 7：学习管理三 Tab 的视觉一致性

适合模型：廉价前端模型。

目标：

统一学习计划、学习日志、日历视图三个 Tab 的卡片密度、标题、筛选区样式。

允许修改：

- `src/pages/StudyCenter.jsx`
- `src/pages/StudyLogs.jsx`
- `src/pages/StudyPlans.jsx`
- `src/pages/CalendarPage.jsx`
- `src/styles/pages.css`

禁止修改：

- 不要改后端 API。
- 不要重做数据结构。
- 不要改番茄钟写日志逻辑。

具体步骤：

1. 阅读三个页面当前布局。
2. 统一筛选区卡片高度和按钮位置。
3. 日/周/月统计切换要明显。
4. 日历事件列表不要太弱，至少突出当天任务。
5. 运行 `npm run build`。

验收标准：

- 三个 Tab 看起来像同一套页面。
- 平板下不挤压。
- 日周月切换仍正常。

主模型审计点：

- 是否影响日期统计逻辑。
- 是否让日历功能变成纯装饰。
- 是否误删番茄钟联动字段。

## 任务 8：音乐播放器元数据失败提示

适合模型：廉价前端/后端均可，但建议前端。

目标：

当音乐元数据识别不到标题、作者、专辑时，前端给出可编辑提示。

允许修改：

- `src/pages/MusicPlayer.jsx`
- `src/styles/pages.css`

禁止修改：

- 不要改上传存储路径。
- 不要改全局播放器 Provider。
- 不要改后端元数据解析库。

具体步骤：

1. 阅读 `MusicPlayer.jsx`。
2. 上传表单中保留手填标题/作者/专辑入口。
3. 如果列表中的 track 缺少 artist/title，用温和提示显示“未识别”。
4. 运行 `npm run build`。

验收标准：

- 缺元数据时用户不困惑。
- 不影响后台播放栏。

主模型审计点：

- 是否破坏 `MusicProvider` 的状态。
- 是否上传后仍能播放。

## 任务 9：后端错误返回文案整理

适合模型：廉价 Go 模型。

目标：

统一后端用户可见错误文案，尤其是 OCR、PDF、Essay 相关错误。

允许修改：

- `backend/internal/handlers/essay.go`
- `backend/internal/handlers/pdf.go`
- `backend/internal/handlers/ocr.go`
- 必要时 `backend/internal/response/response.go`

禁止修改：

- 不要改 HTTP 状态码策略。
- 不要改数据库模型。
- 不要吞掉真实错误。

具体步骤：

1. 搜索 `response.Error`。
2. 只整理用户会看到的 message。
3. PDF 乱码、扫描件、OCR 配置缺失要有明确中文提示。
4. 运行 `go test ./...`。

验收标准：

- 错误信息能指导下一步。
- 不暴露 API Key、Secret、完整本地路径。

主模型审计点：

- 是否隐藏了调试需要的信息。
- 是否泄露敏感信息。
- 是否误把系统错误都变成 200。

## 任务 10：API 封装补充上传 helper

适合模型：廉价前端模型。

目标：

减少 `src/api/essay.js`、`src/api/pdf.js`、`src/api/music.js` 中重复的 FormData 上传和 response unwrap 逻辑。

允许修改：

- `src/api/request.js`
- `src/api/essay.js`
- `src/api/pdf.js`
- `src/api/music.js`

禁止修改：

- 不要改变现有接口路径。
- 不要改变后端。
- 不要引入 axios 重写。

具体步骤：

1. 在 `request.js` 中增加 `upload(path, formData)` helper。
2. 迁移 essay/pdf/music 的上传调用。
3. 确认 JSON API 不受影响。
4. 运行 `npm run build`。

验收标准：

- 上传音乐仍可播放。
- 上传 PDF 测试仍能输出文字。
- 上传申论 PDF 仍会自动解析。

主模型审计点：

- 是否错误设置 `Content-Type`，导致 multipart boundary 丢失。
- 是否破坏统一错误提示。

## 推荐分配顺序

1. 先做任务 5 和任务 6：后端 PDF 质量判断测试 + 文档说明。
2. 再做任务 2 和任务 3：PDF 测试输出 + OCR 展示。
3. 然后做任务 4：申论 chunk 检查体验。
4. 最后做任务 1 和任务 10：lint 与 API helper，风险稍高，适合主模型审计后再合。

## 主模型审计通用清单

每个委派任务完成后，主模型需要检查：

- 是否只改了允许范围内的文件。
- `npm run build` 或 `go test ./...` 是否通过。
- 是否有空白页风险，例如 JSX 里用了未导入的 AntD 组件。
- 是否把 mock、占位、真实接口混在一起。
- 是否把用户数据、密钥、上传文件路径暴露到前端。
- 是否破坏当前端口约定：前端 21073，后端 21080，数据库 21432，其它服务尽量 21000 以后。
