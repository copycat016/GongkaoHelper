import { api } from "./request";

export const getPrompts = (params) => api.get("/prompts", params);
export const createPrompt = (data) => api.post("/prompts", data);
export const updatePrompt = (id, data) => api.put(`/prompts/${id}`, data);
export const deletePrompt = (id) => api.delete(`/prompts/${id}`);
export const testPrompt = (data) => api.post("/prompts/test", data, { mock: { output: "这是 mock 的 Prompt 测试结果。" } });
