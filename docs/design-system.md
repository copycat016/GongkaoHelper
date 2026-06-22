# GongkaoHelper 前端设计系统规范

> 本文件有两个用途：
> 1. **给人看**——一页讲清整套界面的设计语言、准则与标尺；
> 2. **给模型执行**——作为「细节打磨 / 代码清洗」任务的硬约束与验收清单，可直接交给低成本模型按节执行。
>
> dev 端口 **21073**。每次改动后必须跑 `npm run lint` + `npm run build` 验证。

---

## 一、设计语言（母题）

**「日系柔和卡片」**——克制、轻盈、信息优先。

- **气质**：低饱和的晨光配色、大圆角、极淡阴影、留白充足；不喧宾夺主，让内容（备考数据）说话。
- **结构**：一切皆「卡片」。页面 = 标题区(PageHeader) + 若干卡片(AppCard / StatCard)，卡片内再用 SectionHeader 分组。
- **层级靠对比而非装饰**：用字重和留白拉开层级，不用花哨边框/重阴影/emoji 堆砌。
- **去「萌系噪点」**：删除 ✿ 等装饰符号、弱化时钟等非核心元素，避免可爱过度。

---

## 二、十条设计准则（Do / Don't）

1. **只用 token，不写魔法数字**。颜色/间距/圆角/字号/字重/阴影一律引用 `var(--*)`，禁止散落的 `#hex`、`20px`、`box-shadow: 0 ...`。
2. **双色分工不可混用**：`--color-brand`(蓝)=导航/链接/信息；`--color-accent`(行动色)=主按钮/选中态/焦点/进度。CTA 永远用 accent，导航永远用 brand。
3. **间距走 4px 网格**：只用 `--space-1..10`，不出现 `13px`、`18px` 这类非网格值。
4. **圆角分级**：控件/小卡 `--radius-sm/md`，大卡 `--radius-lg`，胶囊 `--radius-pill`。同层级圆角必须一致。
5. **阴影只有三档**：`--shadow-xs`(卡片默认)、`--shadow-sm`(悬浮)、`--shadow-md`(浮层)。不自定义新阴影。
6. **字重偏厚**：标题 `--weight-black(800)`，强调 `--weight-bold/semibold`，正文 `--weight-normal`。日系厚重感来自字重，不来自字号膨胀。
7. **结构用母题组件**：新页面必须由 `src/components/ui` 的 Page/PageHeader/AppCard/SectionHeader/StatCard 等拼装，禁止重新发明卡片/标题。
8. **antd 风格走 ConfigProvider**，不要逐个组件写 `!important` 覆盖（详见第六节迁移目标）。
9. **只动表现层**：业务逻辑 / API / state / localStorage / 事件处理一律不动。
10. **删 CSS 要保守**：只删「整条选择器、全类无引用」的死规则；动态拼接类名（如 `ui-card-${variant}`）不能误删；用计算样式比对验证。

---

## 三、Token 三层架构（唯一真源）

改样式前先读这三个文件，永远在「正确的层」上改：

| 层 | 文件 | 职责 | 谁可以改 |
|----|------|------|----------|
| 基础层 | `src/theme/tokens.css` | 间距/圆角/阴影/字号/字重/动效「标尺」，**不随主题变** | 极少动 |
| 主题层 | `src/theme/palettes.js` | 4 套配色(aozora/sakura/matcha/sumi)的色值，经 `applyThemeConfig` 注入 `:root` | 调色时 |
| 语义层 | `src/styles/theme.css` | `--color-surface/border/text-*/brand/accent/...`，指向主题层或基础层 | 组件只引用这一层 |

**标尺速查**（来自 tokens.css）：

- 间距：`--space-1`=4 / `-2`=8 / `-3`=12 / `-4`=16 / `-5`=20 / `-6`=24 / `-8`=32 / `-10`=40
- 圆角：`--radius-sm`=8 / `-md`=12 / `-lg`=16 / `-xl`=20 / `-pill`=999
- 阴影：`--shadow-xs/sm/md`
- 字号：`--font-xs`=12 → `--font-2xl`=28
- 字重：`--weight-normal`=400 … `--weight-black`=800
- 动效：`--transition-fast`=160ms / `--transition-base`=220ms

CSS 导入顺序（`src/App.css`）：`tokens → theme → layout → components → dashboard → pomodoro → pages → ui.css`。ui.css 在最后，主题权重最高。

---

## 四、双色语义系统

| 语义 token | 含义 | 用在哪 |
|-----------|------|--------|
| `--color-brand` | 品牌蓝 | 导航高亮、链接、eyebrow 小标签、信息性图标 |
| `--color-accent` | 行动色（各主题不同：紫/粉/绿） | 主按钮、选中态、focus 描边、进度、Tab 激活 |
| `--color-accent-soft` | 行动色的极淡底 | 图标块底、选中行底、Tab 激活底 |
| `--color-surface / -muted` | 卡片面 / 次级面 | 卡片背景、表头背景 |
| `--color-border / -soft` | 边框 / 更淡边框 | 卡片描边、分隔线 |
| `--color-text-primary / -secondary` | 主文字 / 次文字 | 标题 / 说明 |

antd 通过 `ThemeProvider.jsx` + `antdTheme.js` 的 `buildAntdTheme` 把 `primary` 设为 accent，统一注入。

---

## 五、母题组件库 `src/components/ui/`

样式集中在 `ui.css`，统一从 `index.js` 导出。新页面拼装范式参考 **`src/pages/StudyLogs.jsx`**。

| 组件 | 用途 |
|------|------|
| `Page` | 页面最外层网格容器（统一 gap + 入场动画） |
| `PageHeader` | 标题 + 描述 + 右侧操作 + 可选 Tabs |
| `AppCard` | 标准卡片，`variant="default\|plain"`（plain=无边框无阴影，用于嵌套分组） |
| `SectionHeader` | 卡片内分组标题（含 accent-soft 图标块） |
| `Toolbar` | 一行操作按钮组 |
| `FormGrid` / `FormCol` | 响应式表单栅格 |
| `StatCard` / `MetricPill` | 指标卡 / 指标胶囊 |
| `EmptyState` | 空态占位 |

---

## 六、代码清洗准则 + 当前待清理「方言」清单

这一节是「细节不满意」的根因所在，也是交给执行模型的**具体任务**。

### 硬约束（任何清洗任务都必须遵守）
- 不动业务逻辑 / API / state / localStorage / 事件；只改样式与类名表现。
- 删 CSS 前 grep 确认零引用；保留动态拼接类名。
- 每完成一节：`npm run lint` + `npm run build` 必须通过，再继续下一节。
- 改动只碰本节列出的文件，避免大范围连锁。

### 待清理项（按优先级，逐条可独立执行）

1. **`src/styles/components.css` 硬编码色 → token**
   - `#f4a2bb` / `#fff7fa` / `#fff2f6`(hover、Tab 激活底) → 应走 `--color-accent` / `--color-accent-soft` 体系。
   - `#f25f8f` / `#e64d7f`(focus 描边、Tab 文字/ink) → `--color-accent`。
   - `#e3e7ef` / `#e5e7ef` / `#edf0f5` / `#e7ebf2`(各种边框) → `--color-border` / `--color-border-soft`。
   - `font-weight: 900` → `--weight-black`(800)，统一字重梯度。
   - 注意：focus 的 `rgba(242,95,143,.12)` 光圈是 accent 派生，需要给 accent 配一个 `-glow` 派生 token 或用 color-mix。

2. **antd `!important` 覆盖收敛到 ConfigProvider**
   - 现状残存：pages.css 33 处、layout.css 14 处、components.css 12 处。
   - 目标：圆角/边框/focus 等能在 `antdTheme.js` 里用 token 配置的，迁过去后删掉对应 `!important` 规则；逐组件验证视觉不变。

3. **其余页面对齐母题组件**
   - 凡仍用 `.glass-card` / 自定义卡片/标题/按钮的页面，改用 `ui` 组件库；以 `StudyLogs.jsx` 为范式。

4. **非 4px 网格间距归一**
   - grep `: 13px` `: 18px` 等非网格值，对到最近的 `--space-*`。

5. **`preview-banner` 暖色块**属功能性提示（预览环境），保留独立配色，不强行 token 化。

---

## 七、执行模型工作流 & 验收清单

交给低成本模型时，按以下循环跑，**一次只做一条待清理项**：

1. 读 `tokens.css` / `theme.css` / `palettes.js` 确认可用 token。
2. grep 目标硬编码值/选择器，确认引用范围。
3. 替换为 token；保守删死规则。
4. `npm run lint && npm run build`。
5. （如能起 dev）切 4 套主题肉眼检查 brand/accent 是否各就各位。
6. 报告：改了哪些文件、删/替了几处、build 结果。

**验收红线**：构建通过、四套主题正常切换、无业务逻辑改动、无新增魔法数字、无误删动态类名。
</content>
</invoke>
