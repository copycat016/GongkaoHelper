import { api } from "./request";

export const getProviders = () => api.get("/llm/providers");
export const createProvider = (data) => api.post("/llm/providers", data);
export const updateProvider = (id, data) => api.put(`/llm/providers/${id}`, data);
export const deleteProvider = (id) => api.delete(`/llm/providers/${id}`);
export const fetchProviderModels = (id) => api.get(`/llm/providers/${id}/models`);

export const getModels = () => api.get("/llm/models");
export const createModel = (data) => api.post("/llm/models", data);
export const updateModel = (id, data) => api.put(`/llm/models/${id}`, data);
export const deleteModel = (id) => api.delete(`/llm/models/${id}`);
