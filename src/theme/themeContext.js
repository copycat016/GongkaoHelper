import { createContext, useContext } from "react";
import { DEFAULT_PALETTE_KEY } from "./palettes";

// 当前主题 key 与切换方法的 Context（与 ThemeProvider 拆开，便于 fast-refresh）。
export const ThemeContext = createContext({
  paletteKey: DEFAULT_PALETTE_KEY,
  setPaletteKey: () => {},
});

export function useThemePalette() {
  return useContext(ThemeContext);
}
