export const dashboardMock = {
  stats: [
    { label: "今日学习", value: "3h 20m", hint: "比昨日 +42m" },
    { label: "番茄钟", value: "5", hint: "专注轮次" },
    { label: "最近申论", value: "综合分析题", hint: "批改记录" },
    { label: "最近录入", value: "0", hint: "待接入统计" },
  ],
  ocrRecords: ["申论材料 PDF 解析", "图片 OCR 修正", "粘贴文本整理"],
  essayRecords: ["综合分析题批改 72 分", "提出对策题批改 78 分"],
  logs: ["08:30-09:20 资料分析", "10:10-11:00 申论材料阅读", "14:00-14:50 申论复盘"],
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

export const logsMock = [
  { id: 1, start: "08:30", end: "09:20", duration: "50m", type: "刷题", subject: "行测", questionType: "资料分析", source: "手动补登", note: "增长率专项" },
  { id: 2, start: "10:10", end: "11:00", duration: "50m", type: "申论", subject: "申论", questionType: "综合分析题", source: "PDF", note: "材料阅读" },
  { id: 3, start: "14:00", end: "14:25", duration: "25m", type: "复习", subject: "申论", questionType: "复盘", source: "申论批改", note: "要点整理" },
];
