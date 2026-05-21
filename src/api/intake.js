import { runOcr } from "./ocr";
import { parsePdfTest } from "./pdf";

export async function intakeFromImage({ file, scene, engine }) {
  const result = await runOcr({ file, scene, engine });
  return normalizeIntakeResult(result, {
    source: "image_ocr",
    fileName: file?.name,
    qualityReason: result?.from_cache ? "OCR cache hit" : "OCR completed",
  });
}

export async function intakeFromPdf(file) {
  const result = await parsePdfTest(file);
  const pageText = (result.pages || [])
    .map((page) => `--- page ${page.page_no || page.PageNo || page.page || ""} ---\n${page.text || page.Text || ""}`)
    .join("\n\n")
    .trim();
  return normalizeIntakeResult(
    { ...result, text: result.text || pageText },
    {
      source: "pdf_file",
      fileName: result.file_name || file?.name,
      qualityReason: result.quality?.reason || "PDF text parsed",
    }
  );
}

export function intakeFromText(text) {
  return normalizeIntakeResult(
    {
      text,
      quality: {
        ok: Boolean(text.trim()),
        reason: text.trim() ? "raw text provided" : "empty text",
      },
    },
    { source: "pasted_text" }
  );
}

export function normalizeIntakeResult(result = {}, fallback = {}) {
  const text = result.text || "";
  return {
    source: result.source || fallback.source || "unknown",
    source_engine: result.source_engine || result.engine || fallback.sourceEngine || "",
    file_name: result.file_name || fallback.fileName || "",
    text,
    editable_text: text,
    char_count: result.total_chars ?? countChars(text),
    line_count: result.line_count ?? countLines(text),
    quality: result.quality || {
      ok: Boolean(text.trim()),
      reason: fallback.qualityReason || "text captured",
    },
    pages: result.pages || [],
    raw: result,
  };
}

function countChars(text) {
  return Array.from(text || "").length;
}

function countLines(text) {
  return (text || "").split("\n").filter((line) => line.trim()).length;
}
