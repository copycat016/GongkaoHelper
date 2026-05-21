import { api } from "./request";

export const listStageGoals = () => api.get("/planning/stage-goals");
export const createStageGoal = (data) => api.post("/planning/stage-goals", data);
export const updateStageGoal = (id, data) => api.put(`/planning/stage-goals/${id}`, data);
export const deleteStageGoal = (id) => api.delete(`/planning/stage-goals/${id}`);

export const listWeeklyTasks = (params) => api.get("/planning/weekly-tasks", params);
export const createWeeklyTask = (data) => api.post("/planning/weekly-tasks", data);
export const updateWeeklyTask = (id, data) => api.put(`/planning/weekly-tasks/${id}`, data);
export const deleteWeeklyTask = (id) => api.delete(`/planning/weekly-tasks/${id}`);

export const listPlanningDailyTasks = (params) => api.get("/planning/daily-tasks", params);
export const createPlanningDailyTask = (data) => api.post("/planning/daily-tasks", data);
export const updatePlanningDailyTask = (id, data) => api.put(`/planning/daily-tasks/${id}`, data);
export const togglePlanningDailyTask = (id) => api.post(`/planning/daily-tasks/${id}/toggle`);
export const deletePlanningDailyTask = (id) => api.delete(`/planning/daily-tasks/${id}`);
