import { api } from "./request";

export const getQuestions = () => api.get("/questions");
export const getQuestion = (id) => api.get(`/questions/${id}`);
export const saveQuestion = (data) => api.post("/questions", data, { mock: data });
export const updateQuestion = (id, data) => api.put(`/questions/${id}`, data);
export const deleteQuestion = (id) => api.delete(`/questions/${id}`);
