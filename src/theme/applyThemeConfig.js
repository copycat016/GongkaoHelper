import { palettes, defaultConfig, resolvePalette } from "./palettes";

// 兼容旧的 import 路径（ThemePanel 等仍从这里取 palettes / defaultConfig）。
export { palettes, defaultConfig };

// 把一套主题配置注入到 :root 的 CSS 变量。
// 这些变量供「非 antd 的自定义元素」使用；antd 组件则由 ThemeProvider 的 ConfigProvider 接管。
// 两者共享同一份 palettes.js 数据源，避免颜色再次打架。
export function applyThemeConfig(config) {
  const next = { ...defaultConfig, ...config };
  const palette = resolvePalette(next.palette);
  const root = document.documentElement;
  root.style.setProperty("--bg-main", palette.bgMain);
  root.style.setProperty("--bg-mask-rgb", palette.bgMaskRgb);
  root.style.setProperty("--primary-blue", palette.primaryBlue);
  root.style.setProperty("--deep-blue-gray", palette.deepBlueGray);
  root.style.setProperty("--light-blue", palette.lightBlue);
  root.style.setProperty("--accent-purple", palette.accentPurple);
  root.style.setProperty("--accent-pink", palette.accentPink);
  root.style.setProperty("--action-pink", palette.actionPink);
  root.style.setProperty("--color-accent-soft", palette.accentSoft);
  root.style.setProperty("--text-blue-gray", palette.textBlueGray);
  root.style.setProperty("--line-soft", `${palette.primaryBlue}29`);
  root.style.setProperty("--bg-blur", `${next.blur}px`);
  root.style.setProperty("--bg-brightness", `${next.brightness / 100}`);
  root.style.setProperty("--bg-mask-opacity", `${next.maskOpacity / 100}`);
  root.style.setProperty("--card-opacity", `${next.cardOpacity / 100}`);
  root.style.setProperty(
    "--bg-image",
    next.backgroundEnabled && next.backgroundImage ? `url("${next.backgroundImage}")` : "none"
  );
  root.style.setProperty("--bg-size", next.backgroundSize);
  root.style.setProperty("--bg-position", next.backgroundPosition);
  // 渐变毛玻璃背景模板：开关挂在 :root[data-glass]，CSS 据此叠加主题渐变 + 卡片磨砂。
  root.dataset.glass = next.glassBackground ? "on" : "off";
  // 左下角时钟卡背景图：用户在主题面板上传后注入；未设置时由 CSS 回退到占位图 /dock-bg.jpg。
  if (next.dockImage) {
    root.style.setProperty("--idle-dock-image", `url("${next.dockImage}")`);
  } else {
    root.style.removeProperty("--idle-dock-image");
  }
  return next;
}
