// 操作按钮行/工具条母题：统一对齐与间距。替代各页手写的 `.xxx-actions` 容器。
// align: "start" | "end" | "center" | "between"
const JUSTIFY = {
  start: "flex-start",
  end: "flex-end",
  center: "center",
  between: "space-between",
};

export default function Toolbar({
  children,
  align = "start",
  gap,
  wrap = true,
  className = "",
  ...rest
}) {
  const style = {
    justifyContent: JUSTIFY[align] || "flex-start",
    flexWrap: wrap ? "wrap" : "nowrap",
    ...(gap != null ? { gap } : null),
  };
  return (
    <div className={`ui-toolbar ${className}`.trim()} style={style} {...rest}>
      {children}
    </div>
  );
}
