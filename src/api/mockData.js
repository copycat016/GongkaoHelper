export const dashboardMock = {
  stats: [
    { label: "今日学习", value: "3h 20m", hint: "比昨日 +42m" },
    { label: "番茄钟", value: "5", hint: "专注轮次" },
    { label: "计划完成", value: "4 / 6", hint: "今日任务" },
    { label: "待复习错题", value: "18", hint: "本周优先" },
  ],
  ocrRecords: ["资料分析图表题 OCR 修正", "言语理解片段阅读", "判断推理图形题"],
  essayRecords: ["综合分析题批改 72 分", "提出对策题批改 78 分"],
  logs: ["08:30-09:20 资料分析", "10:10-11:00 申论材料阅读", "14:00-14:50 错题复盘"],
};

export const providersMock = [
  { id: 1, name: "OpenAI", baseUrl: "https://api.openai.com/v1", enabled: true, note: "高质量任务" },
  { id: 2, name: "Local LLM", baseUrl: "http://localhost:11434/v1", enabled: false, note: "本地测试" },
];

export const modelsMock = [
  { id: 1, provider: "OpenAI", name: "gpt-4.1", alias: "高质量解析", cost: "高", speed: "中", quality: "高", enabled: true },
  { id: 2, provider: "Local LLM", name: "qwen2.5", alias: "快速草稿", cost: "低", speed: "高", quality: "中", enabled: false },
];

export const promptsMock = [
  { id: 1, type: "OCR 自动纠错", name: "题面纠错", model: "高质量解析", version: "v1.0", enabled: true },
  { id: 2, type: "申论批改", name: "申论分项评分", model: "高质量解析", version: "v1.2", enabled: true },
  { id: 3, type: "学习日志总结", name: "日总结", model: "快速草稿", version: "v0.8", enabled: false },
];

export const questionsMock = [
  { id: 1, subject: "行测", level1: "资料分析", level2: "增长率", title: "根据材料计算同比增长率", answer: "B", difficulty: "中等", tags: ["计算", "图表"] },
  { id: 2, subject: "行测", level1: "判断推理", level2: "图形推理", title: "选择最合适的图形", answer: "D", difficulty: "偏难", tags: ["空间", "规律"] },
  { id: 3, subject: "申论", level1: "归纳概括题", level2: "材料概括", title: "概括基层治理问题", answer: "要点完整", difficulty: "中等", tags: ["材料定位"] },
];

export const mistakesMock = [
  { id: 1, title: "资料分析增长率误选", type: "资料分析", reason: "计算错误", mastery: "模糊", reviews: 2, nextReview: "2026-05-10" },
  { id: 2, title: "言语理解主旨题", type: "言语理解与表达", reason: "题意理解错误", mastery: "未掌握", reviews: 1, nextReview: "2026-05-09" },
];

export const logsMock = [
  { id: 1, start: "08:30", end: "09:20", duration: "50m", type: "刷题", subject: "行测", questionType: "资料分析", source: "题库", note: "增长率专项" },
  { id: 2, start: "10:10", end: "11:00", duration: "50m", type: "申论", subject: "申论", questionType: "综合分析题", source: "PDF", note: "材料阅读" },
  { id: 3, start: "14:00", end: "14:25", duration: "25m", type: "复习", subject: "行测", questionType: "错题", source: "错题库", note: "逻辑判断" },
];
