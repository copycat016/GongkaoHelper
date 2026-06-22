import { api, authHeaders } from "./request";

export const getOcrEngines = () => api.get("/ocr/engines");
export const getOcrScenes = () => api.get("/ocr/scenes");
export const getOcrMonthUsage = () => api.get("/ocr/usage/month");
export const getOcrConfig = () => api.get("/ocr/config");
export const updateOcrConfig = (data) => api.put("/ocr/config", data);

export async function runOcr({ file, scene, engine }) {
  if (!file) {
    return { text: "【OCR 原文】某单位 2025 年业务量同比增长 12.5%，问增长量约为多少？" };
  }

  const formData = new FormData();
  formData.append("file", file);
  formData.append("scene", scene || "printed");
  formData.append("engine", engine || "general_basic");

  const response = await fetch("/api/ocr/recognize", {
    method: "POST",
    headers: authHeaders(),
    body: formData,
  });
  const result = await response.json().catch(() => ({}));
  if (!response.ok || result?.code > 0) {
    throw new Error(result?.message || "OCR 识别失败");
  }
  return result.data;
}

export const correctOcrText = (text) => api.post("/ocr/correct", { text }, { mock: { text: "某单位 2025 年业务量同比增长 12.5%，请根据材料计算增长量。" } });
