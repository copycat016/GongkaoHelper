// 页面顶层容器：统一纵向间距与入场动画。替代各页 `.page-grid .xxx-page`。
export default function Page({ children, gap, className = "", ...rest }) {
  const style = gap != null ? { gap } : undefined;
  return (
    <div className={`ui-page ${className}`.trim()} style={style} {...rest}>
      {children}
    </div>
  );
}
