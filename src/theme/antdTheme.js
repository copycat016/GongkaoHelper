// 从一套 palette 生成 antd ConfigProvider 的 theme 配置。
// 这是「单一数据源」的关键：antd 组件的颜色 / 圆角 / 焦点全部由 palette 决定，
// 从而替代 components.css 里那一大坨针对 antd 的 !important 覆盖。
//
// 语义约定：
//   colorPrimary = 行动色（粉，CTA / 主按钮 / 选中态）
//   colorInfo / colorLink = 品牌蓝（导航 / 链接 / 信息性）

const FONT_FAMILY =
  'Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", "Microsoft YaHei", sans-serif';

const SURFACE = "#ffffff";
const SURFACE_MUTED = "#f7f8fb";
const BORDER = "#e5e7ef";
const BORDER_SECONDARY = "#edf0f5";

// 给 6 位 hex 颜色加 alpha，返回 8 位 hex（用于焦点环等半透明色）。
function withAlpha(hex, alpha) {
  const a = Math.round(Math.min(1, Math.max(0, alpha)) * 255)
    .toString(16)
    .padStart(2, "0");
  return `${hex}${a}`;
}

export function buildAntdTheme(palette) {
  const action = palette.actionPink;
  const brand = palette.primaryBlue;

  return {
    token: {
      colorPrimary: action,
      colorInfo: brand,
      colorLink: brand,
      colorLinkHover: action,

      colorText: palette.deepBlueGray,
      colorTextHeading: palette.deepBlueGray,
      colorTextSecondary: palette.textBlueGray,
      colorTextTertiary: palette.textBlueGray,

      colorBgLayout: palette.bgMain,
      colorBgContainer: SURFACE,
      colorBgElevated: SURFACE,

      colorBorder: BORDER,
      colorBorderSecondary: BORDER_SECONDARY,

      borderRadius: 12,
      borderRadiusLG: 12,
      borderRadiusSM: 8,

      fontFamily: FONT_FAMILY,
      fontSize: 14,

      controlOutline: withAlpha(action, 0.12),
      controlOutlineWidth: 3,

      boxShadow: "0 1px 2px rgba(24, 32, 56, 0.04)",
      boxShadowSecondary: "0 4px 12px rgba(70, 118, 168, 0.08)",
      wireframe: false,
    },
    components: {
      Card: {
        headerFontSize: 16,
        headerBg: "transparent",
        colorBorderSecondary: BORDER_SECONDARY,
      },
      Button: {
        primaryShadow: "none",
        defaultShadow: "none",
      },
      Tabs: {
        itemColor: palette.textBlueGray,
        itemHoverColor: action,
        itemSelectedColor: action,
        inkBarColor: action,
      },
      Table: {
        headerBg: SURFACE_MUTED,
        headerColor: palette.textBlueGray,
        rowHoverBg: withAlpha(action, 0.06),
        borderColor: BORDER_SECONDARY,
        headerSplitColor: "transparent",
      },
      Input: {
        activeBorderColor: action,
        hoverBorderColor: withAlpha(action, 0.66),
      },
      Select: {
        optionSelectedBg: withAlpha(action, 0.10),
      },
      Segmented: {
        itemSelectedBg: SURFACE,
        trackBg: SURFACE_MUTED,
      },
    },
  };
}
