import { api } from "./request";

const STORAGE_KEY = "aozora-theme";

export const getThemeConfig = () => {
  const raw = localStorage.getItem(STORAGE_KEY);
  return raw ? JSON.parse(raw) : {};
};

export const saveThemeConfig = (config) => {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(config));
  } catch (error) {
    throw new Error("背景图片过大，已超出浏览器本地存储限制");
  }
  return api.post("/theme", config, { mock: config });
};
