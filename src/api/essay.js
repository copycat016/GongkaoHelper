import { api, authHeaders } from "./request";

function apiUrl(path) {
  return `/api${path}`;
}

async function unwrapResponse(response) {
  const result = await response.json().catch(() => ({}));
  if (!response.ok || result?.code > 0) {
    throw new Error(result?.message || "请求失败");
  }
  return Object.prototype.hasOwnProperty.call(result, "data") ? result.data : result;
}

export async function createEssayDocument({ file, title, rawText, documentRole, sourceGroup, boundaryModelId }) {
  const formData = new FormData();
  if (file) formData.append("file", file);
  if (title) formData.append("title", title);
  if (rawText) formData.append("raw_text", rawText);
  if (documentRole) formData.append("document_role", documentRole);
  if (sourceGroup) formData.append("source_group", sourceGroup);
  if (boundaryModelId) formData.append("boundary_model_id", String(boundaryModelId));

  const response = await fetch(apiUrl("/essay/documents"), {
    method: "POST",
    headers: authHeaders(),
    body: formData,
  });
  return unwrapResponse(response);
}

export const getEssayDocuments = () => api.get("/essay/documents");
export const deleteEssayDocument = (id) => api.delete(`/essay/documents/${id}`);
export const parseEssayDocument = (id, data = {}) => api.post(`/essay/documents/${id}/parse`, data);
export const debugEssayBoundary = (id, data = {}) => api.post(`/essay/documents/${id}/debug-boundary`, data);
export const getEssaySections = (id) => api.get(`/essay/documents/${id}/sections`);
export const getEssayChunks = (id) => api.get(`/essay/documents/${id}/chunks`);
export const classifyEssayChunks = (id, data = {}) => api.post(`/essay/documents/${id}/classify`, data);
export const assembleEssayQuestions = (id) => api.post(`/essay/documents/${id}/assemble`, {});
export const getEssayQuestions = (id) => api.get(`/essay/documents/${id}/questions`);
export const createEssayQuestion = (data) => api.post("/essay/questions", data);
export const updateEssayQuestion = (id, data) => api.put(`/essay/questions/${id}`, data);
export const deleteEssayQuestion = (id) => api.delete(`/essay/questions/${id}`);
export const updateEssaySection = (id, data) => api.put(`/essay/sections/${id}`, data);
export const replaceEssayQuestionRelations = (id, data) => api.post(`/essay/questions/${id}/relations`, data);
export const reviewEssayQuestion = (id, data) => api.post(`/essay/questions/${id}/review`, data);
