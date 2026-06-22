import { api, authHeaders } from "./request";

export const parsePdf = () => api.post("/pdf/parse", null, { mock: { pages: [{ page: 1, text: "PDF 第 1 页解析文本 mock" }] } });

export const getPdfParserInfo = () => api.get("/pdf/parser-info");

export async function parsePdfTest(file) {
  const formData = new FormData();
  formData.append("file", file);

  const response = await fetch("/api/pdf/parse-test", {
    method: "POST",
    headers: authHeaders(),
    body: formData,
  });
  const result = await response.json().catch(() => ({}));
  if (!response.ok || result?.code > 0) {
    throw new Error(result?.message || "PDF 解析测试失败");
  }
  return Object.prototype.hasOwnProperty.call(result, "data") ? result.data : result;
}

export async function parseDocumentTool({ file, rawText, ocrJson, adapter } = {}) {
  const formData = new FormData();
  if (file) formData.append("file", file);
  if (rawText) formData.append("raw_text", rawText);
  if (ocrJson) formData.append("ocr_json", ocrJson);
  if (adapter) formData.append("adapter", adapter);

  const response = await fetch("/api/pdf/parse-tool", {
    method: "POST",
    headers: authHeaders(),
    body: formData,
  });
  const result = await response.json().catch(() => ({}));
  if (!response.ok || result?.code > 0) {
    throw new Error(result?.message || "文档解析失败");
  }
  return Object.prototype.hasOwnProperty.call(result, "data") ? result.data : result;
}
