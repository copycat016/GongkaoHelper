import { api } from "./request";

export const getLogs = (params) => api.get("/logs", params);
export const getLogStats = (params) => api.get("/logs/stats", params);
export const saveLog = (data) => api.post("/logs", data);
