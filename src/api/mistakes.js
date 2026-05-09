import { api } from "./request";

export const getMistakes = (params) => api.get("/mistakes", params);
export const getMistake = (id) => api.get(`/mistakes/${id}`);
export const createMistake = (data) => api.post("/mistakes", data);
export const updateMistake = (id, data) => api.put(`/mistakes/${id}`, data);
export const deleteMistake = (id) => api.delete(`/mistakes/${id}`);
export const reviewMistake = (id, data) => api.post(`/mistakes/${id}/review`, data);
