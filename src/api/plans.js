import { api } from "./request";

export const getPlans = (params) => api.get("/plans", params);
export const createPlan = (data) => api.post("/plans", data);
export const updatePlan = (id, data) => api.put(`/plans/${id}`, data);
export const deletePlan = (id) => api.delete(`/plans/${id}`);
export const completePlan = (id) => api.post(`/plans/${id}/complete`);
