import { Card } from "antd";

// 统计卡母题：label + value + hint + icon。
export default function StatCard({ label, value, hint, icon, className = "" }) {
  return (
    <Card className={`ui-card ui-stat-card ${className}`.trim()}>
      <div className="ui-stat-label">
        {icon}
        {label && <span>{label}</span>}
      </div>
      <div className="ui-stat-value">{value}</div>
      {hint && <div className="ui-stat-hint">{hint}</div>}
    </Card>
  );
}

// 紧凑度量胶囊，提取自 Dashboard MetricPill。
export function MetricPill({ label, value, className = "" }) {
  return (
    <div className={`ui-metric-pill ${className}`.trim()}>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
