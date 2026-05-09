import { Card, Typography } from "antd";

const { Text } = Typography;

function StatCard({ label, value, hint, icon }) {
  return (
    <Card className="glass-card stat-card" bordered={false}>
      <div className="stat-card-label">
        {icon}
        <span>{label}</span>
      </div>
      <div className="stat-card-value">{value}</div>
      {hint && <Text type="secondary">{hint}</Text>}
    </Card>
  );
}

export default StatCard;
