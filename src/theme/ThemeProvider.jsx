import { useMemo, useState } from "react";
import { ConfigProvider } from "antd";
import { resolvePalette, DEFAULT_PALETTE_KEY } from "./palettes";
import { buildAntdTheme } from "./antdTheme";
import { ThemeContext } from "./themeContext";

// 提供当前主题 key 与切换方法。ThemePanel 切换配色时调用 setPaletteKey，
// 触发 ConfigProvider 重渲染，使所有 antd 组件跟随主题变色。
export default function ThemeProvider({ initialPalette = DEFAULT_PALETTE_KEY, children }) {
  const [paletteKey, setPaletteKey] = useState(initialPalette || DEFAULT_PALETTE_KEY);
  const theme = useMemo(() => buildAntdTheme(resolvePalette(paletteKey)), [paletteKey]);
  const value = useMemo(() => ({ paletteKey, setPaletteKey }), [paletteKey]);

  return (
    <ThemeContext.Provider value={value}>
      <ConfigProvider theme={theme}>{children}</ConfigProvider>
    </ThemeContext.Provider>
  );
}
