import { Card } from "antd";

// 统一卡片母题：封装 antd Card + 日系毛玻璃外观（边框/圆角/柔和阴影由 ui.css 的语义 token 控制）。
// 替代散乱的 glass-card / fresh-card / dashboard-hero / dashboard-stage-card。
// variant: "default" | "plain"（plain = 无边框无阴影，用于嵌套/分组）。
// title / extra / 等 antd Card 属性经 ...rest 透传。
export default function AppCard({
  children,
  className = "",
  padding,
  variant = "default",
  ...rest
}) {
  const cls = ["ui-card", variant !== "default" && `ui-card-${variant}`, className]
    .filter(Boolean)
    .join(" ");
  const styles = padding != null ? { body: { padding } } : undefined;
  return (
    <Card className={cls} styles={styles} {...rest}>
      {children}
    </Card>
  );
}
