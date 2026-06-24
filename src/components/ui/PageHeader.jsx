// 页面标题区母题：eyebrow + 标题 + 描述 + 右侧操作 + 可选标签栏。
// 重做自坏掉的旧 components/PageHeader.jsx（旧版只渲染 extra、标题被 CSS 隐藏）。
export default function PageHeader({
  eyebrow,
  title,
  description,
  icon,
  actions,
  tabs,
  className = "",
}) {
  return (
    <header className={`ui-page-header ${className}`.trim()}>
      <div className="ui-page-header-bar">
        <div className="ui-page-header-lead">
          {icon && <div className="ui-page-header-icon">{icon}</div>}
          <div className="ui-page-header-text">
            {eyebrow && <div className="ui-eyebrow">{eyebrow}</div>}
            {title && <h1 className="ui-page-title">{title}</h1>}
            {description && <p className="ui-page-desc">{description}</p>}
          </div>
        </div>
        {actions && <div className="ui-page-header-actions">{actions}</div>}
      </div>
      {tabs && <div className="ui-page-header-tabs">{tabs}</div>}
    </header>
  );
}
