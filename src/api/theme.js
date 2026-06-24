import { api, getAccessToken } from "./request";

const STORAGE_KEY = "aozora-theme";

export const getThemeConfig = async () => {
  try {
    if (getAccessToken()) {
      const serverConfig = await api.get("/theme");
      if (serverConfig) {
        const config = {
          palette: serverConfig.palette || "aozora",
          backgroundEnabled: serverConfig.background_enabled || false,
          backgroundImage: serverConfig.background_image || "",
          blur: serverConfig.blur || 0,
          brightness: serverConfig.brightness || 100,
          maskOpacity: serverConfig.mask_opacity || 34,
          backgroundSize: serverConfig.background_size || "cover",
          backgroundPosition: serverConfig.background_position || "center",
          cardOpacity: serverConfig.card_opacity || 72,
          dockImage: serverConfig.dock_image || "",
          glassBackground: serverConfig.glass_background || false,
        };
        localStorage.setItem(STORAGE_KEY, JSON.stringify(config));
        return config;
      }
    }
  } catch (error) {
    console.warn("从服务器获取主题配置失败，使用本地缓存", error);
  }

  const raw = localStorage.getItem(STORAGE_KEY);
  return raw ? JSON.parse(raw) : {};
};

export const saveThemeConfig = async (config) => {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(config));
  } catch (error) {
    throw new Error("背景图片过大，已超出浏览器本地存储限制", { cause: error });
  }

  const serverConfig = {
    palette: config.palette,
    background_enabled: config.backgroundEnabled,
    background_image: config.backgroundImage,
    blur: config.blur,
    brightness: config.brightness,
    mask_opacity: config.maskOpacity,
    background_size: config.backgroundSize,
    background_position: config.backgroundPosition,
    card_opacity: config.cardOpacity,
    dock_image: config.dockImage,
    glass_background: config.glassBackground,
  };

  return api.post("/theme", serverConfig);
};
