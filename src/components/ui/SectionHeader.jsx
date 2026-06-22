// 区块/卡片内标题母题：icon + eyebrow + 标题 + 右侧 meta/操作。
// 提取自 Dashboard 的 PanelHeader，推广为全局母题。
export default function SectionHeader({
  icon,
  eyebrow,
  title,
  meta,
  action,
  className = "",
}) {
  return (
    <div className={`ui-section-header ${className}`.trim()}>
      <div className="ui-section-header-title">
        {icon && <span className="ui-section-header-icon">{icon}</span>}
        <div>
          {eyebrow && <div className="ui-eyebrow">{eyebrow}</div>}
          {title && <h3 className="ui-section-title">{title}</h3>}
        </div>
      </div>
      {(meta || action) && (
        <div className="ui-section-header-tools">
          {meta && <span className="ui-section-meta">{meta}</span>}
          {action}
        </div>
      )}
    </div>
  );
}
