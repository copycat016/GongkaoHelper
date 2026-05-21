import { api } from "./request";

// 后端约定见 backend/internal/handlers/tasks.go：
//   GET    /api/tasks?date=YYYY-MM-DD     列出当日任务
//   GET    /api/tasks/summary?date=...    当日完成情况摘要
//   POST   /api/tasks                     创建任务
//   PUT    /api/tasks/:id                 更新任务
//   POST   /api/tasks/:id/toggle          翻转完成状态
//   DELETE /api/tasks/:id                 删除任务

export const listDailyTasks = (params) => api.get("/tasks", params);
export const getDailyTaskSummary = (params) => api.get("/tasks/summary", params);
export const createDailyTask = (data) => api.post("/tasks", data);
export const updateDailyTask = (id, data) => api.put(`/tasks/${id}`, data);
export const toggleDailyTask = (id) => api.post(`/tasks/${id}/toggle`);
export const deleteDailyTask = (id) => api.delete(`/tasks/${id}`);
