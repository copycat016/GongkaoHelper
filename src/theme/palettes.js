// 多主题配色表（主题层 token 的数据源）。
// 每套主题除原有 8 个颜色外，新增 actionPink（行动色），
// 用于主按钮 / 选中态 / 焦点等 CTA 场景；品牌蓝 primaryBlue 仍用于导航 / 链接 / 信息性元素。
export const palettes = {
  aozora: {
    label: "Aozora Blue",
    bgMain: "#f5f9ff",
    bgMaskRgb: "245, 249, 255",
    primaryBlue: "#6aa7e8",
    deepBlueGray: "#2f4058",
    lightBlue: "#dceeff",
    accentPurple: "#b9a7f5",
    accentPink: "#f5b7cf",
    actionPink: "#b9a7f5",
    accentSoft: "#f0ebf8",
    textBlueGray: "#5d6f86",
  },
  sakura: {
    label: "Sakura Mist",
    bgMain: "#fff7fb",
    bgMaskRgb: "255, 247, 251",
    primaryBlue: "#7aa8f5",
    deepBlueGray: "#3f4058",
    lightBlue: "#eef3ff",
    accentPurple: "#c6b5ff",
    accentPink: "#f4a9c6",
    actionPink: "#f0609b",
    accentSoft: "#fff2f6",
    textBlueGray: "#75677a",
  },
  matcha: {
    label: "Matcha Morning",
    bgMain: "#f6fbf5",
    bgMaskRgb: "246, 251, 245",
    primaryBlue: "#73b7c8",
    deepBlueGray: "#314c4f",
    lightBlue: "#e2f4f5",
    accentPurple: "#b8c9ef",
    accentPink: "#f2b8bd",
    actionPink: "#5ea88c",
    accentSoft: "#eef8f5",
    textBlueGray: "#607b75",
  },
  sumi: {
    label: "Sumi Soft",
    bgMain: "#f7f8fb",
    bgMaskRgb: "247, 248, 251",
    primaryBlue: "#7d9fd8",
    deepBlueGray: "#313949",
    lightBlue: "#e6ecf7",
    accentPurple: "#a99fe2",
    accentPink: "#e8aebc",
    actionPink: "#8b7fc8",
    accentSoft: "#f3f1f8",
    textBlueGray: "#687287",
  },
};

export const defaultConfig = {
  palette: "aozora",
  backgroundEnabled: false,
  backgroundImage: "",
  blur: 0,
  brightness: 100,
  maskOpacity: 34,
  backgroundSize: "cover",
  backgroundPosition: "center",
  cardOpacity: 72,
  dockImage: "",
  glassBackground: false,
};

export const DEFAULT_PALETTE_KEY = "aozora";

// 取一套有效的 palette，无效 key 时回退到默认主题。
export function resolvePalette(key) {
  return palettes[key] || palettes[DEFAULT_PALETTE_KEY];
}
