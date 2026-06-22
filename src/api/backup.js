import { authHeaders } from "./request";

export async function exportBackup({ includeSecrets = false } = {}) {
  const response = await fetch(`/api/backup/export?include_secrets=${includeSecrets ? "true" : "false"}`, {
    headers: authHeaders(),
  });
  if (!response.ok) {
    const result = await response.json().catch(() => ({}));
    throw new Error(result?.message || "导出备份失败");
  }
  return response.blob();
}
