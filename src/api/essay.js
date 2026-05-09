import { api } from "./request";

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

export async function createEssayDocument({ file, title, rawText, documentRole, sourceGroup }) {
  const formData = new FormData();
  if (file) formData.append("file", file);
  if (title) formData.append("title", title);
  if (rawText) formData.append("raw_text", rawText);
  if (documentRole) formData.append("document_role", documentRole);
  if (sourceGroup) formData.append("source_group", sourceGroup);

  const response = await fetch(apiUrl("/essay/documents"), {
    method: "POST",
    body: formData,
  });
  return unwrapResponse(response);
}

export const getEssayDocuments = () => api.get("/essay/documents");
export const parseEssayDocument = (id, data = {}) => api.post(`/essay/documents/${id}/parse`, data);
export const getEssayChunks = (id) => api.get(`/essay/documents/${id}/chunks`);
export const classifyEssayChunks = (id, data = {}) => api.post(`/essay/documents/${id}/classify`, data);
export const assembleEssayQuestions = (id) => api.post(`/essay/documents/${id}/assemble`, {});
export const getEssayQuestions = (id) => api.get(`/essay/documents/${id}/questions`);
export const reviewEssayQuestion = (id, data) => api.post(`/essay/questions/${id}/review`, data);
